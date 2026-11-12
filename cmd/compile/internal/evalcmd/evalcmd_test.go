package evalcmd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/cmd/compile/internal/evalcmd"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// ball is a sphere of radius r about the origin; r is a slider, the first name declared.
const ball = "@scene(40)\n@range(5, 15) r = 10.0;\nball = length(p) - r;\nreturn ball;\n"

// document writes text to a .gsl in a fresh folder: its path.
func document(t *testing.T, text string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "part.gsl")
	require.NoError(t, os.WriteFile(path, []byte(text), 0644))
	return path
}

// ask runs one request (a JSON object, its "source" filled in when it names no document) through `compile eval -`:
// the decoded response and the exit code.
func ask(t *testing.T, source string, req map[string]any) (map[string]any, int) {
	t.Helper()
	_, hasCode := req["code"]
	_, hasSource := req["source"]
	if !hasCode && !hasSource {
		req["source"] = source
	}
	data, err := json.Marshal(req)
	require.NoError(t, err)
	var stdout, stderr bytes.Buffer
	code := evalcmd.Run(context.Background(), []string{"-"}, bytes.NewReader(data), &stdout, &stderr)
	var response map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &response), stdout.String()+stderr.String())
	return response, code
}

func number(t *testing.T, v any) float64 {
	t.Helper()
	f, ok := v.(float64)
	require.True(t, ok, "a number, not %v", v)
	return f
}

func near(t *testing.T, want float64, got any, tol float64) {
	t.Helper()
	f := number(t, got)
	assert.True(t, math.Abs(f-want) <= tol, "want %g within %g, got %g", want, tol, f)
}

// first is a group's first component at its first point.
func first(group any) any {
	return group.(map[string]any)["values"].([]any)[0].([]any)[0]
}

// Each expression's value and gradient at each point; .dp and .d(var) are expressions like any other.
func TestEval_ValuesGradientsAndDerivatives(t *testing.T) {
	r, code := ask(t, document(t, ball), map[string]any{"id": 7, "op": "eval",
		"exprs": []string{"ball", "ball.dp", "ball.d(r)"}, "points": [][3]float64{{20, 0, 0}, {0, 0, 4}}})
	require.Equal(t, 0, code, r)
	assert.Equal(t, 7.0, r["id"])
	groups := r["groups"].([]any)
	require.Len(t, groups, 3)

	value := groups[0].(map[string]any)
	assert.Equal(t, "ball", value["expr"])
	assert.Equal(t, "num", value["type"])
	values := value["values"].([]any)[0].([]any) // component 0, by point
	near(t, 10, values[0], 1e-5)
	near(t, -6, values[1], 1e-5)
	gradient := value["gradients"].([]any)[0].([]any)[1].([]any) // component 0, point 1
	near(t, 0, gradient[0], 1e-5)
	near(t, 1, gradient[2], 1e-5)

	dp := groups[1].(map[string]any)
	assert.Equal(t, "vec3<num>", dp["type"])
	require.Len(t, dp["values"].([]any), 3) // three components
	near(t, 1, first(dp), 1e-5)

	near(t, -1, first(groups[2]), 1e-5) // d(|p| - r)/dr
}

// defs naming a declared slider replaces it for that query only.
func TestEval_DefsSetASliderForOneQuery(t *testing.T) {
	source := document(t, ball)
	at := [][3]float64{{20, 0, 0}}
	r, code := ask(t, source, map[string]any{"op": "eval", "defs": "r = 12.0;", "exprs": []string{"ball"}, "points": at})
	require.Equal(t, 0, code, r)
	near(t, 8, first(r["groups"].([]any)[0]), 1e-5)

	r, _ = ask(t, source, map[string]any{"op": "eval", "exprs": []string{"ball"}, "points": at})
	near(t, 10, first(r["groups"].([]any)[0]), 1e-5)
}

