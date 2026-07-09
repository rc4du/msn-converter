# GUI Specification — Fyne desktop app for msn-converter

**Source plan:** `GUI_PLAN.md` (design decisions pre-agreed there are treated as confirmed requirements)

## Problem Statement

The current CLI converts exactly one hardcoded XML file per run and panics on any error. Users with hundreds of MSN Messenger logs need bulk conversion with file/folder selection, and the tool must run as a self-contained Windows desktop app.

## Goals

- [x] Bulk-convert MSN Messenger XML logs to human-readable `.txt` via a Fyne GUI
- [x] Single self-contained binary (template embedded, no runtime file dependencies)
- [x] Robust batch behavior: malformed/empty files skipped and reported, never a crash

## Out of Scope

| Feature | Reason |
| ------- | ------ |
| Windows `.exe` packaging (fyne-cross / CI) | Deferred in GUI_PLAN.md — separate step |
| Recursive folder scanning | Deferred in GUI_PLAN.md |
| CRLF line endings / user-editable output format | Deferred in GUI_PLAN.md |
| Cancel button for a running batch | Not in plan; batch is skip-and-report and finishes on its own |
| CLI mode | Plan explicitly drops old CLI behavior |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --------------------- | -------------- | --------- | ---------- |
| Filename sanitization | Date `/`→`_` (existing logic), then any remaining Windows-illegal char (`\ / : * ? " < > \|`) in the base name → `-` | Time `22:45:43` and arbitrary FriendlyNames must be Windows-legal | y (user) |
| Template engine | `text/template` (replaces `html/template`) | Output is `.txt`; HTML entity escaping (`&amp;`, `&#39;`) mangles chat text | y (user) |
| File list after batch | Kept as-is; user clears manually | No surprise state change; summary dialog already signals completion | y (user) |
| Add-files picker filter | Native dialog filtered to `.xml` | Matches folder-picker behavior; prevents junk input | assumption |
| Convert button during a running batch | Disabled (also disabled Add/Remove/Clear mutations are allowed but batch snapshot is taken at start) | Prevents concurrent batches (concurrency guard) | assumption |
| Folder pick / drop containing zero `.xml` | Adds nothing, no error dialog | Low-stakes no-op; list simply doesn't change | assumption |
| Progress granularity | Progress bar advances once per file processed (converted or failed) | Simple, accurate for a sequential batch | assumption |
| Repo restructure | `models/` folds into `internal/converter`; `files/` scaffolding deleted; `files/input/example.xml` moves to `testdata/example.xml` | Mandated by AGENTS.md project structure | y (AGENTS.md) |
| Fyne version | Latest Fyne v2.x | GUI_PLAN.md assumed default | y (plan) |

**Open questions:** none — all resolved or logged above.

---

## User Stories

### P1: Core conversion library ⭐ MVP

**User Story**: As the app, I need a `converter` package that turns one MSN XML log into `.txt` content so the GUI can batch it safely.

**Why P1**: Everything else layers on this; also fixes the latent `Messages[0]` panic.

**Acceptance Criteria**:

1. WHEN a valid MSN XML log is converted THEN the converter SHALL return the templated text (grouped by sender, `{time} - {text}` lines, LF endings) with message text verbatim (no HTML escaping)
2. WHEN the XML is malformed or unreadable THEN the converter SHALL return an error (never panic)
3. WHEN the log decodes but contains zero `<Message>` elements THEN the converter SHALL return an error stating the file has no messages
4. WHEN an output filename is derived THEN it SHALL be `{date}_{time}_{receiver}.txt` from the first message, with date `/`→`_` and any remaining Windows-illegal char (`\ / : * ? " < > |`) replaced by `-` (e.g. `16_9_2010_22-45-43_Ricardo.txt`)
5. WHEN the binary is built THEN the output template SHALL be embedded via `go:embed` (no runtime template file read)

**Independent Test**: Unit tests convert `testdata/example.xml` and malformed/empty fixtures; assert exact output text, filenames, and error cases.

---

### P1: Input queue management ⭐ MVP

**User Story**: As a user, I want to build a list of XML files from pickers so I can choose exactly what gets converted.

**Why P1**: No conversion without input selection.

**Acceptance Criteria**:

1. WHEN the user clicks "Add files" THEN the app SHALL open a native file dialog filtered to `.xml` and append the chosen file to the list
2. WHEN the user picks a folder THEN the app SHALL append every `.xml` file directly inside it (non-recursive)
3. WHEN a file already in the list is added again (any source) THEN the list SHALL remain unchanged (de-dup by absolute path)
4. WHEN the user clicks a list item's remove control THEN that file SHALL leave the list
5. WHEN the user clicks clear-all THEN the list SHALL become empty

