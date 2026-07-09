# GUI Tasks

## Execution Protocol (MANDATORY -- do not skip)

Implement these tasks with the `tlc-spec-driven` skill: **activate it by name and follow its Execute flow and Critical Rules.** Do not search for skill files by filesystem path. The skill is the source of truth for the full flow (per-task cycle, sub-agent delegation, adequacy review, Verifier, discrimination sensor).

**If the skill cannot be activated, STOP and tell the user — do not proceed without it.**

---

**Design**: `.specs/features/gui/design.md`
**Status**: Draft

---

## Test Coverage Matrix

> Generated from codebase, project guidelines, and spec — confirm before Execute. Guidelines found: `AGENTS.md` (Goal-Driven Execution: tests define done; prescribes `converter_test.go` + `testdata/`). No existing tests in repo — strong defaults applied on top.

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| ---------- | ------------------ | -------------------- | ---------------- | ----------- |
| converter (domain) | unit | All branches; 1:1 to spec ACs GUI-01..05, GUI-14; every listed edge case (0-byte, zero messages, illegal chars) | `internal/converter/*_test.go` | `go test ./internal/converter/` |
| queue / scan / batch (domain) | unit | All branches; 1:1 to GUI-07/08/09-logic/11/12/13/14; edge cases (dup add, empty dir, case-insensitive ext, mixed batch) | `internal/ui/*_test.go` | `go test ./internal/ui/` |
| app.go (Fyne wiring) | unit (headless via `fyne.io/fyne/v2/test`) | Enablement matrix (GUI-10), queue mutations via taps, drop classification (GUI-15), batch progress/summary/state (GUI-11/13/18); native OS dialogs excluded → manual smoke at validation | `internal/ui/*_test.go` | `go test ./internal/ui/` |
| main.go / model.go / template asset | none | — build gate only | — | `go build ./...` |

## Gate Check Commands

> Generated from codebase — confirm before Execute.

| Gate Level | When to Use | Command |
| ---------- | ----------- | ------- |
| Quick | After tasks with unit tests only | `go test ./...` |
| Full | After wiring tasks | `go vet ./... && go test ./...` |
| Build | After phase completion / entity-only tasks | `go build ./... && go vet ./... && go test ./...` |

---

## Execution Plan

### Phase 1: Converter core (pure Go)

```
T1 → T2 → T3 → T4
```

### Phase 2: UI logic (pure Go)

```
T5 → T6 → T7
```

### Phase 3: Fyne shell & wiring

```
T8 → T9 → T10 → T11 → T12
```

---

## Task Breakdown

### T1: Move XML structs into internal/converter

**What**: Create `internal/converter/model.go` with the structs from `models/input_message.go` (verbatim, package `converter`); copy `files/input/example.xml` → `testdata/example.xml`. Old `models/` and `files/` stay until T8 (main.go still imports them).
**Where**: `internal/converter/model.go`, `testdata/example.xml`
**Depends on**: None
**Reuses**: `models/input_message.go` (verbatim move)
**Requirement**: supports GUI-01..05

**Tools**: none (stdlib project)

**Done when**:

- [x] Structs compile under package `converter`, fields unchanged
- [x] `testdata/example.xml` exists and `git status` shows it as trackable (not ignored)
- [x] Gate: `go build ./...`

**Tests**: none (schema move — build gate only)
**Gate**: build
**Commit**: `refactor(converter): move XML structs into internal/converter`

---

### T2: Convert() with embedded text template

