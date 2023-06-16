package gordian

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/TroutSoftware/x-tools/gordian/internal/txtar"
)

func TestRuns(t *testing.T) {
	runs, err := filepath.Glob("testdata/*.run")
	if err != nil {
		t.Fatal(err)
	}

	for _, run := range runs {
		t.Run(filepath.Base(run), func(t *testing.T) {
			x, err := txtar.ParseFile(run)
			if err != nil {
				t.Fatalf("error reading %s: %s", run, err)
			}

			pl, err := Compile(string(x.Get("pipeline")))
			if err != nil {
				t.Fatalf("invalid processing pipeline %s: %s", x.Get("pipeline"), err)
			}

			var in io.Reader = bytes.NewReader(x.Get("input"))
			for i := 0; i < len(pl); i++ {
				r, w := io.Pipe()
				go pl[i].Pipe(in, w)
				in = r
			}

			var buf bytes.Buffer
			io.Copy(&buf, in)

			if bytes.Equal(buf.Bytes(), x.Get("output")) {
				return
			}

			for i, f := range x.Files {
				if f.Name == "output" {
					x.Files[i].Data = buf.Bytes()
				}
			}

			res := run[:len(run)-4] + ".results"
			if err := os.WriteFile(res, txtar.Format(x), 0644); err != nil {
				t.Fatalf("writing results to %s: %s", res, err)
			}

			t.Errorf("invalid output. results in %s", res)
		})
	}
}
