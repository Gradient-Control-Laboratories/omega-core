package codegen_test

import (
	"testing"

	"gradientcontrol.com/omega/core/backend/wgsl"
	"gradientcontrol.com/omega/core/assembler"
)

// TestPartsReported checks the render result lists each guarded part by its
// box, for the viewer's Detail slider to measure and draw, and leaves out a
// part whose box moves, with time here.
func TestPartsReported(t *testing.T) {
	const src = `
bolt = boxCentered([0, 0, 0], [2, 2, 4]);
plate = box([-10, -10, -3], [10, 10, -2]);
cap = boxCentered([0, 0, 3 + sin(time)], [1, 1, 1]);
return min(min(bolt, plate), cap);
`
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mockMgr))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	out, err := renderHelper(src, res.Program)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	want := []wgsl.Part{
		{Lo: [3]float64{-1, -1, -2}, Hi: [3]float64{1, 1, 2}},
		{Lo: [3]float64{-10, -10, -3}, Hi: [3]float64{10, 10, -2}},
	}
	if len(out.Parts) != len(want) {
		t.Fatalf("parts %v, want %v", out.Parts, want)
	}
	for i := range want {
		if out.Parts[i] != want[i] {
			t.Errorf("part %d %v, want %v", i, out.Parts[i], want[i])
		}
	}
}
