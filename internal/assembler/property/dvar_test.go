package property

import (
	"math"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// The can of examples/can.gsl: a cylinder of radius r whose height follows
// from a material budget, the side wall and the two lids costing differently
// per unit area. Lengths are in r's units; volume and cost scale with them.
const (
	canBudget   = 40000.0
	canSideCost = 1.0
	canEndCost  = 2.0
)

// canHeight is h(r): the wall the budget pays for once the lids are bought,
// cost = sideCost·2πrh + endCost·2πr².
func canHeight(r float64) float64 {
	return (canBudget - canEndCost*2*math.Pi*r*r) / (canSideCost * 2 * math.Pi * r)
}

func canHeightPrime(r float64) float64 {
	return -canBudget/(2*math.Pi*canSideCost*r*r) - canEndCost/canSideCost
}

func canVolume(r float64) float64 { return math.Pi * r * r * canHeight(r) }

// canDVolume is dV/dr by hand: the side's 2πrh plus the lids' πr²·h'.
func canDVolume(r float64) float64 {
	return 2*math.Pi*r*canHeight(r) + math.Pi*r*r*canHeightPrime(r)
}

// canOptimum is the radius where dV/dr vanishes, h = 2r·endCost/sideCost,
// which the budget constraint turns into budget = 6π·endCost·r².
func canOptimum() float64 { return math.Sqrt(canBudget / (6 * math.Pi * canEndCost)) }

// canDecls declares the can in GSL operands: r a slider, h and volume in
// terms of it, and the can itself max(radial, cap) with radial = √(x²+y²) − r
// and cap = |z| − h/2.
func canDecls() map[string]*program.Operand {
	pi := num(math.Pi)
	r := ref("r")
	h := div(
		sub(ref("budget"), mul(mul(mul(ref("endCost"), num(2)), pi), mul(r, r))),
		mul(mul(mul(ref("sideCost"), num(2)), pi), r))
	volume := mul(mul(pi, mul(r, r)), ref("h"))
	radial := sub(sqrt(add(mul(ref(program.XVarName), ref(program.XVarName)), mul(ref(program.YVarName), ref(program.YVarName)))), r)
	cap := sub(un(program.Operation_ABS, ref(program.ZVarName)), div(ref("h"), num(2)))
	return map[string]*program.Operand{
		"r":        num(30),
		"budget":   num(canBudget),
		"sideCost": num(canSideCost),
		"endCost":  num(canEndCost),
		"h":        h,
		"volume":   volume,
		"can":      bin(program.Operation_MAX, radial, cap),
	}
}

// canVars binds every can declaration to its value at radius r and point pt,
// for evalNum: the derivative keeps a declaration as its live reference.
func canVars(r float64, pt point) map[string][]float64 {
	vars := pt.vars()
	vars["r"] = []float64{r}
	vars["budget"] = []float64{canBudget}
	vars["sideCost"] = []float64{canSideCost}
	vars["endCost"] = []float64{canEndCost}
	vars["h"] = []float64{canHeight(r)}
	vars["volume"] = []float64{canVolume(r)}
	return vars
}

func evalAt(t *testing.T, op *program.Operand, vars map[string][]float64) []float64 {
	t.Helper()
	v, err := evalNum(op, vars)
	if err != nil {
		t.Fatalf("evalNum: %v\n%s", err, op.PrettyPrint(0))
	}
	return v
}

func scalarAt(t *testing.T, op *program.Operand, vars map[string][]float64) float64 {
	t.Helper()
	v := evalAt(t, op, vars)
	if len(v) != 1 {
		t.Fatalf("expected scalar, got len %d: %s", len(v), op.PrettyPrint(0))
	}
	return v[0]
}

// TestDVarChainThroughDependent differentiates the analytic volume with
// respect to r, where h is itself written in terms of r: the chain rule must
// run through h. Checked against the hand derivative and a finite difference,
// and the derivative must vanish at the optimum.
func TestDVarChainThroughDependent(t *testing.T) {
	d := newDerivative()
	d.decls = canDecls()
	dv, err := d.wrt(ref("volume"), "r")
	if err != nil {
		t.Fatalf("wrt: %v", err)
	}
	for _, r := range []float64{20, 30, 45} {
		got := scalarAt(t, dv, canVars(r, point{}))
		if want := canDVolume(r); !closeEnough(got, want) {
			t.Errorf("r=%v: dV/dr=%v, want %v", r, got, want)
		}
		if fd := (canVolume(r+fdStep) - canVolume(r-fdStep)) / (2 * fdStep); !closeEnough(got, fd) {
			t.Errorf("r=%v: dV/dr=%v, finite difference %v", r, got, fd)
		}
	}
	rStar := canOptimum()
	if got := scalarAt(t, dv, canVars(rStar, point{})); math.Abs(got) > 1e-6*canVolume(rStar)/rStar {
		t.Errorf("at the optimum r=%v: dV/dr=%v, want 0", rStar, got)
	}
}

// TestDVarIndependentOfAssignment: the variable differentiated with respect to
// is independent whatever it was assigned. With w = 2r and f = w², f.d(w) is
// 2w (r held fixed) and f.d(r) is 8r (through w).
func TestDVarIndependentOfAssignment(t *testing.T) {
	d := newDerivative()
	d.decls = map[string]*program.Operand{
		"r": num(3),
		"w": mul(num(2), ref("r")),
		"f": mul(ref("w"), ref("w")),
	}
	vars := map[string][]float64{"r": {3}, "w": {6}, "f": {36}}

	dfdw, err := d.wrt(ref("f"), "w")
	if err != nil {
		t.Fatalf("wrt w: %v", err)
	}
	if got := scalarAt(t, dfdw, vars); !closeEnough(got, 12) {
		t.Errorf("df/dw=%v, want 12", got)
	}
	dfdr, err := d.wrt(ref("f"), "r")
	if err != nil {
		t.Fatalf("wrt r: %v", err)
	}
	if got := scalarAt(t, dfdr, vars); !closeEnough(got, 24) {
		t.Errorf("df/dr=%v, want 24", got)
	}
}

// TestDVarCanSurfaceIntegral checks the FIELD derivative against the volume:
// the surface moves at normal speed −f_r/|∇f| as r grows, so integrating that
// over the can's surface is dV/dr. The side (f_r = −1) contributes 2πrh and
// the lids (f_r = −h'/2 each, through h) contribute πr²h', the same two terms
// the analytic volume differentiates to. Midpoint quadrature over the known
// surface; the integrand is piecewise constant, so the rule is exact up to
// rounding.
func TestDVarCanSurfaceIntegral(t *testing.T) {
	d := newDerivative()
	d.decls = canDecls()
	fr, err := d.wrt(ref("can"), "r")
	if err != nil {
		t.Fatalf("wrt: %v", err)
	}
	grad, err := d.gradient(ref("can"))
	if err != nil {
		t.Fatalf("gradient: %v", err)
	}
	dv, err := d.wrt(ref("volume"), "r")
	if err != nil {
		t.Fatalf("wrt volume: %v", err)
	}

	const nTheta, nZ, nRho = 48, 24, 24
	for _, r := range []float64{25, canOptimum(), 40} {
		h := canHeight(r)
		speed := func(x, y, z float64) float64 {
			vars := canVars(r, point{x, y, z, 0})
			g := evalAt(t, grad, vars)
			if len(g) != 3 {
				t.Fatalf("gradient has %d components", len(g))
			}
			norm := math.Sqrt(g[0]*g[0] + g[1]*g[1] + g[2]*g[2])
			return -scalarAt(t, fr, vars) / norm
		}
		var sum float64
		for i := 0; i < nTheta; i++ {
			theta := (float64(i) + 0.5) * 2 * math.Pi / nTheta
			c, s := math.Cos(theta), math.Sin(theta)
			// Side: (r cosθ, r sinθ, z), dA = r dθ dz.
			for j := 0; j < nZ; j++ {
				z := -h/2 + (float64(j)+0.5)*h/nZ
				sum += speed(r*c, r*s, z) * r * (2 * math.Pi / nTheta) * (h / nZ)
			}
			// Lids: (ρ cosθ, ρ sinθ, ±h/2), dA = ρ dρ dθ.
			for k := 0; k < nRho; k++ {
				rho := (float64(k) + 0.5) * r / nRho
				w := rho * (r / nRho) * (2 * math.Pi / nTheta)
				sum += speed(rho*c, rho*s, h/2) * w
				sum += speed(rho*c, rho*s, -h/2) * w
			}
		}
		if want := canDVolume(r); !closeEnough(sum, want) {
			t.Errorf("r=%v: surface integral %v, analytic dV/dr %v", r, sum, want)
		}
		if viaVolume := scalarAt(t, dv, canVars(r, point{})); !closeEnough(sum, viaVolume) {
			t.Errorf("r=%v: surface integral %v, volume.d(r) %v", r, sum, viaVolume)
		}
	}
}

// TestDVarVector differentiates with respect to a literal vector variable, one
// seed per component. For a scalar field the partials collect into a vector:
// with f = |p − c|, f.d(c) = −(p − c)/|p − c|, checked against finite
// differences in each component of c. For a vector field they collect into a
// list of vectors, the Jacobian by columns: (p − c).d(c) = −I. Component access
// on the variable reads through the seed: (c.x·c.y).d(c) = [c.y, c.x, 0].
func TestDVarVector(t *testing.T) {
	d := newDerivative()
	d.decls = map[string]*program.Operand{
		"c": nary(program.Operation_COLLECT, num(1), num(2), num(3)),
		"f": call("length", sub(ref(program.PVarName), ref("c"))),
		"g": sub(ref(program.PVarName), ref("c")),
		"q": mul(nary(program.Operation_ACCESS, ref("c"), ref(program.XVarName)), nary(program.Operation_ACCESS, ref("c"), ref(program.YVarName))),
	}
	pt := point{3, -4, 1, 0}
	varsAt := func(c []float64) map[string][]float64 {
		vars := pt.vars()
		vars["c"] = c
		return vars
	}
	c := []float64{1, 2, 3}

	// Scalar field: a vec3 of partials, each a finite difference in c_k.
	dfdc, err := d.wrt(ref("f"), "c")
	if err != nil {
		t.Fatalf("wrt f: %v", err)
	}
	got := evalAt(t, dfdc, varsAt(c))
	if len(got) != 3 {
		t.Fatalf("f.d(c) has %d components, want 3", len(got))
	}
	for k := 0; k < 3; k++ {
		hi, lo := append([]float64{}, c...), append([]float64{}, c...)
		hi[k] += fdStep
		lo[k] -= fdStep
		fd := (scalarAt(t, d.decls["f"], varsAt(hi)) - scalarAt(t, d.decls["f"], varsAt(lo))) / (2 * fdStep)
		if !closeEnough(got[k], fd) {
			t.Errorf("df/dc[%d]=%v, finite difference %v", k, got[k], fd)
		}
	}

	// Vector field: a list of three vec3 columns, here −I.
	dgdc, err := d.wrt(ref("g"), "c")
	if err != nil {
		t.Fatalf("wrt g: %v", err)
	}
	jac := evalAt(t, dgdc, varsAt(c))
	if len(jac) != 9 {
		t.Fatalf("(p - c).d(c) flattens to %d values, want 9", len(jac))
	}
	for i, v := range jac {
		want := 0.0
		if i%4 == 0 { // the diagonal of a 3x3, flattened column by column
			want = -1
		}
		if !closeEnough(v, want) {
			t.Errorf("jacobian[%d]=%v, want %v", i, v, want)
		}
	}

	// Component access through the seed.
	dqdc, err := d.wrt(ref("q"), "c")
	if err != nil {
		t.Fatalf("wrt q: %v", err)
	}
	got = evalAt(t, dqdc, varsAt(c))
	for k, want := range []float64{2, 1, 0} {
		if k >= len(got) || !closeEnough(got[k], want) {
			t.Errorf("(c.x*c.y).d(c)=%v, want [2 1 0]", got)
			break
		}
	}
}

// TestDVarErrors: the variable must be a top-level scalar or literal vector,
// and the operator must be used at top level, not in a function body (a
// companion).
func TestDVarErrors(t *testing.T) {
	d := newDerivative()
	d.decls = canDecls()
	d.decls["m"] = nary(program.Operation_COLLECT,
		nary(program.Operation_COLLECT, num(1), num(0)),
		nary(program.Operation_COLLECT, num(0), num(1)))

	if _, err := d.wrt(ref("can"), "nope"); err == nil || !strings.Contains(err.Error(), "not a top-level variable") {
		t.Errorf("unknown variable: got %v", err)
	}
	if _, err := d.wrt(ref("can"), "m"); err == nil || !strings.Contains(err.Error(), "list of vectors") {
		t.Errorf("list-of-vectors variable: got %v", err)
	}

	companion := newDerivative()
	companion.decls = canDecls()
	companion.params = map[string]*paramTangent{}
	if _, err := companion.wrt(ref("can"), "r"); err == nil || !strings.Contains(err.Error(), "inside a function body") {
		t.Errorf("companion body: got %v", err)
	}
}

// TestDVarAliases: through expandDeriv, .d(p) is the gradient and .d(time) the
// temporal derivative.
func TestDVarAliases(t *testing.T) {
	d := newDerivative()
	d.decls = canDecls()
	pt := point{3, -4, 1, 0.5}

	viaD, err := d.expandDeriv([]*program.Operand{ref("can"), ref(program.PVarName)})
	if err != nil {
		t.Fatalf("expandDeriv p: %v", err)
	}
	direct, err := d.gradient(ref("can"))
	if err != nil {
		t.Fatalf("gradient: %v", err)
	}
	a, b := evalAt(t, viaD, canVars(30, pt)), evalAt(t, direct, canVars(30, pt))
	if len(a) != 3 || len(b) != 3 {
		t.Fatalf("gradients of length %d and %d", len(a), len(b))
	}
	for i := range a {
		if !closeEnough(a[i], b[i]) {
			t.Errorf(".d(p)[%d]=%v, .dp[%d]=%v", i, a[i], i, b[i])
		}
	}

	viaD, err = d.expandDeriv([]*program.Operand{ref("can"), ref(program.TimeVarName)})
	if err != nil {
		t.Fatalf("expandDeriv time: %v", err)
	}
	if got := scalarAt(t, viaD, canVars(30, pt)); got != 0 {
		t.Errorf(".d(time) of a static can = %v, want 0", got)
	}
}
