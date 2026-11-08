package wgsl

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// TypeInfo represents a concrete WGSL type or synthesized multi-return struct.
type TypeInfo struct {
	Name       string        // Emitted WGSL type name: "f32", "vec3<f32>", "mat3x3<f32>", or "gsl_struct_1"
	ProgType   *program.Type // Backing program.Type (retained from input or flyweight)
	ElemType   *TypeInfo     // Element type for arrays/matrices (nil for scalars/structs)
	FieldNames []string      // Struct member names: ["v_1", "v_2"], nil for single values
	Types      []TypeInfo    // Field types: [TypeF32, TypeVec3] for structs, nil for single values
}

// IsStruct returns true if this TypeInfo represents a multi-value return struct.
func (t TypeInfo) IsStruct() bool {
	return len(t.FieldNames) > 0 || strings.HasPrefix(t.Name, "gsl_struct")
}

// IsVector returns true if this TypeInfo represents a vector (vec2..4 of f32 or bool).
func (t TypeInfo) IsVector() bool {
	return t.IsNumericVector() || t.IsBoolVector()
}

// IsNumericVector returns true if this TypeInfo represents a float vector (vec2..4 of f32).
func (t TypeInfo) IsNumericVector() bool {
	switch t.Name {
	case "vec2<f32>", "vec3<f32>", "vec4<f32>":
		return true
	default:
		return false
	}
}

// IsBoolVector returns true if this TypeInfo represents a boolean vector (vec2..4 of bool).
func (t TypeInfo) IsBoolVector() bool {
	switch t.Name {
	case "vec2<bool>", "vec3<bool>", "vec4<bool>":
		return true
	default:
		return false
	}
}

// IsNumericScalar returns true if this TypeInfo represents a float scalar (f32).
func (t TypeInfo) IsNumericScalar() bool {
	return t.Name == "f32"
}

// IsBoolScalar returns true if this TypeInfo represents a boolean scalar (bool).
func (t TypeInfo) IsBoolScalar() bool {
	return t.Name == "bool"
}

// IsArray returns true if this TypeInfo represents a fixed-size WGSL array.
func (t TypeInfo) IsArray() bool {
	return strings.HasPrefix(t.Name, "array<")
}

// IsMatrix returns true if this TypeInfo represents a matrix (either native matCxR or array-backed matrix).
func (t TypeInfo) IsMatrix() bool {
	return t.ProgType.Kind == program.TypeKind_MATRIX
}

// IsNativeMatrix returns true if this TypeInfo represents a native WGSL matrix (matCxR<f32>).
func (t TypeInfo) IsNativeMatrix() bool {
	return strings.HasPrefix(t.Name, "mat")
}

// MatrixDimensions returns column count, row count, and true if this represents a matrix,
// or (0, 0, false) otherwise.
func (t TypeInfo) MatrixDimensions() (cols, rows int, ok bool) {
	if t.ProgType.Kind == program.TypeKind_MATRIX {
		cols = int(t.ProgType.Size)
		rows = 3
		if t.ProgType.ElemType != nil && t.ProgType.ElemType.Size > 0 {
			rows = int(t.ProgType.ElemType.Size)
		}
		return cols, rows, true
	}
	return 0, 0, false
}

// VectorSize returns the number of components (2, 3, or 4) if this is a vector, or 0 otherwise.
func (t TypeInfo) VectorSize() int {
	if t.ProgType != nil && t.ProgType.Size >= 2 && t.ProgType.Size <= 4 {
		return int(t.ProgType.Size)
	}
	switch t.Name {
	case "vec2<f32>", "vec2<bool>":
		return 2
	case "vec3<f32>", "vec3<bool>":
		return 3
	case "vec4<f32>", "vec4<bool>":
		return 4
	default:
		return 0
	}
}

// ArraySize returns the element count if this is an array, or 0 otherwise.
func (t TypeInfo) ArraySize() int {
	if t.ProgType != nil && t.ProgType.Size > 0 {
		return int(t.ProgType.Size)
	}
	if strings.HasPrefix(t.Name, "array<") {
		inner := strings.TrimPrefix(t.Name, "array<")
		inner = strings.TrimSuffix(inner, ">")
		if commaIdx := strings.LastIndex(inner, ","); commaIdx != -1 {
			sizeStr := strings.TrimSpace(inner[commaIdx+1:])
			var size int
			fmt.Sscanf(sizeStr, "%d", &size)
			return size
		}
	}
	return 0
}

