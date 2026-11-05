package accel_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	commonaccel "gradientcontrol.com/omega/core/backend/wgsl/internal/accel"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/internal/accel"
)

func TestWGSLOps_NoDrift(t *testing.T) {
	pkgDir := filepath.Join("../../../../../..", "stdlib")
	descriptors, err := commonaccel.ValidateAll(pkgDir, "wgsl")
	if err != nil {
		t.Fatalf("validation of package sidecars failed: %v", err)
	}

	generated, err := accel.GenerateOps(descriptors)
	if err != nil {
		t.Fatalf("failed to generate WGSL ops: %v", err)
	}

	committed, err := os.ReadFile("ops_gen.go")
	if err != nil {
		t.Fatalf("failed to read committed ops_gen.go: %v", err)
	}

	if !bytes.Equal(committed, generated) {
		t.Errorf("drift detected: committed ops_gen.go differs from generated code. Run 'go generate ./agent/wgsl/internal/accel/...' to regenerate.")
	}
}
