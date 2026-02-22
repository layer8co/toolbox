// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package biglist

import (
	"fmt"
	"testing"
)

func TestInsert(t *testing.T) {
	l := &List[byte]{
		rootNode:    mkNode[byte](),
		maxLeafSize: 10,
	}
	fmt.Println(l.string(nil))

	l.Insert(0, []byte("hello"))
	fmt.Println(l.string(nil))
	fmt.Println("f:", l.rootNode.getFactor())

	l.Insert(5, []byte("prettyworld"))
	fmt.Println(l.string(nil))
}
