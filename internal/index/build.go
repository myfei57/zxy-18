package index

import (
	"traceflow/internal/store"
)

// BuildInput is a trace record ready for indexing.
type BuildInput struct {
	TraceID    string
	Namespace  string
	RootSpanID string
	SpanCount  int
	State      string
}

// Build indexes aggregated traces into a new generation and makes it current.
func (idx *Index) Build(traces []BuildInput) (int, error) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	number := idx.seq + 1
	gen := newGeneration(number)
	for _, input := range traces {
		gen.Traces[input.TraceID] = TraceRef{
			TraceID:    input.TraceID,
			Namespace:  input.Namespace,
			RootSpanID: input.RootSpanID,
			SpanCount:  input.SpanCount,
			State:      input.State,
		}
	}
	if err := store.WriteJSON(idx.generationPath(number), gen); err != nil {
		return 0, err
	}
	idx.generations[number] = gen
	idx.current = number
	idx.seq = number
	return number, nil
}
