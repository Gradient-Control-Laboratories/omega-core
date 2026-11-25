//go:build cgo

package eval_test

import (
	"context"
	"math"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/eval"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
)

func TestEvalGslPostfixModifierChaining(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)

	doc := `
m = [[1.0, 2.0, 3.0], [4.0, 5.0, 6.0]];
pts = [[10.0, 20.0, 30.0], [40.0, 50.0, 60.0]];
flags = [0.0, 1.0];
@range(0, 10) t = 3.0;
v = [2.0 * t, t * t, 5.0];
body = sphere([1.0, 2.0, 3.0], 0.5);
f(q) { return remap(sphere(2.0), q); }
return body;
`

	queries := []string{
		"m[0][1]",
		"-m[0][1]",
		"pts[0].x",
		"pts[1].yz",
		"body.box[0]",
		"body.box[1].z",
		"!flags[0]",
		"(f(p)).dp.x",
		"v.d(t)[0]",
		"v.d(t)[1]",
	}

	samplePts := [][3]float64{
		{4, 0, 0},
		{0, 4, 0},
	}

	results, err := e.EvalGsl(context.Background(), nil, doc, "", queries, samplePts)
	require.NoError(t, err)
	require.Equal(t, len(queries), len(results))

	assertScalar := func(qi int, wantPt0, wantPt1 float64) {
		t.Helper()
		res := results[qi]
		require.Equal(t, 1, len(res.Samples))
		if got := float64(res.Samples[0][0].Value); math.Abs(got-wantPt0) > 1e-4 {
			t.Errorf("%s at pt0 = %v, want %v", queries[qi], got, wantPt0)
		}
		if got := float64(res.Samples[0][1].Value); math.Abs(got-wantPt1) > 1e-4 {
			t.Errorf("%s at pt1 = %v, want %v", queries[qi], got, wantPt1)
		}
	}

	assertVec := func(qi int, want []float64) {
		t.Helper()
		res := results[qi]
		require.Equal(t, len(want), len(res.Samples))
		for c, w := range want {
			if got := float64(res.Samples[c][0].Value); math.Abs(got-w) > 1e-4 {
				t.Errorf("%s comp %d = %v, want %v", queries[qi], c, got, w)
			}
		}
	}

	assertScalar(0, 2, 2)   // m[0][1]
	assertScalar(1, -2, -2) // -m[0][1]
	assertScalar(2, 10, 10) // pts[0].x
	assertVec(3, []float64{50, 60})
	assertVec(4, []float64{0.5, 1.5, 2.5})
	assertScalar(5, 3.5, 3.5) // body.box[1].z
	assertScalar(6, 1, 1)     // !flags[0]
	assertScalar(7, 1, 0)     // (f(p)).dp.x: unit x-gradient at (4,0,0), 0 at (0,4,0)
	assertScalar(8, 2, 2)     // v.d(t)[0]
	assertScalar(9, 6, 6)     // v.d(t)[1] at t=3
}
