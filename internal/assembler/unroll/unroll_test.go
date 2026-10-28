package unroll_test

import (
	"fmt"
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/unroll"
	"gradientcontrol.com/omega/core/internal/closure"
	"gradientcontrol.com/omega/core/internal/collect"
	"gradientcontrol.com/omega/core/internal/link"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/types"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func runPipeline(src string) (*program.Program, bool, error) {
	srcMgr := &source.MockManager{}
	prog, err := collect.Program([]string{src}, srcMgr, parser.ParsePackage)
	if err != nil {
		return nil, false, err
	}
	if err := link.References(prog, srcMgr, parser.ParsePackage); err != nil {
		return nil, false, err
	}
	if err := closure.Transform(prog); err != nil {
		return nil, false, err
	}
	if err := link.References(prog, srcMgr, parser.ParsePackage); err != nil {
		return nil, false, err
	}
	if err := types.Analyze(prog); err != nil {
		return nil, false, err
	}
	modified, err := unroll.Transform(prog)
	if err != nil {
		return nil, false, err
	}
	if err := link.References(prog, srcMgr, parser.ParsePackage); err != nil {
		return nil, false, err
	}
	if err := types.Analyze(prog); err != nil {
		return nil, false, err
	}
	return prog, modified, nil
}

func getReturnExpr(prog *program.Program) (*program.Expression, error) {
	if prog == nil || prog.Primary == nil || len(prog.Primary.Files) == 0 {
		return nil, fmt.Errorf("empty program")
	}
	file := prog.Primary.Files[0]
	if len(file.Returns) == 0 {
		return nil, fmt.Errorf("no returns in file")
	}
	exprOp, ok := file.Returns[0].T.(*program.Operand_Expression)
	if !ok || exprOp.Expression == nil {
		return nil, fmt.Errorf("return is not an expression: %T", file.Returns[0].T)
	}
	return exprOp.Expression, nil
}

// 1. Wide vector arithmetic (> 4) is unrolled into Collect of scalar operations.
func TestUnroll_WideVectorArithmetic(t *testing.T) {
	src := `
wide = [1.0, 2.0, 3.0, 4.0, 5.0, 6.0];
return wide + wide;
`
	prog, modified, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	if !modified {
		t.Fatalf("expected modified=true")
	}

	expr, err := getReturnExpr(prog)
	if err != nil {
		t.Fatalf("getReturnExpr: %v", err)
	}
	if expr.Operation != program.Operation_COLLECT {
		t.Fatalf("expected return operation to be COLLECT, got %v", expr.Operation)
	}
	if len(expr.Operands) != 6 {
		t.Fatalf("expected 6 unrolled elements, got %d", len(expr.Operands))
	}
	for i, elem := range expr.Operands {
		elemExpr, ok := elem.T.(*program.Operand_Expression)
		if !ok || elemExpr.Expression.Operation != program.Operation_ADD {
			t.Errorf("element %d expected ADD expression, got %T", i, elem.T)
		}
	}
}

// 2. Canonical vectors (width 2, 3, 4) must NOT be unrolled.
func TestUnroll_CanonicalVectorPreserved(t *testing.T) {
	src := `
nrw = [1.0, 2.0, 3.0];
return nrw * 2.0;
`
	prog, modified, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	if modified {
		t.Fatalf("expected modified=false for vec3 operation")
	}

	expr, err := getReturnExpr(prog)
	if err != nil {
		t.Fatalf("getReturnExpr: %v", err)
	}
	if expr.Operation != program.Operation_MUL {
		t.Fatalf("expected Operation_MUL preserved, got %v", expr.Operation)
	}
}

