// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package aes256

import (
	"crypto/aes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"

	"github.com/koonix/x/must"
)

const (
	version       = 0
	argonSaltSize = 16

	argonMemory    = 16 * 1024
	argonMemoryMax = 64 * 1024

	argonTime    = 3
	argonTimeMax = 10

	argonThreads    = 8
	argonThreadsMax = 64
)

var ErrHeaderParamsOutOfRange = errors.New("header params out of range")

type header struct {
	Version      uint8
	ArgonMemory  uint32
	ArgonTime    uint32
	ArgonThreads uint8
	ArgonSalt    [argonSaltSize]byte
	AesIV        [aes.BlockSize]byte
}

func (h header) writeTo(w io.Writer) error {
	err := h.check()
	if err != nil {
		return err
	}
	return binary.Write(w, binary.BigEndian, h)
}

func (h *header) readFrom(r io.Reader) error {
	err := binary.Read(r, binary.BigEndian, h)
	if err != nil {
		return err
	}
	return h.check()
}

func (h header) check() error {
	if h.ArgonMemory > argonMemoryMax {
		return ErrHeaderParamsOutOfRange
	}
	if h.ArgonTime > argonTimeMax {
		return ErrHeaderParamsOutOfRange
	}
	if h.ArgonThreads > argonThreadsMax {
		return ErrHeaderParamsOutOfRange
	}
	return nil
}

func newHeader() header {
	h := header{
		Version:      version,
		ArgonMemory:  argonMemory,
		ArgonTime:    argonTime,
		ArgonThreads: argonThreads,
	}
	must.Get(rand.Read(h.ArgonSalt[:]))
	must.Get(rand.Read(h.AesIV[:]))
	return h
}
