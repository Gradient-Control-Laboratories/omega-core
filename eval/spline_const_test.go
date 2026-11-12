//go:build cgo

package eval_test

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/eval"
)

func TestEvalSplineMatrixArgument(t *testing.T) {
	prog := parseProgram(t, `section(start, control, end) {
  position = start[3];
  return length(p - position.xyz) - 1;
}
branch() {
  start = uiTransform([[1,0,0,0],[0,1,0,0],[0,0,1,0],[3,4,5,1]]);
  middle = uiTransform([[1,0,0,0],[0,1,0,0],[0,0,1,0],[6,7,8,1]]);
  end = uiTransform([[1,0,0,0],[0,1,0,0],[0,0,1,0],[9,10,11,1]]);
  path = uiSpline([start,middle,end]);
  return section(path[0],path[1],path[2]);
}
return branch();`)
	values, err := eval.EvalGrad(prog, []float32{3, 5}, []float32{4, 4}, []float32{5, 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 1 || len(values[0]) != 2 {
		t.Fatalf("unexpected result: %v", values)
	}
	for index, want := range []float32{-1, 1} {
		if math.Abs(float64(values[0][index].Value-want)) > 1e-6 {
			t.Fatalf("matrix placement lost: %v, want %v", values[0][index].Value, want)
		}
	}
}
