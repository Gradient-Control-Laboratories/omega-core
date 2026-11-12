//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// colorEval evaluates colour expressions, each a vec3, at one point, a
// component per query. Every value here is a constant, so the point is only
// there because the evaluator wants one.
func colorEval(t *testing.T, exprs []string) [][3]float64 {
	t.Helper()
	mgr := source.NewLocalManager("../stdlib")
	e := New(mgr, nil)
	queries := make([]string, 0, 3*len(exprs))
	for _, expr := range exprs {
		for _, c := range []string{".x", ".y", ".z"} {
			queries = append(queries, "("+expr+")"+c)
		}
	}
	groups, err := e.EvalGsl(context.Background(), nil, "return p.x;", "", queries, [][3]float64{{0, 0, 0}})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	out := make([][3]float64, len(exprs))
	for i := range exprs {
		for c := 0; c < 3; c++ {
			out[i][c] = float64(groups[3*i+c].Samples[0][0].Value)
		}
	}
	return out
}

// A thin film's colour has four properties worth pinning: a film of zero
// thickness in air reflects nothing; a film of zero thickness on a substrate
// is the substrate's grey, with no tint; and over a metal the film runs
// through the first-order colours at the thicknesses the literature gives
// for steel, yellow near 46 nm and blue near 72 nm, while zero thickness
// gives the metal's own colour back.
func TestThinFilmColours(t *testing.T) {
	steel := "[0.7548, 0.7439, 0.7314]"
	got := colorEval(t, []string{
		"color.thinFilm(0, 1.33, 1)",                  // 0: a soap film of no thickness, in air
		"color.thinFilm(0, 2.42, 13)",                 // 1: bare substrate, a grey
		"color.thinFilm(46, 2.42, 13)",                // 2: first-order yellow
		"color.thinFilm(72, 2.42, 13)",                // 3: first-order blue
		"color.thinFilmOver(0, 2.42, " + steel + ")",  // 4: the metal itself
		"color.thinFilmOver(46, 2.42, " + steel + ")", // 5: straw-tempered steel
		"color.thinFilmOver(72, 2.42, " + steel + ")", // 6: blue-tempered steel
	})
	for c := 0; c < 3; c++ {
		if got[0][c] > 1e-6 {
			t.Errorf("a soap film of no thickness reflects %v, want black", got[0])
		}
	}
	if math.Abs(got[1][0]-got[1][1]) > 1e-4 || math.Abs(got[1][1]-got[1][2]) > 1e-4 || got[1][0] < 0.5 {
		t.Errorf("the bare substrate is %v, want a neutral grey above 0.5", got[1])
	}
	if !(got[2][0] > got[2][2] && got[2][1] > got[2][2]) {
		t.Errorf("46 nm on a metal is %v, want yellow (red and green above blue)", got[2])
	}
	if !(got[3][2] > got[3][0]) {
		t.Errorf("72 nm on a metal is %v, want blue above red", got[3])
	}
	want := [3]float64{0.7548, 0.7439, 0.7314}
	for c := 0; c < 3; c++ {
		if math.Abs(got[4][c]-want[c]) > 1e-3 {
			t.Errorf("no film over steel gives %v, want steel's own %v", got[4], want)
		}
	}
	if !(got[5][0] > got[5][2] && got[5][1] > got[5][2]) {
		t.Errorf("straw-tempered steel is %v, want yellow", got[5])
	}
	if !(got[6][2] > got[6][0]) {
		t.Errorf("blue-tempered steel is %v, want blue", got[6])
	}
}
