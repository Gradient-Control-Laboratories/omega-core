package parser2_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	yaml "gopkg.in/yaml.v3"
	src_test "gradientcontrol.com/omega/core/internal/srctest"
	"gradientcontrol.com/omega/core/internal/test/assert"
	lib "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/parser/internal/parser2"
	"gradientcontrol.com/omega/core/program"
)

type testCase struct {
	Source   string `yaml:"source,omitempty"`
	Expected any    `yaml:"expected"`
}

func TestParse(t *testing.T) {
	src_test.RunAllSourceTests(t, "testdata/*.gsl", testParseOne)
}

func TestParseMeshImport(t *testing.T) {
	const src = "import . \"omega\";\nimportMesh 5 \"gear.stl\";\nreturn mesh(5);\n"

	res := lib.ParsePackage([]string{src})
	assert.NoError(t, res.Error())
	if assert.Equal(t, 1, len(res.Package.Files[0].MeshImports)) {
		imp := res.Package.Files[0].MeshImports[0]
		assert.Equal(t, int32(5), imp.Index)
		assert.Equal(t, "gear.stl", imp.Path)
	}
}

func TestParseStepImport(t *testing.T) {
	const src = "import . \"omega\";\nimportStep 5 \"gear.step\";\nreturn mesh(5);\n"

	res := lib.ParsePackage([]string{src})
	assert.NoError(t, res.Error())
	if assert.Equal(t, 1, len(res.Package.Files[0].StepImports)) {
		imp := res.Package.Files[0].StepImports[0]
		assert.Equal(t, int32(5), imp.Index)
		assert.Equal(t, "gear.step", imp.Path)
	}
}

func testParseOne(t *testing.T, srcFile string) {
	src_test.RunSourceTest(t, srcFile, runTestParseOne, checkTestParseOne, true)
}

func runTestParseOne(t *testing.T, srcFile string, src []byte) (*parser2.Result, error) {
	result := parser2.Parse([]string{string(src)})
	if result == nil {
		return result, fmt.Errorf("no result returned")
	}
	if len(result.Errors) > 0 {
		var err error = &result.Errors[0]
		if _, isBuildErr := err.(*program.Error); !isBuildErr {
			return result, fmt.Errorf("expected build error, got %T: %v", err, err)
		}
		return result, err
	}
	if result.Package == nil {
		return result, fmt.Errorf("no package returned")
	}
	return result, nil
}

func checkTestParseOne(t *testing.T, srcFile string, src []byte, parseResult *parser2.Result) error {
	testName := strings.TrimSuffix(filepath.Base(srcFile), filepath.Ext(srcFile))
	yamlFile := filepath.Join(filepath.Dir(srcFile), testName+".yaml")

	// Read the test case file if it exists
	var tc testCase
	if _, err := os.Stat(yamlFile); err == nil {
		yamlContent, err := os.ReadFile(yamlFile)
		if err != nil {
			return fmt.Errorf("failed to read test file: %s, error: %w", yamlFile, err)
		}
		err = yaml.Unmarshal(yamlContent, &tc)
		if err != nil {
			return fmt.Errorf("failed to unmarshal test case yaml: %s, error: %w", yamlFile, err)
		}
	}

	// Marshal the actual result from Parse to YAML
	actualYaml, err := marshallYaml(parseResult.Package)
	if err != nil {
		return fmt.Errorf("failed to marshal actual result to yaml: %s, error: %w", srcFile, err)
	}

	// If there is no expected yaml, use the actual yaml and write it to the test file
	if tc.Expected == nil {
		tc.Expected, err = unmarshalYaml(actualYaml)
		if err != nil {
			return fmt.Errorf("failed to unmarshal actual result to yaml: %s, error: %w", srcFile, err)
		}
		err = writeYaml(yamlFile, tc)
		if err != nil {
			return fmt.Errorf("failed to write test file: %s, error: %w", yamlFile, err)
		}
	}

	// Marshal the expected part to get a canonical YAML string
	testMarshalYaml(t, tc, actualYaml, parseResult.Package)

	// Traverse the pkg and test attributes
	testPackageAttributes(t, parseResult.Package)

	return nil
}

