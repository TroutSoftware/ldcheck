package gordian

import (
	"compress/bzip2"
	"io"
	"os"
	"regexp"
)

type Bzip2 struct {
	filePath string
}

func (g *Bzip2) Init(r *regexp.Regexp) {
	g.filePath = r.String()

}
func (g *Bzip2) Pipe(in io.Reader, out io.WriteCloser) error {
	file, err := os.Open(g.filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := bzip2.NewReader(file)
	if err != nil {
		return err
	}

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
