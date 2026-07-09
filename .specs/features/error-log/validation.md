# Error Log Validation

**Date**: 2026-07-09
**Spec**: `.specs/features/error-log/spec.md`
**Diff range**: `fc63651..fda3d04` (3 commits: `4917b2f`, `add3a0d`, `fda3d04` — `internal/ui/errorlog.go`, `internal/ui/errorlog_test.go`, `internal/ui/app.go`)
**Verifier**: independent sub-agent (author ≠ verifier), evidence-or-zero

---

## Task Completion

No `tasks.md` for this feature (tasks implicit — Medium scope, per spec traceability note). All 10 requirements marked Complete in spec.md; the three feature commits map 1:1 to the three stories (write log → gate/cleanup → dialog discovery).

---

## Spec-Anchored Acceptance Criteria

| Criterion (WHEN X THEN Y) | Spec-defined outcome | `file:line` + assertion | Result |
| ------------------------- | -------------------- | ----------------------- | ------ |
| **LOG-01** — WHEN batch has ≥1 failure THEN write `conversion-errors.log` (UTF-8 text) into output folder, replacing existing | File named exactly `conversion-errors.log` at outDir, old content replaced, `\n` endings | `internal/ui/errorlog_test.go:55` — `path != stale` (path = `outDir/conversion-errors.log`); `:60` — `strings.Contains(got, "OLD CONTENT")` → fail (overwrite proven); `:63` — `strings.Contains(got, "\r")` → fail (`\n` endings); gating: `errorlog_test.go:213-218` — `path == filepath.Join(outDir, "conversion-errors.log")` + `os.Stat(want)` succeeds; E2E wiring: `errorlog_test.go:368-375` — log read back from outDir after tapped failing batch | ✅ PASS |
| **LOG-02** — WHEN log written THEN header has batch timestamp (`2006-01-02 15:04:05` local), app version (short VCS rev / `unknown`), GOOS/GOARCH, counts line `N converted, M failed` | Exact strings: `2026-07-09 15:04:05`, `abc1234`, `GOOS/GOARCH`, `1 converted, 1 failed` | `internal/ui/errorlog_test.go:81-90` — `strings.Contains(got, want)` for all four exact strings; version derivation: `errorlog_test.go:99` — `versionFromBuildInfo(bi, true) != "a1b2c3d"` → fail (7-char truncation); `:102`, `:105` — `!= "unknown"` → fail | ✅ PASS |
| **LOG-03** — WHEN log written THEN per failed file: full path, error string, size in bytes, mod time, uppercase space-separated hex of first 32 bytes | `40 bytes`, `2026-07-08 10:30:00`, hex `00 01 … 1F` (uppercase, space-separated, exactly 32) | `internal/ui/errorlog_test.go:133-145` — `strings.Contains` for full path, `"parse error: unexpected EOF"`, `"40 bytes"`, `"2026-07-08 10:30:00"`, `wantHex` (contains `0A…1F`, so lowercase would fail); `:146-148` — `strings.Contains(got, wantHex+" 20")` → fail (dump capped at 32 bytes) | ✅ PASS |
| **LOG-04** — WHEN forensics collection fails THEN record collection error for the field, log still written with remaining entries intact | Missing file's entry present with error note; other entry's forensics complete | `internal/ui/errorlog_test.go:186-188` — `Contains(got, missing) && Contains(got, "boom1")`; `:190-192` — `Contains(got, "unavailable")` (collection error recorded); `:194-198` — intact entry has `real`, `"boom2"`, `"2 bytes"`, `"61 62"` | ✅ PASS |
| **LOG-05** — WHEN batch has 0 failures THEN no log written | No file at `outDir/conversion-errors.log`; `("", nil)` returned | `internal/ui/errorlog_test.go:227-232` — `path != "" \|\| err != nil` → fail; `os.Stat(...)`: `!os.IsNotExist(err)` → fail | ✅ PASS |
| **LOG-06** — WHEN log written THEN failure dialog text ends with `Details saved to <full log path> — send this file when reporting bugs.` | Exact suffix incl. path, em dash, trailing period | `internal/ui/errorlog_test.go:313-316` — `strings.HasSuffix(got, "\nDetails saved to /out/conversion-errors.log — send this file when reporting bugs.")`; `:317-319` — `strings.HasPrefix(got, summaryText(s))`; wiring: `internal/ui/app.go:269` (`showSummary` renders `dialogText` in both dialog paths) | ✅ PASS |
| **LOG-07** — WHEN log write failed THEN dialog text ends with `Could not write log: <error>`; app continues (no crash, no retry) | Exact suffix `\nCould not write log: disk full`; no `Details saved`; error returned not panicked | `internal/ui/errorlog_test.go:328-331` — `strings.HasSuffix(got, "\nCould not write log: disk full")`; `:332-334` — `Contains(got, "Details saved")` → fail; error path data: `errorlog_test.go:270-276` — `err == nil` → fatal, `path != ""` → fail (unwritable outDir surfaces error, no crash) | ✅ PASS |
| **LOG-08** — WHEN batch has 0 failures THEN success dialog does not mention the log | Dialog text exactly `3 converted, 0 failed` (equality ⇒ no log mention possible) | `internal/ui/errorlog_test.go:340-342` — `got != "3 converted, 0 failed"` → fail | ✅ PASS |
| **LOG-09** — WHEN 0 failures AND log exists THEN delete it | Stale file gone after clean batch | `internal/ui/errorlog_test.go:247-249` — `!os.IsNotExist(err)` on stale path → fail; E2E: `errorlog_test.go:383-385` — same assertion after tapped all-valid batch | ✅ PASS |
| **LOG-10** — WHEN deletion target missing or deletion fails THEN proceed silently | `("", nil)`, no error surfaced | `internal/ui/errorlog_test.go:256-259` — `path != "" \|\| err != nil` → fail (missing-target branch; matches spec's Independent Test) | ✅ PASS (minor note 4) |

**Status**: ✅ All 10 ACs covered with spec-anchored assertions. 0 spec-precision gaps (spec defines precise outcomes throughout; assertions target them exactly).

**Minor notes** (none blocking):
1. LOG-01 "UTF-8": no explicit `utf8.ValidString` assertion; covered by construction (Go literals + `%02X`/`%d` formatting) and the `\r`-absence check covers the `\n`-endings assumption.
2. LOG-04: collection error asserted via one `Contains(got, "unavailable")`, not per-field (impl records Size/Modified/First-bytes individually — `internal/ui/errorlog.go:87-93`).
3. LOG-07 "no retry": not directly assertable; implementation has no retry path (single `os.WriteFile`, `internal/ui/errorlog.go:76`).
4. LOG-10 "deletion fails" disjunct untested (only "target missing" branch has a test); impl deliberately ignores `os.Remove` error (`internal/ui/errorlog.go:54`); the spec's own Independent Test requires only the missing-file branch.

---

## Edge Cases

- [x] File shorter than 32 bytes → exactly the bytes present: `internal/ui/errorlog_test.go:163-165` — `Contains(got, "61 62 FF\n")` (trailing `\n` proves no extra bytes).
- [x] File deleted before log write → stat/read error recorded, other entries unaffected: `internal/ui/errorlog_test.go:186-198` (LOG-04 test).
- [x] Every file fails → all listed, counts `0 converted, M failed`: `internal/ui/errorlog_test.go:296-303` — `Contains(got, "0 converted, 2 failed")` + both paths/reasons present.
- [x] Output folder unwritable after conversion → could-not-write dialog line: `internal/ui/errorlog_test.go:270-276` (write error surfaces) + `:328-331` (exact dialog line).
- [x] No VCS build info → version `unknown`: `internal/ui/errorlog_test.go:102-107` (empty settings and nil/false both → `unknown`).

---

## Gate Check

- **Gate command**: `go build ./... && go vet ./... && go test ./...` (no tasks.md; standard Go gate)
- **Result**: build OK, vet OK, **68 passed, 0 failed, 0 skipped** in 3 packages
- **Test count before feature** (fc63651): 37 test functions
- **Test count after feature** (fda3d04): 53 test functions
- **Delta**: +16 (all in `internal/ui/errorlog_test.go`; no tests deleted or weakened)
- **Skipped tests**: none
- **Failures**: none

---

## Discrimination Sensor

Scratch method: mutate one committed file in working tree → `go test ./internal/ui/ -count=1` → `git checkout -- <file>`. Baseline: 46 passed.

| Mutation | File:line | Description | Killed? |
| -------- | --------- | ----------- | ------- |
| 1 | `internal/ui/errorlog.go:54` | Removed `os.Remove(...)` stale-log side effect in `handleErrorLog` | ✅ Killed — `TestHandleErrorLogDeletesStaleOnSuccess` (errorlog_test.go:248), `TestConvertTapWritesAndCleansErrorLog` (errorlog_test.go:384) |
| 2 | `internal/ui/errorlog.go:113` | `%02X` → `%02x` (lowercase hex) in `hexPreview` | ✅ Killed — `TestWriteErrorLogForensics` (errorlog_test.go:143), `TestWriteErrorLogShortFile` (errorlog_test.go:164) |
| 3 | `internal/ui/app.go:263` | Dropped `logPath` from discovery line in `dialogText` | ✅ Killed — `TestDialogTextFailureWithLog` (errorlog_test.go:315) |

**Sensor depth**: lightweight (3 behavior-level mutations, one per story)
**Result**: 3/3 killed — PASS ✅
**Post-sensor state**: all mutations reverted; `git status --porcelain` shows only ` M .specs/features/error-log/spec.md` (the pre-existing uncommitted traceability update, untouched). Clean re-run: 46 passed.

---

## Code Quality

| Principle | Status |
| --------- | ------ |
| Minimum code (no features beyond spec; Out-of-Scope respected — no rotation, no crash logging, no new UI) | ✅ |
| Surgical changes (only 3 files, all required) | ✅ |
| No scope creep / no unrelated "improvements" | ✅ |
| Matches patterns (comment style, `Summary`/`FileError` reuse, `fyne.Do` threading discipline; log written off UI thread — app.go:217-227) | ✅ |
| Spec-anchored outcome check (asserted values match spec) | ✅ |
| Per-layer coverage (domain 1:1 ACs; UI wiring happy + edge + error via unit + tap-driven E2E) | ✅ |
| Every test maps to a spec AC / edge case (all 16 new tests carry AC-labelled comments; no unclaimed tests) | ✅ |
| Documented guidelines followed: none found in repo — strong defaults applied | ✅ |

---

## Requirement Traceability Update

| Requirement | Previous Status | New Status |
| ----------- | --------------- | ---------- |
| LOG-01…LOG-10 | Complete | ✅ Verified (recommended; spec.md left untouched — it carries an uncommitted status update) |

---

## Summary

**Overall**: ✅ Ready

**Spec-anchored check**: 10/10 ACs matched spec-defined outcomes; 0 spec-precision gaps; 5/5 edge cases covered
**Sensor**: 3/3 mutations killed
**Gate**: build + vet OK; 68 passed, 0 failed, 0 skipped (+16 new tests)

**What works**: forensic log write with exact header/counts/hex format; failure-gated write; stale-log cleanup on clean batch; exact dialog discovery and write-failure lines; graceful per-field forensics degradation; E2E tap-driven wiring proof.

**Issues found**: none blocking. Four minor observations (see notes under AC table); the only test-strengthening candidate is LOG-10's untested "deletion fails" branch — optional, low value on a single-user desktop app.

**Next steps**: none required; feature may be marked Verified.