// --- Verify parse result ---------------------------------------------

// Test un/marshal yaml parse model
func testMarshalYaml(t *testing.T, tc testCase, actualYaml []byte, pkg *program.Package) {
	expectedYaml, err := marshallYaml(tc.Expected)
	assert.NoError(t, err, "Failed to marshal expected yaml")

	if string(expectedYaml) != string(actualYaml) {
		t.Errorf("expectedYaml:\n%s\nactualYaml:\n%s", string(expectedYaml), string(actualYaml))
	}
	assert.Equal(t, string(expectedYaml), string(actualYaml))
}

func writeYaml(yamlFile string, tc testCase) error {
	content, err := marshallYaml(tc)
	if err != nil {
		return err
	}
	return os.WriteFile(yamlFile, content, 0644)
}

func marshallYaml(target any) ([]byte, error) {
	if pm, ok := target.(proto.Message); ok {
		jsonBytes, err := protojson.Marshal(pm)
		if err == nil {
			var roundtrip any
			if err := json.Unmarshal(jsonBytes, &roundtrip); err == nil {
				target = roundtrip
			}
		}
	} else {
		jsonBytes, err := json.Marshal(target)
		if err == nil {
			var roundtrip any
			if err := json.Unmarshal(jsonBytes, &roundtrip); err == nil {
				target = roundtrip
			}
		}
	}

	target = removeLocations(target)

	buffer := &bytes.Buffer{}
	encoder := yaml.NewEncoder(buffer)
	encoder.SetIndent(2)
	err := encoder.Encode(target)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func removeLocations(v any) any {
	switch val := v.(type) {
	case map[string]any:
		newMap := make(map[string]any)
		for k, child := range val {
			if k == "start" || k == "end" {
				continue
			}
			newMap[k] = removeLocations(child)
		}
		return newMap
	case []any:
		newSlice := make([]any, len(val))
		for i, child := range val {
			newSlice[i] = removeLocations(child)
		}
		return newSlice
	default:
		return v
	}
}

func unmarshalYaml(content []byte) (any, error) {
	target := make(map[string]any)
	err := yaml.Unmarshal(content, &target)
	return target, err
}

// --- Verify parsed model attributes --------------------------------

// Traverse the pkg and test attributes
func testPackageAttributes(t *testing.T, pkg *program.Package) {
	v := &visitor{t: t}
	for _, file := range pkg.Files {
		assert.NotNil(t, file)
		for _, imp := range file.Imports {
			assertImport(t, imp)
		}
		for _, imp := range file.MeshImports {
			assertMeshImport(t, imp)
		}
		for _, assign := range file.Assignments {
			assertAssignment(t, assign, v)
		}
		for _, fnType := range file.FnTypes {
			assertFnType(t, fnType, v)
		}
		for _, op := range file.Returns {
			_, _ = program.VisitOperand(op, v)
		}
	}
}

// visitor implements [program.OperandVisitor].
type visitor struct {
	t *testing.T
}

// [program.OperandVisitor]
func (v *visitor) VisitValue(op *program.Operand, value float64) (*program.Operand, error) {
	assertOp(v.t, op)
	return op, nil
}

// [program.OperandVisitor]
func (v *visitor) VisitReference(op *program.Operand, reference *program.Reference) (*program.Operand, error) {
	assertOp(v.t, op)
	return op, nil
}

// [program.OperandVisitor]
func (v *visitor) VisitExpression(op *program.Operand, expression *program.Expression) (*program.Operand, error) {
	assertOp(v.t, op)
	for _, op := range expression.Operands {
		_, _ = program.VisitOperand(op, v)
	}
	return op, nil
}

// [program.OperandVisitor]
func (v *visitor) VisitFnCall(op *program.Operand, functionCall *program.FnCall) (*program.Operand, error) {
	assertOp(v.t, op)
	for _, arg := range functionCall.Arguments {
		_, _ = program.VisitOperand(arg, v)
	}
	return op, nil
}

// [program.OperandVisitor]
func (v *visitor) VisitPropRef(op *program.Operand, propRef *program.PropertyRef) (*program.Operand, error) {
	assertOp(v.t, op)
	return op, nil
}

func assertImport(t *testing.T, imp *program.Import) {
	assert.NotNil(t, imp)
	if !isValidLocation(imp.Start) {
		t.Errorf("Invalid start location: %v for import %s", imp.Start, imp.Path)
	}
	if !isValidLocation(imp.End) {
		t.Errorf("Invalid end location: %v for import %s", imp.End, imp.Path)
	}
}

func assertMeshImport(t *testing.T, imp *program.MeshImport) {
	assert.NotNil(t, imp)
	if !isValidLocation(imp.Start) {
		t.Errorf("Invalid start location: %v for mesh import %s", imp.Start, imp.Path)
	}
	if !isValidLocation(imp.End) {
		t.Errorf("Invalid end location: %v for mesh import %s", imp.End, imp.Path)
	}
}

func assertAssignment(t *testing.T, assign *program.Assignment, v *visitor) {
	assert.NotNil(t, assign)
	if !isValidLocation(assign.Start) {
		t.Errorf("Invalid start location: %v for assignment", assign.Start)
	}
	if !isValidLocation(assign.End) {
		t.Errorf("Invalid end location: %v for assignment", assign.End)
	}
	for _, variable := range assign.Variables {
		assert.NotNil(t, variable)
		for _, property := range variable.Properties {
			assert.NotNil(t, property)
			if !isValidLocation(property.Start) {
				t.Errorf("Invalid start location: %v for property %s", property.Start, property.Name)
			}
			if !isValidLocation(property.End) {
				t.Errorf("Invalid end location: %v for property %s", property.End, property.Name)
			}
			_, _ = program.VisitOperand(property.Expression, v)
		}
	}
	_, _ = program.VisitOperand(assign.Expression, v)
}

func assertFunction(t *testing.T, fn *program.Function, v *visitor) {
	assert.NotNil(t, fn)
	for _, assign := range fn.Assignments {
		assertAssignment(t, assign, v)
	}
	for _, fnType := range fn.FnTypes {
		assertFnType(t, fnType, v)
	}
	for _, op := range fn.Returns {
		_, _ = program.VisitOperand(op, v)
	}
}

func assertFnType(t *testing.T, fnType *program.FnType, v *visitor) {
	assert.NotNil(t, fnType)
	for _, group := range fnType.Groups {
		assert.NotNil(t, group)
		for _, fn := range group.Functions {
			assertFunction(t, fn, v)
		}
	}
}

func assertOp(t *testing.T, op *program.Operand) {
	assert.NotNil(t, op)
	assert.NotNil(t, op.T)
	if !isValidLocation(op.Start) {
		t.Errorf("Invalid start location: %v for %T", op.Start, op.T)
	}
	if !isValidLocation(op.End) {
		t.Errorf("Invalid end location: %v for %T", op.End, op.T)
	}
}

func isValidLocation(loc *program.Location) bool {
	return loc != nil && loc.Line > 0 && loc.Column > 0
}

func TestParseReturnInNonLastFile(t *testing.T) {
	// First file contains a return statement, second doesn't: should fail!
	sources := []string{
		"import . \"omega\";\nreturn 1;\n",
		"import . \"omega\";\na = 2;\n",
	}
	err := lib.ParsePackage(sources).Error()
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "return statement is only allowed in the last source file")
	}

	// Last file contains return statement, first doesn't: should succeed!
	sources2 := []string{
		"import . \"omega\";\na = 2;\n",
		"import . \"omega\";\nreturn 1;\n",
	}
	res2 := lib.ParsePackage(sources2)
	assert.NoError(t, res2.Error())
	if assert.NotNil(t, res2) {
		assert.NotNil(t, res2.Package)
	}
}

