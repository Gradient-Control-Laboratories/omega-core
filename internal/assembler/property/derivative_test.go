package property

import (
	"fmt"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// point holds concrete values for the implied variables used when numerically
// evaluating operands in tests.
type point struct {
	x, y, z, time float64
}

func (p point) vars() map[string][]float64 {
	return map[string][]float64{
		program.XVarName:    {p.x},
		program.YVarName:    {p.y},
		program.ZVarName:    {p.z},
		program.TimeVarName: {p.time},
		program.PVarName:    {p.x, p.y, p.z},
	}
}

func newDerivative() *derivative {
	return &derivative{decls: map[string]*program.Operand{}, funcs: map[string][]*program.Function{}}
}

func TestDtime(t *testing.T) {
	d := newDerivative()
	pts := []point{{0.3, -1.2, 0.7, 0.4}, {-0.6, 0.9, -0.2, 1.1}, {1.3, 0.2, 0.5, -0.8}}

	cases := []struct {
		name   string
		target *program.Operand
	}{
		{"poly", add(mul(num(3), un(program.Operation_SQUARE, ref(program.TimeVarName))), mul(num(2), ref(program.TimeVarName)))},
		{"product", mul(ref(program.TimeVarName), un(program.Operation_SIN, ref(program.TimeVarName)))},
		{"quotient", div(ref(program.TimeVarName), add(ref(program.TimeVarName), num(2)))},
		{"sqrt", sqrt(add(un(program.Operation_SQUARE, ref(program.TimeVarName)), num(1)))},
		{"trig_chain", un(program.Operation_SIN, mul(num(2), un(program.Operation_COS, ref(program.TimeVarName))))},
		{"exp_ln", un(program.Operation_EXP, un(program.Operation_LN, add(ref(program.TimeVarName), num(3))))},
		{"atan2", nary(program.Operation_ATAN2, un(program.Operation_SIN, ref(program.TimeVarName)), add(ref(program.TimeVarName), num(2)))},
		{"pow_const", bin(program.Operation_POW, add(ref(program.TimeVarName), num(2)), num(2.5))},
		{"min", bin(program.Operation_MIN, mul(num(2), ref(program.TimeVarName)), add(ref(program.TimeVarName), num(1)))},
		{"max", bin(program.Operation_MAX, un(program.Operation_SQUARE, ref(program.TimeVarName)), ref(program.TimeVarName))},
		{"abs", un(program.Operation_ABS, sub(ref(program.TimeVarName), num(1)))},
		{"clamp", call("clamp", ref(program.TimeVarName), num(0), num(1))},
		{"tanh", call("tanh", mul(num(2), ref(program.TimeVarName)))},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			deriv, err := d.temporal(c.target)
			if err != nil {
				t.Fatalf("temporal: %v", err)
			}
			for _, pt := range pts {
				analytic := evalScalar(t, deriv, pt)
				numeric := fdTime(t, c.target, pt)
				if !closeEnough(analytic, numeric) {
					t.Errorf("at %+v: analytic=%v finite-diff=%v", pt, analytic, numeric)
				}
			}
		})
	}
}

func TestDpGradient(t *testing.T) {
	d := newDerivative()
	pts := []point{{0.3, -1.2, 0.7, 0.4}, {-0.6, 0.9, -0.2, 1.1}, {1.3, 0.2, 0.5, -0.8}}

	cases := []struct {
		name   string
		target *program.Operand
	}{
		{"length_p", call("length", ref(program.PVarName))},
		{"quadratic", add(add(un(program.Operation_SQUARE, ref(program.XVarName)), mul(num(2), un(program.Operation_SQUARE, ref(program.YVarName)))), mul(ref(program.XVarName), ref(program.ZVarName)))},
		{"dot_pp", call("dot", ref(program.PVarName), ref(program.PVarName))},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			grad, err := d.gradient(c.target)
			if err != nil {
				t.Fatalf("gradient: %v", err)
			}
			for _, pt := range pts {
				g := evalVec(t, grad, pt)
				if len(g) != 3 {
					t.Fatalf("expected vec3 gradient, got len %d", len(g))
				}
				want := [3]float64{
					fdSpatial(t, c.target, pt, 0),
					fdSpatial(t, c.target, pt, 1),
					fdSpatial(t, c.target, pt, 2),
				}
				for k := 0; k < 3; k++ {
					if !closeEnough(g[k], want[k]) {
						t.Errorf("at %+v axis %d: analytic=%v finite-diff=%v", pt, k, g[k], want[k])
					}
				}
			}
		})
	}
}

