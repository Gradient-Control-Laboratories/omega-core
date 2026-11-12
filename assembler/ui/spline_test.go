package ui_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/assembler/ui"
	"gradientcontrol.com/omega/core/internal/source"
)

func TestExtractLocalSpline(t *testing.T) {
	src := `m = [[1,0,0,0],[0,1,0,0],[0,0,1,0],[9,8,7,1]];
branch() {
  m = [[1,0,0,0],[0,1,0,0],[0,0,1,0],[3,4,5,1]];
  start = uiTransform(m);
  bend = uiTransform([[1,0,0,0],[0,1,0,0],[0,0,1,0],[6,7,8,1]]);
  end = uiTransform([[1,0,0,0],[0,1,0,0],[0,0,1,0],[9,10,11,1]]);
  firstHandle = start;
  path = uiSpline([firstHandle, bend, end]);
  first = path[0];
  last = path[2];
  a = first[3];
  b = last[3];
  return length(p - a.xyz) + length(p - b.xyz);
}
return branch();`
	controls := extract(t, src)
	var spline *ui.Control
	transforms := map[string]ui.Control{}
	for i, c := range controls {
		if c.Kind == ui.KindSpline {
			if spline != nil {
				t.Fatal("duplicate spline")
			}
			spline = &controls[i]
		} else if c.Kind == ui.KindTransform {
			transforms[c.Span.VarName()] = c
		}
	}
	if spline == nil || len(spline.Controls) != 3 || len(spline.Value) != 48 {
		t.Fatalf("missing ordered spline: %+v", controls)
	}
	if got := textAt(src, spline.Span); got != "[firstHandle, bend, end]" {
		t.Fatalf("spline span: %s", got)
	}
	for i, child := range spline.Controls {
		standalone, ok := transforms[child.Span.VarName()]
		if !ok || len(child.Params) != 16 || child.Value[12] != float64(3+i*3) || standalone.Value[12] != child.Value[12] {
			t.Fatalf("control %d lost local value or editable span: %+v", i, child)
		}
	}
}

func TestSplineRejectsInvalidShapes(t *testing.T) {
	matrix := "[[1,0,0,0],[0,1,0,0],[0,0,1,0],[0,0,0,1]]"
	for _, list := range []string{"[1,2,3]", "[m,m]", "[m,m,m,m]", "[[1,2,3],[4,5,6],[7,8,9]]"} {
		_, err := assembler.Assemble([]string{"m = " + matrix + "; return uiSpline(" + list + ");"}, assembler.SourceManager(&source.MockManager{}))
		if err == nil || !strings.Contains(err.Error(), "uiSpline expects") {
			t.Fatalf("%s: expected spline type error, got %v", list, err)
		}
	}
}

func TestExtractAnnotatedSpline(t *testing.T) {
	src := `@transform
start = [[1,0,0,0],[0,1,0,0],[0,0,1,0],[3,4,5,1]];
@transform
bend = [[1,0,0,0],[0,1,0,0],[0,0,1,0],[6,7,8,1]];
@transform
end = [[1,0,0,0],[0,1,0,0],[0,0,1,0],[9,10,11,1]];

@spline
path = [start, bend, end];

first = path[0];
last = path[2];
a = first[3];
b = last[3];
return length(p - a.xyz) + length(p - b.xyz);`
	controls := extract(t, src)
	var spline *ui.Control
	transforms := map[string]ui.Control{}
	for i, c := range controls {
		if c.Kind == ui.KindSpline {
			if spline != nil {
				t.Fatal("duplicate spline")
			}
			spline = &controls[i]
		} else if c.Kind == ui.KindTransform {
			transforms[c.Span.VarName()] = c
		}
	}
	if spline == nil || len(spline.Controls) != 3 || len(spline.Value) != 48 {
		t.Fatalf("missing ordered spline: %+v", controls)
	}
	for i, child := range spline.Controls {
		standalone, ok := transforms[child.Span.VarName()]
		if !ok || len(child.Params) != 16 || child.Value[12] != float64(3+i*3) || standalone.Value[12] != child.Value[12] {
			t.Fatalf("control %d lost local value or editable span: %+v", i, child)
		}
	}
}

