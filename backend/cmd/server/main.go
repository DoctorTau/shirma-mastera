package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"monster-screen/backend/internal/config"
	"monster-screen/backend/internal/crawl"
	"monster-screen/backend/internal/db"
	"monster-screen/backend/internal/dndsu"
	"monster-screen/backend/internal/httpapi"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	client := dndsu.NewClient(cfg.UserAgent, time.Duration(cfg.CrawlRateLimitMs)*time.Millisecond)
	crawlSvc := crawl.NewService(pool, client, cfg.DndsuBaseURL2014, cfg.DndsuBaseURL2024)
	crawl.StartScheduler(ctx, crawlSvc, cfg.CrawlIntervalDays)

	router := httpapi.NewRouter(cfg, pool, crawlSvc)

	server := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: router,
	}

	go func() {
		log.Printf("listening on %s", cfg.ListenAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
