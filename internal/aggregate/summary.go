package aggregate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"traceflow/internal/span"
	"traceflow/internal/store"
)

// Summary is a durable window summary.
type Summary struct {
	ID            string `json:"id"`
	WindowID      string `json:"window_id"`
	Namespace     string `json:"namespace"`
	SpanCount     int    `json:"span_count"`
	TraceCount    int    `json:"trace_count"`
	FinishedCount int    `json:"finished_count"`
	Start         int64  `json:"start"`
	End           int64  `json:"end"`
	ClosedAt      int64  `json:"closed_at"`
}

// BuildSummary computes a window summary from its spans.
func BuildSummary(w Window, spans []span.Span) Summary {
	traces := map[string]bool{}
	finished := 0
	for _, existing := range spans {
		traces[existing.TraceID] = true
		if existing.Finished() {
			finished++
		}
	}
	return Summary{
		ID:            fmt.Sprintf("sum-%s", w.ID),
		WindowID:      w.ID,
		Namespace:     w.Namespace,
		SpanCount:     len(spans),
		TraceCount:    len(traces),
		FinishedCount: finished,
		Start:         w.Start,
		End:           w.End,
		ClosedAt:      time.Now().Unix(),
	}
}

func (a *Aggregator) writeSummary(summary Summary) error {
	return store.WriteJSON(a.summaryPath(summary.ID), summary)
}

func (a *Aggregator) summaryPath(id string) string {
	return filepath.Join(a.paths.SummariesDir(), id+".summary.json")
}

// ReadSummary loads a window summary.
func (a *Aggregator) ReadSummary(id string) (Summary, bool) {
	var summary Summary
	if err := store.ReadJSON(a.summaryPath(id), &summary); err != nil {
		return Summary{}, false
	}
	return summary, true
}

// RecentSummaries returns the newest summaries on disk.
func (a *Aggregator) RecentSummaries(limit int) []Summary {
	entries, err := os.ReadDir(a.paths.SummariesDir())
	if err != nil {
		return nil
	}
	var summaries []Summary
	for _, entry := range entries {
		var summary Summary
		if err := store.ReadJSON(filepath.Join(a.paths.SummariesDir(), entry.Name()), &summary); err != nil {
			continue
		}
		summaries = append(summaries, summary)
	}
	sort.Slice(summaries, func(i, j int) bool { return summaries[i].ClosedAt > summaries[j].ClosedAt })
	if limit > 0 && len(summaries) > limit {
		summaries = summaries[:limit]
	}
	return summaries
}
