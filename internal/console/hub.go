// Package console exposes the TraceFlow HTTP API and embedded control pages.
package console

import (
	"fmt"
	"sort"
	"sync"

	"traceflow/internal/aggregate"
	"traceflow/internal/audit"
	"traceflow/internal/index"
	"traceflow/internal/ns"
	"traceflow/internal/quota"
	"traceflow/internal/report"
	"traceflow/internal/sample"
	"traceflow/internal/settings"
	"traceflow/internal/span"
	"traceflow/internal/trace"
)

// Hub wires all components and implements the console backend.
type Hub struct {
	mu         sync.Mutex
	paths      settings.Paths
	registry   *ns.Registry
	spans      *span.Store
	sampler    *sample.Sampler
	reporter   *report.Reporter
	aggregator *aggregate.Aggregator
	index      *index.Index
	quota      *quota.Ledger
	audit      *audit.Logger
	resolver   *trace.Resolver
}

// NewHub constructs the full backend.
func NewHub(
	paths settings.Paths,
	registry *ns.Registry,
	st *span.Store,
	sampler *sample.Sampler,
	reporter *report.Reporter,
	aggregator *aggregate.Aggregator,
	idx *index.Index,
	quotaLedger *quota.Ledger,
	auditLogger *audit.Logger,
) (*Hub, error) {
	hub := &Hub{
		paths:      paths,
		registry:   registry,
		spans:      st,
		sampler:    sampler,
		reporter:   reporter,
		aggregator: aggregator,
		index:      idx,
		quota:      quotaLedger,
		audit:      auditLogger,
		resolver:   trace.NewResolver(idx),
	}
	return hub, nil
}

// Namespaces lists all registered namespaces.
func (h *Hub) Namespaces() []ns.Namespace {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.registry.List()
}

// RegisterNamespace adds a namespace.
func (h *Hub) RegisterNamespace(name, service string) (ns.Namespace, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	namespace := ns.New(name, service)
	if err := h.registry.Register(namespace); err != nil {
		return ns.Namespace{}, err
	}
	return namespace, nil
}

// DeactivateNamespace marks a namespace inactive.
func (h *Hub) DeactivateNamespace(name string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.registry.Deactivate(name)
}

// IngestSpan receives one span through the sampling pipeline.
func (h *Hub) IngestSpan(req IngestRequest) (span.Span, bool, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	namespace, ok := h.registry.Lookup(req.Namespace)
	if !ok {
		return span.Span{}, false, fmt.Errorf("unknown namespace %s", req.Namespace)
	}
	sp := span.New(req.TraceID, req.SpanID, req.ParentID, req.Service, req.Operation)
	sp.Namespace = namespace.Name
	sp.StartedAt = req.StartedAt
	decided, err := h.sampler.Vote(h.spans, namespace.Name, sp)
	if err != nil {
		h.audit.Record(audit.NewEvent(audit.KindQuota, namespace.Name, req.TraceID, "failed"))
		return sp, false, err
	}
	return sp, decided, nil
}

// FinishSpan attaches a payload and marks a span finished.
func (h *Hub) FinishSpan(namespace, spanID string, payload map[string]string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.spans.Finish(namespace, spanID, payload)
}

// QueryTrace resolves a trace through the current index.
func (h *Hub) QueryTrace(traceID string) (TraceView, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	ref, ok := h.resolver.Query(traceID)
	if !ok {
		return TraceView{}, false
	}
	return h.traceViewLocked(ref), true
}

// RecentTraces lists the newest indexed traces.
func (h *Hub) RecentTraces(limit int) []TraceView {
	h.mu.Lock()
	defer h.mu.Unlock()
	gen := h.index.Current()
	var refs []index.TraceRef
	for _, ref := range gen.Traces {
		refs = append(refs, ref)
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].TraceID < refs[j].TraceID })
	if limit > 0 && len(refs) > limit {
		refs = refs[:limit]
	}
	views := make([]TraceView, 0, len(refs))
	for _, ref := range refs {
		views = append(views, h.traceViewLocked(ref))
	}
	return views
}

func (h *Hub) traceViewLocked(ref index.TraceRef) TraceView {
	view := TraceView{
		TraceID:    ref.TraceID,
		Namespace:  ref.Namespace,
		RootSpanID: ref.RootSpanID,
		SpanCount:  ref.SpanCount,
		State:      ref.State,
	}
	if result, ok := h.aggregator.ReadResult(ref.TraceID); ok {
		view.EndedAt = result.EndedAt
		for _, existing := range result.Spans {
			view.Spans = append(view.Spans, toSpanView(existing, h.spans.PayloadBytes(existing.Namespace, existing.SpanID)))
		}
	}
	return view
}
