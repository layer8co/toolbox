// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package base32

import (
	"encoding/base32"

	"github.com/koonix/x/enc"
)

func Encode[Out, In ~string | ~[]byte](src In) Out {
	return enc.Encode[Out](base32.StdEncoding, src)
}
func EncodeHex[Out, In ~string | ~[]byte](src In) Out {
	return enc.Encode[Out](base32.HexEncoding, src)
}

func Decode[Out, In ~string | ~[]byte](src In) (Out, error) {
	return enc.Decode[Out](base32.StdEncoding, src)
}
func DecodeHex[Out, In ~string | ~[]byte](src In) (Out, error) {
	return enc.Decode[Out](base32.HexEncoding, src)
}
