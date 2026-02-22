// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

// Package biglist implements a list using an AVL order-statistic tree.
package biglist

import (
	"github.com/layer8co/toolbox/math/intmath"
)

type List[T any] struct {
	rootNode    *Node[T]
	firstLeaf   *Node[T]
	maxLeafSize int
}

type Node[T any] struct {
	data   []T
	size   int
	height int
	parent *Node[T]
	left   *Node[T]
	right  *Node[T]
}

func (t *List[T]) IsEmpty() bool {
	return t == nil || t.rootNode == nil
}

func (t *List[T]) Insert(pos int, data []T) {

	// TODO: what if pos is out of bounds?

	// We want to insert lots of data to the middle of
	// a leaf that already has data.
	// First we calculate the amount of data we need to put in the first leaf,
	// and we set that aside (from pos to the end of the leaf's max space) (dataTail).
	// We also set aside that leaf's data from pos to the end of the data (dataHead).
	// We then create leafs and add the data to them.
	// We then add the dataHead to the last leaf that we appended data to.
	// We then add any leftover of dataHead to a new leaf.
	// We then add the dataTail to the first leaf.

	leaf, pos := t.lookupPos(pos)
	tailLeaf := leaf

	x := min(len(data), t.maxLeafSize-pos)
	dataTail := data[:x]
	dataHead := tailLeaf.data[pos:]
	data = data[x:]

	for len(data) > 0 {
		leaf = t.newLeafAfter(leaf)
		x := min(len(data), t.maxLeafSize)
		leaf.data = append(leaf.data, data[:x]...)
		t.balanceAndUpdate(leaf)
		data = data[x:]
	}

	x = min(len(dataHead), t.maxLeafSize-len(leaf.data))
	leaf.data = append(leaf.data, dataHead[:x]...)
	dataHead = dataHead[x:]

	if len(dataHead) > 0 {
		leaf = t.newLeafAfter(leaf)
		leaf.data = append(leaf.data, dataHead...)
		t.balanceAndUpdate(leaf)
	}

	tailLeaf.data = append(tailLeaf.data, dataTail...)
}

func (t *List[T]) Delete(start, end int) {
}

// lookupPos finds the node containing the given position,
// along with the index of the position in that Node.
func (t *List[T]) lookupPos(pos int) (*Node[T], int) {
	node := t.rootNode
	for !node.isLeaf() {
		size := node.left.getSize()
		node = node.left
		if pos <= size {
			size = node.right.getSize()
			node = node.right
		}
		pos = intmath.Abs(pos - size)
	}
	return node, pos
}

// newLeafAfter places a node instead of the given leaf,
// makes the leaf the left child of the node,
// creates a right leaf child for the node,
// balances the tree,
// and returns the new leaf.
func (t *List[T]) newLeafAfter(leaf *Node[T]) (newLeaf *Node[T]) {
	node := &Node[T]{
		size:   len(leaf.data),
		height: 2,
		parent: leaf.parent,
		left:   leaf,
	}
	node.right = &Node[T]{
		height: 1,
		parent: node,
		left:   leaf,
		right:  leaf.right,
	}
	leaf.parent = node
	leaf.right = node.right
	return node.right
}

// balanceAndUpdate moves up the tree from the given node,
// balances the tree, and updates the height and the size of each node
// all the way up to the root.
func (t *List[T]) balanceAndUpdate(n *Node[T]) {
	balanced := false
	for ; n != nil; n = n.parent {
		n.updateHeight()
		n.updateSize()
		if balanced {
			continue
		}
		switch n.getFactor() {
		case 2:
			switch n.right.getFactor() {
			case 0, 1:
				n = n.rotateRight()
			case -1:
				n = n.rotateLeftRight()
			default:
				panic("biglist: invalid AVL tree")
			}
			balanced = true
		case -2:
			switch n.left.getFactor() {
			case 0, -1:
				n = n.rotateLeft()
			case 1:
				n = n.rotateRightLeft()
			default:
				panic("biglist: invalid AVL tree")
			}
			balanced = true
		}
		if n.parent == nil {
			t.rootNode = n
		}
	}
}

func (n *Node[T]) rotateLeftRight() *Node[T] {
	n.right.rotateLeft()
	n.updateHeight()
	n.updateSize()
	return n.rotateRight()
}

func (n *Node[T]) rotateRightLeft() *Node[T] {
	n.left.rotateRight()
	n.updateHeight()
	n.updateSize()
	return n.rotateLeft()
}

func (n *Node[T]) rotateRight() *Node[T] {
	nn := n
	n = n.right
	l := n.left
	n.left = nn
	n.left.right = l
	n.left.updateHeight()
	n.left.updateSize()
	n.updateHeight()
	n.updateSize()
	n.parent = nn.parent
	n.left.parent = n
	if n.parent == nil {
		return n
	}
	if n.parent.right == nn {
		n.parent.right = n
	} else {
		n.parent.left = n
	}
	return n
}

func (n *Node[T]) rotateLeft() *Node[T] {
	nn := n
	n = n.left
	l := n.right
	n.right = nn
	n.right.left = l
	n.right.updateHeight()
	n.right.updateSize()
	n.updateHeight()
	n.updateSize()
	n.parent = nn.parent
	n.right.parent = n
	if n.parent == nil {
		return n
	}
	if n.parent.right == nn {
		n.parent.right = n
	} else {
		n.parent.left = n
	}
	return n
}

