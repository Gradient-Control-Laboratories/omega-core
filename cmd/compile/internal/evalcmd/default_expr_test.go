package evalcmd_test

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// With no expr, a query asks of the document's returns however the return statement is led: by a comment above it,
// or by @scene when it is the first statement.
func TestExpr_TheReturnsAfterACommentOrScene(t *testing.T) {
	docs := map[string]string{
		"after a comment":    "@scene(40)\nball = length(p) - 10;\n// the part\nreturn ball;\n",
		"first after @scene": "@scene(40)\nreturn length(p) - 10;\n",
	}
	exact := 4.0 / 3 * math.Pi * 1000
	for form, doc := range docs {
		r, code := ask(t, document(t, doc), map[string]any{"op": "volume",
			"region": []float64{-12, -12, -12, 12, 12, 12}, "relativeTolerance": 0.05})
		require.Equal(t, 0, code, "%s: %v", form, r)
		assert.True(t, number(t, r["lowerMm3"]) <= exact && exact <= number(t, r["upperMm3"]), "%s: %v brackets %g",
			form, r, exact)
	}
}
