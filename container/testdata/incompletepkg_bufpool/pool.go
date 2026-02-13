// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

// Package bufpool implements a dynamic slice pool.
//
// bufpool can be utilized in situations where slices of arbitrary capacity
// are repeatedly required.
//
// bufpool estimates, based on past data, how big of a slice is needed,
// and returns the requested slice from an appropriate bucket.
// The estimation algorithm is optimized to minimize the memory allocation rate.
//
// Having different [PoolUser]s thus allows making different estimations
// for users of [Pool] with different usage patterns.
// In most applications though, there are no meaningful distinct usage patterns,
// and thus one [PoolUser] will suffice.
//
// # Terminology
//
// For the rest of this writeup, a "request" is getting a slice from the pool,
// using it and potentially growing it, and returning it to the pool.
//
// "A request that uses n bytes" thus is getting a slice from the pool,
// growing it to a capacity of n bytes, and returning it to the pool.
//
// # Why
//
// In applications where the required slice capacity of each request
// is known beforehand, a bucketed pool can be utilized.
//
// However, if the capacity requirement of each request is unknown,
// and thus it's unknown which bucket to even use,
// a bucketed pool is even worse than a regular (single-bucket) slice pool.
//
// The reason a slice pool is not ideal is because slices in a slice pool
// only grow in capacity, but never shrink.
//
// For example, if a slice is acquired from a slice pool, grown to 100 megabytes,
// and returned to the pool, that 100 megabyte slice is going to keep circulating
// in the pool and being handed to users, even if from that point on
// the users require no more than 100 bytes.
// The real reason this is a problem is that if the pool is utilized
// frequently enough, the objects in it will never be garbage collected, which is not ideal
// (unless the memory usage of the progam grows so much that the garbage collector
// starts running much more frequently than the pool utilization,
// and of course at this point the GC CPU usage would be through the roof and not ideal).
//
// This growing-but-not-shrinking problem gets worse
// if n goroutines are hitting the pool in parallel,
// and thus there are n slices in the pool at any given time,
// and if even one request uses 100 megabytes,
// the total memory usage of the pool will grow to n * 100 megabytes.
//
// Typically this problem of slice pools is tackled by not returning slices larger
// than a certain amount to the pool, essentially putting a cap
// on the capacity of each slice in the pool. This is ok in simple cases,
// but when the capacity requirement is truly arbitrary, and in each request,
// we don't know ahead of time if it's going to use 1 byte or 1 gigabyte,
// this simple capping approach is also not ideal.
//
// # Implementation
//
// The bufpool operates in n second periods.
// The estimated usage value (let's call it capEstimate) of each user rises immediately,
// but if the maximum usage in a period is less that capEstimate,
// capEstimate will fall to that lesser maximum.
//
// For example, during the first period,
// capEstimate starts out from 0.
// if the user uses 100 bytes once, capEstimate will rise to 100 bytes.
// If then the user uses 500 bytes, it will become 500 bytes.
// For the next period, capEstimate will also be 500 bytes,
// and again will rise as soon as a single usage is higher than it.
// If then in that period the actual usage doesn't rise above 400 bytes,
// capEstimate will fall to 400 bytes for the next period.
//
// Hopefully it's obvious why capEstimate needs to rise immediately;
// in Go, growing a slice with a capacity of 100 bytes to 101 bytes
// will cause a new 101 byte backing array to be allocated,
// so there is no value in estimations that would be less than the usage.
//
// The reason capEstimate falls to the lesser maximum in one fell swoop,
// instead of, say, walk back to the lesser maximum slowly,
// is because walking back would cause every bucket on the way back
// to be accessed, causing unnecessary memory usage and allocations.
// Beter to just jump to the bucket that we think is the right one.
package bufpool

import (
	"math/bits"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// In the worst case, if you set this value to e.g. 10 seconds,
	// the maximum rate of allocations will be limited to about
	// one allocation every 10 seconds.
	estimationPeriod = 10 * time.Second

	// The number of periods a user needs to be idle
	// for it to be considered idle.
	userIdlePeriods = 6

	// Make the first bucket be for sizes less than or equal 2^minBucketBits.
	minBucketBits = 4

	minBucketSize = 1 << minBucketBits
	bucketsCount  = strconv.IntSize - minBucketBits
)

var testCleanup func() = nil

type Pool[T any] struct {
	buckets  [bucketsCount]sync.Pool
	ticker   *time.Ticker
	mu       *sync.Mutex
	users    map[*PoolUser[T]]struct{}
	userPool sync.Pool
}

