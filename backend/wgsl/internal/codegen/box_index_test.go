package codegen_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/source"
)

// boxIndexOutline declares a profile's box from a reduce whose step reads its
// neighbour by index, points[mod(index + 1, count)], as an exact edge box does.
const boxIndexOutline = `
boxIndexOutlineXY(points) {
  count = reduce([0, 0], (sum, point) { return sum + 1; }, points).x;
  bounds = reduce([nothing, nothing, neg(nothing), neg(nothing)], (box, start, index) {
    end = points[mod(index + 1, count)];
    return [min(box.x, min(start.x, end.x)), min(box.y, min(start.y, end.y)),
            max(box.z, max(start.x, end.x)), max(box.w, max(start.y, end.y))];
  }, points);
  form = length(p.xy) - 1;
  form.box = [[bounds.x, bounds.y, neg(nothing)], [bounds.z, bounds.w, nothing]];
  return form;
}
`

// The box domain left such a reduce's index a name, which only the reduce's
// unrolled step binds. In builtin, the part the declared box bounds lost it
// without a word; read from any library, the box failed to link ("undefined
// identifier: index"). The index now folds to the constant it is, as it does
// for a gradient.

// TestDeclaredBoxReadsReduceIndex_PartBox: a builtin function declaring such a
// box keeps it as its part's box, the points' box padded by boxPad.
func TestDeclaredBoxReadsReduceIndex_PartBox(t *testing.T) {
	files, err := filepath.Glob("../../../../stdlib/builtin/*.gsl")
	if err != nil || len(files) == 0 {
		t.Fatalf("builtin sources: %v", err)
	}
	var builtin []string
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		builtin = append(builtin, string(b))
	}
	mgr := &source.MockManager{PkgSources: map[string][]string{"builtin": append(builtin, boxIndexOutline)}}
	src := "solid = extrude(boxIndexOutlineXY([[0, 0], [10, 0], [10, 10]]), 0, 5);\nreturn solid;\n"
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mgr))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	out, err := renderHelper(src, res.Program)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	for _, want := range []string{"gsl_partLo = vec3<f32>(-0.01, -0.01, 0.0)", "gsl_partHi = vec3<f32>(10.01, 10.01, 5.0)"} {
		if !strings.Contains(out.WGSL, want) {
			t.Errorf("the part's box is not the declared one: no %q in the WGSL", want)
		}
	}
}

// TestDeclaredBoxReadsReduceIndex_Read: a document reads such a box from a
// library function.
func TestDeclaredBoxReadsReduceIndex_Read(t *testing.T) {
	mgr := &source.MockManager{PkgSources: map[string][]string{"shapes/outline": {boxIndexOutline}}}
	src := "import \"shapes/outline\";\nsolid = outline.boxIndexOutlineXY([[0, 0], [10, 0], [10, 10]]);\n" +
		"b = solid.box;\nreturn solid;\n"
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mgr))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	if _, err := renderHelper(src, res.Program); err != nil {
		t.Fatalf("render: %v", err)
	}
}
