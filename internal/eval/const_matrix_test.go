package eval

import (
	"math"
	"testing"

	_ "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func qfc(prefix, name string, args ...*program.Operand) *program.Operand {
	return &program.Operand{T: &program.Operand_FnCall{
		FnCall: &program.FnCall{
			Reference: &program.Reference{Prefix: prefix, Name: name},
			Arguments: args,
		}}}
}

// fnTypeDeclLambdas is fnTypeDecl with body-nested function declarations —
// the shape the parser produces when a body contains lambdas (hoisted to
// gsl_anon* names in the enclosing function's FnTypes).
func fnTypeDeclLambdas(name string, params []string, locals []*program.Assignment, lambdas []*program.FnType, ret *program.Operand) *program.FnType {
	ft := fnTypeDecl(name, params, locals, ret)
	ft.Groups[0].Functions[0].FnTypes = lambdas
	return ft
}

// TestConstLambdaMapFold folds a map over a body-nested lambda that closes
// over the enclosing call's local (which itself depends on a parameter) —
// the shape of every matrix-library body.
func TestConstLambdaMapFold(t *testing.T) {
	lambda := fnTypeDecl(program.AnonFnPrefix, []string{"x"}, nil,
		expr(program.Operation_MUL, ref("x"), ref("k2")))
	fnTypes := []*program.FnType{
		fnTypeDeclLambdas("scaleAll", []string{"xs", "s"},
			[]*program.Assignment{
				assign("k2", expr(program.Operation_MUL, ref("s"), val(2))),
			},
			[]*program.FnType{lambda},
			expr(program.Operation_MAP, ref(program.AnonFnPrefix), ref("xs"))),
	}

	got, err := Const(fc("scaleAll", vec(1, 2, 3), val(10)), nil, fnTypes)
	if err != nil || !approxEqual(got, []float64{20, 40, 60}) {
		t.Fatalf("map over closure lambda: got %v, err %v", got, err)
	}
}

func TestConstReduce(t *testing.T) {
	fnTypes := []*program.FnType{
		fnTypeDecl("add2", []string{"a", "b"}, nil,
			expr(program.Operation_ADD, ref("a"), ref("b"))),
		fnTypeDecl("accum", []string{"acc", "x"}, nil,
			expr(program.Operation_ADD, ref("acc"),
				vec3Op(ref("x"), expr(program.Operation_MUL, ref("x"), ref("x"))))),
	}

	// No-initial form: seed with the first element.
	got, err := Const(expr(program.Operation_REDUCE, ref("add2"), vec(1, 2, 3, 4)), nil, fnTypes)
	if err != nil || !approxEqual(got, []float64{10}) {
		t.Fatalf("reduce(fn, coll): got %v, err %v", got, err)
	}

	// A scalar collection behaves as a list of length one.
	got, err = Const(expr(program.Operation_REDUCE, ref("add2"), val(5)), nil, fnTypes)
	if err != nil || !approxEqual(got, []float64{5}) {
		t.Fatalf("reduce(fn, scalar): got %v, err %v", got, err)
	}

	// Explicit-initial form with a vector accumulator (not expanded).
	got, err = Const(expr(program.Operation_REDUCE, vec(0, 0), ref("accum"), vec(1, 2, 3)), nil, fnTypes)
	if err != nil || !approxEqual(got, []float64{6, 14}) {
		t.Fatalf("reduce(initial, fn, coll): got %v, err %v", got, err)
	}
}

// TestConstMulMatrixOperator locks the shape dispatch of `*`: with a list
// operand involved it is a real matrix product (this folded as a silent
// elementwise Hadamard product before values carried shape), while plain
// vectors keep the elementwise fold.
func TestConstMulMatrixOperator(t *testing.T) {
	a := matOp([]float64{1, 2}, []float64{3, 4})
	b := matOp([]float64{5, 6}, []float64{7, 8})

	got, err := Const(expr(program.Operation_MUL, a, b), nil, nil)
	if err != nil || !approxEqual(got, []float64{23, 34, 31, 46}) {
		t.Fatalf("A*B: got %v, err %v", got, err)
	}
	got, err = Const(expr(program.Operation_MUL, b, a), nil, nil)
	if err != nil || !approxEqual(got, []float64{19, 22, 43, 50}) {
		t.Fatalf("B*A: got %v, err %v", got, err)
	}

	// Rectangular: (2x3)·(3x2) -> 2x2.
	r := matOp([]float64{1, 2}, []float64{3, 4}, []float64{5, 6})
	c := matOp([]float64{1, 0, 1}, []float64{0, 1, 1})
	got, err = Const(expr(program.Operation_MUL, r, c), nil, nil)
	if err != nil || !approxEqual(got, []float64{6, 8, 8, 10}) {
		t.Fatalf("rectangular product: got %v, err %v", got, err)
	}

	// Matrix times column vector.
	got, err = Const(expr(program.Operation_MUL, a, vec(10, 20)), nil, nil)
	if err != nil || !approxEqual(got, []float64{70, 100}) {
		t.Fatalf("A*v: got %v, err %v", got, err)
	}

	// Scalar scaling, either order.
	got, err = Const(expr(program.Operation_MUL, a, val(2)), nil, nil)
	if err != nil || !approxEqual(got, []float64{2, 4, 6, 8}) {
		t.Fatalf("A*s: got %v, err %v", got, err)
	}
	got, err = Const(expr(program.Operation_MUL, val(2), a), nil, nil)
	if err != nil || !approxEqual(got, []float64{2, 4, 6, 8}) {
		t.Fatalf("s*A: got %v, err %v", got, err)
	}

	// Same-length vectors stay elementwise (Hadamard), backward-compatible.
	got, err = Const(expr(program.Operation_MUL, vec(1, 2, 3), vec(4, 5, 6)), nil, nil)
	if err != nil || !approxEqual(got, []float64{4, 10, 18}) {
		t.Fatalf("v*w: got %v, err %v", got, err)
	}

	// Vector times matrix needs a transposed row vector: error.
	if _, err := Const(expr(program.Operation_MUL, vec(1, 2), a), nil, nil); err == nil {
		t.Fatalf("expected error for v*M")
	}
	// Chained products with a matrix operand must be parenthesized: error.
	if _, err := Const(expr(program.Operation_MUL, a, b, a), nil, nil); err == nil {
		t.Fatalf("expected error for chained matrix product")
	}
}

