package ns

import (
	"path/filepath"

	"traceflow/internal/settings"
)

// SpanJournalPath returns the span journal file for a namespace.
func SpanJournalPath(paths settings.Paths, namespace Namespace) string {
	return filepath.Join(paths.SpansDir(), namespace.PathName()+".spans.jnl")
}

// SpanCommitPath returns the commit watermark for a namespace journal.
func SpanCommitPath(paths settings.Paths, namespace Namespace) string {
	return filepath.Join(paths.SpansDir(), namespace.PathName()+".commit.meta")
}
