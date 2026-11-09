package inline_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/inline"
	"gradientcontrol.com/omega/core/internal/closure"
	"gradientcontrol.com/omega/core/internal/collect"
	"gradientcontrol.com/omega/core/internal/link"
	"gradientcontrol.com/omega/core/internal/lower"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/types"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

// sliderChecker marks the named top-level variables as parameters, the way
// the assembler marks a document's sliders, so a reference to one is kept
// rather than folded.
type sliderChecker map[string]bool

func (c sliderChecker) IsParameter(op *program.Operand) bool {
	if ref, ok := op.T.(*program.Operand_Reference); ok && ref.Reference != nil {
		return c[ref.Reference.Name]
	}
	return false
}

// runPipelineWithSliders is runPipeline with the named variables kept as
// parameters through the inline pass.
func runPipelineWithSliders(src string, sliders sliderChecker) (*program.Program, error) {
	srcMgr := &source.MockManager{}
	prog, err := collect.Program([]string{src}, srcMgr, parser.ParsePackage)
	if err != nil {
		return nil, err
	}
	if err := link.References(prog, srcMgr, parser.ParsePackage); err != nil {
		return nil, err
	}
	if err := closure.Transform(prog); err != nil {
		return nil, err
	}
	if err := link.References(prog, srcMgr, parser.ParsePackage); err != nil {
		return nil, err
	}
	if err := types.Analyze(prog); err != nil {
		return nil, err
	}
	if err := lower.Transform(prog); err != nil {
		return nil, err
	}
	if err := link.References(prog, srcMgr, parser.ParsePackage); err != nil {
		return nil, err
	}
	if err := types.Analyze(prog); err != nil {
		return nil, err
	}
	if _, err := inline.TransformWithParameters(prog, sliders); err != nil {
		return nil, err
	}
	if err := link.References(prog, srcMgr, parser.ParsePackage); err != nil {
		return nil, err
	}
	if err := types.Analyze(prog); err != nil {
		return nil, err
	}
	return prog, nil
}

// pattern's local is read twice, so inlining a call to it hoists the local
// into a temporary of the document's own.
const patternFn = `
pattern(s) {
  w = fract(p.x / s);
  return [w, w, 0];
}
`

// orderOf is the names each top-level declaration binds, in order, with the
// line each expression reads the slider x on marked.
func orderOf(prog *program.Program) (names []string, readsX []bool) {
	for _, a := range prog.Primary.Files[0].Assignments {
		name := ""
		for i, v := range a.Variables {
			if i > 0 {
				name += ","
			}
			name += v.Name
		}
		names = append(names, name)
		readsX = append(readsX, reads(a.Expression, "x"))
	}
	return names, readsX
}

func reads(op *program.Operand, name string) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		return t.Reference.Prefix == "" && t.Reference.Name == name
	case *program.Operand_Expression:
		for i, child := range t.Expression.Operands {
			if i == 1 && t.Expression.Operation == program.Operation_ACCESS {
				continue
			}
			if reads(child, name) {
				return true
			}
		}
	case *program.Operand_FnCall:
		for _, arg := range t.FnCall.Arguments {
			if reads(arg, name) {
				return true
			}
		}
	}
	return false
}

func indexOf(names []string, want string) int {
	for i, n := range names {
		if n == want {
			return i
		}
	}
	return -1
}

// A property is written below the part it belongs to, and the document may
// declare a slider between the two: a = sphere(5); x = 8; a.color =
// pattern(x) is in order as written. The temporaries the inlined pattern
// hoists for the property read x, so they are declared after x, not in front
// of a as the part's own temporaries are.
func TestInline_PropertyTemporariesFollowTheSliderTheyRead(t *testing.T) {
	src := patternFn + `
a = sphere(5);
x = 8;
x.min = 0.5;
x.max = 50;
a.color = pattern(x);
return a;
`
	prog, err := runPipelineWithSliders(src, sliderChecker{"x": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	names, readsX := orderOf(prog)
	xAt := indexOf(names, "x")
	if xAt < 0 {
		t.Fatalf("x is not declared: %v", names)
	}
	hoisted := 0
	for i, r := range readsX {
		if !r || names[i] == "x" {
			continue
		}
		hoisted++
		if i < xAt {
			t.Errorf("%s reads x at position %d, before x is declared at %d: %v", names[i], i, xAt, names)
		}
	}
	if hoisted == 0 {
		t.Fatalf("nothing was hoisted out of the property; the test needs a hoisted temporary: %v", names)
	}
	if aAt := indexOf(names, "a"); aAt < 0 || aAt > xAt {
		t.Errorf("a should stay declared before x: %v", names)
	}
}

// With the slider declared above the part, the property's temporaries keep
// their place in front of the part, as before.
func TestInline_PropertyTemporariesStayAheadOfThePartOtherwise(t *testing.T) {
	src := patternFn + `
x = 8;
x.min = 0.5;
x.max = 50;
a = sphere(5);
a.color = pattern(x);
return a;
`
	prog, err := runPipelineWithSliders(src, sliderChecker{"x": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	names, readsX := orderOf(prog)
	aAt := indexOf(names, "a")
	if aAt < 0 {
		t.Fatalf("a is not declared: %v", names)
	}
	hoisted := 0
	for i, r := range readsX {
		if !r || names[i] == "x" {
			continue
		}
		hoisted++
		if i > aAt {
			t.Errorf("%s was moved below a at %d: %v", names[i], aAt, names)
		}
	}
	if hoisted == 0 {
		t.Fatalf("nothing was hoisted out of the property; the test needs a hoisted temporary: %v", names)
	}
}
