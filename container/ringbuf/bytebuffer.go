// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package ringbuf

import (
	"io"
	"unicode/utf8"
	"unsafe"
)

type ByteBuffer struct {
	Buffer[byte]
	runeBuf [utf8.UTFMax]byte
}

func NewByteBuffer(maxLen int, initialCap ...int) *ByteBuffer {
	r := &ByteBuffer{}
	r.maxLen = maxLen
	if len(initialCap) > 0 && initialCap[0] > 0 {
		r.buf = make([]byte, 0, min(initialCap[0], maxLen))
	}
	return r
}

func (b *ByteBuffer) WriteString(s string) (int, error) {
	return b.Write([]byte(s))
}

func (b *ByteBuffer) WriteRune(ru rune) (int, error) {
	n := utf8.EncodeRune(b.runeBuf[:], ru)
	return b.Write(b.runeBuf[:n])
}

func (b *ByteBuffer) ReadRune() (r rune, size int, err error) {
	n, err := b.ReadAt(b.runeBuf[:], int64(b.readPos))
	if err != nil {
		return 0, 0, err
	}
	r, size = utf8.DecodeRune(b.runeBuf[:n])
	b.readPos += size
	return r, size, nil
}

func (b *ByteBuffer) WriteTo(w io.Writer) (n int64, err error) {
	for s := range b.BytesSeq() {
		m, err := w.Write(s)
		n += int64(m)
		b.readPos += m
		if err != nil {
			return n, err
		}
		if m <= len(s) {
			return n, io.ErrShortWrite
		}
	}
	return n, nil
}

func (b *ByteBuffer) ReadFrom(r io.Reader) (n int64, err error) {

	// TODO:
	// If small reads (caused by a small ringbuf)
	// turn out to be too slow even with a fast `r io.Reader`,
	// somehow read to a small 1kb (or however big) buffer.

	if b.maxLen == 0 {
		return 0, nil
	}

	const minRead = 512

	for b.rem() > 0 {
		b.Grow(minRead)
		m, err := r.Read(b.buf[len(b.buf):cap(b.buf)])
		b.buf = b.buf[:len(b.buf)+m]
		n += int64(m)
		if err == io.EOF {
			return n, nil
		}
		if err != nil {
			return n, err
		}
	}

	for {
		m, err := r.Read(b.buf[b.writePos:])
		n += int64(m)
		b.writePos += m
		if b.writePos == b.maxLen {
			b.writePos = 0
		}
		if err == io.EOF {
			return n, nil
		}
		if err != nil {
			return n, err
		}
	}
}

// String returns the contents of the buffer as a string.
// If [*ByteBuffer] is nil, it returns "<nil>".
func (b *ByteBuffer) String() string {
	if b == nil {
		return "<nil>"
	}
	if b.Len() == 0 {
		return ""
	}
	s := make([]byte, b.Len())
	b.ReadAt(s, int64(b.readPos))
	return unsafe.String(&s[0], len(s))
}
