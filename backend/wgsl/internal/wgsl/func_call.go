package wgsl

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// FuncCall implements the [build.Expression] interface.
type FuncCall struct {
	Expression
	WGSLName       string
	Args           []Expression
	PVarIsFirstArg bool
}

func NewFuncCall(op *program.Operand, wgslName string, args []Expression, pVarIsFirstArg bool, returnType TypeInfo) (Expression, error) {
	f := &FuncCall{
		WGSLName:       wgslName,
		Args:           args,
		PVarIsFirstArg: pVarIsFirstArg,
	}
	f.Expression = NewExpr(op, returnType, build.PrecedenceHighest)
	f.Expression.SetOuter(f)
	return f, nil
}

func (f *FuncCall) Codegen(ctx CodegenContext) string {
	argSrcs := make([]string, 0, len(f.Args))
	start := 0

	if f.PVarIsFirstArg {
		// The "p" variable expression that is passed as the first argument
		// is evaluated in terms of the "p" variable in the root context
		pExpr := ctx.P().Expr().(Expression)
		pVal := pExpr.Codegen(ctx.Root())
		if constExpr, ok := pExpr.(build.ConstExpression); ok && constExpr.ConstValues() != nil {
			pVal = program.PVarName + " - " + pVal
		}
		argSrcs = append(argSrcs, pVal)
		start = 1
	}

	for i := start; i < len(f.Args); i++ {
		argSrcs = append(argSrcs, f.Args[i].Codegen(ctx))
	}

	return fmt.Sprintf("%s(%s)", f.WGSLName, strings.Join(argSrcs, ", "))
}

func (f *FuncCall) Reads() []string {
	var list []Expression
	start := 0
	var reads []string
	seen := make(map[string]struct{})
	if f.PVarIsFirstArg {
		reads = append(reads, program.PVarName)
		seen[program.PVarName] = struct{}{}
		start = 1
	}
	for i := start; i < len(f.Args); i++ {
		list = append(list, f.Args[i])
	}
	for _, name := range dedupReads(list...) {
		if _, ok := seen[name]; !ok {
			seen[name] = struct{}{}
			reads = append(reads, name)
		}
	}
	return reads
}
