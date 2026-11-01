package main_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	main "gradientcontrol.com/omega/core/cmd/fix"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func TestRun_NoArgs_PrintsUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := main.Run(nil, nil, &stdout, &stderr)
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Usage of fix:")
}

func TestRun_HelpAndFixerHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := main.Run([]string{"help"}, nil, &stdout, &stderr)
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Registered fixers:")
	assert.Contains(t, stdout.String(), "scene")
	assert.Contains(t, stdout.String(), "-diff")

	stdout.Reset()
	err = main.Run([]string{"help", "scene"}, nil, &stdout, &stderr)
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "@scene")
	assert.Contains(t, stdout.String(), "Before:")
}

func TestRun_InPlaceFileAndIdempotency(t *testing.T) {
	tmpDir := t.TempDir()
	gslPath := filepath.Join(tmpDir, "model.gsl")
	orig := "// Scene Size: 20\n// Scene Grid: 10\n\nreturn sphere(1);\n"
	require.NoError(t, os.WriteFile(gslPath, []byte(orig), 0644))

	var stdout, stderr bytes.Buffer
	err := main.Run([]string{gslPath}, nil, &stdout, &stderr)
	require.NoError(t, err)

	updated, err := os.ReadFile(gslPath)
	require.NoError(t, err)
	assert.Equal(t, "@scene(20)\n\nreturn sphere(1);\n", string(updated))

	// Set a past modification time and verify a second run does not touch the file
	pastTime := time.Now().Add(-10 * time.Minute).Truncate(time.Second)
	require.NoError(t, os.Chtimes(gslPath, pastTime, pastTime))

	err = main.Run([]string{gslPath}, nil, &stdout, &stderr)
	require.NoError(t, err)

	fi, err := os.Stat(gslPath)
	require.NoError(t, err)
	assert.Equal(t, pastTime, fi.ModTime().Truncate(time.Second))
}

func TestRun_RecursiveDirectoryWalk(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	hiddenDir := filepath.Join(tmpDir, ".hidden")
	require.NoError(t, os.MkdirAll(subDir, 0755))
	require.NoError(t, os.MkdirAll(hiddenDir, 0755))

	file1 := filepath.Join(tmpDir, "a.gsl")
	file2 := filepath.Join(subDir, "b.gsl")
	hiddenFile := filepath.Join(hiddenDir, "c.gsl")
	txtFile := filepath.Join(tmpDir, "notes.txt")

	require.NoError(t, os.WriteFile(file1, []byte("// Scene Size: 40\nreturn sphere(1);\n"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("// Scene Size: 100\n// Scene Grid: 20\nreturn sphere(1);\n"), 0644))
	require.NoError(t, os.WriteFile(hiddenFile, []byte("// Scene Size: 10\nreturn sphere(1);\n"), 0644))
	require.NoError(t, os.WriteFile(txtFile, []byte("// Scene Size: 10\n"), 0644))

	var stdout, stderr bytes.Buffer
	err := main.Run([]string{tmpDir}, nil, &stdout, &stderr)
	require.NoError(t, err)

	b1, err := os.ReadFile(file1)
	require.NoError(t, err)
	assert.Equal(t, "@scene(40)\nreturn sphere(1);\n", string(b1))

	b2, err := os.ReadFile(file2)
	require.NoError(t, err)
	assert.Equal(t, "@scene(100, 5)\nreturn sphere(1);\n", string(b2))

	// Hidden directory file and non-.gsl file must remain untouched
	bh, err := os.ReadFile(hiddenFile)
	require.NoError(t, err)
	assert.Equal(t, "// Scene Size: 10\nreturn sphere(1);\n", string(bh))
}

func TestRun_DiffMode(t *testing.T) {
	tmpDir := t.TempDir()
	gslPath := filepath.Join(tmpDir, "diff_test.gsl")
	orig := "// Scene Size: 20\n// Scene Grid: 10\n\nreturn sphere(1);\n"
	require.NoError(t, os.WriteFile(gslPath, []byte(orig), 0644))

	var stdout, stderr bytes.Buffer
	err := main.Run([]string{"-diff", gslPath}, nil, &stdout, &stderr)
	// Per Go fix -diff semantics, exits with non-zero error when diff is not empty
	require.Error(t, err)
	assert.Contains(t, stdout.String(), "--- ")
	assert.Contains(t, stdout.String(), "+++ ")
	assert.Contains(t, stdout.String(), "- // Scene Size: 20")
	assert.Contains(t, stdout.String(), "+ @scene(20)")

	// File on disk must remain unmodified in -diff mode
	diskBytes, readErr := os.ReadFile(gslPath)
	require.NoError(t, readErr)
	assert.Equal(t, orig, string(diskBytes))

	// When file is already fixed, -diff exits with nil error and empty output
	require.NoError(t, os.WriteFile(gslPath, []byte("@scene(20)\n\nreturn sphere(1);\n"), 0644))
	stdout.Reset()
	err = main.Run([]string{"-diff", gslPath}, nil, &stdout, &stderr)
	require.NoError(t, err)
	assert.Equal(t, "", stdout.String())
}

func TestRun_NonexistentPath(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := main.Run([]string{"/nonexistent/path/to/file.gsl"}, nil, &stdout, &stderr)
	require.Error(t, err)
}
