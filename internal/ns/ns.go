// Package ns owns namespace registration and per-namespace isolation.
package ns

import "time"

// Namespace is a tenant boundary for trace ingestion.
type Namespace struct {
	Name      string `json:"name"`
	Service   string `json:"service"`
	CreatedAt int64  `json:"created_at"`
	Active    bool   `json:"active"`
}

// New creates an active namespace record.
func New(name, service string) Namespace {
	return Namespace{Name: name, Service: service, CreatedAt: time.Now().Unix(), Active: true}
}

// PathName returns the filesystem-safe identity used in file names.
func (n Namespace) PathName() string {
	return n.Name
}
