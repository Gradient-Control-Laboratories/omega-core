package assembler_test

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/lang"
	internalsource "gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/program"
)

func TestAssemblePostfixModifierChaining(t *testing.T) {
	mgr := &internalsource.MockManager{
		PkgSources: map[string][]string{
			"geom": {`
mat = [[10, 20, 30], [40, 50, 60]];
body = sphere([2, 4, 6], 1.5);
return body;
`},
		},
	}

	src := `
import "geom";
m = [[1, 2, 3], [4, 5, 6]];
pts = [[10, 20, 30], [40, 50, 60]];
flags = [0, 7];
t = 3;
v = [2 * t, t * t, 5];
body = sphere([1, 2, 3], 0.5);

return
	m[0][1],
	m[1][2],
	-m[0][1],
	!flags[0],
	!flags[1],
	pts[0].x,
	pts[1].yz,
	pts[1].zyx[0],
	body.box[0],
	body.box[1].z,
	v.d(t)[0],
	v.d(t)[1],
	v.d(t).y,
	geom.mat[1][0],
	geom.body.box[0].y,
	["ABC"[1], 99];
`

	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mgr))
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	wants := []struct {
		label string
		vals  []float64
	}{
		{"m[0][1]", []float64{2}},
		{"m[1][2]", []float64{6}},
		{"-m[0][1]", []float64{-2}},
		{"!flags[0]", []float64{1}},
		{"!flags[1]", []float64{0}},
		{"pts[0].x", []float64{10}},
		{"pts[1].yz", []float64{50, 60}},
		{"pts[1].zyx[0]", []float64{60}},
		{"body.box[0]", []float64{0.5, 1.5, 2.5}},
		{"body.box[1].z", []float64{3.5}},
		{"v.d(t)[0]", []float64{2}},
		{"v.d(t)[1]", []float64{6}},
		{"v.d(t).y", []float64{6}},
		{"geom.mat[1][0]", []float64{40}},
		{"geom.body.box[0].y", []float64{2.5}},
		{`["ABC"[1], 99]`, []float64{66, 99}},
	}

	rets := res.Program.Returns()
	if len(rets) != len(wants) {
		t.Fatalf("returns count = %d, want %d", len(rets), len(wants))
	}

	for i, w := range wants {
		got, ok := program.ConstValues(rets[i])
		if !ok {
			t.Errorf("%s: expected constant values %v, got non-const %s", w.label, w.vals, rets[i].PrettyPrint(0))
			continue
		}
		if len(got) != len(w.vals) {
			t.Errorf("%s: got %v, want %v", w.label, got, w.vals)
			continue
		}
		for j := range w.vals {
			if math.Abs(got[j]-w.vals[j]) > 1e-6 {
				t.Errorf("%s[%d] = %v, want %v", w.label, j, got[j], w.vals[j])
			}
		}
	}
}

func TestAssemblePostfixModifierSemanticErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "indexing_scalar_after_2d_index",
			src:  "m = [[1, 2], [3, 4]];\nx = m[0][1][0];\nreturn sphere(x);\n",
			want: "index operator expects collection target",
		},
		{
			name: "non_scalar_index_in_chain",
			src:  "m = [[1, 2], [3, 4]];\nx = m[0][[0, 1]];\nreturn sphere(x);\n",
			want: "index operator expects scalar index",
		},
		{
			name: "swizzle_component_out_of_bounds_after_index",
			src:  "pts = [[1, 2]];\nx = pts[0].z;\nreturn sphere(x);\n",
			want: "z",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := assembler.Assemble([]string{tc.src})
			if err == nil {
				t.Fatalf("expected semantic error containing %q, got nil", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.want)
			}
		})
	}
}

func TestAssemble_AllowedModifierTransitions(t *testing.T) {
	const prelude = `
m = [[1, 2, 3], [4, 5, 6]];
pts = [[10, 20, 30], [40, 50, 60]];
a = sphere(1.0);
a.u = 1.0;
a.v = 2.0;
b = remap(sphere(1.0), p - [3.0, 0.0, 0.0]);
b.u = 3.0;
b.v = 4.0;
both = min(a, b);
shapes = [a, both];
`

	samples := map[lang.ModifierTransition]string{
		{Prev: lang.ModIndex, Cur: lang.ModIndex}:                 "m[0][1]",
		{Prev: lang.ModIndex, Cur: lang.ModSwizzleScalar}:         "pts[0].x",
		{Prev: lang.ModIndex, Cur: lang.ModSwizzleVec}:            "pts[0].xy",
		{Prev: lang.ModIndex, Cur: lang.ModDeriv}:                 "shapes[0].dp",
		{Prev: lang.ModIndex, Cur: lang.ModBox}:                   "shapes[0].box",
		{Prev: lang.ModIndex, Cur: lang.ModSecond}:                "shapes[1].second",
		{Prev: lang.ModIndex, Cur: lang.ModSurfaceFrame}:          "shapes[0].u",
		{Prev: lang.ModSwizzleScalar, Cur: lang.ModSwizzleScalar}: "p.x.x",
		{Prev: lang.ModSwizzleScalar, Cur: lang.ModSwizzleVec}:    "p.x.xy",
		{Prev: lang.ModSwizzleScalar, Cur: lang.ModDeriv}:         "p.x.dp",
		{Prev: lang.ModSwizzleVec, Cur: lang.ModIndex}:            "p.xy[0]",
		{Prev: lang.ModSwizzleVec, Cur: lang.ModSwizzleScalar}:    "p.xy.x",
		{Prev: lang.ModSwizzleVec, Cur: lang.ModSwizzleVec}:       "p.xy.yx",
		{Prev: lang.ModSwizzleVec, Cur: lang.ModDeriv}:            "p.xy.dp",
		{Prev: lang.ModDeriv, Cur: lang.ModIndex}:                 "a.dp[0]",
		{Prev: lang.ModDeriv, Cur: lang.ModSwizzleScalar}:         "a.dp.x",
		{Prev: lang.ModDeriv, Cur: lang.ModSwizzleVec}:            "a.dp.xy",
		{Prev: lang.ModDeriv, Cur: lang.ModDeriv}:                 "a.dp.dp",
		{Prev: lang.ModBox, Cur: lang.ModIndex}:                   "a.box[0]",
		{Prev: lang.ModBox, Cur: lang.ModDeriv}:                   "a.box.dp",
		{Prev: lang.ModSecond, Cur: lang.ModSurfaceFrame}:         "both.second.u",
		{Prev: lang.ModSecond, Cur: lang.ModDeriv}:                "both.second.dp",
		{Prev: lang.ModSurfaceFrame, Cur: lang.ModDeriv}:          "both.u.dp",
	}

	transitions := lang.AllowedModifierTransitions()
	for prev, curs := range transitions {
		for _, cur := range curs {
			tr := lang.ModifierTransition{Prev: prev, Cur: cur}
			expr, ok := samples[tr]
			if !ok {
				t.Errorf("missing integration test sample for allowed transition %s -> %s", prev, cur)
				continue
			}
			t.Run(fmt.Sprintf("%s_to_%s", prev, cur), func(t *testing.T) {
				src := prelude + "result = " + expr + ";\nreturn result;\n"
				if _, err := assembler.Assemble([]string{src}); err != nil {
					t.Fatalf("Assemble(%q) failed: %v", expr, err)
				}
			})
		}
	}

	for tr := range samples {
		if !lang.IsAllowedModifierTransition(tr.Prev, tr.Cur) {
			t.Errorf("sample %s -> %s is no longer in lang.AllowedModifierTransitions()", tr.Prev, tr.Cur)
		}
	}
}
