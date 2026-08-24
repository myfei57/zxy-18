package sample

import "fmt"

// ValidateRate checks that a rate is within 0..100.
func ValidateRate(rate int) error {
	if rate < 0 || rate > 100 {
		return fmt.Errorf("sample rate must be within 0 and 100, got %d", rate)
	}
	return nil
}
