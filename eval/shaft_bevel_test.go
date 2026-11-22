//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"
)

// Shaft Modeling, Phase 12: gears as shafts. shaft.bevel is a bevel gear as a
// segment: its tooth data (its own and its mate's tooth count, module,
// pressure angle, shaft angle, spiral angle, face width and backlash), its
// pitch apex on the axis, and its two flat faces as its planes. The teeth face
// whichever way the apex lies: below `from` they face -Z, toe at `from` and
// heel at `to`, as gears.bevelGear's do; above `to` they face +Z, the same gear
// turned a half turn about X. shaft.gear gains a backlash, and gears gains
// bevelApexToCrown, the drawing dimension the faces are placed from.
//
// The reference is the gear of gears_bevel_test.go: 20 teeth against 20 at a
// right angle (a 45 degree pitch cone), module 2, pressure angle 0.35, face
// width 12, with the back face 2 beyond the crown and the toe face 2 beyond
// the toe's tip corner, so the toe teeth stand 2 proud of it. shaft.bevel's
// root clearance and edge break are shaft.gear's: a quarter module and a
// twentieth of one.

const (
	sbevTeeth  = 20.0
	sbevModule = 2.0
	sbevAlpha  = 0.35
	sbevDelta  = math.Pi / 4
	sbevFace   = 12.0
	sbevRecess = 2.0
)

var (
	sbevOuter = sbevTeeth * sbevModule / (2 * math.Sin(sbevDelta))
	// Apex to the tip corner, per unit of cone distance, as bevelGear has it.
	sbevTipZ = math.Cos(sbevDelta+sbevModule/sbevOuter) / math.Cos(sbevModule/sbevOuter)
	// The crown's radius: the outside radius.
	sbevCrownR = sbevOuter * math.Sin(sbevDelta+sbevModule/sbevOuter) / math.Cos(sbevModule/sbevOuter)
	sbevCrown  = sbevOuter * sbevTipZ              // apex to crown
	sbevToeTip = (sbevOuter - sbevFace) * sbevTipZ // apex to the toe's tip corner
	sbevHeel   = sbevCrown + sbevRecess            // apex to the back face
	sbevToe    = sbevToeTip + sbevRecess           // apex to the toe face
	sbevUpApex = 60.0                              // the apex of the gear that faces +Z
)

// sbevDoc: `reference` is gears.bevelGear with no hub and no bore; `down` the same
// gear as a lone shaft.bevel with its apex at the origin; `up` the gear facing
// +Z with its apex at z = 60; `dia` and `pinion` a pinion on the top end of a
// shaft and the shaft alone; `journal`, `straddle` and `downAlone` a journal
// past the toe of the gear facing -Z; `env` the simplified model.
func sbevDoc() string {
	bevel := func(apex, from, to float64) string {
		return fmt.Sprintf("shaft.bevel(20.0, 20.0, 2.0, 0.35, 1.5707963267948966, spiral, 12.0, backlash, %.10f, %.10f, %.10f)", apex, from, to)
	}
	joint := sbevUpApex - sbevHeel
	return fmt.Sprintf(`
spiral = 0.0;
backlash = 0.0;
reference = gears.bevelGear(20.0, 2.0, 0.35, 0.7853981633974483, spiral, 0.0, 12.0, 2.0, 2.0, backlash, 0.0, 0.0, 0.0, 0.0, 0.25, 0.1);
down = shaft.body(%[1]s);
up = shaft.body(%[2]s);
dia = shaft.body(shaft.diameter(16.0, 0.0, %.10[3]f));
pinion = shaft.body([shaft.diameter(16.0, 0.0, %.10[3]f), %[2]s]);
journal = shaft.body(shaft.diameter(10.0, 5.0, %.10[4]f));
straddle = shaft.body([shaft.diameter(10.0, 5.0, %.10[4]f), %[1]s]);
env = shaft.envelope(%[1]s);
return down;
`, bevel(0, sbevToe, sbevHeel), bevel(sbevUpApex, sbevUpApex-sbevHeel, sbevUpApex-sbevToe), joint, sbevToe)
}

func sbevEval(t *testing.T, defs, expr string, pts [][3]float64) []float64 {
	t.Helper()
	e := New(nil, nil)
	g, err := e.EvalGsl(context.Background(), nil, sbevDoc(), defs, []string{expr}, pts)
	if err != nil {
		t.Fatalf("eval %q: %v", expr, err)
	}
	out := make([]float64, len(pts))
	for i := range pts {
		out[i] = float64(g[0].Samples[0][i].Value)
	}
	return out
}

