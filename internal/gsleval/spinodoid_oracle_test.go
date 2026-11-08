package gsleval

import (
	"context"
	"math"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/fidget/lib"
	"gradientcontrol.com/omega/core/internal/noise/spinodoid"
	"gradientcontrol.com/omega/core/internal/source"
)

// The spinodoid oracles, end to end: parse, bake, fidget.
//
// These check that the oracle fidget evaluates is the lattice table that was
// baked, for both builtins, and that a tape marks the leaf as the lattice, so
// a reader that wants hodge's field knows to swap it in.
//
// Samples come back as float32, so the tolerance is f32 rounding on a field
// whose values are a few length units.
const oracleTolerance = 1e-4

// The arguments spinodoid.gabor(period, theta) passes, and the reference
// table for them.
func referenceParams() spinodoid.Params {
	return spinodoid.Params{
		Theta:       []float64{math.Pi / 2, math.Pi / 2, math.Pi / 2},
		Period:      6,
		PacketWidth: 3,
		Impulses:    8,
		Pool:        4096,
		Seed:        0,
	}
}

func oraclePoints() [][3]float64 {
	return [][3]float64{
		{0, 0, 0},
		{2.5, -3.75, 6.125},
		{-18.5, 11.25, 4},
		{31.5, 27.5, -35.25},
		{-40, -7.5, 22.75},
	}
}

func newSpinodoidEvaluator() *Evaluator {
	return New(source.NewLocalManager("../../stdlib"), nil)
}

// evalField evaluates one scalar GSL expression at points and returns its
// values.
func evalField(t *testing.T, code, expr string, points [][3]float64) []float64 {
	t.Helper()
	e := newSpinodoidEvaluator()
	groups, err := e.EvalGsl(context.Background(), nil, code, "", []string{expr}, points)
	if err != nil {
		t.Fatalf("EvalGsl: %v", err)
	}
	if len(groups) != 1 || len(groups[0].Samples) != 1 {
		t.Fatalf("want one scalar group, got %+v", groups)
	}
	out := make([]float64, len(points))
	for i := range points {
		out[i] = float64(groups[0].Samples[0][i].Value)
	}
	return out
}

// TestGaborSpinodoidOracleMatchesLatticeTable checks the field fidget
// evaluates for spinodoid.gabor against the Go reference for the lattice
// table, the field the preview renders.
func TestGaborSpinodoidOracleMatchesLatticeTable(t *testing.T) {
	p := referenceParams()
	level, err := spinodoid.Level(0.5)
	if err != nil {
		t.Fatalf("Level: %v", err)
	}
	table, err := spinodoid.Bake(p)
	if err != nil {
		t.Fatalf("Bake: %v", err)
	}

	points := oraclePoints()
	got := evalField(t, gaborSpinodoidDoc, "field", points)

	worst := 0.0
	for i, q := range points {
		want, _ := table.Distance(q, level, true)
		if e := math.Abs(got[i] - want); e > worst {
			worst = e
		}
		if math.Abs(got[i]-want) > oracleTolerance {
			t.Errorf("point %v: oracle %v, lattice table %v", q, got[i], want)
		}
	}
	t.Logf("%d points, worst difference %g", len(points), worst)
}

// TestFastGaborSpinodoidOracleMatchesLatticeTable is the same for the
// fastGaborSpinodoid builtin, called directly.
func TestFastGaborSpinodoidOracleMatchesLatticeTable(t *testing.T) {
	p := referenceParams()
	level, err := spinodoid.Level(0.5)
	if err != nil {
		t.Fatalf("Level: %v", err)
	}
	table, err := spinodoid.Bake(p)
	if err != nil {
		t.Fatalf("Bake: %v", err)
	}

	points := oraclePoints()
	got := evalField(t, fastSpinodoidDoc, "field", points)

	worst := 0.0
	for i, q := range points {
		want, _ := table.Distance(q, level, true)
		if e := math.Abs(got[i] - want); e > worst {
			worst = e
		}
		if math.Abs(got[i]-want) > oracleTolerance {
			t.Errorf("point %v: oracle %v, lattice table %v", q, got[i], want)
		}
	}
	t.Logf("%d points, worst difference %g", len(points), worst)
}

// TestSpinodoidBuiltinsAreOneField guards the dispatch: gaborSpinodoid and
// fastGaborSpinodoid read the same lattice table, so with the same arguments
// they agree at every point, bit for bit.
func TestSpinodoidBuiltinsAreOneField(t *testing.T) {
	points := oraclePoints()
	gabor := evalField(t, gaborSpinodoidDoc, "field", points)
	fast := evalField(t, fastSpinodoidDoc, "field", points)

	for i, q := range points {
		if gabor[i] != fast[i] {
			t.Errorf("point %v: gaborSpinodoid %v, fastGaborSpinodoid %v", q, gabor[i], fast[i])
		}
	}
}

// TestGaborSpinodoidTapeLeafIsFast checks what a tape tells its reader: a
// spinodoid.gabor leaf is the lattice field, in the lattice's oracle-id block
// and marked fast, so a reader that wants hodge's field knows to swap it in.
func TestGaborSpinodoidTapeLeafIsFast(t *testing.T) {
	res, err := assembler.Assemble([]string{gaborSpinodoidDoc},
		source.NewLocalManager("../../stdlib"))
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	_, _, oracles, err := lib.GenerateTapeWithSpinodoids(res.Program)
	if err != nil {
		t.Fatalf("GenerateTapeWithSpinodoids: %v", err)
	}
	if len(oracles) != 1 {
		t.Fatalf("want one spinodoid oracle, got %d", len(oracles))
	}
	o := oracles[0]
	if !o.Fast {
		t.Errorf("spinodoid.gabor's tape leaf is not marked fast")
	}
	if want := lib.SpinodoidOracleID(0, lib.SpinodoidDistance); o.AssetID != want {
		t.Errorf("leaf id %#x, want the lattice block's %#x", o.AssetID, want)
	}
}

// The library form both builtins are checked against.
const gaborSpinodoidDoc = "import \"lattice/spinodoid\";\nfield = spinodoid.gabor(6, [pi / 2, pi / 2, pi / 2]);\nreturn field;"

// The fastGaborSpinodoid builtin spelled out: spinodoid.gabor's arguments,
// with the Cartesian poles and an identity frame.
const fastSpinodoidDoc = `poles = [[1, 0, 0], [0, 1, 0], [0, 0, 1]];
field = fastGaborSpinodoid(p, 6, [pi / 2, pi / 2, pi / 2], poles, 3, 8, 4096, 0, 0, 1, 1, poles);
return field;`
