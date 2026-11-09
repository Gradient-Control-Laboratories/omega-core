package build

import (
	"math"

	"gradientcontrol.com/omega/core/program"
)

func (b *Builder) visitBuiltin(builtin *program.BuiltinType, op *program.Operand, fnCall *program.FnCall, args []Expression) (Expression, error) {

	// erfinv has no field form in any backend; it turns a constant (a solid
	// fraction, mostly) into a constant. A literal argument was folded before
	// the build, but a function parameter is only known to be constant here.
	// A non-constant argument falls through, for the backend to refuse.
	if builtin.Name == "erfinv" && allConstArgs(args) {
		vals := args[0].(ConstExpression).ConstValues()
		out := make([]float64, len(vals))
		for i, v := range vals {
			out[i] = math.Erfinv(v)
		}
		result, err := b.buildConstExpressionWithType(op, out, args[0].Type())
		if err != nil {
			return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
		}
		return result, nil
	}

	result, err := b.delegate.BuildBuiltinCall(op, fnCall, builtin, args)
	if err != nil {
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
	}
	return result, nil
}
