//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// unboundedAxis is what an axis .box cannot bound reads as: builtin's nothing.
const unboundedAxis = 1e5

// boxEval evaluates each expression's .box in the document and returns
// [lo.x, lo.y, lo.z, hi.x, hi.y, hi.z] for each. A box does not depend on p, so
// the point is only there because the evaluator wants one.
func boxEval(t *testing.T, e *Evaluator, code string, exprs []string) [][6]float64 {
	t.Helper()
	queries := make([]string, len(exprs))
	for i, x := range exprs {
		queries[i] = "(" + x + ").box"
	}
	groups, err := e.EvalGsl(context.Background(), nil, code, "", queries, [][3]float64{{0, 0, 0}})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	out := make([][6]float64, len(exprs))
	for i := range exprs {
		bx, ok := boxOf(groups[i].Samples)
		if !ok {
			t.Fatalf("%s.box: %d components (%s), want 6", exprs[i], len(groups[i].Samples), groups[i].TypeName)
		}
		out[i] = bx
	}
	return out
}

// boxOf reads a [lo, hi] result at the first point. Samples are indexed by
// component, then by point.
func boxOf(samples [][]GradSample) ([6]float64, bool) {
	var bx [6]float64
	if len(samples) != 6 {
		return bx, false
	}
	for k := range bx {
		bx[k] = float64(samples[k][0].Value)
	}
	return bx, true
}

func newBoxEvaluator() *Evaluator {
	return New(source.NewLocalManager("../stdlib"), nil)
}

