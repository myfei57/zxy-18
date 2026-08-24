package trace

import "fmt"

var order = []string{
	StateReceiving,
	StateSampling,
	StateBuffered,
	StateReported,
	StateAggregated,
	StateQueryable,
}

// Advance moves a trace state forward in the lifecycle.
func Advance(t *Trace, next string) error {
	current := indexOfState(t.State)
	target := indexOfState(next)
	if current < 0 || target < 0 {
		return fmt.Errorf("unknown trace state %q or %q", t.State, next)
	}
	if target <= current {
		return fmt.Errorf("trace state %q cannot move to %q", t.State, next)
	}
	t.State = next
	return nil
}

func indexOfState(state string) int {
	for i, candidate := range order {
		if candidate == state {
			return i
		}
	}
	return -1
}

// Valid reports whether a state belongs to the lifecycle.
func Valid(state string) bool {
	return indexOfState(state) >= 0
}
