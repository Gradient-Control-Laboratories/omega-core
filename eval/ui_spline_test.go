//go:build cgo

package eval

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/source"
)

func TestSplinePreservesNativeField(t *testing.T) {
	code := `m = [[1,0,0,0],[0,1,0,0],[0,0,1,0],[3,4,5,1]];
path = uiSpline([m, m, m]);
last = path[2];
position = last[3];
return length(p - position.xyz) - 2;`
	for _, src := range []string{strings.Replace(code, "uiSpline([m, m, m])", "[m, m, m]", 1), code} {
		res, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
		if err != nil {
			t.Fatal(err)
		}
		got, err := EvalGrad(res.Program, []float32{3, 6, 3}, []float32{4, 4, 4}, []float32{5, 5, 9})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || len(got[0]) != 3 {
			t.Fatalf("unexpected samples: %v", got)
		}
		for i, want := range []float32{-2, 1, 2} {
			if got[0][i].Value != want {
				t.Fatalf("point %d: got %g, want %g", i, got[0][i].Value, want)
			}
		}
	}
}
