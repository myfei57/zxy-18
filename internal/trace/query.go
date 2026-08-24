package trace

import "traceflow/internal/index"

// Resolver answers trace lookups against the live index.
type Resolver struct {
	idx  *index.Index
	view *index.Generation
}

// NewResolver binds a resolver to an index.
func NewResolver(idx *index.Index) *Resolver {
	return &Resolver{idx: idx, view: idx.Snapshot()}
}

// Query resolves a trace through the current index generation.
func (r *Resolver) Query(traceID string) (index.TraceRef, bool) {
	return r.view.Lookup(traceID)
}
