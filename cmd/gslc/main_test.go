package main_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	main "gradientcontrol.com/omega/core/cmd/gslc"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func TestRun_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := main.Run([]string{"--help"}, nil, &stdout, &stderr)
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Usage of gslc")
	assert.Contains(t, stdout.String(), "--target")
}

func TestRun_WGSL_File(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "test_sphere.gsl")
	err := os.WriteFile(sourcePath, []byte("return sqrt(p.x*p.x + p.y*p.y + p.z*p.z) - 1.0;\n"), 0644)
	require.NoError(t, err)

	outPath := filepath.Join(tempDir, "out_sphere.wgsl")
	metaPath := filepath.Join(tempDir, "out_sphere.json")

	var stdout, stderr bytes.Buffer
	err = main.Run([]string{"-o", outPath, sourcePath}, nil, &stdout, &stderr)
	require.NoError(t, err)

	// Verify WGSL file created
	wgslBytes, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.True(t, len(wgslBytes) > 0)
	assert.Contains(t, string(wgslBytes), "sdShape")

	// Verify companion JSON created
	metaBytes, err := os.ReadFile(metaPath)
	require.NoError(t, err)
	assert.True(t, len(metaBytes) > 0)

	var meta map[string]any
	err = json.Unmarshal(metaBytes, &meta)
	require.NoError(t, err)
	assert.Equal(t, "wgsl", meta["target"])
}

func TestRun_WGSL_Stdin(t *testing.T) {
	stdin := strings.NewReader("return sqrt(p.x*p.x + p.y*p.y + p.z*p.z) - 1.0;\n")
	var stdout, stderr bytes.Buffer

	err := main.Run([]string{"-o", "-"}, stdin, &stdout, &stderr)
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "sdShape")
}

func TestRun_WGSL_Standalone(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "standalone.gsl")
	err := os.WriteFile(sourcePath, []byte("return sqrt(p.x*p.x + p.y*p.y + p.z*p.z) - 1.0;\n"), 0644)
	require.NoError(t, err)

	outPath := filepath.Join(tempDir, "standalone.wgsl")
	var stdout, stderr bytes.Buffer
	err = main.Run([]string{"--standalone", "-o", outPath, sourcePath}, nil, &stdout, &stderr)
	require.NoError(t, err)

	wgslBytes, err := os.ReadFile(outPath)
	require.NoError(t, err)
	code := string(wgslBytes)
	assert.Contains(t, code, "sdShape")
	assert.Contains(t, code, "vs_main")
	assert.Contains(t, code, "fs_main")
}

func TestRun_STL(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "model.gsl")
	err := os.WriteFile(sourcePath, []byte("return sqrt(p.x*p.x + p.y*p.y + p.z*p.z) - 1.0;\n"), 0644)
	require.NoError(t, err)

	outPath := filepath.Join(tempDir, "model.stl")
	var stdout, stderr bytes.Buffer
	err = main.Run([]string{"-t", "stl", "--subdivision", "3", "-o", outPath, sourcePath}, nil, &stdout, &stderr)
	require.NoError(t, err)

	stlBytes, err := os.ReadFile(outPath)
	require.NoError(t, err)
	// Binary STL must be at least 84 bytes (80 header + 4 triangle count)
	assert.True(t, len(stlBytes) >= 84)
}

func TestRun_Tape(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "tape.gsl")
	err := os.WriteFile(sourcePath, []byte("return sqrt(p.x*p.x + p.y*p.y + p.z*p.z) - 1.0;\n"), 0644)
	require.NoError(t, err)

	outPath := filepath.Join(tempDir, "tape.tape")
	var stdout, stderr bytes.Buffer
	err = main.Run([]string{"-t", "tape", "-o", outPath, sourcePath}, nil, &stdout, &stderr)
	require.NoError(t, err)

	tapeBytes, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.True(t, len(tapeBytes) > 0)
}

func TestRun_AST(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "ast.gsl")
	err := os.WriteFile(sourcePath, []byte("return 1.0;\n"), 0644)
	require.NoError(t, err)

	outPath := filepath.Join(tempDir, "ast.json")
	var stdout, stderr bytes.Buffer
	err = main.Run([]string{"-t", "ast", "-o", outPath, sourcePath}, nil, &stdout, &stderr)
	require.NoError(t, err)

	astBytes, err := os.ReadFile(outPath)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(astBytes, &parsed)
	require.NoError(t, err)
}

func TestRun_NoMeta(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "nometa.gsl")
	err := os.WriteFile(sourcePath, []byte("return 1.0;\n"), 0644)
	require.NoError(t, err)

	outPath := filepath.Join(tempDir, "nometa.wgsl")
	metaPath := filepath.Join(tempDir, "nometa.json")

	var stdout, stderr bytes.Buffer
	err = main.Run([]string{"--no-meta", "-o", outPath, sourcePath}, nil, &stdout, &stderr)
	require.NoError(t, err)

	_, err = os.Stat(outPath)
	require.NoError(t, err)

	_, err = os.Stat(metaPath)
	assert.True(t, os.IsNotExist(err))
}

func TestRun_InvalidFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := main.Run([]string{"nonexistent_xyz.gsl"}, nil, &stdout, &stderr)
	require.Error(t, err)
}
