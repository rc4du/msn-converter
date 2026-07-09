package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

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
