package crawl

import (
	"context"
	"log"
	"time"
)

// StartScheduler runs an incremental crawl every intervalDays, starting
// after one interval has elapsed (it does not crawl immediately on boot —
// the seed/manual endpoints cover the initial population, per §8.3).
func StartScheduler(ctx context.Context, svc *Service, intervalDays int) {
	if intervalDays <= 0 {
		intervalDays = 30
	}
	interval := time.Duration(intervalDays) * 24 * time.Hour

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Printf("crawl: starting scheduled incremental run (every %d days)", intervalDays)
				if _, err := svc.RunIncremental(ctx); err != nil {
					log.Printf("crawl: scheduled run failed: %v", err)
				}
			}
		}
	}()
}
