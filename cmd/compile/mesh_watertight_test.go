package main_test

import (
	"math"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/assert"
)

// unmatchedDirectedEdges counts the directed edges of tris that are not used exactly once with their reverse used
// exactly once, welding corners by exact position. A closed, consistently wound, manifold surface has none.
func unmatchedDirectedEdges(tris [][3][3]float32) int {
	type edge struct{ a, b [3]float32 }
	uses := map[edge]int{}
	for _, tri := range tris {
		for v := 0; v < 3; v++ {
			uses[edge{tri[v], tri[(v+1)%3]}]++
		}
	}
	unmatched := 0
	for e, n := range uses {
		if n != 1 || uses[edge{e.b, e.a}] != 1 {
			unmatched++
		}
	}
	return unmatched
}

// Marching cubes' mesh is watertight: every edge is used once in each direction. Each cell computes the crossing on
// each of its edges, and the two cells that share a grid edge must compute the same point to the last bit, or the
// mesh opens along the edge. The solids are off the grid so no vertex lands on a grid point.
func TestRun_STL_MarchingCubesIsWatertight(t *testing.T) {
	// A box 20 x 14 x 10 with every edge rounded 2: its volume is the inner box 16 x 10 x 6 grown by 2.
	const x, y, z, r = 16.0, 10.0, 6.0, 2.0
	roundBox := x*y*z + 2*r*(x*y+y*z+z*x) + math.Pi*r*r*(x+y+z) + 4.0/3*math.Pi*r*r*r
	parts := []struct {
		name, source string
		volume       float64
	}{
		{"sphere", "return length(p - [0.31, 0.17, 0.23]) - 10;\n", 4.0 / 3 * math.Pi * 1000},
		{"hollow ball", "r = length(p - [0.31, 0.17, 0.23]);\nreturn max(r - 10, -(r - 6));\n",
			4.0 / 3 * math.Pi * (1000 - 216)},
		{"box with every edge rounded", "d = abs(p - [0.31, 0.17, 0.23]) - [8, 5, 3];\n" +
			"return length(max(d, 0)) + min(max(d.x, d.y, d.z), 0) - 2;\n", roundBox},
	}
	for _, part := range parts {
		for _, depth := range []string{"6", "7"} {
			tris := stlTriangles(t, meshed(t, part.source, "--algo", "marching-cubes", "--subdivision", depth,
				"--bbox", "auto"))
			assert.Equal(t, 0, unmatchedDirectedEdges(tris), "%s, depth %s: unmatched directed edges", part.name, depth)
			closed, volume := closedVolume(tris)
			assert.True(t, closed, "%s, depth %s: every edge shared by exactly two triangles", part.name, depth)
			assert.True(t, math.Abs(volume-part.volume) < 0.03*part.volume,
				"%s, depth %s: volume %.1f, want about %.1f", part.name, depth, volume, part.volume)
		}
	}
}
