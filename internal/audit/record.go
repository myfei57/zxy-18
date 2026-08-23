package audit

import "traceflow/internal/store"

// Logger appends events to a durable JSON-lines file.
type Logger struct {
	path string
}

// NewLogger creates an audit logger.
func NewLogger(path string) *Logger {
	return &Logger{path: path}
}

// Record durably appends an event.
func (l *Logger) Record(event Event) error {
	return store.AppendJSONLine(l.path, event)
}
