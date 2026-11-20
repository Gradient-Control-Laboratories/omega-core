package codegen

import (
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// jsExpression implements build.Expression and stores generated JavaScript code
type jsExpression struct {
	code       string
	operand    *program.Operand
	precedence build.Precedence
	typ        *program.Type
	elements   []build.Expression
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

// jsConst is a jsExpression whose value is known at compile time. Only it
// implements build.ConstExpression: the build layer takes any ConstExpression
// as a constant (bindFoldStep, for one, leaves a fold step unbound if its
// accumulator is one), so a plain jsExpression must not satisfy the interface.
type jsConst struct {
	*jsExpression
	values []float64
}

func (c *jsConst) ConstValues() []float64 {
	return c.values
}

// asJS returns the jsExpression behind expr, constant or not, or nil.
func asJS(expr build.Expression) *jsExpression {
	switch e := expr.(type) {
	case *jsExpression:
		return e
	case *jsConst:
		return e.jsExpression
	}
	return nil
}

// indexTarget returns the code of target ready to be indexed (target[i]):
// parenthesized when it is an operator or ternary expression, since `c ? a :
// b[0]` indexes b alone. Names, literals and calls index as they are (calls
// are tagged PrecedenceLowest here but bind tighter than indexing).
func indexTarget(target build.Expression) string {
	code := getCode(target)
	if p := target.Precedence(); p > build.PrecedenceHighest && p < build.PrecedenceLowest {
		return "(" + code + ")"
	}
	return code
}

// getCode safely extracts JavaScript code from a build.Expression
func getCode(expr build.Expression) string {
	if jsExpr := asJS(expr); jsExpr != nil {
		return jsExpr.Code()
	}
	return "/* unknown expression */"
}
