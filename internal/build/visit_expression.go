package build

import (
	"fmt"
	"math"

	"gradientcontrol.com/omega/core/program"
)

func (b *Builder) VisitExpression(op *program.Operand, expr *program.Expression) (Expression, error) {
	// A folded expression is built from its values, shaped by its type. One a
	// later pass wrote in (an inlined or lowered record's arithmetic) can have
	// values but no type, and several values would then be read as one flat
	// vector, a record's rows lost; it is built from its operands instead,
	// which carry their shapes.
	if len(op.ConstValues) > 0 && !b.isParameter(op) && (op.Type != nil || len(op.ConstValues) == 1) {
		return b.buildConstExpressionWithType(op, op.ConstValues, op.Type)
	}
	return b.visitExpressionImpl(op, expr)
}

func (b *Builder) visitExpressionImpl(op *program.Operand, expr *program.Expression) (Expression, error) {
	switch expr.Operation {
	case program.Operation_ACCESS:
		return b.visitAccess(op, expr)
	case program.Operation_REMAP:
		return b.visitRemap(op, expr)
	case program.Operation_MAP:
		return b.visitMap(op, expr)
	case program.Operation_REDUCE:
		return b.visitReduce(op, expr)
	default:
		// fall through
	}

	operands := expr.Operands
	expressions := make([]Expression, len(operands))
	for i, operand := range operands {
		exp, err := program.VisitOperand(operand, b)
		if err != nil {
			return nil, err
		}
		expressions[i] = exp
	}

	switch expr.Operation {

	case program.Operation_ADD:
		return b.visitOperation(op, expr, expressions, operands, "+", PrecedenceAddition,
			b.calcConstValue(func(i0 int, v0 float64, i1 int, v1 float64) (float64, error) {
				return v0 + v1, nil
			}))
	case program.Operation_SUB:
		return b.visitOperation(op, expr, expressions, operands, "-", PrecedenceAddition,
			b.calcConstValue(func(i0 int, v0 float64, i1 int, v1 float64) (float64, error) {
				// subtract all values from the first value
				if i0 == 0 {
					return v0 - v1, nil
				}
				return v0 + v1, nil
			}))
	case program.Operation_MUL:
		return b.visitOperation(op, expr, expressions, operands, "*", PrecedenceMultiplication,
			b.calcConstValue(func(i0 int, v0 float64, i1 int, v1 float64) (float64, error) {
				return v0 * v1, nil
			}))
	case program.Operation_DIV:
		return b.visitOperation(op, expr, expressions, operands, "/", PrecedenceMultiplication,
			b.calcConstValue(func(i0 int, v0 float64, i1 int, v1 float64) (float64, error) {
				if v1 == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				// divide all values from the first value
				if i0 == 0 {
					return v0 / v1, nil
				}
				return v0 * v1, nil
			}))
	case program.Operation_MODULO:
		return b.visitOperation(op, expr, expressions, operands, "%", PrecedenceMultiplication,
			b.calcConstValue(func(i0 int, v0 float64, i1 int, v1 float64) (float64, error) {
				if v1 == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				// floored modulo which matches wgsl agent "gsl_floored" helper
				return v0 - v1*math.Floor(v0/v1), nil
			}))

	case program.Operation_MIN:
		return b.visitOpCall(op, expr, operands, expressions,
			b.calcConstValue(func(i0 int, v0 float64, i1 int, v1 float64) (float64, error) {
				return min(v0, v1), nil
			}))
	case program.Operation_MAX:
		return b.visitOpCall(op, expr, operands, expressions,
			b.calcConstValue(func(i0 int, v0 float64, i1 int, v1 float64) (float64, error) {
				return max(v0, v1), nil
			}))

	case program.Operation_AND:
		return b.visitOperation(op, expr, expressions, operands, "&&", PrecedenceLogicalAnd,
			b.calcConstValue(func(i0 int, v0 float64, i1 int, v1 float64) (float64, error) {
				if v0 > 0 && v1 > 0 {
					return 1, nil
				}
				return 0, nil
			}))
	case program.Operation_OR:
		return b.visitOperation(op, expr, expressions, operands, "||", PrecedenceLogicalOr,
			b.calcConstValue(func(i0 int, v0 float64, i1 int, v1 float64) (float64, error) {
				if v0 > 0 || v1 > 0 {
					return 1, nil
				}
				return 0, nil
			}))
	case program.Operation_NOT:
		return b.visitNot(op, expr, expressions, operands)

	case program.Operation_SQUARE:
		return b.visitUnaryCall(op, expr, operands, expressions,
			b.calcConstValueUnary(func(v float64) (float64, error) {
				return v * v, nil
			}))
	case program.Operation_SQRT:
		return b.visitUnaryCall(op, expr, operands, expressions,
			b.calcConstValueUnary(func(v float64) (float64, error) {
				return math.Sqrt(v), nil
			}))

	case program.Operation_FLOOR:
		return b.visitUnaryCall(op, expr, operands, expressions,
			b.calcConstValueUnary(func(v float64) (float64, error) {
				return math.Floor(v), nil
			}))
	case program.Operation_CEIL:
		return b.visitUnaryCall(op, expr, operands, expressions,
			b.calcConstValueUnary(func(v float64) (float64, error) {
				return math.Ceil(v), nil
			}))
	case program.Operation_ROUND:
		return b.visitUnaryCall(op, expr, operands, expressions,
			b.calcConstValueUnary(func(v float64) (float64, error) {
				return math.Round(v), nil
			}))

	case program.Operation_NEG:
		return b.visitNeg(op, expr, expressions, operands)
	case program.Operation_ABS:
		return b.visitUnaryCall(op, expr, operands, expressions,
			b.calcConstValueUnary(func(v float64) (float64, error) {
				return math.Abs(v), nil
			}))
	case program.Operation_EXP:
		return b.visitUnaryCall(op, expr, operands, expressions,
			b.calcConstValueUnary(func(v float64) (float64, error) {
				return math.Exp(v), nil
			}))
	case program.Operation_LN:
		return b.visitUnaryCall(op, expr, operands, expressions,
			b.calcConstValueUnary(func(v float64) (float64, error) {
				return math.Log(v), nil
			}))
	case program.Operation_POW:
		return b.visitOpCall(op, expr, operands, expressions,
			b.calcConstValue(func(i0 int, v0 float64, i1 int, v1 float64) (float64, error) {
				return math.Pow(v0, v1), nil
			}))

	case program.Operation_SIN:
		return b.visitUnaryCall(op, expr, operands, expressions,
			b.calcConstValueUnary(func(v float64) (float64, error) {
				return math.Sin(v), nil
			}))
	case program.Operation_COS:
		return b.visitUnaryCall(op, expr, operands, expressions,
			b.calcConstValueUnary(func(v float64) (float64, error) {
				return math.Cos(v), nil
			}))
	case program.Operation_TAN:
		return b.visitUnaryCall(op, expr, operands, expressions,
			b.calcConstValueUnary(func(v float64) (float64, error) {
				return math.Tan(v), nil
			}))
	case program.Operation_ASIN:
		return b.visitUnaryCall(op, expr, operands, expressions,
			b.calcConstValueUnary(func(v float64) (float64, error) {
				return math.Asin(v), nil
			}))
	case program.Operation_ACOS:
		return b.visitUnaryCall(op, expr, operands, expressions,
			b.calcConstValueUnary(func(v float64) (float64, error) {
				return math.Acos(v), nil
			}))
	case program.Operation_ATAN:
		if len(operands) != 1 {
			return nil, ErrLoc(b.delegate.GetScope(), op.Start, fmt.Errorf("atan expects 1 argument, got %d", len(operands)))
		}
		return b.visitUnaryCall(op, expr, operands, expressions,
			b.calcConstValueUnary(func(v float64) (float64, error) {
				return math.Atan(v), nil
			}))
	case program.Operation_ATAN2:
		if len(operands) != 2 {
			return nil, ErrLoc(b.delegate.GetScope(), op.Start, fmt.Errorf("atan2 expects 2 arguments, got %d", len(operands)))
		}
		return b.visitOpCall(op, expr, operands, expressions,
			b.calcConstValue(func(i0 int, v0 float64, i1 int, v1 float64) (float64, error) {
				return math.Atan2(v0, v1), nil
			}))

	case program.Operation_GT:
		return b.visitComparison(op, expr, ">", expressions, operands, PrecedenceComparison,
			func(v0, v1 float64) float64 {
				if v0 > v1 {
					return 1.0
				}
				return 0.0
			})
	case program.Operation_GE:
		return b.visitComparison(op, expr, ">=", expressions, operands, PrecedenceComparison,
			func(v0, v1 float64) float64 {
				if v0 >= v1 {
					return 1.0
				}
				return 0.0
			})
	case program.Operation_LE:
		return b.visitComparison(op, expr, "<=", expressions, operands, PrecedenceComparison,
			func(v0, v1 float64) float64 {
				if v0 <= v1 {
					return 1.0
				}
				return 0.0
			})
	case program.Operation_LT:
		return b.visitComparison(op, expr, "<", expressions, operands, PrecedenceComparison,
			func(v0, v1 float64) float64 {
				if v0 < v1 {
					return 1.0
				}
				return 0.0
			})
	case program.Operation_EQ:
		return b.visitComparison(op, expr, "==", expressions, operands, PrecedenceEquality,
			func(v0, v1 float64) float64 {
				if v0 == v1 {
					return 1.0
				}
				return 0.0
			})
	case program.Operation_NE:
		return b.visitComparison(op, expr, "!=", expressions, operands, PrecedenceEquality,
			func(v0, v1 float64) float64 {
				if v0 != v1 {
					return 1.0
				}
				return 0.0
			})

	case program.Operation_COMPARE:
		return b.visitCompare(op, expr, operands, expressions)
	case program.Operation_COLLECT:
		return b.visitCollect(op, expr, operands, expressions)
	case program.Operation_SELECT:
		return b.visitSelect(op, expr, operands, expressions)
	case program.Operation_INDEX:
		return b.visitIndex(op, expr, operands, expressions)

	case program.Operation_DERIV:
		// The autodiff pass expands DERIV in top-level expressions before the
		// build walk. Reaching one here means it was written inside a function
		// body, which is not supported (differentiating a parameter needs
		// call-site context).
		err := fmt.Errorf("the .dtime/.dp derivative operators are only supported in top-level expressions, not inside function bodies")
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)

	case program.Operation_SECONDCLOSEST:
		// The self-computing-property pass expands .second in top-level
		// expressions (and, via companions, in differentiated function bodies)
		// before the build walk. Reaching one here means it was written
		// somewhere the pass cannot see through.
		err := fmt.Errorf("the .second operator is only supported in top-level expressions, not inside function bodies")
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)

	case program.Operation_BOUNDINGBOX:
		// Expanded in top-level expressions like .second, and for the same
		// reason not inside function bodies.
		err := fmt.Errorf("the .box operator is only supported in top-level expressions, not inside function bodies")
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)

	case program.Operation_SURFACEU, program.Operation_SURFACEV:
		// Expanded in top-level expressions like .box. A material reads the
		// frame of the shape it is on, so the read belongs at the call site,
		// where the shape is named: bar.material = turned(..., bar.u, ...).
		err := fmt.Errorf("the .u and .v operators are only supported in top-level expressions, not inside function bodies; read the frame at the call site and pass it in")
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)

	default:
		err := fmt.Errorf("unhandled operator: %s", expr.Operation)
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
	}
}

