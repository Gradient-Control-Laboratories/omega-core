package joints_test

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/gsleval"
	"gradientcontrol.com/omega/core/internal/source"
)

// TestDeclaredInterfaces: joint.planar's two half-spaces are each negative on
// their own side of the face, and joint.slide's channel stands the declared
// clearance outside its envelope, on the side wall and the top alike.
func TestDeclaredInterfaces(t *testing.T) {
	doc := `above, below = joint.planar([0.0, 0.0, -7.5], vector.dirZ);
channel, envelope = joint.slide([0.0, 17.5, -4.0], vector.dirX, vector.dirZ, 6.0, 5.0, 0.3);
return sphere(1.0);
`
	points := [][3]float64{
		{0, 0, -5},       // above the face
		{0, 0, -9},       // below it
		{10, 17.5, -4},   // the channel's centre line, anywhere along x
		{10, 20.8, -4},   // the channel's side wall: 3 + 0.3 across
		{10, 17.5, -1.2}, // its top: 2.5 + 0.3 up
	}
	want := map[string][]float64{
		"above":    {-2.5, 1.5, -3.5, -3.5, -6.3},
		"below":    {2.5, -1.5, 3.5, 3.5, 6.3},
		"channel":  {14.2, 14.2, -2.8, 0, 0},
		"envelope": {14.5, 14.5, -2.5, 0.3, 0.3},
	}
	names := []string{"above", "below", "channel", "envelope"}
	e := gsleval.New(&source.MockManager{}, nil)
	groups, err := e.EvalGsl(context.Background(), nil, doc, "", names, points)
	if err != nil {
		t.Fatal(err)
	}
	for i, name := range names {
		for j, w := range want[name] {
			if got := float64(groups[i].Samples[0][j].Value); math.Abs(got-w) > 1e-4 {
				t.Errorf("%s at %v = %.4f, want %.4f", name, points[j], got, w)
			}
		}
	}
}
