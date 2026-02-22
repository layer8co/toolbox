// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package moreio

import (
	"io"
)

type AsyncBufio struct {
	w io.Writer
	b []byte
	n int
	// b1 []byte
	// b2 []byte
}

// func (a *AsyncBufio) Write(b []byte) (int, error) {
// 	for len(b) > 0 {
// 		n := copy(a.b[a.n:], b)
// 		b = b[:n]
// 		a.n += n
// 		if a.n == len(a.b) {
// 			// flush a.b
// 			a.n = 0
// 		}
// 	}
// }