func (b *Builder) visitMap(op *program.Operand, expr *program.Expression) (Expression, error) {
	if len(expr.Operands) < 2 {
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, fmt.Errorf("map expects at least 2 operands"))
	}
	fnOp := expr.Operands[0]
	var prefix, name string
	if ref := fnOp.GetReference(); ref != nil {
		prefix = ref.Prefix
		name = ref.Name
	}
	var elements []Expression
	if len(expr.Operands) > 2 {
		for _, arg := range expr.Operands[1:] {
			el, err := program.VisitOperand(arg, b)
			if err != nil {
				return nil, err
			}
			elements = append(elements, el)
		}
	} else {
		var err error
		elements, err = b.multiValue(expr.Operands[1])
		if err != nil {
			return nil, err
		}
	}
	if len(elements) == 0 {
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, fmt.Errorf("expected at least one element in map collection"))
	}

	results := make([]Expression, len(elements))
	for i, elem := range elements {
		var args []Expression
		fnDecl, fn, enclosingScope, err := b.findFunction(&program.FnCall{Reference: &program.Reference{Prefix: prefix, Name: name}}, op, 2)
		if err == nil && fn != nil {
			idxExpr, err := b.delegate.BuildValue(op, float64(i))
			if err != nil {
				return nil, err
			}
			args = []Expression{elem, idxExpr}
		} else {
			fnDecl, fn, enclosingScope, err = b.findFunction(&program.FnCall{Reference: &program.Reference{Prefix: prefix, Name: name}}, op, 1)
			if err != nil {
				return nil, err
			}
			args = []Expression{elem}
		}
		callExpr, err := b.buildFunctionCallExpr(op, fnDecl, fn, enclosingScope, prefix, name, args)
		if err != nil {
			return nil, err
		}
		results[i] = callExpr
	}

	ops := make([]*program.Operand, len(results))
	for i, r := range results {
		ops[i] = r.Operand()
	}
	return b.delegate.BuildCollect(op, expr, ops, results)
}

