# PRD v1.0 — Add: Integer Addition CLI Tool (Project 28)

> **Version**: 1.0
> **Author**: ProductManager (finalized after kickoff)
> **Date**: 2026-07-01
> **Status**: Frozen
> **Repository**: https://github.com/luqz/demo.git
> **Branch**: 28/delivery

> **Changelog**:
> v1.0 — Finalized after kickoff. Major changes from draft (v1.0-draft):
> - Removed stdin input (US-07 → D-11)
> - Added `--help`/`--version` via `flag` package (P0)
> - Changed error messages to English
> - Overflow detection mandatory (including negative overflow AC)
> - stdout uses `fmt.Println` (digit + newline)
> - Exit codes simplified to 0/1 (exit code 2 removed)
> - Main file ≤ 60 lines
> - Added input length limit (40 characters)
> - Removed `docs/requirements.md`
> - Version default `"dev"`, injected via `-ldflags`
> - `flag.Usage` overrides to match Designer's help template

---

## 1. Product Overview

### 1.1 Product Positioning

`add` is a minimal CLI tool that accepts two integer arguments and prints their sum. No subcommands, no interactive mode, no configuration files. Its sole purpose is to perform integer addition quickly in the terminal.

### 1.2 Core Value Proposition

| Dimension | Description |
|-----------|-------------|
| Speed | Binary startup + computation within 100 ms |
| Zero dependencies | Single binary, no runtime dependencies, copy and use |
| Correctness | Covers zero, negative, large integers, overflow, invalid input, etc. |
| Cross-platform | Linux/amd64, Linux/arm64, Darwin/amd64, Darwin/arm64, Windows/amd64 |
| No noise | Stdout contains only the result; stderr contains only error information |

### 1.3 Target Users

| Role | Scenario | Motivation |
|------|----------|------------|
| Developer | Quick integer addition in terminal | Avoids opening calculator or Python REPL |
| Script writer | Integer addition in shell scripts | Needs reliable, pipeable CLI |
| CI system | Numerical calculations in build pipelines (counters, offsets) | Needs predictable exit codes and stdout output |

### 1.4 Project Assumptions

| # | Assumption | If not met |
|---|------------|------------|
| A1 | User inputs are valid integers within int64 range | Tool returns clear error and non-zero exit |
| A2 | Users expect minimal CLI, no interactive mode | If feedback indicates need, v2 can extend |
| A3 | Go 1.21+ available in build environment | Downgrade `go.mod` if needed |
| A4 | Target platforms: Linux, Darwin, Windows × amd64 + arm64 | Additional platforms on demand |
| A5 | Repository serves only this tool | If future tools added, restructure as monorepo |

---

## 2. User Stories

| ID | User Story | Priority | Acceptance Criteria |
|----|------------|----------|---------------------|
| US-01 | As a developer, I want to input two integers and get their sum. | P0 | `add 2 3` outputs `5` with exit code 0 |
| US-02 | As a developer, I want negative numbers to be handled correctly. | P0 | `add -5 10` outputs `5` exit 0; negative first argument requires `--` (e.g., `add -- -5 10`) |
| US-03 | As a script writer, I want erroneous input to return non-zero exit. | P0 | Missing arguments → exit 1, stderr error message |
| US-04 | As a script writer, I want stdout to contain only the result for piping. | P0 | Stdout outputs only the number followed by newline, no prefix or decoration |
| US-05 | As a CI user, I want to see the program version. | P0 (upgraded) | `add --version` or `add -v` outputs version number |
| US-06 | As a new user, I want to see usage help. | P0 (upgraded) | `add --help` or `add -h` outputs usage instructions |

> **Note**: US-07 (stdin piped input) is **deferred** to v1.1 (see D-11).

---

## 3. Acceptance Criteria

### 3.1 Core Function Verification

