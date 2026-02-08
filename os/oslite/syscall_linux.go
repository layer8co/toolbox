// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

//go:build linux

package oslite

import (
	"os"
	"slices"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

const bufPoolMaxSize = 1024

var bufPool = sync.Pool{
	New: func() any {
		return &[]byte{}
	},
}

func open(
	path string,
	flags int,
	mode os.FileMode,
) (
	fd int,
	errno syscall.Errno,
) {

	if strings.IndexByte(path, 0) != -1 {
		return 0, syscall.EINVAL
	}
	buf := bufPool.Get().(*[]byte)
	*buf = slices.Grow(*buf, len(path)+1)
	*buf = append(*buf, path...)
	*buf = append(*buf, 0)
	if len(*buf) <= bufPoolMaxSize {
		defer func() {
			*buf = (*buf)[0:]
			bufPool.Put(buf)
		}()
	}

	var AT_FDCWD int = -0x64

	dirfd := AT_FDCWD
	flags |= syscall.O_CLOEXEC | syscall.O_LARGEFILE
	mode = os.FileMode(syscallMode(mode))

	return untilNoEintr2(func() (int, syscall.Errno) {
		fd, _, errno := syscall.Syscall6(
			//
			syscall.SYS_OPENAT,
			//
			uintptr(dirfd),
			uintptr(unsafe.Pointer(&(*buf)[0])),
			uintptr(flags),
			uintptr(mode),
			0,
			0,
		)
		return int(fd), errno
	})
}

func close(fd int) syscall.Errno {
	_, _, errno := syscall.Syscall(syscall.SYS_CLOSE, uintptr(fd), 0, 0)
	return errno
}

// Notes:
//   - len(b) must be > 0.
//   - n == 0 indicates EOF.
func read(fd int, b []byte) (n int, errno syscall.Errno) {
	return untilNoEintr2(func() (int, syscall.Errno) {
		n, _, errno := syscall.Syscall(
			//
			syscall.SYS_READ,
			//
			uintptr(fd),
			uintptr(unsafe.Pointer(&b[0])),
			uintptr(len(b)),
		)
		return int(n), errno
	})
}

// Similar to syscallMode() in stdlib's os/file_posix.go
func syscallMode(mode os.FileMode) (x uint32) {
	x |= uint32(mode.Perm())
	if mode&os.ModeSetuid != 0 {
		x |= syscall.S_ISUID
	}
	if mode&os.ModeSetgid != 0 {
		x |= syscall.S_ISGID
	}
	if mode&os.ModeSticky != 0 {
		x |= syscall.S_ISVTX
	}
	return
}

// Similar to ignoringEINTR() in stdlib's os/file_posix.go
func untilNoEintr2[T any](fn func() (T, syscall.Errno)) (T, syscall.Errno) {
	for {
		v, errno := fn()
		if errno != syscall.EINTR {
			return v, errno
		}
	}
}
