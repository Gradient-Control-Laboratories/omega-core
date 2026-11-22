//go:build cgo

package eval

import (
	"fmt"
	"math"
	"testing"
)

// Hexagon socket set screws to ISO 4026 (flat point), 4027 (cone point), 4028
// (dog point) and 4029 (cup point), M3 to M12: fastener.setScrew(nominalDia,
// length, point), point fastener.pointFlat, pointCone, pointDog or pointCup.
// Built on a +Z datum, the socket end at z = 0 and the point at z = -length,
// the shank smooth at the major diameter (the thread is a finish, as on every
// fastener here).
//
// The table is the 2003 editions' (DIN EN ISO 4026-4029:2004), from ISO
// 4029:2003 and ISO 4026/4027/4028:1993 themselves (iTeh samples), Westfield
// Fasteners' ISO sheets, fasten.it and amesweb.info, which agree. The
// thresholds of 4026 and 4027 (2003) rest on amesweb alone. ISO issued new
// editions of all four in 2026, not consulted. Point diameters are the
// maxima; the dog's length the maximum of its short or long dog.

// setScrewRow: d, s, t short, t long, then the length at or below which each
// standard takes its short values (4026, 4027, 4028, 4029), then flat dp,
// cone dt, dog dp, short dog z, long dog z, cup dz.
var setScrewRows = [][14]float64{
	{3, 1.5, 1.2, 2, 3, 4, 5, 4, 2.0, 0.75, 2.0, 1.0, 1.75, 1.4},
	{4, 2, 1.5, 2.5, 4, 5, 6, 5, 2.5, 1.0, 2.5, 1.25, 2.25, 2.0},
	{5, 2.5, 2, 3, 5, 6, 6, 5, 3.5, 1.25, 3.5, 1.5, 2.75, 2.5},
	{6, 3, 2, 3.5, 6, 6, 8, 6, 4.0, 1.5, 4.0, 1.75, 3.25, 3.0},
	{8, 4, 3, 5, 8, 8, 10, 8, 5.5, 2.0, 5.5, 2.25, 4.3, 5.0},
	{10, 5, 4, 6, 10, 10, 12, 10, 7.0, 2.5, 7.0, 2.75, 5.3, 6.0},
	{12, 6, 4.8, 8, 12, 12, 16, 12, 8.5, 3.0, 8.5, 3.25, 6.3, 8.0},
}

var setScrewPoints = []string{"pointFlat", "pointCone", "pointDog", "pointCup"}

// The table, read back through the accessors: the socket's width and its
// depth either side of each standard's threshold, and each point's diameter.
func TestSetScrew_Table(t *testing.T) {
	doc := ""
	var exprs []string
	var want []float64
	for i, row := range setScrewRows {
		d := row[0]
		doc += fmt.Sprintf("s%d = fastener.setScrewSocket(%g);\n", i, d)
		exprs, want = append(exprs, fmt.Sprintf("s%d", i)), append(want, row[1])
		for k, point := range setScrewPoints {
			thr := row[4+k]
			doc += fmt.Sprintf("tS%d%d = fastener.setScrewSocketDepth(%g, %g, fastener.%s);\n", i, k, d, thr, point)
			doc += fmt.Sprintf("tL%d%d = fastener.setScrewSocketDepth(%g, %g, fastener.%s);\n", i, k, d, thr+1, point)
			doc += fmt.Sprintf("pd%d%d = fastener.setScrewPointDia(%g, fastener.%s);\n", i, k, d, point)
			exprs = append(exprs, fmt.Sprintf("tS%d%d", i, k), fmt.Sprintf("tL%d%d", i, k), fmt.Sprintf("pd%d%d", i, k))
			want = append(want, row[2], row[3], []float64{row[8], row[9], row[10], row[13]}[k])
		}
		doc += fmt.Sprintf("zS%d = fastener.setScrewDogLength(%g, %g);\nzL%d = fastener.setScrewDogLength(%g, %g);\n", i, d, row[6], i, d, row[6]+1)
		exprs = append(exprs, fmt.Sprintf("zS%d", i), fmt.Sprintf("zL%d", i))
		want = append(want, row[11], row[12])
	}
	doc += "return sphere(1.0);\n"
	got := undercutEval(t, doc, exprs, [][3]float64{{0, 0, 0}})
	for k := range exprs {
		if math.Abs(got[k][0]-want[k]) > 1e-4 {
			t.Errorf("%s = %.4f, want %.4f", exprs[k], got[k][0], want[k])
		}
	}
}

