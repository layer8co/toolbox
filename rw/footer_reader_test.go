// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package rw_test

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"testing"

	"github.com/koonix/x/rw"
)

func TestFooterReader(t *testing.T) {

	type read struct {
		line    int
		len     int
		content string
		err     error
	}

	tests := []struct {
		line      int
		footerLen int
		stream    string
		footer    string
		reads     []read
	}{
		{
			line(),
			10,
			"the quick brown fox jumps over",
			"jumps over",
			[]read{
				{
					line(),
					15,
					"the quick brown",
					nil,
				},
				{
					line(),
					5,
					" fox ",
					nil,
				},
				{
					line(),
					3,
					"",
					io.EOF,
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test%d-line%d", i, test.line), func(t *testing.T) {

			r := rw.NewFooterReader(strings.NewReader(test.stream), make([]byte, test.footerLen))

			for i, read := range test.reads {
				t.Run(fmt.Sprintf("read%d-line%d", i, read.line), func(t *testing.T) {

					b := make([]byte, read.len)
					n, err := r.Read(b)
					content := string(b[:n])

					if read.content != content {
						t.Errorf(
							"incorrect read content: want %q, got %q",
							read.content, content,
						)
					}

					if read.err != err {
						t.Errorf(
							"incorrect error: want %v, got %v",
							read.err, err,
						)
					}
				})
			}

			want := test.footer
			got := string(r.Footer())
			if want != got {
				t.Errorf(
					"incorrect footer: want %q, got %q",
					want, got,
				)
			}
		})
	}
}

// line returns the line number where it's called.
func line() int {
	_, _, line, _ := runtime.Caller(1)
	return line
}
