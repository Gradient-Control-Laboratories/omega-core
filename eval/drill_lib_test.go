//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// TestBoreLibrary checks the bore package end-to-end: every dimension read
// back out of the field where a drawing would measure it, the undercut clipped
// at its own top rather than eating the wall above it, and the drill point
// where the geometry puts it and not where the depth callout does.
//
// The diameters are document VARIABLES on purpose. A feature closure made in
// bore and folded by bore, from a document, captures them — the case that only
// works because a cross-package closure travels with its captures as arguments.
func TestBoreLibrary(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	const doc = `
dia = 12.0;
deep = 30.0;
cbore = 20.0;

h = drill.hole([
  drill.chamfer(cbore, 0.5),
  drill.bore(cbore, 5.0),
  drill.undercut(21.0, 4.0, 5.0),
  drill.chamferAt(dia, 0.5, 5.0),
  drill.blind(dia, deep)
], 0.0);

thru = drill.hole(drill.thru(6.6));
block = boxCentered([0, 0, -20], [60, 60, 40]);
solid = max(block, neg(h));
return solid;
`

	eval := func(expr string, pt [3]float64) float64 {
		t.Helper()
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{expr}, [][3]float64{pt})
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		return float64(groups[0].Samples[0][0].Value)
	}
	near := func(expr string, pt [3]float64, want, tol float64) {
		t.Helper()
		got := eval(expr, pt)
		if math.IsNaN(got) || math.Abs(got-want) > tol {
			t.Errorf("%s at %v = %.4f, want %.3f±%.3f", expr, pt, got, want, tol)
		}
	}
	outside := func(expr string, pt [3]float64) {
		t.Helper()
		got := eval(expr, pt)
		if math.IsNaN(got) || got <= 0 {
			t.Errorf("%s at %v = %.4f, want > 0 (outside the void)", expr, pt, got)
		}
	}

	// The Ø12 drill wall and the Ø20 counterbore wall, each read at its radius.
	near("h", [3]float64{6, 0, -20}, 0, 1e-3)
	near("h", [3]float64{10, 0, -2}, 0, 1e-3)
	// The counterbore floor, 5 below the face — but only outside the drill,
	// since inside it the drilled hole is the deeper void.
	near("h", [3]float64{8, 0, -5}, 0, 1e-3)
	near("h", [3]float64{5, 0, -5}, -1, 1e-3)

	// The drill point. `deep` reaches the full-diameter shoulder at z -30, the
	// way a drawing dimensions it; the 118-degree point adds
	// (12/2)/tan(59) = 3.605 below that and is never called out.
	apex := -30 - 6/math.Tan(59*math.Pi/180)
	near("h", [3]float64{0, 0, apex}, 0, 5e-3)
	outside("h", [3]float64{0, 0, apex - 1})

	// THE POINT IS THE FLOOR, not a cone unioned under one. On the axis at the
	// shoulder there is no surface at all — the nearest one is the cone, 3.09
	// away perpendicular to its flank. A flat cap with a cone stuck beneath it
	// reads 0 here instead, claiming a floor 3.6 above the actual tip, and it
	// points the wrong way: one step deeper it reported +1 toward that phantom
	// disc. Both the value and the sign of the gradient are wrong, and nothing
	// on the wall of the hole shows it.
	near("h", [3]float64{0, 0, -30}, -3.0902, 1e-3)
	near("h", [3]float64{0, 0, -31}, -2.2331, 1e-3)

	// The undercut stands 0.5 proud of the Ø20 counterbore between 4 and 5
	// deep. Its wall reads zero there — and the counterbore wall ABOVE it is
	// still outside the void, which is what proves the undercut is bounded at
	// its top and did not eat the wall it exists to relieve.
	near("h", [3]float64{10.5, 0, -4.5}, 0, 1e-3)
	outside("h", [3]float64{10.2, 0, -3})

	// Both chamfers at 45 degrees: the outer one at the face, and the inner one
	// seated at the counterbore floor, where its bore actually begins.
	near("h", [3]float64{10.25, 0, -0.25}, 0, 1e-3)
	near("h", [3]float64{6.25, 0, -5.25}, 0, 1e-3)

	// A hole of ONE feature, written without a list — a collection literal
	// needs two or more elements, so the singleton spelling is the bare
	// feature. Ø6.6 THRU: the wall at 3.3, and nothing bounding it below.
	near("thru", [3]float64{3.3, 0, -5}, 0, 1e-3)
	near("thru", [3]float64{3.3, 0, -500}, 0, 1e-3)

	// NO SURFACE OF THE VOID MAY SIT ON THE FACE IT OPENS ONTO. Across the
	// whole opening, ON the mouth plane, the void has to be strictly inside
	// itself — never zero, which would mean its boundary lies in that plane.
	// A void that stops exactly on the face evaluates correctly at every point
	// you would think to sample and still renders as no hole at all, so this
	// is checked directly rather than inferred.
	//
	// r sweeps the opening: inside the drill, across the counterbore, and out
	// to just within the outer chamfer's rim at 10.5.
	for _, r := range []float64{0, 2, 4, 5.9, 7, 9.9, 10.4} {
		if got := eval("h", [3]float64{r, 0, 0}); !(got < -1e-6) {
			t.Errorf("h at r=%g on the mouth plane = %.6f, want strictly < 0: "+
				"the void's boundary is coincident with the face there", r, got)
		}
	}
	// And just outside the rim the face is untouched.
	outside("h", [3]float64{10.6, 0, 0})
	// And the part still bounds it: above the block is outside the solid,
	// inside the hole the material is gone, and beside it the wall remains.
	outside("solid", [3]float64{0, 0, 1})
	outside("solid", [3]float64{0, 0, -2})
	near("solid", [3]float64{6, 0, -20}, 0, 1e-3)
	near("solid", [3]float64{20, 0, -2}, -2, 1e-3)
}

