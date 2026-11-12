//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// TestSeatedAtTurnsNotMirrors: a screw driven UP is the library's screw turned
// over, a half turn about x, and not its reflection — a reflected right-hand
// thread is a left-hand one. So seatedAt negates y along with z when it flips,
// and leaves both alone for the library's own downward datum.
func TestSeatedAtTurnsNotMirrors(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	const doc = `
up = fastener.seatedAt([1.0, 1.0, 1.0], vector.dirZ);
down = fastener.seatedAt([1.0, 1.0, 1.0], neg(vector.dirZ));
return sphere(1.0);
`
	pt := [3]float64{2, 3, 4}
	for expr, want := range map[string]float64{
		"up.x": 1, "up.y": -2, "up.z": -3,
		"down.x": 1, "down.y": 2, "down.z": 3,
	} {
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{expr}, [][3]float64{pt})
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		if got := float64(groups[0].Samples[0][0].Value); math.Abs(got-want) > 1e-6 {
			t.Errorf("%s at %v = %v, want %v", expr, pt, got, want)
		}
	}
}
