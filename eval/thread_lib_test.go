//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// TestThreadStandardSizes checks the catalog against published values rather
// than against itself: an M6's coarse pitch, its basic minor and pitch
// diameters, and its tap drill are all numbers you can look up.
func TestThreadStandardSizes(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	const doc = `
m6 = thread.metricPitch(6.0);
m3 = thread.metricPitch(3.0);
m12 = thread.metricPitch(12.0);
between = thread.metricPitch(6.5);
unc14 = thread.uncPitch(0.25);
unf14 = thread.unfPitch(0.25);
minor6 = thread.minorDia(6.0, 1.0);
pitch6 = thread.pitchDia(6.0, 1.0);
drill6 = thread.tapDrill(6.0, 1.0);
minor3 = thread.minorDia(3.0, 0.5);
return m6;
`

	val := func(expr string) float64 {
		t.Helper()
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{expr}, [][3]float64{{0, 0, 0}})
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		return float64(groups[0].Samples[0][0].Value)
	}
	near := func(expr string, want, tol float64) {
		t.Helper()
		got := val(expr)
		if math.IsNaN(got) || math.Abs(got-want) > tol {
			t.Errorf("%s = %.5f, want %.5f±%g", expr, got, want, tol)
		}
	}

	// ISO 261 coarse pitches.
	near("m6", 1.0, 1e-6)
	near("m3", 0.5, 1e-6)
	near("m12", 1.75, 1e-6)
	// A diameter that is not a catalog size is not a thread. The lookup is a
	// hat indicator, so it reads exactly zero between rows rather than
	// interpolating a fictitious size.
	//
	// 6.5 sits between M6 and M7, both real sizes now that the catalog carries
	// ISO 261 second choice as well as first. It used to be 7.0, which stopped
	// being a probe for "not a size" the moment M7 became one.
	near("between", 0.0, 1e-9)

	// 1/4-20 UNC is 20 threads per inch: 25.4/20 mm. 1/4-28 UNF is 25.4/28.
	near("unc14", 25.4/20, 1e-4)
	near("unf14", 25.4/28, 1e-4)

	// ISO 965 basic diameters for M6 x 1: D1 = 4.917, d2 = 5.350. These come
	// out of the profile geometry rather than a table, which is what makes
	// them worth checking against the published values.
	near("minor6", 4.917, 1e-3)
	near("pitch6", 5.350, 1e-3)
	// And M3 x 0.5: D1 = 2.459.
	near("minor3", 2.459, 1e-3)

	// The shop rule for a tap drill: nominal minus pitch. An M6 takes a 5.0.
	near("drill6", 5.0, 1e-6)
}

