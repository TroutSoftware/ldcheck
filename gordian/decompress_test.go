package gordian

// tests for decompress.go
import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestDecompress(t *testing.T) {
	runs, err := filepath.Glob("testdata/compressed/*")
	if err != nil {
		t.Fatal(err)
	}

	for _, run := range runs {
		t.Run(filepath.Base(run), func(t *testing.T) {
			pl, err := Compile("decompress")
			if err != nil {
				t.Fatalf("invalid processing pipeline for file %s: %s", filepath.Base(run), err)
			}

			for i := 0; i < len(pl); i++ {
				r, w := io.Pipe()
				// seems wrong?
				regEx, err := regexp.Compile(run)
				if err != nil {
					t.Fatalf("error parsing file %s: %s", filepath.Base(run), err)
				}
				pl[i].Init(regEx)
				pl[i].Pipe(r, w)

				// clear decompressed file
				os.Remove(run)
			}
		})
	}
}
