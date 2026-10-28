package test_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/cmd/gslc/internal/cli"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func TestCompiler_StdlibAndImports(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "stdlib_test.gsl")
	// Use embedded stdlib packages color and vector
	src := `
r = 1.0;
r.min = 0.5;
r.max = 2.5;

len = sqrt(p.x * p.x + p.y * p.y + p.z * p.z);
return len - r;
`
	err := os.WriteFile(sourcePath, []byte(src), 0644)
	require.NoError(t, err)

	outPath := filepath.Join(tempDir, "stdlib_test.wgsl")
	metaPath := filepath.Join(tempDir, "stdlib_test.json")

	var stdout, stderr bytes.Buffer
	err = cli.Run([]string{"-t", "wgsl", "-o", outPath, sourcePath}, nil, &stdout, &stderr)
	require.NoError(t, err)

	// Verify WGSL shader generated
	wgslData, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Contains(t, string(wgslData), "sdShape")

	// Verify companion metadata contains slider 'r'
	metaData, err := os.ReadFile(metaPath)
	require.NoError(t, err)
	var meta struct {
		Target  string `json:"target"`
		Sliders []struct {
			VariableName string  `json:"variableName"`
			Min          float64 `json:"min"`
			Max          float64 `json:"max"`
		} `json:"sliders"`
	}
	err = json.Unmarshal(metaData, &meta)
	require.NoError(t, err)
	assert.Equal(t, "wgsl", meta.Target)
	require.Equal(t, 1, len(meta.Sliders))
	assert.Equal(t, "r", meta.Sliders[0].VariableName)
	assert.Equal(t, 0.5, meta.Sliders[0].Min)
	assert.Equal(t, 2.5, meta.Sliders[0].Max)
}

func TestCompiler_InliningAndDeadCode(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "opt_test.gsl")
	// usedFunc should be inlined; unusedFunc should be eliminated by deadcode removal
	src := `
usedFunc(v) {
  return v * 0.5;
}

unusedFunc(v) {
  return v * 999.0;
}

return usedFunc(p.x) + p.y - 1.0;
`
	err := os.WriteFile(sourcePath, []byte(src), 0644)
	require.NoError(t, err)

	outPath := filepath.Join(tempDir, "opt_test.wgsl")
	var stdout, stderr bytes.Buffer
	err = cli.Run([]string{"-O", "1", "-o", outPath, sourcePath}, nil, &stdout, &stderr)
	require.NoError(t, err)

	wgslData, err := os.ReadFile(outPath)
	require.NoError(t, err)
	code := string(wgslData)
	assert.Contains(t, code, "sdShape")
	// The 999.0 literal from dead code should not appear
	assert.False(t, strings.Contains(code, "999"))
}

func TestCompiler_AllBackends(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "sphere.gsl")
	src := `return sqrt(p.x * p.x + p.y * p.y + p.z * p.z) - 1.2;`
	err := os.WriteFile(sourcePath, []byte(src), 0644)
	require.NoError(t, err)

	// 1. WGSL
	{
		outPath := filepath.Join(tempDir, "out.wgsl")
		var stdout, stderr bytes.Buffer
		err := cli.Run([]string{"-t", "wgsl", "-o", outPath, sourcePath}, nil, &stdout, &stderr)
		require.NoError(t, err)
		data, err := os.ReadFile(outPath)
		require.NoError(t, err)
		assert.Contains(t, string(data), "sdShape")
	}

	// 2. STL Mesh
	{
		outPath := filepath.Join(tempDir, "out.stl")
		var stdout, stderr bytes.Buffer
		err := cli.Run([]string{"-t", "stl", "--subdivision", "3", "-o", outPath, sourcePath}, nil, &stdout, &stderr)
		require.NoError(t, err)
		data, err := os.ReadFile(outPath)
		require.NoError(t, err)
		assert.True(t, len(data) >= 84)
	}

	// 3. Bytecode Tape
	{
		outPath := filepath.Join(tempDir, "out.tape")
		var stdout, stderr bytes.Buffer
		err := cli.Run([]string{"-t", "tape", "-o", outPath, sourcePath}, nil, &stdout, &stderr)
		require.NoError(t, err)
		data, err := os.ReadFile(outPath)
		require.NoError(t, err)
		assert.True(t, len(data) > 0)
	}

	// 4. AST JSON
	{
		outPath := filepath.Join(tempDir, "out.ast.json")
		var stdout, stderr bytes.Buffer
		err := cli.Run([]string{"-t", "ast", "-o", outPath, sourcePath}, nil, &stdout, &stderr)
		require.NoError(t, err)
		data, err := os.ReadFile(outPath)
		require.NoError(t, err)
		var parsed map[string]any
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)
	}
}
