package parser

import (
	"path/filepath"
	"testing"
)

// TestExamplesParse will parse every program in testdata/examples and compare
// the AST dump to testdata/golden. Until M1 it only checks the fixtures exist.
func TestExamplesParse(t *testing.T) {
	files, err := filepath.Glob("../../testdata/examples/*.loom")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no example programs found in testdata/examples")
	}
	t.Skip("M1: parser not implemented yet")
}

// FuzzParse must never panic or hang on any input.
func FuzzParse(f *testing.F) {
	f.Skip("M1: parser not implemented yet")
}
