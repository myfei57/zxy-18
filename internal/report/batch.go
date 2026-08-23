package report

import (
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"traceflow/internal/audit"
	"traceflow/internal/span"
	"traceflow/internal/store"
)

// Batch is one durable report unit.
type Batch struct {
	ID        string      `json:"id"`
	Namespace string      `json:"namespace"`
	Spans     []span.Span `json:"spans"`
	CreatedAt int64       `json:"created_at"`
}

// BuildBatch assembles spans into a batch.
func BuildBatch(namespace string, spans []span.Span) Batch {
	return Batch{
		ID:        uuid.NewString(),
		Namespace: namespace,
		Spans:     spans,
		CreatedAt: time.Now().Unix(),
	}
}

// Flush durably writes buffered spans as a batch, records the audit event and
// only then advances the cursor.
func (r *Reporter) Flush(namespace string) (Batch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	pending := r.spansSinceCursor(namespace)
	if len(pending) == 0 {
		return Batch{}, nil
	}
	batch := BuildBatch(namespace, pending)
	if err := r.writeBatch(batch); err != nil {
		r.audit.Record(audit.NewEvent(audit.KindReport, namespace, batch.ID, "failed"))
		return Batch{}, err
	}
	r.audit.Record(audit.NewEvent(audit.KindReport, namespace, batch.ID, "ok"))
	if err := r.cursor.Advance(namespace, pending[len(pending)-1].Seq); err != nil {
		return Batch{}, err
	}
	return batch, nil
}

func (r *Reporter) writeBatch(batch Batch) error {
	return store.WriteJSON(r.batchPath(batch.ID), batch)
}

func (r *Reporter) batchPath(id string) string {
	return filepath.Join(r.paths.BatchesDir(), id+".batch.json")
}
