//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// A cross recess is one general feature with the standards as numbers passed
// into it: taper the wing flanks and a driver cams out of it, which is
// Phillips; bring them near vertical and it does not, which is JIS; add four
// ribs at 45 degrees and it is Pozidriv. So what these tests pin is the
// PARAMETERS doing what they say, not any particular standard's dimensions.
//
// Wings 1.6 wide (half 0.8) out to diameter 6, 2 deep, flanks at 10 degrees.
// At depth d the wing half width is 0.8 - d*tan(10) = 0.8 - 0.176327*d.
const crossDoc = `
rec = drill.hole([drill.crossRecess(6.0, 1.6, 2.0, radians(10.0)),
                  drill.crossRecess(6.0, 1.6, 2.0, radians(10.0))], 0.0);
poz = drill.hole([drill.crossRecessRibbed(6.0, 1.6, 2.0, radians(10.0), 0.8, 1.0, radians(10.0)),
                  drill.crossRecessRibbed(6.0, 1.6, 2.0, radians(10.0), 0.8, 1.0, radians(10.0))], 0.0);
return p.x;
`

func crossEval(t *testing.T, expr string, pts [][3]float64) []float64 {
	t.Helper()
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	g, err := e.EvalGsl(context.Background(), nil, crossDoc, "", []string{expr}, pts)
	if err != nil {
		t.Fatalf("eval %s: %v", expr, err)
	}
	out := make([]float64, len(pts))
	for i := range pts {
		out[i] = float64(g[0].Samples[0][i].Value)
	}
	return out
}

// TestCrossRecessFlankTapers is the property the whole feature exists for. The
// wing narrows with depth at exactly the flank angle, so the recess reads zero
// on the tapered face at every depth rather than only at the mouth — which is
// what a Phillips bit rides up, and what a JIS one does not.
func TestCrossRecessFlankTapers(t *testing.T) {
	const halfW, tan10 = 0.8, 0.17632698
	depths := []float64{0.0, 0.5, 1.0, 1.5}
	pts := make([][3]float64, 0, len(depths)*2)
	for _, d := range depths {
		// x = 2.0 is well along the +x arm, so the wing flank governs and the
		// outer radius and the floor do not.
		pts = append(pts, [3]float64{2.0, halfW - tan10*d, -d})
		pts = append(pts, [3]float64{2.0, halfW, -d})
	}
	got := crossEval(t, "rec", pts)

	for i, d := range depths {
		onFlank, atMouthWidth := got[2*i], got[2*i+1]
		if math.Abs(onFlank) > 2e-3 {
			t.Errorf("at depth %.1f the flank is at y=%.6f, and the field reads %+.5f there; "+
				"want 0 — the wing must narrow by tan(flank) per unit depth",
				d, halfW-tan10*d, onFlank)
		}
		// The mouth half width is outside the recess at any depth below the
		// face, by exactly the amount the taper took away.
		want := tan10 * d * math.Cos(math.Pi/18)
		if math.Abs(atMouthWidth-want) > 2e-3 {
			t.Errorf("at depth %.1f, y=%.2f reads %+.5f; want %+.5f — the distance back "+
				"out to the narrowed flank", d, halfW, atMouthWidth, want)
		}
	}
}

// TestCrossRecessIsUnitGradient checks the construction rather than a scale
// factor: a tapered wing is one plane per side, so the field is 1-Lipschitz by
// construction. Sampled off every symmetry plane, since an abs() kink or a
// min() switch makes a central difference cancel and says nothing.
func TestCrossRecessIsUnitGradient(t *testing.T) {
	const h = 1e-3
	for _, q := range [][3]float64{
		{2.10, 0.43, -0.37}, {1.73, 0.61, -0.92}, {2.44, 0.25, -1.31},
		{0.37, 2.16, -0.58}, {0.62, 1.84, -1.44},
	} {
		var grad [3]float64
		for a := 0; a < 3; a++ {
			lo, hi := q, q
			lo[a] -= h
			hi[a] += h
			v := crossEval(t, "rec", [][3]float64{lo, hi})
			grad[a] = (v[1] - v[0]) / (2 * h)
		}
		mag := math.Sqrt(grad[0]*grad[0] + grad[1]*grad[1] + grad[2]*grad[2])
		if math.Abs(mag-1.0) > 1e-3 {
			t.Errorf("|grad| at %v = %.6f, want 1", q, mag)
		}
	}
}

// TestCrossRecessBounds pins the outer diameter and the floor, and
// TestCrossRecessRibs the thing that separates Pozidriv from Phillips.
func TestCrossRecessBounds(t *testing.T) {
	pts := [][3]float64{
		{0, 2.9, -0.5}, // 0 along an arm, inside the radius 3
		{0, 3.2, -0.5}, // 1 past it
		{0, 0, -1.9},   // 2 above the floor at depth 2
		{0, 0, -2.5},   // 3 below it
	}
	got := crossEval(t, "rec", pts)
	if !(got[0] < 0) || !(got[1] > 0) {
		t.Errorf("the arms must end at the recess diameter: r=2.9 read %+.4f (want inside), "+
			"r=3.2 read %+.4f (want outside)", got[0], got[1])
	}
	if !(got[2] < 0) || !(got[3] > 0) {
		t.Errorf("the recess must floor at its depth: z=-1.9 read %+.4f (want inside), "+
			"z=-2.5 read %+.4f (want outside)", got[2], got[3])
	}
}

// TestCrossRecessRibs: on the 45-degree diagonal a plain cross is solid and a
// ribbed one is cut. That is the whole difference between a driver that cams
// out and one that does not, and the four tick marks stamped between the arms
// on a Pozidriv head are these ribs.
func TestCrossRecessRibs(t *testing.T) {
	diag := [][3]float64{{1.4, 1.4, -0.5}}
	plain := crossEval(t, "rec", diag)[0]
	ribbed := crossEval(t, "poz", diag)[0]

	if !(plain > 0) {
		t.Errorf("a plain cross recess must be solid on the diagonal; read %+.4f", plain)
	}
	if !(ribbed < 0) {
		t.Errorf("a ribbed cross recess must be cut on the diagonal; read %+.4f", ribbed)
	}
}
