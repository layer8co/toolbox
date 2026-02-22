// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package moreio

// This is here because there is no io.RuneWriter in the stdlib:
// https://github.com/golang/go/issues/71027
type RuneWriter interface {
	WriteRune(r rune) (int, error)
}
