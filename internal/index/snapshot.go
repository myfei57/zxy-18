package index

import "sort"

// Snapshot returns the current generation as a stable read-only copy.
func (idx *Index) Snapshot() *Generation {
	gen := idx.Current()
	copied := newGeneration(gen.Number)
	for id, ref := range gen.Traces {
		copied.Traces[id] = ref
	}
	return copied
}

// Stats summarises the current generation for the console.
func (idx *Index) Stats() IndexStats {
	gen := idx.Current()
	byNamespace := map[string]int{}
	for _, ref := range gen.Traces {
		byNamespace[ref.Namespace]++
	}
	namespaces := make([]string, 0, len(byNamespace))
	for namespace := range byNamespace {
		namespaces = append(namespaces, namespace)
	}
	sort.Strings(namespaces)
	return IndexStats{
		Generation:  gen.Number,
		TraceCount:  gen.Count(),
		ByNamespace: byNamespace,
		Namespaces:  namespaces,
	}
}

// IndexStats is a generation summary for operators.
type IndexStats struct {
	Generation  int            `json:"generation"`
	TraceCount  int            `json:"trace_count"`
	ByNamespace map[string]int `json:"by_namespace"`
	Namespaces  []string       `json:"namespaces"`
}
