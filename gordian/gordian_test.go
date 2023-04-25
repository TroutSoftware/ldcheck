package gordian

import (
	"io"
	"os"
	"path/filepath"
	"strings"
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
			arch, err := txtar.ParseFile(run)
			if err != nil {
				t.Fatalf("error reading %s: %s", run, err)
			}

			pl, err := Compile(arch.Get("pipeline"))
			if err != nil {
				t.Fatalf("invalid processing pipeline %s: %s", arch.Get("pipeline"), err)
			}

			var in io.Reader = strings.NewReader(arch.Get("input"))
			for i := 0; i < len(pl); i++ {
				r, w := io.Pipe()
				go pl[i].Pipe(in, w)
				in = r
			}

			var buf strings.Builder
			io.Copy(&buf, in)

			if buf.String() == arch.Get("output") {
				return
			}

			for i, f := range arch.Files {
				if f.Name == "output" {
					arch.Files[i].Data = []byte(buf.String())
				}
			}

			res := run[:len(run)-4] + ".results"
			if err := os.WriteFile(res, txtar.Format(arch), 0644); err != nil {
				t.Fatalf("writing results to %s: %s", res, err)
			}

			t.Errorf("invalid output. results in %s", res)
		})
	}
}
