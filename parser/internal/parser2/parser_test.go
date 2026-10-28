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
	api "gradientcontrol.com/omega/core/parser/internal/api"
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

	pkg, err := lib.ParsePackage([]string{src})
	assert.NoError(t, err)
	if assert.Equal(t, 1, len(pkg.Files[0].MeshImports)) {
		imp := pkg.Files[0].MeshImports[0]
		assert.Equal(t, int32(5), imp.Index)
		assert.Equal(t, "gear.stl", imp.Path)
	}
}

func TestParseStepImport(t *testing.T) {
	const src = "import . \"omega\";\nimportStep 5 \"gear.step\";\nreturn mesh(5);\n"

	pkg, err := lib.ParsePackage([]string{src})
	assert.NoError(t, err)
	if assert.Equal(t, 1, len(pkg.Files[0].StepImports)) {
		imp := pkg.Files[0].StepImports[0]
		assert.Equal(t, int32(5), imp.Index)
		assert.Equal(t, "gear.step", imp.Path)
	}
}

func testParseOne(t *testing.T, srcFile string) {
	src_test.RunSourceTest(t, srcFile, runTestParseOne, checkTestParseOne, true)
}

func runTestParseOne(t *testing.T, srcFile string, src []byte) (*api.ParseResult, error) {
	result, err := parser2.Parse([]string{string(src)})
	if err != nil {
		return result, err
	}
	if result == nil {
		return result, fmt.Errorf("no result returned")
	}
	if len(result.Errors) > 0 {
		err = &result.Errors[0]
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

func checkTestParseOne(t *testing.T, srcFile string, src []byte, parseResult *api.ParseResult) error {
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
	_, err := lib.ParsePackage(sources)
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "return statement is only allowed in the last source file")
	}

	// Last file contains return statement, first doesn't: should succeed!
	sources2 := []string{
		"import . \"omega\";\na = 2;\n",
		"import . \"omega\";\nreturn 1;\n",
	}
	pkg2, err := lib.ParsePackage(sources2)
	assert.NoError(t, err)
	assert.NotNil(t, pkg2)
}

func TestParseSingleElementCollection(t *testing.T) {
	// Bracket literal without trailing comma
	pkg1, err := lib.ParsePackage([]string{"import . \"omega\";\na = [1.0];\nreturn a;\n"})
	assert.NoError(t, err)
	assert.NotNil(t, pkg1)

	// Bracket literal with trailing comma
	pkg2, err := lib.ParsePackage([]string{"import . \"omega\";\nb = [2.0,];\nreturn b;\n"})
	assert.NoError(t, err)
	assert.NotNil(t, pkg2)

	// collect() primitive call with 1 argument
	pkg3, err := lib.ParsePackage([]string{"import . \"omega\";\nc = collect(3.0);\nreturn c;\n"})
	assert.NoError(t, err)
	assert.NotNil(t, pkg3)

	// Empty brackets should fail with "collections cannot be empty"
	_, errEmpty := lib.ParsePackage([]string{"import . \"omega\";\nd = [];\nreturn d;\n"})
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
		_, err := lib.ParsePackage(sources)
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
		_, err := lib.ParsePackage(sources)
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
		_, err := lib.ParsePackage(sources)
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
		_, err := lib.ParsePackage(sources)
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
		pkg, err := lib.ParsePackage(sources)
		assert.NoError(t, err)
		assert.NotNil(t, pkg)
	})
}
