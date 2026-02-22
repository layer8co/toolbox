// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package moreio

import "io"

// ErrorCapturingWriter is returned by [NewErrorCapturingWriter].
// See it's documentation for details.
type ErrorCapturingWriter struct {
	Err error

	w AdapterWriter
}

// NewErrorCapturingWriter wraps w so that the first write error encountered
// is stored in [ErrorCapturingWriter.Err] and subsequent writes are no-ops.
//
// It's useful for when many small writes are performed
// and handling the error on each write is overkill.
func NewErrorCapturingWriter(w io.Writer) *ErrorCapturingWriter {
	e := &ErrorCapturingWriter{}
	e.Reset(w)
	return e
}

func (e *ErrorCapturingWriter) Reset(w io.Writer) {
	e.w.Reset(w)
	e.Err = nil
}

func (e *ErrorCapturingWriter) Write(b []byte) (int, error) {
	if e.Err != nil {
		return 0, e.Err
	}
	n, err := e.w.Write(b)
	if err != nil {
		e.Err = err
	}
	return n, err
}

func (e *ErrorCapturingWriter) WriteString(s string) (int, error) {
	if e.Err != nil {
		return 0, e.Err
	}
	n, err := e.w.WriteString(s)
	if err != nil {
		e.Err = err
	}
	return n, err
}

func (e *ErrorCapturingWriter) WriteByte(c byte) error {
	if e.Err != nil {
		return e.Err
	}
	err := e.w.WriteByte(c)
	if err != nil {
		e.Err = err
	}
	return err
}

func (e *ErrorCapturingWriter) WriteRune(r rune) (int, error) {
	if e.Err != nil {
		return 0, e.Err
	}
	n, err := e.w.WriteRune(r)
	if err != nil {
		e.Err = err
	}
	return n, err
}
