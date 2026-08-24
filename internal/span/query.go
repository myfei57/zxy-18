package span

import "sort"

// List returns all stored spans for a namespace ordered by sequence.
func (st *Store) List(namespace string) []Span {
	st.mu.Lock()
	defer st.mu.Unlock()
	out := make([]Span, len(st.spans[namespace]))
	copy(out, st.spans[namespace])
	sort.Slice(out, func(i, j int) bool { return out[i].Seq < out[j].Seq })
	return out
}

// SpanByID returns one span by id.
func (st *Store) SpanByID(namespace, spanID string) (Span, bool) {
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.spanByIDLocked(namespace, spanID)
}

// ForTrace returns spans belonging to a trace ordered by start time.
func (st *Store) ForTrace(namespace, traceID string) []Span {
	st.mu.Lock()
	defer st.mu.Unlock()
	var out []Span
	for _, existing := range st.spans[namespace] {
		if existing.TraceID == traceID {
			out = append(out, existing)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt < out[j].StartedAt })
	return out
}

// Count returns the number of stored spans in a namespace.
func (st *Store) Count(namespace string) int {
	st.mu.Lock()
	defer st.mu.Unlock()
	return len(st.spans[namespace])
}

// TraceIDs returns distinct trace ids stored in a namespace.
func (st *Store) TraceIDs(namespace string) []string {
	st.mu.Lock()
	defer st.mu.Unlock()
	seen := map[string]bool{}
	var out []string
	for _, existing := range st.spans[namespace] {
		if !seen[existing.TraceID] {
			seen[existing.TraceID] = true
			out = append(out, existing.TraceID)
		}
	}
	sort.Strings(out)
	return out
}
