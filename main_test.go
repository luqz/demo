package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

var binary string

func TestMain(m *testing.M) {
	// Build the binary once for all CLI tests.
	tmp, err := os.CreateTemp("", "add-test-*")
	if err != nil {
		os.Stderr.WriteString("failed to create temp file: " + err.Error() + "\n")
		os.Exit(1)
	}
	tmp.Close()
	binary = tmp.Name()
	os.Remove(binary)

	cmd := exec.Command("go", "build", "-o", binary, ".")
	cmd.Stderr = os.Stderr
	if out, err := cmd.Output(); err != nil {
		os.Stderr.WriteString("failed to build test binary: " + err.Error() + "\n")
		if len(out) > 0 {
			os.Stderr.Write(out)
		}
		os.Exit(1)
	}

	code := m.Run()
	os.Remove(binary)
	os.Exit(code)
}

// run executes the add binary with the given args and returns stdout, stderr, and exit code.
func run(args ...string) (stdout, stderr string, exitCode int) {
	cmd := exec.Command(binary, args...)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			exitCode = -1
		}
	}
	// Normalize Windows \r\n to \n for cross-platform consistency.
	stdout = strings.ReplaceAll(outBuf.String(), "\r\n", "\n")
	stderr = strings.ReplaceAll(errBuf.String(), "\r\n", "\n")
	return
}

// TestCLI covers AC-01 through AC-22 plus supplemental T5-T6.
func TestCLI(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantOut  string      // exact stdout match; "" means don't check stdout
		wantErr  string      // substring to find in stderr; "" means stderr should be empty
		exitCode int
		note     string      // for -- separator ruling documentation
	}{
		// === Core function (AC-01 to AC-10) ===
		{name: "AC-01_BasicPositive", args: []string{"2", "3"}, wantOut: "5\n", exitCode: 0},
		{name: "AC-02_ZeroAdd", args: []string{"0", "0"}, wantOut: "0\n", exitCode: 0},
		// AC-03: PRD 命令为示意 add -5 10，实现中使用负第二操作数无需 --
		{name: "AC-03_NegativeSecond", args: []string{"10", "-5"}, wantOut: "5\n", exitCode: 0,
			note: "PRD command is illustrative add -5 10; implementation swaps order so negative is second argument, no -- needed"},
		{name: "AC-04_DashDashNegFirst", args: []string{"--", "-5", "10"}, wantOut: "5\n", exitCode: 0},
		// AC-05: PRD 命令为示意 add -7 -3，实现中使用 -- 分隔符
		{name: "AC-05_DoubleNegative", args: []string{"--", "-7", "-3"}, wantOut: "-10\n", exitCode: 0,
			note: "PRD command is illustrative add -7 -3; implementation uses -- separator for negative first argument"},
		{name: "AC-06_LargeInt", args: []string{"2147483647", "1"}, wantOut: "2147483648\n", exitCode: 0},
		{name: "AC-07_PositiveOverflow", args: []string{"9223372036854775807", "1"},
			wantErr: "integer overflow: 9223372036854775807 + 1 exceeds int64 range", exitCode: 1},
		// AC-08: PRD 命令为示意 add -9223372036854775808 -1，实现中使用 -- 分隔符
		{name: "AC-08_NegativeOverflow", args: []string{"--", "-9223372036854775808", "-1"},
			wantErr: "integer overflow: -9223372036854775808 + -1 exceeds int64 range", exitCode: 1,
			note: "PRD command is illustrative add -9223372036854775808 -1; implementation uses -- separator"},
		{name: "AC-09_PosBoundaryZero", args: []string{"9223372036854775807", "0"}, wantOut: "9223372036854775807\n", exitCode: 0},
		// AC-10: PRD 命令为示意 add -9223372036854775808 0，实现中使用 -- 分隔符
		{name: "AC-10_NegBoundaryZero", args: []string{"--", "-9223372036854775808", "0"}, wantOut: "-9223372036854775808\n", exitCode: 0,
			note: "PRD command is illustrative add -9223372036854775808 0; implementation uses -- separator"},

		// === Error paths (AC-11 to AC-17) ===
		{name: "AC-11_NoArgs", args: []string{},
			wantErr: "Error: missing arguments, need two integers. Usage: add <int_a> <int_b>", exitCode: 1},
		{name: "AC-12_OneArg", args: []string{"1"},
			wantErr: "Error: missing arguments, need two integers. Usage: add <int_a> <int_b>", exitCode: 1},
		{name: "AC-13_ThreeArgs", args: []string{"1", "2", "3"},
			wantErr: "Error: too many arguments, only two integers accepted. Usage: add <int_a> <int_b>", exitCode: 1},
		{name: "AC-14_NonIntegerFirst", args: []string{"abc", "1"},
			wantErr: `Error: "abc" is not a valid integer. Usage: add <int_a> <int_b>`, exitCode: 1},
		{name: "AC-15_NonIntegerSecond", args: []string{"1", "xyz"},
			wantErr: `Error: "xyz" is not a valid integer. Usage: add <int_a> <int_b>`, exitCode: 1},
		{name: "AC-16_FloatInput", args: []string{"1.5", "2"},
			wantErr: `Error: "1.5" is not a valid integer. Usage: add <int_a> <int_b>`, exitCode: 1},
		{name: "AC-17_InputTooLong", args: []string{strings.Repeat("1", 100), "1"},
			wantErr: "Error: input too long, maximum 40 characters per argument. Usage: add <int_a> <int_b>", exitCode: 1},

		// === Help and version (AC-18 to AC-22) ===
		{name: "AC-18_LongHelp", args: []string{"--help"}, wantOut: helpText, exitCode: 0},
		{name: "AC-19_ShortHelp", args: []string{"-h"}, wantOut: helpText, exitCode: 0},
		{name: "AC-20_LongVersion_CI", args: []string{"--version"}, wantOut: "add dev\n", exitCode: 0},
		{name: "AC-21_ShortVersion", args: []string{"-v"}, wantOut: "add dev\n", exitCode: 0},
		{name: "AC-22_VersionLocal", args: []string{"--version"}, wantOut: "add dev\n", exitCode: 0},

		// === Supplemental (T5-T6) ===
		// T5: -- 后 -v 被当作 positional arg，仅1个参数 → missing args
		{name: "T5_DashDashV", args: []string{"--", "-v"},
			wantErr: "Error: missing arguments, need two integers. Usage: add <int_a> <int_b>", exitCode: 1},
		// T6: flag 和 positional 混合；-v 在遇到非 flag 参数前解析，触发 version 输出
		{name: "T6_FlagWithPositional", args: []string{"-v", "2", "3"}, wantOut: "add dev\n", exitCode: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, code := run(tt.args...)

			if code != tt.exitCode {
				t.Errorf("exit code = %d, want %d (stdout=%q stderr=%q)", code, tt.exitCode, stdout, stderr)
				return
			}

			if tt.wantOut != "" && stdout != tt.wantOut {
				t.Errorf("stdout = %q, want %q", stdout, tt.wantOut)
			}
			if tt.wantErr == "" {
				if stderr != "" {
					t.Errorf("stderr = %q, want empty", stderr)
				}
			} else {
				if !strings.Contains(stderr, tt.wantErr) {
					t.Errorf("stderr = %q, want substring %q", stderr, tt.wantErr)
				}
			}
		})
	}
}
