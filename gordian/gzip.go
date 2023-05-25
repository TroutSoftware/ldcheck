package gordian

import (
	"bufio"
	"compress/gzip"
	"io"
	"regexp"
)

type Gzip struct{}

func (g *Gzip) Init(r *regexp.Regexp) {}

func (g *Gzip) Pipe(in io.Reader, out io.WriteCloser) error {
	defer out.Close()

	// Create a gzip reader that reads from the input pipe reader
	reader, err := gzip.NewReader(in)
	if err != nil {
		return err
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Bytes()
		out.Write(line)
		io.WriteString(out, "\n")
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return out.Close()
}