// TestRemapTemporal differentiates length(p) under a time-dependent translation
// remap q = [x+time, y, z]. d/dtime = (x+time)/|q|.
func TestRemapTemporal(t *testing.T) {
	d := newDerivative()
	target := nary(program.Operation_REMAP,
		call("length", ref(program.PVarName)),
		nary(program.Operation_COLLECT, add(ref(program.XVarName), ref(program.TimeVarName)), ref(program.YVarName), ref(program.ZVarName)),
	)
	deriv, err := d.temporal(target)
	if err != nil {
		t.Fatalf("temporal: %v", err)
	}
	inlined := call("length", nary(program.Operation_COLLECT, add(ref(program.XVarName), ref(program.TimeVarName)), ref(program.YVarName), ref(program.ZVarName)))
	for _, pt := range []point{{0.3, -1.2, 0.7, 0.4}, {-0.6, 0.9, -0.2, 1.1}} {
		analytic := evalScalar(t, deriv, pt)
		numeric := fdTime(t, inlined, pt)
		if !closeEnough(analytic, numeric) {
			t.Errorf("at %+v: analytic=%v finite-diff=%v", pt, analytic, numeric)
		}
	}
}

// TestUserFunctionInline differentiates a call to a user-defined function.
func TestUserFunctionInline(t *testing.T) {
	sq := &program.Function{
		Parameters: []*program.Parameter{{Name: "a"}},
		Returns:    []*program.Operand{mul(ref("a"), ref("a"))},
	}
	d := &derivative{decls: map[string]*program.Operand{}, funcs: map[string][]*program.Function{"sq": {sq}}}
	target := call("sq", add(ref(program.TimeVarName), num(1)))
	deriv, err := d.temporal(target)
	if err != nil {
		t.Fatalf("temporal: %v", err)
	}
	inlined := mul(add(ref(program.TimeVarName), num(1)), add(ref(program.TimeVarName), num(1)))
	for _, pt := range []point{{0, 0, 0, 0.4}, {0, 0, 0, 2.0}} {
		analytic := evalScalar(t, deriv, pt)
		numeric := fdTime(t, inlined, pt)
		if !closeEnough(analytic, numeric) {
			t.Errorf("at %+v: analytic=%v finite-diff=%v", pt, analytic, numeric)
		}
	}
}

// TestNamedDeclChain differentiates through a named intermediate that depends
// on time: a = sin(time); target = a * a.
func TestNamedDeclChain(t *testing.T) {
	d := &derivative{
		decls: map[string]*program.Operand{"a": un(program.Operation_SIN, ref(program.TimeVarName))},
		funcs: map[string][]*program.Function{},
	}
	target := mul(ref("a"), ref("a"))
	deriv, err := d.temporal(target)
	if err != nil {
		t.Fatalf("temporal: %v", err)
	}
	// The result references the live declaration `a` (so sliders/uniforms stay
	// live); substitute its body to evaluate numerically.
	deriv = substRef(deriv, "a", un(program.Operation_SIN, ref(program.TimeVarName)))
	inlined := mul(un(program.Operation_SIN, ref(program.TimeVarName)), un(program.Operation_SIN, ref(program.TimeVarName)))
	for _, pt := range []point{{0, 0, 0, 0.4}, {0, 0, 0, 1.7}} {
		analytic := evalScalar(t, deriv, pt)
		numeric := fdTime(t, inlined, pt)
		if !closeEnough(analytic, numeric) {
			t.Errorf("at %+v: analytic=%v finite-diff=%v", pt, analytic, numeric)
		}
	}
}

func TestExpandLeavesDerivInFunctionBody(t *testing.T) {
	// f() { return time.dtime; } -- a DERIV inside a function body is not
	// expanded; Expand leaves it in place and must not block a package whose
	// top-level expressions are valid.
	fn := &program.Function{
		Returns: []*program.Operand{nary(program.Operation_DERIV, ref(program.TimeVarName), ref(program.TimeVarName))},
	}
	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				FnTypes: []*program.FnType{
					{Name: "f", Groups: []*program.FnGroup{{ParamCount: 0, Functions: []*program.Function{fn}}}},
				},
				Returns: []*program.Operand{nary(program.Operation_DERIV, mul(ref(program.TimeVarName), num(2)), ref(program.TimeVarName))},
			},
		},
	}
	prog, err := Expand(&program.Program{Primary: pkg})
	if err != nil {
		t.Fatalf("Expand should not error on a function-body DERIV: %v", err)
	}
	out := prog.Primary
	if operandHasOperator(out.Files[0].Returns[0]) {
		t.Error("top-level return DERIV should have been expanded")
	}
	body := out.Files[0].FnTypes[0].Groups[0].Functions[0].Returns[0]
	if !operandHasOperator(body) {
		t.Error("function-body DERIV should be left in place")
	}
}

