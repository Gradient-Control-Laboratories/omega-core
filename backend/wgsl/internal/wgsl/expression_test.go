package wgsl_test

import (
	"testing"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/program"
)

func TestExpressionReads(t *testing.T) {
	// Literal Value
	val := wgsl.NewValue(nil, 42.0)
	assert.Equal(t, []string(nil), val.Reads())

	// VarRef
	varX := wgsl.NewVar(nil, "x", val, wgsl.TypeF32)
	refX := wgsl.NewVarRef(nil, varX)
	assert.Equal(t, []string{"x"}, refX.Reads())

	varY := wgsl.NewVar(nil, "y", val, wgsl.TypeF32)
	refY := wgsl.NewVarRef(nil, varY)
	assert.Equal(t, []string{"y"}, refY.Reads())

	// NumericOp
	numOp, err := wgsl.NewNumericOp(nil, []build.Expression{refX, refY}, build.PrecedenceAddition, nil, "+")
	assert.NoError(t, err)
	assert.Equal(t, []string{"x", "y"}, numOp.Reads())

	// Paren wrapper delegates
	paren := wgsl.Paren(numOp)
	assert.Equal(t, []string{"x", "y"}, paren.Reads())

	// Conversions delegate via sub Expression
	toBool := numOp.ToBool()
	assert.Equal(t, []string{"x", "y"}, toBool.Reads())
	toNum := toBool.ToNum()
	assert.Equal(t, []string{"x", "y"}, toNum.Reads())

	// Square
	sq, err := wgsl.NewSquare(nil, refX, nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{"x"}, sq.Reads())

	// Member access on Vec
	vecAcc, err := wgsl.NewVecMemberAccess("x", nil, refX)
	assert.NoError(t, err)
	assert.Equal(t, []string{"x"}, vecAcc.Reads())

	// FieldAccess on struct
	fieldAcc := wgsl.NewFieldAccess(nil, refX, "v_1", wgsl.TypeF32)
	assert.Equal(t, []string{"x"}, fieldAcc.Reads())

	// FuncCall
	fc, err := wgsl.NewFuncCall(nil, "helper", []wgsl.Expression{refX, refY}, false, wgsl.TypeF32)
	assert.NoError(t, err)
	assert.Equal(t, []string{"x", "y"}, fc.Reads())

	// FuncCall with PVarIsFirstArg
	fcP, err := wgsl.NewFuncCall(nil, "sphere", []wgsl.Expression{refX, refY}, true, wgsl.TypeF32)
	assert.NoError(t, err)
	assert.Equal(t, []string{program.PVarName, "y"}, fcP.Reads())

	// Select2
	sel2, err := wgsl.NewSelect2(nil, toBool, refX, refY, nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{"x", "y"}, sel2.Reads())
}
