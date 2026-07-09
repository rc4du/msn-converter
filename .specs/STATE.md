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

- **Feature**: error-log — `.specs/features/error-log/spec.md` — **spec APPROVED** (2026-07-09), Execute not started
- **Phase / Task**: Specify complete (Medium scope: design inline, tasks implicit). 10 requirements LOG-01…LOG-10, all decisions user-confirmed in grilling session; closure gate passed.
- **Completed**: none (no code yet)
- **In-progress** (file:line): none
- **Next step**: Execute per implement.md — start by listing atomic steps inline (expected ~5: `internal/ui/errorlog.go` writeErrorLog + forensics; wire into `runBatchAsync`/`finishBatch`/`showSummary` in `internal/ui/app.go`; stale-log delete on success; version via `runtime/debug.ReadBuildInfo`; tests for LOG-01…LOG-10). If listing reveals >5 steps → create formal tasks.md (safety valve). Tests derive from spec ACs; gate before done; one atomic commit per task; Verifier runs after last task.
- **Blockers**: none
- **Uncommitted files**: `.specs/features/error-log/spec.md` (new), `.specs/STATE.md` (this update); `GUI_PLAN.md` deleted in working tree (pre-existing, unrelated)
- **Branch**: main
