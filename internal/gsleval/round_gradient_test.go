package gsleval

import (
	"context"
	"math"
	"testing"
)

// TestRoundIntersectionGradient checks that intersectRound(fields, radius) has a finite, unit gradient everywhere
// near and inside its solid: at radius 0 (every sharp box, cylinder and rectangle is one) the square root of a zero
// sum made it NaN inside, and min(sharp, -radius) tied on the surface, where it took a constant's zero gradient.
// Meshing and lighting read the gradient at every sample, so either broke them.
func TestRoundIntersectionGradient(t *testing.T) {
	code := `
sharpBox = box([0, 0, 0], [10, 20, 30]);
roundBox = box([0, 0, 0], [10, 20, 30], intersectRound(0.8));
slab = max(rectangleXYCentered([5, 10], [10, 20]), abs(p.z - 15) - 15);
return sharpBox;
`
	// A grid through the box and a little beyond it, whose planes hold the box's faces, edges and corners.
	var points [][3]float64
	for x := -1.0; x <= 11; x += 0.5 {
		for y := -1.0; y <= 21; y += 1 {
			for z := -1.0; z <= 31; z += 1 {
				points = append(points, [3]float64{x, y, z})
			}
		}
	}
	e := newBoxEvaluator()
	for _, expr := range []string{"sharpBox", "roundBox", "slab"} {
		groups, err := e.EvalGsl(context.Background(), nil, code, "", []string{expr}, points)
		if err != nil {
			t.Fatalf("%s: %v", expr, err)
		}
		bad := 0
		for i, s := range groups[0].Samples[0] {
			norm := math.Sqrt(float64(s.Dx*s.Dx + s.Dy*s.Dy + s.Dz*s.Dz))
			if math.IsNaN(norm) || math.Abs(norm-1) > 1e-3 {
				if bad < 3 {
					t.Errorf("%s at %v: value %g, gradient (%g, %g, %g), |gradient| %g; want unit length",
						expr, points[i], s.Value, s.Dx, s.Dy, s.Dz, norm)
				}
				bad++
			}
		}
		if bad > 0 {
			t.Errorf("%s: %d of %d points have no unit gradient", expr, bad, len(points))
		}
	}
}

// TestRoundIntersectionValue pins the value the gradient fix must not change: the box's signed distance inside, on
// its faces, and the Euclidean distance outside its edges and corners that radius 0 gives (a sharp max would not).
func TestRoundIntersectionValue(t *testing.T) {
	code := "blk = box([0, 0, 0], [10, 20, 30]);\nreturn blk;\n"
	cases := []struct {
		at   [3]float64
		want float64
	}{
		{[3]float64{5, 10, 15}, -5},              // the centre: 5 to the nearest face
		{[3]float64{0, 10, 15}, 0},               // on a face
		{[3]float64{12, 10, 15}, 2},              // off a face
		{[3]float64{13, 24, 15}, 5},              // off an edge: 3-4-5
		{[3]float64{-1, -2, -2}, 3},              // off a corner: 1-2-2
		{[3]float64{10, 20, 30}, 0},              // a corner
		{[3]float64{9, 19, 29}, -1},              // just inside a corner
		{[3]float64{11, 21, 31}, math.Sqrt(3.0)}, // just outside one
	}
	points := make([][3]float64, len(cases))
	for i, c := range cases {
		points[i] = c.at
	}
	groups, err := newBoxEvaluator().EvalGsl(context.Background(), nil, code, "", []string{"blk"}, points)
	if err != nil {
		t.Fatal(err)
	}
	for i, s := range groups[0].Samples[0] {
		if got := float64(s.Value); math.Abs(got-cases[i].want) > 1e-5 {
			t.Errorf("box at %v = %g, want %g", cases[i].at, got, cases[i].want)
		}
	}
}
