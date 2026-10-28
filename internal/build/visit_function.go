package build

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/program"
)

func (b *Builder) visitFnType(fnType *program.FnType, scope *Scope) error {
	// This assumes the caller has already set the delegate's active scope to the file scope
	// where this function type is declared (as done in the build.go and visit_import.go loops).
	_, err := scope.RegisterFnDecl(fnType.Name, fnType, b.delegate.NewFnVar, b.delegate.GetScope())
	if err != nil {
		return ErrLoc(scope, nil, err)
	}
	return nil
}

// VisitFnCall implements [program.OperandVisitor].
func (b *Builder) VisitFnCall(op *program.Operand, fnCall *program.FnCall) (Expression, error) {
	// Preserve the argument shape when inlining leaves an untyped annotation.
	if fnCall.Reference.Prefix == "" {
		switch fnCall.Reference.Name {
		case "uiPoint", "uiTransform", "uiSpline", "uiColor", "uiDirection":
			return program.VisitOperand(fnCall.Arguments[0], b)
		}
	}
	if len(op.ConstValues) > 0 && !b.isParameter(op) {
		return b.buildConstExpressionWithType(op, op.ConstValues, op.Type)
	}
	return b.visitFnCallImpl(op, fnCall)
}

func (b *Builder) visitFnCallImpl(op *program.Operand, fnCall *program.FnCall) (Expression, error) {
	argOps := fnCall.Arguments
	args := make([]Expression, len(argOps))
	for i, argOp := range argOps {
		argExpr, err := program.VisitOperand(argOp, b)
		if err != nil {
			return nil, ErrLoc(b.delegate.GetScope(), argOp.Start, err)
		}
		args[i] = argExpr
	}

	var builtin *program.BuiltinType
	if fnCall.Reference.Prefix == "" {
		builtin = program.BuiltinNamed(fnCall.Reference.Name)
	}
	if builtin != nil {
		currScope := b.delegate.GetScope()
		foundDecl, foundFn, _ := currScope.FindFunction(fnCall.Reference.Name, len(args))
		if foundDecl == nil || foundFn == nil {
			return b.visitBuiltin(builtin, op, fnCall, args)
		}
	}

	fnDecl, fn, enclosingScope, err := b.findFunction(fnCall, op, len(args))
	if err != nil {
		return nil, err
	}

	// A function that reaches a keyed builtin is inlined whatever its
	// arguments, so the builtin sees its caller's constants (keyed_builtin.go).
	if isInlinableFunction(fn) && enclosingScope.reachesKeyedBuiltin(fn, map[string]bool{}) {
		return b.inlineFuncCall(op, fnDecl, fn, enclosingScope, args)
	}

	if isInlinableFunction(fn) && allConstArgs(args) {
		paramExprTypes := make(map[string]ExpressionType, len(fn.Parameters))
		for _, p := range fn.Parameters {
			paramExprTypes[p.Name] = ConstExprType
		}
		visitedKey := fmt.Sprintf("%s.%s/%d", fnCall.Reference.Prefix, fnCall.Reference.Name, len(args))
		visited := map[string]bool{visitedKey: true}

		// A p/time-dependent function can still be inlined when its local
		// assignments all fold to constants — e.g. a wrapper that forwards a
		// constant string to text() must inline so the string folds to a glyph
		// set rather than emitting the whole font. Only a genuinely p-dependent
		// local keeps it a real function. (Inlining is a safe structural
		// substitution outside a remap; inside one it would drop the coordinate
		// substitution, so the inRemap guard below still applies.)
		enclosing := enclosingScope
		localFns := anonFnDecls(fn)
		assignTypes := make(map[string]ExpressionType, len(paramExprTypes))
		for k, v := range paramExprTypes {
			assignTypes[k] = v
		}
		hasBlockingAssignment := false
		for _, assign := range fn.Assignments {
			rhsType := enclosing.getExpressionType(assign.Expression, assignTypes, visited, localFns)
			// A p/time-dependent local does not block inlining when it
			// forwards a constant collection into a fold (text()'s
			// string→glyph unroll): inlining must keep the collection
			// constant or text() emits the entire font. A Dynamic local, or
			// a p/time local that unlocks no such fold, keeps the function a
			// real function.
			if rhsType != ConstExprType && !hasBlockingAssignment {
				if rhsType == DynamicExprType ||
					!enclosing.unlocksConstUnroll(assign.Expression, assignTypes, visited, localFns) {
					hasBlockingAssignment = true
				}
			}
			for _, v := range assign.Variables {
				assignTypes[v.Name] = rhsType
			}
		}

		shouldInline := len(b.inlineBindings) > 0 && !hasBlockingAssignment
		if !shouldInline {
			switch enclosing.analyzeFnBody(fn, paramExprTypes, visited) {
			case ConstExprType:
				shouldInline = true
			case ExprWithPAndTimeOnly:
				// A p/time-dependent function normally must not be inlined inside
				// a remap, or inlining would bake in the original p and drop the
				// coordinate substitution. The exception is a function that
				// unlocks a constant-collection fold (text() unrolling a constant
				// string into glyphs): it must inline so the string folds rather
				// than emitting the whole font, and the remap delegate re-derives
				// the inlined target under the remapped p (see Remap), so the
				// substitution is preserved.
				shouldInline = !hasBlockingAssignment &&
					(!b.inRemap || enclosing.fnBodyUnlocksConstUnroll(fn, paramExprTypes, visited, localFns))
			}
		}
		if shouldInline {
			return b.inlineFuncCall(op, fnDecl, fn, enclosingScope, args)
		}
	}

	fnInfo, err := b.buildFunction(op, fnDecl, fn, enclosingScope, fnCall, args)
	if err != nil {
		return nil, err
	}

	expr, err := b.delegate.BuildFuncCall(op, fnCall, fnInfo, args)
	if err != nil {
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
	}

	return expr, nil
}

