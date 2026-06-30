package players

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"monster-screen/backend/internal/apiutil"
)

type PlayerCharacter struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	AC                *int      `json:"ac,omitempty"`
	PassivePerception *int      `json:"passivePerception,omitempty"`
	MaxHP             *int      `json:"maxHp,omitempty"`
	Notes             string    `json:"notes"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type repo struct{ pool *pgxpool.Pool }

func RegisterRoutes(r chi.Router, pool *pgxpool.Pool) {
	h := &handler{repo: &repo{pool: pool}}
	r.Route("/api/players", func(pr chi.Router) {
		pr.Get("/", h.list)
		pr.Post("/", h.create)
		pr.Get("/{id}", h.get)
		pr.Put("/{id}", h.update)
		pr.Delete("/{id}", h.del)
	})
}

const cols = `id, name, ac, passive_perception, max_hp, notes, created_at, updated_at`

func scan(row pgx.Row) (PlayerCharacter, error) {
	var p PlayerCharacter
	err := row.Scan(&p.ID, &p.Name, &p.AC, &p.PassivePerception, &p.MaxHP, &p.Notes, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (rp *repo) list(ctx context.Context) ([]PlayerCharacter, error) {
	rows, err := rp.pool.Query(ctx, `SELECT `+cols+` FROM player_characters ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PlayerCharacter
	for rows.Next() {
		p, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (rp *repo) get(ctx context.Context, id uuid.UUID) (PlayerCharacter, error) {
	return scan(rp.pool.QueryRow(ctx, `SELECT `+cols+` FROM player_characters WHERE id = $1`, id))
}

func (rp *repo) create(ctx context.Context, p PlayerCharacter) (PlayerCharacter, error) {
	row := rp.pool.QueryRow(ctx, `
		INSERT INTO player_characters (name, ac, passive_perception, max_hp, notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+cols, p.Name, p.AC, p.PassivePerception, p.MaxHP, p.Notes)
	return scan(row)
}

func (rp *repo) update(ctx context.Context, id uuid.UUID, p PlayerCharacter) (PlayerCharacter, error) {
	row := rp.pool.QueryRow(ctx, `
		UPDATE player_characters SET name=$2, ac=$3, passive_perception=$4, max_hp=$5, notes=$6, updated_at=now()
		WHERE id=$1
		RETURNING `+cols, id, p.Name, p.AC, p.PassivePerception, p.MaxHP, p.Notes)
	return scan(row)
}

func (rp *repo) del(ctx context.Context, id uuid.UUID) error {
	_, err := rp.pool.Exec(ctx, `DELETE FROM player_characters WHERE id = $1`, id)
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
	p, err := h.repo.get(r.Context(), id)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, p)
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	var in PlayerCharacter
	if err := apiutil.DecodeJSON(r, &in); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	p, err := h.repo.create(r.Context(), in)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusCreated, p)
}

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var in PlayerCharacter
	if err := apiutil.DecodeJSON(r, &in); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	p, err := h.repo.update(r.Context(), id, in)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, p)
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
