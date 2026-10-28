package build

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

type Expression interface {
	Operand() *program.Operand
	Start() *program.Location
	Precedence() Precedence
	Type() *program.Type
}

type ConstExpression interface {
	Expression
	ConstValues() []float64
}

type CollectExpression interface {
	Expression
	Elements() []Expression
}

type Precedence int

const (
	PrecedenceHighest        Precedence = iota
	PrecedenceUnary                     // !, +, -
	PrecedenceMultiplication            // *, /, %
	PrecedenceAddition                  // +, -
	PrecedenceComparison                // <, <=, >, >=
	PrecedenceEquality                  // ==, !=
	PrecedenceLogicalAnd                // &&
	PrecedenceLogicalOr                 // ||
	PrecedenceTernary                   // ?:
	PrecedenceAssignment                // =
	PrecedenceLowest
)

type ExpressionType int

const (
	ConstExprType ExpressionType = iota
	ExprWithPAndTimeOnly
	DynamicExprType
)

func (s *Scope) getExpressionType(op *program.Operand, paramExprTypes map[string]ExpressionType, visited map[string]bool, localFns map[string]*program.Function) ExpressionType {
	if op == nil {
		return ConstExprType
	}
	switch t := op.T.(type) {
	case *program.Operand_Value:
		return ConstExprType
	case *program.Operand_Reference:
		if t.Reference.Prefix != "" {
			pkgScope := s.GetNamedImportScope(t.Reference.Prefix)
			if pkgScope == nil {
				return DynamicExprType
			}
			if fnDecl := pkgScope.ExistingFnDecl(t.Reference.Name); fnDecl != nil {
				return ConstExprType
			}
			varInfo := pkgScope.GetVar(t.Reference.Name)
			if varInfo != nil {
				if constExpr, ok := varInfo.Expr().(ConstExpression); ok && constExpr.ConstValues() != nil {
					return ConstExprType
				}
			}
			return DynamicExprType
		}

		name := t.Reference.Name
		if name == program.PVarName || name == program.TimeVarName {
			return ExprWithPAndTimeOnly
		}
		if state, ok := paramExprTypes[name]; ok {
			return state
		}
		if fnDecl := s.ExistingFnDecl(name); fnDecl != nil {
			return ConstExprType
		}
		varInfo := s.GetVar(name)
		if varInfo != nil {
			if constExpr, ok := varInfo.Expr().(ConstExpression); ok && constExpr.ConstValues() != nil {
				return ConstExprType
			}
		}

		return DynamicExprType
	case *program.Operand_PropRef:
		refOp := &program.Operand{T: &program.Operand_Reference{Reference: t.PropRef.Reference}}
		return s.getExpressionType(refOp, paramExprTypes, visited, localFns)
	case *program.Operand_Expression:
		e := t.Expression

		// remap(field, coords) re-evaluates field with p := coords, so the
		// result's dependence on the outer p/time comes from coords alone; the
		// field's own p usage is replaced. Classifying by coords also keeps a
		// field built from non-inlinable helpers (e.g. the glyph dispatch under
		// text()) from reading as Dynamic and wrongly blocking inlining.
		if e.Operation == program.Operation_REMAP && len(e.Operands) == 2 {
			return s.getExpressionType(e.Operands[1], paramExprTypes, visited, localFns)
		}

		// For map/reduce, the function operand is a reference to the lambda;
		// analyze the lambda's body (rather than the bare reference, which
		// reads as constant) so that a body referencing the outer p/time —
		// e.g. an inner remap, as in the text() font reducer — propagates its
		// dependence instead of being mistaken for a constant. The lambda is
		// operand 0 for map and for the no-initial reduce(fn, list) form, and
		// operand 1 for reduce(initial, fn, ...lists).
		lambdaIdx := -1
		switch e.Operation {
		case program.Operation_MAP:
			lambdaIdx = 0
		case program.Operation_REDUCE:
			if len(e.Operands) >= 3 {
				lambdaIdx = 1
			} else {
				lambdaIdx = 0
			}
		}

		state := ConstExprType
		for i, child := range e.Operands {
			if e.Operation == program.Operation_ACCESS && i == 1 {
				continue
			}
			var childState ExpressionType
			if i == lambdaIdx {
				childState = s.lambdaBodyType(child, paramExprTypes, visited, localFns)
			} else {
				childState = s.getExpressionType(child, paramExprTypes, visited, localFns)
			}
			if childState == DynamicExprType {
				return DynamicExprType
			}
			if childState == ExprWithPAndTimeOnly {
				state = ExprWithPAndTimeOnly
			}
		}
		return state
	case *program.Operand_FnCall:
		fc := t.FnCall
		if fc.Reference.Prefix == "" && program.BuiltinNamed(fc.Reference.Name) != nil {
			state := ConstExprType
			for _, arg := range fc.Arguments {
				childState := s.getExpressionType(arg, paramExprTypes, visited, localFns)
				if childState == DynamicExprType {
					return DynamicExprType
				}
				if childState == ExprWithPAndTimeOnly {
					state = ExprWithPAndTimeOnly
				}
			}
			return state
		}
		visitedKey := fmt.Sprintf("%s.%s/%d", fc.Reference.Prefix, fc.Reference.Name, len(fc.Arguments))
		if visited[visitedKey] {
			return DynamicExprType
		}

		// Resolve the callee: a hoisted local lambda (gsl_anon...) is found in
		// localFns and analyzed in this scope; everything else resolves through
		// the scope chain.
		var callee *program.Function
		calleeScope := s
		if fc.Reference.Prefix == "" && localFns != nil {
			callee = localFns[fc.Reference.Name]
		}
		if callee == nil {
			fnScope := s
			if fc.Reference.Prefix != "" {
				fnScope = fnScope.GetNamedImportScope(fc.Reference.Prefix)
			}
			if fnScope == nil {
				return DynamicExprType
			}
			_, callee, calleeScope = fnScope.FindFunction(fc.Reference.Name, len(fc.Arguments))
			if callee == nil || !isInlinableFunction(callee) {
				return DynamicExprType
			}
		}

		fnParams := make(map[string]ExpressionType)
		for i, paramName := range callee.Parameters {
			var argState ExpressionType = ConstExprType
			if i < len(fc.Arguments) {
				argState = s.getExpressionType(fc.Arguments[i], paramExprTypes, visited, localFns)
			}
			fnParams[paramName.Name] = argState
		}

		newVisited := make(map[string]bool, len(visited)+1)
		for k, v := range visited {
			newVisited[k] = v
		}
		newVisited[visitedKey] = true

		return calleeScope.analyzeFnBody(callee, fnParams, newVisited)
	default:
		panic(fmt.Sprintf("internal error: unhandled operand type %T in getExpressionType; please add a new case to this switch", t))
	}
}