// TestThreadProfile checks the ISO 68-1 form itself, and the property that
// matters most about it: a screw and the hole tapped for it MESH.
//
// THE DATUM is phase 0 at z = 0 on +x, where the profile is at its MAJOR
// diameter — a screw's crest, and the root of the hole tapped for it. That is
// the datum fastener's shanks have always used, and TestFastenerLibrary pins
// it independently at z = -4.25 on an M3.
func TestThreadProfile(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	// clash is negative only where screw metal sits somewhere the tapped void
	// is not. The void is the thread form UNION THE TAP DRILL, which is how a
	// tapped hole is actually made: an M6 is drilled 5.0, wider than the basic
	// minor 4.917, so the nut's crest comes back truncated.
	const doc = `
fem = thread.female(6.0, 1.0);
mal = thread.male(6.0, 1.0);
femSin = thread.femaleSin(6.0, 1.0);
band = thread.stripe(6.0, 1.0);
tapped = min(fem, cylinderZ(vector.zero, thread.tapDrill(6.0, 1.0) / 2.0));
clash = max(mal, neg(tapped));
return fem;
`

	at := func(expr string, pt [3]float64) float64 {
		t.Helper()
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{expr}, [][3]float64{pt})
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		return float64(groups[0].Samples[0][0].Value)
	}
	near := func(expr string, pt [3]float64, want, tol float64) {
		t.Helper()
		got := at(expr, pt)
		if math.IsNaN(got) || math.Abs(got-want) > tol {
			t.Errorf("%s at %v = %.5f, want %.5f±%g", expr, pt, got, want, tol)
		}
	}

	// M6 x 1: the form is (5/8)H deep, H = (sqrt3/2)P, so the profile runs
	// between the minor radius 2.4587 and the major radius 3.0. On +x the
	// major-diameter plane is at z = 0 and the minor is half a pitch along.
	const minorR = 3.0 - 0.5412659
	near("fem", [3]float64{3.0, 0, 0}, 0, 1e-3)
	near("fem", [3]float64{minorR, 0, 0.5}, 0, 1e-3)
	// Out at the major radius on a MINOR plane is outside the void, and in at
	// the minor radius on a MAJOR plane is inside it. That is the thread.
	if got := at("fem", [3]float64{3.0, 0, 0.5}); got <= 0 {
		t.Errorf("fem at the major radius on a minor plane = %.4f, want > 0", got)
	}
	if got := at("fem", [3]float64{minorR, 0, 0}); got >= 0 {
		t.Errorf("fem at the minor radius on a major plane = %.4f, want < 0", got)
	}

	// The male form is the same profile on the same datum — one basic profile
	// separates the two threads, so a screw's metal ends exactly where the
	// hole tapped for it begins.
	near("mal", [3]float64{3.0, 0, 0}, 0, 1e-3)
	near("mal", [3]float64{minorR, 0, 0.5}, 0, 1e-3)

	// stripe is the helical phase itself: 0 on the major diameter, 1 on the
	// minor. It reads the same on either radius, since it is a phase and not
	// a distance.
	near("band", [3]float64{minorR, 0, 0}, 0, 1e-3)
	near("band", [3]float64{minorR, 0, 0.5}, 1, 1e-3)

	// The sinusoid is a different form, but it spans the same envelope: it
	// never reaches outside the major diameter nor inside the minor.
	for _, z := range []float64{0, 0.12, 0.25, 0.37, 0.5} {
		if got := at("femSin", [3]float64{3.001, 0, z}); got <= 0 {
			t.Errorf("femSin at z=%g is inside the void beyond the major diameter (%.4f)", z, got)
		}
		if got := at("femSin", [3]float64{minorR - 0.001, 0, z}); got >= 0 {
			t.Errorf("femSin at z=%g is outside the void inside the minor diameter (%.4f)", z, got)
		}
	}

	// THE PROPERTY THAT MATTERS: an M6 screw goes into an M6 tapped hole.
	// Certified over the whole thread, screw metal never reaches anywhere the
	// drilled-and-tapped void is not. Put the two forms half a pitch out of
	// step -- which is what "male" and "female" invite you to do -- and this
	// reads -0.54, the full depth of the form.
	res, err := e.Extremum(ctx, nil, doc, "", "clash",
		[6]float64{-4, -4, -1.5, 4, 4, 1.5}, "min", 0.01, 1)
	if err != nil {
		t.Fatalf("extremum: %v", err)
	}
	if res.Min < -0.02 {
		t.Errorf("an M6 screw interferes with an M6 tapped hole by %.4f at %v",
			-res.Min, res.MinAt)
	}
}

