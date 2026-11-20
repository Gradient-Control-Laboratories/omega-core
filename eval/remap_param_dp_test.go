//go:build cgo

package eval

import (
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/source"
)

// .dp of a remap agrees with fidget's own gradient whether the remapped field is a declaration,
// a thunk called inside the remap, or a parameter of the function that remaps it. The parameter
// was bound to its argument's value at the caller's p, so its remap took that value's derivative
// unmoved: the value was right and .dp lost the chain rule through the remapped point.
func TestRemapOfAParameterDifferentiatesThroughTheRemap(t *testing.T) {
	sources := map[string]string{
		"declaration":   "a = p.z - 1;\nu = remap(a, p + [0, 0, p.x]);\n",
		"thunk":         "a = () { return p.z - 1; };\nu = remap(a(), p + [0, 0, p.x]);\n",
		"parameter":     "f = (v) { return remap(v, p + [0, 0, p.x]); };\nu = f(p.z - 1);\n",
		"nested call":   "f = (v) { return remap(v, p + [0, 0, p.x]); };\ng = (w) { return f(w * 2); };\nu = g(p.z - 1);\n",
		"under a remap": "f = (v) { return remap(v, p + [0, 0, p.x]); };\nu = remap(f(p.z - 1), p + [p.y, 0, 0]);\n",
	}
	xs, ys, zs := []float32{0.5, -1, 2}, []float32{0, 0.7, -0.3}, []float32{0.9, 2, -1}
	for name, body := range sources {
		src := "import . \"omega\";\n" + body + "return u, u.dp.x, u.dp.y, u.dp.z;\n"
		res, err := assembler.Assemble([]string{src}, assembler.SourceManager(source.NewLocalManager("../stdlib")))
		if err != nil {
			t.Fatalf("%s: %v\n%s", name, err, src)
		}
		out, err := EvalGrad(res.Program, xs, ys, zs)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for i := range xs {
			got := [3]float32{out[1][i].Value, out[2][i].Value, out[3][i].Value}
			want := [3]float32{out[0][i].Dx, out[0][i].Dy, out[0][i].Dz}
			for k := range 3 {
				if d := got[k] - want[k]; d > 1e-5 || d < -1e-5 {
					t.Errorf("%s at (%v, %v, %v) axis %d: .dp %v, fidget %v", name, xs[i], ys[i], zs[i], k, got[k], want[k])
				}
			}
		}
	}
}
