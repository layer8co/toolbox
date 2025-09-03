// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package text

import (
	"github.com/koonix/x/internal/noalloc"
)

func Wrap[Dst, Src ~[]byte | ~string](src Src, limit int) Dst {
	dst := make([]byte, 0, len(src)+(len(src)/limit))
	if limit <= 0 {
		panic("Wrap: limit <= 0")
	}
	n := limit
	i := 0
	for i < len(src) {
		dst = append(dst, src[i])
		i++
		n--
		if n == 0 {
			dst = append(dst, '\n')
			n = limit
		}
	}
	return noalloc.FromBytes[Dst](dst)
}
