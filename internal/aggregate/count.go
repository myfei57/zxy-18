package aggregate

import "traceflow/internal/span"

// WindowCount returns the number of spans observed in the current window.
func (a *Aggregator) WindowCount(namespace string) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	window, ok := a.currentWindow(namespace)
	if !ok {
		return 0
	}
	return len(a.windowSpansLocked(window))
}

func (a *Aggregator) windowSpansLocked(w Window) []span.Span {
	var out []span.Span
	for _, existing := range a.spans.List(w.Namespace) {
		if existing.StartedAt >= w.Start && existing.StartedAt < w.End {
			out = append(out, existing)
		}
	}
	return out
}