// TestBoreSections checks the non-round sections and the located features. The
// telling check on a section is at its CORNER: across-flats and across-corners
// differ, and a field that confuses them still reads zero on every flat.
func TestBoreSections(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	const doc = `
hx = drill.hole(drill.hexSocket(10.0, 5.0), 0.0);
sq = drill.hole(drill.squareSocket(10.0, 5.0), 0.0);
dF = drill.hole(drill.dFlat(10.0, 8.0, 5.0), 0.0);
dd = drill.hole(drill.doubleD(10.0, 8.0, 5.0), 0.0);
tp = drill.hole(drill.tap(6.0, 15.0), 0.0);
ctr = drill.hole(drill.centerHole(3.0, 8.0, 5.0), 0.0);
sl = drill.hole(drill.slot(6.0, 20.0, 0.0, 5.0), 0.0);
slCh = drill.hole([drill.slotChamfer(6.0, 20.0, 0.0, 0.4),
                  drill.slot(6.0, 20.0, 0.0, 5.0)], 0.0);
sp = drill.hole([drill.thru(12.0), drill.spline(6.0, 16.0, 3.0, 10.0)], 0.0);
kw = drill.hole([drill.thru(12.0), drill.keyway(4.0, 2.0, 12.0, 10.0)], 0.0);
return hx;
`

	eval := func(expr string, pt [3]float64) float64 {
		t.Helper()
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{expr}, [][3]float64{pt})
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		return float64(groups[0].Samples[0][0].Value)
	}
	near := func(expr string, pt [3]float64, want, tol float64) {
		t.Helper()
		got := eval(expr, pt)
		if math.IsNaN(got) || math.Abs(got-want) > tol {
			t.Errorf("%s at %v = %.4f, want %.3f±%.3f", expr, pt, got, want, tol)
		}
	}
	outside := func(expr string, pt [3]float64) {
		t.Helper()
		got := eval(expr, pt)
		if math.IsNaN(got) || got <= 0 {
			t.Errorf("%s at %v = %.4f, want > 0 (outside the void)", expr, pt, got)
		}
	}

	// Hex across flats 10: the flat is at 5, and 5.5 along the same axis is
	// outside. A field built across CORNERS would read zero out at 5.7735.
	near("hx", [3]float64{5, 0, -2}, 0, 1e-3)
	outside("hx", [3]float64{5.5, 0, -2})
	near("hx", [3]float64{5, 2.8868, -2}, 0, 1e-3)

	// Square across flats 10: flat at 5, and the corner is on the surface too.
	near("sq", [3]float64{5, 0, -2}, 0, 1e-3)
	near("sq", [3]float64{5, 5, -2}, 0, 1e-3)

	// D-hole: Ø10 with one flat, 8 from the flat to the far side, so the flat
	// sits at x = 3. The arc survives on the other three sides.
	near("dF", [3]float64{3, 0, -2}, 0, 1e-3)
	near("dF", [3]float64{0, 5, -2}, 0, 1e-3)
	outside("dF", [3]float64{4, 0, -2})
	// Double-D: two flats 8 apart, so each at 4.
	near("dd", [3]float64{4, 0, -2}, 0, 1e-3)
	near("dd", [3]float64{-4, 0, -2}, 0, 1e-3)
	near("dd", [3]float64{0, 5, -2}, 0, 1e-3)

	// A tapped hole's smooth form is a cylinder at the basic MINOR diameter of
	// its coarse thread — M6x1 at 4.917 — its maximum material condition, which
	// the screw that goes in it interferes with by design.
	near("tp", [3]float64{2.4587, 0, -10}, 0, 1e-3)
	near("tp", [3]float64{0, 0, -15}, 0, 1e-3)

	// An ANSI B94.11 center: a Ø3 pilot under a 60-degree countersink out to
	// Ø8. At the rim the cone is on the surface, and 8 below the mouth the
	// pilot wall is what is left.
	near("ctr", [3]float64{4, 0, 0}, 0, 1e-3)
	near("ctr", [3]float64{1.5, 0, -4.9}, 0, 1e-3)

	// An obround: Ø6 with its end centers 20 apart, so it is on the surface
	// both at the side (y = 3 over the straight) and at the end cap (x = 13).
	near("sl", [3]float64{10, 3, -2}, 0, 1e-3)
	near("sl", [3]float64{13, 0, -2}, 0, 1e-3)
	outside("sl", [3]float64{13.5, 0, -2})

	// A slot's chamfer FOLLOWS the slot. At the mouth its rim is the obround
	// grown by the leg -- 0.4 out at the END (x 13.4) and 0.4 out at the SIDE
	// (y 3.4) alike -- and one leg down it has narrowed to the slot itself.
	//
	// The side is the telling one. A round chamfer big enough to break the
	// ends would reach 13.4 from the centre in every direction, so y 3.4 would
	// be deep inside it and the face would come back with a crescent milled
	// out of it. Here it is exactly on the surface.
	near("slCh", [3]float64{13.4, 0, 0}, 0, 1e-3)
	near("slCh", [3]float64{0, 3.4, 0}, 0, 1e-3)
	near("slCh", [3]float64{0, 3.0, -0.4}, 0, 1e-3)
	outside("slCh", [3]float64{0, 3.6, 0})

	// Six spline grooves out to Ø16 in a Ø12 bore: one groove is centred on
	// +x, and 30 degrees round — halfway to the next — is solid again.
	near("sp", [3]float64{8, 0, -5}, 0, 1e-3)
	outside("sp", [3]float64{5.65, 5.65, -5})

	// A keyway is ONE slot, not a symmetric pair: it cuts to 8 on +x and the
	// bore is untouched at 7 on -x.
	near("kw", [3]float64{8, 0, -5}, 0, 1e-3)
	outside("kw", [3]float64{-7, 0, -5})
}

