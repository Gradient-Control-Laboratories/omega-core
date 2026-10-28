package assembler_test

import (
	"testing"

	"gradientcontrol.com/omega/core/assembler"
)

func TestAssemble_EmptySources(t *testing.T) {
	_, err := assembler.Assemble(nil)
	if err == nil {
		t.Fatal("expected error for empty sources, got nil")
	}
}

func TestAssemble_SimpleGSL(t *testing.T) {
	src := `x = 10;
return x;`
	res, err := assembler.Assemble([]string{src})
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	if res == nil || res.Program == nil {
		t.Fatal("expected non-nil result and program")
	}
}
