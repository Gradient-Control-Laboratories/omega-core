package gsleval_test

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"gradientcontrol.com/omega/core/internal/gsleval"
	"gradientcontrol.com/omega/core/internal/source"
)

// A property of a deeply remapped field — the gradient, the second-nearest
// surface — compiles in about the time the field itself does, and the gradient
// is the field's own slope.
//
// Each remap level used to multiply what the compiler walked by about five:
// the property engine's result was a small shared graph that every pass after
// it copied out as a tree. Eight nested rotations never finished, and neither
// did the gradient of a 24-tooth gear profile, whose library body nests five
// remaps — the query an Astra run hung on for four hours. The gradient is
// checked against central differences of the field, so a faster answer that
// is wrong does not pass.
func TestAPropertyOfADeeplyRemappedFieldCompilesAndIsRight(t *testing.T) {
	var nested strings.Builder
	nested.WriteString("f0() { return max(abs(p.x) - 1.0, abs(p.y) - 2.0, abs(p.z) - 3.0); }\n")
	for i := 1; i <= 8; i++ {
		fmt.Fprintf(&nested, "f%d() { c = cos(0.3); s = sin(0.3); return remap(f%d(), [p.x * c + p.y * s, p.y * c - p.x * s, p.z]); }\n", i, i-1)
	}
	nested.WriteString("return f8();\n")

	cases := []struct{ name, doc, field string }{
		{"eight nested rotations", nested.String(), "f8()"},
		{"a 24-tooth gear profile", "g() { return max(gears.gearProfile(24, 0.65, radians(20), 0.0, 0.25, 0.02), abs(p.z) - 2.0); }\nreturn g();\n", "g()"},
	}
	e := gsleval.New(source.NewLocalManager("testpkg", "../../stdlib"), nil)
	pts := [][3]float64{{0.5, 0.5, 0.5}, {7.9, 1.3, 0.4}, {-3.1, 6.2, -0.7}}
	const h = 1e-3

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defs := "gslWallD = " + c.field + ";"
			type answer struct {
				grad [][3]float64
				err  error
			}
			done := make(chan answer, 1)
			start := time.Now()
			go func() {
				groups, err := e.EvalGsl(context.Background(), nil, c.doc, defs, []string{"gslWallD.dp", "gslWallD.second"}, pts)
				if err != nil {
					done <- answer{err: err}
					return
				}
				// A vec3 comes back one sample list per component: [component][point].
				var a answer
				for i := range pts {
					a.grad = append(a.grad, [3]float64{
						float64(groups[0].Samples[0][i].Value),
						float64(groups[0].Samples[1][i].Value),
						float64(groups[0].Samples[2][i].Value)})
				}
				done <- a
			}()
			var a answer
			select {
			case a = <-done:
			case <-time.After(60 * time.Second):
				t.Fatalf("the properties of %s did not compile in 60 s", c.name)
			}
			if a.err != nil {
				t.Fatalf("eval: %v", a.err)
			}
			t.Logf("%s: .dp and .second in %v", c.name, time.Since(start).Round(time.Millisecond))

			// Central differences of the plain field at the same points.
			var probe [][3]float64
			for _, p := range pts {
				for k := 0; k < 3; k++ {
					plus, minus := p, p
					plus[k] += h
					minus[k] -= h
					probe = append(probe, plus, minus)
				}
			}
			groups, err := e.EvalGsl(context.Background(), nil, c.doc, defs, []string{"gslWallD"}, probe)
			if err != nil {
				t.Fatalf("eval field: %v", err)
			}
			s := groups[0].Samples[0]
			for i := range pts {
				for k := 0; k < 3; k++ {
					fd := (float64(s[i*6+2*k].Value) - float64(s[i*6+2*k+1].Value)) / (2 * h)
					if math.Abs(fd-a.grad[i][k]) > 2e-2 {
						t.Errorf("%s at %v: d/d%c = %.5f, central difference %.5f", c.name, pts[i], "xyz"[k], a.grad[i][k], fd)
					}
				}
			}
		})
	}
}
