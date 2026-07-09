package ui

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

// seedFailedFile writes raw bytes as a failed input file and returns its path.
func seedFailedFile(t *testing.T, dir, name string, content []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, content, 0o644); err != nil {
		t.Fatalf("seeding %s: %v", name, err)
	}
	return p
}

// readLog reads conversion-errors.log from dir, failing the test if absent.
func readLog(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "conversion-errors.log"))
	if err != nil {
		t.Fatalf("reading log: %v", err)
	}
	return string(data)
}

var testNow = time.Date(2026, 7, 9, 15, 4, 5, 0, time.Local)

// LOG-01: a failed batch writes conversion-errors.log into the output folder,
// replacing any existing file of that name, as UTF-8 text with \n endings.
func TestWriteErrorLogCreatesAndOverwrites(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	failed := seedFailedFile(t, inDir, "bad.xml", []byte("<Log"))

	stale := filepath.Join(outDir, "conversion-errors.log")
	if err := os.WriteFile(stale, []byte("OLD CONTENT"), 0o644); err != nil {
		t.Fatalf("seeding stale log: %v", err)
	}

	s := Summary{Converted: 1, Failed: []FileError{{Path: failed, Reason: "boom"}}}
	path, err := writeErrorLog(outDir, s, testNow, "abc1234")
	if err != nil {
		t.Fatalf("writeErrorLog: %v", err)
	}
	if path != stale {
		t.Errorf("returned path = %q; want %q", path, stale)
	}

	got := readLog(t, outDir)
	if strings.Contains(got, "OLD CONTENT") {
		t.Errorf("log was not overwritten: %q", got)
	}
	if strings.Contains(got, "\r") {
		t.Errorf("log contains \\r; want \\n line endings only")
	}
}

// LOG-02: header carries batch timestamp (2006-01-02 15:04:05 local), version,
// GOOS/GOARCH, and the counts line "N converted, M failed".
func TestWriteErrorLogHeader(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	failed := seedFailedFile(t, inDir, "bad.xml", []byte("<Log"))

	s := Summary{Converted: 1, Failed: []FileError{{Path: failed, Reason: "boom"}}}
	if _, err := writeErrorLog(outDir, s, testNow, "abc1234"); err != nil {
		t.Fatalf("writeErrorLog: %v", err)
	}

	got := readLog(t, outDir)
	for _, want := range []string{
		"2026-07-09 15:04:05",
		"abc1234",
		runtime.GOOS + "/" + runtime.GOARCH,
		"1 converted, 1 failed",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("log missing %q:\n%s", want, got)
		}
	}
}

// LOG-02: version is the short (7-char) vcs.revision from build info, or
// "unknown" when build info or the setting is absent (edge case: plain go build).
func TestVersionFromBuildInfo(t *testing.T) {
	bi := &debug.BuildInfo{Settings: []debug.BuildSetting{
		{Key: "vcs.revision", Value: "a1b2c3d4e5f60718293a4b5c6d7e8f9012345678"},
	}}
	if got := versionFromBuildInfo(bi, true); got != "a1b2c3d" {
		t.Errorf("versionFromBuildInfo = %q; want %q", got, "a1b2c3d")
	}
	if got := versionFromBuildInfo(&debug.BuildInfo{}, true); got != "unknown" {
		t.Errorf("no vcs.revision: got %q; want %q", got, "unknown")
	}
	if got := versionFromBuildInfo(nil, false); got != "unknown" {
		t.Errorf("no build info: got %q; want %q", got, "unknown")
	}
}

