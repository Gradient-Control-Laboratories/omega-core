//go:build cgo

package eval

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/program"
)

// An exact round of radius r is the distance to an eroded or dilated sharp intersection, less r:
// for the intersection of A and B, the distance to E = {g1 <= 0} ∩ {g2 <= 0} with g = f + r; for
// their union, by De Morgan, with g = r - f. Outside E, near its edge Γ = {g1 = 0, g2 = 0}, that
// distance is |p - c| for c the point of Γ nearest p. These geometries give Γ and its normal
// cone, so a point p = c + s·u, u a unit vector in the cone at c, is at exactly s from E: the
// tee of two cylinders has no closed form, and needs none.
type spineGeometry struct {
	name   string
	g1, g2 string  // GSL fields whose zero sets meet in Γ
	r      float64 // the round's radius: its surface is at s = r
	// sample returns a point of Γ and the outward unit normals of E's two faces there.
	sample func(rng *rand.Rand) (c, n1, n2 [3]float64)
}

func spineGeometries() []spineGeometry {
	return []spineGeometry{
		{
			// A ball of radius 10 and the half-space z >= 0, intersected and rounded by 2: the rim
			// of a rounded hemisphere. Γ is the circle ρ = √60 at z = 2.
			name: "sphere ∩ plane, r 2",
			g1:   "length(p) - 8",
			g2:   "2 - p.z",
			r:    2,
			sample: func(rng *rand.Rand) (c, n1, n2 [3]float64) {
				a := rng.Float64() * 2 * math.Pi
				rho := math.Sqrt(60)
				c = [3]float64{rho * math.Cos(a), rho * math.Sin(a), 2}
				return c, scale(c, 1/8.0), [3]float64{0, 0, -1}
			},
		},
		{
			// fillet/ball-and-post: a ball of radius 10 and a post of radius 4, united and
			// rounded by 2. Γ is the circle ρ = 6 at z = 6√3, the round's center circle.
			name: "ball ∪ post, r 2",
			g1:   "12 - length(p)",
			g2:   "6 - length(p.xy)",
			r:    2,
			sample: func(rng *rand.Rand) (c, n1, n2 [3]float64) {
				a := rng.Float64() * 2 * math.Pi
				c = [3]float64{6 * math.Cos(a), 6 * math.Sin(a), 6 * math.Sqrt(3)}
				return c, scale(c, -1/12.0), [3]float64{-math.Cos(a), -math.Sin(a), 0}
			},
		},
		{
			// fillet/tee-of-two-cylinders: a bar of radius 6 on x and a stem of radius 3 on z,
			// united and rounded by 1. Γ is where the cylinders of radius 7 and 4 meet: a saddle.
			name: "tee of cylinders, r 1",
			g1:   "7 - length(p.yz)",
			g2:   "4 - length(p.xy)",
			r:    1,
			sample: func(rng *rand.Rand) (c, n1, n2 [3]float64) {
				a := rng.Float64() * 2 * math.Pi
				x, y := 4*math.Cos(a), 4*math.Sin(a)
				z := math.Sqrt(49 - y*y)
				c = [3]float64{x, y, z}
				return c, [3]float64{0, -y / 7, -z / 7}, [3]float64{-x / 4, -y / 4, 0}
			},
		},
	}
}

// seamSamples are points within 0.5 of a geometry's round, p = c + s·u, and their distance s to
// E. For the tee, whose Γ is no circle, it first checks the premise that no point of E is nearer.
func seamSamples(t *testing.T, g spineGeometry, n int, rng *rand.Rand) (xs, ys, zs []float32, want []float64) {
	t.Helper()
	xs, ys, zs, want = make([]float32, n), make([]float32, n), make([]float32, n), make([]float64, n)
	for i := range n {
		c, n1, n2 := g.sample(rng)
		w := rng.Float64()
		u := normalized(add(scale(n1, 1-w), scale(n2, w)))
		s := g.r - 0.5 + rng.Float64()
		pt := add(c, scale(u, s))
		xs[i], ys[i], zs[i] = float32(pt[0]), float32(pt[1]), float32(pt[2])
		want[i] = s
	}
	if strings.HasPrefix(g.name, "tee") {
		for i := range n {
			pt := [3]float64{float64(xs[i]), float64(ys[i]), float64(zs[i])}
			if d := teeDistanceToE(pt); d < want[i]-1e-5 {
				t.Fatalf("tee sample %v: a point of E is %v away, nearer than %v", pt, d, want[i])
			}
		}
	}
	return xs, ys, zs, want
}

