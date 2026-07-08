package converter

import (
	"bytes"
	_ "embed"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"text/template"
)

//go:embed output.txt.tmpl
var templateText string

var outputTmpl = template.Must(template.New("output").Parse(templateText))

// ErrNoMessages reports a log that decoded successfully but contains no messages.
var ErrNoMessages = errors.New("file has no messages")

type Result struct {
	FileName string
	Content  []byte
}

// Convert decodes one MSN XML log and renders it as plain text.
func Convert(r io.Reader) (Result, error) {
	var log Log
	if err := xml.NewDecoder(r).Decode(&log); err != nil {
		return Result{}, fmt.Errorf("decoding xml: %w", err)
	}
	if len(log.Messages) == 0 {
		return Result{}, ErrNoMessages
	}
	var buf bytes.Buffer
	if err := outputTmpl.Execute(&buf, log); err != nil {
		return Result{}, fmt.Errorf("rendering template: %w", err)
	}
	return Result{Content: buf.Bytes()}, nil
}