func TestExpandNoOperatorIsNoOp(t *testing.T) {
	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Returns: []*program.Operand{mul(ref(program.TimeVarName), num(2))},
			},
		},
	}
	prog := &program.Program{Primary: pkg}
	out, err := Expand(prog)
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	// Note: We might clone even if no operators are modified if packageHasOperator returned true,
	// but here since packageHasOperator is false, it returns the same pointer.
	if out != prog {
		t.Error("Expand should return the same program pointer when no operator is present")
	}
}

// --- finite differences -----------------------------------------------------

const fdStep = 1e-6

func fdTime(t *testing.T, op *program.Operand, pt point) float64 {
	hi, lo := pt, pt
	hi.time += fdStep
	lo.time -= fdStep
	return (evalScalar(t, op, hi) - evalScalar(t, op, lo)) / (2 * fdStep)
}

func fdSpatial(t *testing.T, op *program.Operand, pt point, axis int) float64 {
	hi, lo := pt, pt
	switch axis {
	case 0:
		hi.x += fdStep
		lo.x -= fdStep
	case 1:
		hi.y += fdStep
		lo.y -= fdStep
	case 2:
		hi.z += fdStep
		lo.z -= fdStep
	}
	return (evalScalar(t, op, hi) - evalScalar(t, op, lo)) / (2 * fdStep)
}

func closeEnough(a, b float64) bool {
	return math.Abs(a-b) <= 1e-4*(1+math.Abs(b))
}

// substRef replaces every Reference(name) in op with replacement (test helper
// used because derivative results now keep references to live declarations).
func substRef(op *program.Operand, name string, replacement *program.Operand) *program.Operand {
	if op == nil {
		return nil
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		if t.Reference.Name == name {
			return replacement
		}
		return op
	case *program.Operand_Expression:
		ops := make([]*program.Operand, len(t.Expression.Operands))
		for i, sub := range t.Expression.Operands {
			ops[i] = substRef(sub, name, replacement)
		}
		return &program.Operand{T: &program.Operand_Expression{Expression: &program.Expression{
			Operation: t.Expression.Operation, Operands: ops}}}
	case *program.Operand_FnCall:
		args := make([]*program.Operand, len(t.FnCall.Arguments))
		for i, a := range t.FnCall.Arguments {
			args[i] = substRef(a, name, replacement)
		}
		return &program.Operand{T: &program.Operand_FnCall{FnCall: &program.FnCall{
			Reference: t.FnCall.Reference, Arguments: args}}}
	}
	return op
}

func evalScalar(t *testing.T, op *program.Operand, pt point) float64 {
	t.Helper()
	v, err := evalNum(op, pt.vars())
	if err != nil {
		t.Fatalf("evalNum: %v\n%s", err, op.PrettyPrint(0))
	}
	if len(v) != 1 {
		t.Fatalf("expected scalar, got len %d: %s", len(v), op.PrettyPrint(0))
	}
	return v[0]
}

func evalVec(t *testing.T, op *program.Operand, pt point) []float64 {
	t.Helper()
	v, err := evalNum(op, pt.vars())
	if err != nil {
		t.Fatalf("evalNum: %v\n%s", err, op.PrettyPrint(0))
	}
	return v
}

// --- numeric operand evaluator (test-only) ----------------------------------
//
// Evaluates the operand forms the derivative Domain produces, including SELECT
// and swizzle ACCESS, which the production eval.Const helper does not cover.

func evalNum(op *program.Operand, vars map[string][]float64) ([]float64, error) {
	switch v := op.T.(type) {
	case *program.Operand_Value:
		return []float64{v.Value}, nil
	case *program.Operand_Reference:
		val, ok := vars[v.Reference.Name]
		if !ok {
			return nil, fmt.Errorf("unbound reference %q", v.Reference.Name)
		}
		return val, nil
	case *program.Operand_Expression:
		return evalExprNum(v.Expression, vars)
	case *program.Operand_FnCall:
		return evalCallNum(v.FnCall, vars)
	}
	return nil, fmt.Errorf("unsupported operand %T", op.T)
}

