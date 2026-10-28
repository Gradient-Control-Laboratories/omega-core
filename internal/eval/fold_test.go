package eval_test

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/collect"
	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/assert"
	parserlib "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func approxEqual(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.Abs(a[i]-b[i]) > 1e-9 {
			return false
		}
	}
	return true
}

func TestFoldBasic(t *testing.T) {
	src := `
a = 2.0;
b = [1.0, 2.0, 3.0] + a;
return b.zyx;
`
	prog, err := collect.Program([]string{src}, &source.MockManager{}, parserlib.ParsePackage)
	assert.NoError(t, err)

	_, err = eval.Fold(prog, nil)
	assert.NoError(t, err)

	file := prog.Primary.Files[0]
	varA := file.Assignments[0].Variables[0]
	varB := file.Assignments[1].Variables[0]
	opRet := file.Returns[0]

	// Verify constant values are populated
	assert.True(t, approxEqual(varA.ConstValues, []float64{2.0}))
	assert.True(t, approxEqual(varB.ConstValues, []float64{3.0, 4.0, 5.0}))
	assert.True(t, approxEqual(opRet.ConstValues, []float64{5.0, 4.0, 3.0}))
}

func TestFoldClampComponentWise(t *testing.T) {
	src := `
x = [1.0, 5.0, 10.0];
lo = [2.0, 3.0, 4.0];
hi = 6.0;
return clamp(x, lo, hi);
`
	prog, err := collect.Program([]string{src}, &source.MockManager{}, parserlib.ParsePackage)
	assert.NoError(t, err)

	_, err = eval.Fold(prog, nil)
	assert.NoError(t, err)

	file := prog.Primary.Files[0]
	opRet := file.Returns[0]

	assert.True(t, approxEqual(opRet.ConstValues, []float64{2.0, 5.0, 6.0}))
}

func TestFoldSelectConstantPruning_2WayTrue(t *testing.T) {
	src := `
cond = 1;
return cond ? 42 : p.x;
`
	prog, err := collect.Program([]string{src}, &source.MockManager{}, parserlib.ParsePackage)
	assert.NoError(t, err)

	mod, err := eval.Fold(prog, nil)
	assert.NoError(t, err)
	assert.True(t, mod)

	file := prog.Primary.Files[0]
	opRet := file.Returns[0]

	// The select node should have been replaced with the true branch (42)
	assert.True(t, approxEqual(opRet.ConstValues, []float64{42.0}))
	val, ok := opRet.T.(*program.Operand_Value)
	assert.True(t, ok)
	assert.Equal(t, val.Value, 42.0)
}

func TestFoldSelectConstantPruning_2WayFalse(t *testing.T) {
	src := `
cond = 0;
return cond ? p.x : 99;
`
	prog, err := collect.Program([]string{src}, &source.MockManager{}, parserlib.ParsePackage)
	assert.NoError(t, err)

	mod, err := eval.Fold(prog, nil)
	assert.NoError(t, err)
	assert.True(t, mod)

	file := prog.Primary.Files[0]
	opRet := file.Returns[0]

	// The select node should have been replaced with the false branch (99)
	assert.True(t, approxEqual(opRet.ConstValues, []float64{99.0}))
	val, ok := opRet.T.(*program.Operand_Value)
	assert.True(t, ok)
	assert.Equal(t, val.Value, 99.0)
}

func TestFoldSelectConstantPruning_3Way(t *testing.T) {
	// 3-way select: select(c, posBranch, zeroBranch, negBranch)
	srcPos := `
return select(10, 111, 222, 333);
`
	progPos, err := collect.Program([]string{srcPos}, &source.MockManager{}, parserlib.ParsePackage)
	assert.NoError(t, err)
	mod, err := eval.Fold(progPos, nil)
	assert.NoError(t, err)
	assert.True(t, mod)
	retPos := progPos.Primary.Files[0].Returns[0]
	assert.True(t, approxEqual(retPos.ConstValues, []float64{111.0}))

	srcZero := `
return select(0, 111, 222, 333);
`
	progZero, err := collect.Program([]string{srcZero}, &source.MockManager{}, parserlib.ParsePackage)
	assert.NoError(t, err)
	mod, err = eval.Fold(progZero, nil)
	assert.NoError(t, err)
	assert.True(t, mod)
	retZero := progZero.Primary.Files[0].Returns[0]
	assert.True(t, approxEqual(retZero.ConstValues, []float64{222.0}))

	srcNeg := `
return select(-5, 111, 222, 333);
`
	progNeg, err := collect.Program([]string{srcNeg}, &source.MockManager{}, parserlib.ParsePackage)
	assert.NoError(t, err)
	mod, err = eval.Fold(progNeg, nil)
	assert.NoError(t, err)
	assert.True(t, mod)
	retNeg := progNeg.Primary.Files[0].Returns[0]
	assert.True(t, approxEqual(retNeg.ConstValues, []float64{333.0}))
}

func TestFoldFunctionBody_Constants(t *testing.T) {
	src := `
calc(x) {
	scale = 2.0 * 3.0;
	offset = [1.0, 2.0].y;
	return x * scale + offset;
}
return calc(10);
`
	prog, err := collect.Program([]string{src}, &source.MockManager{}, parserlib.ParsePackage)
	assert.NoError(t, err)

	mod, err := eval.Fold(prog, nil)
	assert.NoError(t, err)
	assert.True(t, mod)

	fn := prog.Primary.Files[0].FnTypes[0].Groups[0].Functions[0]
	scaleVar := fn.Assignments[0].Variables[0]
	offsetVar := fn.Assignments[1].Variables[0]

	assert.True(t, approxEqual(scaleVar.ConstValues, []float64{6.0}))
	assert.True(t, approxEqual(offsetVar.ConstValues, []float64{2.0}))
}

func TestFoldFunctionBody_SelectPruning(t *testing.T) {
	srcConst := `
glyphA() {
	code = 65;
	return code == 65 ? (p.x + 10) : (p.y - 10);
}
return glyphA();
`
	progConst, err := collect.Program([]string{srcConst}, &source.MockManager{}, parserlib.ParsePackage)
	assert.NoError(t, err)

	mod, err := eval.Fold(progConst, nil)
	assert.NoError(t, err)
	assert.True(t, mod)

	fn := progConst.Primary.Files[0].FnTypes[0].Groups[0].Functions[0]
	fnRet := fn.Returns[0]

	// The select expression in fn.Returns[0] should be pruned to (p.x + 10)
	exprOp, ok := fnRet.T.(*program.Operand_Expression)
	assert.True(t, ok)
	assert.Equal(t, exprOp.Expression.Operation, program.Operation_ADD)
}
