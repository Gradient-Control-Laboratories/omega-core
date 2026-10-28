package program

import (
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
