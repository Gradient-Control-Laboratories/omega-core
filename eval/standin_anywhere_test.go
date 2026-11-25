//go:build cgo

package eval

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
)

// Shaft Modeling, Phase 22: stand-ins anywhere. A shape's stand-in,
// name.simple = proxy, keeps the viewer's switch (each side evaluated only
// where its weight needs it, the Simple rung, the blend, the stand-in's own
// colour, the crop to a known box), but the weight is measured at the point,
// from the same footprint features use, and the stand-in may sit on any value,
// not only a part. Its size is name.featureSize where declared, else the
// value's box, else it never switches; the fastener library declares its own,
// so a screw placed through a remap keeps its stand-in.

// A fastener placed through a remap keeps the stand-in its library declared.
func TestStandIn_SurvivesRemap(t *testing.T) {
	code, err := pruneShader(t, "screw = remap(fastener.shcs(6.0, 20.0), [p.x - 30.0, p.y, p.z]);\nreturn screw;\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(code, "gsl_simpleRung") {
		t.Errorf("the remapped screw has no stand-in switch")
	}
}

// knob, read by an intersection, is not a part; its stand-in is allowed, a
// substitution needing no identity, and sized by knob's box.
const knobCutDoc = `knob = boxCentered([0, 0, 0], [20, 20, 40]) - 2;
knob.simple = boxCentered([0, 0, 0], [24, 24, 44]);
cut = max(knob, p.z - 10.0);
return cut;
`

func TestStandIn_OnAValueThatIsNotAPart(t *testing.T) {
	code, err := pruneShader(t, knobCutDoc)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(code, "gsl_simpleRung") {
		t.Errorf("the stand-in on a value read by an intersection has no switch")
	}
}

// A value whose box is unbounded switches at its declared size.
func TestStandIn_DeclaredSize(t *testing.T) {
	doc := "bar = max(abs(p.x) - 5.0, abs(p.y) - 5.0);\nbar.simple = length(p.xy) - 5.0;\nbar.featureSize = 10.0;\nreturn bar;\n"
	code, err := pruneShader(t, doc)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(code, "gsl_simpleRung") {
		t.Errorf("the declared size gave no switch")
	}
}

// Unbounded and undeclared, the value has no size and never switches: no
// error, no switch.
func TestStandIn_NoSizeNoSwitch(t *testing.T) {
	doc := "bar = max(abs(p.x) - 5.0, abs(p.y) - 5.0);\nbar.simple = length(p.xy) - 5.0;\nreturn bar;\n"
	code, err := pruneShader(t, doc)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(code, "gsl_simpleRung") {
		t.Errorf("a stand-in with no size switched")
	}
}

// The weight is measured at the point: the switch reads the feature pixel,
// set as sdShape is entered, not the eye's distance to the part's box.
func TestStandIn_WeightAtThePoint(t *testing.T) {
	code, err := pruneShader(t, "bolt = boxCentered([0, 0, 0], [2, 2, 4]) - 0.2;\nbolt.simple = sphere(2);\nreturn bolt;\n")
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(code, "fn gsl_simple(")
	if start < 0 {
		t.Fatal("no gsl_simple")
	}
	body := code[start:]
	body = body[:strings.Index(body, "\n}\n")]
	if !strings.Contains(body, "gsl_featurePixel") || strings.Contains(body, "gsl_vanish(") {
		t.Errorf("the stand-in's weight is not measured at the point:\n%s", body)
	}
	if !strings.Contains(standInDocPart(t, code), "gsl_featurePixel = ") {
		t.Errorf("sdShape does not measure the pixel for the stand-in")
	}
}

func standInDocPart(t *testing.T, code string) string {
	t.Helper()
	start := strings.Index(code, "fn sdShape(")
	if start < 0 {
		t.Fatal("no sdShape")
	}
	end := strings.Index(code[start:], "\n}\n")
	if end < 0 {
		t.Fatal("sdShape has no end")
	}
	return code[:start+end]
}

// Everywhere but the viewer the value is itself.
func TestStandIn_CPUFieldUnchanged(t *testing.T) {
	e := New(&source.MockManager{}, nil)
	pts := [][3]float64{{0, 0, 0}, {11, 0, 0}, {0, 0, 15}}
	g, err := e.EvalGsl(context.Background(), nil, knobCutDoc, "", []string{"cut"}, pts)
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range []float64{-10, -1, 5} {
		if got := float64(g[0].Samples[0][i].Value); math.Abs(got-want) > 1e-4 {
			t.Errorf("at %v: %g, want %g", pts[i], got, want)
		}
	}
}
