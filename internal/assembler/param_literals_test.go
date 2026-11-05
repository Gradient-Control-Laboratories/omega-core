package lib

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/program"
)

// A call with constant arguments whose body reads a slider is not a
// constant: the slider stays live in what the call becomes. The fold used to
// evaluate f(1.0) through f's body to the slider's value and collapse the
// call to 3, which left the slider moving nothing.
func TestAssembleKeepsSliderReadInsideACall(t *testing.T) {
	src := `
r = 2.0;
r.min = 1.0;
r.max = 3.0;
f(x) { return x + r; }
return sphere(f(1.0));
`
	res, err := Assemble([]string{src}, source.NewLocalManager("../../stdlib"))
	if err != nil {
		t.Fatal(err)
	}
	if ret := primaryReturn(res.Program); !reads(ret, "r") {
		t.Fatalf("the return no longer reads the slider r:\n%s", ret.PrettyPrint(0))
	}
}

// A UI point read through a variable is not a constant either: the return
// still reads the point, and the viewer still has it to drag. The fold used
// to collapse pt.x * 2.0 + 1.0 to 3, and with its last reader gone the
// marker's declaration was removed as dead code.
func TestAssembleKeepsUIPointReadThroughVariable(t *testing.T) {
	src := `
pt = uiPoint([1.0, 2.0, 3.0]);
return sphere(pt.x * 2.0 + 1.0);
`
	res, err := Assemble([]string{src}, source.NewLocalManager("../../stdlib"), ExtractUI())
	if err != nil {
		t.Fatal(err)
	}
	if ret := primaryReturn(res.Program); !reads(ret, "pt") {
		t.Fatalf("the return no longer reads the point pt:\n%s", ret.PrettyPrint(0))
	}
	if len(res.UIControls) != 1 {
		t.Fatalf("want the point as the one UI control, got %d", len(res.UIControls))
	}
}

func primaryReturn(prog *program.Program) *program.Operand {
	files := prog.Primary.Files
	return files[len(files)-1].Returns[0]
}

// reads reports whether op reads the top-level name anywhere under it.
func reads(op *program.Operand, name string) bool {
	if op == nil {
		return false
	}
	if ref := op.GetReference(); ref != nil && ref.Name == name {
		return true
	}
	for _, child := range op.GetExpression().GetOperands() {
		if reads(child, name) {
			return true
		}
	}
	for _, arg := range op.GetFnCall().GetArguments() {
		if reads(arg, name) {
			return true
		}
	}
	return false
}
