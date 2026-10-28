package accel

import (
	"strings"
	"testing"
)

func TestParseYAML(t *testing.T) {
	yamlContent := `
package: testpkg
accelerations:
  - name: mul
    kind: operator
    symbol: "*"
    precedence: Multiplication
    signatures:
      - args: [mat, mat]
        return: mat
  - name: neg
    kind: operator
    symbol: "-"
    precedence: Unary
    unary: true
    signatures:
      - args: [mat]
        return: mat
  - name: trans
    kind: builtin
    builtin: transpose
    signatures:
      - args: [mat2x2]
        return: mat2x2
`

	desc, err := ParseYAML([]byte(yamlContent))
	if err != nil {
		t.Fatalf("unexpected error parsing yaml: %v", err)
	}

	if desc.Package != "testpkg" {
		t.Errorf("expected package 'testpkg', got '%s'", desc.Package)
	}
	if len(desc.Accelerations) != 3 {
		t.Fatalf("expected 3 accelerations, got %d", len(desc.Accelerations))
	}

	// 1: Operator
	a0 := desc.Accelerations[0]
	if a0.Name != "mul" || a0.Kind != "operator" || a0.Symbol != "*" || a0.Precedence != "Multiplication" {
		t.Errorf("unexpected acceleration 0: %+v", a0)
	}
	if len(a0.Signatures) != 1 || len(a0.Signatures[0].Args) != 2 || a0.Signatures[0].Return != "mat" {
		t.Errorf("unexpected signature 0: %+v", a0.Signatures)
	}

	// 2: Unary Operator
	a1 := desc.Accelerations[1]
	if a1.Name != "neg" || !a1.Unary {
		t.Errorf("unexpected unary acceleration 1: %+v", a1)
	}

	// 3: Builtin
	a2 := desc.Accelerations[2]
	if a2.Name != "trans" || a2.Kind != "builtin" || a2.Builtin != "transpose" {
		t.Errorf("unexpected builtin acceleration 2: %+v", a2)
	}
}

func TestParseYAMLErrors(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name:    "empty package",
			yaml:    "accelerations: []",
			wantErr: "missing package name",
		},
		{
			name: "missing acceleration name",
			yaml: `
package: testpkg
accelerations:
  - kind: operator
    symbol: "*"
`,
			wantErr: "missing acceleration name",
		},
		{
			name: "unknown kind",
			yaml: `
package: testpkg
accelerations:
  - name: foo
    kind: magic
`,
			wantErr: "invalid kind",
		},
		{
			name: "operator missing symbol",
			yaml: `
package: testpkg
accelerations:
  - name: foo
    kind: operator
`,
			wantErr: "operator missing symbol",
		},
		{
			name: "builtin missing builtin name",
			yaml: `
package: testpkg
accelerations:
  - name: foo
    kind: builtin
`,
			wantErr: "builtin missing builtin name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseYAML([]byte(tt.yaml))
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestParseCommittedSidecars(t *testing.T) {
	matrixPath := "../package/matrix/matrix.wgsl.yaml"
	mDesc, err := ParseFile(matrixPath)
	if err != nil {
		t.Fatalf("failed to parse %s: %v", matrixPath, err)
	}
	if mDesc.Package != "matrix" {
		t.Errorf("expected package 'matrix', got %q", mDesc.Package)
	}
	if len(mDesc.Accelerations) != 11 {
		t.Errorf("expected 11 accelerations in matrix.wgsl.yaml, got %d", len(mDesc.Accelerations))
	}

	vectorPath := "../package/vector/vector.wgsl.yaml"
	vDesc, err := ParseFile(vectorPath)
	if err != nil {
		t.Fatalf("failed to parse %s: %v", vectorPath, err)
	}
	if vDesc.Package != "vector" {
		t.Errorf("expected package 'vector', got %q", vDesc.Package)
	}
	if len(vDesc.Accelerations) != 12 {
		t.Errorf("expected 12 accelerations in vector.wgsl.yaml, got %d", len(vDesc.Accelerations))
	}
}
