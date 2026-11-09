package build

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/internal/types"
	"gradientcontrol.com/omega/core/program"
	errlog "log"
)

// Builder implements [program.OperandVisitor].
type Builder struct {
	rootScope      *Scope
	delegate       Delegate
	importedScopes map[string]*Scope
	progScope      *scope.ProgScope

	// If true, the agent will build every function every time it is called
	// If false, the agent will build each function only once
	buildFnEveryTime bool

	inlineBindings []map[string]Expression
	inRemap        bool

	// inlineSites names the inlined calls being built, outermost first: the
	// callee and the call's source position. Keyed builtins use it to tell
	// their call sites apart (see InlineCallChain).
	inlineSites []string
}

// Build the specified package.
//
// buildFnEveryTime should be set to true
// if the delegate wants to build every function every time it is called.
func Build(delegate Delegate, prog *program.Program, buildFnEveryTime bool) error {
	// Extract a ParameterChecker if provided by the delegate so that delegate-defined
	// dynamic parameters (e.g. runtime parameters or uniforms) are preserved
	// and not erroneously evaluated as compile-time constants.
	var pc eval.ParameterChecker
	if d, ok := delegate.(eval.ParameterChecker); ok {
		pc = d
	}

	// Perform type inference and semantic analysis across all AST packages,
	// populating program.Type metadata on expressions and variables.
	if err := types.Analyze(prog); err != nil {
		return err
	}

	// Constant-fold expressions using inferred type metadata and parameter constraints,
	// simplifying constant subtrees before building scopes and emitting delegate code.
	if _, err := eval.Fold(prog, pc); err != nil {
		return err
	}

	progScope, err := scope.Init(prog)
	if err != nil {
		return err
	}

	b := &Builder{
		delegate:         delegate,
		buildFnEveryTime: buildFnEveryTime,
		importedScopes:   make(map[string]*Scope),
		progScope:        progScope,
	}

	return b.build(prog, progScope)
}

