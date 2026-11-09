package gsleval

import (
	"context"
	"math"
	"strconv"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// rackDoc exercises the rack additions to the gears library: the basic-rack
// profile, its meshing shift, and the assembled rack-and-pinion pair.
const rackDoc = `
prof = gears.rackProfile(1.0, 0.35, 0.0, 0.0, 0.25, 0.02);
loose = gears.rackProfile(1.0, 0.35, 0.0, 0.1, 0.25, 0.02);
bar = gears.rack(1.0, 0.35, 0.0, 0.1, 20.0, 4.0, 3.0, 0.25, 0.02);
return prof;
`

func rackEvaluator(t *testing.T) *Evaluator {
	t.Helper()
	mgr := source.NewLocalManager("testpkg", "../../stdlib")
	return New(mgr, nil)
}

// TestRackProfileDimensions reads the basic-rack proportions back out of the
// field: circular pitch pi*m, tooth half-thickness pi*m/4 at the pitch line,
// tips one module above it, roots (1 + clearance) modules below, and a
// backlash that thins the tooth by exactly that much.
func TestRackProfileDimensions(t *testing.T) {
	e := rackEvaluator(t)
	ctx := context.Background()
	at := func(expr string, x, y float64) float64 {
		t.Helper()
		g, err := e.EvalGsl(ctx, nil, rackDoc, "", []string{expr}, [][3]float64{{x, y, 0}})
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		return float64(g[0].Samples[0][0].Value)
	}
	near := func(label string, got, want, tol float64) {
		t.Helper()
		if math.IsNaN(got) || math.Abs(got-want) > tol {
			t.Errorf("%s = %.4f, want %.3f±%.3f", label, got, want, tol)
		}
	}

	const pitch = math.Pi         // circular pitch at module 1
	const halfThick = math.Pi / 4 // tooth half-thickness at the pitch line

	// A tooth is centered at x = 0 and repeats every circular pitch: the
	// flank surface sits at ±pi/4 from each tooth's centerline.
	near("flank at pitch line", at("prof", halfThick, 0), 0, 1e-3)
	near("flank, next tooth over", at("prof", pitch+halfThick, 0), 0, 1e-3)
	near("flank, previous tooth", at("prof", -pitch-halfThick, 0), 0, 1e-3)
	// Mid-space is empty, mid-tooth is solid.
	if v := at("prof", pitch/2, 0); v <= 0 {
		t.Errorf("tooth space at the pitch line is solid (%.4f)", v)
	}
	if v := at("prof", 0, 0); v >= 0 {
		t.Errorf("tooth centerline at the pitch line is empty (%.4f)", v)
	}
	// Tip one module up, root 1.25 modules down (clearance 0.25).
	near("tooth tip", at("prof", 0, 1.0), 0, 2e-2)
	near("root line", at("prof", pitch/2, -1.25), 0, 2e-2)
	// Backlash 0.1 thins each flank by 0.05, measured normal to the flank.
	tight, slack := at("prof", halfThick, 0), at("loose", halfThick, 0)
	near("backlash thins the flank", slack-tight, 0.05*math.Cos(0.35), 5e-3)
	// The bounded bar stops at its stated depth and half-length.
	near("bar depth", at("bar", pitch/2, -4.0), 0, 3e-2)
	if v := at("bar", 12.0, -1.0); v <= 0 {
		t.Errorf("bar extends past its half-length (%.4f)", v)
	}
}

// TestRackAndPinionMesh is the phase check, and the reason rackShift's
// constant can be trusted: across a full tooth pitch of rack travel the
// mesh clearance must stay positive and essentially constant. A wrong phase
// does not merely shift the number — it swings it deeply negative (the
// half-pitch error this test caught during authoring read -0.69).
func TestRackAndPinionMesh(t *testing.T) {
	e := rackEvaluator(t)
	ctx := context.Background()

	const doc = `
travel = 0.0;
pin, rk = gears.rackAndPinion(24.0, 1.0, 0.35, travel, 0.1, 0.25, 0.02);
meshGap = max(pin, rk);
return min(pin, rk);
`
	// The mesh region: the contact band around the rack's pitch line at
	// y = -pitchRadius = -12.
	region := [6]float64{-8, -15, -0.3, 8, -9, 0.3}

	const steps = 8
	minGap, maxGap := math.Inf(1), math.Inf(-1)
	for i := 0; i < steps; i++ {
		travel := math.Pi * float64(i) / float64(steps-1) // one full circular pitch
		defs := "travel = " + strconv.FormatFloat(travel, 'f', 6, 64) + ";"
		r, err := e.Extremum(ctx, nil, doc, defs, "meshGap", region, "min", 0, 0)
		if err != nil {
			t.Fatalf("extremum at travel %.3f: %v", travel, err)
		}
		if r.Min <= 0 {
			t.Errorf("pinion and rack interfere at travel %.3f: gap %.4f", travel, r.Min)
		}
		minGap, maxGap = math.Min(minGap, r.Min), math.Max(maxGap, r.Min)
	}
	// Rolling contact means the clearance is the same at every travel; a
	// phase or ratio error shows up as variation long before it shows up as
	// interference.
	if spread := maxGap - minGap; spread > 0.005 {
		t.Errorf("mesh clearance varies with travel (%.4f … %.4f, spread %.4f) — the pair is not rolling in step",
			minGap, maxGap, spread)
	}
	t.Logf("mesh clearance %.4f … %.4f across one tooth pitch of travel", minGap, maxGap)
}
