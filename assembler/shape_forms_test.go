package assembler_test

import "testing"

// The centred torus and octahedron are the origin forms moved to the centre.
func TestCentredTorusAndOctahedron(t *testing.T) {
	for _, c := range []struct{ centred, moved string }{
		{"torus([3, -2, 1], 5, 1)", "remap(torus(5, 1), p - [3, -2, 1])"},
		{"octahedron([3, -2, 1], 4)", "remap(octahedron(4), p - [3, -2, 1])"},
	} {
		if d := differenceAt(t, c.centred, c.moved, samplePoints); d > 1e-9 {
			t.Errorf("%s differs from %s by %g summed over the sample points", c.centred, c.moved, d)
		}
	}
}

// A difference's list form on [target, tool] is its two-field form, and its
// operator form, called on the list, is its list form.
func TestDifferenceListAndOperatorForms(t *testing.T) {
	const a, b, c = "sphere(6)", "boxCentered([4, 0, 0], [6, 6, 6])", "sphere([0, 3, 0], 2)"
	const prelude = "applied(op, fields) { return op(fields); }\n"
	for _, f := range []struct{ list, pair, op string }{
		{"differenceSharp([" + a + ", " + b + "])", "differenceSharp(" + a + ", " + b + ")", "applied(differenceSharp, [" + a + ", " + b + "])"},
		{"differenceRound([" + a + ", " + b + "], 0.7)", "differenceRound(" + a + ", " + b + ", 0.7)", "applied(differenceRound(0.7), [" + a + ", " + b + "])"},
		{"differenceLp([" + a + ", " + b + "], 0.7, 3)", "differenceLp(" + a + ", " + b + ", 0.7, 3)", "applied(differenceLp(0.7, 3), [" + a + ", " + b + "])"},
		{"differenceExponential([" + a + ", " + b + "], 0.7)", "differenceExponential(" + a + ", " + b + ", 0.7)", "applied(differenceExponential(0.7), [" + a + ", " + b + "])"},
		{"differenceChamfer([" + a + ", " + b + "], 0.7)", "differenceChamfer(" + a + ", " + b + ", 0.7)", "applied(differenceChamfer(0.7), [" + a + ", " + b + "])"},
		// The chamfer-blend families' list forms are not their two-field forms under a
		// non-sharp blend (the list form blends edge and vertex planes the pair
		// form leaves out), so these compare with the intersection they are made of.
		{"differenceChamferRound([" + a + ", " + b + "], 0.7, 0.3)", "intersectChamferRound([" + a + ", neg(" + b + ")], 0.7, 0.3)", "applied(differenceChamferRound(0.7, 0.3), [" + a + ", " + b + "])"},
		{"differenceChamferBlend([" + a + ", " + b + "], 0.7, intersectRound(0.3))", "intersectChamferBlend([" + a + ", neg(" + b + ")], 0.7, intersectRound(0.3))", "applied(differenceChamferBlend(0.7, intersectRound(0.3)), [" + a + ", " + b + "])"},
		{"differenceRvachev([" + a + ", " + b + "], 0.7)", "differenceRvachev(" + a + ", " + b + ", 0.7)", "applied(differenceRvachev(0.7), [" + a + ", " + b + "])"},
	} {
		if d := differenceAtWith(t, prelude, f.list, f.pair, samplePoints); d > 1e-9 {
			t.Errorf("%s differs from %s by %g summed over the sample points", f.list, f.pair, d)
		}
		if d := differenceAtWith(t, prelude, f.op, f.list, samplePoints); d > 1e-9 {
			t.Errorf("%s differs from %s by %g summed over the sample points", f.op, f.list, d)
		}
	}
	// Three fields: the target minus both tools.
	three := "differenceRound([" + a + ", " + b + ", " + c + "], 0.7)"
	want := "intersectRound([" + a + ", neg(" + b + "), neg(" + c + ")], 0.7)"
	if d := differenceAt(t, three, want, samplePoints); d > 1e-9 {
		t.Errorf("%s differs from %s by %g", three, want, d)
	}
}
