// Package report batches spans into durable files and advances cursors.
package report

import (
	"sync"

	"traceflow/internal/audit"
	"traceflow/internal/settings"
	"traceflow/internal/span"
)

// Reporter batches and sends spans.
type Reporter struct {
	paths  settings.Paths
	spans  *span.Store
	cursor *Cursor
	sink   Sink
	audit  *audit.Logger
	mu     sync.Mutex
}

// NewReporter wires a reporter to its durable state.
func NewReporter(paths settings.Paths, st *span.Store, sink Sink, auditLogger *audit.Logger) (*Reporter, error) {
	cursor := OpenCursor(paths.CursorFile("report"))
	if err := cursor.Load(); err != nil {
		return nil, err
	}
	return &Reporter{paths: paths, spans: st, cursor: cursor, sink: sink, audit: auditLogger}, nil
}

func (r *Reporter) spansSinceCursor(namespace string) []span.Span {
	position := r.cursor.Position(namespace)
	var out []span.Span
	for _, existing := range r.spans.List(namespace) {
		if existing.Seq > position {
			out = append(out, existing)
		}
	}
	return out
}

// Cursor returns the report watermark for a namespace.
func (r *Reporter) Cursor(namespace string) uint64 {
	return r.cursor.Position(namespace)
}

// Cursors returns all report watermarks.
func (r *Reporter) Cursors() map[string]uint64 {
	return r.cursor.Positions()
}
