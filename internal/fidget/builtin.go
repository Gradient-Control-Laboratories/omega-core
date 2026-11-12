package fidget

import (
	"fmt"
	"math"

	"gradientcontrol.com/omega/core/program"
)

// fromBuiltinCall handles built-in function calls that the parser recognizes (see
// common/model/builtin.go builtinsMap). Each builtin is decomposed into fidget tree primitives
// since the tape has no native "function call" concept. Returns (result, error, handled). If
// handled is false, the caller should fall through to user-defined function lookup.
func (s *Scope) fromBuiltinCall(call *program.FnCall) ([]TreeIndex, error, bool) {
	switch call.Reference.Name {
	case "length":
		result, err := s.builtinLength(call)
		return result, err, true
	case "distance":
		result, err := s.builtinDistance(call)
		return result, err, true
	case "dot":
		result, err := s.builtinDot(call)
		return result, err, true
	case "clamp":
		result, err := s.builtinClamp(call)
		return result, err, true
	case "normalize":
		result, err := s.builtinNormalize(call)
		return result, err, true
	case "sign":
		result, err := s.builtinSign(call)
		return result, err, true
	case "fract":
		result, err := s.builtinFract(call)
		return result, err, true
	case "cross":
		result, err := s.builtinCross(call)
		return result, err, true
	case "degrees":
		result, err := s.builtinUnaryScale(call, "degrees", 180.0/math.Pi)
		return result, err, true
	case "radians":
		result, err := s.builtinUnaryScale(call, "radians", math.Pi/180.0)
		return result, err, true
	case "exp2":
		result, err := s.builtinExp2(call)
		return result, err, true
	case "trunc":
		result, err := s.builtinTrunc(call)
		return result, err, true
	case "acosh":
		result, err := s.builtinAcosh(call)
		return result, err, true
	case "asinh":
		result, err := s.builtinAsinh(call)
		return result, err, true
	case "atanh":
		result, err := s.builtinAtanh(call)
		return result, err, true
	case "cosh":
		result, err := s.builtinCosh(call)
		return result, err, true
	case "sinh":
		result, err := s.builtinSinh(call)
		return result, err, true
	case "tanh":
		result, err := s.builtinTanh(call)
		return result, err, true
	}
	return nil, nil, false
}

// ---- vector builtins ----

// length(v) = sqrt(sum(v_i^2))
func (s *Scope) builtinLength(call *program.FnCall) ([]TreeIndex, error) {
	if len(call.Arguments) != 1 {
		return noTreeIndices, fmt.Errorf("length expects 1 argument, got %d", len(call.Arguments))
	}
	v, err := s.MultiValueFromOperand(call.Arguments[0])
	if err != nil {
		return noTreeIndices, fmt.Errorf("length: %v", err)
	}
	if len(v) < 2 {
		return noTreeIndices, fmt.Errorf("length expects a collection, got %d components", len(v))
	}
	return []TreeIndex{vecLength(v)}, nil
}

// distance(a, b) = length(a - b)
func (s *Scope) builtinDistance(call *program.FnCall) ([]TreeIndex, error) {
	if len(call.Arguments) != 2 {
		return noTreeIndices, fmt.Errorf("distance expects 2 arguments, got %d", len(call.Arguments))
	}
	a, err := s.MultiValueFromOperand(call.Arguments[0])
	if err != nil {
		return noTreeIndices, fmt.Errorf("distance arg 0: %v", err)
	}
	b, err := s.MultiValueFromOperand(call.Arguments[1])
	if err != nil {
		return noTreeIndices, fmt.Errorf("distance arg 1: %v", err)
	}
	if len(a) != len(b) {
		return noTreeIndices, fmt.Errorf("distance: mismatched collection sizes %d vs %d", len(a), len(b))
	}
	// Scalar distance = abs(a - b)
	if len(a) == 1 {
		return []TreeIndex{Abs(Sub(a[0], b[0]))}, nil
	}
	diff := make([]TreeIndex, len(a))
	for i := range a {
		diff[i] = Sub(a[i], b[i])
	}
	return []TreeIndex{vecLength(diff)}, nil
}

