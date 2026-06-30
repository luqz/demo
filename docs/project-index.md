# Project Index — add: Integer Addition CLI Tool

## 1. Project Overview (一句话概述)

`add` is a minimal, zero-dependency Go CLI tool that accepts two int64 integers and prints their sum — a single-purpose, pipeable terminal calculator for developers, script writers, and CI systems.

Repository: https://github.com/luqz/demo.git  
Branch: `28/delivery`  
Go version: 1.21  
License: MIT

---

## 2. Core Objectives & Acceptance Criteria

### Core Objectives

| # | Objective | Priority |
|---|-----------|----------|
| 1 | Correct integer addition of two int64 arguments, output to stdout | P0 |
| 2 | Negative number handling, including `--` separator for negative first arg | P0 |
| 3 | Distinguish error scenarios: missing args, extra args, invalid format, overflow, input-too-long | P0 |
| 4 | `--help` / `-h` and `--version` / `-v` via `flag` package | P0 |
| 5 | Cross-platform build (linux/darwin/windows × amd64/arm64), CGO_ENABLED=0 | P0 |
| 6 | Overflow detection (both positive and negative), no silent wrap-around | P0 |

### Acceptance Criteria (30 total, extracted from PRD §3)

- **AC-01 to AC-10** — Core computation (positive, negative, mixed, boundary, overflow)
- **AC-11 to AC-17** — Error paths (missing/extra args, invalid integer, float, input >40 chars)
- **AC-18 to AC-22** — Help and version output
- **AC-23 to AC-30** — Code/build quality (go.mod, line limit ≤60, vet, cross-compile, mod verify)

Exit codes: 0 (success/help/version), 1 (all input errors). Exit code 2 is reserved, not used.

### Explicitly Out of Scope

Floating-point, interactive mode, operations beyond addition, >2 operands, result formatting, config files, arbitrary precision, file input, i18n, package managers, stdin input (deferred to v1.1).

---

## 3. Tech Stack & Key Dependencies

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Language | Go 1.21 | Single binary, zero deps, cross-platform |
| CLI framework | stdlib `flag` (ContinueOnError mode) | PRD constraint, no 3rd-party deps |
| Integer parsing | `strconv.ParseInt(s, 10, 64)` | Explicit int64 bit size |
| Output | `fmt.Println` (stdout), `fmt.Fprintf(os.Stderr, ...)` (stderr) | PRD constraint |
| Testing | stdlib `testing` + `exec.Command` | No external test frameworks |
| Build | `CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=..." -trimpath` | Static, stripped, cross-platform |
| Version injection | `-ldflags -X main.version=v1.0.0`, default `"dev"` | PRD constraint |

Zero external dependencies — `go.mod` has no `require` block.

---

## 4. Main Modules / Directory Structure

```
project-28/
├── main.go              # Entry point: flag parsing, arg validation, output (≤60 lines)
├── add.go               # Pure function add(a, b int64) — overflow detection
├── help.go              # helpText constant (PRD Appendix D verbatim)
├── add_test.go          # Unit tests for add() — 10 cases
├── main_test.go         # CLI integration tests via exec.Command — 24 cases
├── go.mod               # module github.com/luqz/demo, go 1.21
├── go.sum               # Auto-generated checksums
├── Makefile             # build, test, vet, cross, checksum, release, clean
├── .gitignore           # Ignores add, dist/, *.exe
├── .github/workflows/ci.yml  # CI pipeline
├── LICENSE              # MIT
├── README.md            # Project overview, install, usage
└── docs/
    ├── prd.md           # PRD v1.0 (Frozen)
    ├── tech-plan.md     # Architect's Technical Plan v1.0-final
    ├── design.md        # Designer's Design Document v1.0-final
    ├── context-audit.md # Repository context audit
    └── reports/         # Role-specific reports
```

### Module Responsibilities