func evalExprNum(e *program.Expression, vars map[string][]float64) ([]float64, error) {
	switch e.Operation {
	case program.Operation_COLLECT:
		out := make([]float64, 0, len(e.Operands))
		for _, o := range e.Operands {
			v, err := evalNum(o, vars)
			if err != nil {
				return nil, err
			}
			out = append(out, v...)
		}
		return out, nil
	case program.Operation_ACCESS:
		base, err := evalNum(e.Operands[0], vars)
		if err != nil {
			return nil, err
		}
		out := []float64{}
		for _, idx := range e.Operands[1:] {
			ref, ok := idx.T.(*program.Operand_Reference)
			if !ok {
				return nil, fmt.Errorf("ACCESS index must be a swizzle reference")
			}
			for _, c := range ref.Reference.Name {
				i := map[rune]int{'x': 0, 'y': 1, 'z': 2, 'w': 3}[c]
				if i >= len(base) {
					return nil, fmt.Errorf("swizzle %q out of range for len %d", ref.Reference.Name, len(base))
				}
				out = append(out, base[i])
			}
		}
		return out, nil
	case program.Operation_SELECT:
		cond, err := evalNum(e.Operands[0], vars)
		if err != nil {
			return nil, err
		}
		branches := e.Operands[1:]
		var pick *program.Operand
		switch {
		case len(branches) == 2:
			if cond[0] > 0 {
				pick = branches[0]
			} else {
				pick = branches[1]
			}
		case len(branches) == 3:
			switch {
			case cond[0] > 0:
				pick = branches[0]
			case cond[0] == 0:
				pick = branches[1]
			default:
				pick = branches[2]
			}
		default:
			return nil, fmt.Errorf("SELECT with %d branches", len(branches))
		}
		return evalNum(pick, vars)
	}

	args := make([][]float64, len(e.Operands))
	for i, o := range e.Operands {
		v, err := evalNum(o, vars)
		if err != nil {
			return nil, err
		}
		args[i] = v
	}

	switch e.Operation {
	case program.Operation_ADD:
		return fold(args, func(a, b float64) float64 { return a + b })
	case program.Operation_SUB:
		return fold(args, func(a, b float64) float64 { return a - b })
	case program.Operation_MUL:
		return fold(args, func(a, b float64) float64 { return a * b })
	case program.Operation_DIV:
		return fold(args, func(a, b float64) float64 { return a / b })
	case program.Operation_MODULO:
		return fold(args, math.Mod)
	case program.Operation_MIN:
		return fold(args, math.Min)
	case program.Operation_MAX:
		return fold(args, math.Max)
	case program.Operation_POW:
		return fold(args, math.Pow)
	case program.Operation_ATAN2:
		return fold(args, math.Atan2)
	case program.Operation_NEG:
		return unaryNum(args, func(x float64) float64 { return -x })
	case program.Operation_ABS:
		return unaryNum(args, math.Abs)
	case program.Operation_SQUARE:
		return unaryNum(args, func(x float64) float64 { return x * x })
	case program.Operation_SQRT:
		return unaryNum(args, math.Sqrt)
	case program.Operation_FLOOR:
		return unaryNum(args, math.Floor)
	case program.Operation_CEIL:
		return unaryNum(args, math.Ceil)
	case program.Operation_ROUND:
		return unaryNum(args, math.Round)
	case program.Operation_EXP:
		return unaryNum(args, math.Exp)
	case program.Operation_LN:
		return unaryNum(args, math.Log)
	case program.Operation_SIN:
		return unaryNum(args, math.Sin)
	case program.Operation_COS:
		return unaryNum(args, math.Cos)
	case program.Operation_TAN:
		return unaryNum(args, math.Tan)
	case program.Operation_ASIN:
		return unaryNum(args, math.Asin)
	case program.Operation_ACOS:
		return unaryNum(args, math.Acos)
	case program.Operation_ATAN:
		return unaryNum(args, math.Atan)
	case program.Operation_GT:
		return cmp(args, func(a, b float64) bool { return a > b })
	case program.Operation_GE:
		return cmp(args, func(a, b float64) bool { return a >= b })
	case program.Operation_LT:
		return cmp(args, func(a, b float64) bool { return a < b })
	case program.Operation_LE:
		return cmp(args, func(a, b float64) bool { return a <= b })
	case program.Operation_EQ:
		return cmp(args, func(a, b float64) bool { return a == b })
	case program.Operation_NE:
		return cmp(args, func(a, b float64) bool { return a != b })
	}
	return nil, fmt.Errorf("test evaluator: unsupported operation %s", e.Operation.Name())
}