func (b *Builder) visitReduce(op *program.Operand, expr *program.Expression) (Expression, error) {
	if len(expr.Operands) < 2 {
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, fmt.Errorf("reduce expects at least 2 operands"))
	}

	var fnOp *program.Operand
	var initialOp *program.Operand
	var collOps []*program.Operand

	if len(expr.Operands) == 2 {
		fnOp = expr.Operands[0]
		collOps = expr.Operands[1:]
	} else if len(expr.Operands) == 3 {
		if b.isFnOperand(expr.Operands[0]) {
			fnOp = expr.Operands[0]
			collOps = []*program.Operand{expr.Operands[1]}
			initialOp = expr.Operands[2]
		} else {
			initialOp = expr.Operands[0]
			fnOp = expr.Operands[1]
			collOps = []*program.Operand{expr.Operands[2]}
		}
	} else {
		if b.isFnOperand(expr.Operands[0]) {
			fnOp = expr.Operands[0]
			collOps = expr.Operands[1:]
		} else {
			initialOp = expr.Operands[0]
			fnOp = expr.Operands[1]
			collOps = expr.Operands[2:]
		}
	}

	var prefix, name string
	if ref := fnOp.GetReference(); ref != nil {
		prefix = ref.Prefix
		name = ref.Name
	}

	var elements []Expression
	for _, c := range collOps {
		els, err := b.multiValue(c)
		if err != nil {
			return nil, err
		}
		elements = append(elements, els...)
	}
	if len(elements) == 0 {
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, fmt.Errorf("reduce expects at least one element in collection"))
	}

	var acc Expression
	startIdx := 0
	if initialOp != nil {
		var err error
		acc, err = program.VisitOperand(initialOp, b)
		if err != nil {
			return nil, err
		}
		startIdx = 0
		if len(collOps) == 1 {
			looped, ok, err := b.loopFold(op, acc, collOps[0], prefix, name)
			if err != nil {
				return nil, err
			}
			if ok {
				return looped, nil
			}
		}
	} else {
		acc = elements[0]
		startIdx = 1
	}

	for i := startIdx; i < len(elements); i++ {
		var args []Expression
		fnDecl, fn, enclosingScope, err := b.findFunction(&program.FnCall{Reference: &program.Reference{Prefix: prefix, Name: name}}, op, 3)
		if err == nil && fn != nil {
			idxExpr, err := b.delegate.BuildValue(op, float64(i))
			if err != nil {
				return nil, err
			}
			args = []Expression{acc, elements[i], idxExpr}
		} else {
			fnDecl, fn, enclosingScope, err = b.findFunction(&program.FnCall{Reference: &program.Reference{Prefix: prefix, Name: name}}, op, 2)
			if err != nil {
				return nil, err
			}
			args = []Expression{acc, elements[i]}
		}
		acc, err = b.buildFunctionCallExpr(op, fnDecl, fn, enclosingScope, prefix, name, args)
		if err != nil {
			return nil, err
		}
		if i < len(elements)-1 && fn != nil && len(fn.Parameters) > 0 && countReferences(fn, fn.Parameters[0].Name) > 1 {
			acc, err = b.bindFoldStep(op, acc)
			if err != nil {
				return nil, err
			}
		}
	}

	return acc, nil
}

