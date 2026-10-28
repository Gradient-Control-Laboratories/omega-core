package codegen_test

import (
	"strings"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/source"
)

// TestForwardReferenceIsAnError: a top-level variable used above the line
// that declares it is reported, naming the variable, instead of reaching the
// backend with no binding — which panicked here and in the fidget converter.
func TestForwardReferenceIsAnError(t *testing.T) {
	src := "a = max(b, p.x);\n" +
		"b = length(p) - 1.0;\n" +
		"return a;\n"
	res, err := assembler.Assemble([]string{src}, &source.MockManager{})
	if err == nil {
		_, err = renderHelper(src, res.Program)
	}
	if err == nil || !strings.Contains(err.Error(), "b is used before it is declared") {
		t.Fatalf("want a used-before-declared error naming b, got %v", err)
	}
}
