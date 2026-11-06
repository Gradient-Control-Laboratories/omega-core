package program

import (
	"math"
	"testing"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/internal/test/assert"
)

func TestOperationIsBool(t *testing.T) {
	logicalOps := []Operation{
		Operation_EQ, Operation_NE,
		Operation_LT, Operation_LE,
		Operation_GT, Operation_GE,
		Operation_AND, Operation_OR,
		Operation_NOT,
	}
	for _, op := range logicalOps {
		assert.True(t, op.IsBool(), "expected %s to be boolean", op.Name())
	}

	nonLogicalOps := []Operation{
		Operation_ADD, Operation_SUB, Operation_MUL, Operation_DIV,
		Operation_COLLECT, Operation_ACCESS, Operation_REMAP,
	}
	for _, op := range nonLogicalOps {
		assert.False(t, op.IsBool(), "expected %s to not be boolean", op.Name())
	}
}

func TestOperandIsBool(t *testing.T) {
	var nilOp *Operand
	assert.False(t, nilOp.IsBool())

	valOp := NewValue(42.0, nil, nil)
	assert.False(t, valOp.IsBool())

	exprOp := NewExpression(Operation_EQ, []*Operand{valOp, valOp}, nil, nil)
	assert.True(t, exprOp.IsBool())

	addOp := NewExpression(Operation_ADD, []*Operand{valOp, valOp}, nil, nil)
	assert.False(t, addOp.IsBool())
}

func TestProgramMeshSdfSlotsAndGlobalTexturesSerialization(t *testing.T) {
	orig := &Program{
		Primary: &Package{
			Files: []*PackageFile{
				{Path: "main.gsl"},
			},
		},
		MeshSdfSlots: []*MeshSdfSlot{
			{Slot: 0, AssetIndex: 3},
			{Slot: 1, AssetIndex: 7},
		},
		GlobalTextures: &GlobalTextures{
			UsesNoiseTile: true,
			UsesWarpField: false,
			UsesSpinodoid: true,
		},
	}

	data, err := proto.Marshal(orig)
	assert.NoError(t, err)

	decoded, err := DecodeProgram(data)
	assert.NoError(t, err)
	assert.NotNil(t, decoded)

	assert.Len(t, decoded.MeshSdfSlots, 2)
	assert.Equal(t, int32(0), decoded.MeshSdfSlots[0].Slot)
	assert.Equal(t, int32(3), decoded.MeshSdfSlots[0].AssetIndex)
	assert.Equal(t, int32(1), decoded.MeshSdfSlots[1].Slot)
	assert.Equal(t, int32(7), decoded.MeshSdfSlots[1].AssetIndex)

	assert.NotNil(t, decoded.GlobalTextures)
	assert.True(t, decoded.GlobalTextures.UsesNoiseTile)
	assert.False(t, decoded.GlobalTextures.UsesWarpField)
	assert.True(t, decoded.GlobalTextures.UsesSpinodoid)
}

func TestPackageAndProgramReturns(t *testing.T) {
	emptyPkg := &Package{}
	assert.Nil(t, emptyPkg.Returns())
	emptyPkg.SetReturns([]*Operand{NewValue(1, nil, nil)})
	assert.Nil(t, emptyPkg.Returns())

	emptyProg := &Program{Primary: emptyPkg}
	assert.Nil(t, emptyProg.Returns())
	emptyProg.SetReturns([]*Operand{NewValue(1, nil, nil)})
	assert.Nil(t, emptyProg.Returns())

	r1 := NewValue(10, nil, nil)
	r2 := NewValue(20, nil, nil)
	r3 := NewValue(30, nil, nil)
	pkg := &Package{
		Files: []*PackageFile{
			{Path: "first.gsl", Returns: []*Operand{r1}},
			{Path: "last.gsl", Returns: []*Operand{r1, r2}},
		},
	}
	prog := &Program{Primary: pkg}

	assert.Equal(t, []*Operand{r1, r2}, pkg.Returns())
	assert.Equal(t, []*Operand{r1, r2}, prog.Returns())

	prog.SetReturns([]*Operand{r3})
	assert.Equal(t, []*Operand{r3}, pkg.Returns())
	assert.Equal(t, []*Operand{r3}, prog.Returns())
	assert.Equal(t, []*Operand{r1}, pkg.Files[0].Returns)

	pkg.SetReturns(nil)
	assert.Nil(t, prog.Returns())
}

