// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

//go:build !unix

package task

import (
	"os"
	"os/signal"
	"syscall"
)

func trap() (sig <-chan os.Signal, stop func()) {

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(
		interruptChan,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGHUP,
	)

	stop = func() {
		signal.Stop(interruptChan)
	}

	return interruptChan, stop
}
