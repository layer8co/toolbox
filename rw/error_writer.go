// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package rw

import (
	"io"
)

type ErrorWriter struct {
	W   *AdapterWriter
	Err error
}

func NewErrorWriter(w io.Writer) *ErrorWriter {
	return &ErrorWriter{
		W: NewAdapterWriter(w),
	}
}

func (w *ErrorWriter) Write(b []byte) (int, error) {
	if w.Err != nil {
		return 0, w.Err
	}
	n, err := w.W.Write(b)
	if err != nil {
		w.Err = err
	}
	return n, err
}

func (w *ErrorWriter) WriteString(s string) (int, error) {
	if w.Err != nil {
		return 0, w.Err
	}
	n, err := w.W.WriteString(s)
	if err != nil {
		w.Err = err
	}
	return n, err
}

func (w *ErrorWriter) WriteByte(c byte) error {
	if w.Err != nil {
		return w.Err
	}
	err := w.W.WriteByte(c)
	if err != nil {
		w.Err = err
	}
	return err
}

func (w *ErrorWriter) WriteRune(r rune) (int, error) {
	if w.Err != nil {
		return 0, w.Err
	}
	n, err := w.W.WriteRune(r)
	if err != nil {
		w.Err = err
	}
	return n, err
}