// 3. Scalar broadcasting: wide * scalar and scalar * wide.
func TestUnroll_ScalarBroadcasting(t *testing.T) {
	src := `
foo(w, k) {
  return w * k;
}
return foo([1.0, 2.0, 3.0, 4.0, 5.0, 6.0], 2.0);
`
	prog, modified, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	if !modified {
		t.Fatalf("expected modified=true")
	}

	fn := prog.Primary.Files[0].FnTypes[0].Groups[0].Functions[0]
	ret := fn.Returns[0]
	expr, ok := ret.T.(*program.Operand_Expression)
	if !ok || expr.Expression.Operation != program.Operation_COLLECT {
		t.Fatalf("expected function return to be COLLECT, got %T", ret.T)
	}
	if len(expr.Expression.Operands) != 6 {
		t.Fatalf("expected 6 unrolled operands, got %d", len(expr.Expression.Operands))
	}
	for i, op := range expr.Expression.Operands {
		subExpr, ok := op.T.(*program.Operand_Expression)
		if !ok || subExpr.Expression.Operation != program.Operation_MUL {
			t.Fatalf("operand %d expected MUL, got %T", i, op.T)
		}
		// First operand should be index access w[i], second should be scalar k
		if len(subExpr.Expression.Operands) != 2 {
			t.Fatalf("operand %d expected 2 sub-operands, got %d", i, len(subExpr.Expression.Operands))
		}
		idxExpr, ok := subExpr.Expression.Operands[0].T.(*program.Operand_Expression)
		if !ok || idxExpr.Expression.Operation != program.Operation_INDEX {
			t.Errorf("operand %d first arg expected INDEX access, got %T", i, subExpr.Expression.Operands[0].T)
		}
		ref, ok := subExpr.Expression.Operands[1].T.(*program.Operand_Reference)
		if !ok || ref.Reference.Name != "k" {
			t.Errorf("operand %d second arg expected scalar reference 'k', got %T", i, subExpr.Expression.Operands[1].T)
		}
	}
}

// 4. Unary negation on wide collection unrolls into Collect of negated elements.
func TestUnroll_UnaryNegation(t *testing.T) {
	src := `
wide = [1.0, 2.0, 3.0, 4.0, 5.0, 6.0];
return -wide;
`
	prog, modified, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	if !modified {
		t.Fatalf("expected modified=true")
	}

	expr, err := getReturnExpr(prog)
	if err != nil {
		t.Fatalf("getReturnExpr: %v", err)
	}
	if expr.Operation != program.Operation_COLLECT {
		t.Fatalf("expected COLLECT, got %v", expr.Operation)
	}
	if len(expr.Operands) != 6 {
		t.Fatalf("expected 6 elements, got %d", len(expr.Operands))
	}
	for i, elem := range expr.Operands {
		elemExpr, ok := elem.T.(*program.Operand_Expression)
		if !ok || elemExpr.Expression.Operation != program.Operation_NEG {
			t.Errorf("element %d expected NEG expression, got %T", i, elem.T)
		}
	}
}

// 5. Math op calls (abs, min, max, sqrt) on wide collections.
func TestUnroll_MathOpCalls(t *testing.T) {
	src := `
wide = [1.0, 2.0, 3.0, 4.0, 5.0, 6.0];
return max(wide, 0.5);
`
	prog, modified, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	if !modified {
		t.Fatalf("expected modified=true")
	}

	expr, err := getReturnExpr(prog)
	if err != nil {
		t.Fatalf("getReturnExpr: %v", err)
	}
	if expr.Operation != program.Operation_COLLECT {
		t.Fatalf("expected COLLECT, got %v", expr.Operation)
	}
	for i, elem := range expr.Operands {
		elemExpr, ok := elem.T.(*program.Operand_Expression)
		if !ok || elemExpr.Expression.Operation != program.Operation_MAX {
			t.Errorf("element %d expected MAX expression, got %T", i, elem.T)
		}
	}
}

// 6. List of collections (matrices / tables) recurses: unrolls outer list, then inner vectors.
func TestUnroll_ListOfCollections(t *testing.T) {
	src := `
tbl = [[1.0, 2.0, 3.0, 4.0, 5.0, 6.0],
       [7.0, 8.0, 9.0, 10.0, 11.0, 12.0]];
return tbl / tbl;
`
	prog, modified, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	if !modified {
		t.Fatalf("expected modified=true")
	}

	expr, err := getReturnExpr(prog)
	if err != nil {
		t.Fatalf("getReturnExpr: %v", err)
	}
	if expr.Operation != program.Operation_COLLECT {
		t.Fatalf("expected outer COLLECT, got %v", expr.Operation)
	}
	if len(expr.Operands) != 2 {
		t.Fatalf("expected 2 row operands, got %d", len(expr.Operands))
	}
	for rowIdx, rowOp := range expr.Operands {
		rowExpr, ok := rowOp.T.(*program.Operand_Expression)
		if !ok || rowExpr.Expression.Operation != program.Operation_COLLECT {
			t.Fatalf("row %d expected inner COLLECT, got %T", rowIdx, rowOp.T)
		}
		if len(rowExpr.Expression.Operands) != 6 {
			t.Fatalf("row %d expected 6 scalar elements, got %d", rowIdx, len(rowExpr.Expression.Operands))
		}
	}
}

