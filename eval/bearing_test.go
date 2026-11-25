//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
)

// Shaft Modeling, Phase 9: bearings and bushings, the mechanical/bearing
// package. Its tables are held here to their sources, row by row, as the
// catalog's are in fastener_standards_test.go; its bodies to closed forms and
// to proofs that a bearing or a bush seats on its journal without
// interference. Fits are nominal, line to line.

// deepGrooveRow is one deep-groove ball bearing: ISO 15 boundary dimensions
// d, D, B and r_s min, and SKF's abutment and fillet dimensions d_a min,
// D_a max and r_a max.
//
// d, D, B and r_s min: SKF, "Rolling bearings", PUB BU/P1 17000/1 EN, October
// 2018, Table 1.1 "Single row deep groove ball bearings" (boundary dimensions
// ISO 15), PDF pages 266-283,
// https://www.skf.com/binaries/pub12/Images/0901d196802809de-Rolling-bearings---17000_1-EN_tcm_12-121486.pdf
// and, agreeing on every row, NTN-SNR's product pages,
// https://eshop.ntn-snr.com/en/product/6205-SNR/6205 and likewise for each
// designation. ISO 15's free sample stops before series 0, 2 and 3.
//
// d_a min, D_a max, r_a max: the same SKF table, "Abutment and fillet
// dimensions". Makers differ here by up to 2 mm (SNR's differ on 28 rows);
// these are SKF's guidance, not a standard.
type deepGrooveRow struct {
	designation            float64
	d, bigD, b, rs         float64
	daMin, bigDaMax, raMax float64
}

var deepGrooveRows = []deepGrooveRow{
	{6000, 10, 26, 8, 0.3, 12, 24, 0.3},
	{6001, 12, 28, 8, 0.3, 14, 26, 0.3},
	{6002, 15, 32, 9, 0.3, 17, 30, 0.3},
	{6003, 17, 35, 10, 0.3, 19, 33, 0.3},
	{6004, 20, 42, 12, 0.6, 23.2, 38.8, 0.6},
	{6005, 25, 47, 12, 0.6, 28.2, 43.8, 0.6},
	{6006, 30, 55, 13, 1, 34.6, 50, 1},
	{6007, 35, 62, 14, 1, 39.6, 57, 1},
	{6008, 40, 68, 15, 1, 44.6, 63, 1},
	{6009, 45, 75, 16, 1, 51, 69, 1},
	{6010, 50, 80, 16, 1, 55, 75, 1},
	{6200, 10, 30, 9, 0.6, 14.2, 25.8, 0.6},
	{6201, 12, 32, 10, 0.6, 16.2, 27.8, 0.6},
	{6202, 15, 35, 11, 0.6, 19.2, 30.8, 0.6},
	{6203, 17, 40, 12, 0.6, 21.2, 35.8, 0.6},
	{6204, 20, 47, 14, 1, 25.6, 41.4, 1},
	{6205, 25, 52, 15, 1, 30.6, 46.4, 1},
	{6206, 30, 62, 16, 1, 35.6, 56, 1},
	{6207, 35, 72, 17, 1.1, 42, 65, 1},
	{6208, 40, 80, 18, 1.1, 47, 73, 1},
	{6209, 45, 85, 19, 1.1, 52, 78, 1},
	{6210, 50, 90, 20, 1.1, 57, 83, 1},
	{6300, 10, 35, 11, 0.6, 14.2, 30.8, 0.6},
	{6301, 12, 37, 12, 1, 17.6, 31.4, 1},
	{6302, 15, 42, 13, 1, 20.6, 36.4, 1},
	{6303, 17, 47, 14, 1, 22.6, 41.4, 1},
	{6304, 20, 52, 15, 1.1, 27, 45, 1},
	{6305, 25, 62, 17, 1.1, 32, 55, 1},
	{6306, 30, 72, 19, 1.1, 37, 65, 1},
	{6307, 35, 80, 21, 1.5, 44, 71, 1.5},
	{6308, 40, 90, 23, 1.5, 49, 81, 1.5},
	{6309, 45, 100, 25, 1.5, 54, 91, 1.5},
	{6310, 50, 110, 27, 2, 61, 99, 2},
}

