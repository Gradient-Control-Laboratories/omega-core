package meshplan

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/program"
)

// MeshDecl describes an evaluated mesh(...) asset declaration.
type MeshDecl struct {
	// Index identifies the mesh asset by its zero-based session import order.
	Index int `json:"index"`

	// Matrix carries the 4x4 homogeneous placement transform (16 float64 values,
	// column-major order). An empty slice represents identity placement.
	Matrix []float64 `json:"matrix,omitempty"`

	// ShapeIndex is the 1-based shape this mesh belongs to, matching the top-level
	// return slot. Zero if not associated with a specific return shape.
	ShapeIndex int `json:"shapeIndex,omitempty"`
}

// ParseMeshArgs extracts the asset index and 4x4 homogeneous placement matrix
// from a GSL `mesh(...)` function call.
//
// Supported call signatures include:
//   - mesh(index): Identity placement.
//   - mesh(index, matrix): 16-element column-major 4x4 transform matrix.
//   - mesh(index, pos): Translation by 3-element vector [x, y, z].
//   - mesh(index, pos, rot): Rotation matrix with its translation column set to pos.
//
// Dynamic or symbolic placement expressions and referenced constants (such as `pi`
// or imported `matrix.*` routines) are evaluated against module declarations and
// imported packages via `eval.ConstPkg`. Validates that the index argument resolves
// to a non-negative integer slot.
func ParseMeshArgs(
	fc *program.FnCall,
	pkg *program.Package,
	getImport program.GetImportFunc,
) (MeshDecl, error) {
	var decl MeshDecl
	if fc == nil {
		return decl, fmt.Errorf("mesh() fncall is nil")
	}
	args := fc.Arguments

	// Arg 0: mesh slot.
	if len(args) > 0 {
		if v, ok := args[0].T.(*program.Operand_Value); ok {
			decl.Index = int(v.Value)
		} else {
			idx := args[0].ConstValues
			if len(idx) == 0 {
				var err error
				idx, err = eval.ConstPkg(args[0], pkg, getImport)
				if err != nil {
					return decl, fmt.Errorf("mesh() first argument: %w", err)
				}
			}
			if len(idx) != 1 {
				return decl, fmt.Errorf("mesh() first argument must fold to a non-negative integer slot: got %d values", len(idx))
			}
			decl.Index = int(idx[0])
		}
	}

	if len(args) <= 1 {
		return decl, nil // mesh(index): identity placement
	}

	var v []float64
	if len(args[1].ConstValues) > 0 {
		v = args[1].ConstValues
	} else {
		var err error
		v, err = eval.ConstPkg(args[1], pkg, getImport)
		if err != nil {
			return decl, fmt.Errorf("mesh() placement: %w", err)
		}
	}

	switch {
	case len(v) == 16: // mesh(index, M)
		if len(args) > 2 {
			return decl, fmt.Errorf("mesh(index, matrix) takes no further arguments")
		}
		decl.Matrix = append([]float64(nil), v...)
	case len(v) == 3 && len(args) == 2: // mesh(index, pos): translation only
		decl.Matrix = []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, v[0], v[1], v[2], 1}
	case len(v) == 3 && len(args) == 3: // mesh(index, pos, R): rotation matrix + translation column
		var r []float64
		if len(args[2].ConstValues) > 0 {
			r = args[2].ConstValues
		} else {
			var err error
			r, err = eval.ConstPkg(args[2], pkg, getImport)
			if err != nil {
				return decl, fmt.Errorf("mesh() rotation matrix: %w", err)
			}
		}
		if len(r) != 16 {
			return decl, fmt.Errorf("mesh(index, position, R): R must be a 4x4 matrix, got %d values", len(r))
		}
		m := append([]float64(nil), r...)
		m[12], m[13], m[14], m[15] = v[0], v[1], v[2], 1
		decl.Matrix = m
	default:
		return decl, fmt.Errorf("mesh() placement must be a 4x4 matrix, a position [x,y,z], "+
			"or position + rotation matrix; got %d values", len(v))
	}
	return decl, nil
}
