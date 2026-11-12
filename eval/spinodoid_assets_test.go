//go:build cgo

package eval

import (
	"context"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/program"
)

func TestMeshAssetsApplySharedSpinodoidDomainWithoutImportSlotFold(t *testing.T) {
	bounds := []float64{-56.45, -13.2, -8.25, 63.45, 277.2, 41.25}
	e := &Evaluator{fetchAssets: func(context.Context, map[string]string, []byte, bool) (*EvalAssetsResponse, error) {
		return &EvalAssetsResponse{SpinodoidBounds: bounds}, nil
	}}
	prog, err := e.applyMeshAssets(context.Background(), nil, &program.Program{Primary: &program.Package{Scene: &program.Scene{Size: 1105}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(prog.Scene().SpinodoidBounds) != 6 || prog.Scene().SpinodoidBounds[4] != 277.2 || prog.Scene().Size != 1105 {
		t.Fatalf("scene = %v", prog.Scene())
	}
	bounds[4] = 999
	if prog.Scene().SpinodoidBounds[4] != 277.2 {
		t.Fatal("program retained mutable response bounds")
	}
}

func TestEvalGslGradedImportedBodyUsesSharedDomainInsteadOfFramingCube(t *testing.T) {
	e := New(source.NewLocalManager("../stdlib"), func(context.Context, map[string]string, []byte, bool) (*EvalAssetsResponse, error) {
		return &EvalAssetsResponse{SpinodoidBounds: []float64{-56.45, -13.2, -8.25, 63.45, 277.2, 41.25}}, nil
	})
	const document = `// Scene Size: 1105
import "lattice/spinodoid";
importMesh 0 "sole.stl";
field = spinodoid.fastGaborGraded(2, [pi/2,pi/2,pi/2], 1+p.y/300, [[1,0,0],[0,1,0],[0,0,1]]);
return field;`
	groups, err := e.EvalGsl(context.Background(), nil, document, "", []string{"field"}, [][3]float64{{0, 20, 10}, {0, 240, 10}})
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 || len(groups[0].Samples) != 1 || len(groups[0].Samples[0]) != 2 {
		t.Fatalf("unexpected evaluation result: %v", groups)
	}
}
