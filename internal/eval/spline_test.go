package eval

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

func TestConstSplinePreservesMatrixShape(t *testing.T) {
	m := matOp([]float64{1, 0, 0, 0}, []float64{0, 1, 0, 0}, []float64{0, 0, 1, 0}, []float64{3, 4, 5, 1})
	list := expr(program.Operation_COLLECT, fc("uiTransform", m), fc("uiTransform", m), fc("uiTransform", m))
	column := expr(program.Operation_INDEX, expr(program.Operation_INDEX, fc("uiSpline", list), val(2)), val(3))
	got, err := Const(column, nil, nil)
	if err != nil || !approxEqual(got, []float64{3, 4, 5, 1}) {
		t.Fatalf("uiSpline index lost matrix shape: %v, %v", got, err)
	}
}
