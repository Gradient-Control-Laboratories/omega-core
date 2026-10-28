package property

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// Repro for the property Engine's inability to call a thunk-valued parameter.
//
// omega.gsl's higher-order combiners take SDFs as 0-argument function
// parameters and call them internally:
//
//	unionRoundExact(sdf1, sdf2, n1, n2) {
//	  d1 = sdf1();
//	  d2 = sdf2();
//	  ...
//	  n12 = remap(sdf1(), p - d2 * n2);
//	}
//
// and callers pass anonymous thunks, which the parser lifts to named functions
// (gsl_anon_N) and passes by reference. The ordinary build handles this; the
// property abstract interpreter does not. When Engine.inlineCall inlines the
// combiner it binds the thunk arguments as frame *values* — a function
// reference evaluates through lookup's fallthrough to Domain.Opaque, discarding
// the *pb.Function — and then reaches `sdf1()`, whose name resolves against the
// function tables only, so it fails with
//
//	cannot evaluate call to unknown function "sdf1"
//
// The cases below are the derivative-path (.dtime) version of that failure, in
// increasing order of what the fix has to support: a direct call, a call bound
// to a local, and a call under a remap (where the thunk must evaluate in the
// remapped frame, not its defining one).

func assignDecl(name string, body *program.Operand) *program.Assignment {
	return &program.Assignment{
		Variables:  []*program.Variable{{Name: name}},
		Expression: body,
	}
}

// thunkA is `gsl_anon_0() { return x*x + time; }`.
func thunkA() *program.Function {
	return &program.Function{Returns: []*program.Operand{add(mul(ref(program.XVarName), ref(program.XVarName)), ref(program.TimeVarName))}}
}

// thunkB is `gsl_anon_1() { return y*y + 2*time; }`.
func thunkB() *program.Function {
	return &program.Function{Returns: []*program.Operand{add(mul(ref(program.YVarName), ref(program.YVarName)), mul(num(2), ref(program.TimeVarName)))}}
}

// The two thunks' bodies, inlined by hand, for the finite-difference reference.
func bodyA() *program.Operand {
	return add(mul(ref(program.XVarName), ref(program.XVarName)), ref(program.TimeVarName))
}
func bodyB() *program.Operand {
	return add(mul(ref(program.YVarName), ref(program.YVarName)), mul(num(2), ref(program.TimeVarName)))
}

