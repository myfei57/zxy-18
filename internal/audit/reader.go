package audit

import (
	"encoding/json"

	"traceflow/internal/store"
)

// Recent returns the most recent events up to limit.
func (l *Logger) Recent(limit int) []Event {
	rows, err := store.ReadRawLines(l.path)
	if err != nil {
		return nil
	}
	var all []Event
	for _, row := range rows {
		var event Event
		if err := json.Unmarshal(row, &event); err != nil {
			continue
		}
		all = append(all, event)
	}
	if limit <= 0 || len(all) <= limit {
		return all
	}
	return all[len(all)-limit:]
}
