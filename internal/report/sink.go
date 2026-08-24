package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// Sink receives reported batches.
type Sink interface {
	Send(batch Batch) (string, error)
}

// LocalSink persists batches into an outbox directory.
type LocalSink struct {
	outDir string
}

// NewLocalSink creates a file-backed sink.
func NewLocalSink(outDir string) *LocalSink {
	return &LocalSink{outDir: outDir}
}

// Send writes the batch and returns an acknowledgement id.
func (s *LocalSink) Send(batch Batch) (string, error) {
	if err := os.MkdirAll(s.outDir, 0o755); err != nil {
		return "", fmt.Errorf("create sink outbox: %w", err)
	}
	payload, err := json.Marshal(batch)
	if err != nil {
		return "", err
	}
	path := filepath.Join(s.outDir, batch.ID+".json")
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return "", fmt.Errorf("sink write %s: %w", path, err)
	}
	return uuid.NewString(), nil
}
