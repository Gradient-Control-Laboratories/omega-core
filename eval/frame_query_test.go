//go:build cgo

package eval_test

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// The frame reads .u and .v, through the evalGsl query path: each read is the
// frame its shape declares, so a material handed bar.u needs no axis written
// at the call site to be kept in step with the geometry.

// frameDoc is a z-axis bar carrying the cylindrical frame the primitives will
// supply, u the angle about the axis and v the distance along it, plus the
// cases a frame has to survive: the bar moved, a union with a ball carrying
// its own frame, and a shape carrying none.
const frameDoc = `
bar = cylinder([0.0, 0.0, 0.0], [0.0, 0.0, 1.0], 2.0);
bar.u = atan(p.y, p.x);
bar.v = p.z;
moved = remap(bar, p - [10.0, 0.0, 0.0]);
ball = sphere(1.0);
ball.u = 7.0;
both = min(bar, remap(ball, p - [20.0, 0.0, 0.0]));
plain = length(p) - 1.0;
return bar;
`

// frameAt evaluates each expression at the points and returns, per expression,
// the samples of its first component and their gradients.
func frameAt(t *testing.T, pts [][3]float64, exprs ...string) []eval.GradGroup {
	t.Helper()
	e := eval.New(&source.MockManager{}, nil)
	results, err := e.EvalGsl(context.Background(), nil, frameDoc, "", exprs, pts)
	require.NoError(t, err)
	require.Equal(t, len(exprs), len(results))
	return results
}

func within(got float32, want float64) bool { return math.Abs(float64(got)-want) <= 1e-4 }

// A read is the declared frame, evaluated where it is read, and its gradient
// is the coordinate's own: .u.dp runs round the axis at 1/r, the
// cross(axis, offset) / r^2 that machined.turned builds by hand today.
func TestFrameReadIsTheDeclaredFrame(t *testing.T) {
	pts := [][3]float64{{2, 0, 1}, {0, 2, 3}}
	r := frameAt(t, pts, "bar.u", "bar.v")

	wantU := []float64{0, math.Pi / 2}
	wantDu := [][3]float64{{0, 0.5, 0}, {-0.5, 0, 0}}
	for i := range pts {
		s := r[0].Samples[0][i]
		if !within(s.Value, wantU[i]) {
			t.Errorf("bar.u at %v = %v, want %v", pts[i], s.Value, wantU[i])
		}
		if !within(s.Dx, wantDu[i][0]) || !within(s.Dy, wantDu[i][1]) || !within(s.Dz, wantDu[i][2]) {
			t.Errorf("grad bar.u at %v = [%v %v %v], want %v", pts[i], s.Dx, s.Dy, s.Dz, wantDu[i])
		}
		if v := r[1].Samples[0][i].Value; !within(v, pts[i][2]) {
			t.Errorf("bar.v at %v = %v, want %v", pts[i], v, pts[i][2])
		}
	}
}

// .u.dp, read as a value, is that same direction: the lay a material takes.
func TestFrameGradientIsTheLayDirection(t *testing.T) {
	r := frameAt(t, [][3]float64{{0, 2, 0}}, "bar.u.dp")
	want := [3]float64{-0.5, 0, 0}
	for c := 0; c < 3; c++ {
		if got := r[0].Samples[c][0].Value; !within(got, want[c]) {
			t.Errorf("bar.u.dp component %d = %v, want %v", c, got, want[c])
		}
	}
}

// A moved shape carries its frame with it. This is what the feature exists
// for: four bars placed at four places each read their own axis, where one
// material handed one axis reads the wrong one on three of them.
func TestFrameMovesWithItsShape(t *testing.T) {
	r := frameAt(t, [][3]float64{{12, 0, 0}, {10, 2, 0}}, "moved.u")
	for i, want := range []float64{0, math.Pi / 2} {
		if got := r[0].Samples[0][i].Value; !within(got, want) {
			t.Errorf("moved.u sample %d = %v, want %v about the moved axis", i, got, want)
		}
	}
}

// A union reads each part's own frame where that part binds, rather than an
// average of two frames that mean different things.
func TestFrameReadOfAUnionSelects(t *testing.T) {
	r := frameAt(t, [][3]float64{{0, 2, 0}, {21, 0, 0}}, "both.u")
	for i, want := range []float64{math.Pi / 2, 7} {
		if got := r[0].Samples[0][i].Value; !within(got, want) {
			t.Errorf("both.u sample %d = %v, want %v", i, got, want)
		}
	}
}

// A shape with no frame reads zero, whose gradient gives a lay no direction.
func TestFrameOfAFramelessShapeIsZero(t *testing.T) {
	r := frameAt(t, [][3]float64{{1, 0, 0}}, "plain.u")
	s := r[0].Samples[0][0]
	if s.Value != 0 || s.Dx != 0 || s.Dy != 0 || s.Dz != 0 {
		t.Errorf("plain.u = %v with gradient [%v %v %v], want 0 and no direction", s.Value, s.Dx, s.Dy, s.Dz)
	}
}

// bar.second.u is the frame of the second-closest surface: by the bar it is the
// ball's, by the ball it is the bar's. Resolved over both instead of over the
// second-closest distance, the read would answer for the binding surface while
// asking for the other one.
//
// The ball is placed by a declaration, not an inline remap as in both: .second
// itself drops an inline remap of a named shape (both.second reads the ball at
// the origin), which predates frames and is not theirs to work around.
// Differentiated through a declaration, since the grammar takes two accessors
// in a chain, the read is the second surface's own gradient: zero for the
// ball's constant frame, the bar's 1/r round its axis by the ball.
func TestFrameOfTheSecondClosestSurface(t *testing.T) {
	pts := [][3]float64{{0, 2, 0}, {21, 0, 0}}
	const defs = "placed = remap(ball, p - [20.0, 0.0, 0.0]);\npair = min(bar, placed);\nsu = pair.second.u;\n"
	e := eval.New(&source.MockManager{}, nil)
	results, err := e.EvalGsl(context.Background(), nil, frameDoc, defs, []string{"pair.second.u", "su.dp"}, pts)
	require.NoError(t, err)
	require.Equal(t, 2, len(results))

	for i, want := range []float64{7, 0} {
		if got := results[0].Samples[0][i].Value; !within(got, want) {
			t.Errorf("pair.second.u sample %d = %v, want %v", i, got, want)
		}
	}
	want := [][3]float64{{0, 0, 0}, {0, 1.0 / 21, 0}}
	for i := range pts {
		for c := 0; c < 3; c++ {
			if got := results[1].Samples[c][i].Value; !within(got, want[i][c]) {
				t.Errorf("su.dp sample %d component %d = %v, want %v", i, c, got, want[i][c])
			}
		}
	}
}
