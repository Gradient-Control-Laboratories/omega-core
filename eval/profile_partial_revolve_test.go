//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
)

// revolve(profile, axisPoint, axisDirection, angle): the profile turned
// through angle about a line in its plane, from the sketch plane, positive
// toward -z as revolveAngle runs, with flat caps.

// revolveCase is a profile, its axis, and the angle it is turned through.
type revolveCase struct {
	name  string
	loop  refLoop
	at    [2]float64
	dir   [2]float64
	angle float64
}

func revolveCases() []revolveCase {
	return []revolveCase{
		{"L, a quarter turn", lShape(), [2]float64{-10, 0}, [2]float64{0, 1}, math.Pi / 2},
		{"L, three quarters", lShape(), [2]float64{-10, 0}, [2]float64{0, 1}, 3 * math.Pi / 2},
		{"L, a sixth the other way", lShape(), [2]float64{-10, 0}, [2]float64{0, 1}, -math.Pi / 3},
		{"arc L, a third", arcL(), [2]float64{-10, 0}, [2]float64{0, 1}, 2 * math.Pi / 3},
	}
}

// axisCoordinates is p's distance along the axis, its signed distance to the
// right of the axis in the sketch plane, and the unit axis and right normal.
func (c revolveCase) axisCoordinates(q [3]float64) (along, right float64, d, n [2]float64) {
	l := math.Hypot(c.dir[0], c.dir[1])
	d = [2]float64{c.dir[0] / l, c.dir[1] / l}
	n = [2]float64{d[1], -d[0]}
	wx, wy := q[0]-c.at[0], q[1]-c.at[1]
	return wx*d[0] + wy*d[1], wx*n[0] + wy*n[1], d, n
}

// sketchPoint is the sketch-plane point a distance along the axis and r to
// its right.
func (c revolveCase) sketchPoint(along, r float64, d, n [2]float64) [2]float64 {
	return [2]float64{c.at[0] + along*d[0] + r*n[0], c.at[1] + along*d[1] + r*n[1]}
}

// inSweep reports whether the angle theta, from revolveAngle, lies within
// the sweep.
func (c revolveCase) inSweep(theta float64) bool {
	s := math.Copysign(1, c.angle)
	t := math.Mod(s*theta+4*math.Pi, 2*math.Pi)
	return t <= math.Abs(c.angle)
}

// reference is the exact signed distance to the swept solid, from the
// sampled profile: the revolved surface within the sweep, and each cap, the
// profile lying flat in its half-plane, at sqrt(h^2 + max(d, 0)^2).
func (c revolveCase) reference(ref refProfile) func([3]float64) float64 {
	return func(q [3]float64) float64 {
		along, xr, d, n := c.axisCoordinates(q)
		z := q[2]
		rho := math.Hypot(xr, z)
		theta := math.Atan2(-z, xr)
		dRev := ref.distance(c.sketchPoint(along, rho, d, n))
		capAt := func(alpha float64) float64 {
			r := xr*math.Cos(alpha) - z*math.Sin(alpha)
			h := xr*math.Sin(alpha) + z*math.Cos(alpha)
			return math.Hypot(h, math.Max(ref.distance(c.sketchPoint(along, r, d, n)), 0))
		}
		m := math.Min(capAt(0), capAt(c.angle))
		inside := c.inSweep(theta)
		if inside {
			m = math.Min(m, math.Abs(dRev))
		}
		if inside && dRev < 0 {
			return -m
		}
		return m
	}
}

func (c revolveCase) doc() string {
	return fmt.Sprintf("@scene(160)\nsolid = revolve(%s, [%v, %v], [%v, %v], %v);\nreturn solid;\n",
		c.loop.gsl(), c.at[0], c.at[1], c.dir[0], c.dir[1], c.angle)
}

// reach is the solid's farthest distance from the axis, and its extent along it.
func (c revolveCase) reach(ref refProfile) (float64, float64, float64) {
	far, lo, hi := 0.0, math.Inf(1), math.Inf(-1)
	for _, s := range ref.segments {
		along, xr, _, _ := c.axisCoordinates([3]float64{s[0][0], s[0][1], 0})
		far = math.Max(far, math.Abs(xr))
		lo, hi = math.Min(lo, along), math.Max(hi, along)
	}
	return far, lo, hi
}

