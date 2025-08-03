// Copyright 2025 the github.com/koonix/x authors.
// SPDX-License-Identifier: Apache-2.0

package text_test

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/koonix/golibs/text"
)

func TestTruncateRuneRepeats(t *testing.T) {
	tests := []struct {
		line   int
		n      int
		inputs []string
		want   string
	}{
		{
			line(),
			3,
			[]string{
				"XXXXX--",         // 1--
				"--XXXXXX--XX--X", // --2--3--4
				"--X",             // --5
				"XXX--XX",         // 5--6
				"XX--XXX",         // 6--7
				"X",               // 7
			},
			// 1    2    3   4  5    6    7
			"XXX----XXX--XX--X--XXX--XXX--XXX",
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("test%d-line%d", i, test.line), func(t *testing.T) {
			sb := new(strings.Builder)
			filter := text.TruncateRuneRepeats('X', test.n)
			for _, v := range test.inputs {
				sb.Write(filter([]byte(v)))
			}
			got := sb.String()
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("incorrect result (-want +got):\n%s", diff)
			}
		})
	}
}

// line returns the line number where it's called.
func line() int {
	_, _, line, _ := runtime.Caller(1)
	return line
}
