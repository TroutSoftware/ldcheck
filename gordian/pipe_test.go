package gordian

import (
	"bytes"
	"io"
	"testing"
)

func TestPipe1(t *testing.T) {
	in := ` un grand             texte avec          des              espaces`
	r, w := Pipe()

	go func() {
		w.Write([]byte(in))
		w.Close()
	}()

	var buf bytes.Buffer
	removeDupWhiteSpaces(r, nopCloseWriter{&buf})

	t.Log(buf.String())
}

type nopCloseWriter struct{ io.Writer }

func (nopCloseWriter) Close() error                   { return nil }
func (w nopCloseWriter) Write(dt []byte) (int, error) { return w.Writer.Write(dt) }

var whitespace = [256]bool{'\n': true, ' ': true, '\t': true, '\r': true}

func removeDupWhiteSpaces(in *Reader, out io.WriteCloser) error {
	const lookahead = 1024
	// invariant: the loop starts at a non-white space after the first iteration
	for in.Next(lookahead) {
		w := in.Window()
		i := bytes.IndexAny(w, "\n \t\r")

		if i == -1 {
			out.Write(w)
			in.Release(len(w))
			continue
		} else {
			out.Write(w[:i+1])
			i++
		}

		for i < len(w) && whitespace[w[i]] {
			i++

			if i == len(w) {
				if !in.Next(lookahead) {
					return out.Close()
				}
				w = in.Window()
			}
		}
		in.Release(i)
	}

	return out.Close()
}

func TestWrapAround(t *testing.T) {
	r, w := Pipe()
	go func() {
		for i := 0; i < 100; i++ {
			w.Write([]byte("a"))
		}
	}()

	for i := 0; i < 10; i++ {
		if !r.Next(10) {
			t.Fatal("small read")
		}

		t.Logf("reading window %s", r.Window())
		r.Release(10)
	}
}
