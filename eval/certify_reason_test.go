//go:build cgo

package eval

import (
	"context"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// reasonDoc: a ball and a wall with a 0.5 gap between their surfaces, so
// min max(ball, wall) = 0.25 — a threshold of d/2 proves d of clearance. The
// wall sits OFF the x-axis on purpose: with both centres on an axis, the
// axis-aligned subdivision plane lands exactly on the equidistant surface,
// interval max becomes tight there, and even a 0.01 margin resolves in two
// cells — which is a fact about the geometry, not about the instrument.
const reasonDoc = `
ball = length(p) - 3.0;
wall = length(p - [8.0, 3.0, 0.0]) - 5.044;
return ball;
`

var reasonRegion = [6]float64{-8, -8, -8, 12, 8, 8}

// TestCertifyReasonNamesWhyItStopped: an Indeterminate verdict must say which
// of the two ways a proof stops short this was, because they call for
// opposite repairs — budget means the geometry is too tight to resolve (open
// the margin), tolerance means the instrument is too coarse to see a real
// margin (lower the tolerance). A proof that resolves carries no reason.
func TestCertifyReasonNamesWhyItStopped(t *testing.T) {
	t.Run("proven carries no reason", func(t *testing.T) {
		e := New(&source.MockManager{}, nil)
		r, err := e.Certify(context.Background(), nil, reasonDoc, "", "max(ball, wall)", reasonRegion, 0.1, 0)
		if err != nil {
			t.Fatal(err)
		}
		if !r.Proven || r.Reason != "" {
			t.Fatalf("a 0.5 gap against a 0.1 threshold should prove with no reason, got %+v", r)
		}
	})

	t.Run("budget", func(t *testing.T) {
		e := New(&source.MockManager{}, nil)
		// A budget far too small for any real proof: the search is cut short
		// with cells still straddling, and must say so.
		e.SetCertifyBudget(16)
		r, err := e.Certify(context.Background(), nil, reasonDoc, "", "max(ball, wall)", reasonRegion, 0.1, 0)
		if err != nil {
			t.Fatal(err)
		}
		if !r.Indeterminate || r.Reason != ReasonBudget {
			t.Fatalf("a 16-eval budget should stop short with reason %q, got %+v", ReasonBudget, r)
		}
	})

	t.Run("tolerance", func(t *testing.T) {
		e := New(&source.MockManager{}, nil)
		// A 0.15 margin under a tolerance of 2.0: cells stop subdividing at
		// radius 2 while still straddling the tilted equidistant surface, so
		// the search ends undecided with the margin real but unseen.
		r, err := e.Certify(context.Background(), nil, reasonDoc, "", "max(ball, wall)", reasonRegion, 0.1, 2.0)
		if err != nil {
			t.Fatal(err)
		}
		if !r.Indeterminate || r.Reason != ReasonTolerance {
			t.Fatalf("a margin below the tolerance should stop with reason %q, got %+v", ReasonTolerance, r)
		}
	})
}
