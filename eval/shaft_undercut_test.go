//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
	"testing"
)

// Shaft Modeling, Phase 15: DIN 76 thread undercuts. shaft.undercut(form) is
// a thread end: the thread keeps its planes, up to the shoulder, and its last
// g2 becomes the groove a thread is cut into so its tool runs out before the
// shoulder: a floor at the undercut diameter d - I, a 30 degree flank (to the
// axis) rising from it to the major diameter on the thread's side, which cuts
// the last teeth down, and a radius r where the floor meets the end plane.
// mechanical/thread holds the table: thread.undercutDia, undercutLength and
// undercutRadius, with forms thread.din76A (g2 = 3.5 P) and thread.iso4755
// (g2 = 3 P), the same floor, flank and radius but ISO's own radii at P 4.5
// and 5.
//
// The sources. DIN 76-1 (form A): the undercut table HEIDENHAIN's MANUALplus
// manual prints for "undercut DIN 76", columns I (= d - dg), K (= g2), R
// (= r) and W (30 degrees). ISO 4755:1983, through IS 1369-2:1993 as
// transcribed by turngcode.com: dg = d - I, g2 = 3 P, r. The two agree on I
// at every pitch both list, and on r but at P 4.5 and 5. g1, the floor's
// width, is printed by ISO and follows from either: g2 less the flank's run,
// sqrt(3) times the depth (d - dg) / 2. DIN 76-1:2025-10 corrected some g2
// and r values in its table 1, which these do not include.

// undercutRow: pitch, I (d - dg), g2 for DIN 76 form A, r for DIN 76, r for
// ISO 4755 (0 where ISO lists no row).
var undercutRows = [][5]float64{
	{0.2, 0.3, 0.7, 0.1, 0},
	{0.25, 0.4, 0.9, 0.12, 0.12},
	{0.3, 0.5, 1.05, 0.16, 0.16},
	{0.35, 0.6, 1.2, 0.16, 0.16},
	{0.4, 0.7, 1.4, 0.2, 0.2},
	{0.45, 0.7, 1.6, 0.2, 0.2},
	{0.5, 0.8, 1.75, 0.2, 0.2},
	{0.6, 1, 2.1, 0.4, 0.4},
	{0.7, 1.1, 2.45, 0.4, 0.4},
	{0.75, 1.2, 2.6, 0.4, 0.4},
	{0.8, 1.3, 2.8, 0.4, 0.4},
	{1, 1.6, 3.5, 0.6, 0.6},
	{1.25, 2, 4.4, 0.6, 0.6},
	{1.5, 2.3, 5.2, 0.8, 0.8},
	{1.75, 2.6, 6.1, 1, 1},
	{2, 3, 7, 1, 1},
	{2.5, 3.6, 8.7, 1.2, 1.2},
	{3, 4.4, 10.5, 1.6, 1.6},
	{3.5, 5, 12, 1.6, 1.6},
	{4, 5.7, 14, 2, 2},
	{4.5, 6.4, 16, 2, 2.5},
	{5, 7, 17.5, 2.5, 2.5},
	{5.5, 7.7, 19, 3.2, 3.2},
	{6, 8.3, 21, 3.2, 3.2},
}

func undercutEval(t *testing.T, doc string, exprs []string, pts [][3]float64) [][]float64 {
	t.Helper()
	e := New(nil, nil)
	g, err := e.EvalGsl(context.Background(), nil, doc, "", exprs, pts)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	out := make([][]float64, len(exprs))
	for k := range exprs {
		out[k] = make([]float64, len(pts))
		for i := range pts {
			out[k][i] = float64(g[k].Samples[0][i].Value)
			if math.IsNaN(out[k][i]) {
				t.Fatalf("%s at %v is NaN", exprs[k], pts[i])
			}
		}
	}
	return out
}

