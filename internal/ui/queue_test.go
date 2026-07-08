package ui

import (
	"path/filepath"
	"testing"
)

// GUI-08: adding a file already in the list (same absolute path, including
// non-clean variants) leaves the list unchanged and reports 0 added.
func TestQueueAddDeduplicates(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.xml")

	var q Queue
	if got := q.Add(a); got != 1 {
		t.Fatalf("first Add = %d added; want 1", got)
	}
	if got := q.Add(a); got != 0 {
		t.Errorf("duplicate Add = %d added; want 0", got)
	}

	// Non-clean variant of the same absolute path: dir/../<base>/a.xml.
	nonClean := dir + string(filepath.Separator) + ".." + string(filepath.Separator) +
		filepath.Base(dir) + string(filepath.Separator) + "a.xml"
	if got := q.Add(nonClean); got != 0 {
		t.Errorf("non-clean duplicate Add = %d added; want 0", got)
	}

	if got := q.Items(); len(got) != 1 || got[0] != a {
		t.Errorf("Items() = %v; want [%s]", got, a)
	}
	if q.Len() != 1 {
		t.Errorf("Len() = %d; want 1", q.Len())
	}
}

// GUI-08 (order): distinct paths are kept in insertion order.
func TestQueueAddKeepsInsertionOrder(t *testing.T) {
	dir := t.TempDir()
	b := filepath.Join(dir, "b.xml")
	a := filepath.Join(dir, "a.xml")
	c := filepath.Join(dir, "c.xml")

	var q Queue
	if got := q.Add(b, a, c); got != 3 {
		t.Fatalf("Add(b,a,c) = %d added; want 3", got)
	}
	got := q.Items()
	want := []string{b, a, c}
	if len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Errorf("Items() = %v; want %v", got, want)
	}
}

// GUI-09 (logic): removing an item takes that file out of the list.
func TestQueueRemove(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.xml")
	b := filepath.Join(dir, "b.xml")

	var q Queue
	q.Add(a, b)
	q.Remove(a)

	if got := q.Items(); len(got) != 1 || got[0] != b {
		t.Errorf("after Remove(a): Items() = %v; want [%s]", got, b)
	}

	// Removed path can be re-added (it is genuinely gone, not just hidden).
	if got := q.Add(a); got != 1 {
		t.Errorf("re-Add after Remove = %d added; want 1", got)
	}
}

// GUI-09 (logic): clear-all empties the list.
func TestQueueClear(t *testing.T) {
	dir := t.TempDir()

	var q Queue
	q.Add(filepath.Join(dir, "a.xml"), filepath.Join(dir, "b.xml"))
	q.Clear()

	if q.Len() != 0 {
		t.Errorf("Len() after Clear = %d; want 0", q.Len())
	}
	if got := q.Items(); len(got) != 0 {
		t.Errorf("Items() after Clear = %v; want empty", got)
	}
	// Cleared paths can be re-added.
	if got := q.Add(filepath.Join(dir, "a.xml")); got != 1 {
		t.Errorf("Add after Clear = %d added; want 1", got)
	}
}

// Items returns a defensive copy — mutating it does not affect the queue.
func TestQueueItemsDefensiveCopy(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.xml")

	var q Queue
	q.Add(a)

	items := q.Items()
	items[0] = "mutated"

	if got := q.Items(); got[0] != a {
		t.Errorf("queue affected by mutation of Items() result: %v; want [%s]", got, a)
	}
}