func (b *Builder) build(prog *program.Program, progScope *scope.ProgScope) error {
	rootScope := NewScope(b, "root", nil, nil, progScope)
	b.rootScope = rootScope
	primaryScope := NewScope(b, "", nil, rootScope, progScope.Primary)

	b.delegate.SetScope(primaryScope)

	vars, err := b.delegate.NewImpliedVars(primaryScope)
	if err != nil {
		return err
	}
	for _, v := range vars {
		rootScope.AddLocalVar(v)
	}

	err = b.delegate.BuildPackage(prog.Primary)
	if err != nil {
		return err
	}

	// Map to keep track of file scopes
	fileScopes := make([]*Scope, len(prog.Primary.Files))

	// Iterate over each file in the package.
	// 1. Process named imports and register function types globally
	for idx, file := range prog.Primary.Files {
		// Enforce constraint: only the last file can have a return statement
		if idx < len(prog.Primary.Files)-1 && len(file.Returns) > 0 {
			loc := file.Returns[0].Start
			if loc == nil {
				loc = &program.Location{Line: 1, Column: 1, Offset: 0}
			}
			return ErrLoc(primaryScope, loc, fmt.Errorf("only the last file in a package can have a return statement"))
		}

		fileScope := NewScope(b, fmt.Sprintf("file_%d", idx), nil, primaryScope, progScope.Primary.Files[idx])
		fileScope.isFileScope = true
		fileScope.fileIdx = idx
		fileScopes[idx] = fileScope

		for _, imp := range file.Imports {
			err = b.visitImport(imp, rootScope, fileScope)
			if err != nil {
				return err
			}
			CheckScope(b, fileScope)
		}

		b.delegate.SetScope(fileScope)
		for _, fnType := range file.FnTypes {
			err := b.visitFnType(fnType, primaryScope)
			if err != nil {
				return err
			}
			CheckScope(b, fileScope)
		}
	}

	// 2. Compile all assignments in each file's specific scope
	for idx, file := range prog.Primary.Files {
		fileScope := fileScopes[idx]
		b.delegate.SetScope(fileScope)

		for _, assign := range file.Assignments {
			err := b.visitAssignment(assign, fileScope)
			if err != nil {
				return err
			}
			CheckScope(b, fileScope)
		}
	}

	// Restore primaryScope
	b.delegate.SetScope(primaryScope)

	// Compile returns (only from the last file)
	var returnOps []*program.Operand
	var lastFileScope *Scope
	if len(prog.Primary.Files) > 0 {
		lastFileIdx := len(prog.Primary.Files) - 1
		returnOps = prog.Primary.Files[lastFileIdx].Returns
		lastFileScope = fileScopes[lastFileIdx]
	}

	if len(returnOps) == 0 {
		loc := &program.Location{Line: 1, Column: 1, Offset: 0}
		return ErrLoc(primaryScope, loc, fmt.Errorf("expected return as last statement in package"))
	}

	b.delegate.SetScope(lastFileScope)
	defer b.delegate.SetScope(primaryScope)

	expressions := make([]Expression, len(returnOps))
	for i, op := range returnOps {
		expr, err := program.VisitOperand(op, b)
		if err != nil {
			return ErrLoc(lastFileScope, op.Start, err)
		}
		if (op.Type != nil && op.Type.Size == 0) || expr.Type().IsFunction() {
			err = fmt.Errorf("cannot return a function; perhaps forgot parentheses () after function name?")
			return ErrLoc(lastFileScope, op.Start, err)
		}
		expressions[i] = expr
		CheckScope(b, lastFileScope)
	}

	// Hand pre-lowered color returns to delegates that render color.
	// Optional: delegates that don't implement ColorReturnsBuilder render no color.
	// Done before BuildPackageReturns so the delegate can weave the color function
	// into the same module.
	if cb, ok := b.delegate.(ColorReturnsBuilder); ok {
		colorOps := prog.ColorReturns()
		if len(colorOps) > 0 {
			colorExprs := make([]Expression, len(colorOps))
			for i, op := range colorOps {
				if op != nil {
					expr, err := program.VisitOperand(op, b)
					if err != nil {
						return ErrLoc(lastFileScope, op.Start, err)
					}
					colorExprs[i] = expr
				}
				CheckScope(b, lastFileScope)
			}
			if err := cb.BuildPackageColorReturns(colorExprs); err != nil {
				return err
			}
			CheckScope(b, lastFileScope)
		}
	}

	// The same for material returns, for delegates that render materials.
	// Independent of color: a return may have either, both, or neither.
	if mb, ok := b.delegate.(MaterialReturnsBuilder); ok {
		materialOps := prog.MaterialReturns()
		if len(materialOps) > 0 {
			materialExprs := make([]Expression, len(materialOps))
			for i, op := range materialOps {
				if op != nil {
					expr, err := program.VisitOperand(op, b)
					if err != nil {
						return ErrLoc(lastFileScope, op.Start, err)
					}
					materialExprs[i] = expr
				}
				CheckScope(b, lastFileScope)
			}
			if err := mb.BuildPackageMaterialReturns(materialExprs); err != nil {
				return err
			}
			CheckScope(b, lastFileScope)
		}
	}

	err = b.delegate.BuildPackageReturns(returnOps, expressions)
	if err != nil {
		return err
	}
	CheckScope(b, lastFileScope)

	return nil
}

// CheckScope is a function that checks the scope of the builder and the delegate.
// It is used to verify that the delegate is using the correct scope.
// Tests may replace this function to test scope management.
var CheckScope = func(b *Builder, scope *Scope) {
	currScope := b.delegate.GetScope()
	// sanity check - should never happen
	if currScope != scope {
		errlog.Printf("expected delegate.GetScope() to return %s (parent=%v, fileScope=%v), got %s (parent=%v, fileScope=%v)",
			scope, scope.parent, scope.isFileScope, currScope, currScope.parent, currScope.isFileScope)
	}
}

func (b *Builder) pushInlineBindings(bindings map[string]Expression) {
	b.inlineBindings = append(b.inlineBindings, bindings)
}

