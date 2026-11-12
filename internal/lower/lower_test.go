package lower_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/closure"
	"gradientcontrol.com/omega/core/internal/collect"
	"gradientcontrol.com/omega/core/internal/link"
	"gradientcontrol.com/omega/core/internal/lower"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/types"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func runPipeline(src string) (*program.Program, error) {
	srcMgr := &source.MockManager{}
	prog, err := collect.Program([]string{src}, srcMgr, parser.ParsePackage)
	if err != nil {
		return nil, err
	}
	if err := link.References(prog, srcMgr, parser.ParsePackage); err != nil {
		return nil, err
	}
	if err := closure.Transform(prog); err != nil {
		return nil, err
	}
	if err := link.References(prog, srcMgr, parser.ParsePackage); err != nil {
		return nil, err
	}
	if err := types.Analyze(prog); err != nil {
		return nil, err
	}
	if err := lower.Transform(prog); err != nil {
		return nil, err
	}
	// Relink and retype check
	if err := link.References(prog, srcMgr, parser.ParsePackage); err != nil {
		return nil, err
	}
	if err := types.Analyze(prog); err != nil {
		return nil, err
	}
	return prog, nil
}

func TestLower_CSGReduction(t *testing.T) {
	src := `
a = mesh(0);
b = mesh(1);
c = mesh(2);
return meshUnion(a, b, c);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ret := prog.Primary.Files[0].Returns[0]
	// Expect return to be meshUnion(meshUnion(a, b), c)
	fnCall, ok := ret.T.(*program.Operand_FnCall)
	if !ok {
		t.Fatalf("expected function call, got %T", ret.T)
	}
	if fnCall.FnCall.Reference.Name != "meshUnion" {
		t.Errorf("expected meshUnion call, got %q", fnCall.FnCall.Reference.Name)
	}
	if len(fnCall.FnCall.Arguments) != 2 {
		t.Errorf("expected binary call with 2 arguments, got %d", len(fnCall.FnCall.Arguments))
	}

	subCallOp := fnCall.FnCall.Arguments[0]
	subFnCall, ok := subCallOp.T.(*program.Operand_FnCall)
	if !ok {
		t.Fatalf("expected nested function call, got %T", subCallOp.T)
	}
	if subFnCall.FnCall.Reference.Name != "meshUnion" {
		t.Errorf("expected nested meshUnion call, got %q", subFnCall.FnCall.Reference.Name)
	}
}

func TestLower_MatrixMultiply(t *testing.T) {
	src := `
v1 = [1, 2, 3];
v2 = [4, 5, 6];
v3 = [7, 8, 9];
m1 = [v1, v2, v3];
m2 = [v1, v2, v3];
return m1 * m2;
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ret := prog.Primary.Files[0].Returns[0]
	fnCall, ok := ret.T.(*program.Operand_FnCall)
	if !ok {
		t.Fatalf("expected routed function call, got %T", ret.T)
	}
	if fnCall.FnCall.Reference.Prefix != "matrix" || fnCall.FnCall.Reference.Name != "multiply" {
		t.Errorf("expected matrix.multiply, got %s.%s", fnCall.FnCall.Reference.Prefix, fnCall.FnCall.Reference.Name)
	}

	// Verify import was auto-injected
	hasMatrixImport := false
	for _, imp := range prog.Primary.Files[0].Imports {
		if imp.Path == "matrix" {
			hasMatrixImport = true
			break
		}
	}
	if !hasMatrixImport {
		t.Error("expected matrix import to be auto-injected")
	}
}

func TestLower_MatrixScale(t *testing.T) {
	src := `
v1 = [1, 2, 3];
m1 = [v1, v1, v1];
s = 2.0;
return m1 * s;
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ret := prog.Primary.Files[0].Returns[0]
	fnCall, ok := ret.T.(*program.Operand_FnCall)
	if !ok {
		t.Fatalf("expected routed function call, got %T", ret.T)
	}
	if fnCall.FnCall.Reference.Prefix != "matrix" || fnCall.FnCall.Reference.Name != "scale" {
		t.Errorf("expected matrix.scale, got %s.%s", fnCall.FnCall.Reference.Prefix, fnCall.FnCall.Reference.Name)
	}
}

func TestLower_Map_1Param(t *testing.T) {
	src := `
doubleVal(x) { return x * 2; }
xs = [10, 20, 30];
return map(doubleVal, xs);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ret := prog.Primary.Files[0].Returns[0]
	exprOp, ok := ret.T.(*program.Operand_Expression)
	if !ok || exprOp.Expression.Operation != program.Operation_COLLECT {
		t.Fatalf("expected COLLECT expression, got %T (%v)", ret.T, ret)
	}
	if len(exprOp.Expression.Operands) != 3 {
		t.Fatalf("expected 3 unrolled element calls, got %d", len(exprOp.Expression.Operands))
	}
	for i, op := range exprOp.Expression.Operands {
		fnCall, ok := op.T.(*program.Operand_FnCall)
		if !ok {
			t.Fatalf("operand %d: expected Operand_FnCall, got %T", i, op.T)
		}
		if fnCall.FnCall.Reference.Name != "doubleVal" {
			t.Errorf("operand %d: expected doubleVal, got %q", i, fnCall.FnCall.Reference.Name)
		}
		if len(fnCall.FnCall.Arguments) != 1 {
			t.Errorf("operand %d: expected 1 arg, got %d", i, len(fnCall.FnCall.Arguments))
		}
	}
}

