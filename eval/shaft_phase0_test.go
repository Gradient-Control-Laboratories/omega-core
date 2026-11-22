//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// Shaft Modeling, Phase 0 (conductor/tracks/shaft_modeling_20261118): the
// ground truth for the bridge (spec section 3.3) and the three spikes that
// decide whether the closure pass can carry a shaft as the spec describes.
//
// Everything here runs against shaftSpikeLib, a throwaway package injected
// through the mock source manager. No stdlib code depends on it.

// shaftSpikeLib is just enough of a shaft package to put the bridge through the
// real closure pass.
//
// A segment is a factory returning a closure of two extension lengths: seg(0, 0)
// is the segment capped to its own span, seg(big, 0) runs unbounded below it and
// seg(0, big) unbounded above. The fold then needs no positions: the bridge at a
// joint is max(prev(0, big), cur(big, 0)), the two neighbours' sections over
// both their spans together. Each segment is an exact extrusion of its section,
// so the capped segments are exact and their min is exact outside.
//
// Three combinations of the same segments:
//   - plain:   min of the capped segments. The phantom sheet on every joint.
//   - bridged: plain, and the bridges in the same min. Accumulator [d, prevUp].
//   - gated:   the bridges counted only where plain reads <= 0, so outside the
//     field is plain's (exact) and inside it is the deeper of the two. The gate
//     selects between constants and is combined with max, the form that keeps
//     interval bounds. Accumulator [m, b, prevUp].
const shaftSpikeLib = `
import drill "mechanical/drill";

cyl(dia, from, to) {
  seg(eDown, eUp) {
    dr = length(p.xy) - dia / 2.0;
    dz = max((from - eDown) - p.z, p.z - (to + eUp));
    return sqrt(max(dr, 0.0) * max(dr, 0.0) + max(dz, 0.0) * max(dz, 0.0)) + min(max(dr, dz), 0.0);
  }
  return seg;
}

section(field, from, to) {
  seg(eDown, eUp) {
    dz = max((from - eDown) - p.z, p.z - (to + eUp));
    return sqrt(max(field, 0.0) * max(field, 0.0) + max(dz, 0.0) * max(dz, 0.0)) + min(max(field, dz), 0.0);
  }
  return seg;
}

plain(segments) {
  return reduce((acc, f) { return min(acc, f(0.0, 0.0)); }, segments, 100000.0);
}

bridged(segments) {
  acc = reduce((a, f) {
    return [min(a.x, f(0.0, 0.0), max(a.y, f(100000.0, 0.0))), f(0.0, 100000.0)];
  }, segments, [100000.0, 100000.0]);
  return acc.x;
}

gated(segments) {
  acc = reduce((a, f) {
    return [min(a.x, f(0.0, 0.0)), min(a.y, max(a.z, f(100000.0, 0.0))), f(0.0, 100000.0)];
  }, segments, [100000.0, 100000.0, 100000.0]);
  return min(acc.x, max(acc.y, acc.x > 0.0 ? 100000.0 : -100000.0));
}

withHoles(segments, zLo, zHi, bottom, top) {
  acc = reduce((a, f) {
    return [min(a.x, f(0.0, 0.0)), min(a.y, max(a.z, f(100000.0, 0.0))), f(0.0, 100000.0)];
  }, segments, [100000.0, 100000.0, 100000.0]);
  body = min(acc.x, max(acc.y, acc.x > 0.0 ? 100000.0 : -100000.0));
  topVoid = drill.hole(top, zHi);
  bottomVoid = remap(drill.hole(bottom, 0.0), [p.x, -p.y, zLo - p.z]);
  return max(body, -topVoid, -bottomVoid);
}

withVoids(segments, topVoid, bottomVoid) {
  acc = reduce((a, f) {
    return [min(a.x, f(0.0, 0.0), max(a.y, f(100000.0, 0.0))), f(0.0, 100000.0)];
  }, segments, [100000.0, 100000.0]);
  return max(acc.x, -topVoid, -bottomVoid);
}

withVoidsAt(segments, zLo, zHi, topVoid, bottomVoid) {
  acc = reduce((a, f) {
    return [min(a.x, f(0.0, 0.0), max(a.y, f(100000.0, 0.0))), f(0.0, 100000.0)];
  }, segments, [100000.0, 100000.0]);
  return max(acc.x, -remap(topVoid, [p.x, p.y, p.z - zHi]), -remap(bottomVoid, [p.x, -p.y, zLo - p.z]));
}

// chained is spike (a)'s three-read accumulator: a segment knows its length, not
// its place, and the fold carries the running z. seg(z0, eDown, eUp, wantLen)
// returns the length when wantLen is 1, so the fold can advance.
chainCyl(dia, len) {
  seg(z0, eDown, eUp, wantLen) {
    dr = length(p.xy) - dia / 2.0;
    dz = max((z0 - eDown) - p.z, p.z - (z0 + len + eUp));
    d = sqrt(max(dr, 0.0) * max(dr, 0.0) + max(dz, 0.0) * max(dz, 0.0)) + min(max(dr, dz), 0.0);
    return wantLen > 0.5 ? len : d;
  }
  return seg;
}

chained(segments, z0) {
  acc = reduce((a, f) {
    z = a.x;
    return [z + f(z, 0.0, 0.0, 1.0),
            min(a.y, f(z, 0.0, 0.0, 0.0), max(a.z, f(z, 100000.0, 0.0, 0.0))),
            f(z, 0.0, 100000.0, 0.0)];
  }, segments, [z0, 100000.0, 100000.0]);
  return acc.y;
}
`

