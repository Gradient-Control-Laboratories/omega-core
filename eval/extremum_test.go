//go:build cgo

package eval_test

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// TestExtremumSphere finds both extrema of the test document's sphere
// (radius 2 at the origin): the infimum is -2 at the center, the supremum
// sits at a region corner.
func TestExtremumSphere(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)

	region := [6]float64{-5, -5, -5, 5, 5, 5}
	r, err := e.Extremum(context.Background(), nil, testDoc, "", "s", region, "both", 0, 0)
	require.NoError(t, err)

	if !r.MinConverged || !r.MaxConverged {
		t.Fatalf("did not converge: min=%v max=%v evals=%d", r.MinConverged, r.MaxConverged, r.Evals)
	}
	// The certificate must bracket the true infimum, within the adaptive
	// default width (1% of the extremum's magnitude here).
	if r.MinLower > -2+1e-3 || r.Min < -2-1e-3 {
		t.Errorf("certificate [%v, %v] does not bracket -2", r.MinLower, r.Min)
	}
	if r.Min-r.MinLower > 0.01*2+1e-3 {
		t.Errorf("min interval [%v, %v] wider than the adaptive tolerance", r.MinLower, r.Min)
	}
	if d := math.Sqrt(r.MinAt[0]*r.MinAt[0] + r.MinAt[1]*r.MinAt[1] + r.MinAt[2]*r.MinAt[2]); d > 0.03 {
		t.Errorf("MinAt = %v, want near origin", r.MinAt)
	}
	// Supremum: the farthest corner is (±5, ±5, ±5), |p| = 5√3.
	wantMax := 5*math.Sqrt(3) - 2
	if r.Max > wantMax+1e-3 || r.MaxUpper < wantMax-1e-3 {
		t.Errorf("certificate [%v, %v] does not bracket %v", r.Max, r.MaxUpper, wantMax)
	}
	if r.MaxUpper-r.Max > 0.01*wantMax+1e-3 {
		t.Errorf("max interval [%v, %v] wider than the adaptive tolerance", r.Max, r.MaxUpper)
	}
}

// TestExtremumOverlapPredicate checks interference detection via the minimum
// of max(a, b): negative iff the two solids overlap.
func TestExtremumOverlapPredicate(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	region := [6]float64{-8, -8, -8, 8, 8, 8}

	// Disjoint spheres (radius 1 at x = ±3): the intersection field's
	// minimum is +2, at the midpoint of the gap. For exact SDFs this is
	// half the separation distance (surfaces are 4 apart), so one call
	// answers both the overlap predicate (sign) and clearance (2x value).
	disjoint := "max(length(p - [3.0, 0.0, 0.0]) - 1.0, length(p + [3.0, 0.0, 0.0]) - 1.0)"
	r, err := e.Extremum(context.Background(), nil, testDoc, "", disjoint, region, "min", 0, 0)
	require.NoError(t, err)
	if !r.MinConverged {
		t.Fatalf("disjoint case did not converge (evals=%d)", r.Evals)
	}
	if math.Abs(r.Min-2) > 0.03 {
		t.Errorf("disjoint Min = %v, want 2 (half the separation)", r.Min)
	}

	// Overlapping spheres (radius 1 at x = ±0.5): max(a, b) dips to -0.5 at
	// the origin.
	overlap := "max(length(p - [0.5, 0.0, 0.0]) - 1.0, length(p + [0.5, 0.0, 0.0]) - 1.0)"
	r, err = e.Extremum(context.Background(), nil, testDoc, "", overlap, region, "min", 0, 0)
	require.NoError(t, err)
	if !r.MinConverged {
		t.Fatalf("overlap case did not converge (evals=%d)", r.Evals)
	}
	if math.Abs(r.Min+0.5) > 0.01 {
		t.Errorf("overlap Min = %v, want -0.5", r.Min)
	}
	if r.MinAt[0] < -0.51 || r.MinAt[0] > 0.51 || math.Abs(r.MinAt[1]) > 0.02 || math.Abs(r.MinAt[2]) > 0.02 {
		t.Errorf("overlap MinAt = %v, want on the x-axis between the centers", r.MinAt)
	}
}

