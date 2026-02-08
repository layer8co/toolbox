// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package oslite_test

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/layer8co/toolbox/os/oslite"
)

func TestRead(t *testing.T) {

	temp, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		temp.Close()
		os.Remove(temp.Name())
	}()

	text := "hello world"

	_, err = temp.WriteString(text)
	if err != nil {
		t.Fatal(err)
	}

	f, err := oslite.Open(temp.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	buf := make([]byte, 4)
	sb := new(strings.Builder)
	err = nil

	for err != io.EOF {
		var n int
		n, err = f.Read(buf)
		if err != io.EOF && err != nil {
			t.Fatal(err)
		}
		sb.Write(buf[:n])
	}

	want := text
	got := sb.String()

	if want != got {
		t.Fatalf("incorrect read: want %q, got %q", want, got)
	}
}

func BenchmarkRead(b *testing.B) {

	temp, err := os.CreateTemp("", "")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		temp.Close()
		os.Remove(temp.Name())
	}()

	_, err = temp.WriteString("hello world")
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {

		f, err := oslite.Open(temp.Name())
		if err != nil {
			b.Fatal(err)
		}

		buf := make([]byte, 64)

		_, err = f.Read(buf)
		if err != nil {
			b.Fatal(err)
		}

		f.Close()
	}
}

func BenchmarkReadStdlib(b *testing.B) {

	temp, err := os.CreateTemp("", "")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		temp.Close()
		os.Remove(temp.Name())
	}()

	_, err = temp.WriteString("hello world")
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {

		f, err := os.Open(temp.Name())
		if err != nil {
			b.Fatal(err)
		}

		buf := make([]byte, 64)

		_, err = f.Read(buf)
		if err != nil {
			b.Fatal(err)
		}

		f.Close()
	}
}
