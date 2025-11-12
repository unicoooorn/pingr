package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/unicoooorn/pingr/internal/service"
)

type FixedIntervalScheduler struct {
	svc      service.Service
	interval time.Duration
}

// NewFixedIntervalScheduler creates a scheduler that refreshes statuses every given interval.
func NewFixedIntervalScheduler(svc service.Service, interval time.Duration) *FixedIntervalScheduler {
	return &FixedIntervalScheduler{
		svc:      svc,
		interval: interval,
	}
}

// StartMonitoring runs periodic health checks until the context is cancelled.
func (fis *FixedIntervalScheduler) StartMonitoring(ctx context.Context) error {
	ticker := time.NewTicker(fis.interval)
	defer ticker.Stop()

	log.Printf("[scheduler] started with interval %v", fis.interval)

	// Run immediately once at start
	if err := fis.svc.InitiateCheck(ctx); err != nil {
		log.Printf("[scheduler] initial refresh failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("[scheduler] stopping monitoring...")
			return ctx.Err()

		case <-ticker.C:
			if err := fis.svc.InitiateCheck(ctx); err != nil {
				log.Printf("[scheduler] subsytems health check failed: %v", err)
			}
		}
	}
}