// Every row of the table, against its sources, read through the accessors.
func TestThreadUndercut_Table(t *testing.T) {
	var exprs []string
	doc := ""
	for i, row := range undercutRows {
		doc += fmt.Sprintf("dia%d = thread.undercutDia(20.0, %g);\nlenA%d = thread.undercutLength(%g, thread.din76A);\nradA%d = thread.undercutRadius(%g, thread.din76A);\n",
			i, row[0], i, row[0], i, row[0])
		exprs = append(exprs, fmt.Sprintf("dia%d", i), fmt.Sprintf("lenA%d", i), fmt.Sprintf("radA%d", i))
		if row[4] > 0 {
			doc += fmt.Sprintf("lenI%d = thread.undercutLength(%g, thread.iso4755);\nradI%d = thread.undercutRadius(%g, thread.iso4755);\n", i, row[0], i, row[0])
			exprs = append(exprs, fmt.Sprintf("lenI%d", i), fmt.Sprintf("radI%d", i))
		}
	}
	doc += "return sphere(1.0);\n"
	got := undercutEval(t, doc, exprs, [][3]float64{{0, 0, 0}})
	k := 0
	next := func() float64 { v := got[k][0]; k++; return v }
	near := func(label string, v, want float64) {
		if math.Abs(v-want) > 1e-4 {
			t.Errorf("%s = %.4f, want %.4f", label, v, want)
		}
	}
	for _, row := range undercutRows {
		p := row[0]
		near(fmt.Sprintf("P %g undercut diameter on M20", p), next(), 20-row[1])
		near(fmt.Sprintf("P %g DIN 76 g2", p), next(), row[2])
		near(fmt.Sprintf("P %g DIN 76 r", p), next(), row[3])
		if row[4] > 0 {
			near(fmt.Sprintf("P %g ISO 4755 g2", p), next(), 3*p)
			near(fmt.Sprintf("P %g ISO 4755 r", p), next(), row[4])
		}
	}
}

// g1, the floor's width, follows from g2 and the 30 degree flank, and agrees
// with what ISO 4755 prints (1.6 at P 1, 2.5 at P 1.5) to the table's
// rounding.
func TestThreadUndercut_FloorWidth(t *testing.T) {
	for _, c := range []struct{ p, depth, g2, g1 float64 }{{1, 0.8, 3, 1.6}, {1.5, 1.15, 4.5, 2.5}, {2, 1.5, 6, 3.4}} {
		if got := c.g2 - math.Sqrt(3)*c.depth; math.Abs(got-c.g1) > 0.06 {
			t.Errorf("P %g: g2 less the flank's run is %.3f, ISO prints g1 %.1f", c.p, got, c.g1)
		}
	}
}

// undercutGeometry is the reference M10 x 1.5 thread from z = 0 to 30 with a
// DIN 76 undercut at its top, against a 16 mm shoulder from 30 to 40; the
// same thread plainly ended; and the undercut at the bottom instead, on a
// 16 mm shoulder from -10 to 0.
const undercutDoc = `
cut = shaft.body([shaft.thread(10.0, 1.5, 0.0, 30.0, shaft.planeEnd(), shaft.undercut(thread.din76A)), shaft.diameter(16.0, 30.0, 40.0)]);
iso = shaft.body([shaft.thread(10.0, 1.5, 0.0, 30.0, shaft.planeEnd(), shaft.undercut(thread.iso4755)), shaft.diameter(16.0, 30.0, 40.0)]);
plain = shaft.body([shaft.thread(10.0, 1.5, 0.0, 30.0, shaft.planeEnd(), shaft.planeEnd()), shaft.diameter(16.0, 30.0, 40.0)]);
below = shaft.body([shaft.diameter(16.0, -10.0, 0.0), shaft.thread(10.0, 1.5, 0.0, 30.0, shaft.undercut(thread.din76A), shaft.planeEnd())]);
return cut;
`

// The groove's parts, in the half-section at radius r and height z.
var (
	ucMajor = 5.0
	ucFloor = 5.0 - 2.3/2                    // dg / 2
	ucStart = 30.0 - 5.2                     // where the flank leaves the major diameter
	ucFlank = ucStart + math.Sqrt(3)*(2.3/2) // where it meets the floor
	ucR     = 0.8
)

func ucPoint(r, z, phi float64) [3]float64 {
	return [3]float64{r * math.Cos(phi), r * math.Sin(phi), z}
}

// The floor, the fillet at the shoulder and the flank are where the table
// puts them.
func TestShaftUndercut_GrooveShape(t *testing.T) {
	cx, cz := ucFloor+ucR, 30.0-ucR // the fillet's centre
	pts := [][3]float64{
		ucPoint(ucFloor+0.5, 28.0, 0.3),                    // 0: over the floor
		ucPoint(ucFloor-0.3, 28.0, 1.1),                    // 1: in it
		ucPoint(cx-ucR/math.Sqrt2, cz+ucR/math.Sqrt2, 2.0), // 2: on the fillet, between floor and shoulder
		ucPoint(cx, cz, 2.0),                               // 3: the fillet's centre
		ucPoint(ucFloor+0.2, 30.0-0.05, -0.4),              // 4: inside the fillet, by the shoulder
	}
	v := undercutEval(t, undercutDoc, []string{"cut"}, pts)[0]
	if math.Abs(v[0]-0.5) > 1e-3 {
		t.Errorf("over the floor: %.4f, want 0.5", v[0])
	}
	if v[1] > -0.25 {
		t.Errorf("in the floor: %.4f, want solid", v[1])
	}
	if math.Abs(v[2]) > 2e-3 {
		t.Errorf("on the fillet: %.4f, want 0", v[2])
	}
	if math.Abs(v[3]-ucR) > 2e-3 {
		t.Errorf("at the fillet's centre: %.4f, want its radius %.2f", v[3], ucR)
	}
	if v[4] > 0 {
		t.Errorf("inside the fillet by the shoulder: %.4f, want solid", v[4])
	}

	// The flank: nothing stands outside its cone, and somewhere round it a
	// tooth reaches it, cut down to it.
	mid := 0.5
	r := ucMajor + mid*(ucFloor-ucMajor)
	z := ucStart + mid*(ucFlank-ucStart)
	var ring [][3]float64
	for k := 0; k < 72; k++ {
		ring = append(ring, ucPoint(r, z, 2*math.Pi*float64(k)/72))
	}
	flank := undercutEval(t, undercutDoc, []string{"cut"}, ring)[0]
	nearest := math.Inf(1)
	for k, f := range flank {
		if f < -1e-4 {
			t.Errorf("material outside the flank at azimuth %d: %.5f", k, f)
		}
		nearest = math.Min(nearest, f)
	}
	if nearest > 0.02 {
		t.Errorf("no tooth reaches the flank: nearest %.4f", nearest)
	}
}

