// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package ringbuf

// The x:x in the comments are gofmt-compatible spacers.
//
// https://github.com/golang/go/issues/54489

import (
	"slices"
	"unicode/utf8"
)

type Buffer struct {
	buf      []byte
	maxLen   int
	writePos int
}

func New(maxLen int, initialCap ...int) *Buffer {
	x := &Buffer{
		maxLen: maxLen,
	}
	if len(initialCap) > 0 && initialCap[0] > 0 {
		x.buf = make([]byte, 0, min(initialCap[0], maxLen))
	}
	return x
}

func (x *Buffer) Write(b []byte) (int, error) {
	return write(x, b)
}

func (x *Buffer) WriteString(s string) (int, error) {
	return write(x, s)
}

func write[T string | []byte](x *Buffer, b T) (int, error) {

	if len(b) == 0 {
		return 0, nil
	}

	if len(b) >= x.maxLen {
		x.writePos = 0
		x.buf = slices.Grow(x.buf, x.maxLen-len(x.buf))[:x.maxLen]
		copy(x.buf, b[len(b)-x.maxLen:])
		return len(b), nil
	}

	rem := x.maxLen - len(x.buf)
	if n := min(len(b), rem); n > 0 {
		x.buf = append(x.buf, b[:n]...)
		b = b[n:]
	}

	for len(b) > 0 {
		n := copy(x.buf[x.writePos:], b)
		b = b[n:]
		x.writePos += n
		if x.writePos == x.maxLen {
			x.writePos = 0
		}
	}

	return len(b), nil
}

func (x *Buffer) WriteByte(c byte) error {
	x.Write([]byte{c})
	return nil
}

func (x *Buffer) WriteRune(r rune) (int, error) {
	return x.Write(utf8.AppendRune([]byte{}, r))
}

// Bytes returns the contents of the buffer.
//
// The returned slice should only be used for reading,
// since it may alias the buffer content
// at least until the next buffer modification.
//
// Use [Buffer.CloneBytes] if you intend to
// modify the returned slice.
//
//x:x
func (x *Buffer) Bytes() []byte {
	if x.writePos == 0 {
		return x.buf
	}
	ret := make([]byte, x.maxLen)
	n := copy(ret, x.buf[x.writePos:])
	copy(ret[n:], x.buf[:x.writePos])
	return ret
}

// CloneBytes is similar to [Buffer.Bytes],
// but the returned slice is a copy of the underlying data.
//
//x:x
func (x *Buffer) CloneBytes() []byte {
	if x.writePos == 0 {
		return slices.Clone(x.buf)
	}
	return slices.Concat(x.buf[:x.writePos], x.buf[x.writePos:])
}

// String returns the contents of the buffer as a string.
// If the [Buffer] is a nil pointer,
// it returns "<nil>".
//
//x:x
func (x *Buffer) String() string {
	if x == nil {
		return "<nil>"
	}
	return string(x.Bytes())
}

func (x *Buffer) Len() int {
	if x.writePos == 0 {
		return len(x.buf)
	}
	return x.maxLen
}

func (x *Buffer) MaxLen() int {
	return x.maxLen
}

// Truncate grows or shrinks the maximum length of the buffer to n.
func (x *Buffer) Truncate(n int) {
	if x.writePos == 0 {
		x.maxLen = n
		if n < len(x.buf) {
			x.buf = x.buf[len(x.buf)-n:]
		}
		return
	}
	x2 := New(x.Len(), n)
	x2.Write(x.Bytes())
	*x = *x2
}
