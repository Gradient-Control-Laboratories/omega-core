package assembler_test

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
)

// extractMatrixFromSource parses src, extracts its mesh decls against the
// real library sources, and returns the single top-level decl's 16-value
// placement matrix.
func extractMatrixFromSource(t *testing.T, src string) []float64 {
	t.Helper()
	res, err := assembler.Assemble([]string{src}, assembler.ExtractMeshPlan("leaf-"))
	if err != nil {
		t.Fatalf("assemble error: %v", err)
	}

	extract := res.MeshPlan.MeshDecls()
	if len(extract) != 1 {
		t.Fatalf("top-level mesh() count: got %d, want 1 (placement did not fold)", len(extract))
	}
	m := extract[0].Matrix
	if len(m) != 16 {
		t.Fatalf("Matrix length: got %d, want 16", len(m))
	}
	return m
}

// TestExtractMeshDeclsOperatorMultiplyFolds locks the `*` operator form of a
// placement product to the same values as the spelled-out matrix.multiply
// form (see mesh_export_multiply_test.go).
func TestExtractMeshDeclsOperatorMultiplyFolds(t *testing.T) {
	src := `import "matrix";

fromLocal = [[0.0, 1.0, 0.0, 0.0],
             [0.0, 0.0, 1.0, 0.0],
             [1.0, 0.0, 0.0, 0.0],
             [0.0, 0.0, 0.0, 1.0]];
placement = matrix.compose([10.0, 0.0, 5.0], [0.0, 0.0, 0.0]);
return mesh(0, placement * fromLocal);
`
	m := extractMatrixFromSource(t, src)
	want := []float64{
		0, 1, 0, 0,
		0, 0, 1, 0,
		1, 0, 0, 0,
		10, 0, 5, 1,
	}
	for i := range want {
		if math.Abs(m[i]-want[i]) > 1e-9 {
			t.Fatalf("Matrix[%d]: got %v, want %v (full: %v)", i, m[i], want[i], m)
		}
	}
}

// TestExtractMeshDeclsInverseFolds folds a matrix-CONSUMING library op —
// matrix.inverse indexes its argument's columns and runs the map/reduce
// bodies — end-to-end through parseMeshArgs.
func TestExtractMeshDeclsInverseFolds(t *testing.T) {
	src := `import "matrix";

placement = matrix.compose([10.0, 0.0, 5.0], [0.0, 0.0, 0.0]);
return mesh(0, matrix.inverse(placement));
`
	m := extractMatrixFromSource(t, src)
	want := []float64{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		-10, 0, -5, 1,
	}
	for i := range want {
		if math.Abs(m[i]-want[i]) > 1e-9 {
			t.Fatalf("Matrix[%d]: got %v, want %v (full: %v)", i, m[i], want[i], m)
		}
	}
}

