package glb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// parsed is a .glb taken apart: its glTF JSON and its binary chunk.
type parsed struct {
	doc document
	bin []byte
}

// readGLB checks a .glb's framing (header, chunk types, lengths, 4-byte
// alignment) and returns its JSON and binary chunks.
func readGLB(t *testing.T, data []byte) parsed {
	t.Helper()
	require.True(t, len(data) >= headerBytes+chunkHeader, "a header and a JSON chunk")
	le := binary.LittleEndian
	assert.Equal(t, uint32(magic), le.Uint32(data[0:4]))
	assert.Equal(t, uint32(version), le.Uint32(data[4:8]))
	assert.Equal(t, uint32(len(data)), le.Uint32(data[8:12]))
	assert.Equal(t, 0, len(data)%4, "the file is 4-byte aligned")

	jsonLen := int(le.Uint32(data[12:16]))
	assert.Equal(t, uint32(chunkJSON), le.Uint32(data[16:20]))
	assert.Equal(t, 0, jsonLen%4, "the JSON chunk is padded to 4 bytes")
	jsonEnd := headerBytes + chunkHeader + jsonLen
	require.True(t, jsonEnd <= len(data), "the JSON chunk fits")
	var p parsed
	require.NoError(t, json.Unmarshal(data[headerBytes+chunkHeader:jsonEnd], &p.doc))

	if jsonEnd == len(data) {
		return p
	}
	binLen := int(le.Uint32(data[jsonEnd : jsonEnd+4]))
	assert.Equal(t, uint32(chunkBIN), le.Uint32(data[jsonEnd+4:jsonEnd+8]))
	assert.Equal(t, 0, binLen%4, "the BIN chunk is padded to 4 bytes")
	assert.Equal(t, len(data), jsonEnd+chunkHeader+binLen, "the BIN chunk ends the file")
	p.bin = data[jsonEnd+chunkHeader:]
	return p
}

func (p parsed) floats(t *testing.T, accessorIndex int) []float32 {
	t.Helper()
	a := p.doc.Accessors[accessorIndex]
	v := p.doc.BufferViews[a.BufferView]
	out := make([]float32, v.ByteLength/4)
	for i := range out {
		out[i] = math.Float32frombits(binary.LittleEndian.Uint32(p.bin[v.ByteOffset+4*i:]))
	}
	return out
}

func (p parsed) uints(t *testing.T, accessorIndex int) []uint32 {
	t.Helper()
	a := p.doc.Accessors[accessorIndex]
	v := p.doc.BufferViews[a.BufferView]
	out := make([]uint32, v.ByteLength/4)
	for i := range out {
		out[i] = binary.LittleEndian.Uint32(p.bin[v.ByteOffset+4*i:])
	}
	return out
}

func write(t *testing.T, bodies []Body) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, Write(&buf, bodies))
	return buf.Bytes()
}

// Each body is a child of the root node, with a mesh whose positions and
// indices read back exactly, a POSITION min/max, and a material in its color.
func TestWriteLayout(t *testing.T) {
	tri := Body{
		Name:      "tri",
		Color:     "#ff0000",
		Positions: []float32{0, 0, 0, 10, 0, 0, 0, 5, 2.5},
		Indices:   []uint32{0, 1, 2},
	}
	quad := Body{
		Name:      `a "quoted" name`,
		Color:     "#00ff80",
		Positions: []float32{-1, -2, -3, 1, -2, -3, 1, 2, -3, -1, 2, 3},
		Indices:   []uint32{0, 1, 2, 0, 2, 3},
	}
	p := readGLB(t, write(t, []Body{tri, quad}))
	d := p.doc

	assert.Equal(t, "2.0", d.Asset.Version)
	assert.Equal(t, 0, d.Scene)
	require.Len(t, d.Scenes, 1)
	assert.Equal(t, []int{0}, d.Scenes[0].Nodes)

	require.Len(t, d.Nodes, 3)
	root := d.Nodes[0]
	assert.Equal(t, []int{1, 2}, root.Children)
	assert.Nil(t, root.Mesh)
	require.NotNil(t, root.Rotation)
	require.NotNil(t, root.Scale)
	assert.Equal(t, RootRotation, *root.Rotation)
	assert.Equal(t, [3]float64{RootScale, RootScale, RootScale}, *root.Scale)

	require.Len(t, d.Meshes, 2)
	require.Len(t, d.Materials, 2)
	require.Len(t, d.Buffers, 1)
	assert.Equal(t, len(p.bin), d.Buffers[0].ByteLength, "no padding was needed")
	for i, b := range []Body{tri, quad} {
		n := d.Nodes[i+1]
		assert.Equal(t, b.Name, n.Name)
		require.NotNil(t, n.Mesh)
		m := d.Meshes[*n.Mesh]
		require.Len(t, m.Primitives, 1)
		prim := m.Primitives[0]
		assert.Equal(t, modeTriangles, prim.Mode)

		pos := d.Accessors[prim.Attributes["POSITION"]]
		assert.Equal(t, componentFloat, pos.ComponentType)
		assert.Equal(t, "VEC3", pos.Type)
		assert.Equal(t, len(b.Positions)/3, pos.Count)
		assert.Equal(t, b.Positions, p.floats(t, prim.Attributes["POSITION"]))
		lo, hi := bounds(b.Positions)
		assert.Equal(t, lo, pos.Min)
		assert.Equal(t, hi, pos.Max)
		assert.Equal(t, targetArray, d.BufferViews[pos.BufferView].Target)

		idx := d.Accessors[prim.Indices]
		assert.Equal(t, componentUint, idx.ComponentType)
		assert.Equal(t, "SCALAR", idx.Type)
		assert.Equal(t, len(b.Indices), idx.Count)
		assert.Equal(t, b.Indices, p.uints(t, prim.Indices))
		assert.Equal(t, targetElements, d.BufferViews[idx.BufferView].Target)

		mat := d.Materials[prim.Material]
		assert.Equal(t, LinearColor(b.Color), mat.PBR.BaseColorFactor)
		assert.Equal(t, 0.0, mat.PBR.MetallicFactor)
	}
	for _, v := range d.BufferViews {
		assert.Equal(t, 0, v.ByteOffset%4, "every view starts 4-byte aligned")
	}
	lo, hi := bounds(tri.Positions)
	assert.Equal(t, []float64{0, 0, 0}, lo)
	assert.Equal(t, []float64{10, 5, 2.5}, hi)
}