func assembleRound(t *testing.T, src string) *program.Program {
	t.Helper()
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(source.NewLocalManager("../stdlib")))
	if err != nil {
		t.Fatalf("assemble: %v\n%s", err, src)
	}
	return res.Program
}

func evalRound(t *testing.T, prog *program.Program, xs, ys, zs []float32) []float64 {
	t.Helper()
	out, err := EvalGrad(prog, xs, ys, zs)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	vals := make([]float64, len(xs))
	for i := range xs {
		vals[i] = float64(out[0][i].Value)
	}
	return vals
}

func maxAbsError(got, want []float64) float64 {
	worst := 0.0
	for i := range got {
		e := math.Abs(got[i] - want[i])
		if math.IsNaN(e) {
			return math.Inf(1)
		}
		worst = math.Max(worst, e)
	}
	return worst
}

func intersectRoundExactSource(g spineGeometry, iterations int) string {
	return fmt.Sprintf("import . \"omega\";\ng1 = () { return %s; };\ng2 = () { return %s; };\nreturn intersectRoundExact(g1, g2, %d);\n",
		g.g1, g.g2, iterations)
}

// Near each round, the iterated form's distance to E is right to single precision in three steps
// where Γ is a circle and in five on the tee's saddle; the two-argument form's is off by up to 7%
// of the round's radius.
func TestIntersectRoundExactIterations_SeamDistance(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	steps := map[string]int{"sphere ∩ plane, r 2": 3, "ball ∪ post, r 2": 3, "tee of cylinders, r 1": 5}
	for _, g := range spineGeometries() {
		xs, ys, zs, want := seamSamples(t, g, 400, rng)
		today := maxAbsError(evalRound(t, assembleRound(t,
			fmt.Sprintf("import . \"omega\";\ng1 = () { return %s; };\ng2 = () { return %s; };\nreturn intersectRoundExact(g1, g2);\n", g.g1, g.g2)),
			xs, ys, zs), want)
		got := maxAbsError(evalRound(t, assembleRound(t, intersectRoundExactSource(g, steps[g.name])), xs, ys, zs), want)
		t.Logf("%s: %d steps %.2e, two-argument form %.2e", g.name, steps[g.name], got, today)
		if got > 2e-5 {
			t.Errorf("%s: %d steps, error %.2e, want under 2e-5", g.name, steps[g.name], got)
		}
	}
}

// One step's distance to the edge is the two-argument form's: wherever both take the edge's
// branch (their region tests differ, the iterated one judging at c), they agree.
func TestIntersectRoundExactIterations_OneStepIsTheTwoArgumentFormsDistance(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	for _, g := range spineGeometries() {
		xs, ys, zs, _ := seamSamples(t, g, 200, rng)
		two := evalRound(t, assembleRound(t, twoArgumentSource(g)), xs, ys, zs)
		one := evalRound(t, assembleRound(t, intersectRoundExactSource(g, 1)), xs, ys, zs)
		sharp := evalRound(t, assembleRound(t,
			fmt.Sprintf("import . \"omega\";\nreturn max(%s, %s);\n", g.g1, g.g2)), xs, ys, zs)
		compared, worst := 0, 0.0
		for i := range one {
			// Both on the edge's branch: neither is the sharp intersection's value.
			if math.Abs(one[i]-sharp[i]) > 1e-4 && math.Abs(two[i]-sharp[i]) > 1e-4 {
				compared++
				worst = math.Max(worst, math.Abs(one[i]-two[i]))
			}
		}
		t.Logf("%s: %d of %d points on both edge branches, worst difference %.2e", g.name, compared, len(one), worst)
		if compared < len(one)/2 || worst > 2e-5 {
			t.Errorf("%s: %d points compared, worst difference %.2e", g.name, compared, worst)
		}
	}
}

func twoArgumentSource(g spineGeometry) string {
	return fmt.Sprintf("import . \"omega\";\ng1 = () { return %s; };\ng2 = () { return %s; };\nreturn intersectRoundExact(g1, g2);\n", g.g1, g.g2)
}

