package convert_test

import (
	"math"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/fidget/internal/convert"
	"gradientcontrol.com/omega/core/internal/fidget/internal/fidget"
)

// Carriers compose (property/carry.go): a function that returns another
// carrier's part carries what it carries, through a local bound to the call,
// a call returned as it is, a wrapper of a wrapper, and one place of a
// destructured multi-return — here a library joint, whose screw and threaded
// part are steel with thread relief and whose clamped part has no finish.
func TestCarryThroughNestedCalls(t *testing.T) {
	src := `
painted(r, c) {
  s = sphere(r);
  s.color = c;
  s.material = material.withDisplacement(material.standard, -0.1 * max(0.0, p.x));
  return s;
}
bound(r, c) {
  s = painted(r, c);
  return s;
}
direct(r, c) {
  return painted(r, c);
}
twice(r, c) {
  return bound(r, c);
}
overridden(r) {
  s = painted(r, [1.0, 0.0, 0.0]);
  s.color = [0.0, 1.0, 0.0];
  return s;
}
fastened(lidBody, baseBody) {
  screw, lid, base = joint.bolted(lidBody, baseBody, [0.0, 0.0, 0.0], [0.0, 0.0, -1.0], 3.0, 3.0, fastener.fitMedium);
  return screw, lid, base;
}
a = bound(1.0, [1.0, 0.0, 0.0]);
b = remap(direct(1.0, [0.0, 0.0, 1.0]), p - [5.0, 0.0, 0.0]);
c = twice(1.0, [1.0, 1.0, 0.0]);
d = overridden(1.0);
screw, lid, base = fastened(max(abs(p.x) - 10.0, abs(p.y) - 10.0, p.z, -3.0 - p.z),
                            max(abs(p.x) - 10.0, abs(p.y) - 10.0, -3.0 - p.z, -15.0 - p.z));
return a, b, c, d, screw, lid, base;
`
	res, err := assembler.Assemble([]string{src}, mockMgr)
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	err = convert.ToColorShapes(res.Program, 0, func(shapes []convert.ColorShape) error {
		if len(shapes) != 7 {
			t.Fatalf("got %d shapes, want 7", len(shapes))
		}
		colourAt := func(s convert.ColorShape, x, y, z float32) [3]float64 {
			var out [3]float64
			for c := 0; c < 3; c++ {
				r, err := fidget.EvalGrad(fidget.BuildCanvas(s.Color[c]), []float32{x}, []float32{y}, []float32{z})
				if err != nil {
					t.Fatalf("eval colour: %v", err)
				}
				out[c] = float64(r[0].Value)
			}
			return out
		}
		want := []struct {
			name      string
			color     bool
			rgb       []float64
			at        [3]float32
			displaced bool
		}{
			{"a", true, []float64{1, 0, 0}, [3]float32{1, 0, 0}, true},
			{"b", true, []float64{0, 0, 1}, [3]float32{6, 0, 0}, true},
			{"c", true, []float64{1, 1, 0}, [3]float32{1, 0, 0}, true},
			{"d", true, []float64{0, 1, 0}, [3]float32{1, 0, 0}, true},
			{"screw", true, nil, [3]float32{}, true},
			{"lid", false, nil, [3]float32{}, false},
			{"base", true, nil, [3]float32{}, true},
		}
		for i, w := range want {
			s := shapes[i]
			if s.HasColor != w.color {
				t.Errorf("%s: HasColor = %v, want %v", w.name, s.HasColor, w.color)
				continue
			}
			if s.HasMaterial != w.displaced || s.Displaced != w.displaced {
				t.Errorf("%s: HasMaterial = %v, Displaced = %v, want %v", w.name, s.HasMaterial, s.Displaced, w.displaced)
			}
			if w.rgb == nil {
				continue
			}
			got := colourAt(s, w.at[0], w.at[1], w.at[2])
			for c := 0; c < 3; c++ {
				if math.Abs(got[c]-w.rgb[c]) > 1e-4 {
					t.Errorf("%s: colour at %v = %v, want %v", w.name, w.at, got, w.rgb)
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("ToColorShapes: %v", err)
	}
}