// anonFnDecls collects the hoisted anonymous (gsl_anon…) function declarations
// in a function body, keyed by name, so map/reduce lambdas and nested closures
// can be resolved while statically analyzing the body. Returns nil when there
// are none.
func anonFnDecls(fn *program.Function) map[string]*program.Function {
	var m map[string]*program.Function
	for _, fnType := range fn.FnTypes {
		if fnType.IsAnonFn() {
			if m == nil {
				m = make(map[string]*program.Function)
			}
			if len(fnType.Groups) > 0 && len(fnType.Groups[0].Functions) > 0 {
				m[fnType.Name] = fnType.Groups[0].Functions[0]
			}
		}
	}
	return m
}

// analyzeFnBody returns the expression type of a function's body, given the
// types of its parameters. Local assignments are folded in order, and the
// function's own hoisted anonymous functions are made available so map/reduce
// lambdas inside it resolve. The receiver scope is the function's enclosing
// scope (where its free references resolve).
func (s *Scope) analyzeFnBody(fn *program.Function, paramExprTypes map[string]ExpressionType, visited map[string]bool) ExpressionType {
	if fn == nil || len(fn.Returns) != 1 {
		return DynamicExprType
	}
	localFns := anonFnDecls(fn)

	fnParams := make(map[string]ExpressionType, len(paramExprTypes))
	for k, v := range paramExprTypes {
		fnParams[k] = v
	}
	for _, assign := range fn.Assignments {
		rhsType := s.getExpressionType(assign.Expression, fnParams, visited, localFns)
		for _, v := range assign.Variables {
			fnParams[v.Name] = rhsType
		}
	}
	return s.getExpressionType(fn.Returns[0], fnParams, visited, localFns)
}

