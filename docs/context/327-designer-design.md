# Design Document — add CLI (设计文档)

> **Version**: 1.0-final
> **Author**: Designer (finalized after multi-role kickoff, 5 rounds)
> **Date**: 2026-07-01
> **Status**: Frozen
> **Repository**: https://github.com/luqz/demo.git
> **Branch**: 28/delivery
> **Input Documents**: PRD v1.0 (Frozen), Context Audit Report, Decision Document v1.0-final, Tech Plan v1.0-final

---

## 0. Multi-Role Review Summary

### 0.1 Participating Roles and Rounds

| Role | Rounds | Core Contributions |
|------|--------|--------------------|
| ProductManager | 6 | PRD finalization, open question rulings, help text rollback, module path confirmation |
| Designer | 5 | Design document draft, interaction spec, output matrix, help text, error message templates |
| Architect | 6 | Technical review, overflow scheme selection, flag wiring, test strategy, Makefile, cross-platform |

### 0.2 Key PM Rulings (from PRD, Carried into Design)

| # | Ruling | PRD Source | Design Impact |
|---|--------|------------|---------------|
| 1 | `--help`/`--version` is P0, use `flag` package | PRD §2 US-05/US-06 | Introduce `flag` package; `flag.Usage` overridden with Designer template |
| 2 | Error messages in English | PRD Changelog | All error/usage prompts in English, unified format |
| 3 | Overflow detection mandatory | PRD §3.1, §3.2 | Use `int64` + overflow detection; no silent wrap-around |
| 4 | stdout uses `fmt.Println` | PRD Changelog | Pure digit + newline, no prefix or decoration |
| 5 | Exit codes simplified to 0/1 | PRD Changelog | Exit code 2 removed; all errors use exit code 1 |
| 6 | Main file ≤ 60 lines | PRD Changelog | Excluding blank lines and comments (AC-24) |
| 7 | Input length limit 40 characters | PRD Changelog | New validation step; over-length input gets dedicated error |
| 8 | Version default `"dev"`, ldflags injected | PRD Changelog | `var version = "dev"` + `-ldflags "-X main.version=v1.0.0"` |
| 9 | `flag.Usage` overrides to Designer template | PRD Changelog | Help text follows PRD Appendix D verbatim |
| 10 | Deferred piped stdin input | PRD D-11 | v1.0 accepts only command-line arguments |

### 0.3 Open Questions and Resolutions

| ID | Question | Raised By | Resolution | Resolved By |
|----|----------|-----------|------------|-------------|
| Q-01 | Input length check before or after `flag.Parse()`? | Designer (Round 0) | **After.** `flag.Parse()` consumes flags first; length check acts on `flag.Args()` | PM (Round 0) |
| Q-02 | Should `--` separator behavior for negative first argument be explicitly documented beyond help text? | Designer (Round 0) | **Already sufficient** in help text (Examples + Notes); no extra section needed | PM (Round 0) |
| Q-03 | Overflow detection Scheme A (sign consistency) vs Scheme B (boundary pre-check)? | Designer (Round 0) | **Scheme A** (check sign reversal after `a + b`). Rationale: fewer imports (`math` not needed), Go int64 wrap-around behavior is deterministic | Architect (Round 0) |

### 0.4 Additional Rulings Discovered During Discussion

| ID | Ruling | Raised By | Resolution | Resolved By |
|----|--------|-----------|------------|-------------|
| D-23 | Flag parse errors (`--unknown`) are NOT wrapped in `Error: ... Usage: ...` template | Architect (Round 0) | Flag errors differ semantically from positional argument errors; the `Error: ... Usage: ...` template does not apply | PM (Round 1) |
| D-24 | `-h`/`--help` and `-v`/`--version` MUST be dual-registered as independent flags pointing to the same variable | Architect (Round 1) | Help text lists both `-h, --help` and `-v, --version`; code must deliver both | PM (Round 2) |
| D-25 | `flag.CommandLine.SetOutput(io.Discard)` MUST be retained to prevent flag package double-output | Architect (Round 3) | Suppress `ContinueOnError` mode auto-output to stderr; let `main()` control all output | Architect (Round 3) |
| D-26 | Help text MUST revert verbatim to PRD Appendix D original text | Architect (Round 4) | Tech Plan §6.4 had custom wording ("Flag and positional can be mixed") which was misleading; PRD is Frozen and unmodifiable | PM (Round 4) |
| D-27 | `go.mod` module path is `github.com/luqz/demo`, NOT `module add` | Architect (Round 3) | `go install github.com/luqz/demo@latest` requires module declaration to match repository URL | PM (Round 4) |
| D-28 | `28/delivery` branch MUST contain `docs/prd.md`, `docs/tech-plan.md`, `docs/context-audit.md` | Architect (Round 4) | Coder's working branch must have implementation basis; do not fetch documents from other branches | PM (Round 5) |
| D-29 | Overflow detection scheme divergence must be recorded in Tech Plan §10.3 as traceability note | Architect (Round 4) | Rounds 0-2 recommended Scheme A; Tech Plan ultimately selected pre-check scheme; behavior is equivalent | PM (Round 5) |

---

## 1. Design Overview

### 1.1 Design Scope

`add` is a pure command-line tool (CLI). There is no GUI, no web interface, no TUI. The Designer's role focuses on: command format design, input/output specification, error message copy, help/version message copy, and user interaction flow.

What distinguishes Project 28 from Projects 3 and 27: first-time introduction of `--help`/`--version` flags (P0), overflow detection, argument count differentiation (missing vs. extra get distinct messages), and input length limits. These additions are derived directly from the PRD — nothing outside PRD scope is introduced.

