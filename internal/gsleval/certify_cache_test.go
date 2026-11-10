package gsleval

import (
	"context"
	"testing"

	"gradientcontrol.com/omega/core/internal/fidget/lib"
	"gradientcontrol.com/omega/core/internal/source"
)

// cacheBallDoc is paramBallDoc with the radius read through a product with a
// plain literal, which a plain build folds to a number.
const cacheBallDoc = `
rad = 3.0;
rad.min = 1.0;
rad.max = 6.5;
gain = 1.0;
wall = length(p - [8.0, 0.0, 0.0]) - 2.0;
`

// TestCertifyParamsAfterCachedPlainQuery: with queries cached, a plain query
// on the same source must not fold the slider into the program Certify then
// receives. The clearance claim fails at rad = 6.5, so it must never prove.
func TestCertifyParamsAfterCachedPlainQuery(t *testing.T) {
	const expr = "max(length(p) - rad * gain, wall)"
	e := New(&source.MockManager{}, nil)
	e.CacheQueries(8)
	if _, err := e.Volume(context.Background(), cacheBallDoc, expr, paramRegion, 0.1); err != nil {
		t.Fatal(err)
	}
	res, err := e.CertifyOverSliders(context.Background(), nil, cacheBallDoc, "", expr, paramRegion, 0.25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if res.Proven || !res.Disproven {
		t.Fatalf("after a cached plain query, certify must still disprove a claim that fails at rad = 6.5: %+v", res)
	}
}

// TestParamBuildAfterPlainBuild: a plain build of a program must leave the
// sliders free for a later parameter build of the same program.
func TestParamBuildAfterPlainBuild(t *testing.T) {
	e := New(&source.MockManager{}, nil)
	prog, err := e.prepareQuery(context.Background(), nil, cacheBallDoc, "", []string{"length(p) - rad * gain"})
	if err != nil {
		t.Fatal(err)
	}
	box := [][6]float32{{1, 0.5, 0, 1, 0.5, 0}}
	if _, err := lib.EvalInterval(prog, box); err != nil {
		t.Fatal(err)
	}
	ivals, err := lib.EvalIntervalParams(prog, box, map[string][2]float64{"rad": {1, 6}})
	if err != nil {
		t.Fatal(err)
	}
	// |p| = 1.118 at the point, so rad over [1, 6] spans [-4.88, 0.12].
	if got := ivals[0]; got.Upper-got.Lower < 4.9 {
		t.Fatalf("bracket %+v is the default radius alone: the plain build folded rad", got)
	}
}
