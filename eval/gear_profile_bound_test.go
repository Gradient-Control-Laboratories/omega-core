//go:build cgo

package eval

import (
	"fmt"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// gearProfileCase is one gear's profile.
type gearProfileCase struct {
	teeth, module, pressureDeg, clearance, tipBreak float64
}

// oldGearProfile is gearProfile as it was before its flanks were bounded
// (85a0152), built from the still-public involuteRadial and intersectChamfer.
// Its tip chamfer is given the bevel of the new one's, which is kSmall deep:
// intersectChamfer's k for a depth is the depth times |n1 + n2|, the length of
// the sum of the flank's and the tip circle's normals at the edge.
const oldGearProfile = `
oldProfile(numTeeth, moduleVal, pressureAngle, rotOffset, dedendumClearance, kSmall) {
  pitchRadius = (numTeeth / 2.0) * moduleVal;
  baseRadius = pitchRadius * cos(pressureAngle);
  addendumRadius = pitchRadius + moduleVal;
  dedendumRadius = pitchRadius - (1 + dedendumClearance) * moduleVal;
  roundK = dedendumClearance * moduleVal;
  pitchAngle = pi / numTeeth;
  halfPitchAngle = pitchAngle / 2;
  prRatio = pitchRadius / baseRadius;
  gammaP = sqrt(prRatio * prRatio - 1);
  contactAngle = gammaP - atan(gammaP);
  rotAngle = halfPitchAngle - contactAngle;
  cosR = cos(rotAngle);
  sinR = sin(rotAngle);
  rotInv() {
    rx = p.x * cosR + p.y * sinR;
    ry = neg(p.x) * sinR + p.y * cosR;
    return remap(gears.involuteRadial(baseRadius), [rx, ry, p.z]);
  }
  reflInv() {
    return remap(rotInv(), [p.x, neg(p.y), p.z]);
  }
  singleTooth() {
    return min(rotInv(), reflInv());
  }
  cosO = cos(rotOffset);
  sinO = sin(rotOffset);
  opx = p.x * cosO + p.y * sinO;
  opy = neg(p.x) * sinO + p.y * cosO;
  pRho = length([opx, opy]);
  pAngle = atan(opy, opx);
  pAngleShifted = pAngle + pitchAngle;
  twicePitchAngle = pitchAngle * 2;
  pAngleMod = ((pAngleShifted % twicePitchAngle) + twicePitchAngle) % twicePitchAngle;
  pAngleCentered = pAngleMod - pitchAngle;
  rpx = pRho * cos(pAngleCentered);
  rpy = pRho * sin(pAngleCentered);
  repeatedTooth = remap(singleTooth(), [rpx, rpy, p.z]);
  addClip = remap(circleXY([0, 0], addendumRadius), [opx, opy, p.z]);
  tipRoll = sqrt(addendumRadius * addendumRadius - baseRadius * baseRadius) / baseRadius;
  tipCorner = [cos(tipRoll) + tipRoll * sin(tipRoll), sin(tipRoll) - tipRoll * cos(tipRoll)];
  tipNormalSum = length([sin(tipRoll), neg(cos(tipRoll))] + tipCorner / length(tipCorner));
  toothProfile = intersectChamfer(repeatedTooth, addClip, kSmall * tipNormalSum);
  dedBody = remap(circleXY([0, 0], dedendumRadius), [opx, opy, p.z]);
  return unionRound(toothProfile, dedBody, roundK);
}
`

// Bounding the flanks moves no surface: the new profile and the old agree in
// sign at every point, on a fine grid over the teeth and a finer one over a
// tooth's tip, but for the tip chamfer. That is now the straight line between
// the two points where the old bevel, (tooth + tip + setback) / c, met the two
// faces; the old bevel was the level set of that sum, slightly curved, so
// within the chamfer's band the two may differ by under a hundredth of the
// chamfer. (A flank stopped at the tip corner once made the old bevel bend
// round the corner point: an extra surface off the involute side of the
// chamfer, the owner's, in the viewer. A flank run on past the tip, as here,
// does not.)
func TestGearProfile_SameSurface(t *testing.T) {
	cases := map[string]gearProfileCase{
		"17 teeth, as drawn":     {17, 1, 20, 0.25, 0.05},
		"17 teeth, sharp":        {17, 1, 20, 0.0, 0.0},
		"12 teeth, 25 degrees":   {12, 2, 25, 0.25, 0.1},
		"40 teeth, rotated":      {40, 0.5, 20, 0.25, 0.025},
		"the gears example's 20": {20, 4, 0.26 * 180 / math.Pi, 0.25, 0.25},
	}
	for name, g := range cases {
		t.Run(name, func(t *testing.T) {
			args := fmt.Sprintf("%v, %v, radians(%v), 0.0, %v, %v", g.teeth, g.module, g.pressureDeg, g.clearance, g.tipBreak)
			doc := "@scene(140)\n" + oldGearProfile + "newer = gears.gearProfile(" + args + ");\nolder = oldProfile(" + args + ");\nreturn newer, older;\n"
			tip := g.teeth*g.module/2 + g.module
			var pts [][3]float64
			for x := -tip - 1; x <= tip+1; x += tip / 150 {
				for y := -tip - 1; y <= tip+1; y += tip / 150 {
					pts = append(pts, [3]float64{x + 0.00137, y + 0.00291, 0})
				}
			}
			// One pitch of the tip band, a hundredth of a module apart.
			for r := tip - 2*g.module; r <= tip+g.module; r += g.module / 100 {
				for a := 0.0; a <= 2*math.Pi/g.teeth; a += g.module / 100 / r {
					pts = append(pts, [3]float64{r * math.Cos(a), r * math.Sin(a), 0})
				}
			}
			e := New(&source.MockManager{}, nil)
			groups, err := e.EvalGsl(t.Context(), nil, doc, "", []string{"newer", "older"}, pts)
			if err != nil {
				t.Fatalf("eval: %v", err)
			}
			differ, chamferDiffer, chamferWorst := 0, 0, 0.0
			for i, q := range pts {
				n, o := float64(groups[0].Samples[0][i].Value), float64(groups[1].Samples[0][i].Value)
				if math.Abs(o) <= 1e-4 || (n > 0) == (o > 0) {
					continue
				}
				if g.tipBreak > 0 && math.Abs(math.Hypot(q[0], q[1])-tip) < 2*g.tipBreak && math.Max(math.Abs(n), math.Abs(o)) < g.tipBreak/100 {
					chamferDiffer++
					chamferWorst = math.Max(chamferWorst, math.Max(math.Abs(n), math.Abs(o)))
					continue
				}
				if differ++; differ <= 3 {
					t.Errorf("at %v (rho %.4f): new %v, old %v: the surface moved", q, math.Hypot(q[0], q[1]), n, o)
				}
			}
			t.Logf("%d points, %d on the other side; at the chamfer %d, within %.5f mm", len(pts), differ, chamferDiffer, chamferWorst)
		})
	}
}
