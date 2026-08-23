// Package trace assembles spans into traces and exposes query access.
package trace

// State values of the trace lifecycle.
const (
	StateReceiving  = "receiving"
	StateSampling   = "sampling"
	StateBuffered   = "buffered"
	StateReported   = "reported"
	StateAggregated = "aggregated"
	StateQueryable  = "queryable"
)

// Trace is the assembled view of one trace id.
type Trace struct {
	TraceID    string `json:"trace_id"`
	Namespace  string `json:"namespace"`
	RootSpanID string `json:"root_span_id"`
	State      string `json:"state"`
	SpanCount  int    `json:"span_count"`
}

// New creates a receiving trace record.
func New(traceID, namespace string) Trace {
	return Trace{TraceID: traceID, Namespace: namespace, State: StateReceiving}
}
