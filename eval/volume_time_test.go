//go:build cgo

package eval

import (
	"context"
	"testing"
	"time"
)

// volume stops at its time limit with a bracket that still holds, and says time ended it.
func TestVolumeWithin_SaysWhenTheTimeLimitEndedIt(t *testing.T) {
	e := New(nil, nil)
	const code = "s = length(p) - 10.0;\nreturn s;\n"
	region := [6]float64{-12, -12, -12, 12, 12, 12}
	r, err := e.VolumeWithin(context.Background(), nil, code, "", "s", region, 1e-6, time.Nanosecond)
	if err != nil {
		t.Fatal(err)
	}
	const exact = 4188.790204786391
	if !(r.Lower <= exact && exact <= r.Upper) {
		t.Errorf("%v does not bracket %v", r, exact)
	}
	if r.Converged || r.Reason != ReasonTime {
		t.Errorf("converged %v, reason %q: want unconverged by time", r.Converged, r.Reason)
	}
	if r.Evals == 0 {
		t.Errorf("no evaluations before the limit: one batch is always run")
	}
}

// With no limit named, VolumeOf takes the default one, which a quick volume does not reach.
func TestVolumeOf_TakesTheDefaultTimeLimit(t *testing.T) {
	if defaultVolumeTimeLimit != 10*time.Second {
		t.Errorf("default time limit %v, want 10s", defaultVolumeTimeLimit)
	}
	e := New(nil, nil)
	const code = "s = length(p) - 10.0;\nreturn s;\n"
	r, err := e.VolumeOf(context.Background(), nil, code, "", "s", [6]float64{-12, -12, -12, 12, 12, 12}, 0.05)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Converged {
		t.Errorf("a ball at 5%% did not converge: %v", r)
	}
}
