package main_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/cmd/compile/internal/cli"
	"gradientcontrol.com/omega/core/internal/compile"
	"gradientcontrol.com/omega/core/internal/meshimport"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

const fieldSlab = "@scene(40)\n@range(0.5, 8) wall = 2.0;\n\nslab = boxCentered([0, 0, 0], [20, 20, wall]);\nreturn slab;\n"

// fieldOf reads a field bundle as a generic JSON object.
func fieldOf(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var bundle map[string]any
	require.NoError(t, json.Unmarshal(data, &bundle), string(data))
	return bundle
}

// -t field writes <source>.field.json: the WGSL and what a viewer draws it with, byte for byte the bundle
// compile.Process (gsl serve's /api/compile) makes for the same document, and no companion metadata.
func TestRun_Field_IsTheViewersBundle(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "slab.gsl")
	require.NoError(t, os.WriteFile(source, []byte(fieldSlab), 0644))
	var stdout, stderr bytes.Buffer
	require.NoError(t, cli.Run([]string{"-t", "field", source}, nil, &stdout, &stderr), stderr.String())

	data, err := os.ReadFile(filepath.Join(dir, "slab.field.json"))
	require.NoError(t, err)
	_, metaErr := os.Stat(filepath.Join(dir, "slab.json"))
	assert.True(t, os.IsNotExist(metaErr), "no companion metadata beside the bundle")

	want, err := json.Marshal(compile.Process(compile.Request{
		ID: source, Source: fieldSlab, Meshes: meshimport.Dirs(dir),
	}, []string{dir}, nil))
	require.NoError(t, err)
	assert.Equal(t, string(want)+"\n", string(data))

	bundle := fieldOf(t, data)
	assert.True(t, strings.Contains(bundle["wgsl"].(string), "fn sdShape"), "the bundle carries the shader")
	assert.Equal(t, []any{"slab"}, bundle["shapeNames"])
	sliders := bundle["sliders"].([]any)
	require.Equal(t, 1, len(sliders))
	slider := sliders[0].(map[string]any)
	assert.Equal(t, "wall", slider["variableName"])
	assert.Equal(t, 0.5, slider["min"])
	assert.Equal(t, 8.0, slider["max"])
	assert.Equal(t, 2.0, slider["initialValue"])
	assert.True(t, strings.Contains(bundle["wgsl"].(string), slider["paramName"].(string)),
		"the slider is an input of the shader, by the name the bundle gives it")
	assert.Equal(t, 40.0, bundle["scene"].(map[string]any)["size"])
	_, hasErrors := bundle["errors"]
	assert.False(t, hasErrors, "a document that compiles has no errors")
}

// A document with errors still gets its bundle, holding each error with its line and column, and the command
// fails, so the exit status says the bundle draws nothing.
func TestRun_Field_ErrorsGoInTheBundle(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "bad.gsl")
	require.NoError(t, os.WriteFile(source, []byte("@scene(40)\nx = boxCentered([0, 0, 0], [10, 10, 10]);\nreturn y;\n"), 0644))
	out := filepath.Join(dir, "bad.field.json")
	var stdout, stderr bytes.Buffer
	err := cli.Run([]string{"-t", "field", "-o", out, source}, nil, &stdout, &stderr)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "undefined identifier: y")

	data, readErr := os.ReadFile(out)
	require.NoError(t, readErr)
	bundle := fieldOf(t, data)
	_, hasWGSL := bundle["wgsl"]
	assert.False(t, hasWGSL, "a document with errors has no shader")
	errs := bundle["errors"].([]any)
	require.Equal(t, 1, len(errs))
	first := errs[0].(map[string]any)
	assert.Equal(t, 3.0, first["lineNum"])
	assert.Equal(t, 8.0, first["column"])
	assert.Contains(t, first["message"].(string), "undefined identifier: y")
}

// The bundle of a document that reads an imported mesh carries the mesh's preview grid, found beside the document.
func TestRun_Field_ImportedMeshGrid(t *testing.T) {
	data, err := compileFile(t, importedBox(t, "@scene(30)\nimportMesh 0 \"box.stl\";\nreturn bvhSdf(0);\n"), "-t", "field")
	require.NoError(t, err)
	bindings, ok := fieldOf(t, data)["meshSdfBindings"].([]any)
	require.True(t, ok, "the bundle has meshSdfBindings")
	require.Equal(t, 1, len(bindings))
	assert.True(t, len(bindings[0].(map[string]any)["bytesB64"].(string)) > 0, "the grid's bytes are in the bundle")
}

// -t field reads one document and writes no companion metadata, and says so.
func TestRun_Field_Refusals(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.gsl")
	b := filepath.Join(dir, "b.gsl")
	require.NoError(t, os.WriteFile(a, []byte(fieldSlab), 0644))
	require.NoError(t, os.WriteFile(b, []byte(fieldSlab), 0644))
	var stdout, stderr bytes.Buffer

	err := cli.Run([]string{"-t", "field", "-o", filepath.Join(dir, "out.json"), a, b}, nil, &stdout, &stderr)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "one source file")

	err = cli.Run([]string{"-t", "field", "-m", filepath.Join(dir, "meta.json"), a}, nil, &stdout, &stderr)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no companion metadata")
}