func evalCallNum(fc *program.FnCall, vars map[string][]float64) ([]float64, error) {
	args := make([][]float64, len(fc.Arguments))
	for i, a := range fc.Arguments {
		v, err := evalNum(a, vars)
		if err != nil {
			return nil, err
		}
		args[i] = v
	}
	switch fc.Reference.Name {
	case "length":
		return []float64{vecLen(args[0])}, nil
	case "distance":
		diff := make([]float64, len(args[0]))
		for i := range diff {
			diff[i] = args[0][i] - args[1][i]
		}
		return []float64{vecLen(diff)}, nil
	case "dot":
		var s float64
		for i := range args[0] {
			s += args[0][i] * args[1][i]
		}
		return []float64{s}, nil
	case "cross":
		a, b := args[0], args[1]
		return []float64{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}, nil
	case "normalize":
		l := vecLen(args[0])
		out := make([]float64, len(args[0]))
		for i, x := range args[0] {
			out[i] = x / l
		}
		return out, nil
	case "clamp":
		return unaryNum(args[:1], func(x float64) float64 { return math.Max(args[1][0], math.Min(args[2][0], x)) })
	case "sign":
		return unaryNum(args, func(x float64) float64 {
			if x > 0 {
				return 1
			} else if x < 0 {
				return -1
			}
			return 0
		})
	case "trunc":
		return unaryNum(args, math.Trunc)
	case "fract":
		return unaryNum(args, func(x float64) float64 { _, f := math.Modf(x); return f })
	case "exp2":
		return unaryNum(args, math.Exp2)
	case "sinh":
		return unaryNum(args, math.Sinh)
	case "cosh":
		return unaryNum(args, math.Cosh)
	case "tanh":
		return unaryNum(args, math.Tanh)
	case "asinh":
		return unaryNum(args, math.Asinh)
	case "acosh":
		return unaryNum(args, math.Acosh)
	case "atanh":
		return unaryNum(args, math.Atanh)
	case "radians":
		return unaryNum(args, func(x float64) float64 { return x * math.Pi / 180 })
	case "degrees":
		return unaryNum(args, func(x float64) float64 { return x * 180 / math.Pi })
	}
	return nil, fmt.Errorf("test evaluator: unsupported call %q", fc.Reference.Name)
}

func fold(args [][]float64, op func(a, b float64) float64) ([]float64, error) {
	acc := append([]float64(nil), args[0]...)
	for i := 1; i < len(args); i++ {
		var err error
		acc, err = broadcast(acc, args[i], op)
		if err != nil {
			return nil, err
		}
	}
	return acc, nil
}

func broadcast(a, b []float64, op func(a, b float64) float64) ([]float64, error) {
	switch {
	case len(a) == len(b):
		out := make([]float64, len(a))
		for i := range a {
			out[i] = op(a[i], b[i])
		}
		return out, nil
	case len(a) == 1:
		out := make([]float64, len(b))
		for i := range b {
			out[i] = op(a[0], b[i])
		}
		return out, nil
	case len(b) == 1:
		out := make([]float64, len(a))
		for i := range a {
			out[i] = op(a[i], b[0])
		}
		return out, nil
	}
	return nil, fmt.Errorf("cannot broadcast len %d and %d", len(a), len(b))
}

func unaryNum(args [][]float64, op func(float64) float64) ([]float64, error) {
	out := make([]float64, len(args[0]))
	for i, x := range args[0] {
		out[i] = op(x)
	}
	return out, nil
}

func cmp(args [][]float64, pred func(a, b float64) bool) ([]float64, error) {
	if pred(args[0][0], args[1][0]) {
		return []float64{1}, nil
	}
	return []float64{0}, nil
}

func vecLen(v []float64) float64 {
	var s float64
	for _, x := range v {
		s += x * x
	}
	return math.Sqrt(s)
}
