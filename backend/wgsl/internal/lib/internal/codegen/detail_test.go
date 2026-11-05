package codegen

import (
	"os"
	"regexp"
	"strconv"
	"testing"
)

// TestDetailPixel holds the share of a feature's size the detail guard
// switches at to the one builtin's detailValue documents.
func TestDetailPixel(t *testing.T) {
	src, err := os.ReadFile("../../../../../../stdlib/builtin/base.gsl")
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`draws fine where\n// a pixel is under ([0-9.]+) of the size`).FindSubmatch(src)
	if m == nil {
		t.Fatal("builtin/base.gsl does not give detailValue's pixel as the test expects")
	}
	want, err := strconv.ParseFloat(string(m[1]), 64)
	if err != nil {
		t.Fatal(err)
	}
	if detailPixelShare != want {
		t.Errorf("detailPixelShare is %v, want detailValue's %v", detailPixelShare, want)
	}
}
