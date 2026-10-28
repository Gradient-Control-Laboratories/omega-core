package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/meshplan"
	"gradientcontrol.com/omega/core/internal/assembler/topofold"
	"gradientcontrol.com/omega/core/internal/assembler/warpfold"
	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/internal/source"
	src_test "gradientcontrol.com/omega/core/internal/srctest"
	"gradientcontrol.com/omega/core/internal/types"
	"gradientcontrol.com/omega/core/program"
)

func TestAssembleSingle(t *testing.T) {
	pkgMgr := &source.MockManager{}

	res, err := Assemble([]string{"return 42.0;"}, pkgMgr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res == nil || res.Program == nil {
		t.Fatalf("expected non-nil program")
	}
}

func TestAssembleLegacyOmegaVectorConstants(t *testing.T) {
	pkgMgr := source.NewLocalManager("../../../common/source/package")
	_, err := Assemble([]string{`import . "omega";
import "vector";
return length(zero + dirX + dirY + dirZ + vector.zero);`}, pkgMgr)
	if err != nil {
		t.Fatalf("legacy omega constants: %v", err)
	}
}

func TestAssemble_NamedPrimaryPackage(t *testing.T) {
	pkgMgr := source.NewLocalManager("../../../common/source/package")
	src := `
import "vector";
addOne(x) { return x + 1.0; }
a = addOne(vector.zero.x);
return a;
`
	res, err := Assemble([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("assemble failed: %v", err)
	}

	// Explicitly assign primary package path and run reference linking, types, and folding
	res.Program.Primary.Path = "test/root/pkg"
	if err := types.Analyze(res.Program); err != nil {
		t.Fatalf("types.Analyze failed with named primary path: %v", err)
	}
	if _, err := eval.Fold(res.Program, nil); err != nil {
		t.Fatalf("eval.Fold failed with named primary path: %v", err)
	}

	if res.Program.Primary.Path != "test/root/pkg" {
		t.Errorf("expected primary path 'test/root/pkg', got %q", res.Program.Primary.Path)
	}
}

func TestAssemblePropertyExpand(t *testing.T) {
	pkgMgr := &source.MockManager{}
	res, err := Assemble([]string{"dummy = 1.0; return time.dtime;"}, pkgMgr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res == nil || res.Program == nil {
		t.Fatalf("expected non-nil program")
	}

	if len(res.Program.Primary.Files) == 0 || len(res.Program.Primary.Files[0].Returns) == 0 {
		t.Fatalf("expected at least one return")
	}

	retOp := res.Program.Primary.Files[0].Returns[0]
	if hasDeriv(retOp) {
		t.Errorf("expected DERIV operator to be expanded, but found: %s", retOp.PrettyPrint(0))
	}
}

func hasDeriv(op *program.Operand) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		if t.Expression.Operation == program.Operation_DERIV {
			return true
		}
		for _, sub := range t.Expression.Operands {
			if hasDeriv(sub) {
				return true
			}
		}
	case *program.Operand_FnCall:
		for _, a := range t.FnCall.Arguments {
			if hasDeriv(a) {
				return true
			}
		}
	}
	return false
}

func TestAssembleClosureErrors(t *testing.T) {
	pkgMgr := &source.MockManager{}

	tests := []struct {
		name        string
		src         string
		errContains string
	}{
		{
			name: "mixed_return",
			src: `
bad(k) { return (x) { return x + k; }, 42.0; }
f, val = bad(1.0);
return f(val);
`,
			errContains: "a function that returns a function cannot return anything else",
		},
		{
			name: "multi_function_return",
			src: `
bad(k) { return (x) { return x + k; }, (y) { return y * k; }; }
f1, f2 = bad(1.0);
return f1(f2(1.0));
`,
			errContains: "a function that returns a function cannot return anything else",
		},
		{
			name: "uncalled_primary_return",
			src: `
scaleBy(k) { return (a) { return a * k; }; }
return scaleBy(0.5);
`,
			errContains: "cannot return a function; perhaps forgot parentheses () after function name?",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Assemble([]string{tc.src}, pkgMgr)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.errContains)
			}
			if !strings.Contains(err.Error(), tc.errContains) {
				t.Fatalf("expected error containing %q, got %q", tc.errContains, err.Error())
			}
		})
	}
}

func TestAssemble(t *testing.T) {
	src_test.RunAllSourceTests(t, "../internal/testdata/*.gsl", testAssembleOne)
}

func testAssembleOne(t *testing.T, srcFile string) {
	src_test.RunSourceTest(t, srcFile, runTestAssembleOne, checkTestAssembleOne, true)
}

