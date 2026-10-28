package inline

import (
	"strings"

	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/program"
)

// hardwareAccelPkgs registers packages whose operations are accelerated by
// downstream hardware backends (such as WGSL native vectors/matrices) and
// therefore must not be inlined.
var hardwareAccelPkgs = map[string]bool{}

// remapVisit is a parameter remapsParam has asked about, so a recursive
// function is not followed forever.
type remapVisit struct {
	fn    *program.Function
	param int
}

// constEval memoizes one constant evaluation by operand. References are
// followed into the expressions that define them, and a property expansion
// writes its shared nodes as a chain of locals each naming earlier ones more
// than once (property/share.go), so walked as a tree the chain costs 2^n:
// .second of Cracked Mud never finished assembling. Memoized it is linear in
// the DAG. Nothing is rewritten while one evaluation runs, so its answers hold
// for its length.
type constEval struct {
	cond   map[*program.Operand][2]bool
	scalar map[*program.Operand]constScalar
}

type constScalar struct {
	v  float64
	ok bool
}

func newConstEval() *constEval {
	return &constEval{cond: map[*program.Operand][2]bool{}, scalar: map[*program.Operand]constScalar{}}
}

func (in *inliner) shouldInlineCall(fc *program.FnCall, fn *program.Function, calleePkgPath string, currentPkgPath string, fileScope *scope.PkgFileScope) bool {
	if fn == nil || len(fn.Returns) != 1 {
		return false
	}
	if fc.Reference == nil {
		return false
	}
	if strings.HasPrefix(fc.Reference.Name, program.AnonFnPrefix) || strings.Contains(fc.Reference.Name, "_anon_") || strings.HasPrefix(fc.Reference.Name, program.InlinedFnPrefix) {
		return true
	}

	if in.remapsFieldArg(fc, fn, fileScope, calleePkgPath) {
		return true
	}
	if in.isAllConstArgs(fc) {
		// Do not inline standard library operations backed by hardware acceleration sidecars.
		// Inlining would expand their GSL loop/map bodies into closures, breaking hardware acceleration.
		if hardwareAccelPkgs[calleePkgPath] {
			return false
		}
		// Do not inline text and font rendering functions that directly sample p or time.
		// Even when called with constant string or styling arguments, inlining complex glyph SDF
		// evaluation duplicates shader logic and prevents backends from emitting standalone routines.
		if calleePkgPath == "text" && in.directlyUsesPOrTime(fn) {
			return false
		}
		if in.hasPOrTime(fn, fileScope) {
			if in.callCounts != nil {
				count := in.callCounts[fc.Reference.Name]
				if fc.Reference.Prefix != "" {
					count = in.callCounts[fc.Reference.Prefix+"."+fc.Reference.Name]
				}
				if count > 1 {
					return false
				}
			}
		}
		return true
	}
	return false
}

// TODO: remove unused param
func (in *inliner) shouldInlineMultiReturnCall(fc *program.FnCall, fn *program.Function, calleePkgPath string, currentPkgPath string, fileScope *scope.PkgFileScope) bool {
	if fn == nil || len(fn.Returns) < 2 {
		return false
	}
	if fc.Reference == nil {
		return false
	}
	if in.isAllConstArgs(fc) {
		if in.hasPOrTime(fn, fileScope) {
			if in.callCounts != nil {
				count := in.callCounts[fc.Reference.Name]
				if fc.Reference.Prefix != "" {
					count = in.callCounts[fc.Reference.Prefix+"."+fc.Reference.Name]
				}
				if count > 1 {
					return false
				}
			}
		}
		return true
	}
	return false
}

// directlyUsesPOrTime reports whether fn directly references the evaluation point (p)
// or time within its own assignments or returns without following callee functions.
func (in *inliner) directlyUsesPOrTime(fn *program.Function) bool {
	if fn == nil {
		return false
	}
	for _, a := range fn.Assignments {
		if in.operandDirectlyUsesPOrTime(a.Expression) {
			return true
		}
	}
	for _, r := range fn.Returns {
		if in.operandDirectlyUsesPOrTime(r) {
			return true
		}
	}
	return false
}

