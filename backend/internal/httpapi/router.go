package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"monster-screen/backend/internal/apiutil"
	"monster-screen/backend/internal/auth"
	"monster-screen/backend/internal/combat"
	"monster-screen/backend/internal/config"
	"monster-screen/backend/internal/crawl"
	"monster-screen/backend/internal/creatures"
	"monster-screen/backend/internal/encounters"
	"monster-screen/backend/internal/monsters"
	"monster-screen/backend/internal/players"
	"monster-screen/backend/internal/users"
)

func NewRouter(cfg config.Config, pool *pgxpool.Pool, crawlSvc *crawl.Service) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
	}))

	r.Get("/api/health", func(w http.ResponseWriter, req *http.Request) {
		apiutil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	users.RegisterRoutes(r, pool, cfg.JWTSecret)

	r.Group(func(protected chi.Router) {
		protected.Use(auth.Middleware(cfg.JWTSecret))

		protected.Get("/api/me", func(w http.ResponseWriter, req *http.Request) {
			apiutil.WriteJSON(w, http.StatusOK, map[string]bool{"authenticated": true})
		})

		monsters.RegisterRoutes(protected, pool)
		creatures.RegisterRoutes(protected, pool)
		players.RegisterRoutes(protected, pool)
		encounters.RegisterRoutes(protected, pool)
		combat.RegisterRoutes(protected, pool)
		crawl.RegisterRoutes(protected, crawlSvc)
	})

	return r
}
