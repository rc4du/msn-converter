package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// writeXML writes a minimal valid one-message log and returns its path.
// date/time/receiver drive the derived output filename.
func writeXML(t *testing.T, dir, name, date, timeAttr, receiver, text string) string {
	t.Helper()
	content := fmt.Sprintf(`<?xml version="1.0"?>
<Log FirstSessionID="1" LastSessionID="1">
<Message Date=%q Time=%q DateTime="2010-09-17T01:45:43.093Z" SessionID="1">
<From><User FriendlyName="Ana"/></From><To><User FriendlyName=%q/></To>
<Text Style="">%s</Text>
</Message>
</Log>`, date, timeAttr, receiver, text)
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("seeding %s: %v", name, err)
	}
	return p
}

// GUI-12 + GUI-11 (data): mixed batch — failures are skipped and reported with
// reasons, the batch completes, and counts are exact.
func TestRunBatchMixed(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	valid := writeXML(t, inDir, "valid.xml", "16/9/2010", "22:45:43", "Ricardo", "oi")

	malformed := filepath.Join(inDir, "malformed.xml")
	if err := os.WriteFile(malformed, []byte("<Log><Message Date="), 0o644); err != nil {
		t.Fatalf("seeding malformed.xml: %v", err)
	}

	zeroMsg := filepath.Join(inDir, "zero.xml")
	if err := os.WriteFile(zeroMsg, []byte(`<?xml version="1.0"?><Log FirstSessionID="1" LastSessionID="1"></Log>`), 0o644); err != nil {
		t.Fatalf("seeding zero.xml: %v", err)
	}

	missing := filepath.Join(inDir, "does-not-exist.xml")

	files := []string{valid, malformed, zeroMsg, missing}
	s := RunBatch(files, outDir, func(done, total int) {})

	if s.Converted != 1 {
		t.Errorf("Converted = %d; want 1", s.Converted)
	}
	if len(s.Failed) != 3 {
		t.Fatalf("len(Failed) = %d; want 3 (%v)", len(s.Failed), s.Failed)
	}
	wantFailed := []string{malformed, zeroMsg, missing}
	for i, fe := range s.Failed {
		if fe.Path != wantFailed[i] {
			t.Errorf("Failed[%d].Path = %q; want %q", i, fe.Path, wantFailed[i])
		}
		if fe.Reason == "" {
			t.Errorf("Failed[%d].Reason is empty; want non-empty reason", i)
		}
	}

	// The valid file's output landed in outDir.
	if _, err := os.Stat(filepath.Join(outDir, "16_9_2010_22-45-43_Ricardo.txt")); err != nil {
		t.Errorf("expected output file missing: %v", err)
	}
}

// GUI-11 (logic): progress fires exactly once per file with (1..n, n).
func TestRunBatchProgressPerFile(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	files := []string{
		writeXML(t, inDir, "a.xml", "16/9/2010", "22:45:43", "Ricardo", "oi"),
		filepath.Join(inDir, "missing.xml"), // failures count as processed too
		writeXML(t, inDir, "b.xml", "17/9/2010", "10:00:00", "Ana", "oi"),
	}

	var calls [][2]int
	RunBatch(files, outDir, func(done, total int) {
		calls = append(calls, [2]int{done, total})
	})

	if len(calls) != len(files) {
		t.Fatalf("progress called %d times; want %d", len(calls), len(files))
	}
	for i, c := range calls {
		if c[0] != i+1 || c[1] != len(files) {
			t.Errorf("progress call %d = (%d, %d); want (%d, %d)", i, c[0], c[1], i+1, len(files))
		}
	}
}

// GUI-14: two inputs deriving the same output name → one file, last write wins.
func TestRunBatchNameClashLastWriteWins(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	first := writeXML(t, inDir, "first.xml", "16/9/2010", "22:45:43", "Ricardo", "FIRST")
	second := writeXML(t, inDir, "second.xml", "16/9/2010", "22:45:43", "Ricardo", "SECOND")

	s := RunBatch([]string{first, second}, outDir, func(done, total int) {})
	if s.Converted != 2 || len(s.Failed) != 0 {
		t.Fatalf("Summary = %+v; want 2 converted, 0 failed", s)
	}

	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("reading outDir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("outDir has %d files; want 1", len(entries))
	}

	got, err := os.ReadFile(filepath.Join(outDir, "16_9_2010_22-45-43_Ricardo.txt"))
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	want := "16/9/2010 - Ana\n22:45:43 - SECOND\n"
	if string(got) != want {
		t.Errorf("output = %q; want %q (last write wins)", got, want)
	}
}
