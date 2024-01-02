package main

import (
	"encoding/csv"
	"errors"
	"os"
	"regexp"
	"unicode/utf8"
)

var ErrUnterminatedString = errors.New("logfmt: unterminated string")

// take as input a “vm” to filter or output in table
// parsing error => synchronize at next key

// simple state machine to scan a full line
func lgfmtscanner(data []byte, x Exec) error {
	const (
		ST_INTERM = iota
		ST_KEY
		ST_EQUAL
		ST_QVAL
		ST_BQVAL
		ST_IVAL
	)

	clear(x.caps)

	st := ST_INTERM
	i, j := 0, 0
	var key, val string
	for j < len(data) {
		r, sz := utf8.DecodeRune(data[j:])
		// fmt.Printf("cap data %s, next rune %q (state %d)\n", data[i:j], r, st)
		switch st {
		case ST_INTERM:
			if r > ' ' && r != '"' && r != '=' {
				key, val = "", ""
				st = ST_KEY
				i = j
			}
		case ST_KEY:
			switch r {
			case '=':
				key = string(data[i:j])
				i = j + 1
				st = ST_EQUAL
			case ' ':
				// not a logfmt value, skip
				i = j + 1
				st = ST_INTERM
			}
		case ST_EQUAL:
			if r == '"' {
				i++
				st = ST_QVAL
			} else {
				st = ST_IVAL
			}
		case ST_QVAL:
			switch r {
			case '"':
				val = string(data[i:j])
				i = j + 1
				st = ST_INTERM
			case '\\':
				st = ST_BQVAL
			}
		case ST_BQVAL:
			st = ST_QVAL
		case ST_IVAL:
			if r == ' ' {
				val = string(data[i:j])
				i = j
				st = ST_INTERM
			}

		}
		if len(key) > 0 && len(val) > 0 {
			if re, in := x.fs[key]; in {
				if !re.MatchString(val) {
					return nil
				}
			}
			if i, c := x.ps[key]; c {
				x.caps[i] = val
			}
		}
		j += sz
	}

	return x.w.Write(x.caps)
}

type Exec struct {
	fs map[string]*regexp.Regexp
	ps map[string]int

	caps []string
	w    *csv.Writer
}

func NewExec(fs Filters, ps Projects) Exec {
	x := Exec{w: csv.NewWriter(os.Stdout), caps: make([]string, len(ps))}
	x.fs = make(map[string]*regexp.Regexp, len(fs))
	for _, f := range fs {
		x.fs[f.Key] = f.Match
	}
	x.ps = make(map[string]int, len(ps))
	for i, p := range ps {
		x.ps[p] = i
	}
	return x
}

func (x Exec) Flush() error {
	x.w.Flush()
	return x.w.Error()
}
