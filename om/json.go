// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package om

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/koonix/x/must"
)

func (m Map[K, V]) MarshalJSON() ([]byte, error) {

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

func itoa(v any) string {
	x := reflect.ValueOf(v)
	switch x.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(x.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(x.Uint(), 10)
	default:
		return ""
	}
}
