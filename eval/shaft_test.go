//go:build cgo

package eval

import (
	"fmt"
	"math"
	"os"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// Shaft Modeling, Phase 1: a coaxial round stack in mechanical/shaft, held to
// the same exact reference Phase 0 held the bridge to (shaft_phase0_test.go):
// exact outside, a bound inside, no zero on any joint plane, tight surface
// cells and joint cells that prove solid.

// turnedSeg is one segment of a contiguous coaxial stack: a diameter when
// dia0 == dia1, a taper otherwise, from z = from to z = to.
type turnedSeg struct{ dia0, dia1, from, to float64 }

type turnedStack []turnedSeg

// meridian is the stack's section through its axis, mirrored across it, so the
// axis is no edge of it: the 3D distance is this polygon's at (distance from
// the axis, z).
func (s turnedStack) meridian() poly2 {
	var pg poly2
	for _, g := range s {
		pg = append(pg, [2]float64{g.dia0 / 2, g.from}, [2]float64{g.dia1 / 2, g.to})
	}
	for i := len(s) - 1; i >= 0; i-- {
		pg = append(pg, [2]float64{-s[i].dia1 / 2, s[i].to}, [2]float64{-s[i].dia0 / 2, s[i].from})
	}
	return pg
}

func (s turnedStack) reference() func([3]float64) float64 {
	pg := s.meridian()
	return func(q [3]float64) float64 { return pg.distance([2]float64{math.Hypot(q[0], q[1]), q[2]}) }
}

// gsl is the stack as mechanical/shaft segments.
func (s turnedStack) gsl() string {
	parts := make([]string, len(s))
	for i, g := range s {
		if g.dia0 == g.dia1 {
			parts[i] = fmt.Sprintf("shaft.diameter(%v, %v, %v)", g.dia0, g.from, g.to)
		} else {
			parts[i] = fmt.Sprintf("shaft.taper(%v, %v, %v, %v)", g.dia0, g.dia1, g.from, g.to)
		}
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func (s turnedStack) doc() string {
	return fmt.Sprintf("@scene(140)\nsolid = shaft.body(%s);\nreturn solid;\n", s.gsl())
}

func (s turnedStack) reach() float64 {
	r := 0.0
	for _, g := range s {
		r = math.Max(r, math.Max(g.dia0, g.dia1)/2)
	}
	return r
}

// joints are the planes where one segment meets the next, with the smaller of
// the two radii there.
func (s turnedStack) joints() (zs, radii []float64) {
	for i := 0; i+1 < len(s); i++ {
		zs = append(zs, s[i].to)
		radii = append(radii, math.Min(s[i].dia1, s[i+1].dia0)/2)
	}
	return zs, radii
}

type turnedCase struct {
	name  string
	stack turnedStack
}

func turnedCases() []turnedCase {
	return []turnedCase{
		{"one diameter", turnedStack{{20, 20, 0, 30}}},
		{"step down", turnedStack{{20, 20, 0, 20}, {12, 12, 20, 40}}},
		{"step up", turnedStack{{12, 12, 0, 20}, {20, 20, 20, 40}}},
		{"groove", turnedStack{{20, 20, 0, 15}, {14, 14, 15, 18}, {20, 20, 18, 40}}},
		// A taper alone, narrowing upward: above its small end a point can lie
		// nearest the sloped side, which a prism's formula gets wrong.
		{"taper narrowing", turnedStack{{24, 10, 0, 30}}},
		{"taper widening", turnedStack{{10, 24, 0, 30}}},
		// A cone shoulder: tangent-free but continuous joints on both sides.
		{"cone shoulder", turnedStack{{20, 20, 0, 15}, {20, 12, 15, 25}, {12, 12, 25, 40}}},
		// A taper running into a step down, and one widening into a step up.
		{"taper to step", turnedStack{{20, 14, 0, 20}, {10, 10, 20, 35}}},
		{"taper into shoulder", turnedStack{{12, 16, 0, 20}, {24, 24, 20, 30}}},
		// Thin segments, larger and smaller than both neighbours: a bridge
		// between neighbours alone leaves the whole section under them only
		// half their width deep (the owner's collar, in the viewer).
		{"thin collar", turnedStack{{12, 12, 0, 12}, {16, 16, 12, 13.1}, {12, 12, 13.1, 30}}},
		{"thin groove", turnedStack{{12, 12, 0, 12}, {10, 10, 12, 13.1}, {12, 12, 13.1, 30}}},
		{"collar on a step", turnedStack{{20, 20, 0, 12}, {24, 24, 12, 12.8}, {12, 12, 12.8, 30}}},
	}
}

// tooShallow says whether an interior value falls short of what the body
// promises: at least 1/sqrt2 of the true depth. That is the most a square
// reflex corner costs (the field reads max(a, b) where the truth is the
// distance to the corner, sqrt(a^2 + b^2)); a shallow sheet costs far more,
// reading half a thin segment's width however deep the point lies.
func tooShallow(got, want float64) bool {
	return want < -shaftTolerance && got > want/math.Sqrt2+shaftTolerance
}

// No shallow sheet across any segment: on each segment's mid-plane, the field
// is at least 1/sqrt2 of the true depth. A thin segment bridged only to its
// neighbours fails this across its whole section.
func TestShaftBody_NoShallowSheetAcrossSegments(t *testing.T) {
	for _, c := range turnedCases() {
		t.Run(c.name, func(t *testing.T) {
			var points [][3]float64
			for _, g := range c.stack {
				r := math.Min(g.dia0, g.dia1) / 2
				for x := -r; x <= r; x += r / 8 {
					points = append(points, [3]float64{x, 0, (g.from + g.to) / 2})
				}
			}
			ref := c.stack.reference()
			got := evalStdShaft(t, c.stack.doc(), points)
			shallow, worst := 0, 0.0
			for i, q := range points {
				want := ref(q)
				if want < 0 {
					worst = math.Max(worst, got[i]-want)
				}
				if tooShallow(got[i], want) {
					if shallow++; shallow <= 3 {
						t.Errorf("at %v: field %v, reference %v: shallower than 1/sqrt2 of the truth", q, got[i], want)
					}
				}
			}
			t.Logf("%d mid-plane points, worst shortfall %.3g mm, %d too shallow", len(points), worst, shallow)
		})
	}
}

func evalStdShaft(t *testing.T, doc string, points [][3]float64) []float64 {
	t.Helper()
	e := New(&source.MockManager{}, nil)
	groups, err := e.EvalGsl(t.Context(), nil, doc, "", []string{"solid"}, points)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	out := make([]float64, len(points))
	for i := range points {
		out[i] = float64(groups[0].Samples[0][i].Value)
	}
	return out
}

// shaft.body is exact outside and a bound inside against the reference, and
// never the wrong sign.
func TestShaftBody_MatchesReference(t *testing.T) {
	for _, c := range turnedCases() {
		t.Run(c.name, func(t *testing.T) {
			r := c.stack.reach()
			lo := [3]float64{-r, -r, c.stack[0].from}
			hi := [3]float64{r, r, c.stack[len(c.stack)-1].to}
			points := shaftGrid3(lo, hi, 8, 24)
			ref := c.stack.reference()
			got := evalStdShaft(t, c.stack.doc(), points)
			worstOut, worstIn, failures := 0.0, 0.0, 0
			for i, q := range points {
				want := ref(q)
				var bad bool
				if want > 0 {
					miss := math.Abs(got[i] - want)
					worstOut = math.Max(worstOut, miss)
					bad = miss > shaftTolerance
				} else {
					worstIn = math.Max(worstIn, got[i]-want)
					bad = got[i] < want-shaftTolerance || (want < -shaftTolerance && got[i] >= 0) || tooShallow(got[i], want)
				}
				if bad {
					if failures++; failures <= 3 {
						t.Errorf("at %v: field %v, reference %v", q, got[i], want)
					}
				}
			}
			t.Logf("outside worst miss %.3g mm; inside worst shortfall %.3g mm (a bound); %d failures over %d points",
				worstOut, worstIn, failures, len(points))
		})
	}
}

// No zero on any joint plane, and nothing there shallower than 1/sqrt2 of the
// true depth.
func TestShaftBody_NoSheetOnJoints(t *testing.T) {
	for _, c := range turnedCases() {
		zs, radii := c.stack.joints()
		if len(zs) == 0 {
			continue
		}
		t.Run(c.name, func(t *testing.T) {
			points := jointPoints(zs, radii)
			ref := c.stack.reference()
			got := evalStdShaft(t, c.stack.doc(), points)
			zero, shallow := 0, 0
			for i, q := range points {
				want := ref(q)
				if got[i] > -1e-3 {
					zero++
				}
				if tooShallow(got[i], want) {
					shallow++
				}
			}
			if zero > 0 || shallow > 0 {
				t.Errorf("%d of %d joint points not solid, %d shallower than 1/sqrt2 of the truth", zero, len(points), shallow)
			}
		})
	}
}

// Surface cells bound soundly and tightly, and the cells on the joint planes
// prove solid.
func TestShaftBody_SurfaceCellsAreTight(t *testing.T) {
	diagonal := 2 * cellHalfShaft * math.Sqrt(3)
	for _, c := range turnedCases() {
		t.Run(c.name, func(t *testing.T) {
			ref := c.stack.reference()
			surface := turnedBoundaryCentres(c.stack)
			zs, radii := c.stack.joints()
			var joint [][3]float64
			for j, z := range zs {
				for _, x := range []float64{0, radii[j] / 2, radii[j] - 1} {
					joint = append(joint, [3]float64{x, 0, z})
				}
			}
			all := append(append([][3]float64{}, surface...), joint...)
			boxes := make([][6]float32, len(all))
			for i, q := range all {
				boxes[i] = [6]float32{
					float32(q[0] - cellHalfShaft), float32(q[1] - cellHalfShaft), float32(q[2] - cellHalfShaft),
					float32(q[0] + cellHalfShaft), float32(q[1] + cellHalfShaft), float32(q[2] + cellHalfShaft),
				}
			}
			bounds := intervalsWith(t, &source.MockManager{}, c.stack.doc(), boxes)
			widest, unsound := 0.0, 0
			for i, q := range surface {
				lo, hi := float64(bounds[i].Lower), float64(bounds[i].Upper)
				widest = math.Max(widest, (hi-lo)/diagonal)
				if w := ref(q); w < lo-shaftTolerance || w > hi+shaftTolerance {
					if unsound++; unsound <= 3 {
						t.Errorf("cell about %v bounds [%v, %v], reference %v", q, lo, hi, w)
					}
				}
			}
			solid := 0
			for i := range joint {
				if bounds[len(surface)+i].Upper < 0 {
					solid++
				}
			}
			t.Logf("%d surface cells, widest %.2f diagonals, %d unsound; %d of %d joint cells proved solid",
				len(surface), widest, unsound, solid, len(joint))
			if widest > 4 {
				t.Errorf("widest surface cell %.2f diagonals", widest)
			}
			if solid != len(joint) {
				t.Errorf("%d of %d joint cells proved solid", solid, len(joint))
			}
		})
	}
}

// turnedBoundaryCentres are points on the stack's surface, round its meridian
// at a few angles.
func turnedBoundaryCentres(s turnedStack) [][3]float64 {
	pg := s.meridian()
	var out [][3]float64
	for i, a := range pg {
		b := pg[(i+1)%len(pg)]
		if a[0] < 0 || b[0] < 0 {
			continue
		}
		length := math.Hypot(b[0]-a[0], b[1]-a[1])
		for at := 0.0; at < length; at += 1 {
			f := at / length
			rad, z := a[0]+f*(b[0]-a[0]), a[1]+f*(b[1]-a[1])
			for _, ang := range []float64{0.3, 1.9, 4.0} {
				out = append(out, [3]float64{rad * math.Cos(ang), rad * math.Sin(ang), z})
			}
		}
	}
	return out
}

func intervalsWith(t *testing.T, m *source.MockManager, doc string, boxes [][6]float32) []IntervalSample {
	t.Helper()
	ctx := t.Context()
	if err := startQuery(ctx); err != nil {
		t.Fatal(err)
	}
	defer endQuery()
	e := New(m, nil)
	prog, err := e.prepareQuery(ctx, nil, doc, "", []string{"solid"})
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	results, err := EvalInterval(prog, boxes)
	if err != nil {
		t.Fatalf("interval: %v", err)
	}
	return results[0]
}

// The example compiles to WGSL, and its size is logged: the budget later
// phases are measured against. It also evaluates: its middle is solid and the
// space beyond its ends is not.
func TestShaftBody_ExampleCompilesToWGSL(t *testing.T) {
	src, err := os.ReadFile("../examples/shaft_turned.gsl")
	if err != nil {
		t.Fatal(err)
	}
	e := New(&source.MockManager{}, nil)
	res, err := e.Shader(t.Context(), nil, string(src), "", ShaderOptions{Target: "wgsl", OutDir: t.TempDir()})
	if err != nil {
		t.Fatalf("shader: %v", err)
	}
	t.Logf("examples/shaft_turned.gsl: %d bytes of WGSL", len(res.Code))
	groups, err := e.EvalGsl(t.Context(), nil, string(src), "", []string{"turned"}, [][3]float64{{0, 0, 50}, {0, 0, -5}, {0, 0, 105}})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	got := groups[0].Samples[0]
	if got[0].Value >= 0 || got[1].Value <= 0 || got[2].Value <= 0 {
		t.Errorf("middle %v, beyond the ends %v and %v: want solid, then empty twice", got[0].Value, got[1].Value, got[2].Value)
	}
}