func TestParseSingleElementCollection(t *testing.T) {
	// Bracket literal without trailing comma
	res1 := lib.ParsePackage([]string{"import . \"omega\";\na = [1.0];\nreturn a;\n"})
	assert.NoError(t, res1.Error())
	if assert.NotNil(t, res1) {
		assert.NotNil(t, res1.Package)
	}

	// Bracket literal with trailing comma
	res2 := lib.ParsePackage([]string{"import . \"omega\";\nb = [2.0,];\nreturn b;\n"})
	assert.NoError(t, res2.Error())
	if assert.NotNil(t, res2) {
		assert.NotNil(t, res2.Package)
	}

	// collect() primitive call with 1 argument
	res3 := lib.ParsePackage([]string{"import . \"omega\";\nc = collect(3.0);\nreturn c;\n"})
	assert.NoError(t, res3.Error())
	if assert.NotNil(t, res3) {
		assert.NotNil(t, res3.Package)
	}

	// Empty brackets should fail with "collections cannot be empty"
	errEmpty := lib.ParsePackage([]string{"import . \"omega\";\nd = [];\nreturn d;\n"}).Error()
	assert.Error(t, errEmpty)
	if errEmpty != nil {
		assert.Contains(t, errEmpty.Error(), "collections cannot be empty")
	}
}

