package wgsl

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// Collect implements the [build.Expression] interface.
type Collect struct {
	Expression
	Args        []Expression
	ResultType  TypeInfo
	constValues []float64
	// elements are the expressions as they were handed in, before ToNum, so
	// an element that is a constant can still be read as one.
	elements []build.Expression
}

var _ build.ConstExpression = (*Collect)(nil)

func NewCollect(op *program.Operand, expressions []build.Expression, constValues []float64) (build.ConstExpression, error) {
	args := make([]Expression, len(expressions))
	for i, expr := range expressions {
		args[i] = expr.(Expression).ToNum()
	}

	var resultType TypeInfo
	if len(args) > 0 {
		firstType := args[0].WGSLType()
		if firstType.IsVector() || firstType.IsArray() || firstType.IsMatrix() {
			// Collection elements gather into a list of that element type, whatever
			// its width: a row of 2, 3 or 4 is a vecN, a wider one is an
			// array<f32, N>, and a list of rows is itself a legal element. Matching
			// only vecNumType here collapsed a table of wider rows to a flat
			// array<f32, rowCount> — the element type was lost, and reading a row
			// back out of it typed as num, so indexing the row failed with "target
			// must be a collection". The build layer ensures the elements are
			// uniform.
			var progType *program.Type
			if op != nil && op.Type != nil && (op.Type.IsMatrix() || (op.Type.IsList() && op.Type.ElemType != nil)) && int(op.Type.Size) == len(expressions) {
				if op.Type.IsMatrix() {
					progType = op.Type
				} else {
					progType = program.NewMatrixType(op.Type.Size, op.Type.ElemType)
				}
			} else {
				elemPType := firstType.ProgType
				if elemPType == nil {
					if firstType.IsVector() {
						elemPType = program.NewListType(int32(firstType.VectorSize()), program.NewScalarType())
					} else {
						elemPType = program.NewScalarType()
					}
				}
				progType = program.NewMatrixType(int32(len(expressions)), elemPType)
			}
			resultType = WGSLType(progType, false)
		} else {
			switch len(expressions) {
			case 2:
				resultType = TypeVec2
			case 3:
				resultType = TypeVec3
			case 4:
				resultType = TypeVec4
			default:
				name := fmt.Sprintf("array<f32, %d>", len(expressions))
				var progType *program.Type
				if op != nil && op.Type != nil && op.Type.IsList() && int(op.Type.Size) == len(expressions) && (op.Type.ElemType == nil || op.Type.ElemType.IsScalar()) {
					progType = op.Type
				} else {
					progType = program.NewListType(int32(len(expressions)), program.NewScalarType())
				}
				resultType = TypeInfo{
					Name:     name,
					ProgType: progType,
					ElemType: &TypeF32,
				}
			}
		}
	} else if op != nil && op.Type != nil {
		resultType = WGSLType(op.Type, false)
	}

	c := &Collect{
		Args:        args,
		ResultType:  resultType,
		constValues: constValues,
		elements:    expressions,
	}
	c.Expression = NewExprWithReads(op, resultType, build.PrecedenceHighest, dedupReads(args...)...)
	c.Expression.SetOuter(c)
	return c, nil
}

func (c *Collect) Codegen(ctx CodegenContext) string {
	if constValues := c.constValues; constValues != nil {
		src, _ := codegenConst(c.ResultType, constValues)
		return src
	}

	argSrcs := make([]string, 0, len(c.Args))
	for _, arg := range c.Args {
		argSrcs = append(argSrcs, arg.Codegen(ctx))
	}

	return fmt.Sprintf("%s(%s)", c.ResultType.Name, strings.Join(argSrcs, ", "))
}

func codegenConst(t TypeInfo, values []float64) (string, []float64) {
	if t.IsNativeMatrix() {
		cols, rows, _ := t.MatrixDimensions()
		colType := TypeVec3
		if t.ElemType != nil {
			colType = *t.ElemType
		} else {
			colType = WGSLType(program.NewListType(int32(rows), program.NewScalarType()), false)
		}
		args := make([]string, cols)
		for i := 0; i < cols; i++ {
			args[i], values = codegenConst(colType, values)
		}
		return fmt.Sprintf("%s(%s)", t.Name, strings.Join(args, ", ")), values
	}
	if t.IsVector() {
		n := t.VectorSize()
		args := make([]string, n)
		for i := 0; i < n; i++ {
			args[i], values = codegenConst(TypeF32, values)
		}
		return fmt.Sprintf("%s(%s)", t.Name, strings.Join(args, ", ")), values
	}
	if t.IsArray() {
		elemType := TypeF32
		if t.ElemType != nil {
			elemType = *t.ElemType
		}
		size := t.ArraySize()
		args := make([]string, size)
		for i := range args {
			args[i], values = codegenConst(elemType, values)
		}
		return fmt.Sprintf("%s(%s)", t.Name, strings.Join(args, ", ")), values
	}
	if t.IsBoolScalar() {
		if len(values) == 0 {
			return "false", nil
		}
		if values[0] != 0.0 {
			return "true", values[1:]
		}
		return "false", values[1:]
	}
	if len(values) == 0 {
		return "0.0", nil
	}
	return CodegenValue(values[0]), values[1:]
}

func (c *Collect) ConstValues() []float64 {
	return c.constValues
}

func (c *Collect) Elements() []build.Expression {
	res := make([]build.Expression, len(c.Args))
	for i, a := range c.Args {
		res[i] = a
	}
	return res
}

// Element returns the i-th element as it was handed to NewCollect, before its
// conversion to a number, so its concrete type (a constant, an indexed
// constant) is still visible.
func (c *Collect) Element(i int) build.Expression {
	return c.elements[i]
}

func (c *Collect) Reads() []string {
	return dedupReads(c.Args...)
}