// Below the groove the thread is the plainly ended one, and the shaft is
// solid across the stretch where the threaded part and the groove overlap.
func TestShaftUndercut_ThreadAndOverlap(t *testing.T) {
	var away [][3]float64
	for _, z := range []float64{5, 12.3, 20, ucStart - 0.5} {
		for _, r := range []float64{3.9, 4.4, 4.8, 5.2, 7} {
			away = append(away, ucPoint(r, z, 0.7*z))
		}
	}
	v := undercutEval(t, undercutDoc, []string{"cut", "plain"}, away)
	for i := range away {
		if math.Abs(v[0][i]-v[1][i]) > 1e-4 {
			t.Errorf("at %v below the groove: %.5f, the plain thread %.5f", away[i], v[0][i], v[1][i])
		}
	}
	var seam [][3]float64
	for _, z := range []float64{ucStart, ucStart + 0.6, ucStart + 1.3, ucFlank} {
		seam = append(seam, ucPoint(ucFloor-0.4, z, z))
	}
	for i, f := range undercutEval(t, undercutDoc, []string{"cut"}, seam)[0] {
		if f > -0.3 {
			t.Errorf("at %v, where the thread and the groove overlap: %.4f, want solid", seam[i], f)
		}
	}
}

// ISO 4755's groove is shorter: its flank leaves the major diameter 4.5 from
// the shoulder, not 5.2, over the same floor.
func TestShaftUndercut_ISOForm(t *testing.T) {
	isoStart := 30.0 - 4.5
	isoFlank := isoStart + math.Sqrt(3)*(2.3/2)
	pts := [][3]float64{
		ucPoint(ucFloor+0.5, isoFlank+0.9, 0.2), // over the floor
		ucPoint(ucMajor+0.5, ucStart+0.3, 0.2),  // outside, where DIN's flank already falls
	}
	v := undercutEval(t, undercutDoc, []string{"iso", "cut"}, pts)
	if math.Abs(v[0][0]-0.5) > 1e-3 {
		t.Errorf("over ISO's floor: %.4f, want 0.5", v[0][0])
	}
	if !(v[0][1] < v[1][1]) {
		t.Errorf("ISO's groove should start later than DIN's: %.4f vs %.4f", v[0][1], v[1][1])
	}
}

// At the bottom end the groove is the same, turned over: its floor runs up
// from the shoulder at z = 0.
func TestShaftUndercut_BottomEnd(t *testing.T) {
	pts := [][3]float64{ucPoint(ucFloor+0.5, 30.0-28.0, 0.3), ucPoint(ucFloor-0.3, 30.0-28.0, 1.1)}
	v := undercutEval(t, undercutDoc, []string{"below"}, pts)[0]
	if math.Abs(v[0]-0.5) > 1e-3 || v[1] > -0.25 {
		t.Errorf("bottom groove: over the floor %.4f (want 0.5), in it %.4f (want solid)", v[0], v[1])
	}
}

// The box still holds the shaft: the groove only takes material away.
func TestShaftUndercut_Box(t *testing.T) {
	lo, hi := readBox(t, undercutDoc, "cut")
	want := [2][3]float64{{-8, -8, 0}, {8, 8, 40}}
	for k := 0; k < 3; k++ {
		if math.Abs(lo[k]-want[0][k]) > 0.02 || math.Abs(hi[k]-want[1][k]) > 0.02 {
			t.Errorf("box %v .. %v, want %v .. %v", lo, hi, want[0], want[1])
			break
		}
	}
}
