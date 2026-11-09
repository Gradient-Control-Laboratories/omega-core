package assembler_test

import (
	"math"
	"os"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/assembler/meshplan"
)

// TestSliceMultiReturnConstFold is the mesh-placement reproduction for PR #399:
// a destructuring assignment binds `trans` to the constant return of a
// multi-return function, and mesh() placement folds that destructured constant
// to position the asset. Before the fix the fold failed ("reference \"trans\"
// is not a top-level constant"), the mesh map unroll silently fell back, and
// the plan carried zero asset nodes — so this asserts both assets land at the
// folded positions. See testdata/multireturn_const_fold.gsl.
func TestSliceMultiReturnConstFold(t *testing.T) {
	src, err := os.ReadFile("testdata/multireturn_const_fold.gsl")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	res, err := assembler.Assemble([]string{string(src)})
	if err != nil {
		t.Fatalf("assemble error: %v", err)
	}

	plan, _, err := meshplan.SliceProgram(res.Program, "test::leaf-")
	if err != nil {
		t.Fatalf("SliceModule: %v", err)
	}

	var assets []*meshplan.MeshPlanNode
	var walk func(n *meshplan.MeshPlanNode)
	walk = func(n *meshplan.MeshPlanNode) {
		if n == nil {
			return
		}
		if n.Kind == meshplan.MeshPlanKindAsset {
			assets = append(assets, n)
		}
		walk(n.Left)
		walk(n.Right)
	}
	walk(plan.Root)

	// trans folds to [60, 0, -40]; the two instances are placed at trans and
	// trans + [10, 0, 0]. Zero assets here would mean the destructured constant
	// failed to fold (the pre-#399 bug).
	wantPositions := [][3]float64{
		{60, 0, -40},
		{70, 0, -40},
	}
	if len(assets) != len(wantPositions) {
		t.Fatalf("plan asset leaves: got %d, want %d (destructured constant failed to fold?)", len(assets), len(wantPositions))
	}
	for k, want := range wantPositions {
		m := assets[k].Transform.Matrix
		if len(m) != 16 {
			t.Fatalf("plan asset %d Matrix length: got %d, want 16", k, len(m))
		}
		for i := range want {
			if math.Abs(m[12+i]-want[i]) > 1e-6 {
				t.Fatalf("plan asset %d translation: got [%v,%v,%v], want %v", k, m[12], m[13], m[14], want)
			}
		}
	}
}