// bindFoldStep binds one step of a reduce, acc, to a local of its own and
// returns a reference to it, so a next step that uses its accumulator more
// than once refers to it rather than copying its expression into each use:
// copied, a fold whose callback uses its accumulator twice doubles at every
// element. Constants, and folds outside a function, are left as they are.
func (b *Builder) bindFoldStep(op *program.Operand, acc Expression) (Expression, error) {
	if _, ok := acc.(ConstExpression); ok {
		return acc, nil
	}
	scope := b.delegate.GetScope()
	if scope == nil || scope.IsFileScope() {
		return acc, nil
	}
	name := scope.NewVarName("gsl_fold")
	assign := &program.Assignment{Start: op.Start, End: op.End, Variables: []*program.Variable{{Name: name}}, Expression: op}
	varInfo, err := b.delegate.BuildSingleAssignment(assign, op, name, acc)
	if err != nil {
		return nil, err
	}
	if err := scope.AddLocalVar(varInfo); err != nil {
		return nil, err
	}
	return b.delegate.BuildVarRef(op, varInfo)
}

// countReferences counts the references to name in fn's assignments and
// returns, not in the functions it declares.
func countReferences(fn *program.Function, name string) int {
	count := 0
	for _, a := range fn.Assignments {
		count += countReferencesIn(a.Expression, name)
	}
	for _, r := range fn.Returns {
		count += countReferencesIn(r, name)
	}
	return count
}

