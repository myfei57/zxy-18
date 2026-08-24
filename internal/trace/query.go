package trace

import "traceflow/internal/index"

// Resolver answers trace lookups against the live index.
type Resolver struct {
	idx *index.Index
}

// NewResolver binds a resolver to an index.
func NewResolver(idx *index.Index) *Resolver {
	return &Resolver{idx: idx}
}

// Query resolves a trace through the current index generation.
//
// It reads the live current generation on every call instead of a snapshot
// captured at construction time, so traces freshly indexed by Build become
// visible without a process restart.
func (r *Resolver) Query(traceID string) (index.TraceRef, bool) {
	return r.idx.Current().Lookup(traceID)
}
