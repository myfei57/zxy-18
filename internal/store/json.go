package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteJSON durably writes v to path via a temporary file and rename.
func WriteJSON(path string, v any) error {
	dir := filepath.Dir(path)
	payload, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp for %s: %w", path, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(payload); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp for %s: %w", path, err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("sync temp for %s: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp for %s: %w", path, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename onto %s: %w", path, err)
	}
	return nil
}

// ReadJSON loads a JSON document from path.
func ReadJSON(path string, v any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}
