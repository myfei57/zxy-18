package ns

import (
	"fmt"
	"sort"
	"sync"

	"traceflow/internal/store"
)

// Registry persists the namespace set.
type Registry struct {
	mu    sync.Mutex
	path  string
	items map[string]Namespace
}

// NewRegistry creates an in-memory registry backed by path.
func NewRegistry(path string) *Registry {
	return &Registry{path: path, items: map[string]Namespace{}}
}

// Load reads the registry file if present.
func (r *Registry) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var doc struct {
		Namespaces []Namespace `json:"namespaces"`
	}
	if err := store.ReadJSON(r.path, &doc); err != nil {
		if err == store.ErrNotFound {
			return nil
		}
		return err
	}
	for _, namespace := range doc.Namespaces {
		r.items[namespace.Name] = namespace
	}
	return nil
}

// Register adds or updates a namespace and persists it.
func (r *Registry) Register(namespace Namespace) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if namespace.Name == "" {
		return fmt.Errorf("namespace name is required")
	}
	r.items[namespace.Name] = namespace
	return r.saveLocked()
}

// Deactivate marks a namespace inactive without removing its history.
func (r *Registry) Deactivate(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	namespace, ok := r.items[name]
	if !ok {
		return fmt.Errorf("namespace %s not found", name)
	}
	namespace.Active = false
	r.items[name] = namespace
	return r.saveLocked()
}

func (r *Registry) saveLocked() error {
	names := make([]string, 0, len(r.items))
	for name := range r.items {
		names = append(names, name)
	}
	sort.Strings(names)
	doc := struct {
		Namespaces []Namespace `json:"namespaces"`
	}{}
	for _, name := range names {
		doc.Namespaces = append(doc.Namespaces, r.items[name])
	}
	return store.WriteJSON(r.path, doc)
}

// Lookup returns a namespace by name.
func (r *Registry) Lookup(name string) (Namespace, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	namespace, ok := r.items[name]
	return namespace, ok
}

// List returns all namespaces sorted by name.
func (r *Registry) List() []Namespace {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Namespace, 0, len(r.items))
	for _, namespace := range r.items {
		out = append(out, namespace)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Count returns the number of registered namespaces.
func (r *Registry) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.items)
}
