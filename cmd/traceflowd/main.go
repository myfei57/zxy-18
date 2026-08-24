// Command traceflowd runs the TraceFlow distributed tracing platform.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"traceflow/internal/console"
	"traceflow/internal/settings"
)

func main() {
	opts := settings.Defaults()
	flag.StringVar(&opts.DataDir, "data", opts.DataDir, "data directory")
	flag.StringVar(&opts.Addr, "addr", opts.Addr, "http listen address")
	flag.BoolVar(&opts.Seed, "seed", opts.Seed, "seed demo data")
	flag.Parse()

	hub, err := bootstrap(opts)
	if err != nil {
		log.Fatalf("bootstrap: %v", err)
	}
	server := console.NewServer(hub)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go runMaintenance(ctx, hub, 15*time.Second)

	log.Printf("traceflow listening on %s with data at %s", opts.Addr, opts.DataDir)
	if err := server.Start(opts.Addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
