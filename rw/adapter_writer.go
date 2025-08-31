// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package rw

import (
	"io"
	"unicode/utf8"
)

type AdapterWriter struct {
	w io.Writer
	s io.StringWriter
	b io.ByteWriter
	r runeWriter
}

func NewAdapterWriter(w io.Writer) *AdapterWriter {
	if a, ok := w.(*AdapterWriter); ok {
		return a
	}
	a := &AdapterWriter{w: w}
	a.s, _ = w.(io.StringWriter)
	a.b, _ = w.(io.ByteWriter)
	a.r, _ = w.(runeWriter)
	return a
}

func (a *AdapterWriter) Write(b []byte) (int, error) {
	return a.w.Write(b)
}

func (a *AdapterWriter) WriteString(s string) (int, error) {
	if a.s != nil {
		return a.s.WriteString(s)
	}
	return a.w.Write([]byte(s))
}

func (a *AdapterWriter) WriteByte(c byte) error {
	if a.b != nil {
		return a.b.WriteByte(c)
	}
	_, err := a.w.Write([]byte{c})
	return err
}

func (a *AdapterWriter) WriteRune(r rune) (int, error) {
	if a.r != nil {
		return a.r.WriteRune(r)
	}
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	return a.w.Write(buf[:n])
}

// There is no io.RuneWriter in the stdlib: https://github.com/golang/go/issues/71027
type runeWriter interface {
	WriteRune(r rune) (int, error)
}
