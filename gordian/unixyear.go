package gordian

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"time"
)

// One of the most specialized operator: if the date field look like a month, fill in the year.
// Year is current one, except if log month is Dec and current month is january, in case this is past year.
// This heuristic matches broadly what is expected looking at logs.
type UnixYear struct{}

var unixmonth = regexp.MustCompile("(Jan)|(Feb)|(Mar)|(Apr)|(May)|(Jun)|(Jul)|(Aug)|(Sep)|(Oct)|(Nov)|(Dec)")
var itisjan = time.Now().Month() == time.January
var thisyear = time.Now().Year()

func (*UnixYear) Init(_ *regexp.Regexp) {}
func (e *UnixYear) Pipe(in io.Reader, out io.WriteCloser) error {
	scn := bufio.NewScanner(in)
	for scn.Scan() {
		m := unixmonth.FindIndex(scn.Bytes())
		if m == nil {
			out.Write(scn.Bytes())
		} else {
			out.Write(scn.Bytes()[:m[0]])
			if bytes.Equal(scn.Bytes()[m[0]:m[1]], []byte("Dec")) && itisjan {
				fmt.Fprint(out, thisyear-1, " ")
			} else {
				fmt.Fprint(out, thisyear, " ")
			}

			out.Write(scn.Bytes()[m[0]:])
		}
		io.WriteString(out, "\n")

	}

	if err := scn.Err(); err != nil {
		return err
	}

	return out.Close()
}
