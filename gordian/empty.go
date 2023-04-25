package gordian

import (
	"bufio"
	"io"
	"regexp"
)

type NoEmpty struct{}

func (*NoEmpty) Init(_ *regexp.Regexp) {}
func (e *NoEmpty) Pipe(in io.Reader, out io.WriteCloser) error {
	scn := bufio.NewScanner(in)
	for scn.Scan() {
		if len(tleft(scn.Bytes())) == 0 {
			continue
		}

		out.Write(scn.Bytes())
		io.WriteString(out, "\n")
	}

	if err := scn.Err(); err != nil {
		return err
	}

	return out.Close()
}
