// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package moreio

import (
	"io"
	"unicode/utf8"
	"unsafe"
)

// AdapterWriter is returned by [NewAdapterWriter].
// See it's documentation for details.
type AdapterWriter struct {
	w   io.Writer
	s   io.StringWriter
	b   io.ByteWriter
	r   RuneWriter
	buf [utf8.UTFMax]byte
}

// NewAdapterWriter returns a writer
// that forwards the WriteString, WriteByte and WriteRune method calls
// if they're implemented by w,
// otherwise it implements them on top of w.Write().
func NewAdapterWriter(w io.Writer) *AdapterWriter {
	if a, ok := w.(*AdapterWriter); ok {
		return a
	}
	a := &AdapterWriter{}
	a.Reset(w)
	return a
}

func (a *AdapterWriter) Reset(w io.Writer) {
	if a2, ok := w.(*AdapterWriter); ok {
		*a = *a2
	}
	a.w = w
	a.s, _ = w.(io.StringWriter)
	a.b, _ = w.(io.ByteWriter)
	a.r, _ = w.(RuneWriter)
}

func (a *AdapterWriter) Write(b []byte) (int, error) {
	return a.w.Write(b)
}

func (a *AdapterWriter) WriteString(s string) (int, error) {
	if a.s != nil {
		return a.s.WriteString(s)
	}
	return a.w.Write(unsafe.Slice(unsafe.StringData(s), len(s)))
}

func (a *AdapterWriter) WriteByte(c byte) error {
	if a.b != nil {
		return a.b.WriteByte(c)
	}
	a.buf[0] = c
	_, err := a.w.Write(a.buf[:1])
	return err
}

func (a *AdapterWriter) WriteRune(r rune) (int, error) {
	if a.r != nil {
		return a.r.WriteRune(r)
	}
	n := utf8.EncodeRune(a.buf[:], r)
	return a.w.Write(a.buf[:n])
}
