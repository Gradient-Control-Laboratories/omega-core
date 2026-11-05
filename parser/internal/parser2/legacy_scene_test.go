package parser2

import (
	"testing"
)

func TestFromLegacyComments(t *testing.T) {
	// Legacy division count: 20 / 10 = 2.0 mm
	srcDiv := "// Scene Size: 20\n// Scene Grid: 10\nreturn sphere(1);"
	s1 := fromLegacyComments(srcDiv)
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
	s2 := fromLegacyComments(srcMm)
	if s2 == nil {
		t.Fatalf("expected scene from legacy comments, got nil")
	}
	if s2.Size != 50.5 || s2.Grid != 5.0 {
		t.Errorf("got size=%v grid=%v, want 50.5 and 5.0", s2.Size, s2.Grid)
	}

	// Scene Size only, grid omitted defaults to size / 10
	srcSizeOnly := "// Scene Size: 40\nreturn sphere(1);"
	s3 := fromLegacyComments(srcSizeOnly)
	if s3 == nil {
		t.Fatalf("expected scene from legacy comments, got nil")
	}
	if s3.Size != 40 || s3.Grid != 4.0 {
		t.Errorf("got size=%v grid=%v, want 40 and 4.0", s3.Size, s3.Grid)
	}

	// No scene comments
	s4 := fromLegacyComments("return sphere(1);")
	if s4 != nil {
		t.Errorf("expected nil for source without scene comments, got %+v", s4)
	}
}

func TestLegacySceneSizeMillimeters(t *testing.T) {
	for _, size := range []string{"400 mm", "400mm", "4e2 MM", "400"} {
		t.Run(size, func(t *testing.T) {
			got := fromLegacyComments("// Work Order: WO-1\n// Scene Size: " + size + "\n// Scene Grid: 20 mm\nreturn sphere(1);")
			if got == nil || got.Size != 400 || got.Grid != 20 {
				t.Fatalf("scene = %+v, want size 400 and grid 20", got)
			}
		})
	}
	for _, size := range []string{"400 cm", "400 mm trailing", "mm", "-400 mm", "0 mm"} {
		if got := fromLegacyComments("// Scene Size: " + size); got != nil {
			t.Errorf("invalid size %q produced scene %+v", size, got)
		}
	}
}
