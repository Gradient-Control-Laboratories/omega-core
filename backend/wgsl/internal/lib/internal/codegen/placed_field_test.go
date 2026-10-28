package codegen_test

import (
	"strings"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
)

// TestPlacedFieldsMove places a finished screw, whose library sets its
// material and colour (the thread's relief and its paint), with
// transform.translate, and a coloured ball with a document's own wrapper of a
// remap: each field is read in its part's own frame, at p less the offset, as
// it is through a remap written out, and not about the world's origin.
func TestPlacedFieldsMove(t *testing.T) {
	const src = `
import "metal";
steelMat, steelColor = metal.steel();
place(s, o) {
  return transform.translate(s, o);
}
screw = transform.translate(fastener.shcs(4, 8, steelMat, fastener.threadIso), [50, 0, 0]);
ball = sphere(1);
ball.color = [p.x, 0, 0];
placed = place(ball, [0, 30, 0]);
return screw, placed;
`
	res, err := assembler.Assemble([]string{src}, mockMgr)
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	out, err := renderHelper(src, res.Program)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	field := func(name string) string {
		i := strings.Index(out.WGSL, name)
		if i < 0 {
			t.Fatalf("no %s in:\n%s", name, out.WGSL)
		}
		f := out.WGSL[i:]
		return f[:strings.Index(f, "\n}\n")]
	}
	color, material := field("fn gsl_colorField("), field("fn gsl_materialField(")
	if !strings.Contains(material, "p - vec3<f32>(50.0, 0.0, 0.0)") {
		t.Errorf("the screw's material is not read in its own frame:\n%s", material)
	}
	if !strings.Contains(color, "p - vec3<f32>(50.0, 0.0, 0.0)") {
		t.Errorf("the screw's colour is not read in its own frame:\n%s", color)
	}
	if !strings.Contains(color, "p - vec3<f32>(0.0, 30.0, 0.0)") {
		t.Errorf("the placed ball's colour is not read in its own frame:\n%s", color)
	}
}
