# add — Integer Addition CLI Tool

A minimal command-line tool that adds two integers and prints the result.

## Installation

```bash
go install github.com/luqz/demo@latest
```

Or build from source:

```bash
git clone https://github.com/luqz/demo.git
cd demo
make build VERSION=v1.0.0
```

## Usage

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

$ add --version
add dev
```

### Help

```bash
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

## Exit Codes

| Code | Meaning |
|------|---------|
| 0    | Success (result, help, or version output) |
| 1    | Input error (missing args, invalid integer, overflow, etc.) |

## Features

- Integer addition within int64 range
- Overflow detection for both positive and negative overflow
- Zero external dependencies — single static binary
- Cross-platform: Linux, macOS, Windows (amd64, arm64)
- Input validation: length limits, format checks, argument count checks

## Build Targets

```bash
make build          # Build the add binary
make test           # Run all tests
make vet            # Run go vet
make cross          # Cross-compile for all platforms (output in dist/)
make checksum       # Generate SHA256 checksums
make release        # Cross-compile + checksums
make clean          # Remove build artifacts
```

## Build with specific version

```bash
make build VERSION=v1.0.0
CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=v1.0.0" -trimpath -o add
```

## License

MIT — see [LICENSE](LICENSE)
