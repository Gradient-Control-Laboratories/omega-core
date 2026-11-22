//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"
)

// Shaft Modeling, Phase 14: knurls. shaft.knurl(dia, pitch, pattern, from, to)
// is a plain diameter as geometry, the blank at its maximum material
// condition, and its ISO 13444 teeth are a displacement height: 90 degree
// teeth `pitch` apart and pitch / 2 deep, a whole number of them round,
// round(pi dia / pitch), so the pattern closes on itself. shaft.relief
// gathers a list's heights for material.withDisplacement; shaft.paint paints
// the teeth where the displacement is not drawn; shaft.body(segments, finish)
// sets both, and the plain envelope painted as the stand-in.
//
// The height is checked where the renderer reads it, on the displaced
// surface: at a root that is a tooth depth inside the blank, where neither
// bound binds.

const (
	knurlDia   = 20.0
	knurlPitch = 1.0
	knurlFrom  = 0.0
	knurlTo    = 20.0
)

// knurlTooth is a 90 degree tooth's profile, 0 on its crest and -0.5 at its
// root, in pitches, u pitches along.
func knurlTooth(u float64) float64 {
	return -math.Abs(u - math.Floor(u) - 0.5)
}

var knurlTeeth = math.Floor(math.Pi*knurlDia/knurlPitch + 0.5)

// knurlAround and knurlAlong are a point's position in teeth round the axis
// and pitches along it.
func knurlAround(q [3]float64) float64 { return math.Atan2(q[1], q[0]) / (2 * math.Pi) * knurlTeeth }
func knurlAlong(q [3]float64) float64  { return q[2] / knurlPitch }

// tan(30 degrees): a helix at 30 degrees to the axis advances this many teeth
// round it for each pitch along it.
const knurlLead = 0.57735026918962576

// knurlHeight is the reference height of each pattern, between the bounds.
func knurlHeight(pattern string, q [3]float64) float64 {
	a, z := knurlAround(q), knurlAlong(q)
	left, right := knurlTooth(a+knurlLead*z), knurlTooth(a-knurlLead*z)
	var t float64
	switch pattern {
	case "knurlStraight":
		t = knurlTooth(a)
	case "knurlAnnular":
		t = knurlTooth(z)
	case "knurlLeft":
		t = left
	case "knurlRight":
		t = right
	case "knurlDiamond":
		t = math.Min(left, right)
	case "knurlDiamondSunk":
		t = math.Max(left, right)
	}
	return t * knurlPitch
}

func knurlDoc(pattern string) string {
	return fmt.Sprintf(`
segs = [shaft.diameter(16.0, -10.0, 0.0), shaft.knurl(20.0, 1.0, shaft.%s, 0.0, 20.0), shaft.diameter(16.0, 20.0, 30.0)];
h = shaft.relief(segs);
plainBody = shaft.body(segs);
blank = shaft.body([shaft.diameter(16.0, -10.0, 0.0), shaft.diameter(20.0, 0.0, 20.0), shaft.diameter(16.0, 20.0, 30.0)]);
return plainBody;
`, pattern)
}

func knurlEval(t *testing.T, doc, expr string, pts [][3]float64) []float64 {
	t.Helper()
	e := New(nil, nil)
	g, err := e.EvalGsl(context.Background(), nil, doc, "", []string{expr}, pts)
	if err != nil {
		t.Fatalf("eval %q: %v", expr, err)
	}
	out := make([]float64, len(pts))
	for i := range pts {
		out[i] = float64(g[0].Samples[0][i].Value)
	}
	return out
}

// knurlRootPoints are points a tooth depth inside the blank, across azimuth
// and the knurl's middle, the seam at azimuth pi among them.
func knurlRootPoints() [][3]float64 {
	r := knurlDia/2 - knurlPitch/2
	var pts [][3]float64
	for _, phi := range []float64{0, 0.013, 0.05, 1.0, 2.2, math.Pi - 0.004, -math.Pi + 0.004, -1.7} {
		for _, z := range []float64{6.0, 9.3, 10.0, 10.5, 13.71} {
			pts = append(pts, [3]float64{r * math.Cos(phi), r * math.Sin(phi), z})
		}
	}
	return pts
}

