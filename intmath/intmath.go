// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package intmath

import (
	"golang.org/x/exp/constraints"
)

// Abs returns the absolute value of x.
func Abs[T constraints.Integer](x T) T {
	if x >= 0 {
		return x
	}
	return -x
}

// GCD returns the greatest common divisor of x and y.
func GCD[T constraints.Integer](x, y T) T {
	for y != 0 {
		x, y = y, x%y
	}
	return x
}

// Pow returns base^exp, true.
// If the result overflows, it returns 0, false.
func Pow[T constraints.Integer](base, exp T) (result T, ok bool) {

	if exp < 0 {
		panic("intmath.Pow: negative exponent not supported")
	}

	result = 1

	for exp > 0 {
		if exp&1 == 1 {
			prevResult := result
			result *= base
			if base != 0 && result/base != prevResult {
				return 0, false
			}
		}

		exp >>= 1
		if exp == 0 {
			break
		}

		prevBase := base
		base *= base
		if prevBase != 0 && base/prevBase != prevBase {
			return 0, false
		}
	}

	return result, true
}

// MustPow is similar to [Pow], but it panics on overflow.
func MustPow[T constraints.Integer](base, exp T) T {
	result, ok := Pow(base, exp)
	if !ok {
		panic("intmath.MustPow: integer overflow")
	}
	return result
}