func TestMultiFileDeclarationCollisions(t *testing.T) {
	t.Run("VariableRedeclarationAcrossFiles", func(t *testing.T) {
		sources := []string{
			"x = 1.0;\n",
			"x = 2.0;\nreturn x;\n",
		}
		err := lib.ParsePackage(sources).Error()
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), `"x" is an existing variable`)
		}
	})

	t.Run("VariableFunctionCollisionAcrossFiles", func(t *testing.T) {
		sources := []string{
			"x = 1.0;\n",
			"x() { return 2.0; }\nreturn x();\n",
		}
		err := lib.ParsePackage(sources).Error()
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), `"x" is an existing variable`)
		}
	})

	t.Run("FunctionVariableCollisionAcrossFiles", func(t *testing.T) {
		sources := []string{
			"foo() { return 1.0; }\n",
			"foo = 2.0;\nreturn foo;\n",
		}
		err := lib.ParsePackage(sources).Error()
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), `"foo" is an existing function`)
		}
	})

	t.Run("DuplicateFunctionArityAcrossFiles", func(t *testing.T) {
		sources := []string{
			"foo(a) { return a; }\n",
			"foo(b) { return b * 2.0; }\nreturn foo(1.0);\n",
		}
		err := lib.ParsePackage(sources).Error()
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), `function "foo" with 1 parameters is already defined`)
		}
	})

	t.Run("DistinctFunctionArityAcrossFilesAllowed", func(t *testing.T) {
		sources := []string{
			"foo(a) { return a; }\n",
			"foo(a, b) { return a + b; }\nreturn foo(1.0) + foo(2.0, 3.0);\n",
		}
		res := lib.ParsePackage(sources)
		assert.NoError(t, res.Error())
		if assert.NotNil(t, res) {
			assert.NotNil(t, res.Package)
		}
	})
}

