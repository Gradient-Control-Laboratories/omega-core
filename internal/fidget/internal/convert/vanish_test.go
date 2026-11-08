package convert_test

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/fidget/internal/convert"
	"gradientcontrol.com/omega/core/internal/fidget/internal/fidget"
)

// A part a renderer marches as the viewer does (partProxy) is its shape while
// the radius is 0, which is how such a renderer shades: so its gradient, the
// normal it shades by, has to be the shape's. Inside the part's box — where
// every point of its surface is — the box's distance is 0, and a square root's
// slope at 0 once made that gradient NaN, which the CPU renderer drew as
// unlit, unreflecting parts.
func TestPartProxyShadesAsItsShape(t *testing.T) {
	src := "ball = sphere(2);\nreturn ball;\n"
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mockMgr))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	// Inside the box: the centre, near the surface, on it, and in a corner of
	// the box outside the ball.
	xs := []float32{0.1, 1.2, 0, 1.7}
	ys := []float32{0.2, 0.6, 2, 1.7}
	zs := []float32{0.3, 1.0, 0, 1.7}

	gradOf := func(s convert.ColorShape) []fidget.GradResult {
		r, err := fidget.EvalGrad(fidget.BuildCanvas(s.Tree), xs, ys, zs)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}

	var plain []fidget.GradResult
	if err := convert.ToColorShapes(res.Program, 0, func(shapes []convert.ColorShape) error {
		plain = gradOf(shapes[0])
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if err := convert.ToColorShapesWithPixelSize(res.Program, 0, func(shapes []convert.ColorShape, _ fidget.TreeIndex, _ bool, _ fidget.TreeIndex, hasPartProxy bool) error {
		if !hasPartProxy {
			t.Fatal("the ball was not built as a part proxy, so this test checks nothing")
		}
		proxied := gradOf(shapes[0])
		for j := range xs {
			got, want := proxied[j], plain[j]
			for _, v := range []float32{got.Value, got.Dx, got.Dy, got.Dz} {
				if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
					t.Fatalf("(%g, %g, %g): proxied part is %+v, not finite", xs[j], ys[j], zs[j], got)
				}
			}
			if got != want {
				t.Errorf("(%g, %g, %g): proxied part %+v, its shape %+v", xs[j], ys[j], zs[j], got, want)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