// meridianDistance is the signed distance, in a meridian half-plane (ρ >= 0, z), to a region
// bounded by an arc of the circle of radius `radius` about the origin and a segment, given by
// the arc's and segment's nearest-point rules: nearestArc and nearestSegment return the nearest
// point of each piece to (ρ, z). inside says whether (ρ, z) is in the region.
func meridianDistance(rho, z float64, nearestArc, nearestSegment func(rho, z float64) (float64, float64), inside bool) float64 {
	ar, az := nearestArc(rho, z)
	sr, sz := nearestSegment(rho, z)
	d := math.Min(math.Hypot(rho-ar, z-az), math.Hypot(rho-sr, z-sz))
	if inside {
		return -d
	}
	return d
}

// sphereCapDistance is the exact signed distance to E = {|p| <= 8} ∩ {z >= 2}, the sphere ∩ plane
// geometry's eroded intersection: in the meridian, an arc of radius 8 above z = 2 and the chord
// at z = 2, meeting at (√60, 2).
func sphereCapDistance(p [3]float64) float64 {
	rho, z := math.Hypot(p[0], p[1]), p[2]
	edgeR := math.Sqrt(60)
	arc := func(rho, z float64) (float64, float64) {
		l := math.Hypot(rho, z)
		if l > 0 && 8*z/l >= 2 {
			return 8 * rho / l, 8 * z / l
		}
		return edgeR, 2
	}
	segment := func(rho, z float64) (float64, float64) {
		return math.Min(rho, edgeR), 2
	}
	inside := math.Hypot(rho, z) <= 8 && z >= 2
	return meridianDistance(rho, z, arc, segment, inside)
}

// ballPostSeamDistance is the exact signed distance to E = {|p| >= 12} ∩ {ρ >= 6}, the ball ∪ post
// geometry's region two outside both: in the meridian, the arc of radius 12 where ρ >= 6, and the
// line ρ = 6 where |z| >= 6√3, meeting at (6, ±6√3).
func ballPostSeamDistance(p [3]float64) float64 {
	rho, z := math.Hypot(p[0], p[1]), p[2]
	edgeZ := 6 * math.Sqrt(3)
	arc := func(rho, z float64) (float64, float64) {
		l := math.Hypot(rho, z)
		if l > 0 && 12*rho/l >= 6 {
			return 12 * rho / l, 12 * z / l
		}
		return 6, math.Copysign(edgeZ, z)
	}
	line := func(rho, z float64) (float64, float64) {
		if math.Abs(z) >= edgeZ {
			return 6, z
		}
		return 6, math.Copysign(edgeZ, z)
	}
	inside := math.Hypot(rho, z) >= 12 && rho >= 6
	return meridianDistance(rho, z, arc, line, inside)
}

// The whole field, not only near the edge: over a box around each round, inside E, outside it
// nearest a face, nearest the edge, and on the axis, where every point of the circle Γ is equally
// near. The rounded solid's field is this less r.
func TestIntersectRoundExactIterations_WholeField(t *testing.T) {
	cases := []struct {
		g      spineGeometry
		oracle func([3]float64) float64
		lo, hi [3]float64
	}{
		{spineGeometries()[0], sphereCapDistance, [3]float64{-11, -11, -3}, [3]float64{11, 11, 11}},
		// Out past the ball to ρ = 14 and down to z = -2, where a point is nearest a face but the
		// steps can reach the far side of the edge circle, another point where the distance along
		// the edge is stationary.
		{spineGeometries()[1], ballPostSeamDistance, [3]float64{-14, -14, -2}, [3]float64{14, 14, 15}},
	}
	rng := rand.New(rand.NewPCG(5, 6))
	for _, c := range cases {
		const n = 4000
		xs, ys, zs, want := make([]float32, n), make([]float32, n), make([]float32, n), make([]float64, n)
		for i := range n {
			var pt [3]float64
			for k := range 3 {
				pt[k] = c.lo[k] + rng.Float64()*(c.hi[k]-c.lo[k])
			}
			if i < 40 { // on the axis
				pt[0], pt[1] = 0, 0
			}
			xs[i], ys[i], zs[i] = float32(pt[0]), float32(pt[1]), float32(pt[2])
			want[i] = c.oracle([3]float64{float64(xs[i]), float64(ys[i]), float64(zs[i])})
		}
		today := evalRound(t, assembleRound(t, twoArgumentSource(c.g)), xs, ys, zs)
		t.Logf("%s, two-argument form: worst %.2e", c.g.name, maxAbsError(today, want))
		for _, iterations := range []int{1, 3, 5} {
			got := evalRound(t, assembleRound(t, intersectRoundExactSource(c.g, iterations)), xs, ys, zs)
			worst, at := 0.0, 0
			for i := range got {
				if e := math.Abs(got[i] - want[i]); e > worst || math.IsNaN(e) {
					worst, at = e, i
				}
			}
			t.Logf("%s, %d steps: worst %.2e at (%.3f, %.3f, %.3f): %.6f, want %.6f",
				c.g.name, iterations, worst, xs[at], ys[at], zs[at], got[at], want[at])
			if iterations >= 3 && !(worst <= 2e-5) {
				t.Errorf("%s, %d steps: worst error %.2e, want under 2e-5", c.g.name, iterations, worst)
			}
		}
	}
}

