// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package must

func Do(err error) {
	if err != nil {
		panic(err)
	}
}

func Get[T any](v T, err error) T {
	Do(err)
	return v
}

func Get2[X, Y any](x X, y Y, err error) (X, Y) {
	Do(err)
	return x, y
}