// TestThreadIsADistance is the claim the profile exists to make: the field is
// the distance to the thread, not a radial difference scaled until it behaves.
//
// Two things have to hold. It must never overstep — no pair of points may
// differ in value by more than the distance between them, or a sphere trace
// walks through the surface. And it must be TIGHT, which is the half a radial
// approximation cannot do: the value at a point whose distance is known by
// construction has to be that distance, not a fraction of it.
func TestThreadIsADistance(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	const doc = `
mal = thread.male(6.0, 1.0);
return mal;
`
	sample := func(pts [][3]float64) []float64 {
		t.Helper()
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{"mal"}, pts)
		if err != nil {
			t.Fatalf("eval: %v", err)
		}
		out := make([]float64, len(pts))
		for i := range pts {
			out[i] = float64(groups[0].Samples[0][i].Value)
		}
		return out
	}

	// TIGHTNESS, at points whose distance is exact by construction.
	//
	// Straight out from the middle of the crest flat, the nearest surface is
	// that flat, so the value is the radial offset itself. The old radial
	// field read 0.32 here -- three times the true 0.1, an overstep of 3x.
	//
	// And down in the groove at the minor plane on the major radius, the
	// nearest surface is the FLANK, not the root flat below: 0.3789 by
	// perpendicular distance to a 30-degree line, where a radial reading
	// would say 0.5413. Getting this one right is what proves the field
	// measures to the polyline and not along a ray.
	const lead = 0.99791 // cos of the lead angle at the minor diameter
	fixed := []struct {
		pt   [3]float64
		want float64
	}{
		{[3]float64{3.1, 0, 0}, 0.1 * lead},
		{[3]float64{3.5, 0, 0}, 0.5 * lead},
		{[3]float64{3.0, 0, 0.5}, 0.37889 * lead},
		{[3]float64{3.0, 0, 0}, 0},
	}
	pts := make([][3]float64, len(fixed))
	for i, f := range fixed {
		pts[i] = f.pt
	}
	got := sample(pts)
	for i, f := range fixed {
		if math.Abs(got[i]-f.want) > 2e-3 {
			t.Errorf("mal at %v = %.5f, want the true distance %.5f", f.pt, got[i], f.want)
		}
	}

	// NEVER OVERSTEPS. Walk a lattice that straddles the whole form -- inside
	// the core, across both flats and the flanks, and out past the crest --
	// and compare each sample with its neighbour one step along each axis.
	const step = 0.05
	var probe [][3]float64
	for i := 0; i < 10; i++ {
		for j := 0; j < 7; j++ {
			for k := 0; k < 9; k++ {
				r := 1.6 + float64(i)*0.2
				th := float64(j) * 0.4
				z := -0.4 + float64(k)*0.1
				base := [3]float64{r * math.Cos(th), r * math.Sin(th), z}
				probe = append(probe, base)
				for axis := 0; axis < 3; axis++ {
					shifted := base
					shifted[axis] += step
					probe = append(probe, shifted)
				}
			}
		}
	}
	vals := sample(probe)
	worst := 0.0
	var worstAt [3]float64
	for i := 0; i < len(probe); i += 4 {
		for axis := 1; axis <= 3; axis++ {
			d := math.Abs(vals[i+axis]-vals[i]) / step
			if d > worst {
				worst, worstAt = d, probe[i]
			}
		}
	}
	if worst > 1.0+1e-3 {
		t.Errorf("the thread field oversteps: %.5f of value per unit of distance at %v, want <= 1",
			worst, worstAt)
	}
	// And it is not merely safe by being tiny: a field scaled to a half, as
	// the radial form was, would land near 0.5 here.
	if worst < 0.9 {
		t.Errorf("the thread field is slack: best rate %.5f over %d samples, want near 1",
			worst, len(probe)/4)
	}
}