// The steps not taken are dropped at compile time: one step costs about what the two-argument
// form costs, not what eight do. A timing, so it logs and fails only on a gross miss.
func TestIntersectRoundExactIterations_UnusedStepsAreDropped(t *testing.T) {
	g := spineGeometries()[0]
	const n = 200000
	rng := rand.New(rand.NewPCG(7, 8))
	xs, ys, zs := make([]float32, n), make([]float32, n), make([]float32, n)
	for i := range n {
		xs[i], ys[i], zs[i] = float32(rng.Float64()*20-10), float32(rng.Float64()*20-10), float32(rng.Float64()*12-2)
	}
	best := func(src string) time.Duration {
		prog := assembleRound(t, src)
		fastest := time.Duration(math.MaxInt64)
		for range 3 {
			start := time.Now()
			if _, err := EvalGrad(prog, xs, ys, zs); err != nil {
				t.Fatalf("eval: %v", err)
			}
			fastest = min(fastest, time.Since(start))
		}
		return fastest
	}
	one, eight := best(intersectRoundExactSource(g, 1)), best(intersectRoundExactSource(g, 8))
	t.Logf("one step %v, eight steps %v, for %d points", one, eight, n)
	if eight < 2*one {
		t.Errorf("eight steps (%v) cost less than twice one (%v): the unused steps are not dropped", eight, one)
	}
}

// The two iterations the stdlib chose between, as GSL, each written once per field evaluation:
// one evaluation is both fields' values and gradients at one point.
//
// Gradient descent on (g1² + g2²)/2 from p, step 1 (the fields are distances, so their
// gradients are unit): c <- c - g1 ∇g1 - g2 ∇g2. It lands on Γ but not at the nearest point:
// nothing in it asks for p - c to be normal to Γ.
//
// Newton, the stdlib's: the least-norm step onto both zero sets, c <- c - Jᵀ(JJᵀ)⁻¹g, then a slide
// along Γ's tangent t = ∇g1 × ∇g2 to the foot of p. The last evaluation projects without sliding.
const iterationGSL = `
descend = (c) {
  v1 = remap(g1(), c); v2 = remap(g2(), c);
  n1 = remap(g1().dp, c); n2 = remap(g2().dp, c);
  return c - v1 * n1 - v2 * n2;
};
project = (c) {
  v1 = remap(g1(), c); v2 = remap(g2(), c);
  n1 = remap(g1().dp, c); n2 = remap(g2().dp, c);
  a = dot(n1, n1); b = dot(n1, n2); d = dot(n2, n2);
  det = max(a * d - b * b, 0.000000000001);
  return c - ((d * v1 - b * v2) * n1 + (a * v2 - b * v1) * n2) / det;
};
newton = (c) {
  v1 = remap(g1(), c); v2 = remap(g2(), c);
  n1 = remap(g1().dp, c); n2 = remap(g2().dp, c);
  a = dot(n1, n1); b = dot(n1, n2); d = dot(n2, n2);
  det = max(a * d - b * b, 0.000000000001);
  q = c - ((d * v1 - b * v2) * n1 + (a * v2 - b * v1) * n2) / det;
  t = cross(n1, n2);
  return q + dot(p - q, t) / max(dot(t, t), 0.000000000001) * t;
};
`

