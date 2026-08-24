package index

// TraceRef is a queryable pointer to an aggregated trace.
type TraceRef struct {
	TraceID    string `json:"trace_id"`
	Namespace  string `json:"namespace"`
	RootSpanID string `json:"root_span_id"`
	SpanCount  int    `json:"span_count"`
	State      string `json:"state"`
}

// Generation is an immutable index generation.
type Generation struct {
	Number int                 `json:"number"`
	Traces map[string]TraceRef `json:"traces"`
}

func newGeneration(number int) *Generation {
	return &Generation{Number: number, Traces: map[string]TraceRef{}}
}

// Lookup finds a trace ref in the generation.
func (g *Generation) Lookup(traceID string) (TraceRef, bool) {
	ref, ok := g.Traces[traceID]
	return ref, ok
}

// Count returns the number of indexed traces.
func (g *Generation) Count() int {
	return len(g.Traces)
}