// dot(a, b) = sum(a_i * b_i)
func (s *Scope) builtinDot(call *program.FnCall) ([]TreeIndex, error) {
	if len(call.Arguments) != 2 {
		return noTreeIndices, fmt.Errorf("dot expects 2 arguments, got %d", len(call.Arguments))
	}
	a, err := s.MultiValueFromOperand(call.Arguments[0])
	if err != nil {
		return noTreeIndices, fmt.Errorf("dot arg 0: %v", err)
	}
	b, err := s.MultiValueFromOperand(call.Arguments[1])
	if err != nil {
		return noTreeIndices, fmt.Errorf("dot arg 1: %v", err)
	}
	if len(a) != len(b) {
		return noTreeIndices, fmt.Errorf("dot: mismatched collection sizes %d vs %d", len(a), len(b))
	}
	if len(a) < 2 {
		return noTreeIndices, fmt.Errorf("dot expects collections with at least 2 components, got %d", len(a))
	}
	result := Mul(a[0], b[0])
	for i := 1; i < len(a); i++ {
		result = Add(result, Mul(a[i], b[i]))
	}
	return []TreeIndex{result}, nil
}

// cross(a, b) = (a.y*b.z - a.z*b.y, a.z*b.x - a.x*b.z, a.x*b.y - a.y*b.x)
func (s *Scope) builtinCross(call *program.FnCall) ([]TreeIndex, error) {
	if len(call.Arguments) != 2 {
		return noTreeIndices, fmt.Errorf("cross expects 2 arguments, got %d", len(call.Arguments))
	}
	a, err := s.MultiValueFromOperand(call.Arguments[0])
	if err != nil {
		return noTreeIndices, fmt.Errorf("cross arg 0: %v", err)
	}
	b, err := s.MultiValueFromOperand(call.Arguments[1])
	if err != nil {
		return noTreeIndices, fmt.Errorf("cross arg 1: %v", err)
	}
	if len(a) != 3 || len(b) != 3 {
		return noTreeIndices, fmt.Errorf("cross expects vec3 arguments, got %d and %d components", len(a), len(b))
	}
	rx := Sub(Mul(a[1], b[2]), Mul(a[2], b[1]))
	ry := Sub(Mul(a[2], b[0]), Mul(a[0], b[2]))
	rz := Sub(Mul(a[0], b[1]), Mul(a[1], b[0]))
	return []TreeIndex{rx, ry, rz}, nil
}

// normalize(v) = v / length(v)
func (s *Scope) builtinNormalize(call *program.FnCall) ([]TreeIndex, error) {
	if len(call.Arguments) != 1 {
		return noTreeIndices, fmt.Errorf("normalize expects 1 argument, got %d", len(call.Arguments))
	}
	v, err := s.MultiValueFromOperand(call.Arguments[0])
	if err != nil {
		return noTreeIndices, fmt.Errorf("normalize: %v", err)
	}
	if len(v) < 2 {
		return noTreeIndices, fmt.Errorf("normalize expects a collection, got %d components", len(v))
	}
	l := vecLength(v)
	result := make([]TreeIndex, len(v))
	for i := range v {
		result[i] = Div(v[i], l)
	}
	return result, nil
}

// ---- element-wise builtins ----

// clamp(x, lo, hi) = min(max(x, lo), hi) — element-wise with broadcasting
func (s *Scope) builtinClamp(call *program.FnCall) ([]TreeIndex, error) {
	if len(call.Arguments) != 3 {
		return noTreeIndices, fmt.Errorf("clamp expects 3 arguments, got %d", len(call.Arguments))
	}
	x, err := s.MultiValueFromOperand(call.Arguments[0])
	if err != nil {
		return noTreeIndices, fmt.Errorf("clamp arg 0: %v", err)
	}
	lo, err := s.MultiValueFromOperand(call.Arguments[1])
	if err != nil {
		return noTreeIndices, fmt.Errorf("clamp arg 1: %v", err)
	}
	hi, err := s.MultiValueFromOperand(call.Arguments[2])
	if err != nil {
		return noTreeIndices, fmt.Errorf("clamp arg 2: %v", err)
	}
	// Broadcast scalars to match longest vector
	maxLen := len(x)
	if len(lo) > maxLen {
		maxLen = len(lo)
	}
	if len(hi) > maxLen {
		maxLen = len(hi)
	}
	if len(x) == 1 && maxLen > 1 {
		x = expand(x[0], maxLen)
	}
	if len(lo) == 1 && maxLen > 1 {
		lo = expand(lo[0], maxLen)
	}
	if len(hi) == 1 && maxLen > 1 {
		hi = expand(hi[0], maxLen)
	}
	if len(x) != len(lo) || len(x) != len(hi) {
		return noTreeIndices, fmt.Errorf("clamp: mismatched sizes %d, %d, %d", len(x), len(lo), len(hi))
	}
	result := make([]TreeIndex, len(x))
	for i := range x {
		result[i] = Min(Max(x[i], lo[i]), hi[i])
	}
	return result, nil
}

