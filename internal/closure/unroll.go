package closure

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/program"
)

// unrollFunctionFold rewrites, in place, a map or reduce whose COLLECTION is a
// list of functions into the calls that list denotes.
//
// The backends have no function pointers, so a list of functions is only ever
// a compile-time device: the elements have to be resolved here or not at all.
// lower.Transform already unrolls map and reduce, but it runs after this pass
// and after pruneHigherOrderTemplates has deleted the templates the unrolled
// calls would need — and it resolves ONE callee for the whole call, which is
// right for folding scalars and wrong for folding functions. So a fold over
// functions is unrolled here instead, ahead of specialization.
//
// What it produces is ordinary work for the rest of the pass: each emitted
// call passes one element as an argument to a statically named callback, which
// is exactly the shape specializeHigherOrderCall monomorphizes.
//
// Folds over scalars are left untouched for lower.Transform, so nothing that
// compiles today takes a different path.
func (t *transformer) unrollFunctionFold(file *program.PackageFile, enclosingFn *program.Function, op *program.Operand, expr *program.Expression, env map[string]*closureVal) (bool, error) {
	if expr == nil {
		return false, nil
	}
	cbOp, collOp, initialOp, ok := t.foldParts(file, enclosingFn, expr)
	if !ok {
		return false, nil
	}

	elements := t.foldElements(file, enclosingFn, collOp)
	if len(elements) == 0 {
		return false, nil
	}
	if !t.anyFunctionValue(file, enclosingFn, elements) {
		return false, nil
	}

	prefix, name, arity, ok := t.callbackArity(file, enclosingFn, cbOp)
	if !ok {
		return false, fmt.Errorf("cannot resolve the function folding this list of functions: its elements have to be applied by a callback this pass can see")
	}

	if expr.Operation == program.Operation_MAP {
		results := make([]*program.Operand, len(elements))
		for i, elem := range elements {
			args := []*program.Operand{elem}
			if arity == 2 {
				args = append(args, program.NewValue(float64(i), op.Start, op.End))
			}
			results[i] = program.NewFnCall(prefix, name, args, op.Start, op.End)
		}
		op.T = program.NewExpression(program.Operation_COLLECT, results, op.Start, op.End).T
		return true, nil
	}

	acc := initialOp
	start := 0
	if acc == nil {
		acc = elements[0]
		start = 1
	}
	for i := start; i < len(elements); i++ {
		args := []*program.Operand{acc, elements[i]}
		if arity == 3 {
			args = append(args, program.NewValue(float64(i), op.Start, op.End))
		}
		acc = program.NewFnCall(prefix, name, args, op.Start, op.End)
	}
	op.T = acc.T
	return true, nil
}

// foldParts splits a map or reduce expression into its callback, its
// collection, and its initial accumulator if it has one.
//
// This mirrors lowerReduce's disambiguation, including the legacy
// reduce(initial, fn, collection) order that text.left still uses. The
// multi-collection forms — map(fn, a, b, c) and reduce(fn, a, b, c) — are
// declined: a list of functions is one collection, and guessing at a shape
// this pass does not need is worse than leaving it to lower.
func (t *transformer) foldParts(file *program.PackageFile, enclosingFn *program.Function, expr *program.Expression) (cb, coll, initial *program.Operand, ok bool) {
	ops := expr.Operands
	switch expr.Operation {
	case program.Operation_MAP:
		if len(ops) == 2 {
			return ops[0], ops[1], nil, true
		}
	case program.Operation_REDUCE:
		switch len(ops) {
		case 2:
			return ops[0], ops[1], nil, true
		case 3:
			if t.isFunctionValue(ops[0], file, enclosingFn) {
				return ops[0], ops[1], ops[2], true
			}
			return ops[1], ops[2], ops[0], true
		}
	}
	return nil, nil, nil, false
}