- [ ] `add 2 3` → stdout: `5\n`, exit: `0`
- [ ] `add 0 0` → stdout: `0\n`, exit: `0`
- [ ] `add -5 10` → stdout: `5\n`, exit: `0` (second operand negative)
- [ ] `add -- -5 10` → stdout: `5\n`, exit: `0` (first operand negative with `--`)
- [ ] `add -7 -3` → stdout: `-10\n`, exit: `0`
- [ ] `add 2147483647 1` → stdout: `2147483648\n`, exit: `0` (large integer within int64)
- [ ] `add 9223372036854775807 1` → stderr contains overflow error, exit: `1`
- [ ] `add -9223372036854775808 -1` → stderr contains overflow error, exit: `1`
- [ ] `add 9223372036854775807 0` → stdout: `9223372036854775807\n`, exit: `0` (boundary normal)
- [ ] `add -9223372036854775808 0` → stdout: `-9223372036854775808\n`, exit: `0` (boundary normal)

### 3.2 Error Path Verification

- [ ] `add` (no arguments) → stderr: `Error: missing arguments, need two integers. Usage: add <int_a> <int_b>\n`, exit: `1`
- [ ] `add 1` (one argument) → stderr same as above, exit: `1`
- [ ] `add 1 2 3` (too many arguments) → stderr: `Error: too many arguments, only two integers accepted. Usage: add <int_a> <int_b>\n`, exit: `1`
- [ ] `add abc 1` (non-integer) → stderr: `Error: "abc" is not a valid integer. Usage: add <int_a> <int_b>\n`, exit: `1`
- [ ] `add 1 xyz` → stderr: `Error: "xyz" is not a valid integer. Usage: add <int_a> <int_b>\n`, exit: `1`
- [ ] `add 1.5 2` (float) → stderr: `Error: "1.5" is not a valid integer. Usage: add <int_a> <int_b>\n`, exit: `1`
- [ ] `add $(python3 -c 'print("1"*100)') 1` (input too long) → stderr: `Error: input too long, maximum 40 characters per argument. Usage: add <int_a> <int_b>\n`, exit: `1`

### 3.3 Help and Version Verification

- [ ] `add --help` → stdout includes specific help text (see Appendix D), exit: `0`
- [ ] `add -h` → identical to `--help`
- [ ] `add --version` (CI build) → stdout: `add v1.0.0\n`, exit: `0`
- [ ] `add -v` → identical to `--version`
- [ ] `add --version` (local `go build` without ldflags) → stdout: `add dev\n`, exit: `0`

### 3.4 Code and Build Verification

- [ ] `go.mod` declares `go 1.21`
- [ ] main source file ≤ 60 lines (excluding blank lines and comments)
- [ ] `add_test.go` exists with unit tests for `add(a, b int64)` covering 6 cases: normal positive, normal negative, mixed sign, positive overflow, negative overflow, boundary exact
- [ ] `main_test.go` (or combined) uses `exec.Command` for CLI integration tests covering all AC in 3.1–3.3
- [ ] `go vet ./...` clean
- [ ] Cross-platform compilation passes: `GOOS=linux GOARCH=amd64`, `GOOS=darwin GOARCH=amd64`, `GOOS=darwin GOARCH=arm64`, `GOOS=windows GOARCH=amd64`, `GOOS=linux GOARCH=arm64`
- [ ] `go mod verify` passes
- [ ] Build command: `CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=v1.0.0" -trimpath -o add`

### 3.5 Exit Code Specification

| Exit code | Meaning |
|-----------|---------|
| 0 | Success, result printed to stdout |
| 1 | User input error (missing arguments, format error, overflow, input too long, flag parse error with `flag.ContinueOnError`) |

> Exit code 2 is reserved for future runtime errors; not tested in v1.0.

---

## 4. Explicitly Out of Scope

