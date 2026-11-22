package codegen_test

import (
	"os"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
)

// A part's box declares what its corners repeat once (property's
// shareBoxRepeats), so the cylinder example's profile bounds, a min and a max
// over every slider point read again in each corner, are written once each.
// Measured 2026-11-18: 190,178 bytes of WGSL without the sharing, 82,569 with.
func TestBoxShare_CylinderExampleWritesItsBoundsOnce(t *testing.T) {
	src, err := os.ReadFile("../../../../examples/profile_cylinder.gsl")
	if err != nil {
		t.Fatal(err)
	}
	res, err := assembler.Assemble([]string{string(src)}, assembler.SourceManager(mockMgr))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	out, err := renderHelper(string(src), res.Program)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	const limit = 100000
	if !strings.Contains(out.WGSL, "let gslBox") {
		t.Error("the part's box declares nothing it repeats")
	}
	if len(out.WGSL) > limit {
		t.Errorf("WGSL is %d bytes, want at most %d", len(out.WGSL), limit)
	}
	t.Logf("WGSL %d bytes", len(out.WGSL))
}
