//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
	"github.com/Gradient-Control-Laboratories/omega-core/backend/tape"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/fidget"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
	"github.com/Gradient-Control-Laboratories/omega-core/program/noise/spinodoid"
)

// The spinodoid oracles, end to end: parse, bake, fidget.
//
// These check that each builtin's oracle is its own table: the fast table for
// fastGaborSpinodoid, the reference table for gaborSpinodoid, and that a tape
// names which of the two each leaf is.
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
	return New(source.NewLocalManager("../stdlib"), nil)
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

// TestFastGaborLibraryOracleMatchesFastTable checks the field fidget
// evaluates for spinodoid.fastGabor against the Go fast table.
func TestFastGaborLibraryOracleMatchesFastTable(t *testing.T) {
	p := referenceParams()
	level, err := spinodoid.Level(0.5)
	if err != nil {
		t.Fatalf("Level: %v", err)
	}
	table, err := spinodoid.FastGaborBake(p)
	if err != nil {
		t.Fatalf("Bake: %v", err)
	}

	points := oraclePoints()
	got := evalField(t, fastGaborLibraryDoc, "field", points)

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
	table, err := spinodoid.FastGaborBake(p)
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

// TestSpinodoidBuiltinsAreTwoOracles guards the dispatch: with the same
// arguments, spinodoid.gabor is the reference table's field and
// fastGaborSpinodoid the fast table's, and the two differ.
func TestSpinodoidBuiltinsAreTwoOracles(t *testing.T) {
	level, err := spinodoid.Level(0.5)
	if err != nil {
		t.Fatalf("Level: %v", err)
	}
	table, err := spinodoid.ReferenceGaborBake(referenceParams())
	if err != nil {
		t.Fatalf("ReferenceGaborBake: %v", err)
	}
	points := oraclePoints()
	gabor := evalField(t, gaborSpinodoidDoc, "field", points)
	fast := evalField(t, fastSpinodoidDoc, "field", points)

	same := 0
	for i, q := range points {
		want := table.Distance(q, level, true)
		if math.Abs(gabor[i]-want) > oracleTolerance {
			t.Errorf("point %v: gaborSpinodoid %v, reference table %v", q, gabor[i], want)
		}
		if math.Abs(gabor[i]-fast[i]) <= oracleTolerance {
			same++
		}
	}
	if same == len(points) {
		t.Errorf("gaborSpinodoid and fastGaborSpinodoid agree at every point: one table serves both")
	}
}

// TestSpinodoidTapeLeavesNameTheirOracle checks what a tape tells its reader:
// a fastGabor leaf is the fast oracle, in the fast block and marked fast; a
// gabor leaf is the reference, in the reference block and not marked fast.
func TestSpinodoidTapeLeavesNameTheirOracle(t *testing.T) {
	for _, c := range []struct {
		name, doc string
		fast      bool
		id        uint32
	}{
		{"fastGabor", fastGaborLibraryDoc, true, fidget.FastGaborSpinodoidOracleID(0, fidget.SpinodoidDistance)},
		{"gabor", gaborSpinodoidDoc, false, fidget.ReferenceGaborSpinodoidOracleID(0, fidget.SpinodoidDistance)},
	} {
		res, err := assembler.Assemble([]string{c.doc},
			assembler.SourceManager(source.NewLocalManager("../stdlib")))
		if err != nil {
			t.Fatalf("%s: Assemble: %v", c.name, err)
		}
		_, _, oracles, err := tape.GenerateTapeWithSpinodoids(res.Program)
		if err != nil {
			t.Fatalf("%s: GenerateTapeWithSpinodoids: %v", c.name, err)
		}
		if len(oracles) != 1 {
			t.Fatalf("%s: want one spinodoid oracle, got %d", c.name, len(oracles))
		}
		o := oracles[0]
		if o.Fast != c.fast {
			t.Errorf("spinodoid.%s's tape leaf has fast %v, want %v", c.name, o.Fast, c.fast)
		}
		if o.AssetID != c.id {
			t.Errorf("spinodoid.%s's leaf id %#x, want %#x", c.name, o.AssetID, c.id)
		}
	}
}

// The library forms the builtins are checked against: the reference and the
// fast one.
const gaborSpinodoidDoc = "import \"lattice/spinodoid\";\nfield = spinodoid.gabor(6, [pi / 2, pi / 2, pi / 2]);\nreturn field;"
const fastGaborLibraryDoc = "import \"lattice/spinodoid\";\nfield = spinodoid.fastGabor(6, [pi / 2, pi / 2, pi / 2]);\nreturn field;"

// The fastGaborSpinodoid builtin spelled out: spinodoid.fastGabor's arguments,
// with the Cartesian poles and an identity frame.
const fastSpinodoidDoc = `poles = [[1, 0, 0], [0, 1, 0], [0, 0, 1]];
field = fastGaborSpinodoid(p, 6, [pi / 2, pi / 2, pi / 2], poles, 3, 8, 4096, 0, 0, 1, 1, poles, [[0, 0, 0], [0, 0, 0]]);
return field;`