// sbevPoints is a spread of points around the gear facing -Z, clear of the
// axis (bevelGear's zero bore still differs there).
func sbevPoints() [][3]float64 {
	var pts [][3]float64
	for _, r := range []float64{0.6, 3, 8, 12, 15.5, 18, 20, 21.2, 22.5, 25} {
		for _, phi := range []float64{0, 0.05, math.Pi / 40, 0.11, math.Pi / 20, 1.3} {
			for z := 6.0; z <= 24.0; z += 1.25 {
				pts = append(pts, [3]float64{r * math.Cos(phi), r * math.Sin(phi), z})
			}
		}
	}
	return pts
}

// The gear facing -Z with its apex at the origin is gears.bevelGear's,
// straight and spiral, with and without backlash, wherever bevelGear's bore
// does not bind: of zero diameter it still caps the inside at minus the
// distance to the axis, which shaft.bevel, having no bore, does not.
func TestShaftBevel_MatchesBevelGear(t *testing.T) {
	pts := sbevPoints()
	for _, defs := range []string{"", "spiral = 0.4;", "backlash = 0.1;", "spiral = -0.3; backlash = 0.1;"} {
		got := sbevEval(t, defs, "down", pts)
		want := sbevEval(t, defs, "reference", pts)
		worst := 0.0
		for i, q := range pts {
			if want[i] <= -math.Hypot(q[0], q[1])+0.3 {
				continue
			}
			worst = math.Max(worst, math.Abs(got[i]-want[i]))
		}
		if worst > 1e-4 {
			t.Errorf("%q: shaft.bevel differs from gears.bevelGear by up to %.6f", defs, worst)
		}
	}
}

// The gear whose apex lies above `to` is the one facing -Z turned a half turn
// about X, which keeps the hand of its spiral.
func TestShaftBevel_FacingUpIsTheHalfTurn(t *testing.T) {
	pts := sbevPoints()
	turned := make([][3]float64, len(pts))
	for i, q := range pts {
		turned[i] = [3]float64{q[0], -q[1], sbevUpApex - q[2]}
	}
	for _, defs := range []string{"", "spiral = 0.4;"} {
		down := sbevEval(t, defs, "down", pts)
		up := sbevEval(t, defs, "up", turned)
		worst := 0.0
		for i := range pts {
			worst = math.Max(worst, math.Abs(down[i]-up[i]))
		}
		if worst > 1e-4 {
			t.Errorf("%q: the gear facing +Z is not the half-turned gear facing -Z (off by up to %.6f)", defs, worst)
		}
	}
}

// A pinion on the top end of a shaft: solid across the joint, no material
// added or lost against the two pieces alone, and the shaft exact outside away
// from the gear.
func TestShaftBevel_PinionOnAShaft(t *testing.T) {
	joint := sbevUpApex - sbevHeel
	var seam [][3]float64
	for _, r := range []float64{0.5, 3, 6, 7.5} {
		seam = append(seam, [3]float64{r, 0, joint}, [3]float64{0, r, joint})
	}
	for i, v := range sbevEval(t, "", "pinion", seam) {
		if v > -0.2 {
			t.Errorf("joint plane at %v reads %.5f: not solid across the joint", seam[i], v)
		}
	}

	var pts [][3]float64
	for _, r := range []float64{2, 6, 7.9, 8.1, 10, 14, 18, 20, 21.2, 23} {
		for _, phi := range []float64{0, 0.07, math.Pi / 20} {
			for z := 1.0; z < 62; z += 1.7 {
				if math.Abs(z-joint) > 1e-3 {
					pts = append(pts, [3]float64{r * math.Cos(phi), r * math.Sin(phi), z})
				}
			}
		}
	}
	whole := sbevEval(t, "", "pinion", pts)
	shaftAlone := sbevEval(t, "", "dia", pts)
	gearAlone := sbevEval(t, "", "up", pts)
	for i := range pts {
		parts := math.Min(shaftAlone[i], gearAlone[i])
		if math.Abs(parts) > 1e-6 && (whole[i] < 0) != (parts < 0) {
			t.Errorf("at %v the pinion reads %.5f but its pieces %.5f", pts[i], whole[i], parts)
		}
	}

	far := sbevEval(t, "", "pinion", [][3]float64{{12, 0, 10}, {0, 11, 15}})
	for i, want := range []float64{4, 3} {
		if math.Abs(far[i]-want) > 1e-4 {
			t.Errorf("shaft outside, far from the gear: %.5f, want %.5f", far[i], want)
		}
	}
}

