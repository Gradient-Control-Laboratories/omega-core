package codegen

import (
	"github.com/Gradient-Control-Laboratories/omega-core/backend/wgsl/internal/wgsl"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/build"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

func (d *Delegate) BuildCollect(op *program.Operand, _ *program.Expression,
	_ []*program.Operand, expressions []build.Expression) (build.Expression, error) {

	return wgsl.NewCollect(op, expressions, nil)
}

func (d *Delegate) BuildConstCollect(op *program.Operand, _ *program.Expression,
	_ []*program.Operand, expressions []build.Expression, constValues []float64) (build.ConstExpression, error) {

	if err := finiteConstants(d, op, constValues...); err != nil {
		return nil, err
	}
	return wgsl.NewCollect(op, expressions, constValues)
}