| File | Responsibility |
|------|---------------|
| `main.go` | Flag config (init), arg count validation, parseArg, result output, unified exit handling. Split from add/help logic to stay ≤60 lines. |
| `add.go` | Pure `add(a, b int64) (int64, error)` with overflow pre-check (a > MaxInt64-b / a < MinInt64-b). No side effects. |
| `help.go` | Single `helpText` constant — Designer's final help text. Moved out of main.go for line budget. |
| `add_test.go` | 10 table-driven unit tests (6 core ACs + 4 supplemental T1-T4). |
| `main_test.go` | 24 TestMain-pattern integration tests via `exec.Command`, covering AC-01 through AC-22 + T5-T6. |

---

## 5. Key Design Decisions

| ID | Decision | Rationale |
|----|----------|-----------|
| D-01 | Single `package main`, three files | Satisfies main.go ≤60 line constraint; add() testable in isolation |
| D-02 | Overflow: pre-check (a > MaxInt64-b) | Explicit, review-provable, no `math/bits` dependency |

Note: The Design document recommends Scheme A (post-addition sign consistency); the Tech Plan implements a pre-check scheme (a > MaxInt64 - b). Both are behaviorally equivalent for v1.0. See design.md §13.3 and tech-plan.md §10.3 for the divergence record.

| D-03 | `flag.ContinueOnError` + `SetOutput(io.Discard)` | Prevents flag pkg from auto-printing to stderr; `main()` controls all output |
| D-04 | `-h`/`--help` and `-v`/`--version` dual-registered | Both short and long forms bound to same boolean variable per D-24 |
| D-05 | Help text: PRD Appendix D verbatim | Design doc §6.1 rules that PRD Appendix D is the frozen, immutable help text (D-26) |
| D-06 | Module path: `github.com/luqz/demo` | Required for `go install`; overrides Tech Plan's `module add` (D-27) |
| D-07 | Error format: `Error: <desc>. Usage: add <int_a> <int_b>` | Unified template; hardcoded `add`, not `os.Args[0]` |
| D-08 | Flag parse errors NOT wrapped in Error template | Flag errors differ semantically from positional arg errors (D-23) |
| D-09 | `parseArg` uses `%q` formatting | Safe escaping of control characters in error messages (security S4) |
| D-10 | Input length check before ParseInt | 40-char limit prevents passing super-long strings to ParseInt |
| D-11 | Zero config | No config files, env vars, or additional flags beyond --help/--version |
| D-12 | TestMain pattern for integration tests | Build binary once, reuse across all CLI test cases |

---

## 6. How to Run / Test

### Build

```bash
# Local build (version defaults to "dev")
make build
# or directly:
CGO_ENABLED=0 go build -trimpath -o add

# Production build (version injected)
make build VERSION=v1.0.0
CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=v1.0.0" -trimpath -o add
```

### Run

```bash
./add 2 3        # → 5
./add -- -5 10   # → 5 (negative first arg needs -- separator)
./add --help     # → help text to stdout
./add --version  # → add dev or add v1.0.0
```

### Test

```bash
# All tests
make test
go test -v ./...

# Unit tests only (add.go)
go test -v -run TestAdd

# Integration tests only (full CLI)
go test -v -run TestCLI

# Static analysis
make vet
go vet ./...
```

### Cross-Compile

```bash
make cross
# Produces dist/add-{linux,darwin,windows}-{amd64,arm64}[.exe]
```

### Verification Checklist (from Tech Plan §8)

- `go mod verify` — exit 0
- `go vet ./...` — clean
- `grep "^go 1.21" go.mod` — matches
- `grep -cve '^\s*$' -e '^\s*//' main.go` — ≤60
- `make cross` — all 5 platforms pass
- `file add` — statically linked
- 10/10 unit tests PASS, 24/24 integration tests PASS

---

## 7. Related Documents

| Document | Path | Description |
|----------|------|-------------|
| PRD | [prd.md](./prd.md) | Product Requirements Document v1.0 (Frozen) — user stories, 30 ACs, exclusions |
| Tech Plan | [tech-plan.md](./tech-plan.md) | Architect's Technical Implementation Plan v1.0-final — architecture, AC→code mapping, test design, Makefile |
| Design | [design.md](./design.md) | Designer's Design Document v1.0-final — CLI UX, interaction flow, error messages, help text, decision log (37 decisions) |
| Context Audit | [context-audit.md](./context-audit.md) | Repository context audit report |
