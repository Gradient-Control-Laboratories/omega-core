package eval_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/collect"
	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/types"
	parserlib "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

// sliderChecker marks references to ranged variables as runtime parameters,
// as the assembler's checker does.
type sliderChecker map[string]bool

func (c sliderChecker) IsParameter(op *program.Operand) bool {
	ref := op.GetReference()
	return ref != nil && c[ref.Name]
}

func analyzed(t *testing.T, src string) *program.Program {
	t.Helper()
	prog, err := collect.Program([]string{src}, &source.MockManager{}, parserlib.ParsePackage)
	assert.NoError(t, err)
	assert.NoError(t, types.Analyze(prog))
	return prog
}

func isLiteral(op *program.Operand) bool {
	_, ok := op.T.(*program.Operand_Value)
	return ok
}

// A scalar constant the fold has evaluated, whether an expression or a call,
// becomes its literal: same type, value and source span, the subtree gone.
// The second fold finds nothing more to do.
func TestFoldCollapsesScalarConstants(t *testing.T) {
	prog := analyzed(t, `
k = 2.0 * 3.0 + 4.0;
m = max(k, 7.0);
return sphere(m);
`)
	file := prog.Primary.Files[0]
	kRoot, mRoot := file.Assignments[0].Expression, file.Assignments[1].Expression
	kStart, kEnd := kRoot.Start, kRoot.End
	assert.False(t, isLiteral(kRoot))
	assert.False(t, isLiteral(mRoot))

	mod, err := eval.Fold(prog, sliderChecker{})
	assert.NoError(t, err)
	assert.True(t, mod)

	assert.True(t, isLiteral(kRoot), "the expression root should be its literal")
	assert.True(t, isLiteral(mRoot), "the call root should be its literal")
	assert.Equal(t, 10.0, kRoot.GetValue())
	assert.Equal(t, 10.0, mRoot.GetValue())
	assert.True(t, approxEqual(kRoot.ConstValues, []float64{10}))
	assert.Equal(t, program.TypeKind_SCALAR, kRoot.Type.Kind)
	assert.True(t, kRoot.Start == kStart && kRoot.End == kEnd, "the span is the composite's")

	mod, err = eval.Fold(prog, sliderChecker{})
	assert.NoError(t, err)
	assert.False(t, mod, "a collapsed program is a fixed point")
}

// Only scalars collapse: a constant vector keeps its expression and its stamp.
func TestFoldCollapseKeepsVectors(t *testing.T) {
	prog := analyzed(t, `
v = [1.0, 2.0] * 2.0;
return length(p.xy - v);
`)
	_, err := eval.Fold(prog, sliderChecker{})
	assert.NoError(t, err)
	root := prog.Primary.Files[0].Assignments[0].Expression
	assert.False(t, isLiteral(root))
	assert.True(t, approxEqual(root.ConstValues, []float64{2, 4}))
}

// A ranged assignment's root is the literal the slider UI rewrites in place:
// it is stamped but left as written, and what reads the slider is a parameter.
func TestFoldCollapseKeepsRangedRoot(t *testing.T) {
	prog := analyzed(t, `
@range(0, 20) s = 2.0 * 3.0;
t = s + 1.0;
u = 4.0 + 4.0;
return sphere(t + u);
`)
	_, err := eval.Fold(prog, sliderChecker{"s": true})
	assert.NoError(t, err)
	file := prog.Primary.Files[0]
	assert.False(t, isLiteral(file.Assignments[0].Expression), "the ranged root stays as written")
	assert.True(t, approxEqual(file.Assignments[0].Expression.ConstValues, []float64{6}))
	assert.False(t, isLiteral(file.Assignments[1].Expression), "a slider's dependent is not constant")
	assert.Equal(t, 0, len(file.Assignments[1].Expression.ConstValues))
	assert.True(t, isLiteral(file.Assignments[2].Expression), "an unrelated constant still collapses")
}

// Without a checker the fold stamps and leaves the structure: its verdicts
// may be cleared by a later fold that knows the parameters.
func TestFoldNilCheckerDoesNotCollapse(t *testing.T) {
	prog := analyzed(t, `
k = 2.0 * 3.0 + 4.0;
return sphere(k);
`)
	_, err := eval.Fold(prog, nil)
	assert.NoError(t, err)
	root := prog.Primary.Files[0].Assignments[0].Expression
	assert.False(t, isLiteral(root))
	assert.True(t, approxEqual(root.ConstValues, []float64{10}))
}

// Literals inside a UI marker's argument are bound by the UI to their
// operands; a constant expression there stays an expression.
func TestFoldCollapseLeavesUIMarkerArguments(t *testing.T) {
	prog := analyzed(t, `
c = uiPoint([1.0 + 2.0, 0.0, 0.0]);
return sphere(length(p - c));
`)
	_, err := eval.Fold(prog, sliderChecker{})
	assert.NoError(t, err)
	call := prog.Primary.Files[0].Assignments[0].Expression.GetFnCall()
	assert.NotNil(t, call)
	vec := call.Arguments[0].GetExpression()
	assert.NotNil(t, vec)
	assert.False(t, isLiteral(vec.Operands[0]), "the sum inside the marker stays as written")
	assert.True(t, approxEqual(vec.Operands[0].ConstValues, []float64{3}))
}
