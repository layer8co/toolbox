// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package aes256

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"hash"
	"io"

	"github.com/koonix/x/must"
	"github.com/koonix/x/rw"
	"golang.org/x/crypto/argon2"
)

var ErrBadChecksum = errors.New("bad checksum")

const checksumLen = 32

type reader struct {
	src      *rw.FooterReader
	stream   cipher.Stream
	header   header
	hash     hash.Hash
	password []byte
	first    bool
}

func NewReader(src io.Reader, password []byte) io.ReadCloser {
	return &reader{
		src:      rw.NewFooterReader(src, make([]byte, checksumLen)),
		password: password,
		first:    true,
	}
}

func (r *reader) Read(b []byte) (int, error) {

	err := r.initialize()
	if err != nil {
		return 0, err
	}

	n, err := r.src.Read(b)
	if err != nil && err != io.EOF {
		return n, err
	}
	b = b[:n]

	if n > 0 {
		r.hash.Write(b)
		r.stream.XORKeyStream(b, b)
	}

	if err == io.EOF && !hmac.Equal(r.hash.Sum(nil), r.src.Footer()) {
		return n, ErrBadChecksum
	}

	return n, err
}

func (r *reader) Close() error {

	err := r.initialize()
	if err != nil {
		return err
	}

	if !hmac.Equal(r.hash.Sum(nil), r.src.Footer()) {
		return ErrBadChecksum
	}

	return nil
}

func (r *reader) initialize() error {

	if !r.first {
		return nil
	}
	r.first = false

	err := r.header.readFrom(r.src)
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	if err != nil {
		return err
	}

	key := argon2.IDKey(
		r.password,
		r.header.ArgonSalt[:],
		r.header.ArgonTime,
		r.header.ArgonMemory,
		r.header.ArgonThreads,
		aesKeyLen,
	)
	zero(r.password)
	r.password = r.password[:0:0]

	block := must.Get(aes.NewCipher(key))
	r.stream = cipher.NewCTR(block, r.header.AesIV[:])
	r.hash = hmac.New(sha256.New, key)
	zero(key)

	r.header.writeTo(r.hash)

	return nil
}
