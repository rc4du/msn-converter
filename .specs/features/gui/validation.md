# GUI Validation

**Date**: 2026-07-08
**Spec**: `.specs/features/gui/spec.md`
**Diff range**: `d559afa..5813681` (12 commits, T1–T12)
**Verifier**: independent sub-agent (author ≠ verifier); coverage re-derived from spec, evidence-or-zero

**Verdict: PASS ✅**

---

## Gate Check

| Command | Exit code |
| ------- | --------- |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `go test ./...` | 0 |

- **Result**: 52 passed, 0 failed, 0 skipped (3 packages: `internal/converter`, `internal/ui`, root build)
- **Test count before feature** (d559afa): 0
- **Test count after feature**: 52 (incl. subtests) — delta **+52**
- Gate re-run after sensor restores: identical result (52 passed, exit 0)

---

## Spec-Anchored Acceptance Criteria (GUI-01..GUI-16)

| AC | Spec-defined outcome | Evidence (`file:line` — assertion) | Status |
| -- | -------------------- | ---------------------------------- | ------ |
| GUI-01 valid log → templated text, grouped by sender, `{time} - {text}`, LF, verbatim | exact rendered text of `testdata/example.xml`; no CRLF; `&`/`<`/`'` unescaped | `internal/converter/converter_test.go:37` — `if got := string(res.Content); got != want` (full literal expected text lines 24–35); `:40` — `strings.Contains(..., "\r\n")` rejected; verbatim: `:59-61` — `want := "16/9/2010 - Ana\n22:45:43 - Tom & Jerry <3 it's fine\n"` asserted equal | ✅ PASS |
| GUI-02 malformed/unreadable → error, never panic | non-nil error | `internal/converter/converter_test.go:124-127` — `Convert(strings.NewReader("<Log><Message Date=")); err == nil → Fatal`; unreadable path: `converter_test.go:204-208` — `ConvertFile(nonexistent) ... err == nil → Fatal` | ✅ PASS |
| GUI-03 zero `<Message>` → error stating no messages | `errors.Is(err, ErrNoMessages)` | `internal/converter/converter_test.go:142-144` — `if !errors.Is(err, ErrNoMessages) { t.Fatalf(...) }`; `ErrNoMessages = errors.New("file has no messages")` (`converter.go:22`) | ✅ PASS |
| GUI-04 filename `{date /→_}_{time}_{receiver}.txt`, illegal chars → `-` | literal `16_9_2010_22-45-43_Ricardo.txt` | `internal/converter/converter_test.go:78-80` — `if want := "16_9_2010_22-45-43_Ricardo.txt"; res.FileName != want`; all 9 illegal chars (`\ / : * ? " < > \|`) table-tested `:85-120` incl. `[i]Gabriella[/i] → [i]Gabriella[-i]` (`:91`) | ✅ PASS |
| GUI-05 template embedded via go:embed, no runtime file read | conversion succeeds with cwd outside repo | `internal/converter/converter_test.go:230-238` — `t.Chdir(t.TempDir())` then `Convert` succeeds and output equals `want`; directive at `converter.go:16` (`//go:embed output.txt.tmpl`) | ✅ PASS |
| GUI-06 Add-files dialog filtered to `.xml`, appends chosen file | filter matches only `.xml`; file appended | `internal/ui/app_test.go:153-157` — `u.xmlFilter.Matches(.../a.xml)` true, `.../b.txt` false; append: `app_test.go:117-125` — queue equals `[a.xml b.xml]` after `addFiles`; cancel no-op `:166-178`. Native dialog open + `SetFilter` attachment (`app.go:78-79`) → **manual (documented exclusion)** | ✅ PASS (headless part) |
| GUI-07 folder picker adds `*.xml` non-recursive | only direct `.xml` (case-insensitive), absolute paths | `internal/ui/queue_test.go:119-129` — `slices.Equal(got, [B.XML a.xml])` (excludes `c.txt`, `subdir/d.xml`) + `filepath.IsAbs`; wiring: `app_test.go:191-196` — `onFolderPicked` → queue = `[a.xml]` only. Native dialog → **manual** | ✅ PASS |
| GUI-08 de-dup by absolute path | duplicate add → list unchanged, 0 added | `internal/ui/queue_test.go:20-22` — `q.Add(a) != 0` fails; non-clean variant `:27-29` — `Add(dir/../dir/a.xml) != 0` fails; `:31-33` — `Items()` still `[a]` | ✅ PASS |
| GUI-09 per-item remove + clear-all | tapped file leaves list; clear empties | `internal/ui/app_test.go:81-86` — `test.Tap(remove)` → `Items() == [/tmp/b.xml]`; `:99-106` — `test.Tap(u.clearBtn)` → `Len() == 0`; row shows base name `:363-366`; logic: `queue_test.go:58-95` | ✅ PASS |
| GUI-10 Convert disabled unless list non-empty AND outDir set (AND not running) | 5-case enablement matrix | `internal/ui/app_test.go:47-73` — table incl. `{files and outDir → enabled}`, `{running → disabled}`; `if got := !u.convertBtn.Disabled(); got != tc.enabled` | ✅ PASS |
| GUI-11 off-UI goroutine, Convert disabled during run, progress once per file | progress `(1..n, n)` exactly n times; disabled while running; reaches 1.0 | `internal/ui/batch_test.go:89-96` — `len(calls) == len(files)` and each call `(i+1, len(files))`; `app_test.go:262-270` — `beginBatch` → `running`, `convertBtn.Disabled()`, progress reset 0; `app_test.go:305-306` — `u.progress.Value != 1.0` fails; async dispatch `go u.runBatchAsync` (`app.go:203`) structural → responsiveness **manual** | ✅ PASS |
| GUI-12 failing file skipped, batch continues | exact Converted/Failed counts, batch completes | `internal/ui/batch_test.go:51-65` — `s.Converted != 1` fails, `len(s.Failed) != 3` fatal, each `Failed[i].Path` equals expected path, `Reason == ""` fails; end-to-end `app_test.go:301-303` — `s.Converted != 1 \|\| len(s.Failed) != 1` fails | ✅ PASS |
| GUI-13 summary `N converted, M failed` + per-file reasons; list unchanged | exact summary string; queue identical after run | `internal/ui/app_test.go:342-345` — `want := "1 converted, 2 failed\nbad.xml: decoding xml: EOF\nempty.xml: file has no messages"` asserted equal; `:330-332` — `"3 converted, 0 failed"`; queue unchanged `:316-324` — element-wise equality before/after. Dialog visual pop (`ShowInformation`/`ShowCustom`, `app.go:252-259`) → **manual** (tasks.md scoped assertion to summary string) | ✅ PASS |
| GUI-14 name clash → overwrite, last write wins | one output file; content = later input | `internal/converter/converter_test.go:195-200` — `strings.Contains(got, "OLD CONTENT")` fails + prefix check; `internal/ui/batch_test.go:116-127` — `len(entries) != 1` fatal, content `== "16/9/2010 - Ana\n22:45:43 - SECOND\n"` (last write wins) | ✅ PASS |
| GUI-15 DnD: `.xml` + folder's direct `.xml` appended (deduped), others ignored | queue = exactly `[a.xml B.XML sub/d.xml]` from mixed drop | `internal/ui/app_test.go:234-243` — `want := []string{xml, upper, inSub}`, element-wise equality; drop includes txt, nonexistent, duplicate — all excluded. Real OS drop → handler wired `app.go:104-110` → **manual** | ✅ PASS |
| GUI-16 window "MSN Converter", ~600×450 | title exact; size ≥ 600×450 | `internal/ui/app_test.go:37-39` — `w.Title() != "MSN Converter"` fails; `:40-43` — `size.Width < 600 \|\| size.Height < 450` fails | ✅ PASS |