// shaftTolerance is how far an exact field may sit from the reference, in mm:
// float32 at the tens of mm these points lie at.
const shaftTolerance = 2e-5

func shaftSpikeManager() *source.MockManager {
	return &source.MockManager{PkgSources: map[string][]string{"geom/shaft": {shaftSpikeLib}}}
}

func evalShaft(t *testing.T, doc, name string, points [][3]float64) []float64 {
	t.Helper()
	e := New(shaftSpikeManager(), nil)
	groups, err := e.EvalGsl(context.Background(), nil, doc, "", []string{name}, points)
	if err != nil {
		t.Fatalf("eval %s: %v", name, err)
	}
	out := make([]float64, len(points))
	for i := range points {
		out[i] = float64(groups[0].Samples[0][i].Value)
	}
	return out
}

// --- The reference ---------------------------------------------------------

// shaftGrid3 is an n^3 grid over the box grown by margin, offset by irrational
// fractions of a step so few points land on a tie.
func shaftGrid3(lo, hi [3]float64, margin float64, n int) [][3]float64 {
	var out [][3]float64
	var step [3]float64
	for k := range step {
		step[k] = (hi[k] - lo[k] + 2*margin) / float64(n-1)
	}
	offsets := [3]float64{0.1 * math.Sqrt2, 0.1 * math.Pi, 0.1 * math.E}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			for l := 0; l < n; l++ {
				idx := [3]int{i, j, l}
				var q [3]float64
				for k := range q {
					q[k] = lo[k] - margin + (float64(idx[k])+offsets[k])*step[k]
				}
				out = append(out, q)
			}
		}
	}
	return out
}

// poly2 is a closed polygon; distance is exact (straight edges), its sign by
// crossings, negative inside.
type poly2 [][2]float64

func (pg poly2) distance(q [2]float64) float64 {
	best := math.Inf(1)
	inside := false
	for i, a := range pg {
		b := pg[(i+1)%len(pg)]
		ex, ey := b[0]-a[0], b[1]-a[1]
		wx, wy := q[0]-a[0], q[1]-a[1]
		h := 0.0
		if l2 := ex*ex + ey*ey; l2 > 0 {
			h = math.Max(0, math.Min(1, (wx*ex+wy*ey)/l2))
		}
		best = math.Min(best, math.Hypot(wx-h*ex, wy-h*ey))
		if (a[1] > q[1]) != (b[1] > q[1]) && q[0] < a[0]+(q[1]-a[1])*ex/ey {
			inside = !inside
		}
	}
	if inside {
		return -best
	}
	return best
}