// LOG-03: each failed file's entry has full path, error string, size in bytes,
// modification time, and an uppercase space-separated hex dump of the first
// 32 bytes.
func TestWriteErrorLogForensics(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	content := make([]byte, 40)
	for i := range content {
		content[i] = byte(i)
	}
	failed := seedFailedFile(t, inDir, "bad.xml", content)
	mt := time.Date(2026, 7, 8, 10, 30, 0, 0, time.Local)
	if err := os.Chtimes(failed, mt, mt); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	s := Summary{Failed: []FileError{{Path: failed, Reason: "parse error: unexpected EOF"}}}
	if _, err := writeErrorLog(outDir, s, testNow, "abc1234"); err != nil {
		t.Fatalf("writeErrorLog: %v", err)
	}

	got := readLog(t, outDir)
	wantHex := "00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F " +
		"10 11 12 13 14 15 16 17 18 19 1A 1B 1C 1D 1E 1F"
	for _, want := range []string{
		failed, // full input path
		"parse error: unexpected EOF",
		"40 bytes",
		"2026-07-08 10:30:00",
		wantHex,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("log missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, wantHex+" 20") {
		t.Errorf("hex dump exceeds 32 bytes:\n%s", got)
	}
}

// Edge case: a failed file shorter than 32 bytes shows exactly the bytes present.
func TestWriteErrorLogShortFile(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	failed := seedFailedFile(t, inDir, "tiny.xml", []byte{0x61, 0x62, 0xFF})

	s := Summary{Failed: []FileError{{Path: failed, Reason: "boom"}}}
	if _, err := writeErrorLog(outDir, s, testNow, "abc1234"); err != nil {
		t.Fatalf("writeErrorLog: %v", err)
	}

	got := readLog(t, outDir)
	if !strings.Contains(got, "61 62 FF\n") {
		t.Errorf("log missing exact short hex dump %q:\n%s", "61 62 FF", got)
	}
}

// LOG-04 + edge case: a failed file deleted before log writing gets its
// collection error recorded while other entries stay intact and the log is
// still written.
func TestWriteErrorLogMissingFile(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	real := seedFailedFile(t, inDir, "bad.xml", []byte{0x61, 0x62})
	missing := filepath.Join(inDir, "gone.xml")

	s := Summary{Failed: []FileError{
		{Path: missing, Reason: "boom1"},
		{Path: real, Reason: "boom2"},
	}}
	if _, err := writeErrorLog(outDir, s, testNow, "abc1234"); err != nil {
		t.Fatalf("writeErrorLog: %v", err)
	}

	got := readLog(t, outDir)
	if !strings.Contains(got, missing) || !strings.Contains(got, "boom1") {
		t.Errorf("missing file's entry absent:\n%s", got)
	}
	// The collection error for the unreadable file is recorded.
	if !strings.Contains(got, "unavailable") {
		t.Errorf("log does not record collection error for missing file:\n%s", got)
	}
	// The intact file's forensics are unaffected.
	for _, want := range []string{real, "boom2", "2 bytes", "61 62"} {
		if !strings.Contains(got, want) {
			t.Errorf("intact entry missing %q:\n%s", want, got)
		}
	}
}

// LOG-01 (gating) via handleErrorLog: a batch with ≥1 failure writes the log
// and returns its full path.
func TestHandleErrorLogWritesOnFailure(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	failed := seedFailedFile(t, inDir, "bad.xml", []byte("<Log"))

	s := Summary{Converted: 0, Failed: []FileError{{Path: failed, Reason: "boom"}}}
	path, err := handleErrorLog(outDir, s)
	if err != nil {
		t.Fatalf("handleErrorLog: %v", err)
	}
	want := filepath.Join(outDir, "conversion-errors.log")
	if path != want {
		t.Errorf("path = %q; want %q", path, want)
	}
	if _, err := os.Stat(want); err != nil {
		t.Errorf("log not written: %v", err)
	}
}

// LOG-05: a batch with 0 failures writes no log.
func TestHandleErrorLogNoFailuresWritesNothing(t *testing.T) {
	outDir := t.TempDir()

	path, err := handleErrorLog(outDir, Summary{Converted: 3})
	if path != "" || err != nil {
		t.Errorf("handleErrorLog = (%q, %v); want (\"\", nil)", path, err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "conversion-errors.log")); !os.IsNotExist(err) {
		t.Errorf("log exists after all-success batch; stat err = %v", err)
	}
}

// LOG-09: a fully successful batch deletes a leftover conversion-errors.log.
func TestHandleErrorLogDeletesStaleOnSuccess(t *testing.T) {
	outDir := t.TempDir()
	stale := filepath.Join(outDir, "conversion-errors.log")
	if err := os.WriteFile(stale, []byte("stale"), 0o644); err != nil {
		t.Fatalf("seeding stale log: %v", err)
	}

	path, err := handleErrorLog(outDir, Summary{Converted: 2})
	if path != "" || err != nil {
		t.Errorf("handleErrorLog = (%q, %v); want (\"\", nil)", path, err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("stale log still present; stat err = %v", err)
	}
}

// LOG-10: deletion target absent → proceed silently (no error surfaced).
func TestHandleErrorLogMissingStaleIsSilent(t *testing.T) {
	outDir := t.TempDir()

	path, err := handleErrorLog(outDir, Summary{Converted: 1})
	if path != "" || err != nil {
		t.Errorf("handleErrorLog = (%q, %v); want (\"\", nil)", path, err)
	}
}

// LOG-07 (data): an unwritable output folder surfaces the write error to the
// caller instead of crashing.
func TestHandleErrorLogWriteFailure(t *testing.T) {
	inDir := t.TempDir()
	notADir := seedFailedFile(t, inDir, "plainfile", []byte("x"))
	outDir := filepath.Join(notADir, "sub") // child of a regular file: unwritable

	s := Summary{Failed: []FileError{{Path: notADir, Reason: "boom"}}}
	path, err := handleErrorLog(outDir, s)
	if err == nil {
		t.Fatalf("handleErrorLog succeeded writing under a regular file; path = %q", path)
	}
	if path != "" {
		t.Errorf("path = %q on failure; want \"\"", path)
	}
}

// Edge case: every file failed — all failures listed, counts read "0 converted,
// M failed".
func TestWriteErrorLogAllFailed(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	f1 := seedFailedFile(t, inDir, "a.xml", []byte("x"))
	f2 := seedFailedFile(t, inDir, "b.xml", []byte("y"))

	s := Summary{Converted: 0, Failed: []FileError{
		{Path: f1, Reason: "r1"},
		{Path: f2, Reason: "r2"},
	}}
	if _, err := writeErrorLog(outDir, s, testNow, "abc1234"); err != nil {
		t.Fatalf("writeErrorLog: %v", err)
	}

	got := readLog(t, outDir)
	if !strings.Contains(got, "0 converted, 2 failed") {
		t.Errorf("counts line wrong:\n%s", got)
	}
	for _, want := range []string{f1, "r1", f2, "r2"} {
		if !strings.Contains(got, want) {
			t.Errorf("log missing %q:\n%s", want, got)
		}
	}
}
