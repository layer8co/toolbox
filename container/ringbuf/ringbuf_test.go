package ringbuf_test

import (
	"bytes"
	"fmt"
	"io"
	"iter"
	"runtime"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/layer8co/toolbox/container/ringbuf"
)

type testCase struct {
	line      int
	ops       func(b *ringbuf.ByteBuffer) (read []byte)
	wantRead  string
	wantBytes string
}

func TestRingBuf(t *testing.T) {
	runTests(t, 15, nil, []testCase{
		{
			line(),
			func(b *ringbuf.ByteBuffer) []byte { return nil },
			"",
			"",
		},
		{
			line(),
			func(b *ringbuf.ByteBuffer) []byte {
				b.Write([]byte("hello "))
				b.WriteString("world")
				return nil
			},
			"",
			"hello world",
		},
		{
			line(),
			func(b *ringbuf.ByteBuffer) []byte {
				b.Write([]byte(" ab"))
				b.WriteString("cxyz")
				return nil
			},
			"",
			"lo world abcxyz",
		},
	})
}

func runTests(t *testing.T, maxLen int, initialCap *int, testCases []testCase) {

	t.Helper()

	ic := []int{}
	if initialCap != nil {
		ic = append(ic, *initialCap)
	}

	b := ringbuf.NewByteBuffer(maxLen, ic...)

	for i, tt := range testCases {
		t.Run(fmt.Sprintf("test%d-line%d", i, tt.line), func(t *testing.T) {

			gotRead := tt.ops(b)
			diff(t, "wantRead, gotRead", tt.wantRead, gotRead)

			assertEqual(t, "b.Len", b.Len(), len(tt.wantBytes))

			gotBytes := b.Bytes()
			diff(t, "b.Bytes", tt.wantBytes, gotBytes)

			gotBytesSeq := readBytesSeq(b.BytesSeq())
			diff(t, "b.BytesSeq", tt.wantBytes, gotBytesSeq)

			gotString := b.String()
			diff(t, "b.String", tt.wantBytes, gotString)

			gotReadAt := readReaderAt(b)
			diff(t, "b.ReadAt", tt.wantBytes, gotReadAt)
		})
	}
}

func assertEqual[T comparable](t *testing.T, title string, want, got T) {
	t.Helper()
	if want != got {
		t.Errorf("%s: want %v, got %v", title, want, got)
	}
}

func diff[A, B string | []byte](t *testing.T, title string, want A, got B) {
	t.Helper()
	sWant := string(want)
	sGot := string(got)
	if diff := cmp.Diff(sWant, sGot); diff != "" {
		t.Errorf("%s: incorrect result (-want +got):\n%s", title, diff)
	}
}

func readBytesSeq(seq iter.Seq[[]byte]) (b []byte) {
	for s := range seq {
		b = append(b, s...)
	}
	return b
}

func readWriterTo(w io.WriterTo) []byte {
	buf := new(bytes.Buffer)
	w.WriteTo(buf)
	return buf.Bytes()
}

func readReaderAt(r io.ReaderAt) (b []byte) {
	off := 0
	for {
		if off == cap(b) {
			b = slices.Grow(b, 128)
		}
		n, err := r.ReadAt(b[off:cap(b)], int64(off))
		off += n
		b = b[:off]
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
	}
	return b
}

// returns the line number where it's called.
func line() int {
	_, _, line, _ := runtime.Caller(1)
	return line
}

// returns a pointer containing the given var.
func p[T any](v T) *T {
	return &v
}