// TestExtremumClearanceSum checks the sum-form clearance recipe: for two
// disjoint exact SDFs, the minimum of their sum is the separation distance.
// The sum of two 1-Lipschitz fields is 2-Lipschitz, so the pass needs
// lipschitz 2 — and the sum is constant along the whole closest-approach
// segment, so certifying it is slow: this is the expensive fallback to the
// max-form recipe, exercised here with an explicit (looser) tolerance.
func TestExtremumClearanceSum(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	region := [6]float64{-8, -8, -8, 8, 8, 8}

	sum := "(length(p - [3.0, 0.0, 0.0]) - 1.0) + (length(p + [3.0, 0.0, 0.0]) - 1.0)"
	r, err := e.Extremum(context.Background(), nil, testDoc, "", sum, region, "min", 0.1, 2)
	require.NoError(t, err)
	if !r.MinConverged {
		t.Fatalf("did not converge (evals=%d)", r.Evals)
	}
	// Sphere surfaces are 6 - 1 - 1 = 4 apart. The sampled minimum lands on
	// the plateau exactly; only the certificate is tolerance-wide.
	if math.Abs(r.Min-4) > 0.01 {
		t.Errorf("Min = %v, want 4 (separation distance)", r.Min)
	}
	if r.MinLower > 4+1e-6 || r.MinLower < 4-0.11 {
		t.Errorf("MinLower = %v, want within 0.1 below the true infimum 4", r.MinLower)
	}
	// The minimum is attained along the segment between the closest points.
	if math.Abs(r.MinAt[1]) > 0.05 || math.Abs(r.MinAt[2]) > 0.05 || math.Abs(r.MinAt[0]) > 2.01 {
		t.Errorf("MinAt = %v, want on the axis between the spheres", r.MinAt)
	}
}

// pinBoreDoc is the Demo A document, verbatim: a
// block with a vertical bore and a pin lying on its side beside it.
const pinBoreDoc = `
box(x1, y1, z1, x2, y2, z2) {
  return max(
    x1 - p.x, p.x - x2,
    y1 - p.y, p.y - y2,
    z1 - p.z, p.z - z2,
  );
}

cyl(radius, halfLength) {
  return max(length(p.xy) - radius, abs(p.z) - halfLength);
}

boreRadius = 6.0;
housing = max(box(-15.0, -15.0, 0.0, 15.0, 15.0, 20.0),
              -(length(p.xy) - boreRadius));

pinRadius = 5.0;
pinOffset = p - [30.0, 0.0, 5.0];
pin = remap(() { return cyl(pinRadius, 8.0); }(), pinOffset.zyx);

return housing, pin;
`

// TestExtremumPinBoreDemo walks the Demo A verification: place the pin
// coaxially via defs and check interference/clearance with one query. The
// mid-gap minimizers form a 2-D annular sheet, so the certificate may not
// converge within budget — the demo's contract is that the sampled minimum
// is the exact half-clearance and the lower bound still proves the sign.
func TestExtremumPinBoreDemo(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)

	defs := "pinPlaced = remap(() { return cyl(pinRadius, 8.0); }(), p - [0.0, 0.0, 10.0]);"
	region := [6]float64{-16, -16, -1, 16, 16, 21}
	r, err := e.Extremum(context.Background(), nil, pinBoreDoc, defs,
		"max(pinPlaced, housing)", region, "min", 0, 0)
	require.NoError(t, err)

	// Bore r=6, pin r=5: the gap is 1.0, so min = 0.5 at mid-annulus.
	if math.Abs(r.Min-0.5) > 0.01 {
		t.Errorf("Min = %v, want 0.5 (half the radial clearance)", r.Min)
	}
	if r.MinLower <= 0 {
		t.Errorf("MinLower = %v, want > 0 (proves non-interference)", r.MinLower)
	}

	// The original lying-down pin must clear the housing too (it sits at
	// x = 30, well off the block) — and by much more than the bore gap.
	r, err = e.Extremum(context.Background(), nil, pinBoreDoc, "",
		"max(pin, housing)", [6]float64{-20, -20, -10, 45, 20, 25}, "min", 0, 0)
	require.NoError(t, err)
	if r.Min < 2 {
		t.Errorf("lying pin Min = %v, want a wide gap to the housing", r.Min)
	}
}