func TestParseEmptyAndBlankInputsGuaranteePackageFile(t *testing.T) {
	t.Run("NilSources", func(t *testing.T) {
		res := parser2.Parse(nil)
		if assert.NotNil(t, res) && assert.NotNil(t, res.Package) {
			assert.Equal(t, 1, len(res.Errors))
			if assert.Equal(t, 1, len(res.Package.Files)) {
				assert.NotNil(t, res.Package.Files[0])
			}
		}
	})

	t.Run("EmptySourcesSlice", func(t *testing.T) {
		res := parser2.Parse([]string{})
		if assert.NotNil(t, res) && assert.NotNil(t, res.Package) {
			assert.Equal(t, 1, len(res.Errors))
			if assert.Equal(t, 1, len(res.Package.Files)) {
				assert.NotNil(t, res.Package.Files[0])
			}
		}
	})

	t.Run("SingleBlankSource", func(t *testing.T) {
		res := parser2.Parse([]string{""})
		if assert.NotNil(t, res) && assert.NotNil(t, res.Package) {
			if assert.Equal(t, 1, len(res.Package.Files)) {
				assert.NotNil(t, res.Package.Files[0])
			}
		}
	})
}

func TestParsePartialSceneHeaderOnLaterError(t *testing.T) {
	t.Run("ValidSceneDirectiveFollowedBySyntaxError", func(t *testing.T) {
		src := "@scene(20, 5)\nx = ;\n"
		res := parser2.Parse([]string{src})
		assert.Equal(t, 1, len(res.Errors))
		if assert.NotNil(t, res.Package.Scene) {
			assert.Equal(t, 20.0, res.Package.Scene.Size)
			assert.Equal(t, 5.0, res.Package.Scene.Grid)
			assert.True(t, isValidLocation(res.Package.Scene.Start))
			assert.True(t, isValidLocation(res.Package.Scene.End))
		}
		if assert.Equal(t, 1, len(res.Package.Files)) {
			assert.Equal(t, src, res.Package.Files[0].Source)
		}
	})

	t.Run("LegacySceneCommentsFollowedBySyntaxError", func(t *testing.T) {
		src := "// Scene Size: 40\n// Scene Grid: 4\nx = ;\n"
		res := parser2.Parse([]string{src})
		assert.Equal(t, 1, len(res.Errors))
		if assert.NotNil(t, res.Package.Scene) {
			assert.Equal(t, 40.0, res.Package.Scene.Size)
			assert.Equal(t, 10.0, res.Package.Scene.Grid)
			assert.Nil(t, res.Package.Scene.Start)
			assert.Nil(t, res.Package.Scene.End)
		}
		if assert.Equal(t, 1, len(res.Package.Files)) {
			assert.Equal(t, src, res.Package.Files[0].Source)
		}
	})

	t.Run("MalformedSceneDirectivePopulatesSource", func(t *testing.T) {
		src := "@scene(0)\nreturn sphere(1);\n"
		res := parser2.Parse([]string{src})
		assert.Equal(t, 1, len(res.Errors))
		if assert.Equal(t, 1, len(res.Package.Files)) {
			assert.Equal(t, src, res.Package.Files[0].Source)
		}
	})
}

