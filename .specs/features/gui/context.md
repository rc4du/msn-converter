# GUI Context

**Gathered:** 2026-07-08
**Spec:** `.specs/features/gui/spec.md`
**Status:** Ready for design

---

## Feature Boundary

Fyne desktop app that bulk-converts MSN Messenger XML logs to `.txt`: build a de-duplicated file queue (picker / folder / drag-drop), pick an output folder, convert off the UI thread with progress, report a skip-and-report summary. CLI behavior dropped. Packaging deferred.

---

## Implementation Decisions

### Output filename sanitization (user-decided)

- Base name `{date}_{time}_{receiver}.txt` from first message (per GUI_PLAN.md)
- Date: `/` → `_` (existing logic preserved)
- Then every remaining Windows-illegal char (`\ / : * ? " < > |`) in the base name → `-`
- Example: `16_9_2010_22-45-43_Ricardo.txt`; receiver `[i]Gabriella[/i]` → `[i]Gabriella[-i]`

### Template engine (user-decided)

- `text/template`, not `html/template`
- Message text lands in the `.txt` verbatim — no HTML entity escaping
- Template content/logic otherwise unchanged (sender grouping, LF endings)

### File list after batch (user-decided)

- Kept unchanged after the summary dialog; user clears manually via clear-all

### Agent's Discretion

- Widget layout details within the ~600×450 window (list on top, controls below, etc.)
- Exact error message wording for per-file failure reasons
- Internal package APIs (per AGENTS.md structure: `internal/converter`, `internal/ui`)

### Declined / Undiscussed Gray Areas → Assumptions

All recorded in spec.md Assumptions table: `.xml` picker filter, Convert disabled while running, silent no-op on xml-less folders, per-file progress granularity, repo restructure per AGENTS.md.

---

## Specific References

- `GUI_PLAN.md` — authoritative pre-agreed decisions (input sources, overwrite, LF, skip-and-report, window size/title, deferred items)
- `AGENTS.md` — authoritative project structure (`main.go` at root, `internal/converter` owns XML structs + embedded template, `internal/ui`, `testdata/`)

---

## Deferred Ideas

None — discussion stayed within feature scope.
