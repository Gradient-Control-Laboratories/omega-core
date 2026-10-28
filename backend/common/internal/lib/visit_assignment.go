package build

import (
	"fmt"
	"gradientcontrol.com/omega/core/program"
)

func (b *Builder) visitAssignment(assign *program.Assignment, scope *Scope) error {
	names := make([]string, len(assign.Variables))
	for i, v := range assign.Variables {
		names[i] = v.Name
	}
	op := assign.Expression

	if len(names) > 1 {
		exprs, handled, err := b.tryInlineMultiReturn(op, names)
		if err != nil {
			return ErrLoc(scope, op.Start, err)
		}
		if handled {
			for i, name := range names {
				varInfo, err := b.delegate.BuildSingleAssignment(assign, op, name, exprs[i])
				if err != nil {
					return ErrLoc(scope, op.Start, err)
				}
				if err := scope.AddLocalVar(varInfo); err != nil {
					return ErrLoc(scope, op.Start, err)
				}
			}
			return nil
		}
	}

	expr, err := program.VisitOperand(op, b)
	if err != nil {
		return ErrLoc(scope, op.Start, err)
	}

	if len(names) > 1 && op.Type != nil && int(op.Type.Size) != len(names) && op.Type.ElemType == nil {
		return ErrLoc(scope, op.Start, fmt.Errorf("expected %d expressions, got %d", len(names), op.Type.Size))
	}

	if len(names) == 1 {
		if len(b.inlineBindings) > 0 {
			b.inlineBindings[len(b.inlineBindings)-1][names[0]] = expr
		}
		varInfo, err := b.delegate.BuildSingleAssignment(assign, op, names[0], expr)
		if err != nil {
			return ErrLoc(scope, op.Start, err)
		}

		err = scope.AddLocalVar(varInfo)
		if err != nil {
			return ErrLoc(scope, op.Start, err)
		}
		return nil
	}

	var fnCall *program.FnCall

	switch t := op.T.(type) {
	case *program.Operand_FnCall:
		fnCall = t.FnCall

	case *program.Operand_Expression:
		e := t.Expression
		if e.Operation != program.Operation_REMAP {
			err := fmt.Errorf("expected function call for multi variable assignment")
			return ErrLoc(scope, op.Start, err)
		}
		if len(e.Operands) > 0 {
			if fc, ok := e.Operands[0].T.(*program.Operand_FnCall); ok {
				fnCall = fc.FnCall
			}
		}

	case *program.Operand_Value, *program.Operand_Reference, *program.Operand_PropRef:
		err := fmt.Errorf("expected function call for multi variable assignment")
		return ErrLoc(scope, op.Start, err)

	default:
		err := fmt.Errorf("expected function call for multi variable assignment")
		return ErrLoc(scope, op.Start, err)
	}

	varInfos, err := b.delegate.BuildMultiAssignment(assign, op, fnCall, names, expr)
	if err != nil {
		return ErrLoc(scope, op.Start, err)
	}
	for _, varInfo := range varInfos {
		err = scope.AddLocalVar(varInfo)
		if err != nil {
			return ErrLoc(scope, op.Start, err)
		}
	}

	return nil
}
