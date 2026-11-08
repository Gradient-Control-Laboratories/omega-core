package lower

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal"
	"gradientcontrol.com/omega/core/program"
)

// LoopFoldThreshold is the fewest constants a reduce folds for Options'
// LoopFolds to keep it for a loop: 8 keeps gradient noise's corners.
const LoopFoldThreshold = 8

// Options configures TransformWith.
type Options struct {
	// LoopFolds, when above zero, keeps each reduce in a function body that
	// folds a list of at least that many constants with an initial value, for
	// a backend that emits it as a loop rather than one copy of its callback
	// per element (the WGSL one, fold_loop.go). Any other reduce is unrolled
	// as without it. A backend that cannot loop unrolls a kept one itself.
	LoopFolds int
}

// Transform rewrites GSL AST nodes for CSG builtin reductions and matrix
// operator routing into explicit binary calls and library function calls.
func Transform(prog *program.Program) error {
	return TransformWith(prog, Options{})
}

// TransformWith is Transform with options.
func TransformWith(prog *program.Program, opts Options) error {
	if prog == nil {
		return nil
	}

	// Lower primary package
	if prog.Primary != nil {
		if err := transformPackage(prog, prog.Primary, opts); err != nil {
			return err
		}
	}

	// Lower dependent packages
	for _, dep := range prog.Dependencies {
		if dep != nil {
			if err := transformPackage(prog, dep, opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func transformPackage(prog *program.Program, pkg *program.Package, opts Options) error {
	for _, file := range pkg.Files {
		if file != nil {
			if err := walkFile(file, func(op *program.Operand, fnStack []*program.Function) (*program.Operand, error) {
				return lowerOperand(prog, pkg, file, fnStack, op, opts)
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func lowerOperand(prog *program.Program, pkg *program.Package, file *program.PackageFile, fnStack []*program.Function, op *program.Operand, opts Options) (*program.Operand, error) {
	if op == nil {
		return nil, nil
	}
	switch t := op.T.(type) {
	case *program.Operand_FnCall:
		fc := t.FnCall
		if fc.Reference != nil && fc.Reference.Prefix == "" {
			if builtin := isReducingBuiltin(fc.Reference.Name); builtin != nil {
				return reduceCSG(op, fc, builtin)
			}
		}
	case *program.Operand_Expression:
		expr := t.Expression
		switch expr.Operation {
		case program.Operation_MAP:
			return lowerMap(prog, pkg, file, fnStack, op, expr)

		case program.Operation_REDUCE:
			return lowerReduce(prog, pkg, file, fnStack, op, expr)

		case program.Operation_MUL, program.Operation_ADD, program.Operation_SUB, program.Operation_DIV:
			return lowerChainedOp(file, op, expr)

		case program.Operation_NEG:
			// Invariant: parser grammar and primitive validation guarantee exactly 1 operand.
			t0 := expr.Operands[0].Type
			if isHardwareMatrix(t0) {
				ensureImported(file, "matrix", op.Start, op.End)
				res := program.NewFnCall("matrix", "negate", expr.Operands, op.Start, op.End)
				res.Type = t0
				return res, nil
			}
			if isHardwareVector(t0) {
				ensureImported(file, "vector", op.Start, op.End)
				res := program.NewFnCall("vector", "negate", expr.Operands, op.Start, op.End)
				res.Type = t0
				return res, nil
			}
		}
	}

	return op, nil
}

type chainedOperand struct {
	op      *program.Operand
	origIdx int // 1-based index in the original expr.Operands slice
}

// lowerChainedOp lowers flat N-ary arithmetic expressions (such as A * B * C or 2 * 3 * v)
// containing hardware vector or matrix operands into a left-folded tree of binary operations.
//
// What it does:
//  1. Verifies if any operand is a hardware vector or matrix. If all operands are scalars,
//     it preserves the flat N-ary expression so downstream backends and constant folding can
//     process it without overhead.
//  2. For associative operations (* and +), groups any contiguous runs of >= 2 scalar operands
//     into nested scalar sub-expressions. For example, in `s1 * s2 * v1 * s3 * s4`, the scalar
//     runs `[s1, s2]` and `[s3, s4]` are grouped into `(s1 * s2)` and `(s3 * s4)`, yielding the
//     intermediate sequence `[(s1 * s2), v1, (s3 * s4)]`.
//  3. Sequentially left-folds the resulting sequence from left to right via lowerBinaryOp:
//     - `(s1 * s2) * v1` lowers to `vector.multiplyScalar(v1, (s1 * s2))`
//     - `acc * (s3 * s4)` lowers to `vector.multiplyScalar(acc, (s3 * s4))`
//     Original operand indices are tracked throughout the fold for accurate diagnostics.
//
// Why it does this:
//   - The GSL parser produces flat N-ary expressions for chained arithmetic, but hardware acceleration
//     sidecars (WGSL/GPU) only implement binary operations (e.g., matrix-matrix, matrix-vector).
//   - Matrix and vector operations are non-commutative, requiring strict left-to-right evaluation order
//     to produce correct intermediate types and dimensions at each step.
//   - Grouping contiguous scalars in associative operations avoids redundant vector/matrix promotions:
//     `(s1 * s2)` and `(s3 * s4)` remain scalar expressions, allowing constant folding to evaluate each
//     cluster to a single scalar constant before GPU code generation and preventing redundant
//     per-component vector scaling passes on the GPU (e.g., 2 vector scalings instead of 3).
func lowerChainedOp(file *program.PackageFile, op *program.Operand, expr *program.Expression) (*program.Operand, error) {
	// Invariant: parser syntax rules guarantee at least 2 operands for arithmetic operations.
	// Upstream type analysis (types.Analyze) runs prior to lowering, guaranteeing
	// that all operand types in valid ASTs are populated and non-nil.
	anyMatrixOrVector := false
	for _, subOp := range expr.Operands {
		if subOp.Type == nil {
			return op, nil
		}
		if isHardwareMatrix(subOp.Type) || isHardwareVector(subOp.Type) {
			anyMatrixOrVector = true
		}
	}

	if !anyMatrixOrVector {
		// Pure scalar expression: preserve flat N-ary expression for backend codegen & constant folding
		return op, nil
	}

	// Contiguous scalar grouping applies to associative operations (* and +).
	// Subtraction and division are left-associative and not associative.
	var items []chainedOperand
	if expr.Operation == program.Operation_MUL || expr.Operation == program.Operation_ADD {
		for i := 0; i < len(expr.Operands); {
			if expr.Operands[i].Type.IsScalar() {
				start := i
				for i < len(expr.Operands) && expr.Operands[i].Type.IsScalar() {
					i++
				}
				runLen := i - start
				if runLen >= 2 {
					groupOperands := make([]*program.Operand, runLen)
					copy(groupOperands, expr.Operands[start:i])
					grpExpr := program.NewExpression(expr.Operation, groupOperands, expr.Operands[start].Start, expr.Operands[i-1].End)
					grpExpr.Type = program.NewScalarType()
					items = append(items, chainedOperand{
						op:      grpExpr,
						origIdx: i, // 1-based index of the last operand in the run
					})
				} else {
					items = append(items, chainedOperand{
						op:      expr.Operands[start],
						origIdx: start + 1,
					})
				}
			} else {
				items = append(items, chainedOperand{
					op:      expr.Operands[i],
					origIdx: i + 1,
				})
				i++
			}
		}
	} else {
		items = make([]chainedOperand, len(expr.Operands))
		for i, subOp := range expr.Operands {
			items[i] = chainedOperand{
				op:      subOp,
				origIdx: i + 1,
			}
		}
	}

	acc := items[0].op
	for i := 1; i < len(items); i++ {
		step := items[i].op
		origIdx := items[i].origIdx
		lowered, err := lowerBinaryOp(file, expr.Operation, acc, step, origIdx)
		if err != nil {
			return nil, err
		}
		acc = lowered
	}

	return acc, nil
}

func lowerBinaryOp(file *program.PackageFile, op program.Operation, acc, step *program.Operand, opIdx int) (*program.Operand, error) {
	derivedType, err := program.DeriveBinaryOpType(op, acc.Type, step.Type, opIdx)
	if err != nil {
		return nil, err
	}

	start := acc.Start
	end := step.End

	lMat := isHardwareMatrix(acc.Type)
	rMat := isHardwareMatrix(step.Type)
	lVec := isHardwareVector(acc.Type)
	rVec := isHardwareVector(step.Type)
	lScal := acc.Type.IsScalar()
	rScal := step.Type.IsScalar()

	var result *program.Operand

	switch op {
	case program.Operation_MUL:
		if lScal && rScal {
			result = program.NewExpression(op, []*program.Operand{acc, step}, start, end)
		} else if lScal && rMat {
			ensureImported(file, "matrix", start, end)
			result = program.NewFnCall("matrix", "scale", []*program.Operand{acc, step}, start, end)
		} else if lMat && rScal {
			ensureImported(file, "matrix", start, end)
			result = program.NewFnCall("matrix", "scale", []*program.Operand{step, acc}, start, end)
		} else if lMat && rMat {
			ensureImported(file, "matrix", start, end)
			result = program.NewFnCall("matrix", "multiply", []*program.Operand{acc, step}, start, end)
		} else if lMat && rVec {
			ensureImported(file, "vector", start, end)
			result = program.NewFnCall("vector", "multiply", []*program.Operand{acc, step}, start, end)
		} else if lScal && rVec {
			ensureImported(file, "vector", start, end)
			result = program.NewFnCall("vector", "multiplyScalar", []*program.Operand{step, acc}, start, end)
		} else if lVec && rScal {
			ensureImported(file, "vector", start, end)
			result = program.NewFnCall("vector", "multiplyScalar", []*program.Operand{acc, step}, start, end)
		} else if lVec && rVec {
			ensureImported(file, "vector", start, end)
			result = program.NewFnCall("vector", "multiplyComponents", []*program.Operand{acc, step}, start, end)
		}

	case program.Operation_ADD:
		if lScal && rScal {
			result = program.NewExpression(op, []*program.Operand{acc, step}, start, end)
		} else if lMat && rMat {
			ensureImported(file, "matrix", start, end)
			result = program.NewFnCall("matrix", "plus", []*program.Operand{acc, step}, start, end)
		} else if lVec && rVec {
			ensureImported(file, "vector", start, end)
			result = program.NewFnCall("vector", "addVector", []*program.Operand{acc, step}, start, end)
		} else if lVec && rScal {
			ensureImported(file, "vector", start, end)
			result = program.NewFnCall("vector", "addScalar", []*program.Operand{acc, step}, start, end)
		} else if lScal && rVec {
			ensureImported(file, "vector", start, end)
			result = program.NewFnCall("vector", "addScalar", []*program.Operand{step, acc}, start, end)
		}

	case program.Operation_SUB:
		if lScal && rScal {
			result = program.NewExpression(op, []*program.Operand{acc, step}, start, end)
		} else if lMat && rMat {
			ensureImported(file, "matrix", start, end)
			result = program.NewFnCall("matrix", "minus", []*program.Operand{acc, step}, start, end)
		} else if lVec && rVec {
			ensureImported(file, "vector", start, end)
			result = program.NewFnCall("vector", "subtractVector", []*program.Operand{acc, step}, start, end)
		} else if lVec && rScal {
			ensureImported(file, "vector", start, end)
			result = program.NewFnCall("vector", "subtractScalarFrom", []*program.Operand{acc, step}, start, end)
		} else if lScal && rVec {
			ensureImported(file, "vector", start, end)
			result = program.NewFnCall("vector", "subtractFromScalar", []*program.Operand{acc, step}, start, end)
		}

	case program.Operation_DIV:
		if lScal && rScal {
			result = program.NewExpression(op, []*program.Operand{acc, step}, start, end)
		} else if lMat && rScal {
			ensureImported(file, "matrix", start, end)
			inv := program.NewExpression(program.Operation_DIV, []*program.Operand{
				program.NewValue(1.0, step.Start, step.End),
				step,
			}, step.Start, step.End)
			inv.Type = program.NewScalarType()
			result = program.NewFnCall("matrix", "scale", []*program.Operand{inv, acc}, start, end)
		} else if lVec && rVec {
			ensureImported(file, "vector", start, end)
			result = program.NewFnCall("vector", "divideComponents", []*program.Operand{acc, step}, start, end)
		} else if lVec && rScal {
			ensureImported(file, "vector", start, end)
			result = program.NewFnCall("vector", "divideByScalar", []*program.Operand{acc, step}, start, end)
		} else if lScal && rVec {
			ensureImported(file, "vector", start, end)
			result = program.NewFnCall("vector", "divideScalarBy", []*program.Operand{acc, step}, start, end)
		}
	}

	if result == nil {
		return nil, internal.InternalErr(fmt.Errorf("unhandled lowering for operation %v at operand %d", op, opIdx))
	}

	result.Type = derivedType
	return result, nil
}

func isReducingBuiltin(name string) *program.BuiltinType {
	builtin := program.BuiltinNamed(name)
	if builtin != nil && builtin.Reduce != program.NoReduce {
		return builtin
	}
	return nil
}

func reduceCSG(op *program.Operand, fnCall *program.FnCall, builtin *program.BuiltinType) (*program.Operand, error) {
	if builtin.Reduce == program.ReduceTail {
		// ReduceTail: meshDifference
		if len(fnCall.Arguments) != 2 {
			return nil, fmt.Errorf("meshDifference expects exactly 2 arguments, got %d", len(fnCall.Arguments))
		}
		baseElements := flattenOperand(fnCall.Arguments[0])
		baseOp, err := foldBinary(op, "meshUnion", baseElements)
		if err != nil {
			return nil, err
		}

		subtrahends := flattenOperand(fnCall.Arguments[1])
		return foldBinary(op, builtin.Name, append([]*program.Operand{baseOp}, subtrahends...))
	} else {
		// ReduceAll: meshUnion, meshIntersect
		var elements []*program.Operand
		if len(fnCall.Arguments) == 1 {
			elements = flattenOperand(fnCall.Arguments[0])
		} else {
			elements = fnCall.Arguments
		}
		return foldBinary(op, builtin.Name, elements)
	}
}

func foldBinary(op *program.Operand, name string, elements []*program.Operand) (*program.Operand, error) {
	if len(elements) == 0 {
		return nil, fmt.Errorf("%s expects at least one mesh, got empty list", name)
	}
	acc := elements[0]
	for i := 1; i < len(elements); i++ {
		acc = program.NewFnCall("", name, []*program.Operand{acc, elements[i]}, op.Start, op.End)
	}
	return acc, nil
}

func flattenOperand(argOp *program.Operand) []*program.Operand {
	if expr, ok := argOp.T.(*program.Operand_Expression); ok && expr.Expression.Operation == program.Operation_COLLECT {
		return expr.Expression.Operands
	}

	if argOp.Type != nil && argOp.Type.Size > 1 {
		elements := make([]*program.Operand, argOp.Type.Size)
		for i := 0; i < int(argOp.Type.Size); i++ {
			elements[i] = makeIndexAccess(argOp, i, argOp.Start, argOp.End)
		}
		return elements
	}

	return []*program.Operand{argOp}
}

func makeIndexAccess(target *program.Operand, index int, start, end *program.Location) *program.Operand {
	return program.NewExpression(
		program.Operation_INDEX,
		[]*program.Operand{
			target,
			program.NewValue(float64(index), start, end),
		},
		start,
		end,
	)
}

func isHardwareMatrix(t *program.Type) bool {
	return t.IsMatrix() && t.ElemType != nil && t.ElemType.Size <= 4
}

func isHardwareVector(t *program.Type) bool {
	return t.IsList() && t.Size >= 2 && t.Size <= 4 && t.ElemType.IsScalar()
}

func ensureImported(file *program.PackageFile, pkg string, start, end *program.Location) {
	for _, imp := range file.Imports {
		if imp.Prefix == pkg || (imp.Prefix == "" && imp.SimpleName() == pkg) {
			return
		}
	}
	file.AddImport(program.NewImport(pkg, pkg, start, end))
}

// AST Walker Helpers

func walkOperand(op *program.Operand, f func(*program.Operand) (*program.Operand, error)) (*program.Operand, error) {
	if op == nil {
		return nil, nil
	}

	switch t := op.T.(type) {
	case *program.Operand_Expression:
		for i, sub := range t.Expression.Operands {
			walked, err := walkOperand(sub, f)
			if err != nil {
				return nil, err
			}
			t.Expression.Operands[i] = walked
		}
	case *program.Operand_FnCall:
		for i, arg := range t.FnCall.Arguments {
			walked, err := walkOperand(arg, f)
			if err != nil {
				return nil, err
			}
			t.FnCall.Arguments[i] = walked
		}
	}

	return f(op)
}

func walkAssignments(assignments []*program.Assignment, f func(*program.Operand) (*program.Operand, error)) error {
	for _, assign := range assignments {
		walked, err := walkOperand(assign.Expression, f)
		if err != nil {
			return err
		}
		assign.Expression = walked

		for _, v := range assign.Variables {
			for _, prop := range v.Properties {
				walkedProp, err := walkOperand(prop.Expression, f)
				if err != nil {
					return err
				}
				prop.Expression = walkedProp
			}
		}
	}
	return nil
}

func walkReturns(returns []*program.Operand, f func(*program.Operand) (*program.Operand, error)) error {
	for i, ret := range returns {
		walked, err := walkOperand(ret, f)
		if err != nil {
			return err
		}
		returns[i] = walked
	}
	return nil
}

func walkFnTypes(fnTypes []*program.FnType, fnStack []*program.Function, f func(*program.Operand, []*program.Function) (*program.Operand, error)) error {
	for _, ft := range fnTypes {
		for _, fg := range ft.Groups {
			for _, fn := range fg.Functions {
				currentStack := append(fnStack, fn)
				if err := walkAssignments(fn.Assignments, func(op *program.Operand) (*program.Operand, error) {
					return f(op, currentStack)
				}); err != nil {
					return err
				}
				if err := walkReturns(fn.Returns, func(op *program.Operand) (*program.Operand, error) {
					return f(op, currentStack)
				}); err != nil {
					return err
				}
				if err := walkFnTypes(fn.FnTypes, currentStack, f); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func walkFile(file *program.PackageFile, f func(*program.Operand, []*program.Function) (*program.Operand, error)) error {
	if err := walkAssignments(file.Assignments, func(op *program.Operand) (*program.Operand, error) {
		return f(op, nil)
	}); err != nil {
		return err
	}
	if err := walkReturns(file.Returns, func(op *program.Operand) (*program.Operand, error) {
		return f(op, nil)
	}); err != nil {
		return err
	}
	if err := walkReturns(file.ColorReturns, func(op *program.Operand) (*program.Operand, error) {
		return f(op, nil)
	}); err != nil {
		return err
	}
	if err := walkReturns(file.MaterialReturns, func(op *program.Operand) (*program.Operand, error) {
		return f(op, nil)
	}); err != nil {
		return err
	}
	if err := walkFnTypes(file.FnTypes, nil, f); err != nil {
		return err
	}
	return nil
}
