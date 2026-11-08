package gsleval

import (
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	fidlib "gradientcontrol.com/omega/core/internal/fidget/lib"
	"gradientcontrol.com/omega/core/internal/source"
)

// TestSecondClosest evaluates the .second (second-closest distance) property
// end-to-end through parse -> property.Expand -> fidget, against the omega
// boolean combiners. Three unit spheres on the x-axis: a@-2, b@0, c@3. At the
// origin the distances are da=1, db=-1, dc=2, so the second-nearest surface is a
// (distance 1); the second-nearest of a c-minus-b difference is c (distance 2).
func TestSecondClosest(t *testing.T) {
	srcMgr := source.NewLocalManager("../../stdlib")

	base := `import . "omega";
a = sphere([-2,0,0], 1);
b = sphere([0,0,0], 1);
c = sphere([3,0,0], 1);
balls = unionRound([a, b], 0.5);
`
	cases := []struct {
		expr string
		want float32
	}{
		// sharp booleans (bottom out in min/max). min keeps the two smallest, max
		// the two largest; for three operands the 2nd-smallest and 2nd-largest are
		// both the middle (=1). differenceSharp(c,b) = max(c,-b), a 2-way max, so
		// its .second is the smaller of {c, -b} = min(2, 1) = 1.
		{"unionSharp([a,b,c]).second", 1},
		{"intersectSharp([a,b,c]).second", 1},
		{"differenceSharp(c, b).second", 1},
		{"min(a,b,c).second", 1},
		{"max(a,b,c).second", 1},
		{"min(min(a,b),c).second", 1}, // nested min flattens

		// smooth round / LP booleans: first = the blended surface (the combiner's
		// own SDF), second = the nearest piece of the result's topology that the
		// point is not already on. A blend has more pieces than the surfaces it
		// was handed: each input contributes a face, and the fillets between them
		// are pieces in their own right, so the fillet is usually what a point on
		// a face is nearest to.
		//
		// These used to expect the nearest input SIDE, which at the origin is
		// db = -1 — b's own surface, the piece the point is already on. Naming
		// the active piece as the runner-up is the flat fold that the piece rule
		// replaced.
		//
		// Worked from the clamped terms the combiner itself uses. At the origin
		// da=1, db=-1, dc=2, k=0.5. For a union, read in the intersect frame
		// (e = -d): the excesses c = max(e+k, 0) are (0, 1.5, 0), so only b is
		// active and sum = 2.25. A neighbouring piece swaps one term's clamped
		// excess for the part clamped away, u = min(e+k, 0):
		//   a: 2.25 - 0 + 0.5^2 = 2.5   -> sqrt(2.5) - 0.5 = 1.0811 -> -1.0811
		//   b: 2.25 - 2.25 + 0  = 0     -> empties the blend, not a piece
		//   c: 2.25 - 0 + 1.5^2 = 4.5   -> sqrt(4.5) - 0.5 = 1.6213 -> -1.6213
		// The nearest is the a-b fillet at -1.0811.
		{"unionRound([a,b,c], 0.5).second", -1.0811388},
		{"unionRound(a, b, 0.5).second", -1.0811388},
		{"unionRound(a, b, c, 0.5).second", -1.0811388},
		// Intersection reads in its own frame: c = (1.5, 0, 2.5), sum = 8.5, and
		// the three swaps give 2.0, 2.4580 and sqrt(2.25) - 0.5 = 1.0.
		{"intersectRound([a,b,c], 0.5).second", 1},
		// Same swap under the Lp norm: 1.5^4 - 0 + 0.5^4 = 5.125, and
		// 5.125^(1/4) - 0.5 = 1.0046.
		{"unionLp([a,b,c], 0.5, 4).second", -1.0046083},
		// A difference compares {c, -b}; the nearest piece is unchanged from the
		// side reading here.
		{"differenceRound(c, b, 0.5).second", 1},
		{"differenceLp(c, b, 0.5, 4).second", 1},

		// composition: a named round union nested in another (Color's balls/body).
		// The a-b fillet is nearest here too, and balls carries its own runner-up
		// so the inner blend's pieces are not lost.
		{"unionRound([balls, c], 0.5).second", -1.0811388},

		// affine offsets carry through
		{"(min(a,b,c) + 0.5).second", 1.5},
		{"(unionRound([a,b,c], 0.5) - 0.25).second", -1.3311388},

		// a lone surface has no second-closest -> the sentinel
		{"a.second", 1e30},
	}
	for _, tc := range cases {
		src := base + "return " + tc.expr + ";\n"
		res, err := assembler.Assemble([]string{src}, srcMgr)
		if err != nil {
			t.Errorf("%s: assemble: %v", tc.expr, err)
			continue
		}
		out, err := fidlib.EvalGrad(res.Program, []float32{0}, []float32{0}, []float32{0})
		if err != nil {
			t.Errorf("%s: eval: %v", tc.expr, err)
			continue
		}
		got := out[0][0].Value
		if tc.want > 1e20 {
			if got < 1e20 {
				t.Errorf("%s = %v, want the ~1e30 sentinel", tc.expr, got)
			}
			continue
		}
		if d := got - tc.want; d > 1e-3 || d < -1e-3 {
			t.Errorf("%s = %v, want %v", tc.expr, got, tc.want)
		}
	}
}

// TestSecondClosestRoundFillet checks the round-union topology: the blended
// surface is a candidate, so at a fillet point a round union's .second equals the
// nearest input side (= the sharp union), while its own SDF (the blended
// boundary) is strictly nearer — the gap first-to-second being the fillet fill.
func TestSecondClosestRoundFillet(t *testing.T) {
	srcMgr := source.NewLocalManager("../../stdlib")
	src := `import . "omega";
a = sphere([-1,0,0], 1);
b = sphere([1,0,0], 1);
return unionRound([a,b], 0.6).second, unionSharp([a,b]), unionRound([a,b], 0.6);
`
	res, err := assembler.Assemble([]string{src}, srcMgr)
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	// A point above the seam, inside the fillet band.
	out, err := fidlib.EvalGrad(res.Program, []float32{0}, []float32{1.0}, []float32{0})
	if err != nil {
		t.Fatal(err)
	}
	second, sharp, blend := out[0][0].Value, out[1][0].Value, out[2][0].Value
	if d := second - sharp; d > 1e-3 || d < -1e-3 {
		t.Errorf("round .second = %v, want the sharp union %v (nearest side)", second, sharp)
	}
	if blend >= second {
		t.Errorf("in the fillet expected blend SDF %v < .second %v", blend, second)
	}
}
