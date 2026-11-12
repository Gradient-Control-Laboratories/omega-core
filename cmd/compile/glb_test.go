package main_test

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gradientcontrol.com/omega/core/cmd/compile/internal/cli"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// glbNodeNames is a .glb's node names after the root, in order.
func glbNodeNames(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.True(t, len(data) > 20, "a .glb")
	assert.Equal(t, "glTF", string(data[0:4]))
	n := binary.LittleEndian.Uint32(data[12:16])
	var doc struct {
		Nodes []struct {
			Name string `json:"name"`
		} `json:"nodes"`
	}
	require.NoError(t, json.Unmarshal(data[20:20+n], &doc))
	var names []string
	for _, node := range doc.Nodes[1:] {
		names = append(names, node.Name)
	}
	return names
}

// -t glb writes each body as its own node, named after its return, to
// <source>.glb by default.
func TestRun_GLB_SeveralBodies(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "pair.gsl")
	require.NoError(t, os.WriteFile(source, []byte(twoBodies), 0644))
	var stdout, stderr bytes.Buffer
	require.NoError(t, cli.Run([]string{"-t", "glb", "--subdivision", "5", "--no-meta", source}, nil, &stdout, &stderr),
		stderr.String())
	assert.Equal(t, []string{"ball", "blk"}, glbNodeNames(t, filepath.Join(dir, "pair.glb")))
}

// One body's node is named after the source file, as a one-body 3MF's object is.
func TestRun_GLB_OneBodyNamedAfterSource(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "knob.gsl")
	require.NoError(t, os.WriteFile(source, []byte("@scene(40)\nball = length(p) - 5;\nreturn ball;\n"), 0644))
	out := filepath.Join(dir, "out.glb")
	var stdout, stderr bytes.Buffer
	require.NoError(t, cli.Run([]string{"-t", "glb", "--subdivision", "5", "--no-meta", "-o", out, source}, nil, &stdout, &stderr),
		stderr.String())
	assert.Equal(t, []string{"knob"}, glbNodeNames(t, out))
}
