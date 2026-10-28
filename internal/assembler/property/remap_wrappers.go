package property

import (
	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/program"
)

// A FUNCTION THAT ONLY MOVES ITS SHAPE IS A REMAP. transform.translate is
//
//	translate(shape, offset) { return remap(shape, p - offset); }
//
// and a document's own place(shape, offset) { return transform.translate(shape,
// offset); } is the same remap one call further out. A colour or material moves
// with the shape through a remap written out (colorMarker), but not through
// such a call: the Engine binds a function's parameters to values evaluated in
// the caller's frame, so the remap inside can no longer re-evaluate the shape
// at the moved point, and a library call is opaque to the colour gradient
// besides. So a part placed with transform.translate drew the colour and
// material its library gave it — a finished screw's thread — about the world's
// origin instead of its own.
//
// unwrapRemaps rewrites each call to such a function into the remap itself,
// remap(shape argument, coordinate with the other arguments in place of the
// parameters), before the fields are resolved. A wrapper of a wrapper unwraps
// too. A function qualifies when its whole body is that one return; the
// coordinate may read p, its parameters and builtins, and nothing from another
// package, which the calling package might not import.

// maxRemapUnwrap bounds how many wrappers deep a call is unwrapped.
const maxRemapUnwrap = 8

// unwrapRemaps returns op with every call to a remap wrapper beneath it
// replaced by its remap. op itself is never changed: a node with a rewritten
// child is copied.
func (r *resolver) unwrapRemaps(op *program.Operand) *program.Operand {
	return r.unwrapRemapsDepth(op, 0)
}

func (r *resolver) unwrapRemapsDepth(op *program.Operand, depth int) *program.Operand {
	if op == nil || depth > maxRemapUnwrap {
		return op
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		if t.Expression == nil {
			return op
		}
		var ops []*program.Operand
		for i, o := range t.Expression.Operands {
			u := r.unwrapRemapsDepth(o, depth)
			if u != o && ops == nil {
				ops = append([]*program.Operand(nil), t.Expression.Operands...)
			}
			if ops != nil {
				ops[i] = u
			}
		}
		if ops == nil {
			return op
		}
		e := proto.Clone(t.Expression).(*program.Expression)
		e.Operands = ops
		return &program.Operand{T: &program.Operand_Expression{Expression: e}}
	case *program.Operand_FnCall:
		fc := t.FnCall
		if fc == nil {
			return op
		}
		args := make([]*program.Operand, len(fc.Arguments))
		changed := false
		for i, a := range fc.Arguments {
			args[i] = r.unwrapRemapsDepth(a, depth)
			changed = changed || args[i] != a
		}
		if remap := r.remapOf(fc, args, depth); remap != nil {
			return remap
		}
		if !changed {
			return op
		}
		c := proto.Clone(fc).(*program.FnCall)
		c.Arguments = args
		return &program.Operand{T: &program.Operand_FnCall{FnCall: c}}
	}
	return op
}

// remapOf is the remap a call to a remap wrapper stands for, with args in
// place of its parameters, or nil when the called function is not one.
func (r *resolver) remapOf(fc *program.FnCall, args []*program.Operand, depth int) *program.Operand {
	fn := r.calledFunction(fc, len(args))
	if fn == nil || len(fn.Returns) != 1 || len(fn.Assignments) != 0 || len(fn.FnTypes) != 0 ||
		len(fn.Parameters) != len(args) {
		return nil
	}
	body := r.unwrapRemapsDepth(fn.Returns[0], depth+1)
	e, ok := body.T.(*program.Operand_Expression)
	if !ok || e.Expression == nil || e.Expression.Operation != program.Operation_REMAP || len(e.Expression.Operands) != 2 {
		return nil
	}
	params := make(map[string]*program.Operand, len(fn.Parameters))
	for i, prm := range fn.Parameters {
		params[prm.Name] = args[i]
	}
	target, ok := e.Expression.Operands[0].T.(*program.Operand_Reference)
	if !ok || target.Reference == nil || target.Reference.Prefix != "" || params[target.Reference.Name] == nil {
		return nil
	}
	coord := e.Expression.Operands[1]
	if !localCoordinate(coord, params) {
		return nil
	}
	return nary(program.Operation_REMAP, params[target.Reference.Name], substituteParams(coord, params))
}

// calledFunction resolves a call to the function it names: a qualified call in
// that import's namespace, an unqualified one among the package's own and
// dot-imported functions, by parameter count.
func (r *resolver) calledFunction(fc *program.FnCall, arity int) *program.Function {
	if fc.Reference == nil || program.IsBuiltin(fc.Reference.Name) {
		return nil
	}
	funcs := r.overloads
	if fc.Reference.Prefix != "" {
		ns := r.imports[fc.Reference.Prefix]
		if ns == nil {
			return nil
		}
		funcs = ns.funcs
	}
	for _, fn := range funcs[fc.Reference.Name] {
		if len(fn.Parameters) == arity {
			return fn
		}
	}
	return nil
}

// localCoordinate reports whether a wrapper's coordinate reads only p (and x,
// y, z), its parameters, literals and builtins: nothing qualified, which would
// name an import of the wrapper's package, and no other name.
func localCoordinate(op *program.Operand, params map[string]*program.Operand) bool {
	if op == nil {
		return true
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		ref := t.Reference
		if ref == nil || ref.Prefix != "" {
			return false
		}
		switch ref.Name {
		case program.PVarName, program.XVarName, program.YVarName, program.ZVarName:
			return true
		}
		return params[ref.Name] != nil
	case *program.Operand_Expression:
		if t.Expression == nil {
			return true
		}
		for _, o := range t.Expression.Operands {
			if !localCoordinate(o, params) {
				return false
			}
		}
		return true
	case *program.Operand_FnCall:
		fc := t.FnCall
		if fc == nil || fc.Reference == nil || fc.Reference.Prefix != "" || !program.IsBuiltin(fc.Reference.Name) {
			return false
		}
		for _, a := range fc.Arguments {
			if !localCoordinate(a, params) {
				return false
			}
		}
		return true
	case *program.Operand_Value:
		return true
	}
	return false
}

// substituteParams is op with each reference to a parameter replaced by a copy
// of its argument.
func substituteParams(op *program.Operand, params map[string]*program.Operand) *program.Operand {
	if op == nil {
		return nil
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		if t.Reference != nil && t.Reference.Prefix == "" {
			if a, ok := params[t.Reference.Name]; ok {
				return proto.Clone(a).(*program.Operand)
			}
		}
		return op
	case *program.Operand_Expression:
		if t.Expression == nil {
			return op
		}
		e := proto.Clone(t.Expression).(*program.Expression)
		for i, o := range e.Operands {
			e.Operands[i] = substituteParams(o, params)
		}
		return &program.Operand{T: &program.Operand_Expression{Expression: e}}
	case *program.Operand_FnCall:
		if t.FnCall == nil {
			return op
		}
		c := proto.Clone(t.FnCall).(*program.FnCall)
		for i, a := range c.Arguments {
			c.Arguments[i] = substituteParams(a, params)
		}
		return &program.Operand{T: &program.Operand_FnCall{FnCall: c}}
	}
	return op
}