// IsScalarOrVector returns true if this TypeInfo represents a scalar (f32, bool) or vector.
func (t TypeInfo) IsScalarOrVector() bool {
	return t.IsNumericScalar() || t.IsBoolScalar() || t.IsVector()
}

// String implements fmt.Stringer, returning the emitted WGSL type name.
func (t TypeInfo) String() string {
	return t.Name
}

// Equal returns true if t and other represent the same WGSL type and structure.
func (t TypeInfo) Equal(other TypeInfo) bool {
	if t.Name != other.Name || len(t.FieldNames) != len(other.FieldNames) || len(t.Types) != len(other.Types) {
		return false
	}
	for i, f := range t.FieldNames {
		if f != other.FieldNames[i] {
			return false
		}
	}
	for i, ty := range t.Types {
		if !ty.Equal(other.Types[i]) {
			return false
		}
	}
	return true
}

// Predefined flyweights for common types to eliminate heap allocations.
var (
	TypeF32      = TypeInfo{Name: "f32", ProgType: program.NewScalarType()}
	TypeBool     = TypeInfo{Name: "bool", ProgType: program.NewScalarType()}
	TypeVec2     = TypeInfo{Name: "vec2<f32>", ProgType: program.NewListType(2, program.NewScalarType()), ElemType: &TypeF32}
	TypeVec3     = TypeInfo{Name: "vec3<f32>", ProgType: program.NewListType(3, program.NewScalarType()), ElemType: &TypeF32}
	TypeVec4     = TypeInfo{Name: "vec4<f32>", ProgType: program.NewListType(4, program.NewScalarType()), ElemType: &TypeF32}
	TypeVec2Bool = TypeInfo{Name: "vec2<bool>", ProgType: program.NewListType(2, program.NewScalarType()), ElemType: &TypeBool}
	TypeVec3Bool = TypeInfo{Name: "vec3<bool>", ProgType: program.NewListType(3, program.NewScalarType()), ElemType: &TypeBool}
	TypeVec4Bool = TypeInfo{Name: "vec4<bool>", ProgType: program.NewListType(4, program.NewScalarType()), ElemType: &TypeBool}
	TypeMat2x2   = TypeInfo{
		Name:     "mat2x2<f32>",
		ProgType: program.NewMatrixType(2, program.NewListType(2, program.NewScalarType())),
		ElemType: &TypeVec2,
	}
	TypeMat2x3 = TypeInfo{
		Name:     "mat2x3<f32>",
		ProgType: program.NewMatrixType(2, program.NewListType(3, program.NewScalarType())),
		ElemType: &TypeVec3,
	}
	TypeMat2x4 = TypeInfo{
		Name:     "mat2x4<f32>",
		ProgType: program.NewMatrixType(2, program.NewListType(4, program.NewScalarType())),
		ElemType: &TypeVec4,
	}
	TypeMat3x2 = TypeInfo{
		Name:     "mat3x2<f32>",
		ProgType: program.NewMatrixType(3, program.NewListType(2, program.NewScalarType())),
		ElemType: &TypeVec2,
	}
	TypeMat3x3 = TypeInfo{
		Name:     "mat3x3<f32>",
		ProgType: program.NewMatrixType(3, program.NewListType(3, program.NewScalarType())),
		ElemType: &TypeVec3,
	}
	TypeMat3x4 = TypeInfo{
		Name:     "mat3x4<f32>",
		ProgType: program.NewMatrixType(3, program.NewListType(4, program.NewScalarType())),
		ElemType: &TypeVec4,
	}
	TypeMat4x2 = TypeInfo{
		Name:     "mat4x2<f32>",
		ProgType: program.NewMatrixType(4, program.NewListType(2, program.NewScalarType())),
		ElemType: &TypeVec2,
	}
	TypeMat4x3 = TypeInfo{
		Name:     "mat4x3<f32>",
		ProgType: program.NewMatrixType(4, program.NewListType(3, program.NewScalarType())),
		ElemType: &TypeVec3,
	}
	TypeMat4x4 = TypeInfo{
		Name:     "mat4x4<f32>",
		ProgType: program.NewMatrixType(4, program.NewListType(4, program.NewScalarType())),
		ElemType: &TypeVec4,
	}
)

