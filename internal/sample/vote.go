package sample

import "traceflow/internal/span"

// Vote decides whether a span is sampled and durably records the verdict only
// after the span data has been stored.
func (s *Sampler) Vote(st *span.Store, namespace string, sp span.Span) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Capacity <= 0 {
		s.Rejected++
		return false, s.saveLocked()
	}
	if s.Token >= s.Capacity {
		s.Rejected++
		return false, s.saveLocked()
	}
	s.Token++
	limit := s.Capacity * s.Rate / 100
	if s.Token > limit {
		s.Rejected++
		return false, s.saveLocked()
	}
	if _, err := st.Append(sp); err != nil {
		return false, err
	}
	s.Sampled++
	return true, s.saveLocked()
}