// operandDirectlyUsesPOrTime reports whether op directly references p, time, or pixelSize.
func (in *inliner) operandDirectlyUsesPOrTime(op *program.Operand) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		if t.Reference.Prefix == "" && (t.Reference.Name == program.PVarName ||
			t.Reference.Name == program.TimeVarName || t.Reference.Name == program.PixelSize) {
			return true
		}
	case *program.Operand_PropRef:
		if t.PropRef != nil && t.PropRef.Reference != nil && t.PropRef.Reference.Prefix == "" &&
			(t.PropRef.Reference.Name == program.PVarName ||
				t.PropRef.Reference.Name == program.TimeVarName ||
				t.PropRef.Reference.Name == program.PixelSize) {
			return true
		}
	case *program.Operand_Expression:
		for _, child := range t.Expression.Operands {
			if in.operandDirectlyUsesPOrTime(child) {
				return true
			}
		}
	case *program.Operand_FnCall:
		if t.FnCall.Reference != nil && t.FnCall.Reference.Prefix == "" &&
			t.FnCall.Reference.Name == program.PixelSize {
			return true
		}
		for _, arg := range t.FnCall.Arguments {
			if in.operandDirectlyUsesPOrTime(arg) {
				return true
			}
		}
	}
	return false
}

// pOrTimeKey is hasPOrTime's cache key: the function asked about and the
// innermost function being transformed when it was asked.
type pOrTimeKey struct {
	fn, enclosing *program.Function
}

func (in *inliner) hasPOrTime(fn *program.Function, fileScope *scope.PkgFileScope) bool {
	if fn == nil {
		return false
	}
	// The answer depends on the functions being transformed around this one:
	// findCallbackFunction resolves a callback by name through them first. So
	// it is cached for that innermost function too, and a lambda of one
	// function never decides the answer for another.
	key := pOrTimeKey{fn: fn}
	if n := len(in.enclosingFns); n > 0 {
		key.enclosing = in.enclosingFns[n-1]
	}
	if in.pOrTimeCache != nil {
		if val, ok := in.pOrTimeCache[key]; ok {
			return val
		}
	}
	visited := make(map[*program.Function]bool)
	res := in.hasPOrTimeHelper(fn, fileScope, visited)
	if in.pOrTimeCache != nil {
		in.pOrTimeCache[key] = res
	}
	return res
}

func (in *inliner) hasPOrTimeHelper(fn *program.Function, fileScope *scope.PkgFileScope, visited map[*program.Function]bool) bool {
	if fn == nil || visited[fn] {
		return false
	}
	visited[fn] = true

	for _, assign := range fn.Assignments {
		if in.operandHasPOrTime(assign.Expression, fileScope, visited) {
			return true
		}
	}
	for _, ret := range fn.Returns {
		if in.operandHasPOrTime(ret, fileScope, visited) {
			return true
		}
	}
	return false
}

func (in *inliner) operandHasPOrTime(op *program.Operand, fileScope *scope.PkgFileScope, visited map[*program.Function]bool) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		if t.Reference.Prefix == "" && (t.Reference.Name == program.PVarName ||
			t.Reference.Name == program.TimeVarName || t.Reference.Name == program.PixelSize) {
			return true
		}
	case *program.Operand_Expression:
		if t.Expression.Operation == program.Operation_REDUCE || t.Expression.Operation == program.Operation_MAP {
			cbIdx := 0
			if t.Expression.Operation == program.Operation_REDUCE && len(t.Expression.Operands) >= 3 {
				cbIdx = 1
			}
			if cbIdx < len(t.Expression.Operands) {
				cbFn := in.findCallbackFunction(t.Expression.Operands[cbIdx], fileScope)
				if cbFn != nil && in.hasPOrTimeHelper(cbFn, fileScope, visited) {
					return true
				}
			}
		}
		for _, child := range t.Expression.Operands {
			if in.operandHasPOrTime(child, fileScope, visited) {
				return true
			}
		}
	case *program.Operand_FnCall:
		fc := t.FnCall
		if fc.Reference != nil && fc.Reference.Prefix == "" && fc.Reference.Name == program.PixelSize {
			return true
		}
		for _, arg := range fc.Arguments {
			if in.operandHasPOrTime(arg, fileScope, visited) {
				return true
			}
		}
		if !in.isNativeOrSpecial(fc) {
			_, calleeFn, _ := in.findFunction(fc, fileScope, len(fc.Arguments))
			if calleeFn != nil && in.hasPOrTimeHelper(calleeFn, fileScope, visited) {
				return true
			}
		}
	}
	return false
}

