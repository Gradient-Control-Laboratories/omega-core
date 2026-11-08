package wgsl

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// Select2 implements the [build.Expression] interface.
type Select2 struct {
	Expression
	CondExpr  Expression
	TrueExpr  Expression
	FalseExpr Expression
}

func NewSelect2(
	op *program.Operand,
	conditionExpr build.Expression,
	trueExpr build.Expression,
	falseExpr build.Expression,
	resultType *program.Type,
) (Expression, error) {
	wCondExpr := conditionExpr.(Expression).ToBool()
	wTrueExpr := trueExpr.(Expression)
	wFalseExpr := falseExpr.(Expression)

	var wResultType TypeInfo
	if resultType != nil {
		wResultType = WGSLType(resultType, false)
	} else {
		wResultType = wTrueExpr.WGSLType()
	}

	if !wTrueExpr.WGSLType().Equal(wFalseExpr.WGSLType()) {
		wTrueExpr = wTrueExpr.ToNum()
		wFalseExpr = wFalseExpr.ToNum()
	}

	s := &Select2{
		CondExpr:  wCondExpr.Unwrap(),
		TrueExpr:  wTrueExpr,
		FalseExpr: wFalseExpr,
	}
	s.Expression = NewExpr(op, wResultType, build.PrecedenceHighest)
	s.Expression.SetOuter(s)
	return s, nil
}

func NewSelectRaw(op *program.Operand, cond Expression, falseExpr Expression, trueExpr Expression, resultType TypeInfo) *Select2 {
	s := &Select2{
		CondExpr:  cond,
		FalseExpr: falseExpr,
		TrueExpr:  trueExpr,
	}
	s.Expression = NewExpr(op, resultType, build.PrecedenceHighest)
	s.Expression.SetOuter(s)
	return s
}

func (s *Select2) Codegen(ctx CodegenContext) string {
	return fmt.Sprintf("select(%s, %s, %s)",
		s.FalseExpr.Codegen(ctx),
		s.TrueExpr.Codegen(ctx),
		s.CondExpr.Codegen(ctx))
}

func (s *Select2) Reads() []string {
	return dedupReads(s.CondExpr, s.TrueExpr, s.FalseExpr)
}

// Select3 implements the [build.Expression] interface.
type Select3 struct {
	Expression
	ValueWgslExpr  Expression
	LtZeroWgslExpr Expression
	EqZeroWgslExpr Expression
	GtZeroWgslExpr Expression
	FnName         string
}

func NewSelect3(
	op *program.Operand,
	valueExpr build.Expression,
	ltZeroResultExpr build.Expression,
	eqZeroResultExpr build.Expression,
	gtZeroResultExpr build.Expression,
	resultType *program.Type,
	addPredefinedFunc func(name string, args []build.Expression, ret TypeInfo, src string) error,
) (Expression, error) {

	valueWgslExpr := valueExpr.(Expression).ToNum()
	ltZeroWgslExpr := ltZeroResultExpr.(Expression).ToNum()
	eqZeroWgslExpr := eqZeroResultExpr.(Expression).ToNum()
	gtZeroWgslExpr := gtZeroResultExpr.(Expression).ToNum()

	rType := WGSLType(resultType, false)

	fnName := "gsl_select3"
	err := addPredefinedFunc(fnName, GslSelect3ArgsNum, TypeF32, GslSelect3NumSrc)
	if err != nil {
		return nil, err
	}
	switch rType.Name {
	case TypeVec2.Name:
		fnName = "gsl_select3_vec2"
		err = addPredefinedFunc(fnName, GslSelect3ArgsVec2, TypeVec2, GslSelect3Vec2Src)
	case TypeVec3.Name:
		fnName = "gsl_select3_vec3"
		err = addPredefinedFunc(fnName, GslSelect3ArgsVec3, TypeVec3, GslSelect3Vec3Src)
	case TypeVec4.Name:
		fnName = "gsl_select3_vec4"
		err = addPredefinedFunc(fnName, GslSelect3ArgsVec4, TypeVec4, GslSelect3Vec4Src)
	}
	if err != nil {
		return nil, err
	}

	s := &Select3{
		ValueWgslExpr:  valueWgslExpr.Unwrap(),
		LtZeroWgslExpr: ltZeroWgslExpr.Unwrap(),
		EqZeroWgslExpr: eqZeroWgslExpr.Unwrap(),
		GtZeroWgslExpr: gtZeroWgslExpr.Unwrap(),
		FnName:         fnName,
	}
	s.Expression = NewExpr(op, rType, build.PrecedenceHighest)
	s.Expression.SetOuter(s)
	return s, nil
}

