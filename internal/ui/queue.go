package ui

import (
	"path/filepath"
	"slices"
)

// Queue is an ordered, de-duplicated (by absolute path) list of input files.
type Queue struct {
	items []string
	seen  map[string]bool
}

// Add appends each path not already present (compared by absolute, cleaned
// path), preserving insertion order. It returns how many were added.
func (q *Queue) Add(paths ...string) int {
	added := 0
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if q.seen[abs] {
			continue
		}
		if q.seen == nil {
			q.seen = make(map[string]bool)
		}
		q.seen[abs] = true
		q.items = append(q.items, abs)
		added++
	}
	return added
}

// Remove takes the given path out of the queue, if present.
func (q *Queue) Remove(path string) {
	abs, err := filepath.Abs(path)
	if err != nil || !q.seen[abs] {
		return
	}
	delete(q.seen, abs)
	i := slices.Index(q.items, abs)
	q.items = slices.Delete(q.items, i, i+1)
}

// Clear empties the queue.
func (q *Queue) Clear() {
	q.items = nil
	q.seen = nil
}

// Items returns a snapshot copy of the queued paths.
func (q *Queue) Items() []string {
	return slices.Clone(q.items)
}

// Len reports how many files are queued.
func (q *Queue) Len() int {
	return len(q.items)
}
