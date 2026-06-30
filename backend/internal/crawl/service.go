// Package crawl orchestrates the dndsu crawler against the database: it
// walks the 2014 catalog, follows each card's cross-link to discover its
// 2024 counterpart (see the plan's de-risking notes — next.dnd.su has no
// standalone catalog endpoint we could find), and upserts both editions
// linked to each other. Re-runs diff against stored content hashes so only
// new/changed cards are re-fetched (§8.3).
package crawl

import (
	"context"
	"log"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"monster-screen/backend/internal/dndsu"
	"monster-screen/backend/internal/monsters"
)

type Service struct {
	pool      *pgxpool.Pool
	client    *dndsu.Client
	repo      *monsters.Repo
	baseURL14 string
	baseURL24 string

	mu      sync.Mutex
	running bool
}

func NewService(pool *pgxpool.Pool, client *dndsu.Client, baseURL14, baseURL24 string) *Service {
	return &Service{
		pool:      pool,
		client:    client,
		repo:      monsters.NewRepo(pool),
		baseURL14: baseURL14,
		baseURL24: baseURL24,
	}
}

type RunStats struct {
	Added   int
	Updated int
	Errors  []string
}

// IsRunning reports whether a crawl is currently in progress, so the API can
// reject overlapping triggers instead of racing the rate limiter.
func (s *Service) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// RunSeed does a full crawl: every card is re-fetched and re-parsed
// regardless of whether it looks unchanged. Used for the initial seed and
// for a manual "force resync".
func (s *Service) RunSeed(ctx context.Context) (RunStats, error) {
	return s.run(ctx, "seed", true)
}

// RunIncremental only re-fetches cards whose catalog-level fields changed
// since the last crawl (§8.3 diff). Used by the periodic scheduler and the
// "update now" manual trigger.
func (s *Service) RunIncremental(ctx context.Context) (RunStats, error) {
	return s.run(ctx, "incremental", false)
}

func (s *Service) run(ctx context.Context, kind string, force bool) (RunStats, error) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return RunStats{}, errAlreadyRunning
	}
	s.running = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	runID, err := s.startRun(ctx, kind)
	if err != nil {
		return RunStats{}, err
	}

	stats := RunStats{}
	catalog, err := dndsu.FetchCatalog2014(ctx, s.client, s.baseURL14)
	if err != nil {
		s.finishRun(ctx, runID, "failed", stats)
		return stats, err
	}

	previousHashes, err := s.loadCatalogHashes(ctx)
	if err != nil {
		s.finishRun(ctx, runID, "failed", stats)
		return stats, err
	}

	diff := dndsu.DiffCatalog(catalog, previousHashes, dndsu.HashCatalogEntry)
	toProcess := diff.New
	toProcess = append(toProcess, diff.Changed...)
	if force {
		toProcess = catalog
	}

	for _, entry := range toProcess {
		select {
		case <-ctx.Done():
			s.finishRun(ctx, runID, "cancelled", stats)
			return stats, ctx.Err()
		default:
		}

		isNew := previousHashes[entry.DndsuID] == ""
		if err := s.processEntry(ctx, entry); err != nil {
			log.Printf("crawl: entry %d (%s) failed: %v", entry.DndsuID, entry.Slug, err)
			stats.Errors = append(stats.Errors, entry.Slug+": "+err.Error())
			continue
		}
		if isNew {
			stats.Added++
		} else {
			stats.Updated++
		}
	}

	// Always keep catalog_entries current for the entries we saw, even the
	// unchanged ones, so last_seen_at reflects reality.
	if err := s.touchCatalogEntries(ctx, diff.Same); err != nil {
		log.Printf("crawl: touch unchanged entries: %v", err)
	}

	s.finishRun(ctx, runID, "completed", stats)
	return stats, nil
}

// processEntry fetches and upserts a single monster's 2014 card and, if
// linked, its 2024 counterpart.
func (s *Service) processEntry(ctx context.Context, entry dndsu.CatalogEntry) error {
	card2014, err := dndsu.FetchCard(ctx, s.client, entry.URL, "2014")
	if err != nil {
		return err
	}

	id2014, err := s.upsertMonster(ctx, card2014)
	if err != nil {
		return err
	}
	if err := s.upsertCatalogEntry(ctx, entry, "2014"); err != nil {
		return err
	}

	if card2014.OtherEditionURL == "" {
		return nil
	}

	card2024, err := dndsu.FetchCard(ctx, s.client, card2014.OtherEditionURL, "2024")
	if err != nil {
		// 2024 link existing but unreachable is not fatal to the 2014 record.
		log.Printf("crawl: fetch 2024 counterpart of %s failed: %v", entry.Slug, err)
		return nil
	}
	id2024, err := s.upsertMonster(ctx, card2024)
	if err != nil {
		log.Printf("crawl: upsert 2024 counterpart of %s failed: %v", entry.Slug, err)
		return nil
	}
	if err := s.repo.LinkEditions(ctx, id2014, id2024); err != nil {
		return err
	}

	catalogEntry2024 := dndsu.CatalogEntry{
		DndsuID: card2024.DndsuID,
		Slug:    card2024.Slug,
		URL:     card2024.SourceURL,
		NameRu:  card2024.NameRu,
		NameEn:  card2024.NameEn,
		CR:      card2024.StatBlock.ChallengeRating,
	}
	return s.upsertCatalogEntry(ctx, catalogEntry2024, "2024")
}

var errAlreadyRunning = &crawlError{"a crawl is already running"}

type crawlError struct{ msg string }

func (e *crawlError) Error() string { return e.msg }