func TestParsePartialImportsAndAssetImportsOnError(t *testing.T) {
	t.Run("ValidImportsFollowedBySyntaxErrorInImport", func(t *testing.T) {
		src := "@scene(25)\nimport \"math\";\nimport . \"omega\";\nimport bad_unquoted;\nreturn 1;\n"
		res := parser2.Parse([]string{src})
		assert.Equal(t, 1, len(res.Errors))
		assert.NotNil(t, res.Package.Scene)
		if assert.Equal(t, 1, len(res.Package.Files)) {
			file := res.Package.Files[0]
			if assert.Equal(t, 2, len(file.Imports)) {
				assert.Equal(t, "math", file.Imports[0].Path)
				assertImport(t, file.Imports[0])
				assert.Equal(t, "vector", file.Imports[1].Path)
				assert.Equal(t, ".", file.Imports[1].Prefix)
				assertImport(t, file.Imports[1])
			}
		}
	})

	t.Run("ValidImportsFollowedBySemanticErrorInImport", func(t *testing.T) {
		src := "import \"math\";\nimport \"invalid.gsl\";\nreturn 1;\n"
		res := parser2.Parse([]string{src})
		assert.Equal(t, 1, len(res.Errors))
		if assert.Equal(t, 1, len(res.Package.Files)) {
			file := res.Package.Files[0]
			if assert.Equal(t, 1, len(file.Imports)) {
				assert.Equal(t, "math", file.Imports[0].Path)
				assertImport(t, file.Imports[0])
			}
		}
	})

	t.Run("ValidAssetImportsFollowedBySyntaxOrBuildError", func(t *testing.T) {
		src := "import \"math\";\nimportStep 0 \"a.step\";\nimportMesh 1 \"b.stl\";\nimportWarp 2 \"c.warp\";\nimportStep 3 ;\n"
		res := parser2.Parse([]string{src})
		assert.Equal(t, 1, len(res.Errors))
		if assert.Equal(t, 1, len(res.Package.Files)) {
			file := res.Package.Files[0]
			assert.Equal(t, 1, len(file.Imports))
			if assert.Equal(t, 1, len(file.StepImports)) {
				assert.Equal(t, "a.step", file.StepImports[0].Path)
			}
			if assert.Equal(t, 1, len(file.MeshImports)) {
				assert.Equal(t, "b.stl", file.MeshImports[0].Path)
			}
			if assert.Equal(t, 1, len(file.WarpImports)) {
				assert.Equal(t, "c.warp", file.WarpImports[0].Path)
			}
		}
	})

	t.Run("ValidAssetImportsFollowedByDuplicateSlotBuildError", func(t *testing.T) {
		src := "importStep 0 \"a.step\";\nimportStep 0 \"dup.step\";\nreturn 1;\n"
		res := parser2.Parse([]string{src})
		assert.Equal(t, 1, len(res.Errors))
		if assert.Equal(t, 1, len(res.Package.Files)) {
			file := res.Package.Files[0]
			if assert.Equal(t, 1, len(file.StepImports)) {
				assert.Equal(t, "a.step", file.StepImports[0].Path)
			}
		}
	})
}

