// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package aes256_test

import (
	"bytes"
	"testing"

	"github.com/koonix/x/aes256"
)

func Test(t *testing.T) {
	pass := []byte("pass123")
	text := []byte("hello world")
	ad := []byte("some text")
	ciphertext := aes256.Encrypt(pass, text, ad)
	gotText, err := aes256.Decrypt(pass, ciphertext, ad)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotText, text) {
		t.Fatalf("incorrect gotText; want %q, got %q", text, gotText)
	}
}

func TestIncorrectCipherText(t *testing.T) {
	pass := []byte("pass123")
	text := []byte("hello world")
	ad := []byte("some text")
	ciphertext := aes256.Encrypt(pass, text, ad)
	ciphertext[len(ciphertext)-1]++
	_, err := aes256.Decrypt(pass, ciphertext, ad)
	if err == nil {
		t.Fatal("want authentication error, got nil")
	}
}

func TestIncorrectAD(t *testing.T) {
	pass := []byte("pass123")
	text := []byte("hello world")
	ad1 := []byte("some text")
	ad2 := []byte("some other text")
	ciphertext := aes256.Encrypt(pass, text, ad1)
	_, err := aes256.Decrypt(pass, ciphertext, ad2)
	if err == nil {
		t.Fatal("want authentication error, got nil")
	}
}

func BenchmarkSecretBox(b *testing.B) {
	pass := []byte("pass123")
	text := []byte("hello world")
	for b.Loop() {
		ciphertext := aes256.Encrypt(pass, text, nil)
		gotText, err := aes256.Decrypt(pass, ciphertext, nil)
		if err != nil {
			b.Fatal(err)
		}
		if !bytes.Equal(gotText, text) {
			b.Fatalf("incorrect gotText; want %q, got %q", text, gotText)
		}
	}
}
