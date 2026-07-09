package ui

import (
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
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
	win       fyne.Window
	queue     *Queue
	outDir    string
	running   bool
	xmlFilter storage.FileFilter

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

	u.xmlFilter = storage.NewExtensionFileFilter([]string{".xml"})
	u.addFilesBtn = widget.NewButton("Add files", func() {
		d := dialog.NewFileOpen(u.onFilePicked, u.win)
		d.SetFilter(u.xmlFilter)
		d.Show()
	})
	u.addFolderBtn = widget.NewButton("Add folder", func() {
		dialog.NewFolderOpen(u.onFolderPicked, u.win).Show()
	})
	u.clearBtn = widget.NewButton("Clear all", func() {
		u.queue.Clear()
		u.refresh()
	})

	u.chooseOutBtn = widget.NewButton("Choose output folder…", func() {
		dialog.NewFolderOpen(u.onOutputPicked, u.win).Show()
	})
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
	u.win.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		paths := make([]string, 0, len(uris))
		for _, uri := range uris {
			paths = append(paths, uri.Path())
		}
		u.addDropped(paths)
	})

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

// addFiles appends paths to the queue (de-duplicated) and refreshes the UI.
func (u *appUI) addFiles(paths []string) {
	u.queue.Add(paths...)
	u.refresh()
}

// setOutputDir records the chosen output folder, shows it, and re-gates Convert.
func (u *appUI) setOutputDir(path string) {
	u.outDir = path
	u.outLabel.SetText(path)
	u.updateConvertState()
}

// addDropped classifies dropped paths (GUI-15): folders contribute their
// direct .xml files, .xml files (case-insensitive) are added, anything else
// (including nonexistent paths) is ignored.
func (u *appUI) addDropped(paths []string) {
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		switch {
		case info.IsDir():
			xmls, err := ListXML(p)
			if err != nil {
				continue
			}
			files = append(files, xmls...)
		case strings.EqualFold(filepath.Ext(p), ".xml"):
			files = append(files, p)
		}
	}
	u.addFiles(files)
}

// onFilePicked handles the add-files dialog result. Cancel (nil URI) is a no-op.
func (u *appUI) onFilePicked(rc fyne.URIReadCloser, err error) {
	if err != nil || rc == nil {
		return
	}
	defer rc.Close()
	u.addFiles([]string{rc.URI().Path()})
}

// onFolderPicked adds every .xml directly inside the picked folder. Cancel is a no-op.
func (u *appUI) onFolderPicked(dir fyne.ListableURI, err error) {
	if err != nil || dir == nil {
		return
	}
	files, err := ListXML(dir.Path())
	if err != nil {
		return
	}
	u.addFiles(files)
}

// onOutputPicked records the picked output folder. Cancel is a no-op.
func (u *appUI) onOutputPicked(dir fyne.ListableURI, err error) {
	if err != nil || dir == nil {
		return
	}
	u.setOutputDir(dir.Path())
}
