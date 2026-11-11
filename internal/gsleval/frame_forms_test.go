package gsleval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// frameEval evaluates scalar expressions in a document binding `bar`, a
// 20 mm cylinder along z through the origin, at the given points.
func frameEval(t *testing.T, exprs []string, points [][3]float64) [][]float64 {
	t.Helper()
	mgr := source.NewLocalManager("../../stdlib")
	e := New(mgr, nil)
	doc := "bar = cylinder([0, 0, 0], [0, 0, 1], 10);\nreturn bar;\n"
	groups, err := e.EvalGsl(context.Background(), nil, doc, "", exprs, points)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	// A scalar expression's samples are one row, a value per point.
	out := make([][]float64, len(exprs))
	for i := range exprs {
		for _, s := range groups[i].Samples[0] {
			out[i] = append(out[i], float64(s.Value))
		}
	}
	return out
}

func onBar(deg, z float64) [3]float64 {
	r := deg * math.Pi / 180
	return [3]float64{10 * math.Cos(r), 10 * math.Sin(r), z}
}

// A checkerboard on a part's own frame counts an even number of cells round
// u, the nearest to the circumference's, so the board closes on itself: on a
// 20 mm bar with 5 mm cells that is twelve, a cell every 30 degrees, each
// the other colour from its neighbours round and along, and the cell after
// the seam the same colour as the cell before the first.
func TestCheckerboardFrameFormClosesRoundTheBar(t *testing.T) {
	points := [][3]float64{
		onBar(15, 1), onBar(45, 1), onBar(75, 1), onBar(345, 1), // cells 0, 1, 2 and 11 round, first row along
		onBar(15, 6), // cell 0 round, second row along
	}
	got := frameEval(t, []string{"playful.checkerboard(5, [1, 0, 0], [0, 0, 1], [bar.u, bar.v], 20).x"}, points)[0]
	red := func(v float64) bool { return v > 0.9 }
	blue := func(v float64) bool { return v < 0.1 }
	if !(red(got[0]) && blue(got[1]) && red(got[2])) {
		t.Errorf("cells 0, 1, 2 round the bar read %v, want red, blue, red", got[:3])
	}
	if !blue(got[3]) {
		t.Errorf("cell 11, the last before the seam, reads %v, want blue so the board closes against cell 0", got[3])
	}
	if !blue(got[4]) {
		t.Errorf("cell 0 of the second row along reads %v, want blue, the other colour from the row before", got[4])
	}
}

// Polka dots on a part's own frame count a whole number of honeycomb cells
// round u, each sqrt(3) spacing wide, the nearest to the circumference's: on
// a 20 mm bar with 8 mm spacing that is five, a cell every 72 degrees, so a
// dot centre, the same point one cell on and one cell back across the seam
// all read the dot, the point between cells reads the base, and a row up at
// one spacing reads a dot again; and the drift form at time zero through the
// origin is the direction form.
func TestPolkaDotFrameFormClosesRoundTheBar(t *testing.T) {
	points := [][3]float64{onBar(0, 0), onBar(72, 0), onBar(-72, 0), onBar(36, 0), onBar(0, 4), onBar(0, 8)}
	got := frameEval(t, []string{
		"playful.polkaDot(8, 0.3, [1, 0, 0], [0, 0, 1], [bar.u, bar.v], 20).x",
		"playful.polkaDot(8, 0.3, [0, 0, 0], [0, 0, 1], [1, 0, 0], [0, 0, 1], [2, 0, 0]).x",
		"playful.polkaDot(8, 0.3, [0, 0, 1], [1, 0, 0], [0, 0, 1]).x",
	}, points)
	frame := got[0]
	dot := func(v float64) bool { return v < 0.1 }
	base := func(v float64) bool { return v > 0.9 }
	if !(dot(frame[0]) && dot(frame[1]) && dot(frame[2])) {
		t.Errorf("a dot centre, one cell on and one cell back across the seam read %v, want the dot at all three", frame[:3])
	}
	if !base(frame[3]) || !base(frame[4]) {
		t.Errorf("between cells and between rows read %v, want the base colour", frame[3:5])
	}
	if !dot(frame[5]) {
		t.Errorf("one spacing up the row reads %v, want a dot", frame[5])
	}
	for i := range points {
		if math.Abs(got[1][i]-got[2][i]) > 1e-6 {
			t.Errorf("the drift form at time zero differs from the direction form at %v: %v vs %v", points[i], got[1][i], got[2][i])
		}
	}
}

// The rainbow bands take a direction: along +X the colour changes with x and
// holds with z, and the two-argument form is the direction form along z.
func TestRainbowBandsTakeADirection(t *testing.T) {
	points := [][3]float64{{0, 0, 0}, {2.5, 0, 0}, {0, 0, 2.5}, {0, 0, 5}}
	got := frameEval(t, []string{
		"playful.rainbowSheen(10, [1, 0, 0], color.rainbow).x",
		"playful.rainbowSheen(10, color.rainbow).x",
		"playful.rainbowSheen(10, [0, 0, 1], color.rainbow).x",
	}, points)
	alongX := got[0]
	if math.Abs(alongX[0]-alongX[2]) > 1e-4 {
		t.Errorf("bands along x change with z: %v at the origin, %v 2.5 mm up", alongX[0], alongX[2])
	}
	if math.Abs(alongX[0]-alongX[1]) < 1e-3 {
		t.Errorf("bands along x do not change with x: %v at the origin, %v 2.5 mm along", alongX[0], alongX[1])
	}
	for i := range points {
		if math.Abs(got[1][i]-got[2][i]) > 1e-6 {
			t.Errorf("the two-argument form differs from the direction form along z at %v: %v vs %v", points[i], got[1][i], got[2][i])
		}
	}
}
