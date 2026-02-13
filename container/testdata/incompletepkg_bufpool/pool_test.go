// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package bufpool_test

import (
	"math/rand/v2"
	"slices"
	"sync"
	"testing"

	bufpool "github.com/layer8co/toolbox/container/testdata/incompletepkg_bufpool"
)

// TODO: The tests need to be closer to the real world.
// Don't just raw dog things in the benchmark loop,
// actually create a function that utilizies the pool
// similar to how it would be used in a real world setting.
// Don't just measure alloc/sec; measure memory usage after GC too.
//
// The issue with the current benches is that the regular single-bucket pool
// very quickly grows to the max, and so the real shrinking capabilities
// of our stuff does not shine through.
// Also, because the growth with slices.Grow happens chunkily,
// the bucketed nature of our stuff doesn't show itself.

func BenchmarkGetPut(b *testing.B) {

	b.Run("bufpool", func(b *testing.B) {
		pool := bufpool.New[byte]()
		user := pool.NewUser()
		for b.Loop() {
			buf := user.Get()
			*buf = append(*buf, "ayylmao"...)
			user.Put(buf, len(*buf))
		}
	})

	b.Run("stdlib", func(b *testing.B) {
		pool := sync.Pool{
			New: func() any {
				b := make([]byte, 0, 16)
				return &b
			},
		}
		for b.Loop() {
			buf := pool.Get().(*[]byte)
			*buf = append(*buf, "ayylmao"...)
			pool.Put(buf)
		}
	})
}

func BenchmarkGetPutRising(b *testing.B) {

	b.Run("bufpool", func(b *testing.B) {
		pool := bufpool.New[byte]()
		user := pool.NewUser()
		i := uint16(0)
		for b.Loop() {
			buf := user.Get()
			*buf = slices.Grow(*buf, int(i))
			user.Put(buf, cap(*buf))
			i++
		}
	})

	b.Run("stdlib", func(b *testing.B) {
		pool := sync.Pool{
			New: func() any {
				b := make([]byte, 0, 16)
				return &b
			},
		}
		i := uint16(0)
		for b.Loop() {
			buf := pool.Get().(*[]byte)
			*buf = slices.Grow(*buf, int(i))
			pool.Put(buf)
			i++
		}
	})
}

func BenchmarkGetPutRand(b *testing.B) {

	x := 100
	y := 100_000

	b.Run("bufpool", func(b *testing.B) {
		pool := bufpool.New[byte]()
		user := pool.NewUser()
		for b.Loop() {
			buf := user.Get()
			*buf = slices.Grow(*buf, randBetween(x, y))
			user.Put(buf, cap(*buf))
		}
	})

	b.Run("stdlib", func(b *testing.B) {
		pool := sync.Pool{
			New: func() any {
				b := make([]byte, 0, 16)
				return &b
			},
		}
		for b.Loop() {
			buf := pool.Get().(*[]byte)
			*buf = slices.Grow(*buf, randBetween(x, y))
			pool.Put(buf)
		}
	})
}

// randBetween returns random number n such that a <= n <= b.
func randBetween(a, b int) int {
	if a > b {
		panic("a > b")
	}
	return rand.IntN(b-a+1) + a
}

// func grow[T any](b []T, n int) []T {
// 	if cap(b) >= n {
// 		return b[:n]
// 	}
// 	return make([]T, n)
// }

// func BenchmarkPoolCap(b *testing.B) {
//
// 	pool := bufpool.New[byte]()
// 	expectedCap := 512
//
// 	for b.Loop() {
// 		buf := pool.Get(expectedCap)
// 		*buf.Buf = append(*buf.Buf, "ayylmao"...)
// 		buf.Put()
// 	}
// }
//
// func BenchmarkPoolId(b *testing.B) {
//
// 	pool := bufpool.New[byte]()
// 	uniq := new(byte)
//
// 	for b.Loop() {
// 		buf := pool.Get(uniq)
// 		*buf.Buf = append(*buf.Buf, "wut"...)
// 		buf.Put()
// 	}
// }