### 1.2 Design Principles

| Principle | Description | Source |
|-----------|-------------|--------|
| Zero cognitive load | Users should be able to use the tool correctly without reading help. Command format is self-evident. | Designer |
| Least surprise | Behavior follows Unix tool conventions: normal output to stdout, errors to stderr, exit codes 0/1. | Designer |
| Err on the side of doing less | Do not guess user intent. Strictly validate input and give precise error messages. | Designer |
| Precise error messages | Distinguish five error scenarios: missing args, extra args, invalid format, overflow, over-length. Each gets a dedicated message. | Designer (from PRD AC) |
| Consistent format | All error output follows unified template: `Error: <description>. Usage: add <int_a> <int_b>` | PRD Appendix C |
| Safety over convenience | Overflow detection is mandatory — prefer reporting an error over returning an incorrect result. | PRD §3.1 |
| Help is design | `--help` output is the user's "first impression." Carefully designed help text reduces learning cost. | Designer |

### 1.3 Design Constraints

| Constraint | Source | Details |
|------------|--------|---------|
| Pure CLI, no GUI/Web/TUI | PRD §1.1 | All interaction through terminal text |
| Use `flag` package for flag parsing | PRD Changelog | `--help`/`-h`, `--version`/`-v` via `flag` package |
| `flag.Usage` overridden | PRD Appendix D | Help text format specified by Designer (verbatim PRD Appendix D) |
| No subcommands | PRD §1.1 | Single `add` command |
| Zero external dependencies | PRD §1.2 | Go standard library only |
| `main.go` ≤ 60 lines | PRD Changelog, §3.4 | Excluding blank lines and comments |
| Error messages in English | PRD Changelog | Target users are developers |
| Exit codes 0/1 only | PRD Changelog | Exit code 2 removed |
| stdout is result only | PRD §1.2 | `fmt.Println` outputs pure digit + newline |
| No ANSI colors | Inherited from Projects 3/27 | Plain text compatible with all terminals |
| No config files / environment variables | PRD §4 D-06 | All configuration via command-line arguments |
| No stdin input | PRD D-11 | v1.0 accepts only command-line arguments |

### 1.4 Target Users

| Role | Typical Scenario | Core Need |
|------|-----------------|-----------|
| Developer | Quick integer addition in terminal | Avoid opening calculator or Python REPL |
| Script writer | Integer addition in shell scripts | Reliable, pipeable CLI with capturable output |
| CI system | Numeric calculations in build pipelines (counters, offsets) | Predictable exit codes and stdout output |

---

## 2. CLI Execution Experience

### 2.1 Happy Path: Successful Calculation

User provides two valid integers. Tool outputs the result.

```
$ add 2 3
5
$ add 0 0
0
$ add -5 10
5
$ add -- -5 10
5
$ add -7 -3
-10
$ add 2147483647 1
2147483648
$ add 9223372036854775807 0
9223372036854775807
$ add -9223372036854775808 0
-9223372036854775808
```

Design notes:
- stdout outputs only the digit + newline (`fmt.Println`), no prefix, label, or decoration
- Output can be directly captured by shell: `result=$(add 2 3)`
- When the first argument is negative, use `--` separator (`add -- -5 10`); this is standard `flag` package behavior

### 2.2 Flag Path: Help (`--help` / `-h`)

```
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
```

Design notes:
- Output to stdout, exit code 0 (not an error path)
- `--help` and `-h` behave identically; both handled by `flag` package
- Help text is PRD Appendix D verbatim — no modification permitted (D-26)
- Text uses hardcoded `add` program name, not `os.Args[0]`, ensuring golden file reproducibility

### 2.3 Flag Path: Version (`--version` / `-v`)

```
$ add --version          # CI build (ldflags injected)
add v1.0.0
$ add --version          # Local go build (no ldflags)
add dev
```

Design notes:
- Output to stdout, exit code 0 (not an error path)
- `--version` and `-v` behave identically; both bound to the same `showVersion` boolean variable (D-24)
- Version string injected via ldflags: `-ldflags="-X main.version=v1.0.0"`
- Source file has `var version = "dev"` as default
- Format: `add <version>` (program name + space + version number), followed by newline

### 2.4 Failure Path: Missing Arguments

When positional arguments are fewer than two (0 or 1):

```
$ add
Error: missing arguments, need two integers. Usage: add <int_a> <int_b>
$ add 1
Error: missing arguments, need two integers. Usage: add <int_a> <int_b>
```

Design notes:
- Output to stderr, exit code 1
- `flag` package consumes all flags first (`--help`, `--version`, etc.), then check `flag.NArg()`
- Missing and extra arguments use different error messages (precise problem location)

### 2.5 Failure Path: Extra Arguments

When positional arguments exceed two:

```
$ add 1 2 3
Error: too many arguments, only two integers accepted. Usage: add <int_a> <int_b>
```

Design notes:
- Output to stderr, exit code 1
- Distinguished from missing-argument scenario to help user quickly locate the problem

### 2.6 Failure Path: Input Too Long

When any argument exceeds 40 characters:

```
$ add $(python3 -c 'print("1"*100)') 1
Error: input too long, maximum 40 characters per argument. Usage: add <int_a> <int_b>
```

Design notes:
- Output to stderr, exit code 1
- Length check runs before integer parsing (prevents passing super-long strings to ParseInt)
- Limit is 40 characters (explicit in PRD)
- Reports the first over-length argument (left-to-right validation)

### 2.7 Failure Path: Invalid Integer

