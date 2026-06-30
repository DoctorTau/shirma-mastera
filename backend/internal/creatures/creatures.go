package creatures

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"monster-screen/backend/internal/apiutil"
	"monster-screen/backend/internal/statblock"
)

type CreatedCreature struct {
	ID        uuid.UUID           `json:"id"`
	NameRu    string              `json:"nameRu"`
	NameEn    string              `json:"nameEn"`
	StatBlock statblock.StatBlock `json:"statblock"`
	Notes     string              `json:"notes"`
	CreatedAt time.Time           `json:"createdAt"`
	UpdatedAt time.Time           `json:"updatedAt"`
}

type repo struct{ pool *pgxpool.Pool }

func RegisterRoutes(r chi.Router, pool *pgxpool.Pool) {
	h := &handler{repo: &repo{pool: pool}}
	r.Route("/api/creatures", func(cr chi.Router) {
		cr.Get("/", h.list)
		cr.Post("/", h.create)
		cr.Get("/{id}", h.get)
		cr.Put("/{id}", h.update)
		cr.Delete("/{id}", h.del)
	})
}

const cols = `id, name_ru, name_en, statblock, notes, created_at, updated_at`

func scan(row pgx.Row) (CreatedCreature, error) {
	var c CreatedCreature
	var sb []byte
	if err := row.Scan(&c.ID, &c.NameRu, &c.NameEn, &sb, &c.Notes, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return c, err
	}
	if len(sb) > 0 {
		_ = json.Unmarshal(sb, &c.StatBlock)
	}
	return c, nil
}

func (rp *repo) list(ctx context.Context) ([]CreatedCreature, error) {
	rows, err := rp.pool.Query(ctx, `SELECT `+cols+` FROM created_creatures ORDER BY name_ru ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CreatedCreature
	for rows.Next() {
		c, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (rp *repo) get(ctx context.Context, id uuid.UUID) (CreatedCreature, error) {
	return scan(rp.pool.QueryRow(ctx, `SELECT `+cols+` FROM created_creatures WHERE id = $1`, id))
}

func (rp *repo) create(ctx context.Context, c CreatedCreature) (CreatedCreature, error) {
	sb, _ := json.Marshal(c.StatBlock)
	row := rp.pool.QueryRow(ctx, `
		INSERT INTO created_creatures (name_ru, name_en, statblock, notes)
		VALUES ($1, $2, $3, $4)
		RETURNING `+cols, c.NameRu, c.NameEn, sb, c.Notes)
	return scan(row)
}

func (rp *repo) update(ctx context.Context, id uuid.UUID, c CreatedCreature) (CreatedCreature, error) {
	sb, _ := json.Marshal(c.StatBlock)
	row := rp.pool.QueryRow(ctx, `
		UPDATE created_creatures SET name_ru=$2, name_en=$3, statblock=$4, notes=$5, updated_at=now()
		WHERE id=$1
		RETURNING `+cols, id, c.NameRu, c.NameEn, sb, c.Notes)
	return scan(row)
}

func (rp *repo) del(ctx context.Context, id uuid.UUID) error {
	_, err := rp.pool.Exec(ctx, `DELETE FROM created_creatures WHERE id = $1`, id)
	return err
}

type handler struct{ repo *repo }

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	out, err := h.repo.list(r.Context())
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, out)
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	c, err := h.repo.get(r.Context(), id)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, c)
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	var in CreatedCreature
	if err := apiutil.DecodeJSON(r, &in); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	c, err := h.repo.create(r.Context(), in)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusCreated, c)
}

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var in CreatedCreature
	if err := apiutil.DecodeJSON(r, &in); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	c, err := h.repo.update(r.Context(), id, in)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, c)
}

func (h *handler) del(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.repo.del(r.Context(), id); err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusNoContent, nil)
}
