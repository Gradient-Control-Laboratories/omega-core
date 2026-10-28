package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/deadcode"
	"gradientcontrol.com/omega/core/internal/collect"
	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/internal/link"
	"gradientcontrol.com/omega/core/internal/lower"
	"gradientcontrol.com/omega/core/internal/source"
	src_test "gradientcontrol.com/omega/core/internal/srctest"
	"gradientcontrol.com/omega/core/internal/types"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func TestRebuild(t *testing.T) {
	src_test.RunAllSourceTests(t, "../internal/testdata/*.gsl", testRebuildOne)
}

func testRebuildOne(t *testing.T, srcFile string) {
	src_test.RunSourceTest(t, srcFile, runTestRebuildOne, checkTestRebuildOne, true)
}

func runTestRebuildOne(t *testing.T, srcFile string, src []byte) (*program.Program, error) {
	pkgMgr := source.NewLocalManager("../../../common/source/package")

	prog, err := collect.Program([]string{string(src)}, pkgMgr, parser.ParsePackage)
	if err != nil {
		return nil, err
	}

	pkgName := filepath.Base(srcFile)
	if prog.Primary != nil {
		for _, file := range prog.Primary.Files {
			if file.Path == "" {
				file.Path = pkgName
			}
		}
	}

	err = link.References(prog, pkgMgr, parser.ParsePackage)
	if err != nil {
		return nil, err
	}
	err = types.Analyze(prog)
	if err != nil {
		return nil, err
	}
	err = lower.Transform(prog)
	if err != nil {
		return nil, err
	}
	err = link.References(prog, pkgMgr, parser.ParsePackage)
	if err != nil {
		return nil, err
	}
	err = types.Analyze(prog)
	if err != nil {
		return nil, err
	}
	_, err = eval.Fold(prog, nil)
	if err != nil {
		return nil, err
	}

	newProg, err := deadcode.Remove(prog)
	if err != nil {
		return nil, err
	}

	for _, dep := range newProg.Dependencies {
		depName := filepath.Base(dep.Path)
		for i, file := range dep.Files {
			if file.Path == "" {
				if len(dep.Files) == 1 {
					file.Path = depName + ".gsl"
				} else {
					file.Path = fmt.Sprintf("%s_%d.gsl", depName, i)
				}
			}
		}
	}

	assertAllReferencesLinked(t, newProg)

	return newProg, nil
}

