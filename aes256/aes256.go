// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package aes256

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/koonix/x/must"
	"golang.org/x/crypto/argon2"
)

const (
	version = 0
	keyLen  = 32
)

func Encrypt(password, plaintext, additionalData []byte) []byte {
	h := newHeader()
	b := encodeHeader(h)
	aead := getAEAD(password, h)
	return aead.Seal(b, h.Nonce[:], plaintext, additionalData)
}

func Decrypt(password, ciphertext, additionalData []byte) ([]byte, error) {
	ciphertext, h, err := decodeHeader(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("could not decode header: %w", err)
	}
	aead := getAEAD(password, h)
	return aead.Open(nil, h.Nonce[:], ciphertext, additionalData)
}

func getAEAD(password []byte, h header) cipher.AEAD {
	key := argon2.IDKey(password, h.Salt[:], h.Time, h.Memory, h.Threads, keyLen)
	defer zero(key)
	block := must.Get(aes.NewCipher(key))
	return must.Get(cipher.NewGCM(block))
}

const (
	saltSize  = 16
	nonceSize = 12
	headerLen = 0 +
		1 + // version
		1 + // threads
		4 + // time
		4 + // memory
		saltSize +
		nonceSize
)

type header struct {
	Version uint8
	Threads uint8
	Time    uint32
	Memory  uint32
	Salt    [saltSize]byte
	Nonce   [nonceSize]byte
}

func newHeader() header {
	c := header{
		Version: version,
		Threads: 8,
		Time:    3,
		Memory:  64 * 1024,
	}
	rand.Read(c.Salt[:])
	rand.Read(c.Nonce[:])
	return c
}

func encodeHeader(h header) []byte {
	b := make([]byte, 0, headerLen)
	b = append(b, h.Version, h.Threads)
	b = binary.BigEndian.AppendUint32(b, h.Time)
	b = binary.BigEndian.AppendUint32(b, h.Memory)
	b = append(b, h.Salt[:]...)
	b = append(b, h.Nonce[:]...)
	return b
}

func decodeHeader(b []byte) ([]byte, header, error) {

	h := header{}

	if len(b) < headerLen {
		return b, h, fmt.Errorf(
			"invalid length (want >= %d, got %d)",
			headerLen, len(b),
		)
	}

	h.Version = b[0]
	b = b[1:]

	if h.Version != version {
		return b, h, fmt.Errorf(
			"unsupported version (want %d, got %d)",
			version, h.Version,
		)
	}

	h.Threads = b[0]
	b = b[1:]

	h.Time = binary.BigEndian.Uint32(b)
	b = b[4:]

	h.Memory = binary.BigEndian.Uint32(b)
	b = b[4:]

	h.Salt = [saltSize]byte(b[:saltSize])
	b = b[saltSize:]

	h.Nonce = [nonceSize]byte(b[:nonceSize])
	b = b[nonceSize:]

	return b, h, nil
}

func zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
