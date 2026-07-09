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

- **Feature**: error-log — **COMPLETE & VERIFIED** (2026-07-09). Verifier PASS: 10/10 ACs evidence-matched, 5/5 edge cases, sensor 3/3 mutants killed. Report: `.specs/features/error-log/validation.md`.
- **Phase / Task**: Execute done — 3 atomic commits: `4917b2f` (errorlog core: writeErrorLog + forensics + version), `add3a0d` (handleErrorLog gating + stale cleanup), `fda3d04` (UI wiring + dialog discovery line). Gate: build + vet + 68 tests green.
- **Completed**: LOG-01…LOG-10 (spec traceability marked Complete)
- **In-progress** (file:line): none
- **Next step**: none for this feature. Candidate follow-up from spec Out-of-Scope: UTF-16 input decoding bug (log will prove it in the wild first).
- **Blockers**: none
- **Uncommitted files**: none after closing docs commit
- **Branch**: main