// foldElements resolves the collection of a fold to its elements.
//
// A list written at a call site and bound into a specialized clone is found in
// listBindings; anything else has to be visible from the AST, which is what
// program.StaticElements decides. Returns nil when the collection cannot be
// resolved, which for a list of scalars simply means lower will handle it.
func (t *transformer) foldElements(file *program.PackageFile, enclosingFn *program.Function, collOp *program.Operand) []*program.Operand {
	if collOp == nil {
		return nil
	}
	if ref, ok := collOp.T.(*program.Operand_Reference); ok && ref.Reference != nil && ref.Reference.Prefix == "" {
		if byParam, ok := t.listBindings[enclosingFn]; ok {
			if els, ok := byParam[ref.Reference.Name]; ok {
				return cloneOperands(els)
			}
		}
	}
	var fnStack []*program.Function
	if enclosingFn != nil {
		fnStack = []*program.Function{enclosingFn}
	}
	return cloneOperands(program.StaticElements(t.prog, file, fnStack, collOp))
}

// cloneOperands copies the elements a collection was resolved to. They are
// spliced into a body and then rewritten in place by specialization, and the
// list they came from — a literal at a call site, or a named variable that may
// still be read elsewhere — must not be rewritten along with them.
func cloneOperands(ops []*program.Operand) []*program.Operand {
	if ops == nil {
		return nil
	}
	out := make([]*program.Operand, len(ops))
	for i, op := range ops {
		out[i] = proto.Clone(op).(*program.Operand)
	}
	return out
}

// functionListElements returns the elements of an argument that is a list
// holding at least one function — written inline at the call, or named by a
// variable this pass can resolve to one. Returns nil for anything else,
// including a list of plain values, which crosses a call as an ordinary
// argument and needs none of this.
func (t *transformer) functionListElements(file *program.PackageFile, enclosingFn *program.Function, arg *program.Operand) []*program.Operand {
	if arg == nil {
		return nil
	}
	switch arg.T.(type) {
	case *program.Operand_Expression, *program.Operand_Reference:
	default:
		return nil
	}
	els := t.foldElements(file, enclosingFn, arg)
	if len(els) == 0 || !t.anyFunctionValue(file, enclosingFn, els) {
		return nil
	}
	return els
}

// capturesAreLiteral reports whether every capture value is a closed literal —
// a number, or a collection of them. Those resolve in any file, so baking them
// into the definer's is safe. Anything else is an expression from the creating
// file's scope and only survives as an argument.
func capturesAreLiteral(captures []captureBinding) bool {
	for _, c := range captures {
		if !isLiteralOperand(c.value) {
			return false
		}
	}
	return true
}

func isLiteralOperand(op *program.Operand) bool {
	if op == nil {
		return true
	}
	switch v := op.T.(type) {
	case *program.Operand_Value:
		return true
	case *program.Operand_Expression:
		if v.Expression == nil || v.Expression.Operation != program.Operation_COLLECT {
			return false
		}
		for _, sub := range v.Expression.Operands {
			if !isLiteralOperand(sub) {
				return false
			}
		}
		return true
	}
	return false
}

func (t *transformer) anyFunctionValue(file *program.PackageFile, enclosingFn *program.Function, elements []*program.Operand) bool {
	for _, el := range elements {
		if t.isFunctionValue(el, file, enclosingFn) {
			return true
		}
	}
	return false
}

// callbackArity resolves the fold's callback to a name this pass can emit a
// call to, and the number of parameters it declares — 2 or 3 for reduce
// (accumulator and element, plus an optional index), 1 or 2 for map.
func (t *transformer) callbackArity(file *program.PackageFile, enclosingFn *program.Function, cbOp *program.Operand) (string, string, int, bool) {
	if cbOp == nil {
		return "", "", 0, false
	}
	ref, ok := cbOp.T.(*program.Operand_Reference)
	if !ok || ref.Reference == nil {
		return "", "", 0, false
	}
	prefix, name := ref.Reference.Prefix, ref.Reference.Name
	_, ft := t.lookupFnTypeInScope(enclosingFn, file, prefix, name)
	if ft == nil {
		return "", "", 0, false
	}
	for _, fg := range ft.Groups {
		if len(fg.Functions) > 0 {
			return prefix, name, len(fg.Functions[0].Parameters), true
		}
	}
	return "", "", 0, false
}
