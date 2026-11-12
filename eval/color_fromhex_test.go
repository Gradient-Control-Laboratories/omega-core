//go:build cgo

package eval_test

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/source"
)

// color.fromHex reads six hex digits, with or without '#', in either case, as the sRGB vec3 a color takes.
func TestColorFromHex(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	queries := []string{`color.fromHex("#8fa3b8")`, `color.fromHex("8FA3B8")`, `color.fromHex("#000000")`,
		`color.fromHex("#ffffff")`}
	want := [][3]float64{{0x8f, 0xa3, 0xb8}, {0x8f, 0xa3, 0xb8}, {0, 0, 0}, {255, 255, 255}}
	groups, err := e.EvalGsl(t.Context(), nil, "", "", queries, [][3]float64{{0, 0, 0}})
	if err != nil {
		t.Fatal(err)
	}
	for i, group := range groups {
		if len(group.Samples) != 3 {
			t.Fatalf("%s: %d components, want 3", queries[i], len(group.Samples))
		}
		for k := 0; k < 3; k++ {
			got := float64(group.Samples[k][0].Value)
			if math.Abs(got-want[i][k]/255) > 1e-6 {
				t.Errorf("%s component %d = %g, want %g", queries[i], k, got, want[i][k]/255)
			}
		}
	}
}
