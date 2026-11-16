//go:build cgo

package eval

import (
	"context"
	"testing"
)

// A tolerance the budget cannot reach: the bracket is still sound, and the result says it did not converge.
func TestVolumeOf_SaysWhenTheBudgetEndedIt(t *testing.T) {
	defer func(was int) { volumeBudget = was }(volumeBudget)
	volumeBudget = 20_000

	e := New(nil, nil)
	const code = "s = length(p) - 10.0;\nreturn s;\n"
	region := [6]float64{-12, -12, -12, 12, 12, 12}
	r, err := e.VolumeOf(context.Background(), nil, code, "", "s", region, 1e-6)
	if err != nil {
		t.Fatal(err)
	}
	const exact = 4188.790204786391
	if !(r.Lower <= exact && exact <= r.Upper) {
		t.Errorf("%v does not bracket %v", r, exact)
	}
	if r.Converged || r.Reason != ReasonBudget {
		t.Errorf("converged %v, reason %q: want unconverged by budget", r.Converged, r.Reason)
	}

	volumeBudget = 64_000_000
	r, err = e.VolumeOf(context.Background(), nil, code, "", "s", region, 0.05)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Converged || r.Reason != "" {
		t.Errorf("converged %v, reason %q: want converged", r.Converged, r.Reason)
	}
}
