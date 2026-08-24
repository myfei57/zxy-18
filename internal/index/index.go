package index

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"traceflow/internal/store"
)

// Index persists queryable trace generations under a directory.
type Index struct {
	mu          sync.Mutex
	dir         string
	generations map[int]*Generation
	current     int
	seq         int
}

// NewIndex creates an index rooted at dir.
func NewIndex(dir string) *Index {
	return &Index{
		dir:         dir,
		generations: map[int]*Generation{},
		current:     0,
		seq:         0,
	}
}

// Load restores the newest generation from disk.
func (idx *Index) Load() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	entries, err := os.ReadDir(idx.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var numbers []int
	for _, entry := range entries {
		var number int
		if _, err := fmt.Sscanf(entry.Name(), "gen-%d.json", &number); err == nil {
			numbers = append(numbers, number)
		}
	}
	if len(numbers) == 0 {
		return nil
	}
	sort.Ints(numbers)
	last := numbers[len(numbers)-1]
	var gen Generation
	if err := store.ReadJSON(idx.generationPath(last), &gen); err != nil {
		return err
	}
	idx.generations[last] = &gen
	idx.current = last
	idx.seq = last
	return nil
}

func (idx *Index) generationPath(number int) string {
	return filepath.Join(idx.dir, fmt.Sprintf("gen-%d.json", number))
}
