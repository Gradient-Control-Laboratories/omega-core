package meshplan

import (
	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/program"
)

// inlineFnCall inlines single-expression functions returning mesh calls or SDFs.
func inlineFnCall(fc *program.FnCall, module *program.Package) *program.Operand {
	if fc == nil || fc.Reference == nil || fc.Reference.Prefix != "" {
		return nil
	}
	fn := moduleFunction(module, fc.Reference.Name, len(fc.Arguments))
	if fn == nil || len(fn.Returns) != 1 || len(fn.Assignments) != 0 || len(fn.FnTypes) != 0 {
		return nil
	}
	bindings := make(map[string]*program.Operand, len(fn.Parameters))
	for i, p := range fn.Parameters {
		if i < len(fc.Arguments) {
			bindings[p.Name] = fc.Arguments[i]
		}
	}
	return substituteRefs(fn.Returns[0], bindings)
}

// moduleFunction finds a module-local function declaration with the given
// name and parameter count.
func moduleFunction(module *program.Package, name string, paramCount int) *program.Function {
	if module == nil {
		return nil
	}
	for _, f := range module.Files {
		for _, ft := range f.FnTypes {
			if ft.Name != name {
				continue
			}
			for _, g := range ft.Groups {
				if int(g.ParamCount) == paramCount {
					if len(g.Functions) > 0 {
						return g.Functions[0]
					}
				}
			}
		}
	}
	return nil
}

// substituteRefs returns a deep copy of op with references to the bound
// names replaced by copies of the bound operands.
func substituteRefs(op *program.Operand, bindings map[string]*program.Operand) *program.Operand {
	if op == nil {
		return nil
	}
	cloned := proto.Clone(op).(*program.Operand)
	substituteRefsInPlace(cloned, bindings)
	return cloned
}

func substituteRefsInPlace(op *program.Operand, bindings map[string]*program.Operand) {
	if op == nil {
		return
	}
	switch v := op.T.(type) {
	case *program.Operand_Reference:
		if v.Reference != nil {
			if sub, ok := bindings[v.Reference.Name]; ok {
				op.T = proto.Clone(sub).(*program.Operand).T
			}
		}
	case *program.Operand_Expression:
		expr := v.Expression
		if expr == nil {
			return
		}
		operands := expr.Operands
		// An ACCESS member (v.x) and a map's function are name references
		// that must never be substituted as values.
		switch expr.Operation {
		case program.Operation_ACCESS:
			if len(operands) > 1 {
				operands = operands[:1]
			}
		case program.Operation_MAP:
			if len(operands) > 1 {
				operands = operands[1:]
			}
		}
		for _, sub := range operands {
			substituteRefsInPlace(sub, bindings)
		}
	case *program.Operand_FnCall:
		if v.FnCall == nil {
			return
		}
		for _, arg := range v.FnCall.Arguments {
			substituteRefsInPlace(arg, bindings)
		}
	}
}
