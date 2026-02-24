// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package zeroctx

import (
	"context"
	"runtime"
	"sync"
)

type cancelCtx struct {
	context.Context // TODO: Remove this once we have implemented all it's methods.

	parent chan roChan
	sub    chan rwChan
	unsub  chan roChan
}

var (
	cancelCtxPool = sync.Pool{
		New: func() any {
			return newCancelCtx()
		},
	}
	chanPool = sync.Pool{
		New: func() any {
			return make(rwChan, 1)
		},
	}
)

func newCancelCtx() *cancelCtx {

	var (
		gc     = make(rwChan)
		parent = make(chan roChan)
		sub    = make(chan rwChan)
		unsub  = make(chan roChan)
	)

	go func() {

		var (
			parentDone roChan
			nextSub    rwChan

			isDone = false
		)

		// TODO: Use a map pool here once we have an implementation of it.
		subs := make(map[roChan]rwChan)

		for {

			if nextSub == nil {
				nextSub = chanPool.Get().(rwChan)
			}

			select {

			case <-gc:
				return

			case ch := <-parent:
				parentDone = ch

			case <-parentDone:
				isDone = true
				for _, s := range subs {
					s <- struct{}{}
				}

			case sub <- nextSub:
				if isDone {
					nextSub <- struct{}{}
				} else {
					subs[nextSub] = nextSub
				}
				nextSub = nil

			case s := <-unsub:
				rw := subs[s]
				delete(subs, s)
				select {
				case <-s:
				default:
				}
				chanPool.Put(rw)
			}
		}
	}()

	ctx := &cancelCtx{
		sub:   sub,
		unsub: unsub,
	}

	runtime.AddCleanup(ctx, func(ch rwChan) { close(ch) }, gc)

	return ctx
}

func (c *cancelCtx) GetDone() <-chan struct{} {
	return <-c.sub
}

func (c *cancelCtx) PutDone(done <-chan struct{}) {
	c.unsub <- done
}
