package convert_test

import (
	"math"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/fidget/internal/convert"
	"gradientcontrol.com/omega/core/internal/fidget/internal/fidget"
)

// A colour and a material a function sets on the part it returns come back
// with the part (property/carry.go): on a name bound to the call, on a call
// nested in a remap, which is sampled in the remapped frame, and not where the
// caller sets the property itself. A function that sets none carries none.
func TestCarryReturnProperties(t *testing.T) {
	src := `
painted(r, c) {
  s = sphere(r);
  s.color = c;
  s.material = material.withDisplacement(material.standard, -0.1 * max(0.0, p.x));
  return s;
}
plainBall(r) {
  return sphere(r);
}
a = painted(1.0, [1.0, 0.0, 0.0]);
b = remap(painted(1.0, [0.0, 0.0, 1.0]), p - [5.0, 0.0, 0.0]);
c = painted(1.0, [1.0, 0.0, 0.0]);
c.color = [0.0, 1.0, 0.0];
d = remap(plainBall(1.0), p - [-5.0, 0.0, 0.0]);
return a, b, c, d;
`
	res, err := assembler.Assemble([]string{src}, mockMgr)
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	err = convert.ToColorShapes(res.Program, 0, func(shapes []convert.ColorShape) error {
		if len(shapes) != 4 {
			t.Fatalf("got %d shapes, want 4", len(shapes))
		}
		// colourAt evaluates a shape's colour at one point.
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
			rgb       [3]float64
			at        [3]float32
			displaced bool
		}{
			{"a", true, [3]float64{1, 0, 0}, [3]float32{1, 0, 0}, true},
			{"b", true, [3]float64{0, 0, 1}, [3]float32{6, 0, 0}, true},
			{"c", true, [3]float64{0, 1, 0}, [3]float32{1, 0, 0}, true},
			{"d", false, [3]float64{}, [3]float32{}, false},
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
			if !w.color {
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
