//go:build cgo

package eval_test

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
	"gradientcontrol.com/omega/core/program"
)

// parseProgram compiles GSL source to a program.Program for eval tests.
func parseProgram(t *testing.T, gslSrc string) *program.Program {
	t.Helper()
	res, err := assembler.Assemble([]string{gslSrc}, assembler.SourceManager(&source.MockManager{}))
	require.NoError(t, err)
	return res.Program
}

func TestEvalGradSphere(t *testing.T) {
	// A unit sphere: |p| - 1. Distance and gradient have closed forms.
	prog := parseProgram(t, `
return length(p) - 1.0;
`)

	xs := []float32{0, 2, 0, 0.5}
	ys := []float32{0, 0, 3, 0}
	zs := []float32{0, 0, 0, 0}

	results, err := eval.EvalGrad(prog, xs, ys, zs)
	require.NoError(t, err)
	require.Equal(t, 1, len(results))

	samples := results[0]
	require.Equal(t, len(xs), len(samples))

	// Distances: |p| - 1
	expected := []float32{-1, 1, 2, -0.5}
	for i, want := range expected {
		if math.Abs(float64(samples[i].Value-want)) > 1e-4 {
			t.Errorf("point %d: value = %v, want %v", i, samples[i].Value, want)
		}
	}

	// Gradient at (2,0,0) is the unit radial direction (1,0,0).
	g := samples[1]
	if math.Abs(float64(g.Dx-1)) > 1e-4 || math.Abs(float64(g.Dy)) > 1e-4 || math.Abs(float64(g.Dz)) > 1e-4 {
		t.Errorf("gradient at (2,0,0) = (%v, %v, %v), want (1, 0, 0)", g.Dx, g.Dy, g.Dz)
	}
	// Gradient at (0,3,0) is (0,1,0).
	g = samples[2]
	if math.Abs(float64(g.Dx)) > 1e-4 || math.Abs(float64(g.Dy-1)) > 1e-4 || math.Abs(float64(g.Dz)) > 1e-4 {
		t.Errorf("gradient at (0,3,0) = (%v, %v, %v), want (0, 1, 0)", g.Dx, g.Dy, g.Dz)
	}
}

func TestEvalGradMultipleReturns(t *testing.T) {
	// Two returns evaluate independently: a unit sphere and a translated copy.
	prog := parseProgram(t, `
a = length(p) - 1.0;
b = length(p - [3.0, 0.0, 0.0]) - 1.0;
return a, b;
`)

	xs := []float32{0, 3}
	ys := []float32{0, 0}
	zs := []float32{0, 0}

	results, err := eval.EvalGrad(prog, xs, ys, zs)
	require.NoError(t, err)
	require.Equal(t, 2, len(results))

	// Return a: -1 at origin, 2 at (3,0,0).
	if math.Abs(float64(results[0][0].Value+1)) > 1e-4 {
		t.Errorf("a at origin = %v, want -1", results[0][0].Value)
	}
	if math.Abs(float64(results[0][1].Value-2)) > 1e-4 {
		t.Errorf("a at (3,0,0) = %v, want 2", results[0][1].Value)
	}
	// Return b (moved to x=3): 2 at origin, -1 at (3,0,0).
	if math.Abs(float64(results[1][0].Value-2)) > 1e-4 {
		t.Errorf("b at origin = %v, want 2", results[1][0].Value)
	}
	if math.Abs(float64(results[1][1].Value+1)) > 1e-4 {
		t.Errorf("b at (3,0,0) = %v, want -1", results[1][1].Value)
	}
}

func TestEvalGradWithSlider(t *testing.T) {
	// A slider-driven sphere: eval uses the slider's current literal value.
	prog := parseProgram(t, `
radius = 2.0;
radius.min = 1.0;
radius.max = 3.0;
return length(p) - radius;
`)

	xs := []float32{4}
	ys := []float32{0}
	zs := []float32{0}

	results, err := eval.EvalGrad(prog, xs, ys, zs)
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	// |(4,0,0)| - 2 = 2
	if math.Abs(float64(results[0][0].Value-2)) > 1e-4 {
		t.Errorf("value = %v, want 2", results[0][0].Value)
	}
}

func TestEvalGradMismatchedSlices(t *testing.T) {
	prog := parseProgram(t, `
return length(p) - 1.0;
`)
	_, err := eval.EvalGrad(prog, []float32{0, 1}, []float32{0}, []float32{0})
	require.Error(t, err)
}