func TestLower_Map_2Param(t *testing.T) {
	src := `
scaleByIndex(x, i) { return x * (i + 1); }
return map(scaleByIndex, [10, 20]);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ret := prog.Primary.Files[0].Returns[0]
	exprOp, ok := ret.T.(*program.Operand_Expression)
	if !ok || exprOp.Expression.Operation != program.Operation_COLLECT {
		t.Fatalf("expected COLLECT expression, got %T", ret.T)
	}
	if len(exprOp.Expression.Operands) != 2 {
		t.Fatalf("expected 2 unrolled calls, got %d", len(exprOp.Expression.Operands))
	}
	for i, op := range exprOp.Expression.Operands {
		fnCall, ok := op.T.(*program.Operand_FnCall)
		if !ok {
			t.Fatalf("operand %d: expected Operand_FnCall, got %T", i, op.T)
		}
		if len(fnCall.FnCall.Arguments) != 2 {
			t.Fatalf("operand %d: expected 2 args, got %d", i, len(fnCall.FnCall.Arguments))
		}
		// Second argument should be index literal i
		idxVal, ok := fnCall.FnCall.Arguments[1].T.(*program.Operand_Value)
		if !ok || idxVal.Value != float64(i) {
			t.Errorf("operand %d: expected index %d, got %v", i, i, fnCall.FnCall.Arguments[1].T)
		}
	}
}

func TestLower_Map_Matrix(t *testing.T) {
	src := `
myFn(v, idx) { return v * (idx + 1); }
m = [[1, 2], [3, 4]];
return map(myFn, m);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ret := prog.Primary.Files[0].Returns[0]
	matrixCollect, ok := ret.T.(*program.Operand_Expression)
	if !ok || matrixCollect.Expression.Operation != program.Operation_COLLECT {
		t.Fatalf("expected matrix COLLECT expression, got %T", ret.T)
	}
	if len(matrixCollect.Expression.Operands) != 2 {
		t.Fatalf("expected 2 column element calls, got %d", len(matrixCollect.Expression.Operands))
	}
	for colIdx, colOp := range matrixCollect.Expression.Operands {
		fnCall, ok := colOp.T.(*program.Operand_FnCall)
		if !ok {
			t.Fatalf("col %d: expected FnCall, got %T", colIdx, colOp.T)
		}
		if len(fnCall.FnCall.Arguments) != 2 {
			t.Fatalf("col %d: expected 2 args, got %d", colIdx, len(fnCall.FnCall.Arguments))
		}
		idxVal, ok := fnCall.FnCall.Arguments[1].T.(*program.Operand_Value)
		if !ok || idxVal.Value != float64(colIdx) {
			t.Errorf("col %d: expected index %d, got %v", colIdx, colIdx, fnCall.FnCall.Arguments[1].T)
		}
	}
}

func TestLower_Reduce_NoInitial(t *testing.T) {
	src := `
myAdd(a, b) { return a + b; }
return reduce(myAdd, [1, 2, 3]);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ret := prog.Primary.Files[0].Returns[0]
	// Expect myAdd(myAdd(1, 2), 3)
	fnCall, ok := ret.T.(*program.Operand_FnCall)
	if !ok {
		t.Fatalf("expected FnCall, got %T", ret.T)
	}
	if fnCall.FnCall.Reference.Name != "myAdd" {
		t.Errorf("expected myAdd, got %q", fnCall.FnCall.Reference.Name)
	}
	if len(fnCall.FnCall.Arguments) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(fnCall.FnCall.Arguments))
	}
}

func TestLower_Reduce_WithInitial(t *testing.T) {
	src := `
myAdd(a, b) { return a + b; }
return reduce(myAdd, [1, 2], 100);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ret := prog.Primary.Files[0].Returns[0]
	// Expect myAdd(myAdd(100, 1), 2)
	fnCall, ok := ret.T.(*program.Operand_FnCall)
	if !ok {
		t.Fatalf("expected FnCall, got %T", ret.T)
	}
	if fnCall.FnCall.Reference.Name != "myAdd" {
		t.Errorf("expected myAdd, got %q", fnCall.FnCall.Reference.Name)
	}
}

