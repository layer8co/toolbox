// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

// Package ringbuf implements a dynamically growing ring buffer.
package ringbuf

import (
	"errors"
	"io"
	"iter"
	"slices"
)

// TODO:
//
//   - Add a Truncate() method.
//       Of course if cap hasn't been reached, truncation is super easy.
//       Otherwise, if the buffer ain't that full,
//       we can shift things around in the buffer itself.
//       If it's really full, put the extra shit
//       in a lil buffer you got from a pool.
//
//   - Add the ability to lower the cap of the buffer.
//       of course if the new cap(r.buf) is less than len(r.buf),
//       we need to jettison SOME data.
//       Do we jettison the newest data? The oldest data?
//
//   - Add Seek().

var ErrNegativeOffset = errors.New("negative offset")

type Buffer[T any] struct {
	buf      []T
	maxLen   int
	writePos int // writePos is an actual position in buf.
	readPos  int // readPos is relative to writePos. It wraps around buf.
	byteBuf  [1]T
}

// NewBuffer returns a new ring buffer.
//
// Incoming data grows the buffer from initialCap
// (which is zero if not provided)
// to maxLen, then overwrites old data.
func NewBuffer[T any](maxLen int, initialCap ...int) *Buffer[T] {
	r := &Buffer[T]{
		maxLen: maxLen,
	}
	if len(initialCap) > 0 && initialCap[0] > 0 {
		r.buf = make([]T, 0, min(initialCap[0], maxLen))
	}
	return r
}

func (b *Buffer[T]) Write(src []T) (n int, err error) {

	n = len(src)

	if len(src) == 0 || b.maxLen == 0 {
		return n, nil
	}

	if len(src) >= b.maxLen {
		b.writePos = 0
		b.buf = slices.Grow(b.buf, b.rem())[:b.maxLen]
		copy(b.buf, src[len(src)-b.maxLen:])
		return n, nil
	}

	if m := min(len(src), b.rem()); m > 0 {
		b.buf = append(b.buf, src[:m]...)
		src = src[m:]
	}

	for len(src) > 0 {
		m := copy(b.buf[b.writePos:], src)
		src = src[m:]
		b.writePos += m
		if b.writePos == b.maxLen {
			b.writePos = 0
		}
	}

	return n, nil
}

func (b *Buffer[T]) WriteByte(v T) error {
	b.byteBuf[0] = v
	b.Write(b.byteBuf[:])
	return nil
}

func (b *Buffer[T]) Read(dest []T) (int, error) {
	n, err := b.ReadAt(dest, int64(b.readPos))
	b.readPos += n
	return n, err
}

func (b *Buffer[T]) ReadByte() (T, error) {
	_, err := b.Read(b.byteBuf[:])
	return b.byteBuf[0], err
}

func (b *Buffer[T]) ReadAt(dest []T, offset int64) (n int, err error) {
	if offset < 0 {
		return 0, ErrNegativeOffset
	}
	if offset >= int64(len(b.buf)) {
		return 0, io.EOF
	}
	if len(dest) == 0 {
		return 0, nil
	}
	for b := range b.seq(int(offset)) {
		m := copy(dest, b)
		dest = dest[m:]
		n += m
	}
	return n, nil
}

// Bytes returns a copy of the unread elements of the buffer.
//
// If you want to access buffer data without copying/allocation,
// consider using [Buffer.BytesSeq].
func (r *Buffer[T]) Bytes() []T {
	b := make([]T, r.Len())
	r.ReadAt(b, int64(r.readPos))
	return b
}

// BytesSeq returns an iterator over the unread elements of the buffer.
// It's similar to [Buffer.Bytes]
//
// The returned slices alias the buffer content
// at least until the next buffer modification.
func (b *Buffer[T]) BytesSeq() iter.Seq[[]T] {
	return b.seq(b.readPos)
}

// Next returns a copy of the first n unread elements of the buffer,
// advancing the buffer as if the data had been returned by [Buffer.Read].
//
// If you want to access this data without allocations,
// consider using [Buffer.NextSeq].
func (b *Buffer[T]) Next(n int) []T {
	s := make([]T, min(n, b.Len()))
	b.Read(s)
	return s
}

// NextSeq returns an iterator over the next n elements of the buffer,
// advancing the buffer as if the data had been returned by [Buffer.Read].
//
// The returned slices alias the buffer content
// at least until the next buffer modification.
func (b *Buffer[T]) NextSeq(n int) iter.Seq[[]T] {
	if n < 0 {
		panic("ringbuf.Buffer.NextSeq: n < 0")
	}
	return func(yield func([]T) bool) {
		for s := range b.seq(b.readPos) {
			if n == 0 {
				break
			}
			s = s[:min(len(s), n)]
			if !yield(s) {
				break
			}
			n -= len(s)
		}
		b.readPos += n
	}
}

// seq returns an iterator over the segments of the buffer.
// It does not iterate if the buffer is empty.
func (b *Buffer[T]) seq(offset int) iter.Seq[[]T] {

	if offset < 0 {
		panic("ringbuf.Buffer.seq: offset < 0")
	}

	return func(yield func([]T) bool) {

		if len(b.buf) == 0 {
			return
		}

		s := b.buf[b.writePos:]
		if offset < len(s) && !yield(s[offset:]) {
			return
		}

		offset = max(offset-len(s), 0)
		s = b.buf[offset:b.writePos]
		if len(s) > 0 && !yield(s[offset:]) {
			return
		}
	}
}

// // Truncate discards all but the first unread bytes from the buffer
// // but continues to use the same allocated storage.
// //
// // It panics if n is negative or greater than r.Len().
// func (r *Buffer[T]) Truncate(n int) {
// 	if n < 0 {
// 		panic("ringbuf.Buffer.Truncate: n < 0")
// 	}
// 	if n > r.Len() {
// 		panic("ringbuf.Buffer.Truncate: n > r.Len()")
// 	}
// 	if n == r.Len() {
// 		return
// 	}
// 	if r.writePos == 0 {
// 		r.buf = r.buf[:r.readPos+n]
// 		return
// 	}
// 	if end := r.writePos + r.readPos + n; end < len(r.buf) && /* overflow check: */ end > 0 {
// 		copy(r.buf, r.buf[r.writePos:n])
// 		r.buf = r.buf[:n-r.writePos]
// 		return
// 	}
// }

func (b *Buffer[T]) Reset() {
	b.buf = b.buf[:0]
	b.writePos = 0
	b.readPos = 0
}

// Len returns the number of unread elements of the buffer;
// r.Len() == len(r.Bytes())
func (b *Buffer[T]) Len() int {
	return len(b.buf) - b.readPos
}

func (b *Buffer[T]) MaxLen() int {
	return b.maxLen
}

func (b *Buffer[T]) Cap() int {
	return cap(b.buf)
}

func (b *Buffer[T]) Grow(n int) {
	b.buf = slices.Grow(b.buf, min(n, b.rem()))
}

// rem returns the remaining length until maxLen is reached.
func (b *Buffer[T]) rem() int {
	return b.maxLen - len(b.buf)
}
