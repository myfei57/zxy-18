// Package aggregate merges spans into durable trace results.
package aggregate

import (
	"sync"

	"traceflow/internal/audit"
	"traceflow/internal/settings"
	"traceflow/internal/span"
)

// Aggregator runs the merge pipeline for one instance.
type Aggregator struct {
	paths  settings.Paths
	spans  *span.Store
	cursor *Cursor
	audit  *audit.Logger
	mu     sync.Mutex
}

// NewAggregator wires the aggregator to its durable state.
func NewAggregator(paths settings.Paths, st *span.Store, auditLogger *audit.Logger) (*Aggregator, error) {
	cursor := OpenCursor(paths.CursorFile("aggregate"))
	if err := cursor.Load(); err != nil {
		return nil, err
	}
	return &Aggregator{paths: paths, spans: st, cursor: cursor, audit: auditLogger}, nil
}

// PendingTraces returns trace ids not yet aggregated.
func (a *Aggregator) PendingTraces(namespace string) []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	position := a.cursor.Position(namespace)
	var out []string
	for _, traceID := range a.spans.TraceIDs(namespace) {
		for _, existing := range a.spans.ForTrace(namespace, traceID) {
			if existing.Seq > position {
				out = append(out, traceID)
				break
			}
		}
	}
	return out
}

// CursorPosition returns the aggregation watermark for a namespace.
func (a *Aggregator) CursorPosition(namespace string) uint64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.cursor.Position(namespace)
}
