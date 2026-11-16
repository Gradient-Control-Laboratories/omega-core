//go:build cgo

package eval_test

import (
	"context"
	"testing"

	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// A five-lobed knob drawn in one lobe's wedge, on a hub: the angle is folded with atan and a remainder, whose
// intervals span the whole turn across the angle's seam, so no box that touches the seam can be ruled out. The
// part is 50 across its tips. The search ends on leaves it cannot drop, with budget to spare, 9 mm off it in x.
const foldedKnobDoc = `@scene(64)
wedge = tau / 5;
rad = length(p.xy);
fold = abs((atan(p.y, p.x) + tau + wedge / 2) % wedge - wedge / 2);
q = [rad * cos(fold), rad * sin(fold)];
outline = min(length(q - [16.0, 0.0]) - 9.0, rad - 17.0);
slab = abs(p.z - 10.0) - 10.0;
body = intersectRound(outline, slab, 3.0);
hub = max(rad - 10.0, -12.0 - p.z, p.z - 1.0);
knob = max(min(body, hub), 4.2 - rad);
return knob;
`

// A box the search could not bring near the part is not answered as the part's box, however the search ended:
// it is the scan's, unproven, and the reason says the intervals were too loose.
func TestEnclose_ABoxLeftLooseIsNotProven(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	r, err := e.Enclose(context.Background(), nil, foldedKnobDoc, "", "knob", [6]float64{}, 0)
	require.NoError(t, err)
	require.False(t, r.Proven, "proven with the box %v mm off the part", r.CellSize)
	require.Equal(t, eval.ReasonLoose, r.Reason)
	require.True(t, r.Found)
}