**Status**: 16/16 ACs evidenced (GUI-06/07/11/13/15 have documented manual sub-items — see Manual Validation).

---

## Edge Cases

| Edge case | Evidence | Status |
| --------- | -------- | ------ |
| 0-byte / truncated XML skipped and reported, no panic | `converter_test.go:131-135` (0-byte → error), `:123-127` (truncated → error); batch-level: `batch_test.go:36-39,54-64` (malformed in mixed batch → Failed with reason, batch completes) | ✅ |
| Two inputs → same output name: later overwrites | `batch_test.go:100-128` — 1 file in outDir, content is second input's | ✅ |
| Output folder unwritable/deleted → each file fails with reason, batch completes | Indirect composition: `converter_test.go:212-216` (nonexistent outDir → `ConvertFile` error) + `batch.go:26-28` single uniform error path + `batch_test.go:51-65` (every `ConvertFile` error → `Failed` entry with reason, loop continues). No direct RunBatch-with-deleted-outDir test | ✅ (indirect — see note) |
| Picked folder with zero `.xml` → list unchanged, no error | `queue_test.go:133-140` (empty dir → empty, nil err), `:144-156` (txt-only dir → empty, nil err) | ✅ |
| `.XML` uppercase treated as `.xml` | `queue_test.go:101,119` — `B.XML` included in scan result; `app_test.go:213,226,235` — dropped `B.XML` added | ✅ |

**5/5 covered** (one via sound two-step composition rather than a single direct test — acceptable because `RunBatch` has exactly one error path, itself under test; noted for completeness, not a gap).

---

## Discrimination Sensor

Scratch-state mutations, one at a time; each restored with `git checkout -- <file>` and verified with `git diff --quiet` before the next. Final full gate re-run green.

