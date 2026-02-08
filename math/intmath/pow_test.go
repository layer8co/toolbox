// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package intmath_test

import (
	"fmt"
	"math"
	"runtime"
	"testing"

	"github.com/layer8co/toolbox/math/intmath"
	"golang.org/x/exp/constraints"
)

type powTest struct {
	line         int
	v            any // []Type{base, exponent, want}
	wantOverflow ovf
}

type ovf bool

const (
	ok       ovf = true
	overflow ovf = false
)

func (o ovf) String() string {
	if o {
		return "overflow"
	} else {
		return "ok"
	}
}

func TestPow(t *testing.T) {

	tests := []powTest{

		// Format: {line(), []int64{base, exponent, want}, wantOverflow}

		// int64 tests
		{line(), []int64{0, 10, 0}, ok},
		{line(), []int64{0, 0, 1}, ok},
		{line(), []int64{1, 0, 1}, ok},
		{line(), []int64{-1, 0, 1}, ok},
		{line(), []int64{10, 0, 1}, ok},
		{line(), []int64{-10, 0, 1}, ok},
		{line(), []int64{1, 1, 1}, ok},
		{line(), []int64{2, 1, 2}, ok},
		{line(), []int64{1, 10, 1}, ok},
		{line(), []int64{2, 10, 1024}, ok},
		{line(), []int64{-2, 10, 1024}, ok},
		{line(), []int64{-10, 2, 100}, ok},
		{line(), []int64{-10, 3, -1000}, ok},
		{line(), []int64{2, 63, 0}, overflow},
		{line(), []int64{-2, 63, math.MinInt64}, ok},
		{line(), []int64{2, 64, 0}, overflow},
		{line(), []int64{-2, 64, 0}, overflow},
		{line(), []int64{2, 100, 0}, overflow},
		{line(), []int64{-2, 100, 0}, overflow},
		{line(), []int64{math.MaxInt64, 1, math.MaxInt64}, ok},
		{line(), []int64{math.MinInt64, 1, math.MinInt64}, ok},
		{line(), []int64{math.MaxInt64, 2, 0}, overflow},
		{line(), []int64{math.MinInt64, 2, 0}, overflow},

		// uint64 tests
		{line(), []uint64{0, 10, 0}, ok},
		{line(), []uint64{0, 0, 1}, ok},
		{line(), []uint64{1, 0, 1}, ok},
		{line(), []uint64{10, 0, 1}, ok},
		{line(), []uint64{1, 1, 1}, ok},
		{line(), []uint64{1, 10, 1}, ok},
		{line(), []uint64{2, 10, 1024}, ok},
		{line(), []uint64{2, 63, math.MaxInt64 + 1}, ok},
		{line(), []uint64{2, 64, 0}, overflow},
		{line(), []uint64{math.MaxUint64, 1, math.MaxUint64}, ok},
		{line(), []uint64{math.MaxUint64, 2, 0}, overflow},

		// int32 tests
		{line(), []int32{2, 10, 1024}, ok},
		{line(), []int32{-3, 3, -27}, ok},
		{line(), []int32{2, 31, 0}, overflow},
		{line(), []int32{-2, 31, math.MinInt32}, ok},
		{line(), []int32{2, 32, 0}, overflow},
		{line(), []int32{-2, 32, 0}, overflow},
		{line(), []int32{math.MaxInt32, 1, math.MaxInt32}, ok},
		{line(), []int32{math.MinInt32, 1, math.MinInt32}, ok},
		{line(), []int32{math.MaxInt32, 2, 0}, overflow},
		{line(), []int32{math.MinInt32, 2, 0}, overflow},

		// uint32 tests
		{line(), []uint32{2, 10, 1024}, ok},
		{line(), []uint32{2, 31, math.MaxInt32 + 1}, ok},
		{line(), []uint32{2, 32, 0}, overflow},
		{line(), []uint32{math.MaxUint32, 1, math.MaxUint32}, ok},
		{line(), []uint32{math.MaxUint32, 2, 0}, overflow},

		// int16 tests
		{line(), []int16{2, 10, 1024}, ok},
		{line(), []int16{-3, 3, -27}, ok},
		{line(), []int16{2, 15, 0}, overflow},
		{line(), []int16{-2, 15, math.MinInt16}, ok},
		{line(), []int16{2, 16, 0}, overflow},
		{line(), []int16{-2, 16, 0}, overflow},
		{line(), []int16{math.MaxInt16, 1, math.MaxInt16}, ok},
		{line(), []int16{math.MinInt16, 1, math.MinInt16}, ok},
		{line(), []int16{math.MaxInt16, 2, 0}, overflow},
		{line(), []int16{math.MinInt16, 2, 0}, overflow},

		// uint16 tests
		{line(), []uint16{2, 10, 1024}, ok},
		{line(), []uint16{2, 15, math.MaxInt16 + 1}, ok},
		{line(), []uint16{2, 16, 0}, overflow},
		{line(), []uint16{math.MaxUint16, 1, math.MaxUint16}, ok},
		{line(), []uint16{math.MaxUint16, 2, 0}, overflow},

		// int8 tests
		{line(), []int8{2, 3, 8}, ok},
		{line(), []int8{-3, 3, -27}, ok},
		{line(), []int8{2, 7, 0}, overflow},
		{line(), []int8{-2, 7, math.MinInt8}, ok},
		{line(), []int8{2, 8, 0}, overflow},
		{line(), []int8{-2, 8, 0}, overflow},
		{line(), []int8{math.MaxInt8, 1, math.MaxInt8}, ok},
		{line(), []int8{math.MinInt8, 1, math.MinInt8}, ok},
		{line(), []int8{math.MaxInt8, 2, 0}, overflow},
		{line(), []int8{math.MinInt8, 2, 0}, overflow},

		// uint8 tests
		{line(), []uint8{2, 3, 8}, ok},
		{line(), []uint8{2, 7, math.MaxInt8 + 1}, ok},
		{line(), []uint8{2, 8, 0}, overflow},
		{line(), []uint8{math.MaxUint8, 1, math.MaxUint8}, ok},
		{line(), []uint8{math.MaxUint8, 2, 0}, overflow},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test%d-line%d", i, test.line), func(t *testing.T) {
			switch test.v.(type) {
			case []int64:
				testPow[[]int64](t, test)
			case []uint64:
				testPow[[]uint64](t, test)
			case []int32:
				testPow[[]int32](t, test)
			case []uint32:
				testPow[[]uint32](t, test)
			case []int16:
				testPow[[]int16](t, test)
			case []uint16:
				testPow[[]uint16](t, test)
			case []int8:
				testPow[[]int8](t, test)
			case []uint8:
				testPow[[]uint8](t, test)
			default:
				t.Fatalf("unsupported type %T", test.v)
			}
		})
	}
}

func testPow[S ~[]E, E constraints.Integer](t *testing.T, test powTest) {
	t.Helper()
	base := test.v.(S)[0]
	exp := test.v.(S)[1]
	want := test.v.(S)[2]
	got, gotOverflow := intmath.Pow(base, exp)
	if want != got || test.wantOverflow != ovf(gotOverflow) {
		t.Errorf(
			"%d^%d: want %d %s, got %d %s",
			base, exp, want, test.wantOverflow, got, ovf(gotOverflow),
		)
	}
}

func BenchmarkPow(b *testing.B) {
	for b.Loop() {
		intmath.Pow[int64](-2, 63)
	}
}

func BenchmarkPowStdlib(b *testing.B) {
	for b.Loop() {
		math.Pow(-2, 63)
	}
}

// line returns the line number where it's called.
func line() int {
	_, _, line, _ := runtime.Caller(1)
	return line
}
