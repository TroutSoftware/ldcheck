package gordian

import (
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
	"unicode/utf8"
)

type Transform interface {
	Init(*regexp.Regexp)
	Pipe(io.Reader, io.WriteCloser) error
}

type Pipeline = []Transform

// language syntax:
//
//	expr = trans { "|" trans }.
//	trans = name { arg } .
//	arg = regexp | string .
//	regexp = "/" validregexp "/" .
//	string = "\"" validgostring "\"" .
func Compile(prg string) (Pipeline, error) {
	lx := &lexer{src: prg}

	var tr Transform
	var pipeline Pipeline

	for lx.next() {
		switch lx.lex {
		case lTransform:
			t := lx.tk()
			typ, ok := transforms[t]
			if !ok {
				return nil, fmt.Errorf("unknown transform %s", lx.tk())
			}

			tr = reflect.New(typ).Interface().(Transform)
			pipeline = append(pipeline, tr)
		case lString:
			re, err := regexp.Compile(regexp.QuoteMeta(lx.tk()))
			if err != nil {
				return nil, fmt.Errorf("at %d: invalid string %s: %w", lx.off, lx.tk(), err)
			}
			tr.Init(re)
		case lRegexp:
			re, err := regexp.Compile(lx.tk())
			if err != nil {
				return nil, fmt.Errorf("at %d: invalid regexp %s: %w", lx.off, lx.tk(), err)
			}
			tr.Init(re)
		case lPipe:
			tr = nil
		case lUnknown:
			return nil, fmt.Errorf("unknown token at %d", lx.off)
		}
	}

	return pipeline, nil
}

var transforms = map[string]reflect.Type{
	"bunzip2":  reflect.TypeOf(Bzip2{}),
	"bunzip2?": reflect.TypeOf(MaybeBzip2{}),
	"dragend":  reflect.TypeOf(DragEnd{}),
	"groupml":  reflect.TypeOf(GroupML{}),
	"gunzip":   reflect.TypeOf(Gzip{}),
	"gunzip?":  reflect.TypeOf(MaybeGzip{}),
	"ignore":   reflect.TypeOf(Ignore{}),
	"noempty":  reflect.TypeOf(NoEmpty{}),
	"only":     reflect.TypeOf(Only{}),
	"unixyear": reflect.TypeOf(UnixYear{}),
}

const (
	lUnknown = iota
	lTransform
	lString
	lRegexp
	lPipe
)

type lexer struct {
	src string

	off, len int
	skip     int
	lex      int
}

func (l *lexer) tk() string { return l.src[l.off : l.off+l.len] }

func (l *lexer) next() bool {
	l.off += l.len + l.skip

	l.lex = lUnknown
	l.len = 0
	l.skip = 0

	for l.off < len(l.src) {
		switch l.src[l.off] {
		case ' ', '\n', '\r', '\t':
			l.off++
		case '"':
			l.off++
			l.lex = lString
			return l.until("\"")
		case '/':
			l.off++
			l.lex = lRegexp
			return l.until("/")
		case '|':
			l.lex = lPipe
			l.len = 1
			return true
		case '#':
			l.until("\n")
			l.off += l.len + l.skip
		default:
			l.lex = lTransform
			if !l.until(" \n\r\t") {
				l.len = len(l.src) - l.off // accept all remaining
			}

			return true // next step will check validity
		}
	}
	return false
}

func (l *lexer) until(chars string) bool {
	i := strings.IndexAny(l.src[l.off:], chars)
	if i == -1 {
		return false
	}

	_, l.skip = utf8.DecodeRuneInString(l.src[l.off+i:])
	l.len = i
	return true
}
