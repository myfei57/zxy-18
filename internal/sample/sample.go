// Package sample decides which spans are kept and tracks the sampling budget.
package sample

import (
	"sync"

	"traceflow/internal/store"
)

// Sampler persists the sampling rate and budget decisions.
type Sampler struct {
	mu       sync.Mutex
	path     string
	Rate     int
	Capacity int
	Token    int
	Sampled  int
	Rejected int
}

// NewSampler creates a sampler backed by path.
func NewSampler(path string) *Sampler {
	return &Sampler{path: path}
}

// Load reads sampler state, opening a fresh budget window.
func (s *Sampler) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var doc struct {
		Rate     int `json:"rate"`
		Capacity int `json:"capacity"`
		Token    int `json:"token"`
		Sampled  int `json:"sampled"`
		Rejected int `json:"rejected"`
	}
	if err := store.ReadJSON(s.path, &doc); err != nil {
		if err == store.ErrNotFound {
			return nil
		}
		return err
	}
	s.Rate = doc.Rate
	s.Capacity = doc.Capacity
	s.Token = 0
	s.Sampled = doc.Sampled
	s.Rejected = doc.Rejected
	return nil
}

func (s *Sampler) saveLocked() error {
	doc := struct {
		Rate     int `json:"rate"`
		Capacity int `json:"capacity"`
		Token    int `json:"token"`
		Sampled  int `json:"sampled"`
		Rejected int `json:"rejected"`
	}{
		Rate:     s.Rate,
		Capacity: s.Capacity,
		Token:    s.Token,
		Sampled:  s.Sampled,
		Rejected: s.Rejected,
	}
	return store.WriteJSON(s.path, doc)
}

// State returns a console-friendly snapshot.
func (s *Sampler) State() State {
	s.mu.Lock()
	defer s.mu.Unlock()
	budget := s.Capacity*s.Rate/100 - s.Token
	if budget < 0 {
		budget = 0
	}
	return State{
		Rate:     s.Rate,
		Capacity: s.Capacity,
		Token:    s.Token,
		Sampled:  s.Sampled,
		Rejected: s.Rejected,
		Budget:   budget,
	}
}
