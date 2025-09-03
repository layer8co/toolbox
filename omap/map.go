// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

// Package omap provides an ordered map implementation.
package omap

import (
	"iter"
	"slices"
)

// Map is an ordered map.
type Map[K comparable, V any] struct {
	*_map[K, V]
}

type _map[K comparable, V any] struct {
	s []tuple[K, V]
}

type tuple[K comparable, V any] struct {
	key K
	val V
}

func New[K comparable, V any](size ...int) Map[K, V] {
	m := Map[K, V]{
		_map: new(_map[K, V]),
	}
	if len(size) > 0 {
		m.s = make([]tuple[K, V], 0, size[0])
	}
	return m
}

func (m *Map[K, V]) init() {
	if m._map == nil {
		m._map = new(_map[K, V])
	}
}

func (m Map[K, V]) Get(key K) (val V, has bool) {
	i := m.index(key)
	if i == -1 {
		return val, false
	}
	return m.s[i].val, true
}

func (m *Map[K, V]) Set(key K, val V) {
	i := m.index(key)
	if i == -1 {
		m.s = append(m.s, tuple[K, V]{
			key: key,
			val: val,
		})
	} else {
		m.s[i].val = val
	}
}

func (m *Map[K, V]) Delete(key K) (val V, has bool) {
	i := m.index(key)
	if i == -1 {
		return val, false
	}
	m.s = slices.Delete(m.s, i, i+1)
	return m.s[i].val, true
}

func (m Map[K, V]) Map() map[K]V {
	x := make(map[K]V, len(m.s))
	for _, t := range m.s {
		x[t.key] = t.val
	}
	return x
}

func (m Map[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, t := range m.s {
			if !yield(t.key, t.val) {
				return
			}
		}
	}
}

func (m Map[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for _, t := range m.s {
			if !yield(t.key) {
				return
			}
		}
	}
}

func (m Map[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, t := range m.s {
			if !yield(t.val) {
				return
			}
		}
	}
}

func (m Map[K, V]) index(key K) int {
	return slices.IndexFunc(m.s, func(t tuple[K, V]) bool {
		return key == t.key
	})
}
