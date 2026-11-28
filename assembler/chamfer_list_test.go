package assembler_test

import (
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
)

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

// A sharp chamfered box's .second names the runner-up piece, a face slab or a
// chamfer plane, with that piece's gradient, and carries no sentinel: the
// chamfer helpers' constant-index ternaries are settled before .second is
// expanded, so no 1e30 reaches a min or max.
func TestSharpChamferSecondNamesTheRunnerUp(t *testing.T) {
	const prelude = "slab = box([-14.0, -7.0, 4.0], [14.0, 7.0, 15.0], intersectChamfer(0.5));\n"
	for _, c := range []struct {
		at                 [3]float64
		second, dx, dy, dz string
	}{
		{[3]float64{0, 3, 10}, "-5.0", "0", "0", "1"},                            // the z slab
		{[3]float64{0, 1, 14}, "-6.5 / sqrt2", "0", "1 / sqrt2", "1 / sqrt2"},    // the top/y chamfer plane
		{[3]float64{0, 6.9, 10}, "-4.6 / sqrt2", "0", "1 / sqrt2", "1 / sqrt2"},  // the same plane, from the y face
		{[3]float64{13, 0.5, 12}, "-3.5 / sqrt2", "1 / sqrt2", "0", "1 / sqrt2"}, // the x/z chamfer plane
	} {
		pts := [][3]float64{c.at}
		for _, e := range []struct{ expr, want string }{
			{"slab.second", c.second}, {"slab.second.dp.x", c.dx}, {"slab.second.dp.y", c.dy}, {"slab.second.dp.z", c.dz},
		} {
			if d := differenceAtWith(t, prelude, e.expr, e.want, pts); d > 1e-9 {
				t.Errorf("%s at %v is %g from %s", e.expr, c.at, d, e.want)
			}
		}
	}

	res, err := assembler.Assemble([]string{prelude + "s = slab.second;\nreturn s;\n"})
	if err != nil {
		t.Fatal(err)
	}
	doc, _, _ := strings.Cut(res.Program.PrettyPrint(0), "// ====")
	if n := strings.Count(doc, "1e+30"); n > 0 {
		t.Errorf("slab.second carries the no-second sentinel %d times:\n%s", n, doc)
	}
}
