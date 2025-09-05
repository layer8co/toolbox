// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package hex

import (
	"encoding/hex"

	"github.com/koonix/x/enc"
)

var Encoding enc.Encoding = e{}

type e struct{}

func (e) EncodedLen(n int) int {
	return hex.EncodedLen(n)
}

func (e) DecodedLen(n int) int {
	return hex.DecodedLen(n)
}

func (e) Encode(dst, src []byte) {
	hex.Encode(dst, src)
}

func (e) Decode(dst, src []byte) (int, error) {
	return hex.Decode(dst, src)
}

func Encode[Out, In ~string | ~[]byte](src In) Out {
	return enc.Encode[Out](Encoding, src)
}

func Decode[Out, In ~string | ~[]byte](src In) (Out, error) {
	return enc.Decode[Out](Encoding, src)
}