// TestBoundingBox checks each .box rule end to end, parse through fidget.
func TestBoundingBox(t *testing.T) {
	const u = unboundedAxis
	code := `
a = sphere([1, 2, 3], 0.5);
b = box([0, 0, 0], [1, 2, 3]);
r = 0.5;
r.min = 0.1;
r.max = 2.0;
mySphere(c, rad) { return length(p - c) - rad; }
return a;
`
	cases := []struct {
		expr string
		want [6]float64
	}{
		// primitives
		{"a", [6]float64{0.5, 1.5, 2.5, 1.5, 2.5, 3.5}},
		{"b", [6]float64{0, 0, 0, 1, 2, 3}},
		{"box(-1, 2)", [6]float64{-1, -1, -1, 2, 2, 2}},
		{"sphere(2)", [6]float64{-2, -2, -2, 2, 2, 2}},
		{"boxCentered([1, 1, 1], [2, 4, 6])", [6]float64{0, -1, -2, 2, 3, 4}},
		{"torus(2, 0.5)", [6]float64{-2.5, -2.5, -0.5, 2.5, 2.5, 0.5}},
		{"cylinderRounded([0, 0, 0], [0, 0, 4], 1)", [6]float64{-1, -1, -1, 1, 1, 5}},
		{"cylinderCapped([0, 0, 0], [0, 4, 0], 1)", [6]float64{-1, -1, -1, 1, 5, 1}},
		{"coneCapped([0, 0, 0], [0, 0, 2], 1, 0.5)", [6]float64{-1, -1, -1, 1, 1, 3}},
		{"lineSegment([0, 0, 0], [3, 0, 0]) - 0.5", [6]float64{-0.5, -0.5, -0.5, 3.5, 0.5, 0.5}},
		// an infinite cylinder is unbounded only along its axis
		{"cylinderZ([1, 2, 0], 0.5)", [6]float64{0.5, 1.5, -u, 1.5, 2.5, u}},
		// the octahedron reads 1/sqrt(3) of the distance past a vertex
		{"octahedron(1) - 0.1", [6]float64{-1.1732051, -1.1732051, -1.1732051, 1.1732051, 1.1732051, 1.1732051}},

		// combinations
		{"min(a, b)", [6]float64{0, 0, 0, 1.5, 2.5, 3.5}},
		{"max(a, b)", [6]float64{0.5, 1.5, 2.5, 1, 2, 3}},
		{"unionSharp([a, b])", [6]float64{0, 0, 0, 1.5, 2.5, 3.5}},
		{"differenceSharp(b, a)", [6]float64{0, 0, 0, 1, 2, 3}},
		{"differenceRound(b, a, 0.2)", [6]float64{0, 0, 0, 1, 2, 3}},
		{"intersectRound(a, b, 0.2)", [6]float64{0.5, 1.5, 2.5, 1, 2, 3}},
		{"unionRound(a, b, 0.3)", [6]float64{-0.3, -0.3, -0.3, 1.8, 2.8, 3.8}},
		{"(p.x > 0) ? a : b", [6]float64{0, 0, 0, 1.5, 2.5, 3.5}},
		{"min(a, nothing)", [6]float64{0.5, 1.5, 2.5, 1.5, 2.5, 3.5}},

		// offsets, scale and shells
		{"min(a, b) - 0.25", [6]float64{-0.25, -0.25, -0.25, 1.75, 2.75, 3.75}},
		{"a + 0.25", [6]float64{0.5, 1.5, 2.5, 1.5, 2.5, 3.5}},
		{"0.5 * a - 1", [6]float64{-1.5, -0.5, 0.5, 3.5, 4.5, 5.5}},
		{"a / 2 - 1", [6]float64{-1.5, -0.5, 0.5, 3.5, 4.5, 5.5}},
		{"abs(a) - 0.1", [6]float64{0.4, 1.4, 2.4, 1.6, 2.6, 3.6}},

		// a box that depends on a slider reads at the slider's value; a slider
		// used as a field is not the positive constant it starts as
		{"sphere(r)", [6]float64{-0.5, -0.5, -0.5, 0.5, 0.5, 0.5}},
		{"min(a, r)", [6]float64{-u, -u, -u, u, u, u}},
		{"max(a, r)", [6]float64{0.5, 1.5, 2.5, 1.5, 2.5, 3.5}},

		// a box read back as geometry
		{"box(a.box[0], a.box[1])", [6]float64{0.5, 1.5, 2.5, 1.5, 2.5, 3.5}},

		// transforms carry the box through the remap inside them; b spans
		// [0, 1] x [0, 2] x [0, 3], and rotations are counterclockwise
		{"transform.translate(b, [1, 2, 3])", [6]float64{1, 2, 3, 2, 4, 6}},
		{"transform.rotateZ(b, pi / 2)", [6]float64{-2, 0, 0, 0, 1, 3}},
		{"transform.rotateX(b, pi / 2)", [6]float64{0, -3, 0, 1, 0, 2}},
		{"transform.rotateY(b, pi / 2)", [6]float64{0, 0, -1, 3, 2, 0}},
		{"transform.rotate(b, [0, 0, 0], [0, 0, 1], pi / 2)", [6]float64{-2, 0, 0, 0, 1, 3}},
		{"transform.rotate(b, [1, 0, 0], [0, 0, 2], pi)", [6]float64{1, -2, 0, 2, 0, 3}},
		{"transform.scale(b, 2)", [6]float64{0, 0, 0, 2, 4, 6}},
		{"transform.affine(b, matrix.translation([1, 2, 3]))", [6]float64{1, 2, 3, 2, 4, 6}},
		{"transform.affine(b, matrix.scaling([2, 1, 3]))", [6]float64{0, 0, 0, 2, 2, 9}},
		{"transform.translate(sphere(r), [1, 0, 0])", [6]float64{0.5, -0.5, -0.5, 1.5, 0.5, 0.5}},
		// a turned box's hull grows by up to r(|cos| + |sin|) under an offset
		{"transform.rotateZ(b, pi / 4) - 0.1", [6]float64{-1.5556349, -0.1414214, -0.1414214, 0.8485281, 2.2627417, 3.1414214}},
		// an infinite cylinder turned about its own axis stays unbounded along it
		{"transform.rotateZ(cylinderZ([1, 2, 0], 0.5), pi / 2)", [6]float64{-2.5, 0.5, -u, -1.5, 1.5, u}},

		// a complement has its side's surface, though its solid is unbounded,
		// so it bounds a union with it, and a difference is its target
		{"-a", [6]float64{0.5, 1.5, 2.5, 1.5, 2.5, 3.5}},
		{"-2 * a", [6]float64{0.5, 1.5, 2.5, 1.5, 2.5, 3.5}},
		{"min(-a, b)", [6]float64{0, 0, 0, 1.5, 2.5, 3.5}},
		{"max(b, -a)", [6]float64{0, 0, 0, 1, 2, 3}},

		// a user's own function is read by its body
		{"mySphere([0, 0, 0], 1)", [6]float64{-1, -1, -1, 1, 1, 1}},

		// unbounded: a remap, an offset past the positive constant that
		// emptied one side, an offset complement, and a ternary with a
		// complement, which steps from inside -a to outside b across all of
		// x = 0
		{"remap(a, p - [1, 0, 0])", [6]float64{-u, -u, -u, u, u, u}},
		{"transform.translate(remap(a, p * 2), [1, 0, 0])", [6]float64{-u, -u, -u, u, u, u}},
		{"min(a, 5) - 10", [6]float64{-u, -u, -u, u, u, u}},
		{"-a - 0.1", [6]float64{-u, -u, -u, u, u, u}},
		{"(p.x > 0) ? -a : b", [6]float64{-u, -u, -u, u, u, u}},
	}
	e := newBoxEvaluator()
	exprs := make([]string, len(cases))
	for i, c := range cases {
		exprs[i] = c.expr
	}
	got := boxEval(t, e, code, exprs)
	for i, c := range cases {
		for k := 0; k < 6; k++ {
			if d := got[i][k] - c.want[k]; math.Abs(d) > 2e-5*math.Max(1, math.Abs(c.want[k])) {
				t.Errorf("%s.box = %v, want %v", c.expr, got[i], c.want)
				break
			}
		}
	}
}

