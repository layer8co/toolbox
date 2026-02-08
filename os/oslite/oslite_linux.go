// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

//go:build linux

package oslite

import (
	"io"
	"os"
)

func (f *File) Open(path string) error {
	return f.OpenFile(path, os.O_RDONLY, 0)
}

func (f *File) Create(path string) error {
	return f.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

func (f *File) OpenFile(path string, flag int, perm os.FileMode) error {
	fd, errno := open(path, flag, perm)
	if errno != 0 {
		return errno
	}
	f.fd = fd
	return nil
}

func (f File) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	n, errno := read(f.fd, b)
	if errno != 0 {
		return 0, errno
	}
	if n == 0 {
		return 0, io.EOF
	}
	return n, nil
}

func (f File) Close() error {
	errno := close(f.fd)
	if errno != 0 {
		return errno
	}
	return nil
}
