// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package biglist

import (
	"fmt"
	"strings"
	"testing"
)

func TestBalanceRight(t *testing.T) {

	root, src, nr, wantBefore, wantAfter := testBalanceRight1[int]()
	testBalance(t, root, src, nr, wantBefore, wantAfter)

	root, src, nr, wantBefore, wantAfter = testBalanceRight2[int]()
	testBalance(t, root, src, nr, wantBefore, wantAfter)

	root, src, nr, wantBefore, wantAfter = testBalanceRight3[int]()
	testBalance(t, root, src, nr, wantBefore, wantAfter)

	root, src, nr, wantBefore, wantAfter = testBalanceLeft1[int]()
	testBalance(t, root, src, nr, wantBefore, wantAfter)

	root, src, nr, wantBefore, wantAfter = testBalanceLeft2[int]()
	testBalance(t, root, src, nr, wantBefore, wantAfter)

	root, src, nr, wantBefore, wantAfter = testBalanceLeft3[int]()
	testBalance(t, root, src, nr, wantBefore, wantAfter)
}

func TestNewLeafAfter(t *testing.T) {
	testNewLeafAfter[int]()
}

func testNewLeafAfter[T any]() {
	t := &List[T]{
		rootNode: mkLeaf[T](10),
	}
	fmt.Println(t.string(nil))
	t.balanceAndUpdate(t.newLeafAfter(t.rootNode))
	fmt.Println(t.string(nil))
}

func testBalanceRight1[T any]() (
	root, src *Node[T],
	nr nodeChars[T],
	wantBefore, wantAfter string,
) {

	wantBefore = `
       x          h=4,s=40
     /   \
   a       y      h=1,s=10 h=3,s=30
          / \
         b   z    h=1,s=10 h=2,s=20
            / \
            c d   h=1,s=10 h=1,s=10
`
	wantAfter = `
   y      h=3,s=40
  / \
 x   z    h=2,s=20 h=2,s=20
/ \ / \
a b c d   h=1,s=10 h=1,s=10 h=1,s=10 h=1,s=10
`

	nr = make(nodeChars[T])

	z := mkNode(mkLeaf[T](10), mkLeaf[T](10))
	nr[z] = 'z'
	nr[z.left] = 'c'
	nr[z.right] = 'd'

	y := mkNode(mkLeaf[T](10), z)
	nr[y] = 'y'
	nr[y.left] = 'b'

	x := mkNode(mkLeaf[T](10), y)
	nr[x] = 'x'
	nr[x.left] = 'a'

	return x, z, nr, wantBefore, wantAfter
}

func testBalanceRight2[T any]() (
	root, src *Node[T],
	nr nodeChars[T],
	wantBefore, wantAfter string,
) {

	wantBefore = `
       x          h=4,s=30
     /   \
   a       z      h=1,s=10 h=3,s=20
          / \
         y   c    h=2,s=10 h=1,s=10
        /
        b         h=1,s=10
`
	wantAfter = `
   y      h=3,s=30
  / \
 x   z    h=2,s=20 h=2,s=10
/ \   \
a b   c   h=1,s=10 h=1,s=10 h=1,s=10
`

	nr = make(nodeChars[T])

	y := mkNode(mkLeaf[T](10))
	nr[y] = 'y'
	nr[y.left] = 'b'

	z := mkNode(y, mkLeaf[T](10))
	nr[z] = 'z'
	nr[z.right] = 'c'

	x := mkNode(mkLeaf[T](10), z)
	nr[x] = 'x'
	nr[x.left] = 'a'

	return x, y, nr, wantBefore, wantAfter
}

func testBalanceRight3[T any]() (
	root, src *Node[T],
	nr nodeChars[T],
	wantBefore, wantAfter string,
) {

	wantBefore = `
   x      h=3,s=10
    \
     y    h=2,s=10
      \
      z   h=1,s=10
`
	wantAfter = `
 y    h=2,s=10
/ \
x z   h=1,s=0 h=1,s=10
`

	nr = make(nodeChars[T])

	z := mkLeaf[T](10)
	nr[z] = 'z'

	y := mkNode(nil, z)
	nr[y] = 'y'

	x := mkNode(nil, y)
	nr[x] = 'x'

	return x, z, nr, wantBefore, wantAfter
}

