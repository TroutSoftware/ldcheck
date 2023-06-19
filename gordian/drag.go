package gordian

import (
	"bufio"
	"io"
	"regexp"
)

type DragEnd struct{ re *regexp.Regexp }

func (d *DragEnd) Init(re *regexp.Regexp) { d.re = re }

func (d *DragEnd) Pipe(in io.Reader, out io.WriteCloser) error {
	scn := bufio.NewScanner(in)
	var match string

	for scn.Scan() {
		line := scn.Bytes()
		m := d.re.FindIndex(line)
		if m != nil {
			match = string(line[m[0]:m[1]])
			out.Write(line)
		} else {
			out.Write(line)
			io.WriteString(out, match)
		}
		io.WriteString(out, "\n")
	}

	if err := scn.Err(); err != nil {
		return err
	}

	return out.Close()
}