// crateDoc is the Demo B document, verbatim: three
// parts in an open crate, with a block/drum interference and the ball
// poking through the +y wall.
const crateDoc = `
boxAt(center, half) {
  q = p - center;
  return max(
    -half.x - q.x, q.x - half.x,
    -half.y - q.y, q.y - half.y,
    -half.z - q.z, q.z - half.z,
  );
}

cylAt(center, radius, halfLength) {
  q = p - center;
  return max(length(q.xy) - radius, abs(q.z) - halfLength);
}

sphereAt(center, radius) {
  return length(p - center) - radius;
}

// Open crate: interior 30 x 30 x 20, walls and floor 1 thick, no lid.
crateOuter = boxAt([0.0, 0.0, 9.5], [16.0, 16.0, 10.5]);
crateInner = boxAt([0.0, 0.0, 10.0], [15.0, 15.0, 10.0]);
crate = max(crateOuter, -crateInner);

blockCenter = [-6.0, -5.0, 6.0];
block = boxAt(blockCenter, [5.0, 4.0, 5.0]);

drumCenter = [1.0, -3.0, 6.0];
drum = cylAt(drumCenter, 4.0, 5.0);

ballCenter = [8.0, 12.0, 8.0];
ball = sphereAt(ballCenter, 5.0);

return crate, block, drum, ball;
`

// TestExtremumAssemblyRepairDemo walks the Demo B verification: one batched
// sweep finds the two defects (block/drum overlap, ball escaping the
// crate), then a defs override of the document's center variables — the
// same move the model makes to test a fix before editing — proves the
// repaired layout clean.
func TestExtremumAssemblyRepairDemo(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	region := [6]float64{-18, -18, -2, 18, 18, 22}
	ctx := context.Background()

	minOf := func(t *testing.T, defs, expr string) *eval.ExtremumResult {
		t.Helper()
		r, err := e.Extremum(ctx, nil, crateDoc, defs, expr, region, "min", 0, 0)
		require.NoError(t, err)
		return r
	}

	// The discovery sweep: three pairwise interferences, three containments.
	if r := minOf(t, "", "max(block, drum)"); r.Min > -0.5 {
		t.Errorf("block/drum Min = %v, want a clear interference (< -0.5)", r.Min)
	}
	if r := minOf(t, "", "max(block, ball)"); r.MinLower <= 0 {
		t.Errorf("block/ball MinLower = %v, want proven clear (> 0)", r.MinLower)
	}
	if r := minOf(t, "", "max(drum, ball)"); r.MinLower <= 0 {
		t.Errorf("drum/ball MinLower = %v, want proven clear (> 0)", r.MinLower)
	}
	if r := minOf(t, "", "max(ball, -crateInner)"); r.Min > -0.5 {
		t.Errorf("ball containment Min = %v, want a clear escape (< -0.5)", r.Min)
	}
	if r := minOf(t, "", "max(block, -crateInner)"); r.Min <= 0 {
		t.Errorf("block containment Min = %v, want contained (> 0)", r.Min)
	}
	if r := minOf(t, "", "max(drum, -crateInner)"); r.Min <= 0 {
		t.Errorf("drum containment Min = %v, want contained (> 0)", r.Min)
	}

	// The repair, tested the way the model should: override the document's
	// own center variables in defs and re-run the same checks.
	fix := "drumCenter = [6.0, -3.0, 6.0]; ballCenter = [8.0, 9.0, 8.0];"
	for _, expr := range []string{
		"max(block, drum)",
		"max(block, ball)",
		"max(drum, ball)",
	} {
		if r := minOf(t, fix, expr); r.Min < 0.5 {
			t.Errorf("repaired %s Min = %v, want clearance >= 1 (min >= 0.5)", expr, r.Min)
		}
	}
	for _, expr := range []string{
		"max(block, -crateInner)",
		"max(drum, -crateInner)",
		"max(ball, -crateInner)",
	} {
		r := minOf(t, fix, expr)
		if r.Min < 0.4 {
			t.Errorf("repaired %s Min = %v, want wall clearance", expr, r.Min)
		}
		if r.MinLower <= 0 {
			t.Errorf("repaired %s MinLower = %v, want proven contained (> 0)", expr, r.MinLower)
		}
	}
}

// TestExtremumRejectsVector rejects non-scalar expressions.
func TestExtremumRejectsVector(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	_, err := e.Extremum(context.Background(), nil, testDoc, "", "[s, s]",
		[6]float64{-1, -1, -1, 1, 1, 1}, "min", 0, 0)
	if err == nil {
		t.Fatal("expected an error for a vector-valued expression")
	}
}
