package mesh

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/backend/mesh/glb"
	"gradientcontrol.com/omega/core/backend/mesh/threemf"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// glbJSON is a .glb's glTF JSON chunk, decoded loosely.
func glbJSON(t *testing.T, data []byte) map[string]any {
	t.Helper()
	require.True(t, len(data) > 20, "a .glb")
	assert.Equal(t, "glTF", string(data[0:4]))
	n := binary.LittleEndian.Uint32(data[12:16])
	var doc map[string]any
	require.NoError(t, json.Unmarshal(data[20:20+n], &doc))
	return doc
}

// WriteGLB names each body's node as MeshBodies names it, colors it as
// Write3MF does (colors by return index, the palette otherwise), and holds
// its corners welded as for 3MF.
func TestWriteGLBColors(t *testing.T) {
	src := "@scene(40)\nball = length(p - [8, 0, 0]) - 5;\nblk = box([-14, -4, -4], [-4, 4, 4]);\nreturn ball, blk;\n"
	res, err := assembler.Assemble([]string{src})
	require.NoError(t, err)
	bodies, err := MeshBodies(res.Program, bodiesOpts()...)
	require.NoError(t, err)
	require.Len(t, bodies, 2)

	var buf bytes.Buffer
	require.NoError(t, WriteGLB(&buf, bodies, map[int]string{2: "#abcdef"}))
	doc := glbJSON(t, buf.Bytes())

	nodes := doc["nodes"].([]any)
	require.Len(t, nodes, 3)
	materials := doc["materials"].([]any)
	accessors := doc["accessors"].([]any)
	meshes := doc["meshes"].([]any)
	for i, want := range []struct {
		name  string
		color string
	}{{"ball", threemf.PaletteColor(1)}, {"blk", "#abcdef"}} {
		node := nodes[i+1].(map[string]any)
		assert.Equal(t, want.name, node["name"])
		prim := meshes[int(node["mesh"].(float64))].(map[string]any)["primitives"].([]any)[0].(map[string]any)

		mat := materials[int(prim["material"].(float64))].(map[string]any)
		factor := mat["pbrMetallicRoughness"].(map[string]any)["baseColorFactor"].([]any)
		linear := glb.LinearColor(want.color)
		for k := range factor {
			assert.Equal(t, linear[k], factor[k].(float64))
		}

		welded, err := threemf.WeldSTL(bodies[i].STL)
		require.NoError(t, err)
		pos := accessors[int(prim["attributes"].(map[string]any)["POSITION"].(float64))].(map[string]any)
		idx := accessors[int(prim["indices"].(float64))].(map[string]any)
		assert.Equal(t, float64(len(welded.Positions)/3), pos["count"])
		assert.Equal(t, float64(len(welded.Indices)), idx["count"])
	}
}