// TestThunkParamCall differentiates through a higher-order combiner that calls
// its function-valued parameters. Points are chosen so the two branches are far
// apart, keeping min() smooth for the finite-difference comparison.
func TestThunkParamCall(t *testing.T) {
	pts := []point{{0.3, -1.2, 0.7, 0.4}, {1.3, 0.2, 0.5, -0.8}}

	cases := []struct {
		name    string
		combine *program.Function
		target  *program.Operand
		inlined *program.Operand
	}{
		{
			// combine(sdf1, sdf2) { return min(sdf1(), sdf2()); }
			name: "direct",
			combine: &program.Function{
				Parameters: []*program.Parameter{{Name: "sdf1"}, {Name: "sdf2"}},
				Returns:    []*program.Operand{bin(program.Operation_MIN, call("sdf1"), call("sdf2"))},
			},
			target:  call("combine", ref("gsl_anon_0"), ref("gsl_anon_1")),
			inlined: bin(program.Operation_MIN, bodyA(), bodyB()),
		},
		{
			// combine(sdf1, sdf2) { d1 = sdf1(); d2 = sdf2(); return min(d1, d2); }
			// -- the shape unionRoundExact actually uses.
			name: "via_local",
			combine: &program.Function{
				Parameters: []*program.Parameter{{Name: "sdf1"}, {Name: "sdf2"}},
				Assignments: []*program.Assignment{
					assignDecl("d1", call("sdf1")),
					assignDecl("d2", call("sdf2")),
				},
				Returns: []*program.Operand{bin(program.Operation_MIN, ref("d1"), ref("d2"))},
			},
			target:  call("combine", ref("gsl_anon_0"), ref("gsl_anon_1")),
			inlined: bin(program.Operation_MIN, bodyA(), bodyB()),
		},
		{
			// combine(sdf1, sdf2) { return remap(sdf1(), [x+time, y, z]) + sdf2(); }
			// -- unionRoundExact's n12/n21: the thunk is called under a
			// coordinate change, so it must evaluate in the remapped frame.
			name: "under_remap",
			combine: &program.Function{
				Parameters: []*program.Parameter{{Name: "sdf1"}, {Name: "sdf2"}},
				Returns: []*program.Operand{add(
					nary(program.Operation_REMAP, call("sdf1"),
						nary(program.Operation_COLLECT, add(ref(program.XVarName), ref(program.TimeVarName)), ref(program.YVarName), ref(program.ZVarName))),
					call("sdf2"))},
			},
			target: call("combine", ref("gsl_anon_0"), ref("gsl_anon_1")),
			// x*x + time under q = [x+time, y, z] is (x+time)*(x+time) + time;
			// time is not remapped.
			inlined: add(
				add(mul(add(ref(program.XVarName), ref(program.TimeVarName)), add(ref(program.XVarName), ref(program.TimeVarName))), ref(program.TimeVarName)),
				bodyB()),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := &derivative{
				decls: map[string]*program.Operand{},
				funcs: map[string][]*program.Function{
					"gsl_anon_0": {thunkA()},
					"gsl_anon_1": {thunkB()},
					"combine":    {c.combine},
				},
			}
			deriv, err := d.temporal(c.target)
			if err != nil {
				t.Fatalf("temporal: %v", err)
			}
			for _, pt := range pts {
				analytic := evalScalar(t, deriv, pt)
				numeric := fdTime(t, c.inlined, pt)
				if !closeEnough(analytic, numeric) {
					t.Errorf("at %+v: analytic=%v finite-diff=%v", pt, analytic, numeric)
				}
			}
		})
	}
}

// TestThunkParamGradient is the same failure on the spatial path (.dp): the
// gradient of a combiner that calls its thunk parameters.
func TestThunkParamGradient(t *testing.T) {
	// combine(sdf1, sdf2) { d1 = sdf1(); d2 = sdf2(); return min(d1, d2); }
	combine := &program.Function{
		Parameters: []*program.Parameter{{Name: "sdf1"}, {Name: "sdf2"}},
		Assignments: []*program.Assignment{
			assignDecl("d1", call("sdf1")),
			assignDecl("d2", call("sdf2")),
		},
		Returns: []*program.Operand{bin(program.Operation_MIN, ref("d1"), ref("d2"))},
	}
	d := &derivative{
		decls: map[string]*program.Operand{},
		funcs: map[string][]*program.Function{
			"gsl_anon_0": {thunkA()},
			"gsl_anon_1": {thunkB()},
			"combine":    {combine},
		},
	}
	target := call("combine", ref("gsl_anon_0"), ref("gsl_anon_1"))
	inlined := bin(program.Operation_MIN, bodyA(), bodyB())

	grad, err := d.gradient(target)
	if err != nil {
		t.Fatalf("gradient: %v", err)
	}
	for _, pt := range []point{{0.3, -1.2, 0.7, 0.4}, {1.3, 0.2, 0.5, -0.8}} {
		g := evalVec(t, grad, pt)
		if len(g) != 3 {
			t.Fatalf("expected vec3 gradient, got len %d", len(g))
		}
		want := [3]float64{
			fdSpatial(t, inlined, pt, 0),
			fdSpatial(t, inlined, pt, 1),
			fdSpatial(t, inlined, pt, 2),
		}
		for k := 0; k < 3; k++ {
			if !closeEnough(g[k], want[k]) {
				t.Errorf("at %+v axis %d: analytic=%v finite-diff=%v", pt, k, g[k], want[k])
			}
		}
	}
}
