package crawl

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"monster-screen/backend/internal/dndsu"
	"monster-screen/backend/internal/monsters"
)

func (s *Service) startRun(ctx context.Context, kind string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, `INSERT INTO crawl_runs (kind) VALUES ($1) RETURNING id`, kind).Scan(&id)
	return id, err
}

func (s *Service) finishRun(ctx context.Context, id uuid.UUID, status string, stats RunStats) {
	errBytes, _ := json.Marshal(stats.Errors)
	_, err := s.pool.Exec(ctx, `
		UPDATE crawl_runs SET status=$2, added=$3, updated=$4, errors=$5, finished_at=now()
		WHERE id=$1
	`, id, status, stats.Added, stats.Updated, errBytes)
	if err != nil {
		// best-effort logging table; nothing else to do if this write fails
		_ = err
	}
}

// loadCatalogHashes returns dndsu_id -> content_hash for every 2014 catalog
// entry we've stored, so the incremental run can diff against it.
func (s *Service) loadCatalogHashes(ctx context.Context) (map[int]string, error) {
	rows, err := s.pool.Query(ctx, `SELECT dndsu_id, content_hash FROM catalog_entries WHERE edition = '2014'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[int]string{}
	for rows.Next() {
		var id int
		var hash string
		if err := rows.Scan(&id, &hash); err != nil {
			return nil, err
		}
		out[id] = hash
	}
	return out, rows.Err()
}

func (s *Service) touchCatalogEntries(ctx context.Context, entries []dndsu.CatalogEntry) error {
	for _, e := range entries {
		_, err := s.pool.Exec(ctx, `
			UPDATE catalog_entries SET last_seen_at = now() WHERE section = 'monster' AND edition = '2014' AND dndsu_id = $1
		`, e.DndsuID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) upsertCatalogEntry(ctx context.Context, e dndsu.CatalogEntry, edition string) error {
	hash := dndsu.HashCatalogEntry(e)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO catalog_entries (section, edition, dndsu_id, slug, name_ru, name_en, cr, url, is_unique_npc, content_hash, last_seen_at)
		VALUES ('monster', $1, $2, $3, $4, $5, $6, $7, $8, $9, now())
		ON CONFLICT (section, edition, dndsu_id) DO UPDATE SET
			slug = EXCLUDED.slug,
			name_ru = EXCLUDED.name_ru,
			name_en = EXCLUDED.name_en,
			cr = EXCLUDED.cr,
			url = EXCLUDED.url,
			is_unique_npc = EXCLUDED.is_unique_npc,
			content_hash = EXCLUDED.content_hash,
			last_seen_at = now()
	`, edition, e.DndsuID, e.Slug, e.NameRu, e.NameEn, e.CR, e.URL, e.IsUniqueNPC, hash)
	return err
}

func (s *Service) upsertMonster(ctx context.Context, card *dndsu.ParsedCard) (uuid.UUID, error) {
	var imagePtr *string
	if card.ImageURL != "" {
		imagePtr = &card.ImageURL
	}
	var sourcePtr *string
	if card.SourceBook != "" {
		sourcePtr = &card.SourceBook
	}

	return s.repo.Upsert(ctx, monsters.UpsertInput{
		DndsuID:     card.DndsuID,
		Slug:        card.Slug,
		Edition:     card.Edition,
		NameRu:      card.NameRu,
		NameEn:      card.NameEn,
		CR:          card.StatBlock.ChallengeRating,
		CRNumeric:   dndsu.CRToNumeric(card.StatBlock.ChallengeRating),
		Type:        card.StatBlock.Type,
		Size:        card.StatBlock.SizeRu,
		Alignment:   card.StatBlock.Alignment,
		StatBlock:   card.StatBlock,
		ImageURL:    imagePtr,
		SourceBook:  sourcePtr,
		SourceURL:   card.SourceURL,
		RawHTML:     card.RawHTML,
		ContentHash: dndsu.ContentHash(card.RawHTML),
		IsUniqueNPC: card.IsUniqueNPC,
	})
}