func testBalanceLeft1[T any]() (
	root, src *Node[T],
	nr nodeChars[T],
	wantBefore, wantAfter string,
) {

	wantBefore = `
       x          h=4,s=40
     /   \
   y       a      h=3,s=30 h=1,s=10
  / \
 z   b            h=2,s=20 h=1,s=10
/ \
d c               h=1,s=10 h=1,s=10
`
	wantAfter = `
   y      h=3,s=40
  / \
 z   x    h=2,s=20 h=2,s=20
/ \ / \
d c b a   h=1,s=10 h=1,s=10 h=1,s=10 h=1,s=10
`

	nr = make(nodeChars[T])

	z := mkNode(mkLeaf[T](10), mkLeaf[T](10))
	nr[z] = 'z'
	nr[z.left] = 'd'
	nr[z.right] = 'c'

	y := mkNode(z, mkLeaf[T](10))
	nr[y] = 'y'
	nr[y.right] = 'b'

	x := mkNode(y, mkLeaf[T](10))
	nr[x] = 'x'
	nr[x.right] = 'a'

	return x, z, nr, wantBefore, wantAfter
}

func testBalanceLeft2[T any]() (
	root, src *Node[T],
	nr nodeChars[T],
	wantBefore, wantAfter string,
) {

	wantBefore = `
       x          h=4,s=30
     /   \
   z       a      h=3,s=20 h=1,s=10
  / \
 c   y            h=1,s=10 h=2,s=10
      \
      b           h=1,s=10
`
	wantAfter = `
   y      h=3,s=30
  / \
 z   x    h=2,s=10 h=2,s=20
/   / \
c   b a   h=1,s=10 h=1,s=10 h=1,s=10
`

	nr = make(nodeChars[T])

	y := mkNode(nil, mkLeaf[T](10))
	nr[y] = 'y'
	nr[y.right] = 'b'

	z := mkNode(mkLeaf[T](10), y)
	nr[z] = 'z'
	nr[z.left] = 'c'

	x := mkNode(z, mkLeaf[T](10))
	nr[x] = 'x'
	nr[x.right] = 'a'

	return x, y, nr, wantBefore, wantAfter
}

func testBalanceLeft3[T any]() (
	root, src *Node[T],
	nr nodeChars[T],
	wantBefore, wantAfter string,
) {

	wantBefore = `
   x      h=3,s=10
  /
 y        h=2,s=10
/
z         h=1,s=10
`
	wantAfter = `
 y    h=2,s=10
/ \
z x   h=1,s=10 h=1,s=0
`

	nr = make(nodeChars[T])

	z := mkLeaf[T](10)
	nr[z] = 'z'

	y := mkNode(z)
	nr[y] = 'y'

	x := mkNode(y)
	nr[x] = 'x'

	return x, z, nr, wantBefore, wantAfter
}

func mkNode[T any](leftRight ...*Node[T]) *Node[T] {
	n := &Node[T]{}
	if len(leftRight) >= 2 {
		n.right = leftRight[1]
	}
	if len(leftRight) >= 1 {
		n.left = leftRight[0]
	}
	n.updateHeight()
	n.updateSize()
	for _, lr := range leftRight {
		if lr != nil {
			lr.parent = n
		}
	}
	return n
}

func mkLeaf[T any](size int) *Node[T] {
	return &Node[T]{
		height: 1,
		data:   make([]T, size),
	}
}

func testBalance[T any](
	t *testing.T,
	rootNode, balanceNode *Node[T],
	nr nodeChars[T],
	wantBefore, wantAfter string,
) {
	t.Helper()

	wantBefore = strings.Trim(wantBefore, "\n")
	wantAfter = strings.Trim(wantAfter, "\n")

	tree := &List[T]{
		rootNode: rootNode,
	}

	gotBefore := tree.string(nr)
	if wantBefore != gotBefore {
		t.Errorf(
			"incorrect before value:\nwant:\n%s\ngot:\n%s\n",
			wantBefore, gotBefore,
		)
	}

	tree.balanceAndUpdate(balanceNode)

	gotAfter := tree.string(nr)
	if wantAfter != gotAfter {
		t.Errorf(
			"incorrect after value:\nwant:\n%s\ngot:\n%s\n",
			wantAfter, gotAfter,
		)
	}
}
