package accel_test

import (
	"os"
	"path/filepath"
	"testing"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/accel"
)

func TestValidateSidecar_RealFiles(t *testing.T) {
	matrixYAML := filepath.Join("..", "package", "matrix", "matrix.wgsl.yaml")
	if _, err := accel.ValidateSidecar(matrixYAML); err != nil {
		t.Fatalf("ValidateSidecar(%s) failed: %v", matrixYAML, err)
	}

	vectorYAML := filepath.Join("..", "package", "vector", "vector.wgsl.yaml")
	if _, err := accel.ValidateSidecar(vectorYAML); err != nil {
		t.Fatalf("ValidateSidecar(%s) failed: %v", vectorYAML, err)
	}
}

func TestValidateAll_RealPackages(t *testing.T) {
	pkgRoot := filepath.Join("..", "package")
	descriptors, err := accel.ValidateAll(pkgRoot, "wgsl")
	if err != nil {
		t.Fatalf("ValidateAll(%s, \"wgsl\") failed: %v", pkgRoot, err)
	}
	if len(descriptors) < 2 {
		t.Fatalf("expected at least 2 descriptors (matrix and vector), got %d", len(descriptors))
	}
}

func TestSiblingGSLPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/path/to/matrix.wgsl.yaml", "/path/to/matrix.gsl"},
		{"/path/to/vector.js.yaml", "/path/to/vector.gsl"},
		{"/path/to/simple.yaml", "/path/to/simple.gsl"},
	}

	for _, tt := range tests {
		actual := accel.SiblingGSLPath(tt.input)
		if actual != tt.expected {
			t.Errorf("SiblingGSLPath(%q) = %q, expected %q", tt.input, actual, tt.expected)
		}
	}
}

func TestValidateDescriptor_PackageMismatch(t *testing.T) {
	desc := &accel.PackageDescriptor{
		Package: "wrong_package",
		Accelerations: []accel.Acceleration{
			{
				Name: "test_fn",
				Kind: "operator",
				Signatures: []accel.Signature{
					{Args: []string{"a"}, Return: "a"},
				},
			},
		},
	}
	gslSource := "test_fn(a) { return a; }"

	err := accel.ValidateDescriptor(desc, gslSource, "correct_package")
	if err == nil {
		t.Fatal("expected error for package mismatch, got nil")
	}
}

func TestValidateDescriptor_FunctionNotFound(t *testing.T) {
	desc := &accel.PackageDescriptor{
		Package: "testpkg",
		Accelerations: []accel.Acceleration{
			{
				Name: "missing_fn",
				Kind: "operator",
				Signatures: []accel.Signature{
					{Args: []string{"a"}, Return: "a"},
				},
			},
		},
	}
	gslSource := "existing_fn(a) { return a; }"

	err := accel.ValidateDescriptor(desc, gslSource, "testpkg")
	if err == nil {
		t.Fatal("expected error for non-existent function, got nil")
	}
}

func TestValidateDescriptor_ArityMismatch(t *testing.T) {
	desc := &accel.PackageDescriptor{
		Package: "testpkg",
		Accelerations: []accel.Acceleration{
			{
				Name: "my_fn",
				Kind: "operator",
				Signatures: []accel.Signature{
					{Args: []string{"a", "b", "c"}, Return: "a"},
				},
			},
		},
	}
	// my_fn has only 2 parameters in GSL
	gslSource := "my_fn(a, b) { return a; }"

	err := accel.ValidateDescriptor(desc, gslSource, "testpkg")
	if err == nil {
		t.Fatal("expected error for arity mismatch, got nil")
	}
}

func TestValidateSidecar_MissingSiblingGSL(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "orphan.wgsl.yaml")
	content := []byte("package: orphan\naccelerations: []\n")
	if err := os.WriteFile(yamlPath, content, 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	_, err := accel.ValidateSidecar(yamlPath)
	if err == nil {
		t.Fatal("expected error for missing sibling .gsl file, got nil")
	}
}
