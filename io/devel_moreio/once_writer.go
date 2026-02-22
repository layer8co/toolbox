// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package moreio

import (
	"io"
	"sync"
)

type MutexWriter struct {
	Mu     *sync.Mutex
	Locked bool

	w AdapterWriter
}

// NewMutexWriter creates a writer that locks the mutex
// before the first write operation after creation
// or [MutexWriter.Locked] being set to false.
//
// This is useful for synchronizing writes to a writer that is written to
// through interfaces like [bufio.Writer] where normally you'd have to
// lock the mutex before the first write operation of the bufio,
// but that would be wasteful since it might be a while until the bufio
// actually started writing to the downstream writer;
// using this writer, however, the locking can be postponed
// until the actual writes begin.
func NewMutexWriter(w io.Writer, mu *sync.Mutex) *MutexWriter {
	m := &MutexWriter{}
	m.Reset(w, mu)
	return m
}

func (m *MutexWriter) Reset(w io.Writer, mu *sync.Mutex) {
	m.w.Reset(w)
	m.Mu = mu
	m.Locked = false
}

func (m *MutexWriter) Write(b []byte) (int, error) {
	m.lock()
	return m.w.Write(b)
}

func (m *MutexWriter) WriteString(s string) (int, error) {
	m.lock()
	return m.w.WriteString(s)
}

func (m *MutexWriter) WriteByte(c byte) error {
	m.lock()
	return m.w.WriteByte(c)
}

func (m *MutexWriter) WriteRune(r rune) (int, error) {
	m.lock()
	return m.w.WriteRune(r)
}

func (m *MutexWriter) lock() {
	if !m.Locked {
		m.Locked = true
		m.Mu.Lock()
	}
}