// A body with no triangles is a node without a mesh, since glTF accessors
// cannot be empty; with nothing to store there is no BIN chunk or buffer.
func TestWriteEmptyBody(t *testing.T) {
	p := readGLB(t, write(t, []Body{{Name: "gone", Color: "#123456"}}))
	require.Len(t, p.doc.Nodes, 2)
	assert.Equal(t, "gone", p.doc.Nodes[1].Name)
	assert.Nil(t, p.doc.Nodes[1].Mesh)
	assert.Empty(t, p.doc.Meshes)
	assert.Empty(t, p.doc.Accessors)
	assert.Empty(t, p.doc.Buffers)
	assert.Nil(t, p.bin)

	p = readGLB(t, write(t, []Body{
		{Name: "gone"},
		{Name: "tri", Positions: []float32{0, 0, 0, 1, 0, 0, 0, 1, 0}, Indices: []uint32{0, 1, 2}},
	}))
	require.Len(t, p.doc.Nodes, 3)
	assert.Nil(t, p.doc.Nodes[1].Mesh)
	require.NotNil(t, p.doc.Nodes[2].Mesh)
	assert.Equal(t, 0, *p.doc.Nodes[2].Mesh)
}

func TestWriteRejectsMalformedBodies(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, []Body{{Name: "past", Positions: []float32{0, 0, 0, 1, 0, 0, 0, 1, 0}, Indices: []uint32{0, 1, 3}}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "past its 3 vertices")

	err = Write(&buf, []Body{{Name: "ragged", Positions: []float32{0, 0}, Indices: []uint32{0, 0, 0}}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not whole vertices")
}

func TestWriteNoBodies(t *testing.T) {
	var buf bytes.Buffer
	require.Error(t, Write(&buf, nil))
	assert.Equal(t, 0, buf.Len())
}

// Colors are sRGB "#rrggbb" decoded to glTF's linear baseColorFactor;
// anything else is #808080.
func TestLinearColor(t *testing.T) {
	assert.Equal(t, [4]float64{1, 1, 1, 1}, LinearColor("#ffffff"))
	assert.Equal(t, [4]float64{0, 0, 0, 1}, LinearColor("#000000"))
	gray := LinearColor("#808080")
	assert.True(t, math.Abs(gray[0]-0.2158605) < 1e-6, "sRGB 128 is linear 0.2158605, got %v", gray[0])
	assert.Equal(t, gray, LinearColor("red"))
	assert.Equal(t, gray, LinearColor("#80808"))
	assert.Equal(t, LinearColor("#ABCDEF"), LinearColor("#abcdef"))
}

// The root rotation takes the meshes' Z up to glTF's Y up, as the viewer's
// Z-up world turns: +Z goes to +Y and +Y to -Z, X stays.
func TestRootRotationTakesZUpToYUp(t *testing.T) {
	q := RootRotation
	rotate := func(v [3]float64) [3]float64 {
		// v' = v + 2w(u×v) + 2u×(u×v), u the quaternion's vector part.
		u := [3]float64{q[0], q[1], q[2]}
		cross := func(a, b [3]float64) [3]float64 {
			return [3]float64{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
		}
		uv := cross(u, v)
		uuv := cross(u, uv)
		return [3]float64{v[0] + 2*(q[3]*uv[0]+uuv[0]), v[1] + 2*(q[3]*uv[1]+uuv[1]), v[2] + 2*(q[3]*uv[2]+uuv[2])}
	}
	near := func(a, b [3]float64) bool {
		return math.Abs(a[0]-b[0]) < 1e-12 && math.Abs(a[1]-b[1]) < 1e-12 && math.Abs(a[2]-b[2]) < 1e-12
	}
	assert.True(t, near(rotate([3]float64{0, 0, 1}), [3]float64{0, 1, 0}), "+Z goes to +Y")
	assert.True(t, near(rotate([3]float64{0, 1, 0}), [3]float64{0, 0, -1}), "+Y goes to -Z")
	assert.True(t, near(rotate([3]float64{1, 0, 0}), [3]float64{1, 0, 0}), "X stays")
	norm := q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3]
	assert.True(t, math.Abs(norm-1) < 1e-15, "a unit quaternion, as glTF requires")
}
