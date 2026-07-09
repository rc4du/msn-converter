package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// newTestUI builds the app UI against a headless toolkit.
func newTestUI(t *testing.T) *appUI {
	t.Helper()
	a := test.NewApp()
	t.Cleanup(a.Quit)
	return newAppUI(a)
}

// listRow renders row id of the queue list and returns its label and remove button.
func listRow(t *testing.T, u *appUI, id int) (*widget.Label, *widget.Button) {
	t.Helper()
	row := u.list.CreateItem().(*fyne.Container)
	u.list.UpdateItem(id, row)
	return row.Objects[0].(*widget.Label), row.Objects[1].(*widget.Button)
}

// GUI-16: window titled "MSN Converter", sized ~600×450.
func TestMainWindowShell(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	w := newMainWindow(a)

	if got := w.Title(); got != "MSN Converter" {
		t.Errorf("window title = %q; want %q", got, "MSN Converter")
	}
	size := w.Canvas().Size()
	if size.Width < 600 || size.Height < 450 {
		t.Errorf("window size = %v; want at least 600x450", size)
	}
}

// GUI-10: Convert enabled iff queue non-empty AND outDir set AND not running.
func TestConvertEnablementMatrix(t *testing.T) {
	cases := []struct {
		name    string
		files   []string
		outDir  string
		running bool
		enabled bool
	}{
		{"empty queue, no outDir", nil, "", false, false},
		{"files, no outDir", []string{"/tmp/a.xml"}, "", false, false},
		{"empty queue, outDir set", nil, "/tmp/out", false, false},
		{"files and outDir", []string{"/tmp/a.xml"}, "/tmp/out", false, true},
		{"files and outDir but running", []string{"/tmp/a.xml"}, "/tmp/out", true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := newTestUI(t)
			u.queue.Add(tc.files...)
			u.outDir = tc.outDir
			u.running = tc.running
			u.updateConvertState()
			if got := !u.convertBtn.Disabled(); got != tc.enabled {
				t.Errorf("convert enabled = %v; want %v", got, tc.enabled)
			}
		})
	}
}

// GUI-09: tapping a row's ✕ removes exactly that file from the list.
func TestRemoveButtonRemovesRow(t *testing.T) {
	u := newTestUI(t)
	u.queue.Add("/tmp/a.xml", "/tmp/b.xml")
	u.list.Refresh()

	_, remove := listRow(t, u, 0)
	test.Tap(remove)

	if got := u.queue.Items(); len(got) != 1 || got[0] != "/tmp/b.xml" {
		t.Errorf("queue after remove = %v; want [/tmp/b.xml]", got)
	}
}

// GUI-09 + GUI-10: clear-all empties the list and Convert re-disables.
func TestClearAllEmptiesListAndDisablesConvert(t *testing.T) {
	u := newTestUI(t)
	u.queue.Add("/tmp/a.xml", "/tmp/b.xml")
	u.outDir = "/tmp/out"
	u.updateConvertState()
	if u.convertBtn.Disabled() {
		t.Fatal("precondition failed: Convert should be enabled with files + outDir")
	}

	test.Tap(u.clearBtn)

	if u.queue.Len() != 0 {
		t.Errorf("queue.Len() = %d after clear-all; want 0", u.queue.Len())
	}
	if !u.convertBtn.Disabled() {
		t.Error("Convert enabled after clear-all emptied the list; want disabled")
	}
}

// GUI-09: rows display the file's base name.
func TestListRowShowsFileName(t *testing.T) {
	u := newTestUI(t)
	u.queue.Add("/tmp/logs/a.xml")
	u.list.Refresh()

	label, _ := listRow(t, u, 0)
	if label.Text != "a.xml" {
		t.Errorf("row label = %q; want %q", label.Text, "a.xml")
	}
}