// TestBoundingBoxTransformsSound proves, by interval arithmetic, that each
// transformed shape has no solid outside its box: the transform rules are
// hand-derived, and this is what keeps them honest.
func TestBoundingBoxTransformsSound(t *testing.T) {
	code := `
b = box([0, 0, 0], [2, 1, 1]);
c = cylinderRounded([0, 0, 0], [0, 0, 3], 0.5);
m = matrix.compose([1, -2, 0.5], [0.3, -0.7, 1.1], [1.5, 0.5, 2]);
return b;
`
	exprs := []string{
		"transform.translate(b, [1, 2, 3])",
		"transform.rotateX(b, 0.4)",
		"transform.rotateY(c, -1.3)",
		"transform.rotateZ(unionRound(b, c, 0.2), 2.5)",
		"transform.rotate(c, [1, 1, 0], [1, 2, 3], 0.9)",
		"transform.scale(c, 0.7) - 0.2",
		"transform.affine(b, m)",
		"transform.affine(c, m) - 0.3",
		"transform.rotateZ(b, 0.6) - 0.25",
	}
	e := newBoxEvaluator()
	e.SetCertifyBudget(200000)
	boxes := boxEval(t, e, code, exprs)
	for i, x := range exprs {
		if !anyBounded(boxes[i]) {
			t.Errorf("%s: no box", x)
			continue
		}
		// A transformed box's hull touches the solid only at a vertex, where a
		// field scaled down by the matrix rises slowly; prove it a little
		// way out.
		for _, slab := range outsideSlabsBy(boxes[i], 0.05) {
			res, err := e.Certify(context.Background(), nil, code, "", x, slab, 0, 0)
			if err != nil {
				t.Fatalf("%s: certify: %v", x, err)
			}
			if res.Disproven {
				t.Errorf("%s has solid outside its box %v, in %v", x, boxes[i], res.CounterBox)
				break
			}
			if res.Indeterminate {
				t.Errorf("%s: could not prove its box %v in %v (%s)", x, boxes[i], slab, res.Reason)
				break
			}
		}
	}
}

// TestBoundingBoxSurfacesSound proves, by interval arithmetic, that the values
// whose surface box is tighter than their solid's, the complements and what is
// built from them, have no surface outside it.
func TestBoundingBoxSurfacesSound(t *testing.T) {
	code := `
a = sphere([1, 2, 3], 0.5);
b = box([0, 0, 0], [1, 2, 3]);
return b;
`
	exprs := []string{
		"-a",
		"-2 * a",
		"min(-a, b)",
		"max(b, -a)",
		"abs(-a)",
		"transform.translate(-a, [1, 0, 0])",
		"transform.rotateZ(-b, 0.7)",
	}
	e := newBoxEvaluator()
	e.SetCertifyBudget(200000)
	boxes := boxEval(t, e, code, exprs)
	for i, x := range exprs {
		if !anyBounded(boxes[i]) {
			t.Errorf("%s: no box", x)
			continue
		}
		// A turned box's hull touches the surface only at a vertex; prove it a
		// little way out, as for the transforms.
		for _, slab := range outsideSlabsBy(boxes[i], 0.05) {
			res, err := certifyOneSign(context.Background(), e, code, x, slab)
			if err != nil {
				t.Fatalf("%s: certify: %v", x, err)
			}
			if res.Disproven {
				t.Errorf("%s has surface outside its box %v, in %v", x, boxes[i], res.CounterBox)
				break
			}
			if res.Indeterminate {
				t.Errorf("%s: could not prove its box %v in %v (%s)", x, boxes[i], slab, res.Reason)
				break
			}
		}
	}
}

