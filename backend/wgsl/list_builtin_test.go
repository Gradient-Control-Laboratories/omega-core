package wgsl_test

import (
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
	"github.com/Gradient-Control-Laboratories/omega-core/backend/wgsl"
)

// A list of shapes built from list(n, fn) and a constant-length helper reach
// the shader.
func TestListBuiltinBuildsShapes(t *testing.T) {
	src := `row(n, spacing) { return list(n, (i) { return sphere([i * spacing, 0, 0], 1); }); }
return unionSharp(row(5, 3)), unionRound(row(3, 2), 0.5);`
	res, err := assembler.Assemble([]string{src})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wgsl.Render(src, res.Program); err != nil {
		t.Fatal(err)
	}
}
