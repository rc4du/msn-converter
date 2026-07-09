package ui

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
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

// GUI-06/08: addFiles appends, de-dups, and refreshes gating.
func TestAddFilesAppendsDedupsAndRefreshes(t *testing.T) {
	u := newTestUI(t)
	u.outDir = "/tmp/out"

	u.addFiles([]string{"/tmp/a.xml", "/tmp/b.xml"})
	u.addFiles([]string{"/tmp/a.xml"}) // duplicate — list unchanged

	got := u.queue.Items()
	want := []string{"/tmp/a.xml", "/tmp/b.xml"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("queue = %v; want %v", got, want)
	}
	// Refresh happened: gating was recomputed after the mutation.
	if u.convertBtn.Disabled() {
		t.Error("Convert disabled after addFiles with outDir set; want enabled (refresh ran)")
	}
}

// GUI-10: setOutputDir updates the path label and Convert gating.
func TestSetOutputDirUpdatesLabelAndGating(t *testing.T) {
	u := newTestUI(t)
	u.queue.Add("/tmp/a.xml")
	u.updateConvertState()
	if !u.convertBtn.Disabled() {
		t.Fatal("precondition failed: Convert should be disabled without outDir")
	}

	u.setOutputDir("/tmp/out")

	if u.outDir != "/tmp/out" {
		t.Errorf("outDir = %q; want %q", u.outDir, "/tmp/out")
	}
	if u.outLabel.Text != "/tmp/out" {
		t.Errorf("output label = %q; want %q", u.outLabel.Text, "/tmp/out")
	}
	if u.convertBtn.Disabled() {
		t.Error("Convert disabled after setOutputDir with non-empty queue; want enabled")
	}
}

// GUI-06: the add-files dialog filter accepts only .xml.
func TestXMLFilterAcceptsOnlyXML(t *testing.T) {
	u := newTestUI(t)
	if !u.xmlFilter.Matches(storage.NewFileURI("/tmp/a.xml")) {
		t.Error("filter rejects /tmp/a.xml; want match")
	}
	if u.xmlFilter.Matches(storage.NewFileURI("/tmp/b.txt")) {
		t.Error("filter matches /tmp/b.txt; want reject")
	}
}

// Cancelled pickers (nil URI) are no-ops: queue, outDir and label unchanged.
func TestPickerCancelIsNoOp(t *testing.T) {
	u := newTestUI(t)
	labelBefore := u.outLabel.Text

	u.onFilePicked(nil, nil)
	u.onFolderPicked(nil, nil)
	u.onOutputPicked(nil, nil)

	if u.queue.Len() != 0 {
		t.Errorf("queue.Len() = %d after cancelled pickers; want 0", u.queue.Len())
	}
	if u.outDir != "" {
		t.Errorf("outDir = %q after cancelled picker; want empty", u.outDir)
	}
	if u.outLabel.Text != labelBefore {
		t.Errorf("output label = %q after cancelled picker; want %q", u.outLabel.Text, labelBefore)
	}
}

// GUI-07: picking a folder appends every .xml directly inside it.
func TestOnFolderPickedAddsXMLFiles(t *testing.T) {
	u := newTestUI(t)
	dir := t.TempDir()
	for _, name := range []string{"a.xml", "c.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatalf("seeding %s: %v", name, err)
		}
	}

	u.onFolderPicked(mustListable(t, dir), nil)

	got := u.queue.Items()
	if len(got) != 1 || filepath.Base(got[0]) != "a.xml" {
		t.Errorf("queue = %v; want just a.xml from %s", got, dir)
	}
}

// GUI-15: mixed drop — .xml files (case-insensitive) and folders' direct .xml
// are appended (deduped); non-xml files and nonexistent paths are ignored.
func TestAddDroppedMixed(t *testing.T) {
	u := newTestUI(t)
	dir := t.TempDir()

	seed := func(rel string) string {
		p := filepath.Join(dir, rel)
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatalf("seeding %s: %v", rel, err)
		}
		return p
	}
	xml := seed("a.xml")
	upper := seed("B.XML")
	txt := seed("c.txt")

	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	inSub := filepath.Join(sub, "d.xml")
	if err := os.WriteFile(inSub, []byte("x"), 0o644); err != nil {
		t.Fatalf("seeding sub/d.xml: %v", err)
	}

	u.addDropped([]string{
		xml,
		upper,
		txt,                               // non-xml → ignored
		sub,                               // folder → its direct .xml added
		filepath.Join(dir, "missing.xml"), // nonexistent → ignored
		xml,                               // duplicate → deduped
	})

	got := u.queue.Items()
	want := []string{xml, upper, inSub}
	if len(got) != len(want) {
		t.Fatalf("queue = %v; want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("queue[%d] = %q; want %q", i, got[i], want[i])
		}
	}
}

