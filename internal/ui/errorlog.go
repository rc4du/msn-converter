package ui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

// errorLogName is fixed (spec: overwrite, no rotation) so the user instruction
// stays one sentence.
const errorLogName = "conversion-errors.log"

// logTimeLayout is the spec-mandated local-time format for the batch header;
// modification times reuse it for consistency.
const logTimeLayout = "2006-01-02 15:04:05"

// previewLen bounds the hex dump: enough for BOM/XML header forensics without
// embedding chat content (privacy).
const previewLen = 32

// appVersion identifies the running build for the log header.
func appVersion() string {
	bi, ok := debug.ReadBuildInfo()
	return versionFromBuildInfo(bi, ok)
}

// versionFromBuildInfo extracts the short (7-char) vcs.revision, or "unknown"
// when the binary was built without VCS stamping.
func versionFromBuildInfo(bi *debug.BuildInfo, ok bool) string {
	if !ok || bi == nil {
		return "unknown"
	}
	for _, s := range bi.Settings {
		if s.Key == "vcs.revision" && s.Value != "" {
			if len(s.Value) > 7 {
				return s.Value[:7]
			}
			return s.Value
		}
	}
	return "unknown"
}

// handleErrorLog is the post-batch entry point: with failures it writes the
// log and returns its path; with none it removes any stale log (LOG-09/10,
// errors deliberately ignored) so users never email an outdated file.
func handleErrorLog(outDir string, s Summary) (string, error) {
	if len(s.Failed) == 0 {
		os.Remove(filepath.Join(outDir, errorLogName))
		return "", nil
	}
	return writeErrorLog(outDir, s, time.Now(), appVersion())
}

// writeErrorLog writes the forensic log for a failed batch into outDir,
// replacing any previous log. It returns the full path written.
func writeErrorLog(outDir string, s Summary, now time.Time, version string) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "MSN Converter — conversion error log\n")
	fmt.Fprintf(&b, "Batch:    %s\n", now.Format(logTimeLayout))
	fmt.Fprintf(&b, "Version:  %s\n", version)
	fmt.Fprintf(&b, "System:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&b, "%d converted, %d failed\n", s.Converted, len(s.Failed))
	for _, fe := range s.Failed {
		b.WriteString("\n")
		fmt.Fprintf(&b, "File:     %s\n", fe.Path)
		fmt.Fprintf(&b, "Error:    %s\n", fe.Reason)
		writeForensics(&b, fe.Path)
	}
	path := filepath.Join(outDir, errorLogName)
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// writeForensics appends size, modification time and hex preview for one
// failed file. Collection errors are recorded per field so one unreadable
// file never suppresses the rest of the log (LOG-04).
func writeForensics(b *strings.Builder, path string) {
	if info, err := os.Stat(path); err != nil {
		fmt.Fprintf(b, "Size:     unavailable: %v\n", err)
		fmt.Fprintf(b, "Modified: unavailable: %v\n", err)
	} else {
		fmt.Fprintf(b, "Size:     %d bytes\n", info.Size())
		fmt.Fprintf(b, "Modified: %s\n", info.ModTime().Format(logTimeLayout))
	}
	fmt.Fprintf(b, "First bytes: %s\n", hexPreview(path))
}

// hexPreview returns an uppercase space-separated hex dump of the file's
// first previewLen bytes (fewer when the file is shorter), or a collection
// error note when the file cannot be read.
func hexPreview(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Sprintf("unavailable: %v", err)
	}
	defer f.Close()

	buf := make([]byte, previewLen)
	n, err := io.ReadFull(f, buf)
	if n == 0 && err != nil && err != io.EOF {
		return fmt.Sprintf("unavailable: %v", err)
	}
	parts := make([]string, n)
	for i := 0; i < n; i++ {
		parts[i] = fmt.Sprintf("%02X", buf[i])
	}
	return strings.Join(parts, " ")
}
