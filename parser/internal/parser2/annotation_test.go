package parser2_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/parser/internal/parser2"
	"gradientcontrol.com/omega/core/program"
)

// parseFile parses one source and returns its single package file, failing the
// test on any parse error.
func parseFile(t *testing.T, src string) *program.PackageFile {
	t.Helper()
	res := parser2.Parse([]string{src})
	if len(res.Errors) > 0 {
		t.Fatalf("unexpected parse errors: %v", res.Errors)
	}
	if res.Package == nil || len(res.Package.Files) != 1 {
		t.Fatalf("expected one package file, got %+v", res.Package)
	}
	return res.Package.Files[0]
}

// parseErr parses one source expecting it to be rejected, and returns the
// message so the caller can check which rule rejected it.
func parseErr(t *testing.T, src string) string {
	t.Helper()
	res := parser2.Parse([]string{src})
	if res == nil || len(res.Errors) == 0 {
		t.Fatalf("expected a parse error, got none")
	}
	msgs := make([]string, len(res.Errors))
	for i, e := range res.Errors {
		msgs[i] = e.Error()
	}
	return strings.Join(msgs, "; ")
}

func TestMaterialAnnotationOnConstant(t *testing.T) {
	file := parseFile(t, "@material brass = 0.3;\nbrassColor = [1, 1, 0];\nreturn sphere(1);")

	if len(file.Annotations) != 1 {
		t.Fatalf("expected one annotation, got %d", len(file.Annotations))
	}
	ann := file.Annotations[0]
	if ann.Kind != "material" || ann.Symbol != "brass" {
		t.Errorf("got kind=%q symbol=%q, want material and brass", ann.Kind, ann.Symbol)
	}
	// The span runs from the '@' to one past the end of the annotated name, so a
	// browser can jump from a swatch to the declaration behind it.
	if ann.Start == nil || ann.Start.Line != 1 || ann.Start.Column != 1 {
		t.Errorf("got start=%+v, want line 1 column 1", ann.Start)
	}
	if ann.End == nil || ann.End.Column != int32(len("@material brass")+1) {
		t.Errorf("got end=%+v, want column %d", ann.End, len("@material brass")+1)
	}

	// Recorded, not lowered: the assignment is built exactly as it would be
	// without the annotation, so nothing downstream sees a wrapper call.
	if len(file.Assignments) != 2 {
		t.Fatalf("expected both assignments, got %d", len(file.Assignments))
	}
	brass := file.Assignments[0]
	if len(brass.Variables) != 1 || brass.Variables[0].Name != "brass" {
		t.Fatalf("expected the brass assignment first, got %+v", brass.Variables)
	}
	if call, wrapped := brass.Expression.T.(*program.Operand_FnCall); wrapped {
		t.Errorf("annotation wrapped the expression in a call to %q", call.FnCall.Reference.Name)
	}
}

func TestMaterialAnnotationOnFunction(t *testing.T) {
	// The grammar already admits an annotated function declaration — @point and
	// @transform simply refuse one — so a material that computes itself from the
	// point carries the annotation the same way a constant does.
	file := parseFile(t, `
@material walnut(axisOrigin, axisDir) {
  return axisOrigin + axisDir, [0.4, 0.26, 0.13];
}
return sphere(1);
`)

	if len(file.Annotations) != 1 {
		t.Fatalf("expected one annotation, got %d", len(file.Annotations))
	}
	if ann := file.Annotations[0]; ann.Kind != "material" || ann.Symbol != "walnut" {
		t.Errorf("got kind=%q symbol=%q, want material and walnut", ann.Kind, ann.Symbol)
	}
	if len(file.FnTypes) != 1 || file.FnTypes[0].Name != "walnut" {
		t.Fatalf("expected the walnut function to be declared, got %+v", file.FnTypes)
	}
}

func TestMaterialAnnotationOnOverloadsSharesOneEntry(t *testing.T) {
	// An annotation marks a name, not an arity, so only one overload carries it
	// and every arity of walnut is covered.
	file := parseFile(t, `
@material walnut(axisOrigin, axisDir) {
  return walnut(axisOrigin, axisDir, 0.3);
}
walnut(axisOrigin, axisDir, oil) {
  return axisOrigin + axisDir * oil, [0.4, 0.26, 0.13];
}
return sphere(1);
`)

	if len(file.Annotations) != 1 {
		t.Fatalf("expected one annotation, got %d", len(file.Annotations))
	}
	if len(file.FnTypes) != 1 || len(file.FnTypes[0].Groups) != 2 {
		t.Fatalf("expected one function name with two arities, got %+v", file.FnTypes)
	}
}

func TestFinishAnnotationIsRecordedWithItsOwnKind(t *testing.T) {
	// @finish is recorded the same way @material is; only the kind differs, and
	// the kind is what the validator and the browser switch on.
	file := parseFile(t, "@finish matte = 1.0;\nreturn sphere(1);")

	if len(file.Annotations) != 1 {
		t.Fatalf("expected one annotation, got %d", len(file.Annotations))
	}
	if ann := file.Annotations[0]; ann.Kind != "finish" || ann.Symbol != "matte" {
		t.Errorf("got kind=%q symbol=%q, want finish and matte", ann.Kind, ann.Symbol)
	}
}

