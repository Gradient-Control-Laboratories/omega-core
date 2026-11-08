package wgsl_test

import (
	"testing"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/program"
)

func TestWGSLType(t *testing.T) {
	// Nil fallback
	assert.Equal(t, wgsl.TypeF32, wgsl.WGSLType(nil, false))
	assert.Equal(t, wgsl.TypeBool, wgsl.WGSLType(nil, true))

	// Scalar
	scalarType := program.NewScalarType()
	assert.Equal(t, wgsl.TypeF32, wgsl.WGSLType(scalarType, false))
	assert.Equal(t, wgsl.TypeBool, wgsl.WGSLType(scalarType, true))

	// Numeric Vectors
	assert.Equal(t, wgsl.TypeVec2, wgsl.WGSLType(program.NewListType(2, scalarType), false))
	assert.Equal(t, wgsl.TypeVec3, wgsl.WGSLType(program.NewListType(3, scalarType), false))
	assert.Equal(t, wgsl.TypeVec4, wgsl.WGSLType(program.NewListType(4, scalarType), false))

	// Boolean Vectors
	assert.Equal(t, wgsl.TypeVec2Bool, wgsl.WGSLType(program.NewListType(2, scalarType), true))
	assert.Equal(t, wgsl.TypeVec3Bool, wgsl.WGSLType(program.NewListType(3, scalarType), true))
	assert.Equal(t, wgsl.TypeVec4Bool, wgsl.WGSLType(program.NewListType(4, scalarType), true))

	// Large Lists / Arrays
	arrType := wgsl.WGSLType(program.NewListType(8, scalarType), false)
	assert.Equal(t, "array<f32, 8>", arrType.Name)
	assert.Equal(t, &wgsl.TypeF32, arrType.ElemType)
	assert.Equal(t, program.NewListType(8, scalarType), arrType.ProgType)

	arrBoolType := wgsl.WGSLType(program.NewListType(5, scalarType), true)
	assert.Equal(t, "array<bool, 5>", arrBoolType.Name)
	assert.Equal(t, &wgsl.TypeBool, arrBoolType.ElemType)

	// Matrices: Primary tier (native matCxR<f32> for C, R in [2, 4])
	mat2 := program.NewMatrixType(2, program.NewListType(2, scalarType))
	assert.Equal(t, wgsl.TypeMat2x2, wgsl.WGSLType(mat2, false))

	mat3 := program.NewMatrixType(3, program.NewListType(3, scalarType))
	assert.Equal(t, wgsl.TypeMat3x3, wgsl.WGSLType(mat3, false))

	mat4 := program.NewMatrixType(4, program.NewListType(4, scalarType))
	assert.Equal(t, wgsl.TypeMat4x4, wgsl.WGSLType(mat4, false))

	mat3x2 := program.NewMatrixType(3, program.NewListType(2, scalarType))
	mat3x2WGSL := wgsl.WGSLType(mat3x2, false)
	assert.Equal(t, "mat3x2<f32>", mat3x2WGSL.Name)
	assert.Equal(t, "vec2<f32>", mat3x2WGSL.ElemType.Name)

	mat4x3 := program.NewMatrixType(4, program.NewListType(3, scalarType))
	mat4x3WGSL := wgsl.WGSLType(mat4x3, false)
	assert.Equal(t, "mat4x3<f32>", mat4x3WGSL.Name)
	assert.Equal(t, "vec3<f32>", mat4x3WGSL.ElemType.Name)

	// Secondary tier (C > 4, R in [2, 4]: array<vecR<f32>, C>)
	mat5x3 := program.NewMatrixType(5, program.NewListType(3, scalarType))
	mat5x3WGSL := wgsl.WGSLType(mat5x3, false)
	assert.Equal(t, "array<vec3<f32>, 5>", mat5x3WGSL.Name)
	assert.Equal(t, "vec3<f32>", mat5x3WGSL.ElemType.Name)

	// Tertiary tier (R > 4: array<array<f32, R>, C>)
	mat3x6 := program.NewMatrixType(3, program.NewListType(6, scalarType))
	mat3x6WGSL := wgsl.WGSLType(mat3x6, false)
	assert.Equal(t, "array<array<f32, 6>, 3>", mat3x6WGSL.Name)
	assert.Equal(t, "array<f32, 6>", mat3x6WGSL.ElemType.Name)

	mat6x6 := program.NewMatrixType(6, program.NewListType(6, scalarType))
	mat6x6WGSL := wgsl.WGSLType(mat6x6, false)
	assert.Equal(t, "array<array<f32, 6>, 6>", mat6x6WGSL.Name)
	assert.Equal(t, "array<f32, 6>", mat6x6WGSL.ElemType.Name)
}

func TestTypeInfoStruct(t *testing.T) {
	assert.False(t, wgsl.TypeF32.IsStruct())
	assert.Equal(t, "f32", wgsl.TypeF32.String())
	assert.Equal(t, program.NewScalarType(), wgsl.TypeF32.ProgType)

	structInfo := wgsl.TypeInfo{
		Name:       "gsl_struct_1",
		FieldNames: []string{"v_1", "v_2"},
		Types:      []wgsl.TypeInfo{wgsl.TypeF32, wgsl.TypeVec3},
	}
	assert.True(t, structInfo.IsStruct())
	assert.Equal(t, "gsl_struct_1", structInfo.String())
	assert.Equal(t, "f32", structInfo.Types[0].Name)
	assert.Equal(t, "vec3<f32>", structInfo.Types[1].Name)
}

