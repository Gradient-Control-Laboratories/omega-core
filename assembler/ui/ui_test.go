package ui_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/assembler/ui"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/program/edit"
)

func extract(t *testing.T, src string) []ui.Control {
	t.Helper()
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
	assert.NoError(t, err)
	return ui.Extract(res.Program.Primary, nil)
}

func textAt(src string, span edit.Span) string {
	line := strings.Split(src, "\n")[span.StartLine-1]
	return line[span.StartColumn-1 : span.EndColumn-1]
}

func TestExtractDirectPoint(t *testing.T) {
	src := "return uiPoint([1, 2, 3]);\n"
	controls := extract(t, src)
	assert.Equal(t, 1, len(controls))
	assert.Equal(t, ui.KindPoint, controls[0].Kind)
	assert.Equal(t, []float64{1, 2, 3}, controls[0].Value)
	assert.Equal(t, "[1, 2, 3]", textAt(src, controls[0].Span))
	assert.Equal(t, 3, len(controls[0].Params))
	assert.Equal(t, []string{"gsl_param_1_17", "gsl_param_1_20", "gsl_param_1_23"}, []string{
		controls[0].Params[0].ParamName,
		controls[0].Params[1].ParamName,
		controls[0].Params[2].ParamName,
	})
	assert.Equal(t, "2", textAt(src, controls[0].Params[1].Span))
}

func TestExtractReferenceTransform(t *testing.T) {
	src := "m = [[1, 0, 0, 0], [0, 1, 0, 0], [0, 0, 1, 0], [3, 4, 5, 1]];\n" +
		"return uiTransform(m);\n"
	controls := extract(t, src)
	assert.Equal(t, 1, len(controls))
	assert.Equal(t, ui.KindTransform, controls[0].Kind)
	assert.Equal(t, []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 3, 4, 5, 1}, controls[0].Value)
	assert.Equal(t, "[[1, 0, 0, 0], [0, 1, 0, 0], [0, 0, 1, 0], [3, 4, 5, 1]]", textAt(src, controls[0].Span))
	assert.Equal(t, 16, len(controls[0].Params))
	assert.Equal(t, []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 3, 4, 5, 1}, values(controls[0].Params))
	assert.Equal(t, "3", textAt(src, controls[0].Params[12].Span))
}

func values(params []ui.Param) []float64 {
	out := make([]float64, len(params))
	for i, p := range params {
		out[i] = p.Value
	}
	return out
}