// TestBoreAxisAt checks placement on an arbitrary normal, which nothing in the
// library could do before: fastener.seatedAt reads only the SIGN of its
// direction's z. The -Z case is the one that matters — a shortest-arc rotation
// from +Z divides by zero there, and this basis does not.
func TestBoreAxisAt(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	const doc = `
h = drill.hole(drill.blind(6.0, 10.0), 0.0);
up = remap(h, drill.axisAt([10.0, 0.0, 0.0], vector.dirZ));
down = remap(h, drill.axisAt(vector.zero, neg(vector.dirZ)));
alongX = remap(h, drill.axisAt(vector.zero, vector.dirX));
obl = remap(h, drill.axisAt(vector.zero, [1.0, 1.0, 1.0]));
return up;
`

	eval := func(expr string, pt [3]float64) float64 {
		t.Helper()
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{expr}, [][3]float64{pt})
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		return float64(groups[0].Samples[0][0].Value)
	}
	near := func(expr string, pt [3]float64, want, tol float64) {
		t.Helper()
		got := eval(expr, pt)
		if math.IsNaN(got) || math.Abs(got-want) > tol {
			t.Errorf("%s at %v = %.4f, want %.3f±%.3f", expr, pt, got, want, tol)
		}
	}
	outside := func(expr string, pt [3]float64) {
		t.Helper()
		got := eval(expr, pt)
		if math.IsNaN(got) || got <= 0 {
			t.Errorf("%s at %v = %.4f, want > 0 (outside the void)", expr, pt, got)
		}
	}

	// A BLIND drill, so the floor makes the direction testable: an unbounded
	// one is symmetric about its axis and would read the same either way.

	// Moved 10 along +x, still axis-aligned: the Ø6 wall is at x = 13, and
	// the floor stops it 10 down.
	near("up", [3]float64{13, 0, -5}, 0, 1e-3)
	near("up", [3]float64{10, 0, -5}, -3, 1e-3)
	outside("up", [3]float64{10, 0, -15})

	// Entering along -Z: the hole runs UP from the origin. The wall is there
	// at +z and the floor is above, not below.
	near("down", [3]float64{3, 0, 5}, 0, 1e-3)
	near("down", [3]float64{0, 0, 5}, -3, 1e-3)
	outside("down", [3]float64{0, 0, 15})

	// Entering along +X: the hole runs in -x, with its floor out that way.
	near("alongX", [3]float64{-5, 3, 0}, 0, 1e-3)
	near("alongX", [3]float64{-5, 0, 0}, -3, 1e-3)
	outside("alongX", [3]float64{-15, 0, 0})

	// An oblique axis: 2 along it from the mouth, 3 off it, is on the wall.
	s3 := 1 / math.Sqrt(3)
	s2 := 1 / math.Sqrt(2)
	pt := [3]float64{-2*s3 + 3*s2, -2*s3 - 3*s2, -2 * s3}
	near("obl", pt, 0, 1e-3)
}

