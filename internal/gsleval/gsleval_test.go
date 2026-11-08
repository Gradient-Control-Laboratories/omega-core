package gsleval_test

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/gsleval"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

var mockMgr = source.NewLocalManager("../../stdlib")

const testDoc = `
radius = 2.0;
s = length(p) - radius;
return s;
`

func TestEvalGsl(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)

	// Query a document name and an ad-hoc expression at two points.
	results, err := e.EvalGsl(context.Background(), nil, testDoc, "",
		[]string{"s", "length(p) - 0.5"},
		[][3]float64{{4, 0, 0}, {0, 0, 0}})
	require.NoError(t, err)
	require.Equal(t, 2, len(results))

	expected := [][]float32{
		{2, -2},    // s = |p| - 2
		{3.5, -.5}, // |p| - 0.5
	}
	for i, want := range expected {
		require.Equal(t, 1, len(results[i].Samples))
		samples := results[i].Samples[0]
		require.Equal(t, len(want), len(samples))
		for j, w := range want {
			if math.Abs(float64(samples[j].Value-w)) > 1e-4 {
				t.Errorf("expr %d point %d: value = %v, want %v", i, j, samples[j].Value, w)
			}
		}
	}

	// Gradient of s at (4,0,0) is radial: (1,0,0).
	g := results[0].Samples[0][0]
	if math.Abs(float64(g.Dx-1)) > 1e-4 || math.Abs(float64(g.Dy)) > 1e-4 || math.Abs(float64(g.Dz)) > 1e-4 {
		t.Errorf("gradient = (%v, %v, %v), want (1, 0, 0)", g.Dx, g.Dy, g.Dz)
	}
}

func TestEvalGslVectorExpr(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)

	// A vector expression decomposes into one group with a component per
	// element, so models can ask for whole vectors and matrices.
	results, err := e.EvalGsl(context.Background(), nil, testDoc, "",
		[]string{"[s, s + 1.0]"},
		[][3]float64{{4, 0, 0}})
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, 2, len(results[0].Samples))
	if math.Abs(float64(results[0].Samples[0][0].Value-2)) > 1e-4 {
		t.Errorf("component 0 = %v, want 2", results[0].Samples[0][0].Value)
	}
	if math.Abs(float64(results[0].Samples[1][0].Value-3)) > 1e-4 {
		t.Errorf("component 1 = %v, want 3", results[0].Samples[1][0].Value)
	}
}

func TestEvalGslDefs(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)

	// defs adds scratch variables in the document's scope; the query
	// expressions can use both defs names and document names.
	results, err := e.EvalGsl(context.Background(), nil, testDoc,
		"shift = 10.0; shifted = s + shift;",
		[]string{"shifted"},
		[][3]float64{{4, 0, 0}})
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	if math.Abs(float64(results[0].Samples[0][0].Value-12)) > 1e-4 {
		t.Errorf("value = %v, want 12", results[0].Samples[0][0].Value)
	}
}

func TestEvalGslDefsDedup(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)

	// A name assigned twice in defs keeps the last value instead of erroring
	// on an SSA collision.
	results, err := e.EvalGsl(context.Background(), nil, testDoc,
		"w = 1.0; w = 2.0;",
		[]string{"w"},
		[][3]float64{{0, 0, 0}})
	require.NoError(t, err)
	if math.Abs(float64(results[0].Samples[0][0].Value-2)) > 1e-4 {
		t.Errorf("value = %v, want 2 (last assignment wins)", results[0].Samples[0][0].Value)
	}
}

func TestEvalGslDefsOverrideDocument(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)

	// A def may redefine a name the document already declares; the def wins
	// for the query instead of colliding. testDoc declares radius = 2.0.
	results, err := e.EvalGsl(context.Background(), nil, testDoc,
		"radius = 5.0;",
		[]string{"radius"},
		[][3]float64{{0, 0, 0}})
	require.NoError(t, err)
	if math.Abs(float64(results[0].Samples[0][0].Value-5)) > 1e-4 {
		t.Errorf("value = %v, want 5 (def overrides document)", results[0].Samples[0][0].Value)
	}
}

func TestEvalGslDefsImportStripped(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)

	// Import statements in defs are dropped (libraries are auto-imported),
	// so they don't cause a mid-document parse error.
	results, err := e.EvalGsl(context.Background(), nil, testDoc,
		"import . \"omega\";\nx = 3.0;",
		[]string{"x"},
		[][3]float64{{0, 0, 0}})
	require.NoError(t, err)
	if math.Abs(float64(results[0].Samples[0][0].Value-3)) > 1e-4 {
		t.Errorf("value = %v, want 3", results[0].Samples[0][0].Value)
	}
}

func TestEvalGslMatrixExpr(t *testing.T) {
	e := gsleval.New(mockMgr, nil)

	// A whole 4x4 matrix comes back as one 16-component group.
	results, err := e.EvalGsl(context.Background(), nil, testDoc, "",
		[]string{"matrix.rotationEuler([0.0, 0.0, 0.0])"},
		[][3]float64{{0, 0, 0}})
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, 16, len(results[0].Samples))
	// Identity: components 0, 5, 10, 15 are 1 (column-major or row-major —
	// the diagonal is the diagonal either way).
	for _, idx := range []int{0, 5, 10, 15} {
		if math.Abs(float64(results[0].Samples[idx][0].Value-1)) > 1e-5 {
			t.Errorf("component %d = %v, want 1", idx, results[0].Samples[idx][0].Value)
		}
	}
}