func runTestAssembleOne(t *testing.T, srcFile string, src []byte) (*program.Program, error) {
	pkgMgr := source.NewLocalManager("../../../common/source/package")
	res, err := Assemble([]string{string(src)}, pkgMgr)
	if err != nil {
		return nil, err
	}
	return res.Program, nil
}

func checkTestAssembleOne(t *testing.T, srcFile string, src []byte, result *program.Program) error {
	goldenFile := strings.TrimSuffix(srcFile, filepath.Ext(srcFile)) + ".assembled"

	actualContent := result.PrettyPrint(0)

	_, err := os.Stat(goldenFile)
	if os.IsNotExist(err) || os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.WriteFile(goldenFile, []byte(actualContent), 0644); err != nil {
			return fmt.Errorf("failed to write golden file: %w", err)
		}
		t.Logf("Generated golden file: %s", goldenFile)
		return nil
	}

	expectedBytes, err := os.ReadFile(goldenFile)
	if err != nil {
		return fmt.Errorf("failed to read golden file %s: %w", goldenFile, err)
	}

	expectedContent := string(expectedBytes)
	actualStr := strings.TrimSpace(actualContent)
	expectedStr := strings.TrimSpace(expectedContent)
	if actualStr != expectedStr {
		return fmt.Errorf("assembled content mismatch for %s\nExpected:\n%s\nActual:\n%s", goldenFile, expectedStr, actualStr)
	}
	return nil
}

type mockGeometryResolver struct {
	volume float64
}

func (m *mockGeometryResolver) GetMesh(slot int) (topofold.MeshGeometry, bool) {
	return nil, false
}

func (m *mockGeometryResolver) GetAnnotatedMesh(slot int) (topofold.AnnotatedMesh, bool, error) {
	return nil, false, nil
}

type mockWarpResolver struct {
	slot int
}

var _ warpfold.WarpResolver = (*mockWarpResolver)(nil)
var _ topofold.GeometryResolver = (*mockGeometryResolver)(nil)

func (m *mockWarpResolver) ResolveWarpedSlot(meshSlot int, pkgDir, warpPath string, scale float64) (int, error) {
	return m.slot, nil
}

func TestAssembleOptions(t *testing.T) {
	pkgMgr := source.NewLocalManager("../../../common/source/package")

	// 1. Test PrimaryPackage option
	res, err := Assemble([]string{"return 123.0;"}, pkgMgr, PrimaryPackage("trees/WO-1"))
	if err != nil {
		t.Fatalf("PrimaryPackage option failed: %v", err)
	}
	if res == nil || res.Program == nil || res.Program.Primary == nil || res.Program.Primary.Path != "trees/WO-1" {
		t.Fatalf("expected program with primary package path 'trees/WO-1', got: %v", res)
	}

	// 2. Test GeometryResolver option
	geomResolver := &mockGeometryResolver{volume: 100.0}
	res2, err := Assemble([]string{"return 42.0;"}, pkgMgr, GeometryResolver(geomResolver))
	if err != nil {
		t.Fatalf("GeometryResolver option failed: %v", err)
	}
	if res2 == nil || res2.Program == nil {
		t.Fatalf("expected non-nil program")
	}

	// 3. Test WarpResolver option
	warpRes := &mockWarpResolver{slot: 7}
	res3, err := Assemble([]string{"return 42.0;"}, pkgMgr, WarpResolver(warpRes))
	if err != nil {
		t.Fatalf("WarpResolver option failed: %v", err)
	}
	if res3 == nil || res3.Program == nil {
		t.Fatalf("expected non-nil program")
	}

	// 4. Test DisableMeshSlotFolding option
	res4, err := Assemble([]string{"return 42.0;"}, pkgMgr, DisableMeshSlotFolding())
	if err != nil {
		t.Fatalf("DisableMeshSlotFolding option failed: %v", err)
	}
	if res4 == nil || res4.Program == nil {
		t.Fatalf("expected non-nil program")
	}

	// 5. Test DisableDeadcodeRemoval option
	srcWithDead := "unused() { return 100.0; }\nreturn 42.0;\n"
	res5, err := Assemble([]string{srcWithDead}, pkgMgr, DisableDeadcodeRemoval())
	if err != nil {
		t.Fatalf("DisableDeadcodeRemoval option failed: %v", err)
	}
	hasUnused := false
	for _, f := range res5.Program.Primary.Files {
		for _, fn := range f.FnTypes {
			if fn.Name == "unused" {
				hasUnused = true
			}
		}
	}
	if !hasUnused {
		t.Errorf("expected unused function to be retained when DisableDeadcodeRemoval is set")
	}

	// Verify default Assemble removes unused function
	resDefault, err := Assemble([]string{srcWithDead}, pkgMgr)
	if err != nil {
		t.Fatalf("default Assemble failed: %v", err)
	}
	hasUnusedDefault := false
	for _, f := range resDefault.Program.Primary.Files {
		for _, fn := range f.FnTypes {
			if fn.Name == "unused" {
				hasUnusedDefault = true
			}
		}
	}
	if hasUnusedDefault {
		t.Errorf("expected unused function to be removed by default Assemble")
	}
}

