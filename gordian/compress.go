package gordian

import (
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"io"
	"regexp"
)

type Bzip2 struct{}

func (g *Bzip2) Init(_ *regexp.Regexp) {}

func (g *Bzip2) Pipe(in io.Reader, out io.WriteCloser) error {
	in = bzip2.NewReader(in)
	io.Copy(out, in)
	return out.Close()
}

type Gzip struct{}

func (g *Gzip) Init(_ *regexp.Regexp) {}

func (g *Gzip) Pipe(in io.Reader, out io.WriteCloser) error {
	defer out.Close()
	reader, err := gzip.NewReader(in)
	if err != nil {
		return err
	}
	io.Copy(out, reader)
	return reader.Close()
}

type MaybeBzip2 Bzip2

func (g *MaybeBzip2) Init(_ *regexp.Regexp) {}

func (g *MaybeBzip2) Pipe(in io.Reader, out io.WriteCloser) error {
	const bz2Header = "\x42\x5a\x68"

	header := make([]byte, len(bz2Header))
	in.Read(header)

	in = io.MultiReader(bytes.NewReader(header), in)
	if string(header) == bz2Header {
		return (*Bzip2)(g).Pipe(in, out)
	}

	io.Copy(out, in)
	return out.Close()
}

type MaybeGzip Gzip

func (g *MaybeGzip) Init(_ *regexp.Regexp) {}

func (g *MaybeGzip) Pipe(in io.Reader, out io.WriteCloser) error {
	const gzHeader = "\x1f\x8b"

	header := make([]byte, len(gzHeader))
	in.Read(header)

	in = io.MultiReader(bytes.NewReader(header), in)
	if string(header) == gzHeader {
		return (*Gzip)(g).Pipe(in, out)
	}

	io.Copy(out, in)
	return out.Close()
}