type PoolUser[T any] struct {
	pool             *Pool[T]
	sizeEstimate     atomic.Int64
	sizeEstimateNext atomic.Int64
	registered       atomic.Bool
	idlePeriods      int
}

func New[T any]() *Pool[T] {

	ticker := time.NewTicker(estimationPeriod)
	ticker.Stop()

	mu := new(sync.Mutex)
	users := make(map[*PoolUser[T]]struct{})

	p := &Pool[T]{
		ticker: ticker,
		mu:     mu,
		users:  users,
		userPool: sync.Pool{
			New: func() any {
				return &PoolUser[T]{}
			},
		},
	}
	for i := range p.buckets {
		p.buckets[i].New = func() any {
			b := make([]T, 0, bucketSize(i))
			return &b
		}
	}

	tick := func() {

		mu.Lock()
		defer mu.Unlock()

		for u := range users {

			next := u.sizeEstimateNext.Swap(0)
			u.sizeEstimate.Store(next)

			if isIdle := next == 0; !isIdle {
				u.idlePeriods = 0
				continue
			}

			u.idlePeriods++

			if u.idlePeriods >= userIdlePeriods {
				u.idlePeriods = 0
				delete(users, u)
				u.registered.Store(false)
			}
		}

		if len(users) == 0 {
			ticker.Stop()
		}
	}

	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				tick()
			case <-done:
				if testCleanup != nil {
					testCleanup()
				}
				return
			}
		}
	}()

	runtime.AddCleanup(p, func(ch chan struct{}) { close(ch) }, done)

	return p
}

// Calling [PoolUser.Done]
// merely saves on allocations and is not necessary.
func (p *Pool[T]) NewUser() *PoolUser[T] {
	u := p.userPool.Get().(*PoolUser[T])
	u.pool = p
	return u
}

func (p *Pool[T]) get(size int) *[]T {
	return p.buckets[bucketIndex(size)].Get().(*[]T)
}

func (p *Pool[T]) put(b *[]T) {
	p.buckets[bucketIndex(cap(*b))].Put(b)
}

// ==================================================

func (u *PoolUser[T]) Get(size ...int) *[]T {
	if len(size) > 0 {
		if size[0] < 0 {
			panic("bufpool.PoolUser.Get: size[0] < 0")
		}
		return u.pool.get(size[0])
	}
	return u.pool.get(int(u.sizeEstimate.Load()))
}

// A negative size bypasses updating the size estimation values.
func (u *PoolUser[T]) Put(b *[]T, size int) {
	if size >= 0 {
		updateIfIncrease(&u.sizeEstimate, int64(size))
		updateIfIncrease(&u.sizeEstimateNext, int64(size))
		if !u.registered.Load() {
			u.register()
		}
	}
	*b = (*b)[:0]
	u.pool.put(b)
}

func (u *PoolUser[T]) Reset() {
	// Because of the lock-free way we update these atomics
	// in the background goroutine, it's important here
	// that we zero sizeEstimateNext before sizeEstimate.
	u.sizeEstimateNext.Store(0)
	u.sizeEstimate.Store(0)
}

func (u *PoolUser[T]) Done() {

	p := u.pool

	if p == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.users, u)
	if len(p.users) == 0 {
		p.ticker.Stop()
	}

	*u = PoolUser[T]{}
	p.userPool.Put(u)
}

func (u *PoolUser[T]) register() {

	p := u.pool

	p.mu.Lock()
	defer p.mu.Unlock()

	if u.registered.Load() {
		return
	}

	p.users[u] = struct{}{}
	u.registered.Store(true)
	if len(p.users) == 1 {
		p.ticker.Reset(estimationPeriod)
	}
}

// ==================================================

func bucketIndex(size int) int {
	if size <= minBucketSize {
		return 0
	}
	return bits.Len(uint(size-1)) - minBucketBits
}

func bucketSize(index int) int {
	// TODO: This overflows on (index + minBucketBits) == 0;
	// it should be `(1 << (index + minBucketSize)) - 1`,
	// but then bucketIndex would need adjustments to
	// to correctly report the index of each given size.
	// At the end of the day, we should make bucketIndex(math.MaxInt)
	// possible without overflows.
	return 1 << (index + minBucketBits)
}

// updateIfIncrease atomically updates the given atomic.Int64
// if new is larger than it.
func updateIfIncrease(a *atomic.Int64, new int64) {
	// Atomically compare and update the atomic using a CAS loop.
	for {
		old := a.Load()
		if new <= old {
			return
		}
		if a.CompareAndSwap(old, new) {
			return
		}
	}
}