// lambdaBodyType returns the expression type of a map/reduce function operand's
// body. The operand references the lambda (a hoisted local gsl_anon function,
// or a named function); its own bound parameters (value/index,
// accumulator/value) are scoped constant — any dependence those carry from a
// p-dependent collection is already captured by analyzing the collection
// operands — so that only an outer p/time reference in the body (e.g. an inner
// remap) propagates. Returns ConstExprType when the lambda can't be resolved,
// leaving the caller's other operands to determine the result.
func (s *Scope) lambdaBodyType(op *program.Operand, paramExprTypes map[string]ExpressionType, visited map[string]bool, localFns map[string]*program.Function) ExpressionType {
	var name, pkgName string
	if ref, ok := op.T.(*program.Operand_Reference); ok {
		name = ref.Reference.Name
		pkgName = ref.Reference.Prefix
	} else {
		return ConstExprType
	}

	var fn *program.Function
	fnScope := s
	closesOver := false
	if pkgName == "" && localFns != nil {
		fn = localFns[name]
		// A hoisted local lambda closes over the enclosing function's params
		// and locals, so those types must remain visible while analyzing it.
		closesOver = fn != nil
	}
	if fn == nil {
		lookupScope := s
		if pkgName != "" {
			lookupScope = s.GetNamedImportScope(pkgName)
			if lookupScope == nil {
				return ConstExprType
			}
		}
		fnDecl := lookupScope.ExistingFnDecl(name)
		if fnDecl == nil {
			return ConstExprType
		}

		for _, g := range fnDecl.Groups {
			if g != nil && len(g.Functions) > 0 {
				fn = g.Functions[0]
				fnScope = lookupScope.FindFnScope(name)
				break
			}
		}
		if fn == nil {
			return ConstExprType
		}
	}

	visitedKey := fmt.Sprintf("%s.%s/%d", pkgName, name, len(fn.Parameters))
	if visited[visitedKey] {
		return DynamicExprType
	}
	newVisited := make(map[string]bool, len(visited)+1)
	for k, v := range visited {
		newVisited[k] = v
	}
	newVisited[visitedKey] = true

	// The lambda's own bound parameters (value/index, accumulator/value) are
	// constant; a closure also sees the enclosing scope's params/locals.
	fnParams := make(map[string]ExpressionType, len(fn.Parameters))
	if closesOver {
		for k, v := range paramExprTypes {
			fnParams[k] = v
		}
	}
	for _, paramName := range fn.Parameters {
		fnParams[paramName.Name] = ConstExprType
	}
	return fnScope.analyzeFnBody(fn, fnParams, newVisited)
}

// foldCollectionStart returns the operand index at which a reduce/map's
// collection arguments begin, or -1 if e is not a fold. For map the collection
// is operands[1:]; for reduce it is operands[2:] (initial, fn, …collections) or
// operands[1:] for the no-initial reduce(fn, list) form.
func foldCollectionStart(e *program.Expression) int {
	switch e.Operation {
	case program.Operation_COLLECT:
		return 0
	case program.Operation_MAP:
		return 1
	case program.Operation_REDUCE:
		if len(e.Operands) >= 3 {
			return 2
		}
		return 1
	}
	return -1
}

