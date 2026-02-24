// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package streamcrypt

import (
	"crypto/cipher"
	"crypto/sha3"
	"errors"
	"fmt"
	"io"

	"github.com/layer8co/toolbox/io/moreio"
	"golang.org/x/crypto/argon2"
)

const checksumLen = 32

var (
	ErrBadChecksum = errors.New("bad checksum")
	ErrClosed      = errors.New("already closed")
)

// PasswordFunc is used by [NewDecryptor].
// See it's documentation for details.
//
// The returned []byte is zeroed after use,
// so return a copy of it if it's in use elsewhere.
type PasswordFunc func() ([]byte, error)

// Decryptor is returned by [NewDecryptor].
// See it's documentation for details.
//
// Decryptor implements [io.ReadCloser].
type Decryptor struct {
	src       *moreio.FooterReader
	stream    cipher.Stream
	header    header
	hash      *sha3.SHAKE
	passFunc  PasswordFunc
	firstTime bool
	closed    bool
}

// NewDecryptor returns a [Decryptor]
// which is an [io.ReadCloser]
// that reads ciphertext from src and retrieves the plaintext.
//
// The encryption password is retrieved using passFunc during the first read.
// The []byte that passFunc returns is zeroed after use,
// so return a copy of it if it's in use elsewhere.
//
// The authentication of the ciphertext is checked
// upon reaching EOF or calling [Decryptor.Close].
// Therefore, calling Close after reaching EOF is unnecessary.
//
// After either reaching EOF or calling Close,
// calls to Read will result in an [ErrClosed] error,
// and calls to Close will be a no-op.
//
// The following options can be used to configure the decryption behavior:
//   - [WithArgonTimeMax] (default: 10)
//   - [WithArgonMemoryMax] (default: 64*1024)
//   - [WithArgonThreadsMax] (default: 64)
func NewDecryptor(
	src io.Reader,
	passFunc PasswordFunc,
	options ...Option,
) *Decryptor {
	return &Decryptor{
		src:       moreio.NewFooterReader(src, make([]byte, checksumLen)),
		passFunc:  passFunc,
		firstTime: true,
		header:    newHeaderForDecryptor(getConfig(options)),
	}
}

func (d *Decryptor) Read(b []byte) (int, error) {

	if d.closed {
		return 0, ErrClosed
	}

	err := d.readHeader()
	if err != nil {
		return 0, err
	}

	n, err := d.src.Read(b)
	if err != nil && err != io.EOF {
		return n, err
	}
	b = b[:n]

	if n > 0 {
		d.hash.Write(b)
		d.stream.XORKeyStream(b, b)
	}

	if err == io.EOF {
		d.closed = true
		if !equal(getChecksum(d.hash), d.src.Footer()) {
			return n, ErrBadChecksum
		}
	}

	return n, err
}

func (d *Decryptor) Close() error {

	if d.closed {
		return nil
	}

	err := d.readHeader()
	if err != nil {
		return err
	}

	d.closed = true

	if !equal(getChecksum(d.hash), d.src.Footer()) {
		return ErrBadChecksum
	}

	return nil
}

var (
	testingBadChecksum = false
	badChecksumBytes   = []byte{'x'}
)

func (d *Decryptor) readHeader() error {

	if !d.firstTime {
		return nil
	}
	d.firstTime = false

	err := d.header.readFrom(d.src)
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	if err != nil {
		return err
	}

	// TODO: Utilize runtime/secret if or when it becomes available.
	// https://go.dev/doc/go1.26#new-experimental-runtimesecret-package

	password, err := d.passFunc()
	if err != nil {
		return fmt.Errorf("could not retrieve password: %w", err)
	}

	key := argon2.IDKey(
		password,
		d.header.ArgonSalt[:],
		d.header.ArgonTime,
		d.header.ArgonMemory,
		d.header.ArgonThreads,
		d.header.keyLen(),
	)
	clear(password)

	d.stream = d.header.getStream(key)
	d.hash = sha3.NewSHAKE256()
	d.hash.Write(key)
	clear(key)

	d.header.writeTo(d.hash)

	if testingBadChecksum {
		d.hash.Write(badChecksumBytes)
	}

	return nil
}
