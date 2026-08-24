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
	s.Sampled++
	if _, err := st.Append(sp); err != nil {
		// Span data failed to land: roll back the verdict so the sampling
		// budget is not consumed for a span that cannot be queried.
		s.Sampled--
		s.Token--
		return false, err
	}
	if err := s.saveLocked(); err != nil {
		return false, err
	}
	return true, nil
}
