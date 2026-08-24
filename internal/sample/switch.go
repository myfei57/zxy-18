package sample

// Switch changes the sampling rate and immediately rebuilds the budget from
// the current aggregation window count.
func (s *Sampler) Switch(rate, windowCount int) error {
	if err := ValidateRate(rate); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Rate = rate
	s.Capacity = windowCount
	s.Token = 0
	return s.saveLocked()
}
