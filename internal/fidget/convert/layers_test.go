package convert_test

import (
	"math"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/fidget"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/fidget/convert"
)

// A coat set on a shape inside a return moves that shape, before the
// booleans that use it, by what the coat adds over what it inherits
// (property.LowerDisplacedReturns): a cut taken after it stays a plane, and a
// coat laid over the cut raises the cut face and the outer faces alike. Each
// point names the distance the lit renderer marches there.
func TestLayersDisplaceBeforeTheBooleans(t *testing.T) {
	const doc = `
blank = box([-10, -10, -10], [10, 10, 10]);
blank.material = material.standard;
blank.color = [0, 0, 1];
coat = blank;
coat.material = material.withBuild(material.standard, 0.5);
coat.color = [1, 0, 0];
cutter = box([5, -20, -20], [20, 20, 20]);
cut = differenceSharp(coat, cutter);
`
	for _, c := range []struct {
		name, ret string
		pts       [][3]float32
		want      []float32
	}{
		{
			// The coat stands 0.5 proud of the top; the cut face stays at x = 5.
			name: "paint then cut",
			ret:  "return cut;",
			pts:  [][3]float32{{0, 0, 10.5}, {0, 0, 10}, {5.25, 0, 0}, {5, 0, 0}},
			want: []float32{0, -0.5, 0.25, 0},
		},
		{
			// A second coat, 0.25 over the first, raises both by 0.25.
			name: "paint, cut, paint",
			ret: `again = cut;
again.material = coating.levelled(material.standard, material.withBuild(material.standard, 0.5), 0.25, 1.0);
return again;`,
			pts:  [][3]float32{{0, 0, 10.75}, {0, 0, 10.5}, {5.25, 0, 0}, {5, 0, 0}},
			want: []float32{0, -0.25, 0, -0.25},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			res, err := assembler.Assemble([]string{doc + c.ret}, assembler.SourceManager(mockMgr))
			if err != nil {
				t.Fatalf("assemble: %v", err)
			}
			xs, ys, zs := splitPoints(c.pts)
			if err := convert.ToColorShapesWithPixelSize(res.Program, 0, func(shapes []convert.ColorShape, _ fidget.TreeIndex, _ bool, _ fidget.TreeIndex, _ bool) error {
				if len(shapes) != 1 || !shapes[0].HasInTree {
					t.Fatalf("want one shape displaced inside its booleans, got %d (in tree %v)", len(shapes), len(shapes) == 1 && shapes[0].HasInTree)
				}
				got, err := fidget.EvalGrad(fidget.BuildCanvas(shapes[0].InTree), xs, ys, zs)
				if err != nil {
					t.Fatal(err)
				}
				for j, w := range c.want {
					if math.Abs(float64(got[j].Value-w)) > 1e-4 {
						t.Errorf("%v: marched distance %g, want %g", c.pts[j], got[j].Value, w)
					}
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// A cut through a coat shows what the coat was laid over inside the solid
// beneath it, and the coat itself in its own thickness and on its outer face.
func TestLayersCutShowsTheLayerBeneath(t *testing.T) {
	const doc = `
blank = box([-10, -10, -10], [10, 10, 10]);
blank.material = material.standard;
blank.color = [0, 0, 1];
coat = blank;
coat.material = material.withBuild(material.standard, 0.5);
coat.color = [1, 0, 0];
cutter = box([5, -20, -20], [20, 20, 20]);
cut = differenceSharp(coat, cutter);
return cut;
`
	res, err := assembler.Assemble([]string{doc}, assembler.SourceManager(mockMgr))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	// On the cut face in the wood, on the cut face in the coat, and on the
	// coat's outer face.
	pts := [][3]float32{{5, 0, 0}, {5, 0, 10.25}, {0, 0, 10.5}}
	wantRed := []float32{0, 1, 1}
	xs, ys, zs := splitPoints(pts)
	if err := convert.ToColorShapesWithPixelSize(res.Program, 0, func(shapes []convert.ColorShape, _ fidget.TreeIndex, _ bool, _ fidget.TreeIndex, _ bool) error {
		if !shapes[0].HasColor {
			t.Fatal("no colour field")
		}
		red, err := fidget.EvalGrad(fidget.BuildCanvas(shapes[0].Color[0]), xs, ys, zs)
		if err != nil {
			t.Fatal(err)
		}
		blue, err := fidget.EvalGrad(fidget.BuildCanvas(shapes[0].Color[2]), xs, ys, zs)
		if err != nil {
			t.Fatal(err)
		}
		for j, w := range wantRed {
			if math.Abs(float64(red[j].Value-w)) > 1e-4 || math.Abs(float64(blue[j].Value-(1-w))) > 1e-4 {
				t.Errorf("%v: colour (%g, _, %g), want red %g", pts[j], red[j].Value, blue[j].Value, w)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

// A coat reads the height beneath it at the closest point of the surface it
// stands on, not where it stands: a relief that varies through the volume
// (here 0.01 * z, wood's pores in the large) is the same under every coat. On
// the top, 10.6 is the wood's 0.1 plus the coat's 0.5; read at the coat's own
// height it would be 0.006 more.
func TestLayersReadTheSurfaceBeneath(t *testing.T) {
	src := `
blank = box([-10, -10, -10], [10, 10, 10]);
grain = material.withDisplacement(material.standard, 0.01 * p.z);
coat = blank;
coat.material = grain;
paint = coat;
paint.material = material.withBuild(grain, 0.5);
cutter = box([5, -20, -20], [20, 20, 20]);
cut = differenceSharp(paint, cutter);
return cut;
`
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mockMgr))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	xs, ys, zs := splitPoints([][3]float32{{0, 0, 10.6}})
	if err := convert.ToColorShapesWithPixelSize(res.Program, 0, func(shapes []convert.ColorShape, _ fidget.TreeIndex, _ bool, _ fidget.TreeIndex, _ bool) error {
		got, err := fidget.EvalGrad(fidget.BuildCanvas(shapes[0].InTree), xs, ys, zs)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(float64(got[0].Value)) > 1e-4 {
			t.Errorf("(0, 0, 10.6): marched distance %g, want 0 (the heights read at z = 10)", got[0].Value)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

// A return that is one shape with its material set on it is moved by that
// material as a whole, as it always was: it has no displaced distance of its
// own.
func TestLayersWholeReturnUnchanged(t *testing.T) {
	src := "ball = sphere(2);\nball.material = material.withDisplacement(material.standard, 0.3);\nmoved = remap(ball, p - [1, 0, 0]);\nreturn moved;\n"
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mockMgr))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	if d := res.Program.DisplacedReturns(); d != nil {
		t.Fatalf("displaced returns %v, want none", d)
	}
}

func splitPoints(pts [][3]float32) (xs, ys, zs []float32) {
	for _, p := range pts {
		xs = append(xs, p[0])
		ys = append(ys, p[1])
		zs = append(zs, p[2])
	}
	return xs, ys, zs
}
