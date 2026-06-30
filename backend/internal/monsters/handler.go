package monsters

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"monster-screen/backend/internal/apiutil"
)

type Handler struct {
	repo *Repo
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	offset, _ := strconv.Atoi(q.Get("offset"))

	list, err := h.repo.List(r.Context(), q.Get("search"), q.Get("edition"), q.Get("type"), limit, offset)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, list)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	m, err := h.repo.Get(r.Context(), id)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, m)
}
