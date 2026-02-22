// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package file

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

	f, err := OpenBufferedAtomic(path, os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
	}()

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
