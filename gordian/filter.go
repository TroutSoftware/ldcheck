package main

import (
	"bufio"
	"io"
	"regexp"
)

type Ignore struct{ re *regexp.Regexp }
type Only struct{ re *regexp.Regexp }

func (i *Ignore) Init(re *regexp.Regexp) { i.re = re }
func (o *Only) Init(re *regexp.Regexp)   { o.re = re }

func (i *Ignore) Pipe(in io.Reader, out io.WriteCloser) error {
	scn := bufio.NewScanner(in)
	for scn.Scan() {
		line := scn.Bytes()
		m := i.re.FindIndex(line)
		if m == nil {
			out.Write(line)
		} else {
			out.Write(line[:m[0]])
			out.Write(line[m[1]:])
		}
		io.WriteString(out, "\n")
	}

	if err := scn.Err(); err != nil {
		return err
	}

	return out.Close()
}

func (o *Only) Pipe(in io.Reader, out io.WriteCloser) error {
	scn := bufio.NewScanner(in)
	for scn.Scan() {
		line := scn.Bytes()
		m := o.re.FindIndex(line)
		if m == nil {
			continue
		} else {
			out.Write(line[:m[0]])
			out.Write(line[m[1]:])
		}
		io.WriteString(out, "\n")
	}

	if err := scn.Err(); err != nil {
		return err
	}

	return out.Close()
}
