package span

import (
	"fmt"
	"time"
)

// Finish marks a span finished only after its payload is durable.
func (st *Store) Finish(namespace, spanID string, payload map[string]string) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	existing, ok := st.spanByIDLocked(namespace, spanID)
	if !ok {
		return fmt.Errorf("span %s not found", spanID)
	}
	if existing.Finished() {
		return nil
	}
	// Land the payload first: only once the span data is fully durable do we
	// record the finish marker. A crash between these steps can then never
	// leave a span marked finished with an empty payload — at worst the span
	// stays receiving and the (idempotent) finish is retried.
	key := st.payloadKey(namespace, spanID)
	if err := st.writePayload(key, Payload{
		TraceID: existing.TraceID,
		SpanID:  spanID,
		Keys:    payload,
		Bytes:   payloadBytes(payload),
	}); err != nil {
		return err
	}
	existing.FinishedAt = time.Now().Unix()
	existing.Status = StatusFinished
	if _, err := st.appendRecordLocked(namespace, journalRecord{Kind: "finish", Span: existing}); err != nil {
		return err
	}
	st.spans[namespace] = upsertSpan(st.spans[namespace], existing)
	return nil
}