func TestExtractAnnotatedSplineWithPromotedElements(t *testing.T) {
	src := `@spline
path = [
  [[1,0,0,0],[0,1,0,0],[0,0,1,0],[3,4,5,1]],
  [[1,0,0,0],[0,1,0,0],[0,0,1,0],[6,7,8,1]],
  [[1,0,0,0],[0,1,0,0],[0,0,1,0],[9,10,11,1]]
];

first = path[0];
last = path[2];
a = first[3];
b = last[3];
return length(p - a.xyz) + length(p - b.xyz);`
	controls := extract(t, src)
	var spline *ui.Control
	for i, c := range controls {
		if c.Kind == ui.KindSpline {
			if spline != nil {
				t.Fatal("duplicate spline")
			}
			spline = &controls[i]
		}
	}
	if spline == nil || len(spline.Controls) != 3 || len(spline.Value) != 48 {
		t.Fatalf("missing ordered spline with promoted elements: %+v", controls)
	}
	for i, child := range spline.Controls {
		if len(child.Params) != 16 || child.Value[12] != float64(3+i*3) {
			t.Fatalf("control %d lost local value: %+v", i, child)
		}
	}
}

func TestRewriteSplineControls(t *testing.T) {
	src := `frame(pos) {
  p0 = pos;
  return [[1, 0, 0, 0], [0, 1, 0, 0], [0, 0, 1, 0], [p0.x, p0.y, p0.z, 1]];
}
a = uiTransform(frame([1, 2, 3]));
b = uiTransform(frame([4, 5, 6]));
c = uiTransform([[1, 0, 0, 0], [0, 1, 0, 0], [0, 0, 1, 0], [7, 8, 9, 1]]);
path = uiSpline([a, b, c]);
first = path[0];
p4 = first[3];
return length(p - p4.xyz);`
	controls := extract(t, src)
	rewritten, splines, controlCount, err := ui.RewriteSplineControls(src, controls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if splines != 1 || controlCount != 3 {
		t.Fatalf("expected 1 spline and 3 controls, got %d splines and %d controls", splines, controlCount)
	}
	if strings.Contains(rewritten, "uiTransform(frame(") {
		t.Fatalf("expected computed frame calls inside uiTransform to be replaced by matrix literals:\n%s", rewritten)
	}
	if !strings.Contains(rewritten, "[[1,0,0,0],[0,1,0,0],[0,0,1,0],[1,2,3,1]]") ||
		!strings.Contains(rewritten, "[[1,0,0,0],[0,1,0,0],[0,0,1,0],[4,5,6,1]]") {
		t.Fatalf("missing marshaled matrix literals in rewritten source:\n%s", rewritten)
	}

	controlsAfter := extract(t, rewritten)
	rewrittenAgain, splines2, controlCount2, err := ui.RewriteSplineControls(rewritten, controlsAfter)
	if err != nil {
		t.Fatalf("unexpected error on second rewrite: %v", err)
	}
	if splines2 != 1 || controlCount2 != 3 || rewrittenAgain != rewritten {
		t.Fatalf("expected idempotent rewrite on literal controls")
	}

	analyzed, err := assembler.AnalyzePackage(rewritten, nil)
	if err != nil {
		t.Fatalf("AnalyzePackage: %v", err)
	}
	if got := ui.Extract(analyzed, nil); len(got) == 0 {
		t.Fatal("expected ui.Extract on un-inlined analyzed package to extract controls")
	}

	if _, _, _, err := ui.RewriteSplineControls(src, []ui.Control{{
		Kind:     ui.KindSpline,
		Controls: []ui.Control{{Span: controls[0].Span, Value: []float64{1, 2}}},
	}}); err == nil {
		t.Fatal("expected error for non-16 control value length")
	}
	if _, _, _, err := ui.RewriteSplineControls(src, []ui.Control{{
		Kind:     ui.KindSpline,
		Controls: []ui.Control{{Value: make([]float64, 16)}},
	}}); err == nil {
		t.Fatal("expected error for invalid 0-based control span")
	}
}
