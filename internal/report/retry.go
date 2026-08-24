package report

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"traceflow/internal/store"
)

// BatchView couples a batch with its durable ack state.
type BatchView struct {
	Batch Batch `json:"batch"`
	Acked bool  `json:"acked"`
}

// RecentBatches returns the newest batches on disk.
func (r *Reporter) RecentBatches(limit int) []BatchView {
	r.mu.Lock()
	defer r.mu.Unlock()
	entries, err := os.ReadDir(r.paths.BatchesDir())
	if err != nil {
		return nil
	}
	var views []BatchView
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".batch.json") {
			continue
		}
		var batch Batch
		if err := store.ReadJSON(filepath.Join(r.paths.BatchesDir(), name), &batch); err != nil {
			continue
		}
		views = append(views, BatchView{Batch: batch, Acked: r.Acked(batch.ID)})
	}
	sort.Slice(views, func(i, j int) bool { return views[i].Batch.CreatedAt > views[j].Batch.CreatedAt })
	if limit > 0 && len(views) > limit {
		views = views[:limit]
	}
	return views
}