func (in *inliner) isAllConstArgs(fc *program.FnCall) bool {
	for _, arg := range fc.Arguments {
		if !in.isConstOperand(arg) || in.callsPOrTime(arg, in.currentFileScope) {
			return false
		}
	}
	return true
}

func (in *inliner) isConstOperand(op *program.Operand) bool {
	if op == nil {
		return false
	}
	if in.callsPOrTime(op, in.currentFileScope) {
		return false
	}
	if len(op.ConstValues) > 0 {
		return true
	}
	switch t := op.T.(type) {
	case *program.Operand_Value:
		return true
	case *program.Operand_Reference:
		if t.Reference != nil && t.Reference.Prefix == "" {
			return in.isConstRef(t.Reference.Name)
		}
	case *program.Operand_Expression:
		if t.Expression.Operation == program.Operation_REDUCE {
			cbIdx := 0
			if len(t.Expression.Operands) >= 3 {
				cbIdx = 1
			}
			for i, child := range t.Expression.Operands {
				if i == cbIdx {
					continue
				}
				if !in.isConstOperand(child) {
					return false
				}
			}
			return true
		} else if t.Expression.Operation == program.Operation_MAP {
			for i, child := range t.Expression.Operands {
				if i == 0 {
					continue
				}
				if !in.isConstOperand(child) {
					return false
				}
			}
			return true
		}
		for _, child := range t.Expression.Operands {
			if !in.isConstOperand(child) {
				return false
			}
		}
		return true
	case *program.Operand_FnCall:
		if t.FnCall.Reference != nil && t.FnCall.Reference.Prefix == "" &&
			t.FnCall.Reference.Name == program.PixelSize {
			return false
		}
		for _, arg := range t.FnCall.Arguments {
			if !in.isConstOperand(arg) {
				return false
			}
		}
		return true
	}
	return false
}

// isHoistableConst reports whether op can be copied into a function hoisted to
// file level: a constant that names nothing local to the enclosing function.
//
// isConstOperand is not that test. It also accepts a reference whose value
// constant folding has worked out, which can be a LOCAL of the enclosing
// function — a closure's baked capture, `major = 2` — and that is right for an
// expression substituted in place, where the local is in scope. A nested
// function of the callee is hoisted to file level instead, outside the
// enclosing function, and the local does not resolve there: inlining
// tab.lookup(major), whose table fold captures its argument, wrote `major`
// into the hoisted lambda and failed to link. So a constant also has to name
// no local of a function it is inside.
func (in *inliner) isHoistableConst(op *program.Operand) bool {
	return in.isConstOperand(op) && !in.namesEnclosingLocal(op)
}

// namesEnclosingLocal reports whether op refers, unqualified, to a parameter
// or assignment of a function being transformed — or to a temporary about to
// be hoisted into one — none of which a file-level function can see.
func (in *inliner) namesEnclosingLocal(op *program.Operand) bool {
	if op == nil || len(in.enclosingFns) == 0 {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		return t.Reference != nil && t.Reference.Prefix == "" && in.isEnclosingLocal(t.Reference.Name)
	case *program.Operand_Expression:
		for _, child := range t.Expression.Operands {
			if in.namesEnclosingLocal(child) {
				return true
			}
		}
	case *program.Operand_FnCall:
		for _, arg := range t.FnCall.Arguments {
			if in.namesEnclosingLocal(arg) {
				return true
			}
		}
	}
	return false
}

