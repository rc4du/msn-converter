package ui

import (
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
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
	return newAppUI(a).win
}

// appUI holds the window state and widgets. Everything below it is Fyne-free;
// this is the single wiring point.
type appUI struct {
	win     fyne.Window
	queue   *Queue
	outDir  string
	running bool

	list         *widget.List
	outLabel     *widget.Label
	progress     *widget.ProgressBar
	convertBtn   *widget.Button
	addFilesBtn  *widget.Button
	addFolderBtn *widget.Button
	clearBtn     *widget.Button
	chooseOutBtn *widget.Button
}

func newAppUI(a fyne.App) *appUI {
	u := &appUI{
		win:   a.NewWindow("MSN Converter"),
		queue: &Queue{},
	}
	u.win.Resize(fyne.NewSize(600, 450))

	u.list = widget.NewList(
		func() int { return u.queue.Len() },
		func() fyne.CanvasObject {
			return container.NewBorder(nil, nil, nil, widget.NewButton("✕", nil), widget.NewLabel(""))
		},
		func(id widget.ListItemID, o fyne.CanvasObject) {
			items := u.queue.Items()
			if id < 0 || id >= len(items) {
				return
			}
			path := items[id]
			row := o.(*fyne.Container)
			row.Objects[0].(*widget.Label).SetText(filepath.Base(path))
			row.Objects[1].(*widget.Button).OnTapped = func() {
				u.queue.Remove(path)
				u.refresh()
			}
		},
	)

	u.addFilesBtn = widget.NewButton("Add files", nil)
	u.addFolderBtn = widget.NewButton("Add folder", nil)
	u.clearBtn = widget.NewButton("Clear all", func() {
		u.queue.Clear()
		u.refresh()
	})

	u.chooseOutBtn = widget.NewButton("Choose output folder…", nil)
	u.outLabel = widget.NewLabel("No output folder selected")
	u.progress = widget.NewProgressBar()
	u.convertBtn = widget.NewButton("Convert", nil)

	top := container.NewHBox(u.addFilesBtn, u.addFolderBtn, u.clearBtn)
	bottom := container.NewVBox(
		container.NewBorder(nil, nil, u.chooseOutBtn, nil, u.outLabel),
		u.progress,
		u.convertBtn,
	)
	u.win.SetContent(container.NewBorder(top, bottom, nil, nil, u.list))

	u.updateConvertState()
	return u
}

// refresh redraws the queue list and recomputes Convert enablement. Call after
// every queue mutation.
func (u *appUI) refresh() {
	u.list.Refresh()
	u.updateConvertState()
}

// updateConvertState applies GUI-10: Convert is enabled only when the queue is
// non-empty, an output folder is chosen, and no batch is running.
func (u *appUI) updateConvertState() {
	if u.queue.Len() > 0 && u.outDir != "" && !u.running {
		u.convertBtn.Enable()
	} else {
		u.convertBtn.Disable()
	}
}
