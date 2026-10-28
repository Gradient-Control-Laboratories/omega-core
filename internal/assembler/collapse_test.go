package lib

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/program"
)

// An assembled program carries every scalar constant as a literal. A catalog
// lookup inlined with a constant argument used to stay as its evaluated tree
// at every use; Fastener Cosmetics was 23 MB of such trees around 1,088
// operands that read p.
func TestAssembleCollapsesScalarConstantsToLiterals(t *testing.T) {
	src := `
d = 5.0;
pitch = thread.metricPitch(d);
depth = 5.0 * 0.866 * pitch / 8.0;
return cylinderZ([0.0, 0.0, 0.0], d / 2.0 - depth);
`
	res, err := Assemble([]string{src}, source.NewLocalManager("../../../common/source/package"))
	if err != nil {
		t.Fatal(err)
	}
	composite := 0
	var walk func(op *program.Operand)
	walk = func(op *program.Operand) {
		if op == nil {
			return
		}
		if len(op.ConstValues) == 1 {
			if _, lit := op.T.(*program.Operand_Value); !lit {
				composite++
			}
		}
		if ex := op.GetExpression(); ex != nil {
			for _, o := range ex.Operands {
				walk(o)
			}
		}
		if fc := op.GetFnCall(); fc != nil {
			for _, o := range fc.Arguments {
				walk(o)
			}
		}
	}
	// The radius, d/2 less the ISO basic thread depth 5H/8 at the M5 coarse
	// pitch 0.8, reaches the cylinder as one literal.
	const radius = 5.0/2.0 - 5.0*0.866*0.8/8.0
	radiusLiteral := false
	var find func(op *program.Operand)
	find = func(op *program.Operand) {
		if op == nil {
			return
		}
		if v, ok := op.T.(*program.Operand_Value); ok && math.Abs(v.Value-radius) < 1e-9 {
			radiusLiteral = true
		}
		if ex := op.GetExpression(); ex != nil {
			for _, o := range ex.Operands {
				find(o)
			}
		}
		if fc := op.GetFnCall(); fc != nil {
			for _, o := range fc.Arguments {
				find(o)
			}
		}
	}
	for _, f := range res.Program.Primary.Files {
		for _, a := range f.Assignments {
			walk(a.Expression)
			find(a.Expression)
		}
		for _, r := range f.Returns {
			walk(r)
			find(r)
		}
	}
	if composite != 0 {
		t.Errorf("%d scalar constants are still composite expressions or calls", composite)
	}
	if !radiusLiteral {
		t.Errorf("the radius %g, through the catalog pitch, is not a literal in the program", radius)
	}
}

// A constant under a non-constant parent collapses too, including one an
// inlined body of this package brings in untyped: the thread form's radius
// stays a literal whether or not what surrounds it reads p.
func TestAssembleCollapsesConstantsUnderVaryingParents(t *testing.T) {
	src := `
n = 20.0;
m = 0.4;
f(a, b) {
  r = a / 2.0 * b * cos(0.26);
  return acos(r / max(length(p), r)) + r * r;
}
return f(n, m);
`
	res, err := Assemble([]string{src}, source.NewLocalManager("../../../common/source/package"))
	if err != nil {
		t.Fatal(err)
	}
	composite := 0
	var walk func(op *program.Operand)
	walk = func(op *program.Operand) {
		if op == nil {
			return
		}
		if len(op.ConstValues) == 1 && op.GetReference() == nil {
			if _, lit := op.T.(*program.Operand_Value); !lit {
				composite++
			}
		}
		if ex := op.GetExpression(); ex != nil {
			for _, o := range ex.Operands {
				walk(o)
			}
		}
		if fc := op.GetFnCall(); fc != nil {
			for _, o := range fc.Arguments {
				walk(o)
			}
		}
	}
	for _, f := range res.Program.Primary.Files {
		for _, a := range f.Assignments {
			walk(a.Expression)
		}
		for _, r := range f.Returns {
			walk(r)
		}
	}
	if composite != 0 {
		t.Errorf("%d scalar constants are still composite expressions or calls", composite)
	}
}
