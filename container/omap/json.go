// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package omap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/layer8co/toolbox/must"
)

func (m Map[K, V]) MarshalJSON() ([]byte, error) {

	if m.IsNil() {
		return []byte("null"), nil
	}

	b := new(bytes.Buffer)
	b.WriteByte('{')

	for i, t := range m.s {

		if i > 0 {
			b.WriteByte(',')
		}

		key, err := json.Marshal(t.key)
		if err != nil {
			return nil, err
		}

		if len(key) == 0 || key[0] != '"' {
			s := itoa(t.key)
			if s == "" {
				return nil, fmt.Errorf("unsupported key type %T", t.key)
			}
			key, err = json.Marshal(s)
			if err != nil {
				return nil, err
			}
		}

		b.Write(key)
		b.WriteByte(':')

		val, err := json.Marshal(t.val)
		if err != nil {
			return nil, err
		}

		b.Write(val)
	}

	b.WriteByte('}')
	return b.Bytes(), nil
}

func (m *Map[K, V]) UnmarshalJSON(b []byte) error {

	m.init()

	dec := json.NewDecoder(bytes.NewReader(b))

	t, err := dec.Token()
	if err != nil {
		return err
	}

	delim, ok := t.(json.Delim)
	if !ok || delim != '{' {
		return fmt.Errorf("expected '{', got %v", t)
	}

	for dec.More() {

		t, err := dec.Token()
		if err != nil {
			return err
		}

		keyStr, ok := t.(string)
		if !ok {
			return fmt.Errorf("expected string key, got %T", t)
		}

		var key K
		var val V

		err = json.Unmarshal(must.Get(json.Marshal(keyStr)), &key)
		if err != nil {
			return fmt.Errorf(
				"could not decode key %q into %T: %w",
				keyStr, key, err,
			)
		}

		if err := dec.Decode(&val); err != nil {
			return err
		}

		m.Set(key, val)
	}

	t, err = dec.Token()
	if err != nil {
		return err
	}

	delim, ok = t.(json.Delim)
	if !ok || delim != '}' {
		return fmt.Errorf("expected '}', got %v", t)
	}

	return nil
}

func itoa[T any](v T) string {
	switch v := any(v).(type) {
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uintptr:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	default:
		return ""
	}
}
