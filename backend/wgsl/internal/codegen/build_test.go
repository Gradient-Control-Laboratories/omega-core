package codegen_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/backend/wgsl"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsltest"
	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/source"
	src_test "gradientcontrol.com/omega/core/internal/srctest"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/program"
)

var mockMgr = source.NewLocalManager("testpkg", "../../../../stdlib")

func renderHelper(
	src string,
	prog *program.Program,
) (*wgsl.Result, error) {
	return wgsl.Render(src, prog)
}

func TestBuild(t *testing.T) {
	src_test.RunAllSourceTests(t, "testdata/*.gsl", testBuildOne)
}

func TestAutoImportInSubpackage(t *testing.T) {
	sources, _, err := mockMgr.Sources("branch_pkg")
	if err != nil {
		t.Fatalf("failed to get branch_pkg sources: %v", err)
	}
	res, err := assembler.Assemble(sources, assembler.SourceManager(mockMgr))
	if err != nil {
		t.Fatalf("failed to assemble program: %v", err)
	}
	_, err = renderHelper(sources[0], res.Program)
	if err != nil {
		t.Fatalf("WGSL render failed: %v", err)
	}
}

func TestMatrixOperatorsInFunction(t *testing.T) {
	src := []byte(`m = [[1, 0], [0, 1]];
blend(u) { return m * u + m * (1 - u); }
bm = blend(p.x);
col = bm[0];
return col[0];`)
	if _, err := runTestBuildOne(t, "", src); err != nil {
		t.Fatal(err)
	}
}

func testBuildOne(t *testing.T, srcFile string) {
	if !strings.HasSuffix(srcFile, "literal_to_param.gsl") {
		src_test.RunSourceTest(t, srcFile, runTestBuildOne, checkTestBuildOne, true)
	}
}

func runTestBuildOne(t *testing.T, srcFile string, src []byte) (*wgsl.Result, error) {
	res, err := assembler.Assemble([]string{string(src)}, assembler.SourceManager(mockMgr))
	if err != nil {
		return nil, err
	}
	return renderHelper(string(src), res.Program)
}

func checkTestBuildOne(t *testing.T, srcFile string, src []byte, result *wgsl.Result) error {
	if result == nil {
		return fmt.Errorf("no result returned")
	}
	baseName := filepath.Base(srcFile)
	testName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	wgslFilename := testName + ".wgsl"
	goldenPath := filepath.Join(filepath.Dir(srcFile), wgslFilename)
	wgsltest.AssertWGSLEqual(t, goldenPath, result.WGSL, true)
	return nil
}

func TestWGSL_UnprunedProgram(t *testing.T) {
	src := `
		unusedVar = 999.0;
		unusedFn() { return 123.0; }
		r = 10.0;
		return sphere(p, r);
	`
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mockMgr))
	assert.NoError(t, err)

	rendered, err := renderHelper(src, res.Program)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(rendered.WGSL, "unusedVar"))
	assert.False(t, strings.Contains(rendered.WGSL, "unusedFn"))
	assert.True(t, strings.Contains(rendered.WGSL, "sdShape"))
}

func TestWGSL_PrecomputedSlotsAndTexturesIngestion(t *testing.T) {
	src := `
n = noiseHash3([1.0, 2.0, 3.0]);
return bvhSdf(3) + n;
`
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mockMgr))
	assert.NoError(t, err)

	// Verify assembler populated MeshSdfSlots and GlobalTextures
	if assert.NotNil(t, res.Program.GlobalTextures) {
		assert.True(t, res.Program.GlobalTextures.UsesNoiseTile)
	}
	assert.Len(t, res.Program.MeshSdfSlots, 1)

	rendered, err := renderHelper(src, res.Program)
	assert.NoError(t, err)
	assert.NotNil(t, rendered)
	assert.Len(t, rendered.MeshSdfSlots, 1)
	assert.Equal(t, 0, rendered.MeshSdfSlots[0].Slot)
	assert.Equal(t, 3, rendered.MeshSdfSlots[0].AssetIndex)
	assert.True(t, strings.Contains(rendered.WGSL, "noiseTex"))
	assert.True(t, strings.Contains(rendered.WGSL, "sdfTex0"))
}

func TestWGSL_DirectSlotsAndTexturesEmpty(t *testing.T) {
	src := `return sphere(10.0);`
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mockMgr))
	assert.NoError(t, err)

	assert.Empty(t, res.Program.MeshSdfSlots)
	assert.NotNil(t, res.Program.GlobalTextures)
	assert.False(t, res.Program.GlobalTextures.UsesNoiseTile)
	assert.False(t, res.Program.GlobalTextures.UsesSpinodoid)
	assert.False(t, res.Program.GlobalTextures.UsesWarpField)

	rendered, err := renderHelper(src, res.Program)
	assert.NoError(t, err)
	assert.NotNil(t, rendered)
	assert.Empty(t, rendered.MeshSdfSlots)
	assert.False(t, strings.Contains(rendered.WGSL, "noiseTex"))
	assert.False(t, strings.Contains(rendered.WGSL, "sdfTex"))
	assert.False(t, strings.Contains(rendered.WGSL, "spinodoidTex"))
	assert.False(t, strings.Contains(rendered.WGSL, "warpTex"))
}
