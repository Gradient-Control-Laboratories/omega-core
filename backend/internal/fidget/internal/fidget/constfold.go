package fidget

import "math"

// fidget stores and folds constants in f32. A constant subexpression built
// here (0.4 * 0.4) would round each operand to f32 and then round the result
// again, landing an ulp or more away from the correctly rounded value. So
// constant subexpressions are folded here in f64 and reach fidget as a single
// constant, rounded once. The op semantics mirror the f64 Context::eval of
// fidget 0.4, which produced the same values before fidget moved to f32.
//
// NewConst records each value; guarded by the same Lock as the FFI calls,
// reset by Clear.
var constValues = map[TreeIndex]float64{}

func fold1(a TreeIndex, f func(float64) float64) (TreeIndex, bool) {
	x, ok := constValues[a]
	if !ok {
		return 0, false
	}
	return NewConst(f(x)), true
}

func fold2(a, b TreeIndex, f func(float64, float64) float64) (TreeIndex, bool) {
	x, ok := constValues[a]
	if !ok {
		return 0, false
	}
	y, ok := constValues[b]
	if !ok {
		return 0, false
	}
	return NewConst(f(x, y)), true
}

// Rust's f64::min / f64::max return the other operand when one is NaN.
func rustMin(a, b float64) float64 {
	if math.IsNaN(a) {
		return b
	}
	if math.IsNaN(b) {
		return a
	}
	return math.Min(a, b)
}

func rustMax(a, b float64) float64 {
	if math.IsNaN(a) {
		return b
	}
	if math.IsNaN(b) {
		return a
	}
	return math.Max(a, b)
}

func foldCompare(a, b float64) float64 {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	case a == b:
		return 0
	}
	return math.NaN()
}

// Rust's f64::rem_euclid.
func foldMod(a, b float64) float64 {
	r := math.Mod(a, b)
	if r < 0 {
		r += math.Abs(b)
	}
	return r
}

func foldAnd(a, b float64) float64 {
	if a == 0 {
		return a
	}
	return b
}

func foldOr(a, b float64) float64 {
	if a != 0 {
		return a
	}
	return b
}

func foldNot(a float64) float64 {
	if a == 0 {
		return 1
	}
	return 0
}

// Tree::pow: exponentiation by squaring, one fold per tree op it builds.
func foldPowInt(x float64, n int64) float64 {
	switch {
	case n < 0:
		n = -n
		x = 1 / x
	case n == 0:
		return 1
	}
	y, haveY := 0.0, false
	for n > 1 {
		if n%2 == 1 {
			if haveY {
				y = x * y
			} else {
				y, haveY = x, true
			}
			n--
		}
		x = x * x
		n /= 2
	}
	if haveY {
		x *= y
	}
	return x
}
