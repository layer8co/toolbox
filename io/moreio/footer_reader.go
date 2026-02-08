// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package moreio

import (
	"io"
)

// 100 seems to be a good value:
// https://github.com/golang/go/blob/go1.25.0/src/bufio/bufio.go#L45
const maxConsecutiveEmptyReads = 100

// FooterReader is returned by [NewFooterReader].
// See it's documentation for details.
type FooterReader struct {
	r          io.Reader
	footer     []byte
	footerSize int
	extra      []byte
}

// NewFooterReader returns a reader that reads data from r,
// excluding the last len(footer) bytes of the stream,
// which can be retrieved via [FooterReader.Footer].
//
// Optionally, an extra buffer with the same length as footer
// can be provided to prevent it being allocated once
// during the internal operations.
//
// [FooterReader.Footer] always returns the latest read bytes,
// regardless of errors, whether EOF has been achieved,
// or whether len(footer) bytes have been read from r yet.
//
// [FooterReader.Read] is safe to call again
// after errors (including EOF).
func NewFooterReader(r io.Reader, footer []byte, extra ...[]byte) *FooterReader {
	f := &FooterReader{}
	f.ResetWithFooter(r, footer, extra...)
	return f
}

func (f *FooterReader) Reset(r io.Reader) {
	f.r = r
	f.footerSize = 0
}

func (f *FooterReader) ResetWithFooter(r io.Reader, footer []byte, extra ...[]byte) {
	f.r = r
	f.footer = footer
	f.footerSize = 0
	if len(extra) > 0 {
		if len(extra[0]) != len(footer) {
			panic("NewFooterReader: len(extra[0]) != len(footer)")
		}
		f.extra = extra[0]
	}
}

func (r *FooterReader) Read(b []byte) (int, error) {

	i := 0
	for r.footerSize != len(r.footer) {
		x := len(r.footer) - r.footerSize
		n, err := r.r.Read(r.footer[:x])
		copy(r.footer[x-n:x], r.footer[:n])
		r.footerSize += n
		if err == io.EOF {
			return 0, io.ErrUnexpectedEOF
		}
		if err != nil {
			return 0, err
		}
		if n == 0 {
			i++
		} else {
			i = 0
		}
		if i >= maxConsecutiveEmptyReads {
			return 0, io.ErrNoProgress
		}
	}

	n, err := r.r.Read(b)

	if n == 0 {
		return n, err
	}

	if r.extra == nil {
		r.extra = make([]byte, len(r.footer))
	}
	copy(r.extra, r.footer)

	f := len(r.footer)
	x := min(n, f)

	f -= copy(r.footer[f-x:f], b[n-x:n])
	copy(r.footer[:f], r.extra[len(r.extra)-f:])

	copy(b[x:n], b[:n-x])
	copy(b[:x], r.extra[:len(r.extra)-f])

	return n, err
}

func (r *FooterReader) Footer() []byte {
	return r.footer[len(r.footer)-r.footerSize:]
}
