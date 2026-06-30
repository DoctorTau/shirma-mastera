package encounters

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"monster-screen/backend/internal/apiutil"
	"monster-screen/backend/internal/auth"
)

type Encounter struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	Round             int        `json:"round"`
	ActiveCombatantID *uuid.UUID `json:"activeCombatantId,omitempty"`
	Status            string     `json:"status"`
	Combatants        []Combatant `json:"combatants,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type Combatant struct {
	ID              uuid.UUID `json:"id"`
	EncounterID     uuid.UUID `json:"encounterId"`
	SourceType      string    `json:"sourceType"`
	SourceID        *uuid.UUID `json:"sourceId,omitempty"`
	MonsterEdition  *string   `json:"monsterEdition,omitempty"`
	DisplayName     string    `json:"displayName"`
	MaxHP           *int      `json:"maxHp,omitempty"`
	CurrentHP       *int      `json:"currentHp,omitempty"`
	TempHP          int       `json:"tempHp"`
	Initiative      *int      `json:"initiative,omitempty"`
	Conditions      []string  `json:"conditions"`
	Notes           string    `json:"notes"`
	IsPC            bool      `json:"isPc"`
	SortOrder       int       `json:"sortOrder"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type AddCombatantInput struct {
	SourceType     string     `json:"sourceType"`
	SourceID       *uuid.UUID `json:"sourceId,omitempty"`
	MonsterEdition *string    `json:"monsterEdition,omitempty"`
	DisplayName    string     `json:"displayName"`
	MaxHP          *int       `json:"maxHp,omitempty"`
	IsPC           bool       `json:"isPc"`
	Count          int        `json:"count"`
}

type repo struct{ pool *pgxpool.Pool }

func RegisterRoutes(r chi.Router, pool *pgxpool.Pool) {
	h := &handler{repo: &repo{pool: pool}}
	r.Route("/api/encounters", func(er chi.Router) {
		er.Get("/", h.list)
		er.Post("/", h.create)
		er.Get("/{id}", h.get)
		er.Put("/{id}", h.update)
		er.Delete("/{id}", h.del)
		er.Post("/{id}/duplicate", h.duplicate)
		er.Post("/{id}/combatants", h.addCombatants)
		er.Put("/{id}/combatants/{cid}", h.updateCombatant)
		er.Patch("/{id}/combatants/{cid}", h.patchCombatant)
		er.Delete("/{id}/combatants/{cid}", h.deleteCombatant)
	})
}

const encCols = `id, name, round, active_combatant_id, status, created_at, updated_at`
const combCols = `id, encounter_id, source_type, source_id, monster_edition, display_name, max_hp, current_hp, temp_hp, initiative, conditions, notes, is_pc, sort_order, updated_at`