// TestThreadRelief checks the relief forms: a thread drawn as a displacement
// height on a part built at its maximum material condition.
//
// The claim is that subtracting the height from the smooth part draws EXACTLY
// the thread male() and female() are distances to — not something close to
// it — so the displaced field must agree in sign with the true one everywhere
// the true one is not on its surface. And the height must be zero beyond its
// bounds, with no slab of exact zero where it runs out into an end face.
func TestThreadRelief(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	// shaftDisp is a plain M6 shank at the major diameter with the external
	// relief taken off it. The two parts are M6 plates, one bored at the basic
	// minor diameter (the maximum material condition) and one at the 5.0 tap
	// drill, each with the internal relief. endCap is the shank stopped by an
	// end face at z = -5, where the relief runs out.
	const doc = `
rel = thread.maleRelief(6.0, 1.0, 5.0, -5.0);
shaftDisp = (length(p.xy) - 3.0) - rel;
mal = thread.male(6.0, 1.0);
fem = thread.female(6.0, 1.0);
partMin = (thread.minorDia(6.0, 1.0) / 2.0 - length(p.xy)) - thread.femaleRelief(6.0, 1.0, 5.0, -5.0);
voidMin = neg(fem);
partDrill = (2.5 - length(p.xy)) - thread.femaleRelief(6.0, 1.0, 5.0, 5.0, -5.0);
voidDrill = neg(min(fem, length(p.xy) - 2.5));
femMin = thread.femaleRelief(6.0, 1.0, 5.0, -5.0);
femDrill = thread.femaleRelief(6.0, 1.0, 5.0, 5.0, -5.0);
endCap = max(length(p.xy) - 3.0, -5.0 - p.z) - rel;
return mal;
`
	sample := func(expr string, pts [][3]float64) []float64 {
		t.Helper()
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{expr}, pts)
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		out := make([]float64, len(pts))
		for i := range pts {
			out[i] = float64(groups[0].Samples[0][i].Value)
		}
		return out
	}
	near := func(expr string, pt [3]float64, want, tol float64) {
		t.Helper()
		got := sample(expr, [][3]float64{pt})[0]
		if math.IsNaN(got) || math.Abs(got-want) > tol {
			t.Errorf("%s at %v = %.5f, want %.5f±%g", expr, pt, got, want, tol)
		}
	}

	// ON THE PROFILE the height is the profile's own radial depth. M6 x 1: the
	// crest flat at the major radius, the root flat at the minor radius 2.4587
	// half a pitch along, and the middle of a flank 0.21875 along the axis, which
	// is sqrt3 * (0.21875 - 1/16) = 0.27063 below the crest. The displaced shank
	// is zero at each of them: it is on the thread.
	const minorR = 3.0 - 0.5412659
	near("rel", [3]float64{3.0, 0, 0}, 0, 1e-4)
	near("rel", [3]float64{minorR, 0, 0.5}, -0.5412659, 1e-4)
	near("rel", [3]float64{3.0 - 0.2706329, 0, 0.21875}, -0.2706329, 1e-4)
	near("shaftDisp", [3]float64{minorR, 0, 0.5}, 0, 1e-4)
	near("shaftDisp", [3]float64{3.0 - 0.2706329, 0, 0.21875}, 0, 1e-4)

	// THE SAME THREAD. Walk a lattice through the whole form, well inside the
	// bounds, and every displaced field must agree in sign with the true one
	// wherever the true one is clear of its own surface: the external thread,
	// the internal one on a basic-minor bore, and the internal one on a drilled
	// bore, whose crests come back truncated to the drill.
	var lattice [][3]float64
	for i := 0; i < 12; i++ {
		for j := 0; j < 12; j++ {
			for k := 0; k < 12; k++ {
				r := 2.0 + 1.3*float64(i)/11
				th := 2 * math.Pi * float64(j) / 12
				z := -3.0 + 6.0*float64(k)/11
				lattice = append(lattice, [3]float64{r * math.Cos(th), r * math.Sin(th), z})
			}
		}
	}
	for _, pair := range [][2]string{
		{"shaftDisp", "mal"},
		{"partMin", "voidMin"},
		{"partDrill", "voidDrill"},
	} {
		disp, truth := sample(pair[0], lattice), sample(pair[1], lattice)
		n, bad := 0, 0
		var badAt [3]float64
		for i := range lattice {
			if math.Abs(truth[i]) < 2e-3 {
				continue
			}
			n++
			if (disp[i] < 0) != (truth[i] < 0) {
				bad++
				badAt = lattice[i]
			}
		}
		if bad > 0 {
			t.Errorf("%s disagrees in sign with %s at %d of %d points, e.g. %v",
				pair[0], pair[1], bad, n, badAt)
		}
		if n < len(lattice)/2 {
			t.Errorf("%s: only %d of %d lattice points were clear of the surface", pair[0], n, len(lattice))
		}
	}

	// BOUNDED: nothing above zTop, below zBottom, or outside the band the thread
	// can reach — so a head, a face or a neighbouring part keeps its own surface
	// whichever material it wears.
	for _, expr := range []string{"rel", "femMin", "femDrill"} {
		for _, pt := range [][3]float64{{2.6, 0, 5.2}, {2.6, 0, -5.2}, {4.2, 0, 0.5}, {2.6, 0, 5.0}} {
			near(expr, pt, 0, 1e-6)
		}
	}

	// NO MEMBRANE at the end face. The relief runs out at half the rate the
	// face's own distance grows, so just inside the face the displaced shank is
	// still strictly inside the part. At a rate of one it would read exactly
	// zero through a slab a form depth thick, and render as a skin.
	var slab [][3]float64
	for _, r := range []float64{0.5, 1.2, 2.0, 2.3, 2.45} {
		for _, th := range []float64{0, 1.3, 2.6, 3.9, 5.2} {
			for _, dz := range []float64{0.001, 0.01, 0.05, 0.2, 0.5} {
				slab = append(slab, [3]float64{r * math.Cos(th), r * math.Sin(th), -5.0 + dz})
			}
		}
	}
	for i, v := range sample("endCap", slab) {
		if v >= 0 {
			t.Errorf("endCap at %v = %.6f, want < 0 just inside the end face", slab[i], v)
		}
	}
}
