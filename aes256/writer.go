// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package aes256

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"hash"
	"io"
	"io/fs"

	"github.com/koonix/x/must"
	"golang.org/x/crypto/argon2"
)

const aesKeyLen = 32

type writer struct {
	dest       io.Writer
	stream     cipher.Stream
	header     header
	hash       hash.Hash
	first      bool
	done       bool
	ciphertext []byte // Buffer used for encryption.
}

func NewWriter(dest io.Writer, password []byte) io.WriteCloser {

	w := &writer{
		dest:   dest,
		first:  true,
		header: newHeader(),
	}

	key := argon2.IDKey(
		password,
		w.header.ArgonSalt[:],
		w.header.ArgonTime,
		w.header.ArgonMemory,
		w.header.ArgonThreads,
		aesKeyLen,
	)

	block := must.Get(aes.NewCipher(key))
	w.stream = cipher.NewCTR(block, w.header.AesIV[:])
	w.hash = hmac.New(sha256.New, key)
	zero(key)

	w.header.writeTo(w.hash)

	return w
}

func (w *writer) Write(plaintext []byte) (int, error) {

	if w.done {
		return 0, fs.ErrClosed
	}

	err := w.initialize()
	if err != nil {
		return 0, err
	}

	if len(plaintext) == 0 {
		return 0, nil
	}

	if cap(w.ciphertext) < len(plaintext) {
		w.ciphertext = make([]byte, len(plaintext))
	}
	w.ciphertext = w.ciphertext[:len(plaintext)]

	w.stream.XORKeyStream(w.ciphertext, plaintext)
	w.hash.Write(w.ciphertext)
	return w.dest.Write(w.ciphertext)
}

func (w *writer) Close() error {
	w.done = true
	err := w.initialize()
	if err != nil {
		return err
	}
	_, err = w.dest.Write(w.hash.Sum(nil))
	return err
}

func (w *writer) initialize() error {
	if w.first {
		w.first = false
		err := w.header.writeTo(w.dest)
		if err != nil {
			return err
		}
	}
	return nil
}

func zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
