package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// Record is one recovered journal entry.
type Record struct {
	Seq     uint64          `json:"seq"`
	Payload json.RawMessage `json:"payload"`
}

// Journal is an append-only, sync-per-write log with a durable commit watermark.
type Journal struct {
	path      string
	commit    string
	committed uint64
	file      *os.File
}

// OpenJournal opens (creating) the log and its commit marker.
func OpenJournal(path, commit string) (*Journal, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open journal %s: %w", path, err)
	}
	j := &Journal{path: path, commit: commit, file: file}
	var meta struct {
		Committed uint64 `json:"committed"`
	}
	if err := ReadJSON(commit, &meta); err == nil {
		j.committed = meta.Committed
	} else if err != ErrNotFound {
		_ = file.Close()
		return nil, err
	}
	return j, nil
}

// Append writes one record line and syncs the log.
func (j *Journal) Append(seq uint64, payload []byte) error {
	line, err := json.Marshal(Record{Seq: seq, Payload: payload})
	if err != nil {
		return err
	}
	if _, err := j.file.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("journal append %s: %w", j.path, err)
	}
	return j.file.Sync()
}

// Commit durably advances the commit watermark.
func (j *Journal) Commit(seq uint64) error {
	if seq <= j.committed {
		return nil
	}
	if err := WriteJSON(j.commit, struct {
		Committed uint64 `json:"committed"`
	}{Committed: seq}); err != nil {
		return err
	}
	j.committed = seq
	return nil
}

// Close closes the underlying log.
func (j *Journal) Close() error {
	return j.file.Close()
}

// Recover replays a journal and returns the committed records in order.
func Recover(path, commit string) ([]Record, error) {
	var meta struct {
		Committed uint64 `json:"committed"`
	}
	_ = ReadJSON(commit, &meta)
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open journal %s: %w", path, err)
	}
	defer file.Close()
	var records []Record
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var rec Record
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			return nil, fmt.Errorf("decode journal %s: %w", path, err)
		}
		if rec.Seq <= meta.Committed {
			records = append(records, rec)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return records, nil
}
