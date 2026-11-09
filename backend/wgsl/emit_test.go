package wgsl_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/backend/wgsl"
)

func TestRender_Sphere(t *testing.T) {
	src := `return sphere(10);`
	res, err := assembler.Assemble([]string{src})
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	progressCalled := false
	progressFn := func(stage string, completed, total uint32) {
		progressCalled = true
	}

	renderRes, err := wgsl.Render(src, res.Program, wgsl.Standalone(true), wgsl.WithProgress(progressFn))
	if err != nil {
		t.Fatalf("wgsl.Render failed: %v", err)
	}
	if renderRes == nil {
		t.Fatal("expected non-nil Result")
	}
	if !strings.Contains(renderRes.WGSL, "@compute") && !strings.Contains(renderRes.WGSL, "fn ") {
		t.Errorf("expected WGSL output to contain function or compute shader, got: %s", renderRes.WGSL)
	}
	if !progressCalled {
		t.Error("expected progress function to be called")
	}
}
