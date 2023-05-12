package gordian

import (
	"compress/gzip"
	"io"
	"os"
	"regexp"
)

type Gzip struct {
	filePath string
}

func (g *Gzip) Init(r *regexp.Regexp) {
	g.filePath = r.String()

}
func (g *Gzip) Pipe(in io.Reader, out io.WriteCloser) error {
	file, err := os.Open(g.filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Create an output file
	outFile, err := os.Create(g.filePath + "_temp")
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, reader)
	if err != nil {
		return err
	}

	// Rename the temporary file to the original file
	err = os.Rename(g.filePath+"_temp", g.filePath)
	if err != nil {
		return err
	}

	return nil
}