// 7. Literal unpacking: [a, b, c, d, e] + [f, g, h, i, j] unpacks literals without index accesses.
func TestUnroll_LiteralUnpacking(t *testing.T) {
	src := `
return [1.0, 2.0, 3.0, 4.0, 5.0, 6.0] + [10.0, 20.0, 30.0, 40.0, 50.0, 60.0];
`
	prog, modified, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	if !modified {
		t.Fatalf("expected modified=true")
	}

	expr, err := getReturnExpr(prog)
	if err != nil {
		t.Fatalf("getReturnExpr: %v", err)
	}
	for i, elem := range expr.Operands {
		elemExpr := elem.T.(*program.Operand_Expression).Expression
		// Operands of the ADD should be literal values, NOT index accesses
		val0, ok0 := elemExpr.Operands[0].T.(*program.Operand_Value)
		val1, ok1 := elemExpr.Operands[1].T.(*program.Operand_Value)
		if !ok0 || !ok1 {
			t.Errorf("element %d expected unpacked literal values, got %T and %T", i, elemExpr.Operands[0].T, elemExpr.Operands[1].T)
		} else {
			if val0.Value != float64(i+1) || val1.Value != float64((i+1)*10) {
				t.Errorf("element %d expected %v and %v, got %v and %v", i, float64(i+1), float64((i+1)*10), val0.Value, val1.Value)
			}
		}
	}
}

// 8. Compound expression hoisting: f() + wide synthesizes a temporary assignment for f().
func TestUnroll_CompoundExpressionHoisting(t *testing.T) {
	src := `
makeRow() {
  return [1.0, 2.0, 3.0, 4.0, 5.0, 6.0];
}
return makeRow() * 2.0;
`
	prog, modified, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	if !modified {
		t.Fatalf("expected modified=true")
	}

	file := prog.Primary.Files[0]
	// Should have hoisted makeRow() call into a temporary assignment in file.Assignments
	foundHoisted := false
	for _, assign := range file.Assignments {
		if assign.Expression != nil {
			if fnCall, ok := assign.Expression.T.(*program.Operand_FnCall); ok && fnCall.FnCall.Reference.Name == "makeRow" {
				foundHoisted = true
				break
			}
		}
	}
	if !foundHoisted {
		t.Errorf("expected makeRow() call to be hoisted into a temporary assignment")
	}
}

// 9. Scalar broadcasting with scalar on left: scalar * wide.
func TestUnroll_ScalarBroadcasting_LeftScalar(t *testing.T) {
	src := `
foo(k, w) {
  return k * w;
}
return foo(2.0, [1.0, 2.0, 3.0, 4.0, 5.0, 6.0]);
`
	prog, modified, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	if !modified {
		t.Fatalf("expected modified=true")
	}

	fn := prog.Primary.Files[0].FnTypes[0].Groups[0].Functions[0]
	ret := fn.Returns[0]
	expr, ok := ret.T.(*program.Operand_Expression)
	if !ok || expr.Expression.Operation != program.Operation_COLLECT {
		t.Fatalf("expected function return to be COLLECT, got %T", ret.T)
	}
	if len(expr.Expression.Operands) != 6 {
		t.Fatalf("expected 6 unrolled operands, got %d", len(expr.Expression.Operands))
	}
	for i, op := range expr.Expression.Operands {
		subExpr, ok := op.T.(*program.Operand_Expression)
		if !ok || subExpr.Expression.Operation != program.Operation_MUL {
			t.Fatalf("operand %d expected MUL, got %T", i, op.T)
		}
		// First operand should be scalar reference 'k', second should be index access w[i]
		ref, ok := subExpr.Expression.Operands[0].T.(*program.Operand_Reference)
		if !ok || ref.Reference.Name != "k" {
			t.Errorf("operand %d first arg expected scalar reference 'k', got %T", i, subExpr.Expression.Operands[0].T)
		}
		idxExpr, ok := subExpr.Expression.Operands[1].T.(*program.Operand_Expression)
		if !ok || idxExpr.Expression.Operation != program.Operation_INDEX {
			t.Errorf("operand %d second arg expected INDEX access, got %T", i, subExpr.Expression.Operands[1].T)
		}
	}
}

