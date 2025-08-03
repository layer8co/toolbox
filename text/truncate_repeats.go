// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package text

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"unicode/utf8"
)

// FilterFunc transforms consecutive chunks of a data stream.
type FilterFunc func([]byte) []byte

// TruncateRuneRepeats returns a filter that collapses sequences of
// more than n consecutive c runes down to n.
func TruncateRuneRepeats(c rune, n int) FilterFunc {
	re := regexp.MustCompile(fmt.Sprintf("%c{%d,}", c, n+1))
	replace := bytes.Repeat([]byte(string(c)), n)
	trailing := 0
	return func(b []byte) []byte {
		b = keepFromLeft(b, c, n-trailing)
		b = re.ReplaceAllLiteral(b, replace)
		trailing = countRight(b, c)
		return b
	}
}

func FilterWriter(w io.Writer, filter FilterFunc) io.Writer {
	return writer{f: func(b []byte) (int, error) {
		_, err := w.Write(filter(b))
		return len(b), err
	}}
}

func FilterReader(r io.Reader, filter FilterFunc) io.Reader {
	return reader{f: func(b []byte) (int, error) {
		_, err := r.Read(b)
		return copy(b, filter(b)), err
	}}
}

type writer struct{ f func(b []byte) (int, error) }
type reader struct{ f func(b []byte) (int, error) }

func (w writer) Write(b []byte) (int, error) { return w.f(b) }
func (r reader) Read(b []byte) (int, error)  { return r.f(b) }

// countRight returns the number of consecutive occurences of rune c
// at the end of b.
func countRight(b []byte, c rune) (n int) {
	for len(b) > 0 {
		r, size := utf8.DecodeLastRune(b)
		if r != c {
			break
		}
		n++
		b = b[:len(b)-size]
	}
	return n
}

// keepFromLeft keeps at most n consecutive leading occurences of rune c
// from the beginning of b.
func keepFromLeft(b []byte, c rune, n int) []byte {
	b2 := b
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		if r != c {
			break
		}
		if n > 0 {
			n--
		} else {
			b2 = b2[size:]
		}
		b = b[size:]
	}
	return b2
}
