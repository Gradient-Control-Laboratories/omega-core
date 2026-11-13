package eval_test

import (
	"testing"

	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/test/assert"
)

// A return's expressions are its own text, whatever leads the statement: a comment above it, the @scene directive
// when it is the document's first statement, or both.
func TestTopLevelReturnExprsSkipWhatLeadsTheReturn(t *testing.T) {
	docs := map[string]string{
		"after a declaration":        "@scene(40)\nball = length(p) - 10;\nreturn ball;\n",
		"after a comment":            "@scene(40)\nball = length(p) - 10;\n// the part\nreturn ball;\n",
		"after a block comment":      "@scene(40)\nball = length(p) - 10;\n/* the part */ return ball;\n",
		"first after @scene":         "@scene(40)\nreturn ball;\n",
		"after @scene and a comment": "@scene(40)\n// the part\nreturn ball;\n",
	}
	for form, doc := range docs {
		assert.Equal(t, []string{"ball"}, eval.TopLevelReturnExprs(doc), form)
	}
	assert.Equal(t, []string{"a", "max(b, -c)"},
		eval.TopLevelReturnExprs("@scene(40)\n// two\nreturn a, max(b, -c);\n"))
}
