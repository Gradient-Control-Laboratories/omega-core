package tape_test

import (
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/backend/tape"
)

func TestCompile_Sphere(t *testing.T) {
	src := `return sphere(10);`
	res, err := assembler.Assemble([]string{src})
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	progressCalled := false
	progressFn := func(stage string, completed, total uint32) {
		progressCalled = true
	}

	tapeBytes, err := tape.Compile(res.Program, tape.WithProgress(progressFn))
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	if len(tapeBytes) == 0 {
		t.Fatal("expected non-empty tape bytes")
	}
	if !progressCalled {
		t.Error("expected progress callback to be invoked")
	}
}