// ringGrooveRow is DIN 471's groove for a retaining ring on a shaft d1: groove
// diameter d2, width m (H13) and minimum edge margin n. DIN 471:2011 as
// reproduced by Westfield Fasteners, Tables 1-2,
// https://westfieldfasteners.co.uk/Standards/Clip-CircExt.pdf, and Mädler,
// catalogue page 726, https://static.mercateo.com/ff/e2d28082af85416a8a9b3a6f8a237259/pdf/K43_726_EN.pdf,
// which agree on every row both carry (Mädler omits 13, 21, 29, 34, 36, 42
// and 48). DIN 471:2026 changes nothing in this range.
type ringGrooveRow struct{ d1, d2, m, n float64 }

var ringGrooveRows = []ringGrooveRow{
	{10, 9.6, 1.1, 0.6}, {11, 10.5, 1.1, 0.8}, {12, 11.5, 1.1, 0.8}, {13, 12.4, 1.1, 0.9},
	{14, 13.4, 1.1, 0.9}, {15, 14.3, 1.1, 1.1}, {16, 15.2, 1.1, 1.2}, {17, 16.2, 1.1, 1.2},
	{18, 17.0, 1.3, 1.5}, {19, 18.0, 1.3, 1.5}, {20, 19.0, 1.3, 1.5}, {21, 20.0, 1.3, 1.5},
	{22, 21.0, 1.3, 1.5}, {24, 22.9, 1.3, 1.7}, {25, 23.9, 1.3, 1.7}, {26, 24.9, 1.3, 1.7},
	{28, 26.6, 1.6, 2.1}, {29, 27.6, 1.6, 2.1}, {30, 28.6, 1.6, 2.1}, {32, 30.3, 1.6, 2.6},
	{34, 32.3, 1.6, 2.6}, {35, 33.0, 1.6, 3.0}, {36, 34.0, 1.85, 3.0}, {38, 36.0, 1.85, 3.0},
	{40, 37.5, 1.85, 3.8}, {42, 39.5, 1.85, 3.8}, {45, 42.5, 1.85, 3.8}, {48, 45.5, 1.85, 3.8},
	{50, 47.0, 2.15, 4.5},
}

// bushRow is ISO 4379's copper alloy bush for a bore d1: Type C's three
// outside diameters (series 1 to 3, as the table lists them), Type F's two
// series of outside diameter, flange diameter and flange thickness, and the
// 45-degree chamfer's maximum. ISO 4379:2018 Tables 1 and 2,
// https://cdn.standards.iteh.ai/samples/75584/f3f8c8dbadb842f9b451f38b7bdf6a77/ISO-4379-2018.pdf
// matching ISO 4379:1993 in every cell,
// https://cdn.standards.iteh.ai/samples/10265/46b854248ed14ceaaeec8a8937f3ba2e/ISO-4379-1993.pdf
// Bores in parentheses in the standard (24, 27, 33, 36) are for special
// applications only and are carried.
type bushRow struct {
	d1               float64
	c1, c2, c3       float64 // Type C d2, series 1-3
	f1d2, f1d3, f1b2 float64
	f2d2, f2d3, f2b2 float64
	chamfer          float64
}

