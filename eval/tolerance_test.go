//go:build cgo

package eval

import (
	"fmt"
	"math"
	"testing"
)

// Shaft Modeling, Phase 17 (b): ISO 286 fits. mechanical/tolerance gives the
// limit deviations of shaft zones j5, j6, k5, k6, m5, m6, n6, p6 and hole
// zones H6, H7, J6, J7, K6, K7, M6, M7, and the grades IT5, IT6, IT7, for
// nominal sizes over 6 up to 120 mm, in millimetres:
// tolerance.shaftUpper/shaftLower(nominal, zone) (es, ei),
// tolerance.holeUpper/holeLower(nominal, zone) (ES, EI) and
// tolerance.grade(nominal, it). bearing.seatDia and bearing.housingDia size a
// bearing's journal and housing bore in a zone at its maximum or minimum
// material.
//
// The values (micrometres), from Simply Bearings' ISO 286 tables (ISO 286-2)
// and Miki Pulley's excerpt of JIS B 0401 (JIS B 0401:2016, taken to be ISO
// 286:2010), which agree but for K6 at 6-10 mm, where Simply Bearings prints
// -6: -7 is right, since +2 - (-7) is IT6's 9. IT grades from ISO 286-1 as
// Wikipedia tabulates them; every zone's width is its grade.

var toleranceRanges = [][2]float64{{6, 10}, {10, 18}, {18, 30}, {30, 50}, {50, 80}, {80, 120}}

var toleranceGrades = map[string][6]float64{
	"IT5": {6, 8, 9, 11, 13, 15},
	"IT6": {9, 11, 13, 16, 19, 22},
	"IT7": {15, 18, 21, 25, 30, 35},
}

// Upper and lower deviations per range, micrometres.
var toleranceShafts = map[string][6][2]float64{
	"j5": {{4, -2}, {5, -3}, {5, -4}, {6, -5}, {6, -7}, {6, -9}},
	"j6": {{7, -2}, {8, -3}, {9, -4}, {11, -5}, {12, -7}, {13, -9}},
	"k5": {{7, 1}, {9, 1}, {11, 2}, {13, 2}, {15, 2}, {18, 3}},
	"k6": {{10, 1}, {12, 1}, {15, 2}, {18, 2}, {21, 2}, {25, 3}},
	"m5": {{12, 6}, {15, 7}, {17, 8}, {20, 9}, {24, 11}, {28, 13}},
	"m6": {{15, 6}, {18, 7}, {21, 8}, {25, 9}, {30, 11}, {35, 13}},
	"n6": {{19, 10}, {23, 12}, {28, 15}, {33, 17}, {39, 20}, {45, 23}},
	"p6": {{24, 15}, {29, 18}, {35, 22}, {42, 26}, {51, 32}, {59, 37}},
}

var toleranceHoles = map[string][6][2]float64{
	"H6": {{9, 0}, {11, 0}, {13, 0}, {16, 0}, {19, 0}, {22, 0}},
	"H7": {{15, 0}, {18, 0}, {21, 0}, {25, 0}, {30, 0}, {35, 0}},
	"J6": {{5, -4}, {6, -5}, {8, -5}, {10, -6}, {13, -6}, {16, -6}},
	"J7": {{8, -7}, {10, -8}, {12, -9}, {14, -11}, {18, -12}, {22, -13}},
	"K6": {{2, -7}, {2, -9}, {2, -11}, {3, -13}, {4, -15}, {4, -18}},
	"K7": {{5, -10}, {6, -12}, {6, -15}, {7, -18}, {9, -21}, {10, -25}},
	"M6": {{-3, -12}, {-4, -15}, {-4, -17}, {-4, -20}, {-5, -24}, {-6, -28}},
	"M7": {{0, -15}, {0, -18}, {0, -21}, {0, -25}, {0, -30}, {0, -35}},
}

// Every zone at a size inside each range and at each range's upper end, which
// belongs to it ("over 6 up to 10" holds 10), and every grade.
func TestTolerance_Table(t *testing.T) {
	doc := ""
	var exprs []string
	var want []float64
	n := 0
	add := func(expr string, w float64) {
		name := fmt.Sprintf("v%d", n)
		n++
		doc += name + " = " + expr + ";\n"
		exprs = append(exprs, name)
		want = append(want, w)
	}
	for r, rg := range toleranceRanges {
		for _, size := range []float64{(rg[0] + rg[1]) / 2, rg[1]} {
			for zone, devs := range toleranceShafts {
				add(fmt.Sprintf("tolerance.shaftUpper(%g, tolerance.%s)", size, zone), devs[r][0]/1000)
				add(fmt.Sprintf("tolerance.shaftLower(%g, tolerance.%s)", size, zone), devs[r][1]/1000)
			}
			for zone, devs := range toleranceHoles {
				add(fmt.Sprintf("tolerance.holeUpper(%g, tolerance.%s)", size, zone), devs[r][0]/1000)
				add(fmt.Sprintf("tolerance.holeLower(%g, tolerance.%s)", size, zone), devs[r][1]/1000)
			}
			for grade, its := range toleranceGrades {
				add(fmt.Sprintf("tolerance.grade(%g, tolerance.%s)", size, grade), its[r]/1000)
			}
		}
	}
	doc += "return sphere(1.0);\n"
	got := undercutEval(t, doc, exprs, [][3]float64{{0, 0, 0}})
	for k := range exprs {
		if math.Abs(got[k][0]-want[k]) > 1e-6 {
			t.Errorf("%s = %.4f, want %.4f", doc[0:0]+exprsLine(doc, exprs[k]), got[k][0], want[k])
		}
	}
}

// exprsLine is the declaration of name in doc, for a readable failure.
func exprsLine(doc, name string) string {
	for _, line := range splitLines(doc) {
		if len(line) > len(name)+3 && line[:len(name)+3] == name+" = " {
			return line
		}
	}
	return name
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}

// A 6205's journal in k5 and its housing bore in H7, at each material
// condition: bore 25 (18-30: k5 +11 / +2), outside 52 (50-80: H7 +30 / 0).
func TestBearing_SeatsInAZone(t *testing.T) {
	doc := `
jMax = bearing.seatDia(6205.0, tolerance.k5, tolerance.maxMaterial);
jMin = bearing.seatDia(6205.0, tolerance.k5, tolerance.minMaterial);
hMax = bearing.housingDia(6205.0, tolerance.H7, tolerance.maxMaterial);
hMin = bearing.housingDia(6205.0, tolerance.H7, tolerance.minMaterial);
return sphere(1.0);
`
	got := undercutEval(t, doc, []string{"jMax", "jMin", "hMax", "hMin"}, [][3]float64{{0, 0, 0}})
	for i, w := range []float64{25.011, 25.002, 52.0, 52.030} {
		if math.Abs(got[i][0]-w) > 1e-5 {
			t.Errorf("seat %d: %.4f, want %.4f", i, got[i][0], w)
		}
	}
}
