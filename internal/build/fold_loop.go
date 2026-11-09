package build

import (
	"gradientcontrol.com/omega/core/program"
)

// FoldLooper is a Delegate that can build a reduce as a loop: one copy of its
// callback, called once per element, rather than a copy per element. A
// shader's compile time grows with its code, so a scan over a neighbourhood
// of cells costs a fraction as a loop of what it does unrolled.
type FoldLooper interface {
	// BuildFoldLoop returns the fold of elements, constants all of one type,
	// from init. step returns one call of the callback on the accumulator,
	// element and index expressions it is given, which the delegate makes to
	// stand for the loop's. ok is false when the delegate would rather the
	// reduce were unrolled; it then builds nothing.
	BuildFoldLoop(op *program.Operand, init Expression, elements []ConstExpression,
		step func(acc, element, index Expression) (Expression, error)) (result Expression, ok bool, err error)
}

// constElements returns the elements of the list collOp writes out, each
// built as a constant, or nil if it is not a list written out or any element
// is not a constant.
func (b *Builder) constElements(collOp *program.Operand) ([]ConstExpression, error) {
	expr := collOp.GetExpression()
	if expr == nil || expr.Operation != program.Operation_COLLECT {
		return nil, nil
	}
	// Only numbers written out are built, so nothing else is built twice.
	for _, el := range expr.Operands {
		if !writtenNumbers(el) {
			return nil, nil
		}
	}
	consts := make([]ConstExpression, len(expr.Operands))
	for i, el := range expr.Operands {
		built, err := program.VisitOperand(el, b)
		if err != nil {
			return nil, err
		}
		c, ok := built.(ConstExpression)
		if !ok || c.ConstValues() == nil {
			return nil, nil
		}
		consts[i] = c
	}
	return consts, nil
}

// loopFold builds the reduce of the list collOp from init by the callback
// prefix.name as a loop, when the delegate can, the list is constants written
// out, and the callback is a function of the accumulator, the element and, if
// it takes one, the index. ok is false when the reduce is to be unrolled
// instead.
func (b *Builder) loopFold(op *program.Operand, init Expression, collOp *program.Operand, prefix, name string) (Expression, bool, error) {
	looper, ok := b.delegate.(FoldLooper)
	if !ok {
		return nil, false, nil
	}
	consts, err := b.constElements(collOp)
	if err != nil || consts == nil {
		return nil, false, err
	}
	fnCall := &program.FnCall{Reference: &program.Reference{Prefix: prefix, Name: name}}
	withIndex := true
	fnDecl, fn, enclosingScope, err := b.findFunction(fnCall, op, 3)
	if err != nil || fn == nil {
		withIndex = false
		fnDecl, fn, enclosingScope, err = b.findFunction(fnCall, op, 2)
		if err != nil || fn == nil {
			return nil, false, nil
		}
	}
	if len(fn.Returns) != 1 {
		return nil, false, nil
	}
	return looper.BuildFoldLoop(op, init, consts, func(acc, element, index Expression) (Expression, error) {
		args := []Expression{acc, element}
		if withIndex {
			args = append(args, index)
		}
		fnInfo, err := b.buildFunction(op, fnDecl, fn, enclosingScope, fnCall, args)
		if err != nil {
			return nil, err
		}
		return b.delegate.BuildFuncCall(op, fnCall, fnInfo, args)
	})
}

// writtenNumbers reports whether op is a number, a negated one, or a list of
// those, written out.
func writtenNumbers(op *program.Operand) bool {
	if op == nil {
		return false
	}
	if _, ok := op.T.(*program.Operand_Value); ok {
		return true
	}
	expr := op.GetExpression()
	if expr == nil || (expr.Operation != program.Operation_NEG && expr.Operation != program.Operation_COLLECT) {
		return false
	}
	for _, sub := range expr.Operands {
		if !writtenNumbers(sub) {
			return false
		}
	}
	return true
}
