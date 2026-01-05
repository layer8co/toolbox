// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package oslite_test

import (
	"os"
	"testing"

	"github.com/layer8co/toolbox/oslite"
)

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

func BenchmarkOsRead(b *testing.B) {

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
