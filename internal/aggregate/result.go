package aggregate

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	"traceflow/internal/span"
	"traceflow/internal/store"
	"traceflow/internal/trace"
)

// Result is one durable aggregation output.
type Result struct {
	TraceID    string      `json:"trace_id"`
	Namespace  string      `json:"namespace"`
	RootSpanID string      `json:"root_span_id"`
	State      string      `json:"state"`
	SpanCount  int         `json:"span_count"`
	EndedAt    int64       `json:"ended_at"`
	Spans      []span.Span `json:"spans"`
}

// BuildResult constructs a result from an assembled trace.
func BuildResult(tr trace.Trace, spans []span.Span) Result {
	return Result{
		TraceID:    tr.TraceID,
		Namespace:  tr.Namespace,
		RootSpanID: tr.RootSpanID,
		State:      tr.State,
		SpanCount:  len(spans),
		EndedAt:    time.Now().Unix(),
		Spans:      spans,
	}
}

func (a *Aggregator) writeResult(result Result) error {
	return store.WriteJSON(a.resultPath(result.TraceID), result)
}

func (a *Aggregator) resultPath(traceID string) string {
	return filepath.Join(a.paths.ResultsDir(), traceID+".result.json")
}

// ReadResult loads an aggregation result.
func (a *Aggregator) ReadResult(traceID string) (Result, bool) {
	var result Result
	if err := store.ReadJSON(a.resultPath(traceID), &result); err != nil {
		return Result{}, false
	}
	return result, true
}

// RecentResults returns the newest results on disk.
func (a *Aggregator) RecentResults(limit int) []Result {
	entries, err := os.ReadDir(a.paths.ResultsDir())
	if err != nil {
		return nil
	}
	var results []Result
	for _, entry := range entries {
		var result Result
		if err := store.ReadJSON(filepath.Join(a.paths.ResultsDir(), entry.Name()), &result); err != nil {
			continue
		}
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].EndedAt > results[j].EndedAt })
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results
}