func TestBoundsSphere(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)

	// The document's sphere has radius 2 at the origin; the fitted box
	// should be within a refined cell of [-2, 2]^3.
	r, err := e.Bounds(context.Background(), nil, testDoc, "", "s",
		[6]float64{-10, -10, -10, 10, 10, 10}, 32)
	require.NoError(t, err)
	require.True(t, r.Found)

	tol := 2 * r.CellSize[0]
	for k := 0; k < 3; k++ {
		if math.Abs(r.BboxMin[k]+2) > tol || math.Abs(r.BboxMax[k]-2) > tol {
			t.Errorf("bbox axis %d = [%v, %v], want about [-2, 2] (tol %v)",
				k, r.BboxMin[k], r.BboxMax[k], tol)
		}
	}
}

func TestEvalGslLibraryAutoImport(t *testing.T) {
	e := gsleval.New(mockMgr, nil)

	// The document imports nothing; query expressions still get the
	// built-in libraries by qualified name.
	results, err := e.EvalGsl(context.Background(), nil, testDoc, "",
		[]string{"vector.multiplyVector(matrix.rotationEuler([0.0, 0.0, 0.0]), [1.0, 2.0, 3.0]).x"},
		[][3]float64{{0, 0, 0}})
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	if math.Abs(float64(results[0].Samples[0][0].Value-1)) > 1e-5 {
		t.Errorf("value = %v, want 1", results[0].Samples[0][0].Value)
	}
}

func TestBoundsClampedField(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)

	// A narrow-band field in vdbSdf's shape: clamped to ±band away from the
	// surface (interior clamps negative). With coarse cells the half-diagonal
	// exceeds the band, so a threshold-only scan would mark every cell
	// occupied; the negative-first rule must still find the tight box.
	r, err := e.Bounds(context.Background(), nil, testDoc, "", "clamp(length(p) - 20.0, -1.7, 1.7)",
		[6]float64{-100, -100, -100, 100, 100, 100}, 16)
	require.NoError(t, err)
	require.True(t, r.Found)
	for k := 0; k < 3; k++ {
		if r.BboxMin[k] < -20-2*r.CellSize[k] || r.BboxMax[k] > 20+2*r.CellSize[k] {
			t.Errorf("bbox axis %d = [%v, %v], want within a couple cells of [-20, 20]",
				k, r.BboxMin[k], r.BboxMax[k])
		}
	}
}

func TestBoundsUninformativeClamp(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)

	// Pathological: an upper-only clamp whose inside region is smaller than
	// any cell. Every sample reads the constant background — there is no
	// information at this resolution, and the honest answer is "not found",
	// not a whole-region box.
	r, err := e.Bounds(context.Background(), nil, testDoc, "", "min(length(p) - 2.0, 1.7)",
		[6]float64{-100, -100, -100, 100, 100, 100}, 16)
	require.NoError(t, err)
	require.False(t, r.Found)
}

func TestBoundsThinShellFallback(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)

	// A shell far thinner than any cell: no cell center is inside, so the
	// near-band fallback has to locate it.
	r, err := e.Bounds(context.Background(), nil, testDoc, "", "abs(length(p) - 50.0) - 0.1",
		[6]float64{-100, -100, -100, 100, 100, 100}, 16)
	require.NoError(t, err)
	require.True(t, r.Found)
	for k := 0; k < 3; k++ {
		if r.BboxMax[k] < 40 || r.BboxMax[k] > 60 {
			t.Errorf("bbox axis %d max = %v, want near 50", k, r.BboxMax[k])
		}
	}
}

func TestBoundsDefs(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)

	// The bounds expression can reference defs names.
	r, err := e.Bounds(context.Background(), nil, testDoc,
		"grown = s - 1.0;", "grown",
		[6]float64{-10, -10, -10, 10, 10, 10}, 32)
	require.NoError(t, err)
	require.True(t, r.Found)
	tol := 2 * r.CellSize[0]
	if math.Abs(r.BboxMax[0]-3) > tol {
		t.Errorf("bbox max x = %v, want about 3", r.BboxMax[0])
	}
}

func TestBoundsNothingInside(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)

	// Always-positive field: nothing inside, but minValue/minAt report the
	// nearest approach.
	r, err := e.Bounds(context.Background(), nil, testDoc, "", "length(p) + 100.0",
		[6]float64{-10, -10, -10, 10, 10, 10}, 16)
	require.NoError(t, err)
	require.False(t, r.Found)
	if r.MinValue < 100 || r.MinValue > 120 {
		t.Errorf("minValue = %v, want just above 100", r.MinValue)
	}
}

func TestEvalGslQueryParseError(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)
	_, err := e.EvalGsl(context.Background(), nil, testDoc, "",
		[]string{"length(p"},
		[][3]float64{{0, 0, 0}})
	require.Error(t, err)
}

func TestEvalGslUndefinedName(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)
	_, err := e.EvalGsl(context.Background(), nil, testDoc, "",
		[]string{"noSuchShape"},
		[][3]float64{{0, 0, 0}})
	require.Error(t, err)
}
