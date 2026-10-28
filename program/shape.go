package program

import "fmt"

// IsHardwareMatrix returns true if t is a matrix with <= 4 rows.
func (t *Type) IsHardwareMatrix() bool {
	return t != nil && t.IsMatrix() && t.ElemType != nil && t.ElemType.Size <= 4
}

// IsHardwareVector returns true if t is a list of length 2..4 with scalar elements.
func (t *Type) IsHardwareVector() bool {
	return t != nil && t.IsList() && t.Size >= 2 && t.Size <= 4 && (t.ElemType == nil || t.ElemType.IsScalar())
}

// MatrixDims returns (rows, cols) for a matrix type.
func MatrixDims(t *Type) (rows, cols int) {
	if t == nil {
		return 0, 0
	}
	cols = int(t.Size)
	if t.ElemType != nil {
		rows = int(t.ElemType.Size)
	}
	return rows, cols
}

// DeriveBinaryOpType derives the resulting Type and validates linear algebra
// shape conformability for a binary operation between left and right.
// opIdx is the 1-based operand index of the right operand in the expression.
func DeriveBinaryOpType(op Operation, left, right *Type, opIdx int) (*Type, error) {
	if left == nil || right == nil {
		return nil, nil
	}

	lMat := left.IsHardwareMatrix()
	rMat := right.IsHardwareMatrix()
	lVec := left.IsHardwareVector()
	rVec := right.IsHardwareVector()
	lScal := left.IsScalar()
	rScal := right.IsScalar()

	switch op {
	case Operation_MUL:
		if lScal && rScal {
			return NewScalarType(), nil
		}
		if lScal && rMat {
			return right, nil
		}
		if lMat && rScal {
			return left, nil
		}
		if lMat && rMat {
			r1, c1 := MatrixDims(left)
			r2, c2 := MatrixDims(right)
			if c1 != r2 {
				return nil, fmt.Errorf("matrix multiplication dimension mismatch at operand %d: %dx%d vs %dx%d", opIdx, r1, c1, r2, c2)
			}
			return NewMatrixType(int32(c2), NewListType(int32(r1), NewScalarType())), nil
		}
		if lMat && rVec {
			r, c := MatrixDims(left)
			vLen := int(right.Size)
			if c != vLen {
				return nil, fmt.Errorf("matrix-vector dimension mismatch at operand %d: matrix %dx%d, vector len %d", opIdx, r, c, vLen)
			}
			return NewListType(int32(r), NewScalarType()), nil
		}
		if lVec && rMat {
			return nil, fmt.Errorf("cannot multiply vector by matrix at operand %d: write matrix * vector instead", opIdx)
		}
		if lScal && rVec {
			return right, nil
		}
		if lVec && rScal {
			return left, nil
		}
		if lVec && rVec {
			l1 := int(left.Size)
			l2 := int(right.Size)
			if l1 != l2 {
				return nil, fmt.Errorf("vector multiplication dimension mismatch at operand %d: len %d vs len %d", opIdx, l1, l2)
			}
			return left, nil
		}
		return nil, fmt.Errorf("unsupported operand types for '*' at operand %d: %s and %s", opIdx, left.TypeString(), right.TypeString())

	case Operation_ADD:
		if lScal && rScal {
			return NewScalarType(), nil
		}
		if lMat && rMat {
			r1, c1 := MatrixDims(left)
			r2, c2 := MatrixDims(right)
			if r1 != r2 || c1 != c2 {
				return nil, fmt.Errorf("matrix addition dimension mismatch at operand %d: %dx%d vs %dx%d", opIdx, r1, c1, r2, c2)
			}
			return left, nil
		}
		if (lMat && rScal) || (lScal && rMat) {
			return nil, fmt.Errorf("cannot add matrix and scalar at operand %d: matrix addition requires matrices", opIdx)
		}
		if (lMat && rVec) || (lVec && rMat) {
			return nil, fmt.Errorf("cannot add matrix and vector at operand %d", opIdx)
		}
		if lVec && rVec {
			l1 := int(left.Size)
			l2 := int(right.Size)
			if l1 != l2 {
				return nil, fmt.Errorf("vector addition dimension mismatch at operand %d: len %d vs len %d", opIdx, l1, l2)
			}
			return left, nil
		}
		if lVec && rScal {
			return left, nil
		}
		if lScal && rVec {
			return right, nil
		}
		return nil, fmt.Errorf("unsupported operand types for '+' at operand %d: %s and %s", opIdx, left.TypeString(), right.TypeString())

	case Operation_SUB:
		if lScal && rScal {
			return NewScalarType(), nil
		}
		if lMat && rMat {
			r1, c1 := MatrixDims(left)
			r2, c2 := MatrixDims(right)
			if r1 != r2 || c1 != c2 {
				return nil, fmt.Errorf("matrix subtraction dimension mismatch at operand %d: %dx%d vs %dx%d", opIdx, r1, c1, r2, c2)
			}
			return left, nil
		}
		if (lMat && rScal) || (lScal && rMat) {
			return nil, fmt.Errorf("cannot subtract matrix and scalar at operand %d", opIdx)
		}
		if (lMat && rVec) || (lVec && rMat) {
			return nil, fmt.Errorf("cannot subtract matrix and vector at operand %d", opIdx)
		}
		if lVec && rVec {
			l1 := int(left.Size)
			l2 := int(right.Size)
			if l1 != l2 {
				return nil, fmt.Errorf("vector subtraction dimension mismatch at operand %d: len %d vs len %d", opIdx, l1, l2)
			}
			return left, nil
		}
		if lVec && rScal {
			return left, nil
		}
		if lScal && rVec {
			return right, nil
		}
		return nil, fmt.Errorf("unsupported operand types for '-' at operand %d: %s and %s", opIdx, left.TypeString(), right.TypeString())

	case Operation_DIV:
		if lScal && rScal {
			return NewScalarType(), nil
		}
		if (lScal && rMat) || (lVec && rMat) {
			return nil, fmt.Errorf("division by matrix at operand %d is unsupported; use matrix.inverse", opIdx)
		}
		if lMat && rMat {
			return nil, fmt.Errorf("matrix division at operand %d is unsupported; multiply by matrix.inverse", opIdx)
		}
		if lMat && rScal {
			return left, nil
		}
		if lMat && rVec {
			return nil, fmt.Errorf("cannot divide matrix and vector at operand %d", opIdx)
		}
		if lVec && rVec {
			l1 := int(left.Size)
			l2 := int(right.Size)
			if l1 != l2 {
				return nil, fmt.Errorf("vector division dimension mismatch at operand %d: len %d vs len %d", opIdx, l1, l2)
			}
			return left, nil
		}
		if lVec && rScal {
			return left, nil
		}
		if lScal && rVec {
			return right, nil
		}
		return nil, fmt.Errorf("unsupported operand types for '/' at operand %d: %s and %s", opIdx, left.TypeString(), right.TypeString())
	}

	return nil, nil
}
