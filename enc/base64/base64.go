// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package base64

import (
	"encoding/base64"

	"github.com/koonix/x/enc"
)

func Encode[Out, In ~string | ~[]byte](src In) Out {
	return enc.Encode[Out](base64.StdEncoding, src)
}
func EncodeURL[Out, In ~string | ~[]byte](src In) Out {
	return enc.Encode[Out](base64.URLEncoding, src)
}
func EncodeRaw[Out, In ~string | ~[]byte](src In) Out {
	return enc.Encode[Out](base64.RawStdEncoding, src)
}
func EncodeRawURL[Out, In ~string | ~[]byte](src In) Out {
	return enc.Encode[Out](base64.RawURLEncoding, src)
}

func Decode[Out, In ~string | ~[]byte](src In) (Out, error) {
	return enc.Decode[Out](base64.StdEncoding, src)
}
func DecodeURL[Out, In ~string | ~[]byte](src In) (Out, error) {
	return enc.Decode[Out](base64.URLEncoding, src)
}
func DecodeRaw[Out, In ~string | ~[]byte](src In) (Out, error) {
	return enc.Decode[Out](base64.RawStdEncoding, src)
}
func DecodeRawURL[Out, In ~string | ~[]byte](src In) (Out, error) {
	return enc.Decode[Out](base64.RawURLEncoding, src)
}