// TestBoundingBoxRawForms checks fields written out in p's coordinates, the way
// library parts are: planes, slabs, and distances to a point or an axis, and a
// fastener's head and shank built from them. Each box is exact, and interval
// arithmetic proves no surface lies outside it.
//
// The head and shank are fastener's cylindricalHead and plainShank as they were
// before the library declared their boxes, so that their boxes are still read
// from their fields.
func TestBoundingBoxRawForms(t *testing.T) {
	const u = unboundedAxis
	code := `
head(dk, k) {
  return max(length(p.xy) - dk / 2.0, -p.z, p.z - k,
             ((length(p.xy) - dk / 2.0) + (p.z - k) + 0.2) * 0.7071);
}
shank(dia, shankLength) {
  return max(length(p.xy) - dia / 2.0, -shankLength - p.z, p.z,
             ((length(p.xy) - dia / 2.0)
              + (-p.z - shankLength) + 0.15 * dia) * 0.7071);
}
return sphere(1);
`
	cases := []struct {
		expr string
		want [6]float64
	}{
		// a plane's surface is the plane; its solid, a half-space, is wider
		{"p.z - 2", [6]float64{-u, -u, 2, u, u, 2}},
		{"-p.z", [6]float64{-u, -u, 0, u, u, 0}},
		{"2 * p.x - 4", [6]float64{2, -u, -u, 2, u, u}},
		{"(1 - p.y) / 2", [6]float64{-u, 1, -u, u, 1, u}},
		// a slab, a ball and cylinders along an axis
		{"abs(p.z) - 1", [6]float64{-u, -u, -1, u, u, 1}},
		{"abs(p.z - 1) - 1", [6]float64{-u, -u, 0, u, u, 2}},
		{"length(p) - 2", [6]float64{-2, -2, -2, 2, 2, 2}},
		{"length(p - [1, 2, 3]) - 0.5", [6]float64{0.5, 1.5, 2.5, 1.5, 2.5, 3.5}},
		{"length(p.xy) - 2", [6]float64{-2, -2, -u, 2, 2, u}},
		{"length(p.xz - [1, 0]) - 1", [6]float64{0, -u, -1, 2, u, 1}},
		{"length([p.y, p.z]) - 1", [6]float64{-u, -1, -1, u, 1, 1}},
		// cut by planes into a solid
		{"max(length(p.xy) - 1, -p.z, p.z - 3)", [6]float64{-1, -1, 0, 1, 1, 3}},
		{"max(length(p.xy) - 1, abs(p.z - 1) - 1)", [6]float64{-1, -1, 0, 1, 1, 2}},
		// a fastener's head, shank and the screw made of them
		{"head(10, 6)", [6]float64{-5, -5, 0, 5, 5, 6}},
		{"shank(6, 18)", [6]float64{-3, -3, -18, 3, 3, 0}},
		{"fastener.screw(head(10, 6), shank(6, 18))", [6]float64{-5, -5, -18, 5, 5, 6}},
		// a remap still clears it, and a vector that repeats an axis bounds
		// nothing
		{"remap(length(p) - 1, p - [1, 0, 0])", [6]float64{-u, -u, -u, u, u, u}},
		{"length([p.x, p.x]) - 1", [6]float64{-u, -u, -u, u, u, u}},
	}
	e := newBoxEvaluator()
	e.SetCertifyBudget(200000)
	exprs := make([]string, len(cases))
	for i, c := range cases {
		exprs[i] = c.expr
	}
	got := boxEval(t, e, code, exprs)
	for i, c := range cases {
		for k := 0; k < 6; k++ {
			if d := got[i][k] - c.want[k]; math.Abs(d) > 2e-5*math.Max(1, math.Abs(c.want[k])) {
				t.Errorf("%s.box = %v, want %v", c.expr, got[i], c.want)
				break
			}
		}
		if !anyBounded(got[i]) {
			continue
		}
		for _, slab := range outsideSlabs(got[i]) {
			res, err := certifyOneSign(context.Background(), e, code, c.expr, slab)
			if err != nil {
				t.Fatalf("%s: certify: %v", c.expr, err)
			}
			if res.Disproven {
				t.Errorf("%s has surface outside its box %v, in %v", c.expr, got[i], res.CounterBox)
				break
			}
			if res.Indeterminate {
				t.Errorf("%s: could not prove its box %v in %v (%s)", c.expr, got[i], slab, res.Reason)
				break
			}
		}
	}
}