func newTestProgram() *program.Program {
	return &program.Program{
		Primary: pkg([]*program.Import{
			{Path: "matrix", Prefix: "matrix"},
			{Path: "vector", Prefix: "vector"},
		}, nil, nil),
	}
}

// TestConstMatrixLibraryFolds folds the map/reduce/column-indexing bodies of
// the real matrix and vector libraries — none of which folded before values
// carried shape (only closed-form constructors and a multiply shim did).
func TestConstMatrixLibraryFolds(t *testing.T) {
	prog := newTestProgram()

	// multiply: composing two translations translates by the sum.
	got, err := ConstPkg(qfc("matrix", "multiply",
		qfc("matrix", "translation", vec(1, 0, 0)),
		qfc("matrix", "translation", vec(0, 2, 0))), prog.Primary, prog.GetImport)
	if err != nil {
		t.Fatalf("matrix.multiply of translations: %v", err)
	}
	want := []float64{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		1, 2, 0, 1,
	}
	if !approxEqual(got, want) {
		t.Fatalf("matrix.multiply of translations: got %v, want %v", got, want)
	}

	// transpose4 of a numbered matrix.
	m := matOp(
		[]float64{1, 2, 3, 4},
		[]float64{5, 6, 7, 8},
		[]float64{9, 10, 11, 12},
		[]float64{13, 14, 15, 16},
	)
	got, err = ConstPkg(qfc("matrix", "transpose4", m), prog.Primary, prog.GetImport)
	wantT := []float64{
		1, 5, 9, 13,
		2, 6, 10, 14,
		3, 7, 11, 15,
		4, 8, 12, 16,
	}
	if err != nil || !approxEqual(got, wantT) {
		t.Fatalf("matrix.transpose4: got %v, err %v", got, err)
	}

	// determinant4 of an axis scaling is the product of the factors.
	got, err = ConstPkg(qfc("matrix", "determinant4",
		qfc("matrix", "scaling", vec(2, 3, 4))), prog.Primary, prog.GetImport)
	if err != nil || !approxEqual(got, []float64{24}) {
		t.Fatalf("matrix.determinant4: got %v, err %v", got, err)
	}

	// The affine inverse of a translation is the negated translation.
	got, err = ConstPkg(qfc("matrix", "inverse",
		qfc("matrix", "translation", vec(1, 2, 3))), prog.Primary, prog.GetImport)
	wantInv := []float64{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		-1, -2, -3, 1,
	}
	if err != nil || !approxEqual(got, wantInv) {
		t.Fatalf("matrix.inverse of translation: got %v, err %v", got, err)
	}

	// A line rotation by pi about the z axis flips x (convention-free), and
	// multiplyPoint applies the affine matrix to a point.
	got, err = ConstPkg(qfc("vector", "multiplyPoint",
		qfc("matrix", "rotation", vec(0, 0, 0), vec(0, 0, 1), val(math.Pi)),
		vec(1, 0, 0)), prog.Primary, prog.GetImport)
	if err != nil || !approxEqual(got, []float64{-1, 0, 0}) {
		t.Fatalf("rotation about a line applied to a point: got %v, err %v", got, err)
	}
}

// TestConstMatrixMultiplyGeneralRect proves the resolvable library path
// handles rectangular products the square-only fallback cannot.
func TestConstMatrixMultiplyGeneralRect(t *testing.T) {
	prog := newTestProgram()
	r := matOp([]float64{1, 2}, []float64{3, 4}, []float64{5, 6})
	c := matOp([]float64{1, 0, 1}, []float64{0, 1, 1})
	got, err := ConstPkg(qfc("matrix", "multiply", r, c), prog.Primary, prog.GetImport)
	if err != nil || !approxEqual(got, []float64{6, 8, 8, 10}) {
		t.Fatalf("rectangular matrix.multiply: got %v, err %v", got, err)
	}
}

func TestConstNormalizeFolds(t *testing.T) {
	got, err := Const(fc("normalize", vec(3, 4)), nil, nil)
	if err != nil || !approxEqual(got, []float64{0.6, 0.8}) {
		t.Fatalf("normalize: got %v, err %v", got, err)
	}
}

// TestConstMatrixMultiplyFallbackNoImports exercises the import-free
// fallback with shaped list arguments, which multiplies as matrices
// (rectangular included) rather than requiring the flat square form.
func TestConstMatrixMultiplyFallbackNoImports(t *testing.T) {
	r := matOp([]float64{1, 2}, []float64{3, 4}, []float64{5, 6})
	c := matOp([]float64{1, 0, 1}, []float64{0, 1, 1})
	got, err := ConstPkg(qfc("matrix", "multiply", r, c), &program.Package{}, nil)
	if err != nil || !approxEqual(got, []float64{6, 8, 8, 10}) {
		t.Fatalf("fallback list multiply: got %v, err %v", got, err)
	}
}