// An extremum is a true sample, and its bound says the Lipschitz constant it assumed.
func TestExtremum_SaysWhatItAssumes(t *testing.T) {
	source := document(t, ball)
	region := []float64{-20, -20, -20, 20, 20, 20}
	r, code := ask(t, source, map[string]any{"op": "extremum", "region": region})
	require.Equal(t, 0, code, r)
	assert.Equal(t, "ball", r["expr"])
	assert.Equal(t, "min", r["mode"])
	assert.Equal(t, 1.0, r["assumes"].(map[string]any)["lipschitz"])
	near(t, -10, r["min"], 0.1)
	assert.True(t, number(t, r["minLower"]) <= number(t, r["min"]))
	_, hasMax := r["max"]
	assert.False(t, hasMax)

	r, _ = ask(t, source, map[string]any{"op": "extremum", "region": region, "mode": "both", "lipschitz": 2})
	assert.Equal(t, 2.0, r["assumes"].(map[string]any)["lipschitz"])
	near(t, 20*math.Sqrt(3)-10, r["max"], 0.5) // at the region's corners
}

// A proof holds over the region; over the sliders it names the values where it fails.
func TestCertify_ProvenThenDisprovenOverTheSliders(t *testing.T) {
	source := document(t, ball)
	outside := []float64{12, -2, -2, 16, 2, 2}
	r, code := ask(t, source, map[string]any{"op": "certify", "region": outside, "threshold": 0})
	require.Equal(t, 0, code, r)
	assert.Equal(t, "proven", r["verdict"])

	r, code = ask(t, source, map[string]any{"op": "certify", "region": outside, "threshold": 0, "overSliders": true})
	require.Equal(t, 0, code, r) // a disproof is an answer
	assert.Equal(t, "disproven", r["verdict"])
	assert.Equal(t, []any{5.0, 15.0}, r["params"].(map[string]any)["r"])
	failing := r["counterParams"].(map[string]any)["r"].([]any)
	assert.True(t, number(t, failing[0]) >= 12, "fails only where the ball reaches the region: %v", failing)
	require.Len(t, r["counterBox"].([]any), 6)

	r, _ = ask(t, source, map[string]any{"op": "certify", "region": outside, "threshold": 0,
		"params": map[string][2]float64{"r": {5, 11}}})
	assert.Equal(t, "proven", r["verdict"])
}

func TestBoundsSurfaceVolumeAndDescribe(t *testing.T) {
	source := document(t, ball)
	r, code := ask(t, source, map[string]any{"op": "bounds"})
	require.Equal(t, 0, code, r)
	assert.Equal(t, true, r["found"])

	r, _ = ask(t, source, map[string]any{"op": "surface", "points": [][3]float64{{30, 0, 0}, {0, -3, 0}}})
	samples := r["samples"].([]any)
	require.Len(t, samples, 2)
	near(t, 10, samples[0].(map[string]any)["point"].([]any)[0], 1e-3)
	near(t, -10, samples[1].(map[string]any)["point"].([]any)[1], 1e-3)
	assert.Equal(t, true, samples[0].(map[string]any)["converged"])

	r, _ = ask(t, source, map[string]any{"op": "volume", "region": []float64{-12, -12, -12, 12, 12, 12},
		"relativeTolerance": 0.05})
	exact := 4.0 / 3 * math.Pi * 1000
	assert.True(t, number(t, r["lowerMm3"]) <= exact && exact <= number(t, r["upperMm3"]), "%v brackets %g", r, exact)

	r, _ = ask(t, source, map[string]any{"op": "describe"})
	assert.Equal(t, []any{5.0, 15.0}, r["sliders"].(map[string]any)["r"])
	assert.Equal(t, []any{"ball"}, r["returns"])
}

// A volume request without relativeTolerance takes the default, 0.05, rather than failing.
func TestVolume_DefaultTolerance(t *testing.T) {
	r, code := ask(t, document(t, ball), map[string]any{"op": "volume", "region": []float64{-12, -12, -12, 12, 12, 12}})
	require.Equal(t, 0, code, r)
	lower, upper := number(t, r["lowerMm3"]), number(t, r["upperMm3"])
	exact := 4.0 / 3 * math.Pi * 1000
	assert.True(t, lower <= exact && exact <= upper, "%v brackets %g", r, exact)
	assert.True(t, upper-lower <= 0.05*upper*(1+1e-9), "%v is within the default 0.05", r)
}

