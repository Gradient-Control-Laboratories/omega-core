package build

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

func (b *Builder) visitRemap(op *program.Operand, expr *program.Expression) (Expression, error) {

	// parser should ensure that the expression has 2 operands

	pOp := expr.Operands[1]
	pExpr, err := program.VisitOperand(pOp, b)
	if err != nil {
		return nil, ErrLoc(b.delegate.GetScope(), pOp.Start, err)
	}

	pType := pExpr.Type()
	if pType == nil && pOp != nil {
		pType = pOp.Type
	}
	if pType == nil || pType.Size != 3 {
		err := fmt.Errorf("remap 2nd argument must be or return a vec3: %s", pOp.String())
		return nil, ErrLoc(b.delegate.GetScope(), pOp.Start, err)
	}

	targetOp := expr.Operands[0]

	// Build the expression

	prevInRemap := b.inRemap
	b.inRemap = true
	targetExpr, err := program.VisitOperand(targetOp, b)
	b.inRemap = prevInRemap
	if err != nil {
		return nil, ErrLoc(b.delegate.GetScope(), targetOp.Start, err)
	}

	remapExpr, err := b.delegate.Remap(op, targetOp, pExpr, targetExpr)
	if err != nil {
		return nil, ErrLoc(b.delegate.GetScope(), targetOp.Start, err)
	}

	return remapExpr, nil
}
