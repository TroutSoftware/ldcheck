package main

import (
	"bufio"
	"io"
	"regexp"
)

// GroupML re-assemble records spread over multiple lines.
// Given a specific pattern (regular expression), lines are grouped according to heuristic:
//
//  1. if the current line matches the pattern, *as long as*
//  2. the next line does not the pattern, *and*
//  3. the next line is not an empty line
//
// then they all belong to the same record.
type GroupML struct{ re *regexp.Regexp }

func (grp *GroupML) Init(re *regexp.Regexp) { grp.re = re }

func (grp *GroupML) Pipe(in io.Reader, out io.WriteCloser) error {
	ml := false

	scn := bufio.NewScanner(in)
	for scn.Scan() {
		switch {
		case grp.re.Match(scn.Bytes()):
			if ml {
				// terminate previous match
				io.WriteString(out, "\n")
			}
			ml = true
			out.Write(tright(scn.Bytes()))

		case ml:
			// normalize spaces to single
			line := tleft(scn.Bytes())
			if len(line) == 0 {
				// rule #3: end capture
				out.Write(scn.Bytes())
				io.WriteString(out, "\n")
				ml = false
				continue
			}
			io.WriteString(out, " ")
			out.Write(line)

		default:
			out.Write(scn.Bytes())
			// add missing cr
			io.WriteString(out, "\n")
		}
	}

	if err := scn.Err(); err != nil {
		return err
	}

	return out.Close()
}

func tleft(b []byte) []byte {
	sp := 0
	for ; sp < len(b); sp++ {
		if asciiSpace[b[sp]] == 0 {
			break
		}
	}
	return b[sp:]
}

func tright(b []byte) []byte {
	sp := len(b)
	for ; sp > 0; sp-- {
		if asciiSpace[b[sp-1]] == 0 {
			break
		}
	}
	return b[:sp]
}

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}
