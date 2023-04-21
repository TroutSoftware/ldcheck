package main

import (
	"fmt"
	"io"
	"reflect"
	"unicode/utf8"
)

type Transform interface {
	Init(int, string) error
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
			typ, ok := transforms[lx.tk()]
			if !ok {
				return nil, fmt.Errorf("unknown transform %s", lx.tk())
			}

			tr = reflect.New(typ).Interface().(Transform)
			pipeline = append(pipeline, tr)
		case lString, lRegexp:
			tr.Init(lx.lex, lx.tk())
		case lPipe:
			tr = nil
		case lUnknown:
			return nil, fmt.Errorf("unknown token at %d", lx.off)
		}
	}

	return pipeline, nil
}

var transforms = map[string]reflect.Type{
	"groupml": reflect.TypeOf(GroupML{}),
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

func (l *lexer) until(r rune) bool {
	for i, s := range l.src[l.off:] {
		if s == r {
			l.skip = utf8.RuneLen(r)
			l.len = i
			return true
		}
	}
	return false
}

func (l *lexer) next() bool {
	l.off += l.len + l.skip
	if l.off == len(l.src) {
		return false
	}

	l.lex = lUnknown
	l.len = 0
	l.skip = 0

	for {
		switch l.src[l.off] {
		case ' ', '\n', '\r', '\t':
			l.off++

		case '"':
			l.off++
			l.lex = lString
			return l.until('"')
		case '/':
			l.off++
			l.lex = lRegexp
			return l.until('/')
		case '|':
			l.lex = lPipe
			l.len = 1
			return true
		default:
			l.lex = lTransform
			return l.until(' ')
		}
	}
}