func (in *inliner) isEnclosingLocal(name string) bool {
	for _, fn := range in.enclosingFns {
		for _, p := range fn.Parameters {
			if p.Name == name {
				return true
			}
		}
		for _, a := range fn.Assignments {
			for _, v := range a.Variables {
				if v.Name == name {
					return true
				}
			}
		}
	}
	for _, a := range in.hoistedAssignments {
		for _, v := range a.Variables {
			if v.Name == name {
				return true
			}
		}
	}
	return false
}

func (in *inliner) isConstRef(name string) bool {
	if in.constRefMemo == nil {
		in.constRefMemo = make(map[string]bool)
	}
	if val, ok := in.constRefMemo[name]; ok {
		return val
	}
	if in.currentFileScope != nil {
		elem := in.currentFileScope.NamedElem(name)
		if v, ok := elem.(*scope.Var); ok && v != nil && v.Parent != nil && v.Parent.Assign != nil {
			in.constRefMemo[name] = false // guard against recursion
			isConst := in.isConstOperand(v.Parent.Assign.Expression)
			in.constRefMemo[name] = isConst
			return isConst
		}
	}
	return false
}

// isCopyableConst reports whether a constant argument or local can be copied
// into each place it is read. isConstOperand also accepts a call whose
// arguments are constant, but when the called function reads p or time, as a
// texture does, every copy evaluates it again: color.mixOklab reads its input
// nine times, so the texture ran once per read. Such a call is bound once
// instead. p itself, read directly, is cheap to copy and still is.
func (in *inliner) isCopyableConst(op *program.Operand, fileScope *scope.PkgFileScope) bool {
	return in.isConstOperand(op) && !in.callsPOrTime(op, fileScope)
}

func (in *inliner) isSimpleOperand(op *program.Operand) bool {
	if op == nil {
		return true
	}
	switch op.T.(type) {
	case *program.Operand_Value, *program.Operand_Reference:
		return true
	}
	return false
}

// evalConstCondition reports whether op is a condition known at compile time,
// and its value.
func (in *inliner) evalConstCondition(op *program.Operand) (bool, bool) {
	return in.evalConstConditionMemo(op, newConstEval())
}

func (in *inliner) evalConstConditionMemo(op *program.Operand, m *constEval) (bool, bool) {
	if op == nil {
		return false, false
	}
	if r, ok := m.cond[op]; ok {
		return r[0], r[1]
	}
	isConst, val := in.evalConstConditionOnce(op, m)
	m.cond[op] = [2]bool{isConst, val}
	return isConst, val
}

func (in *inliner) evalConstConditionOnce(op *program.Operand, m *constEval) (bool, bool) {
	if len(op.ConstValues) > 0 {
		return true, op.ConstValues[0] != 0
	}
	switch t := op.T.(type) {
	case *program.Operand_Value:
		return true, t.Value != 0
	case *program.Operand_Reference:
		if t.Reference.Prefix == "" && in.currentFileScope != nil {
			elem := in.currentFileScope.NamedElem(t.Reference.Name)
			if v, ok := elem.(*scope.Var); ok && v != nil && v.Parent != nil && v.Parent.Assign != nil {
				return in.evalConstConditionMemo(v.Parent.Assign.Expression, m)
			}
		}
	case *program.Operand_Expression:
		expr := t.Expression
		switch expr.Operation {
		case program.Operation_LT, program.Operation_GT, program.Operation_LE, program.Operation_GE, program.Operation_EQ, program.Operation_NE:
			if len(expr.Operands) == 2 {
				val1, ok1 := in.evalConstScalarMemo(expr.Operands[0], m)
				val2, ok2 := in.evalConstScalarMemo(expr.Operands[1], m)
				if ok1 && ok2 {
					switch expr.Operation {
					case program.Operation_LT:
						return true, val1 < val2
					case program.Operation_GT:
						return true, val1 > val2
					case program.Operation_LE:
						return true, val1 <= val2
					case program.Operation_GE:
						return true, val1 >= val2
					case program.Operation_EQ:
						return true, val1 == val2
					case program.Operation_NE:
						return true, val1 != val2
					}
				}
			}
		case program.Operation_NOT:
			if len(expr.Operands) == 1 {
				if ok, b := in.evalConstConditionMemo(expr.Operands[0], m); ok {
					return true, !b
				}
			}
		case program.Operation_AND:
			if len(expr.Operands) == 2 {
				ok1, b1 := in.evalConstConditionMemo(expr.Operands[0], m)
				ok2, b2 := in.evalConstConditionMemo(expr.Operands[1], m)
				if ok1 && ok2 {
					return true, b1 && b2
				}
			}
		case program.Operation_OR:
			if len(expr.Operands) == 2 {
				ok1, b1 := in.evalConstConditionMemo(expr.Operands[0], m)
				ok2, b2 := in.evalConstConditionMemo(expr.Operands[1], m)
				if ok1 && ok2 {
					return true, b1 || b2
				}
			}
		}
	}
	return false, false
}