// GUI-11: starting a batch snapshots the queue, marks running, disables
// Convert and resets progress — synchronously, before the goroutine runs.
func TestBeginBatchDisablesConvertDuringRun(t *testing.T) {
	u := newTestUI(t)
	u.queue.Add("/tmp/a.xml", "/tmp/b.xml")
	u.setOutputDir("/tmp/out")
	u.progress.SetValue(0.7)

	files, outDir := u.beginBatch()

	if len(files) != 2 || files[0] != "/tmp/a.xml" || files[1] != "/tmp/b.xml" {
		t.Errorf("snapshot = %v; want [/tmp/a.xml /tmp/b.xml]", files)
	}
	if outDir != "/tmp/out" {
		t.Errorf("outDir snapshot = %q; want %q", outDir, "/tmp/out")
	}
	if !u.running {
		t.Error("running = false after beginBatch; want true")
	}
	if !u.convertBtn.Disabled() {
		t.Error("Convert enabled during run; want disabled")
	}
	if u.progress.Value != 0 {
		t.Errorf("progress = %v at batch start; want 0", u.progress.Value)
	}
}

// GUI-11/12/13/14/18: tapping Convert converts real fixtures off the UI
// goroutine, drives progress to 1.0, re-enables Convert, and keeps the queue.
func TestConvertTapRunsBatch(t *testing.T) {
	u := newTestUI(t)
	inDir := t.TempDir()
	outDir := t.TempDir()

	valid := writeXML(t, inDir, "valid.xml", "16/9/2010", "22:45:43", "Ricardo", "oi")
	malformed := filepath.Join(inDir, "malformed.xml")
	if err := os.WriteFile(malformed, []byte("<Log><Message Date="), 0o644); err != nil {
		t.Fatalf("seeding malformed.xml: %v", err)
	}

	u.addFiles([]string{valid, malformed})
	u.setOutputDir(outDir)
	queueBefore := u.queue.Items()

	done := make(chan Summary, 1)
	u.batchDone = func(s Summary) { done <- s }

	test.Tap(u.convertBtn)
	s := <-done

	// Expected output file written (GUI-11 mechanics via ConvertFile).
	if _, err := os.Stat(filepath.Join(outDir, "16_9_2010_22-45-43_Ricardo.txt")); err != nil {
		t.Errorf("expected output file missing: %v", err)
	}
	// Skip-and-report (GUI-12): counts exact.
	if s.Converted != 1 || len(s.Failed) != 1 {
		t.Errorf("Summary = %+v; want 1 converted, 1 failed", s)
	}
	// Progress reached 1.0 (GUI-11).
	if u.progress.Value != 1.0 {
		t.Errorf("progress = %v after batch; want 1.0", u.progress.Value)
	}
	// Re-enabled after run (GUI-10) and no longer running.
	if u.running {
		t.Error("running = true after batch; want false")
	}
	if u.convertBtn.Disabled() {
		t.Error("Convert disabled after batch finished; want re-enabled")
	}
	// Queue unchanged (GUI-18 / spec: list kept as-is).
	queueAfter := u.queue.Items()
	if len(queueAfter) != len(queueBefore) {
		t.Fatalf("queue after batch = %v; want unchanged %v", queueAfter, queueBefore)
	}
	for i := range queueBefore {
		if queueAfter[i] != queueBefore[i] {
			t.Errorf("queue[%d] = %q; want %q", i, queueAfter[i], queueBefore[i])
		}
	}
}

// GUI-13: summary is "N converted, M failed" plus per-file failure reasons.
func TestSummaryText(t *testing.T) {
	allOK := Summary{Converted: 3}
	if got := summaryText(allOK); got != "3 converted, 0 failed" {
		t.Errorf("summaryText = %q; want %q", got, "3 converted, 0 failed")
	}

	mixed := Summary{
		Converted: 1,
		Failed: []FileError{
			{Path: "/tmp/bad.xml", Reason: "decoding xml: EOF"},
			{Path: "/tmp/empty.xml", Reason: "file has no messages"},
		},
	}
	got := summaryText(mixed)
	want := "1 converted, 2 failed\nbad.xml: decoding xml: EOF\nempty.xml: file has no messages"
	if got != want {
		t.Errorf("summaryText = %q; want %q", got, want)
	}
}

func mustListable(t *testing.T, dir string) fyne.ListableURI {
	t.Helper()
	l, err := storage.ListerForURI(storage.NewFileURI(dir))
	if err != nil {
		t.Fatalf("lister for %s: %v", dir, err)
	}
	return l
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