// TestBoreNoCoincidentSurface is the library's standing invariant, stated
// directly rather than inferred: no surface of a hole's void may lie in the
// face it opens onto.
//
// Two coincident surfaces cannot be resolved, and the failure is invisible to
// the kind of checking the rest of this file does — such a hole reads the
// right number at every point you would think to sample, passes its dimension
// tests, and renders as a solid plate. So each feature is built on its own
// with its mouth at z = 0 and swept ACROSS ITS OPENING in that exact plane,
// where it must be strictly inside itself. Zero there means its boundary lies
// in the face.
//
// Every case found while writing this library would have been caught here: the
// bodies clipped at the mouth, the cones clipped at the mouth, and the spline
// and keyway that stopped on it.
func TestBoreNoCoincidentSurface(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	const doc = `
drillThru = drill.hole(drill.thru(12.0), 0.0);
drillBlind = drill.hole(drill.blind(12.0, 30.0), 0.0);
flat = drill.hole(drill.bore(20.0, 5.0), 0.0);
cham = drill.hole(drill.chamfer(20.0, 0.5), 0.0);
csk = drill.hole(drill.countersink(18.0, radians(90.0)), 0.0);
tpr = drill.hole(drill.taper(18.0, radians(60.0)), 0.0);
ctr = drill.hole(drill.centerHole(3.0, 8.0, 5.0), 0.0);
hx = drill.hole(drill.hexSocket(10.0, 5.0), 0.0);
sq = drill.hole(drill.squareSocket(10.0, 5.0), 0.0);
dF = drill.hole(drill.dFlat(10.0, 8.0, 5.0), 0.0);
dd = drill.hole(drill.doubleD(10.0, 8.0, 5.0), 0.0);
tp = drill.hole(drill.tap(6.0, 15.0), 0.0);
thr = drill.hole(drill.tapThreaded(6.0, 1.0, 15.0), 0.0);
sl = drill.hole(drill.slot(6.0, 20.0, 0.0, 5.0), 0.0);
slCh = drill.hole([drill.slotChamfer(6.0, 20.0, 0.0, 0.4),
                  drill.slot(6.0, 20.0, 0.0, 5.0)], 0.0);
kw = drill.hole([drill.thru(12.0), drill.keyway(4.0, 2.0, 12.0, 10.0)], 0.0);
sp = drill.hole([drill.thru(12.0), drill.spline(6.0, 16.0, 3.0, 10.0)], 0.0);
kwThru = drill.hole([drill.thru(12.0), drill.keyway(4.0, 2.0, 12.0)], 0.0);
spThru = drill.hole([drill.thru(12.0), drill.spline(6.0, 16.0, 3.0)], 0.0);
slThru = drill.hole(drill.slot(6.0, 20.0, 0.0), 0.0);
dFThru = drill.hole(drill.dFlat(10.0, 8.0), 0.0);
hxThru = drill.hole(drill.hexSocket(10.0), 0.0);
raised = drill.hole(drill.thru(12.0), 7.5);
return drillThru;
`

	// Each case names a feature and points INSIDE its opening, in the mouth
	// plane. The last point of each is deliberately close to the rim.
	cases := []struct {
		expr string
		in   [][2]float64
	}{
		{"drillThru", [][2]float64{{0, 0}, {3, 0}, {0, 5.9}}},
		{"drillBlind", [][2]float64{{0, 0}, {3, 0}, {0, 5.9}}},
		{"flat", [][2]float64{{0, 0}, {5, 0}, {0, 9.9}}},
		{"cham", [][2]float64{{0, 0}, {5, 0}, {10.4, 0}}},
		{"csk", [][2]float64{{0, 0}, {4, 0}, {8.9, 0}}},
		{"tpr", [][2]float64{{0, 0}, {4, 0}, {8.9, 0}}},
		{"ctr", [][2]float64{{0, 0}, {2, 0}, {3.9, 0}}},
		{"hx", [][2]float64{{0, 0}, {3, 0}, {4.9, 0}, {0, 4.9}}},
		{"sq", [][2]float64{{0, 0}, {3, 3}, {4.9, 4.9}}},
		{"dF", [][2]float64{{0, 0}, {2.9, 0}, {0, 4.9}}},
		{"dd", [][2]float64{{0, 0}, {3.9, 0}, {0, 4.9}}},
		{"tp", [][2]float64{{0, 0}, {1, 0}, {2.35, 0}}},
		{"thr", [][2]float64{{0, 0}, {1, 0}, {2, 0}}},
		{"sl", [][2]float64{{0, 0}, {10, 0}, {12.9, 0}, {0, 2.9}}},
		// The chamfer opens the face, so its own rim is what must not lie in
		// it -- swept right out to 13.3 and 3.3, inside a rim at 13.4 and 3.4.
		{"slCh", [][2]float64{{0, 0}, {10, 0}, {13.3, 0}, {0, 3.3}}},
		// The located features, in their groove rather than in the bore.
		{"kw", [][2]float64{{7, 0}, {7.9, 0}}},
		{"sp", [][2]float64{{7, 0}, {7.9, 0}}},
		// And a mouth that is NOT at z = 0, so the rule is about the mouth
		// plane rather than about the origin.
		{"raised", [][2]float64{{0, 0}, {3, 0}, {0, 5.9}}},
	}

	for _, c := range cases {
		z := 0.0
		if c.expr == "raised" {
			z = 7.5
		}
		pts := make([][3]float64, len(c.in))
		for i, xy := range c.in {
			pts[i] = [3]float64{xy[0], xy[1], z}
		}
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{c.expr}, pts)
		if err != nil {
			t.Fatalf("eval %q: %v", c.expr, err)
		}
		for i, xy := range c.in {
			got := float64(groups[0].Samples[0][i].Value)
			if math.IsNaN(got) || got >= -1e-9 {
				t.Errorf("%s at (%g, %g) in the mouth plane z=%g = %.9f, want strictly < 0: "+
					"a surface of the void lies in the face it opens onto, which renders as no hole",
					c.expr, xy[0], xy[1], z, got)
			}
		}
	}
	// THE SAME RULE AT THE EXIT FACE. A feature meant to go through must not be
	// bounded on the far side either: a groove stopped at exactly the back face
	// leaves an infinitely thin web there, which is the defect upside down and
	// just as invisible to sampling anywhere else.
	//
	// These are swept in the plane a 14-thick part's back face would occupy,
	// and again well past it, since a through feature is bounded by neither.
	thru := []struct {
		expr string
		in   [][2]float64
	}{
		{"drillThru", [][2]float64{{0, 0}, {3, 0}, {0, 5.9}}},
		{"kwThru", [][2]float64{{7, 0}, {7.9, 0}}},
		{"spThru", [][2]float64{{7, 0}, {7.9, 0}}},
		{"slThru", [][2]float64{{0, 0}, {12.9, 0}}},
		{"dFThru", [][2]float64{{0, 0}, {2.9, 0}}},
		{"hxThru", [][2]float64{{0, 0}, {4.9, 0}}},
	}
	for _, c := range thru {
		for _, z := range []float64{-14, -40} {
			pts := make([][3]float64, len(c.in))
			for i, xy := range c.in {
				pts[i] = [3]float64{xy[0], xy[1], z}
			}
			groups, err := e.EvalGsl(ctx, nil, doc, "", []string{c.expr}, pts)
			if err != nil {
				t.Fatalf("eval %q: %v", c.expr, err)
			}
			for i, xy := range c.in {
				got := float64(groups[0].Samples[0][i].Value)
				if math.IsNaN(got) || got >= -1e-9 {
					t.Errorf("%s at (%g, %g) in an exit plane z=%g = %.9f, want strictly < 0: "+
						"it is bounded on the far face and leaves an infinitely thin web there",
						c.expr, xy[0], xy[1], z, got)
				}
			}
		}
	}
}
