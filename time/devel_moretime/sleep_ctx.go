// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package moretime

import (
	"context"
	"sync"
	"time"
)

var timerPool = sync.Pool{
	New: func() any {
		t := time.NewTimer(time.Hour)
		return t
	},
}

// SleepContext sleeps for at least the duration d, or until ctx is done.
//
// It returns false if the return was due to ctx becoming done, and true otherwise.
func SleepContext(ctx context.Context, d time.Duration) bool {

	// TODO: Add zeroctx support.

	if d <= 0 {
		return isCtxAlive(ctx)
	}

	t := timerPool.Get().(*time.Timer)
	t.Reset(d)

	select {

	case <-t.C:
		timerPool.Put(t)
		return true

	case <-ctx.Done():
		t.Stop()
		timerPool.Put(t)
		return false
	}
}

func isCtxAlive(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		return true
	}
}