func (in *inliner) evalConstScalarMemo(op *program.Operand, m *constEval) (float64, bool) {
	if op == nil {
		return 0, false
	}
	if r, ok := m.scalar[op]; ok {
		return r.v, r.ok
	}
	v, ok := in.evalConstScalarOnce(op, m)
	m.scalar[op] = constScalar{v, ok}
	return v, ok
}

func (in *inliner) evalConstScalarOnce(op *program.Operand, m *constEval) (float64, bool) {
	if len(op.ConstValues) > 0 {
		return op.ConstValues[0], true
	}
	switch t := op.T.(type) {
	case *program.Operand_Value:
		return t.Value, true
	case *program.Operand_Reference:
		if t.Reference.Prefix == "" && in.currentFileScope != nil {
			elem := in.currentFileScope.NamedElem(t.Reference.Name)
			if v, ok := elem.(*scope.Var); ok && v != nil && v.Parent != nil && v.Parent.Assign != nil {
				return in.evalConstScalarMemo(v.Parent.Assign.Expression, m)
			}
		}
	case *program.Operand_Expression:
		expr := t.Expression
		if len(expr.Operands) == 2 {
			v1, ok1 := in.evalConstScalarMemo(expr.Operands[0], m)
			v2, ok2 := in.evalConstScalarMemo(expr.Operands[1], m)
			if ok1 && ok2 {
				switch expr.Operation {
				case program.Operation_ADD:
					return v1 + v2, true
				case program.Operation_SUB:
					return v1 - v2, true
				case program.Operation_MUL:
					return v1 * v2, true
				case program.Operation_DIV:
					if v2 != 0 {
						return v1 / v2, true
					}
				}
			}
		}
	}
	return 0, false
}

// callsPOrTime reports whether op calls a function whose body reads p or time.
func (in *inliner) callsPOrTime(op *program.Operand, fileScope *scope.PkgFileScope) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		if t.Reference.Prefix == "" && (t.Reference.Name == program.PVarName ||
			t.Reference.Name == program.TimeVarName || t.Reference.Name == program.PixelSize) {
			return true
		}
	case *program.Operand_Expression:
		if t.Expression.Operation == program.Operation_REDUCE || t.Expression.Operation == program.Operation_MAP {
			cbIdx := 0
			if t.Expression.Operation == program.Operation_REDUCE && len(t.Expression.Operands) >= 3 {
				cbIdx = 1
			}
			if cbIdx < len(t.Expression.Operands) {
				cbFn := in.findCallbackFunction(t.Expression.Operands[cbIdx], fileScope)
				if cbFn != nil && in.hasPOrTime(cbFn, fileScope) {
					return true
				}
			}
		}
		for _, child := range t.Expression.Operands {
			if in.callsPOrTime(child, fileScope) {
				return true
			}
		}
	case *program.Operand_FnCall:
		fc := t.FnCall
		if fc.Reference != nil && fc.Reference.Prefix == "" && fc.Reference.Name == program.PixelSize {
			return true
		}
		for _, arg := range fc.Arguments {
			if in.callsPOrTime(arg, fileScope) {
				return true
			}
		}
		if !in.isNativeOrSpecial(fc) {
			_, calleeFn, _ := in.findFunction(fc, fileScope, len(fc.Arguments))
			if calleeFn != nil && in.hasPOrTime(calleeFn, fileScope) {
				return true
			}
		}
	}
	return false
}

