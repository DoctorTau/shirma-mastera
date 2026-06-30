package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL       string
	JWTSecret         string
	ListenAddr        string
	CrawlIntervalDays int
	CrawlRateLimitMs  int
	DndsuBaseURL2014  string
	DndsuBaseURL2024  string
	UserAgent         string
}

func Load() Config {
	return Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/monster_screen?sslmode=disable"),
		JWTSecret:         getEnv("JWT_SECRET", "dev-secret"),
		ListenAddr:        getEnv("LISTEN_ADDR", ":8080"),
		CrawlIntervalDays: getEnvInt("CRAWL_INTERVAL_DAYS", 30),
		CrawlRateLimitMs:  getEnvInt("CRAWL_RATE_LIMIT_MS", 600),
		DndsuBaseURL2014:  getEnv("DNDSU_BASE_URL_2014", "https://dnd.su"),
		DndsuBaseURL2024:  getEnv("DNDSU_BASE_URL_2024", "https://next.dnd.su"),
		UserAgent:         getEnv("CRAWL_USER_AGENT", "monster-screen-dm-companion/1.0 (personal use; contact: owner)"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