// TestBoundingBoxDeclared checks name.box = [lo, hi]: the declaration is cut to
// its box grown by boxPad, so a correct box changes nothing and a tight one
// crops. fidget, like the mesher, caps the cut; the viewer leaves it open, which
// leaves the solid whole, and the boxes hold for both. A box declared inside a
// function follows the function's parameters.
func TestBoundingBoxDeclared(t *testing.T) {
	const pad = 0.01
	code := `
ball = sphere(1);
ball.box = [[-1, -1, -1], [1, 1, 1]];
tight = sphere(1);
tight.box = [[-0.5, -0.5, -0.5], [0.5, 0.5, 0.5]];
warped = remap(sphere(1), p * [1, 2, 1]);
warped.box = [[-1, -0.5, -1], [1, 0.5, 1]];
ring(r) {
  s = remap(torus(r, 0.25), [p.y, p.x, p.z]);
  s.box = [[-r - 0.25, -r - 0.25, -0.25], [r + 0.25, r + 0.25, 0.25]];
  return s;
}
big = ring(3);
ground = p.z;
ground.box = [[-1, -1, -1], [1, 1, 0]];
return ball;
`
	e := newBoxEvaluator()
	cases := []struct {
		expr string
		want [6]float64
	}{
		// the declared box, grown by the pad, is the box: the shape inside it
		// is not evaluated
		{"ball", [6]float64{-1 - pad, -1 - pad, -1 - pad, 1 + pad, 1 + pad, 1 + pad}},
		{"tight", [6]float64{-0.5 - pad, -0.5 - pad, -0.5 - pad, 0.5 + pad, 0.5 + pad, 0.5 + pad}},
		// a remap cleared the inferred box; the declared one stands
		{"warped", [6]float64{-1 - pad, -0.5 - pad, -1 - pad, 1 + pad, 0.5 + pad, 1 + pad}},
		// declared inside a function, in terms of its parameter
		{"big", [6]float64{-3.25 - pad, -3.25 - pad, -0.25 - pad, 3.25 + pad, 3.25 + pad, 0.25 + pad}},
		// and carried like any other box
		{"transform.translate(big, [0, 0, 1])", [6]float64{-3.25 - pad, -3.25 - pad, 0.75 - pad, 3.25 + pad, 3.25 + pad, 1.25 + pad}},
		// a half-space's surface is cut to its box, but its solid is not, so
		// the viewer shows the part of a sphere below the ground, far outside
		// the ground's box
		{"ground", [6]float64{-1 - pad, -1 - pad, -1 - pad, 1 + pad, 1 + pad, pad}},
		{"max(ground, sphere([0, 0, -3], 1))", [6]float64{-1, -1, -4, 1, 1, -2}},
	}
	exprs := make([]string, len(cases))
	for i, c := range cases {
		exprs[i] = c.expr
	}
	got := boxEval(t, e, code, exprs)
	for i, c := range cases {
		for k := 0; k < 6; k++ {
			if math.Abs(got[i][k]-c.want[k]) > 2e-5*math.Max(1, math.Abs(c.want[k])) {
				t.Errorf("%s.box = %v, want %v", c.expr, got[i], c.want)
				break
			}
		}
	}

	// A correct box changes nothing; a tight one crops.
	pts := [][3]float64{{0.9, 0, 0}, {0, 0, 0}, {3, 0, 0.1}, {-0.4, 0.7, -0.5}, {0.2, -0.1, 1.4}}
	groups, err := e.EvalGsl(context.Background(), nil, code, "", []string{"ball", "sphere(1)", "tight"}, pts)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	for j := range pts {
		ball, sphere, tight := groups[0].Samples[0][j].Value, groups[1].Samples[0][j].Value, groups[2].Samples[0][j].Value
		if math.Abs(float64(ball-sphere)) > 1e-6 {
			t.Errorf("at %v a correct box changed the field: %v, want %v", pts[j], ball, sphere)
		}
		if tight < sphere-1e-6 {
			t.Errorf("at %v the crop added material: %v < %v", pts[j], tight, sphere)
		}
	}
	if v := groups[2].Samples[0][0].Value; math.Abs(float64(v)-(0.9-0.5-pad)) > 1e-5 {
		t.Errorf("tight at (0.9, 0, 0) = %v, want %v (outside its crop)", v, 0.9-0.5-pad)
	}

	// A box names one declaration, so a destructuring assignment cannot carry one.
	bad := `
pair() { return sphere(1), sphere(2); }
a, b = pair();
a.box = [[-1, -1, -1], [1, 1, 1]];
return a;
`
	if _, err := e.EvalGsl(context.Background(), nil, bad, "", []string{"a"}, pts[:1]); err == nil {
		t.Errorf("a box on a destructured name was accepted")
	}
}

