package report

import (
	"sync"

	"traceflow/internal/store"
)

// Cursor persists per-namespace progression watermarks.
type Cursor struct {
	mu   sync.Mutex
	path string
	pos  map[string]uint64
}

// OpenCursor creates a cursor backed by path.
func OpenCursor(path string) *Cursor {
	return &Cursor{path: path, pos: map[string]uint64{}}
}

// Load reads the cursor file.
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

// Position returns the durable watermark for a namespace.
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

// Positions returns a copy of all watermarks.
func (c *Cursor) Positions() map[string]uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]uint64, len(c.pos))
	for namespace, value := range c.pos {
		out[namespace] = value
	}
	return out
}
