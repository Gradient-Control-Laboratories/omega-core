package lib

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/program"
)

// The frame reads .u and .v refuse the two programs they cannot mean. Their
// values are checked through the evaluator in gsleval's frame_query_test.go.

// A frame defined by its own reading is a cycle, not an infinite expansion.
func TestFrameReadingItselfIsACycle(t *testing.T) {
	_, err := Assemble([]string{"bar = sphere(1.0);\nbar.u = bar.u + 1.0;\nreturn bar;\n"}, &source.MockManager{})
	if err == nil || !strings.Contains(err.Error(), "cyclic dependency") {
		t.Errorf("err = %v, want a cyclic dependency", err)
	}
}

// Inside a function body there is no shape to name, so the read is refused
// with the fix: read the frame at the call site and pass it in.
func TestFrameReadInAFunctionBodyIsRefused(t *testing.T) {
	src := "readU(s) { return s.u; }\nbar = sphere(1.0);\nbar.u = p.x;\nf = readU(bar);\nreturn f;\n"
	_, err := Assemble([]string{src}, &source.MockManager{})
	if err == nil || !strings.Contains(err.Error(), "read the frame at the call site") {
		t.Errorf("err = %v, want the call-site advice", err)
	}
}

// An accessor written straight into a property expression expands as it would
// in a declaration, so a material reads the frame where the shape is named:
// bar.material = turned(..., bar.u, ...). Each case assembles, and nothing
// left for the property pass reaches the lowered colour and material.
func TestAccessorInsideAPropertyExpands(t *testing.T) {
	const bar = "bar = cylinder([0.0, 0.0, 0.0], [0.0, 0.0, 1.0], 2.0);\nbar.u = atan(p.y, p.x);\n"
	cases := []struct {
		name, src string
		// readsFrame: the lowered return depends on p through bar's declared
		// frame, where a frameless read folds to a constant zero.
		readsFrame bool
	}{
		{"dp in color", "s = sphere(1.0);\ns.color = [abs(s.dp.x), 0.2, 0.2];\nreturn s;\n", false},
		{"u in color", bar + "bar.color = [abs(sin(bar.u)), 0.2, 0.2];\nreturn bar;\n", true},
		{"u.dp in material", bar + "bar.material = material.withLay(material.make(0.3, 1.0, 1.5), bar.u.dp, 1.0);\nreturn bar;\n", true},
		{"second in color", "a = sphere(1.0);\nb = remap(sphere(1.0), p - [1.5, 0.0, 0.0]);\nboth = min(a, b);\nboth.color = [clamp(both.second, 0.0, 1.0), 0.2, 0.2];\nreturn both;\n", false},
		{"box in color", "s = sphere(1.0);\ns.color = s.box[1] * 0.5;\nreturn s;\n", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := Assemble([]string{c.src}, &source.MockManager{})
			if err != nil {
				t.Fatalf("Assemble: %v", err)
			}
			lowered := append(res.Program.ColorReturns(), res.Program.MaterialReturns()...)
			if len(lowered) == 0 {
				t.Fatalf("no lowered colour or material return")
			}
			sawFrame := false
			for _, op := range lowered {
				if op := findOperation(op, propertyOperators...); op != program.Operation_UNSPECIFIED_OPERAND {
					t.Errorf("lowered return still holds %s", op)
				}
				if readsP(op) {
					sawFrame = true
				}
			}
			if c.readsFrame && !sawFrame {
				t.Errorf("lowered return does not read bar's frame")
			}
		})
	}
}

// propertyOperators are the operations the property pass expands, which no
// lowered return may still hold.
var propertyOperators = []program.Operation{
	program.Operation_DERIV, program.Operation_SECONDCLOSEST, program.Operation_BOUNDINGBOX,
	program.Operation_SURFACEU, program.Operation_SURFACEV,
}

// findOperation returns the first of ops used anywhere in op, or
// UNSPECIFIED_OPERAND if none is.
func findOperation(op *program.Operand, ops ...program.Operation) program.Operation {
	if op == nil {
		return program.Operation_UNSPECIFIED_OPERAND
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		for _, want := range ops {
			if t.Expression.Operation == want {
				return want
			}
		}
		for _, sub := range t.Expression.Operands {
			if got := findOperation(sub, ops...); got != program.Operation_UNSPECIFIED_OPERAND {
				return got
			}
		}
	case *program.Operand_FnCall:
		for _, a := range t.FnCall.Arguments {
			if got := findOperation(a, ops...); got != program.Operation_UNSPECIFIED_OPERAND {
				return got
			}
		}
	}
	return program.Operation_UNSPECIFIED_OPERAND
}

// readsP reports whether op refers to the implicit point p anywhere.
func readsP(op *program.Operand) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		return t.Reference.Prefix == "" && t.Reference.Name == program.PVarName
	case *program.Operand_Expression:
		for _, sub := range t.Expression.Operands {
			if readsP(sub) {
				return true
			}
		}
	case *program.Operand_FnCall:
		for _, a := range t.FnCall.Arguments {
			if readsP(a) {
				return true
			}
		}
	}
	return false
}

// bar.u.second asks for the second-closest surface of a frame coordinate,
// which is a value on a surface and not a distance to one, so it is refused
// with the read it most likely meant, written directly or through a
// declaration. It would otherwise read as the "no second surface" sentinel.
func TestFrameHasNoSecondClosestSurface(t *testing.T) {
	const both = "a = sphere(1.0);\na.u = 1.0;\nb = remap(sphere(1.0), p - [3.0, 0.0, 0.0]);\nb.u = 2.0;\nboth = min(a, b);\n"
	for name, src := range map[string]string{
		"inline":         both + "f = both.u.second;\nreturn f;\n",
		"through decl":   both + "fu = both.u;\nf = fu.second;\nreturn f;\n",
		"v through decl": both + "fv = both.v;\nf = fv.second;\nreturn f;\n",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Assemble([]string{src}, &source.MockManager{})
			if err == nil || !strings.Contains(err.Error(), "did you mean .second.u") {
				t.Errorf("err = %v, want the .second.u advice", err)
			}
		})
	}
}
