package scene_test

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
	"gradientcontrol.com/omega/core/program/scene"
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