// unlocksConstUnroll reports whether evaluating op reaches a reduce/map whose
// collection operand(s) are constant — the fold text() uses to unroll a constant
// string into its handful of glyphs rather than emitting the entire font
// (~280 KB). A wrapper that forwards a constant collection into such a fold must
// be inlined to keep the collection constant, even when its own locals are
// p/time dependent (e.g. `d = font.left(codes, …); return d;`). Resolution
// mirrors getExpressionType; only inlinable user callees are followed, since
// those are all the inliner can fold through.
func (s *Scope) unlocksConstUnroll(op *program.Operand, paramExprTypes map[string]ExpressionType, visited map[string]bool, localFns map[string]*program.Function) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		e := t.Expression
		if collStart := foldCollectionStart(e); collStart >= 0 && len(e.Operands) > collStart {
			allConst := true
			for _, c := range e.Operands[collStart:] {
				if s.getExpressionType(c, paramExprTypes, visited, localFns) != ConstExprType {
					allConst = false
					break
				}
			}
			if allConst {
				return true
			}
		}
		if e.Operation == program.Operation_INDEX && len(e.Operands) == 2 {
			if s.getExpressionType(e.Operands[1], paramExprTypes, visited, localFns) == ConstExprType {
				return true
			}
		}
		for i, child := range e.Operands {
			if e.Operation == program.Operation_ACCESS && i == 1 {
				continue
			}
			if s.unlocksConstUnroll(child, paramExprTypes, visited, localFns) {
				return true
			}
		}
		return false
	case *program.Operand_FnCall:
		fc := t.FnCall
		for _, arg := range fc.Arguments {
			if s.unlocksConstUnroll(arg, paramExprTypes, visited, localFns) {
				return true
			}
		}
		if fc.Reference.Prefix == "" && program.BuiltinNamed(fc.Reference.Name) != nil {
			return false
		}
		visitedKey := fmt.Sprintf("%s.%s/%d", fc.Reference.Prefix, fc.Reference.Name, len(fc.Arguments))
		if visited[visitedKey] {
			return false
		}

		var callee *program.Function
		calleeScope := s
		if fc.Reference.Prefix == "" && localFns != nil {
			callee = localFns[fc.Reference.Name]
		}
		if callee == nil {
			fnScope := s
			if fc.Reference.Prefix != "" {
				fnScope = fnScope.GetNamedImportScope(fc.Reference.Prefix)
			}
			if fnScope == nil {
				return false
			}
			_, callee, calleeScope = fnScope.FindFunction(fc.Reference.Name, len(fc.Arguments))
			if callee == nil || !isInlinableFunction(callee) {
				return false
			}
		}

		fnParams := make(map[string]ExpressionType, len(callee.Parameters))
		for i, paramName := range callee.Parameters {
			argState := ConstExprType
			if i < len(fc.Arguments) {
				argState = s.getExpressionType(fc.Arguments[i], paramExprTypes, visited, localFns)
			}
			fnParams[paramName.Name] = argState
		}
		newVisited := make(map[string]bool, len(visited)+1)
		for k, v := range visited {
			newVisited[k] = v
		}
		newVisited[visitedKey] = true
		return calleeScope.fnBodyUnlocksConstUnroll(callee, fnParams, newVisited, anonFnDecls(callee))
	case *program.Operand_PropRef:
		refOp := &program.Operand{T: &program.Operand_Reference{Reference: t.PropRef.Reference}}
		return s.unlocksConstUnroll(refOp, paramExprTypes, visited, localFns)
	case *program.Operand_Value, *program.Operand_Reference:
		return false
	default:
		panic(fmt.Sprintf("internal error: unhandled operand type %T in unlocksConstUnroll; please add a new case to this switch", t))
	}
}