func TestAssembleIterativeCascade(t *testing.T) {
	pkgMgr := source.NewLocalManager("../../../common/source/package")
	src := `
import font "text";
return font.left("a", [0, 0], 1, 0.04, 0.3);
`
	res, err := Assemble([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	for _, ret := range res.Program.Primary.Files[0].Returns {
		t.Logf("Primary Return: %s", ret.PrettyPrint(0))
	}
	for _, dep := range res.Program.Dependencies {
		t.Logf("Dep package: %s, FnTypes: %d", dep.Path, len(dep.Files[0].FnTypes))
		for _, ft := range dep.Files[0].FnTypes {
			t.Logf("Dep %s FnType: %s", dep.Path, ft.Name)
			if strings.Contains(ft.Name, "dispatch") {
				t.Errorf("found dispatch function in dep %s: %s", dep.Path, ft.Name)
			}
		}
	}
}

func TestAssembleTextGlyphs(t *testing.T) {
	pkgMgr := source.NewLocalManager("../../../common/source/package")
	content, err := os.ReadFile("../../../common/web/example/Text Glyphs.gsl")
	if err != nil {
		t.Fatalf("failed to read Text Glyphs.gsl: %v", err)
	}
	res, err := Assemble([]string{string(content)}, pkgMgr)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	for _, assign := range res.Program.Primary.Files[0].Assignments {
		for _, v := range assign.Variables {
			t.Logf("Assign: %s = %s", v.Name, assign.Expression.PrettyPrint(0))
		}
	}
	for _, ft := range res.Program.Primary.Files[0].FnTypes {
		for _, g := range ft.Groups {
			t.Logf("Primary FnType: %s, ParamCount: %d, NumFns: %d, Params: %d", ft.Name, g.ParamCount, len(g.Functions), len(g.Functions[0].Parameters))
		}
	}
	for _, dep := range res.Program.Dependencies {
		t.Logf("Dep package: %s, FnTypes: %d", dep.Path, len(dep.Files[0].FnTypes))
		for _, ft := range dep.Files[0].FnTypes {
			if strings.Contains(ft.Name, "dispatch") {
				t.Errorf("found dispatch function in dep %s: %s", dep.Path, ft.Name)
			}
		}
	}
}

func TestAssemble_PrimaryPackageMeshSlotFolding(t *testing.T) {
	pkgMgr := source.NewLocalManager("../../../common/source/package")
	src := `
importId = importStep "A356-H01.STEP";
partXf = [[1, 0, 0, 0], [0, 1, 0, 0], [0, 0, 1, 0], [0, 0, 0, 1]];
msh = mesh(importId, partXf);
return msh;
`
	res, err := Assemble([]string{src}, pkgMgr, PrimaryPackage("branches/A356_H01"))
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	if res.Program.Primary == nil || res.Program.Primary.Path != "branches/A356_H01" {
		t.Fatalf("expected primary package path 'branches/A356_H01', got: %v", res.Program)
	}

	// Verify importStep expression was folded to a constant value
	for _, a := range res.Program.Primary.Files[0].Assignments {
		if len(a.Variables) == 1 && a.Variables[0].Name == "importId" {
			if _, ok := a.Expression.T.(*program.Operand_Value); !ok {
				t.Fatalf("expected importId to fold to Operand_Value, got: %s", a.Expression.PrettyPrint(0))
			}
		}
	}
}

func TestAssembleWithSlots(t *testing.T) {
	pkgMgr := source.NewLocalManager("../../../common/source/package")
	src := `
importId = importStep "A356-H01.STEP";
partXf = [[1, 0, 0, 0], [0, 1, 0, 0], [0, 0, 1, 0], [0, 0, 0, 1]];
msh = mesh(importId, partXf);
return msh;
`
	res, err := Assemble([]string{src}, pkgMgr, PrimaryPackage("branches/A356_H01"))
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	if res == nil || res.Program == nil {
		t.Fatalf("expected non-nil program")
	}
	if len(res.Slots) != 1 {
		t.Fatalf("expected 1 import slot, got %d", len(res.Slots))
	}
	if res.Slots[0].Slot != 0 || res.Slots[0].Path != "A356-H01.STEP" {
		t.Errorf("unexpected slot entry: %+v", res.Slots[0])
	}
}

func TestAssembleTreesManifold(t *testing.T) {
	commonRoot := "../../../common/source/package"
	docRoot := "../../../../dm/documents"
	src, err := os.ReadFile(docRoot + "/trees/manifold/manifold.gsl")
	if err != nil {
		t.Fatalf("read manifold: %v", err)
	}
	srcMgr := source.NewLocalManager(commonRoot, docRoot)
	res, err := Assemble([]string{string(src)}, srcMgr)
	if err != nil {
		t.Fatalf("Assemble trees/manifold failed: %+v", err)
	}
	if res == nil || res.Program == nil {
		t.Fatalf("expected non-nil program")
	}
}

func TestAssembleS304_C02(t *testing.T) {
	commonRoot := "../../../common/source/package"
	docRoot := "../../../../dm/documents"
	src, err := os.ReadFile(docRoot + "/branches/S304_C02/S304_C02.gsl")
	if err != nil {
		t.Fatalf("read S304_C02: %v", err)
	}
	srcMgr := source.NewLocalManager(commonRoot, docRoot)
	res, err := Assemble([]string{string(src)}, srcMgr, PrimaryPackage("branches/S304_C02"))
	if err != nil {
		t.Fatalf("Assemble branches/S304_C02 failed: %+v", err)
	}
	if err := types.Analyze(res.Program); err != nil {
		t.Fatalf("types.Analyze failed: %+v", err)
	}
}

func TestAssembleS304_T001(t *testing.T) {
	commonRoot := "../../../common/source/package"
	docRoot := "../../../../dm/documents"
	src, err := os.ReadFile(docRoot + "/trees/S304-T001/S304-T001.gsl")
	if err != nil {
		t.Fatalf("read S304-T001: %v", err)
	}
	srcMgr := source.NewLocalManager(commonRoot, docRoot)
	res, err := Assemble([]string{string(src)}, srcMgr, PrimaryPackage("trees/S304-T001"))
	if err != nil {
		t.Fatalf("Assemble trees/S304-T001 failed: %+v", err)
	}
	if err := types.Analyze(res.Program); err != nil {
		t.Fatalf("types.Analyze failed: %+v", err)
	}
}

func getOperandConstValue(elem *program.Operand) (float64, bool) {
	if val, ok := elem.T.(*program.Operand_Value); ok {
		return val.Value, true
	}
	if len(elem.ConstValues) > 0 {
		return elem.ConstValues[0], true
	}
	if elem.Type != nil && elem.Type.ConstValue != nil {
		if val, ok := elem.Type.ConstValue.T.(*program.Operand_Value); ok {
			return val.Value, true
		}
		if len(elem.Type.ConstValue.ConstValues) > 0 {
			return elem.Type.ConstValue.ConstValues[0], true
		}
	}
	return 0, false
}

func TestAssemble_WideCollectionUnrolledAndFolded(t *testing.T) {
	pkgMgr := &source.MockManager{}
	src := `
return [1.0, 2.0, 3.0, 4.0, 5.0, 6.0] * 2.0;
`
	res, err := Assemble([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	if res == nil || res.Program == nil || res.Program.Primary == nil {
		t.Fatalf("expected non-nil assembled program")
	}

	file := res.Program.Primary.Files[0]
	if len(file.Returns) == 0 {
		t.Fatalf("expected at least 1 return statement")
	}
	retOp := file.Returns[0]
	expr, ok := retOp.T.(*program.Operand_Expression)
	if !ok || expr.Expression.Operation != program.Operation_COLLECT {
		t.Fatalf("expected return operand to be COLLECT, got %T (%v)", retOp.T, retOp)
	}
	if len(expr.Expression.Operands) != 6 {
		t.Fatalf("expected 6 unrolled elements, got %d", len(expr.Expression.Operands))
	}
	expectedValues := []float64{2.0, 4.0, 6.0, 8.0, 10.0, 12.0}
	for i, elem := range expr.Expression.Operands {
		val, ok := getOperandConstValue(elem)
		if !ok {
			t.Fatalf("element %d expected constant value, got %T (%v)", i, elem.T, elem)
		}
		if val != expectedValues[i] {
			t.Errorf("element %d: expected %v, got %v", i, expectedValues[i], val)
		}
	}
}

func TestAssemble_CanonicalVectorPreserved(t *testing.T) {
	pkgMgr := &source.MockManager{}
	src := `
return p * 2.0;
`
	res, err := Assemble([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	if res == nil || res.Program == nil || res.Program.Primary == nil {
		t.Fatalf("expected non-nil assembled program")
	}

	file := res.Program.Primary.Files[0]
	if len(file.Returns) == 0 {
		t.Fatalf("expected at least 1 return statement")
	}
	retOp := file.Returns[0]
	t.Logf("retOp: %s (T=%T)", retOp.PrettyPrint(0), retOp.T)
	fc, ok := retOp.T.(*program.Operand_FnCall)
	if !ok || fc.FnCall == nil || fc.FnCall.Reference == nil || fc.FnCall.Reference.Name != "multiplyScalar" {
		t.Fatalf("expected vector.multiplyScalar call, got %s", retOp.PrettyPrint(0))
	}
}

func TestAssemble_InlinedFunctionExposesWideCollection(t *testing.T) {
	pkgMgr := &source.MockManager{}
	src := `
scaleRow(r, k) {
  return r * k;
}
return scaleRow([1.0, 2.0, 3.0, 4.0, 5.0, 6.0], 3.0);
`
	res, err := Assemble([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	if res == nil || res.Program == nil || res.Program.Primary == nil {
		t.Fatalf("expected non-nil assembled program")
	}

	file := res.Program.Primary.Files[0]
	if len(file.Returns) == 0 {
		t.Fatalf("expected at least 1 return statement")
	}
	retOp := file.Returns[0]
	expr, ok := retOp.T.(*program.Operand_Expression)
	if !ok || expr.Expression.Operation != program.Operation_COLLECT {
		t.Fatalf("expected inlined function call return to be COLLECT, got %T (%v)", retOp.T, retOp)
	}
	if len(expr.Expression.Operands) != 6 {
		t.Fatalf("expected 6 unrolled elements, got %d", len(expr.Expression.Operands))
	}
	expectedValues := []float64{3.0, 6.0, 9.0, 12.0, 15.0, 18.0}
	for i, elem := range expr.Expression.Operands {
		val, ok := getOperandConstValue(elem)
		if !ok {
			t.Fatalf("element %d expected constant value, got %T: %+v", i, elem.T, elem)
		}
		if val != expectedValues[i] {
			t.Errorf("element %d: expected %v, got %v", i, expectedValues[i], val)
		}
	}
}

func TestAssemble_ColorAndMaterialLoweredIntoReturns(t *testing.T) {
	pkgMgr := &source.MockManager{}
	src := `
target = [1.0, 2.0, 3.0];
target.color = [1.0, 0.0, 0.0];
target.material = [0.5, 0.8, 0.1];
return target;
`
	res, err := Assemble([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	if res == nil || res.Program == nil || res.Program.Primary == nil {
		t.Fatalf("expected non-nil assembled program")
	}

	colorReturns := res.Program.ColorReturns()
	if len(colorReturns) != 1 {
		t.Fatalf("expected 1 color return, got %d", len(colorReturns))
	}
	matReturns := res.Program.MaterialReturns()
	if len(matReturns) != 1 {
		t.Fatalf("expected 1 material return, got %d", len(matReturns))
	}

	file := res.Program.Primary.Files[0]
	if len(file.ColorReturns) != 1 {
		t.Fatalf("expected 1 color return on file, got %d", len(file.ColorReturns))
	}
	if len(file.MaterialReturns) != 1 {
		t.Fatalf("expected 1 material return on file, got %d", len(file.MaterialReturns))
	}

	if colorReturns[0].Type == nil {
		t.Errorf("expected colorReturn to have Type resolved")
	}
	if matReturns[0].Type == nil {
		t.Errorf("expected matReturn to have Type resolved")
	}
}

func TestAssemble_OnlyColorLowered(t *testing.T) {
	pkgMgr := &source.MockManager{}
	src := `
ball = [1.0, 2.0, 3.0];
ball.color = [0.0, 1.0, 0.0];
return ball;
`
	res, err := Assemble([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	if len(res.Program.ColorReturns()) != 1 {
		t.Fatalf("expected 1 color return, got %d", len(res.Program.ColorReturns()))
	}
	if len(res.Program.MaterialReturns()) != 0 {
		t.Fatalf("expected 0 material returns, got %d", len(res.Program.MaterialReturns()))
	}
}

func TestAssemble_OnlyMaterialLowered(t *testing.T) {
	pkgMgr := &source.MockManager{}
	src := `
boxTarget = [1.0, 2.0, 3.0];
boxTarget.material = [0.2, 0.4, 0.6];
return boxTarget;
`
	res, err := Assemble([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	if len(res.Program.ColorReturns()) != 0 {
		t.Fatalf("expected 0 color returns, got %d", len(res.Program.ColorReturns()))
	}
	if len(res.Program.MaterialReturns()) != 1 {
		t.Fatalf("expected 1 material return, got %d", len(res.Program.MaterialReturns()))
	}
}

func TestAssemble_NeitherColorNorMaterial(t *testing.T) {
	pkgMgr := &source.MockManager{}
	src := `
return 42.0;
`
	res, err := Assemble([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	if len(res.Program.ColorReturns()) != 0 {
		t.Fatalf("expected 0 color returns, got %d", len(res.Program.ColorReturns()))
	}
	if len(res.Program.MaterialReturns()) != 0 {
		t.Fatalf("expected 0 material returns, got %d", len(res.Program.MaterialReturns()))
	}
}

func TestAssemble_SphereScene(t *testing.T) {
	pkgMgr := &source.MockManager{}
	src := "// Scene Size: 5\n// Scene Grid: 10\nreturn sphere(0.8);"
	res, err := Assemble([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	sc := res.Program.Scene()
	if sc == nil || sc.Size != 5 || sc.Grid != 0.5 {
		t.Fatalf("expected scene size 5, grid 0.5, got %+v", sc)
	}
}

func TestAssemble_SliderParameterConstantFoldingPreserved(t *testing.T) {
	pkgMgr := &source.MockManager{}
	src := `
@range(1.0, 5.0)
radius = 2.0;

scale = radius * 2.0;
staticVal = 3.0 * 4.0;

return scale + staticVal;
`
	res, err := Assemble([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	var scaleAssign, staticAssign *program.Assignment
	for _, a := range res.Program.Primary.Files[0].Assignments {
		if len(a.Variables) > 0 {
			if a.Variables[0].Name == "scale" {
				scaleAssign = a
			}
			if a.Variables[0].Name == "staticVal" {
				staticAssign = a
			}
		}
	}

	if staticAssign == nil {
		t.Fatalf("expected staticVal assignment")
	}
	if len(staticAssign.Expression.ConstValues) == 0 || staticAssign.Expression.ConstValues[0] != 12.0 {
		t.Fatalf("expected staticVal to be folded to 12.0, got %+v", staticAssign.Expression)
	}

	if scaleAssign == nil {
		t.Fatalf("expected scale assignment")
	}
	if len(scaleAssign.Expression.ConstValues) > 0 {
		t.Fatalf("expected scale to NOT have ConstValues stamped, got %+v", scaleAssign.Expression.ConstValues)
	}
}

func TestAssemble_UniversalReduceRetained(t *testing.T) {
	pkgMgr := &source.MockManager{}
	src := `
scan(q) {
  return reduce([1000, 0], (best, i) {
    d = length(q - [i, 0, 0]);
    return d < best.x ? [d, i] : best;
  }, [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]);
}
return sphere(p, 10 + scan(p).x);
`
	res, err := Assemble([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	hasReduce := false
	var checkOp func(op *program.Operand)
	checkOp = func(op *program.Operand) {
		if op == nil {
			return
		}
		if e := op.GetExpression(); e != nil {
			if e.Operation == program.Operation_REDUCE {
				hasReduce = true
			}
			for _, sub := range e.Operands {
				checkOp(sub)
			}
		}
		if fc := op.GetFnCall(); fc != nil {
			for _, arg := range fc.Arguments {
				checkOp(arg)
			}
		}
	}
	for _, f := range res.Program.Primary.Files {
		for _, ft := range f.FnTypes {
			for _, g := range ft.Groups {
				for _, fn := range g.Functions {
					for _, a := range fn.Assignments {
						checkOp(a.Expression)
					}
					for _, r := range fn.Returns {
						checkOp(r)
					}
				}
			}
		}
	}
	if !hasReduce {
		t.Errorf("expected Assemble without options to retain Operation_REDUCE in AST, but none found")
	}
}

func TestAssemble_MapUnrollSlotsFromComputedCollection(t *testing.T) {
	pkgMgr := source.NewLocalManager("../../../common/source/package")
	src := `
makeMesh(id) {
	partXf = [[1, 0, 0, 0], [0, 1, 0, 0], [0, 0, 1, 0], [0, 0, 0, 1]];
	return mesh(id, partXf);
}
id1 = importStep "A356-H01.STEP";
id2 = importStep "B123-H02.STEP";
ids = [id1] + [id2];
return map(makeMesh, ids);
`
	res, err := Assemble([]string{src}, pkgMgr, PrimaryPackage("branches/A356_H01"))
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	if len(res.Slots) != 2 {
		t.Fatalf("expected 2 mesh slots extracted after map unroll, got %d", len(res.Slots))
	}
}

func TestAssemble_MapUnrollInlinedCollection(t *testing.T) {
	pkgMgr := &source.MockManager{}
	src := `
makeCollection() {
	a = [10.0, 20.0];
	return [a.x, a.y, a.x * 2.0, a.y * 2.0];
}
doubleVal(x) { return x * 2.0; }
xs = makeCollection();
return map(doubleVal, xs);
`
	res, err := Assemble([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	ret := res.Program.Primary.Files[0].Returns[0]
	hasMap := false
	var checkOp func(op *program.Operand)
	checkOp = func(op *program.Operand) {
		if op == nil {
			return
		}
		if e := op.GetExpression(); e != nil {
			if e.Operation == program.Operation_MAP {
				hasMap = true
			}
			for _, sub := range e.Operands {
				checkOp(sub)
			}
		}
		if fc := op.GetFnCall(); fc != nil {
			for _, arg := range fc.Arguments {
				checkOp(arg)
			}
		}
	}
	checkOp(ret)
	if hasMap {
		t.Errorf("expected map to be lowered and unrolled, but Operation_MAP remains in return")
	}
}

func TestAssemble_SliderVariablePreservedInPreLoweringFold(t *testing.T) {
	pkgMgr := &source.MockManager{}
	src := `
radius = 25.0;
radius.min = 5.0;
radius.max = 50.0;
scaled = radius * 2.0;
return sphere(scaled);
`
	res, err := Assemble([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	foundRadiusVar := false
	for _, a := range res.Program.Primary.Files[0].Assignments {
		for _, v := range a.Variables {
			if v.Name == "radius" {
				foundRadiusVar = true
			}
		}
	}
	if !foundRadiusVar {
		t.Errorf("expected slider variable 'radius' to be preserved, but it was eliminated")
	}
}

func TestAssemble_ExtractUI_Populated(t *testing.T) {
	pkgMgr := &source.MockManager{}
	src := `@scene(200, 10)
radius = 25.0;
radius.min = 5.0;
radius.max = 50.0;
return sphere(radius);
`
	res, err := Assemble([]string{src}, pkgMgr, ExtractUI())
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	if res.Scene == nil {
		t.Fatalf("expected res.Scene to be populated with ExtractUI()")
	}
	if res.Scene.Size != 200 {
		t.Errorf("expected Scene.Size 200, got %v", res.Scene.Size)
	}
	if len(res.Sliders) != 1 {
		t.Fatalf("expected 1 slider with ExtractUI(), got %d", len(res.Sliders))
	}
	if res.Sliders[0].VariableName != "radius" {
		t.Errorf("expected slider variable 'radius', got %q", res.Sliders[0].VariableName)
	}
}

func TestAssemble_ExtractUI_ZeroHeadlessOverhead(t *testing.T) {
	pkgMgr := &source.MockManager{}
	src := `@scene(200, 10)
radius = 25.0;
radius.min = 5.0;
radius.max = 50.0;
return sphere(radius);
`
	res, err := Assemble([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	if res.Scene != nil {
		t.Errorf("expected res.Scene to be nil when ExtractUI() is omitted, got %+v", res.Scene)
	}
	if len(res.Sliders) != 0 {
		t.Errorf("expected 0 sliders when ExtractUI() is omitted, got %d", len(res.Sliders))
	}
	if len(res.UIControls) != 0 {
		t.Errorf("expected 0 UIControls when ExtractUI() is omitted, got %d", len(res.UIControls))
	}
}

func TestAssemble_MeshSdfSlotsAndGlobalTextures(t *testing.T) {
	pkgMgr := &source.MockManager{}
	src := `
unused() {
	return bvhSdf(99);
}

n = noiseHash3([1.0, 2.0, 3.0]);
return bvhSdf(3) + n;
`
	res, err := Assemble([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	if res.GlobalTextures == nil {
		t.Fatalf("expected res.GlobalTextures to be populated")
	}
	if !res.GlobalTextures.UsesNoiseTile {
		t.Errorf("expected UsesNoiseTile to be true")
	}
	if res.GlobalTextures.UsesWarpField {
		t.Errorf("expected UsesWarpField to be false")
	}
	if res.GlobalTextures.UsesSpinodoid {
		t.Errorf("expected UsesSpinodoid to be false")
	}

	if len(res.MeshSdfSlots) != 1 {
		t.Fatalf("expected 1 mesh SDF slot, got %d", len(res.MeshSdfSlots))
	}
	if res.MeshSdfSlots[0].Slot != 0 || res.MeshSdfSlots[0].AssetIndex != 3 {
		t.Errorf("expected slot 0 for asset 3, got slot %d for asset %d",
			res.MeshSdfSlots[0].Slot, res.MeshSdfSlots[0].AssetIndex)
	}

	// Verify also populated on res.Program
	if res.Program == nil || res.Program.GlobalTextures == nil {
		t.Fatalf("expected res.Program.GlobalTextures to be populated")
	}
	if !res.Program.GlobalTextures.UsesNoiseTile {
		t.Errorf("expected res.Program.GlobalTextures.UsesNoiseTile to be true")
	}
	if len(res.Program.MeshSdfSlots) != 1 {
		t.Fatalf("expected 1 slot on res.Program.MeshSdfSlots, got %d", len(res.Program.MeshSdfSlots))
	}
}

func TestAssemble_ExtractMeshPlan(t *testing.T) {
	pkgMgr := &source.MockManager{}
	srcWithMesh := `
importMesh 0 "p.stl";
meshOut = mesh(0, [1.0, 2.0, 3.0]);
return sphere(1.0), meshOut;
`

	t.Run("DefaultWithoutOptionReturnsNilMeshPlan", func(t *testing.T) {
		res, err := Assemble([]string{srcWithMesh}, pkgMgr)
		if err != nil {
			t.Fatalf("Assemble failed: %v", err)
		}
		if res.MeshPlan != nil {
			t.Errorf("expected res.MeshPlan == nil, got %v", res.MeshPlan)
		}
		if res.SdfSubtrees != nil {
			t.Errorf("expected res.SdfSubtrees == nil, got %v", res.SdfSubtrees)
		}
	})

	t.Run("ExtractMeshPlanPopulatesPlanAndSubtrees", func(t *testing.T) {
		res, err := Assemble([]string{srcWithMesh}, pkgMgr, ExtractMeshPlan("leaf-"))
		if err != nil {
			t.Fatalf("Assemble failed: %v", err)
		}
		var plan *meshplan.MeshExportPlan = res.MeshPlan
		var subtrees []meshplan.SdfSubtree = res.SdfSubtrees
		if plan == nil {
			t.Fatalf("expected res.MeshPlan != nil")
		}
		if plan.Root == nil {
			t.Fatalf("expected res.MeshPlan.Root != nil")
		}
		if len(subtrees) != 1 {
			t.Fatalf("expected 1 SdfSubtree, got %d", len(subtrees))
		}
		leaf := res.SdfSubtrees[0]
		if !strings.HasPrefix(leaf.ID, "leaf-") {
			t.Errorf("expected leaf ID to have prefix 'leaf-', got %s", leaf.ID)
		}
		leafProg := meshplan.SubtreeProgram(res.Program, leaf.Return)
		if leaf.Return == nil || leafProg.Primary == nil {
			t.Fatalf("expected valid subprogram in leaf")
		}
		// Verify subprogram contains 0 mesh(...) calls
		for _, f := range leafProg.Primary.Files {
			for _, r := range f.Returns {
				if op, ok := r.T.(*program.Operand_FnCall); ok && op.FnCall != nil && op.FnCall.Reference != nil {
					if op.FnCall.Reference.Name == "mesh" {
						t.Errorf("subprogram should not contain mesh() call in returns")
					}
				}
			}
		}
	})

	t.Run("CompilationErrorsSkipStep8", func(t *testing.T) {
		badSrc := "syntax error {"
		res, err := Assemble([]string{badSrc}, pkgMgr, ExtractMeshPlan("leaf-"))
		if err == nil {
			t.Fatalf("expected compilation error, got nil")
		}
		if res != nil {
			t.Errorf("expected res == nil on error, got %v", res)
		}
	})

	t.Run("NoGeometryReturnsRootNilAndEmptySubtrees", func(t *testing.T) {
		prog := &program.Program{
			Primary: &program.Package{
				Files: []*program.PackageFile{
					{},
				},
			},
		}
		plan, subtrees, err := meshplan.SliceProgram(prog, "leaf-")
		if err != nil {
			t.Fatalf("SliceProgram failed: %v", err)
		}
		if plan == nil {
			t.Fatalf("expected plan != nil")
		}
		if plan.Root != nil {
			t.Errorf("expected plan.Root == nil for program with no geometry, got %v", plan.Root)
		}
		if len(subtrees) != 0 {
			t.Errorf("expected 0 SdfSubtrees, got %d", len(subtrees))
		}
	})
}
