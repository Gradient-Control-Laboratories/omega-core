package glb

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/assert"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
)

func normalsFixture() []Body {
	return []Body{
		{Name: "tri", Color: "#ff0000", Positions: []float32{0, 0, 0, 10, 0, 0, 0, 5, 2.5}, Indices: []uint32{0, 1, 2}},
		{Name: "quad", Color: "#00ff80", Positions: []float32{-1, -2, -3, 1, -2, -3, 1, 2, -3, -1, 2, 3},
			Indices: []uint32{0, 1, 2, 0, 2, 3}},
	}
}

// A body without normals is written as it was before normals existed: these are the bytes Write gave for this
// fixture then.
func TestWriteWithoutNormalsIsUnchanged(t *testing.T) {
	sum := sha256.Sum256(write(t, normalsFixture()))
	assert.Equal(t, "5da66db678ac8255bd60163a3159ddb62e887c976b3e8777544efec2d158c7f7", hex.EncodeToString(sum[:]))
}

// A body with normals carries them as a NORMAL attribute, one VEC3 per vertex, read back exactly; a body without
// them beside it has none.
func TestWriteNormals(t *testing.T) {
	bodies := normalsFixture()
	bodies[1].Normals = []float32{0, 0, 1, 0, 0, 1, 0, 1, 0, 1, 0, 0}
	p := readGLB(t, write(t, bodies))

	tri := p.doc.Meshes[0].Primitives[0]
	_, has := tri.Attributes["NORMAL"]
	assert.False(t, has, "the body without normals has no NORMAL attribute")

	quad := p.doc.Meshes[1].Primitives[0]
	normal, has := quad.Attributes["NORMAL"]
	require.True(t, has, "the body with normals has a NORMAL attribute")
	accessor := p.doc.Accessors[normal]
	assert.Equal(t, "VEC3", accessor.Type)
	assert.Equal(t, componentFloat, accessor.ComponentType)
	assert.Equal(t, 4, accessor.Count)
	assert.Equal(t, bodies[1].Normals, p.floats(t, normal))
	assert.Equal(t, bodies[1].Positions, p.floats(t, quad.Attributes["POSITION"]))
	assert.Equal(t, bodies[1].Indices, p.uints(t, quad.Indices))
}

// Normals that are not one per vertex are refused.
func TestWriteRejectsMismatchedNormals(t *testing.T) {
	bodies := normalsFixture()
	bodies[0].Normals = []float32{0, 0, 1}
	var buf bytes.Buffer
	assert.Error(t, Write(&buf, bodies))
}
