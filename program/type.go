package program

import "fmt"

// NewScalarType returns a new scalar Type.
func NewScalarType() *Type {
	return &Type{
		Kind: TypeKind_SCALAR,
		Size: 1,
	}
}

// NewListType returns a new list/vector Type with the given element count and element type.
func NewListType(size int32, elem *Type) *Type {
	if elem == nil {
		elem = NewScalarType()
	}
	return &Type{
		Kind:     TypeKind_LIST,
		Size:     size,
		ElemType: elem,
	}
}

// NewMatrixType returns a new matrix Type with the given column count and column vector element type.
func NewMatrixType(cols int32, elem *Type) *Type {
	return &Type{
		Kind:     TypeKind_MATRIX,
		Size:     cols,
		ElemType: elem,
	}
}

// NewFunctionType returns a new function Type referencing the given function path.
func NewFunctionType(pkgPath, fnPath string) *Type {
	return &Type{
		Kind:    TypeKind_FUNCTION,
		PkgPath: pkgPath,
		FnPath:  fnPath,
	}
}

func (t *Type) TypeString() string {
	if t == nil {
		return "unknown"
	}
	if t.IsFunction() {
		return fmt.Sprintf("fn %s", t.FnPath)
	}
	if t.IsScalar() {
		return "num"
	}
	if t.IsMatrix() {
		rows := 0
		if t.ElemType != nil {
			rows = int(t.ElemType.Size)
		}
		return fmt.Sprintf("list<vec%d>", rows)
	}
	if t.IsList() {
		if t.Size == 1 {
			return "list1"
		}
		if t.Size >= 2 && t.Size <= 4 {
			return fmt.Sprintf("vec%d", t.Size)
		}
		return fmt.Sprintf("list%d", t.Size)
	}
	return fmt.Sprintf("vec%d", t.Size)
}

// IsScalar returns true if the type is a scalar numeric value.
func (t *Type) IsScalar() bool {
	return t != nil && t.Kind == TypeKind_SCALAR
}

// IsList returns true if the type is a 1D sequence/vector of elements.
func (t *Type) IsList() bool {
	return t != nil && t.Kind == TypeKind_LIST
}

// IsMatrix returns true if the type is a 2D matrix (list of column vectors).
func (t *Type) IsMatrix() bool {
	return t != nil && t.Kind == TypeKind_MATRIX
}

// IsFunction returns true if the type is a function reference.
func (t *Type) IsFunction() bool {
	return t != nil && t.Kind == TypeKind_FUNCTION
}

func (t *Type) IsNumeric() bool {
	return t != nil && (t.Kind == TypeKind_SCALAR || t.Kind == TypeKind_LIST || t.Kind == TypeKind_MATRIX)
}

// Equal returns true if two types are structurally identical.
func (t *Type) Equal(other *Type) bool {
	if t == nil {
		return other == nil
	}
	if other == nil || t.Kind != other.Kind {
		return false
	}
	switch t.Kind {
	case TypeKind_FUNCTION:
		return t.FnPath == other.FnPath && t.PkgPath == other.PkgPath
	case TypeKind_SCALAR:
		return true
	case TypeKind_LIST, TypeKind_MATRIX:
		return t.Size == other.Size && t.ElemType.Equal(other.ElemType)
	default:
		return false
	}
}
