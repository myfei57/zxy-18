package report

import (
	"path/filepath"
	"time"

	"traceflow/internal/store"
)

// ackRecord is a durable sink acknowledgement.
type ackRecord struct {
	BatchID   string `json:"batch_id"`
	Namespace string `json:"namespace"`
	AckID     string `json:"ack_id"`
	At        int64  `json:"at"`
}

// persistAck durably records a sink acknowledgement.
func (r *Reporter) persistAck(namespace string, batch Batch, ackID string) error {
	return store.WriteJSON(r.ackPath(batch.ID), ackRecord{
		BatchID:   batch.ID,
		Namespace: namespace,
		AckID:     ackID,
		At:        time.Now().Unix(),
	})
}

func (r *Reporter) ackPath(batchID string) string {
	return filepath.Join(r.paths.AckDir(), batchID+".ack.json")
}

// Acked reports whether a batch has a durable ack.
func (r *Reporter) Acked(batchID string) bool {
	return store.Exists(r.ackPath(batchID))
}
