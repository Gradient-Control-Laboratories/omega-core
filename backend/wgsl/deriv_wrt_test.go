package wgsl_test

import (
	"os"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/backend/wgsl"
)

// TestRender_DerivWrtVariable takes a program through the whole pipeline with
// a .d(var) on a scalar and on a shape: the autodiff pass must expand it, the
// variable's definition held independent and its dependents differentiated
// through, and the result must lower to WGSL.
func TestRender_DerivWrtVariable(t *testing.T) {
	src := `r = 3.0;
h = 2 * r;
can = max(length(p.xy) - r, abs(p.z) - h / 2);
speed = -can.d(r) / length(can.dp);
volume = pi * r * r * h;
dv = volume.d(r);
return can, speed, dv;
`
	res, err := assembler.Assemble([]string{src})
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	renderRes, err := wgsl.Render(src, res.Program)
	if err != nil {
		t.Fatalf("wgsl.Render failed: %v", err)
	}
	if renderRes == nil || renderRes.WGSL == "" {
		t.Fatal("expected WGSL output")
	}
}

// TestRender_DerivWrtVector takes a .d(c) on a literal vector variable through
// the pipeline: the vec3 of partials of a scalar field, and the Jacobian (a
// list of vec3 columns) of a vector field, both consumed downstream.
func TestRender_DerivWrtVector(t *testing.T) {
	src := `c = [1, 2, 3];
f = length(p - c);
g = f.d(c);
j = (p - c).d(c);
return f, length(g), length(j[0]) + length(j[2]);
`
	res, err := assembler.Assemble([]string{src})
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	renderRes, err := wgsl.Render(src, res.Program)
	if err != nil {
		t.Fatalf("wgsl.Render failed: %v", err)
	}
	if renderRes == nil || renderRes.WGSL == "" {
		t.Fatal("expected WGSL output")
	}
}

// TestRender_CanExample compiles the shipped can example, which is the
// showcase for .d(var): the volume's derivative through the budget-constrained
// height, and the can's own surface speed.
func TestRender_CanExample(t *testing.T) {
	data, err := os.ReadFile("../../examples/can.gsl")
	if err != nil {
		t.Fatalf("read example: %v", err)
	}
	src := string(data)
	res, err := assembler.Assemble([]string{src})
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	renderRes, err := wgsl.Render(src, res.Program)
	if err != nil {
		t.Fatalf("wgsl.Render failed: %v", err)
	}
	if renderRes == nil || renderRes.WGSL == "" {
		t.Fatal("expected WGSL output")
	}
}

// TestAssemble_DerivWrtVariableErrors: .d(var) is rejected inside a function
// body (a companion has no tangent for a top-level variable), on a list of
// vectors, and on a name that is not declared.
func TestAssemble_DerivWrtVariableErrors(t *testing.T) {
	cases := []struct {
		name, src, want string
	}{
		{"function_body", `width = 10;
grow(s) {
  return s.d(width);
}
f = grow(sphere(width));
return f;
`, "inside a function body"},
		{"list_of_vectors", `m = [[1, 0, 0], [0, 1, 0]];
f = length(p - m[0]);
g = f.d(m);
return f, g;
`, "list of vectors"},
		{"undeclared", `f = sphere(2);
g = f.d(undeclaredVariable);
return g;
`, "undeclaredVariable"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := assembler.Assemble([]string{c.src})
			if err == nil {
				t.Fatalf("expected an error containing %q", c.want)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error %q does not contain %q", err.Error(), c.want)
			}
		})
	}
}
