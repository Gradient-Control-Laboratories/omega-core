package evalcmd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/cmd/compile/internal/evalcmd"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/assert"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
)

// plated is a 20 mm cube under a 0.1 mm coat.
const plated = "@scene(40)\n" +
	"blank = boxCentered(vector.zero, [20.0, 20.0, 20.0]);\n" +
	"blank.material = material.make(0.5, 0, 1.5);\n" +
	"coat = blank;\n" +
	"coat.material = material.withBuild(material.make(0.3, 0, 1.5), 0.1);\n" +
	"return coat;\n"

// "condition" sets where a field query reads the coated surface, request by request: a session's evaluator serves
// them all, and a request that names none reads the field as built.
func TestCondition_PerRequest(t *testing.T) {
	source := document(t, plated)
	quoted, err := json.Marshal(source)
	require.NoError(t, err)
	query := `"op": "eval", "source": ` + string(quoted) + `, "exprs": ["coat"], "points": [[10.05, 0, 0]]`
	lines := strings.Join([]string{
		`{"id": 1, ` + query + `, "condition": "nominal"}`,
		`{"id": 2, ` + query + `}`,
		`{"id": 3, ` + query + `, "condition": "as-made"}`,
		`{"id": 4, ` + query + `, "condition": "as-built"}`,
	}, "\n")
	var stdout, stderr bytes.Buffer
	code := evalcmd.Run(context.Background(), []string{"--stdio"}, strings.NewReader(lines), &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	responses := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	require.Len(t, responses, 4)
	for i, want := range []float64{-0.05, 0.05, -0.05, 0.05} {
		var r map[string]any
		require.NoError(t, json.Unmarshal([]byte(responses[i]), &r))
		require.Equal(t, true, r["ok"], r)
		near(t, want, first(r["groups"].([]any)[0]), 1e-5)
	}

	near10 := []float64{10.02, -1, -1, 10.08, 1, 1}
	r, code := ask(t, source, map[string]any{"op": "certify", "region": near10, "threshold": 0})
	require.Equal(t, 0, code, r)
	assert.Equal(t, "proven", r["verdict"])
	r, code = ask(t, source, map[string]any{"op": "certify", "region": near10, "threshold": 0, "condition": "nominal"})
	require.Equal(t, 0, code, r)
	assert.Equal(t, "disproven", r["verdict"])
}

// A condition the documents cannot express yet, one that is no condition, and one on an op that draws or describes
// the document are malformed requests.
func TestCondition_Refused(t *testing.T) {
	source := document(t, plated)
	for _, c := range []struct {
		req  map[string]any
		says string
	}{
		{map[string]any{"op": "eval", "exprs": []string{"coat"}, "points": [][3]float64{{0, 0, 0}}, "condition": "min-material"}, "relief depth"},
		{map[string]any{"op": "bounds", "condition": "as-drawn"}, "as-built, nominal, as-made or max-material"},
		{map[string]any{"op": "describe", "condition": "nominal"}, "describe does not take it"},
		{map[string]any{"op": "render", "out": "x.png", "condition": "as-made"}, "render does not take it"},
	} {
		r, code := ask(t, source, c.req)
		assert.Equal(t, 2, code, c.req)
		assert.Equal(t, false, r["ok"])
		assert.Contains(t, r["error"].(string), c.says)
	}
}

// toleranced is plated with its coat's thickness a working tolerance, 0.1 ± 0.02 mm.
var toleranced = strings.NewReplacer(
	"coat = blank;", "thick = 0.1;\nthick.tolerance = tolerance.symmetric(0.02);\ncoat = blank;",
	"material.withBuild(material.make(0.3, 0, 1.5), 0.1);", "material.withBuild(material.make(0.3, 0, 1.5), thick);",
).Replace(plated)

// certify over the document's tolerances: clear of the thickest coat it is
// proven, inside it disproven, the failing thicknesses named as shares of the
// tolerance.
func TestCertify_OverTolerances(t *testing.T) {
	source := document(t, toleranced)
	r, code := ask(t, source, map[string]any{"op": "certify", "region": []float64{10.13, -1, -1, 10.14, 1, 1},
		"threshold": 0, "condition": "nominal", "overTolerances": true})
	require.Equal(t, 0, code, r)
	assert.Equal(t, "proven", r["verdict"])

	r, code = ask(t, source, map[string]any{"op": "certify", "region": []float64{10.105, -1, -1, 10.11, 1, 1},
		"threshold": 0, "condition": "nominal", "overTolerances": true})
	require.Equal(t, 0, code, r)
	assert.Equal(t, "disproven", r["verdict"])
	share := r["counterParams"].(map[string]any)["thick"].([]any)
	assert.True(t, number(t, share[0]) >= 0.5, "fails only for coats past 0.11 mm: %v", share)

	r, code = ask(t, source, map[string]any{"op": "certify", "region": []float64{10.13, -1, -1, 10.14, 1, 1},
		"threshold": 0, "overTolerances": true, "overSliders": true})
	assert.Equal(t, 2, code, r)
	assert.Contains(t, r["error"].(string), "overTolerances")
}
