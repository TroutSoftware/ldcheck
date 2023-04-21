package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
)

func main() {
	prg := flag.String("H", "", "gordian transforms to apply")
	flag.Parse()

	pl, err := Compile(*prg)
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
