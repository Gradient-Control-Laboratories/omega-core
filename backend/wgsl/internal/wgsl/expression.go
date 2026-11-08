package wgsl

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

type CodegenContext interface {
	P() *VarInfo
	WithP(p *VarInfo) CodegenContext
	Root() CodegenContext
	AssignUnique(prefix string, target string, reads ...string) string
	EmitVar(ref *VarInfo)
}

type Expression interface {
	build.Expression

	Codegen(ctx CodegenContext) string
	ParenIf(precedence build.Precedence) Expression
	ToCommon(commonType TypeInfo) (Expression, error)
	ToBool() Expression
	ToNum() Expression
	Unwrap() Expression
	WGSLType() TypeInfo
	Reads() []string
	SetOuter(outer Expression)
}

// expressionImpl implements the [build.Expression] and the [Expression] interface.
type expressionImpl struct {
	operand    *program.Operand
	exprType   TypeInfo
	precedence build.Precedence
	sub        Expression

	// outer points to the enclosing concrete struct (e.g. *VarRef, *NumericOp)
	// that embeds this expressionImpl.
	outer Expression

	reads []string
}

func (e *expressionImpl) SetOuter(outer Expression) {
	e.outer = outer
}

func (e *expressionImpl) getOuter() Expression {
	if e.outer != nil {
		return e.outer
	}
	return e
}

func NewExpr(
	operand *program.Operand,
	exprType TypeInfo,
	precedence build.Precedence,
) Expression {
	if exprType.Name == "" && operand != nil && operand.Type != nil {
		exprType = WGSLType(operand.Type, operand.IsBool())
	}
	return &expressionImpl{
		operand:    operand,
		precedence: precedence,
		exprType:   exprType,
	}
}

func NewExprWithReads(
	operand *program.Operand,
	exprType TypeInfo,
	precedence build.Precedence,
	reads ...string,
) Expression {
	if exprType.Name == "" && operand != nil && operand.Type != nil {
		exprType = WGSLType(operand.Type, operand.IsBool())
	}
	return &expressionImpl{
		operand:    operand,
		precedence: precedence,
		exprType:   exprType,
		reads:      reads,
	}
}

func NewSubExpr(
	sub Expression,
	operand *program.Operand,
	exprType TypeInfo,
	precedence build.Precedence,
) Expression {
	if exprType.Name == "" {
		if operand != nil && operand.Type != nil {
			exprType = WGSLType(operand.Type, operand.IsBool())
		} else if sub != nil {
			exprType = sub.WGSLType()
		}
	}
	return &expressionImpl{
		operand:    operand,
		precedence: precedence,
		exprType:   exprType,
		sub:        sub,
	}
}

type parenExpr struct {
	Expression
}

func (p *parenExpr) Reads() []string {
	return p.Expression.Reads()
}

func (p *parenExpr) SetOuter(outer Expression) {
	p.Expression.SetOuter(outer)
}

func (p *parenExpr) Codegen(ctx CodegenContext) string {
	return "(" + p.Expression.Codegen(ctx) + ")"
}

func (p *parenExpr) Unwrap() Expression {
	return p.Expression.Unwrap()
}

func (p *parenExpr) Precedence() build.Precedence {
	return build.PrecedenceHighest
}

func (p *parenExpr) ParenIf(precedence build.Precedence) Expression {
	return p
}

func (p *parenExpr) WGSLType() TypeInfo {
	return p.Expression.WGSLType()
}

func (p *parenExpr) ToCommon(commonType TypeInfo) (Expression, error) {
	e, err := p.Expression.ToCommon(commonType)
	if err != nil {
		return nil, err
	}
	if e == p.Expression {
		return p, nil
	}
	return Paren(e), nil
}

func (p *parenExpr) ToBool() Expression {
	e := p.Expression.ToBool()
	if e == p.Expression {
		return p
	}
	return Paren(e)
}

func (p *parenExpr) ToNum() Expression {
	e := p.Expression.ToNum()
	if e == p.Expression {
		return p
	}
	return Paren(e)
}

func Paren(e Expression) Expression {
	return &parenExpr{Expression: e}
}

func (e *expressionImpl) Unwrap() Expression {
	return e.getOuter()
}

func (e *expressionImpl) ParenIf(precedence build.Precedence) Expression {
	if e.precedence > precedence {
		return Paren(e.getOuter())
	}
	return e.getOuter()
}

// ----------------------------------------------------------------
// Implement build.Expression
// ----------------------------------------------------------------

func (e *expressionImpl) Operand() *program.Operand {
	return e.operand
}

func (e *expressionImpl) Precedence() build.Precedence {
	return e.precedence
}

func (e *expressionImpl) Start() *program.Location {
	if op := e.operand; op != nil {
		return op.Start
	}
	return nil
}

