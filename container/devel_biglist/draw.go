// Copyright 2025 the toolbox authors.
// SPDX-License-Identifier: Apache-2.0

package biglist

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/layer8co/toolbox/io/moreio"
	"github.com/layer8co/toolbox/math/intmath"
)

const (
	nodeRune       = 'O'
	leftSlashRune  = '/'
	rightSlashRune = '\\'
)

type (
	nodeChars[T any] map[*Node[T]]rune
	rowInfo          map[int][]string
)

func (t *List[T]) string(nc nodeChars[T]) string {
	var b bytes.Buffer
	t.draw(&b, nc)
	return b.String()
}

func (t *List[T]) draw(w io.Writer, nc nodeChars[T]) (retErr error) {

	ew := moreio.NewErrorCapturingWriter(w)
	defer func() {
		errors.Join(retErr, ew.Err)
	}()

	if t.IsEmpty() {
		ew.WriteString("<empty>")
		return
	}

	rows := (t.rootNode.height * 2) - 1
	width := intmath.MustPow(2, t.rootNode.height)

	grid := make([][]rune, rows)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	ri := make(rowInfo)

	rootNodeStartingCol := (1 << (t.rootNode.height - 1)) - 1 // (2^(height-1))-1
	t.rootNode.drawNode(0, rootNodeStartingCol, &grid, nc, ri)

	for i, rowRunes := range grid {
		if i != 0 {
			ew.WriteByte('\n')
		}
		info := ri[i]
		s := string(rowRunes)
		if len(info) == 0 {
			s = strings.TrimRight(s, " ")
		}
		ew.WriteString(s)
		if len(info) > 0 {
			ew.WriteString("  ")
			ew.WriteString(strings.Join(info, " "))
		}
	}

	return nil
}

func (n *Node[T]) drawNode(
	row, col int,
	grid *[][]rune,
	nc nodeChars[T],
	ri rowInfo,
) {

	if n == nil {
		return
	}

	if row >= len(*grid) || col >= len((*grid)[0]) {
		panic("out of bounds")
	}

	char, has := nc[n]
	if !has {
		char = nodeRune
	}
	(*grid)[row][col] = char

	s := new(strings.Builder)

	fmt.Fprintf(s, "h=%d,s=%d", n.height, n.getSize())

	switch data := any(n.data).(type) {
	case []byte:
		if n.isLeaf() {
			fmt.Fprintf(s, ",%q", data)
		}
	}

	ri[row] = append(ri[row], s.String())

	if n.isLeaf() {
		return
	}

	// Horizontal distance from parent to the slash
	slashDist := 1
	if n.height > 2 {
		slashDist = 1 << (n.height - 3) // 2^(n.height-3)
	}
	slashRow := row + 1

	// Horizontal distance from parent to child
	childDist := 1 << (n.height - 2) // 2^(n.height-2)
	childRow := row + 2

	if n.left != nil {
		leftSlashCol := col - slashDist
		leftChildCol := col - childDist
		if leftSlashCol >= 0 {
			if leftSlashCol >= len((*grid)[0]) {
				panic("out of bounds")
			}
			(*grid)[slashRow][leftSlashCol] = leftSlashRune
		}
		n.left.drawNode(childRow, leftChildCol, grid, nc, ri)
	}

	if n.right != nil {
		rightSlashCol := col + slashDist
		rightChildCol := col + childDist
		if rightSlashCol >= 0 {
			if rightSlashCol >= len((*grid)[0]) {
				panic("out of bounds")
			}
			(*grid)[slashRow][rightSlashCol] = rightSlashRune
		}
		n.right.drawNode(childRow, rightChildCol, grid, nc, ri)
	}
}