// fnBodyReturnsAllConst threads a function's local assignments the same way
// analyzeFnBody does (which itself only handles single-return functions) and
// reports whether every return expression folds to a constant given the
// call-site parameter types. Used to decide whether a multi-return function can
// be inlined at a multi-value assignment so each value stays a compile-time
// constant instead of a runtime struct field.
func (s *Scope) fnBodyReturnsAllConst(fn *program.Function, paramExprTypes map[string]ExpressionType, visited map[string]bool) bool {
	if fn == nil || len(fn.Returns) == 0 {
		return false
	}
	localFns := anonFnDecls(fn)
	fnParams := make(map[string]ExpressionType, len(paramExprTypes))
	for k, v := range paramExprTypes {
		fnParams[k] = v
	}
	for _, assign := range fn.Assignments {
		rhsType := s.getExpressionType(assign.Expression, fnParams, visited, localFns)
		if rhsType != ConstExprType && len(assign.Variables) > 1 {
			// getExpressionType conservatively types a multi-return call
			// Dynamic. A nested const destructure (base, dir = pointNormal())
			// inlines recursively at build time, so recurse here too or the
			// outer function is wrongly rejected.
			if fc, ok := assign.Expression.T.(*program.Operand_FnCall); ok &&
				s.callReturnsAllConst(fc.FnCall, len(assign.Variables), fnParams, visited, localFns) {
				rhsType = ConstExprType
			}
		}
		for _, v := range assign.Variables {
			fnParams[v.Name] = rhsType
		}
	}
	for _, ret := range fn.Returns {
		if s.getExpressionType(ret, fnParams, visited, localFns) != ConstExprType {
			return false
		}
	}
	return true
}

// callReturnsAllConst reports whether a direct call to a multi-return function
// with the given number of expected return values folds every value to a
// constant. Mirrors the callee resolution in unlocksConstUnroll's function-call
// case and requires the shape inlineMultiFuncCall can rebuild, so the constant
// claim is only made when build-time inlining will actually realize it.
func (s *Scope) callReturnsAllConst(fc *program.FnCall, numReturns int, paramExprTypes map[string]ExpressionType, visited map[string]bool, localFns map[string]*program.Function) bool {
	if fc.Reference.Prefix == "" && program.BuiltinNamed(fc.Reference.Name) != nil {
		return false
	}
	visitedKey := fmt.Sprintf("%s.%s/%d", fc.Reference.Prefix, fc.Reference.Name, len(fc.Arguments))
	if visited[visitedKey] {
		return false
	}

	fnScope := s
	if fc.Reference.Prefix != "" {
		fnScope = fnScope.GetNamedImportScope(fc.Reference.Prefix)
	}
	if fnScope == nil {
		return false
	}
	_, callee, calleeScope := fnScope.FindFunction(fc.Reference.Name, len(fc.Arguments))
	if callee == nil || !isMultiReturnInlinable(callee) {
		return false
	}
	if len(callee.Returns) != numReturns {
		return false
	}

	fnParams := make(map[string]ExpressionType, len(callee.Parameters))
	for i, paramName := range callee.Parameters {
		argState := ConstExprType
		if i < len(fc.Arguments) {
			argState = s.getExpressionType(fc.Arguments[i], paramExprTypes, visited, localFns)
		}
		fnParams[paramName.Name] = argState
	}
	newVisited := make(map[string]bool, len(visited)+1)
	for k, v := range visited {
		newVisited[k] = v
	}
	newVisited[visitedKey] = true
	return calleeScope.fnBodyReturnsAllConst(callee, fnParams, newVisited)
}

// fnBodyUnlocksConstUnroll threads a function's local assignments (so a forwarded
// constant collection keeps its constant type) and reports whether any local or
// return expression unlocks a constant fold. Handles multi-return functions
// (the motivating sprue.labeled case) by checking every return.
func (s *Scope) fnBodyUnlocksConstUnroll(fn *program.Function, paramExprTypes map[string]ExpressionType, visited map[string]bool, localFns map[string]*program.Function) bool {
	if fn == nil || len(fn.Returns) == 0 {
		return false
	}
	localFns = anonFnDecls(fn)
	fnParams := make(map[string]ExpressionType, len(paramExprTypes))
	for k, v := range paramExprTypes {
		fnParams[k] = v
	}
	for _, assign := range fn.Assignments {
		if s.unlocksConstUnroll(assign.Expression, fnParams, visited, localFns) {
			return true
		}
		rhsType := s.getExpressionType(assign.Expression, fnParams, visited, localFns)
		for _, v := range assign.Variables {
			fnParams[v.Name] = rhsType
		}
	}
	for _, ret := range fn.Returns {
		if s.unlocksConstUnroll(ret, fnParams, visited, localFns) {
			return true
		}
	}
	return false
}
