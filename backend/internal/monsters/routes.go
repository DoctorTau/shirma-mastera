package monsters

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(r chi.Router, pool *pgxpool.Pool) {
	h := &Handler{repo: NewRepo(pool)}
	r.Get("/api/monsters", h.List)
	r.Get("/api/monsters/{id}", h.Get)
}
