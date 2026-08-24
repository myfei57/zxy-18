package index

// Current returns the live current generation.
func (idx *Index) Current() *Generation {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	if gen, ok := idx.generations[idx.current]; ok {
		return gen
	}
	return newGeneration(0)
}