// tryInlineMultiReturn handles a multi-value assignment whose right-hand side is
// a direct call to a multi-return function that either returns all-constant
// values or forwards a constant collection into a fold. The constant case
// (e.g. `point, normal = pointNormal()` returning literal collections) must
// inline as separate assignments — built as a real call, each value becomes a
// runtime struct field and stops being a compile-time constant, breaking any
// downstream use that requires one (collection indexes, text, vdb params). The
// fold case (e.g. `shaft, vents = sprue.labeled("Manifold", "A36")`, where
// labeled() routes the string into text()) must be inlined or the string stays
// a runtime parameter and text() emits the entire font. Single-return inlining
// lives in VisitFunctionCall; multi-return values are only ever consumed by a
// multi-value assignment, so the decision is made here.
//
// It returns (exprs, true, nil) with one expression per return value when it
// inlines, or (nil, false, nil) to fall through to the normal call path. Cheap
// static checks gate the decision before any argument is built, so a fall-
// through does no redundant work.
func (b *Builder) tryInlineMultiReturn(op *program.Operand, names []string) ([]Expression, bool, error) {
	fc := op.GetFnCall()
	if fc == nil {
		return nil, false, nil
	}
	if fc.Reference.Prefix == "" && program.BuiltinNamed(fc.Reference.Name) != nil {
		return nil, false, nil
	}
	// Inlining inside a remap would drop the coordinate substitution.
	if b.inRemap {
		return nil, false, nil
	}

	fnDecl, fn, enclosingScope, err := b.findFunction(fc, op, len(fc.Arguments))
	if err != nil {
		// Let the normal path surface the resolution error with its context.
		return nil, false, nil
	}
	if !isMultiReturnInlinable(fn) {
		return nil, false, nil
	}
	if len(fn.Returns) != len(names) {
		return nil, false, nil
	}

	// All arguments must be constant, and the function must forward a constant
	// collection into a fold; otherwise inlining is unwarranted and we keep the
	// function real. Classify arguments statically so no expression is built on
	// the fall-through path.
	scope := b.delegate.GetScope()
	paramExprTypes := make(map[string]ExpressionType, len(fn.Parameters))
	argVisited := map[string]bool{}
	for i, param := range fn.Parameters {
		if i >= len(fc.Arguments) {
			return nil, false, nil
		}
		if scope.getExpressionType(fc.Arguments[i], nil, argVisited, nil) != ConstExprType {
			return nil, false, nil
		}
		paramExprTypes[param.Name] = ConstExprType
	}
	visited := map[string]bool{
		fmt.Sprintf("%s.%s/%d", fc.Reference.Prefix, fc.Reference.Name, len(fc.Arguments)): true,
	}
	enclosing := enclosingScope
	if !enclosing.fnBodyReturnsAllConst(fn, paramExprTypes, visited) &&
		!enclosing.fnBodyUnlocksConstUnroll(fn, paramExprTypes, visited, anonFnDecls(fn)) {
		return nil, false, nil
	}

	// Eligible: build the arguments and inline.
	args := make([]Expression, len(fc.Arguments))
	for i, argOp := range fc.Arguments {
		argExpr, err := program.VisitOperand(argOp, b)
		if err != nil {
			return nil, false, ErrLoc(b.delegate.GetScope(), argOp.Start, err)
		}
		args[i] = argExpr
	}
	exprs, err := b.inlineMultiFuncCall(op, fnDecl, fn, enclosingScope, args)
	if err != nil {
		return nil, false, err
	}
	return exprs, true, nil
}

