// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package kv

import (
	"iter"
	"slices"
)

type Kv[K comparable, V any] struct {
	s []tuple[K, V]
}

type tuple[K comparable, V any] struct {
	key K
	val V
}

func (v Kv[K, V]) Get(key K) (val V, has bool) {
	i := v.index(key)
	if i == -1 {
		return val, false
	}
	return v.s[i].val, true
}

func (v *Kv[K, V]) Set(key K, val V) {
	i := v.index(key)
	if i == -1 {
		v.s = append(v.s, tuple[K, V]{
			key: key,
			val: val,
		})
	} else {
		v.s[i].val = val
	}
}

func (v *Kv[K, V]) Delete(key K) (val V, has bool) {
	i := v.index(key)
	if i == -1 {
		return val, false
	}
	v.s = slices.Delete(v.s, i, i+1)
	return v.s[i].val, true
}

func (v Kv[K, V]) Map() map[K]V {
	m := make(map[K]V, len(v.s))
	for _, t := range v.s {
		m[t.key] = t.val
	}
	return m
}

func (v Kv[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, t := range v.s {
			if !yield(t.key, t.val) {
				return
			}
		}
	}
}

func (v Kv[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for _, t := range v.s {
			if !yield(t.key) {
				return
			}
		}
	}
}

func (v Kv[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, t := range v.s {
			if !yield(t.val) {
				return
			}
		}
	}
}

func (v Kv[K, V]) index(key K) int {
	return slices.IndexFunc(v.s, func(t tuple[K, V]) bool {
		return key == t.key
	})
}
