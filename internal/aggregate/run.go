package aggregate

import (
	"traceflow/internal/audit"
	"traceflow/internal/span"
	"traceflow/internal/trace"
)

// Run merges one trace, durably writes the result and only then advances the
// cursor.
func (a *Aggregator) Run(namespace, traceID string) (Result, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	spans := a.spans.ForTrace(namespace, traceID)
	tr, _, err := trace.Assemble(namespace, traceID, spans)
	if err != nil {
		return Result{}, err
	}
	if err := trace.Advance(&tr, trace.StateAggregated); err != nil {
		return Result{}, err
	}
	result := BuildResult(tr, spans)
	if err := a.writeResult(result); err != nil {
		a.audit.Record(audit.NewEvent(audit.KindAggregate, namespace, traceID, "failed"))
		return Result{}, err
	}
	if err := a.cursor.Advance(namespace, maxSeq(spans)); err != nil {
		a.audit.Record(audit.NewEvent(audit.KindAggregate, namespace, traceID, "failed"))
		return Result{}, err
	}
	a.audit.Record(audit.NewEvent(audit.KindAggregate, namespace, traceID, "ok"))
	return result, nil
}

func maxSeq(spans []span.Span) uint64 {
	var max uint64
	for _, existing := range spans {
		if existing.Seq > max {
			max = existing.Seq
		}
	}
	return max
}
