// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package aes256_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/koonix/x/aes256"
	"github.com/koonix/x/must"
)

func TestEncryptDecrypt(t *testing.T) {

	input := []byte("hello world")
	password := []byte("mypass123")

	ciphertext := aes256.Encrypt(input, password)
	output, err := aes256.Decrypt(ciphertext, password)
	if err != nil {
		t.Fatalf("could not decrypt: %s", err)
	}

	if diff := cmp.Diff(input, output); diff != "" {
		t.Errorf("incorrect result (-want +got):\n%s", diff)
	}
}

func BenchmarkWrite(b *testing.B) {

	password := []byte("mypass123")
	plaintext := []byte("hello world")

	w := aes256.NewWriter(io.Discard, password)

	b.ResetTimer()

	for b.Loop() {
		w.Write(plaintext)
	}
}

func BenchmarkRead(b *testing.B) {

	password := []byte("mypass123")
	ciphertextBuf := new(bytes.Buffer)

	w := aes256.NewWriter(ciphertextBuf, password)
	w.Write([]byte("hello world"))

	rr := &repeatReader{
		b: ciphertextBuf.Bytes(),
	}

	r := aes256.NewReader(rr, password)
	readBuf := make([]byte, ciphertextBuf.Len())

	b.ResetTimer()

	for b.Loop() {
		must.Get(r.Read(readBuf))
	}
}

func BenchmarkEncrypt(b *testing.B) {

	password := []byte("mypass123")
	plaintext := []byte("hello world")

	b.ResetTimer()

	for b.Loop() {
		aes256.Encrypt(plaintext, password)
	}

	b.ReportMetric(
		float64(b.Elapsed().Milliseconds())/float64(b.N),
		"ms/op",
	)
}

func BenchmarkDecrypt(b *testing.B) {

	password := []byte("mypass123")
	plaintext := []byte("hello world")
	ciphertext := aes256.Encrypt(plaintext, password)

	b.ResetTimer()

	for b.Loop() {
		aes256.Decrypt(ciphertext, password)
	}

	b.ReportMetric(
		float64(b.Elapsed().Milliseconds())/float64(b.N),
		"ms/op",
	)
}

type repeatReader struct {
	b []byte
}

func (r *repeatReader) Read(b []byte) (int, error) {
	return copy(b, r.b), nil
}