func (b *Builder) findFunction(fnCall *program.FnCall, op *program.Operand, paramCount int) (*program.FnType, *program.Function, *Scope, error) {
	scope := b.delegate.GetScope()
	if fnCall.Reference.Prefix != "" {
		scope = scope.GetNamedImportScope(fnCall.Reference.Prefix)
		if scope == nil {
			return nil, nil, nil, ErrLoc(b.delegate.GetScope(), op.Start, fmt.Errorf("package %q not found", fnCall.Reference.Prefix))
		}
	}

	fnVar := scope.GetVar(fnCall.Reference.Name)
	if fnVar == nil {
		return nil, nil, nil, ErrLoc(b.delegate.GetScope(), op.Start, fmt.Errorf("function %q not defined", fnCall.Reference.Name))
	}
	lookupName := fnCall.Reference.Name
	fnDecl := scope.ExistingFnDecl(lookupName)
	if fnDecl == nil && fnVar.Type() != nil && fnVar.Type().FnPath != "" {
		fnPath := fnVar.Type().FnPath
		lookupName = fnPath
		if idx := strings.LastIndex(fnPath, "/"); idx >= 0 {
			lookupName = fnPath[idx+1:]
		}
		fnDecl = scope.ExistingFnDecl(lookupName)
	}
	if fnDecl == nil {
		return nil, nil, nil, ErrLoc(b.delegate.GetScope(), op.Start, fmt.Errorf("%q is not a function", fnCall.Reference.Name))
	}

	foundDecl, fn, enclosingScope := scope.FindFunction(lookupName, paramCount)
	if foundDecl != nil {
		fnDecl = foundDecl
	}
	if fn == nil {
		var sigs []string
		fullName := fnCall.Reference.Name
		if fnCall.Reference.Prefix != "" {
			fullName = fnCall.Reference.Prefix + "." + fnCall.Reference.Name
		}
		if fnDecl != nil {
			for _, g := range fnDecl.Groups {
				if g == nil || len(g.Functions) == 0 {
					continue
				}
				f := g.Functions[0]
				var params []string
				for _, p := range f.Parameters {
					params = append(params, p.Name)
				}
				sigs = append(sigs, fmt.Sprintf("%s(%s)", fullName, strings.Join(params, ", ")))
			}
		}
		var suggestion string
		if len(sigs) > 0 {
			suggestion = "; perhaps one of these? " + strings.Join(sigs, ", ")
		}
		return nil, nil, nil, ErrLoc(b.delegate.GetScope(), op.Start, fmt.Errorf("no function named %q with %d parameters%s", fnCall.Reference.Name, paramCount, suggestion))
	}
	return fnDecl, fn, enclosingScope, nil
}

func (b *Builder) buildFunction(op *program.Operand, fnDecl *program.FnType, fn *program.Function, enclosingScope *Scope, fnCall *program.FnCall, args []Expression) (FnInfo, error) {

	// If the function has already been built
	// and the agent has not been asked to build it each time it is called,
	// return the cached function info.
	fnInfo := enclosingScope.InfoForArgs(fn, args)
	if fnInfo != nil && !b.buildFnEveryTime {
		return fnInfo, nil
	}

	fnBodyScope := NewScope(b, fnDecl.Name, op.Start, enclosingScope, b.resolveFnScope(enclosingScope, fnDecl.Name, fn))

	prevScope := b.delegate.GetScope()
	b.delegate.SetScope(fnBodyScope)
	defer b.delegate.SetScope(prevScope)

	// Inline bindings substitute the parameters of a function being inlined
	// at the call site; they must not be visible while building this
	// function's body, or a body reference that shares a name with an
	// enclosing inline parameter (e.g. a reduce lambda's "value") would
	// silently resolve to the caller's binding instead of this function's
	// own parameter or local.
	prevBindings := b.inlineBindings
	b.inlineBindings = nil
	defer func() {
		b.inlineBindings = prevBindings
	}()

	fnInfo, paramVars, err := b.delegate.BuildFunc(fnDecl, fn, fnCall.Reference.Prefix, args)
	if err != nil {
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
	}
	if fnInfo == nil {
		err := fmt.Errorf("missing info for function %q", fnCall.Reference.Name)
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, program.InternalErr(err))
	}
	CheckScope(b, fnBodyScope)

	for _, param := range paramVars {
		fnBodyScope.addParam(param)
	}

	for _, nestedFnType := range fn.FnTypes {
		err := b.visitFnType(nestedFnType, fnBodyScope)
		if err != nil {
			return nil, ErrLoc(b.delegate.GetScope(), nil, err)
		}
		CheckScope(b, fnBodyScope)
	}

	for _, assign := range fn.Assignments {
		err := b.visitAssignment(assign, fnBodyScope)
		if err != nil {
			return nil, ErrLoc(b.delegate.GetScope(), assign.Start, err)
		}
		CheckScope(b, fnBodyScope)
	}

	returnOps := fn.Returns
	returnExprs := make([]Expression, len(returnOps))
	for i, exprOp := range returnOps {
		expr, err := program.VisitOperand(exprOp, b)
		if err != nil {
			return nil, ErrLoc(b.delegate.GetScope(), exprOp.Start, err)
		}
		returnExprs[i] = expr
		CheckScope(b, fnBodyScope)
	}
	fnInfo, err = b.delegate.BuildFuncReturn(fnInfo, returnOps, returnExprs)
	if err != nil {
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
	}
	CheckScope(b, fnBodyScope)

	err = enclosingScope.AddFnInfo(fn, fnInfo)
	if err != nil {
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
	}

	return fnInfo, nil
}