// TestBoundingBoxCalleeLocal checks that a called function's local is its own,
// even when its caller has already evaluated one of the same name: catalog's
// table lookups keep their key in a local k, and a caller with a k of its own
// used to hand its value in.
func TestBoundingBoxCalleeLocal(t *testing.T) {
	code := `
twice(n) { k = n; return k * 2; }
outer(n) { k = n * 10; return sphere(k * 0 + twice(n)); }
return outer(1);
`
	got := boxEval(t, newBoxEvaluator(), code, []string{"outer(1)"})[0]
	want := [6]float64{-2, -2, -2, 2, 2, 2}
	for k := 0; k < 6; k++ {
		if math.Abs(got[k]-want[k]) > 1e-6 {
			t.Fatalf("outer(1).box = %v, want %v", got, want)
		}
	}
}

// TestBoundingBoxUserShadow checks that a document's own box() is not read as
// builtin's: its box is whatever its body says, which here is a unit sphere's,
// not builtin's [0, 1] cube.
func TestBoundingBoxUserShadow(t *testing.T) {
	code := `
box(a, b) { return length(p) - 1; }
return box([0, 0, 0], [1, 1, 1]);
`
	got := boxEval(t, newBoxEvaluator(), code, []string{"box([0, 0, 0], [1, 1, 1])"})[0]
	for k := 0; k < 3; k++ {
		if math.Abs(got[k]+1) > 1e-6 || math.Abs(got[k+3]-1) > 1e-6 {
			t.Fatalf("a user's box() read as builtin's: %v", got)
		}
	}
}

// certifyOneSign proves that expr has no zero in slab. A slab is connected, and
// a field with no zero in it keeps one sign there: >= 0, as a solid inside its
// box is, or <= 0, as a complement is. It proves the first and tries the second
// only when the first is disproven, so a disproof means both signs are there,
// and a zero between them.
func certifyOneSign(ctx context.Context, e *Evaluator, code, expr string, slab [6]float64) (*CertifyResult, error) {
	res, err := e.Certify(ctx, nil, code, "", expr, slab, 0, 0)
	if err != nil || !res.Disproven {
		return res, err
	}
	return e.Certify(ctx, nil, code, "", "-("+expr+")", slab, 0, 0)
}

func finiteAxis(v float64) bool { return math.Abs(v) < unboundedAxis/2 }

func anyBounded(bx [6]float64) bool {
	for _, v := range bx {
		if finiteAxis(v) {
			return true
		}
	}
	return false
}

// outsideSlabs covers a search region minus the box, grown by a hair, with one
// slab per bounded side. The region reaches well past the box on every side,
// and is capped where an axis is unbounded.
func outsideSlabs(bx [6]float64) [][6]float64 {
	return outsideSlabsBy(bx, 0)
}

// outsideSlabsBy is outsideSlabs with the box grown by margin as well as the
// hair.
func outsideSlabsBy(bx [6]float64, margin float64) [][6]float64 {
	reach := 1.0
	for _, v := range bx {
		if finiteAxis(v) {
			reach = math.Max(reach, math.Abs(v))
		}
	}
	reach = 2*reach + 10
	hair := 1e-3*reach + margin
	var region [6]float64
	for k := 0; k < 3; k++ {
		region[k], region[k+3] = -reach, reach
	}
	var out [][6]float64
	for k := 0; k < 3; k++ {
		if lo := bx[k]; finiteAxis(lo) && lo-hair > -reach {
			s := region
			s[k+3] = lo - hair
			out = append(out, s)
		}
		if hi := bx[k+3]; finiteAxis(hi) && hi+hair < reach {
			s := region
			s[k] = hi + hair
			out = append(out, s)
		}
	}
	return out
}
