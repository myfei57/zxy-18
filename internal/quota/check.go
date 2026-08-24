package quota

import "fmt"

// Check verifies that adding delta to the namespace usage stays within limit.
func (l *Ledger) Check(namespace string, delta int) error {
	status := l.Status(namespace)
	if status.Used+delta > status.Limit {
		return fmt.Errorf("quota exceeded for namespace %s: used %d limit %d", namespace, status.Used, status.Limit)
	}
	return nil
}
