package unroll

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

// Transform lowers collection arithmetic, unary negation, and math op calls
// on collections wider than a vec4 (width > 4) and lists of collections into
// canonical Collect expressions. Operations on canonical vectors (width 2..4)
// are preserved.
//
// It returns true if any AST modifications were made.
func Transform(prog *program.Program) (bool, error) {
	if prog == nil {
		return false, nil
	}

	ctx := &unrollContext{
		prog: prog,
	}

	if prog.Primary != nil {
		if err := ctx.transformPackage(prog.Primary); err != nil {
			return false, err
		}
	}

	for _, dep := range prog.Dependencies {
		if err := ctx.transformPackage(dep); err != nil {
			return false, err
		}
	}

	return ctx.modified, nil
}

type unrollContext struct {
	prog        *program.Program
	pkg         *program.Package
	file        *program.PackageFile
	fnStack     []*program.Function
	hoisted     []*program.Assignment
	tempCounter int
	modified    bool
}

func (c *unrollContext) nextTempName() string {
	c.tempCounter++
	return fmt.Sprintf("_unroll_t%d", c.tempCounter)
}

func (c *unrollContext) transformPackage(pkg *program.Package) error {
	c.pkg = pkg
	for _, file := range pkg.Files {
		if err := c.transformFile(file); err != nil {
			return err
		}
	}
	return nil
}

func (c *unrollContext) transformFile(file *program.PackageFile) error {
	c.file = file

	// 1. Walk file assignments
	var newAssignments []*program.Assignment
	for _, assign := range file.Assignments {
		c.hoisted = nil
		walked, err := c.unrollOperand(assign.Expression)
		if err != nil {
			return err
		}
		assign.Expression = walked

		for _, v := range assign.Variables {
			if v != nil {
				for _, prop := range v.Properties {
					if prop != nil {
						walkedProp, err := c.unrollOperand(prop.Expression)
						if err != nil {
							return err
						}
						prop.Expression = walkedProp
					}
				}
			}
		}

		if len(c.hoisted) > 0 {
			newAssignments = append(newAssignments, c.hoisted...)
			c.hoisted = nil
		}
		newAssignments = append(newAssignments, assign)
	}
	file.Assignments = newAssignments

	// 2. Walk file returns
	for i, ret := range file.Returns {
		c.hoisted = nil
		walked, err := c.unrollOperand(ret)
		if err != nil {
			return err
		}
		file.Returns[i] = walked
		if len(c.hoisted) > 0 {
			file.Assignments = append(file.Assignments, c.hoisted...)
			c.hoisted = nil
		}
	}
	for i, ret := range file.ColorReturns {
		if ret == nil {
			continue
		}
		c.hoisted = nil
		walked, err := c.unrollOperand(ret)
		if err != nil {
			return err
		}
		file.ColorReturns[i] = walked
		if len(c.hoisted) > 0 {
			file.Assignments = append(file.Assignments, c.hoisted...)
			c.hoisted = nil
		}
	}
	for i, ret := range file.MaterialReturns {
		if ret == nil {
			continue
		}
		c.hoisted = nil
		walked, err := c.unrollOperand(ret)
		if err != nil {
			return err
		}
		file.MaterialReturns[i] = walked
		if len(c.hoisted) > 0 {
			file.Assignments = append(file.Assignments, c.hoisted...)
			c.hoisted = nil
		}
	}

	// 3. Walk functions
	if err := c.transformFnTypes(file.FnTypes); err != nil {
		return err
	}

	return nil
}

