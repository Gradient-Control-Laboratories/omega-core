package codegen

import (
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// jsExpression implements build.Expression and stores generated JavaScript code
type jsExpression struct {
	code        string
	operand     *program.Operand
	precedence  build.Precedence
	typ         *program.Type
	constValues []float64
	elements    []build.Expression
}

func newExpression(code string) *jsExpression {
	return &jsExpression{
		code:       code,
		operand:    nil,
		precedence: build.PrecedenceHighest,
		typ:        jsNumType,
	}
}

func (e *jsExpression) Operand() *program.Operand {
	return e.operand
}

func (e *jsExpression) Start() *program.Location {
	if e.operand != nil {
		return e.operand.Start
	}
	return nil
}

func (e *jsExpression) Precedence() build.Precedence {
	return e.precedence
}

func (e *jsExpression) Type() *program.Type {
	return e.typ
}

// Code returns the generated JavaScript code
func (e *jsExpression) Code() string {
	return e.code
}

func (e *jsExpression) ConstValues() []float64 {
	return e.constValues
}

func (e *jsExpression) Elements() []build.Expression {
	return e.elements
}

func (e *jsExpression) ParenIf(precedence build.Precedence) *jsExpression {
	if e.precedence > precedence {
		return &jsExpression{
			code:       "(" + e.code + ")",
			operand:    e.operand,
			precedence: build.PrecedenceHighest,
			typ:        e.typ,
		}
	}
	return e
}

// getCode safely extracts JavaScript code from a build.Expression
func getCode(expr build.Expression) string {
	if jsExpr, ok := expr.(*jsExpression); ok {
		return jsExpr.Code()
	}
	return "/* unknown expression */"
}