// roundSeg is one coaxial round segment of a contiguous stack.
type roundSeg struct{ dia, from, to float64 }

// roundStack is a contiguous coaxial stack: segment i+1 starts where i ends.
type roundStack []roundSeg

// meridian is the stack's section through its axis, mirrored across it: the
// 3D distance is this polygon's at (distance from the axis, z), and the axis
// itself is no edge of it.
func (s roundStack) meridian() poly2 {
	var pg poly2
	for _, g := range s {
		pg = append(pg, [2]float64{g.dia / 2, g.from}, [2]float64{g.dia / 2, g.to})
	}
	for i := len(s) - 1; i >= 0; i-- {
		pg = append(pg, [2]float64{-s[i].dia / 2, s[i].to}, [2]float64{-s[i].dia / 2, s[i].from})
	}
	return pg
}

func (s roundStack) reference() func([3]float64) float64 {
	pg := s.meridian()
	return func(q [3]float64) float64 { return pg.distance([2]float64{math.Hypot(q[0], q[1]), q[2]}) }
}

// gsl is the stack as a list of spike segments.
func (s roundStack) gsl() string {
	parts := make([]string, len(s))
	for i, g := range s {
		parts[i] = fmt.Sprintf("spike.cyl(%v, %v, %v)", g.dia, g.from, g.to)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func (s roundStack) reach() float64 {
	r := 0.0
	for _, g := range s {
		r = math.Max(r, g.dia/2)
	}
	return r
}

// joints are the planes where one segment meets the next, with the smaller of
// the two radii there.
func (s roundStack) joints() (zs, radii []float64) {
	for i := 0; i+1 < len(s); i++ {
		zs = append(zs, s[i].to)
		radii = append(radii, math.Min(s[i].dia, s[i+1].dia)/2)
	}
	return zs, radii
}

type stackCase struct {
	name  string
	stack roundStack
}

func roundStackCases() []stackCase {
	return []stackCase{
		{"step down", roundStack{{20, 0, 20}, {12, 20, 40}}},
		{"step up", roundStack{{12, 0, 20}, {20, 20, 40}}},
		// A smaller middle segment: a bridge across it would add material.
		{"groove", roundStack{{20, 0, 15}, {14, 15, 18}, {20, 18, 40}}},
		// Nearly equal diameters: where an ungated bridge reads under the true
		// distance, diagonally off the ends.
		{"near equal", roundStack{{20, 0, 20}, {19.8, 20, 40}}},
	}
}

var shaftVariants = []string{"plain", "bridged", "gated"}

func shaftDoc(variant, segments string) string {
	return fmt.Sprintf("@scene(140)\nimport spike \"geom/shaft\";\nsolid = spike.%s(%s);\nreturn solid;\n", variant, segments)
}

// jointPoints are points on each joint plane strictly inside both neighbours,
// where a min of capped segments reads 0.
func jointPoints(zs, radii []float64) [][3]float64 {
	var out [][3]float64
	for j, z := range zs {
		r := radii[j] - 0.05
		for x := -r; x <= r; x += r / 6 {
			for y := -r; y <= r; y += r / 6 {
				if math.Hypot(x, y) < r {
					out = append(out, [3]float64{x, y, z})
				}
			}
		}
	}
	return out
}

// sheetFailures counts joint points where the field is not solid, and where an
// inset of 1 mm opens a slit although the true distance is deeper than that.
func sheetFailures(got, want []float64) (zero, slit int) {
	for i := range got {
		if got[i] > -1e-3 {
			zero++
		}
		if want[i] < -1.05 && got[i]+1 >= 0 {
			slit++
		}
	}
	return zero, slit
}

// --- The phantom, and the bridge's answer to it ------------------------------

// The phantom pinned: a min of capped segments is not solid on its joint planes,
// and an inset opens it. The check has teeth (plain fails it everywhere); the
// bridge, ungated or gated, passes it.
func TestShaftPhase0_NoSheetOnJoints(t *testing.T) {
	for _, c := range roundStackCases() {
		t.Run(c.name, func(t *testing.T) {
			points := jointPoints(c.stack.joints())
			ref := c.stack.reference()
			want := make([]float64, len(points))
			for i, q := range points {
				want[i] = ref(q)
			}
			for _, v := range shaftVariants {
				zero, slit := sheetFailures(evalShaft(t, shaftDoc(v, c.stack.gsl()), "solid", points), want)
				t.Logf("%s: %d of %d joint points not solid, %d slit under a 1 mm inset", v, zero, len(points), slit)
				if v == "plain" {
					if zero != len(points) {
						t.Errorf("plain: want every joint point to read 0 (the phantom), got %d of %d", zero, len(points))
					}
					continue
				}
				if zero > 0 || slit > 0 {
					t.Errorf("%s: %d joint points not solid, %d slit", v, zero, slit)
				}
			}
		})
	}
}

// Against the exact reference, both bridges are exact outside and a bound
// inside: never deeper than the true distance, never the wrong sign. Each logs
// its worst miss outside and its worst shortfall inside.
//
// The ungated bridge is exact outside although it is a max: each extended
// segment is an exact semi-infinite extrusion, so outside, wherever p is
// within a neighbour's span the bridge reads at least that neighbour's own
// distance, and beyond both spans at least the end segment's.
func TestShaftPhase0_MatchesReference(t *testing.T) {
	for _, c := range roundStackCases() {
		t.Run(c.name, func(t *testing.T) {
			r := c.stack.reach()
			lo := [3]float64{-r, -r, c.stack[0].from}
			hi := [3]float64{r, r, c.stack[len(c.stack)-1].to}
			points := shaftGrid3(lo, hi, 8, 24)
			ref := c.stack.reference()
			for _, v := range []string{"bridged", "gated"} {
				got := evalShaft(t, shaftDoc(v, c.stack.gsl()), "solid", points)
				worstOut, worstIn, failures := 0.0, 0.0, 0
				for i, q := range points {
					want := ref(q)
					var bad bool
					if want > 0 {
						miss := math.Abs(got[i] - want)
						worstOut = math.Max(worstOut, miss)
						bad = miss > shaftTolerance
					} else {
						// Inside, a bound reads no deeper than the truth, and not
						// outside: want <= got <= 0.
						worstIn = math.Max(worstIn, got[i]-want)
						bad = got[i] < want-shaftTolerance || (want < -shaftTolerance && got[i] >= 0)
					}
					if bad {
						if failures++; failures <= 3 {
							t.Errorf("%s at %v: field %v, reference %v", v, q, got[i], want)
						}
					}
				}
				t.Logf("%s: outside worst miss %.3g mm; inside worst shortfall %.3g mm (a bound); %d failures",
					v, worstOut, worstIn, failures)
			}
		})
	}
}

// --- Mixed convexity ---------------------------------------------------------

// hexAF is the across-flats size of the hex, and roundDia the round beside it,
// between the hex's flats and its corners: at some angles the hex is the larger
// side of the joint and at others the smaller.
const (
	hexAF    = 20.0
	roundDia = 22.0
)

func hexPolygon(af float64) poly2 {
	r := af / math.Sqrt(3)
	var pg poly2
	for i := 0; i < 6; i++ {
		a := math.Pi / 3 * float64(i)
		pg = append(pg, [2]float64{r * math.Cos(a), r * math.Sin(a)})
	}
	return pg
}

func hexProfileGSL(af float64) string {
	pts := make([]string, 0, 6)
	for _, pt := range hexPolygon(af) {
		x, y := fmt.Sprintf("%.9f", pt[0]), fmt.Sprintf("%.9f", pt[1])
		if pt[0] < 0 {
			x = fmt.Sprintf("neg(%.9f)", -pt[0])
		}
		if pt[1] < 0 {
			y = fmt.Sprintf("neg(%.9f)", -pt[1])
		}
		pts = append(pts, "["+x+", "+y+"]")
	}
	return "profileXY([" + strings.Join(pts, ", ") + "])"
}

// shaftExtruded is the exact distance to a 2D section kept between two planes.
func shaftExtruded(d2 float64, z, low, high float64) float64 {
	h := math.Abs(z-(low+high)/2) - (high-low)/2
	return math.Hypot(math.Max(d2, 0), math.Max(h, 0)) + math.Min(math.Max(d2, h), 0)
}

// A hex beside a round between its flats and corners, in both orders. Outside,
// the reference is exact (the min of two exact solids); inside, the sign comes
// from containment. The joint plane is solid wherever both sides are, and where
// the hex's corners stand proud of the round its end face is a real face: that
// is outside, so the exact check covers it.
func TestShaftPhase0_MixedConvexity(t *testing.T) {
	hex := hexPolygon(hexAF)
	orders := map[string]bool{"hex below": true, "hex above": false}
	for name, hexBelow := range orders {
		t.Run(name, func(t *testing.T) {
			hexLo, hexHi, rLo, rHi := 0.0, 20.0, 20.0, 40.0
			if !hexBelow {
				hexLo, hexHi, rLo, rHi = 20, 40, 0, 20
			}
			hexSeg := fmt.Sprintf("spike.section(%s, %v, %v)", hexProfileGSL(hexAF), hexLo, hexHi)
			roundSeg := fmt.Sprintf("spike.cyl(%v, %v, %v)", roundDia, rLo, rHi)
			segs := "[" + hexSeg + ", " + roundSeg + "]"
			if !hexBelow {
				segs = "[" + roundSeg + ", " + hexSeg + "]"
			}
			ref := func(q [3]float64) float64 {
				dh := shaftExtruded(hex.distance([2]float64{q[0], q[1]}), q[2], hexLo, hexHi)
				dr := shaftExtruded(math.Hypot(q[0], q[1])-roundDia/2, q[2], rLo, rHi)
				return math.Min(dh, dr)
			}
			r := hexAF / math.Sqrt(3)
			points := shaftGrid3([3]float64{-r, -r, 0}, [3]float64{r, r, 40}, 6, 26)
			// The joint plane, inside both; and just above and below it round the
			// corners, where the convexity flips.
			joint := jointPoints([]float64{20}, []float64{hexAF / 2})
			var rim [][3]float64
			for a := 0.0; a < 2*math.Pi; a += math.Pi / 90 {
				for _, rad := range []float64{10.2, 10.8, 11.2, 11.4} {
					for _, dz := range []float64{-0.3, -0.05, 0.05, 0.3} {
						rim = append(rim, [3]float64{rad * math.Cos(a), rad * math.Sin(a), 20 + dz})
					}
				}
			}
			points = append(append(points, joint...), rim...)
			for _, v := range []string{"bridged", "gated"} {
				got := evalShaft(t, shaftDoc(v, segs), "solid", points)
				worstOut, notSolid, failures := 0.0, 0, 0
				for i, q := range points {
					want := ref(q)
					if want > 0 {
						miss := math.Abs(got[i] - want)
						worstOut = math.Max(worstOut, miss)
						if miss > shaftTolerance {
							if failures++; failures <= 3 {
								t.Errorf("%s outside at %v: field %v, reference %v", v, q, got[i], want)
							}
						}
					} else if got[i] > -1e-3 && want < -1e-3 {
						if notSolid++; notSolid <= 3 {
							t.Errorf("%s inside at %v: field %v, reference %v", v, q, got[i], want)
						}
					}
				}
				t.Logf("%s: outside worst miss %.3g mm over %d points; %d inside points not solid", v, worstOut, len(points), notSolid)
			}
		})
	}
}

// --- Provability -------------------------------------------------------------

const cellHalfShaft = 0.25

func shaftIntervals(t *testing.T, doc string, boxes [][6]float32) []IntervalSample {
	t.Helper()
	ctx := context.Background()
	if err := startQuery(ctx); err != nil {
		t.Fatal(err)
	}
	defer endQuery()
	e := New(shaftSpikeManager(), nil)
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

// boundaryCentres are points on the stack's surface, round its meridian at a
// few angles.
func (s roundStack) boundaryCentres() [][3]float64 {
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

// Surface cells bound soundly and tightly, the gate's select included; and the
// cells on the joint planes, where the phantom was, prove solid.
func TestShaftPhase0_SurfaceCellsAreTight(t *testing.T) {
	diagonal := 2 * cellHalfShaft * math.Sqrt(3)
	for _, c := range roundStackCases() {
		t.Run(c.name, func(t *testing.T) {
			ref := c.stack.reference()
			surface := c.stack.boundaryCentres()
			zs, radii := c.stack.joints()
			var joint [][3]float64
			for j, z := range zs {
				for _, x := range []float64{0, radii[j] / 2, radii[j] - 1} {
					joint = append(joint, [3]float64{x, 0, z})
				}
			}
			for _, v := range shaftVariants {
				doc := shaftDoc(v, c.stack.gsl())
				boxes := make([][6]float32, 0, len(surface)+len(joint))
				for _, q := range append(append([][3]float64{}, surface...), joint...) {
					boxes = append(boxes, [6]float32{
						float32(q[0] - cellHalfShaft), float32(q[1] - cellHalfShaft), float32(q[2] - cellHalfShaft),
						float32(q[0] + cellHalfShaft), float32(q[1] + cellHalfShaft), float32(q[2] + cellHalfShaft),
					})
				}
				bounds := shaftIntervals(t, doc, boxes)
				widest, unsound := 0.0, 0
				for i, q := range surface {
					lo, hi := float64(bounds[i].Lower), float64(bounds[i].Upper)
					widest = math.Max(widest, (hi-lo)/diagonal)
					if w := ref(q); w < lo-shaftTolerance || w > hi+shaftTolerance {
						unsound++
					}
				}
				solid := 0
				for i := range joint {
					if bounds[len(surface)+i].Upper < 0 {
						solid++
					}
				}
				t.Logf("%s: %d surface cells, widest %.2f diagonals, %d unsound; %d of %d joint cells proved solid",
					v, len(surface), widest, unsound, solid, len(joint))
				if unsound > 0 || widest > 4 {
					t.Errorf("%s: %d unsound surface cells, widest %.2f diagonals", v, unsound, widest)
				}
				// The gate costs this: on a joint cell the plain pool straddles 0, so
				// the select is undecided and the cell's upper bound stays plain's.
				// Logged for gated, required of bridged.
				if v == "bridged" && solid != len(joint) {
					t.Errorf("%s: %d of %d joint cells proved solid", v, solid, len(joint))
				}
			}
		})
	}
}

// certify proves the space just off the shaft empty: beside its larger
// diameter, beyond each end, and in the corner of a step.
func TestShaftPhase0_CertifiesClearance(t *testing.T) {
	stack := roundStack{{20, 0, 20}, {12, 20, 40}}
	const gap = 0.25
	regions := map[string][6]float64{
		"beside":           {10 + gap, -12, -5, 15, 12, 45},
		"beyond the top":   {-12, -12, 40 + gap, 12, 12, 45},
		"in the step":      {6 + gap, -3, 20 + gap, 9, 3, 30},
		"under the bottom": {-12, -12, -5, 12, 12, -gap},
	}
	for _, v := range []string{"bridged", "gated"} {
		for where, region := range regions {
			e := New(shaftSpikeManager(), nil)
			r, err := e.Certify(context.Background(), nil, shaftDoc(v, stack.gsl()), "", "solid", region, 0, 0)
			if err != nil {
				t.Fatalf("certify %s %s: %v", v, where, err)
			}
			t.Logf("%s, %s: proven=%v in %d evals", v, where, r.Proven, r.Evals)
			if !r.Proven {
				t.Errorf("%s, %s: want proven, got disproven=%v indeterminate=%v [%v, %v]",
					v, where, r.Disproven, r.Indeterminate, r.RegionLower, r.RegionUpper)
			}
		}
	}
}

// --- Spikes ------------------------------------------------------------------

// Spike (a): WGSL size against the number of segments, for each fold. Logged
// only; the spec records the decision.
func TestShaftPhase0_SpikeAccumulatorGrowth(t *testing.T) {
	e := New(shaftSpikeManager(), nil)
	list := func(xs []string) string { return "[" + strings.Join(xs, ", ") + "]" }
	for _, n := range []int{2, 4, 8, 16} {
		var baseline, chain []string
		for i := 0; i < n; i++ {
			dia := 12.0 + 4*float64(i%3)
			baseline = append(baseline, fmt.Sprintf("spike.cyl(%v, %v, %v)", dia, 5*i, 5*(i+1)))
			chain = append(chain, fmt.Sprintf("spike.chainCyl(%v, 5.0)", dia))
		}
		docs := map[string]string{
			"plain":   shaftDoc("plain", list(baseline)),
			"bridged": shaftDoc("bridged", list(baseline)),
			"gated":   shaftDoc("gated", list(baseline)),
			"chained": fmt.Sprintf("@scene(140)\nimport spike \"geom/shaft\";\nsolid = spike.chained(%s, 0.0);\nreturn solid;\n", list(chain)),
		}
		for _, v := range []string{"plain", "bridged", "gated", "chained"} {
			res, err := e.Shader(t.Context(), nil, docs[v], "", ShaderOptions{Target: "wgsl", OutDir: t.TempDir()})
			if err != nil {
				t.Errorf("%d segments, %s: %v", n, v, err)
				continue
			}
			t.Logf("%2d segments, %-7s: %7d bytes of WGSL", n, v, len(res.Code))
		}
	}
}

// The chained fold reads the same shaft as its baseline twin.
func TestShaftPhase0_SpikeChainedMatchesBaseline(t *testing.T) {
	stack := roundStack{{20, 0, 15}, {14, 15, 18}, {20, 18, 40}}
	chain := "[spike.chainCyl(20.0, 15.0), spike.chainCyl(14.0, 3.0), spike.chainCyl(20.0, 22.0)]"
	points := shaftGrid3([3]float64{-10, -10, 0}, [3]float64{10, 10, 40}, 4, 14)
	base := evalShaft(t, shaftDoc("bridged", stack.gsl()), "solid", points)
	chained := evalShaft(t, fmt.Sprintf("@scene(140)\nimport spike \"geom/shaft\";\nsolid = spike.chained(%s, 0.0);\nreturn solid;\n", chain), "solid", points)
	for i, q := range points {
		if math.Abs(base[i]-chained[i]) > shaftTolerance {
			t.Fatalf("at %v: chained %v, baseline %v", q, chained[i], base[i])
		}
	}
}

const endHolesPrelude = `@scene(140)
import spike "geom/shaft";
import drill "mechanical/drill";
segs = [spike.cyl(20.0, 0.0, 20.0), spike.cyl(12.0, 20.0, 40.0)];
`

// Spike (b), as the drill library imagined it: drill closures made in the
// document and passed through the shaft package's function into drill.hole,
// one end a list and the other a lone feature, the bottom one flipped by a
// remap of the fold's result. Refused in Phase 0, where it was pinned: a lone
// function forwarded across packages did not fold as a list of one. Since
// Phase 2 the closure pass follows a list, or a lone function, handed on to
// another function's fold, and the holes are cut where the voids spike below
// cuts them.
func TestShaftPhase0_SpikeHoleListsThroughShaft(t *testing.T) {
	doc := endHolesPrelude + `solid = spike.withHoles(segs, 0.0, 40.0, drill.centerHole(3.0, 8.0, 10.0), [drill.blind(5.0, 12.0), drill.chamfer(5.0, 0.5)]);
return solid;
`
	cases := []struct {
		q    [3]float64
		want string
	}{
		{[3]float64{0, 0, 35}, "void"},    // in the top hole, 5 below its mouth
		{[3]float64{2, 0, 35}, "void"},    // inside the 5 mm bore
		{[3]float64{3.5, 0, 35}, "solid"}, // in the wall beyond it
		{[3]float64{0, 0, 26}, "solid"},   // below the top hole's 12 depth and point
		{[3]float64{0, 0, 5}, "void"},     // in the center's pilot, 5 above the bottom face
		{[3]float64{3.5, 0, 0.5}, "void"}, // in the center's 60-degree countersink
		{[3]float64{0, 0, 20}, "solid"},   // the joint plane, on the axis
		{[3]float64{8, 0, 10}, "solid"},
	}
	points := make([][3]float64, len(cases))
	for i, c := range cases {
		points[i] = c.q
	}
	got := evalShaft(t, doc, "solid", points)
	for i, c := range cases {
		state := "solid"
		if got[i] > 0 {
			state = "void"
		}
		if state != c.want {
			t.Errorf("at %v: field %v reads %s, want %s", c.q, got[i], state, c.want)
		}
	}
}

// Spike (b), the ways round it: the document folds each hole itself and hands
// the shaft the voids. "placed by the caller" places them too (the bottom one
// flipped by a remap of the fold's result, in the document); "placed by shaft"
// hands them over with their mouths at 0 and lets the shaft package remap them,
// which is where a remap of a shape that arrives as a parameter has been a
// silent no-op.
func TestShaftPhase0_SpikeEndHoleVoids(t *testing.T) {
	docs := map[string]string{
		"placed by the caller": endHolesPrelude + `top = drill.hole([drill.blind(5.0, 12.0), drill.chamfer(5.0, 0.5)], 40.0);
bottom = remap(drill.hole(drill.centerHole(3.0, 8.0, 10.0), 0.0), [p.x, -p.y, 0.0 - p.z]);
solid = spike.withVoids(segs, top, bottom);
return solid;
`,
		"placed by shaft": endHolesPrelude + `top = drill.hole([drill.blind(5.0, 12.0), drill.chamfer(5.0, 0.5)], 0.0);
bottom = drill.hole(drill.centerHole(3.0, 8.0, 10.0), 0.0);
solid = spike.withVoidsAt(segs, 0.0, 40.0, top, bottom);
return solid;
`,
	}
	cases := []struct {
		q    [3]float64
		want string
	}{
		{[3]float64{0, 0, 35}, "void"},    // in the top hole, 5 below its mouth
		{[3]float64{2, 0, 35}, "void"},    // inside the 5 mm bore
		{[3]float64{3.5, 0, 35}, "solid"}, // in the wall beyond it
		{[3]float64{0, 0, 26}, "solid"},   // below the top hole's 12 depth and point
		{[3]float64{0, 0, 5}, "void"},     // in the center's pilot, 5 above the bottom face
		{[3]float64{3.5, 0, 0.5}, "void"}, // in the center's 60-degree countersink
		{[3]float64{0, 0, 20}, "solid"},   // the joint plane, on the axis
		{[3]float64{8, 0, 10}, "solid"},
	}
	points := make([][3]float64, len(cases))
	for i, c := range cases {
		points[i] = c.q
	}
	for _, name := range []string{"placed by the caller", "placed by shaft"} {
		t.Run(name, func(t *testing.T) {
			got := evalShaft(t, docs[name], "solid", points)
			for i, c := range cases {
				state := "solid"
				if got[i] > 0 {
					state = "void"
				}
				t.Logf("at %v: %.4f (%s)", c.q, got[i], state)
				if state != c.want {
					t.Errorf("at %v: field %v reads %s, want %s", c.q, got[i], state, c.want)
				}
			}
		})
	}
}
