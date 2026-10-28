package codegen

import (
	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) BuildCollect(op *program.Operand, _ *program.Expression,
	_ []*program.Operand, expressions []build.Expression) (build.Expression, error) {

	return wgsl.NewCollect(op, expressions, nil)
}

func (d *Delegate) BuildConstCollect(op *program.Operand, _ *program.Expression,
	_ []*program.Operand, expressions []build.Expression, constValues []float64) (build.ConstExpression, error) {

	return wgsl.NewCollect(op, expressions, constValues)
}
