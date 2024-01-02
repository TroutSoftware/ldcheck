package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Filters []Filter

func (f *Filters) String() string {
	var buf strings.Builder
	for i, f := range *f {
		pfix := ", "
		if i == 0 {
			pfix = ""
		}
		fmt.Fprintf(&buf, pfix+"%s:%s", f.Key, f.Match)
	}
	return buf.String()
}

func (f *Filters) Set(value string) error {
	k, mr, ok := strings.Cut(value, "=")
	if !ok {
		return fmt.Errorf("invalid filter %s: want <key>=<value>", value)
	}
	m, err := regexp.Compile(mr)
	if err != nil {
		return fmt.Errorf("invalid regexp %s: %w", mr, err)
	}

	if *f == nil {
		*f = Filters{{k, m}}
	} else {
		*f = append(*f, Filter{k, m})
	}
	return nil
}

type Filter struct {
	Key   string
	Match *regexp.Regexp
}

type Projects []string

func (p *Projects) String() string { return strings.Join([]string(*p), ",") }
func (p *Projects) Set(value string) error {
	if *p == nil {
		*p = Projects{value}
	} else {
		*p = append(*p, value)
	}
	return nil
}

func main() {
	var filters Filters
	var projects Projects

	flag.Var(&filters, "f", `Filter lines according to key=value.
	key is an exact string match, and value a regular expression.
	Pass this option multiple times to filter more (logical and).`)
	flag.Var(&projects, "e", `Extract a specific value at key. 
	Pass this option multiple times to extract multiple fields.
	The order of the fields return matches the order of the arguments.`)
	flag.Parse()

	in := os.Stdin
	// TODO(rdo) scan multiple files
	if flag.NArg() == 1 {
		fh, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot open %s: %s\n", flag.Arg(1), err)
			os.Exit(1)
		}
		in = fh
		defer fh.Close()
	}

	sc := bufio.NewScanner(in)
	x := NewExec(filters, projects)
	for sc.Scan() {
		lgfmtscanner(sc.Bytes(), x)
	}
	if sc.Err() != nil {
		fmt.Fprintf(os.Stderr, "error reading data: %s\n", sc.Err())
		os.Exit(1)
	}
	if err := x.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "error writing csv: %s\n", err)
	}
}
