// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package file

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type BufferedAtomicFile struct {
	*os.File
	w    *bufio.Writer
	dest string
	once sync.Once
}

func OpenBufferedAtomic(path string, flag int, perm os.FileMode) (*BufferedAtomicFile, error) {

	f, err := CreateTemp(path, flag, perm)
	if err != nil {
		return nil, fmt.Errorf(
			"could not create temporary file in %q: %w",
			filepath.Dir(path), err,
		)
	}

	return &BufferedAtomicFile{
		File: f,
		dest: path,
		w:    bufio.NewWriter(f),
	}, nil
}

func (f *BufferedAtomicFile) Close() (retErr error) {
	f.once.Do(func() {
		err := f.w.Flush()
		if err != nil {
			retErr = err
			return
		}
		retErr = close(f.File, f.dest)
	})
	return
}

func (f *BufferedAtomicFile) CloseOnSuccess(errPtr *error) {
	f.once.Do(func() {
		if *errPtr != nil {
			f.RemoveTemp()
			return
		}
		*errPtr = f.w.Flush()
		if *errPtr != nil {
			return
		}
		*errPtr = close(f.File, f.dest)
	})
}

func (f *BufferedAtomicFile) RemoveTemp() error {
	return os.Remove(f.File.Name())
}

func (f *BufferedAtomicFile) Write(b []byte) (int, error) {
	return f.w.Write(b)
}

func (f *BufferedAtomicFile) WriteString(s string) (int, error) {
	return f.w.WriteString(s)
}

func (f *BufferedAtomicFile) WriteByte(c byte) error {
	return f.w.WriteByte(c)
}

func (f *BufferedAtomicFile) WriteRune(r rune) (size int, err error) {
	return f.w.WriteRune(r)
}

func (f *BufferedAtomicFile) ReadFrom(r io.Reader) (int64, error) {
	return f.w.ReadFrom(r)
}

func (f *BufferedAtomicFile) Seek(offset int64, whence int) (int64, error) {
	err := f.w.Flush()
	if err != nil {
		return 0, fmt.Errorf("could not flush the buffer: %w", err)
	}
	return f.File.Seek(offset, whence)
}

func (f *BufferedAtomicFile) Sync() error {
	err := f.w.Flush()
	if err != nil {
		return fmt.Errorf("could not flush the buffer: %w", err)
	}
	return f.File.Sync()
}
