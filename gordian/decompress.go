package gordian

import (
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

type Decompress struct {
	filePath string
}

func (g *Decompress) Init(r *regexp.Regexp) {
	g.filePath = r.String()

}
func (g *Decompress) Pipe(in io.Reader, out io.WriteCloser) error {
	fileName := filepath.Base(g.filePath)
	fileExt := filepath.Ext(fileName)
	filePathWithoutExt := g.filePath[:len(g.filePath)-len(fileExt)]

	file, err := os.Open(g.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var reader io.Reader
	switch fileExt {
	case ".gz":
		reader, err = gzip.NewReader(file)
		if err != nil {
			return err
		}
	case ".bz2":
		reader = bzip2.NewReader(file)
	default:
		return fmt.Errorf("unknown file extension: %s", fileExt)
	}

	// Create an output file
	outFile, err := os.Create(filePathWithoutExt)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, reader)
	if err != nil {
		return err
	}

	return nil
}
