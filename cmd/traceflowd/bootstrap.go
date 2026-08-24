package main

import (
	"log"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"traceflow/internal/aggregate"
	"traceflow/internal/audit"
	"traceflow/internal/console"
	"traceflow/internal/index"
	"traceflow/internal/ns"
	"traceflow/internal/quota"
	"traceflow/internal/report"
	"traceflow/internal/sample"
	"traceflow/internal/settings"
	"traceflow/internal/span"
)

func bootstrap(opts settings.Options) (*console.Hub, error) {
	paths := settings.NewPaths(opts.DataDir)
	if err := paths.EnsureAll(); err != nil {
		return nil, err
	}

	registry := ns.NewRegistry(paths.RegistryFile())
	if err := registry.Load(); err != nil {
		return nil, err
	}
	if _, ok := registry.Lookup("order-center"); !ok {
		if err := registry.Register(ns.New("order-center", "order-api")); err != nil {
			return nil, err
		}
	}

	ledger := quota.NewLedger(paths.QuotaFile())
	if err := ledger.Load(); err != nil {
		return nil, err
	}
	if _, ok := ledger.Limits()["order-center"]; !ok {
		if err := ledger.SetLimit("order-center", 1000); err != nil {
			return nil, err
		}
	}

	sampler := sample.NewSampler(paths.SampleFile())
	if err := sampler.Load(); err != nil {
		return nil, err
	}
	if sampler.Rate == 0 {
		if err := sampler.Switch(100, 1000); err != nil {
			return nil, err
		}
	}

	st, err := span.NewStore(paths, ledger, registry)
	if err != nil {
		return nil, err
	}
	auditLogger := audit.NewLogger(paths.AuditFile())
	reporter, err := report.NewReporter(paths, st, report.NewLocalSink(filepath.Join(paths.Root(), "sink")), auditLogger)
	if err != nil {
		return nil, err
	}
	aggregator, err := aggregate.NewAggregator(paths, st, auditLogger)
	if err != nil {
		return nil, err
	}
	idx := index.NewIndex(paths.IndexDir())
	if err := idx.Load(); err != nil {
		return nil, err
	}

	hub, err := console.NewHub(paths, registry, st, sampler, reporter, aggregator, idx, ledger, auditLogger)
	if err != nil {
		return nil, err
	}
	if err := hub.EnsureWindows(); err != nil {
		return nil, err
	}
	if opts.Seed {
		seedDemo(hub)
	}
	return hub, nil
}

func seedDemo(hub *console.Hub) {
	now := time.Now().Unix()
	seeds := []struct {
		traceID string
		spans   []struct {
			service   string
			operation string
		}
	}{
		{"demo-checkout", []struct {
			service   string
			operation string
		}{{"order-api", "checkout"}, {"payment-api", "charge"}, {"inventory-api", "reserve"}}},
		{"demo-refund", []struct {
			service   string
			operation string
		}{{"order-api", "refund"}, {"payment-api", "reverse"}}},
		{"demo-search", []struct {
			service   string
			operation string
		}{{"search-api", "query"}, {"index-api", "match"}}},
	}
	for _, seed := range seeds {
		var rootID string
		for i, sp := range seed.spans {
			spanID := uuid.NewString()
			if i == 0 {
				rootID = spanID
			}
			parentID := ""
			if i > 0 {
				parentID = rootID
			}
			ingested, decided, err := hub.IngestSpan(console.IngestRequest{
				Namespace: "order-center",
				TraceID:   seed.traceID,
				SpanID:    spanID,
				ParentID:  parentID,
				Service:   sp.service,
				Operation: sp.operation,
				StartedAt: now,
			})
			if err != nil {
				log.Printf("seed ingest: %v", err)
				continue
			}
			if decided {
				if err := hub.FinishSpan("order-center", ingested.SpanID, map[string]string{"seed": seed.traceID}); err != nil {
					log.Printf("seed finish: %v", err)
				}
			}
		}
	}
	if err := hub.Maintenance(); err != nil {
		log.Printf("seed maintenance: %v", err)
	}
}
