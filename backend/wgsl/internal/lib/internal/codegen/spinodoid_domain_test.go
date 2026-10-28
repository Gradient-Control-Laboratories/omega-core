package codegen_test

import (
	"strings"
	"testing"
)

func TestSpinodoidDomainMetadataAndConservativeOutsideBuild(t *testing.T) {
	const src = `// Scene Size: 300
upright = [[1, 0, 0], [0, 1, 0], [0, 0, 1]];
field = gaborSpinodoid(p, 2, [pi / 2, pi / 2, pi / 2], upright, 3, 8, 4096, 0, 0.5, 1, 1 + 0.001 * p.y, upright);
gradient = gaborSpinodoidGradient(p, 3, [pi / 2, pi / 2, pi / 2], upright, 3, 8, 4096, 0, 0.5, 1, 1 + 0.001 * p.y, upright);
return field + length(gradient) * 0.01;
`
	res, err := runTestBuildOne(t, "", []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	// A slot's data row is metadata, never slot*64: rectangular tables have
	// different row counts, and scalar/gradient helpers must both honor it.
	for _, want := range []string{
		"fn gsl_spinodoid(", "fn gsl_spinodoidGrad(",
		"spinodoidTex, 0,", "spinodoidTex, 1,",
		"let row = i32(originRow.w);",
		"let index = (wy * size.z + wz) * size.x + wx;",
		"row + index / 4096",
		"return length(outsideDelta);", "return normalize(outsideDelta);",
	} {
		if !strings.Contains(res.WGSL, want) {
			t.Errorf("generated shader lacks %q", want)
		}
	}
	if strings.Contains(res.WGSL, "return 1e20") || strings.Contains(res.WGSL, "spinodoidTex, 64,") {
		t.Fatal("shader retained an infinite outside value or fixed-size slot offset")
	}
}
