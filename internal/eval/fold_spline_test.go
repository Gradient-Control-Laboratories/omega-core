package eval_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/collect"
	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/types"
	parserlib "gradientcontrol.com/omega/core/parser"
)

func TestFoldLocalSplinePreservesNestedMatrix(t *testing.T) {
	prog, err := collect.Program([]string{`frames() {
  frame = uiTransform([[1,0,0,0],[0,1,0,0],[0,0,1,0],[3,4,5,1]]);
  path = uiSpline([frame, frame, frame]);
  selected = path[2];
  position = selected[3];
  return position;
}
return frames();`}, &source.MockManager{}, parserlib.ParsePackage)
	if err != nil {
		t.Fatal(err)
	}
	if err := types.Analyze(prog); err != nil {
		t.Fatal(err)
	}
	if _, err := eval.Fold(prog, nil); err != nil {
		t.Fatal(err)
	}
	fn := prog.Primary.Files[0].FnTypes[0].Groups[0].Functions[0]
	if got := fn.Assignments[2].Expression.ConstValues; !approxEqual(got, []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 3, 4, 5, 1}) {
		t.Fatalf("spline element lost matrix shape: %v", got)
	}
	if got := fn.Assignments[3].Expression.ConstValues; !approxEqual(got, []float64{3, 4, 5, 1}) {
		t.Fatalf("matrix column lost vector shape: %v", got)
	}
}
