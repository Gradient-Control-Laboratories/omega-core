package wgsl

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// Comparison implements the [build.Expression] interface.
type Comparison struct {
	Expression
	Expr1  Expression
	Expr2  Expression
	Symbol string
}

func NewComparison(
	op *program.Operand,
	operation program.Operation,
	symbol string,
	expressions []build.Expression,
	resultType *program.Type,
	precedence build.Precedence,
) (Expression, error) {
	expr1 := expressions[0].(Expression)
	expr2 := expressions[1].(Expression)

	if operation != program.Operation_EQ && operation != program.Operation_NE {
		expr1 = expr1.ToNum().ParenIf(precedence)
		expr2 = expr2.ToNum()
		if expr2.Precedence() > precedence {
			expr2 = Paren(expr2)
		}
	} else {
		expr1 = expr1.ParenIf(precedence)
		if expr2.Precedence() > precedence {
			expr2 = Paren(expr2)
		}
	}

	boolType := WGSLType(resultType, true)

	c := &Comparison{
		Expr1:  expr1,
		Expr2:  expr2,
		Symbol: symbol,
	}
	c.Expression = NewExpr(op, boolType, precedence)
	c.Expression.SetOuter(c)
	return c, nil
}

func NewCompareRaw(op *program.Operand, expr1 Expression, symbol string, expr2 Expression, boolType TypeInfo) *Comparison {
	c := &Comparison{
		Expr1:  expr1,
		Expr2:  expr2,
		Symbol: symbol,
	}
	c.Expression = NewExpr(op, boolType, build.PrecedenceComparison)
	c.Expression.SetOuter(c)
	return c
}

func (c *Comparison) Codegen(ctx CodegenContext) string {
	return fmt.Sprintf("%s %s %s", c.Expr1.Codegen(ctx), c.Symbol, c.Expr2.Codegen(ctx))
}

func (c *Comparison) Reads() []string {
	return dedupReads(c.Expr1, c.Expr2)
}

// Compare implements the [build.Expression] interface.
type Compare struct {
	Expression
	Expr1  Expression
	Expr2  Expression
	FnName string
}

func NewCompare(
	op *program.Operand,
	expr1 build.Expression,
	expr2 build.Expression,
	resultType *program.Type,
	addPredefinedFunc func(name string, args []build.Expression, ret TypeInfo, src string) error,
) (Expression, error) {
	wgslExpr1 := expr1.(Expression).ToNum()
	wgslExpr2 := expr2.(Expression).ToNum()

	rType := WGSLType(resultType, false)

	fnName := "gsl_compare2"
	err := addPredefinedFunc(fnName, GslCompare2NumArgs, TypeF32, GslCompare2NumSrc)
	if err != nil {
		return nil, err
	}
	switch rType.Name {
	case TypeVec2.Name:
		fnName = "gsl_compare2_vec2"
		err = addPredefinedFunc(fnName, GslCompare2Vec2Args, TypeVec2, GslCompare2Vec2Src)
	case TypeVec3.Name:
		fnName = "gsl_compare2_vec3"
		err = addPredefinedFunc(fnName, GslCompare2Vec3Args, TypeVec3, GslCompare2Vec3Src)
	case TypeVec4.Name:
		fnName = "gsl_compare2_vec4"
		err = addPredefinedFunc(fnName, GslCompare2Vec4Args, TypeVec4, GslCompare2Vec4Src)
	}
	if err != nil {
		return nil, err
	}

	c := &Compare{
		Expr1:  wgslExpr1,
		Expr2:  wgslExpr2,
		FnName: fnName,
	}
	c.Expression = NewExpr(op, rType, build.PrecedenceHighest)
	c.Expression.SetOuter(c)
	return c, nil
}

func (c *Compare) Codegen(ctx CodegenContext) string {
	return fmt.Sprintf("%s(%s, %s)", c.FnName, c.Expr1.Codegen(ctx), c.Expr2.Codegen(ctx))
}

func (c *Compare) Reads() []string {
	return dedupReads(c.Expr1, c.Expr2)
}

var GslCompare2NumArgs = []build.Expression{
	NewOpaqueExpr(nil, TypeF32, build.PrecedenceHighest, "x"),
	NewOpaqueExpr(nil, TypeF32, build.PrecedenceHighest, "y")}

const GslCompare2NumSrc = `
fn gsl_compare2(x : f32, y : f32) -> f32 {
  if x > y { return 1.0; }
  if x < y { return -1.0; }
  return 0.0;
}`

var GslCompare2Vec2Args = []build.Expression{
	NewOpaqueExpr(nil, TypeVec2, build.PrecedenceHighest, "x"),
	NewOpaqueExpr(nil, TypeVec2, build.PrecedenceHighest, "y")}

const GslCompare2Vec2Src = `
fn gsl_compare2_vec2(x : vec2<f32>, y : vec2<f32>) -> vec2<f32> {
  return vec2(gsl_compare2(x.x, y.x), gsl_compare2(x.y, y.y));
}`

var GslCompare2Vec3Args = []build.Expression{
	NewOpaqueExpr(nil, TypeVec3, build.PrecedenceHighest, "x"),
	NewOpaqueExpr(nil, TypeVec3, build.PrecedenceHighest, "y")}

const GslCompare2Vec3Src = `
fn gsl_compare2_vec3(x : vec3<f32>, y : vec3<f32>) -> vec3<f32> {
  return vec3(gsl_compare2(x.x, y.x), gsl_compare2(x.y, y.y), gsl_compare2(x.z, y.z));
}`

var GslCompare2Vec4Args = []build.Expression{
	NewOpaqueExpr(nil, TypeVec4, build.PrecedenceHighest, "x"),
	NewOpaqueExpr(nil, TypeVec4, build.PrecedenceHighest, "y")}

const GslCompare2Vec4Src = `
fn gsl_compare2_vec4(x : vec4<f32>, y : vec4<f32>) -> vec4<f32> {
  return vec4(gsl_compare2(x.x, y.x), gsl_compare2(x.y, y.y), gsl_compare2(x.z, y.z), gsl_compare2(x.w, y.w));
}`
