package console

import (
	"sort"

	"traceflow/internal/index"
	"traceflow/internal/span"
)

// IngestRequest is the span ingestion payload.
type IngestRequest struct {
	Namespace string `json:"namespace"`
	TraceID   string `json:"trace_id"`
	SpanID    string `json:"span_id"`
	ParentID  string `json:"parent_id,omitempty"`
	Service   string `json:"service"`
	Operation string `json:"operation"`
	StartedAt int64  `json:"started_at"`
}

// TraceView is the console representation of an indexed trace.
type TraceView struct {
	TraceID    string     `json:"trace_id"`
	Namespace  string     `json:"namespace"`
	RootSpanID string     `json:"root_span_id"`
	SpanCount  int        `json:"span_count"`
	State      string     `json:"state"`
	EndedAt    int64      `json:"ended_at,omitempty"`
	Spans      []spanView `json:"spans,omitempty"`
}

type spanView struct {
	TraceID      string `json:"trace_id"`
	SpanID       string `json:"span_id"`
	ParentID     string `json:"parent_id,omitempty"`
	Service      string `json:"service"`
	Operation    string `json:"operation"`
	StartedAt    int64  `json:"started_at"`
	FinishedAt   int64  `json:"finished_at,omitempty"`
	Status       string `json:"status"`
	Seq          uint64 `json:"seq"`
	PayloadBytes int    `json:"payload_bytes"`
}

// SpanDetail is a span with its payload body.
type SpanDetail struct {
	Span         spanView          `json:"span"`
	Payload      map[string]string `json:"payload,omitempty"`
	PayloadBytes int               `json:"payload_bytes"`
}

func toSpanView(s span.Span, payloadBytes int) spanView {
	return spanView{
		TraceID:      s.TraceID,
		SpanID:       s.SpanID,
		ParentID:     s.ParentID,
		Service:      s.Service,
		Operation:    s.Operation,
		StartedAt:    s.StartedAt,
		FinishedAt:   s.FinishedAt,
		Status:       s.Status,
		Seq:          s.Seq,
		PayloadBytes: payloadBytes,
	}
}

func countByNamespace(refs map[string]index.TraceRef) map[string]int {
	out := map[string]int{}
	for _, ref := range refs {
		out[ref.Namespace]++
	}
	return out
}

func sortedNamespaces(refs map[string]index.TraceRef) []string {
	seen := map[string]bool{}
	for _, ref := range refs {
		seen[ref.Namespace] = true
	}
	out := make([]string, 0, len(seen))
	for namespace := range seen {
		out = append(out, namespace)
	}
	sort.Strings(out)
	return out
}
