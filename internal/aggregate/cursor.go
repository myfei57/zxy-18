package aggregate

import (
	"sync"

	"traceflow/internal/store"
)

// Cursor persists the aggregation watermark per namespace.
type Cursor struct {
	mu   sync.Mutex
	path string
	pos  map[string]uint64
}

// OpenCursor creates the aggregate cursor.
func OpenCursor(path string) *Cursor {
	return &Cursor{path: path, pos: map[string]uint64{}}
}

// Load reads the watermark file.
func (c *Cursor) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var doc struct {
		Positions map[string]uint64 `json:"positions"`
	}
	if err := store.ReadJSON(c.path, &doc); err != nil {
		if err == store.ErrNotFound {
			return nil
		}
		return err
	}
	if doc.Positions != nil {
		c.pos = doc.Positions
	}
	return nil
}

// Position returns the watermark for a namespace.
func (c *Cursor) Position(namespace string) uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.pos[namespace]
}

// Advance durably moves the watermark forward.
func (c *Cursor) Advance(namespace string, value uint64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if value > c.pos[namespace] {
		c.pos[namespace] = value
	}
	doc := struct {
		Positions map[string]uint64 `json:"positions"`
	}{Positions: c.pos}
	return store.WriteJSON(c.path, doc)
}