func (b *Builder) resolveFnScope(enclosing *Scope, fnName string, fn *program.Function) scope.Scope {
	if enclosing == nil || enclosing.uScope == nil {
		return nil
	}
	elem := enclosing.uScope.NamedElem(fnName)
	if fnTypeScope, ok := elem.(*scope.FnTypeScope); ok && fnTypeScope != nil {
		paramCount := int32(len(fn.Parameters))
		if gScope, ok := fnTypeScope.Groups[paramCount]; ok && gScope != nil {
			for _, fScope := range gScope.Functions {
				if fScope.Function == fn {
					return fScope
				}
			}
		}
	}
	return nil
}

func (b *Builder) inlineFuncCall(op *program.Operand, fnDecl *program.FnType, fn *program.Function, enclosingScope *Scope, args []Expression) (Expression, error) {
	b.pushInlineSite(fnDecl, op)
	defer b.popInlineSite()

	bindings := make(map[string]Expression)
	for i, paramName := range fn.Parameters {
		if i < len(args) {
			bindings[paramName.Name] = args[i]
		}
	}

	b.pushInlineBindings(bindings)
	defer b.popInlineBindings()

	prevScope := b.delegate.GetScope()
	inlineScope := NewScope(b, fnDecl.Name, op.Start, enclosingScope, nil)
	inlineScope.inlined = true
	for i, paramName := range fn.Parameters {
		if i < len(args) {
			paramVar := b.delegate.NewParamVar(inlineScope, paramName.Name, args[i], args[i].Type())
			inlineScope.addParam(paramVar)
		}
	}
	b.delegate.SetScope(inlineScope)
	defer func() {
		b.delegate.SetScope(prevScope)
	}()

	if len(fn.Returns) != 1 {
		return nil, fmt.Errorf("inlined function must have exactly one return expression")
	}

	for _, nestedFn := range fn.FnTypes {
		err := b.visitFnType(nestedFn, inlineScope)
		if err != nil {
			return nil, err
		}
	}

	for _, assign := range fn.Assignments {
		err := b.visitAssignment(assign, inlineScope)
		if err != nil {
			return nil, err
		}
	}

	return program.VisitOperand(fn.Returns[0], b)
}

func (b *Builder) inlineMultiFuncCall(op *program.Operand, fnDecl *program.FnType, fn *program.Function, enclosingScope *Scope, args []Expression) ([]Expression, error) {
	b.pushInlineSite(fnDecl, op)
	defer b.popInlineSite()

	bindings := make(map[string]Expression)
	for i, paramName := range fn.Parameters {
		if i < len(args) {
			bindings[paramName.Name] = args[i]
		}
	}

	b.pushInlineBindings(bindings)
	defer b.popInlineBindings()

	prevScope := b.delegate.GetScope()
	inlineScope := NewScope(b, fnDecl.Name, op.Start, enclosingScope, nil)
	inlineScope.inlined = true
	for i, paramName := range fn.Parameters {
		if i < len(args) {
			paramVar := b.delegate.NewParamVar(inlineScope, paramName.Name, args[i], args[i].Type())
			inlineScope.addParam(paramVar)
		}
	}
	b.delegate.SetScope(inlineScope)
	defer func() {
		b.delegate.SetScope(prevScope)
	}()

	for _, nestedFn := range fn.FnTypes {
		err := b.visitFnType(nestedFn, inlineScope)
		if err != nil {
			return nil, err
		}
	}

	for _, assign := range fn.Assignments {
		err := b.visitAssignment(assign, inlineScope)
		if err != nil {
			return nil, err
		}
	}

	exprs := make([]Expression, len(fn.Returns))
	for i, ret := range fn.Returns {
		expr, err := program.VisitOperand(ret, b)
		if err != nil {
			return nil, err
		}
		exprs[i] = expr
	}
	return exprs, nil
}

// Return true if all of the expressions are ConstExpression and have values
func allConstArgs(args []Expression) bool {
	for _, arg := range args {
		if ce, ok := arg.(ConstExpression); !ok || ce.ConstValues() == nil {
			return false
		}
	}
	return true
}