func TestParsePartialDeclarationsAndReturnOnError(t *testing.T) {
	t.Run("SyntaxErrorOnLaterDeclarationPreservesPrecedingDeclarations", func(t *testing.T) {
		src := `@scene(30, 5)
import "lib/part";
importStep 0 "stmt.step";
importId = importStep "expr.stp";
m = importMesh "expr.stl";
w = importWarp "sim.warp";
@range(0, 10) r = 5;
r.color = [1, 0, 0];
helper(a) {
  return a * 2;
}
broken = ;
return importId;
`
		res := parser2.Parse([]string{src})
		assert.Equal(t, 1, len(res.Errors))
		assert.NotNil(t, res.Package.Scene)
		if assert.Equal(t, 1, len(res.Package.Files)) {
			file := res.Package.Files[0]
			assert.Equal(t, src, file.Source)
			assert.Equal(t, 1, len(file.Imports))
			assert.Equal(t, 1, len(file.StepImports))
			// id, m, w, r
			if assert.Equal(t, 4, len(file.Assignments)) {
				assert.NotNil(t, file.Assignments[3].Range)
				rVar := file.Assignments[3].Variables[0]
				assert.Equal(t, "r", rVar.Name)
				assert.Equal(t, 1, len(rVar.Properties))
				assert.Equal(t, "color", rVar.Properties[0].Name)
			}
			if assert.Equal(t, 1, len(file.FnTypes)) {
				assert.Equal(t, "helper", file.FnTypes[0].Name)
			}
			testPackageAttributes(t, res.Package)
		}
	})

	t.Run("SemanticErrorOnLaterDeclarationPreservesPrecedingDeclarations", func(t *testing.T) {
		src := `x = 1.0;
y = 2.0;
x = 3.0;
return x + y;
`
		res := parser2.Parse([]string{src})
		assert.Equal(t, 1, len(res.Errors))
		assert.Contains(t, res.Errors[0].Message, `"x" is an existing variable`)
		if assert.Equal(t, 1, len(res.Package.Files)) {
			file := res.Package.Files[0]
			if assert.Equal(t, 2, len(file.Assignments)) {
				assert.Equal(t, "x", file.Assignments[0].Variables[0].Name)
				assert.Equal(t, "y", file.Assignments[1].Variables[0].Name)
			}
		}
	})

	t.Run("ErrorInReturnPreservesDeclarations", func(t *testing.T) {
		src := `@scene(20)
import "math";
a = 10;
return ;
`
		res := parser2.Parse([]string{src})
		assert.Equal(t, 1, len(res.Errors))
		assert.Contains(t, res.Errors[0].Message, "expected at least one return element")
		assert.NotNil(t, res.Package.Scene)
		if assert.Equal(t, 1, len(res.Package.Files)) {
			file := res.Package.Files[0]
			assert.Equal(t, 1, len(file.Imports))
			assert.Equal(t, 1, len(file.Assignments))
		}
	})

	t.Run("MultiFileErrorInSecondFilePreservesFirstAndPartialSecond", func(t *testing.T) {
		src0 := "@scene(20)\nimport \"math\";\na = 1;\n"
		src1 := "import \"vector\";\nb = 2;\nbroken = ;\nreturn a + b;\n"
		res := parser2.Parse([]string{src0, src1})
		assert.Equal(t, 1, len(res.Errors))
		assert.NotNil(t, res.Package.Scene)
		if assert.Equal(t, 2, len(res.Package.Files)) {
			assert.Equal(t, 1, len(res.Package.Files[0].Imports))
			assert.Equal(t, 1, len(res.Package.Files[0].Assignments))
			assert.Equal(t, 1, len(res.Package.Files[1].Imports))
			assert.Equal(t, 1, len(res.Package.Files[1].Assignments))
			assert.Equal(t, "b", res.Package.Files[1].Assignments[0].Variables[0].Name)
			assert.Equal(t, src1, res.Package.Files[1].Source)
		}
	})
}

func TestParseDiagnosticMessageRegressions(t *testing.T) {
	t.Run("MissingSemicolonBeforeReturn", func(t *testing.T) {
		res := parser2.Parse([]string{"a = 1\nreturn a;\n"})
		if assert.Equal(t, 1, len(res.Errors)) {
			assert.Contains(t, res.Errors[0].Message, `missing ";" before return`)
		}
	})

	t.Run("ForgotSemicolonAtEndOfPreviousStatement", func(t *testing.T) {
		res := parser2.Parse([]string{"a = 1\n}\n"})
		if assert.Equal(t, 1, len(res.Errors)) {
			assert.Contains(t, res.Errors[0].Message, `perhaps forgot ";" at the end of a previous statement?`)
		}
	})

	t.Run("ExpectedAssignmentNamedFunctionOrReturn", func(t *testing.T) {
		res := parser2.Parse([]string{"a = 1;\n123;\n"})
		if assert.Equal(t, 1, len(res.Errors)) {
			assert.Contains(t, res.Errors[0].Message, "expected assignment, named function declaration, or return")
		}
		assert.Equal(t, 1, len(res.Package.Files[0].Assignments))
	})

	t.Run("ImportErrorHint", func(t *testing.T) {
		res := parser2.Parse([]string{"import unquoted;\nreturn 1;\n"})
		if assert.Equal(t, 1, len(res.Errors)) {
			assert.Contains(t, res.Errors[0].Message, `expected 'import <optional-prefix> "<path>";'`)
		}
	})

	t.Run("NestedTernaryDecompositionHint", func(t *testing.T) {
		res := parser2.Parse([]string{"return ?;\n"})
		if assert.Equal(t, 1, len(res.Errors)) {
			assert.Contains(t, res.Errors[0].Message, "nested ternaries cannot be inlined")
		}
	})
}