**What**: `internal/converter/converter.go`: `//go:embed output.txt.tmpl` (content moved from `templates/output.txt`, unchanged), parsed with `text/template`; `Convert(r io.Reader) (Result, error)` decodes XML, returns `ErrNoMessages` for zero-message logs, renders template. `Result{FileName, Content}` (FileName filled by T3 — stub empty here or land T3's func signature as unexported placeholder returning "").
**Where**: `internal/converter/converter.go`, `internal/converter/output.txt.tmpl`, `internal/converter/converter_test.go`
**Depends on**: T1
**Reuses**: `templates/output.txt` content; `main.go` decode logic
**Requirement**: GUI-01, GUI-02, GUI-03, GUI-05

**Tools**: none

**Done when**:

- [x] Valid `testdata/example.xml` → exact expected text (sender grouping, `{time} - {text}` lines, LF, verbatim text — assert a message containing `&`/`<` stays unescaped via inline fixture)
- [x] Malformed XML → error, no panic; 0-byte input → error, no panic
- [x] Zero-message log → `ErrNoMessages` (via `errors.Is`)
- [x] Binary contains template (no file read at runtime)
- [x] Gate: `go test ./internal/converter/` passes

**Tests**: unit (GUI-01/02/03/05 1:1)
**Gate**: quick
**Commit**: `feat(converter): add Convert with embedded text template`

---

### T3: Windows-safe output filename derivation

**What**: Filename rule per AD-001 in `converter.go`: `{date /→_}_{time}_{receiver}` + sanitize `\ / : * ? " < > |` → `-` + `.txt`, from first message; wire into `Convert`'s `Result.FileName`.
**Where**: `internal/converter/converter.go`, `internal/converter/converter_test.go`
**Depends on**: T2
**Reuses**: `main.go:50` date logic
**Requirement**: GUI-04

**Tools**: none

**Done when**:

- [x] `example.xml` → `16_9_2010_22-45-43_Ricardo.txt`
- [x] Receiver `[i]Gabriella[/i]` → `[i]Gabriella[-i]` in name; each illegal char class covered
- [x] Gate: `go test ./internal/converter/` passes

**Tests**: unit (GUI-04 1:1 + illegal-char edge cases)
**Gate**: quick
**Commit**: `feat(converter): derive Windows-safe output filenames`

---

### T4: ConvertFile with overwrite semantics

**What**: `ConvertFile(xmlPath, outDir string) (string, error)` — open input, `Convert`, write via `filepath.Join(outDir, name)` with `os.Create` (truncate = overwrite), return written filename.
**Where**: `internal/converter/converter.go`, `internal/converter/converter_test.go`
**Depends on**: T3
**Reuses**: T2/T3 core
**Requirement**: GUI-14 (mechanics)

**Tools**: none

**Done when**:

- [x] Writes expected file into `t.TempDir()`; content matches `Convert` output
- [x] Pre-existing same-name file gets overwritten (old content gone)
- [x] Unreadable input path → error; nonexistent outDir → error (no partial panic)
- [x] Gate: `go test ./internal/converter/` passes

**Tests**: unit
**Gate**: quick
**Commit**: `feat(converter): add ConvertFile with overwrite semantics`

---

### T5: De-duplicated file queue

**What**: `internal/ui/queue.go`: `Queue` with `Add(paths ...string) int` (abs-path de-dup, insertion order), `Remove`, `Clear`, `Items() []string` (copy), `Len()`. No Fyne imports.
**Where**: `internal/ui/queue.go`, `internal/ui/queue_test.go`
**Depends on**: None (parallel-safe after T1, but executes in order)
**Reuses**: —
**Requirement**: GUI-08, GUI-09 (logic)

**Tools**: none

**Done when**:

- [x] Add of duplicate (same abs path, incl. non-clean variants like `dir/../dir/a.xml`) → list unchanged, returns 0 added
- [x] Remove/Clear behave; Items returns defensive copy
- [x] Gate: `go test ./internal/ui/` passes

**Tests**: unit (GUI-08 1:1)
**Gate**: quick
**Commit**: `feat(ui): add de-duplicated file queue`

---

### T6: Folder XML scan

**What**: `ListXML(dir string) ([]string, error)` in `internal/ui/queue.go` — non-recursive `os.ReadDir`, case-insensitive `.xml` match, absolute paths, skips subdirs.
**Where**: `internal/ui/queue.go`, `internal/ui/queue_test.go`
**Depends on**: T5
**Reuses**: —
**Requirement**: GUI-07 (logic)

**Tools**: none

**Done when**:

- [x] Mixed temp dir (`a.xml`, `B.XML`, `c.txt`, subdir with `d.xml`) → returns only `a.xml`+`B.XML`
- [x] Empty dir / no-xml dir → empty slice, nil error
- [x] Gate: `go test ./internal/ui/` passes

**Tests**: unit (GUI-07 + case-insensitive + empty-folder edge cases)
**Gate**: quick
**Commit**: `feat(ui): add non-recursive folder XML scan`

---

### T7: Skip-and-report batch runner

**What**: `internal/ui/batch.go`: `FileError{Path, Reason}`, `Summary{Converted int; Failed []FileError}`, `RunBatch(files []string, outDir string, progress func(done, total int)) Summary` — sequential `converter.ConvertFile`, failures recorded and skipped, progress once per file. No Fyne imports.
**Where**: `internal/ui/batch.go`, `internal/ui/batch_test.go`
**Depends on**: T4
**Reuses**: `converter.ConvertFile`
**Requirement**: GUI-11 (logic), GUI-12, GUI-13 (data), GUI-14

**Tools**: none

**Done when**:

- [x] Mixed batch (valid, malformed, zero-message, nonexistent path) → `Converted` and `Failed` counts exact, reasons non-empty, batch completes
- [x] progress called exactly `len(files)` times with `(1..n, n)`
- [x] Two inputs mapping to same name → one output file, last write wins
- [x] Gate: `go test ./internal/ui/` passes

**Tests**: unit (GUI-12/14 1:1 + mixed-batch edge case)
**Gate**: quick
**Commit**: `feat(ui): add skip-and-report batch runner`

---

### T8: Fyne dependency + app shell, drop CLI

**What**: `go get fyne.io/fyne/v2@v2.7.4`; `internal/ui/app.go` with `Run()` (window "MSN Converter", Resize 600×450, placeholder content); rewrite `main.go` to `ui.Run()`; delete `models/`, `templates/`, `files/` (CLI dropped). One deliverable: repo compiles as a Fyne app.
**Where**: `go.mod`, `go.sum`, `internal/ui/app.go`, `main.go`, deletions
**Depends on**: T1–T7 (deletions require nothing else importing old paths)
**Reuses**: —
**Requirement**: GUI-16

**Tools**: none (network for `go get`)

**Done when**:

- [x] `go build ./...` green with old dirs deleted; no imports of `models/` remain
- [x] Headless test (fyne `test` pkg) asserts window title "MSN Converter"
- [ ] macOS toolchain sanity: `go run .` opens a window (manual, dev machine — deferred to user validation)
- [x] Gate: `go build ./... && go vet ./... && go test ./...`

**Tests**: unit (headless title check; rest is shell)
**Gate**: build
**Commit**: `feat(app): replace CLI with Fyne app shell`

---

### T9: Layout, queue list widget, Convert gating

**What**: Border layout per design (toolbar top / `widget.List` center / output-row + progress + Convert bottom); list rows show filename + per-item ✕; clear-all button; `updateConvertState()` implementing GUI-10 (`Len>0 && outDir!="" && !running`).
**Where**: `internal/ui/app.go`, `internal/ui/app_test.go`
**Depends on**: T8
**Reuses**: `Queue` (T5)
**Requirement**: GUI-09 (widget), GUI-10, GUI-16

**Tools**: none

**Done when**:

- [ ] Headless tests: enablement matrix — empty+noDir / files+noDir / empty+dir → disabled; files+dir → enabled
- [ ] Tap ✕ removes row; tap clear-all empties list; Convert re-disables when list empties
- [ ] Gate: `go vet ./... && go test ./...`

**Tests**: unit (headless, GUI-10 1:1)
**Gate**: full
**Commit**: `feat(ui): add queue list layout and convert gating`

---

### T10: File, folder and output pickers

**What**: Wire "Add files" (`dialog.NewFileOpen` + `storage.NewExtensionFileFilter([]string{".xml"})`), "Add folder" (`dialog.NewFolderOpen` → `ListXML` → `Queue.Add`), "Choose output folder" (`dialog.NewFolderOpen` → outDir + label). Dialog callbacks delegate to plain funcs `addFiles(paths []string)` / `setOutputDir(path string)` so logic is headless-testable.
**Where**: `internal/ui/app.go`, `internal/ui/app_test.go`
**Depends on**: T9
**Reuses**: `ListXML` (T6), `Queue` (T5)
**Requirement**: GUI-06, GUI-07

**Tools**: none

**Done when**:

- [ ] Headless: `addFiles` appends+dedups and refreshes list; `setOutputDir` updates label + gating
- [ ] File dialog has `.xml` filter set; cancel (nil URI) is a no-op (callback guard tested)
- [ ] Gate: `go vet ./... && go test ./...`

**Tests**: unit (headless)
**Gate**: full
**Commit**: `feat(ui): wire file, folder and output pickers`

---

### T11: Drag-and-drop input

**What**: `win.SetOnDropped(func(_ fyne.Position, uris []fyne.URI))` → `addDropped(paths []string)`: `os.Stat` each; dir → `ListXML`; `.xml` file (case-insensitive) → add; else ignore.
**Where**: `internal/ui/app.go`, `internal/ui/app_test.go`
**Depends on**: T10
**Reuses**: `ListXML`, `Queue`
**Requirement**: GUI-15

**Tools**: none

**Done when**:

- [ ] Headless tests on `addDropped`: mixed drop (xml file, txt file, folder with xml, nonexistent path) → only xml + folder-xml added, deduped, no error
- [ ] Gate: `go vet ./... && go test ./...`

**Tests**: unit (headless, GUI-15 1:1)
**Gate**: full
**Commit**: `feat(ui): add drag-and-drop input`

---

### T12: Batch execution wiring — goroutine, progress, summary

**What**: Convert tap → snapshot `queue.Items()`, `running=true` + disable, `go RunBatch(...)` with progress wrapped in `fyne.Do` (ProgressBar.SetValue), completion via `fyne.Do`: summary dialog (`ShowInformation` when 0 failed; `ShowCustom` scrollable reasons otherwise), `running=false`, re-enable, queue untouched.
**Where**: `internal/ui/app.go`, `internal/ui/app_test.go`
**Depends on**: T11
**Reuses**: `RunBatch` (T7)
**Requirement**: GUI-11, GUI-12, GUI-13, GUI-14, GUI-18 (list kept)

**Tools**: none

**Done when**:

- [ ] Headless: tapping Convert with real temp fixtures produces expected `.txt` files; progress reaches 1.0; Convert disabled during run, re-enabled after; queue unchanged after run
- [ ] Summary content: `N converted, M failed` + per-file reasons (assert on generated summary string/objects)
- [ ] Manual smoke on macOS: `go run .`, mixed batch, window responsive
- [ ] Gate: `go build ./... && go vet ./... && go test ./...`

**Tests**: unit (headless integration-style)
**Gate**: build
**Commit**: `feat(ui): run batch off UI thread with progress and summary`

---

## Phase Execution Map

```
Phase 1 → Phase 2 → Phase 3

Phase 1:  T1 ──→ T2 ──→ T3 ──→ T4
Phase 2:  T5 ──→ T6 ──→ T7
Phase 3:  T8 ──→ T9 ──→ T10 ──→ T11 ──→ T12
```

12 tasks → packs into 2 batches (Phase 1+2 = 7 tasks; Phase 3 = 5 tasks) → sub-agent offer applies.

---

## Task Granularity Check

| Task | Scope | Status |
| ---- | ----- | ------ |
| T1 struct move + fixture | 1 file + 1 asset move | ✅ Granular |
| T2 Convert + embed | 1 function + asset | ✅ Granular |
| T3 filename rule | 1 function | ✅ Granular |
| T4 ConvertFile | 1 function | ✅ Granular |
| T5 Queue | 1 type | ✅ Granular |
| T6 ListXML | 1 function | ✅ Granular |
| T7 RunBatch | 1 function + 2 types | ✅ Granular (cohesive) |
| T8 dep + shell + CLI drop | several files | ⚠️ Cohesive — deletions and entry-point swap are build-atomic, cannot split without broken commits |
| T9 layout + gating | 1 file | ✅ Granular |
| T10 pickers | 1 file | ✅ Granular |
| T11 DnD | 1 function + handler | ✅ Granular |
| T12 batch wiring | 1 file | ✅ Granular |

## Diagram-Definition Cross-Check

| Task | Depends On (body) | Diagram Shows | Status |
| ---- | ----------------- | ------------- | ------ |
| T1 | none | phase start | ✅ Match |
| T2 | T1 | T1→T2 | ✅ Match |
| T3 | T2 | T2→T3 | ✅ Match |
| T4 | T3 | T3→T4 | ✅ Match |
| T5 | none (ordered after Phase 1) | Phase 2 start | ✅ Match |
| T6 | T5 | T5→T6 | ✅ Match |
| T7 | T4 (+ after T6 in order) | Phase 2 after Phase 1; T6→T7 | ✅ Match |
| T8 | T1–T7 | Phase 3 after Phase 2 | ✅ Match |
| T9 | T8 | T8→T9 | ✅ Match |
| T10 | T9 | T9→T10 | ✅ Match |
| T11 | T10 | T10→T11 | ✅ Match |
| T12 | T11 | T11→T12 | ✅ Match |

## Test Co-location Validation

| Task | Code Layer | Matrix Requires | Task Says | Status |
| ---- | ---------- | --------------- | --------- | ------ |
| T1 | model/schema | none (build gate) | none/build | ✅ OK |
| T2 | converter domain | unit | unit | ✅ OK |
| T3 | converter domain | unit | unit | ✅ OK |
| T4 | converter domain | unit | unit | ✅ OK |
| T5 | ui domain | unit | unit | ✅ OK |
| T6 | ui domain | unit | unit | ✅ OK |
| T7 | ui domain | unit | unit | ✅ OK |
| T8 | app shell | headless unit (title) + build | unit+build | ✅ OK |
| T9 | app wiring | headless unit | unit | ✅ OK |
| T10 | app wiring | headless unit | unit | ✅ OK |
| T11 | app wiring | headless unit | unit | ✅ OK |
| T12 | app wiring | headless unit | unit | ✅ OK |