// WGSLType directly inspects a *program.Type and returns its corresponding TypeInfo.
func WGSLType(t *program.Type, isBool bool) TypeInfo {
	if t == nil {
		if isBool {
			return TypeBool
		}
		return TypeF32
	}

	if isBool {
		if t.Kind == program.TypeKind_SCALAR || t.Size <= 1 {
			return TypeBool
		}
		switch t.Size {
		case 2:
			return TypeVec2Bool
		case 3:
			return TypeVec3Bool
		case 4:
			return TypeVec4Bool
		default:
			return TypeInfo{
				Name:     fmt.Sprintf("array<bool, %d>", t.Size),
				ProgType: t,
				ElemType: &TypeBool,
			}
		}
	}

	if t.Kind == program.TypeKind_MATRIX {
		cols := int(t.Size)
		rows := 3
		if t.ElemType != nil && t.ElemType.Size > 0 {
			rows = int(t.ElemType.Size)
		}

		if cols >= 2 && cols <= 4 && rows >= 2 && rows <= 4 {
			switch cols {
			case 2:
				switch rows {
				case 2:
					return TypeMat2x2
				case 3:
					return TypeMat2x3
				case 4:
					return TypeMat2x4
				}
			case 3:
				switch rows {
				case 2:
					return TypeMat3x2
				case 3:
					return TypeMat3x3
				case 4:
					return TypeMat3x4
				}
			case 4:
				switch rows {
				case 2:
					return TypeMat4x2
				case 3:
					return TypeMat4x3
				case 4:
					return TypeMat4x4
				}
			}
		}

		if rows >= 2 && rows <= 4 {
			elemVec := WGSLType(program.NewListType(int32(rows), program.NewScalarType()), false)
			return TypeInfo{
				Name:     fmt.Sprintf("array<%s, %d>", elemVec.Name, cols),
				ProgType: t,
				ElemType: &elemVec,
			}
		}

		elemArr := WGSLType(program.NewListType(int32(rows), program.NewScalarType()), false)
		return TypeInfo{
			Name:     fmt.Sprintf("array<%s, %d>", elemArr.Name, cols),
			ProgType: t,
			ElemType: &elemArr,
		}
	}

	if t.Kind == program.TypeKind_LIST {
		if t.ElemType != nil && (t.ElemType.IsList() || t.ElemType.IsMatrix()) {
			elemType := WGSLType(t.ElemType, false)
			return TypeInfo{
				Name:     fmt.Sprintf("array<%s, %d>", elemType.Name, t.Size),
				ProgType: t,
				ElemType: &elemType,
			}
		}
		switch t.Size {
		case 0, 1:
			return TypeF32
		case 2:
			return TypeVec2
		case 3:
			return TypeVec3
		case 4:
			return TypeVec4
		default:
			return TypeInfo{
				Name:     fmt.Sprintf("array<f32, %d>", t.Size),
				ProgType: t,
				ElemType: &TypeF32,
			}
		}
	}

	return TypeF32
}

// TypeFor resolves the TypeInfo for an expression sequence, synthesizing and caching
// multi-value return structs when len(expressions) > 1.
func TypeFor(
	expressions []build.Expression,
	structTypeCache map[string]TypeInfo,
) TypeInfo {
	if len(expressions) == 0 {
		return TypeF32
	}
	if len(expressions) == 1 {
		if w, ok := expressions[0].(Expression); ok {
			return w.WGSLType()
		}
		return WGSLType(expressions[0].Type(), false)
	}

	fieldNames := make([]string, len(expressions))
	fieldTypes := make([]TypeInfo, len(expressions))
	typeNames := make([]string, len(expressions))
	for i, expr := range expressions {
		fieldNames[i] = fmt.Sprintf("v_%d", i+1)
		var ft TypeInfo
		if w, ok := expr.(Expression); ok {
			ft = w.WGSLType()
		} else {
			ft = WGSLType(expr.Type(), false)
		}
		fieldTypes[i] = ft
		typeNames[i] = ft.Name
	}

	typeKey := "struct," + strings.Join(typeNames, ",")
	if cached, ok := structTypeCache[typeKey]; ok {
		return cached
	}

	structInfo := TypeInfo{
		Name:       newStructName(structTypeCache),
		ProgType:   program.NewListType(int32(len(fieldTypes)), program.NewScalarType()),
		FieldNames: fieldNames,
		Types:      fieldTypes,
	}
	structTypeCache[typeKey] = structInfo
	return structInfo
}

func newStructName(structTypeCache map[string]TypeInfo) string {
	if len(structTypeCache) == 0 {
		return "gsl_struct"
	}
	index := len(structTypeCache) + 1
outer:
	for {
		structName := fmt.Sprintf("gsl_struct_%d", index)
		for _, s := range structTypeCache {
			if s.Name == structName {
				index++
				continue outer
			}
		}
		return structName
	}
}