// A journal past the toe, as a straddle-mounted pinion has: the toe teeth
// that stand proud of the toe face, into the journal's span, are kept.
func TestShaftBevel_StraddleKeepsTheToeTeeth(t *testing.T) {
	tooth := math.Pi / sbevTeeth
	var pts [][3]float64
	for _, r := range []float64{11.5, 12, 12.2} {
		for _, z := range []float64{sbevToeTip + 0.3, 0.5 * (sbevToeTip + sbevToe), sbevToe - 0.3} {
			pts = append(pts, [3]float64{r * math.Cos(tooth), r * math.Sin(tooth), z})
		}
	}
	got := sbevEval(t, "", "straddle", pts)
	want := sbevEval(t, "", "down", pts)
	for i := range pts {
		if math.Abs(got[i]-want[i]) > 1e-4 {
			t.Errorf("toe tooth at %v reads %.5f with the journal, %.5f without", pts[i], got[i], want[i])
		}
	}
}

// The box holds the teeth that stand proud of the toe face: it runs from the
// toe's tip corner to the back face, out to the crown.
func TestShaftBevel_BoxHoldsTheToeTeeth(t *testing.T) {
	cases := []struct {
		name   string
		lo, hi float64
	}{
		{"down", sbevToeTip, sbevHeel},
		{"up", sbevUpApex - sbevHeel, sbevUpApex - sbevToeTip},
	}
	for _, c := range cases {
		lo, hi := readBox(t, sbevDoc(), c.name)
		want := [2][3]float64{{-sbevCrownR, -sbevCrownR, c.lo}, {sbevCrownR, sbevCrownR, c.hi}}
		for k := 0; k < 3; k++ {
			if math.Abs(lo[k]-want[0][k]) > 0.02 || math.Abs(hi[k]-want[1][k]) > 0.02 {
				t.Errorf("%s: box %v .. %v, want %v .. %v", c.name, lo, hi, want[0], want[1])
				break
			}
		}
	}
}

// The simplified model is the blank: it holds the gear, and outside, away from
// its ends, it is the tip cone exactly.
func TestShaftBevel_EnvelopeIsTheBlank(t *testing.T) {
	pts := sbevPoints()
	detail := sbevEval(t, "", "down", pts)
	blank := sbevEval(t, "", "env", pts)
	for i := range pts {
		if detail[i] < 0 && blank[i] >= 0 {
			t.Errorf("at %v the gear is solid (%.5f) but its envelope is not (%.5f)", pts[i], detail[i], blank[i])
		}
	}
	tipAngle := sbevDelta + sbevModule/sbevOuter
	var outside [][3]float64
	var want []float64
	for _, cone := range []float64{20, 23, 26} {
		for _, off := range []float64{0.2, 0.6} {
			th := tipAngle + off/cone
			for _, phi := range []float64{0, 0.1} {
				outside = append(outside, bevelPoint(cone, th, phi))
				want = append(want, cone*math.Sin(th-tipAngle))
			}
		}
	}
	got := sbevEval(t, "", "env", outside)
	for i := range outside {
		if math.Abs(got[i]-want[i]) > 1e-4 {
			t.Errorf("envelope at %v reads %.5f, the tip cone %.5f", outside[i], got[i], want[i])
		}
	}
}

// sbevThickness is the circular thickness, on the circle of radius r in the
// plane z, of the tooth centred on azimuth pi/n beside the space centred on 0,
// found by root-finding expr along that circle.
func sbevThickness(t *testing.T, doc, expr string, n, r, z float64) float64 {
	t.Helper()
	e := New(nil, nil)
	lo, hi := 0.0, math.Pi/n
	for pass := 0; pass < 2; pass++ {
		const k = 400
		pts := make([][3]float64, k+1)
		for i := range pts {
			phi := lo + (hi-lo)*float64(i)/k
			pts[i] = [3]float64{r * math.Cos(phi), r * math.Sin(phi), z}
		}
		g, err := e.EvalGsl(context.Background(), nil, doc, "", []string{expr}, pts)
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		found := false
		for i := 0; i < k; i++ {
			a, b := float64(g[0].Samples[0][i].Value), float64(g[0].Samples[0][i+1].Value)
			if (a < 0) != (b < 0) {
				a0, b0 := lo+(hi-lo)*float64(i)/k, lo+(hi-lo)*float64(i+1)/k
				if pass == 1 {
					flank := a0 + (b0-a0)*a/(a-b)
					return 2 * (math.Pi/n - flank) * r
				}
				lo, hi, found = a0, b0, true
				break
			}
		}
		if !found {
			t.Fatalf("%s: no flank between the space and the tooth at r %.3f", expr, r)
		}
	}
	return math.NaN()
}