// sign(x) = compare(x, 0) — element-wise
func (s *Scope) builtinSign(call *program.FnCall) ([]TreeIndex, error) {
	if len(call.Arguments) != 1 {
		return noTreeIndices, fmt.Errorf("sign expects 1 argument, got %d", len(call.Arguments))
	}
	v, err := s.MultiValueFromOperand(call.Arguments[0])
	if err != nil {
		return noTreeIndices, fmt.Errorf("sign: %v", err)
	}
	zero := NewConst(0)
	result := make([]TreeIndex, len(v))
	for i := range v {
		result[i] = Compare(v[i], zero)
	}
	return result, nil
}

// fract(x) = x - floor(x) — element-wise
func (s *Scope) builtinFract(call *program.FnCall) ([]TreeIndex, error) {
	if len(call.Arguments) != 1 {
		return noTreeIndices, fmt.Errorf("fract expects 1 argument, got %d", len(call.Arguments))
	}
	v, err := s.MultiValueFromOperand(call.Arguments[0])
	if err != nil {
		return noTreeIndices, fmt.Errorf("fract: %v", err)
	}
	result := make([]TreeIndex, len(v))
	for i := range v {
		result[i] = Sub(v[i], Floor(v[i]))
	}
	return result, nil
}

// trunc(x) = sign(x) * floor(abs(x)) — element-wise
func (s *Scope) builtinTrunc(call *program.FnCall) ([]TreeIndex, error) {
	if len(call.Arguments) != 1 {
		return noTreeIndices, fmt.Errorf("trunc expects 1 argument, got %d", len(call.Arguments))
	}
	v, err := s.MultiValueFromOperand(call.Arguments[0])
	if err != nil {
		return noTreeIndices, fmt.Errorf("trunc: %v", err)
	}
	zero := NewConst(0)
	result := make([]TreeIndex, len(v))
	for i := range v {
		result[i] = Mul(Compare(v[i], zero), Floor(Abs(v[i])))
	}
	return result, nil
}

// builtinUnaryScale multiplies each component by a constant — used for degrees and radians
func (s *Scope) builtinUnaryScale(call *program.FnCall, name string, factor float64) ([]TreeIndex, error) {
	if len(call.Arguments) != 1 {
		return noTreeIndices, fmt.Errorf("%s expects 1 argument, got %d", name, len(call.Arguments))
	}
	v, err := s.MultiValueFromOperand(call.Arguments[0])
	if err != nil {
		return noTreeIndices, fmt.Errorf("%s: %v", name, err)
	}
	k := NewConst(factor)
	result := make([]TreeIndex, len(v))
	for i := range v {
		result[i] = Mul(v[i], k)
	}
	return result, nil
}

// exp2(x) = exp(x * ln(2)) — element-wise
func (s *Scope) builtinExp2(call *program.FnCall) ([]TreeIndex, error) {
	if len(call.Arguments) != 1 {
		return noTreeIndices, fmt.Errorf("exp2 expects 1 argument, got %d", len(call.Arguments))
	}
	v, err := s.MultiValueFromOperand(call.Arguments[0])
	if err != nil {
		return noTreeIndices, fmt.Errorf("exp2: %v", err)
	}
	ln2 := NewConst(math.Ln2)
	result := make([]TreeIndex, len(v))
	for i := range v {
		result[i] = Exp(Mul(v[i], ln2))
	}
	return result, nil
}

// ---- hyperbolic builtins ----

// sinh(x) = (exp(x) - exp(-x)) / 2 — element-wise
func (s *Scope) builtinSinh(call *program.FnCall) ([]TreeIndex, error) {
	if len(call.Arguments) != 1 {
		return noTreeIndices, fmt.Errorf("sinh expects 1 argument, got %d", len(call.Arguments))
	}
	v, err := s.MultiValueFromOperand(call.Arguments[0])
	if err != nil {
		return noTreeIndices, fmt.Errorf("sinh: %v", err)
	}
	two := NewConst(2)
	result := make([]TreeIndex, len(v))
	for i := range v {
		ep := Exp(v[i])
		en := Exp(Neg(v[i]))
		result[i] = Div(Sub(ep, en), two)
	}
	return result, nil
}

