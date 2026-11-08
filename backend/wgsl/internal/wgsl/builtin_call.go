package wgsl

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// BuiltinCall implements the [build.Expression] interface.
type BuiltinCall struct {
	Expression
	WGSLName string
	Args     []Expression
}

func NewBuiltinCall(op *program.Operand, fnName string, expressions []build.Expression) (Expression, error) {
	wgslBuiltin, overload, err := FindBuiltin(fnName, expressions)
	if err != nil {
		return nil, err
	}

	// Apparently numeric consts are considered to be abstract values
	// and cannot be automatically converted to vectors if they don't have a type.
	// In that case, we need to convert them to the correct type.
	args := make([]Expression, len(expressions))
	for i, expr := range expressions {
		args[i], err = expr.(Expression).ToCommon(overload.ParamTypes[i])
		if err != nil {
			return nil, err
		}
	}

	b := &BuiltinCall{
		WGSLName: wgslBuiltin.WGSLName,
		Args:     args,
	}
	b.Expression = NewExpr(op, overload.ReturnType, build.PrecedenceHighest)
	b.Expression.SetOuter(b)
	return b, nil
}

// NewCustomBuiltinCall constructs a BuiltinCall with an explicit function name and return type,
// for WGSL built-ins (such as transpose and determinant) that do not originate from GSL language built-ins.
func NewCustomBuiltinCall(op *program.Operand, name string, expressions []build.Expression, returnType TypeInfo) Expression {
	args := make([]Expression, len(expressions))
	for i, expr := range expressions {
		args[i] = expr.(Expression)
	}
	b := &BuiltinCall{
		WGSLName: name,
		Args:     args,
	}
	b.Expression = NewExpr(op, returnType, build.PrecedenceHighest)
	b.Expression.SetOuter(b)
	return b
}

func (b *BuiltinCall) Codegen(ctx CodegenContext) string {
	argSrcs := make([]string, 0, len(b.Args))
	for _, arg := range b.Args {
		argSrcs = append(argSrcs, arg.Codegen(ctx))
	}

	return fmt.Sprintf("%s(%s)", b.WGSLName, strings.Join(argSrcs, ", "))
}

func (b *BuiltinCall) Reads() []string {
	return dedupReads(b.Args...)
}