// Backlash thins every tooth by that much across its flanks, so its circular
// thickness on the pitch circle falls by backlash / cos(pressure angle): on
// shaft.gear, and on shaft.bevel at the outer end, in proportion to cone
// distance inward of it.
func TestShaft_GearAndBevelBacklash(t *testing.T) {
	const b = 0.1
	gearDoc := `
plain = shaft.body(shaft.gear(20.0, 2.0, 0.35, 0.0, 0.0, 20.0, shaft.planeEnd(), shaft.planeEnd()));
thinned = shaft.body(shaft.gear(20.0, 2.0, 0.35, 0.0, 0.0, 20.0, shaft.planeEnd(), shaft.planeEnd(), 0.1));
return thinned;
`
	full := math.Pi * sbevModule / 2
	thin := full - b/math.Cos(sbevAlpha)
	if got := sbevThickness(t, gearDoc, "plain", 20, 20, 10); math.Abs(got-full) > 2e-3 {
		t.Errorf("shaft.gear tooth thickness %.5f, want %.5f", got, full)
	}
	if got := sbevThickness(t, gearDoc, "thinned", 20, 20, 10); math.Abs(got-thin) > 2e-3 {
		t.Errorf("shaft.gear with backlash %.2f: tooth thickness %.5f, want %.5f", b, got, thin)
	}

	// The bevel, sampled on its pitch cone a unit inside the back cone: the
	// circle there has radius cone*sin(delta) at height cone*cos(delta).
	cone := sbevOuter - 1
	doc := strings.Replace(sbevDoc(), "backlash = 0.0;", "backlash = 0.1;", 1)
	got := sbevThickness(t, doc, "down", sbevTeeth, cone*math.Sin(sbevDelta), cone*math.Cos(sbevDelta))
	want := thin * cone / sbevOuter
	if math.Abs(got-want) > 2e-3 {
		t.Errorf("shaft.bevel with backlash %.2f: tooth thickness %.5f at cone %.3f, want %.5f", b, got, cone, want)
	}
}

// gears.bevelApexToCrown is the distance along the axis from the pitch apex
// to the crown, the outer tooth-tip corner, as bevelGear builds it: a shade
// from the textbook R cos(delta) - m sin(delta), which takes the addendum
// angle as atan(m / R) where the teeth have m / R.
func TestGears_BevelApexToCrown(t *testing.T) {
	cases := []struct{ z, mate, m, sigma float64 }{
		{20, 20, 2, math.Pi / 2},
		{15, 25, 2, math.Pi / 2},
		{16, 24, 5, 1.2},
	}
	for _, c := range cases {
		delta := math.Atan2(math.Sin(c.sigma), c.mate/c.z+math.Cos(c.sigma))
		outer := c.z * c.m / (2 * math.Sin(delta))
		gamma := c.m / outer
		want := outer * math.Cos(delta+gamma) / math.Cos(gamma)
		doc := fmt.Sprintf("crown = gears.bevelApexToCrown(%.1f, %.1f, %.1f, %.10f);\nreturn length(p) - crown;\n", c.z, c.mate, c.m, c.sigma)
		e := New(nil, nil)
		g, err := e.EvalGsl(context.Background(), nil, doc, "", []string{"crown"}, [][3]float64{{0, 0, 0}})
		if err != nil {
			t.Fatalf("eval: %v", err)
		}
		got := float64(g[0].Samples[0][0].Value)
		if math.Abs(got-want) > 1e-4 {
			t.Errorf("%v: apex to crown %.6f, want %.6f", c, got, want)
		}
		textbook := outer*math.Cos(delta) - c.m*math.Sin(delta)
		if math.Abs(got-textbook) > 0.01*c.m {
			t.Errorf("%v: apex to crown %.6f, the textbook %.6f", c, got, textbook)
		}
	}
}