func (s *Select3) Codegen(ctx CodegenContext) string {
	return fmt.Sprintf("%s(%s, %s, %s, %s)",
		s.FnName,
		s.ValueWgslExpr.Codegen(ctx),
		s.LtZeroWgslExpr.Codegen(ctx),
		s.EqZeroWgslExpr.Codegen(ctx),
		s.GtZeroWgslExpr.Codegen(ctx))
}

func (s *Select3) Reads() []string {
	return dedupReads(s.ValueWgslExpr, s.LtZeroWgslExpr, s.EqZeroWgslExpr, s.GtZeroWgslExpr)
}

var GslSelect3ArgsNum = []build.Expression{
	NewOpaqueExpr(nil, TypeF32, build.PrecedenceHighest, "v"),
	NewOpaqueExpr(nil, TypeF32, build.PrecedenceHighest, "a"),
	NewOpaqueExpr(nil, TypeF32, build.PrecedenceHighest, "b"),
	NewOpaqueExpr(nil, TypeF32, build.PrecedenceHighest, "c")}

const GslSelect3NumSrc = `
fn gsl_select3(v : f32, a : f32, b : f32, c : f32) -> f32 {
  if v > 0.0 { return a; }
  if v < 0.0 { return c; }
  return b;
}`

var GslSelect3ArgsVec2 = []build.Expression{
	NewOpaqueExpr(nil, TypeVec2, build.PrecedenceHighest, "v"),
	NewOpaqueExpr(nil, TypeVec2, build.PrecedenceHighest, "a"),
	NewOpaqueExpr(nil, TypeVec2, build.PrecedenceHighest, "b"),
	NewOpaqueExpr(nil, TypeVec2, build.PrecedenceHighest, "c")}

const GslSelect3Vec2Src = `
fn gsl_select3_vec2(v : vec2<f32>, a : vec2<f32>, b : vec2<f32>, c : vec2<f32>) -> vec2<f32> {
  return vec2(gsl_select3(v.x, a.x, b.x, c.x), gsl_select3(v.y, a.y, b.y, c.y));
}`

var GslSelect3ArgsVec3 = []build.Expression{
	NewOpaqueExpr(nil, TypeVec3, build.PrecedenceHighest, "v"),
	NewOpaqueExpr(nil, TypeVec3, build.PrecedenceHighest, "a"),
	NewOpaqueExpr(nil, TypeVec3, build.PrecedenceHighest, "b"),
	NewOpaqueExpr(nil, TypeVec3, build.PrecedenceHighest, "c")}

const GslSelect3Vec3Src = `
fn gsl_select3_vec3(v : vec3<f32>, a : vec3<f32>, b : vec3<f32>, c : vec3<f32>) -> vec3<f32> {
  return vec3(gsl_select3(v.x, a.x, b.x, c.x), gsl_select3(v.y, a.y, b.y, c.y), gsl_select3(v.z, a.z, b.z, c.z));
}`

var GslSelect3ArgsVec4 = []build.Expression{
	NewOpaqueExpr(nil, TypeVec4, build.PrecedenceHighest, "v"),
	NewOpaqueExpr(nil, TypeVec4, build.PrecedenceHighest, "a"),
	NewOpaqueExpr(nil, TypeVec4, build.PrecedenceHighest, "b"),
	NewOpaqueExpr(nil, TypeVec4, build.PrecedenceHighest, "c")}

const GslSelect3Vec4Src = `
fn gsl_select3_vec4(v : vec4<f32>, a : vec4<f32>, b : vec4<f32>, c : vec4<f32>) -> vec4<f32> {
  return vec4(gsl_select3(v.x, a.x, b.x, c.x), gsl_select3(v.y, a.y, b.y, c.y), gsl_select3(v.z, a.z, b.z, c.z), gsl_select3(v.w, a.w, b.w, c.w));
}`
