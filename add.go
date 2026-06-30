package main

import (
	"fmt"
	"math"
)

// add returns the sum of two int64 integers, or an error if overflow occurs.
// Overflow detection uses pre-check: for same-sign operands, verify the
// result stays within int64 bounds before performing addition.
func add(a, b int64) (int64, error) {
	// Positive overflow: both positive, sum would exceed MaxInt64
	if a > 0 && b > 0 && a > math.MaxInt64-b {
		return 0, fmt.Errorf("integer overflow: %d + %d exceeds int64 range", a, b)
	}
	// Negative overflow: both negative, sum would underflow MinInt64
	if a < 0 && b < 0 && a < math.MinInt64-b {
		return 0, fmt.Errorf("integer overflow: %d + %d exceeds int64 range", a, b)
	}
	return a + b, nil
}
