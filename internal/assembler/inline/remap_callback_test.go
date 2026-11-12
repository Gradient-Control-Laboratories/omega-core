package inline_test

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// A field whose only read of p sits in a map or reduce function its callee
// declares still depends on p, so a function that remaps it is inlined. The
// callback is the callee's own, not one of the functions being transformed, and
// is looked for there. time keeps nearest a call, so its reduce stays in its
// body; called twice, moved is inlined only because its field reads p.
func TestInline_RemapOfFieldReadingPInACalleesReduce(t *testing.T) {
	src := `
nearest(points) {
	return reduce(1000, (best, q, i) { return min(best, length(p.xy - q)); }, points);
}
moved(shape, offset) {
	return remap(shape, p - offset);
}
a = moved(nearest([[0, 0], [time, 0]]), [-5, 0, 0]);
b = moved(nearest([[0, 0], [time, 0]]), [5, 0, 0]);
return min(a, b);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, a := range prog.Primary.Files[0].Assignments {
		for _, name := range callsIn(a.Expression) {
			if name == "moved" {
				t.Fatalf("expected %s's call to moved inlined into a remap, still calling it:\n%s",
					a.Variables[0].Name, a.Expression.PrettyPrint(0))
			}
		}
	}
}

// revolve, a builtin package function, remaps its profile. A profile reads p
// only inside the reduce functions of the builtin functions it calls, and
// revolve is still inlined: left a call, a backend emits it as a function that
// receives the profile already evaluated at the caller's point and returns it
// unrevolved.
func TestInline_RevolveOfProfileIsInlined(t *testing.T) {
	src := `
profile = profileXY([[1, 0], [2, 0], [2, 1], [1, 1]], [0, 0, 0, 0]);
return revolve(profile, [0, 0], [0, 1]);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ret := prog.Primary.Files[0].Returns[0]
	if callsTo(ret, "revolve") {
		t.Fatalf("expected revolve inlined into a remap, still calling it:\n%s", ret.PrettyPrint(0))
	}
}

// callsTo reports whether op calls the unqualified function name, builtin or not.
func callsTo(op *program.Operand, name string) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_FnCall:
		if t.FnCall.Reference != nil && t.FnCall.Reference.Prefix == "" && t.FnCall.Reference.Name == name {
			return true
		}
		for _, a := range t.FnCall.Arguments {
			if callsTo(a, name) {
				return true
			}
		}
	case *program.Operand_Expression:
		for _, c := range t.Expression.Operands {
			if callsTo(c, name) {
				return true
			}
		}
	}
	return false
}
