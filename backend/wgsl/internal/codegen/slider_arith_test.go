package codegen_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/slider"
	"gradientcontrol.com/omega/core/internal/source"
)

// TestSliderArithmetic covers a slider used inside otherwise-constant
// arithmetic. The assembler folds with no ParameterChecker, so every
// expression over the slider arrives at codegen already stamped with a
// constant; the codegen-side fold must clear those stamps rather than bake
// the slider's current value into the shader. Direct references
// (sphere(radius)) were always live — the derived forms are the regression.
func TestSliderArithmetic(t *testing.T) {
	src := "radius = 2;\n" +
		"radius.min = 1;\n" +
		"radius.max = 5;\n" +
		"sphere(r) { return length(p) - r; }\n" +
		"scale = radius * 2;\n" +
		"a = sphere(radius);\n" +
		"b = sphere(radius * 2);\n" +
		"c = sphere(radius + 1);\n" +
		"d = sphere(scale + 1);\n" +
		"return a, b, c, d;\n"

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
	// ...and every use — direct, in arithmetic, and through an intermediate
	// variable — must reference the live value, not a folded constant. Each of
	// a, b, c and d is a part, so its call is written inside the guard that
	// lets it shrink away (vanish.go) rather than as its own let.
	for _, want := range []string{
		"sphere(p, radius)",
		"sphere(p, radius * 2.0)",
		"sphere(p, radius + 1.0)",
		"let scale = radius * 2.0;",
		"sphere(p, scale + 1.0)",
	} {
		if !strings.Contains(rendered.WGSL, want) {
			t.Errorf("missing %q (slider baked in?):\n%s", want, rendered.WGSL)
		}
	}
}