// operandUsesP reports whether op depends on p, through the declarations it
// names and the functions it calls. Unlike operandHasPOrTime it leaves time out:
// a value that varies only in time is the same at every point, so a remap has
// nothing to re-evaluate.
func (in *inliner) operandUsesP(op *program.Operand, fileScope *scope.PkgFileScope, visited map[*program.Function]bool, seen map[string]bool) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		if t.Reference.Prefix != "" {
			return false
		}
		name := t.Reference.Name
		if name == program.PVarName {
			return true
		}
		if seen[name] || fileScope == nil {
			return false
		}
		seen[name] = true
		if v, ok := fileScope.NamedElem(name).(*scope.Var); ok && v != nil && v.Parent != nil && v.Parent.Assign != nil {
			return in.operandUsesP(v.Parent.Assign.Expression, fileScope, visited, seen)
		}
	case *program.Operand_Expression:
		if t.Expression.Operation == program.Operation_REDUCE || t.Expression.Operation == program.Operation_MAP {
			cbIdx := 0
			if t.Expression.Operation == program.Operation_REDUCE && len(t.Expression.Operands) >= 3 {
				cbIdx = 1
			}
			if cbIdx < len(t.Expression.Operands) {
				cbFn := in.findCallbackFunction(t.Expression.Operands[cbIdx], fileScope)
				if cbFn != nil && !visited[cbFn] {
					visited[cbFn] = true
					for _, a := range cbFn.Assignments {
						if in.operandUsesP(a.Expression, fileScope, visited, seen) {
							return true
						}
					}
					for _, r := range cbFn.Returns {
						if in.operandUsesP(r, fileScope, visited, seen) {
							return true
						}
					}
				}
			}
		}
		for _, child := range t.Expression.Operands {
			if in.operandUsesP(child, fileScope, visited, seen) {
				return true
			}
		}
	case *program.Operand_FnCall:
		fc := t.FnCall
		for _, arg := range fc.Arguments {
			if in.operandUsesP(arg, fileScope, visited, seen) {
				return true
			}
		}
		// A builtin package function (sphere) has a body, which may read p; only
		// a native builtin, remap and the like have none to look into.
		if !in.isNativeOrSpecial(fc) {
			_, calleeFn, _ := in.findFunction(fc, fileScope, len(fc.Arguments))
			if calleeFn != nil && !visited[calleeFn] {
				visited[calleeFn] = true
				for _, a := range calleeFn.Assignments {
					if in.operandUsesP(a.Expression, fileScope, visited, seen) {
						return true
					}
				}
				for _, r := range calleeFn.Returns {
					if in.operandUsesP(r, fileScope, visited, seen) {
						return true
					}
				}
			}
		}
	}
	return false
}

// remapsFieldArg reports whether the call hands fn a field that fn remaps:
// remap(shape, q) where shape is a parameter (or a local computed from one) and
// the argument for it depends on p. A remap evaluates its target again at q,
// which needs the argument's expression; a backend that emits fn as a function
// of its own receives the parameter as a value and can only return it,
// evaluated at the caller's p. So such a call is inlined, as fidget
// effectively does already, however often fn is called. An argument that does
// not depend on p is the same at any point, and keeps the call. fn lives in
// calleePkgPath, whose imports resolve the calls in its body.
func (in *inliner) remapsFieldArg(fc *program.FnCall, fn *program.Function, fileScope *scope.PkgFileScope, calleePkgPath string) bool {
	for i, param := range fn.Parameters {
		if i >= len(fc.Arguments) ||
			!in.operandUsesP(fc.Arguments[i], fileScope, map[*program.Function]bool{}, map[string]bool{}) {
			continue
		}
		if in.remapsParam(fn, param.Name, in.fileScopeOf(fn, calleePkgPath), map[remapVisit]bool{}) {
			return true
		}
	}
	return false
}

