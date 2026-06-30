package monsters

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"monster-screen/backend/internal/statblock"
)

type Monster struct {
	ID               uuid.UUID          `json:"id"`
	DndsuID          int                `json:"dndsuId"`
	Slug             string             `json:"slug"`
	Edition          string             `json:"edition"`
	NameRu           string             `json:"nameRu"`
	NameEn           string             `json:"nameEn"`
	CR               string             `json:"cr"`
	Type             string             `json:"type"`
	Size             string             `json:"size"`
	Alignment        string             `json:"alignment"`
	StatBlock        statblock.StatBlock `json:"statblock"`
	ImageURL         *string            `json:"imageUrl,omitempty"`
	SourceBook       *string            `json:"sourceBook,omitempty"`
	SourceURL        string             `json:"sourceUrl"`
	LinkedMonsterID  *uuid.UUID         `json:"linkedMonsterId,omitempty"`
	IsUniqueNPC      bool               `json:"isUniqueNpc"`
	LastFetchedAt    *time.Time         `json:"lastFetchedAt,omitempty"`
	UpdatedAt        time.Time          `json:"updatedAt"`
}

type UpsertInput struct {
	DndsuID         int
	Slug            string
	Edition         string
	NameRu          string
	NameEn          string
	CR              string
	CRNumeric       *float64
	Type            string
	Size            string
	Alignment       string
	StatBlock       statblock.StatBlock
	ImageURL        *string
	SourceBook      *string
	SourceURL       string
	RawHTML         string
	ContentHash     string
	IsUniqueNPC     bool
}

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

const selectColumns = `id, dndsu_id, slug, edition, name_ru, name_en, cr, type, size, alignment,
	statblock, image_url, source_book, source_url, linked_monster_id, is_unique_npc, last_fetched_at, updated_at`

func scanMonster(row pgx.Row) (Monster, error) {
	var m Monster
	var sb []byte
	if err := row.Scan(&m.ID, &m.DndsuID, &m.Slug, &m.Edition, &m.NameRu, &m.NameEn, &m.CR, &m.Type, &m.Size,
		&m.Alignment, &sb, &m.ImageURL, &m.SourceBook, &m.SourceURL, &m.LinkedMonsterID, &m.IsUniqueNPC,
		&m.LastFetchedAt, &m.UpdatedAt); err != nil {
		return m, err
	}
	if len(sb) > 0 {
		_ = json.Unmarshal(sb, &m.StatBlock)
	}
	return m, nil
}

func (r *Repo) List(ctx context.Context, search, edition, monsterType string, limit, offset int) ([]Monster, error) {
	query := `SELECT ` + selectColumns + ` FROM monsters WHERE 1=1`
	args := []any{}
	idx := 1

	if search != "" {
		query += fmt.Sprintf(` AND (name_ru ILIKE $%d OR name_en ILIKE $%d)`, idx, idx)
		args = append(args, "%"+search+"%")
		idx++
	}
	if edition != "" {
		query += fmt.Sprintf(` AND edition = $%d`, idx)
		args = append(args, edition)
		idx++
	}
	if monsterType != "" {
		query += fmt.Sprintf(` AND type = $%d`, idx)
		args = append(args, monsterType)
		idx++
	}
	query += fmt.Sprintf(` ORDER BY name_ru ASC LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Monster
	for rows.Next() {
		m, err := scanMonster(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (Monster, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+selectColumns+` FROM monsters WHERE id = $1`, id)
	return scanMonster(row)
}

func (r *Repo) GetBySlugEdition(ctx context.Context, slug, edition string) (Monster, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+selectColumns+` FROM monsters WHERE slug = $1 AND edition = $2`, slug, edition)
	return scanMonster(row)
}

func (r *Repo) GetContentHash(ctx context.Context, slug, edition string) (string, error) {
	var hash string
	err := r.pool.QueryRow(ctx, `SELECT content_hash FROM monsters WHERE slug = $1 AND edition = $2`, slug, edition).Scan(&hash)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	return hash, err
}

// Upsert inserts or updates a monster by (slug, edition) and returns its id.
func (r *Repo) Upsert(ctx context.Context, in UpsertInput) (uuid.UUID, error) {
	sb, err := json.Marshal(in.StatBlock)
	if err != nil {
		return uuid.Nil, err
	}

	var id uuid.UUID
	err = r.pool.QueryRow(ctx, `
		INSERT INTO monsters (dndsu_id, slug, edition, name_ru, name_en, cr, cr_numeric, type, size, alignment,
			statblock, image_url, source_book, source_url, raw_html, content_hash, is_unique_npc, last_fetched_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17, now(), now())
		ON CONFLICT (slug, edition) DO UPDATE SET
			dndsu_id = EXCLUDED.dndsu_id,
			name_ru = EXCLUDED.name_ru,
			name_en = EXCLUDED.name_en,
			cr = EXCLUDED.cr,
			cr_numeric = EXCLUDED.cr_numeric,
			type = EXCLUDED.type,
			size = EXCLUDED.size,
			alignment = EXCLUDED.alignment,
			statblock = EXCLUDED.statblock,
			image_url = EXCLUDED.image_url,
			source_book = EXCLUDED.source_book,
			source_url = EXCLUDED.source_url,
			raw_html = EXCLUDED.raw_html,
			content_hash = EXCLUDED.content_hash,
			is_unique_npc = EXCLUDED.is_unique_npc,
			last_fetched_at = now(),
			updated_at = now()
		RETURNING id
	`, in.DndsuID, in.Slug, in.Edition, in.NameRu, in.NameEn, in.CR, in.CRNumeric, in.Type, in.Size, in.Alignment,
		sb, in.ImageURL, in.SourceBook, in.SourceURL, in.RawHTML, in.ContentHash, in.IsUniqueNPC).Scan(&id)
	return id, err
}

func (r *Repo) LinkEditions(ctx context.Context, id2014, id2024 uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE monsters SET linked_monster_id = $2, updated_at = now() WHERE id = $1`, id2014, id2024)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `UPDATE monsters SET linked_monster_id = $2, updated_at = now() WHERE id = $1`, id2024, id2014)
	return err
}
