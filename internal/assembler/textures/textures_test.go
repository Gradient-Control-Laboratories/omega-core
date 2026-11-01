package textures_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/textures"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/program"
)

func fnCall(name string, args ...*program.Operand) *program.Operand {
	return &program.Operand{
		T: &program.Operand_FnCall{
			FnCall: &program.FnCall{
				Reference: &program.Reference{Name: name},
				Arguments: args,
			},
		},
	}
}

func val(v float64) *program.Operand {
	return &program.Operand{
		T: &program.Operand_Value{
			Value: v,
		},
	}
}

func expr(op program.Operation, operands ...*program.Operand) *program.Operand {
	return &program.Operand{
		T: &program.Operand_Expression{
			Expression: &program.Expression{
				Operation: op,
				Operands:  operands,
			},
		},
	}
}

func TestExtract_MeshSdfSlots(t *testing.T) {
	prog := &program.Program{
		Primary: &program.Package{
			Files: []*program.PackageFile{
				{
					Returns: []*program.Operand{
						expr(program.Operation_MIN, fnCall("bvhSdf", val(7)), fnCall("vdbSdf", val(3), val(64), val(5))),
					},
					Assignments: []*program.Assignment{
						{
							Expression: fnCall("bvhSdf", val(3)),
							Variables: []*program.Variable{
								{
									Name: "shapeWithProp",
									Properties: []*program.Property{
										{
											Name:       "color",
											Expression: fnCall("vdbSdf", val(9), val(128), val(2.5)),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Dependencies: []*program.Package{
			{
				Files: []*program.PackageFile{
					{
						Returns: []*program.Operand{
							fnCall("bvhSdf", val(1)),
						},
					},
				},
			},
		},
	}

	slots, global := textures.Extract(prog)

	// Distinct asset indices in sorted order: 1, 3, 7, 9
	assert.Len(t, slots, 4)
	assert.Equal(t, int32(0), slots[0].Slot)
	assert.Equal(t, int32(1), slots[0].AssetIndex)
	assert.Equal(t, program.MeshSdfKind_MESH_SDF_KIND_BVH, slots[0].Kind)
	assert.Equal(t, int32(0), slots[0].VoxelDivisor)
	assert.Equal(t, float64(0), slots[0].HalfBandVoxels)

	assert.Equal(t, int32(1), slots[1].Slot)
	assert.Equal(t, int32(3), slots[1].AssetIndex)
	assert.Equal(t, program.MeshSdfKind_MESH_SDF_KIND_VDB, slots[1].Kind)
	assert.Equal(t, int32(64), slots[1].VoxelDivisor)
	assert.Equal(t, float64(5), slots[1].HalfBandVoxels)

	assert.Equal(t, int32(2), slots[2].Slot)
	assert.Equal(t, int32(7), slots[2].AssetIndex)
	assert.Equal(t, program.MeshSdfKind_MESH_SDF_KIND_BVH, slots[2].Kind)
	assert.Equal(t, int32(0), slots[2].VoxelDivisor)
	assert.Equal(t, float64(0), slots[2].HalfBandVoxels)

	assert.Equal(t, int32(3), slots[3].Slot)
	assert.Equal(t, int32(9), slots[3].AssetIndex)
	assert.Equal(t, program.MeshSdfKind_MESH_SDF_KIND_VDB, slots[3].Kind)
	assert.Equal(t, int32(128), slots[3].VoxelDivisor)
	assert.Equal(t, float64(2.5), slots[3].HalfBandVoxels)

	// No noise, warp, or spinodoid
	assert.NotNil(t, global)
	assert.False(t, global.UsesNoiseTile)
	assert.False(t, global.UsesWarpField)
	assert.False(t, global.UsesSpinodoid)
}

func TestExtract_GlobalTextures(t *testing.T) {
	tests := []struct {
		name          string
		callName      string
		wantNoise     bool
		wantWarp      bool
		wantSpinodoid bool
	}{
		{"noiseHash3", "noiseHash3", true, false, false},
		{"worleyF1", "worleyF1", true, false, false},
		{"worleyFeature", "worleyFeature", true, false, false},
		{"warpField", "warpField", false, true, false},
		{"gaborSpinodoid", "gaborSpinodoid", false, false, true},
		{"gaborSpinodoidGradient", "gaborSpinodoidGradient", false, false, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prog := &program.Program{
				Primary: &program.Package{
					Files: []*program.PackageFile{
						{
							Returns: []*program.Operand{
								fnCall(tc.callName, val(0)),
							},
						},
					},
				},
			}

			_, global := textures.Extract(prog)
			assert.NotNil(t, global)
			assert.Equal(t, tc.wantNoise, global.UsesNoiseTile)
			assert.Equal(t, tc.wantWarp, global.UsesWarpField)
			assert.Equal(t, tc.wantSpinodoid, global.UsesSpinodoid)
		})
	}
}

func TestExtract_EmptyProgram(t *testing.T) {
	slots, global := textures.Extract(&program.Program{Primary: &program.Package{}})
	assert.Nil(t, slots)
	assert.NotNil(t, global)
	assert.False(t, global.UsesNoiseTile)
	assert.False(t, global.UsesWarpField)
	assert.False(t, global.UsesSpinodoid)
}

func TestExtract_ZeroReturnsIgnoresUnreferencedAssignments(t *testing.T) {
	prog := &program.Program{
		Primary: &program.Package{
			Files: []*program.PackageFile{
				{
					Returns: nil,
					Assignments: []*program.Assignment{
						{Expression: fnCall("bvhSdf", val(3))},
						{Expression: fnCall("noiseHash3", val(0))},
					},
				},
			},
		},
	}

	slots, global := textures.Extract(prog)
	assert.Nil(t, slots)
	assert.NotNil(t, global)
	assert.False(t, global.UsesNoiseTile)
	assert.False(t, global.UsesWarpField)
	assert.False(t, global.UsesSpinodoid)
}