// Every pattern's height is its teeth, between the bounds.
func TestShaftKnurl_PatternHeights(t *testing.T) {
	pts := knurlRootPoints()
	for _, pattern := range []string{"knurlStraight", "knurlAnnular", "knurlLeft", "knurlRight", "knurlDiamond", "knurlDiamondSunk"} {
		got := knurlEval(t, knurlDoc(pattern), "h", pts)
		worst := 0.0
		for i, q := range pts {
			worst = math.Max(worst, math.Abs(got[i]-knurlHeight(pattern, q)))
		}
		if worst > 2e-4 {
			t.Errorf("%s: height off its teeth by up to %.5f", pattern, worst)
		}
	}
}

// The straight teeth run along the axis and the annular rings round it; a
// helix keeps its height along itself; crest 0, root a tooth depth.
func TestShaftKnurl_PatternInvariants(t *testing.T) {
	r := knurlDia/2 - knurlPitch/2
	at := func(phi, z float64) [3]float64 { return [3]float64{r * math.Cos(phi), r * math.Sin(phi), z} }
	tooth := 2 * math.Pi / knurlTeeth

	straight := knurlEval(t, knurlDoc("knurlStraight"), "h", [][3]float64{at(0, 10), at(tooth/2, 10), at(0.3, 6), at(0.3, 14)})
	if math.Abs(straight[0]+knurlPitch/2) > 2e-4 || math.Abs(straight[1]) > 2e-4 {
		t.Errorf("straight: root %.5f, crest %.5f, want %.2f and 0", straight[0], straight[1], -knurlPitch/2)
	}
	if math.Abs(straight[2]-straight[3]) > 2e-4 {
		t.Errorf("straight teeth change along the axis: %.5f and %.5f", straight[2], straight[3])
	}

	annular := knurlEval(t, knurlDoc("knurlAnnular"), "h", [][3]float64{at(0, 10), at(0, 10.5), at(1.1, 12.3), at(-2.4, 12.3)})
	if math.Abs(annular[0]+knurlPitch/2) > 2e-4 || math.Abs(annular[1]) > 2e-4 {
		t.Errorf("annular: root %.5f, crest %.5f, want %.2f and 0", annular[0], annular[1], -knurlPitch/2)
	}
	if math.Abs(annular[2]-annular[3]) > 2e-4 {
		t.Errorf("annular rings change round the axis: %.5f and %.5f", annular[2], annular[3])
	}

	// A right-hand helix turns counterclockwise, seen from +Z, as it rises.
	dz := 2.0
	turn := knurlLead * dz / knurlPitch * tooth
	for _, c := range []struct {
		pattern string
		sign    float64
	}{{"knurlRight", 1}, {"knurlLeft", -1}} {
		h := knurlEval(t, knurlDoc(c.pattern), "h", [][3]float64{at(0.4, 8), at(0.4+c.sign*turn, 8+dz)})
		if math.Abs(h[0]-h[1]) > 2e-4 {
			t.Errorf("%s: height changes along its own helix: %.5f and %.5f", c.pattern, h[0], h[1])
		}
	}
}

// The relief runs out: nothing on and beyond the end planes, nothing at the
// axis (a section cut shows no pattern there), and nothing on the plain
// diameters either side.
func TestShaftKnurl_ReliefBounds(t *testing.T) {
	r := knurlDia / 2
	pts := [][3]float64{
		{r, 0, knurlFrom}, {r, 0, knurlTo}, {0, r, knurlFrom - 1}, {0, r, knurlTo + 1}, // on and past the planes
		{0, 0, 10}, {1, 0, 10}, // at the axis, and well inside
		{8, 0, -5}, {0, 8, 25}, // the plain diameters
	}
	for _, pattern := range []string{"knurlStraight", "knurlDiamond", "knurlAnnular"} {
		for i, v := range knurlEval(t, knurlDoc(pattern), "h", pts) {
			if math.Abs(v) > 1e-5 {
				t.Errorf("%s: relief %.5f at %v, want none", pattern, v, pts[i])
			}
		}
	}
}