| # | Exclusion | Decision | Rationale |
|---|-----------|----------|-----------|
| D-01 | Floating-point/decimal addition | Not done | Integers only; float input treated as format error |
| D-02 | Interactive mode / REPL | Not done | Pure CLI, one call one calculation |
| D-03 | Subtraction, multiplication, division | Not done | Single responsibility for addition |
| D-04 | Three or more operands | Not done | Strictly two integers; extra arguments produce error |
| D-05 | Result formatting (commas, hex, etc.) | Not done | Output pure decimal integer |
| D-06 | Configuration files / environment variables | Not done | Zero-config design |
| D-07 | Big integer / arbitrary precision beyond int64 | Not done | Limit to int64; overflow reports error |
| D-08 | Reading input from file via `--file` flag | Not done | Supports only command-line positional arguments (not file path as argument) |
| D-09 | Localization / i18n | Not done | All messages in English |
| D-10 | Install scripts / package manager releases | Not done | Provide source for `go build` / `go install` only |
| D-11 | **Piped stdin input** | Not done in v1.0 | Deferred to v1.1; only command-line arguments are accepted |

---

## 5. Conflict

No conflicts. The repository is clean (only `# demo` README). No existing code, documentation, or architecture to reconcile. All decisions made during kickoff are recorded here.

---

## 6. Appendix

### Appendix A: Reference Projects

| Project | Type | Relation |
|---------|------|----------|
| Project 3 | Go CLI addition tool | Reference for structure; note: Project 3 uses `strconv.Atoi` and does not detect overflow |
| Project 27 | Go CLI addition tool | Reference for PRD structure template |

### Appendix B: Glossary

| Term | Definition |
|------|------------|
| stdout | Standard output stream, for normal result |
| stderr | Standard error stream, for error and diagnostic messages |
| Exit code | Return value of program; 0 success, non-0 failure |
| int64 | 64-bit signed integer, range [-9223372036854775808, 9223372036854775807] |
| stdin | Standard input stream (not used in v1.0) |
| `--` separator | Used to separate flags from positional arguments; required when first argument is negative under `flag` package |

### Appendix C: Error Message Templates

All error messages follow this format (English):

```
Error: <description>. Usage: add <int_a> <int_b>
```

- Prefix `Error: ` followed by English description
- Argument names are quoted with ASCII double quotes (e.g., `"abc"`)
- Format consists of error explanation and usage hint

Specific golden strings (see AC 3.2 for exact mappings):

1. Missing arguments → `Error: missing arguments, need two integers. Usage: add <int_a> <int_b>`
2. Too many arguments → `Error: too many arguments, only two integers accepted. Usage: add <int_a> <int_b>`
3. Invalid integer → `Error: "<input>" is not a valid integer. Usage: add <int_a> <int_b>`
4. Integer overflow → `Error: integer overflow: <a> + <b> exceeds int64 range. Usage: add <int_a> <int_b>`
5. Input too long → `Error: input too long, maximum 40 characters per argument. Usage: add <int_a> <int_b>`

### Appendix D: Help Text (used in `flag.Usage` override)

```
add — integer addition tool

Usage:
  add [-h] [-v] <int_a> <int_b>

Arguments:
  <int_a>    first integer (int64 range)
  <int_b>    second integer (int64 range)

Options:
  -h, --help       show this help message
  -v, --version    show version number

Examples:
  add 2 3           outputs 5
  add -5 10         outputs 5
  add -- -5 10      use -- separator when the first argument is negative

Notes:
  Negative numbers must directly follow the minus sign (e.g., -5, not - 5).
  When the first argument is negative, use -- to separate flags from args.
```

### Appendix E: File List

| File | Description | Priority |
|------|-------------|----------|
| `main.go` | Entry point: flag parsing, input validation, addition, output | P0 |
| `add_test.go` | Unit tests for `add()` function (6 cases) | P0 |
| `main_test.go` | CLI integration tests via `exec.Command` covering all AC | P0 |
| `go.mod` | Go module definition (go 1.21) | P0 |
| `Makefile` | Build targets: `build`, `test`, `vet`, `cross`, `clean` | P1 |
| `README.md` | Project overview, install instructions, usage examples | P0 |
| `docs/prd.md` | This document | P0 |
| `docs/context-audit.md` | Repository context audit report | P0 |
| `docs/reports/` | Role-specific reports directory | P0 |