func (b *Builder) popInlineBindings() {
	if len(b.inlineBindings) > 0 {
		b.inlineBindings = b.inlineBindings[:len(b.inlineBindings)-1]
	}
}

func (b *Builder) getInlinedBinding(name string) (Expression, bool) {
	for i := len(b.inlineBindings) - 1; i >= 0; i-- {
		if expr, ok := b.inlineBindings[i][name]; ok {
			return expr, true
		}
	}
	return nil, false
}

func (b *Builder) isParameter(op *program.Operand) bool {
	if pc, ok := b.delegate.(eval.ParameterChecker); ok {
		return pc.IsParameter(op)
	}
	return false
}

// buildConstExpressionWithType shapes a flattened slice of constant numeric values (vals)
// into a structured ConstExpression matching the AST Type t (scalar, 1D vector, or nested matrix/array).
func (b *Builder) buildConstExpressionWithType(op *program.Operand, vals []float64, t *program.Type) (ConstExpression, error) {
	// Case 1: Untyped constant.
	// If no explicit type metadata is available, interpret a single value as a scalar,
	// or collect multiple values into a flat vector constant.
	if t == nil {
		if len(vals) == 1 {
			return b.delegate.BuildValue(op, vals[0])
		}
		vecLen := len(vals)
		childExprs := make([]Expression, vecLen)
		for i, val := range vals {
			child, err := b.delegate.BuildValue(op, val)
			if err != nil {
				return nil, err
			}
			childExprs[i] = child
		}
		return b.delegate.BuildConstCollect(op, nil, nil, childExprs, vals)
	}

	// Case 2: Explicit scalar type (Size == 1 and no element subtype).
	// Validate that exactly one value is provided and build the scalar constant.
	if t.Size == 1 && t.ElemType == nil {
		if len(vals) != 1 {
			return nil, fmt.Errorf("constant value length mismatch for scalar: expected 1, got %d", len(vals))
		}
		return b.delegate.BuildValue(op, vals[0])
	}

	// Case 3: Compound or multi-dimensional type (e.g., matrices or arrays of vectors).
	// Recursively partition the flattened values slice into chunks sized by ElemType,
	// build each child expression recursively, and aggregate them into a parent collect expression.
	if t.ElemType != nil {
		elemLen := 1
		for elem := t.ElemType; elem != nil; elem = elem.ElemType {
			elemLen *= max(1, int(elem.Size))
		}
		numElems := int(t.Size)
		if len(vals) != numElems*elemLen {
			return nil, fmt.Errorf("constant value length mismatch: expected %d, got %d", numElems*elemLen, len(vals))
		}
		childExprs := make([]Expression, numElems)
		for i := 0; i < numElems; i++ {
			chunk := vals[i*elemLen : (i+1)*elemLen]
			child, err := b.buildConstExpressionWithType(op, chunk, t.ElemType)
			if err != nil {
				return nil, err
			}
			childExprs[i] = child
		}
		return b.delegate.BuildConstCollect(op, nil, nil, childExprs, vals)
	}

	// Case 4: Flat 1D vector (Size > 1 and no element subtype).
	// Build individual scalar expressions for each component and aggregate them into a vector collect expression.
	if t.Size > 1 {
		vecLen := int(t.Size)
		if len(vals) != vecLen {
			return nil, fmt.Errorf("constant value length mismatch: expected %d, got %d", vecLen, len(vals))
		}
		childExprs := make([]Expression, vecLen)
		for i, val := range vals {
			child, err := b.delegate.BuildValue(op, val)
			if err != nil {
				return nil, err
			}
			childExprs[i] = child
		}
		return b.delegate.BuildConstCollect(op, nil, nil, childExprs, vals)
	}

	// Case 5: Fallback for scalar constants with Size <= 1.
	if len(vals) != 1 {
		return nil, fmt.Errorf("constant value length mismatch for scalar: expected 1, got %d", len(vals))
	}
	return b.delegate.BuildValue(op, vals[0])
}
