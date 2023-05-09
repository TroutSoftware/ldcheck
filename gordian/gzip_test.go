package gordian

// tests for decompress.go
import (
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestGzip(t *testing.T) {
	file := "testdata/test.txt.gz"
	t.Run(filepath.Base(file), func(t *testing.T) {
		createAndCompressGzip(t, file)
		var pl []Transform
		pl, err := Compile("gunzip")
		if err != nil {
			t.Fatalf("invalid processing pipeline for file %s: %s", filepath.Base(file), err)
		}

		for i := 0; i < len(pl); i++ {
			r, w := io.Pipe()
			// seems wrong?
			regEx, err := regexp.Compile(file)
			if err != nil {
				t.Fatalf("error parsing file %s: %s", filepath.Base(file), err)
			}
			pl[i].Init(regEx)
			pl[i].Pipe(r, w)

			// clear decompressed file
			os.Remove(file)
		}
	})
}

func createAndCompressGzip(t *testing.T, filename string) {
	lines := []string{
		"10.0.0.1 - - [08/May/2023:12:34:56 -0500] \"GET /index.html HTTP/1.1\" 200 1234 \"-\" \"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36 Edge/16.16299\"\n",
		"10.0.0.2 - - [08/May/2023:12:35:02 -0500] \"GET /about.html HTTP/1.1\" 200 987 \"-\" \"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36 Edge/16.16299\"\n",
		"10.0.0.3 - - [08/May/2023:12:36:12 -0500] \"POST /submit.php HTTP/1.1\" 302 0 \"/form.html\" \"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36 Edge/16.16299\"\n",
		"10.0.0.4 - - [08/May/2023:12:36:23 -0500] \"GET /images/logo.png HTTP/1.1\" 304 0 \"http://example.com/index.html\" \"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36 Edge/16.16299\"\n",
		"10.0.0.5 - - [08/May/2023:12:37:01 -0500] \"GET /blog.html HTTP/1.1\" 404 0 \"-\" \"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36 Edge/16.16299\"\n",
	}

	// Create the file
	file, err := os.Create(filename + "_temp")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	// Write the Apache access log data to the file
	for _, line := range lines {
		_, err = file.WriteString(line)
		if err != nil {
			t.Fatal(err)
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	writer := gzip.NewWriter(f)
	defer writer.Close()

	_, err = file.Seek(0, 0) // Reset the file offset to the beginning
	if err != nil {
		log.Fatal(err)
	}

	_, err = io.Copy(writer, file) // Compress the file
	if err != nil {
		log.Fatal(err)
	}

	err = f.Sync()
	if err != nil {
		log.Fatal(err)
	}

	// delete uncompressed file
	os.Remove(filename + "_temp")
}
