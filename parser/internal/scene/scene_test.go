package scene_test

import (
	"testing"

	"gradientcontrol.com/omega/core/parser/internal/scene"
	"gradientcontrol.com/omega/core/program"
)

func TestSceneNew(t *testing.T) {
	// Single argument defaults grid to size / 10
	s, err := scene.New(20, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Size != 20 {
		t.Errorf("got size %v, want 20", s.Size)
	}
	if s.Grid != 2.0 {
		t.Errorf("got grid %v, want 2.0", s.Grid)
	}

	// Explicit grid argument
	g := 5.0
	s2, err := scene.New(20, &g, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s2.Size != 20 || s2.Grid != 5.0 {
		t.Errorf("got size=%v grid=%v, want size=20 grid=5.0", s2.Size, s2.Grid)
	}

	// Reject zero or negative size
	if _, err := scene.New(0, nil, nil, nil); err == nil {
		t.Errorf("expected error for size=0")
	}
	if _, err := scene.New(-10, nil, nil, nil); err == nil {
		t.Errorf("expected error for size=-10")
	}

	// Reject zero or negative grid
	negGrid := -1.0
	if _, err := scene.New(20, &negGrid, nil, nil); err == nil {
		t.Errorf("expected error for negative grid")
	}
	zeroGrid := 0.0
	if _, err := scene.New(20, &zeroGrid, nil, nil); err == nil {
		t.Errorf("expected error for zero grid")
	}
}

func TestFromLegacyComments(t *testing.T) {
	// Legacy division count: 20 / 10 = 2.0 mm
	srcDiv := "// Scene Size: 20\n// Scene Grid: 10\nreturn sphere(1);"
	s1 := scene.FromLegacyComments(srcDiv)
	if s1 == nil {
		t.Fatalf("expected scene from legacy comments, got nil")
	}
	if s1.Size != 20 {
		t.Errorf("got size %v, want 20", s1.Size)
	}
	if s1.Grid != 2.0 {
		t.Errorf("got grid %v, want 2.0", s1.Grid)
	}

	// Explicit mm grid
	srcMm := "// Scene Size: 50.5\n// Scene Grid: 5 mm\nreturn sphere(1);"
	s2 := scene.FromLegacyComments(srcMm)
	if s2 == nil {
		t.Fatalf("expected scene from legacy comments, got nil")
	}
	if s2.Size != 50.5 || s2.Grid != 5.0 {
		t.Errorf("got size=%v grid=%v, want 50.5 and 5.0", s2.Size, s2.Grid)
	}

	// Scene Size only, grid omitted defaults to size / 10
	srcSizeOnly := "// Scene Size: 40\nreturn sphere(1);"
	s3 := scene.FromLegacyComments(srcSizeOnly)
	if s3 == nil {
		t.Fatalf("expected scene from legacy comments, got nil")
	}
	if s3.Size != 40 || s3.Grid != 4.0 {
		t.Errorf("got size=%v grid=%v, want 40 and 4.0", s3.Size, s3.Grid)
	}

	// No scene comments
	s4 := scene.FromLegacyComments("return sphere(1);")
	if s4 != nil {
		t.Errorf("expected nil for source without scene comments, got %+v", s4)
	}
}

func TestResolveWorldExtent(t *testing.T) {
	s := &program.Scene{Size: 40, Grid: 4}
	if ext := scene.ResolveWorldExtent(s, 100); ext != 20.0 {
		t.Errorf("got worldExtent %v, want 20.0", ext)
	}

	if ext := scene.ResolveWorldExtent(nil, 50.0); ext != 50.0 {
		t.Errorf("got fallback %v, want 50.0", ext)
	}

	if ext := scene.ResolveWorldExtent(nil, 0); ext != 10.0 {
		t.Errorf("got default %v, want 10.0", ext)
	}
}
