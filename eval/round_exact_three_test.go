package eval

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
	"testing"
)

// The three-field exact round is the opening of the intersection by a ball of the radius: the
// fields shrunk by the radius, the exact distance to their intersection, grown back. These tests
// compare it with that construction computed independently, by Dykstra's alternating projection
// onto the shrunk sets, which converges to the nearest point of an intersection of convex sets.

// convexSet is one shrunk set: its exact signed distance and its nearest-point projection.
type convexSet struct {
	gsl     string // the field, before shrinking, as GSL
	dist    func(p [3]float64, r float64) float64
	project func(p [3]float64, r float64) [3]float64
}

func halfSpace(n [3]float64, h float64) convexSet {
	n = normalized(n)
	return convexSet{
		gsl: fmt.Sprintf("dot(p, [%.17g, %.17g, %.17g]) - %.17g", n[0], n[1], n[2], h),
		dist: func(p [3]float64, r float64) float64 {
			return n[0]*p[0] + n[1]*p[1] + n[2]*p[2] - (h - r)
		},
		project: func(p [3]float64, r float64) [3]float64 {
			s := n[0]*p[0] + n[1]*p[1] + n[2]*p[2] - (h - r)
			return add(p, scale(n, -math.Max(s, 0)))
		},
	}
}

func ball(radius float64) convexSet {
	return convexSet{
		gsl:  fmt.Sprintf("length(p) - %.17g", radius),
		dist: func(p [3]float64, r float64) float64 { return math.Sqrt(p[0]*p[0]+p[1]*p[1]+p[2]*p[2]) - (radius - r) },
		project: func(p [3]float64, r float64) [3]float64 {
			l := math.Sqrt(p[0]*p[0] + p[1]*p[1] + p[2]*p[2])
			if l <= radius-r {
				return p
			}
			return scale(p, (radius-r)/l)
		},
	}
}

// openedDistance is the opening's signed field at p: inside the shrunk intersection the largest
// shrunk field (exact there: the largest ball about p inside it is the smallest one inside each),
// outside the distance to Dykstra's nearest point; then less the radius.
func openedDistance(sets []convexSet, r float64, p [3]float64) float64 {
	inside := math.Inf(-1)
	for _, s := range sets {
		inside = math.Max(inside, s.dist(p, r))
	}
	if inside <= 0 {
		return inside - r
	}
	x := p
	inc := make([][3]float64, len(sets))
	for range 20000 {
		for k, s := range sets {
			y := add(x, inc[k])
			z := s.project(y, r)
			inc[k] = add(y, scale(z, -1))
			x = z
		}
	}
	d := add(p, scale(x, -1))
	return math.Sqrt(d[0]*d[0]+d[1]*d[1]+d[2]*d[2]) - r
}

type threeFieldCase struct {
	name   string
	sets   []convexSet
	r      float64
	center [3]float64 // the sharp corner, around which points are sampled
	reach  float64
	tol    float64
}

func pyramidPlanes() []convexSet {
	// The pyramid of fillet/rectangular-pyramid-all-edges: the rectangle (±10, ±6, 0), apex (0, 0, 12).
	apex := [3]float64{0, 0, 12}
	base := [][3]float64{{10, 6, 0}, {-10, 6, 0}, {-10, -6, 0}, {10, -6, 0}}
	sets := []convexSet{halfSpace([3]float64{0, 0, -1}, 0)}
	for i := range base {
		a, b := base[i], base[(i+1)%len(base)]
		u, v := add(b, scale(a, -1)), add(apex, scale(a, -1))
		n := normalized([3]float64{u[1]*v[2] - u[2]*v[1], u[2]*v[0] - u[0]*v[2], u[0]*v[1] - u[1]*v[0]})
		if n[0]*a[0]+n[1]*a[1]+n[2]*a[2] < 0 {
			n = scale(n, -1)
		}
		sets = append(sets, halfSpace(n, n[0]*a[0]+n[1]*a[1]+n[2]*a[2]))
	}
	return sets
}

func threeFieldCases() []threeFieldCase {
	pyr := pyramidPlanes()
	return []threeFieldCase{
		{
			name:   "box corner, square faces, r 2",
			sets:   []convexSet{halfSpace([3]float64{1, 0, 0}, 10), halfSpace([3]float64{0, 1, 0}, 10), halfSpace([3]float64{0, 0, 1}, 10)},
			r:      2,
			center: [3]float64{10, 10, 10},
			reach:  4,
			tol:    2e-5,
		},
		{
			// The base corner (10, 6, 0): the base and the two sloped faces through it.
			name:   "pyramid base corner, r 1.5",
			sets:   []convexSet{pyr[0], pyr[1], pyr[4]},
			r:      1.5,
			center: [3]float64{10, 6, 0},
			reach:  4,
			tol:    2e-5,
		},
		{
			// A ball cut by two planes at right angles to each other but not to the sphere.
			name:   "ball cut by two planes, r 2",
			sets:   []convexSet{ball(10), halfSpace([3]float64{0, 0, -1}, 0), halfSpace([3]float64{1, 0, 0}, 6)},
			r:      2,
			center: [3]float64{6, 8, 0},
			reach:  3,
			tol:    1e-4,
		},
	}
}