When any argument cannot be parsed as int64 by `strconv.ParseInt`:

```
$ add abc 1
Error: "abc" is not a valid integer. Usage: add <int_a> <int_b>
$ add 1 xyz
Error: "xyz" is not a valid integer. Usage: add <int_a> <int_b>
$ add 1.5 2
Error: "1.5" is not a valid integer. Usage: add <int_a> <int_b>
```

Design notes:
- Output to stderr, exit code 1
- Invalid argument is wrapped in ASCII double quotes (`"abc"`), matching PRD Appendix C
- Uses Go `%q` formatting verb for safe handling (escapes control characters per security checklist S4)
- Reports only the first invalid argument (left-to-right validation)

### 2.8 Failure Path: Integer Overflow

When the sum of two int64 integers exceeds the int64 range:

```
$ add 9223372036854775807 1
Error: integer overflow: 9223372036854775807 + 1 exceeds int64 range. Usage: add <int_a> <int_b>
$ add -9223372036854775808 -1
Error: integer overflow: -9223372036854775808 + -1 exceeds int64 range. Usage: add <int_a> <int_b>
```

Design notes:
- Output to stderr, exit code 1
- Overflow detection is mandatory (explicit PRD requirement; differs from Projects 3/27)
- Detection algorithm: Scheme A — sign consistency check (see §3.4)
- Message includes specific operand values to help user understand overflow cause

### 2.9 Failure Path: Flag Parse Error

When an unknown flag is passed (`flag.ContinueOnError` mode):

```
$ add --unknown 1 2
flag provided but not defined: -unknown
```

Design notes:
- Flag package outputs error to stderr natively; exit code 1
- `flag.CommandLine.Init("add", flag.ContinueOnError)` — does not call `os.Exit`
- Flag errors are NOT wrapped in `Error: ... Usage: ...` template (D-23); they differ semantically from positional argument errors

### 2.10 Output Matrix

| Scenario | stdout | stderr | Exit Code |
|----------|--------|--------|-----------|
| Normal calculation | Result (pure digit + `\n`) | None | 0 |
| `--help` / `-h` | Help text (PRD Appendix D verbatim) | None | 0 |
| `--version` / `-v` | `add <version>\n` | None | 0 |
| Missing args (0 or 1) | None | `Error: missing arguments, need two integers. Usage: add <int_a> <int_b>\n` | 1 |
| Extra args (≥3) | None | `Error: too many arguments, only two integers accepted. Usage: add <int_a> <int_b>\n` | 1 |
| Input too long (>40 chars) | None | `Error: input too long, maximum 40 characters per argument. Usage: add <int_a> <int_b>\n` | 1 |
| Invalid integer | None | `Error: "<input>" is not a valid integer. Usage: add <int_a> <int_b>\n` | 1 |
| Integer overflow | None | `Error: integer overflow: <a> + <b> exceeds int64 range. Usage: add <int_a> <int_b>\n` | 1 |
| Flag parse error | None | flag package native output (e.g., `flag provided but not defined: -unknown\n`) | 1 |

### 2.11 Terminal Output Regions

| Region | Trigger | Stream | Content |
|--------|---------|--------|---------|
| Result output | Calculation success | stdout | Pure digit + `\n` |
| Help output | `--help`/`-h` | stdout | Structured help text |
| Version output | `--version`/`-v` | stdout | `add <version>\n` |
| Error output | Validation failure | stderr | `Error: <description>. Usage: ...` |

No separators, no titles, no footers, no summaries. Zero decoration.

### 2.12 ANSI Color Specification

Not used. The only output is a pure digit (stdout) or an error message (stderr). No color markup. Consistent with PRD "no colored output" convention.

---

## 3. Error Handling Design

### 3.1 Exit Code Semantics

| Exit Code | Meaning | Trigger Conditions |
|-----------|---------|--------------------|
| 0 | Success | Calculation complete (result to stdout), help output, version output |
| 1 | Input error | Missing arguments, extra arguments, input too long, invalid integer, overflow, flag parse error |

> Exit code 2 has been removed from the PRD; reserved for future runtime errors.

### 3.2 Error Message Template

All error messages follow a unified format:

```
Error: <specific description>. Usage: add <int_a> <int_b>\n
```

- Prefix `Error: ` (English)
- Specific description (English, distinguishes different error scenarios)
- Period separator
- `Usage: add <int_a> <int_b>` suffix (fixed, helps user correct input)
- Program name `add` is hardcoded (not `os.Args[0]`, ensuring golden file reproducibility)

### 3.3 Error Categories and Golden Strings

| # | Category | Error Message Template | Detection Method |
|---|----------|------------------------|------------------|
| 1 | Missing args | `Error: missing arguments, need two integers. Usage: add <int_a> <int_b>` | `flag.NArg() < 2` |
| 2 | Extra args | `Error: too many arguments, only two integers accepted. Usage: add <int_a> <int_b>` | `flag.NArg() > 2` |
| 3 | Input too long | `Error: input too long, maximum 40 characters per argument. Usage: add <int_a> <int_b>` | `len(arg) > 40` |
| 4 | Invalid integer | `Error: "<input>" is not a valid integer. Usage: add <int_a> <int_b>` | `strconv.ParseInt()` returns error |
| 5 | Integer overflow | `Error: integer overflow: <a> + <b> exceeds int64 range. Usage: add <int_a> <int_b>` | Sign consistency detection |

> Flag parse errors (unknown flags) use flag package native output — they are NOT wrapped in this template (D-23).

### 3.4 Validation Priority Chain

Validation order is fixed, following the principle "structure before content, safety before calculation":

