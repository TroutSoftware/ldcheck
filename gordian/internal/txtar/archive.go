// Evolution of golang.org/x/tools/txtar, handling encoding for binary data
package txtar

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

// An Archive is a collection of files.
type Archive struct {
	Comment []byte
	Files   []File
}

// A File is a single file in an archive.
type File struct {
	Name string // name of file ("foo/bar.txt")
	Data []byte // text content of file
}

// Format returns the serialized form of an Archive.
// It is assumed that the Archive data structure is well-formed:
// a.Comment and all a.File[i].Data contain no file marker lines,
// and all a.File[i].Name is non-empty.
func Format(a *Archive) []byte {
	var buf bytes.Buffer
	buf.Write(fixNL(a.Comment))
	for _, f := range a.Files {
		fmt.Fprintf(&buf, "-- %s --\n", f.Name)
		buf.Write(fixNL(f.Data))
	}
	return buf.Bytes()
}

// ParseFile parses the named file as an archive.
func ParseFile(file string) (*Archive, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return Parse(data), nil
}

// Parse parses the serialized form of an Archive.
// The returned Archive holds slices of data.
func Parse(data []byte) *Archive {
	a := new(Archive)
	var name, enc string
	a.Comment, name, data = findFileMarker(data)
	for name != "" {
		name, enc = findEncodingMarker(name)
		f := File{name, nil}
		f.Data, name, data = findFileMarker(data)
		switch enc {
		case "base64":
			buf := make([]byte, base64.StdEncoding.DecodedLen(len(f.Data)))
			sz, err := base64.StdEncoding.Decode(buf, f.Data)
			if err != nil {
				panic(err)
			}
			f.Data = buf[:sz]
		}
		a.Files = append(a.Files, f)
	}
	return a
}

// Get return file at name
// This panics if the file does not exist, and should only be used in tests
func (a *Archive) Get(name string) []byte {
	for _, f := range a.Files {
		if f.Name == name {
			return f.Data
		}
	}
	panic("file " + name + " not in archive")
}

var (
	newlineMarker = []byte("\n-- ")
	marker        = []byte("-- ")
	markerEnd     = []byte(" --")
)

// findFileMarker finds the next file marker in data,
// extracts the file name, and returns the data before the marker,
// the file name, and the data after the marker.
// If there is no next marker, findFileMarker returns before = fixNL(data), name = "", after = nil.
func findFileMarker(data []byte) (before []byte, name string, after []byte) {
	var i int
	for {
		if name, after = isMarker(data[i:]); name != "" {
			return data[:i], name, after
		}
		j := bytes.Index(data[i:], newlineMarker)
		if j < 0 {
			return fixNL(data), "", nil
		}
		i += j + 1 // positioned at start of new possible marker
	}
}

func findEncodingMarker(name string) (string, string) {
	const encodingMarker = "; "
	j := strings.Index(name, encodingMarker)
	if j < 0 {
		return name, ""
	}

	return name[:j], name[j+len(encodingMarker):]
}

// isMarker checks whether data begins with a file marker line.
// If so, it returns the name from the line and the data after the line.
// Otherwise it returns name == "" with an unspecified after.
func isMarker(data []byte) (name string, after []byte) {
	if !bytes.HasPrefix(data, marker) {
		return "", nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		data, after = data[:i], data[i+1:]
	}
	if !(bytes.HasSuffix(data, markerEnd) && len(data) >= len(marker)+len(markerEnd)) {
		return "", nil
	}
	return strings.TrimSpace(string(data[len(marker) : len(data)-len(markerEnd)])), after
}

// If data is empty or ends in \n, fixNL returns data.
// Otherwise fixNL returns a new slice consisting of data with a final \n added.
func fixNL(data []byte) []byte {
	if len(data) == 0 || data[len(data)-1] == '\n' {
		return data
	}
	d := make([]byte, len(data)+1)
	copy(d, data)
	d[len(data)] = '\n'
	return d
}
