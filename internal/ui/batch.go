package ui

import (
	"github.com/rrc4du/msn-converter/internal/converter"
)

// FileError records why one input file failed to convert.
type FileError struct {
	Path   string
	Reason string
}

// Summary is the outcome of a batch run.
type Summary struct {
	Converted int
	Failed    []FileError
}

// RunBatch converts each file sequentially into outDir. Failures are recorded
// and skipped — the batch always completes. progress is called once per file
// processed (converted or failed) with (done, total).
func RunBatch(files []string, outDir string, progress func(done, total int)) Summary {
	var s Summary
	total := len(files)
	for i, f := range files {
		if _, err := converter.ConvertFile(f, outDir); err != nil {
			s.Failed = append(s.Failed, FileError{Path: f, Reason: err.Error()})
		} else {
			s.Converted++
		}
		progress(i+1, total)
	}
	return s
}