func checkTestRebuildOne(t *testing.T, srcFile string, src []byte, result *program.Program) error {
	dir := filepath.Dir(srcFile)
	testName := strings.TrimSuffix(filepath.Base(srcFile), filepath.Ext(srcFile))
	expectedDir := filepath.Join(dir, testName)

	if err := os.MkdirAll(expectedDir, 0755); err != nil {
		return fmt.Errorf("failed to create expected dir: %w", err)
	}

	checkFile := func(pkgPath, filePath, content string) error {
		var expectedFilePath string
		if pkgPath == "" {
			expectedFilePath = filepath.Join(expectedDir, filePath)
		} else {
			expectedFilePath = filepath.Join(expectedDir, pkgPath, filePath)
		}
		if err := os.MkdirAll(filepath.Dir(expectedFilePath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for expected file: %w", err)
		}

		_, err := os.Stat(expectedFilePath)
		if os.IsNotExist(err) || os.Getenv("UPDATE_GOLDEN") != "" {
			if err := os.WriteFile(expectedFilePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write expected pretty-print file: %w", err)
			}
			t.Logf("Generated expected pretty-print file: %s", expectedFilePath)
			return nil
		}

		expectedContent, err := os.ReadFile(expectedFilePath)
		if err != nil {
			return fmt.Errorf("failed to read expected file %s: %w", expectedFilePath, err)
		}

		actualStr := strings.TrimSpace(content)
		expectedStr := strings.TrimSpace(string(expectedContent))
		if actualStr != expectedStr {
			return fmt.Errorf("pretty-printed content mismatch for %s\nExpected:\n%s\nActual:\n%s", expectedFilePath, expectedStr, actualStr)
		}
		return nil
	}

	if result.Primary != nil {
		for _, file := range result.Primary.Files {
			if err := checkFile("", file.Path, file.PrettyPrint(0)); err != nil {
				return err
			}
		}
	}

	for _, dep := range result.Dependencies {
		for _, file := range dep.Files {
			if err := checkFile(dep.Path, file.Path, file.PrettyPrint(0)); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestReferencesInPlace(t *testing.T) {
	pkgMgr := &source.MockManager{}
	src := `
		y = 10.0;
		x = 42.0;
		used() {
			return 200.0;
		}
		return x + used();
	`

	prog, err := collect.Program([]string{src}, pkgMgr, parser.ParsePackage)
	if err != nil {
		t.Fatalf("failed to collect: %v", err)
	}

	// Run in-place reference resolution
	err = link.References(prog, pkgMgr, parser.ParsePackage)
	if err != nil {
		t.Fatalf("References failed: %v", err)
	}

	// Verify that references are resolved, but dead code (y) is still present!
	hasY := false
	hasX := false
	for _, file := range prog.Primary.Files {
		for _, assign := range file.Assignments {
			for _, v := range assign.Variables {
				if v.Name == "y" {
					hasY = true
				}
				if v.Name == "x" {
					hasX = true
				}
			}
		}
	}

	if !hasY {
		t.Error("expected dead assignment 'y' to still be present in-place after linking")
	}
	if !hasX {
		t.Error("expected assignment 'x' to be present")
	}

	// Verify that references are resolved (e.g. x operand has RefPath)
	foundRef := false
	for _, file := range prog.Primary.Files {
		for _, ret := range file.Returns {
			if expr, ok := ret.T.(*program.Operand_Expression); ok {
				if len(expr.Expression.Operands) > 0 {
					if refOp, ok := expr.Expression.Operands[0].T.(*program.Operand_Reference); ok {
						if refOp.Reference.RefPath == "x" {
							foundRef = true
						}
					}
				}
			}
		}
	}

	if !foundRef {
		t.Error("expected reference 'x' to have RefPath resolved to 'x'")
	}
}

func TestReferences(t *testing.T) {
	src_test.RunAllSourceTests(t, "../internal/testdata/*.gsl", testReferencesOne)
}

func testReferencesOne(t *testing.T, srcFile string) {
	src_test.RunSourceTest(t, srcFile, runTestReferencesOn, checkTestReferencesOne, true)
}

func runTestReferencesOn(t *testing.T, srcFile string, src []byte) (*program.Program, error) {
	pkgMgr := source.NewLocalManager("../../../common/source/package")

	prog, err := collect.Program([]string{string(src)}, pkgMgr, parser.ParsePackage)
	if err != nil {
		return nil, err
	}

	pkgName := filepath.Base(srcFile)
	if prog.Primary != nil {
		for _, file := range prog.Primary.Files {
			if file.Path == "" {
				file.Path = pkgName
			}
		}
	}

	err = link.References(prog, pkgMgr, parser.ParsePackage)
	if err != nil {
		return nil, err
	}

	visitor := &refCheckVisitor{t: t}
	walkPackage(prog.Primary, visitor)
	for _, dep := range prog.Dependencies {
		if dep.Path == "builtin" {
			continue
		}
		walkPackage(dep, visitor)
	}

	return prog, nil
}

func checkTestReferencesOne(t *testing.T, srcFile string, src []byte, result *program.Program) error {
	return nil
}

type refCheckVisitor struct {
	t *testing.T
}

func (v *refCheckVisitor) VisitValue(op *program.Operand, val float64) (any, error) {
	return nil, nil
}

func (v *refCheckVisitor) VisitReference(op *program.Operand, ref *program.Reference) (any, error) {
	if ref == nil {
		return nil, nil
	}
	if program.IsBuiltin(ref.Name) || ref.Name == program.PVarName || ref.Name == program.TimeVarName {
		return nil, nil
	}
	if ref.RefPath == "" {
		v.t.Errorf("Reference %q (prefix: %q) has empty RefPath at %v", ref.Name, ref.Prefix, op.Start)
	}
	return nil, nil
}

func (v *refCheckVisitor) VisitExpression(op *program.Operand, expr *program.Expression) (any, error) {
	if expr == nil {
		return nil, nil
	}
	if expr.Operation == program.Operation_ACCESS {
		return program.VisitOperand(expr.Operands[0], v)
	}
	for _, operand := range expr.Operands {
		if _, err := program.VisitOperand(operand, v); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (v *refCheckVisitor) VisitFnCall(op *program.Operand, fc *program.FnCall) (any, error) {
	if fc == nil {
		return nil, nil
	}
	if fc.Reference != nil {
		ref := fc.Reference
		if !(program.IsBuiltin(ref.Name) || ref.Name == program.PVarName || ref.Name == program.TimeVarName) {
			if ref.RefPath == "" {
				v.t.Errorf("FnCall reference %q (prefix: %q) has empty RefPath at %v", ref.Name, ref.Prefix, op.Start)
			}
		}
	}
	for _, arg := range fc.Arguments {
		if _, err := program.VisitOperand(arg, v); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (v *refCheckVisitor) VisitPropRef(op *program.Operand, pr *program.PropertyRef) (any, error) {
	if pr == nil {
		return nil, nil
	}
	if pr.Reference != nil {
		ref := pr.Reference
		if !(program.IsBuiltin(ref.Name) || ref.Name == program.PVarName || ref.Name == program.TimeVarName) {
			if ref.RefPath == "" {
				v.t.Errorf("PropertyRef reference %q (prefix: %q) has empty RefPath at %v", ref.Name, ref.Prefix, op.Start)
			}
		}
	}
	return nil, nil
}

func walkPackage(pkg *program.Package, visitor *refCheckVisitor) {
	if pkg == nil {
		return
	}
	for _, file := range pkg.Files {
		walkFile(file, visitor)
	}
}

func walkFile(file *program.PackageFile, visitor *refCheckVisitor) {
	if file == nil {
		return
	}
	for _, assign := range file.Assignments {
		walkAssignment(assign, visitor)
	}
	for _, fnType := range file.FnTypes {
		walkFnType(fnType, visitor)
	}
	for _, ret := range file.Returns {
		program.VisitOperand(ret, visitor)
	}
}

func walkAssignment(assign *program.Assignment, visitor *refCheckVisitor) {
	if assign == nil {
		return
	}
	program.VisitOperand(assign.Expression, visitor)
	for _, v := range assign.Variables {
		if v != nil {
			for _, prop := range v.Properties {
				if prop != nil {
					program.VisitOperand(prop.Expression, visitor)
				}
			}
		}
	}
}

func walkFnType(fnType *program.FnType, visitor *refCheckVisitor) {
	if fnType == nil {
		return
	}
	for _, group := range fnType.Groups {
		if group != nil {
			for _, fn := range group.Functions {
				walkFunction(fn, visitor)
			}
		}
	}
}

func walkFunction(fn *program.Function, visitor *refCheckVisitor) {
	if fn == nil {
		return
	}
	for _, assign := range fn.Assignments {
		walkAssignment(assign, visitor)
	}
	for _, fnType := range fn.FnTypes {
		walkFnType(fnType, visitor)
	}
	for _, ret := range fn.Returns {
		program.VisitOperand(ret, visitor)
	}
}

func TestLinkConstantEvaluation(t *testing.T) {
	pkgMgr := &source.MockManager{}
	src := `
		a = 1.0;
		b = 2.0;
		c = a + b;
		mult() {
			return [10.0, 20.0];
		}
		x, y = mult();
		d = p + 5.0;
		return 0.0;
	`

	prog, err := collect.Program([]string{src}, pkgMgr, parser.ParsePackage)
	if err != nil {
		t.Fatalf("failed to collect: %v", err)
	}

	prog.Primary.Path = "test"
	err = link.References(prog, pkgMgr, parser.ParsePackage)
	if err != nil {
		t.Fatalf("References failed: %v", err)
	}
	_, err = eval.Fold(prog, nil)
	if err != nil {
		t.Fatalf("References failed: %v", err)
	}

	// Helper to find variable by name in the primary package
	findVar := func(name string) *program.Variable {
		for _, file := range prog.Primary.Files {
			for _, assign := range file.Assignments {
				for _, v := range assign.Variables {
					if v.Name == name {
						return v
					}
				}
			}
		}
		return nil
	}

	assertConst := func(name string, expected []float64) {
		t.Helper()
		v := findVar(name)
		if v == nil {
			t.Fatalf("variable %q not found", name)
		}
		if len(v.ConstValues) == 0 {
			t.Errorf("variable %q: expected ConstValues %v, got none", name, expected)
			return
		}
		if len(v.ConstValues) != len(expected) {
			t.Errorf("variable %q: expected ConstValues length %d, got %d", name, len(expected), len(v.ConstValues))
			return
		}
		for i, val := range expected {
			if v.ConstValues[i] != val {
				t.Errorf("variable %q index %d: expected %f, got %f", name, i, val, v.ConstValues[i])
			}
		}
	}

	assertNotConst := func(name string) {
		t.Helper()
		v := findVar(name)
		if v == nil {
			t.Fatalf("variable %q not found", name)
		}
		if len(v.ConstValues) > 0 {
			t.Errorf("variable %q: expected no ConstValues, got %v", name, v.ConstValues)
		}
	}

	assertConst("a", []float64{1.0})
	assertConst("b", []float64{2.0})
	assertConst("c", []float64{3.0})
	assertConst("x", []float64{10.0})
	assertConst("y", []float64{20.0})
	assertNotConst("d")
}

func TestLinkConstantEvaluationWithImports(t *testing.T) {
	depSrc := `
		val = 42.0;
		return val;
	`
	pkgMgr := &source.MockManager{
		PkgSources: map[string][]string{
			"dep": {depSrc},
		},
	}
	src := `
		import "dep";
		res = dep.val + 2.0;
		return 0.0;
	`

	prog, err := collect.Program([]string{src}, pkgMgr, parser.ParsePackage)
	if err != nil {
		t.Fatalf("failed to collect: %v", err)
	}

	prog.Primary.Path = "test"
	err = link.References(prog, pkgMgr, parser.ParsePackage)
	if err != nil {
		t.Fatalf("References failed: %v", err)
	}
	_, err = eval.Fold(prog, nil)
	if err != nil {
		t.Fatalf("References failed: %v", err)
	}

	// Helper to find variable by name in the primary package
	findVar := func(name string) *program.Variable {
		for _, file := range prog.Primary.Files {
			for _, assign := range file.Assignments {
				for _, v := range assign.Variables {
					if v.Name == name {
						return v
					}
				}
			}
		}
		return nil
	}

	v := findVar("res")
	if v == nil {
		t.Fatalf("variable 'res' not found")
	}
	expected := []float64{44.0}
	if len(v.ConstValues) == 0 {
		t.Fatalf("variable 'res': expected ConstValues %v, got none", expected)
	}
	if len(v.ConstValues) != len(expected) || v.ConstValues[0] != expected[0] {
		t.Errorf("variable 'res': expected ConstValues %v, got %v", expected, v.ConstValues)
	}
}

func assertAllReferencesLinked(t *testing.T, prog *program.Program) {
	var checkReference func(ref *program.Reference)
	var checkOperand func(op *program.Operand)

	checkReference = func(ref *program.Reference) {
		if ref == nil {
			return
		}
		if program.IsBuiltin(ref.Name) || ref.Name == program.PVarName || ref.Name == program.TimeVarName {
			return
		}
		if ref.RefPath == "" {
			t.Errorf("unresolved reference: %s (prefix: %s)", ref.Name, ref.Prefix)
		}
	}

	checkOperand = func(op *program.Operand) {
		if op == nil {
			return
		}
		switch v := op.T.(type) {
		case *program.Operand_Reference:
			checkReference(v.Reference)
		case *program.Operand_Expression:
			expr := v.Expression
			if expr.Operation == program.Operation_ACCESS {
				if len(expr.Operands) > 0 {
					checkOperand(expr.Operands[0])
				}
			} else {
				for _, child := range expr.Operands {
					checkOperand(child)
				}
			}
		case *program.Operand_FnCall:
			fc := v.FnCall
			checkReference(fc.Reference)
			for _, arg := range fc.Arguments {
				checkOperand(arg)
			}
		}
	}

	var checkFnType func(ft *program.FnType)
	checkFnType = func(ft *program.FnType) {
		if ft == nil {
			return
		}
		for _, group := range ft.Groups {
			for _, fn := range group.Functions {
				for _, assign := range fn.Assignments {
					checkOperand(assign.Expression)
					for _, v := range assign.Variables {
						for _, p := range v.Properties {
							checkOperand(p.Expression)
						}
					}
				}
				for _, ret := range fn.Returns {
					checkOperand(ret)
				}
				for _, subFt := range fn.FnTypes {
					checkFnType(subFt)
				}
			}
		}
	}

	checkPackage := func(pkg *program.Package) {
		if pkg == nil {
			return
		}
		for _, file := range pkg.Files {
			for _, assign := range file.Assignments {
				checkOperand(assign.Expression)
				for _, v := range assign.Variables {
					for _, p := range v.Properties {
						checkOperand(p.Expression)
					}
				}
			}
			for _, ret := range file.Returns {
				checkOperand(ret)
			}
			for _, ft := range file.FnTypes {
				checkFnType(ft)
			}
		}
	}

	checkPackage(prog.Primary)
	for _, dep := range prog.Dependencies {
		checkPackage(dep)
	}
}
