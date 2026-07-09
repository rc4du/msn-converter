# STATE

## Decisions

### AD-001
- **Decision**: Output filenames are sanitized as: date `/`→`_`, then every remaining Windows-illegal char (`\ / : * ? " < > |`) in the base name → `-`.
- **Reason**: Primary target is Windows; MSN `Time` attrs contain `:` and FriendlyNames are arbitrary user text.
- **Trade-off**: Filenames can differ from what macOS/Linux would tolerate (`:` allowed there); cross-platform names win over fidelity.
- **Scope**: Any code that derives filesystem names from log content (converter and future exporters).
- **Date**: 2026-07-08
- **Status**: active

### AD-002
- **Decision**: Plain-text output is rendered with `text/template`, never `html/template`.
- **Reason**: `.txt` output must contain message text verbatim; `html/template` HTML-escapes (`&` → `&amp;`, `'` → `&#39;`), corrupting chat content.
- **Trade-off**: Loses auto-escaping if an HTML export format is ever added (that format would opt into `html/template` explicitly).
- **Scope**: All text-producing templates in the project.
- **Date**: 2026-07-08
- **Status**: active

## Handoff

- **Feature**: gui — `.specs/features/gui/` — **COMPLETE, Verifier PASS** (2026-07-08)
- **Phase / Task**: all 12 tasks executed via 2 batch workers; Verifier report at `.specs/features/gui/validation.md` (16/16 ACs evidenced, 5/5 edge cases, sensor 5 injected / 5 killed, 52 tests green)
- **Completed**: T1–T12, commits `32357e0..5813681` on main (one atomic commit per task)
- **In-progress** (file:line): none
- **Next step**: manual smoke on macOS — `go run .`: window "MSN Converter" ~600×450 opens; native pickers (.xml filter) feed queue; real drag-and-drop works; mixed batch shows summary dialog, window stays responsive. Then (deferred scope): Windows `.exe` packaging via fyne-cross/CI.
- **Blockers**: none
- **Uncommitted files**: `.specs/` (all planning + validation docs), `AGENTS.md`, `GUI_PLAN.md` (untracked); `.gitignore` (modified); `README.md` (deleted) — pre-existing, intentionally not committed with feature tasks
- **Branch**: main
