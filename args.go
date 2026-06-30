package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

var errInputTooLong = errors.New("input too long, maximum 40 characters per argument")

// parseArg validates and parses a raw string as an int64 argument.
// Returns an error if the input is too long (>40 chars) or not a valid integer.
func parseArg(raw string) (int64, error) {
	if len(raw) > 40 {
		return 0, errInputTooLong
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%q is not a valid integer", raw)
	}
	return n, nil
}

// printError writes a formatted error message to stderr.
// All user-facing argument errors follow the unified template:
//
//	Error: <description>. Usage: add <int_a> <int_b>
//
// D-23: Flag parse errors are NOT wrapped in this template; they use native output.
func printError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s. Usage: add <int_a> <int_b>\n", msg)
}
