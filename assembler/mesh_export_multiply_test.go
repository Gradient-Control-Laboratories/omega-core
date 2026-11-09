package assembler_test

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
)

// TestExtractMeshDeclsMultiplyFolds guards the (temporarily closed-form)
// matrix.multiply against the canvas constant evaluator: the exact idiom
// models emit for mesh placements — a placement composed with a digest
// matrix — must fold to a 16-value Matrix.
func TestExtractMeshDeclsMultiplyFolds(t *testing.T) {
	src := `import "matrix";

fromLocal = [[0.0, 1.0, 0.0, 0.0],
             [0.0, 0.0, 1.0, 0.0],
             [1.0, 0.0, 0.0, 0.0],
             [0.0, 0.0, 0.0, 1.0]];
placement = matrix.compose([10.0, 0.0, 5.0], [0.0, 0.0, 0.0]);
return mesh(0, matrix.multiply(placement, fromLocal));
`
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
	// translate(10,0,5) times the axis permutation: the permutation's
	// columns pass through, the homogeneous column picks up the translation.
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

