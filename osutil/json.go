// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package osutil

import (
	"encoding/json"
	"os"
)

func WriteJson(
	path string,
	data any,
	perm os.FileMode,
	prefix, indent string,
) (
	err error,
) {

	f, err := OpenFileAtomic(path, os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer f.Close(&err)

	e := json.NewEncoder(f)
	e.SetIndent(prefix, indent)

	return e.Encode(data)
}

func ReadJson(path string, dest any) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(dest)
}
