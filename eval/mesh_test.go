//go:build cgo

package eval_test

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/eval"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/assert"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
)

// meshBallDoc is a ball about the origin whose radius is a slider.
const meshBallDoc = "@scene(40)\n@range(2, 12) r = 5.0;\nball = length(p) - r;\nreturn ball;\n"

// glbNode is one body of a .glb read back: its name, and its positions and normals, xyz per vertex.
type glbNode struct {
	name      string
	positions []float32
	normals   []float32
}

// readGLBNodes reads the bodies of a .glb as Mesh writes it: the root's children, each with one primitive.
func readGLBNodes(t *testing.T, data []byte) []glbNode {
	t.Helper()
	require.True(t, len(data) > 20, "a .glb")
	require.Equal(t, "glTF", string(data[0:4]))
	jsonLen := int(binary.LittleEndian.Uint32(data[12:16]))
	var doc struct {
		Nodes []struct {
			Name string `json:"name"`
			Mesh *int   `json:"mesh"`
		} `json:"nodes"`
		Meshes []struct {
			Primitives []struct {
				Attributes map[string]int `json:"attributes"`
			} `json:"primitives"`
		} `json:"meshes"`
		Accessors []struct {
			BufferView int `json:"bufferView"`
			Count      int `json:"count"`
		} `json:"accessors"`
		BufferViews []struct {
			ByteOffset int `json:"byteOffset"`
			ByteLength int `json:"byteLength"`
		} `json:"bufferViews"`
	}
	require.NoError(t, json.Unmarshal(data[20:20+jsonLen], &doc))
	bin := data[20+jsonLen+8:]
	floats := func(accessor int) []float32 {
		view := doc.BufferViews[doc.Accessors[accessor].BufferView]
		out := make([]float32, view.ByteLength/4)
		for i := range out {
			out[i] = math.Float32frombits(binary.LittleEndian.Uint32(bin[view.ByteOffset+4*i:]))
		}
		return out
	}
	var nodes []glbNode
	for _, n := range doc.Nodes {
		if n.Mesh == nil {
			continue
		}
		attributes := doc.Meshes[*n.Mesh].Primitives[0].Attributes
		node := glbNode{name: n.Name, positions: floats(attributes["POSITION"])}
		if normal, ok := attributes["NORMAL"]; ok {
			node.normals = floats(normal)
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// reach is the farthest a node's vertices get from the origin.
func reach(n glbNode) float64 {
	far := 0.0
	for i := 0; i+2 < len(n.positions); i += 3 {
		far = math.Max(far, math.Sqrt(float64(n.positions[i]*n.positions[i]+n.positions[i+1]*n.positions[i+1]+n.positions[i+2]*n.positions[i+2])))
	}
	return far
}

// A ball meshes to one closed body whose volume is the ball's and whose normals are the field's gradient: radial,
// which no average of facets is exactly.
func TestMeshBall(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	var buf bytes.Buffer
	r, err := e.Mesh(t.Context(), nil, meshBallDoc, "", &buf, eval.MeshOptions{Subdivision: 6})
	require.NoError(t, err)

	require.Len(t, r.Bodies, 1)
	body := r.Bodies[0]
	assert.Equal(t, "ball", body.Name)
	assert.Equal(t, 1, body.Index)
	assert.Equal(t, 0, body.OpenEdges)
	assert.Equal(t, 0, body.NonManifoldEdges)
	exact := 4 * math.Pi * 125 / 3
	assert.True(t, math.Abs(body.VolumeMm3-exact) <= 0.02*exact, "mesh volume %g, want %g within 2%%", body.VolumeMm3, exact)
	for k := 0; k < 3; k++ {
		assert.True(t, r.Region[k] < -5 && r.Region[k+3] > 5, "the region used, %v, holds the ball", r.Region)
	}

	nodes := readGLBNodes(t, buf.Bytes())
	require.Len(t, nodes, 1)
	node := nodes[0]
	assert.Equal(t, "ball", node.name)
	assert.Equal(t, body.Vertices, len(node.positions)/3)
	require.Equal(t, len(node.positions), len(node.normals), "a normal per vertex")
	worst := 1.0
	for i := 0; i+2 < len(node.positions); i += 3 {
		px, py, pz := float64(node.positions[i]), float64(node.positions[i+1]), float64(node.positions[i+2])
		nx, ny, nz := float64(node.normals[i]), float64(node.normals[i+1]), float64(node.normals[i+2])
		length := math.Sqrt(nx*nx + ny*ny + nz*nz)
		require.True(t, math.Abs(length-1) < 1e-4, "normal %d has length %g", i/3, length)
		worst = math.Min(worst, (px*nx+py*ny+pz*nz)/math.Sqrt(px*px+py*py+pz*pz))
	}
	assert.True(t, worst > math.Cos(0.5*math.Pi/180), "every normal within half a degree of radial; the worst cosine is %g", worst)
}

// defs move the part that is meshed, and the box found for it.
func TestMeshAppliesDefs(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	var buf bytes.Buffer
	r, err := e.Mesh(t.Context(), nil, meshBallDoc, "r = 10.0;", &buf, eval.MeshOptions{Subdivision: 6})
	require.NoError(t, err)
	nodes := readGLBNodes(t, buf.Bytes())
	require.Len(t, nodes, 1)
	assert.True(t, math.Abs(reach(nodes[0])-10) < 0.2, "the vertices reach %g, want 10", reach(nodes[0]))
	exact := 4 * math.Pi * 1000 / 3
	assert.True(t, math.Abs(r.Bodies[0].VolumeMm3-exact) <= 0.02*exact, "mesh volume %g, want %g", r.Bodies[0].VolumeMm3, exact)
}

// Each return is a body of its own, named as the return is; a region given is the region used.
func TestMeshBodiesAndRegion(t *testing.T) {
	doc := "@scene(40)\nball = length(p - [8, 0, 0]) - 5;\nblk = box([-14, -4, -4], [-4, 4, 4]);\nreturn ball, blk;\n"
	e := eval.New(&source.MockManager{}, nil)
	region := [6]float64{-20, -20, -20, 20, 20, 20}
	var buf bytes.Buffer
	r, err := e.Mesh(t.Context(), nil, doc, "", &buf, eval.MeshOptions{Subdivision: 6, Region: &region})
	require.NoError(t, err)
	assert.Equal(t, region, r.Region)
	require.Len(t, r.Bodies, 2)
	assert.Equal(t, "ball", r.Bodies[0].Name)
	assert.Equal(t, "blk", r.Bodies[1].Name)
	assert.Equal(t, 2, r.Bodies[1].Index)
	assert.True(t, math.Abs(r.Bodies[1].VolumeMm3-640) <= 0.03*640, "the block's mesh volume %g, want 640", r.Bodies[1].VolumeMm3)
	nodes := readGLBNodes(t, buf.Bytes())
	require.Len(t, nodes, 2)
	assert.Equal(t, "ball", nodes[0].name)
	assert.Equal(t, "blk", nodes[1].name)
	for _, n := range nodes {
		assert.Equal(t, len(n.positions), len(n.normals), "%s has a normal per vertex", n.name)
	}
}

// A document that returns nothing has nothing to mesh.
func TestMeshNeedsAReturn(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	var buf bytes.Buffer
	_, err := e.Mesh(t.Context(), nil, "r = 5.0;\n", "", &buf, eval.MeshOptions{})
	assert.Error(t, err)
}
