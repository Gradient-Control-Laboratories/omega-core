package inline_test

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

func isCallTo(name string) func(*program.Operand) bool {
	return func(op *program.Operand) bool {
		fc := op.GetFnCall()
		return fc != nil && fc.Reference != nil && fc.Reference.Prefix == "" && fc.Reference.Name == name
	}
}

// A document's own function that shares its name with a builtin's is that
// function, and is inlined on the first pass like any other. It used to be
// mistaken for the builtin and skipped, until a later pass — if the optimizer
// took one — found the builtin's unused sphere removed as dead code and the
// name free: whether it was inlined depended on how many passes ran.
func TestInlinesUserFunctionShadowingBuiltinName(t *testing.T) {
	src := `
sphere(radius) {
  return length(p) - radius * 2.0;
}
return sphere(2.5);
`
	ret := primaryReturn(inlineWithParameters(t, src, []string{}))
	if n := count(ret, isCallTo("sphere")); n != 0 {
		t.Errorf("the document's sphere was taken for the builtin and not inlined: %s", ret.PrettyPrint(0))
	}
	if n := count(ret, isOperation(program.Operation_MUL)); n != 1 {
		t.Errorf("want the body's radius * 2.0 inlined once, got %d: %s", n, ret.PrettyPrint(0))
	}
}

// Without a definition of its own, a call to a builtin's function is the
// builtin's, and this pass leaves it as it did before.
func TestLeavesBuiltinFunctionCall(t *testing.T) {
	src := `
return sphere(2.5);
`
	ret := primaryReturn(inlineWithParameters(t, src, []string{}))
	if n := count(ret, isCallTo("sphere")); n != 1 {
		t.Errorf("want the builtin sphere call left in place, got %d: %s", n, ret.PrettyPrint(0))
	}
}
