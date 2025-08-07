// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package osutil

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

func WriteFileAtomic(path string, data []byte, perm os.FileMode) (err error) {

	f, err := OpenFileAtomic(path, os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer f.Close(&err)

	_, err = f.Write(data)
	return err
}

func OpenFileAtomic(path string, flag int, perm os.FileMode) (File, error) {

	f, err := CreateTemp(path, os.O_WRONLY, perm)
	if err != nil {
		return File{}, fmt.Errorf(
			"could not create temporary file in %q: %w",
			filepath.Dir(path), err,
		)
	}

	var once sync.Once

	close := func(e *error) {
		once.Do(func() {

			if *e != nil {
				os.Remove(f.Name())
				return
			}

			err := f.Close()
			if err != nil {
				os.Remove(f.Name())
				*e = fmt.Errorf(
					"could not close temporary file %q: %w",
					f.Name(), err,
				)
			}

			err = os.Rename(f.Name(), path)
			if err != nil {
				os.Remove(f.Name())
				*e = fmt.Errorf(
					"could not move temporary file %q to %q: %w",
					f.Name(), path, err,
				)
			}
		})
	}

	return File{
		File:  f,
		close: close,
	}, nil
}

type File struct {
	*os.File
	close func(*error)
}

func (f File) Close(err *error) {
	f.close(err)
}
