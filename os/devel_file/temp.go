// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package file

import (
	"math/rand/v2"
	"os"
	"strconv"
)

// CreateTemp is similar to [os.CreateTemp]
// but it allows setting the permissions of the temporary file as well.
func CreateTemp(path string, flag int, perm os.FileMode) (*os.File, error) {
	try := 0
	for {
		flag |= os.O_CREATE | os.O_EXCL
		tempPath := path + "-temp" + strconv.Itoa(rand10())
		f, err := os.OpenFile(tempPath, flag, perm)
		if os.IsExist(err) {
			if try++; try < 10000 {
				continue
			}
			return nil, &os.PathError{
				Op:   "createtemp",
				Path: path + "*",
				Err:  os.ErrExist,
			}
		}
		return f, err
	}
}

// rand10 returns a random 10 digit int.
func rand10() int {
	return rand.IntN(9_000_000_000) + 1_000_000_000
}