func (c *unrollContext) transformFnTypes(fnTypes []*program.FnType) error {
	for _, ft := range fnTypes {
		for _, fg := range ft.Groups {
			for _, fn := range fg.Functions {
				if err := c.transformFunction(fn); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c *unrollContext) transformFunction(fn *program.Function) error {
	c.fnStack = append(c.fnStack, fn)
	defer func() {
		c.fnStack = c.fnStack[:len(c.fnStack)-1]
	}()

	var newAssignments []*program.Assignment
	for _, assign := range fn.Assignments {
		c.hoisted = nil
		walked, err := c.unrollOperand(assign.Expression)
		if err != nil {
			return err
		}
		assign.Expression = walked

		for _, v := range assign.Variables {
			if v != nil {
				for _, prop := range v.Properties {
					if prop != nil {
						walkedProp, err := c.unrollOperand(prop.Expression)
						if err != nil {
							return err
						}
						prop.Expression = walkedProp
					}
				}
			}
		}

		if len(c.hoisted) > 0 {
			newAssignments = append(newAssignments, c.hoisted...)
			c.hoisted = nil
		}
		newAssignments = append(newAssignments, assign)
	}
	fn.Assignments = newAssignments

	for i, ret := range fn.Returns {
		c.hoisted = nil
		walked, err := c.unrollOperand(ret)
		if err != nil {
			return err
		}
		fn.Returns[i] = walked
		if len(c.hoisted) > 0 {
			fn.Assignments = append(fn.Assignments, c.hoisted...)
			c.hoisted = nil
		}
	}

	if err := c.transformFnTypes(fn.FnTypes); err != nil {
		return err
	}

	return nil
}

func (c *unrollContext) unrollOperand(op *program.Operand) (*program.Operand, error) {
	switch t := op.T.(type) {
	case *program.Operand_FnCall:
		fc := t.FnCall
		for i, arg := range fc.Arguments {
			walked, err := c.unrollOperand(arg)
			if err != nil {
				return nil, err
			}
			fc.Arguments[i] = walked
		}
		// Unroll the "normalize" function call for single-element lists.
		if fc.Reference != nil &&
			(fc.Reference.Prefix == "" || fc.Reference.Prefix == "builtin") &&
			fc.Reference.Name == "normalize" &&
			len(fc.Arguments) == 1 {
			argType := c.operandType(fc.Arguments[0])
			if argType != nil && argType.IsList() && argType.Size == 1 {
				elem, err := c.elementAt(fc.Arguments[0], 0)
				if err != nil {
					return nil, err
				}
				signCall := &program.Operand{
					T: &program.Operand_FnCall{
						FnCall: &program.FnCall{
							Reference: &program.Reference{Name: "sign"},
							Arguments: []*program.Operand{elem},
						},
					},
					Start: op.Start,
					End:   op.End,
					Type:  program.NewScalarType(),
				}
				res := program.NewExpression(program.Operation_COLLECT, []*program.Operand{signCall}, op.Start, op.End)
				res.Type = argType
				c.modified = true
				return res, nil
			}
		}
		return op, nil

	case *program.Operand_Expression:
		expr := t.Expression
		for i, sub := range expr.Operands {
			walked, err := c.unrollOperand(sub)
			if err != nil {
				return nil, err
			}
			expr.Operands[i] = walked
		}

		if isUnrollableOp(expr.Operation) {
			resType := c.effectiveType(op, expr)
			if resType != nil {
				width := unrollWidth(resType)
				if width > 0 {
					return c.unrollExpression(op, expr, resType, width)
				}
			}
		}
		return op, nil
	}

	return op, nil
}

func (c *unrollContext) unrollExpression(op *program.Operand, expr *program.Expression, resType *program.Type, width int) (*program.Operand, error) {
	c.modified = true
	results := make([]*program.Operand, width)

	var elemType *program.Type
	if resType.IsMatrix() && resType.ElemType != nil {
		elemType = resType.ElemType
	} else {
		elemType = program.NewScalarType()
	}

	for k := 0; k < width; k++ {
		subOperands := make([]*program.Operand, len(expr.Operands))
		for j, sub := range expr.Operands {
			elem, err := c.elementAt(sub, k)
			if err != nil {
				return nil, err
			}
			subOperands[j] = elem
		}

		elemExpr := program.NewExpression(expr.Operation, subOperands, op.Start, op.End)
		elemExpr.Type = elemType
		// Recursively unroll in case of nested collections (lists of wide vectors)
		unrolledElem, err := c.unrollOperand(elemExpr)
		if err != nil {
			return nil, err
		}
		results[k] = unrolledElem
	}

	res := program.NewExpression(program.Operation_COLLECT, results, op.Start, op.End)
	res.Type = resType
	return res, nil
}

func (c *unrollContext) elementAt(sub *program.Operand, k int) (*program.Operand, error) {
	// 1. Scalar operand broadcasts: returns the scalar operand directly.
	if st := c.operandType(sub); st != nil && st.IsScalar() {
		return sub, nil
	}

	// 2. Collect literal or reference assigned to Collect: unpack directly.
	if elements := program.StaticElements(c.prog, c.file, c.fnStack, sub); elements != nil {
		if k < len(elements) {
			return elements[k], nil
		}
	}

	// 3. Atom (reference or existing index access): index directly.
	if isAtom(sub) {
		return makeIndexAccess(sub, k, sub.Start, sub.End), nil
	}

	// 4. Non-atomic compound expression: hoist to temporary assignment to avoid duplicate evaluation.
	tmpName := c.nextTempName()
	assign := &program.Assignment{
		Variables:  []*program.Variable{{Name: tmpName}},
		Expression: sub,
	}
	c.hoisted = append(c.hoisted, assign)
	ref := program.NewReference("", tmpName, sub.Start, sub.End)
	return makeIndexAccess(ref, k, sub.Start, sub.End), nil
}

func isAtom(op *program.Operand) bool {
	switch t := op.T.(type) {
	case *program.Operand_Reference, *program.Operand_Value:
		return true
	case *program.Operand_Expression:
		return t.Expression != nil && t.Expression.Operation == program.Operation_INDEX
	}
	return false
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

// unrollWidth returns the collection width if the type must be unrolled, or 0 if preserved as-is.
//
// Returns non-zero (t.Size) when hardware SIMD cannot execute the operation directly:
// - 1-element lists: WGSL lacks hardware vec1 SIMD and prohibits array operator arithmetic.
// - Wide lists (width > 4): GPUs only provide native hardware SIMD vectors up to vec4.
// - Matrices (lists of collections): desugared into component collections.
//
// Returns 0 for types supported natively:
// - Scalars: executed natively without collection containers.
// - Hardware vectors (vec2..vec4): executed directly via GPU SIMD vector instructions.
func unrollWidth(t *program.Type) int {
	// Maximum dimension supported by GPU SIMD vector types (vec2..vec4).
	const vecOperatorLimit = 4

	if t == nil {
		return 0
	}
	if t.IsList() {
		if t.Size == 1 || t.Size > vecOperatorLimit {
			return int(t.Size)
		}
	}
	if t.IsMatrix() {
		return int(t.Size)
	}
	return 0
}

func (c *unrollContext) lookupVarType(name string) *program.Type {
	if len(c.fnStack) > 0 {
		fn := c.fnStack[len(c.fnStack)-1]
		for _, p := range fn.Parameters {
			if p != nil && p.Name == name && p.Type != nil {
				return p.Type
			}
		}
		for _, a := range fn.Assignments {
			if a != nil {
				for _, v := range a.Variables {
					if v != nil && v.Name == name && v.Type != nil {
						return v.Type
					}
				}
			}
		}
	}
	if c.file != nil {
		for _, a := range c.file.Assignments {
			if a != nil {
				for _, v := range a.Variables {
					if v != nil && v.Name == name && v.Type != nil {
						return v.Type
					}
				}
			}
		}
	}
	return nil
}

func (c *unrollContext) operandType(op *program.Operand) *program.Type {
	if op.Type != nil {
		return op.Type
	}
	if ref := op.GetReference(); ref != nil {
		return c.lookupVarType(ref.Name)
	}
	if ex := op.GetExpression(); ex != nil {
		if ex.Operation == program.Operation_COLLECT && len(ex.Operands) > 0 {
			elemT := c.operandType(ex.Operands[0])
			if elemT != nil && !elemT.IsScalar() {
				return program.NewMatrixType(int32(len(ex.Operands)), elemT)
			}
			return program.NewListType(int32(len(ex.Operands)), program.NewScalarType())
		}
		if isUnrollableOp(ex.Operation) {
			for _, sub := range ex.Operands {
				st := c.operandType(sub)
				if st != nil && !st.IsScalar() {
					return st
				}
			}
		}
	}
	return nil
}

func (c *unrollContext) effectiveType(op *program.Operand, expr *program.Expression) *program.Type {
	if op.Type != nil {
		return op.Type
	}
	for _, sub := range expr.Operands {
		st := c.operandType(sub)
		if st != nil && !st.IsScalar() {
			return st
		}
	}
	return nil
}

func isUnrollableOp(op program.Operation) bool {
	switch op {
	case program.Operation_ADD,
		program.Operation_SUB,
		program.Operation_MUL,
		program.Operation_DIV,
		program.Operation_MODULO,
		program.Operation_NEG,
		program.Operation_ABS,
		program.Operation_MIN,
		program.Operation_MAX,
		program.Operation_SQRT,
		program.Operation_SQUARE,
		program.Operation_FLOOR,
		program.Operation_CEIL,
		program.Operation_ROUND,
		program.Operation_EXP,
		program.Operation_LN,
		program.Operation_SIN,
		program.Operation_COS,
		program.Operation_TAN,
		program.Operation_ASIN,
		program.Operation_ACOS,
		program.Operation_ATAN,
		program.Operation_ATAN2,
		program.Operation_POW,
		program.Operation_EQ,
		program.Operation_NE,
		program.Operation_LT,
		program.Operation_LE,
		program.Operation_GT,
		program.Operation_GE,
		program.Operation_COMPARE:
		return true
	}
	return false
}
