// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

// Package zeroctx is a zero-allocation context implementation.
package zeroctx

import (
	"context"
)

// ==========

// TODO: Make [Put] take care of [Reset]-ing too.

// ==========

// TODO:
//
// Implement the context methods:
//
//	 func (c *cancelCtx) Deadline() (deadline time.Time, ok bool) {}
//	 func (c *cancelCtx) Done() <-chan struct{} {}
//	 func (c *cancelCtx) Err() error {}
//	 func (c *cancelCtx) Value(key any) any {}

// ==========

// TODO:
//
// Ditch the actor-style architecture and use lock-based synchronization,
// since it's quite a bit slower than the stdlib context,
// but with a lock-based approach,
// we'd be quite a bit faster than stdlib:
//
//	 func (c *cancelCtx) GetDone() <-chan struct{} {
//	 	done := chanPool.Get().(rwChan)
//	 	if c.done.Load() {
//	 		done <- struct{}{}
//	 	} else {
//	 		c.mu.Lock()
//	 		c.subs[done] = done
//	 		c.mu.Unlock()
//	 	}
//	 	return done
//	 }
//
//	 func (c *cancelCtx) PutDone(done <-chan struct{}) {
//	 	c.mu.Lock()
//	 	rw := c.subs[done]
//	 	delete(c.subs, done)
//	 	c.mu.Unlock()
//	 	drain(done)
//	 	chanPool.Put(rw)
//	 }
//
//	 func drain[T any](ch <-chan T) {
//	 	select {
//	 	case <-ch:
//	 	default:
//	 	}
//	 }

// ==========

type Context interface {
	context.Context
	GetDone() <-chan struct{}
	PutDone(<-chan struct{})
	// TODO: `IsDone() bool`
}

func WithCancel(parent context.Context) Context {
	ctx := cancelCtxPool.Get().(*cancelCtx)
	ch := parent.Done()
	if ch != nil {
		ctx.parent <- ch
	}
	return ctx
}

func Put(ctx Context) {
	cancelCtxPool.Put(ctx.(*cancelCtx))
}

func Reset(ctx Context) Context {
	c := ctx.(*cancelCtx)
	*c = cancelCtx{}
	return c
}

type (
	rwChan = chan struct{}
	roChan = <-chan struct{}
)
