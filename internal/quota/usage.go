package quota

import "sort"

// Used returns the durable usage for a namespace.
func (l *Ledger) Used(namespace string) int {
	return l.Status(namespace).Used
}

// Limits returns the configured limits.
func (l *Ledger) Limits() map[string]int {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make(map[string]int, len(l.limits))
	for namespace, limit := range l.limits {
		out[namespace] = limit
	}
	return out
}

// Namespaces returns namespaces with any quota state.
func (l *Ledger) Namespaces() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	seen := map[string]bool{}
	for namespace := range l.used {
		seen[namespace] = true
	}
	for namespace := range l.limits {
		seen[namespace] = true
	}
	out := make([]string, 0, len(seen))
	for namespace := range seen {
		out = append(out, namespace)
	}
	sort.Strings(out)
	return out
}
