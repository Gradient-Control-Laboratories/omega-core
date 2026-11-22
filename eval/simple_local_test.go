//go:build cgo

package eval

import (
	"regexp"
	"testing"
)

// standIns finds a stand-in declared for a colour whose stand-in is base: the
// viewer's WGSL starts it at the stand-in, `var gsl_simpleValue… = base`,
// wherever the colour's expression is lowered (inline in gsl_colorField, or in
// a helper it calls).
func standInFor(code, base string) bool {
	return regexp.MustCompile(`var gsl_simpleValue[_0-9]* = ` + regexp.QuoteMeta(base) + `;`).MatchString(code)
}

// A colour built in a local, given a stand-in there, and set on a part
// inside a function keeps its stand-in out of the function, as a colour
// returned from one does: the viewer draws the stand-in below the Detail
// slider's Textures rung. Both parts here must carry it.
func TestSimpleValue_LocalColourCarried(t *testing.T) {
	code := fixesShader(t, `
@range(0, 1) fade = 0.5;
shade(base) {
  painted = base * (1.0 - 0.5 * fade * (0.5 + 0.5 * sin(p.x * 3.0)));
  painted.simple = base;
  return painted;
}
local() {
  s = sphere(1.0);
  t = [0.8, 0.2, 0.2] * (1.0 - 0.5 * fade * (0.5 + 0.5 * sin(p.x * 3.0)));
  t.simple = [0.8, 0.2, 0.2];
  s.color = t;
  return s;
}
viaCall() {
  s = sphere(1.0);
  s.color = shade([0.2, 0.8, 0.2]);
  return s;
}
a = local();
b = transform.translate(viaCall(), [3.0, 0.0, 0.0]);
return a, b;
`)
	if !standInFor(code, "vec3<f32>(0.8, 0.2, 0.2)") {
		t.Errorf("the colour built in a local lost its stand-in")
	}
	if !standInFor(code, "vec3<f32>(0.2, 0.8, 0.2)") {
		t.Errorf("the colour returned from a function lost its stand-in")
	}
}

// The finished knurled shaft's paint goes plain steel with textures off, as a
// fastener's thread paint does: its colour carries a stand-in.
func TestShaftKnurl_PaintGoesPlainWithTextures(t *testing.T) {
	code := fixesShader(t, `
segs = [shaft.diameter(16.0, -10.0, 0.0), shaft.knurl(20.0, 1.0, shaft.knurlDiamond, 0.0, 20.0), shaft.diameter(16.0, 20.0, 30.0)];
steelMat, steelColor = metal.steel();
knurled = shaft.body(segs, steelMat);
return knurled;
`)
	if !regexp.MustCompile(`var gsl_simpleValue[_0-9]* = `).MatchString(code) {
		t.Errorf("the knurl's paint has no stand-in: it stays painted with textures off")
	}
}
