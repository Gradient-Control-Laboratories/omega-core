//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/fidget"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/fidget/convert"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
)

// Shaft Modeling, Phase 13: fixes to what the shaft work found.
//
//   13a  The box analysis folds a comparison of literals but not a logical
//        and/or/not of them, so a select on `1 && 0` keeps both branches,
//        and a branch with no box leaves the whole box unbounded.
//   13b  A part's box corners are written into its vanish guard as one tree,
//        so the subexpressions a turn shares are written again at every use,
//        and nested turns multiply them.
//   13c  The shaft's fold bridges its first segment with the seed, a
//        bridge that can never win, and evaluates the segment for it anyway.
//   13d  The PNG framing scan keeps only the cells whose centre is inside
//        once any is, and loses a part too thin for any centre to fall in.

// 13a: a select whose condition is a logical and of literals takes one
// branch, so the other's lack of a box does not leak into it. The remapped
// sphere has no box (the analysis clears a remap's), the plain one has.
func TestShaftFixes_BoxFoldsLogicalLiterals(t *testing.T) {
	for _, c := range []struct{ cond, want string }{
		{"1.0 > 0.5 && 1.0 < 0.5", "plain"},
		{"1.0 > 0.5 || 1.0 < 0.5", "remapped"},
		{"!(1.0 > 0.5)", "plain"},
	} {
		doc := fmt.Sprintf("picked = %s ? remap(sphere(1.0), [p.x, p.y, p.z * 0.5]) : sphere(2.0);\nreturn picked;\n", c.cond)
		lo, hi := readBox(t, doc, "picked")
		bounded := hi[0] < 1000
		if c.want == "plain" && (!bounded || math.Abs(hi[0]-2.0) > 0.02) {
			t.Errorf("%q: box %v .. %v, want the plain sphere's, +-2", c.cond, lo, hi)
		}
		if c.want == "remapped" && bounded {
			t.Errorf("%q: box %v .. %v, want unbounded (the remap's)", c.cond, lo, hi)
		}
	}
}

// 13a, as the shaft met it: a segment of the caller's own whose field branch
// is a helper that does not guard itself, reached by the box questions only
// through `want > 0.5 && want < 2.5`, still leaves the shaft its box.
func TestShaftFixes_UnguardedSegmentKeepsTheBox(t *testing.T) {
	doc := `
rod(dia, from, to) {
  seg(eDown, eUp, want) {
    return want > 0.5 && want < 2.5 ? [want < 1.5 ? from : to, 0.0, 0.0, 0.0]
         : (want > 4.5 ? [want > 5.5 ? nothing : dia / 2.0, 0.0, 0.0, 0.0]
                      : (want > 2.5 && want < 3.5 ? [nothing, 0.0, 0.0, 0.0]
                                                  : [rodField(dia, from, to), rodField(dia, from - 100000.0, to),
                                                     rodField(dia, from, to + 100000.0), rodField(dia, from - 100000.0, to + 100000.0)]));
  }
  return seg;
}
rodField(dia, low, high) {
  return max(length(p.xy) - dia / 2.0, low - p.z, p.z - high);
}
solid = shaft.body([shaft.diameter(10.0, 0.0, 5.0), rod(6.0, 5.0, 20.0)]);
return solid;
`
	lo, hi := readBox(t, doc, "solid")
	want := [2][3]float64{{-5, -5, 0}, {5, 5, 20}}
	for k := 0; k < 3; k++ {
		if math.Abs(lo[k]-want[0][k]) > 0.02 || math.Abs(hi[k]-want[1][k]) > 0.02 {
			t.Errorf("box %v .. %v, want %v .. %v", lo, hi, want[0], want[1])
			break
		}
	}
}

// fixesShader compiles a document to WGSL.
func fixesShader(t *testing.T, doc string) string {
	t.Helper()
	e := New(&source.MockManager{}, nil)
	res, err := e.Shader(t.Context(), nil, doc, "", ShaderOptions{Target: "wgsl", OutDir: t.TempDir()})
	if err != nil {
		t.Fatalf("shader: %v", err)
	}
	return res.Code
}

// 13b: a part turned once, twice and three times, its box sized by a slider
// so nothing folds. Each turn adds its own arithmetic, so the shader grows by
// about as much at each depth; written as trees, the third turn added more
// than four times what the second did.
func TestShaftFixes_TurnedPartBoxGrowsLinearly(t *testing.T) {
	sizes := make([]int, 4)
	for depth := 1; depth <= 3; depth++ {
		shape := "box([neg(r), neg(2.0 * r), neg(3.0 * r)], [r, 2.0 * r, 3.0 * r])"
		for k, turn := range []string{"rotateX", "rotateY", "rotateZ"}[:depth] {
			shape = fmt.Sprintf("transform.%s(%s, %s)", turn, shape, []string{"a", "b", "c"}[k])
		}
		doc := "@range(1, 5) r = 2.0;\n@range(0, 3) a = 0.3;\n@range(0, 3) b = 0.4;\n@range(0, 3) c = 0.5;\n" +
			"turnedBox = " + shape + ";\nreturn turnedBox;\n"
		code := fixesShader(t, doc)
		if !strings.Contains(code, "gsl_vanish(") {
			t.Fatalf("depth %d: the turned box is not a part", depth)
		}
		sizes[depth] = len(code)
	}
	first, second := sizes[2]-sizes[1], sizes[3]-sizes[2]
	t.Logf("WGSL by depth: %d, %d, %d bytes (+%d, +%d)", sizes[1], sizes[2], sizes[3], first, second)
	if second > 2*first {
		t.Errorf("the third turn added %d bytes, more than twice the second's %d", second, first)
	}
}