func (n *Node[T]) isLeaf() bool {
	return n.height == 1
}

func (n *Node[T]) getFactor() int {
	return n.right.getHeight() - n.left.getHeight()
}

func (n *Node[T]) updateHeight() {
	n.height = n.calculateHeight()
}

func (n *Node[T]) updateSize() {
	n.size = n.calculateSize()
}

// calculateHeight calculates the height of the node
// based on the height of it's children.
func (n *Node[T]) calculateHeight() int {
	if n.isLeaf() {
		return 1
	}
	return max(n.right.getHeight(), n.left.getHeight()) + 1
}

// calculateSize calculates the size of the node
// based on the size of it's children.
func (n *Node[T]) calculateSize() int {
	return n.right.getSize() + n.left.getSize()
}

func (n *Node[T]) getHeight() int {
	if n == nil {
		return 0
	}
	return n.height
}

func (n *Node[T]) getSize() int {
	if n == nil {
		return 0
	}
	if n.isLeaf() {
		return len(n.data)
	}
	return n.size
}

// // n is an extra node you provide to reduce allocations.
// func newTwoLeafNode[T any](
// 	leftData, rightData []T,
// 	prev, next, n *Node[T],
// ) *Node[T] {
// 	*n = Node[T]{}
// 	left := &Node[T]{
// 		data:   leftData,
// 		parent: n,
// 		left:   prev,
// 	}
// 	right := &Node[T]{
// 		data:   rightData,
// 		parent: n,
// 		left:   left,
// 		right:  next,
// 	}
// 	left.right = right
// 	return n
// }

// func (n *Node[T]) All() iter.Seq[[]T] {
// 	list.New()
// 	n = n.getFirstLeaf()
// 	f := n
// 	return func(yield func([]T) bool) {
// 		if n == nil {
// 			return
// 		}
// 		for {
// 			if !yield(n.data) {
// 				return
// 			}
// 			n = n.next
// 			if n == f {
// 				return
// 			}
// 		}
// 	}
// }

// func (n *Node[T]) Append(data ...T) {
// }
//
// func (n *Node[T]) rebalance() *Node[T] {
// 	return createRope(n.getFirstLeaf(), n.leafCount)
// }
//
// func createRope[T any](firstLeaf *Node[T], leafCount int) *Node[T] {
// 	if leafCount <= 1 {
// 		return firstLeaf
// 	}
// 	if leafCount == 2 {
// 		return concat(firstLeaf, firstLeaf.next)
// 	}
// 	middle := leafCount / 2
// 	middleLeaf := firstLeaf
// 	for range middle {
// 		middleLeaf = middleLeaf.next
// 	}
// 	return concat(
// 		createRope(firstLeaf, middle),
// 		createRope(middleLeaf, leafCount-middle),
// 	)
// }
//
// func concat[T any](a, b *Node[T]) *Node[T] {
// 	return &Node[T]{
// 		leftChild:   a,
// 		leftWeight:  a.leftWeight + a.rightWeight,
// 		rightChild:  b,
// 		rightWeight: b.leftWeight + b.rightWeight,
// 	}
// }
//
// func (n *Node[T]) getFirstLeaf() *Node[T] {
// 	if n == nil {
// 		return nil
// 	}
// 	for {
// 		if n.next != nil {
// 			return n
// 		}
// 		n = n.leftChild
// 	}
// }
//
// func (n *Node[T]) isEmpty() bool {
// 	return n == nil || n.leftChild == nil || len(n.data) == 0
// }

// func abs(n int) int {
// 	const intSize = 32 << (^uint(0) >> 63)
// 	const k = intSize - 1
// 	mask := n >> k
// 	return (n ^ mask) - mask
// }

// // balanceAndUpdate moves up the tree and balances the tree
// // and updates the height and the size of each node
// // all the way up to the root.
// func (n *Node[T]) balanceAndUpdate() (newRoot *Node[T]) {
// 	balanced := false
// 	for ; n != nil; n = n.parent {
// 		n.updateHeight()
// 		n.updateSize()
// 		if balanced {
// 			continue
// 		}
// 		f := n.getFactor()
// 		if f >= 2 {
// 			if n.right.getFactor() >= 1 {
// 				n = n.rotateRight()
// 			} else {
// 				n = n.rotateLeftRight()
// 			}
// 			balanced = true
// 		} else if f <= -2 {
// 			if n.right.getFactor() <= -1 {
// 				n = n.rotateLeft()
// 			} else {
// 				n = n.rotateRightLeft()
// 			}
// 			balanced = true
// 		}
// 		if n.parent == nil {
// 			newRoot = n
// 		}
// 	}
// 	return
// }

// func (leaf *Node[T]) divideLeaf() *Node[T] {
// 	n := &Node[T]{
// 		size:   len(leaf.data),
// 		height: 2,
// 		parent: leaf.parent,
// 		left:   leaf,
// 	}
// 	n.right = &Node[T]{
// 		height: 1,
// 		parent: n,
// 		left:   leaf,
// 		right:  leaf.right,
// 	}
// 	leaf.parent = n
// 	n.right.left = n
// 	return n
// }