func (e *expressionImpl) Type() *program.Type {
	if e.operand != nil && e.operand.Type != nil && e.operand.Type.IsFunction() {
		return e.operand.Type
	}
	if e.exprType.ProgType != nil {
		return e.exprType.ProgType
	}
	if e.operand != nil && e.operand.Type != nil {
		return e.operand.Type
	}
	if e.sub != nil {
		return e.sub.Type()
	}
	return nil
}

func (e *expressionImpl) WGSLType() TypeInfo {
	return e.exprType
}

// ----------------------------------------------------------------

func (e *expressionImpl) Reads() []string {
	if e.reads != nil {
		r := make([]string, len(e.reads))
		copy(r, e.reads)
		return r
	}
	if e.sub != nil {
		return e.sub.Reads()
	}
	return nil
}

func (e *expressionImpl) Codegen(ctx CodegenContext) string {
	if e.outer != nil && e.outer != e {
		return e.outer.Codegen(ctx)
	}
	panic("Codegen called on raw expressionImpl without outer implementation")
}

func (e *expressionImpl) ToCommon(commonType TypeInfo) (Expression, error) {
	outer := e.getOuter()
	if e.exprType.Equal(commonType) || e.exprType.Name == commonType.Name {
		return outer, nil
	}

	if e.exprType.IsNumericScalar() && commonType.IsNumericVector() {
		return NewCastExpr(e.operand, commonType, outer), nil
	}
	if e.exprType.IsBoolScalar() && commonType.IsBoolVector() {
		return NewCastExpr(e.operand, commonType, outer), nil
	}

	return nil, fmt.Errorf("failed to convert %s to %s", e.exprType.Name, commonType.Name)
}

func (e *expressionImpl) ToBool() Expression {
	outer := e.getOuter()
	switch e.WGSLType().Name {
	case TypeF32.Name:
		zero := NewOpaqueExpr(nil, TypeF32, build.PrecedenceHighest, "0.0")
		return NewCompareRaw(e.operand, outer.ParenIf(build.PrecedenceComparison), ">", zero, TypeBool)
	case TypeVec2.Name:
		zero := NewOpaqueExpr(nil, TypeVec2, build.PrecedenceHighest, "0.0")
		return NewCompareRaw(e.operand, outer.ParenIf(build.PrecedenceComparison), ">", zero, TypeVec2Bool)
	case TypeVec3.Name:
		zero := NewOpaqueExpr(nil, TypeVec3, build.PrecedenceHighest, "0.0")
		return NewCompareRaw(e.operand, outer.ParenIf(build.PrecedenceComparison), ">", zero, TypeVec3Bool)
	case TypeVec4.Name:
		zero := NewOpaqueExpr(nil, TypeVec4, build.PrecedenceHighest, "0.0")
		return NewCompareRaw(e.operand, outer.ParenIf(build.PrecedenceComparison), ">", zero, TypeVec4Bool)
	}
	return outer
}

func (e *expressionImpl) ToNum() Expression {
	outer := e.getOuter()
	unwrapped := e.Unwrap()
	cond := unwrapped
	if unwrapped.Precedence() >= build.PrecedenceComparison {
		cond = Paren(cond)
	}
	switch e.WGSLType().Name {
	case TypeBool.Name:
		zero := NewOpaqueExpr(nil, TypeF32, build.PrecedenceHighest, "0.0")
		one := NewOpaqueExpr(nil, TypeF32, build.PrecedenceHighest, "1.0")
		return NewSelectRaw(e.operand, cond, zero, one, TypeF32)
	case TypeVec2Bool.Name:
		zero := NewOpaqueExpr(nil, TypeVec2, build.PrecedenceHighest, "vec2<f32>(0.0)")
		one := NewOpaqueExpr(nil, TypeVec2, build.PrecedenceHighest, "vec2<f32>(1.0)")
		return NewSelectRaw(e.operand, cond, zero, one, TypeVec2)
	case TypeVec3Bool.Name:
		zero := NewOpaqueExpr(nil, TypeVec3, build.PrecedenceHighest, "vec3<f32>(0.0)")
		one := NewOpaqueExpr(nil, TypeVec3, build.PrecedenceHighest, "vec3<f32>(1.0)")
		return NewSelectRaw(e.operand, cond, zero, one, TypeVec3)
	case TypeVec4Bool.Name:
		zero := NewOpaqueExpr(nil, TypeVec4, build.PrecedenceHighest, "vec4<f32>(0.0)")
		one := NewOpaqueExpr(nil, TypeVec4, build.PrecedenceHighest, "vec4<f32>(1.0)")
		return NewSelectRaw(e.operand, cond, zero, one, TypeVec4)
	}
	return outer
}

// dedupReads returns deduplicated WGSL identifier names read by the given expressions,
// preserving the first-seen appearance order.
func dedupReads(exprs ...Expression) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, expr := range exprs {
		for _, name := range expr.Reads() {
			if _, ok := seen[name]; !ok {
				seen[name] = struct{}{}
				result = append(result, name)
			}
		}
	}
	return result
}
