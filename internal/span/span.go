// Package span stores sampled trace spans and their payloads durably.
package span

import "fmt"

// Status values recorded on spans.
const (
	StatusReceiving = "receiving"
	StatusFinished  = "finished"
)

// Span is one remote call interval inside a trace.
type Span struct {
	TraceID    string            `json:"trace_id"`
	SpanID     string            `json:"span_id"`
	ParentID   string            `json:"parent_id,omitempty"`
	Namespace  string            `json:"namespace"`
	Service    string            `json:"service"`
	Operation  string            `json:"operation"`
	StartedAt  int64             `json:"started_at"`
	FinishedAt int64             `json:"finished_at,omitempty"`
	Status     string            `json:"status"`
	Seq        uint64            `json:"seq"`
	Payload    map[string]string `json:"payload,omitempty"`
}

// New creates a receiving span.
func New(traceID, spanID, parentID, service, operation string) Span {
	return Span{
		TraceID:   traceID,
		SpanID:    spanID,
		ParentID:  parentID,
		Service:   service,
		Operation: operation,
		Status:    StatusReceiving,
	}
}

// Finished reports whether the span carries a durable finish marker.
func (s Span) Finished() bool {
	return s.Status == StatusFinished
}

// Validate checks the minimal invariants of an incoming span.
func (s Span) Validate() error {
	if s.TraceID == "" {
		return fmt.Errorf("span trace_id is required")
	}
	if s.SpanID == "" {
		return fmt.Errorf("span span_id is required")
	}
	if s.Service == "" {
		return fmt.Errorf("span service is required")
	}
	if s.Operation == "" {
		return fmt.Errorf("span operation is required")
	}
	return nil
}
