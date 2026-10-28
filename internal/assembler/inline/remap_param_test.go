package inline_test

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// A shape passed into a function that remaps it must inline. Built as a real
// function, WGSL received the shape already evaluated at the caller's p and the
// remap returned it unmoved, while fidget moved it (Driver Bits' standing() and
// lying() helpers).
func TestInline_RemappedShapeParameter(t *testing.T) {
	src := `
moved(f) {
	return remap(f, [p.x - 5.0, p.y, p.z]);
}
s = length(p) - 1.0;
return moved(s);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ret := prog.Primary.Files[0].Returns[0]
	exprOp, ok := ret.T.(*program.Operand_Expression)
	if !ok {
		t.Fatalf("expected moved(s) inlined to its remap, got %T", ret.T)
	}
	if exprOp.Expression.Operation != program.Operation_REMAP {
		t.Errorf("expected Operation_REMAP, got %v", exprOp.Expression.Operation)
	}
}

// Through a function that hands its parameter on to one that remaps it.
func TestInline_RemappedShapeParameterPassedOn(t *testing.T) {
	src := `
moved(f) {
	return remap(f, [p.x - 5.0, p.y, p.z]);
}
outer(g) {
	return moved(g);
}
s = length(p) - 1.0;
return outer(s);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ret := prog.Primary.Files[0].Returns[0]
	exprOp, ok := ret.T.(*program.Operand_Expression)
	if !ok {
		t.Fatalf("expected outer(s) inlined through moved to its remap, got %T", ret.T)
	}
	if exprOp.Expression.Operation != program.Operation_REMAP {
		t.Errorf("expected Operation_REMAP, got %v", exprOp.Expression.Operation)
	}
}

// A scalar reads no p and passes by value correctly, even an animated one, so
// the call stays a function.
func TestInline_RemappedScalarParameterStaysACall(t *testing.T) {
	src := `
sized(r) {
	return remap(length(p) - r, [p.x - 5.0, p.y, p.z]);
}
return sized(mod(time, 1.0) + 1.0);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ret := prog.Primary.Files[0].Returns[0]
	fc, ok := ret.T.(*program.Operand_FnCall)
	if !ok {
		t.Fatalf("expected sized(...) to stay a call, got %T", ret.T)
	}
	if fc.FnCall.Reference.Name != "sized" {
		t.Errorf("expected a call to sized, got %q", fc.FnCall.Reference.Name)
	}
}
