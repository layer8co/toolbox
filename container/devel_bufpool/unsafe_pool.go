// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package bufpool

import (
	"sync"
	"sync/atomic"
	"time"
)

// Any references stored in the items
// that you return to this pool would be stored as bytes
// and not be able to be detected by the GC,
// thus if you DO want to store references in the items
// you return to the pool, and you want the GC to
// not, well, GC them, make sure they have references
// elsewhere.
// Or just don't store references in the items you return to the pool,
// (or do not access the references that are stored in the items
// you get from the pool).
type UnsafePool struct {
	buckets  [64]sync.Pool
	ticker   *time.Ticker
	mu       *sync.Mutex
	users    map[*UnsafePoolUser]struct{}
	userPool sync.Pool
}

type UnsafePoolUser struct {
	pool             *UnsafePool
	sizeEstimate     atomic.Int64
	sizeEstimateNext atomic.Int64
}
