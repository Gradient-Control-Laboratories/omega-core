package lib

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/program"
)

// countUnder returns how many operands under op satisfy pred.
func countUnder(op *program.Operand, pred func(*program.Operand) bool) int {
	if op == nil {
		return 0
	}
	n := 0
	if pred(op) {
		n++
	}
	for _, child := range op.GetExpression().GetOperands() {
		n += countUnder(child, pred)
	}
	for _, arg := range op.GetFnCall().GetArguments() {
		n += countUnder(arg, pred)
	}
	return n
}

// countInPackage sums countUnder over a package's assignments, returns and
// function bodies.
func countInPackage(pkg *program.Package, pred func(*program.Operand) bool) int {
	n := 0
	var inFnType func(ft *program.FnType)
	inFnType = func(ft *program.FnType) {
		for _, g := range ft.Groups {
			for _, fn := range g.Functions {
				for _, a := range fn.Assignments {
					n += countUnder(a.Expression, pred)
				}
				for _, r := range fn.Returns {
					n += countUnder(r, pred)
				}
				for _, nested := range fn.FnTypes {
					inFnType(nested)
				}
			}
		}
	}
	for _, f := range pkg.Files {
		for _, a := range f.Assignments {
			n += countUnder(a.Expression, pred)
		}
		for _, r := range f.Returns {
			n += countUnder(r, pred)
		}
		for _, ft := range f.FnTypes {
			inFnType(ft)
		}
	}
	return n
}

// TestAssembleKeepsSliderSelect: a ternary whose condition reads a slider
// stays in the program, with the slider live. The viewer moves a slider
// without recompiling, so a branch chosen at the slider's current value would
// freeze — which is what happened when the mesh-slot fold ran a
// parameter-blind eval.Fold over the program: `r > 0.5 ? 5 : 2` became 5 and
// r disappeared from the program and from the slider list.
func TestAssembleKeepsSliderSelect(t *testing.T) {
	src := `r = 1.0;
r.min = 0.0;
r.max = 2.0;
b = sphere(r > 0.5 ? 5 : 2);
return b;
`
	res, err := Assemble([]string{src}, &source.MockManager{}, ExtractUI())
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	if len(res.Sliders) != 1 || res.Sliders[0].VariableName != "r" {
		t.Fatalf("want the slider r to survive assembly, got %+v", res.Sliders)
	}
	isSelect := func(op *program.Operand) bool {
		e := op.GetExpression()
		return e != nil && e.Operation == program.Operation_SELECT
	}
	readsR := func(op *program.Operand) bool {
		ref := op.GetReference()
		return ref != nil && ref.Prefix == "" && ref.Name == "r"
	}
	if n := countInPackage(res.Program.Primary, isSelect); n == 0 {
		t.Errorf("the select on the slider was pruned: %s", res.Program.Primary.Files[0].Assignments[len(res.Program.Primary.Files[0].Assignments)-1].Expression.PrettyPrint(0))
	}
	if n := countInPackage(res.Program.Primary, readsR); n == 0 {
		t.Errorf("nothing reads the slider r after assembly")
	}
}