// 10. Scalar broadcasting in op calls: max(wide, scalar).
func TestUnroll_ScalarBroadcasting_OpCall(t *testing.T) {
	src := `
foo(w, k) {
  return max(w, k);
}
return foo([1.0, 2.0, 3.0, 4.0, 5.0, 6.0], 0.0);
`
	prog, modified, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	if !modified {
		t.Fatalf("expected modified=true")
	}

	fn := prog.Primary.Files[0].FnTypes[0].Groups[0].Functions[0]
	ret := fn.Returns[0]
	expr, ok := ret.T.(*program.Operand_Expression)
	if !ok || expr.Expression.Operation != program.Operation_COLLECT {
		t.Fatalf("expected function return to be COLLECT, got %T", ret.T)
	}
	if len(expr.Expression.Operands) != 6 {
		t.Fatalf("expected 6 unrolled operands, got %d", len(expr.Expression.Operands))
	}
	for i, op := range expr.Expression.Operands {
		subExpr, ok := op.T.(*program.Operand_Expression)
		if !ok || subExpr.Expression.Operation != program.Operation_MAX {
			t.Fatalf("operand %d expected MAX, got %T", i, op.T)
		}
		idxExpr, ok := subExpr.Expression.Operands[0].T.(*program.Operand_Expression)
		if !ok || idxExpr.Expression.Operation != program.Operation_INDEX {
			t.Errorf("operand %d first arg expected INDEX access, got %T", i, subExpr.Expression.Operands[0].T)
		}
		ref, ok := subExpr.Expression.Operands[1].T.(*program.Operand_Reference)
		if !ok || ref.Reference.Name != "k" {
			t.Errorf("operand %d second arg expected scalar reference 'k', got %T", i, subExpr.Expression.Operands[1].T)
		}
	}
}

// 11. Single-element list arithmetic unrolls into scalar math wrapped in COLLECT.
func TestUnroll_SingleElementListArithmetic(t *testing.T) {
	src := `
foo(a, b) {
  return a + b;
}
return foo([1.0], [2.0]);
`
	prog, modified, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	if !modified {
		t.Fatalf("expected modified=true")
	}

	fn := prog.Primary.Files[0].FnTypes[0].Groups[0].Functions[0]
	ret := fn.Returns[0]
	expr, ok := ret.T.(*program.Operand_Expression)
	if !ok || expr.Expression.Operation != program.Operation_COLLECT {
		t.Fatalf("expected COLLECT, got %v", expr.Expression.Operation)
	}
	if len(expr.Expression.Operands) != 1 {
		t.Fatalf("expected 1 unrolled operand, got %d", len(expr.Expression.Operands))
	}
	subExpr, ok := expr.Expression.Operands[0].T.(*program.Operand_Expression)
	if !ok || subExpr.Expression.Operation != program.Operation_ADD {
		t.Fatalf("expected ADD, got %T", expr.Expression.Operands[0].T)
	}
}

// 12. Single-element list normalize desugars into [sign(x)].
func TestUnroll_SingleElementListNormalize(t *testing.T) {
	src := `
foo(a) {
  return normalize(a);
}
return foo([2.0]);
`
	prog, modified, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	if !modified {
		t.Fatalf("expected modified=true")
	}

	fn := prog.Primary.Files[0].FnTypes[0].Groups[0].Functions[0]
	ret := fn.Returns[0]
	expr, ok := ret.T.(*program.Operand_Expression)
	if !ok || expr.Expression.Operation != program.Operation_COLLECT {
		t.Fatalf("expected COLLECT, got %v", expr.Expression.Operation)
	}
	if len(expr.Expression.Operands) != 1 {
		t.Fatalf("expected 1 unrolled operand, got %d", len(expr.Expression.Operands))
	}
	subCall, ok := expr.Expression.Operands[0].T.(*program.Operand_FnCall)
	if !ok || subCall.FnCall.Reference.Name != "sign" {
		t.Fatalf("expected sign call, got %T", expr.Expression.Operands[0].T)
	}
}
