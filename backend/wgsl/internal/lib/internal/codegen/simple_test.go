package codegen_test

import (
	"strings"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
)

// TestSimpleStandIn builds a part with a simpler stand-in, name.simple. The
// viewer's switch to the stand-in is written inside the part's own guard, so
// neither model is evaluated where the part is skipped, and its colour blends
// to the stand-in's by the same weight.
func TestSimpleStandIn(t *testing.T) {
	const src = `
bolt = boxCentered([0, 0, 0], [2, 2, 4]) - 0.2;
bolt.color = [1, 0.5 + 0.5 * sin(10 * p.x), 0];
boltProxy = sphere(2);
boltProxy.color = [0, 0, 1];
bolt.simple = boltProxy;
plate = box([-10, -10, -3], [10, 10, -2]);
return min(bolt, plate);
`
	res, err := assembler.Assemble([]string{src}, mockMgr)
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	out, err := renderHelper(src, res.Program)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	wgsl := out.WGSL
	for _, want := range []string{
		"var<private> gsl_simpleRung: f32;",
		"fn gsl_simple(lo: vec3<f32>, hi: vec3<f32>) -> f32 {",
		"fn gsl_boxSdf(",
	} {
		if !strings.Contains(wgsl, want) {
			t.Errorf("missing %q", want)
		}
	}

	// The switch sits inside the guard's branch: the full model and the
	// stand-in are both read after the part's vanish weight is known.
	shape := wgsl[strings.Index(wgsl, "fn sdShape("):]
	shape = shape[:strings.Index(shape, "\n}\n")]
	vanish := strings.Index(shape, "gsl_vanish(")
	simple := strings.Index(shape, "gsl_simple(")
	crop := strings.Index(shape, "gsl_boxSdf(")
	if vanish < 0 || simple < 0 || crop < 0 || !(vanish < simple && simple < crop) {
		t.Errorf("switch not inside the part's guard (vanish %d, simple %d, crop %d):\n%s", vanish, simple, crop, shape)
	}
	if n := strings.Count(shape, "gsl_simple("); n != 1 {
		t.Errorf("%d switches in sdShape, want 1 (only bolt has a stand-in)", n)
	}

	// The colour field blends from the stand-in's blue to the part's own.
	color := wgsl[strings.Index(wgsl, "fn gsl_colorField("):]
	if !strings.Contains(color, "gsl_simple(") {
		t.Errorf("colour field does not blend to the stand-in:\n%s", color)
	}
}

// TestSimpleNotAPart rejects a stand-in on a value that is not a part: one a
// remap reads, here.
func TestSimpleNotAPart(t *testing.T) {
	const src = `
bolt = sphere(2);
bolt.simple = sphere(1);
return remap(bolt, p - [1, 0, 0]);
`
	_, err := assembler.Assemble([]string{src}, mockMgr)
	if err == nil || !strings.Contains(err.Error(), "bolt.simple: bolt is not a part") {
		t.Fatalf("got %v, want bolt.simple: bolt is not a part", err)
	}
}

// TestSimpleValue builds a colour with a stand-in, c.simple, declared in the
// document before the stand-in it names: the colour's switch reads the
// viewer's weight, and the stand-in is not a part's, so nothing in sdShape
// switches.
func TestSimpleValue(t *testing.T) {
	const src = `
stripes = [1, 0.5 + 0.5 * sin(10 * p.x), 0];
stripes.simple = flat;
flat = [1, 0.5, 0];
ball = sphere(2);
ball.color = stripes;
return ball;
`
	res, err := assembler.Assemble([]string{src}, mockMgr)
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	out, err := renderHelper(src, res.Program)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(out.WGSL, "var<private> gsl_simpleDetail: f32 = 1.0;") {
		t.Errorf("the colour's weight is not declared")
	}
	color := out.WGSL[strings.Index(out.WGSL, "fn gsl_colorField("):]
	if !strings.Contains(color, "if (gsl_simpleDetail > 0.0) {") {
		t.Errorf("colour field does not switch to the stand-in:\n%s", color)
	}
	if strings.Contains(out.WGSL, "gsl_simpleRung") {
		t.Errorf("a colour's stand-in switched a part's shape")
	}
}