// The union of the returns is the expr when none is named.
func TestExpr_DefaultsToTheUnionOfTheReturns(t *testing.T) {
	two := "@scene(40)\na = length(p - [5, 0, 0]) - 2;\nb = length(p + [5, 0, 0]) - 2;\nreturn a, b;\n"
	r, code := ask(t, document(t, two), map[string]any{"op": "bounds"})
	require.Equal(t, 0, code, r)
	assert.Equal(t, "min(a, b)", r["expr"])

	// A document with no sliders describes them as none; one that does not parse is an error.
	r, code = ask(t, document(t, two), map[string]any{"op": "describe"})
	require.Equal(t, 0, code, r)
	assert.Equal(t, map[string]any{}, r["sliders"])
	assert.Equal(t, []any{"a", "b"}, r["returns"])
	r, code = ask(t, document(t, "@scene(40)\na = = 2;\nreturn a;\n"), map[string]any{"op": "describe"})
	assert.Equal(t, 1, code, r)
}

// A query that fails is answered in the evaluator's words with exit 1; a request that cannot be read, or lacks what
// its op needs, with exit 2; neither prints anything but the response.
func TestFailures_AreAnsweredWithTheirExitCodes(t *testing.T) {
	source := document(t, ball)
	origin := [][3]float64{{0, 0, 0}}
	unit := []float64{0, 0, 0, 1, 1, 1}
	cases := []struct {
		req  map[string]any
		code int
		says string
	}{
		{map[string]any{"op": "eval", "exprs": []string{"nope"}, "points": origin}, 1, "undefined identifier: nope"},
		{map[string]any{"op": "eval", "source": filepath.Join(t.TempDir(), "gone.gsl"), "exprs": []string{"x"},
			"points": origin}, 1, "gone.gsl"},
		{map[string]any{"op": "eval", "code": "@scene(9)\nm = importMesh(\"a.stl\");\nreturn m;\n",
			"exprs": []string{"m"}, "points": origin}, 1, "mesh assets (importMesh)"},
		{map[string]any{"op": "certify", "threshold": 0}, 2, "names its region"},
		{map[string]any{"op": "certify", "region": unit}, 2, "names its threshold"},
		{map[string]any{"op": "extremum", "region": []float64{0, 0, 0, -1, 1, 1}}, 2, "max must exceed its min"},
		{map[string]any{"op": "volume", "region": unit, "defs": "r = 3.0;"}, 2, "no defs"},
		{map[string]any{"op": "eval", "exprs": []string{"ball"}}, 2, "names its points"},
		{map[string]any{"op": "bounds", "exprs": []string{"ball"}}, 2, "exprs is eval's"},
		{map[string]any{"op": "bounds", "regoin": unit}, 2, "unknown field \"regoin\""},
		{map[string]any{"op": "spin"}, 2, "op \"spin\""},
	}
	for _, c := range cases {
		r, code := ask(t, source, c.req)
		assert.Equal(t, c.code, code, c.req)
		assert.Equal(t, false, r["ok"])
		assert.Contains(t, r["error"].(string), c.says)
	}

	var stdout, stderr bytes.Buffer
	code := evalcmd.Run(context.Background(), []string{"-"}, strings.NewReader(`{"id": 4, "op": `), &stdout, &stderr)
	assert.Equal(t, 2, code)
	assert.Contains(t, stdout.String(), "a malformed request")
	assert.Equal(t, "", stderr.String())
}

// A session answers every line in turn, a failed request included, until stdin closes.
func TestSession_AnswersEveryLineAndGoesOn(t *testing.T) {
	quoted, err := json.Marshal(document(t, ball))
	require.NoError(t, err)
	source := `"source": ` + string(quoted)
	lines := strings.Join([]string{
		`{"id": "a", "op": "eval", ` + source + `, "exprs": ["nope"], "points": [[0, 0, 0]]}`,
		`not json`,
		``,
		`{"id": "b", "op": "eval", ` + source + `, "exprs": ["ball"], "points": [[20, 0, 0]]}`,
		`{"id": "c", "op": "describe", ` + source + `}`,
	}, "\n")
	var stdout, stderr bytes.Buffer
	code := evalcmd.Run(context.Background(), []string{"--stdio"}, strings.NewReader(lines), &stdout, &stderr)
	require.Equal(t, 0, code, stderr.String())
	responses := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	require.Len(t, responses, 4)
	var ids, oks []any
	for _, line := range responses {
		var r map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &r))
		ids = append(ids, r["id"])
		oks = append(oks, r["ok"])
	}
	assert.Equal(t, []any{"a", nil, "b", "c"}, ids)
	assert.Equal(t, []any{false, false, true, true}, oks)
}

func TestUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	assert.Equal(t, 0, evalcmd.Run(context.Background(), []string{"--help"}, nil, &stdout, &stderr))
	assert.Contains(t, stdout.String(), "gsl eval [-I DIR]... --stdio")
	assert.Equal(t, 2, evalcmd.Run(context.Background(), nil, nil, &stdout, &stderr))
	assert.Contains(t, stderr.String(), "name one request file")
}

// pngOf decodes the PNG at path: its size and how many colours it holds (sampled every 2 pixels).
func pngOf(t *testing.T, path string) (int, int, int) {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)
	b := img.Bounds()
	colours := map[[3]uint32]bool{}
	for y := b.Min.Y; y < b.Max.Y; y += 2 {
		for x := b.Min.X; x < b.Max.X; x += 2 {
			r, g, bl, _ := img.At(x, y).RGBA()
			colours[[3]uint32{r >> 8, g >> 8, bl >> 8}] = true
		}
	}
	return b.Dx(), b.Dy(), len(colours)
}

// render writes the picture compile -t png draws -- views tiled two across, a direction among them, the light
// ground -- at out, and says its size.
func TestRender_WritesThePictureAtOut(t *testing.T) {
	source := document(t, ball)
	out := filepath.Join(t.TempDir(), "pics", "ball.png")
	r, code := ask(t, source, map[string]any{"op": "render", "out": out, "views": "iso,-1,1,0.6", "width": 160,
		"height": 120, "light": true})
	require.Equal(t, 0, code, r)
	assert.Equal(t, out, r["out"])
	assert.Equal(t, 2.0*160+4, r["width"]) // two panels and the 4-pixel gap
	assert.Equal(t, 120.0, r["height"])
	w, h, colours := pngOf(t, out)
	assert.Equal(t, 324, w)
	assert.Equal(t, 120, h)
	assert.True(t, colours >= 20, "a shaded ball, not a blank canvas: %d colours", colours)

	r, _ = ask(t, source, map[string]any{"op": "render", "out": out})
	assert.Equal(t, 640.0, r["width"]) // compile -t png's size when none is named
	assert.Equal(t, 480.0, r["height"])
}

// defs and expr reach the picture: a slider's value at a fixed extent, a section cut by an expression.
func TestRender_DrawsDefsAndExpr(t *testing.T) {
	source := document(t, ball)
	dir := t.TempDir()
	draw := func(name string, extra map[string]any) []byte {
		t.Helper()
		req := map[string]any{"op": "render", "out": filepath.Join(dir, name), "width": 96, "height": 72, "extent": 20}
		for k, v := range extra {
			req[k] = v
		}
		r, code := ask(t, source, req)
		require.Equal(t, 0, code, r)
		data, err := os.ReadFile(filepath.Join(dir, name))
		require.NoError(t, err)
		return data
	}
	whole := draw("whole.png", nil)
	assert.False(t, bytes.Equal(whole, draw("small.png", map[string]any{"defs": "r = 5.0;"})))
	assert.False(t, bytes.Equal(whole, draw("cut.png", map[string]any{"expr": "max(ball, p.y)"})))
}

func TestRender_RefusesWhatItCannotDraw(t *testing.T) {
	source := document(t, ball)
	out := filepath.Join(t.TempDir(), "x.png")
	for _, c := range []struct {
		req  map[string]any
		says string
	}{
		{map[string]any{"op": "render"}, "names out"},
		{map[string]any{"op": "render", "out": out, "views": "sideways"}, "views:"},
		{map[string]any{"op": "render", "out": out, "width": -3}, "each side is at least 1"},
		{map[string]any{"op": "render", "out": out, "extent": -1}, "extent -1"},
	} {
		r, code := ask(t, source, c.req)
		assert.Equal(t, 2, code, c.req)
		assert.Contains(t, r["error"].(string), c.says)
	}
	_, err := os.Stat(out)
	assert.True(t, os.IsNotExist(err), "nothing is written for a refused request")
}
