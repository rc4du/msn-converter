# Error Log Specification

## Problem Statement

When a batch conversion fails on a user's machine, the failure details exist only in a transient dialog: the user closes it and the information is gone, and `err.Error()` alone is too shallow to debug remotely (e.g. encoding issues need the file's leading bytes). Users need a persistent, detailed artifact they can email to the developer.

## Goals

- [ ] A failed batch leaves a self-contained log file the user can find and send without instructions beyond one dialog line.
- [ ] The log carries enough forensics (version, OS, per-file error + leading bytes) to diagnose conversion bugs — especially encoding bugs — without asking the user for the original chat file.

## Out of Scope

| Feature | Reason |
| ------- | ------ |
| Crash / panic / startup logging | Different architecture (logger at startup, panic recovery); feature targets conversion failures only (grilling Q1) |
| Fixing UTF-16 input decoding | Separate bug; this log will *prove* it in the wild first |
| Timestamped / rotating logs | Fixed name + overwrite chosen (grilling Q4); reproduce-then-send flow makes history unnecessary |
| Embedding input file content in log | MSN logs are personal conversations; privacy (grilling Q5) |
| Telemetry / auto-upload | User emails file manually; no network code |
| "Open log folder" button or other new UI | One dialog line suffices (grilling Q6) |

---

## Assumptions & Open Questions

All decisions resolved with user in grilling session (2026-07-09).

| Decision | Chosen | Rationale | Confirmed? |
| -------- | ------ | --------- | ---------- |
| Gap addressed | Persistence + detail; no crash logging | Matches "send me the file on conversion error" flow | y |
| Location | Output folder | User just picked it, guaranteed to exist at batch time, zero new UI | y |
| When written | Only when ≥1 file fails | No failures = nothing to investigate; no clutter | y |
| Rotation | Fixed name `conversion-errors.log`, overwritten | Latest failure is the one debugged; one-sentence user instruction | y |
| Content tier | Header + per-failure forensics incl. 32-byte hex preview | Error strings alone can't diagnose encoding bugs; preview ≈ XML header/BOM, not chat text | y |
| Discovery | One line appended to failure dialog | Seen at exact moment of failure | y |
| Stale log | Deleted on fully successful batch | Prevents user emailing outdated log | y |
| App version source | `runtime/debug.ReadBuildInfo()` → `vcs.revision` (short), fallback `unknown` | Zero build config; identifies build well enough | assumption (not grilled) |
| Log encoding | UTF-8 plain text, `\n` line endings | Notepad (Win10+) handles both; keeps template simple | assumption (not grilled) |

**Open questions:** none — all resolved or logged above.

---

## User Stories

### P1: Persistent error log ⭐ MVP

**User Story**: As the developer receiving bug reports, I want every failed batch to write a detailed log file into the output folder so that users can send me one file that lets me diagnose the failure.

**Why P1**: The entire feature — without the file there is nothing to send.

**Acceptance Criteria**:

1. WHEN a batch completes with ≥1 failed file THEN the system SHALL write `conversion-errors.log` (UTF-8 plain text) into the chosen output folder, replacing any existing file of that name.
2. WHEN the log is written THEN it SHALL contain a header with: batch timestamp (local time, `2006-01-02 15:04:05` format), app version (short VCS revision from build info, or `unknown` when unavailable), `GOOS/GOARCH`, and the counts line `N converted, M failed`.
3. WHEN the log is written THEN it SHALL contain, per failed file: the full input path, the error string, file size in bytes, modification time, and an uppercase space-separated hex dump of the file's first 32 bytes (fewer when the file is shorter).
4. WHEN forensics collection for a failed file itself fails (file deleted, unreadable) THEN the log SHALL record the collection error for that field and SHALL still be written with all remaining entries intact.
5. WHEN a batch completes with 0 failures THEN the system SHALL NOT write a log.

**Independent Test**: Run a batch with one valid and one malformed XML into a temp output dir → `conversion-errors.log` exists with header, counts `1 converted, 1 failed`, and a forensics block for the malformed file. Run an all-valid batch → no log created.

