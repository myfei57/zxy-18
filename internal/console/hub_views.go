package console

import (
	"sort"

	"traceflow/internal/aggregate"
	"traceflow/internal/audit"
	"traceflow/internal/index"
	"traceflow/internal/quota"
	"traceflow/internal/report"
	"traceflow/internal/sample"
	"traceflow/internal/trace"
)

// ReportStatus is the console-facing reporter snapshot.
type ReportStatus struct {
	Cursors      map[string]uint64  `json:"cursors"`
	Batches      []report.BatchView `json:"batches"`
	PendingSpans map[string]int     `json:"pending_spans"`
}

// ReportStatus returns reporter state.
func (h *Hub) ReportStatus() ReportStatus {
	h.mu.Lock()
	defer h.mu.Unlock()
	status := ReportStatus{
		Cursors:      h.reporter.Cursors(),
		Batches:      h.reporter.RecentBatches(20),
		PendingSpans: map[string]int{},
	}
	for _, namespace := range h.registry.List() {
		position := h.reporter.Cursor(namespace.Name)
		count := 0
		for _, existing := range h.spans.List(namespace.Name) {
			if existing.Seq > position {
				count++
			}
		}
		status.PendingSpans[namespace.Name] = count
	}
	return status
}

// Flush reports one namespace.
func (h *Hub) Flush(namespace string) (report.Batch, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.reporter.Flush(namespace)
}

// Send reports one namespace to the sink.
func (h *Hub) Send(namespace string) (report.Batch, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.reporter.Send(namespace)
}

// Aggregate runs one trace through the aggregator.
func (h *Hub) Aggregate(namespace, traceID string) (aggregate.Result, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.aggregator.Run(namespace, traceID)
}

// AggregateCursor returns the aggregation watermark for a namespace.
func (h *Hub) AggregateCursor(namespace string) uint64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.aggregator.CursorPosition(namespace)
}

// AggregateResults lists recent results.
func (h *Hub) AggregateResults(limit int) []aggregate.Result {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.aggregator.RecentResults(limit)
}

// OpenWindows lists open aggregation windows.
func (h *Hub) OpenWindows() []aggregate.Window {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.aggregator.OpenWindows()
}

// CloseWindow closes one aggregation window.
func (h *Hub) CloseWindow(namespace, id string) (aggregate.Summary, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.aggregator.CloseWindow(namespace, id)
}

// Summaries lists recent window summaries.
func (h *Hub) Summaries(limit int) []aggregate.Summary {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.aggregator.RecentSummaries(limit)
}

// TraceTree returns the assembled call tree for a trace.
func (h *Hub) TraceTree(namespace, traceID string) ([]trace.Node, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	spans := h.spans.ForTrace(namespace, traceID)
	if len(spans) == 0 {
		return nil, false
	}
	_, nodes, err := trace.Assemble(namespace, traceID, spans)
	if err != nil {
		return nil, false
	}
	return nodes, true
}

// IndexStatus returns the current generation stats.
func (h *Hub) IndexStatus() index.IndexStats {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.index.Stats()
}

// IndexSnapshot returns a frozen copy of the current generation.
func (h *Hub) IndexSnapshot() index.IndexStats {
	h.mu.Lock()
	defer h.mu.Unlock()
	gen := h.index.Snapshot()
	return index.IndexStats{
		Generation:  gen.Number,
		TraceCount:  gen.Count(),
		ByNamespace: countByNamespace(gen.Traces),
		Namespaces:  sortedNamespaces(gen.Traces),
	}
}

// SwitchRate changes the sampling rate with a rebuilt budget.
func (h *Hub) SwitchRate(rate int) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	total := 0
	for _, namespace := range h.registry.List() {
		total += h.aggregator.WindowCount(namespace.Name)
	}
	if err := h.sampler.Switch(rate, total); err != nil {
		return err
	}
	h.audit.Record(audit.NewEvent(audit.KindSample, "", "rate", "ok"))
	return nil
}

// SamplerStatus returns sampler state.
func (h *Hub) SamplerStatus() sample.State {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.sampler.State()
}

// QuotaStatus returns quota state for a namespace.
func (h *Hub) QuotaStatus(namespace string) quota.Status {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.quota.Status(namespace)
}

// QuotaList returns quota state for every known namespace.
func (h *Hub) QuotaList() []quota.Status {
	h.mu.Lock()
	defer h.mu.Unlock()
	names := map[string]bool{}
	for _, namespace := range h.registry.List() {
		names[namespace.Name] = true
	}
	for _, namespace := range h.quota.Namespaces() {
		names[namespace] = true
	}
	var out []quota.Status
	for namespace := range names {
		limit := h.quota.Limit(namespace)
		used := h.quota.Used(namespace)
		out = append(out, quota.Status{
			Namespace: namespace,
			Limit:     limit,
			Used:      used,
			Remaining: limit - used,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Namespace < out[j].Namespace })
	return out
}

// SetQuota configures a namespace limit.
func (h *Hub) SetQuota(namespace string, limit int) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.quota.SetLimit(namespace, limit)
}

// SpanList lists stored spans for a namespace.
func (h *Hub) SpanList(namespace string, limit int) []spanView {
	h.mu.Lock()
	defer h.mu.Unlock()
	all := h.spans.List(namespace)
	if limit > 0 && len(all) > limit {
		all = all[len(all)-limit:]
	}
	views := make([]spanView, 0, len(all))
	for _, existing := range all {
		views = append(views, toSpanView(existing, h.spans.PayloadBytes(namespace, existing.SpanID)))
	}
	return views
}

// SpanDetail returns one span with its payload.
func (h *Hub) SpanDetail(namespace, spanID string) (SpanDetail, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	existing, ok := h.spans.SpanByID(namespace, spanID)
	if !ok {
		return SpanDetail{}, false
	}
	detail := SpanDetail{Span: toSpanView(existing, h.spans.PayloadBytes(namespace, spanID))}
	if payload, err := h.spans.ReadPayload(namespace, spanID); err == nil {
		detail.Payload = payload.Keys
		detail.PayloadBytes = payload.Bytes
	}
	return detail, true
}

// AuditRecent returns recent audit events.
func (h *Hub) AuditRecent(limit int) []audit.Event {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.audit.Recent(limit)
}

// AuditStats returns audit counts per kind.
func (h *Hub) AuditStats() map[string]int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.audit.Stats()
}

func (h *Hub) aggregatedTracesLocked() []index.BuildInput {
	results := h.aggregator.RecentResults(0)
	traces := make([]index.BuildInput, 0, len(results))
	for _, result := range results {
		traces = append(traces, index.BuildInput{
			TraceID:    result.TraceID,
			Namespace:  result.Namespace,
			RootSpanID: result.RootSpanID,
			State:      result.State,
			SpanCount:  result.SpanCount,
		})
	}
	return traces
}
