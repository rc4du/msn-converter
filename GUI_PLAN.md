# GUI Plan — Fyne desktop app for msn-converter

Design decisions for wrapping the MSN XML→text converter in a [Fyne](https://github.com/fyne-io/fyne) GUI. Primary target: Windows.

## Goal

Replace the current hardcoded single-file CLI with a Fyne desktop app that bulk-converts MSN Messenger XML chat logs into human-readable `.txt` files.

## Architecture

- Extract the conversion logic (decode XML → run template → write `.txt`) into a `converter` package.
  - Returns errors instead of panicking.
- `main.go` becomes the Fyne application entry point. The old CLI behavior is dropped.
- The output template is embedded into the binary via `go:embed`, producing a single self-contained `.exe` with no runtime file dependency.

## Input selection

A single de-duplicated file list (keyed by absolute path) fed by three sources:

1. **Add files** button — opens the native Fyne file picker. Fyne has no multi-select, so this adds one file per open; selections accumulate in the list.
2. **Folder picker** — adds every `.xml` in the chosen folder (non-recursive).
3. **Drag and drop** — dropping files or folders onto the window adds their `.xml` contents.

The list widget shows queued files with per-item remove and a clear-all action.

## Output

- User selects one output folder.
- Filename keeps existing logic: `{date}_{time}_{receiver}.txt` (derived from the first message).
- **Overwrite** on name clash (existing file or two inputs producing the same name — last write wins).
- **LF** (`\n`) line endings (unchanged from current template).

## Batch execution

- Conversion runs off the UI goroutine so the window stays responsive; a progress bar reflects progress.
- **Skip-and-report** error handling: malformed or empty files are skipped (this also fixes a latent `Messages[0]` panic on empty logs). A summary at the end reports `N converted, M failed` with per-file reasons.

## Windows support

- Development on macOS via `go run`. Windows packaging (the CGO `.exe` build) is deferred and handled as a separate step.
- Code kept Windows-safe regardless:
  - `filepath.Join` for all path building (current code uses string concatenation).
  - No console/stdout assumptions.

## Assumed defaults

- Convert button disabled until the file list is non-empty **and** an output folder is set.
- Progress and summary surfaced via Fyne dialogs.
- Latest Fyne v2.x.
- Window ~600×450, title "MSN Converter".

## Deferred / out of scope

- Windows `.exe` packaging strategy (`fyne-cross` via Docker, build-on-Windows, or GitHub Actions CI) — decide later.
- Recursive folder scanning.
- CRLF line endings and user-editable output format.
