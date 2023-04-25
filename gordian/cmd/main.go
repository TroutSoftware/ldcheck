package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"runtime/pprof"

	"github.com/TroutSoftware/x-tools/gordian"
)

func main() {
	prg := flag.String("H", "", "gordian transforms to apply")
	cpuprofile := flag.String("prof", "", "write cpu profile to file")
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	pl, err := gordian.Compile(*prg)
	if err != nil {
		log.Fatal(err)
	}

	var in io.Reader = bufio.NewReader(os.Stdin)
	for i := 0; i < len(pl); i++ {
		r, w := io.Pipe()
		go pl[i].Pipe(in, w)
		in = r
	}

	io.Copy(os.Stdout, in)
}
