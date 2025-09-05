// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package enc

import (
	"encoding/base64"

	"github.com/koonix/x/internal/noalloc"
)

type Encoding interface {
	EncodedLen(n int) int
	DecodedLen(n int) int
	Encode(dst, src []byte)
	Decode(dst, src []byte) (int, error)
}

func Encode[Out, In ~string | ~[]byte](enc Encoding, src In) Out {
	dst := make([]byte, enc.EncodedLen(len(src)))
	enc.Encode(dst, []byte(src))
	return noalloc.ConvertBytes[Out](dst)
}

func Decode[Out, In ~string | ~[]byte](enc Encoding, src In) (Out, error) {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(src)))
	_, err := enc.Decode(dst, []byte(src))
	return noalloc.ConvertBytes[Out](dst), err
}
