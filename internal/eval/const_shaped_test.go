package eval

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// matOp builds a list-of-vectors literal [[..],[..],..] as the parser lowers
// it: a COLLECT of COLLECTs.
func matOp(cols ...[]float64) *program.Operand {
	ops := make([]*program.Operand, len(cols))
	for i, col := range cols {
		ops[i] = vec(col...)
	}
	return expr(program.Operation_COLLECT, ops...)
}

func TestConstFlattenMatrixLiteral(t *testing.T) {
	m := matOp(
		[]float64{1, 2, 3, 4},
		[]float64{5, 6, 7, 8},
		[]float64{9, 10, 11, 12},
		[]float64{13, 14, 15, 16},
	)
	got, err := Const(m, nil, nil)
	want := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	if err != nil || !approxEqual(got, want) {
		t.Fatalf("matrix literal flat value: got %v, err %v", got, err)
	}

	// A mixed scalar/vector collect keeps the same flat values as before.
	got, err = Const(expr(program.Operation_COLLECT, val(1), vec(2, 3)), nil, nil)
	if err != nil || !approxEqual(got, []float64{1, 2, 3}) {
		t.Fatalf("mixed collect flat value: got %v, err %v", got, err)
	}
}

func TestConstIndexListLiteral(t *testing.T) {
	m := matOp([]float64{1, 2, 3}, []float64{4, 5, 6})
	got, err := Const(expr(program.Operation_INDEX, m, val(1)), nil, nil)
	if err != nil || !approxEqual(got, []float64{4, 5, 6}) {
		t.Fatalf("list literal column index: got %v, err %v", got, err)
	}

	entry := expr(program.Operation_INDEX, expr(program.Operation_INDEX, m, val(1)), val(2))
	got, err = Const(entry, nil, nil)
	if err != nil || !approxEqual(got, []float64{6}) {
		t.Fatalf("list literal entry index: got %v, err %v", got, err)
	}
}

// TestConstMatrixParamColumnIndex proves a matrix keeps its list structure
// across a function parameter boundary: m[2] inside a body yields the whole
// column (when parameters bound flat this mis-indexed to the 3rd scalar).
func TestConstMatrixParamColumnIndex(t *testing.T) {
	fnTypes := []*program.FnType{
		fnTypeDecl("column", []string{"m"}, nil,
			expr(program.Operation_INDEX, ref("m"), val(2))),
		fnTypeDecl("entry", []string{"m"}, nil,
			expr(program.Operation_INDEX,
				expr(program.Operation_INDEX, ref("m"), val(2)), val(1))),
	}
	m := matOp(
		[]float64{1, 2, 3, 4},
		[]float64{5, 6, 7, 8},
		[]float64{9, 10, 11, 12},
		[]float64{13, 14, 15, 16},
	)

	got, err := Const(fc("column", m), nil, fnTypes)
	if err != nil || !approxEqual(got, []float64{9, 10, 11, 12}) {
		t.Fatalf("m[2] through a param: got %v, err %v", got, err)
	}
	got, err = Const(fc("entry", m), nil, fnTypes)
	if err != nil || !approxEqual(got, []float64{10}) {
		t.Fatalf("m[2][1] through a param: got %v, err %v", got, err)
	}
}

// TestConstDestructuredMatrixColumn indexes a matrix bound by a destructuring
// assignment to a whole column, with a non-constant sibling return untouched.
func TestConstDestructuredMatrixColumn(t *testing.T) {
	fnTypes := []*program.FnType{
		multiFnTypeDecl("build", nil,
			[]*program.Assignment{
				assign("sdf", expr(program.Operation_ADD, ref("p"), val(1))),
				assign("m", matOp([]float64{1, 2, 3}, []float64{4, 5, 6})),
			},
			ref("sdf"), ref("m")),
	}
	assignments := []*program.Assignment{
		multiAssign([]string{"shape", "m"}, fc("build")),
	}

	got, err := Const(expr(program.Operation_INDEX, ref("m"), val(1)), assignments, fnTypes)
	if err != nil || !approxEqual(got, []float64{4, 5, 6}) {
		t.Fatalf("destructured matrix column: got %v, err %v", got, err)
	}
}

// TestConstUiTransformColumnIndex proves the uiPoint/uiTransform markers pass
// their argument through with shape intact, so a wrapped placement stays
// column-indexable.
func TestConstUiTransformColumnIndex(t *testing.T) {
	m := matOp([]float64{1, 2, 3}, []float64{4, 5, 6})
	got, err := Const(expr(program.Operation_INDEX, fc("uiTransform", m), val(0)), nil, nil)
	if err != nil || !approxEqual(got, []float64{1, 2, 3}) {
		t.Fatalf("uiTransform(m)[0]: got %v, err %v", got, err)
	}
}

// TestConstElementsPkgMatrix locks the element contract the canvas map-unroll
// relies on: a list yields one value per element (a column vector each), a
// plain vector one scalar per component.
func TestConstElementsPkgMatrix(t *testing.T) {
	pkg := &program.Package{}
	m := matOp([]float64{1, 2, 3}, []float64{4, 5, 6})
	elems, err := ConstElementsPkg(m, pkg, nil)
	if err != nil || len(elems) != 2 ||
		!approxEqual(elems[0], []float64{1, 2, 3}) || !approxEqual(elems[1], []float64{4, 5, 6}) {
		t.Fatalf("list elements: got %v, err %v", elems, err)
	}

	elems, err = ConstElementsPkg(vec(7, 8), pkg, nil)
	if err != nil || len(elems) != 2 ||
		!approxEqual(elems[0], []float64{7}) || !approxEqual(elems[1], []float64{8}) {
		t.Fatalf("vector elements: got %v, err %v", elems, err)
	}
}
