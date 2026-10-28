package convert_test

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/convert"
	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/fidget"
	assembler "gradientcontrol.com/omega/core/internal/assembler"
)

// A MULTI-RETURN function carries each part's colour and material out to the
// name it is destructured into (property/carry.go): each name gets the part at
// its own place in the return, a part the function leaves unpainted carries
// nothing, and a colour the caller sets on a name still wins.
func TestCarryMultiReturnProperties(t *testing.T) {
	src := `
trio(r) {
  a = sphere(r);
  a.color = [1.0, 0.0, 0.0];
  a.material = material.withDisplacement(material.standard, -0.1 * max(0.0, p.x));
  b = remap(sphere(r), p - [5.0, 0.0, 0.0]);
  b.color = [0.0, 0.0, 1.0];
  c = remap(sphere(r), p - [10.0, 0.0, 0.0]);
  return a, b, c;
}
x, y, z = trio(1.0);
u, v, w = trio(1.0);
v.color = [0.0, 1.0, 0.0];
return x, y, z, remap(v, p - [0.0, 20.0, 0.0]);
`
	res, err := assembler.Assemble([]string{src}, mockMgr)
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	err = convert.ToColorShapes(res.Program, 0, func(shapes []convert.ColorShape) error {
		if len(shapes) != 4 {
			t.Fatalf("got %d shapes, want 4", len(shapes))
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
			rgb       [3]float64
			at        [3]float32
			displaced bool
		}{
			{"x", true, [3]float64{1, 0, 0}, [3]float32{1, 0, 0}, true},
			{"y", true, [3]float64{0, 0, 1}, [3]float32{6, 0, 0}, false},
			{"z", false, [3]float64{}, [3]float32{}, false},
			{"v", true, [3]float64{0, 1, 0}, [3]float32{6, 20, 0}, false},
		}
		for i, w := range want {
			s := shapes[i]
			if s.HasColor != w.color {
				t.Errorf("%s: HasColor = %v, want %v", w.name, s.HasColor, w.color)
				continue
			}
			if s.Displaced != w.displaced {
				t.Errorf("%s: Displaced = %v, want %v", w.name, s.Displaced, w.displaced)
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
