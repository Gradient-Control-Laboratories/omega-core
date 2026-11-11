package gsleval_test

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/gsleval"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// valueAt is expr's value at the origin with defs applied to doc.
func valueAt(t *testing.T, doc, defs, expr string) float64 {
	t.Helper()
	e := gsleval.New(&source.MockManager{}, nil)
	results, err := e.EvalGsl(context.Background(), nil, doc, defs, []string{expr}, [][3]float64{{0, 0, 0}})
	require.NoError(t, err)
	return float64(results[0].Samples[0][0].Value)
}

// A def overrides a name however the document declares it: first after the @scene directive (which has no `;`, so
// it leads the statement after it), behind an annotation, or with property lines (`r.min = ...`) that set and
// declare nothing.
func TestEvalGslDefsOverrideTheFormsDocumentsDeclare(t *testing.T) {
	docs := map[string]string{
		"first after @scene": "@scene(40)\nr = 10.0;\nball = length(p) - r;\nreturn ball;\n",
		"with .min and .max": "@scene(40)\nr = 10.0;\nr.min = 5.0;\nr.max = 15.0;\nball = length(p) - r;\nreturn ball;\n",
		"behind @range":      "@scene(40)\n@range(5, 15) r = 10.0;\nball = length(p) - r;\nreturn ball;\n",
		"after a comment":    "@scene(40)\n// the radius\n@range(5, 15) r = 10.0;\nball = length(p) - r;\nreturn ball;\n",
	}
	for form, doc := range docs {
		got := valueAt(t, doc, "r = 12.0;", "r")
		assert.True(t, math.Abs(got-12) < 1e-4, "%s: r = %v, want 12", form, got)
		got = valueAt(t, doc, "", "r")
		assert.True(t, math.Abs(got-10) < 1e-4, "%s: without the def r = %v, want 10", form, got)
	}
}

// An override replaces the declaration with its annotation, so a query may set a slider outside the range its
// control allows; the @scene directive before it stays.
func TestEvalGslDefsOverrideIsNotHeldToTheSliderRange(t *testing.T) {
	doc := "@scene(40)\n@range(5, 15) r = 10.0;\nball = length(p) - r;\nreturn ball;\n"
	got := valueAt(t, doc, "r = 20.0;", "r")
	assert.True(t, math.Abs(got-20) < 1e-4, "r = %v, want 20", got)
}

func TestTopLevelNamesSeeAnnotatedDeclarationsAndNotPropertyLines(t *testing.T) {
	doc := "@scene(40)\n@range(5, 15) r = 10.0;\nk = 2.0;\nk.min = 1.0;\nk.max = 3.0;\nball = length(p) - r * k;\n" +
		"ball.color = [1, 0, 0];\nreturn ball;\n"
	names := gsleval.TopLevelNames(doc)
	for _, name := range []string{"r", "k", "ball"} {
		assert.True(t, names[name], "%s is declared: %v", name, names)
	}
	assert.Equal(t, []string{"ball"}, gsleval.TopLevelReturnExprs(doc))
}
