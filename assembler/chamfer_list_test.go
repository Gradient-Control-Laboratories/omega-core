package assembler_test

import "testing"

// chamferPlanes are three tilted planes meeting near the origin, so the
// sample points fall on faces, edges and the vertex as well as in between.
const chamferPlanes = "a = dot(p, normalize([1.0, 0.2, 0.1])) - 1.0;\n" +
	"b = dot(p, normalize([0.1, 1.0, -0.3])) - 0.5;\n" +
	"c = dot(p, normalize([-0.2, 0.3, 1.0])) - 0.8;\n"

// The chamfer blends over a list are the same field as their two- and
// three-field forms, under every blend: the list form takes the strongest
// edge and vertex planes, which are the binding ones, and nothing else.
func TestChamferBlendListFormsMatchFieldForms(t *testing.T) {
	for _, c := range []struct{ fields, list string }{
		{"intersectChamferRound(a, b, 0.4, 0.6)", "intersectChamferRound([a, b], 0.4, 0.6)"},
		{"intersectChamferRound(a, b, c, 0.4, 0.6)", "intersectChamferRound([a, b, c], 0.4, 0.6)"},
		{"unionChamferRound(a, b, 0.4, 0.6)", "unionChamferRound([a, b], 0.4, 0.6)"},
		{"unionChamferRound(a, b, c, 0.4, 0.6)", "unionChamferRound([a, b, c], 0.4, 0.6)"},
		{"differenceChamferRound(a, b, 0.4, 0.6)", "differenceChamferRound([a, b], 0.4, 0.6)"},
		{"intersectChamferBlend(a, b, c, 0.4, intersectLp(0.6, 3.0))", "intersectChamferBlend([a, b, c], 0.4, intersectLp(0.6, 3.0))"},
		{"intersectChamferBlend(a, b, c, 0.4, intersectExponential(0.6))", "intersectChamferBlend([a, b, c], 0.4, intersectExponential(0.6))"},
		{"unionChamferBlend(a, b, c, 0.4, unionLp(0.6, 3.0))", "unionChamferBlend([a, b, c], 0.4, unionLp(0.6, 3.0))"},
		{"differenceChamferBlend(a, b, 0.4, intersectExponential(0.6))", "differenceChamferBlend([a, b], 0.4, intersectExponential(0.6))"},
	} {
		if d := differenceAtWith(t, chamferPlanes, c.fields, c.list, samplePoints); d > 1e-9 {
			t.Errorf("%s differs from %s by %g summed over the sample points", c.list, c.fields, d)
		}
	}
}

// A round chamfer over a list leaves a face far from every edge where it is:
// the value there is the face's own distance.
func TestChamferRoundListKeepsFarFaces(t *testing.T) {
	const prelude = "a = p.x - 5;\nb = p.y - 5;\nc = p.z - 5;\n"
	points := [][3]float64{{5, -20, -20}, {4.5, -20, -20}, {5.5, -20, -20}}
	if d := differenceAtWith(t, prelude, "intersectChamferRound([a, b, c], 1.0, 1.0)", "a", points); d > 1e-9 {
		t.Errorf("the list form moves a far face: %g summed over the points", d)
	}
}
