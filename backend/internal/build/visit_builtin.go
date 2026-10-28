package build

import (
	"gradientcontrol.com/omega/core/program"
)

func (b *Builder) visitBuiltin(builtin *program.BuiltinType, op *program.Operand, fnCall *program.FnCall, args []Expression) (Expression, error) {

	result, err := b.delegate.BuildBuiltinCall(op, fnCall, builtin, args)
	if err != nil {
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
	}
	return result, nil
}
