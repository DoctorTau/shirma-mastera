package crawl

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"monster-screen/backend/internal/apiutil"
)

func RegisterRoutes(r chi.Router, svc *Service) {
	h := &handler{svc: svc}
	r.Route("/api/admin/crawl", func(cr chi.Router) {
		cr.Post("/seed", h.seed)
		cr.Post("/resync", h.resync)
		cr.Get("/status", h.status)
	})
}

type handler struct{ svc *Service }

func (h *handler) seed(w http.ResponseWriter, r *http.Request) {
	h.runAsync(w, r.Context(), h.svc.RunSeed)
}

func (h *handler) resync(w http.ResponseWriter, r *http.Request) {
	h.runAsync(w, r.Context(), h.svc.RunIncremental)
}

func (h *handler) status(w http.ResponseWriter, r *http.Request) {
	apiutil.WriteJSON(w, http.StatusOK, map[string]bool{"running": h.svc.IsRunning()})
}

// runAsync kicks the crawl off in the background and returns immediately: a
// full seed walks ~2900 cards at a polite rate limit, which takes minutes,
// far too long for a single HTTP request to wait on.
func (h *handler) runAsync(w http.ResponseWriter, reqCtx context.Context, fn func(context.Context) (RunStats, error)) {
	if h.svc.IsRunning() {
		apiutil.WriteError(w, http.StatusConflict, "a crawl is already running")
		return
	}
	go func() {
		// deliberately not tied to the request's context: the crawl must
		// keep running after the HTTP response is sent.
		ctx := context.Background()
		if _, err := fn(ctx); err != nil {
			log.Printf("crawl run failed: %v", err)
		}
	}()
	apiutil.WriteJSON(w, http.StatusAccepted, map[string]string{"status": "started"})
}
