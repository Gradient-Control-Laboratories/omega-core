package gsleval_test

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/gsleval"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// The frame each primitive declares (builtin/3d.gsl), read off a plain call:
// lengths in mm, angles in radians, u the coordinate that wraps where one does.
// A cylinder, a cone and an axis are cylindrical about their axis; a sphere is
// azimuth and polar angle; a capped cylinder's v turns from along the axis on
// the side to across it on a cap; a plane is its in-plane offsets; a box gives
// each face its own two in-plane axes, cycled so no axis is u on two faces
// that meet.
const primitiveDoc = `
s0 = sphere(1.0);
s1 = sphere([10.0, 0.0, 0.0], 1.0);
s2 = sphere([0.0, 0.0, 0.0], 1.0, [1.0, 0.0, 0.0], [0.0, 1.0, 0.0]);
ax = axisZ([0.0, 0.0, 0.0]);
cy = cylinder([5.0, 0.0, 0.0], [0.0, 0.0, 1.0], 2.0);
cz = cylinderZ([0.0, 0.0, 0.0], 1.0);
cc = cylinderCapped([0.0, 0.0, 0.0], [0.0, 0.0, 4.0], 1.0);
cn = cone([0.0, 0.0, 0.0], [0.0, 0.0, 1.0], 0.5);
pl = plane([0.0, 0.0, 0.0], [0.0, 0.0, 1.0]);
bx = box([-1.0, -1.0, -1.0], [1.0, 1.0, 1.0]);
bc = boxCentered([0.0, 0.0, 0.0], [2.0, 2.0, 2.0]);
return s0;
`

func TestPrimitiveFrames(t *testing.T) {
	const pi = math.Pi
	cases := []struct {
		shape string
		pts   [][3]float64
		u, v  []float64
	}{
		{"s0", [][3]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}, []float64{0, pi / 2, 0}, []float64{pi / 2, pi / 2, 0}},
		{"s1", [][3]float64{{10, 1, 0}, {10, 0, -1}}, []float64{pi / 2, 0}, []float64{pi / 2, pi}},
		// About the pole +X from the reference +Y.
		{"s2", [][3]float64{{0, 1, 0}, {0, 0, 1}}, []float64{0, pi / 2}, []float64{pi / 2, pi / 2}},
		{"ax", [][3]float64{{1, 0, 5}, {0, 1, 2}}, []float64{0, pi / 2}, []float64{5, 2}},
		{"cy", [][3]float64{{7, 0, 1}, {5, 2, 3}}, []float64{0, pi / 2}, []float64{1, 3}},
		{"cz", [][3]float64{{1, 0, 3}, {0, 1, -2}}, []float64{0, pi / 2}, []float64{3, -2}},
		// Two points on the side, v from the middle at z = 2; one on a cap, v from the axis.
		{"cc", [][3]float64{{1, 0, 2}, {0, 1, 3}, {0.5, 0, 4}}, []float64{0, pi / 2, 0}, []float64{0, 1, 0.5}},
		{"cn", [][3]float64{{1, 0, 3}, {0, 1, 2}}, []float64{0, pi / 2}, []float64{3, 2}},
		{"pl", [][3]float64{{2, 3, 0}, {-1, 4, 0}}, []float64{2, -1}, []float64{3, 4}},
		// An x face (y, z), a y face (z, x), a z face (x, y).
		{"bx", [][3]float64{{1, 0.3, 0.5}, {0.2, 1, 0.4}, {0.6, 0.7, 1}}, []float64{0.3, 0.4, 0.6}, []float64{0.5, 0.2, 0.7}},
		{"bc", [][3]float64{{1, 0.3, 0.5}, {0.2, 1, 0.4}, {0.6, 0.7, 1}}, []float64{0.3, 0.4, 0.6}, []float64{0.5, 0.2, 0.7}},
	}
	e := gsleval.New(&source.MockManager{}, nil)
	for _, c := range cases {
		t.Run(c.shape, func(t *testing.T) {
			results, err := e.EvalGsl(context.Background(), nil, primitiveDoc, "", []string{c.shape + ".u", c.shape + ".v"}, c.pts)
			require.NoError(t, err)
			require.Equal(t, 2, len(results))
			for i := range c.pts {
				if got := results[0].Samples[0][i].Value; !within(got, c.u[i]) {
					t.Errorf("%s.u at %v = %v, want %v", c.shape, c.pts[i], got, c.u[i])
				}
				if got := results[1].Samples[0][i].Value; !within(got, c.v[i]) {
					t.Errorf("%s.v at %v = %v, want %v", c.shape, c.pts[i], got, c.v[i])
				}
			}
		})
	}
}

// A torus's frame is a sphere's bent round the ring: u the azimuth, v the
// angle round the tube from its top, which unlike a sphere's polar angle wraps,
// along the bottom of the tube.
func TestTorusFrame(t *testing.T) {
	const doc = "ring = torus(3.0, 1.0);\nreturn ring;\n"
	const pi = math.Pi
	// Outer equator, top, outer equator a quarter round, inner equator.
	pts := [][3]float64{{4, 0, 0}, {3, 0, 1}, {0, 4, 0}, {2, 0, 0}}
	wantU := []float64{0, 0, pi / 2, 0}
	wantV := []float64{pi / 2, 0, pi / 2, -pi / 2}
	e := gsleval.New(&source.MockManager{}, nil)
	results, err := e.EvalGsl(context.Background(), nil, doc, "", []string{"ring.u", "ring.v"}, pts)
	require.NoError(t, err)
	require.Equal(t, 2, len(results))
	for i := range pts {
		if got := results[0].Samples[0][i].Value; !within(got, wantU[i]) {
			t.Errorf("ring.u at %v = %v, want %v", pts[i], got, wantU[i])
		}
		if got := results[1].Samples[0][i].Value; !within(got, wantV[i]) {
			t.Errorf("ring.v at %v = %v, want %v", pts[i], got, wantV[i])
		}
	}
}
