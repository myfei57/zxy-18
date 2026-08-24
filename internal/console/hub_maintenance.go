package console

import (
	"time"

	"traceflow/internal/audit"
)

// FlushAll flushes every namespace in order.
func (h *Hub) FlushAll() (int, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	count := 0
	for _, namespace := range h.registry.List() {
		if _, err := h.reporter.Flush(namespace.Name); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// AggregateAll aggregates all pending traces.
func (h *Hub) AggregateAll() (int, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	count := 0
	for _, namespace := range h.registry.List() {
		for _, traceID := range h.aggregator.PendingTraces(namespace.Name) {
			if _, err := h.aggregator.Run(namespace.Name, traceID); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}

// BuildIndex indexes all aggregated results into a fresh generation.
func (h *Hub) BuildIndex() (int, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	traces := h.aggregatedTracesLocked()
	number, err := h.index.Build(traces)
	if err != nil {
		return 0, err
	}
	h.audit.Record(audit.NewEvent(audit.KindIndex, "", "generation", "ok"))
	return number, nil
}

// CloseExpiredWindows closes open windows whose end has passed.
func (h *Hub) CloseExpiredWindows() (int, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	now := time.Now().Unix()
	count := 0
	for _, w := range h.aggregator.OpenWindows() {
		if w.End < now {
			if _, err := h.aggregator.CloseWindow(w.Namespace, w.ID); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}

// Maintenance runs the full pipeline once.
func (h *Hub) Maintenance() error {
	if _, err := h.FlushAll(); err != nil {
		return err
	}
	if _, err := h.AggregateAll(); err != nil {
		return err
	}
	if _, err := h.BuildIndex(); err != nil {
		return err
	}
	_, err := h.CloseExpiredWindows()
	return err
}

// EnsureWindows opens the current time window for every namespace.
func (h *Hub) EnsureWindows() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	now := time.Now().Unix()
	start := now - now%3600
	for _, namespace := range h.registry.List() {
		if _, err := h.aggregator.EnsureCurrentWindow(namespace.Name, start, start+3600); err != nil {
			return err
		}
	}
	return nil
}
