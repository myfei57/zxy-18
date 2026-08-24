// Package audit records durable events for reporting and aggregation.
package audit

import (
	"time"

	"github.com/google/uuid"
)

// Event kinds.
const (
	KindReport    = "report"
	KindSend      = "send"
	KindAggregate = "aggregate"
	KindQuota     = "quota"
	KindSample    = "sample"
	KindIndex     = "index"
)

// Event is one durable audit record.
type Event struct {
	ID        string            `json:"id"`
	Kind      string            `json:"kind"`
	Namespace string            `json:"namespace,omitempty"`
	Target    string            `json:"target,omitempty"`
	Result    string            `json:"result"`
	Detail    map[string]string `json:"detail,omitempty"`
	At        int64             `json:"at"`
}

// NewEvent creates an event with a uuid identity.
func NewEvent(kind, namespace, target, result string) Event {
	return Event{
		ID:        uuid.NewString(),
		Kind:      kind,
		Namespace: namespace,
		Target:    target,
		Result:    result,
		At:        time.Now().Unix(),
	}
}
