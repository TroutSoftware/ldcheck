package gordian

// tests for decompress.go
import (
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/TroutSoftware/x-tools/gordian/internal/txtar"
	"github.com/google/go-cmp/cmp"
)

func TestDecompressed(t *testing.T) {
	runs, err := filepath.Glob("testdata/compressed/*.run")
	if err != nil {
		t.Fatal(err)
	}

	for _, run := range runs {
		t.Run(filepath.Base(run), func(t *testing.T) {
			arch, err := txtar.ParseFile(run)
			if err != nil {
				t.Fatalf("error reading %s: %s", run, err)
			}

			var in io.Reader = strings.NewReader(arch.Get("input")[7:])
			r := base64.NewDecoder(base64.StdEncoding, in)
			f := "testdata/output.txt"

			out, err := os.Create(f)
			if err != nil {
				t.Fatalf("error creating file %s", err)
			}

			// write data to file
			_, err = io.Copy(out, r)
			if err != nil {
				t.Fatalf("error writing file %s", err)
			}

			var pl []Transform
			pl, err = Compile(arch.Get("pipeline"))
			if err != nil {
				t.Fatalf("invalid processing pipeline: %s", err)
			}

			for i := 0; i < len(pl); i++ {
				r, w := io.Pipe()
				regEx, err := regexp.Compile(f)
				if err != nil {
					t.Fatalf("error parsing file %s", err)
				}
				pl[i].Init(regEx)
				pl[i].Pipe(r, w)

				// validate output
				// read data from file
				var buf strings.Builder
				result, err := os.Open(f)
				if err != nil {
					t.Fatalf("error opening results %s", err)
				}

				_, err = io.Copy(&buf, result)
				if err != nil {
					t.Fatalf("error reading results %s", err)
				}

				out := arch.Get("output")[3:]
				if !cmp.Equal(buf.String(), out) {
					t.Errorf("invalid output. got: %s want: %s", out, buf.String())
				}

				// clear decompressed file
				os.Remove(f)
			}
		})
	}
}
