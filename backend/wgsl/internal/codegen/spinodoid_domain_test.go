package codegen_test

import (
	"strings"
	"testing"
)

func TestSpinodoidDomainMetadataAndConservativeOutsideBuild(t *testing.T) {
	const src = `// Scene Size: 300
upright = [[1, 0, 0], [0, 1, 0], [0, 0, 1]];
field = fastGaborSpinodoid(p, 2, [pi / 2, pi / 2, pi / 2], upright, 3, 8, 4096, 0, 0, 1, 1 + 0.001 * p.y, upright, [[-150, -150, -150], [150, 150, 150]]);
gradient = fastGaborSpinodoidGradient(p, 3, [pi / 2, pi / 2, pi / 2], upright, 3, 8, 4096, 0, 0, 1, 1 + 0.001 * p.y, upright, [[-150, -150, -150], [150, 150, 150]]);
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

// The reference spinodoid is its own helper over its own table: the 27 cells
// about the point, two texels a kernel, behind the same slot header.
func TestReferenceSpinodoidBuildsItsOwnHelpers(t *testing.T) {
	const src = `upright = [[1, 0, 0], [0, 1, 0], [0, 0, 1]];
field = gaborSpinodoid(p, 2, [pi / 2, pi / 2, pi / 2], upright, 3, 8, 4096, 0, 0, 1, 1 + 0.001 * p.y, upright, [[-150, -150, -150], [150, 150, 150]]);
gradient = gaborSpinodoidGradient(p, 3, [pi / 2, pi / 2, pi / 2], upright, 3, 8, 4096, 0, 0, 1, 1 + 0.001 * p.y, upright, [[-150, -150, -150], [150, 150, 150]]);
fast = fastGaborSpinodoid(p, 2, [pi / 2, pi / 2, pi / 2], upright, 3, 8, 4096, 0, 0, 1, 1, upright, [[0, 0, 0], [0, 0, 0]]);
return field + length(gradient) * 0.01 + fast * 0.001;
`
	res, err := runTestBuildOne(t, "", []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"fn gsl_spinodoidRef(", "fn gsl_spinodoidRefGrad(", "fn gsl_spinodoid(",
		"gsl_spinodoidRef(", "gsl_spinodoidRefGrad(",
		"spinodoidTex, 0,", "spinodoidTex, 1,", "spinodoidTex, 2,",
		"let home = vec3<i32>(floor(q / cell));",
		"let index = (first + n) * 2;",
		"return length(outsideDelta);", "return normalize(outsideDelta);",
	} {
		if !strings.Contains(res.WGSL, want) {
			t.Errorf("generated shader lacks %q", want)
		}
	}
	if strings.Contains(res.WGSL, "fn gsl_spinodoidGrad(") {
		t.Error("no fast gradient was called, but its helper was emitted")
	}
}