var bushRows = []bushRow{
	{6, 8, 10, 12, 8, 10, 1, 12, 14, 3, 0.3},
	{8, 10, 12, 14, 10, 12, 1, 14, 18, 3, 0.3},
	{10, 12, 14, 16, 12, 14, 1, 16, 20, 3, 0.3},
	{12, 14, 16, 18, 14, 16, 1, 18, 22, 3, 0.5},
	{14, 16, 18, 20, 16, 18, 1, 20, 25, 3, 0.5},
	{15, 17, 19, 21, 17, 19, 1, 21, 27, 3, 0.5},
	{16, 18, 20, 22, 18, 20, 1, 22, 28, 3, 0.5},
	{18, 20, 22, 24, 20, 22, 1, 24, 30, 3, 0.5},
	{20, 23, 24, 26, 23, 26, 1.5, 26, 32, 3, 0.5},
	{22, 25, 26, 28, 25, 28, 1.5, 28, 34, 3, 0.5},
	{24, 27, 28, 30, 27, 30, 1.5, 30, 36, 3, 0.5},
	{25, 28, 30, 32, 28, 31, 1.5, 32, 38, 4, 0.5},
	{27, 30, 32, 34, 30, 33, 1.5, 34, 40, 4, 0.5},
	{28, 32, 34, 36, 32, 36, 2, 36, 42, 4, 0.5},
	{30, 34, 36, 38, 34, 38, 2, 38, 44, 4, 0.5},
	{32, 36, 38, 40, 36, 40, 2, 40, 46, 4, 0.8},
	{33, 37, 40, 42, 37, 41, 2, 42, 48, 5, 0.8},
	{35, 39, 41, 45, 39, 43, 2, 45, 50, 5, 0.8},
	{36, 40, 42, 46, 40, 44, 2, 46, 52, 5, 0.8},
	{38, 42, 45, 48, 42, 46, 2, 48, 54, 5, 0.8},
	{40, 44, 48, 50, 44, 48, 2, 50, 58, 5, 0.8},
	{42, 46, 50, 52, 46, 50, 2, 52, 60, 5, 0.8},
	{45, 50, 53, 55, 50, 55, 2.5, 55, 63, 5, 0.8},
	{48, 53, 56, 58, 53, 58, 2.5, 58, 66, 5, 0.8},
	{50, 55, 58, 60, 55, 60, 2.5, 60, 68, 5, 0.8},
}

// expectAll evaluates every expression in one document and compares each to
// its value.
func expectAll(t *testing.T, exprs []string, want []float64) {
	t.Helper()
	e := New(source.NewLocalManager("testpkg", "../stdlib"), nil)
	groups, err := e.EvalGsl(context.Background(), nil, "x = 0.0;\nreturn sphere(1.0);\n", "", exprs, [][3]float64{{0, 0, 0}})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	bad := 0
	for i, g := range groups {
		got := float64(g.Samples[0][0].Value)
		if math.IsNaN(got) || math.Abs(got-want[i]) > 1e-4 {
			if bad++; bad <= 10 {
				t.Errorf("%s = %.5f, want %.5f", exprs[i], got, want[i])
			}
		}
	}
	t.Logf("%d values against their sources, %d wrong", len(exprs), bad)
}

func TestBearing_DeepGrooveTableAgainstSources(t *testing.T) {
	var exprs []string
	var want []float64
	add := func(f string, des, v float64) {
		exprs = append(exprs, fmt.Sprintf("bearing.%s(%v)", f, des))
		want = append(want, v)
	}
	for _, r := range deepGrooveRows {
		add("bore", r.designation, r.d)
		add("outerDia", r.designation, r.bigD)
		add("width", r.designation, r.b)
		add("chamfer", r.designation, r.rs)
		add("shoulderDia", r.designation, r.daMin)
		add("housingShoulderDia", r.designation, r.bigDaMax)
		add("filletMax", r.designation, r.raMax)
	}
	// A designation with no row builds nothing.
	add("outerDia", 6011, 0)
	expectAll(t, exprs, want)
}

func TestBearing_RingGrooveTableAgainstSources(t *testing.T) {
	var exprs []string
	var want []float64
	for _, r := range ringGrooveRows {
		exprs = append(exprs, fmt.Sprintf("bearing.ringGrooveDia(%v)", r.d1), fmt.Sprintf("bearing.ringGrooveWidth(%v)", r.d1),
			fmt.Sprintf("bearing.ringGrooveMargin(%v)", r.d1))
		want = append(want, r.d2, r.m, r.n)
	}
	// 23 is not a DIN 471 size.
	exprs = append(exprs, "bearing.ringGrooveDia(23.0)")
	want = append(want, 0)
	expectAll(t, exprs, want)
}