| # | Mutation | File | Tests run | Result |
| - | -------- | ---- | --------- | ------ |
| 1 | Stop replacing `:` in filename sanitization (`` `\/:*?"<>|` `` → `` `\/*?"<>|` ``) | `internal/converter/converter.go:69` | `go test ./internal/converter/ ./internal/ui/` | ✅ KILLED — `TestConvertFileNameSanitization` (colon + example cases), `TestConvertFileWritesOutput`, `TestConvertFileOverwritesExisting`, `TestRunBatchMixed`, `TestRunBatchNameClashLastWriteWins`, `TestConvertTapRunsBatch` all fail |
| 2 | Disable zero-message guard (`len(log.Messages) == 0` → `< 0`) | `internal/converter/converter.go:35` | same | ✅ KILLED — `TestConvertZeroMessages` fails (panic on `Messages[0]` caught by test harness), `TestRunBatchMixed` fails |
| 3 | Remove de-dup in `Queue.Add` (always append) | `internal/ui/queue.go:25-27` | `go test ./internal/ui/` | ✅ KILLED — `TestQueueAddDeduplicates`, `TestAddFilesAppendsDedupsAndRefreshes`, `TestAddDroppedMixed` fail |
| 4 | Stop appending `FileError` on failure in `RunBatch` | `internal/ui/batch.go:27` | `go test ./internal/ui/` | ✅ KILLED — `TestRunBatchMixed` (`len(Failed) = 0; want 3`), `TestConvertTapRunsBatch` (`want 1 converted, 1 failed`) fail |
| 5 | Flip enablement condition (`u.outDir != ""` → `== ""`) in `updateConvertState` | `internal/ui/app.go:126` | `go test -timeout 60s ./internal/ui/` | ✅ KILLED — `TestConvertEnablementMatrix` (2 cases), `TestClearAllEmptiesListAndDisablesConvert`, `TestAddFilesAppendsDedupsAndRefreshes`, `TestSetOutputDirUpdatesLabelAndGating` fail (and `TestConvertTapRunsBatch` deadlocks — disabled button never starts batch) |

**Sensor depth**: 5 mutations (above lightweight default)
**Result**: 5/5 killed — ✅ PASS. Working tree verified clean of mutations after each restore and at end (`git status --porcelain` shows only pre-existing unrelated changes: `.gitignore`, `README.md` deletion, untracked `.specs/` docs).

---

## Payload / Conjunction Rule Check

- `Result.FileName` asserted on literal value (`converter_test.go:78`), `Result.Content` on full literal text (`:24-38`, `:59-61`).
- `Summary.Converted` / `len(Summary.Failed)` / `Failed[i].Path` / `Failed[i].Reason` asserted on values (`batch_test.go:51-65`, `app_test.go:301-303`).
- Persisted output files asserted on existence AND content (`converter_test.go:174`, `batch_test.go:124-126`).
- No "no error only" stand-ins found for spec-defined payloads.

---

## Spec-Precision Notes

1. **GUI-16 "~600×450"**: spec is approximate by design ("~"); test asserts `>= 600×450` canvas size. Acceptable; no action.
2. **Failure reason text**: spec requires "per-file failure reasons" without defining wording. Batch test asserts non-empty; `TestSummaryText` pins exact rendered strings (`decoding xml: EOF`, `file has no messages`). Adequate; flagged only for awareness.
3. **"Off the UI goroutine" (GUI-11)**: goroutine dispatch is structural (`go u.runBatchAsync`, `app.go:203`); headless tests cannot assert thread identity. Observable consequences (disable-during-run, progress, re-enable) are asserted. Residual → manual responsiveness smoke.

No blocking spec-precision gaps.

---

## Manual Validation Items Outstanding (documented exclusions — not gaps)

Per tasks.md coverage matrix, native OS dialogs and window smokes are excluded from headless tests:

1. `go run .` opens a 600×450 window titled "MSN Converter" on macOS (T8 deferred item).
2. "Add files" opens the native file dialog with the `.xml` filter actually attached (`app.go:78-80`).
3. "Add folder" / "Choose output folder…" open native folder pickers and feed the tested callbacks.
4. Real OS drag-and-drop onto the window reaches `addDropped` (URI→path conversion, `app.go:104-110`).
5. Mixed-batch smoke: summary dialog visually appears (info vs. scrollable custom) and window stays responsive during a large batch (success criterion 4, T12 deferred item).

---

## Task Completion

T1–T12 all committed (`32357e0`..`5813681`), each with its stated commit message; tasks.md Done-when boxes checked except the two explicitly deferred manual items (T8 `go run .`, T12 macOS smoke) listed above.

---

## Requirement Traceability Update

| Requirement | Previous | New |
| ----------- | -------- | --- |
| GUI-01..GUI-16 | Pending | ✅ Verified (GUI-06/07/11/13/15 pending manual dialog/window smoke for the excluded native parts) |

---

## Summary

**Overall**: ✅ Ready (pending listed manual smokes)
**Spec-anchored check**: 16/16 ACs matched spec outcomes; 0 blocking spec-precision gaps
**Edge cases**: 5/5 covered
**Sensor**: 5/5 mutations killed
**Gate**: build 0 / vet 0 / test 0 — 52 passed, 0 failed (delta +52 vs. pre-feature 0)
**Issues found**: none blocking. Optional hardening (not required): a direct `RunBatch` test with a deleted output folder would make the "unwritable outDir" edge case first-class instead of compositional.
