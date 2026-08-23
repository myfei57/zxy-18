package main

import (
	"context"
	"log"
	"time"

	"traceflow/internal/console"
)

func runMaintenance(ctx context.Context, hub *console.Hub, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := hub.Maintenance(); err != nil {
				log.Printf("maintenance: %v", err)
			}
		}
	}
}
