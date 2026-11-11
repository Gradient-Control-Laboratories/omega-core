//go:build cgo

package eval_test

import (
	"context"
	"testing"

	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// TestCertifyDisjointProven certifies non-interference soundly: two disjoint
// spheres, max(a, b) >= 0 everywhere, proven by interval arithmetic with no
// Lipschitz constant.
func TestCertifyDisjointProven(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	region := [6]float64{-8, -8, -8, 8, 8, 8}
	disjoint := "max(length(p - [3.0, 0.0, 0.0]) - 1.0, length(p + [3.0, 0.0, 0.0]) - 1.0)"
	r, err := e.Certify(context.Background(), nil, testDoc, "", disjoint, region, 0, 0)
	require.NoError(t, err)
	if !r.Proven {
		t.Fatalf("disjoint: want Proven, got proven=%v disproven=%v indet=%v region=[%.3f,%.3f] evals=%d",
			r.Proven, r.Disproven, r.Indeterminate, r.RegionLower, r.RegionUpper, r.Evals)
	}
}

// TestCertifyOverlapDisproven certifies interference: two overlapping spheres,
// max(a, b) < 0 near the origin, disproven with a proven counterexample box.
func TestCertifyOverlapDisproven(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	region := [6]float64{-8, -8, -8, 8, 8, 8}
	overlap := "max(length(p - [0.5, 0.0, 0.0]) - 1.0, length(p + [0.5, 0.0, 0.0]) - 1.0)"
	r, err := e.Certify(context.Background(), nil, testDoc, "", overlap, region, 0, 0)
	require.NoError(t, err)
	if !r.Disproven {
		t.Fatalf("overlap: want Disproven, got proven=%v disproven=%v indet=%v evals=%d",
			r.Proven, r.Disproven, r.Indeterminate, r.Evals)
	}
	if r.CounterUpper >= 0 {
		t.Errorf("counterexample upper = %v, want < 0 (proven violation)", r.CounterUpper)
	}
	// The counterexample box must lie where the field is negative — near the
	// origin overlap, so it brackets x = 0.
	if b := r.CounterBox; b[0] > 0 || b[3] < 0 {
		t.Errorf("counterexample box %v does not bracket the origin overlap", b)
	}
}

// TestCertifyPinBoreProven walks the campaign's headline: certify Demo A's
// placed pin does not interfere with the housing — a sound, constant-free
// proof of non-interference over the whole region.
func TestCertifyPinBoreProven(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	defs := "pinPlaced = remap(() { return cyl(pinRadius, 8.0); }(), p - [0.0, 0.0, 10.0]);"
	region := [6]float64{-16, -16, -1, 16, 16, 21}
	r, err := e.Certify(context.Background(), nil, pinBoreDoc, defs, "max(pinPlaced, housing)", region, 0, 0)
	require.NoError(t, err)
	if !r.Proven {
		t.Fatalf("pin/bore: want Proven non-interference, got proven=%v disproven=%v indet=%v worstLower=%v evals=%d",
			r.Proven, r.Disproven, r.Indeterminate, r.WorstLower, r.Evals)
	}
}
