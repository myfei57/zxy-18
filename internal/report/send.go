package report

import "traceflow/internal/audit"

// Send durably batches, delivers to the sink, records the acknowledgement and
// only then advances the cursor.
func (r *Reporter) Send(namespace string) (Batch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	pending := r.spansSinceCursor(namespace)
	if len(pending) == 0 {
		return Batch{}, nil
	}
	batch := BuildBatch(namespace, pending)
	if err := r.writeBatch(batch); err != nil {
		return Batch{}, err
	}
	ackID, err := r.sink.Send(batch)
	if err != nil {
		return Batch{}, err
	}
	if err := r.cursor.Advance(namespace, pending[len(pending)-1].Seq); err != nil {
		return Batch{}, err
	}
	if err := r.persistAck(namespace, batch, ackID); err != nil {
		return Batch{}, err
	}
	r.audit.Record(audit.NewEvent(audit.KindSend, namespace, batch.ID, "ok"))
	return batch, nil
}