func TestBearing_BushTablesAgainstSources(t *testing.T) {
	var exprs []string
	var want []float64
	add := func(e string, v float64) { exprs = append(exprs, e); want = append(want, v) }
	for _, r := range bushRows {
		add(fmt.Sprintf("bearing.bushingOuterDia(%v, 1.0)", r.d1), r.c1)
		add(fmt.Sprintf("bearing.bushingOuterDia(%v, 2.0)", r.d1), r.c2)
		add(fmt.Sprintf("bearing.bushingOuterDia(%v, 3.0)", r.d1), r.c3)
		add(fmt.Sprintf("bearing.flangedOuterDia(%v, 1.0)", r.d1), r.f1d2)
		add(fmt.Sprintf("bearing.flangeDia(%v, 1.0)", r.d1), r.f1d3)
		add(fmt.Sprintf("bearing.flangeThickness(%v, 1.0)", r.d1), r.f1b2)
		add(fmt.Sprintf("bearing.flangedOuterDia(%v, 2.0)", r.d1), r.f2d2)
		add(fmt.Sprintf("bearing.flangeDia(%v, 2.0)", r.d1), r.f2d3)
		add(fmt.Sprintf("bearing.flangeThickness(%v, 2.0)", r.d1), r.f2b2)
		add(fmt.Sprintf("bearing.bushingChamfer(%v)", r.d1), r.chamfer)
	}
	expectAll(t, exprs, want)
}

// The labelled approximation for the balls, which ISO 15 does not give: a
// diameter of 0.3 (D - d), and as many as fit round the pitch circle at 1.5
// diameters apart. A 6205 gets 9 and a 6204 8, as real ones have.
func TestBearing_BallApproximation(t *testing.T) {
	expectAll(t,
		[]string{"bearing.ballDia(6205.0)", "bearing.ballCount(6205.0)", "bearing.ballCount(6204.0)", "bearing.ballCount(6310.0)"},
		[]float64{8.1, 9, 8, math.Floor(math.Pi * 80 / (1.5 * 0.3 * 60))})
}

// roundedRingDistance is the exact distance to a ring from radius ri to ro
// and z from z0 to z1, its four meridian corners rounded by r.
func roundedRingDistance(q [3]float64, ri, ro, z0, z1, r float64) float64 {
	rho := math.Hypot(q[0], q[1])
	cr, cz := (ri+ro)/2, (z0+z1)/2
	hx, hz := (ro-ri)/2-r, (z1-z0)/2-r
	dx, dz := math.Abs(rho-cr)-hx, math.Abs(q[2]-cz)-hz
	return math.Hypot(math.Max(dx, 0), math.Max(dz, 0)) + math.Min(math.Max(dx, dz), 0) - r
}

func evalBearingDoc(t *testing.T, doc, name string, pts [][3]float64) []float64 {
	t.Helper()
	e := New(source.NewLocalManager("testpkg", "../stdlib"), nil)
	groups, err := e.EvalGsl(context.Background(), nil, doc, "", []string{name}, pts)
	if err != nil {
		t.Fatalf("eval %s: %v", name, err)
	}
	out := make([]float64, len(pts))
	for i := range pts {
		out[i] = float64(groups[0].Samples[0][i].Value)
	}
	return out
}

// The envelope of a 6205 with its face at z = 10 is the rounded ring of its
// boundary dimensions, exactly, inside and out.
func TestBearing_EnvelopeIsExact(t *testing.T) {
	doc := "@scene(80)\nenv = bearing.deepGrooveEnvelope(6205.0, 10.0);\nreturn env;\n"
	var pts [][3]float64
	for x := -30.0; x <= 30; x += 0.9 {
		for y := -30.0; y <= 30; y += 1.3 {
			for z := 6.0; z <= 29; z += 0.7 {
				pts = append(pts, [3]float64{x, y, z})
			}
		}
	}
	got := evalBearingDoc(t, doc, "env", pts)
	bad, worst := 0, 0.0
	for i, q := range pts {
		want := roundedRingDistance(q, 12.5, 26, 10, 25, 1)
		d := math.Abs(got[i] - want)
		worst = math.Max(worst, d)
		if d > 1e-4 {
			if bad++; bad <= 3 {
				t.Errorf("at %v: envelope %v, distance %v", q, got[i], want)
			}
		}
	}
	t.Logf("6205 envelope: %d points, worst miss %.2g mm", len(pts), worst)
}

