//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// The ball and the wall of reasonDoc come closest along the line from the
// origin to the wall's centre (8, 3, 0), 8.544 away: the ball's surface is 3
// out along it and the wall's 3.5, so max(ball, wall) bottoms out at 0.25
// midway, 3.25 from the origin.
var reasonClosest = [3]float64{3.25 * 8 / math.Sqrt(73), 3.25 * 3 / math.Sqrt(73), 0}

// TestCertifyWorstBoxIsUndecided: when a proof stops short, the box it names
// is one it could not decide, where the claim grazes — not the region. The
// least lower bound ever seen is always the root's, so a worst box taken over
// the whole search named the region every time and read as a search that had
// never subdivided.
func TestCertifyWorstBoxIsUndecided(t *testing.T) {
	check := func(t *testing.T, r *CertifyResult) {
		t.Helper()
		if !r.Indeterminate {
			t.Fatalf("want an indeterminate verdict, got %+v", r)
		}
		if r.WorstBox == reasonRegion {
			t.Fatalf("the worst box is the whole region, %v", r.WorstBox)
		}
		if r.WorstLower <= r.RegionLower {
			t.Errorf("worst lower %.4f is no better than the region's own bound %.4f — it is the root's",
				r.WorstLower, r.RegionLower)
		}
		b := r.WorstBox
		c := [3]float64{(b[0] + b[3]) / 2, (b[1] + b[4]) / 2, (b[2] + b[5]) / 2}
		if d := math.Sqrt((c[0]-reasonClosest[0])*(c[0]-reasonClosest[0]) +
			(c[1]-reasonClosest[1])*(c[1]-reasonClosest[1]) +
			(c[2]-reasonClosest[2])*(c[2]-reasonClosest[2])); d > boxRadius(b)+0.5 {
			t.Errorf("worst box %v is centred %.3f from where the claim grazes, %v — farther than its own radius %.3f",
				b, d, reasonClosest, boxRadius(b))
		}
	}

	t.Run("budget", func(t *testing.T) {
		e := New(&source.MockManager{}, nil)
		e.SetCertifyBudget(4000)
		// 0.25 is the true minimum: the claim holds and grazes, so no
		// budget decides it.
		r, err := e.Certify(context.Background(), nil, reasonDoc, "", "max(ball, wall)", reasonRegion, 0.25, 0)
		if err != nil {
			t.Fatal(err)
		}
		if r.Reason != ReasonBudget {
			t.Fatalf("want reason %q, got %+v", ReasonBudget, r)
		}
		check(t, r)
	})

	t.Run("tolerance", func(t *testing.T) {
		e := New(&source.MockManager{}, nil)
		r, err := e.Certify(context.Background(), nil, reasonDoc, "", "max(ball, wall)", reasonRegion, 0.1, 2.0)
		if err != nil {
			t.Fatal(err)
		}
		if r.Reason != ReasonTolerance {
			t.Fatalf("want reason %q, got %+v", ReasonTolerance, r)
		}
		check(t, r)
	})
}

// TestCertifyDefaultRunsToTheBudget: with no tolerance named, a real margin
// thinner than 1e-3 of the region diagonal — the old default tolerance, 0.03
// over this region — still proves. The claim is 0.01 inside the true minimum.
// A seated driver bit is this case: it touches its screw, and the claim worth
// proving is that it goes no deeper than a hundredth.
func TestCertifyDefaultRunsToTheBudget(t *testing.T) {
	e := New(&source.MockManager{}, nil)
	r, err := e.Certify(context.Background(), nil, reasonDoc, "", "max(ball, wall)", reasonRegion, 0.24, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Proven {
		t.Fatalf("a 0.01 margin with no tolerance named should prove within the default budget, got %+v", r)
	}

	// Named, the tolerance is still a floor: the same claim at the old
	// default gives up.
	r, err = e.Certify(context.Background(), nil, reasonDoc, "", "max(ball, wall)", reasonRegion, 0.24, 0.03)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Indeterminate || r.Reason != ReasonTolerance {
		t.Fatalf("at a named tolerance of 0.03 the 0.01 margin should stop on the tolerance, got %+v", r)
	}
}

// paramReasonDoc is reasonDoc with the wall on a slider: over gap in
// [0, 0.1] the wall's surface stands 0.4 to 0.5 off the ball's, so
// max(ball, wall) is never below 0.2.
const paramReasonDoc = `
gap = 0.1;
gap.min = 0.0;
gap.max = 0.1;
ball = length(p) - 3.0;
wall = length(p - [8.0, 3.0, 0.0]) - (5.144 - gap);
return ball;
`

// TestCertifyParamsDefaultRunsToTheBudget is the parametric search's twin:
// with no tolerance named it splits past the old default down to what the
// budget allows, so a 0.01 margin over a slider's whole range proves, and a
// grazing claim names an undecided cell rather than the root.
func TestCertifyParamsDefaultRunsToTheBudget(t *testing.T) {
	e := New(&source.MockManager{}, nil)
	params := map[string][2]float64{"gap": {0, 0.1}}
	r, err := e.CertifyParams(context.Background(), nil, paramReasonDoc, "", "max(ball, wall)", reasonRegion, 0.19, 0, params)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Proven {
		t.Fatalf("a 0.01 margin over gap in [0, 0.1] should prove with no tolerance named, got %+v", r)
	}

	e.SetCertifyBudget(4000)
	r, err = e.CertifyParams(context.Background(), nil, paramReasonDoc, "", "max(ball, wall)", reasonRegion, 0.2, 0, params)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Indeterminate || r.Reason != ReasonBudget {
		t.Fatalf("the grazing claim should stop on the budget, got %+v", r)
	}
	if r.WorstBox == reasonRegion {
		t.Fatalf("the worst box is the whole region, %v", r.WorstBox)
	}
}