// An M6 x 10 of each point: the end where its standard puts it, the socket
// 3 across flats and 3.5 deep (10 is past every M6 threshold but 4028's, whose
// threshold of 8 is passed too), solid between.
func TestSetScrew_Points(t *testing.T) {
	doc := ""
	for _, point := range setScrewPoints {
		doc += fmt.Sprintf("%s = fastener.setScrew(6.0, 10.0, fastener.%s);\n", point, point)
	}
	doc += "return pointFlat;\n"
	check := func(point string, pts [][3]float64, want []float64, tol float64) {
		t.Helper()
		v := undercutEval(t, doc, []string{point}, pts)[0]
		for i := range pts {
			switch {
			case math.IsInf(want[i], -1):
				if v[i] > -0.2 {
					t.Errorf("%s at %v: %.4f, want solid", point, pts[i], v[i])
				}
			case math.IsInf(want[i], 1):
				if v[i] < 0.05 {
					t.Errorf("%s at %v: %.4f, want open", point, pts[i], v[i])
				}
			default:
				if math.Abs(v[i]-want[i]) > tol {
					t.Errorf("%s at %v: %.4f, want %.4f", point, pts[i], v[i], want[i])
				}
			}
		}
	}
	solid, open := math.Inf(-1), math.Inf(1)
	// Every one: the socket open down its axis to 3.5, solid below it and
	// across the shank's middle.
	for _, point := range setScrewPoints {
		check(point, [][3]float64{{0, 0, -1}, {0, 0, -3.3}, {0, 0, -4.2}, {2.5, 0, -5}}, []float64{open, open, solid, solid}, 0)
	}
	// Flat: an end face 4 across at z = -10.
	check("pointFlat", [][3]float64{{0, 0, -10.5}, {1.5, 0, -10.3}, {2.9, 0, -9.95}}, []float64{0.5, 0.3, open}, 2e-3)
	// Cone, 90 degrees past the threshold: a tip flat 1.5 across at z = -10.
	check("pointCone", [][3]float64{{0, 0, -10.5}, {1.2, 0, -10.0}, {0.4, 0, -9.6}}, []float64{0.5, open, solid}, 2e-3)
	// Dog, long: 4 across, from z = -10 up 3.25 to the thread's end.
	check("pointDog", [][3]float64{{0, 0, -10.5}, {1.8, 0, -8}, {2.5, 0, -8}, {2.2, 0, -5.5}}, []float64{0.5, solid, 0.5, solid}, 2e-3)
	// Cup: a rim 3 across at z = -10 around a 120 degree cone 0.866 deep.
	check("pointCup", [][3]float64{{0, 0, -9.6}, {0, 0, -8.9}, {0, 0, -10.5}}, []float64{open, solid, open}, 0)
}

// The box is the screw's own: the shank's diameter, from the point to the
// socket end.
func TestSetScrew_Box(t *testing.T) {
	doc := "screw = fastener.setScrew(8.0, 12.0, fastener.pointCup);\nreturn screw;\n"
	lo, hi := readBox(t, doc, "screw")
	want := [2][3]float64{{-4, -4, -12}, {4, 4, 0}}
	for k := 0; k < 3; k++ {
		if math.Abs(lo[k]-want[0][k]) > 0.02 || math.Abs(hi[k]-want[1][k]) > 0.02 {
			t.Errorf("box %v .. %v, want %v .. %v", lo, hi, want[0], want[1])
			break
		}
	}
}
