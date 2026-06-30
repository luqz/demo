# Add — Integer Addition CLI Tool: Final Documentation

> Project 28
> Repository: https://github.com/luqz/demo.git
> Branch: 28/delivery
> Version: 1.0.0
> Date: 2026-07-01

---

## Table of Contents

1. [Quick Start (README)](#1-quick-start-readme)
2. [CLI Interface Reference](#2-cli-interface-reference)
3. [Project Structure](#3-project-structure)
4. [Change Log & Release Notes](#4-change-log--release-notes)
5. [Build & Test](#5-build--test)

---

## 1. Quick Start (README)

### What is add?

`add` is a minimal, zero-dependency command-line tool that adds two integers and prints the result. It is written in Go, compiles to a single static binary, and runs on Linux, macOS, and Windows.

### Installation

Via `go install`:

```bash
go install github.com/luqz/demo@latest
```

From source:

```bash
git clone https://github.com/luqz/demo.git
cd demo
make build VERSION=v1.0.0
```

### Usage

```
add <int_a> <int_b>
```

### Examples

```bash
$ add 2 3
5

$ add -5 10
5

$ add -- -5 10
5

$ add -- -9223372036854775808 0
-9223372036854775808

$ add --help
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

$ add --version
add v1.0.0
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (result printed to stdout, or help/version output) |
| 1 | Input error (missing args, invalid integer, overflow, input too long) |

### Features

- Integer addition within int64 range [-9223372036854775808, 9223372036854775807]
- Overflow detection for both positive and negative overflow
- Zero external dependencies — single static binary, no runtime requirements
- Cross-platform: Linux (amd64/arm64), macOS (amd64/arm64), Windows (amd64)
- Input validation: length limits (40 chars), format checks, argument count checks
- Pristine stdout: only the result digit, suitable for shell piping

---

## 2. CLI Interface Reference

### 2.1 Command Synopsis

```
add [-h] [-v] <int_a> <int_b>
```

`add` is a single-command program; there are no subcommands.

### 2.2 Arguments

| Argument | Type | Range | Required | Description |
|----------|------|-------|----------|-------------|
| `<int_a>` | int64 | [-9223372036854775808, 9223372036854775807] | Yes | First integer operand |
| `<int_b>` | int64 | [-9223372036854775808, 9223372036854775807] | Yes | Second integer operand |

Per-argument input length limit: 40 characters. Inputs exceeding this limit produce an error before integer parsing begins.

When the first argument is negative (e.g., `-5`), use the `--` separator to prevent the Go `flag` package from interpreting it as a flag:

```bash
add -- -5 10
```

The second argument is never ambiguous, so no separator is needed when only the second argument is negative:

```bash
add 10 -5
```

### 2.3 Options

| Flag | Long Form | Description |
|------|-----------|-------------|
| `-h` | `--help` | Print help text to stdout, exit 0 |
| `-v` | `--version` | Print version number (`add <version>`) to stdout, exit 0 |

The short and long forms of each flag are functionally identical — both `-h` and `--help` trigger the same code path, as do `-v` and `--version`.

Flags are parsed before positional arguments. When a flag and positional arguments are mixed (e.g., `add -v 2 3`), the flag takes priority and the program exits after printing the version — the positional arguments are ignored.

### 2.4 stdout Output Specification

All successful output goes to stdout. The format varies by scenario:

| Scenario | stdout Format | Example |
|----------|--------------|---------|
| Normal addition | `<result>\n` (pure decimal integer + newline) | `5\n` |
| Help | Multi-line help text (see §2.7) | — |
| Version | `add <version>\n` | `add v1.0.0\n` |

Stdout contains exactly the result digit — no prefix, no label, no decoration. This allows direct capture in shell scripts:

```bash
result=$(add 2 3)   # result = "5"
```

### 2.5 stderr Output Specification

All error output goes to stderr. All user-facing argument errors follow a unified template:

```
Error: <description>. Usage: add <int_a> <int_b>\n
```

Exception: Flag parse errors (unknown flags like `add --unknown`) use the Go `flag` package's native output format (e.g., `flag provided but not defined: -unknown\n`) and are NOT wrapped in the template above. This distinction is by design — flag errors differ semantically from positional argument errors.

### 2.6 Error Scenarios — Golden Strings

The following error messages are exact golden strings verified by integration tests. They are the authoritative reference for downstream tooling that parses add's output.

| # | Trigger | stderr (exact) | Exit |
|---|---------|----------------|------|
| E1 | Zero or one positional arguments (`add`, `add 1`) | `Error: missing arguments, need two integers. Usage: add <int_a> <int_b>\n` | 1 |
| E2 | Three or more positional arguments (`add 1 2 3`) | `Error: too many arguments, only two integers accepted. Usage: add <int_a> <int_b>\n` | 1 |
| E3 | Argument exceeds 40 characters | `Error: input too long, maximum 40 characters per argument. Usage: add <int_a> <int_b>\n` | 1 |
| E4 | Argument is not a valid int64 (`add abc 1`, `add 1.5 2`) | `Error: "<input>" is not a valid integer. Usage: add <int_a> <int_b>\n` | 1 |
| E5 | Addition overflows int64 range | `Error: integer overflow: <a> + <b> exceeds int64 range. Usage: add <int_a> <int_b>\n` | 1 |
| E6 | Unknown flag (`add --unknown`) | `flag provided but not defined: -unknown\n` (flag package native) | 1 |

E4 uses Go's `%q` formatting verb for the invalid argument, which wraps it in ASCII double quotes and escapes any control characters. For example, `add abc 1` prints `"abc"` — not `abc`.

### 2.7 Help Text (Full Reference)

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

### 2.8 Validation Chain

Input is validated in a fixed order. The first failure halts execution immediately:

1. Flag parsing (`flag.Parse` in `ContinueOnError` mode)
2. `--help` / `-h` check → print help, exit 0
3. `--version` / `-v` check → print version, exit 0
4. Argument count (`flag.Args()` length): < 2 → missing; > 2 → too many
5. Input length: each argument ≤ 40 characters
6. Integer parsing: `strconv.ParseInt(raw, 10, 64)`
7. Overflow detection: pre-check before addition
8. Addition → print result

### 2.9 Overflow Detection

Addition uses int64 arithmetic with explicit overflow detection via pre-check. The algorithm verifies that the sum stays within int64 bounds before performing the addition:

- Positive overflow: both operands positive, and `a > math.MaxInt64 - b`
- Negative overflow: both operands negative, and `a < math.MinInt64 - b`

When overflow is detected, an error is returned — the program never silently wraps around. This is a deliberate divergence from Go's native int64 wrap-around behavior and from reference projects 3 and 27, which did not detect overflow.

### 2.10 Boundary Values

| Input | Result | Notes |
|-------|--------|-------|
| `9223372036854775807 0` | `9223372036854775807` | Max int64 + 0 = normal |
| `-9223372036854775808 0` | `-9223372036854775808` | Min int64 + 0 = normal |
| `9223372036854775807 1` | Overflow error | Positive overflow |
| `-9223372036854775808 -1` | Overflow error | Negative overflow |
| `9223372036854775807 -1` | `9223372036854775806` | Mixed sign, no overflow |
| `-9223372036854775808 1` | `-9223372036854775807` | Mixed sign, no overflow |

### 2.11 No REST / HTTP API

This is a pure command-line tool. There are no HTTP endpoints, no REST API, no gRPC services, no WebSocket endpoints. The only interface is the CLI as documented above. The project has no database, no schema, no migrations, no persistent state.

---

## 3. Project Structure

### 3.1 Directory Tree

```
project-28/
├── main.go              Entry point: flag parsing, argument validation, output
├── add.go               Core logic: add(a,b int64) with overflow detection
├── help.go              Help text constant (PRD Appendix D verbatim)
├── args.go              Argument parsing (parseArg) and error formatting (printError)
├── add_test.go          Unit tests for add() function (10 cases)
├── main_test.go         CLI integration tests via exec.Command (24 cases)
├── go.mod               Go module definition (github.com/luqz/demo, go 1.21)
├── Makefile             Build, test, vet, cross-compile, release targets
├── README.md            Project overview, installation, usage examples
├── LICENSE              MIT License
├── .gitignore           Ignores build outputs and IDE files
└── docs/
    ├── prd.md           Product Requirements Document v1.0 (Frozen)
    ├── design.md        CLI Interaction Design Document v1.0-final (Frozen)
    ├── tech-plan.md     Technical Implementation Plan v1.0-final (Frozen)
    ├── context-audit.md Repository Context Audit Report
    ├── final-docs.md    This document
    ├── context/         Role-specific working documents (drafts, intermediate outputs)
    └── reports/         Role-specific final reports (coder, reviewer, QA, integrator, security)
```

### 3.2 Source File Descriptions

#### main.go (68 lines)

The entry point. Responsibilities:

- Registers `-v`/`--version` flags (both bound to the same `showVersion` boolean)
- Configures `flag` package: `ContinueOnError` mode, stdout redirected to `io.Discard`
- Parses command-line flags and positional arguments
- Checks for `--help` (returns `flag.ErrHelp`) and `--version` flags
- Validates argument count (exactly 2 required)
- Calls `parseArg()` on each argument, then `add()`
- Prints result to stdout or error to stderr
- All exit paths: `os.Exit(0)` for success/help/version, `os.Exit(1)` for errors

Imports: `flag`, `fmt`, `io`, `os` (all standard library)

#### add.go (21 lines)

Pure function implementing int64 addition with overflow detection:

```
func add(a, b int64) (int64, error)
```

Overflow detection uses the pre-check method:
- Positive overflow: `a > 0 && b > 0 && a > math.MaxInt64 - b`
- Negative overflow: `a < 0 && b < 0 && a < math.MinInt64 - b`

Overflow errors include the specific operand values in the message.

Imports: `fmt`, `math` (all standard library)

#### help.go (26 lines)

Contains the `helpText` constant — the exact help text shown by `add --help`. This is PRD Appendix D verbatim (as mandated by ruling D-26). The text is stored in a separate file to keep `main.go` under the 60-line constraint.

#### args.go (33 lines)

Two functions:

`parseArg(raw string) (int64, error)`:
- Checks input length (≤ 40 characters, returns `errInputTooLong` if exceeded)
- Parses with `strconv.ParseInt(raw, 10, 64)`
- Returns `"<input>" is not a valid integer` on parse failure (using `%q` formatting)

`printError(msg string)`:
- Formats error to stderr: `Error: <msg>. Usage: add <int_a> <int_b>\n`
- Centralized error formatting ensures consistent golden strings

Imports: `errors`, `fmt`, `os`, `strconv` (all standard library)

### 3.3 Test Files

#### add_test.go (54 lines)

10 table-driven unit tests directly calling `add(a, b int64)`:

| Test | a | b | Expected |
|------|---|---|----------|
| NormalPositive | 2 | 3 | 5, nil |
| NormalNegative | -7 | -3 | -10, nil |
| MixedSign | -5 | 10 | 5, nil |
| PositiveOverflow | MaxInt64 | 1 | 0, error |
| NegativeOverflow | MinInt64 | -1 | 0, error |
| BoundaryExact | MaxInt64 | 0 | MaxInt64, nil |
| ZeroBoundary | 0 | 5 | 5, nil |
| NegBoundaryExact | MinInt64 | 0 | MinInt64, nil |
| MixedSignNoOverflow1 | MaxInt64 | -1 | MaxInt64-1, nil |
| MixedSignNoOverflow2 | MinInt64 | 1 | MinInt64+1, nil |

#### main_test.go (144 lines)

24 CLI integration tests using `exec.Command` against a compiled binary. Uses Go's `TestMain` pattern — builds the binary once, then runs all cases against it. Covers:

- All 22 acceptance criteria (AC-01 through AC-22) from the PRD
- 2 supplemental cases (T5, T6) for `--` separator and flag/positional mixing behavior
- For cases where PRD command strings are illustrative (negative first argument), actual test arguments use `--` separator with an explanatory comment

### 3.4 Dependencies

The project has zero external dependencies. All imports are from the Go standard library:

| Package | Used In | Purpose |
|---------|---------|---------|
| `flag` | main.go | CLI flag parsing (`-h`/`--help`, `-v`/`--version`) |
| `fmt` | main.go, add.go, args.go | Output formatting (stdout, stderr, error messages) |
| `io` | main.go | `io.Discard` to suppress flag package auto-output |
| `os` | main.go, args.go | `os.Exit()`, `os.Stderr`, `os.Args` |
| `strconv` | args.go | `ParseInt` for string-to-int64 conversion |
| `math` | add.go | `MaxInt64`, `MinInt64` for overflow pre-checks |
| `errors` | args.go | Sentinel error `errInputTooLong` |
| `testing` | add_test.go, main_test.go | Go test framework |
| `os/exec` | main_test.go | Running CLI binary in integration tests |
| `strings` | main_test.go | String matching in test assertions |

go.mod:
```
module github.com/luqz/demo
go 1.21
```

### 3.5 Build System

The Makefile provides 7 targets:

| Target | Command | Description |
|--------|---------|-------------|
| `build` | `CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(VERSION)" -trimpath -o add` | Build the binary. VERSION defaults to `dev` |
| `test` | `go test -v ./...` | Run all unit and integration tests |
| `vet` | `go vet ./...` | Run static analysis |
| `cross` | 5 GOOS/GOARCH combinations | Cross-compile to `dist/add-{os}-{arch}[.exe]` |
| `checksum` | `sha256sum` in dist/ | Generate SHA256 checksums |
| `release` | cross + checksum | Full release build |
| `clean` | Remove add and dist/ | Clean build artifacts |

Cross-compilation targets:
- `linux/amd64` → `dist/add-linux-amd64`
- `linux/arm64` → `dist/add-linux-arm64`
- `darwin/amd64` → `dist/add-darwin-amd64`
- `darwin/arm64` → `dist/add-darwin-arm64`
- `windows/amd64` → `dist/add-windows-amd64.exe`

All builds use `CGO_ENABLED=0` for fully static binaries with no system library dependencies.

---

## 4. Change Log & Release Notes

### v1.0.0 (2026-07-01) — Initial Release

First public release of the `add` CLI tool.

**Features:**
- Integer addition of two int64 operands
- Overflow detection for both positive and negative overflow
- `--help` / `-h` for usage instructions
- `--version` / `-v` for version number display
- Input validation: argument count, length limit (40 chars), integer format
- Five distinct error messages with exact golden strings
- Unified error format: `Error: <description>. Usage: add <int_a> <int_b>`
- Zero external dependencies — single static Go binary
- Cross-platform builds: Linux, macOS, Windows (amd64 + arm64)

**Key Design Decisions (from PRD & multi-role kickoff):**
- Overflow detection is mandatory (differs from earlier reference projects that silently wrapped around)
- Error messages in English, no localization
- Version injected via ldflags (`-X main.version`), default is `"dev"`
- `--` separator required when the first positional argument is negative (standard Go `flag` package behavior)
- Piped stdin input deferred to v1.1
- Exit codes simplified to 0/1 only

**Commit History:**
| Commit | Description |
|--------|-------------|
| `3af674e` | feat: implement add CLI tool — integer addition with overflow detection |
| `28372d4` | design.md added |
| `756b183` | docs: final tech-plan.md |
| `72eb6d0` | prd.md added |
| `24d1207` | docs: context audit and discovery reports |
| `1261e16` | Initial commit (empty repo) |

**Test Coverage:**
- 10 unit tests (`add_test.go`) covering normal, overflow, boundary, and mixed-sign cases
- 24 CLI integration tests (`main_test.go`) covering all 22 acceptance criteria plus 2 supplemental cases
- All 34 tests pass; 0 failures, 0 skipped
- `go vet ./...` clean

**Known Limitations:**
- No stdin/piped input support (planned for v1.1)
- No support for operands exceeding int64 range (big integers)
- No support for floating-point or decimal addition
- No subcommands (subtraction, multiplication, division not included)

---

## 5. Build & Test

### Prerequisites

- Go 1.21 or later

### Build

```bash
# Default build (version = "dev")
make build

# Build with specific version
make build VERSION=v1.0.0

# Cross-compile for all platforms
make cross

# Full release (cross-compile + checksums)
make release VERSION=v1.0.0
```

### Test

```bash
# Run all tests
make test

# Run only unit tests
go test -run TestAdd -v

# Run only CLI integration tests
go test -run TestCLI -v

# Static analysis
make vet
```

### Direct Go Commands

```bash
# Build without Makefile
CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=v1.0.0" -trimpath -o add

# Single-platform cross-compile
GOOS=linux GOARCH=amd64 go build -o add

# Install via go install
go install github.com/luqz/demo@latest
```

---

## Appendix A: Document Status

| Document | Path | Status | Version |
|----------|------|--------|---------|
| PRD | docs/prd.md | Frozen | v1.0 |
| Design | docs/design.md | Frozen | v1.0-final |
| Tech Plan | docs/tech-plan.md | Frozen | v1.0-final |
| Context Audit | docs/context-audit.md | Complete | — |
| Final Docs | docs/final-docs.md | Complete | v1.0.0 |
| README | README.md | Complete | — |
| LICENSE | LICENSE | Complete | MIT |

## Appendix B: Key Rulings Reference

| ID | Ruling | Source |
|----|--------|--------|
| D-11 | Piped stdin input deferred to v1.1 | PRD |
| D-23 | Flag parse errors NOT wrapped in `Error: ... Usage: ...` template | Design |
| D-24 | `-h`/`--help` and `-v`/`--version` dual-registered to same variables | Design |
| D-25 | `flag.CommandLine.SetOutput(io.Discard)` required to prevent double-output | Design |
| D-26 | Help text is PRD Appendix D verbatim — no modification permitted | Design |
| D-27 | Module path is `github.com/luqz/demo` | Design |
| Q-03 | Overflow: pre-check scheme selected (final implementation) | Design / Tech Plan |
