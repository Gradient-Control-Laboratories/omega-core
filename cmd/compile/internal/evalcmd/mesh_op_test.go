package evalcmd_test

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/assert"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
)

// mesh writes the part, with defs applied, as a .glb at out, and answers with the box it meshed over and what each
// body's triangles say: closed, and the volume of the ball at the radius the defs set.
func TestMesh_WritesAGLBAndReportsItsBodies(t *testing.T) {
	out := filepath.Join(t.TempDir(), "parts", "ball.glb")
	r, code := ask(t, document(t, ball), map[string]any{"op": "mesh", "out": out, "defs": "r = 5.0;", "subdivision": 6})
	require.Equal(t, 0, code, r)
	assert.Equal(t, out, r["out"])

	region := r["region"].([]any)
	require.Len(t, region, 6)
	assert.True(t, number(t, region[0]) < -5 && number(t, region[3]) > 5, "the region %v holds the ball", region)

	bodies := r["bodies"].([]any)
	require.Len(t, bodies, 1)
	body := bodies[0].(map[string]any)
	assert.Equal(t, "ball", body["name"])
	assert.Equal(t, 1.0, body["index"])
	assert.Equal(t, 0.0, body["openEdges"])
	assert.Equal(t, 0.0, body["nonManifoldEdges"])
	assert.True(t, number(t, body["vertices"]) > 100 && number(t, body["triangles"]) > 100, "%v has a mesh", body)
	exact := 4.0 / 3 * math.Pi * 125
	near(t, exact, body["volumeMm3"], 0.02*exact)

	data, err := os.ReadFile(out)
	require.NoError(t, err)
	require.True(t, len(data) > 20, "a file was written")
	assert.Equal(t, "glTF", string(data[0:4]))
}

// A region named is the region meshed, and each return is a body.
func TestMesh_RegionAndBodies(t *testing.T) {
	doc := "@scene(40)\nball = length(p - [8, 0, 0]) - 5;\nblk = box([-14, -4, -4], [-4, 4, 4]);\nreturn ball, blk;\n"
	out := filepath.Join(t.TempDir(), "two.glb")
	r, code := ask(t, document(t, doc), map[string]any{"op": "mesh", "out": out, "subdivision": 6,
		"region": []float64{-20, -20, -20, 20, 20, 20}, "algorithm": "mdc"})
	require.Equal(t, 0, code, r)
	assert.Equal(t, []any{-20.0, -20.0, -20.0, 20.0, 20.0, 20.0}, r["region"])
	bodies := r["bodies"].([]any)
	require.Len(t, bodies, 2)
	assert.Equal(t, "ball", bodies[0].(map[string]any)["name"])
	assert.Equal(t, "blk", bodies[1].(map[string]any)["name"])
	assert.Equal(t, 2.0, bodies[1].(map[string]any)["index"])
}

// mesh writes a file, so it names one; a subdivision or an algorithm the mesher does not have is a malformed
// request, not a failed one.
func TestMesh_MalformedRequests(t *testing.T) {
	source := document(t, ball)
	out := filepath.Join(t.TempDir(), "ball.glb")
	for _, c := range []struct {
		req  map[string]any
		says string
	}{
		{map[string]any{"op": "mesh"}, "names out"},
		{map[string]any{"op": "mesh", "out": out, "subdivision": -1}, "subdivision"},
		{map[string]any{"op": "mesh", "out": out, "algorithm": "guess"}, "algorithm"},
		{map[string]any{"op": "mesh", "out": out, "region": []float64{0, 0, 0, -1, 1, 1}}, "max must exceed its min"},
		{map[string]any{"op": "mesh", "out": out, "expr": "ball"}, "expr"},
	} {
		r, code := ask(t, source, c.req)
		assert.Equal(t, 2, code, "%v: %v", c.req, r)
		assert.Equal(t, false, r["ok"])
		msg, _ := r["error"].(string)
		assert.Contains(t, msg, c.says)
	}
	_, err := os.Stat(out)
	assert.True(t, os.IsNotExist(err), "no file from a malformed request")
}
