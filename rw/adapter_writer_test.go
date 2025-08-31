// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package rw_test

import (
	"testing"

	"github.com/koonix/x/rw"
)

func BenchmarkAdapterWriter_Write(b *testing.B) {
	d := discard{}
	a := rw.NewAdapterWriter(d)
	v := []byte{'x'}
	b.ResetTimer()
	for b.Loop() {
		a.Write(v)
	}
}

func BenchmarkAdapterWriter_WriteString_Direct(b *testing.B) {
	d := discardFull{}
	a := rw.NewAdapterWriter(d)
	b.ResetTimer()
	for b.Loop() {
		a.WriteString("x")
	}
}

func BenchmarkAdapterWriter_WriteString_Convert(b *testing.B) {
	d := discard{}
	a := rw.NewAdapterWriter(d)
	b.ResetTimer()
	for b.Loop() {
		a.WriteString("x")
	}
}

func BenchmarkAdapterWriter_WriteByte_Direct(b *testing.B) {
	d := discardFull{}
	a := rw.NewAdapterWriter(d)
	b.ResetTimer()
	for b.Loop() {
		a.WriteByte('x')
	}
}

func BenchmarkAdapterWriter_WriteByte_Convert(b *testing.B) {
	d := discard{}
	a := rw.NewAdapterWriter(d)
	b.ResetTimer()
	for b.Loop() {
		a.WriteByte('x')
	}
}

func BenchmarkAdapterWriter_WriteRune_Direct(b *testing.B) {
	d := discardFull{}
	a := rw.NewAdapterWriter(d)
	b.ResetTimer()
	for b.Loop() {
		a.WriteRune('x')
	}
}

func BenchmarkAdapterWriter_WriteRune_Convert(b *testing.B) {
	d := discard{}
	a := rw.NewAdapterWriter(d)
	b.ResetTimer()
	for b.Loop() {
		a.WriteRune('x')
	}
}

type discard struct{}
type discardFull struct {
	discard
}

var (
	global_b []byte
	global_s string
	global_c byte
	global_r rune
)

func (discard) Write(b []byte) (int, error) {
	global_b = b
	return len(b), nil
}
func (discardFull) WriteString(s string) (int, error) {
	global_s = s
	return len(s), nil
}
func (discardFull) WriteByte(c byte) error {
	global_c = c
	return nil
}
func (discardFull) WriteRune(r rune) (int, error) {
	global_r = r
	return 0, nil
}
