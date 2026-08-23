package span

import (
	"path/filepath"

	"traceflow/internal/store"
)

// Payload is the durable body attached to a finished span.
type Payload struct {
	TraceID string            `json:"trace_id"`
	SpanID  string            `json:"span_id"`
	Keys    map[string]string `json:"keys"`
	Bytes   int               `json:"bytes"`
}

func (st *Store) payloadKey(namespace, spanID string) string {
	return filepath.Join(st.paths.PayloadsDir(), namespace+"."+spanID+".payload")
}

func payloadBytes(keys map[string]string) int {
	total := 0
	for key, value := range keys {
		total += len(key) + len(value)
	}
	return total
}

func (st *Store) writePayload(key string, p Payload) error {
	return store.WriteJSON(key, p)
}

// ReadPayload loads a span payload by id.
func (st *Store) ReadPayload(namespace, spanID string) (Payload, error) {
	var p Payload
	if err := store.ReadJSON(st.payloadKey(namespace, spanID), &p); err != nil {
		return Payload{}, err
	}
	return p, nil
}

// PayloadBytes returns the durable payload size for a span.
func (st *Store) PayloadBytes(namespace, spanID string) int {
	p, err := st.ReadPayload(namespace, spanID)
	if err != nil {
		return 0
	}
	return p.Bytes
}
