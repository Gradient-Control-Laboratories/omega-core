package inline_test

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// A function that remaps a field it was handed has to be inlined when the
// field depends on p: emitted on its own, it would receive the field as a value
// already evaluated at the caller's p, and the remap could only return it.
func TestInline_RemapOfFieldParameter(t *testing.T) {
	src := `
moved(shape, offset) {
	return remap(shape, p - offset);
}
turned(shape, k) {
	return remap(shape, p * k);
}
return turned(moved(length(p) - 1, [1, 0, 0]), 2);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ret := prog.Primary.Files[0].Returns[0]
	if calls := callsIn(ret); len(calls) != 0 {
		t.Fatalf("expected both calls inlined into nested remaps, still calling %v:\n%s", calls, ret.PrettyPrint(0))
	}
	outer, ok := ret.T.(*program.Operand_Expression)
	if !ok || outer.Expression.Operation != program.Operation_REMAP {
		t.Fatalf("expected the outer remap, got:\n%s", ret.PrettyPrint(0))
	}
	inner, ok := outer.Expression.Operands[0].T.(*program.Operand_Expression)
	if !ok || inner.Expression.Operation != program.Operation_REMAP {
		t.Fatalf("expected the inner remap as the outer's target, got:\n%s", ret.PrettyPrint(0))
	}
}

// A wrapper that hands its field straight on to a function that remaps it
// remaps that field too, so it is inlined as well: emitted on its own, it would
// pass the remapping function a value already evaluated at the caller's p.
func TestInline_RemapThroughWrapper(t *testing.T) {
	src := `
moved(shape, offset) {
	return remap(shape, p - offset);
}
place(shape, offset) {
	return moved(shape, offset);
}
return place(length(p) - 1, [1, 0, 0]);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ret := prog.Primary.Files[0].Returns[0]
	if calls := callsIn(ret); len(calls) != 0 {
		t.Fatalf("expected the wrapper and its callee inlined into a remap, still calling %v:\n%s", calls, ret.PrettyPrint(0))
	}
	if e, ok := ret.T.(*program.Operand_Expression); !ok || e.Expression.Operation != program.Operation_REMAP {
		t.Fatalf("expected a remap, got:\n%s", ret.PrettyPrint(0))
	}
}

// A builtin package function has a body that reads p, so a field built from one
// depends on p: each call remapping sphere(25) is inlined, however many there
// are, not only the single call another rule would inline anyway.
func TestInline_RemapOfBuiltinShape(t *testing.T) {
	src := `
moved(shape, offset) {
	return remap(shape, p - offset);
}
a = moved(sphere(25), [-50, 0, 0]);
b = moved(sphere(25), [50, 0, 0]);
return a, b;
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, a := range prog.Primary.Files[0].Assignments {
		if calls := callsIn(a.Expression); len(calls) != 0 && calls[0] == "moved" {
			t.Fatalf("expected %s's call inlined into a remap, still calling %v:\n%s",
				a.Variables[0].Name, calls, a.Expression.PrettyPrint(0))
		}
	}
}

// A remap's target that calls a function with a field is inlined, even though
// the function remaps nothing itself: emitted on its own, it would be handed
// the remapped point but the field built against the caller's.
func TestInline_RemapTargetWithFieldArgument(t *testing.T) {
	src := `
cap(shape, h) {
	return max(shape, p.z - h);
}
return remap(cap(length(p) - 1, 0.5), p - [3, 0, 0]);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ret := prog.Primary.Files[0].Returns[0]
	if calls := callsIn(ret); len(calls) != 0 {
		t.Fatalf("expected the target inlined into the remap, still calling %v:\n%s", calls, ret.PrettyPrint(0))
	}
	if e, ok := ret.T.(*program.Operand_Expression); !ok || e.Expression.Operation != program.Operation_REMAP {
		t.Fatalf("expected a remap, got:\n%s", ret.PrettyPrint(0))
	}
}

// A parameter that does not depend on p is the same at any point, so remapping
// through it does not force a call inline: the call stays when its arguments
// are not constant.
func TestInline_RemapOfScalarParameterStays(t *testing.T) {
	src := `
ring(radius) {
	return remap(length(p.xy) - radius, [p.x, p.y, 0]);
}
r = 2;
r.min = 1;
r.max = 3;
return ring(r * time);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ret := prog.Primary.Files[0].Returns[0]
	if calls := callsIn(ret); len(calls) != 1 || calls[0] != "ring" {
		t.Fatalf("expected the call to ring to stay, got calls %v:\n%s", calls, ret.PrettyPrint(0))
	}
}

// callsIn lists the names of the user function calls in op.
func callsIn(op *program.Operand) []string {
	var out []string
	var walk func(o *program.Operand)
	walk = func(o *program.Operand) {
		if o == nil {
			return
		}
		switch t := o.T.(type) {
		case *program.Operand_FnCall:
			if t.FnCall.Reference != nil && t.FnCall.Reference.Prefix == "" && !program.IsBuiltin(t.FnCall.Reference.Name) {
				out = append(out, t.FnCall.Reference.Name)
			}
			for _, a := range t.FnCall.Arguments {
				walk(a)
			}
		case *program.Operand_Expression:
			for _, c := range t.Expression.Operands {
				walk(c)
			}
		}
	}
	walk(op)
	return out
}
