package wgsl

import (
	"strconv"
	"strings"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// Value implements the [build.Expression] interface.
type Value struct {
	Expression
	Value float64
}

func NewValue(op *program.Operand, value float64) *Value {
	v := &Value{
		Value: value,
	}
	v.Expression = NewExpr(op, TypeF32, build.PrecedenceHighest)
	v.Expression.SetOuter(v)
	return v
}

func (v *Value) ConstValues() []float64 {
	return []float64{v.Value}
}

func (v *Value) Codegen(ctx CodegenContext) string {
	return CodegenValue(v.Value)
}

func (v *Value) Reads() []string {
	return nil
}

// CodegenValue returns a WGSL float literal for value: the shortest plain
// decimal that round-trips as f32, which is what an unsuffixed WGSL float
// literal is. Plain decimal rather than %g on purpose — no golden this pipeline
// has ever produced uses exponent notation, and there is no shader validator
// here to prove a browser would take one, so the literal style stays as it was
// and only the digits change.
//
// This used to be %f, six decimal places, which broke at both ends. Anything
// below about 5e-7 printed as 0.0, so a guard written as max(x, 1e-9) reached
// the shader as max(x, 0.0) and guarded nothing while working correctly in the
// mesher, which keeps the real value — omega.gsl's intersectLp carries exactly
// that. And every irrational constant lost its tail: pi emitted as 3.141593,
// about two f32 ulps off a type that can represent 3.1415927.
func CodegenValue(value float64) string {
	src := strconv.FormatFloat(value, 'f', -1, 32)

	// Keep it a float literal — WGSL reads a bare digit string as an integer.
	if !strings.Contains(src, ".") {
		src += ".0"
	}

	return src
}
