package gsleval

import (
	"context"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// paramBallDoc: a ball whose radius is a slider, near a fixed wall. At the
// default radius there is comfortable clearance; at the slider maximum the
// ball overlaps the wall.
const paramBallDoc = `
rad = 3.0;
rad.min = 1.0;
rad.max = 6.5;
ball = length(p) - rad;
wall = length(p - [8.0, 0.0, 0.0]) - 2.0;
return ball;
`

var paramRegion = [6]float64{-8, -8, -8, 12, 8, 8}

// TestCertifyParamsRefusesWhatPointCertifyProves is the demo's discrimination
// pair: at the default radius the clearance claim proves, but over the
// slider's full range it must NOT prove — rad = 6.5 overlaps the wall.
func TestCertifyParamsRefusesWhatPointCertifyProves(t *testing.T) {
	e := New(&source.MockManager{}, nil)
	point, err := e.Certify(context.Background(), nil, paramBallDoc, "", "max(ball, wall)", paramRegion, 0.25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !point.Proven {
		t.Fatalf("point certify at the default radius should prove clearance: %+v", point)
	}

	ranged, err := e.CertifyOverSliders(context.Background(), nil, paramBallDoc, "", "max(ball, wall)", paramRegion, 0.25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if ranged.Proven {
		t.Fatalf("parametric certify proved a claim that fails at rad=6.5: %+v", ranged)
	}
	if ranged.Params["rad"] != [2]float64{1.0, 6.5} {
		t.Errorf("params = %v, want rad over its declared range", ranged.Params)
	}
	// With parameter subdivision the refusal sharpens to a DISPROOF carrying
	// a counterexample in parameter space: the clearance claim fails once
	// the gap drops below 0.5, i.e. for radii above 5.5.
	if !ranged.Disproven {
		t.Fatalf("expected a parametric disproof with a counterexample, got %+v", ranged)
	}
	cp, ok := ranged.CounterParams["rad"]
	if !ok {
		t.Fatalf("no CounterParams for rad: %+v", ranged)
	}
	if cp[0] < 5.0 {
		t.Errorf("counterexample rad range %v should sit in the high end (> 5)", cp)
	}
}

// TestCertifyParamsProvesOverWholeRange: with the slider capped where
// clearance survives, the parametric proof closes — sound for EVERY radius.
func TestCertifyParamsProvesOverWholeRange(t *testing.T) {
	e := New(&source.MockManager{}, nil)
	res, err := e.CertifyParams(context.Background(), nil, paramBallDoc, "", "max(ball, wall)",
		paramRegion, 0.25, 0, map[string][2]float64{"rad": {1.0, 5.0}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Proven {
		t.Fatalf("clearance holds for every rad in [1, 5] and must prove: %+v", res)
	}
}

// TestCertifyParamsRejectsNonConstant: a parameter whose declaration is not
// constant cannot become a tape var, and a certificate over a range the tape
// never saw would be unsound — refuse loudly. (A folded constant like
// `a = 1.0 + 2.0` is fine: the tape still sees a var at its value.)
func TestCertifyParamsRejectsNonConstant(t *testing.T) {
	e := New(&source.MockManager{}, nil)
	doc := "a = p.x + 1.0;\nball = length(p) - 2.0 + 0.0 * a;\nreturn ball;\n"
	_, err := e.CertifyParams(context.Background(), nil, doc, "", "ball",
		[6]float64{-5, -5, -5, 5, 5, 5}, 0, 0, map[string][2]float64{"a": {1, 4}})
	if err == nil {
		t.Fatal("expected an error for a non-constant parameter")
	}
	_, err = e.CertifyParams(context.Background(), nil, doc, "", "ball",
		[6]float64{-5, -5, -5, 5, 5, 5}, 0, 0, map[string][2]float64{"missing": {1, 4}})
	if err == nil {
		t.Fatal("expected an error for an unknown parameter name")
	}
}
