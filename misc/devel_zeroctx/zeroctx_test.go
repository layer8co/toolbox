// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package zeroctx_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/layer8co/toolbox/misc/zeroctx"
)

// ==========

func BenchmarkCtxNoop(b *testing.B) {
	for b.Loop() {
		ctx, cancel := context.WithCancel(context.Background())
		_ = ctx
		cancel()
	}
}

func BenchmarkZeroCtxNoop(b *testing.B) {
	for b.Loop() {
		ctx := zeroctx.WithCancel(context.Background())
		zeroctx.Put(ctx)
	}
}

// ==========

func BenchmarkCtxParentNoop(b *testing.B) {
	for b.Loop() {
		ctx, cancel := context.WithCancel(context.Background())
		_ = ctx
		cancel()
	}
}

func BenchmarkZeroCtxParentNoop(b *testing.B) {
	for b.Loop() {
		ctx := zeroctx.WithCancel(context.Background())
		zeroctx.Put(ctx)
	}
}

// ==========

func BenchmarkCtxDone(b *testing.B) {
	for b.Loop() {
		ctx, cancel := context.WithCancel(context.Background())
		ch := ctx.Done()
		_ = ch
		cancel()
	}
}

func BenchmarkZeroCtxDone(b *testing.B) {
	for b.Loop() {
		ctx := zeroctx.WithCancel(context.Background())
		ch := ctx.GetDone()
		ctx.PutDone(ch)
		zeroctx.Put(ctx)
	}
}

// ==========

var timerPool = sync.Pool{
	New: func() any {
		t := time.NewTimer(time.Hour)
		t.Stop()
		return t
	},
}

func sleepZeroCtx(ctx zeroctx.Context, d time.Duration) bool {

	if d <= 0 {
		return isZeroCtxAlive(ctx)
	}

	done := ctx.GetDone()
	defer ctx.PutDone(done)

	t := timerPool.Get().(*time.Timer)
	t.Reset(d)
	defer func() {
		t.Stop()
		timerPool.Put(t)
	}()

	select {
	case <-t.C:
		return true
	case <-done:
		return false
	}
}

// SleepCtx sleeps for duration d or until ctx is done.
// Returns true if the sleep completed, false if the context was canceled/deadlined.
func sleepCtx(ctx context.Context, d time.Duration) bool {

	if d <= 0 {
		return isCtxAlive(ctx)
	}

	t := timerPool.Get().(*time.Timer)
	t.Reset(d)
	defer func() {
		t.Stop()
		timerPool.Put(t)
	}()

	select {
	case <-t.C:
		return true
	case <-ctx.Done():
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

func isZeroCtxAlive(ctx zeroctx.Context) bool {
	done := ctx.GetDone()
	ctx.PutDone(done)
	select {
	case <-done:
		return false
	default:
		return true
	}
}