func threeFieldSource(c threeFieldCase, form string) string {
	var b strings.Builder
	b.WriteString("import . \"omega\";\n")
	for i, s := range c.sets {
		fmt.Fprintf(&b, "f%d = %s;\n", i+1, s.gsl)
	}
	fmt.Fprintf(&b, "return %s;\n", fmt.Sprintf(form, c.r))
	return b.String()
}

func threeFieldSamples(c threeFieldCase, n int, rng *rand.Rand) (xs, ys, zs []float32, want []float64) {
	for range n {
		p := [3]float64{
			c.center[0] + c.reach*(2*rng.Float64()-1),
			c.center[1] + c.reach*(2*rng.Float64()-1),
			c.center[2] + c.reach*(2*rng.Float64()-1),
		}
		xs, ys, zs = append(xs, float32(p[0])), append(ys, float32(p[1])), append(zs, float32(p[2]))
		want = append(want, openedDistance(c.sets, c.r, p))
	}
	return xs, ys, zs, want
}

// Around a corner of three faces, every point: inside, nearest a face, an edge's round or the
// corner's ball.
func TestIntersectRoundExactThreeFields_Opening(t *testing.T) {
	rng := rand.New(rand.NewPCG(7, 8))
	for _, c := range threeFieldCases() {
		xs, ys, zs, want := threeFieldSamples(c, 400, rng)
		got := evalRound(t, assembleRound(t, threeFieldSource(c, "intersectRoundExact(f1, f2, f3, %.17g)")), xs, ys, zs)
		worst := maxAbsError(got, want)
		t.Logf("%s: worst error %.2e over %d points", c.name, worst, len(xs))
		if worst > c.tol {
			t.Errorf("%s: worst error %.2e, want under %.0e", c.name, worst, c.tol)
		}
	}
}

// Where the faces are planes, one step is exact.
func TestIntersectRoundExactThreeFields_PlanesInOneStep(t *testing.T) {
	rng := rand.New(rand.NewPCG(9, 10))
	c := threeFieldCases()[1]
	xs, ys, zs, want := threeFieldSamples(c, 400, rng)
	got := evalRound(t, assembleRound(t, threeFieldSource(c, "intersectRoundExactSteps(f1, f2, f3, %.17g, 1)")), xs, ys, zs)
	if worst := maxAbsError(got, want); worst > c.tol {
		t.Errorf("%s in one step: worst error %.2e, want under %.0e", c.name, worst, c.tol)
	}
}

// The union is the intersection's De Morgan dual: the round of the union of three solids is the
// complement's opening, so the fields are each other's negatives.
func TestUnionRoundExactThreeFields_IsTheDual(t *testing.T) {
	rng := rand.New(rand.NewPCG(11, 12))
	c := threeFieldCases()[2]
	xs, ys, zs, _ := threeFieldSamples(c, 200, rng)
	inter := evalRound(t, assembleRound(t, threeFieldSource(c, "intersectRoundExact(f1, f2, f3, %.17g)")), xs, ys, zs)
	union := evalRound(t, assembleRound(t, threeFieldSource(c, "neg(unionRoundExact(neg(f1), neg(f2), neg(f3), %.17g))")), xs, ys, zs)
	if worst := maxAbsError(union, inter); worst > 1e-6 {
		t.Errorf("union is not the intersection's dual: worst difference %.2e", worst)
	}
}

// Three faces whose normals are coplanar make no corner (three faces of a triangular prism): the
// corner's term must drop out and leave the edges' rounds, the prism's exact opening.
func TestIntersectRoundExactThreeFields_NoCornerWhereNormalsAreCoplanar(t *testing.T) {
	rng := rand.New(rand.NewPCG(13, 14))
	c := threeFieldCase{
		name: "triangular prism along z, r 1",
		sets: []convexSet{
			halfSpace([3]float64{0, -1, 0}, 0),
			halfSpace([3]float64{math.Sqrt(3), 1, 0}, 10*math.Sqrt(3)/2),
			halfSpace([3]float64{-math.Sqrt(3), 1, 0}, 10*math.Sqrt(3)/2),
		},
		r:      1,
		center: [3]float64{0, 3, 0},
		reach:  12,
		tol:    2e-5,
	}
	xs, ys, zs, want := threeFieldSamples(c, 400, rng)
	got := evalRound(t, assembleRound(t, threeFieldSource(c, "intersectRoundExact(f1, f2, f3, %.17g)")), xs, ys, zs)
	worst := maxAbsError(got, want)
	t.Logf("%s: worst error %.2e", c.name, worst)
	if worst > c.tol {
		t.Errorf("%s: worst error %.2e, want under %.0e", c.name, worst, c.tol)
	}
}