func (c revolveCase) grid(ref refProfile, n int) [][3]float64 {
	far, lo, hi := c.reach(ref)
	// The axis runs along y here; x is the axis's right and z the depth.
	return grid3([3]float64{c.at[0] - far, lo, -far}, [3]float64{c.at[0] + far, hi, far}, 4, n)
}

// The partial revolve is exact: it agrees with the reference everywhere,
// inside the sweep and out, behind each cap and past the axis.
func TestRevolve_PartialMatchesReference(t *testing.T) {
	for _, c := range revolveCases() {
		t.Run(c.name, func(t *testing.T) {
			ref := newRefProfile(t, c.loop)
			points := c.grid(ref, 24)
			compareToReference(t, c.name, points, evalProfileField(t, c.doc(), "solid", points), c.reference(ref))
		})
	}
}

// Independently of that formula: the field's magnitude is the distance to a
// dense cloud of points on the solid's boundary (the revolved surface within
// the sweep and both cap regions), to within the cloud's spacing.
func TestRevolve_PartialMatchesAPointCloud(t *testing.T) {
	const spacing = 0.25
	for _, c := range revolveCases()[:2] {
		t.Run(c.name, func(t *testing.T) {
			ref := newRefProfile(t, c.loop)
			far, lo, hi := c.reach(ref)
			l := math.Hypot(c.dir[0], c.dir[1])
			d := [2]float64{c.dir[0] / l, c.dir[1] / l}
			n := [2]float64{d[1], -d[0]}
			place := func(along, r, theta float64) [3]float64 {
				s := c.sketchPoint(along, r*math.Cos(theta), d, n)
				return [3]float64{s[0], s[1], -r * math.Sin(theta)}
			}
			var cloud [][3]float64
			// The revolved surface: each boundary sample turned through the sweep.
			for _, b := range ref.boundarySamples(spacing) {
				along, r, _, _ := c.axisCoordinates([3]float64{b[0], b[1], 0})
				steps := int(math.Ceil(math.Abs(c.angle) * r / spacing))
				for k := 0; k <= steps; k++ {
					cloud = append(cloud, place(along, r, c.angle*float64(k)/float64(steps)))
				}
			}
			// The caps: the profile's inside, on a grid, at both ends.
			for along := lo; along <= hi; along += spacing {
				for r := 0.0; r <= far; r += spacing {
					if ref.distance(c.sketchPoint(along, r, d, n)) < 0 {
						cloud = append(cloud, place(along, r, 0), place(along, r, c.angle))
					}
				}
			}
			points := c.grid(ref, 9)
			got := evalProfileField(t, c.doc(), "solid", points)
			worst := 0.0
			for i, q := range points {
				best := math.Inf(1)
				for _, s := range cloud {
					best = math.Min(best, math.Hypot(math.Hypot(q[0]-s[0], q[1]-s[1]), q[2]-s[2]))
				}
				diff := math.Abs(math.Abs(got[i]) - best)
				worst = math.Max(worst, diff)
				if diff > spacing {
					t.Errorf("at %v |solid| = %v, the cloud is %v away", q, math.Abs(got[i]), best)
				}
			}
			t.Logf("%s: %d cloud points, worst %.3g mm over %d points", c.name, len(cloud), worst, len(points))
		})
	}
}

// Cells on its surface, revolved and capped, bound it soundly and tightly.
func TestRevolve_PartialSurfaceCellsAreTight(t *testing.T) {
	for _, c := range revolveCases() {
		t.Run(c.name, func(t *testing.T) {
			ref := newRefProfile(t, c.loop)
			l := math.Hypot(c.dir[0], c.dir[1])
			d := [2]float64{c.dir[0] / l, c.dir[1] / l}
			n := [2]float64{d[1], -d[0]}
			var centres [][3]float64
			for _, b := range ref.boundarySamples(2) {
				along, r, _, _ := c.axisCoordinates([3]float64{b[0], b[1], 0})
				for _, f := range []float64{0, 0.37, 0.71, 1} {
					theta := c.angle * f
					s := c.sketchPoint(along, r*math.Cos(theta), d, n)
					centres = append(centres, [3]float64{s[0], s[1], -r * math.Sin(theta)})
				}
			}
			for _, q := range gridAround(ref.box(), 0, 12) {
				if ref.distance(q) < 0 {
					along, r, _, _ := c.axisCoordinates([3]float64{q[0], q[1], 0})
					s := c.sketchPoint(along, r*math.Cos(c.angle), d, n)
					centres = append(centres, [3]float64{q[0], q[1], 0}, [3]float64{s[0], s[1], -r * math.Sin(c.angle)})
				}
			}
			checkSurfaceCells(t, c.name, c.doc(), centres, c.reference(ref))
		})
	}
}

