package settings

import (
	"fmt"
	"os"
	"path/filepath"
)

// Paths resolves every durable file location from a single data root.
type Paths struct {
	root string
}

// NewPaths builds a Paths rooted at root.
func NewPaths(root string) Paths {
	return Paths{root: root}
}

// Root returns the data root.
func (p Paths) Root() string {
	return p.root
}

// RegistryFile returns the namespace registry file.
func (p Paths) RegistryFile() string {
	return filepath.Join(p.root, "registry.json")
}

// QuotaFile returns the quota ledger file.
func (p Paths) QuotaFile() string {
	return filepath.Join(p.root, "quota.json")
}

// SampleFile returns the sampler state file.
func (p Paths) SampleFile() string {
	return filepath.Join(p.root, "sample.json")
}

// AuditFile returns the audit event log.
func (p Paths) AuditFile() string {
	return filepath.Join(p.root, "audit.log")
}

// SpansDir returns the span journal directory.
func (p Paths) SpansDir() string {
	return filepath.Join(p.root, "spans")
}

// PayloadsDir returns the span payload directory.
func (p Paths) PayloadsDir() string {
	return filepath.Join(p.root, "payloads")
}

// BatchesDir returns the report batch directory.
func (p Paths) BatchesDir() string {
	return filepath.Join(p.root, "batches")
}

// ResultsDir returns the aggregation result directory.
func (p Paths) ResultsDir() string {
	return filepath.Join(p.root, "results")
}

// SummariesDir returns the window summary directory.
func (p Paths) SummariesDir() string {
	return filepath.Join(p.root, "summaries")
}

// WindowsDir returns the window state directory.
func (p Paths) WindowsDir() string {
	return filepath.Join(p.root, "windows")
}

// IndexDir returns the index generation directory.
func (p Paths) IndexDir() string {
	return filepath.Join(p.root, "index")
}

// AckDir returns the sink acknowledgement directory.
func (p Paths) AckDir() string {
	return filepath.Join(p.root, "acks")
}

// CursorFile returns the cursor file for a named progression.
func (p Paths) CursorFile(name string) string {
	return filepath.Join(p.root, "cursors", name+".json")
}

// EnsureAll creates every directory TraceFlow writes to.
func (p Paths) EnsureAll() error {
	dirs := []string{
		p.SpansDir(),
		p.PayloadsDir(),
		p.BatchesDir(),
		p.ResultsDir(),
		p.SummariesDir(),
		p.WindowsDir(),
		p.IndexDir(),
		p.AckDir(),
		filepath.Dir(p.CursorFile("report")),
		filepath.Dir(p.CursorFile("aggregate")),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create data dir %s: %w", dir, err)
		}
	}
	return nil
}
