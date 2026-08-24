package audit

import (
	"encoding/json"

	"traceflow/internal/store"
)

// Stats returns event counts per kind.
func (l *Logger) Stats() map[string]int {
	counts := map[string]int{}
	rows, err := store.ReadRawLines(l.path)
	if err != nil {
		return counts
	}
	for _, row := range rows {
		var event Event
		if err := json.Unmarshal(row, &event); err != nil {
			continue
		}
		counts[event.Kind]++
	}
	return counts
}
