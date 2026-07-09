package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

// Run builds the main window and enters the Fyne event loop.
func Run() {
	a := app.New()
	newMainWindow(a).ShowAndRun()
}

// newMainWindow assembles the application window. Kept separate from Run so
// tests can build it against a headless app.
func newMainWindow(a fyne.App) fyne.Window {
	w := a.NewWindow("MSN Converter")
	w.Resize(fyne.NewSize(600, 450))
	w.SetContent(widget.NewLabel("MSN Converter"))
	return w
}