func TestMaterialAnnotationRejections(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{{
		name: "multiple names",
		src: "walnut(axisOrigin, axisDir) {\n  return axisOrigin, axisDir;\n}\n" +
			"@material m, c = walnut([0, 0, 0], [0, 0, 1]);\nreturn sphere(1);",
		want: "annotates a single name",
	}, {
		name: "property assignment",
		src:  "s = sphere(1);\n@material s.color = [1, 0, 0];\nreturn s;",
		want: "can only annotate a variable assignment or a function declaration",
	}, {
		name: "inside a function body",
		src:  "f(x) {\n  @material m = x;\n  return m;\n}\nreturn sphere(f(1));",
		want: "can only annotate a top-level declaration",
	}, {
		name: "annotated twice",
		src:  "@material brass = 0.3;\n@material brass = 0.4;\nreturn sphere(1);",
		want: "already annotated @material",
	}, {
		name: "unknown annotation",
		src:  "@texture brass = 0.3;\nreturn sphere(1);",
		// The grammar admits only the annotations it knows, so an unknown one
		// is refused by the parser rather than by buildAnnotatedDeclaration.
		want: `unexpected token "texture"`,
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseErr(t, tc.src)
			if !strings.Contains(got, tc.want) {
				t.Errorf("got %q, want it to contain %q", got, tc.want)
			}
		})
	}
}

func TestSurfaceAnnotationMarksTheOperatorsName(t *testing.T) {
	// A @surface preset is an overload of the operator it bakes defaults for,
	// taking only the slots a document supplies: the substrate first, its
	// colour, the body's frame. The annotation marks the name, so one entry
	// covers the preset and the operator, and @frame is recorded on the
	// parameter as @color is.
	file := parseFile(t, `
@surface turned(baseMaterial, @color baseColor, @frame frame) {
  cut, tone = turned(baseMaterial, baseColor, frame, 0.15);
  return cut, tone;
}
turned(baseMaterial, baseColor, frame, feed) {
  return baseMaterial + feed, baseColor[0] + frame.x;
}
return sphere(1);
`)

	if len(file.Annotations) != 1 {
		t.Fatalf("expected one annotation, got %d", len(file.Annotations))
	}
	if ann := file.Annotations[0]; ann.Kind != "surface" || ann.Symbol != "turned" {
		t.Errorf("got kind=%q symbol=%q, want surface and turned", ann.Kind, ann.Symbol)
	}
	if len(file.FnTypes) != 1 || len(file.FnTypes[0].Groups) != 2 {
		t.Fatalf("expected one function name with two arities, got %+v", file.FnTypes)
	}
	var preset *program.Function
	for _, g := range file.FnTypes[0].Groups {
		if g.ParamCount == 3 {
			preset = g.Functions[0]
		}
	}
	if preset == nil {
		t.Fatalf("expected the three-parameter preset, got %+v", file.FnTypes[0].Groups)
	}
	if got := preset.Parameters[1].Kind; got != "color" {
		t.Errorf("baseColor kind = %q, want color", got)
	}
	if got := preset.Parameters[2].Kind; got != "frame" {
		t.Errorf("frame kind = %q, want frame", got)
	}
	if got := preset.Parameters[0].Kind; got != "" {
		t.Errorf("baseMaterial kind = %q, want none", got)
	}
}

func TestSurfaceAndFrameRejections(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{{
		name: "surface on a constant",
		src:  "@surface matte = 1.0;\nreturn sphere(1);",
		want: "@surface annotation can only be applied to function declarations",
	}, {
		name: "surface with an argument",
		src:  "@surface(4) coated(m) {\n  return m;\n}\nreturn sphere(1);",
		want: "@surface annotation does not take arguments",
	}, {
		name: "surface on a parameter",
		src:  "coated(@surface m) {\n  return m;\n}\nreturn sphere(1);",
		want: "unsupported parameter annotation @surface",
	}, {
		name: "frame on a variable",
		src:  "@frame f = [[1, 0, 0], [0, 1, 0]];\nreturn sphere(1);",
		want: "@frame annotation can only be applied to function parameters",
	}, {
		name: "frame with an argument",
		src:  "g(@frame(x) f) {\n  return f;\n}\nreturn sphere(1);",
		want: "@frame annotation does not take arguments",
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseErr(t, tc.src)
			if !strings.Contains(got, tc.want) {
				t.Errorf("got %q, want it to contain %q", got, tc.want)
			}
		})
	}
}

func TestPointAnnotationStillLowers(t *testing.T) {
	// @point and @transform keep their lowering: the annotated expression is
	// wrapped in a marker call that common/build strips before codegen.
	file := parseFile(t, "@point center = [1, 2, 3];\nreturn sphere(1);")

	if len(file.Annotations) != 0 {
		t.Errorf("a lowered annotation should record nothing, got %+v", file.Annotations)
	}
	if len(file.Assignments) != 1 {
		t.Fatalf("expected one assignment, got %d", len(file.Assignments))
	}
	call, wrapped := file.Assignments[0].Expression.T.(*program.Operand_FnCall)
	if !wrapped {
		t.Fatalf("expected the expression to be wrapped in a call, got %T",
			file.Assignments[0].Expression.T)
	}
	if call.FnCall.Reference.Name != "uiPoint" {
		t.Errorf("got a call to %q, want uiPoint", call.FnCall.Reference.Name)
	}
}
