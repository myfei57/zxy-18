package quota

import (
	"fmt"
	"sync"

	"traceflow/internal/store"
)

// Ledger persists usage and limits per namespace.
type Ledger struct {
	mu     sync.Mutex
	path   string
	used   map[string]int
	limits map[string]int
}

// NewLedger creates a quota ledger.
func NewLedger(path string) *Ledger {
	return &Ledger{path: path, used: map[string]int{}, limits: map[string]int{}}
}

// Load reads quota state from disk.
func (l *Ledger) Load() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	var doc struct {
		Used   map[string]int `json:"used"`
		Limits map[string]int `json:"limits"`
	}
	if err := store.ReadJSON(l.path, &doc); err != nil {
		if err == store.ErrNotFound {
			return nil
		}
		return err
	}
	if doc.Used != nil {
		l.used = doc.Used
	}
	if doc.Limits != nil {
		l.limits = doc.Limits
	}
	return nil
}

func (l *Ledger) saveLocked() error {
	doc := struct {
		Used   map[string]int `json:"used"`
		Limits map[string]int `json:"limits"`
	}{Used: l.used, Limits: l.limits}
	return store.WriteJSON(l.path, doc)
}

// SetLimit configures the ceiling for a namespace.
func (l *Ledger) SetLimit(namespace string, limit int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if limit < 0 {
		return fmt.Errorf("negative quota limit")
	}
	l.limits[namespace] = limit
	return l.saveLocked()
}

// Occupy records a durable usage increment.
func (l *Ledger) Occupy(namespace string, delta int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.used[namespace] += delta
	return l.saveLocked()
}

// Status returns the quota snapshot for a namespace.
func (l *Ledger) Status(namespace string) Status {
	l.mu.Lock()
	defer l.mu.Unlock()
	limit := l.limits[namespace]
	if limit == 0 {
		limit = DefaultLimit
	}
	return Status{
		Namespace: namespace,
		Limit:     limit,
		Used:      l.used[namespace],
		Remaining: limit - l.used[namespace],
	}
}

// Limit returns the configured ceiling for a namespace.
func (l *Ledger) Limit(namespace string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	limit := l.limits[namespace]
	if limit == 0 {
		return DefaultLimit
	}
	return limit
}
