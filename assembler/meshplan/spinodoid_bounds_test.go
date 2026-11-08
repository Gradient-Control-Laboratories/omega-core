package meshplan

import (
	"math"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

func requireBounds6(t *testing.T, got []float64, want [6]float64) {
	t.Helper()
	if len(got) != 6 {
		t.Fatalf("bounds = %v, want %v", got, want)
	}
	for i := range got {
		if math.Abs(got[i]-want[i]) > 1e-8 {
			t.Fatalf("bounds = %v, want %v", got, want)
		}
	}
}

func TestComputeSpinodoidBounds_AnalyticalScene(t *testing.T) {
	sc := &program.Scene{Size: 80, Grid: 8}
	got, err := computeSpinodoidBounds(sc, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireBounds6(t, got, [6]float64{-40, -40, -40, 40, 40, 40})
}

func TestComputeSpinodoidBounds_HeaderlessAnalytical(t *testing.T) {
	got, err := computeSpinodoidBounds(nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty bounds for headerless analytical scene, got %v", got)
	}

	prog := &program.Program{
		Primary: &program.Package{
			Scene: &program.Scene{Size: 0, SpinodoidBounds: []float64{-10, -10, -10, 10, 10, 10}},
		},
	}
	if err := SetProgramSpinodoidBounds(prog, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prog.Scene().GetSpinodoidBounds()) != 0 {
		t.Fatalf("expected SpinodoidBounds cleared for non-positive scene size, got %v", prog.Scene().GetSpinodoidBounds())
	}
}

func TestComputeSpinodoidBounds_ImportedMeshAndTransformedPlacement(t *testing.T) {
	bboxes := map[int][6]float64{
		0: {1, 2, 0, 3, 5, 6},
	}
	decls := []MeshDecl{
		{
			Index:  0,
			Matrix: []float64{0, 1, 0, 0, -1, 0, 0, 0, 0, 0, 1, 0, 100, 300, 20, 1},
		},
	}
	prog := &program.Program{Primary: &program.Package{}}
	if err := SetProgramSpinodoidBounds(prog, bboxes, decls); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireBounds6(t, prog.Scene().GetSpinodoidBounds(), [6]float64{-3.85, -13.05, -1.3, 102.85, 318.05, 27.3})
}

func TestComputeSpinodoidBounds_InvalidAndNonFiniteBounds(t *testing.T) {
	// Inverted min > max
	_, err := computeSpinodoidBounds(nil, map[int][6]float64{
		2: {5, 0, 0, 1, 1, 1},
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "mesh slot 2 has invalid bounds") {
		t.Fatalf("expected invalid bounds error for slot 2, got %v", err)
	}

	// Non-finite placement transform
	_, err = computeSpinodoidBounds(nil, map[int][6]float64{
		0: {0, 0, 0, 1, 1, 1},
	}, []MeshDecl{
		{
			Index:  0,
			Matrix: []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, math.NaN(), 0, 0, 1},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "imported geometry has non-finite bounds") {
		t.Fatalf("expected non-finite bounds error, got %v", err)
	}
}
