// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

//go:build unix

package task

import (
	"os"
	"os/signal"
	"syscall"
)

func trap() (sig <-chan os.Signal, stop func()) {

	Sig := make(chan os.Signal, 1)
	signal.Notify(
		Sig,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGHUP,
	)

	sigpipe := make(chan os.Signal, 1)
	signal.Notify(sigpipe, syscall.SIGPIPE)

	stop = func() {
		signal.Stop(Sig)
		signal.Stop(sigpipe)
		close(sigpipe)
	}

	return Sig, stop
}
