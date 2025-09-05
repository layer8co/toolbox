// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package noalloc

import "unsafe"

func ConvertBytes[Out ~string | ~[]byte](b []byte) Out {
	var v Out
	if len(b) == 0 {
		return v
	}
	switch any(v).(type) {
	case []byte:
		return Out(b)
	case string:
		return Out(unsafe.String(&b[0], len(b)))
	default:
		panic("unreachable")
	}
}
