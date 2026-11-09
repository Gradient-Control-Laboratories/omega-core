package program_test

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

func TestTypeShapeMethods(t *testing.T) {
	scalarType := &program.Type{
		Kind: program.TypeKind_SCALAR,
		Size: 1,
	}
	if !scalarType.IsScalar() {
		t.Errorf("expected scalarType.IsScalar() to be true")
	}
	if scalarType.IsList() {
		t.Errorf("expected scalarType.IsList() to be false")
	}
	if scalarType.IsMatrix() {
		t.Errorf("expected scalarType.IsMatrix() to be false")
	}
	if str := scalarType.TypeString(); str != "num" {
		t.Errorf("expected TypeString() == 'num', got %q", str)
	}

	list1Type := &program.Type{
		Kind: program.TypeKind_LIST,
		Size: 1,
		ElemType: &program.Type{
			Kind: program.TypeKind_SCALAR,
			Size: 1,
		},
	}
	if list1Type.IsScalar() {
		t.Errorf("expected list1Type.IsScalar() to be false")
	}
	if !list1Type.IsList() {
		t.Errorf("expected list1Type.IsList() to be true")
	}
	if list1Type.IsMatrix() {
		t.Errorf("expected list1Type.IsMatrix() to be false")
	}
	if str := list1Type.TypeString(); str != "list1" {
		t.Errorf("expected TypeString() == 'list1', got %q", str)
	}

	vec3Type := &program.Type{
		Kind: program.TypeKind_LIST,
		Size: 3,
		ElemType: &program.Type{
			Kind: program.TypeKind_SCALAR,
			Size: 1,
		},
	}
	if !vec3Type.IsList() {
		t.Errorf("expected vec3Type.IsList() to be true")
	}
	if str := vec3Type.TypeString(); str != "vec3" {
		t.Errorf("expected TypeString() == 'vec3', got %q", str)
	}

	list6Type := &program.Type{
		Kind: program.TypeKind_LIST,
		Size: 6,
		ElemType: &program.Type{
			Kind: program.TypeKind_SCALAR,
			Size: 1,
		},
	}
	if !list6Type.IsList() {
		t.Errorf("expected list6Type.IsList() to be true")
	}
	if str := list6Type.TypeString(); str != "list6" {
		t.Errorf("expected TypeString() == 'list6', got %q", str)
	}

	mat3x3Type := &program.Type{
		Kind: program.TypeKind_MATRIX,
		Size: 3,
		ElemType: &program.Type{
			Kind: program.TypeKind_LIST,
			Size: 3,
			ElemType: &program.Type{
				Kind: program.TypeKind_SCALAR,
				Size: 1,
			},
		},
	}
	if !mat3x3Type.IsMatrix() {
		t.Errorf("expected mat3x3Type.IsMatrix() to be true")
	}
	if mat3x3Type.IsList() {
		t.Errorf("expected mat3x3Type.IsList() to be false")
	}
	if str := mat3x3Type.TypeString(); str != "list<vec3>" {
		t.Errorf("expected TypeString() == 'list<vec3>', got %q", str)
	}

	// Verify constructor functions
	cScalar := program.NewScalarType()
	if !cScalar.IsScalar() || cScalar.Kind != program.TypeKind_SCALAR || cScalar.Size != 1 {
		t.Errorf("unexpected NewScalarType(): %+v", cScalar)
	}
	cList := program.NewListType(3, nil)
	if !cList.IsList() || cList.Kind != program.TypeKind_LIST || cList.Size != 3 || !cList.ElemType.IsScalar() {
		t.Errorf("unexpected NewListType(): %+v", cList)
	}
	cMat := program.NewMatrixType(3, cList)
	if !cMat.IsMatrix() || cMat.Kind != program.TypeKind_MATRIX || cMat.Size != 3 {
		t.Errorf("unexpected NewMatrixType(): %+v", cMat)
	}
	cFn := program.NewFunctionType("pkg", "fn")
	if !cFn.IsFunction() || cFn.Kind != program.TypeKind_FUNCTION || cFn.PkgPath != "pkg" || cFn.FnPath != "fn" {
		t.Errorf("unexpected NewFunctionType(): %+v", cFn)
	}
	if cFn.IsNumeric() {
		t.Errorf("expected cFn.IsNumeric() to be false")
	}
	if !cScalar.IsNumeric() {
		t.Errorf("expected cScalar.IsNumeric() to be true")
	}
	unspec := &program.Type{Kind: program.TypeKind_TYPE_UNSPECIFIED}
	if unspec.IsNumeric() {
		t.Errorf("expected unspecified type to not be numeric")
	}
}

func TestTypeEqual(t *testing.T) {
	var nilType *program.Type
	if !nilType.Equal(nil) {
		t.Errorf("expected nil.Equal(nil) to be true")
	}

	scalar := program.NewScalarType()
	if nilType.Equal(scalar) {
		t.Errorf("expected nil.Equal(scalar) to be false")
	}
	if scalar.Equal(nilType) {
		t.Errorf("expected scalar.Equal(nil) to be false")
	}
	if !scalar.Equal(program.NewScalarType()) {
		t.Errorf("expected scalar.Equal(scalar) to be true")
	}

	list1 := program.NewListType(1, scalar)
	if scalar.Equal(list1) {
		t.Errorf("expected scalar.Equal(list1) to be false")
	}
	if list1.Equal(scalar) {
		t.Errorf("expected list1.Equal(scalar) to be false")
	}
	if !list1.Equal(program.NewListType(1, program.NewScalarType())) {
		t.Errorf("expected list1.Equal(list1) to be true")
	}

	vec3 := program.NewListType(3, scalar)
	if list1.Equal(vec3) {
		t.Errorf("expected list1.Equal(vec3) to be false")
	}
	if !vec3.Equal(program.NewListType(3, scalar)) {
		t.Errorf("expected vec3.Equal(vec3) to be true")
	}

	mat3 := program.NewMatrixType(3, vec3)
	if mat3.Equal(vec3) {
		t.Errorf("expected mat3.Equal(vec3) to be false")
	}
	if !mat3.Equal(program.NewMatrixType(3, program.NewListType(3, scalar))) {
		t.Errorf("expected mat3.Equal(mat3) to be true")
	}

	fn1 := program.NewFunctionType("pkg", "fn")
	fn2 := program.NewFunctionType("pkg", "other")
	if !fn1.Equal(program.NewFunctionType("pkg", "fn")) {
		t.Errorf("expected fn1.Equal(fn1) to be true")
	}
	if fn1.Equal(fn2) {
		t.Errorf("expected fn1.Equal(fn2) to be false")
	}
	if fn1.Equal(scalar) {
		t.Errorf("expected fn1.Equal(scalar) to be false")
	}
	if scalar.Equal(fn1) {
		t.Errorf("expected scalar.Equal(fn1) to be false")
	}
}
