// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package omap

import (
	"fmt"

	"go.yaml.in/yaml/v4"
)

func (m Map[K, V]) MarshalYAML() (any, error) {

	node := &yaml.Node{
		Kind: yaml.MappingNode,
	}

	if m.IsNil() {
		return node, nil
	}

	for _, t := range m.s {

		key := &yaml.Node{}
		val := &yaml.Node{}

		err := key.Encode(t.key)
		if err != nil {
			return nil, err
		}

		err = val.Encode(t.val)
		if err != nil {
			return nil, err
		}

		node.Content = append(node.Content, key, val)
	}

	return node, nil
}

func (m *Map[K, V]) UnmarshalYAML(node *yaml.Node) error {

	m.init()

	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("expected yaml mapping node, got %v", node.Kind)
	}

	for i := 0; i < len(node.Content); i += 2 {

		var key K
		var val V

		err := node.Content[i].Decode(&key)
		if err != nil {
			return err
		}

		err = node.Content[i+1].Decode(&val)
		if err != nil {
			return err
		}

		m.Set(key, val)
	}

	return nil
}
