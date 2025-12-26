// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package ringbuf

import (
	"iter"
	"slices"
)

// TODO:
//   - Add Reset, Read, ReadFrom, WriteTo, and all the other crazy methods of bytes.Buffer.
//   - Add a method for changing maxLen.
//   - Add a method for getting the n-th latest byte?

type Buffer[T any] struct {
	buf      []T
	maxLen   int
	writePos int
	byte     [1]T
}

func NewBuffer[T any](maxLen int, initialCap ...int) *Buffer[T] {
	r := &Buffer[T]{
		maxLen: maxLen,
	}
	if len(initialCap) > 0 && initialCap[0] > 0 {
		r.buf = make([]T, 0, min(initialCap[0], maxLen))
	}
	return r
}

func (r *Buffer[T]) Write(b []T) (int, error) {

	if len(b) == 0 {
		return 0, nil
	}

	if len(b) >= r.maxLen {
		r.writePos = 0
		r.buf = slices.Grow(r.buf, r.maxLen-len(r.buf))[:r.maxLen]
		copy(r.buf, b[len(b)-r.maxLen:])
		return len(b), nil
	}

	rem := r.maxLen - len(r.buf)
	if n := min(len(b), rem); n > 0 {
		r.buf = append(r.buf, b[:n]...)
		b = b[n:]
	}

	for len(b) > 0 {
		n := copy(r.buf[r.writePos:], b)
		b = b[n:]
		r.writePos += n
		if r.writePos == r.maxLen {
			r.writePos = 0
		}
	}

	return len(b), nil
}

func (r *Buffer[T]) WriteByte(v T) error {
	r.byte[0] = v
	r.Write(r.byte[:])
	return nil
}

// Bytes returns a copy of the contents of the buffer.
// If you want to access buffer data without copying/allocation,
// consider using [Buffer.BytesSeq], [Buffer.CopyTo] or [Buffer.AppendTo].
//
//x:x
func (r *Buffer[T]) Bytes() []T {
	return r.AppendTo(nil)
}

// BytesSeq returns an iterator over the segments of the buffer.
// The iterator iterates at least once.
func (r *Buffer[T]) BytesSeq() iter.Seq[[]T] {
	return func(yield func([]T) bool) {
		if !yield(r.buf[r.writePos:]) {
			return
		}
		if r.writePos > 0 {
			yield(r.buf[:r.writePos])
		}
	}
}

func (r *Buffer[T]) CopyTo(s []T) int {
	n := copy(s, r.buf[r.writePos:])
	n += copy(s[n:], r.buf[:r.writePos])
	return n
}

func (r *Buffer[T]) AppendTo(s []T) []T {
	s = slices.Grow(s, r.Len())
	r.CopyTo(s)
	return s
}

func (r *Buffer[T]) Len() int {
	if r.writePos == 0 {
		return len(r.buf)
	}
	return r.maxLen
}

func (r *Buffer[T]) MaxLen() int {
	return r.maxLen
}

// // Truncate grows or shrinks the maximum length of the buffer to n.
// func (r *Buffer[T]) Truncate(n int) {
// 	if r.writePos == 0 {
// 		r.maxLen = n
// 		if n < len(r.buf) {
// 			r.buf = r.buf[len(r.buf)-n:]
// 		}
// 		return
// 	}
// 	r2 := NewBuffer[T](r.Len(), n)
// 	r2.Write(r.Bytes())
// 	*r = *r2
// }
