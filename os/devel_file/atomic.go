// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package file

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type AtomicFile struct {
	*os.File
	dest string
	once sync.Once
}

func WriteAtomic(path string, data []byte, perm os.FileMode) (retErr error) {

	f, err := OpenAtomic(path, os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer func() {
		errors.Join(retErr, f.Close())
	}()

	_, err = f.Write(data)
	return err
}

func OpenAtomic(path string, flag int, perm os.FileMode) (*AtomicFile, error) {

	f, err := CreateTemp(path, flag, perm)
	if err != nil {
		return nil, fmt.Errorf(
			"could not create temporary file in %q: %w",
			filepath.Dir(path), err,
		)
	}

	return &AtomicFile{
		File: f,
		dest: path,
	}, nil
}

func (f *AtomicFile) Close() (err error) {
	f.once.Do(func() {
		err = close(f.File, f.dest)
	})
	return
}

func (f *AtomicFile) CloseOnSuccess(errPtr *error) {
	f.once.Do(func() {
		if *errPtr != nil {
			f.RemoveTemp()
			return
		}
		*errPtr = close(f.File, f.dest)
	})
}

func (f *AtomicFile) RemoveTemp() error {
	return os.Remove(f.File.Name())
}

func close(f *os.File, dest string) error {

	err := f.Close()
	if err != nil {
		os.Remove(f.Name())
		return fmt.Errorf("could not close temporary file %q: %w", f.Name(), err)
	}

	err = os.Rename(f.Name(), dest)
	if err != nil {
		os.Remove(f.Name())
		return fmt.Errorf("could not move temporary file %q to %q: %w", f.Name(), dest, err)
	}

	return nil
}