// inlinesRemapTarget reports whether op, a call, is a remap's target and hands
// its callee a field: an argument that depends on p. A backend that emits the
// callee as a function of its own passes it the remapped point, but the field
// is an argument built against the caller's, so remap(f(length(p) - 1), q)
// would move f and leave the sphere where it was. Inlined, the target is an
// expression the remap re-derives whole. op stays the target as it is
// inlined, so a wrapper keeps inlining until the target is no longer a call.
func (in *inliner) inlinesRemapTarget(op *program.Operand, fc *program.FnCall, fileScope *scope.PkgFileScope) bool {
	if !in.remapTargets[op] {
		return false
	}
	for _, arg := range fc.Arguments {
		if in.operandUsesP(arg, fileScope, map[*program.Function]bool{}, map[string]bool{}) {
			return true
		}
	}
	return false
}

// remapsParam reports whether fn remaps its parameter name, directly, through a
// local computed from it, or by handing it on to a function that does. fn is
// declared in fileScope.
func (in *inliner) remapsParam(fn *program.Function, name string, fileScope *scope.PkgFileScope, visited map[remapVisit]bool) bool {
	fromParam := map[string]bool{name: true}
	for _, a := range fn.Assignments {
		if a == nil || !in.usesAny(a.Expression, fromParam) {
			continue
		}
		for _, v := range a.Variables {
			if v != nil {
				fromParam[v.Name] = true
			}
		}
	}
	for _, a := range fn.Assignments {
		if a != nil && in.remapsAny(a.Expression, fromParam, fileScope, visited) {
			return true
		}
	}
	for _, ret := range fn.Returns {
		if in.remapsAny(ret, fromParam, fileScope, visited) {
			return true
		}
	}
	return false
}

// remapsAny reports whether op contains a remap whose target uses a name in
// names, or passes such a name to a function that remaps that parameter: a
// wrapper place(shape, o) { return transform.translate(shape, o); } remaps
// shape as surely as the remap in translate does. Calls resolve in fileScope.
func (in *inliner) remapsAny(op *program.Operand, names map[string]bool, fileScope *scope.PkgFileScope, visited map[remapVisit]bool) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		if t.Expression.Operation == program.Operation_REMAP && len(t.Expression.Operands) > 0 &&
			in.usesAny(t.Expression.Operands[0], names) {
			return true
		}
		for _, child := range t.Expression.Operands {
			if in.remapsAny(child, names, fileScope, visited) {
				return true
			}
		}
	case *program.Operand_FnCall:
		fc := t.FnCall
		for i, arg := range fc.Arguments {
			if in.remapsAny(arg, names, fileScope, visited) {
				return true
			}
			if in.isBuiltinOrSpecial(fc) || !in.usesAny(arg, names) {
				continue
			}
			_, callee, calleePkgPath := in.findFunction(fc, fileScope, len(fc.Arguments))
			if callee == nil || i >= len(callee.Parameters) {
				continue
			}
			key := remapVisit{callee, i}
			if visited[key] {
				continue
			}
			visited[key] = true
			if in.remapsParam(callee, callee.Parameters[i].Name, in.fileScopeOf(callee, calleePkgPath), visited) {
				return true
			}
		}
	}
	return false
}

// usesAny reports whether op references any name in names.
func (in *inliner) usesAny(op *program.Operand, names map[string]bool) bool {
	for name := range names {
		if in.countUsesInOperand(op, name) > 0 {
			return true
		}
	}
	return false
}

func (in *inliner) countUsesInOperand(op *program.Operand, paramName string) int {
	if op == nil {
		return 0
	}
	count := 0
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		if t.Reference.Prefix == "" && t.Reference.Name == paramName {
			count++
		}
	case *program.Operand_PropRef:
		if t.PropRef != nil && t.PropRef.Reference != nil && t.PropRef.Reference.Prefix == "" && t.PropRef.Reference.Name == paramName {
			count++
		}
	case *program.Operand_Expression:
		for _, child := range t.Expression.Operands {
			count += in.countUsesInOperand(child, paramName)
		}
	case *program.Operand_FnCall:
		for _, arg := range t.FnCall.Arguments {
			count += in.countUsesInOperand(arg, paramName)
		}
	}
	return count
}
