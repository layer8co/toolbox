// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package oslite

import "os"

func (f *File) Open(path string) error {
	file, err := os.Open(path)
	f.file = file
	return err
}

func (f *File) Create(path string) error {
	file, err := os.Create(path)
	f.file = file
	return err
}

func (f *File) OpenFile(path string, flag int, perm os.FileMode) error {
	file, err := os.OpenFile(path, flag, perm)
	f.file = file
	return err
}

func (f File) Read(b []byte) (int, error) {
	return f.file.Read(b)
}

func (f File) Close() error {
	return f.file.Close()
}
