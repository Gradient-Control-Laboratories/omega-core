//go:build cgo

package eval

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/source"
)

// A map over a list of fields whose function applies .dp to its element, directly or through a
// function it calls, gives what the same calls written out give. The property pass, which expands
// .dp, runs before map is unrolled, and a parameter's derivative needs the argument its call site
// passes; a map's function is called by the map. Such a fold is unrolled first, its lambda renamed
// apart (every lambda is gsl_anon, and a nested one would be taken for a call to itself), given what
// it reads of the function it is written in as parameters, and moved to the file's own functions.
func TestMapWhoseFunctionDifferentiatesIsTheCallsWrittenOut(t *testing.T) {
	const pre = "import . \"omega\";\na = length(p - [0.5, 0, 0]) - 1;\nb = p.z;\nc = p.x - 1.5;\nfs = [a, b, c];\n"
	cases := []struct{ name, mapped, written string }{
		{
			"one map, .dp in a called function",
			"g(f) { return f.dp.x; }\nu = intersectSharp(map((x) { return g(x); }, [a, b]));\n",
			"u = max(a.dp.x, b.dp.x);\n",
		},
		{
			"one map, .dp in the lambda, with its index",
			"u = intersectSharp(map((x, i) { return x.dp.y * (i + 1); }, fs));\n",
			"u = max(a.dp.y, b.dp.y * 2, c.dp.y * 3);\n",
		},
		{
			"every pair, nested maps",
			"u = intersectSharp(map((x, i) { return intersectSharp(map((y, j) { return j > i ? edgeDistance(x, y, 3) : x; }, fs)); }, fs));\n",
			"u = max(edgeDistance(a, b, 3), edgeDistance(a, c, 3), edgeDistance(b, c, 3));\n",
		},
		{
			"every three, three nested maps",
			"u = intersectSharp(map((x, i) { return intersectSharp(map((y, j) { return intersectSharp(map((z, k) { return k > j && j > i ? cornerDistance(x, y, z, 3) : x; }, fs)); }, fs)); }, fs));\n",
			"u = cornerDistance(a, b, c, 3);\n",
		},
	}
	xs, ys, zs := []float32{1.2, 2, 2.3, -0.4}, []float32{0.3, 0.1, 0.4, 1.1}, []float32{0.7, -0.4, 0.5, 0.2}
	eval := func(t *testing.T, body string) [][4]float64 {
		src := pre + body + "return u;\n"
		res, err := assembler.Assemble([]string{src}, assembler.SourceManager(source.NewLocalManager("../stdlib")))
		if err != nil {
			t.Fatalf("%v\n%s", err, src)
		}
		out, err := EvalGrad(res.Program, xs, ys, zs)
		if err != nil {
			t.Fatalf("%v", err)
		}
		vals := make([][4]float64, len(xs))
		for i := range xs {
			o := out[0][i]
			vals[i] = [4]float64{float64(o.Value), float64(o.Dx), float64(o.Dy), float64(o.Dz)}
		}
		return vals
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, want := eval(t, c.mapped), eval(t, c.written)
			for i := range got {
				for k := range 4 {
					if d := math.Abs(got[i][k] - want[i][k]); !(d <= 1e-6) {
						t.Errorf("at point %d, component %d: mapped %v, written out %v", i, k, got[i][k], want[i][k])
					}
				}
			}
		})
	}
}
