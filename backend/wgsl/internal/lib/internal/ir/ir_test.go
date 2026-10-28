package ir_test

import (
	"bytes"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/internal/ir"
	"gradientcontrol.com/omega/core/internal/test/assert"
)

func TestStmtFormatting(t *testing.T) {
	var buf bytes.Buffer

	letStmt := ir.NewLetStmt("x", "f32", "1.5", nil)
	assert.NoError(t, letStmt.Format(&buf, "  "))
	assert.Equal(t, "  let x: f32 = 1.5;\n", buf.String())

	buf.Reset()
	varStmt := ir.NewVarStmt("y", "", "vec3<f32>(0.0)", nil)
	assert.NoError(t, varStmt.Format(&buf, "  "))
	assert.Equal(t, "  var y = vec3<f32>(0.0);\n", buf.String())

	buf.Reset()
	assignStmt := ir.NewAssignStmt("y", "vec3<f32>(1.0)", nil)
	assert.NoError(t, assignStmt.Format(&buf, "  "))
	assert.Equal(t, "  y = vec3<f32>(1.0);\n", buf.String())

	buf.Reset()
	retStmt := ir.NewReturnStmt("y", []string{"y"})
	assert.NoError(t, retStmt.Format(&buf, "  "))
	assert.Equal(t, "  return y;\n", buf.String())

	buf.Reset()
	opaqueStmt := ir.NewOpaqueStmt([]string{"line1;", "line2;"}, nil, nil)
	assert.NoError(t, opaqueStmt.Format(&buf, "  "))
	assert.Equal(t, "  line1;\n  line2;\n", buf.String())
}

func TestGuardStmtFormatting(t *testing.T) {
	guardBody := ir.NewBlock()
	guardBody.Add(ir.NewAssignStmt("part", "1.0", nil))

	guard := ir.NewGuardStmt("weight > 0.0", guardBody, []string{"weight"})
	var buf bytes.Buffer
	assert.NoError(t, guard.Format(&buf, "  "))

	expected := "  if (weight > 0.0) {\n    part = 1.0;\n  }\n"
	assert.Equal(t, expected, buf.String())
}

func TestBlockReadsAndWrites(t *testing.T) {
	block := ir.NewBlock()
	block.Add(ir.NewLetStmt("a", "f32", "x + y", []string{"x", "y"}))
	block.Add(ir.NewLetStmt("b", "f32", "a * 2.0", []string{"a"}))
	block.Add(ir.NewAssignStmt("out", "b + z", []string{"b", "z"}))

	// External reads: x, y, z (a and b are internal local lets)
	reads := block.Reads()
	assert.Equal(t, []string{"x", "y", "z"}, reads)

	// Writes: a, b, out
	writes := block.Writes()
	assert.Equal(t, []string{"a", "b", "out"}, writes)
}

func TestDeclarationsFormatting(t *testing.T) {
	var buf bytes.Buffer

	st := ir.NewStructDecl("gsl_shape", []ir.StructField{
		{Name: "d", Type: "f32"},
		{Name: "i", Type: "u32"},
	})
	assert.NoError(t, st.Format(&buf))
	expectedStruct := "struct gsl_shape {\n  d : f32,\n  i : u32,\n}"
	assert.Equal(t, expectedStruct, buf.String())

	buf.Reset()
	u := ir.NewUniformDecl(0, 1, "params", "SimParams")
	assert.NoError(t, u.Format(&buf))
	assert.Equal(t, "@group(0) @binding(1) var<uniform> params: SimParams;\n", buf.String())

	buf.Reset()
	fnBody := ir.NewBlock()
	fnBody.Add(ir.NewReturnStmt("length(p)", []string{"p"}))
	fn := ir.NewFuncDecl("sphere", []*ir.Param{{Name: "p", Type: "vec3<f32>"}}, "f32", fnBody)
	assert.NoError(t, fn.Format(&buf))
	expectedFn := "fn sphere(p: vec3<f32>) -> f32 {\n  return length(p);\n}\n"
	assert.Equal(t, expectedFn, buf.String())
}

func TestModuleCanonicalOrder(t *testing.T) {
	mod := ir.NewModule()
	mod.AddUniform(ir.NewUniformDecl(0, 0, "cam", "Camera"))
	mod.AddPredefined("fn gsl_box() -> f32 { return 0.0; }")
	mod.AddImport(ir.NewFuncDecl("lib_math", nil, "f32", ir.NewBlock()))
	mod.AddFunction(ir.NewFuncDecl("sdShape", nil, "f32", ir.NewBlock()))
	mod.AddRuntimeSnippet("var<private> gsl_vanishView: vec4<f32>;")
	mod.AddStruct(ir.NewStructDecl("gsl_shape", []ir.StructField{{Name: "d", Type: "f32"}}))

	rendered := mod.WGSL()

	// Verify order: Uniforms -> Predefined -> Imports -> Functions -> RuntimeSnippets -> Structs
	camIdx := strings.Index(rendered, "var<uniform> cam")
	boxIdx := strings.Index(rendered, "fn gsl_box")
	mathIdx := strings.Index(rendered, "fn lib_math")
	shapeIdx := strings.Index(rendered, "fn sdShape")
	snippetIdx := strings.Index(rendered, "var<private> gsl_vanishView")
	structIdx := strings.Index(rendered, "struct gsl_shape")

	assert.True(t, camIdx != -1)
	assert.True(t, boxIdx > camIdx)
	assert.True(t, mathIdx > boxIdx)
	assert.True(t, shapeIdx > mathIdx)
	assert.True(t, snippetIdx > shapeIdx)
	assert.True(t, structIdx > snippetIdx)
}
