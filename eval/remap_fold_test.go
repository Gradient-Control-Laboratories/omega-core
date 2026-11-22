//go:build cgo

package eval

import (
	"context"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// Shaft Modeling, Phase 19: a remap to a constant point folds. remap(field,
// [x, y, z]) with constant coordinates is the field at one point, a number
// whenever the field reads nothing but p and constants, so the assembler
// writes the number; drill.hole's size probe is one, and costs nothing.

// The probe, the unit ball's field at (0, 0, 5), is 4: no remap is left in
// the shader, and the CPU reads the same field.
func TestFold_RemapToAConstantPoint(t *testing.T) {
	doc := "probe = remap(length(p) - 1.0, [0.0, 0.0, 5.0]);\nball = length(p) - 2.0 + probe - 4.0;\nreturn ball;\n"
	code, err := pruneShader(t, doc)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(code, "gsl_remap") {
		t.Errorf("the remap to a constant point was not folded")
	}
	e := New(&source.MockManager{}, nil)
	pts := [][3]float64{{0, 0, 1.5}, {1, 0, 0}}
	g, err := e.EvalGsl(context.Background(), nil, doc, "", []string{"ball"}, pts)
	if err != nil {
		t.Fatal(err)
	}
	for i, p := range pts {
		want := math.Sqrt(p[0]*p[0]+p[1]*p[1]+p[2]*p[2]) - 2
		if got := float64(g[0].Samples[0][i].Value); math.Abs(got-want) > 1e-5 {
			t.Errorf("at %v: %g, want %g", p, got, want)
		}
	}
}

// A drilled 3 mm thru hole's feature size is the number 3, folded from its
// probe.
func TestFeaturePrune_DrillProbeFolds(t *testing.T) {
	code, err := pruneShader(t, "plate = box([-20.0, -20.0, -5.0], [20.0, 20.0, 5.0]);\nplated = max(plate, neg(drill.hole(drill.thru(3.0), 5.0)));\nreturn plated;\n")
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`let gsl_featureSize = (-?[0-9.e+-]+);`).FindStringSubmatch(code)
	if m == nil {
		t.Fatalf("the drill's feature size is not a number:\n%s", code)
	}
	if v, _ := strconv.ParseFloat(m[1], 64); math.Abs(v-3) > 1e-4 {
		t.Errorf("the drill's feature size is %g, want 3", v)
	}
}