func scanEncounter(row pgx.Row) (Encounter, error) {
	var e Encounter
	err := row.Scan(&e.ID, &e.Name, &e.Round, &e.ActiveCombatantID, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func scanCombatant(row pgx.Row) (Combatant, error) {
	var c Combatant
	err := row.Scan(&c.ID, &c.EncounterID, &c.SourceType, &c.SourceID, &c.MonsterEdition, &c.DisplayName,
		&c.MaxHP, &c.CurrentHP, &c.TempHP, &c.Initiative, &c.Conditions, &c.Notes, &c.IsPC, &c.SortOrder, &c.UpdatedAt)
	return c, err
}

func (rp *repo) ownsEncounter(ctx context.Context, userID, encounterID uuid.UUID) (bool, error) {
	var exists bool
	err := rp.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM encounters WHERE id = $1 AND user_id = $2)`,
		encounterID, userID).Scan(&exists)
	return exists, err
}

func (rp *repo) list(ctx context.Context, userID uuid.UUID) ([]Encounter, error) {
	rows, err := rp.pool.Query(ctx, `SELECT `+encCols+` FROM encounters WHERE user_id = $1 ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Encounter
	for rows.Next() {
		e, err := scanEncounter(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (rp *repo) combatantsFor(ctx context.Context, encounterID uuid.UUID) ([]Combatant, error) {
	rows, err := rp.pool.Query(ctx, `SELECT `+combCols+` FROM combatants WHERE encounter_id = $1 ORDER BY sort_order ASC`, encounterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Combatant
	for rows.Next() {
		c, err := scanCombatant(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (rp *repo) get(ctx context.Context, userID, id uuid.UUID) (Encounter, error) {
	e, err := scanEncounter(rp.pool.QueryRow(ctx, `SELECT `+encCols+` FROM encounters WHERE id = $1 AND user_id = $2`, id, userID))
	if err != nil {
		return e, err
	}
	combatants, err := rp.combatantsFor(ctx, id)
	if err != nil {
		return e, err
	}
	e.Combatants = combatants
	return e, nil
}

func (rp *repo) create(ctx context.Context, userID uuid.UUID, name string) (Encounter, error) {
	row := rp.pool.QueryRow(ctx, `INSERT INTO encounters (name, user_id) VALUES ($1, $2) RETURNING `+encCols, name, userID)
	return scanEncounter(row)
}

func (rp *repo) update(ctx context.Context, userID, id uuid.UUID, name string) (Encounter, error) {
	row := rp.pool.QueryRow(ctx, `UPDATE encounters SET name=$3, updated_at=now() WHERE id=$1 AND user_id=$2 RETURNING `+encCols, id, userID, name)
	return scanEncounter(row)
}

func (rp *repo) del(ctx context.Context, userID, id uuid.UUID) error {
	_, err := rp.pool.Exec(ctx, `DELETE FROM encounters WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (rp *repo) duplicate(ctx context.Context, userID, id uuid.UUID) (Encounter, error) {
	src, err := rp.get(ctx, userID, id)
	if err != nil {
		return Encounter{}, err
	}
	dst, err := rp.create(ctx, userID, src.Name+" (копия)")
	if err != nil {
		return Encounter{}, err
	}
	for _, c := range src.Combatants {
		_, err := rp.pool.Exec(ctx, `
			INSERT INTO combatants (encounter_id, source_type, source_id, monster_edition, display_name, max_hp, current_hp, temp_hp, conditions, notes, is_pc, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6,$7,0,$8,$9,$10,$11)
		`, dst.ID, c.SourceType, c.SourceID, c.MonsterEdition, c.DisplayName, c.MaxHP, c.MaxHP, c.Conditions, c.Notes, c.IsPC, c.SortOrder)
		if err != nil {
			return Encounter{}, err
		}
	}
	return rp.get(ctx, userID, dst.ID)
}

// addCombatants inserts `count` (or 1) combatants, auto-numbering duplicates
// when count > 1 (e.g. "Гоблин 1", "Гоблин 2", ...).
func (rp *repo) addCombatants(ctx context.Context, encounterID uuid.UUID, in AddCombatantInput) ([]Combatant, error) {
	count := in.Count
	if count < 1 {
		count = 1
	}

	var maxOrder int
	_ = rp.pool.QueryRow(ctx, `SELECT COALESCE(MAX(sort_order), -1) FROM combatants WHERE encounter_id = $1`, encounterID).Scan(&maxOrder)

	var created []Combatant
	for i := 1; i <= count; i++ {
		name := in.DisplayName
		if count > 1 {
			name = fmt.Sprintf("%s %d", in.DisplayName, i)
		}
		maxOrder++
		row := rp.pool.QueryRow(ctx, `
			INSERT INTO combatants (encounter_id, source_type, source_id, monster_edition, display_name, max_hp, current_hp, temp_hp, is_pc, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6,$6,0,$7,$8)
			RETURNING `+combCols, encounterID, in.SourceType, in.SourceID, in.MonsterEdition, name, in.MaxHP, in.IsPC, maxOrder)
		c, err := scanCombatant(row)
		if err != nil {
			return nil, err
		}
		created = append(created, c)
	}
	_, _ = rp.pool.Exec(ctx, `UPDATE encounters SET updated_at = now() WHERE id = $1`, encounterID)
	return created, nil
}

type PatchCombatantInput struct {
	DisplayName *string   `json:"displayName,omitempty"`
	CurrentHP   *int      `json:"currentHp,omitempty"`
	TempHP      *int      `json:"tempHp,omitempty"`
	MaxHP       *int      `json:"maxHp,omitempty"`
	Initiative  *int      `json:"initiative,omitempty"`
	Conditions  *[]string `json:"conditions,omitempty"`
	Notes       *string   `json:"notes,omitempty"`
	SortOrder   *int      `json:"sortOrder,omitempty"`
}

func (rp *repo) patchCombatant(ctx context.Context, encounterID, id uuid.UUID, in PatchCombatantInput) (Combatant, error) {
	row := rp.pool.QueryRow(ctx, `
		UPDATE combatants SET
			display_name = COALESCE($3, display_name),
			current_hp = CASE WHEN $4::boolean THEN $5 ELSE current_hp END,
			temp_hp = COALESCE($6, temp_hp),
			max_hp = COALESCE($7, max_hp),
			initiative = CASE WHEN $8::boolean THEN $9 ELSE initiative END,
			conditions = COALESCE($10, conditions),
			notes = COALESCE($11, notes),
			sort_order = COALESCE($12, sort_order),
			updated_at = now()
		WHERE id = $1 AND encounter_id = $2
		RETURNING `+combCols,
		id, encounterID,
		in.DisplayName,
		in.CurrentHP != nil, in.CurrentHP,
		in.TempHP,
		in.MaxHP,
		in.Initiative != nil, in.Initiative,
		in.Conditions,
		in.Notes,
		in.SortOrder,
	)
	c, err := scanCombatant(row)
	if err == nil {
		_, _ = rp.pool.Exec(ctx, `UPDATE encounters SET updated_at = now() WHERE id = $1`, encounterID)
	}
	return c, err
}

func (rp *repo) updateCombatant(ctx context.Context, encounterID, id uuid.UUID, c Combatant) (Combatant, error) {
	row := rp.pool.QueryRow(ctx, `
		UPDATE combatants SET display_name=$3, max_hp=$4, current_hp=$5, temp_hp=$6, initiative=$7,
			conditions=$8, notes=$9, is_pc=$10, sort_order=$11, updated_at=now()
		WHERE id=$1 AND encounter_id=$2
		RETURNING `+combCols, id, encounterID, c.DisplayName, c.MaxHP, c.CurrentHP, c.TempHP, c.Initiative,
		c.Conditions, c.Notes, c.IsPC, c.SortOrder)
	out, err := scanCombatant(row)
	if err == nil {
		_, _ = rp.pool.Exec(ctx, `UPDATE encounters SET updated_at = now() WHERE id = $1`, encounterID)
	}
	return out, err
}

func (rp *repo) deleteCombatant(ctx context.Context, encounterID, id uuid.UUID) error {
	_, err := rp.pool.Exec(ctx, `DELETE FROM combatants WHERE id = $1 AND encounter_id = $2`, id, encounterID)
	if err == nil {
		_, _ = rp.pool.Exec(ctx, `UPDATE encounters SET updated_at = now() WHERE id = $1`, encounterID)
	}
	return err
}

type handler struct{ repo *repo }

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	out, err := h.repo.list(r.Context(), userID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, out)
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	e, err := h.repo.get(r.Context(), userID, id)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, e)
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	var in struct {
		Name string `json:"name"`
	}
	if err := apiutil.DecodeJSON(r, &in); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	e, err := h.repo.create(r.Context(), userID, in.Name)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusCreated, e)
}

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var in struct {
		Name string `json:"name"`
	}
	if err := apiutil.DecodeJSON(r, &in); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	e, err := h.repo.update(r.Context(), userID, id, in.Name)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, e)
}

func (h *handler) del(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.repo.del(r.Context(), userID, id); err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusNoContent, nil)
}

func (h *handler) duplicate(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	e, err := h.repo.duplicate(r.Context(), userID, id)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusCreated, e)
}

// requireOwnedEncounter checks the encounter exists and belongs to the authenticated
// user, writing a 404 response and returning false if not.
func (h *handler) requireOwnedEncounter(w http.ResponseWriter, r *http.Request, encounterID uuid.UUID) bool {
	userID, _ := auth.UserIDFromContext(r.Context())
	owns, err := h.repo.ownsEncounter(r.Context(), userID, encounterID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return false
	}
	if !owns {
		apiutil.WriteError(w, http.StatusNotFound, "not found")
		return false
	}
	return true
}

func (h *handler) addCombatants(w http.ResponseWriter, r *http.Request) {
	encounterID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if !h.requireOwnedEncounter(w, r, encounterID) {
		return
	}
	var in AddCombatantInput
	if err := apiutil.DecodeJSON(r, &in); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	out, err := h.repo.addCombatants(r.Context(), encounterID, in)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusCreated, out)
}

func (h *handler) updateCombatant(w http.ResponseWriter, r *http.Request) {
	encounterID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if !h.requireOwnedEncounter(w, r, encounterID) {
		return
	}
	cid, err := uuid.Parse(chi.URLParam(r, "cid"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid combatant id")
		return
	}
	var in Combatant
	if err := apiutil.DecodeJSON(r, &in); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	c, err := h.repo.updateCombatant(r.Context(), encounterID, cid, in)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, c)
}

func (h *handler) patchCombatant(w http.ResponseWriter, r *http.Request) {
	encounterID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if !h.requireOwnedEncounter(w, r, encounterID) {
		return
	}
	cid, err := uuid.Parse(chi.URLParam(r, "cid"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid combatant id")
		return
	}
	var in PatchCombatantInput
	if err := apiutil.DecodeJSON(r, &in); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	c, err := h.repo.patchCombatant(r.Context(), encounterID, cid, in)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, c)
}

func (h *handler) deleteCombatant(w http.ResponseWriter, r *http.Request) {
	encounterID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if !h.requireOwnedEncounter(w, r, encounterID) {
		return
	}
	cid, err := uuid.Parse(chi.URLParam(r, "cid"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid combatant id")
		return
	}
	if err := h.repo.deleteCombatant(r.Context(), encounterID, cid); err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusNoContent, nil)
}
