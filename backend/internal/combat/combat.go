package combat

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"monster-screen/backend/internal/apiutil"
	"monster-screen/backend/internal/auth"
)

type State struct {
	Round             int        `json:"round"`
	ActiveCombatantID *uuid.UUID `json:"activeCombatantId,omitempty"`
	Status            string     `json:"status"`
}

type repo struct{ pool *pgxpool.Pool }

func RegisterRoutes(r chi.Router, pool *pgxpool.Pool) {
	h := &handler{repo: &repo{pool: pool}}
	r.Route("/api/encounters/{id}/combat", func(cr chi.Router) {
		cr.Post("/start", h.start)
		cr.Post("/end", h.end)
		cr.Patch("/state", h.patchState)
	})
}

func (rp *repo) setState(ctx context.Context, userID, id uuid.UUID, status string, round int, active *uuid.UUID) (State, error) {
	var s State
	err := rp.pool.QueryRow(ctx, `
		UPDATE encounters SET status=$3, round=$4, active_combatant_id=$5, updated_at=now()
		WHERE id=$1 AND user_id=$2
		RETURNING status, round, active_combatant_id
	`, id, userID, status, round, active).Scan(&s.Status, &s.Round, &s.ActiveCombatantID)
	return s, err
}

func (rp *repo) patchState(ctx context.Context, userID, id uuid.UUID, round *int, active *uuid.UUID, hasActive bool) (State, error) {
	var s State
	err := rp.pool.QueryRow(ctx, `
		UPDATE encounters SET
			round = COALESCE($3, round),
			active_combatant_id = CASE WHEN $4::boolean THEN $5 ELSE active_combatant_id END,
			updated_at = now()
		WHERE id = $1 AND user_id = $2
		RETURNING status, round, active_combatant_id
	`, id, userID, round, hasActive, active).Scan(&s.Status, &s.Round, &s.ActiveCombatantID)
	return s, err
}

type handler struct{ repo *repo }

func (h *handler) start(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var firstCombatant *uuid.UUID
	row := h.repo.pool.QueryRow(r.Context(), `SELECT id FROM combatants WHERE encounter_id = $1 ORDER BY initiative DESC NULLS LAST, sort_order ASC LIMIT 1`, id)
	var first uuid.UUID
	if err := row.Scan(&first); err == nil {
		firstCombatant = &first
	}
	s, err := h.repo.setState(r.Context(), userID, id, "active", 1, firstCombatant)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, s)
}

func (h *handler) end(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	s, err := h.repo.setState(r.Context(), userID, id, "completed", 0, nil)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, s)
}

func (h *handler) patchState(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var in struct {
		Round             *int       `json:"round,omitempty"`
		ActiveCombatantID *uuid.UUID `json:"activeCombatantId,omitempty"`
		HasActive         bool       `json:"-"`
	}
	if err := apiutil.DecodeJSON(r, &in); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	s, err := h.repo.patchState(r.Context(), userID, id, in.Round, in.ActiveCombatantID, in.ActiveCombatantID != nil)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, s)
}
