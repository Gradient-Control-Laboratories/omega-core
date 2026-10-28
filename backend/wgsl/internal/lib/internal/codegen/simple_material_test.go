package codegen_test

import (
	"strings"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
)

// A material is a mat4x4<f32>, and a stand-in's switch blends it column by
// column: a part's stand-in in a material of its own, and a material's flat
// look, its .simple, which the Textures rung draws.

// materialField is the body of the WGSL material field src renders to.
func materialField(t *testing.T, src string) string {
	t.Helper()
	res, err := assembler.Assemble([]string{src}, mockMgr)
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	out, err := renderHelper(src, res.Program)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	at := strings.Index(out.WGSL, "fn gsl_materialField(")
	if at < 0 {
		t.Fatalf("no material field")
	}
	body := out.WGSL[at:]
	return body[:strings.Index(body, "\n}\n")]
}

// A part's stand-in wears its own material: the field switches to it by the
// part's weight, here from a displaced matte to a plain chrome.
func TestSimpleStandInMaterial(t *testing.T) {
	field := materialField(t, `
bolt = boxCentered([0, 0, 0], [2, 2, 4]);
bolt.material = material.withDisplacement(material.matte, p.x * 0.01);
boltProxy = sphere(2);
boltProxy.material = material.chrome;
bolt.simple = boltProxy;
return bolt;
`)
	if !strings.Contains(field, "gsl_simple(") {
		t.Fatalf("the material field does not switch to the stand-in's:\n%s", field)
	}
	if !strings.Contains(field, "mat4x4<f32>(mix(") {
		t.Errorf("the material is not blended column by column:\n%s", field)
	}
}

// A material's flat look draws in its place below the Textures rung.
func TestSimpleValueMaterial(t *testing.T) {
	field := materialField(t, `
rough = material.withRoughness(material.matte, 0.5 + 0.5 * sin(10 * p.x));
rough.simple = material.matte;
bolt = boxCentered([0, 0, 0], [2, 2, 4]);
bolt.material = rough;
return bolt;
`)
	if !strings.Contains(field, "gsl_simpleDetail") {
		t.Errorf("the material does not switch to its flat look by the Textures rung:\n%s", field)
	}
}
