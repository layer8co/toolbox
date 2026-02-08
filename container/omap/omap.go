// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

// Package omap provides an ordered map implementation.
package omap

import (
	"fmt"
	"iter"
	"slices"
	"strings"
)

// Map is an ordered map.
type Map[K comparable, V any] struct {
	*omap[K, V]
}

type omap[K comparable, V any] struct {
	s []tuple[K, V]
}

type tuple[K comparable, V any] struct {
	key K
	val V
}

func New[K comparable, V any](size ...int) Map[K, V] {
	m := Map[K, V]{
		omap: new(omap[K, V]),
	}
	if len(size) > 0 {
		m.s = make([]tuple[K, V], 0, size[0])
	}
	return m
}

func Init[K comparable, V any](m *Map[K, V], size ...int) {
	if m.IsNil() {
		*m = New[K, V](size...)
	}
}

func (m Map[K, V]) IsNil() bool {
	return m.omap == nil
}

func (m Map[K, V]) Get(key K) (val V, has bool) {
	if m.IsNil() {
		return val, false
	}
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
	if m.IsNil() {
		return val, false
	}
	i := m.index(key)
	if i == -1 {
		return val, false
	}
	m.s = slices.Delete(m.s, i, i+1)
	return m.s[i].val, true
}

func (m Map[K, V]) Len() int {
	if m.IsNil() {
		return 0
	}
	return len(m.s)
}

func (m Map[K, V]) Map() map[K]V {
	if m.IsNil() {
		return nil
	}
	x := make(map[K]V, len(m.s))
	for _, t := range m.s {
		x[t.key] = t.val
	}
	return x
}

func (m Map[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		if m.IsNil() {
			return
		}
		for _, t := range m.s {
			if !yield(t.key, t.val) {
				return
			}
		}
	}
}

func (m Map[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		if m.IsNil() {
			return
		}
		for _, t := range m.s {
			if !yield(t.key) {
				return
			}
		}
	}
}

func (m Map[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		if m.IsNil() {
			return
		}
		for _, t := range m.s {
			if !yield(t.val) {
				return
			}
		}
	}
}

func (m Map[K, V]) String() string {
	if m.IsNil() {
		return "omap[]"
	}
	var sb strings.Builder
	sb.WriteString("omap[")
	f := "%v:%v"
	for i, t := range m.s {
		if i == 1 {
			f = " " + f
		}
		fmt.Fprintf(&sb, f, t.key, t.val)
	}
	sb.WriteString("]")
	return sb.String()
}

func (m *Map[K, V]) init() {
	if m.omap == nil {
		m.omap = new(omap[K, V])
	}
}

func (m Map[K, V]) index(key K) int {
	return slices.IndexFunc(m.s, func(t tuple[K, V]) bool {
		return key == t.key
	})
}