func TestTypeInfoVectorMethods(t *testing.T) {
	assert.True(t, wgsl.TypeF32.IsNumericScalar())
	assert.False(t, wgsl.TypeF32.IsBoolScalar())
	assert.False(t, wgsl.TypeF32.IsVector())
	assert.False(t, wgsl.TypeF32.IsNumericVector())
	assert.False(t, wgsl.TypeF32.IsBoolVector())
	assert.True(t, wgsl.TypeF32.IsScalarOrVector())

	assert.False(t, wgsl.TypeBool.IsNumericScalar())
	assert.True(t, wgsl.TypeBool.IsBoolScalar())
	assert.False(t, wgsl.TypeBool.IsVector())
	assert.True(t, wgsl.TypeBool.IsScalarOrVector())

	assert.False(t, wgsl.TypeVec2.IsNumericScalar())
	assert.False(t, wgsl.TypeVec2.IsBoolScalar())
	assert.True(t, wgsl.TypeVec2.IsVector())
	assert.True(t, wgsl.TypeVec2.IsNumericVector())
	assert.False(t, wgsl.TypeVec2.IsBoolVector())

	assert.False(t, wgsl.TypeVec3Bool.IsNumericScalar())
	assert.False(t, wgsl.TypeVec3Bool.IsBoolScalar())
	assert.True(t, wgsl.TypeVec3Bool.IsVector())
	assert.False(t, wgsl.TypeVec3Bool.IsNumericVector())
	assert.True(t, wgsl.TypeVec3Bool.IsBoolVector())

	assert.False(t, wgsl.TypeMat3x3.IsNumericScalar())
	assert.False(t, wgsl.TypeMat3x3.IsBoolScalar())
	assert.False(t, wgsl.TypeMat3x3.IsVector())
	assert.False(t, wgsl.TypeMat3x3.IsScalarOrVector())
	assert.True(t, wgsl.TypeMat3x3.IsMatrix())
	assert.False(t, wgsl.TypeMat3x3.IsArray())

	arrType := wgsl.WGSLType(program.NewListType(6, program.NewScalarType()), false)
	assert.True(t, arrType.IsArray())
	assert.False(t, arrType.IsMatrix())
	assert.Equal(t, 6, arrType.ArraySize())

	assert.Equal(t, 2, wgsl.TypeVec2.VectorSize())
	assert.Equal(t, 3, wgsl.TypeVec3.VectorSize())
	assert.Equal(t, 4, wgsl.TypeVec4.VectorSize())
	assert.Equal(t, 3, wgsl.TypeVec3Bool.VectorSize())
	assert.Equal(t, 0, wgsl.TypeF32.VectorSize())
}

func TestMatrixDimensions(t *testing.T) {
	c, r, ok := wgsl.TypeMat2x2.MatrixDimensions()
	assert.True(t, ok)
	assert.Equal(t, 2, c)
	assert.Equal(t, 2, r)
	assert.True(t, wgsl.TypeMat2x2.IsNativeMatrix())
	assert.True(t, wgsl.TypeMat2x2.IsMatrix())

	c, r, ok = wgsl.TypeMat3x3.MatrixDimensions()
	assert.True(t, ok)
	assert.Equal(t, 3, c)
	assert.Equal(t, 3, r)
	assert.True(t, wgsl.TypeMat3x3.IsNativeMatrix())

	c, r, ok = wgsl.TypeMat4x4.MatrixDimensions()
	assert.True(t, ok)
	assert.Equal(t, 4, c)
	assert.Equal(t, 4, r)
	assert.True(t, wgsl.TypeMat4x4.IsNativeMatrix())

	// Non-square native matrix
	c, r, ok = wgsl.TypeMat4x3.MatrixDimensions()
	assert.True(t, ok)
	assert.Equal(t, 4, c)
	assert.Equal(t, 3, r)
	assert.True(t, wgsl.TypeMat4x3.IsNativeMatrix())
	assert.True(t, wgsl.TypeMat4x3.IsMatrix())

	// Array-backed matrix
	mat5x3 := wgsl.TypeInfo{
		Name:     "array<vec3<f32>, 5>",
		ProgType: program.NewMatrixType(5, program.NewListType(3, program.NewScalarType())),
	}
	c, r, ok = mat5x3.MatrixDimensions()
	assert.True(t, ok)
	assert.Equal(t, 5, c)
	assert.Equal(t, 3, r)
	assert.False(t, mat5x3.IsNativeMatrix())
	assert.True(t, mat5x3.IsMatrix())

	// Non-matrix
	c, r, ok = wgsl.TypeVec3.MatrixDimensions()
	assert.False(t, ok)
	assert.Equal(t, 0, c)
	assert.Equal(t, 0, r)
	assert.False(t, wgsl.TypeVec3.IsNativeMatrix())
	assert.False(t, wgsl.TypeVec3.IsMatrix())
}
