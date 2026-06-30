package main

import (
	"math"
	"strings"
	"testing"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		name    string
		a, b    int64
		want    int64
		wantErr bool
		errMsg  string // substring to look for in error; empty means no check
	}{
		// Core cases (PRD §3.1 AC-01,03,05,07,08,09)
		{name: "NormalPositive", a: 2, b: 3, want: 5},                                     // AC-01
		{name: "NormalNegative", a: -7, b: -3, want: -10},                                  // AC-05
		{name: "MixedSign", a: -5, b: 10, want: 5},                                         // AC-03
		{name: "PositiveOverflow", a: math.MaxInt64, b: 1, wantErr: true, errMsg: "integer overflow"},   // AC-07
		{name: "NegativeOverflow", a: math.MinInt64, b: -1, wantErr: true, errMsg: "integer overflow"},  // AC-08
		{name: "BoundaryExact", a: math.MaxInt64, b: 0, want: math.MaxInt64},               // AC-09

		// Supplemental cases (T1-T4 from multi-role consensus revision #6)
		{name: "ZeroBoundary", a: 0, b: 5, want: 5},                                        // AC-02 / T1: zero
		{name: "NegBoundaryExact", a: math.MinInt64, b: 0, want: math.MinInt64},             // AC-10
		{name: "MixedSignNoOverflow1", a: math.MaxInt64, b: -1, want: math.MaxInt64 - 1},    // T3
		{name: "MixedSignNoOverflow2", a: math.MinInt64, b: 1, want: math.MinInt64 + 1},     // T4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := add(tt.a, tt.b)
			if tt.wantErr {
				if err == nil {
					t.Errorf("add(%d, %d) expected error, got nil", tt.a, tt.b)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("add(%d, %d) error = %q, want substring %q", tt.a, tt.b, err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("add(%d, %d) unexpected error: %v", tt.a, tt.b, err)
					return
				}
				if got != tt.want {
					t.Errorf("add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
				}
			}
		})
	}
}