// cosh(x) = (exp(x) + exp(-x)) / 2 — element-wise
func (s *Scope) builtinCosh(call *program.FnCall) ([]TreeIndex, error) {
	if len(call.Arguments) != 1 {
		return noTreeIndices, fmt.Errorf("cosh expects 1 argument, got %d", len(call.Arguments))
	}
	v, err := s.MultiValueFromOperand(call.Arguments[0])
	if err != nil {
		return noTreeIndices, fmt.Errorf("cosh: %v", err)
	}
	two := NewConst(2)
	result := make([]TreeIndex, len(v))
	for i := range v {
		ep := Exp(v[i])
		en := Exp(Neg(v[i]))
		result[i] = Div(Add(ep, en), two)
	}
	return result, nil
}

// tanh(x) = (exp(x) - exp(-x)) / (exp(x) + exp(-x)) — element-wise
func (s *Scope) builtinTanh(call *program.FnCall) ([]TreeIndex, error) {
	if len(call.Arguments) != 1 {
		return noTreeIndices, fmt.Errorf("tanh expects 1 argument, got %d", len(call.Arguments))
	}
	v, err := s.MultiValueFromOperand(call.Arguments[0])
	if err != nil {
		return noTreeIndices, fmt.Errorf("tanh: %v", err)
	}
	result := make([]TreeIndex, len(v))
	for i := range v {
		ep := Exp(v[i])
		en := Exp(Neg(v[i]))
		result[i] = Div(Sub(ep, en), Add(ep, en))
	}
	return result, nil
}

// asinh(x) = ln(x + sqrt(x² + 1)) — element-wise
func (s *Scope) builtinAsinh(call *program.FnCall) ([]TreeIndex, error) {
	if len(call.Arguments) != 1 {
		return noTreeIndices, fmt.Errorf("asinh expects 1 argument, got %d", len(call.Arguments))
	}
	v, err := s.MultiValueFromOperand(call.Arguments[0])
	if err != nil {
		return noTreeIndices, fmt.Errorf("asinh: %v", err)
	}
	one := NewConst(1)
	result := make([]TreeIndex, len(v))
	for i := range v {
		result[i] = Ln(Add(v[i], Sqrt(Add(Square(v[i]), one))))
	}
	return result, nil
}

// acosh(x) = ln(x + sqrt(x² - 1)) — element-wise
func (s *Scope) builtinAcosh(call *program.FnCall) ([]TreeIndex, error) {
	if len(call.Arguments) != 1 {
		return noTreeIndices, fmt.Errorf("acosh expects 1 argument, got %d", len(call.Arguments))
	}
	v, err := s.MultiValueFromOperand(call.Arguments[0])
	if err != nil {
		return noTreeIndices, fmt.Errorf("acosh: %v", err)
	}
	one := NewConst(1)
	result := make([]TreeIndex, len(v))
	for i := range v {
		result[i] = Ln(Add(v[i], Sqrt(Sub(Square(v[i]), one))))
	}
	return result, nil
}

// atanh(x) = 0.5 * ln((1 + x) / (1 - x)) — element-wise
func (s *Scope) builtinAtanh(call *program.FnCall) ([]TreeIndex, error) {
	if len(call.Arguments) != 1 {
		return noTreeIndices, fmt.Errorf("atanh expects 1 argument, got %d", len(call.Arguments))
	}
	v, err := s.MultiValueFromOperand(call.Arguments[0])
	if err != nil {
		return noTreeIndices, fmt.Errorf("atanh: %v", err)
	}
	one := NewConst(1)
	half := NewConst(0.5)
	result := make([]TreeIndex, len(v))
	for i := range v {
		result[i] = Mul(half, Ln(Div(Add(one, v[i]), Sub(one, v[i]))))
	}
	return result, nil
}

// ---- helpers ----

// vecLength computes sqrt(sum(v_i^2))
func vecLength(v []TreeIndex) TreeIndex {
	sum := Square(v[0])
	for i := 1; i < len(v); i++ {
		sum = Add(sum, Square(v[i]))
	}
	return Sqrt(sum)
}