---

### P1: User discovers the log ⭐ MVP

**User Story**: As a non-technical user hitting a conversion error, I want the failure dialog to tell me a log file was saved and what to do with it so that I can report the bug without guidance.

**Why P1**: Without discovery the file is never sent; the feature silently fails its purpose.

**Acceptance Criteria**:

1. WHEN the log was written successfully THEN the failure dialog text SHALL end with the line `Details saved to <full log path> — send this file when reporting bugs.`
2. WHEN writing the log failed THEN the failure dialog text SHALL instead end with `Could not write log: <error>` and the app SHALL continue normally (no crash, no retry).
3. WHEN a batch has 0 failures THEN the success dialog SHALL NOT mention the log.

**Independent Test**: Force a failing batch → dialog shows the exact path line. Make output dir read-only mid-flow (or inject write error in test) → dialog shows the could-not-write line.

---

### P2: Stale log cleanup

**User Story**: As the developer, I want a fully successful batch to remove a leftover `conversion-errors.log` so that users never email me an outdated log.

**Why P2**: Support-noise prevention; feature works without it, timestamp in header is the fallback guard.

**Acceptance Criteria**:

1. WHEN a batch completes with 0 failures AND `conversion-errors.log` exists in the output folder THEN the system SHALL delete it.
2. WHEN the deletion target does not exist or deletion fails THEN the system SHALL proceed silently (no dialog change, no error surfaced).

**Independent Test**: Place a dummy `conversion-errors.log` in output dir, run all-valid batch → file gone; run same batch with dir lacking the file → no error.

---

## Edge Cases

- WHEN a failed input file is shorter than 32 bytes THEN the hex preview SHALL show exactly the bytes present.
- WHEN a failed input file was deleted between conversion failure and log writing THEN its entry SHALL state the stat/read error and other entries SHALL be unaffected (LOG-04).
- WHEN every file in the batch fails THEN the log SHALL list every failure (no truncation) and the counts line SHALL read `0 converted, M failed`.
- WHEN the output folder becomes unwritable after conversion (log write fails) THEN the dialog SHALL show the could-not-write line (LOG-07) — conversion results already on disk are unaffected.
- WHEN the binary lacks VCS build info (plain `go build` without git) THEN the version field SHALL read `unknown`.

Remaining implicit-requirement dimensions (auth, rate limits, concurrency beyond the existing single-batch gate, external dependencies, idempotency): N/A for this scope — single-user desktop app, one batch at a time enforced by existing `running` flag, no network.

---

## Requirement Traceability

| Requirement ID | Story | AC | Phase | Status |
| -------------- | ----- | -- | ----- | ------ |
| LOG-01 | P1: Persistent error log | 1 | Execute | Complete |
| LOG-02 | P1: Persistent error log | 2 | Execute | Complete |
| LOG-03 | P1: Persistent error log | 3 | Execute | Complete |
| LOG-04 | P1: Persistent error log | 4 | Execute | Complete |
| LOG-05 | P1: Persistent error log | 5 | Execute | Complete |
| LOG-06 | P1: User discovers the log | 1 | Execute | Complete |
| LOG-07 | P1: User discovers the log | 2 | Execute | Complete |
| LOG-08 | P1: User discovers the log | 3 | Execute | Complete |
| LOG-09 | P2: Stale log cleanup | 1 | Execute | Complete |
| LOG-10 | P2: Stale log cleanup | 2 | Execute | Complete |

**Coverage:** 10 total, 10 covered by tests (tasks implicit — Medium scope), 0 unmapped.

---

## Success Criteria

- [ ] A user with a failing batch can locate and email `conversion-errors.log` guided only by the dialog line.
- [ ] A received log identifies the failing build (version), environment (OS/arch), and per-file cause — an encoding bug is recognizable from the hex preview alone (e.g. `FF FE` BOM).
- [ ] No log file appears for fully successful batches; leftover logs from earlier failures are removed by a later clean run.
