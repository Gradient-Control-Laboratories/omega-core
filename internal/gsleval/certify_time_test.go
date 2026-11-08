package gsleval

import (
	"context"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// timeBallDoc: a ball translating with time toward a fixed wall. Early in the
// motion there is comfortable clearance; late in the motion the ball reaches
// the wall — so a claim over time proves or disproves depending on how much
// of the motion the range covers.
const timeBallDoc = `
ball = length(p - [2.0 * time, 0.0, 0.0]) - 1.0;
wall = length(p - [8.0, 0.0, 0.0]) - 2.0;
return ball;
`

var timeRegion = [6]float64{-4, -4, -4, 12, 4, 4}

// TestCertifyOverTimeMotionEnvelope is the motion-envelope discrimination
// trio: frozen at time 0 the clearance claim proves; over time in [0, 1] it
// still proves — sound at EVERY instant, not sampled frames; over [0, 3] it
// must disprove with the failing time span as the counterexample (surfaces
// come within 0.5 once time passes 2.25).
func TestCertifyOverTimeMotionEnvelope(t *testing.T) {
	e := New(&source.MockManager{}, nil)
	point, err := e.Certify(context.Background(), nil, timeBallDoc, "", "max(ball, wall)", timeRegion, 0.25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !point.Proven {
		t.Fatalf("point certify at time 0 should prove clearance: %+v", point)
	}

	early, err := e.CertifyParams(context.Background(), nil, timeBallDoc, "", "max(ball, wall)",
		timeRegion, 0.25, 0, map[string][2]float64{"time": {0, 1}})
	if err != nil {
		t.Fatal(err)
	}
	if !early.Proven {
		t.Fatalf("clearance holds at every instant of [0, 1] and must prove: %+v", early)
	}
	if early.Params["time"] != [2]float64{0, 1} {
		t.Errorf("params = %v, want the time range echoed", early.Params)
	}

	full, err := e.CertifyParams(context.Background(), nil, timeBallDoc, "", "max(ball, wall)",
		timeRegion, 0.25, 0, map[string][2]float64{"time": {0, 3}})
	if err != nil {
		t.Fatal(err)
	}
	if full.Proven {
		t.Fatalf("certify proved a claim that fails late in the motion: %+v", full)
	}
	if !full.Disproven {
		t.Fatalf("expected a disproof naming the failing time span, got %+v", full)
	}
	ct, ok := full.CounterParams["time"]
	if !ok {
		t.Fatalf("no CounterParams for time: %+v", full)
	}
	if ct[0] < 2.0 {
		t.Errorf("counterexample time span %v should sit late in the motion (> 2)", ct)
	}
}

// TestCertifyTimeComposesWithSliders: time and a document slider range
// together — one proof over the whole motion AND every configuration.
func TestCertifyTimeComposesWithSliders(t *testing.T) {
	doc := `
rad = 1.0;
rad.min = 0.5;
rad.max = 1.5;
ball = length(p - [2.0 * time, 0.0, 0.0]) - rad;
wall = length(p - [8.0, 0.0, 0.0]) - 2.0;
return ball;
`
	e := New(&source.MockManager{}, nil)
	res, err := e.CertifyParams(context.Background(), nil, doc, "", "max(ball, wall)",
		timeRegion, 0.25, 0, map[string][2]float64{"time": {0, 1}, "rad": {0.5, 1.5}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Proven {
		t.Fatalf("clearance holds for every (time, rad) and must prove: %+v", res)
	}
	if len(res.Params) != 2 {
		t.Errorf("params = %v, want both time and rad", res.Params)
	}

	late, err := e.CertifyParams(context.Background(), nil, doc, "", "max(ball, wall)",
		timeRegion, 0.25, 0, map[string][2]float64{"time": {0, 3}, "rad": {0.5, 1.5}})
	if err != nil {
		t.Fatal(err)
	}
	if !late.Disproven {
		t.Fatalf("expected a disproof over the longer motion, got %+v", late)
	}
	if _, ok := late.CounterParams["time"]; !ok {
		t.Fatalf("counterexample should name the failing time span: %+v", late)
	}
}