1. **Flag parsing** — `flag.Parse()`, `flag.ContinueOnError` mode
2. **Argument count** — `flag.NArg()`, distinguish missing (<2) vs extra (>2) (Q-01: after flag parsing)
3. **Input length** — each argument ≤ 40 characters
4. **Integer parsing** — `strconv.ParseInt(arg, 10, 64)`, check error
5. **Overflow detection** — after addition, check if result sign is inconsistent with operand signs (Scheme A)
6. **Result output** — perform addition, `fmt.Println` the result

First error found at any step terminates execution.

### 3.5 Overflow Detection Scheme (Scheme A — Sign Consistency)

```go
func add(a, b int64) (int64, error) {
    result := a + b
    // Positive overflow: both operands positive, result negative
    if a > 0 && b > 0 && result < 0 {
        return 0, fmt.Errorf("integer overflow: %d + %d exceeds int64 range", a, b)
    }
    // Negative overflow: both operands negative, result positive
    if a < 0 && b < 0 && result > 0 {
        return 0, fmt.Errorf("integer overflow: %d + %d exceeds int64 range", a, b)
    }
    return result, nil
}
```

Rationale (Q-03): Go int64 wrap-around behavior is deterministic (two's complement). Checking sign reversal after addition requires no `math` import and is trivially provable by code review. The Architect's Tech Plan documents the alternative pre-check scheme in §10.2; both are behaviorally equivalent for v1.0.

---

## 4. User Interaction Flow

### 4.1 Main Flow Diagram

```
User enters command
    │
    ▼
┌─────────────────────────────────┐
│ Flag parsing                    │
│ flag.CommandLine.Init(...)      │
│ flag.Parse()                    │
│ flag.ContinueOnError            │
└───────────┬─────────────────────┘
            │
     ┌──────┴──────┐
     │ Failure      │ Success
     ▼              ▼
┌────────────┐  ┌─────────────────────────┐
│ stderr:    │  │ --help / -h triggered?   │
│ flag error  │  └───────────┬─────────────┘
│ exit 1     │              │
└────────────┘       ┌──────┴──────┐
                      │ Yes          │ No
                      ▼              ▼
              ┌────────────┐  ┌─────────────────────────┐
              │ stdout:    │  │ --version / -v triggered? │
              │ help text  │  └───────────┬─────────────┘
              │ exit 0     │              │
              └────────────┘       ┌──────┴──────┐
                                    │ Yes          │ No
                                    ▼              ▼
                            ┌────────────┐  ┌───────────────────────┐
                            │ stdout:    │  │ flag.NArg() == 2 ?    │
                            │ add <ver>  │  └───────────┬───────────┘
                            │ exit 0     │              │
                            └────────────┘       ┌──────┴──────┐
                                          ┌──────┴──────┐      │
                                          │ < 2 (missing)│      │ Yes
                                          ▼              ▼      │
                                   ┌────────────┐ ┌──────┐      │
                                   │ stderr:    │ │ > 2  │      │
                                   │ missing    │ │ extra │      │
                                   │ arguments  │ │      │      │
                                   │ exit 1     │ └──┬───┘      │
                                   └────────────┘    │         │
                                                     ▼         │
                                              ┌────────────┐   │
                                              │ stderr:    │   │
                                              │ too many   │   │
                                              │ arguments  │   │
                                              │ exit 1     │   │
                                              └────────────┘   │
                                                               ▼
                                              ┌─────────────────────────────┐
                                              │ len(argv[0]) ≤ 40 ?         │
                                              │ len(argv[1]) ≤ 40 ?         │
                                              └───────────┬─────────────────┘
                                                          │
                                                   ┌──────┴──────┐
                                                   │ No           │ Yes
                                                   ▼              ▼
                                            ┌────────────┐  ┌─────────────────────────┐
                                            │ stderr:    │  │ ParseInt(argv[0], 10, 64)│
                                            │ input too  │  └───────────┬─────────────┘
                                            │ long       │              │
                                            │ exit 1     │       ┌──────┴──────┐
                                            └────────────┘       │ Failure      │ Success
                                                                  ▼              ▼
                                                          ┌────────────┐  ┌─────────────────────────┐
                                                          │ stderr:    │  │ ParseInt(argv[1], 10, 64)│
                                                          │ not a valid│  └───────────┬─────────────┘
                                                          │ integer    │              │
                                                          │ exit 1     │       ┌──────┴──────┐
                                                          └────────────┘       │ Failure      │ Success
                                                                                ▼              ▼
                                                                        ┌────────────┐  ┌─────────────────────┐
                                                                        │ stderr:    │  │ Overflow check       │
                                                                        │ not a valid│  │ a + b exceeds int64? │
                                                                        │ integer    │  └───────────┬─────────┘
                                                                        │ exit 1     │              │
                                                                        └────────────┘       ┌──────┴──────┐
                                                                                              │ Yes          │ No
                                                                                              ▼              ▼
                                                                                      ┌────────────┐  ┌────────────┐
                                                                                      │ stderr:    │  │ stdout:    │
                                                                                      │ overflow   │  │ a + b      │
                                                                                      │ exit 1     │  │ exit 0     │
                                                                                      └────────────┘  └────────────┘
```

### 4.2 Exit Branch Count

| Exit Type | Count | Description |
|-----------|-------|-------------|
| Success (exit 0) | 3 | Normal calculation, help, version |
| Failure (exit 1) | 7 | Flag parse error, missing args, extra args, input too long, invalid integer (A), invalid integer (B), overflow |
| Total | 10 | Exit branches |

---

## 5. Edge Case Design

### 5.1 Integer Overflow Detection

Core distinction from Projects 3/27: Project 28 MUST detect overflow and report an error; it must not rely on Go's native silent wrap-around.

Final Scheme: Scheme A — Sign Consistency Detection (see §3.5 for code).

### 5.2 Boundary Value Behavior

| Input | Behavior | Notes |
|-------|----------|-------|
| `9223372036854775807 0` | Output `9223372036854775807`, exit 0 | int64 max + 0 = normal |
| `-9223372036854775808 0` | Output `-9223372036854775808`, exit 0 | int64 min + 0 = normal |
| `9223372036854775807 1` | Overflow error, exit 1 | Positive overflow |
| `-9223372036854775808 -1` | Overflow error, exit 1 | Negative overflow |
| `9223372036854775807 -1` | Output `9223372036854775806`, exit 0 | Boundary normal computation |
| `-9223372036854775808 1` | Output `-9223372036854775807`, exit 0 | Boundary normal computation |

### 5.3 Input Length Restrictions

| Input | len | Behavior |
|-------|-----|----------|
| `"42"` | 2 | Pass, continue to parse |
| `"-9223372036854775808"` | 20 | Pass (int64 minimum, 20 chars) |
| `"9223372036854775807"` | 19 | Pass (int64 maximum, 19 chars) |
| `"1" * 41` | 41 | Reject: `input too long` |
| `"1" * 40` | 40 | Pass (boundary value), continue to parse |

### 5.4 `--` Separator Behavior

When the first positional argument is negative, the `flag` package will misparse it as a flag. Solution:

```
$ add -5 10        # Error: flag package treats -5 as unknown flag
$ add -- -5 10     # Correct: after --, -5 is recognized as positional argument, output 5
```

This is standard `flag` package behavior and is documented in the help text. No custom flag parsing logic is needed to bypass this limitation.

### 5.5 Special Numeric Values

| Input | ParseInt Behavior | Design Behavior |
|-------|-------------------|-----------------|
| `+5` | Parses as `5` | Normal calculation |
| `-0` | Parses as `0` | Normal calculation |
| `0005` | Parses as `5` | Normal calculation |
| `0xFF` | Returns error | `Error: "0xFF" is not a valid integer` |
| `1e3` | Returns error | `Error: "1e3" is not a valid integer` |
| `""` (empty) | Returns error | `Error: "" is not a valid integer` |

### 5.6 Cross-Platform Newlines

- Go `fmt.Println` outputs CRLF (`\r\n`) on Windows, LF (`\n`) on Unix
- Pure digit output has no compatibility issues — shell `$()` capture automatically handles trailing newlines
- Help text and error messages also use `fmt.Println`/`fmt.Fprintln`, newline behavior is consistent
- Integration tests normalize newlines via `strings.ReplaceAll(got, "\r\n", "\n")`

---

## 6. Help Text Design

### 6.1 Complete Help Text (PRD Appendix D — Verbatim, Immutable per D-26)

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

### 6.2 Help Text Design Principles

| Principle | Description |
|-----------|-------------|
| Clear structure | Title → Usage → Arguments → Options → Examples → Notes; information density from high to low |
| Hardcoded program name | Uses `add`, not `os.Args[0]`, ensuring golden file reproducibility |
| Explicit argument range | `int64 range` phrasing; does not list specific numeric range (too verbose) |
| Practical examples | Three examples covering: normal input, negative as second argument, negative as first argument (`--` separator) |
| Concise notes | Only two key pitfalls; does not enumerate all PRD constraints |

### 6.3 Help Text Triggering

- `add --help` → stdout outputs help text, exit 0
- `add -h` → identical to `--help`
- `flag` package automatically handles `--help`/`-h` signals via `flag.ErrHelp`; `flag.Usage` function is set as fallback
- `flag.CommandLine.SetOutput(io.Discard)` prevents flag package from auto-outputting to stderr; `main()` explicitly calls `fmt.Print(helpText)` on `ErrHelp` (D-25)

---

## 7. Version Information Design

### 7.1 Version Output Format

```
add v1.0.0    # CI build (ldflags injected)
add dev       # Local build (default)
```

Format: `add <version>\n`

- Program name hardcoded `add`
- Space separator
- Version number: CI injected via ldflags (e.g., `v1.0.0`), local default `dev`
- Followed by newline

### 7.2 Implementation

```go
var version = "dev"              // package-level default

// In main():
// var showVersion bool
// flag.BoolVar(&showVersion, "v", false, "show version number")
// flag.BoolVar(&showVersion, "version", false, "show version number")
// ...
// if showVersion {
//     fmt.Printf("add %s\n", version)
//     os.Exit(0)
// }
```

Build command:
```
CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=v1.0.0" -trimpath -o add
```

Both `-v` and `--version` are bound to the same `showVersion` boolean variable (D-24).

---

## 8. Configuration Experience

### 8.1 Configuration Method

No configuration. The tool accepts no configuration files, environment variables, or additional command-line flags (beyond `--help`/`--version`). All behavior is controlled through positional arguments. Consistent with PRD §4 D-06 "no config files / environment variables."

### 8.2 Omitted Configuration Domains

This tool is a minimal CLI addition calculator with no configurable runtime behavior. The following configuration domains are explicitly not applicable:

| Domain | Omission Rationale |
|--------|--------------------|
| Environment variables | PRD §4 D-06 explicitly excluded |
| Configuration files | PRD §4 D-06 explicitly excluded |
| Additional flags | Only `--help`/`--version` (P0); no other flags |
| Output format switching | Sole output is pure digit; no format variants |
| Language switching | English only; PRD §4 D-09 no i18n |

---

## 9. Security UX Design

### 9.1 Omitted Security Domains

This tool processes pure integer arithmetic. It involves no passwords, keys, tokens, external network requests, or filesystem access. The following security domains are explicitly not applicable:

| Security Domain | Omission Rationale |
|-----------------|--------------------|
| Password/key lifecycle | No sensitive data |
| Response body truncation/sanitization | No external requests |
| Production environment protection | No production/test environment distinction |
| Output sanitization | Pure digit output; no injectable content |

### 9.2 Security-Related Design Decisions

| Decision | Details |
|----------|---------|
| Input length limit 40 characters | Prevents super-large strings from being passed to ParseInt, mitigating potential DoS (though Go stdlib is already robust) |
| Overflow detection | Ensures calculation correctness — no silent wrong results, critical for CI counters and shell scripts |
| `strconv.ParseInt` not `Atoi` | `ParseInt` explicitly limits bitSize=64, avoiding `int` type range differences on 32-bit platforms |
| `%q` formatting for error messages | `parseArg` uses `%q` to format invalid input; naturally escapes control characters (security checklist S4) |

---

## 10. Excluded Features

The following features are explicitly excluded by the PRD and not included in this design:

| Exclusion | Reason | PRD Ref |
|-----------|--------|---------|
| Floating-point/decimal addition | Integers only | D-01 |
| Interactive mode / REPL | Pure CLI, single invocation | D-02 |
| Subtraction, multiplication, division | Single responsibility | D-03 |
| Three or more operands | Strictly two integers | D-04 |
| Result formatting (commas, hex, etc.) | Pure decimal | D-05 |
| Config files / environment variables | Zero-config design | D-06 |
| Big integer / arbitrary precision beyond int64 | int64 limit | D-07 |
| File input via `--file` flag | Command-line args only | D-08 |
| Localization / i18n | English only | D-09 |
| Install scripts / package manager releases | Source only | D-10 |
| Piped stdin input | Deferred to v1.1 | D-11 |

---

## 11. Project Structure Design

### 11.1 Target Directory Tree

```
project-28/
├── main.go              # Entry: flag parsing, validation, output (≤ 60 lines)
├── add.go               # Pure function add(a, b int64) (int64, error) — overflow detection
├── help.go              # helpText constant (PRD Appendix D verbatim)
├── add_test.go          # Unit tests for add() (10 test cases)
├── main_test.go         # CLI integration tests via exec.Command (24 test cases)
├── go.mod               # module github.com/luqz/demo (go 1.21)
├── go.sum               # Checksums (auto-generated by go mod tidy)
├── Makefile             # Build targets: build, test, vet, cross, checksum, release, clean
├── .gitignore           # Ignores add, dist/, *.exe
├── .github/workflows/ci.yml  # CI pipeline
├── LICENSE              # MIT
├── README.md            # Project overview
└── docs/
    ├── prd.md           # PRD v1.0
    ├── design.md        # This document
    ├── tech-plan.md     # Architect's Tech Plan v1.0-final
    ├── context-audit.md # Context audit report
    └── reports/         # Role-specific reports directory
```

### 11.2 Key Structural Decisions

| Decision | Rationale |
|----------|-----------|
| `main.go` + `add.go` + `help.go` separation | `add.go` contains pure function for isolated testing; `help.go` keeps the large constant out of `main.go` to satisfy ≤ 60 line constraint |
| `add_test.go` independent | Unit tests test only the `add()` function; 10 test cases covering all overflow scenarios |
| `main_test.go` integration tests | Test complete CLI behavior via `exec.Command`, covering all PRD §3.1-3.3 ACs |
| Zero external dependencies | `go.mod` declares only `go 1.21`; no `require` block |
| Module path `github.com/luqz/demo` | Required for `go install github.com/luqz/demo@latest` (D-27) |
| Makefile | Encapsulates build/test/vet/cross/checksum/release/clean; PRD §E lists as P1 deliverable |

### 11.3 Differences from Projects 3 and 27

| Dimension | Project 3 | Project 27 | Project 28 |
|-----------|-----------|------------|------------|
| Source file structure | Single main.go | Single main.go | main.go + add.go + help.go |
| Test files | main_test.go | main_test.go | add_test.go + main_test.go |
| Makefile | None | None | Yes (P1 deliverable) |
| main.go line limit | Not specified | ≤ 50 | ≤ 60 (excluding blank lines and comments) |

---

## 12. Design Decision Log

| ID | Decision | Rationale | Source |
|----|----------|-----------|--------|
| D-01 | Use `flag` package for `--help`/`--version` | PRD US-05/US-06 upgraded to P0; Changelog explicitly requires `flag` package | PM Kickoff (PRD) |
| D-02 | `flag.ContinueOnError` mode | Does not call `os.Exit(2)`; unified exit codes 0/1 | PM Kickoff (PRD §3.5) |
| D-03 | `flag.Usage` overridden with Designer template | PRD Appendix D explicitly specifies help text format | PM Kickoff (PRD Appendix D) |
| D-04 | Version default `"dev"`, ldflags injected | PRD Changelog explicit requirement | PM Kickoff (PRD) |
| D-05 | Missing args and extra args use different error messages | More precise than Projects 3/27's single usage hint; helps user quickly locate problem | Designer (derived from PRD AC) |
| D-06 | All error messages carry `Usage: add <int_a> <int_b>` suffix | PRD Appendix C unified format | PM Kickoff (PRD §C) |
| D-07 | Program name hardcoded `add` in error messages | Ensures golden file reproducibility; avoids `go run` temp path interference | Designer (inherits P27 D-02) |
| D-08 | stdout outputs only the digit | Enables direct shell script capture | Designer (Unix convention) |
| D-09 | stderr carries all error output | Unix convention | Designer |
| D-10 | Exit codes 0/1 binary | PRD explicitly removed exit code 2 | PM Kickoff (PRD §3.5) |
| D-11 | Input length limit 40 characters | PRD Changelog explicit requirement | PM Kickoff (PRD) |
| D-12 | Overflow detection uses int64 + algorithmic check | PRD §3.1 explicitly requires covering both positive and negative overflow | PM Kickoff (PRD) |
| D-13 | `strconv.ParseInt` instead of `strconv.Atoi` | Needs bitSize=64 explicit semantics for overflow detection; Atoi returns `int` which is not platform-guaranteed | Designer |
| D-14 | English-only error messages | Target users are developers; English is universal | Designer (PRD D-09) |
| D-15 | No ANSI color output | Plain text compatible with all terminals | Designer |
| D-16 | `main.go` + `add.go` + `help.go` separation | Pure function `add()` tested independently; `helpText` in separate file to satisfy ≤ 60 line constraint | Designer |
| D-17 | `add_test.go` separate from `main_test.go` | Unit tests (10 cases) and integration tests (24 cases) have distinct responsibilities | Designer |
| D-18 | `--help` outputs to stdout with exit 0 | Help is not an error; Unix convention that `--help` returns 0 | Designer (Unix convention) |
| D-19 | `--version` outputs to stdout with exit 0 | Version is not an error | Designer (Unix convention) |
| D-20 | Version output format `add <version>` | Concise, consistent with program name style in help text | Designer |
| D-21 | CGO_ENABLED=0 static compilation | Ensures cross-platform portability | Designer (PRD §1.2) |
| D-22 | Makefile as P1 deliverable | PRD Appendix E explicitly listed | PRD §E |
| D-23 | Flag parse errors NOT wrapped in `Error: ... Usage: ...` template | Flag errors differ semantically from positional argument errors | Architect (Round 0) / PM (Round 1) |
| D-24 | `-h`/`--help` and `-v`/`--version` dual-registered | Code must deliver what the help text promises | Architect (Round 1) / PM (Round 2) |
| D-25 | `SetOutput(io.Discard)` retained | Prevents flag package double-output in `ContinueOnError` mode | Architect (Round 3) |
| D-26 | Help text reverts verbatim to PRD Appendix D | Tech Plan §6.4 custom wording was misleading; PRD is Frozen and unmodifiable | Architect (Round 4) / PM (Round 4) |
| D-27 | Module path `github.com/luqz/demo` | `go install` requires module path to match repository URL | Architect (Round 3) / PM (Round 4) |
| D-28 | `28/delivery` branch must contain `docs/prd.md`, `docs/tech-plan.md`, `docs/context-audit.md` | Coder needs implementation basis in working branch | Architect (Round 4) / PM (Round 5) |
| D-29 | Overflow detection scheme divergence recorded in Tech Plan §10.3 | Traceability for Scheme A (recommended Rounds 0-2) vs pre-check (Tech Plan final) | Architect (Round 4) / PM (Round 5) |
| D-30 | Error messages distinguish five scenarios | Missing/extra/format/overflow/over-length each get dedicated message | Designer (from PRD AC) |
| D-31 | Program name hardcoded `add` in all user-visible output | Not dependent on `os.Args[0]` | Designer |
| D-32 | `strconv.ParseInt(s, 10, 64)` not `strconv.Atoi` | Explicit bitSize=64 | Designer |
| D-33 | Only decimal integers accepted; `0x`/`0o`/`0b` prefixes treated as invalid | Consistent error handling for non-decimal formats | Designer |
| D-34 | stdout is pure digit + newline, no prefix/suffix/label | Pipeable output | Designer |
| D-35 | Cross-platform newlines normalized via `strings.ReplaceAll(got, "\r\n", "\n")` | Test portability | Designer |
| D-36 | `add()` function's error must include operand values (`a` and `b`) | Enables `main` to format the overflow error message | Designer |
| D-37 | Unit tests (`add_test.go`) test only mathematical correctness; formatting belongs to integration tests (`main_test.go`) | Separation of concerns | Designer |

---

## 13. Design Boundaries

### 13.1 Explicitly Not Designed (Delegated to Architect / Coder / DevOps)

| Area | Delegated To | Rationale |
|------|-------------|-----------|
| Overflow detection concrete implementation code | Architect | Designer specifies the algorithm (Scheme A); Architect selects exact code pattern |
| `flag` package `Init`/`SetOutput`/`Usage` wiring details | Architect | Designer specifies behavioral requirements; Architect produces technical wiring |
| Makefile target implementation | Architect | Designer specifies target names; Architect provides Makefile content |
| CI pipeline configuration (`.github/workflows/ci.yml`) | Architect / DevOps | Designer specifies what CI must verify; DevOps implements pipeline |
| Test table structure and naming conventions | Architect | Designer specifies coverage requirements; Architect defines test architecture |
| Build flags (`-ldflags`, `-trimpath`, `CGO_ENABLED=0`) | Architect | Specified in PRD; Architect confirms and documents |
| `TestMain` pattern implementation | Architect | Designer confirms test coverage split; Architect implements |
| Cross-platform binary naming convention (`dist/add-{os}-{arch}[.exe]`) | Architect | Build artifact organization |

### 13.2 Mapping to PRD Sections

| PRD Section | Design Section(s) | Description |
|-------------|-------------------|-------------|
| §1.1 Product Positioning | §1.1 Design Scope, §1.3 Design Constraints | CLI-only, no subcommands |
| §1.2 Core Value Proposition | §1.2 Design Principles | Speed, zero deps, correctness, cross-platform, no noise |
| §1.3 Target Users | §1.4 Target Users | Developer, script writer, CI system |
| §2 User Stories | §2 CLI Execution Experience | Each user story maps to a path in the experience journey |
| §3.1 Core Function AC | §2.1 Happy Path, §5.2 Boundary Values | 10 core ACs |
| §3.2 Error Path AC | §2.4-2.8 Failure Paths, §3.3 Error Categories | 7 error ACs |
| §3.3 Help & Version AC | §2.2-2.3 Flag Paths, §6, §7 | 5 help/version ACs |
| §3.4 Code & Build AC | §11 Project Structure | Structurally covered; implementation details delegated |
| §3.5 Exit Code Spec | §3.1 Exit Code Semantics, §2.10 Output Matrix | 0/1 only |
| §4 Out of Scope | §10 Excluded Features | All 11 exclusions |
| Appendix C Error Templates | §3.2-3.3 Error Messages | Golden strings |
| Appendix D Help Text | §6.1 Complete Help Text | Verbatim text |
| Appendix E File List | §11.1 Target Directory Tree | All files accounted for |

### 13.3 Coordination with Architect's Tech Plan

| Design Decision | Tech Plan Reference | Consistency |
|-----------------|---------------------|-------------|
| Scheme A overflow detection | Tech Plan §4.3 (pre-check scheme) | Behaviorally equivalent; divergence noted as D-29 |
| Help text content | Tech Plan §6.4 (Designer "final" text) | Overridden by D-26 — must use PRD Appendix D verbatim |
| Module path | Tech Plan §4.1 (`module add`) | Overridden by D-27 — must be `module github.com/luqz/demo` |
| File structure (add.go, help.go) | Tech Plan §4.1 | Consistent |
| flag wiring (SetOutput + ErrHelp) | Tech Plan §3.3 | Consistent |
| Test strategy (unit + integration) | Tech Plan §5 | Consistent |
| `--` separator test handling | Tech Plan §3.5 | Consistent; design specifies UX expectation |

---

## 14. Appendix

### Appendix A: Reference Projects

| Project | Type | Relation |
|---------|------|----------|
| Project 3 | Go CLI addition tool | Reference for structure; note: uses `strconv.Atoi`, does not detect overflow |
| Project 27 | Go CLI addition tool | Reference for PRD structure template |

### Appendix B: Error Message Golden Strings (from PRD Appendix C)

| # | Scenario | Exact stderr String |
|---|----------|---------------------|
| 1 | Missing arguments | `Error: missing arguments, need two integers. Usage: add <int_a> <int_b>\n` |
| 2 | Too many arguments | `Error: too many arguments, only two integers accepted. Usage: add <int_a> <int_b>\n` |
| 3 | Invalid integer | `Error: "<input>" is not a valid integer. Usage: add <int_a> <int_b>\n` |
| 4 | Integer overflow | `Error: integer overflow: <a> + <b> exceeds int64 range. Usage: add <int_a> <int_b>\n` |
| 5 | Input too long | `Error: input too long, maximum 40 characters per argument. Usage: add <int_a> <int_b>\n` |

### Appendix C: Comparison with Projects 3 and 27

Project 28 is the third delivery in the same domain (integer addition CLI tool). Key design differences:

| Dimension | Project 3 | Project 27 | Project 28 |
|-----------|-----------|------------|------------|
| `--help`/`-h` | Not implemented | Not implemented (PRD D-01 excluded) | Implemented (P0, flag package) |
| `--version`/`-v` | Not implemented | Not implemented | Implemented (P0, flag package + ldflags) |
| flag package | Not used (os.Args direct) | Not used (os.Args direct) | Used (flag package) |
| Overflow detection | Not detected (silent wrap) | Not detected (silent wrap) | Mandatory detection (positive + negative overflow) |
| Integer type | `int` (Atoi) | `int` (Atoi) | `int64` (ParseInt) |
| Error message granularity | None (single usage hint) | None (single usage hint) | 5 distinct categories |
| Program name | Hardcoded `add` | Static `<int> <int>` | Hardcoded `add` |
| Input length limit | None | None | 40 characters |
| stdout method | fmt.Println | fmt.Println | fmt.Println |
| Exit codes | 0/1 | 0/1 | 0/1 (code 2 removed) |
| Main line limit | Not specified | ≤ 50 | ≤ 60 |
| Source file structure | Single main.go | Single main.go | main.go + add.go + help.go |
| Test files | Single main_test.go | Single main_test.go | add_test.go + main_test.go |
| Makefile | None | None | Yes (P1) |
| Go version | go 1.26.4 | go 1.21 | go 1.21 |
| stdin piped input | None | None | Deferred to v1.1 (D-11) |

**Reused design elements** (same as Projects 3/27):
- stdout outputs only digit + newline
- stderr prefixed with `Error: `
- Exit codes 0/1
- English error messages
- No config files, environment variables, ANSI colors, interactive mode
- CGO_ENABLED=0 static compilation
- Left-to-right validation, reporting only the first error

---

> **Design document finalized.** All open questions (Q-01, Q-02, Q-03) resolved. All 20 decisions from the multi-role kickoff discussion recorded in §12 Decision Log. Help text locked to PRD Appendix D verbatim (per D-26). Module path locked to `github.com/luqz/demo` (per D-27). Ready for Coder implementation against this spec.
