package scene_test

import (
	"testing"

	"gradientcontrol.com/omega/core/parser/internal/scene"
)

func TestLegacySceneSizeMillimeters(t *testing.T) {
	for _, size := range []string{"400 mm", "400mm", "4e2 MM", "400"} {
		t.Run(size, func(t *testing.T) {
			got := scene.FromLegacyComments("// Work Order: WO-1\n// Scene Size: " + size + "\n// Scene Grid: 20 mm\nreturn sphere(1);")
			if got == nil || got.Size != 400 || got.Grid != 20 {
				t.Fatalf("scene = %+v, want size 400 and grid 20", got)
			}
		})
	}
	for _, size := range []string{"400 cm", "400 mm trailing", "mm", "-400 mm", "0 mm"} {
		if got := scene.FromLegacyComments("// Scene Size: " + size); got != nil {
			t.Errorf("invalid size %q produced scene %+v", size, got)
		}
	}
}