// 13c: the first segment is written into the shader no more often than a
// segment in the middle, which has neighbours on both sides to bridge to.
// Each section carries a radius no other part of the program has, so its
// copies can be counted.
func TestShaftFixes_FirstSegmentNotBridgedToTheSeed(t *testing.T) {
	code := fixesShader(t, `
solid = shaft.body([shaft.section(length(p.xy) - 11.1111, 0.0, 10.0),
                    shaft.section(length(p.xy) - 12.2222, 10.0, 20.0),
                    shaft.section(length(p.xy) - 13.3333, 20.0, 30.0)]);
return solid;
`)
	first, middle := strings.Count(code, "11.1111"), strings.Count(code, "12.2222")
	t.Logf("copies: first %d, middle %d, last %d", first, middle, strings.Count(code, "13.3333"))
	if first > middle {
		t.Errorf("the first segment is written %d times, more than the middle one's %d", first, middle)
	}
}

// 13c: dropping the first segment's bridges changes no value: the shaft reads
// as the min of its capped segments outside, and solid across both joints.
func TestShaftFixes_FirstSegmentFieldUnchanged(t *testing.T) {
	doc := `
solid = shaft.body([shaft.diameter(20.0, 0.0, 10.0), shaft.diameter(30.0, 10.0, 20.0), shaft.diameter(16.0, 20.0, 40.0)]);
return solid;
`
	e := New(nil, nil)
	// Outside, the nearest point is a side, an end face or a shoulder's
	// corner: (20, 0, 5) is nearest the 30 mm segment's corner at (15, 10),
	// and (20, 0, 30) the same segment's corner at (15, 20).
	pts := [][3]float64{{20, 0, 5}, {20, 0, 15}, {20, 0, 30}, {0, 0, -5}, {0, 0, 45}, {5, 0, 10}, {5, 0, 20}, {12, 0, 5}}
	want := []float64{math.Sqrt(50), 5, math.Sqrt(125), 5, 5, -5, -3, 2}
	g, err := e.EvalGsl(context.Background(), nil, doc, "", []string{"solid"}, pts)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	for i := range pts {
		got := float64(g[0].Samples[0][i].Value)
		if want[i] > 0 && math.Abs(got-want[i]) > 1e-4 {
			t.Errorf("at %v: %.5f, want %.5f", pts[i], got, want[i])
		}
		if want[i] < 0 && got > -0.5 {
			t.Errorf("at %v: %.5f, want solid (the joint is a sheet)", pts[i], got)
		}
	}
}

// 13d: framing finds a rod too thin for any of the coarse scan's cell centres
// to fall inside, beside a block that holds several: the framed sphere reaches
// the rod's far end.
func TestShaftFixes_FramingKeepsAThinPart(t *testing.T) {
	src := `
block = box([-60.0, -60.0, -60.0], [60.0, 60.0, 60.0]);
rod = max(length(p.yz) - 3.0, 60.0 - p.x, p.x - 200.0);
return min(block, rod);
`
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	var center [3]float32
	var scale float32
	framed := false
	err = convert.ToTrees(res.Program, func(trees []fidget.TreeIndex, _ []string) error {
		center, scale, framed = frameCanvas(fidget.BuildCanvas(trees[0]))
		return nil
	})
	if err != nil || !framed {
		t.Fatalf("framing failed: %v (framed %v)", err, framed)
	}
	radius := float64(scale) / 1.25
	far := math.Hypot(math.Hypot(195-float64(center[0]), float64(center[1])), float64(center[2]))
	t.Logf("framed at %v, radius %.1f; the rod's end is %.1f from the centre", center, radius, far)
	if far > radius {
		t.Errorf("the rod's far end lies %.1f from the framed centre, outside its radius %.1f", far, radius)
	}
}

// Each shaft segment evaluates its 2D section once and returns all four
// capped/extended distances as a vec4, preventing 4x WGSL duplication.
func TestShaftFixes_SegmentSectionEvaluatedOnce(t *testing.T) {
	cases := []struct {
		name     string
		maxBytes int
	}{
		{"shaft_pinion", 45000}, // was 87,263 B before vec4 segment return
		{"shaft_hobbed", 40000}, // was 50,963 B before vec4 segment return; includes .simple stand-in across remap
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src, err := os.ReadFile("../examples/" + tc.name + ".gsl")
			if err != nil {
				t.Fatal(err)
			}
			code := fixesShader(t, string(src))
			t.Logf("%s WGSL size: %d bytes", tc.name, len(code))
			if len(code) > tc.maxBytes {
				t.Errorf("%s WGSL is %d bytes, want <= %d", tc.name, len(code), tc.maxBytes)
			}
		})
	}
}
