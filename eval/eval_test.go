//go:build cgo

package eval_test

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/eval/joints"
)

func TestEvalAPI(t *testing.T) {
	e := eval.New(nil, nil)
	ctx := context.Background()
	code := `
rad = 5.0;
rad.min = 1.0;
rad.max = 10.0;
s = sphere(rad);
return s;
`
	groups, err := e.EvalGsl(ctx, nil, code, "", []string{"s"}, [][3]float64{{0, 0, 0}, {5, 0, 0}})
	if err != nil {
		t.Fatalf("EvalGsl: %v", err)
	}
	if len(groups) != 1 || len(groups[0].Samples) != 1 || len(groups[0].Samples[0]) != 2 {
		t.Fatalf("unexpected groups: %+v", groups)
	}
	if got := float64(groups[0].Samples[0][0].Value); math.Abs(got - -5.0) > 1e-4 {
		t.Errorf("sample[0] = %v, want -5.0", got)
	}
	if got := float64(groups[0].Samples[0][1].Value); math.Abs(got) > 1e-4 {
		t.Errorf("sample[1] = %v, want 0.0", got)
	}

	bounds, err := e.Bounds(ctx, nil, code, "", "s", [6]float64{-10, -10, -10, 10, 10, 10}, 16)
	if err != nil {
		t.Fatalf("Bounds: %v", err)
	}
	if !bounds.Found {
		t.Fatal("Bounds.Found = false, want true")
	}

	cert, err := e.Certify(ctx, nil, code, "", "s", [6]float64{6, 6, 6, 10, 10, 10}, 0, 0)
	if err != nil {
		t.Fatalf("Certify: %v", err)
	}
	if !cert.Proven {
		t.Errorf("Certify.Proven = false, want true")
	}

	ranges, err := eval.SliderRanges(code)
	if err != nil {
		t.Fatalf("SliderRanges: %v", err)
	}
	if ranges["rad"] != ([2]float64{1, 10}) {
		t.Errorf("SliderRanges[rad] = %v, want [1, 10]", ranges["rad"])
	}

	names := eval.TopLevelNames(code)
	if !names["rad"] || !names["s"] {
		t.Errorf("TopLevelNames = %v, want rad and s", names)
	}

	returns := eval.TopLevelReturnExprs(code)
	if len(returns) != 1 || returns[0] != "s" {
		t.Errorf("TopLevelReturnExprs = %v, want [s]", returns)
	}
}

func TestJointsAPI(t *testing.T) {
	found, returns, skipped := joints.Find(`
screw, clamped, threaded = joint.bolted(top, bot, [0, 0, 5], [0, 0, -1], 4, 5, 0);
return screw, clamped, threaded;
`)
	if len(skipped) != 0 {
		t.Errorf("skipped = %v, want empty", skipped)
	}
	if len(found) != 1 || found[0].Kind != "bolted" {
		t.Errorf("found = %+v, want 1 bolted joint", found)
	}
	if len(returns) != 3 {
		t.Errorf("returns = %v, want 3 expressions", returns)
	}
}
