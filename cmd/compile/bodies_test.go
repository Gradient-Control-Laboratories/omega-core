package main_test

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	main "gradientcontrol.com/omega/core/cmd/compile"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

const twoBodies = "@scene(40)\nball = length(p - [8, 0, 0]) - 5;\nblk = box([-14, -4, -4], [-4, 4, 4]);\nreturn ball, blk;\n"

func zipFiles(t *testing.T, path string) map[string][]byte {
	t.Helper()
	zr, err := zip.OpenReader(path)
	require.NoError(t, err)
	defer zr.Close()
	files := map[string][]byte{}
	for _, f := range zr.File {
		rc, err := f.Open()
		require.NoError(t, err)
		b, err := io.ReadAll(rc)
		require.NoError(t, err)
		rc.Close()
		files[f.Name] = b
	}
	return files
}

// Several returns mesh as several bodies: -t stl writes a zip of <name>.stl,
// named after the source by default.
func TestRun_STL_SeveralBodiesZip(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "pair.gsl")
	require.NoError(t, os.WriteFile(source, []byte(twoBodies), 0644))
	var stdout, stderr bytes.Buffer
	require.NoError(t, main.Run([]string{"-t", "stl", "--subdivision", "5", "--no-meta", source}, nil, &stdout, &stderr),
		stderr.String())

	_, err := os.Stat(filepath.Join(dir, "pair.stl"))
	assert.True(t, os.IsNotExist(err), "no pair.stl")
	files := zipFiles(t, filepath.Join(dir, "pair.zip"))
	require.Len(t, files, 2)
	for _, name := range []string{"ball.stl", "blk.stl"} {
		stl := files[name]
		require.True(t, len(stl) > 84, "%s is an STL", name)
	}
}

// An output named .stl cannot hold several bodies: the error says to name a .zip.
func TestRun_STL_SeveralBodiesRefuseSTLPath(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "pair.gsl")
	require.NoError(t, os.WriteFile(source, []byte(twoBodies), 0644))
	out := filepath.Join(dir, "out.stl")
	var stdout, stderr bytes.Buffer
	err := main.Run([]string{"-t", "stl", "--subdivision", "5", "--no-meta", "-o", out, source}, nil, &stdout, &stderr)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "out.zip"), err.Error())
	_, statErr := os.Stat(out)
	assert.True(t, os.IsNotExist(statErr), "nothing written")
}

// -t 3mf writes each body as its own object.
func TestRun_3MF_SeveralBodies(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "pair.gsl")
	require.NoError(t, os.WriteFile(source, []byte(twoBodies), 0644))
	var stdout, stderr bytes.Buffer
	require.NoError(t, main.Run([]string{"-t", "3mf", "--subdivision", "5", "--no-meta", source}, nil, &stdout, &stderr),
		stderr.String())
	model := string(zipFiles(t, filepath.Join(dir, "pair.3mf"))["3D/3dmodel.model"])
	assert.Equal(t, 2, strings.Count(model, "<object "))
	assert.True(t, strings.Contains(model, `name="ball"`))
	assert.True(t, strings.Contains(model, `name="blk"`))
}
