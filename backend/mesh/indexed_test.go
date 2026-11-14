package mesh

import (
	"bytes"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// tetra is the unit corner tetrahedron, its four faces wound outward.
func tetra() ([]float32, []uint32) {
	return []float32{0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1}, []uint32{0, 2, 1, 0, 1, 3, 0, 3, 2, 1, 2, 3}
}

// A closed mesh has no open and no non-manifold edges, and its signed volume is the solid's.
func TestMeasureIndexedClosed(t *testing.T) {
	positions, indices := tetra()
	s := MeasureIndexed(positions, indices)
	assert.Equal(t, 4, s.Vertices)
	assert.Equal(t, 4, s.Triangles)
	assert.Equal(t, 0, s.OpenEdges)
	assert.Equal(t, 0, s.NonManifoldEdges)
	assert.True(t, math.Abs(s.SignedVolume-1.0/6) < 1e-12, "volume %g, want 1/6", s.SignedVolume)
}

// A missing face leaves its three edges open; turning every face over turns the volume's sign.
func TestMeasureIndexedOpenAndInsideOut(t *testing.T) {
	positions, indices := tetra()
	s := MeasureIndexed(positions, indices[:9])
	assert.Equal(t, 3, s.OpenEdges)
	assert.Equal(t, 0, s.NonManifoldEdges)

	flipped := make([]uint32, len(indices))
	for i := 0; i < len(indices); i += 3 {
		flipped[i], flipped[i+1], flipped[i+2] = indices[i], indices[i+2], indices[i+1]
	}
	s = MeasureIndexed(positions, flipped)
	assert.Equal(t, 0, s.OpenEdges)
	assert.True(t, math.Abs(s.SignedVolume+1.0/6) < 1e-12, "volume %g, want -1/6", s.SignedVolume)
}

// An edge three triangles share is non-manifold; the fin's two other edges are open.
func TestMeasureIndexedNonManifold(t *testing.T) {
	positions, indices := tetra()
	positions = append(positions, 1, 1, 1)
	indices = append(indices, 0, 1, 4)
	s := MeasureIndexed(positions, indices)
	assert.Equal(t, 1, s.NonManifoldEdges)
	assert.Equal(t, 2, s.OpenEdges)
}

// GLBNormals is asked for each body's normals over its welded vertices, and what it returns is the body's NORMAL
// attribute; without the option a body has none.
func TestWriteGLBNormalsOption(t *testing.T) {
	src := "@scene(40)\nball = length(p - [8, 0, 0]) - 5;\nblk = box([-14, -4, -4], [-4, 4, 4]);\nreturn ball, blk;\n"
	res, err := assembler.Assemble([]string{src})
	require.NoError(t, err)
	bodies, err := MeshBodies(res.Program, bodiesOpts()...)
	require.NoError(t, err)

	var plain bytes.Buffer
	require.NoError(t, WriteGLB(&plain, bodies, nil))
	for _, m := range glbJSON(t, plain.Bytes())["meshes"].([]any) {
		attributes := m.(map[string]any)["primitives"].([]any)[0].(map[string]any)["attributes"].(map[string]any)
		_, has := attributes["NORMAL"]
		assert.False(t, has, "no NORMAL without the option")
	}

	var asked []string
	var buf bytes.Buffer
	require.NoError(t, WriteGLB(&buf, bodies, nil, GLBNormals(func(body Body, positions []float32, indices []uint32) ([]float32, error) {
		asked = append(asked, body.Name)
		assert.True(t, MeasureIndexed(positions, indices).OpenEdges == 0, "%s arrives welded and closed", body.Name)
		normals := make([]float32, len(positions))
		for i := 2; i < len(normals); i += 3 {
			normals[i] = 1
		}
		return normals, nil
	})))
	assert.Equal(t, []string{"ball", "blk"}, asked)
	doc := glbJSON(t, buf.Bytes())
	accessors := doc["accessors"].([]any)
	for _, m := range doc["meshes"].([]any) {
		attributes := m.(map[string]any)["primitives"].([]any)[0].(map[string]any)["attributes"].(map[string]any)
		normal, has := attributes["NORMAL"]
		require.True(t, has, "a NORMAL with the option")
		position := accessors[int(attributes["POSITION"].(float64))].(map[string]any)
		assert.Equal(t, position["count"], accessors[int(normal.(float64))].(map[string]any)["count"])
	}
}