// certifyBearing proves expr >= 0 over a region, or reports that it is
// disproven. Line-to-line contact over a whole bore is a large surface to
// tile with cells finer than the slack, so the budget is 20 million interval
// evaluations, not the default 200 thousand: each seat proves in 1 to 4 s
// (the owner's call, 2026-11-20).
func certifyBearing(t *testing.T, doc, expr string, region [6]float64) *CertifyResult {
	t.Helper()
	e := New(source.NewLocalManager("testpkg", "../stdlib"), nil)
	e.SetCertifyBudget(20000000)
	r, err := e.Certify(context.Background(), nil, doc, "", expr, region, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// checkWithinEnvelope: on a dense grid, the detailed body never reads below
// its envelope, so nothing of it lies outside. Sampled, not proven: the two
// share the bore, the outside diameter and the faces, and interval arithmetic
// cannot separate coincident surfaces (max(detail, -envelope) straddles zero
// over the whole shared area), so certify runs out of any budget there (the
// owner's call, 2026-11-20).
func checkWithinEnvelope(t *testing.T, doc, detail, env string, pts [][3]float64) {
	t.Helper()
	d := evalBearingDoc(t, doc, detail, pts)
	e := evalBearingDoc(t, doc, env, pts)
	bad := 0
	for i, q := range pts {
		if d[i] < e[i]-1e-5 {
			if bad++; bad <= 3 {
				t.Errorf("%s at %v reads %v, below its envelope's %v", detail, q, d[i], e[i])
			}
		}
	}
	t.Logf("%s within %s: %d points, %d below", detail, env, len(pts), bad)
}

// The detailed 6205 looks like a bearing: nothing of it lies outside its
// envelope; its bore and outside diameter are the envelope's; a ball sits at
// each of its nine places on the pitch circle, and the space between two
// balls on the pitch circle is empty.
func TestBearing_DeepGrooveLooksLikeABearing(t *testing.T) {
	doc := `@scene(80)
detail = bearing.deepGroove(6205.0, 10.0);
env = bearing.deepGrooveEnvelope(6205.0, 10.0);
return detail;
`
	var grid [][3]float64
	for x := -28.0; x <= 28; x += 0.47 {
		for y := -28.0; y <= 28; y += 0.53 {
			for z := 8.0; z <= 27; z += 0.61 {
				grid = append(grid, [3]float64{x, y, z})
			}
		}
	}
	checkWithinEnvelope(t, doc, "detail", "env", grid)
	pitchR, ball, z := (25.0+52.0)/4, 8.1, 17.5
	var pts [][3]float64
	for k := 0; k < 9; k++ {
		a := 2 * math.Pi * float64(k) / 9
		pts = append(pts, [3]float64{pitchR * math.Cos(a), pitchR * math.Sin(a), z})
	}
	for k := 0; k < 9; k++ {
		a := 2 * math.Pi * (float64(k) + 0.5) / 9
		pts = append(pts, [3]float64{pitchR * math.Cos(a), pitchR * math.Sin(a), z})
	}
	// Just inside and just outside the bore and the outside diameter, half way
	// across the width.
	pts = append(pts, [3]float64{12.5 + 0.01, 0, z}, [3]float64{12.5 - 0.01, 0, z},
		[3]float64{26 - 0.01, 0, z}, [3]float64{26 + 0.01, 0, z})
	got := evalBearingDoc(t, doc, "detail", pts)
	for k := 0; k < 9; k++ {
		if math.Abs(got[k]+ball/2) > 1e-3 {
			t.Errorf("ball %d's centre reads %v, want %v", k, got[k], -ball/2)
		}
		if got[9+k] <= 0 {
			t.Errorf("between balls %d and %d the pitch circle reads %v, want empty", k, k+1, got[9+k])
		}
	}
	if got[18] >= 0 || got[19] <= 0 || got[20] >= 0 || got[21] <= 0 {
		t.Errorf("bore in %v out %v, outside diameter in %v out %v", got[18], got[19], got[20], got[21])
	}
}

// A 6205 on a journal at its bore, against a shoulder at d_a min with a
// fillet of r_a max, its face on the shoulder: no interference, proven, for
// the envelope and for the detailed body; its bore meets the journal and its
// face the shoulder, line to line. A fillet a millimetre larger than the
// bearing's chamfer interferes, and is disproven.
func TestBearing_SeatsOnJournalAndShoulder(t *testing.T) {
	seat := func(fillet string) string {
		return fmt.Sprintf(`@scene(80)
journal = shaft.body([shaft.diameter(bearing.bore(6205.0), 0.0, 30.0),
                      shaft.fillet(%s),
                      shaft.diameter(bearing.shoulderDia(6205.0), 30.0, 40.0)]);
env = bearing.deepGrooveEnvelope(6205.0, 30.0 - bearing.width(6205.0));
detail = bearing.deepGroove(6205.0, 30.0 - bearing.width(6205.0));
clearEnv = max(journal, env) + 0.02;
clearDetail = max(journal, detail) + 0.02;
return journal;
`, fillet)
	}
	region := [6]float64{-27, -27, 13, 27, 27, 32}
	ok := seat("bearing.filletMax(6205.0)")
	for _, expr := range []string{"clearEnv", "clearDetail"} {
		if r := certifyBearing(t, ok, expr, region); !r.Proven {
			t.Errorf("%s: the bearing should seat without interference, got %+v", expr, r)
		}
	}
	if r := certifyBearing(t, seat("bearing.chamfer(6205.0) + 1.0"), "clearEnv", region); !r.Disproven {
		t.Errorf("a fillet larger than the chamfer should interfere, got %+v", r)
	}
	// Line to line: on the bore half way along, and on the face between the
	// bore's chamfer and the shoulder's edge, both read zero.
	pts := [][3]float64{{12.5, 0, 22.5}, {14.0, 0, 30.0}}
	journal := evalBearingDoc(t, ok, "journal", pts)
	env := evalBearingDoc(t, ok, "env", pts)
	for i, q := range pts {
		if math.Abs(journal[i]) > 1e-4 || math.Abs(env[i]) > 1e-4 {
			t.Errorf("at %v: journal %v, bearing %v; want both on their surfaces", q, journal[i], env[i])
		}
	}
}

// meridianRingDistance is the exact distance to the solid of revolution of a
// closed polygon in the half plane (distance from the axis, z) that does not
// reach the axis.
func meridianRingDistance(pg poly2) func([3]float64) float64 {
	return func(q [3]float64) float64 { return pg.distance([2]float64{math.Hypot(q[0], q[1]), q[2]}) }
}

// ISO 4379 bushes for a bore of 20 at z = 0, 20 long: Type C series 2 (d2 24)
// and Type F series 1 (d2 23, flange 26 by 1.5, at the bottom). Their
// envelopes are exact; the detailed bushes lie within them and are them half
// way along; and each seats on a journal at its bore without interference.
func TestBearing_Bushings(t *testing.T) {
	doc := `@scene(60)
plainEnv = bearing.bushingEnvelope(20.0, 2.0, 20.0, 0.0);
plain = bearing.bushing(20.0, 2.0, 20.0, 0.0);
flangedEnv = bearing.flangedBushingEnvelope(20.0, 1.0, 20.0, 0.0);
flanged = bearing.flangedBushing(20.0, 1.0, 20.0, 0.0);
journal = shaft.body(shaft.diameter(20.0, -10.0, 30.0));
plainSeat = max(journal, plain) + 0.02;
flangedSeat = max(journal, flanged) + 0.02;
return plain;
`
	plainRef := meridianRingDistance(poly2{{10, 0}, {12, 0}, {12, 20}, {10, 20}})
	flangedRef := meridianRingDistance(poly2{{10, 0}, {13, 0}, {13, 1.5}, {11.5, 1.5}, {11.5, 20}, {10, 20}})
	var pts [][3]float64
	for x := -16.0; x <= 16; x += 0.37 {
		for z := -3.0; z <= 23; z += 0.41 {
			pts = append(pts, [3]float64{x, 0.3 * x, z})
		}
	}
	for _, c := range []struct {
		name string
		ref  func([3]float64) float64
	}{{"plainEnv", plainRef}, {"flangedEnv", flangedRef}} {
		got := evalBearingDoc(t, doc, c.name, pts)
		bad, worst := 0, 0.0
		for i, q := range pts {
			d := math.Abs(got[i] - c.ref(q))
			worst = math.Max(worst, d)
			if d > 1e-4 {
				if bad++; bad <= 3 {
					t.Errorf("%s at %v: %v, distance %v", c.name, q, got[i], c.ref(q))
				}
			}
		}
		t.Logf("%s: %d points, worst miss %.2g mm", c.name, len(pts), worst)
	}
	var grid [][3]float64
	for x := -15.0; x <= 15; x += 0.29 {
		for y := -15.0; y <= 15; y += 0.31 {
			for z := -2.0; z <= 22; z += 0.37 {
				grid = append(grid, [3]float64{x, y, z})
			}
		}
	}
	checkWithinEnvelope(t, doc, "plain", "plainEnv", grid)
	checkWithinEnvelope(t, doc, "flanged", "flangedEnv", grid)
	region := [6]float64{-15, -15, -2, 15, 15, 22}
	for _, expr := range []string{"plainSeat", "flangedSeat"} {
		if r := certifyBearing(t, doc, expr, region); !r.Proven {
			t.Errorf("%s should be proven, got %+v", expr, r)
		}
	}
	mid := [][3]float64{{10, 0, 10}, {12, 0, 10}, {11, 0, 10}}
	plain := evalBearingDoc(t, doc, "plain", mid)
	plainEnv := evalBearingDoc(t, doc, "plainEnv", mid)
	for i := range mid {
		if math.Abs(plain[i]-plainEnv[i]) > 1e-4 {
			t.Errorf("half way along at %v the bush reads %v, its envelope %v", mid[i], plain[i], plainEnv[i])
		}
	}
	// The flanged bush's tube likewise: its wall from 10 to 11.5, half way up.
	// The one-sided check above cannot see a wall that reads empty; this does
	// (the owner's call, 2026-11-20, after the flange's top chamfer cut the
	// tube).
	flangedMid := [][3]float64{{10, 0, 10}, {11.5, 0, 10}, {10.75, 0, 10}}
	flanged := evalBearingDoc(t, doc, "flanged", flangedMid)
	flangedEnv := evalBearingDoc(t, doc, "flangedEnv", flangedMid)
	for i := range flangedMid {
		if math.Abs(flanged[i]-flangedEnv[i]) > 1e-4 {
			t.Errorf("half way up at %v the flanged bush reads %v, its envelope %v", flangedMid[i], flanged[i], flangedEnv[i])
		}
	}
}

// The control: the closed-form rounded ring against a polygon with its corner
// arcs sampled, built independently, over the 6205's meridian.
func TestBearing_RoundedRingControl(t *testing.T) {
	const ri, ro, z0, z1, r = 12.5, 26.0, 10.0, 25.0, 1.0
	corner := func(cx, cz, from float64) [][2]float64 {
		var out [][2]float64
		const n = 2000
		for i := 0; i <= n; i++ {
			a := from + (math.Pi/2)*float64(i)/n
			out = append(out, [2]float64{cx + r*math.Cos(a), cz + r*math.Sin(a)})
		}
		return out
	}
	var pg poly2
	pg = append(pg, corner(ro-r, z0+r, -math.Pi/2)...)
	pg = append(pg, corner(ro-r, z1-r, 0)...)
	pg = append(pg, corner(ri+r, z1-r, math.Pi/2)...)
	pg = append(pg, corner(ri+r, z0+r, math.Pi)...)
	ref := meridianRingDistance(pg)
	worst := 0.0
	for x := 0.0; x <= 30; x += 0.23 {
		for z := 6.0; z <= 29; z += 0.19 {
			q := [3]float64{x, 0, z}
			worst = math.Max(worst, math.Abs(roundedRingDistance(q, ri, ro, z0, z1, r)-ref(q)))
		}
	}
	if worst > 1e-5 {
		t.Errorf("the closed form misses the polygon by %.2g mm", worst)
	}
	t.Logf("rounded ring against its polygon: worst miss %.2g mm", worst)
}

// The bearings example compiles to WGSL (its size logged as Phase 9's budget)
// and evaluates: the shaft solid in its journal, a ball of the left bearing at
// its place (on the placed part, the shaft laid along X), the bush's wall solid,
// and empty beyond the shaft's end.
func TestBearing_BearingsExampleCompilesToWGSL(t *testing.T) {
	src, err := os.ReadFile("../examples/shaft_bearings.gsl")
	if err != nil {
		t.Fatal(err)
	}
	e := New(nil, nil)
	res, err := e.Shader(t.Context(), nil, string(src), "", ShaderOptions{Target: "wgsl", OutDir: t.TempDir()})
	if err != nil {
		t.Fatalf("shader: %v", err)
	}
	t.Logf("examples/shaft_bearings.gsl: %d bytes of WGSL", len(res.Code))
	groups, err := e.EvalGsl(t.Context(), nil, string(src), "", []string{"bearingShaft", "laidLeftBearing", "bush"},
		[][3]float64{{0, 0, 50}, {45, 0, 19.25}, {10.7, 0, 115}, {0, 0, 130}})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	shaftAt, bearingsAt, bushAt := groups[0].Samples[0], groups[1].Samples[0], groups[2].Samples[0]
	if shaftAt[0].Value >= 0 || bearingsAt[1].Value >= 0 || bushAt[2].Value >= 0 || shaftAt[3].Value <= 0 {
		t.Errorf("shaft %v, ball %v, bush wall %v, beyond the end %v: want solid three times, then empty",
			shaftAt[0].Value, bearingsAt[1].Value, bushAt[2].Value, shaftAt[3].Value)
	}
}

// Each bearing and bush body declares its box, exactly its table dimensions:
// across its outside (or flange) diameter, from its face along its width or
// length; and every solid point of it lies inside (the analysis grows a box
// by 0.01, so it is checked to within 0.02).
func TestBearing_BoxesHoldTheirBodies(t *testing.T) {
	cases := []struct {
		expr   string
		r      float64
		z0, z1 float64
	}{
		{"bearing.deepGroove(6205.0, 10.0)", 26, 10, 25},
		{"bearing.deepGrooveEnvelope(6205.0, 10.0)", 26, 10, 25},
		{"bearing.bushing(20.0, 2.0, 20.0, 5.0)", 12, 5, 25},
		{"bearing.bushingEnvelope(20.0, 2.0, 20.0, 5.0)", 12, 5, 25},
		{"bearing.flangedBushing(20.0, 1.0, 20.0, 5.0)", 13, 5, 25},
		{"bearing.flangedBushingEnvelope(20.0, 1.0, 20.0, 5.0)", 13, 5, 25},
	}
	for _, c := range cases {
		doc := fmt.Sprintf("@scene(80)\nsolid = %s;\nreturn solid;\n", c.expr)
		lo, hi := readBox(t, doc, "solid")
		want := [2][3]float64{{-c.r, -c.r, c.z0}, {c.r, c.r, c.z1}}
		for k := 0; k < 3; k++ {
			if math.Abs(lo[k]-want[0][k]) > 0.02 || math.Abs(hi[k]-want[1][k]) > 0.02 {
				t.Errorf("%s: box %v .. %v, want %v .. %v", c.expr, lo, hi, want[0], want[1])
				break
			}
		}
		var pts [][3]float64
		for x := -c.r - 3; x <= c.r+3; x += 0.43 {
			for y := -c.r - 3; y <= c.r+3; y += 0.47 {
				for z := c.z0 - 3; z <= c.z1+3; z += 0.41 {
					pts = append(pts, [3]float64{x, y, z})
				}
			}
		}
		f := evalBearingDoc(t, doc, "solid", pts)
		inside, out := 0, 0
		for i, q := range pts {
			if f[i] >= 0 {
				continue
			}
			inside++
			for k := 0; k < 3; k++ {
				if q[k] < lo[k] || q[k] > hi[k] {
					if out++; out <= 3 {
						t.Errorf("%s: solid point %v outside its box %v .. %v", c.expr, q, lo, hi)
					}
					break
				}
			}
		}
		t.Logf("%s: box %.2f .. %.2f; %d solid points, %d outside", c.expr, lo, hi, inside, out)
	}
}
