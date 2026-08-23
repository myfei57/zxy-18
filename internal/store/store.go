// Package store provides durable file helpers shared across TraceFlow components.
package store

import (
	"errors"
	"os"
)

// ErrNotFound is returned when a durable record is absent.
var ErrNotFound = errors.New("store: record not found")

// Exists reports whether path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
