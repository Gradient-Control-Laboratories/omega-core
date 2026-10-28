package ui_test

import (
	"strings"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/ui"
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
		_, err := assembler.Assemble([]string{"m = " + matrix + "; return uiSpline(" + list + ");"}, &source.MockManager{})
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
