// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package ringbuf

import (
	"unicode/utf8"
)

type ByteBuffer struct {
	Buffer[byte]
	rune [utf8.UTFMax]byte
}

func NewByteBuffer(maxLen int, initialCap ...int) *ByteBuffer {
	r := &ByteBuffer{}
	r.maxLen = maxLen
	if len(initialCap) > 0 && initialCap[0] > 0 {
		r.buf = make([]byte, 0, min(initialCap[0], maxLen))
	}
	return r
}

func (r *ByteBuffer) WriteString(s string) (int, error) {
	return r.Write([]byte(s))
}

func (r *ByteBuffer) WriteRune(ru rune) (int, error) {
	n := utf8.EncodeRune(r.rune[:], ru)
	return r.Write(r.rune[:n])
}

// String returns the contents of the buffer as a string.
// If [*ByteBuffer] is nil, it returns "<nil>".
//
//x:x
func (r *ByteBuffer) String() string {
	if r == nil {
		return "<nil>"
	}
	return string(r.Bytes())
}
