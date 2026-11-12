//go:build cgo

package eval_test

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// evalScalars evaluates each expression at one point and returns its values.
func evalScalars(t *testing.T, doc string, exprs []string, point [3]float64) []float64 {
	t.Helper()
	e := eval.New(&source.MockManager{}, nil)
	groups, err := e.EvalGsl(context.Background(), nil, doc, "", exprs, [][3]float64{point})
	require.NoError(t, err)
	out := make([]float64, len(exprs))
	for i := range exprs {
		out[i] = float64(groups[i].Samples[0][0].Value)
	}
	return out
}

func assertNear(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-4 {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

// TestNormalizeFieldInPlainDocument: builtin's normalizeField takes v.dp of
// its parameter, so a call needs a companion carrying the argument's
// gradient. A document with no derivative of its own used to be skipped by
// the expansion, leaving v.dp unexpanded; and the companion's tangent of
// cylinderY could not read vector.dirY.
func TestNormalizeFieldInPlainDocument(t *testing.T) {
	doc := `
s = normalizeField(sphere(1) * 2);
c = normalizeField(cylinderY(vector.zero, 1) * 2);
return s;
`
	got := evalScalars(t, doc, []string{"s", "c"}, [3]float64{0.6, 0.3, 0.2})
	// Doubling a distance and normalizing it gives the distance back.
	assertNear(t, "s", got[0], math.Sqrt(0.49)-1)
	assertNear(t, "c", got[1], math.Sqrt(0.40)-1)
}

// TestNormalizeCurve: a multi-return function whose body differentiates gets a
// companion that returns both values.
func TestNormalizeCurve(t *testing.T) {
	doc := `
dist, angle = normalizeCurve(sphere(1), p.x);
return dist - 0.5;
`
	got := evalScalars(t, doc, []string{"dist", "angle"}, [3]float64{0.6, 0.3, 0.2})
	// At (0.6, 0.3, 0.2): sphere(1) = -0.3 with gradient p/0.7, and p.x = 0.6
	// with gradient [1, 0, 0].
	n := [3]float64{0.6 / 0.7, 0.3 / 0.7, 0.2 / 0.7}
	sum := 0.3 / math.Sqrt((n[0]+1)*(n[0]+1)+n[1]*n[1]+n[2]*n[2])
	difference := -0.9 / math.Sqrt((n[0]-1)*(n[0]-1)+n[1]*n[1]+n[2]*n[2])
	assertNear(t, "dist", got[0], math.Hypot(sum, difference))
	assertNear(t, "angle", got[1], math.Atan2(difference, sum))
}

// TestCompanionParameterShadowsDeclaration: a document declaring its own `a`
// gets the same normalizeCurve as one that doesn't; the companion's parameter
// a is not the document's a.
func TestCompanionParameterShadowsDeclaration(t *testing.T) {
	doc := `
a = 5;
dist, angle = normalizeCurve(sphere(1), p.x);
return dist - 0.5 + a * 0;
`
	got := evalScalars(t, doc, []string{"dist", "angle"}, [3]float64{0.6, 0.3, 0.2})
	plain := evalScalars(t, `
dist, angle = normalizeCurve(sphere(1), p.x);
return dist - 0.5;
`, []string{"dist", "angle"}, [3]float64{0.6, 0.3, 0.2})
	assertNear(t, "dist", got[0], plain[0])
	assertNear(t, "angle", got[1], plain[1])
}

// TestCompanionTangentReadsQualifiedImport: inside a companion body, the
// tangent of an argument that reads vector.dirY resolves the vector package.
func TestCompanionTangentReadsQualifiedImport(t *testing.T) {
	doc := `
slope(d) {
  return d.dp.x;
}
outer(x) {
  return slope(dot(p, vector.dirX)) + x.dp.y;
}
o = outer(p.y);
return sphere(1);
`
	got := evalScalars(t, doc, []string{"o"}, [3]float64{0.6, 0.3, 0.2})
	assertNear(t, "o", got[0], 2)
}