func countReferencesIn(op *program.Operand, name string) int {
	if op == nil {
		return 0
	}
	count := 0
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		if t.Reference.Prefix == "" && t.Reference.Name == name {
			count++
		}
	case *program.Operand_PropRef:
		if t.PropRef != nil && t.PropRef.Reference != nil && t.PropRef.Reference.Prefix == "" && t.PropRef.Reference.Name == name {
			count++
		}
	case *program.Operand_Expression:
		for _, child := range t.Expression.Operands {
			count += countReferencesIn(child, name)
		}
	case *program.Operand_FnCall:
		for _, arg := range t.FnCall.Arguments {
			count += countReferencesIn(arg, name)
		}
	}
	return count
}

func (b *Builder) buildFunctionCallExpr(op *program.Operand, fnDecl *program.FnType, fn *program.Function, enclosingScope *Scope, prefix, name string, args []Expression) (Expression, error) {
	if isInlinableFunction(fn) {
		return b.inlineFuncCall(op, fnDecl, fn, enclosingScope, args)
	}
	fnCall := &program.FnCall{Reference: &program.Reference{Prefix: prefix, Name: name}}
	fnInfo, err := b.buildFunction(op, fnDecl, fn, enclosingScope, fnCall, args)
	if err != nil {
		return nil, err
	}
	return b.delegate.BuildFuncCall(op, fnCall, fnInfo, args)
}

func (b *Builder) multiValue(op *program.Operand) ([]Expression, error) {
	if op == nil {
		return nil, nil
	}
	expr, err := program.VisitOperand(op, b)
	if err != nil {
		return nil, err
	}

	if col, ok := expr.(CollectExpression); ok && len(col.Elements()) > 0 {
		return col.Elements(), nil
	}

	typ := expr.Type()
	if typ == nil && op != nil {
		typ = op.Type
	}

	if typ.IsList() || typ.IsMatrix() {
		isList := typ.IsMatrix()
		vecLen := int(typ.Size)

		if constExpr, ok := expr.(ConstExpression); ok && !isList && constExpr.ConstValues() != nil {
			constValues := constExpr.ConstValues()
			elements := make([]Expression, vecLen)
			for i := range elements {
				val := constValues[min(i, len(constValues)-1)]
				element, err := b.delegate.BuildValue(op, val)
				if err != nil {
					return nil, err
				}
				elements[i] = element
			}
			return elements, nil
		}

		elements := make([]Expression, vecLen)
		for i := range elements {
			element, err := b.delegate.BuildIndexAccess(op, op.GetExpression(), op, op, expr, i)
			if err != nil {
				return nil, err
			}
			elements[i] = element
		}
		return elements, nil
	}

	return []Expression{expr}, nil
}

func (b *Builder) isFnOperand(op *program.Operand) bool {
	if op == nil {
		return false
	}
	if op.Type != nil && op.Type.FnPath != "" {
		return true
	}
	if ref := op.GetReference(); ref != nil {
		scope := b.delegate.GetScope()
		if scope != nil && scope.GetVar(ref.Name) != nil {
			return true
		}
	}
	return false
}