func methodSource(g spineGeometry, method string, evals int) string {
	var b strings.Builder
	b.WriteString("import . \"omega\";\n")
	fmt.Fprintf(&b, "g1 = () { return %s; };\ng2 = () { return %s; };\n", g.g1, g.g2)
	b.WriteString(iterationGSL)
	b.WriteString("c0 = p;\n")
	for i := 1; i <= evals; i++ {
		step := method
		if method == "newton" && i == evals {
			step = "project"
		}
		fmt.Fprintf(&b, "c%d = %s(c%d);\n", i, step, i-1)
	}
	fmt.Fprintf(&b, "return length(p - c%d);\n", evals)
	return b.String()
}

// TestExactRoundMethods logs, for each geometry, the error near the round after 1 to 5 field
// evaluations of gradient descent and of Newton: the measurement the stdlib's choice of Newton
// rests on. It fails only if a source does not evaluate.
func TestExactRoundMethods(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for _, g := range spineGeometries() {
		xs, ys, zs, want := seamSamples(t, g, 400, rng)
		t.Logf("%s: max |error| after 1..5 evaluations", g.name)
		for _, m := range []string{"descend", "newton"} {
			var row []string
			for n := 1; n <= 5; n++ {
				row = append(row, fmt.Sprintf("%.1e", maxAbsError(evalRound(t, assembleRound(t, methodSource(g, m, n)), xs, ys, zs), want)))
			}
			t.Logf("  %-8s %s", m, strings.Join(row, "  "))
		}
	}
}

// teeDistanceToE is the distance from p, inside both offset cylinders, to E = {y² + z² >= 49} ∩
// {x² + y² >= 16} near the top of the tee, in float64: the nearer of Γ (searched densely, then
// refined) and each face's nearest point where that point lies in E.
func teeDistanceToE(p [3]float64) float64 {
	onGamma := func(a float64) [3]float64 {
		x, y := 4*math.Cos(a), 4*math.Sin(a)
		return [3]float64{x, y, math.Sqrt(49 - y*y)}
	}
	dist := func(q [3]float64) float64 {
		return math.Sqrt((p[0]-q[0])*(p[0]-q[0]) + (p[1]-q[1])*(p[1]-q[1]) + (p[2]-q[2])*(p[2]-q[2]))
	}
	best, bestA := math.Inf(1), 0.0
	const n = 20000
	for i := range n {
		a := 2 * math.Pi * float64(i) / n
		if d := dist(onGamma(a)); d < best {
			best, bestA = d, a
		}
	}
	lo, hi := bestA-2*math.Pi/n, bestA+2*math.Pi/n
	for range 100 {
		m1, m2 := lo+(hi-lo)/3, hi-(hi-lo)/3
		if dist(onGamma(m1)) < dist(onGamma(m2)) {
			hi = m2
		} else {
			lo = m1
		}
	}
	best = math.Min(best, dist(onGamma((lo+hi)/2)))
	// The bar's face: p pushed out to radius 7 about x, in E if outside the stem's cylinder of 4.
	if ryz := math.Hypot(p[1], p[2]); ryz > 0 {
		q := [3]float64{p[0], 7 * p[1] / ryz, 7 * p[2] / ryz}
		if q[0]*q[0]+q[1]*q[1] >= 16 {
			best = math.Min(best, dist(q))
		}
	}
	// The stem's face: p pushed out to radius 4 about z, in E if outside the bar's cylinder of 7.
	if rxy := math.Hypot(p[0], p[1]); rxy > 0 {
		q := [3]float64{4 * p[0] / rxy, 4 * p[1] / rxy, p[2]}
		if q[1]*q[1]+q[2]*q[2] >= 49 {
			best = math.Min(best, dist(q))
		}
	}
	return best
}

func add(a, b [3]float64) [3]float64 { return [3]float64{a[0] + b[0], a[1] + b[1], a[2] + b[2]} }

func scale(a [3]float64, s float64) [3]float64 { return [3]float64{a[0] * s, a[1] * s, a[2] * s} }

func normalized(a [3]float64) [3]float64 {
	return scale(a, 1/math.Sqrt(a[0]*a[0]+a[1]*a[1]+a[2]*a[2]))
}
