package gsleval

import (
	"context"
	"math"
	"math/rand"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// scalarSamples evaluates scalar expressions at pts, a row per expression.
func scalarSamples(t *testing.T, exprs []string, pts [][3]float64) [][]float64 {
	t.Helper()
	mgr := source.NewLocalManager("../../stdlib")
	e := New(mgr, nil)
	groups, err := e.EvalGsl(context.Background(), nil, "return p.x;", "", exprs, pts)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	out := make([][]float64, len(exprs))
	for i := range exprs {
		out[i] = make([]float64, len(pts))
		for j := range pts {
			out[i][j] = float64(groups[i].Samples[0][j].Value)
		}
	}
	return out
}

func randomPoints(n int, reach float64) [][3]float64 {
	r := rand.New(rand.NewSource(11))
	pts := make([][3]float64, n)
	for i := range pts {
		pts[i] = [3]float64{(r.Float64() - 0.5) * reach, (r.Float64() - 0.5) * reach, (r.Float64() - 0.5) * reach}
	}
	return pts
}

// The Voronoi scan in GSL sits on the cellular builtins' own lattice, so at
// full randomness and no smoothing its Euclidean F1 is cellular(size); the
// metrics order as Chebyshev <= Euclidean <= Manhattan, since they do point
// by point; smoothing only lowers F1 and never below 0; and the true edge
// distance is never under (F2 - F1) / 2, which is what it is where the two
// nearest points face each other squarely and less than it is elsewhere.
func TestVoronoiAgreesWithTheCellularLattice(t *testing.T) {
	pts := randomPoints(500, 120)
	s := scalarSamples(t, []string{
		"noise.voronoi(5, 1, 0, 0)",
		"noise.cellular(5)",
		"noise.voronoi(5, 1, 1, 0)",
		"noise.voronoi(5, 1, 2, 0)",
		"noise.voronoi(5, 1, 0, 0.3)",
		"noise.voronoiEdge(5, 1)",
		"noise.cellularEdges(5) / 2",
	}, pts)
	for j := range pts {
		if math.Abs(s[0][j]-s[1][j]) > 1e-3 {
			t.Fatalf("at %v the Euclidean Voronoi is %g, cellular F1 %g", pts[j], s[0][j], s[1][j])
		}
		if s[3][j] > s[0][j]+1e-4 || s[0][j] > s[2][j]+1e-4 {
			t.Fatalf("at %v the metrics are out of order: Chebyshev %g, Euclidean %g, Manhattan %g", pts[j], s[3][j], s[0][j], s[2][j])
		}
		if s[4][j] > s[0][j]+1e-4 || s[4][j] < 0 {
			t.Fatalf("at %v smooth F1 is %g against F1 %g", pts[j], s[4][j], s[0][j])
		}
		if s[5][j] < s[6][j]-1e-3 {
			t.Fatalf("at %v the edge distance %g is under (F2 - F1) / 2 = %g", pts[j], s[5][j], s[6][j])
		}
	}
}

// The filtered fractals keep the ranges their docs state, measured off a
// screen where every octave is drawn: fbm and anisotropic centred on 0
// inside +-1, turbulence, ridged and hybrid inside [0, 1], and the warp
// moving the point no further than its amount.
func TestFilteredFractalsKeepTheirRanges(t *testing.T) {
	pts := randomPoints(1000, 200)
	exprs := []string{
		"noise.fbm(10, 6, 0.5, 2)",
		"noise.anisotropic(4, [1, 0, 0], 10, 4)",
		"noise.turbulence(10, 6, 0.5, 2)",
		"noise.ridged(10, 6, 0.5, 2, 1)",
		"noise.hybrid(10, 6, 0.5, 2, 0.7)",
		"length(noise.warp(10, 3) - p)",
		"texture.bands(6, [1, 0, 0], 0, 0.5, 3)",
	}
	lo := []float64{-1, -1, 0, 0, 0, 0, 0}
	hi := []float64{1, 1, 1, 1, 1, 3.3, 1}
	s := scalarSamples(t, exprs, pts)
	for i, row := range s {
		mean := 0.0
		for _, v := range row {
			if v < lo[i] || v > hi[i] {
				t.Errorf("%s = %g, outside [%g, %g]", exprs[i], v, lo[i], hi[i])
				break
			}
			mean += v / float64(len(row))
		}
		if i < 2 && math.Abs(mean) > 0.05 {
			t.Errorf("%s has mean %g, want about 0", exprs[i], mean)
		}
	}
}

// Detail is fractional: one octave is the lattice noise itself, doubled.
func TestFbmOfOneOctaveIsTheLattice(t *testing.T) {
	pts := randomPoints(200, 100)
	s := scalarSamples(t, []string{"noise.fbm(7, 1, 0.5, 2)", "2 * noise.gradient(7)"}, pts)
	for j := range pts {
		if math.Abs(s[0][j]-s[1][j]) > 1e-4 {
			t.Fatalf("at %v fbm of detail 1 is %g, the doubled lattice %g", pts[j], s[0][j], s[1][j])
		}
	}
}

// Triplanar weights sum to 1 and follow the normal: a normal along an axis
// puts all the weight on the plane facing it.
func TestTriplanarWeights(t *testing.T) {
	s := scalarSamples(t, []string{
		"texture.triplanarWeights([0.3, -0.5, 0.8], 4).x + texture.triplanarWeights([0.3, -0.5, 0.8], 4).y + texture.triplanarWeights([0.3, -0.5, 0.8], 4).z",
		"texture.triplanarWeights([0, 0, 2], 4).z",
		"texture.triplanarWeights([0, -1, 0], 4).y",
	}, [][3]float64{{0, 0, 0}})
	if math.Abs(s[0][0]-1) > 1e-5 || math.Abs(s[1][0]-1) > 1e-5 || math.Abs(s[2][0]-1) > 1e-5 {
		t.Errorf("triplanar weights: sum %g, +Z face %g, -Y face %g; want 1 each", s[0][0], s[1][0], s[2][0])
	}
}
