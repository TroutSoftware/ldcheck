package gordian

import (
	"bufio"
	"compress/bzip2"
	"io"
	"regexp"
)

type Bzip2 struct{}

func (g *Bzip2) Init(r *regexp.Regexp) {}

func (g *Bzip2) Pipe(in io.Reader, out io.WriteCloser) error {
	defer out.Close()

	// Create a bzip2 reader that reads from the input pipe reader
	reader := bzip2.NewReader(in)
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
