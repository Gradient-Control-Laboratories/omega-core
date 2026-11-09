package wgsl

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// Mod implements the [build.Expression] interface.
type Mod struct {
	Expression
	Left   Expression
	Right  Expression
	FnName string
}

func NewMod(
	op *program.Operand,
	expressions []build.Expression,
	resultType *program.Type,
	addPredefinedFunc func(name string, args []build.Expression, ret TypeInfo, src string) error,
) (Expression, error) {
	rType := WGSLType(resultType, false)
	if rType.Name == "" {
		rType = TypeF32
	}

	// The gsl_floored helpers take both operands at the result type, so a
	// scalar operand mixed with a vector one (vec % scalar broadcasts in GSL)
	// must be splatted — otherwise the generated call is invalid WGSL.
	leftCommon, err := expressions[0].(Expression).ToNum().ToCommon(rType)
	if err != nil {
		return nil, err
	}
	rightCommon, err := expressions[1].(Expression).ToNum().ToCommon(rType)
	if err != nil {
		return nil, err
	}
	leftWgslExpr := leftCommon.ParenIf(build.PrecedenceMultiplication)
	rightWgslExpr := rightCommon.ParenIf(build.PrecedenceMultiplication)

	var fnName string
	switch rType.Name {
	case TypeF32.Name:
		fnName = "gsl_floored"
		err = addPredefinedFunc(fnName, flooredNumArgs, TypeF32, flooredNumSrc)
	case TypeVec2.Name:
		fnName = "gsl_floored_vec2"
		err = addPredefinedFunc(fnName, flooredVec2Args, TypeVec2, flooredVec2Src)
	case TypeVec3.Name:
		fnName = "gsl_floored_vec3"
		err = addPredefinedFunc(fnName, flooredVec3Args, TypeVec3, flooredVec3Src)
	case TypeVec4.Name:
		fnName = "gsl_floored_vec4"
		err = addPredefinedFunc(fnName, flooredVec4Args, TypeVec4, flooredVec4Src)
	}
	if err != nil {
		return nil, err
	}

	m := &Mod{
		Left:   leftWgslExpr,
		Right:  rightWgslExpr,
		FnName: fnName,
	}
	m.Expression = NewExpr(op, rType, build.PrecedenceHighest)
	m.Expression.SetOuter(m)
	return m, nil
}

func (m *Mod) Codegen(ctx CodegenContext) string {
	return fmt.Sprintf("%s(%s, %s)", m.FnName, m.Left.Codegen(ctx), m.Right.Codegen(ctx))
}

func (m *Mod) Reads() []string {
	return dedupReads(m.Left, m.Right)
}

var flooredNumArgs = []build.Expression{
	NewOpaqueExpr(nil, TypeF32, build.PrecedenceHighest, "a"),
	NewOpaqueExpr(nil, TypeF32, build.PrecedenceHighest, "n"),
}

const flooredNumSrc = `
fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}`

var flooredVec2Args = []build.Expression{
	NewOpaqueExpr(nil, TypeVec2, build.PrecedenceHighest, "a"),
	NewOpaqueExpr(nil, TypeVec2, build.PrecedenceHighest, "n"),
}

const flooredVec2Src = `
fn gsl_floored_vec2(a: vec2<f32>, n: vec2<f32>) -> vec2<f32> {
  return a - n * floor(a / n);
}`

var flooredVec3Args = []build.Expression{
	NewOpaqueExpr(nil, TypeVec3, build.PrecedenceHighest, "a"),
	NewOpaqueExpr(nil, TypeVec3, build.PrecedenceHighest, "n"),
}

const flooredVec3Src = `
fn gsl_floored_vec3(a: vec3<f32>, n: vec3<f32>) -> vec3<f32> {
  return a - n * floor(a / n);
}`

var flooredVec4Args = []build.Expression{
	NewOpaqueExpr(nil, TypeVec4, build.PrecedenceHighest, "a"),
	NewOpaqueExpr(nil, TypeVec4, build.PrecedenceHighest, "n"),
}

const flooredVec4Src = `
fn gsl_floored_vec4(a: vec4<f32>, n: vec4<f32>) -> vec4<f32> {
  return a - n * floor(a / n);
}`
