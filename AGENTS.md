# AGENTS.md

This file provides guidance to LLM's when working with code in this repository.

## Project Overview

MSN-Converter is a simple Golang program that converts Microsoft MSN Messenger chat logs from XML format to a more human readable text format.

**Key Features:**
- The user can chose a lot of XML files and bulk convert then.
- The user can choose the folder where the converted files will be saved.

The GUI is built with [Fyne](https://github.com/fyne-io/fyne).

## Project Structure

Single-binary Fyne desktop app. No `cmd/` layout (only one binary); app code lives under `internal/` so nothing external can import it.

```
msn-converter/
├── go.mod
├── main.go                    # thin entry: build Fyne app, hand off to ui
├── internal/
│   ├── converter/
│   │   ├── converter.go       # Convert(xml) → txt; returns errors, never panics
│   │   ├── converter_test.go
│   │   ├── model.go           # Log/Message/User XML structs
│   │   └── output.txt.tmpl    # go:embed target (no runtime template file)
│   └── ui/
│       ├── app.go             # window, layout, wiring
│       ├── filelist.go        # de-duplicated file queue widget
│       └── batch.go           # off-UI-goroutine run + progress
└── testdata/                  # sample .xml fixtures for tests
    └── example.xml
```

Rules:
- `main.go` stays at the repo root — no `cmd/msn-converter/` until a second binary exists.
- The `converter` package owns the XML structs (the old top-level `models/` folds into it).
- The output template is embedded via `//go:embed output.txt.tmpl` so the built `.exe` is self-contained.
- Runtime paths (input files, output folder) come from the GUI, not hardcoded dirs. No `files/` scaffolding.
- Use `path/filepath` for all path building (Windows target).

## Building

Use the `Makefile`:

| Command | Purpose |
|---------|---------|
| `make build` | Native binary for the host OS |
| `make run` | Build + run locally |
| `make test` | Run all tests |
| `make tidy` | Sync `go.mod`/`go.sum` |
| `make windows` | Cross-build Windows `.exe` via `fyne-cross` (needs Docker running) |
| `make windows-mingw` | Cross-build Windows `.exe` on host via `mingw-w64` (needs `brew install mingw-w64` + `Icon.png` in root) |
| `make clean` | Remove build artifacts |

Fyne uses CGO (OpenGL/GLFW), so `GOOS=windows go build` alone does **not** work — a Windows C toolchain is required. `make windows` (Docker) is the reliable path; `make windows-mingw` avoids Docker but needs host mingw + an `Icon.png`.

## Instructions

### 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

### 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

### 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

### 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.