func octahedronPlanes() []convexSet {
	// The octahedron |x| + |y| + |z| <= 10 of fillet/octahedron-all-edges.
	var sets []convexSet
	for _, sx := range []float64{1, -1} {
		for _, sy := range []float64{1, -1} {
			for _, sz := range []float64{1, -1} {
				sets = append(sets, halfSpace([3]float64{sx, sy, sz}, 10/math.Sqrt(3)))
			}
		}
	}
	return sets
}

func listSource(sets []convexSet, form string) string {
	var b strings.Builder
	b.WriteString("import . \"omega\";\n")
	var names []string
	for i, s := range sets {
		fmt.Fprintf(&b, "f%d = %s;\n", i+1, s.gsl)
		names = append(names, fmt.Sprintf("f%d", i+1))
	}
	fmt.Fprintf(&b, "fs = [%s];\nresult = %s;\nreturn result;\n", strings.Join(names, ", "), form)
	return b.String()
}

// The list form, with any number of faces to a corner: the pyramid's apex, where four faces meet
// and no ball touches all four, so the corner splits into a short round between two patches; and
// six of the octahedron's faces, among them two pairs of parallel ones, which have no edge.
func TestIntersectRoundExactList_Opening(t *testing.T) {
	rng := rand.New(rand.NewPCG(15, 16))
	for _, c := range []threeFieldCase{
		{name: "pyramid, 5 faces, near the apex", sets: pyramidPlanes(), r: 1.5, center: [3]float64{0, 0, 11}, reach: 4, tol: 2e-5},
		{name: "pyramid, 5 faces, whole", sets: pyramidPlanes(), r: 1.5, center: [3]float64{0, 0, 6}, reach: 12, tol: 2e-5},
		{name: "octahedron, 6 of its faces", sets: octahedronPlanes()[:6], r: 2, center: [3]float64{9, 0, 0}, reach: 4, tol: 2e-5},
	} {
		xs, ys, zs, want := threeFieldSamples(c, 400, rng)
		got := evalRound(t, assembleRound(t, listSource(c.sets, fmt.Sprintf("intersectRoundExact(fs, %.17g)", c.r))), xs, ys, zs)
		worst := maxAbsError(got, want)
		t.Logf("%s: worst error %.2e over %d points", c.name, worst, len(xs))
		if worst > c.tol {
			t.Errorf("%s: worst error %.2e, want under %.0e", c.name, worst, c.tol)
		}
	}
}

// For three fields the list form, whose gradients are differences, is the three-field form,
// whose gradients are .dp's.
func TestIntersectRoundExactList_IsTheThreeFieldForm(t *testing.T) {
	rng := rand.New(rand.NewPCG(17, 18))
	for _, c := range threeFieldCases() {
		xs, ys, zs, _ := threeFieldSamples(c, 200, rng)
		three := evalRound(t, assembleRound(t, threeFieldSource(c, "intersectRoundExact(f1, f2, f3, %.17g)")), xs, ys, zs)
		list := evalRound(t, assembleRound(t, listSource(c.sets, fmt.Sprintf("intersectRoundExact(fs, %.17g)", c.r))), xs, ys, zs)
		if worst := maxAbsError(list, three); worst > 1e-5 {
			t.Errorf("%s: list and three-field forms differ by %.2e", c.name, worst)
		}
	}
}

// The whole octahedron by its symmetry: with q = abs(p) the point of it nearest p is nearest q in
// the first octant, so the distance to the octahedron shrunk by the radius is the list distance to
// its face there and the three mirror planes, which are no faces and are not shrunk; inside, it is
// the face's alone.
func TestIntersectDistanceList_FoldedOctahedron(t *testing.T) {
	rng := rand.New(rand.NewPCG(19, 20))
	r := 2.0
	face := fmt.Sprintf("(p.x + p.y + p.z - %.17g) / sqrt(3)", 10-r*math.Sqrt(3))
	src := fmt.Sprintf("import . \"omega\";\nresult = remap(%[1]s <= 0 ? %[1]s : intersectDistance([%[1]s, neg(p.x), neg(p.y), neg(p.z)], 4), abs(p)) - %.17[2]g;\nreturn result;\n", face, r)
	for _, c := range []threeFieldCase{
		{name: "near a corner", sets: octahedronPlanes(), r: r, center: [3]float64{9, 0, 0}, reach: 4, tol: 2e-5},
		{name: "whole", sets: octahedronPlanes(), r: r, center: [3]float64{0, 0, 0}, reach: 12, tol: 2e-5},
	} {
		xs, ys, zs, want := threeFieldSamples(c, 400, rng)
		got := evalRound(t, assembleRound(t, src), xs, ys, zs)
		worst := maxAbsError(got, want)
		t.Logf("folded octahedron, %s: worst error %.2e", c.name, worst)
		if worst > c.tol {
			t.Errorf("folded octahedron, %s: worst error %.2e, want under %.0e", c.name, worst, c.tol)
		}
	}
}
