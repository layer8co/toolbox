// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package bufpool

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestPoolCleanup(t *testing.T) {

	New[byte]()

	cleanedUp := make(chan struct{})
	testCleanup = func() {
		close(cleanedUp)
	}

	runtime.GC()

	select {
	case <-cleanedUp:
		// Success.
	case <-time.After(500 * time.Millisecond):
		t.Fatal("cleanup did not run")
	}
}

func isPoolEmpty(pool *sync.Pool) bool {
	var v any
	allocs := testing.AllocsPerRun(1, func() {
		v = pool.Get()
	})
	if allocs > 0 {
		return true
	}
	pool.Put(v)
	return false
}