**Independent Test**: Queue-state unit tests (add/add-dup/remove/clear); manual demo adding files and a folder.

---

### P1: Batch conversion & reporting ⭐ MVP

**User Story**: As a user, I want to convert the whole queue into a chosen output folder and see what succeeded so I can trust the result.

**Why P1**: This is the product.

**Acceptance Criteria**:

1. WHEN no output folder is set OR the file list is empty THEN the Convert button SHALL be disabled
2. WHEN both an output folder is set AND the list is non-empty THEN the Convert button SHALL be enabled
3. WHEN Convert is clicked THEN conversion SHALL run off the UI goroutine, the Convert button SHALL disable for the duration, and a progress bar SHALL advance once per file processed
4. WHEN a file fails (malformed, empty, unreadable, write error) THEN the batch SHALL skip it and continue with the remaining files
5. WHEN the batch finishes THEN the app SHALL show a summary dialog: `N converted, M failed` plus per-file failure reasons, and the file list SHALL remain unchanged
6. WHEN an output filename already exists (pre-existing file or two inputs mapping to the same name) THEN the app SHALL overwrite it (last write wins)

**Independent Test**: Run app with a mixed queue (valid + malformed + empty XML) → summary shows correct counts; output folder contains expected `.txt` files.

---

### P2: Drag and drop input

**User Story**: As a user, I want to drop files or folders onto the window so building the queue is fast.

**Why P2**: Convenience — pickers already cover input fully.

**Acceptance Criteria**:

1. WHEN `.xml` files are dropped onto the window THEN they SHALL be appended to the list (deduped)
2. WHEN a folder is dropped THEN every `.xml` directly inside it SHALL be appended (non-recursive)
3. WHEN non-`.xml` files are dropped THEN they SHALL be ignored

**Independent Test**: Drop a mixed selection (xml, txt, folder) → only xml + folder's xml appear.

---

## Edge Cases

- WHEN a 0-byte or truncated XML file is in the batch THEN it SHALL be skipped and reported (no panic — fixes latent `Messages[0]` crash)
- WHEN two queued inputs derive the same output filename THEN the later one SHALL overwrite (last write wins)
- WHEN the output folder becomes unwritable (deleted after selection) THEN each file SHALL fail with a reported reason; batch completes
- WHEN a picked folder contains zero `.xml` THEN the list SHALL be unchanged and no error shown
- WHEN a `.XML` (uppercase) file is picked from a folder THEN it SHALL be treated as `.xml` (case-insensitive extension match)

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| -------------- | ----- | ----- | ------ |
| GUI-01 | P1 Converter: valid log → templated text, verbatim, LF | Execute | Implemented ✅ |
| GUI-02 | P1 Converter: malformed → error, no panic | Execute | Implemented ✅ |
| GUI-03 | P1 Converter: zero messages → error | Execute | Implemented ✅ |
| GUI-04 | P1 Converter: filename derivation + Windows sanitization | Execute | Implemented ✅ |
| GUI-05 | P1 Converter: template embedded via go:embed | Execute | Implemented ✅ |
| GUI-06 | P1 Queue: add-files dialog (.xml filter) | Execute | Implemented ✅ |
| GUI-07 | P1 Queue: folder picker adds *.xml non-recursive | Execute | Implemented ✅ |
| GUI-08 | P1 Queue: de-dup by absolute path | Execute | Implemented ✅ |
| GUI-09 | P1 Queue: per-item remove + clear-all | Execute | Implemented ✅ |
| GUI-10 | P1 Batch: Convert enablement rule | Execute | Implemented ✅ |
| GUI-11 | P1 Batch: off-UI-goroutine + progress per file | Execute | Implemented ✅ |
| GUI-12 | P1 Batch: skip-and-report, batch continues | Execute | Implemented ✅ |
| GUI-13 | P1 Batch: summary dialog N/M + reasons | Execute | Implemented ✅ |
| GUI-14 | P1 Batch: overwrite on name clash | Execute | Implemented ✅ |
| GUI-15 | P2 DnD: files/folders dropped → .xml appended | Execute | Implemented ✅ |
| GUI-16 | App shell: window ~600×450, title "MSN Converter" | Execute | Implemented ✅ |

**Coverage:** 16 total, 16 implemented and verified (validation.md PASS, 2026-07-08); native-dialog/window smoke items remain manual

---

## Success Criteria

- [x] A mixed batch (valid, malformed, empty files) completes with correct `N converted, M failed` and no crash
- [x] Generated filenames are Windows-legal for the sample fixtures
- [x] `go build` produces a binary that converts without `templates/` present on disk
- [ ] Window remains responsive (movable, repainting) during a large batch — manual smoke outstanding
