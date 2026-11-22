//go:build cgo

package eval

import (
	"context"
	"math"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// Shaft Modeling, Phase 19: feature pruning. A small feature gives way when it
// is small on screen though its part is not. A feature opts in with
// name.featureSize = size (mm; 0 asks for its box's smallest extent), and is
// then builtin pruneBelow(field, size): the field everywhere but the viewer's
// WGSL, which measures the pixel at the point from the world position and the
// eye (gsl_featurePixel, set as sdShape is entered) and drops the feature
// once it is too small to see. Only a union operand or a subtracted one may be
// pruned, where nothing is the identity. gears.gearProfile is re-authored as
// its pitch disc with the tips added and the gaps taken away, so pruned teeth
// leave the pitch cylinder.

func pruneShader(t *testing.T, doc string) (string, error) {
	t.Helper()
	e := New(&source.MockManager{}, nil)
	res, err := e.Shader(t.Context(), nil, doc, "", ShaderOptions{Target: "wgsl", OutDir: t.TempDir()})
	if err != nil {
		return "", err
	}
	return res.Code, nil
}

// A plate 40 x 40 x 10 with a hole 1 across through it, the hole a feature
// when featured is set.
func plateDoc(featured bool) string {
	doc := "plate = box([-20.0, -20.0, -5.0], [20.0, 20.0, 5.0]);\nhole = length(p.xy) - 0.5;\n"
	if featured {
		doc += "hole.featureSize = 1.0;\n"
	}
	return doc + "plated = max(plate, neg(hole));\nreturn plated;\n"
}

// Only a feature that opts in is pruned: a document without one compiles with
// no footprint at all.
func TestFeaturePrune_OptIn(t *testing.T) {
	plain, err := pruneShader(t, plateDoc(false))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(plain, "gsl_featurePixel") {
		t.Errorf("a document with no feature measures the feature pixel")
	}
	featured, err := pruneShader(t, plateDoc(true))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(featured, "gsl_featurePixel") {
		t.Errorf("a featured hole is not guarded by the feature pixel")
	}
}

// The footprint is measured once, as sdShape is entered, from the world
// position and the eye, before anything reads it.
func TestFeaturePrune_FootprintAtEntry(t *testing.T) {
	code, err := pruneShader(t, plateDoc(true))
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(code, "fn sdShape(")
	if start < 0 {
		t.Fatal("no sdShape")
	}
	body := code[start:]
	body = body[:strings.Index(body, "\n}\n")]
	at := strings.Index(body, "gsl_featurePixel")
	if at < 0 {
		t.Fatal("sdShape does not set gsl_featurePixel")
	}
	line := body[at:]
	line = line[:strings.IndexByte(line, '\n')]
	if !strings.HasPrefix(line, "gsl_featurePixel = ") || !strings.Contains(line, "gsl_vanishEye") || !strings.Contains(line, "p") {
		t.Errorf("sdShape's first use of the feature pixel is not its measure from p and the eye: %s", line)
	}
}

// A feature intersected with something is an error: pruned to nothing, it
// would take the whole intersection with it. The error names the property
// and where a feature may stand, a union or a difference.
func TestFeaturePrune_IntersectionIsAnError(t *testing.T) {
	doc := "a = length(p) - 5.0;\nb = length(p.xy) - 1.0;\nb.featureSize = 1.0;\nc = max(a, b);\nreturn c;\n"
	_, err := pruneShader(t, doc)
	if err == nil || !strings.Contains(err.Error(), "featureSize") || !strings.Contains(err.Error(), "union") {
		t.Errorf("a featured operand of an intersection compiled, or failed for another reason (err %v)", err)
	}
}

// A union operand and a subtracted one are both allowed.
func TestFeaturePrune_UnionAndDifferenceAllowed(t *testing.T) {
	for _, combine := range []string{"min(a, b)", "max(a, neg(b))"} {
		doc := "a = length(p) - 5.0;\nb = length(p.xy) - 1.0;\nb.featureSize = 1.0;\nc = " + combine + ";\nreturn c;\n"
		if _, err := pruneShader(t, doc); err != nil {
			t.Errorf("%s: %v", combine, err)
		}
	}
}

// Everywhere but the viewer the feature is its field: the CPU reads the hole.
func TestFeaturePrune_CPUFieldUnchanged(t *testing.T) {
	e := New(&source.MockManager{}, nil)
	pts := [][3]float64{{0, 0, 0}, {0.25, 0, 1}, {3, 2, -1}}
	for _, featured := range []bool{false, true} {
		g, err := e.EvalGsl(context.Background(), nil, plateDoc(featured), "", []string{"plated"}, pts)
		if err != nil {
			t.Fatal(err)
		}
		if got := float64(g[0].Samples[0][0].Value); math.Abs(got-0.5) > 1e-4 {
			t.Errorf("featured %v: the field in the hole is %g, want 0.5", featured, got)
		}
		if got := float64(g[0].Samples[0][2].Value); got > -1 {
			t.Errorf("featured %v: the plate away from the hole reads %g", featured, got)
		}
	}
}

// The library's features opt in: a drilled hole and a gear's teeth.
func TestFeaturePrune_LibraryFeaturesOptIn(t *testing.T) {
	for name, doc := range map[string]string{
		"drill": "plate = box([-20.0, -20.0, -5.0], [20.0, 20.0, 5.0]);\nplated = max(plate, neg(drill.hole(drill.thru(3.0), 5.0)));\nreturn plated;\n",
		"gear":  "g = shaft.body([shaft.gear(60.0, 2.0, 0.3490658503988659, 0.0, 0.0, 20.0)]);\nreturn g;\n",
	} {
		code, err := pruneShader(t, doc)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if !strings.Contains(code, "gsl_featurePixel") {
			t.Errorf("%s: the library feature is not guarded by the feature pixel", name)
		}
	}
}

// gears.gearProfile as it was before Phase 19, frozen as the reference the
// re-authored profile must keep: the same solid.
const oldGearProfileDoc = `
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
    left = rotInv();
    right = reflInv();
    return min(left, right);
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
  toothProfile = intersectChamferNormalized(repeatedTooth, addClip, kSmall,
                                            [sin(tipRoll), neg(cos(tipRoll))], tipCorner / length(tipCorner));
  dedBody = remap(circleXY([0, 0], dedendumRadius), [opx, opy, p.z]);
  return unionRound(toothProfile, dedBody, roundK);
}
newP = gears.gearProfile(60.0, 2.0, 0.3490658503988659, 0.0, 0.25, 0.1);
oldP = oldProfile(60.0, 2.0, 0.3490658503988659, 0.0, 0.25, 0.1);
return newP;
`

// profilePoints samples two teeth's worth of the 60-tooth module-2 profile,
// from below its root to above its tip.
func profilePoints() [][3]float64 {
	var pts [][3]float64
	for i := 0; i <= 80; i++ {
		r := 56.0 + 7.0*float64(i)/80
		for j := 0; j <= 120; j++ {
			a := -2 * math.Pi / 60 * float64(j-60) / 60
			pts = append(pts, [3]float64{r * math.Cos(a), r * math.Sin(a), 0})
		}
	}
	return pts
}

// The re-authored profile is the same solid: wherever the old one is clearly
// inside or outside, so is the new one.
func TestGearProfile_SameSolid(t *testing.T) {
	e := New(&source.MockManager{}, nil)
	pts := profilePoints()
	g, err := e.EvalGsl(context.Background(), nil, oldGearProfileDoc, "", []string{"newP", "oldP"}, pts)
	if err != nil {
		t.Fatal(err)
	}
	bad := 0
	for i := range pts {
		n, o := float64(g[0].Samples[0][i].Value), float64(g[1].Samples[0][i].Value)
		if math.Abs(o) > 0.04 && (n < 0) != (o < 0) {
			if bad < 5 {
				t.Errorf("at %.3f, %.3f: new %g, old %g", pts[i][0], pts[i][1], n, o)
			}
			bad++
		}
	}
	if bad > 0 {
		t.Errorf("%d of %d points disagree on inside or outside", bad, len(pts))
	}
}

// Across the pitch circle inside a tooth, where the disc and the tips meet,
// the new profile is as deep as the old one's tooth is there, at least half:
// no sheet of zero, nor a shallow one, lies inside the solid.
func TestGearProfile_NoShallowSheetAtPitch(t *testing.T) {
	e := New(&source.MockManager{}, nil)
	var pts [][3]float64
	for j := 0; j <= 200; j++ {
		a := -2 * math.Pi / 60 * float64(j-100) / 100
		pts = append(pts, [3]float64{60 * math.Cos(a), 60 * math.Sin(a), 0})
	}
	g, err := e.EvalGsl(context.Background(), nil, oldGearProfileDoc, "", []string{"newP", "oldP"}, pts)
	if err != nil {
		t.Fatal(err)
	}
	deepest := 0
	for i := range pts {
		if g[1].Samples[0][i].Value < g[1].Samples[0][deepest].Value {
			deepest = i
		}
	}
	n, o := float64(g[0].Samples[0][deepest].Value), float64(g[1].Samples[0][deepest].Value)
	if o > -0.5 {
		t.Fatalf("the reference tooth is only %g deep at the pitch circle", o)
	}
	if n > 0.5*o {
		t.Errorf("at the tooth's middle on the pitch circle the new profile reads %g, the old %g", n, o)
	}
}
