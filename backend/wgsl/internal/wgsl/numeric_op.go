package wgsl

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// NumericOp implements the [build.Expression] interface.
type NumericOp struct {
	Expression
	Symbol string
	Args   []Expression
}

func NewNumericOp(op *program.Operand, expressions []build.Expression, precedence build.Precedence, resultType *program.Type, symbol string) (Expression, error) {
	wgslExprs := make([]Expression, len(expressions))
	wgslExprs[0] = expressions[0].(Expression).ToNum().ParenIf(precedence)
	for i := 1; i < len(expressions); i++ {
		e := expressions[i].(Expression).ToNum()
		if e.Precedence() > precedence ||
			(e.Precedence() == precedence && (symbol == "-" || symbol == "/" || symbol == "%")) {
			e = Paren(e)
		}
		wgslExprs[i] = e
	}

	rType := WGSLType(resultType, false)
	if rType.Name == "" || (resultType == nil && len(wgslExprs) >= 1) {
		t0 := wgslExprs[0].WGSLType()
		if len(wgslExprs) == 1 {
			rType = t0
		} else if len(wgslExprs) == 2 {
			t1 := wgslExprs[1].WGSLType()
			switch symbol {
			case "*":
				switch {
				case t0.IsNativeMatrix() && t1.IsVector():
					_, rows, _ := t0.MatrixDimensions()
					rType = WGSLType(program.NewListType(int32(rows), program.NewScalarType()), false)
				case t0.IsVector() && t1.IsNativeMatrix():
					cols, _, _ := t1.MatrixDimensions()
					rType = WGSLType(program.NewListType(int32(cols), program.NewScalarType()), false)
				case t0.IsNativeMatrix() && t1.IsNativeMatrix():
					_, rows1, _ := t0.MatrixDimensions()
					cols2, _, _ := t1.MatrixDimensions()
					rType = WGSLType(program.NewMatrixType(int32(cols2), program.NewListType(int32(rows1), program.NewScalarType())), false)
				case t0.IsNativeMatrix():
					rType = t0
				case t1.IsNativeMatrix():
					rType = t1
				case t0.IsVector():
					rType = t0
				case t1.IsVector():
					rType = t1
				default:
					rType = TypeF32
				}
			case "/":
				rType = t0
			case "+", "-":
				switch {
				case t0.IsNativeMatrix():
					rType = t0
				case t0.IsVector():
					rType = t0
				case t1.IsVector():
					rType = t1
				default:
					rType = TypeF32
				}
			default:
				rType = t0
			}
		}
	}

	n := &NumericOp{
		Symbol: symbol,
		Args:   wgslExprs,
	}
	n.Expression = NewExpr(op, rType, precedence)
	n.Expression.SetOuter(n)
	return n, nil
}

func (n *NumericOp) Codegen(ctx CodegenContext) string {
	if len(n.Args) == 1 {
		return n.Symbol + n.Args[0].Codegen(ctx)
	}
	if n.Symbol == "/" && len(n.Args) == 2 && n.Args[0].WGSLType().IsNativeMatrix() {
		return fmt.Sprintf("%s * (1.0 / (%s))", n.Args[0].Codegen(ctx), n.Args[1].Codegen(ctx))
	}
	src := n.Args[0].Codegen(ctx)
	for i := 1; i < len(n.Args); i++ {
		src += " " + n.Symbol + " " + n.Args[i].Codegen(ctx)
	}
	return src
}

func (n *NumericOp) Reads() []string {
	return dedupReads(n.Args...)
}
