package server_test

import (
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// format glb comes back as model/gltf-binary: one node per visible body,
// and colors set a body's material by its return index.
func TestMeshEndpointGLB(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	res, data := postMesh(t, ts.URL, map[string]any{
		"id":          "m-glb",
		"source":      "ball = length(p - [8, 0, 0]) - 5;\nblk = box([-14, -4, -4], [-4, 4, 4]);\nlid = length(p - [0, 10, 0]) - 3;\nreturn ball, blk, lid;\n",
		"subdivision": 5,
		"bbox":        []float32{-20, -20, -20, 20, 20, 20},
		"algo":        "mdc",
		"format":      "glb",
		"bodies":      []int{1, 3},
		"colors":      map[string]string{"3": "#ffffff"},
	})
	require.Equal(t, http.StatusOK, res.StatusCode, string(data))
	assert.Equal(t, "model/gltf-binary", res.Header.Get("Content-Type"))
	assert.Contains(t, res.Header.Get("Content-Disposition"), "model.glb")

	require.True(t, len(data) > 20, "a .glb")
	assert.Equal(t, "glTF", string(data[0:4]))
	assert.Equal(t, uint32(len(data)), binary.LittleEndian.Uint32(data[8:12]))
	n := binary.LittleEndian.Uint32(data[12:16])
	var doc struct {
		Nodes []struct {
			Name string `json:"name"`
			Mesh *int   `json:"mesh"`
		} `json:"nodes"`
		Meshes []struct {
			Primitives []struct {
				Material int `json:"material"`
			} `json:"primitives"`
		} `json:"meshes"`
		Materials []struct {
			PBR struct {
				BaseColorFactor []float64 `json:"baseColorFactor"`
			} `json:"pbrMetallicRoughness"`
		} `json:"materials"`
	}
	require.NoError(t, json.Unmarshal(data[20:20+n], &doc))
	require.Len(t, doc.Nodes, 3)
	assert.Equal(t, "ball", doc.Nodes[1].Name)
	assert.Equal(t, "lid", doc.Nodes[2].Name)
	require.NotNil(t, doc.Nodes[2].Mesh)
	lid := doc.Materials[doc.Meshes[*doc.Nodes[2].Mesh].Primitives[0].Material]
	assert.Equal(t, []float64{1, 1, 1, 1}, lid.PBR.BaseColorFactor)
}