// As geometry the knurl is its blank: the shaft is the same as one with a
// plain diameter in its place.
func TestShaftKnurl_GeometryIsTheBlank(t *testing.T) {
	var pts [][3]float64
	for _, r := range []float64{0, 5, 8.5, 9.8, 10, 10.3, 12} {
		for _, z := range []float64{-12, -5, 0, 0.2, 10, 19.9, 20, 25, 32} {
			pts = append(pts, [3]float64{r * 0.6, r * 0.8, z})
		}
	}
	doc := knurlDoc("knurlDiamond")
	got, want := knurlEval(t, doc, "plainBody", pts), knurlEval(t, doc, "blank", pts)
	for i := range pts {
		if math.Abs(got[i]-want[i]) > 1e-5 {
			t.Errorf("at %v the knurled shaft reads %.5f, its blank %.5f", pts[i], got[i], want[i])
		}
	}
}

// Paint darkens a root where the displacement is not drawn (it is never drawn
// here), and leaves a crest and the plain diameters the base colour.
func TestShaftKnurl_PaintDarkensTheRoots(t *testing.T) {
	r := knurlDia/2 - knurlPitch/2
	tooth := 2 * math.Pi / knurlTeeth
	doc := strings.Replace(knurlDoc("knurlStraight"), "return plainBody;",
		"tint = shaft.paint([0.6, 0.6, 0.6], segs);\nreturn plainBody;", 1)
	pts := [][3]float64{{r, 0, 10}, {r * math.Cos(tooth/2), r * math.Sin(tooth/2), 10}, {8, 0, -5}}
	e := New(nil, nil)
	g, err := e.EvalGsl(context.Background(), nil, doc, "", []string{"tint"}, pts)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	red := func(i int) float64 { return float64(g[0].Samples[0][i].Value) }
	if red(0) > 0.6*0.9 {
		t.Errorf("a root is painted %.4f, want darker than the base 0.6", red(0))
	}
	if math.Abs(red(1)-0.6) > 1e-3 || math.Abs(red(2)-0.6) > 1e-3 {
		t.Errorf("a crest is painted %.4f and a plain diameter %.4f, want the base 0.6", red(1), red(2))
	}
}

// The finished body is the plain one as geometry, and carries the relief as
// its material's displacement and its envelope as its stand-in.
func TestShaftKnurl_FinishedBody(t *testing.T) {
	doc := strings.Replace(knurlDoc("knurlDiamond"), "return plainBody;",
		"steelMat, steelColor = metal.steel();\nknurled = shaft.body(segs, steelMat);\nreturn knurled;", 1)
	pts := [][3]float64{{10, 0, 10}, {9, 0, 10}, {12, 0, 25}, {0, 0, -11}}
	got, want := knurlEval(t, doc, "knurled", pts), knurlEval(t, doc, "plainBody", pts)
	for i := range pts {
		if math.Abs(got[i]-want[i]) > 1e-5 {
			t.Errorf("at %v the finished body reads %.5f, the plain one %.5f", pts[i], got[i], want[i])
		}
	}
	// A lone displaced part's height rides in its material record, which the
	// viewer reads (gsl_sdShapeDisplaced is written only for several parts of
	// different heights); its stand-in is the gsl_simple guard.
	code := fixesShader(t, doc)
	materialField := ""
	if i := strings.Index(code, "fn gsl_materialField"); i >= 0 {
		materialField = code[i:]
		if j := strings.Index(materialField, "\n}"); j >= 0 {
			materialField = materialField[:j]
		}
	}
	if !strings.Contains(materialField, "material_withDisplacement(") {
		t.Errorf("the finished body's material carries no displacement")
	}
	if !strings.Contains(code, "fn gsl_simple") {
		t.Errorf("the finished body's shader has no stand-in")
	}
}

// A shaft with no knurl compiles as before: its relief and paint answers fold
// away. Pinned by the shipped examples' WGSL, which the phase's report
// compares; here, a plain shaft's relief is the constant 0.
func TestShaftKnurl_PlainShaftHasNoRelief(t *testing.T) {
	doc := `
segs = [shaft.diameter(16.0, 0.0, 10.0), shaft.fillet(1.0), shaft.diameter(24.0, 10.0, 20.0)];
h = shaft.relief(segs);
plainBody = shaft.body(segs);
return plainBody;
`
	for i, v := range knurlEval(t, doc, "h", [][3]float64{{8, 0, 5}, {12, 0, 15}, {0, 0, 0}}) {
		if v != 0 {
			t.Errorf("point %d: a plain shaft's relief is %.5f, want 0", i, v)
		}
	}
}
