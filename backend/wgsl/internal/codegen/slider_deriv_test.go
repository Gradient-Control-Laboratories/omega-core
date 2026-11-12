package codegen_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/assembler/slider"
	"gradientcontrol.com/omega/core/internal/source"
)

// TestSliderDerivative covers a derivative that depends on a slider (a literal
// lifted to a WGSL uniform). It must (1) not crash — synthesized derivative
// literals have no source span and previously nil-derefed the span-lifting loop
// in BuildValue — and (2) keep the slider live in the derivative (reference the
// uniform, not bake in its value), so dragging the slider updates the result.
func TestSliderDerivative(t *testing.T) {
	src := "radius = 5;\n" +
		"radius.min = 1;\n" +
		"radius.max = 10;\n" +
		"center = radius * [sin(time), cos(time), 0];\n" +
		"vel = center.dtime;\n" +
		"return length(p - center) - 1 + vel.x;\n"

	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}

	sliders := slider.Extract(src, res.Program.Primary)
	if len(sliders) != 1 {
		t.Fatalf("expected 1 slider, got %d", len(sliders))
	}
	param := sliders[0].ParamName

	rendered, err := renderHelper(src, res.Program)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// The slider literal must be lifted to a uniform...
	if !strings.Contains(rendered.WGSL, "let radius = "+param) {
		t.Fatalf("slider was not lifted to uniform %q:\n%s", param, rendered.WGSL)
	}
	// ...and the derivative-derived `vel` must reference the live `radius`
	// (tracking the slider), not bake in its literal value.
	var velLine string
	for _, line := range strings.Split(rendered.WGSL, "\n") {
		if strings.Contains(line, "let vel ") {
			velLine = line
		}
	}
	if velLine == "" {
		t.Fatalf("no vel binding in output:\n%s", rendered.WGSL)
	}
	if !strings.Contains(velLine, "radius") {
		t.Errorf("derivative does not track the slider (no radius reference):\n%s", velLine)
	}
}
