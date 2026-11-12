package server_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func postMesh(t *testing.T, url string, payload map[string]any) (*http.Response, []byte) {
	t.Helper()
	body, err := json.Marshal(payload)
	require.NoError(t, err)
	res, err := http.Post(url+"/api/mesh", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return res, data
}

// Several returns come back as several bodies: a zip of <name>.stl, or with
// format 3mf one 3MF holding an object each.
func TestMeshEndpointSeveralBodies(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	payload := map[string]any{
		"id":          "m-bodies",
		"source":      "ball = length(p - [8, 0, 0]) - 5;\nblk = box([-14, -4, -4], [-4, 4, 4]);\nreturn ball, blk;\n",
		"subdivision": 5,
		"bbox":        []float32{-20, -20, -20, 20, 20, 20},
		"algo":        "mdc",
	}

	res, data := postMesh(t, ts.URL, payload)
	require.Equal(t, http.StatusOK, res.StatusCode, string(data))
	assert.Equal(t, "application/zip", res.Header.Get("Content-Type"))
	assert.True(t, strings.Contains(res.Header.Get("Content-Disposition"), "model.zip"))
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	var names []string
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	assert.Equal(t, []string{"ball.stl", "blk.stl"}, names)

	payload["format"] = "3mf"
	res, data = postMesh(t, ts.URL, payload)
	require.Equal(t, http.StatusOK, res.StatusCode, string(data))
	assert.Equal(t, "model/3mf", res.Header.Get("Content-Type"))
	zr, err = zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	var model []byte
	for _, f := range zr.File {
		if f.Name == "3D/3dmodel.model" {
			rc, err := f.Open()
			require.NoError(t, err)
			model, err = io.ReadAll(rc)
			require.NoError(t, err)
			rc.Close()
		}
	}
	assert.Equal(t, 2, strings.Count(string(model), "<object "))

	payload["format"] = "obj"
	res, _ = postMesh(t, ts.URL, payload)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

// bodies narrows the export to the visible returns, and colors sets a 3MF
// body's color by its return index.
func TestMeshEndpointVisibleBodiesAndColors(t *testing.T) {
	srv, _, _ := setupTestServer(t)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	res, data := postMesh(t, ts.URL, map[string]any{
		"id":          "m-visible",
		"source":      "ball = length(p - [8, 0, 0]) - 5;\nblk = box([-14, -4, -4], [-4, 4, 4]);\nlid = length(p - [0, 10, 0]) - 3;\nreturn ball, blk, lid;\n",
		"subdivision": 5,
		"bbox":        []float32{-20, -20, -20, 20, 20, 20},
		"algo":        "mdc",
		"format":      "3mf",
		"bodies":      []int{1, 3},
		"colors":      map[string]string{"3": "#abcdef"},
	})
	require.Equal(t, http.StatusOK, res.StatusCode, string(data))
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	var model string
	for _, f := range zr.File {
		if f.Name == "3D/3dmodel.model" {
			rc, err := f.Open()
			require.NoError(t, err)
			b, err := io.ReadAll(rc)
			require.NoError(t, err)
			rc.Close()
			model = string(b)
		}
	}
	assert.Equal(t, 2, strings.Count(model, "<object "))
	assert.True(t, strings.Contains(model, `name="ball"`))
	assert.False(t, strings.Contains(model, `name="blk"`), "a hidden body is not exported")
	assert.True(t, strings.Contains(model, "   <m:color color=\"#8987FF\"/>\n   <m:color color=\"#ABCDEF\"/>\n"), "palette, then the color given")

	res, _ = postMesh(t, ts.URL, map[string]any{
		"id": "m-none", "source": "return length(p) - 5;\n", "subdivision": 4,
		"bbox": []float32{-8, -8, -8, 8, 8, 8}, "bodies": []int{},
	})
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}