// certify proves the side of the sketch plane the sweep turns away from
// empty, when the sweep is less than half a turn, and bounds proves a box
// that holds the solid.
func TestRevolve_PartialProves(t *testing.T) {
	for _, c := range revolveCases() {
		if math.Abs(c.angle) >= math.Pi {
			continue
		}
		t.Run(c.name, func(t *testing.T) {
			ref := newRefProfile(t, c.loop)
			far, lo, hi := c.reach(ref)
			// A positive sweep turns toward -z, a negative one toward +z.
			zLo, zHi := 0.25, far+5
			if c.angle < 0 {
				zLo, zHi = -far-5, -0.25
			}
			region := [6]float64{c.at[0] - far - 5, lo - 5, zLo, c.at[0] + far + 5, hi + 5, zHi}
			e := New(&source.MockManager{}, nil)
			r, err := e.Certify(context.Background(), nil, c.doc(), "", "solid", region, 0, 0)
			if err != nil {
				t.Fatalf("certify: %v", err)
			}
			if !r.Proven {
				t.Errorf("behind the sketch plane: want proven, got proven=%v disproven=%v in %d evals", r.Proven, r.Disproven, r.Evals)
			}

			box := [6]float64{c.at[0] - far - 10, lo - 10, -far - 10, c.at[0] + far + 10, hi + 10, far + 10}
			en, err := e.Enclose(context.Background(), nil, c.doc(), "", "solid", box, 0)
			if err != nil {
				t.Fatalf("enclose: %v", err)
			}
			if !en.Proven {
				t.Fatalf("enclose: want a proven box, got %q", en.Reason)
			}
			for _, q := range c.grid(ref, 14) {
				if c.reference(ref)(q) < -0.01 {
					for k := 0; k < 3; k++ {
						if q[k] < en.BboxMin[k]-1e-4 || q[k] > en.BboxMax[k]+1e-4 {
							t.Fatalf("a point inside, %v, lies outside the proven box %v %v", q, en.BboxMin, en.BboxMax)
						}
					}
				}
			}
			t.Logf("%s: proven box %v to %v", c.name, en.BboxMin, en.BboxMax)
		})
	}
}

// Its frame is the full revolve's: .u the angle round the axis, .v the
// profile's own .u where the revolve reads the profile.
func TestRevolve_PartialFrameIsTheFullRevolves(t *testing.T) {
	c := revolveCases()[3]
	ref := newRefProfile(t, c.loop)
	doc := fmt.Sprintf("@scene(160)\nswept = revolve(%s, [%v, %v], [%v, %v], %v);\nwhole = revolve(%s, [%v, %v], [%v, %v]);\nreturn swept, whole;\n",
		c.loop.gsl(), c.at[0], c.at[1], c.dir[0], c.dir[1], c.angle, c.loop.gsl(), c.at[0], c.at[1], c.dir[0], c.dir[1])
	points := c.grid(ref, 7)
	e := New(&source.MockManager{}, nil)
	groups, err := e.EvalGsl(context.Background(), nil, doc, "", []string{"swept.u", "whole.u", "swept.v", "whole.v"}, points)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	for k, name := range []string{"u", "v"} {
		for i, q := range points {
			got, want := float64(groups[2*k].Samples[0][i].Value), float64(groups[2*k+1].Samples[0][i].Value)
			if math.Abs(got-want) > 1e-4 {
				t.Fatalf("swept.%s at %v = %v, whole.%s %v", name, q, got, name, want)
			}
		}
	}
}
