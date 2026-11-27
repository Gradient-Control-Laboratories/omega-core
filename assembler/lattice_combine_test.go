package assembler_test

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
)

// differenceAt sums |a - b| over the points by constant folding one
// expression of remap(field, point) terms to a number.
func differenceAt(t *testing.T, a, b string, points [][3]float64) float64 {
	t.Helper()
	var terms []string
	for _, p := range points {
		pt := fmt.Sprintf("[%g, %g, %g]", p[0], p[1], p[2])
		terms = append(terms, fmt.Sprintf("abs(remap(%s, %s) - remap(%s, %s))", a, pt, b, pt))
	}
	src := "return " + strings.Join(terms, " + ") + ";"
	res, err := assembler.Assemble([]string{src})
	if err != nil {
		t.Fatalf("%v\n%s", err, src)
	}
	text := res.Program.PrettyPrint(0)
	// The document's own return comes first; its dependencies follow.
	i := strings.Index(text, "return ")
	if i < 0 {
		t.Fatalf("no return in:\n%s", text)
	}
	line, _, _ := strings.Cut(text[i+len("return "):], "\n")
	v, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(line), ";"), 64)
	if err != nil {
		t.Fatalf("%s - %s did not fold to a number:\n%s", a, b, text)
	}
	return v
}

// samplePoints spreads points over several cells, near centres, faces, edges
// and corners as well as in between.
var samplePoints = [][3]float64{
	{0.1, 0.2, 0.3}, {4.9, 6.4, 3.9}, {5.1, 6.6, 4.1}, {2.3, 9.7, -1.4}, {-7.2, 3.3, 12.6},
	{10.02, 13.1, 7.95}, {14.7, -2.2, 6.1}, {-3.3, -8.8, -2.5}, {7.5, 3.25, 2.0}, {19.9, 26.1, 15.9},
	{1.0, 6.5, 0.0}, {5.0, 0.0, 4.0},
}

// A blendable form joined sharply is the folded lattice.
func TestBeamLatticeCombineFormsMatchFolded(t *testing.T) {
	const cell = "[10, 13, 8]"
	for _, c := range []struct{ folded, explicit string }{
		{"beam.bcc(" + cell + ")", "beam.bcc(" + cell + ", unionSharp)"},
		{"beam.fcc(" + cell + ")", "beam.fcc(" + cell + ", unionSharp)"},
		{"beam.octet(" + cell + ")", "beam.octet(" + cell + ", unionSharp)"},
		{"beam.truncatedOctahedral(" + cell + ")", "beam.truncatedOctahedral(" + cell + ", unionSharp)"},
		{"beam.truncatedCubic(" + cell + ", 0.3)", "beam.truncatedCubic(" + cell + ", 0.3, unionSharp)"},
		{"beam.truncatedCubic(" + cell + ", 0)", "beam.truncatedCubic(" + cell + ", 0, unionSharp)"},
		{"beam.kelvinCell(" + cell + ", 0.4)", "beam.kelvinCell(" + cell + ", 0.4, unionSharp)"},
		{"beam.fluorite(" + cell + ")", "beam.fluorite(" + cell + ", unionSharp)"},
		{"beam.cubicDouble(" + cell + ")", "beam.cubicDouble(" + cell + ", unionSharp)"},
		{"foam.kelvin(" + cell + ")", "foam.kelvin(" + cell + ", unionSharp)"},
		{"foam.rhombicDodecahedral(" + cell + ")", "foam.rhombicDodecahedral(" + cell + ", unionSharp)"},
	} {
		if d := differenceAt(t, c.folded, c.explicit, samplePoints); d > 1e-9 {
			t.Errorf("%s differs from %s by %g summed over the sample points", c.explicit, c.folded, d)
		}
	}
}

// The deprecated lattice.bcc is beam.bcc less the beam radius, as gsl fix
// rewrites it.
func TestDeprecatedLatticeBccIsBeamBcc(t *testing.T) {
	for _, old := range []string{"lattice.bcc(10, 1)", "lattice.bcc(10, 1, unionSharp)"} {
		if d := differenceAt(t, old, "(beam.bcc(10) - 1)", samplePoints); d > 1e-9 {
			t.Errorf("%s differs from beam.bcc(10) - 1 by %g summed over the sample points", old, d)
		}
	}
}

// voronoiWall is the exact distance from p to the boundary of its Voronoi cell
// among the lattice points offset by offsets in cubic cells of the given size.
func voronoiWall(p [3]float64, size float64, offsets [][3]float64) float64 {
	var sites [][3]float64
	base := [3]float64{math.Floor(p[0] / size), math.Floor(p[1] / size), math.Floor(p[2] / size)}
	for i := -2.0; i <= 2; i++ {
		for j := -2.0; j <= 2; j++ {
			for k := -2.0; k <= 2; k++ {
				for _, o := range offsets {
					sites = append(sites, [3]float64{(base[0]+i)*size + o[0], (base[1]+j)*size + o[1], (base[2]+k)*size + o[2]})
				}
			}
		}
	}
	d2 := func(a, b [3]float64) float64 {
		x, y, z := a[0]-b[0], a[1]-b[1], a[2]-b[2]
		return x*x + y*y + z*z
	}
	s0 := sites[0]
	for _, s := range sites {
		if d2(p, s) < d2(p, s0) {
			s0 = s
		}
	}
	w := math.Inf(1)
	for _, s := range sites {
		if n := math.Sqrt(d2(s, s0)); n > 1e-12 {
			w = math.Min(w, (d2(p, s)-d2(p, s0))/(2*n))
		}
	}
	return w
}

// The Voronoi foams are the walls of their points' Voronoi cells.
func TestVoronoiFoamsAreVoronoiWalls(t *testing.T) {
	const size, h = 10.0, 5.0
	for _, c := range []struct {
		field   string
		offsets [][3]float64
	}{
		{"foam.kelvin(10)", [][3]float64{{0, 0, 0}, {h, h, h}}},
		{"foam.rhombicDodecahedral(10)", [][3]float64{{h, h, h}, {0, 0, h}, {0, h, 0}, {h, 0, 0}}},
	} {
		// One term per point, so errors at different points cannot cancel.
		var terms []string
		for _, p := range samplePoints {
			terms = append(terms, fmt.Sprintf("abs(remap(%s, [%g, %g, %g]) - %.17g)", c.field, p[0], p[1], p[2], voronoiWall(p, size, c.offsets)))
		}
		got := differenceAt(t, strings.Join(terms, " + "), "0", [][3]float64{{0, 0, 0}})
		if got > 1e-6 {
			t.Errorf("%s differs from the Voronoi wall distance by %g summed over the sample points", c.field, got)
		}
	}
}