func TestLower_Map_SingleElementList(t *testing.T) {
	src := `
myFn(x) { return x * 2.0; }
return map(myFn, [5.0]);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ret := prog.Primary.Files[0].Returns[0]
	collectExpr, ok := ret.T.(*program.Operand_Expression)
	if !ok || collectExpr.Expression.Operation != program.Operation_COLLECT {
		t.Fatalf("expected COLLECT expression, got %T", ret.T)
	}
	if len(collectExpr.Expression.Operands) != 1 {
		t.Fatalf("expected 1 element in collect, got %d", len(collectExpr.Expression.Operands))
	}
	fnCall, ok := collectExpr.Expression.Operands[0].T.(*program.Operand_FnCall)
	if !ok || fnCall.FnCall.Reference.Name != "myFn" {
		t.Fatalf("expected call to myFn, got %T", collectExpr.Expression.Operands[0].T)
	}
}

func TestLower_Reduce_SingleElementList(t *testing.T) {
	srcNoInit := `
myAdd(a, b) { return a + b; }
return reduce(myAdd, [42.0]);
`
	prog1, err := runPipeline(srcNoInit)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ret1 := prog1.Primary.Files[0].Returns[0]
	val1, ok := ret1.T.(*program.Operand_Value)
	if !ok || val1.Value != 42.0 {
		t.Fatalf("expected direct value 42.0, got %v", ret1.T)
	}

	srcInit := `
myAdd(a, b) { return a + b; }
return reduce(myAdd, [42.0], 10.0);
`
	prog2, err := runPipeline(srcInit)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ret2 := prog2.Primary.Files[0].Returns[0]
	fnCall2, ok := ret2.T.(*program.Operand_FnCall)
	if !ok || fnCall2.FnCall.Reference.Name != "myAdd" {
		t.Fatalf("expected call to myAdd, got %T", ret2.T)
	}
	if len(fnCall2.FnCall.Arguments) != 2 {
		t.Fatalf("expected 2 arguments to myAdd, got %d", len(fnCall2.FnCall.Arguments))
	}
}

func TestLower_VectorAddition(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected string
	}{
		{
			name:     "vec_plus_vec",
			src:      "v1 = [1.0, 2.0, 3.0]; v2 = [4.0, 5.0, 6.0]; return v1 + v2;",
			expected: "vector.addVector",
		},
		{
			name:     "vec_plus_scalar",
			src:      "v = [1.0, 2.0, 3.0]; s = 2.0; return v + s;",
			expected: "vector.addScalar",
		},
		{
			name:     "scalar_plus_vec",
			src:      "v = [1.0, 2.0, 3.0]; s = 2.0; return s + v;",
			expected: "vector.addScalar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prog, err := runPipeline(tc.src)
			if err != nil {
				t.Fatalf("runPipeline failed: %v", err)
			}
			ret := prog.Primary.Files[0].Returns[0]
			fnCall, ok := ret.T.(*program.Operand_FnCall)
			if !ok {
				t.Fatalf("expected FnCall, got %T", ret.T)
			}
			routed := fnCall.FnCall.Reference.Prefix + "." + fnCall.FnCall.Reference.Name
			if routed != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, routed)
			}
		})
	}
}

func TestLower_VectorSubtraction(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected string
	}{
		{
			name:     "vec_minus_vec",
			src:      "v1 = [1.0, 2.0, 3.0]; v2 = [4.0, 5.0, 6.0]; return v1 - v2;",
			expected: "vector.subtractVector",
		},
		{
			name:     "vec_minus_scalar",
			src:      "v = [1.0, 2.0, 3.0]; s = 2.0; return v - s;",
			expected: "vector.subtractScalarFrom",
		},
		{
			name:     "scalar_minus_vec",
			src:      "v = [1.0, 2.0, 3.0]; s = 2.0; return s - v;",
			expected: "vector.subtractFromScalar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prog, err := runPipeline(tc.src)
			if err != nil {
				t.Fatalf("runPipeline failed: %v", err)
			}
			ret := prog.Primary.Files[0].Returns[0]
			fnCall, ok := ret.T.(*program.Operand_FnCall)
			if !ok {
				t.Fatalf("expected FnCall, got %T", ret.T)
			}
			routed := fnCall.FnCall.Reference.Prefix + "." + fnCall.FnCall.Reference.Name
			if routed != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, routed)
			}
		})
	}
}

func TestLower_VectorMultiplication(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected string
	}{
		{
			name:     "vec_mul_scalar",
			src:      "v = [1.0, 2.0, 3.0]; s = 2.0; return v * s;",
			expected: "vector.multiplyScalar",
		},
		{
			name:     "scalar_mul_vec",
			src:      "v = [1.0, 2.0, 3.0]; s = 2.0; return s * v;",
			expected: "vector.multiplyScalar",
		},
		{
			name:     "vec_mul_vec",
			src:      "v1 = [1.0, 2.0, 3.0]; v2 = [4.0, 5.0, 6.0]; return v1 * v2;",
			expected: "vector.multiplyComponents",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prog, err := runPipeline(tc.src)
			if err != nil {
				t.Fatalf("runPipeline failed: %v", err)
			}
			ret := prog.Primary.Files[0].Returns[0]
			fnCall, ok := ret.T.(*program.Operand_FnCall)
			if !ok {
				t.Fatalf("expected FnCall, got %T", ret.T)
			}
			routed := fnCall.FnCall.Reference.Prefix + "." + fnCall.FnCall.Reference.Name
			if routed != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, routed)
			}
		})
	}
}

func TestLower_VectorDivision(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected string
	}{
		{
			name:     "vec_div_scalar",
			src:      "v = [1.0, 2.0, 3.0]; s = 2.0; return v / s;",
			expected: "vector.divideByScalar",
		},
		{
			name:     "scalar_div_vec",
			src:      "v = [1.0, 2.0, 3.0]; s = 2.0; return s / v;",
			expected: "vector.divideScalarBy",
		},
		{
			name:     "vec_div_vec",
			src:      "v1 = [1.0, 2.0, 3.0]; v2 = [4.0, 5.0, 6.0]; return v1 / v2;",
			expected: "vector.divideComponents",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prog, err := runPipeline(tc.src)
			if err != nil {
				t.Fatalf("runPipeline failed: %v", err)
			}
			ret := prog.Primary.Files[0].Returns[0]
			fnCall, ok := ret.T.(*program.Operand_FnCall)
			if !ok {
				t.Fatalf("expected FnCall, got %T", ret.T)
			}
			routed := fnCall.FnCall.Reference.Prefix + "." + fnCall.FnCall.Reference.Name
			if routed != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, routed)
			}
		})
	}
}

func TestLower_VectorNegation(t *testing.T) {
	src := "v = [1.0, 2.0, 3.0]; return -v;"
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("runPipeline failed: %v", err)
	}
	ret := prog.Primary.Files[0].Returns[0]
	fnCall, ok := ret.T.(*program.Operand_FnCall)
	if !ok {
		t.Fatalf("expected FnCall, got %T", ret.T)
	}
	routed := fnCall.FnCall.Reference.Prefix + "." + fnCall.FnCall.Reference.Name
	if routed != "vector.negate" {
		t.Errorf("expected vector.negate, got %s", routed)
	}

	// Verify import was auto-injected
	hasVectorImport := false
	for _, imp := range prog.Primary.Files[0].Imports {
		if imp.Path == "vector" {
			hasVectorImport = true
			break
		}
	}
	if !hasVectorImport {
		t.Error("expected vector import to be auto-injected")
	}
}

func TestLower_ChainedOperations(t *testing.T) {
	tests := []struct {
		name       string
		src        string
		validateFn func(t *testing.T, ret *program.Operand)
	}{
		{
			name: "pure_scalar_mul_preserved",
			src:  "s1 = 2.0; s2 = 3.0; s3 = 4.0; return s1 * s2 * s3;",
			validateFn: func(t *testing.T, ret *program.Operand) {
				expr, ok := ret.T.(*program.Operand_Expression)
				if !ok || expr.Expression.Operation != program.Operation_MUL {
					t.Fatalf("expected Operation_MUL expression, got %T", ret.T)
				}
				if len(expr.Expression.Operands) != 3 {
					t.Errorf("expected 3 flat operands for pure scalar, got %d", len(expr.Expression.Operands))
				}
			},
		},
		{
			name: "pure_scalar_add_preserved",
			src:  "s1 = 2.0; s2 = 3.0; s3 = 4.0; return s1 + s2 + s3;",
			validateFn: func(t *testing.T, ret *program.Operand) {
				expr, ok := ret.T.(*program.Operand_Expression)
				if !ok || expr.Expression.Operation != program.Operation_ADD {
					t.Fatalf("expected Operation_ADD expression, got %T", ret.T)
				}
				if len(expr.Expression.Operands) != 3 {
					t.Errorf("expected 3 flat operands for pure scalar, got %d", len(expr.Expression.Operands))
				}
			},
		},
		{
			name: "scalar_matrix_mul",
			src:  "s = 2.0; M = [[1.0, 0.0], [0.0, 1.0]]; return s * M;",
			validateFn: func(t *testing.T, ret *program.Operand) {
				fc, ok := ret.T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected FnCall, got %T", ret.T)
				}
				routed := fc.FnCall.Reference.Prefix + "." + fc.FnCall.Reference.Name
				if routed != "matrix.scale" {
					t.Errorf("expected matrix.scale, got %s", routed)
				}
			},
		},
		{
			name: "matrix_scalar_mul",
			src:  "s = 2.0; M = [[1.0, 0.0], [0.0, 1.0]]; return M * s;",
			validateFn: func(t *testing.T, ret *program.Operand) {
				fc, ok := ret.T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected FnCall, got %T", ret.T)
				}
				routed := fc.FnCall.Reference.Prefix + "." + fc.FnCall.Reference.Name
				if routed != "matrix.scale" {
					t.Errorf("expected matrix.scale, got %s", routed)
				}
			},
		},
		{
			name: "matrix_matrix_matrix_mul",
			src:  "M1 = [[1.0, 0.0], [0.0, 1.0]]; M2 = [[2.0, 0.0], [0.0, 2.0]]; M3 = [[3.0, 0.0], [0.0, 3.0]]; return M1 * M2 * M3;",
			validateFn: func(t *testing.T, ret *program.Operand) {
				fc, ok := ret.T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected FnCall, got %T", ret.T)
				}
				routed := fc.FnCall.Reference.Prefix + "." + fc.FnCall.Reference.Name
				if routed != "matrix.multiply" {
					t.Fatalf("expected outer matrix.multiply, got %s", routed)
				}
				inner, ok := fc.FnCall.Arguments[0].T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected inner FnCall, got %T", fc.FnCall.Arguments[0].T)
				}
				innerRouted := inner.FnCall.Reference.Prefix + "." + inner.FnCall.Reference.Name
				if innerRouted != "matrix.multiply" {
					t.Errorf("expected inner matrix.multiply, got %s", innerRouted)
				}
			},
		},
		{
			name: "matrix_vector_mul",
			src:  "M = [[1.0, 0.0], [0.0, 1.0]]; v = [3.0, 4.0]; return M * v;",
			validateFn: func(t *testing.T, ret *program.Operand) {
				fc, ok := ret.T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected FnCall, got %T", ret.T)
				}
				routed := fc.FnCall.Reference.Prefix + "." + fc.FnCall.Reference.Name
				if routed != "vector.multiply" {
					t.Errorf("expected vector.multiply, got %s", routed)
				}
			},
		},
		{
			name: "scalar_vector_mul",
			src:  "s = 2.0; v = [3.0, 4.0]; return s * v;",
			validateFn: func(t *testing.T, ret *program.Operand) {
				fc, ok := ret.T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected FnCall, got %T", ret.T)
				}
				routed := fc.FnCall.Reference.Prefix + "." + fc.FnCall.Reference.Name
				if routed != "vector.multiplyScalar" {
					t.Errorf("expected vector.multiplyScalar, got %s", routed)
				}
			},
		},
		{
			name: "vector_scalar_mul",
			src:  "s = 2.0; v = [3.0, 4.0]; return v * s;",
			validateFn: func(t *testing.T, ret *program.Operand) {
				fc, ok := ret.T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected FnCall, got %T", ret.T)
				}
				routed := fc.FnCall.Reference.Prefix + "." + fc.FnCall.Reference.Name
				if routed != "vector.multiplyScalar" {
					t.Errorf("expected vector.multiplyScalar, got %s", routed)
				}
			},
		},
		{
			name: "vector_vector_mul",
			src:  "v1 = [1.0, 2.0]; v2 = [3.0, 4.0]; return v1 * v2;",
			validateFn: func(t *testing.T, ret *program.Operand) {
				fc, ok := ret.T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected FnCall, got %T", ret.T)
				}
				routed := fc.FnCall.Reference.Prefix + "." + fc.FnCall.Reference.Name
				if routed != "vector.multiplyComponents" {
					t.Errorf("expected vector.multiplyComponents, got %s", routed)
				}
			},
		},
		{
			name: "extreme_heterogeneous_multiplication_chain",
			src: `
s1 = 2.0; s2 = 3.0;
M1 = [[1.0, 0.0], [0.0, 1.0]];
s3 = 4.0;
M2 = [[1.0, 0.0], [0.0, 1.0]];
s4 = 5.0; s5 = 6.0;
v1 = [1.0, 2.0];
s6 = 7.0;
return s1 * s2 * M1 * s3 * M2 * s4 * s5 * v1 * s6;
`,
			validateFn: func(t *testing.T, ret *program.Operand) {
				fc, ok := ret.T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected outer FnCall, got %T", ret.T)
				}
				routed := fc.FnCall.Reference.Prefix + "." + fc.FnCall.Reference.Name
				if routed != "vector.multiplyScalar" {
					t.Fatalf("expected outermost vector.multiplyScalar, got %s", routed)
				}
			},
		},
		{
			name: "chained_vector_scalar_addition",
			src:  "v1 = [1.0, 2.0]; s1 = 3.0; s2 = 4.0; v2 = [5.0, 6.0]; return v1 + s1 + s2 + v2;",
			validateFn: func(t *testing.T, ret *program.Operand) {
				fc, ok := ret.T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected outer FnCall, got %T", ret.T)
				}
				routed := fc.FnCall.Reference.Prefix + "." + fc.FnCall.Reference.Name
				if routed != "vector.addVector" {
					t.Fatalf("expected outer vector.addVector, got %s", routed)
				}
				inner, ok := fc.FnCall.Arguments[0].T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected inner FnCall, got %T", fc.FnCall.Arguments[0].T)
				}
				innerRouted := inner.FnCall.Reference.Prefix + "." + inner.FnCall.Reference.Name
				if innerRouted != "vector.addScalar" {
					t.Fatalf("expected inner vector.addScalar, got %s", innerRouted)
				}
			},
		},
		{
			name: "chained_vector_subtraction",
			src:  "v1 = [1.0, 2.0]; v2 = [3.0, 4.0]; v3 = [5.0, 6.0]; return v1 - v2 - v3;",
			validateFn: func(t *testing.T, ret *program.Operand) {
				fc, ok := ret.T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected outer FnCall, got %T", ret.T)
				}
				routed := fc.FnCall.Reference.Prefix + "." + fc.FnCall.Reference.Name
				if routed != "vector.subtractVector" {
					t.Fatalf("expected outer vector.subtractVector, got %s", routed)
				}
			},
		},
		{
			name: "chained_matrix_addition",
			src:  "M1 = [[1.0, 0.0], [0.0, 1.0]]; M2 = [[2.0, 0.0], [0.0, 2.0]]; M3 = [[3.0, 0.0], [0.0, 3.0]]; return M1 + M2 + M3;",
			validateFn: func(t *testing.T, ret *program.Operand) {
				fc, ok := ret.T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected outer FnCall, got %T", ret.T)
				}
				routed := fc.FnCall.Reference.Prefix + "." + fc.FnCall.Reference.Name
				if routed != "matrix.plus" {
					t.Fatalf("expected outer matrix.plus, got %s", routed)
				}
			},
		},
		{
			name: "chained_matrix_subtraction",
			src:  "M1 = [[1.0, 0.0], [0.0, 1.0]]; M2 = [[2.0, 0.0], [0.0, 2.0]]; M3 = [[3.0, 0.0], [0.0, 3.0]]; return M1 - M2 - M3;",
			validateFn: func(t *testing.T, ret *program.Operand) {
				fc, ok := ret.T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected outer FnCall, got %T", ret.T)
				}
				routed := fc.FnCall.Reference.Prefix + "." + fc.FnCall.Reference.Name
				if routed != "matrix.minus" {
					t.Fatalf("expected outer matrix.minus, got %s", routed)
				}
			},
		},
		{
			name: "chained_matrix_scalar_division",
			src:  "M = [[1.0, 0.0], [0.0, 1.0]]; s1 = 2.0; s2 = 4.0; return M / s1 / s2;",
			validateFn: func(t *testing.T, ret *program.Operand) {
				fc, ok := ret.T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected outer FnCall, got %T", ret.T)
				}
				routed := fc.FnCall.Reference.Prefix + "." + fc.FnCall.Reference.Name
				if routed != "matrix.scale" {
					t.Fatalf("expected outer matrix.scale, got %s", routed)
				}
				inner, ok := fc.FnCall.Arguments[1].T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected inner FnCall, got %T", fc.FnCall.Arguments[1].T)
				}
				innerRouted := inner.FnCall.Reference.Prefix + "." + inner.FnCall.Reference.Name
				if innerRouted != "matrix.scale" {
					t.Fatalf("expected inner matrix.scale, got %s", innerRouted)
				}
			},
		},
		{
			name: "chained_vector_scalar_division",
			src:  "v = [1.0, 2.0]; s1 = 2.0; s2 = 4.0; return v / s1 / s2;",
			validateFn: func(t *testing.T, ret *program.Operand) {
				fc, ok := ret.T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected outer FnCall, got %T", ret.T)
				}
				routed := fc.FnCall.Reference.Prefix + "." + fc.FnCall.Reference.Name
				if routed != "vector.divideByScalar" {
					t.Fatalf("expected outer vector.divideByScalar, got %s", routed)
				}
			},
		},
		{
			name: "chained_vector_vector_division",
			src:  "v1 = [1.0, 2.0]; v2 = [3.0, 4.0]; v3 = [5.0, 6.0]; return v1 / v2 / v3;",
			validateFn: func(t *testing.T, ret *program.Operand) {
				fc, ok := ret.T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected outer FnCall, got %T", ret.T)
				}
				routed := fc.FnCall.Reference.Prefix + "." + fc.FnCall.Reference.Name
				if routed != "vector.divideComponents" {
					t.Fatalf("expected outer vector.divideComponents, got %s", routed)
				}
			},
		},
		{
			name: "scalar_vector_division",
			src:  "s = 10.0; v = [2.0, 5.0]; return s / v;",
			validateFn: func(t *testing.T, ret *program.Operand) {
				fc, ok := ret.T.(*program.Operand_FnCall)
				if !ok {
					t.Fatalf("expected FnCall, got %T", ret.T)
				}
				routed := fc.FnCall.Reference.Prefix + "." + fc.FnCall.Reference.Name
				if routed != "vector.divideScalarBy" {
					t.Fatalf("expected vector.divideScalarBy, got %s", routed)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prog, err := runPipeline(tc.src)
			if err != nil {
				t.Fatalf("runPipeline failed: %v", err)
			}
			ret := prog.Primary.Files[0].Returns[0]
			tc.validateFn(t, ret)
		})
	}
}

func TestLower_ChainedOperations_Diagnostics(t *testing.T) {
	tests := []struct {
		name        string
		src         string
		errContains string
	}{
		{
			name:        "illegal_ordering_v_times_M",
			src:         "v = [1.0, 2.0]; M = [[1.0, 0.0], [0.0, 1.0]]; return v * M;",
			errContains: "cannot multiply vector by matrix at operand 2: write matrix * vector instead",
		},
		{
			name:        "illegal_ordering_chain_s_v_M",
			src:         "s1 = 2.0; s2 = 3.0; v1 = [1.0, 2.0]; M1 = [[1.0, 0.0], [0.0, 1.0]]; return s1 * s2 * v1 * M1;",
			errContains: "cannot multiply vector by matrix at operand 4: write matrix * vector instead",
		},
		{
			name: "matrix_matrix_dim_mismatch",
			src: `
s1 = 2.0; s2 = 3.0;
M1 = [[1.0, 0.0, 0.0], [0.0, 1.0, 0.0], [0.0, 0.0, 1.0], [1.0, 1.0, 1.0]];
s3 = 4.0;
M2 = [[1.0, 0.0], [0.0, 1.0], [1.0, 1.0]];
return s1 * s2 * M1 * s3 * M2;
`,
			errContains: "matrix multiplication dimension mismatch at operand 5: 3x4 vs 2x3",
		},
		{
			name: "matrix_vector_dim_mismatch",
			src: `
s1 = 2.0; s2 = 3.0;
M1 = [[1.0, 0.0, 0.0], [0.0, 1.0, 0.0], [0.0, 0.0, 1.0]];
s3 = 4.0;
M2 = [[1.0, 0.0, 0.0], [0.0, 1.0, 0.0], [0.0, 0.0, 1.0]];
s4 = 5.0; s5 = 6.0;
v1 = [1.0, 2.0, 3.0, 4.0];
return s1 * s2 * M1 * s3 * M2 * s4 * s5 * v1;
`,
			errContains: "matrix-vector dimension mismatch at operand 8: matrix 3x3, vector len 4",
		},
		{
			name:        "vector_vector_dim_mismatch",
			src:         "v1 = [1.0, 2.0]; v2 = [3.0, 4.0, 5.0]; return v1 * v2;",
			errContains: "vector multiplication dimension mismatch at operand 2: len 2 vs len 3",
		},
		{
			name:        "matrix_scalar_add",
			src:         "M = [[1.0, 0.0], [0.0, 1.0]]; s = 2.0; return M + s;",
			errContains: "cannot add matrix and scalar at operand 2: matrix addition requires matrices",
		},
		{
			name:        "matrix_scalar_sub",
			src:         "M = [[1.0, 0.0], [0.0, 1.0]]; s = 2.0; return M - s;",
			errContains: "cannot subtract matrix and scalar at operand 2",
		},
		{
			name:        "matrix_vector_add",
			src:         "M = [[1.0, 0.0], [0.0, 1.0]]; v = [1.0, 2.0]; return M + v;",
			errContains: "cannot add matrix and vector at operand 2",
		},
		{
			name:        "matrix_vector_sub",
			src:         "M = [[1.0, 0.0], [0.0, 1.0]]; v = [1.0, 2.0]; return M - v;",
			errContains: "cannot subtract matrix and vector at operand 2",
		},
		{
			name:        "scalar_matrix_div",
			src:         "s = 2.0; M = [[1.0, 0.0], [0.0, 1.0]]; return s / M;",
			errContains: "division by matrix at operand 2 is unsupported; use matrix.inverse",
		},
		{
			name:        "vector_matrix_div",
			src:         "v = [1.0, 2.0]; M = [[1.0, 0.0], [0.0, 1.0]]; return v / M;",
			errContains: "division by matrix at operand 2 is unsupported; use matrix.inverse",
		},
		{
			name:        "matrix_matrix_div",
			src:         "M1 = [[1.0, 0.0], [0.0, 1.0]]; M2 = [[1.0, 0.0], [0.0, 1.0]]; return M1 / M2;",
			errContains: "matrix division at operand 2 is unsupported; multiply by matrix.inverse",
		},
		{
			name: "matrix_addition_shape_mismatch",
			src: `
M1 = [[1.0, 0.0], [0.0, 1.0]];
M2 = [[1.0, 0.0, 0.0], [0.0, 1.0, 0.0], [0.0, 0.0, 1.0]];
return M1 + M2;
`,
			errContains: "matrix addition dimension mismatch at operand 2: 2x2 vs 3x3",
		},
		{
			name: "matrix_subtraction_shape_mismatch",
			src: `
M1 = [[1.0, 0.0], [0.0, 1.0]];
M2 = [[1.0, 0.0, 0.0], [0.0, 1.0, 0.0], [0.0, 0.0, 1.0]];
return M1 - M2;
`,
			errContains: "matrix subtraction dimension mismatch at operand 2: 2x2 vs 3x3",
		},
		{
			name:        "vector_addition_length_mismatch",
			src:         "v1 = [1.0, 2.0]; v2 = [3.0, 4.0, 5.0]; return v1 + v2;",
			errContains: "vector addition dimension mismatch at operand 2: len 2 vs len 3",
		},
		{
			name:        "vector_subtraction_length_mismatch",
			src:         "v1 = [1.0, 2.0]; v2 = [3.0, 4.0, 5.0]; return v1 - v2;",
			errContains: "vector subtraction dimension mismatch at operand 2: len 2 vs len 3",
		},
		{
			name:        "vector_division_length_mismatch",
			src:         "v1 = [1.0, 2.0]; v2 = [3.0, 4.0, 5.0]; return v1 / v2;",
			errContains: "vector division dimension mismatch at operand 2: len 2 vs len 3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := runPipeline(tc.src)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.errContains)
			}
			if !strings.Contains(err.Error(), tc.errContains) {
				t.Errorf("expected error containing %q, got %q", tc.errContains, err.Error())
			}
		})
	}
}

func TestLower_Map_StructurallyIndexedCollection(t *testing.T) {
	src := `
doubleVal(x) { return x * 2.0; }
a = [10.0, 20.0];
xs = [a.x, a.y, a.x * 2.0, a.y * 2.0];
return map(doubleVal, xs);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("runPipeline error: %v", err)
	}

	ret := prog.Primary.Files[0].Returns[0]
	exprOp, ok := ret.T.(*program.Operand_Expression)
	if !ok || exprOp.Expression.Operation != program.Operation_COLLECT {
		t.Fatalf("expected COLLECT expression, got %T (%v)", ret.T, ret)
	}
	if len(exprOp.Expression.Operands) != 4 {
		t.Fatalf("expected 4 unrolled element calls, got %d", len(exprOp.Expression.Operands))
	}
}