func TestModuleFunction(t *testing.T) {
	fnTypes, fn0 := AddFunction(nil, "helper", nil)
	fn0.Returns = []*Operand{NewValue(1, nil, nil)}
	fnTypes, fn2 := AddFunction(fnTypes, "helper", []string{"a", "b"})
	fn2.Returns = []*Operand{NewValue(2, nil, nil)}

	pkg := &Package{
		Files: []*PackageFile{
			{},
			{FnTypes: fnTypes},
		},
	}

	assert.Same(t, fn0, ModuleFunction(pkg, "helper", 0))
	assert.Same(t, fn2, ModuleFunction(pkg, "helper", 2))
	assert.Nil(t, ModuleFunction(pkg, "helper", 1))
	assert.Nil(t, ModuleFunction(pkg, "missing", 0))
	assert.Nil(t, ModuleFunction(nil, "helper", 0))
}

func TestTopLevelAssignment(t *testing.T) {
	scaleExpr := NewValue(1.018, NewLoc(1, 1, 0), NewLoc(1, 6, 5))
	multiAssign := NewAssignment([]string{"a", "b"}, NewFnCall("", "pair", nil, nil, nil), nil, nil)
	singleAssign := NewAssignment([]string{"partScale"}, scaleExpr, nil, nil)

	pkg := &Package{
		Files: []*PackageFile{
			{Assignments: []*Assignment{multiAssign}},
			{Assignments: []*Assignment{singleAssign}},
		},
	}

	assert.Same(t, scaleExpr, TopLevelAssignment(pkg, "partScale"))
	assert.Nil(t, TopLevelAssignment(pkg, "a"))
	assert.Nil(t, TopLevelAssignment(pkg, "missing"))
	assert.Nil(t, TopLevelAssignment(nil, "partScale"))
}

func TestConstValues(t *testing.T) {
	stamped := NewReference("", "x", nil, nil)
	stamped.ConstValues = []float64{10, 20, 30}
	got, ok := ConstValues(stamped)
	assert.True(t, ok)
	assert.Equal(t, []float64{10, 20, 30}, got)

	valOp := NewValue(42.5, nil, nil)
	got, ok = ConstValues(valOp)
	assert.True(t, ok)
	assert.Equal(t, []float64{42.5}, got)

	negOp := NewExpression(Operation_NEG, []*Operand{NewValue(3.5, nil, nil)}, nil, nil)
	got, ok = ConstValues(negOp)
	assert.True(t, ok)
	assert.Equal(t, []float64{-3.5}, got)

	collectOp := NewExpression(Operation_COLLECT, []*Operand{
		NewValue(1, nil, nil),
		NewExpression(Operation_COLLECT, []*Operand{NewValue(2, nil, nil), negOp}, nil, nil),
	}, nil, nil)
	got, ok = ConstValues(collectOp)
	assert.True(t, ok)
	assert.Equal(t, []float64{1, 2, -3.5}, got)

	for _, invalid := range []*Operand{
		nil,
		NewReference("", "unresolved", nil, nil),
		NewValue(math.NaN(), nil, nil),
		NewValue(math.Inf(1), nil, nil),
		{ConstValues: []float64{1, math.NaN()}},
		{ConstValues: []float64{math.Inf(-1)}},
		NewExpression(Operation_ADD, []*Operand{NewValue(1, nil, nil), NewValue(2, nil, nil)}, nil, nil),
	} {
		_, ok := ConstValues(invalid)
		assert.False(t, ok)
	}
}

func TestEqualModuloLocations(t *testing.T) {
	a := NewExpression(Operation_ADD, []*Operand{
		NewReference("", "partScale", NewLoc(1, 1, 0), NewLoc(1, 10, 9)),
		NewValue(2, NewLoc(1, 13, 12), NewLoc(1, 14, 13)),
	}, NewLoc(1, 1, 0), NewLoc(1, 14, 13))
	a.ConstValues = []float64{3.018}

	b := NewExpression(Operation_ADD, []*Operand{
		NewReference("", "partScale", NewLoc(10, 5, 100), NewLoc(10, 14, 109)),
		NewValue(2, NewLoc(10, 17, 112), NewLoc(10, 18, 113)),
	}, NewLoc(10, 5, 100), NewLoc(10, 18, 113))

	assert.True(t, EqualModuloLocations(a, b))
	assert.NotNil(t, a.Start)
	assert.Len(t, a.ConstValues, 1)

	c := NewExpression(Operation_MUL, []*Operand{
		NewReference("", "partScale", nil, nil),
		NewValue(2, nil, nil),
	}, nil, nil)
	assert.False(t, EqualModuloLocations(a, c))
	assert.False(t, EqualModuloLocations(a, nil))
	assert.False(t, EqualModuloLocations(nil, b))
	assert.False(t, EqualModuloLocations(nil, nil))
}
