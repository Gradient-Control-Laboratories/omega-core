package gsleval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// TestCatalogProductStandards checks the product-standard tables against the
// standards themselves — every value here is one you can look up — and the
// Torx table against ISO 10664:2014, which the values it used to carry were
// wrong about from T27 up.
func TestCatalogProductStandards(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	const doc = `
x = 0.0;
return sphere(1.0);
`
	near := func(expr string, want float64) {
		t.Helper()
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{expr}, [][3]float64{{0, 0, 0}})
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		got := float64(groups[0].Samples[0][0].Value)
		if math.IsNaN(got) || math.Abs(got-want) > 1e-4 {
			t.Errorf("%s = %.5f, want %.5f", expr, got, want)
		}
	}

	// ISO 7045 M6 pan head: dk 12, k 4.6, rf 10, a PH3 of m 6.9 penetrating
	// 3.6, a PZ3 of m 6.7 penetrating 3.46. And an M2 takes a PH0 — the one
	// place a zero in these tables is a size rather than an absence.
	near("catalog.panHeadDia(6.0)", 12.0)
	near("catalog.panHeadHeight(6.0)", 4.6)
	near("catalog.panCrown(6.0)", 10.0)
	near("catalog.panRecess(6.0)", 3.0)
	near("catalog.panPhillipsDia(6.0)", 6.9)
	near("catalog.panPhillipsDepth(6.0)", 3.6)
	near("catalog.panPozidrivDia(6.0)", 6.7)
	near("catalog.panPozidrivDepth(6.0)", 3.46)
	near("catalog.panRecess(2.0)", 0.0)
	near("catalog.panPhillipsDia(2.0)", 1.9)
	// No ISO 7045 screw in M12: nothing to build.
	near("catalog.panHeadDia(12.0)", 0.0)

	// ISO 7380-1 M6 button head, ISO 4017 M10 hex head.
	near("catalog.buttonHeadDia(6.0)", 10.5)
	near("catalog.buttonSocket(6.0)", 4.0)
	near("catalog.buttonSocketDepth(6.0)", 2.08)
	near("catalog.hexHeadAcrossFlats(10.0)", 16.0)
	near("catalog.hexHeadHeight(10.0)", 6.4)

	// The Torx pairings: ISO 14579 puts a T55 in an M12 cap screw, ISO 14583 a
	// T15 in an M3.5 pan head, and they agree on T20 at M4 but not on its depth.
	near("catalog.shcsTorx(12.0)", 55.0)
	near("catalog.panTorx(3.5)", 15.0)
	near("catalog.shcsTorx(4.0)", 20.0)
	near("catalog.panTorx(4.0)", 20.0)
	near("catalog.shcsTorxDepth(4.0)", 1.80)
	near("catalog.panTorxDepth(4.0)", 1.66)

	// ISO 10664:2014 Table 1. The old table's T27 row was the standard's T30.
	near("catalog.torxMajor(27.0)", 5.1)
	near("catalog.torxMinor(27.0)", 3.68)
	near("catalog.torxMajor(30.0)", 5.6)
	near("catalog.torxMajor(40.0)", 6.75)
	near("catalog.torxMajor(55.0)", 11.35)

	// ISO 4757 recesses by number.
	near("catalog.phillipsCore(3.0)", 2.41)
	near("catalog.phillipsFlank(2.0)", 5.75*math.Pi/180)
	near("catalog.pozidrivWing(3.0)", 1.42)
	near("catalog.pozidrivCore(3.0)", 2.72)
	near("catalog.pozidrivKnee(3.0)", 3.86)
	near("catalog.robertsonAcross(2.0)", 2.84)

	// Where a driver's tip seats in an M4 pan head: the recess seat, 3.1 - 6.5
	// + sqrt(6.5^2 - (m/2)^2), less the gauge penetration.
	near("fastener.panPhillipsFloor(4.0)", 3.1-6.5+math.Sqrt(6.5*6.5-2.2*2.2)-2.4)
	near("fastener.panPozidrivFloor(4.0)", 3.1-6.5+math.Sqrt(6.5*6.5-2.15*2.15)-2.34)

	// ISO 4014's thread length: 2d + 6 up to 125 long, 2d + 12 to 200, and the
	// whole shank when that is shorter.
	near("fastener.hexBoltThread(6.0, 40.0)", 18.0)
	near("fastener.hexBoltThread(6.0, 150.0)", 24.0)
	near("fastener.hexBoltThread(6.0, 10.0)", 10.0)
}

// TestProductTablesUseCatalogSizes holds the product-standard tables to the
// guarantee the metric table gave by construction: a size is a row of metric.
// A key with no pitch is a screw whose thread does not exist.
func TestProductTablesUseCatalogSizes(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	for table, rows := range map[string]float64{
		"pan7045": 10, "button7380": 8, "hex4017": 13, "torx14579": 9, "torx14583": 9,
	} {
		doc := `
missing = reduce((acc, v) { return acc + v; },
                 map((r) { return 1.0 - clamp(sign(catalog.pitch(r[0])), 0.0, 1.0); },
                     catalog.` + table + `));
count = reduce((acc, v) { return acc + v; }, map((r) { return 1.0; }, catalog.` + table + `));
return sphere(1.0);
`
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{"missing", "count"}, [][3]float64{{0, 0, 0}})
		if err != nil {
			t.Fatalf("%s: %v", table, err)
		}
		if missing := float64(groups[0].Samples[0][0].Value); missing != 0 {
			t.Errorf("catalog.%s has %v sizes the metric table does not", table, missing)
		}
		if count := float64(groups[1].Samples[0][0].Value); count != rows {
			t.Errorf("catalog.%s has %v rows, want %v", table, count, rows)
		}
	}
}

// TestStandardScrewsAndDrives probes the new standard screws and recesses at
// points whose side is known from the standards: a recess is open where the
// drive goes and closed where it does not, down to the depth the standard
// gives, and no deeper.
func TestStandardScrewsAndDrives(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	const doc = `
ph = fastener.panPhillips(4.0, 10.0);
pz = fastener.panPozidriv(4.0, 10.0);
flat = drill.robertsonAt(0.0, 2.84, 2.92, 0.0);
pyramid = drill.robertsonAt(0.0, 2.84, 2.92);
hx = fastener.hexScrew(6.0, 12.0);
bt = fastener.button(6.0, 12.0);
st = fastener.shcsTorx(6.0, 12.0);
return ph;
`
	at := func(expr string, pt [3]float64) float64 {
		t.Helper()
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{expr}, [][3]float64{pt})
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		return float64(groups[0].Samples[0][0].Value)
	}
	// A solid screw is negative in metal; a void is negative inside the void.
	metal := func(expr string, pt [3]float64) {
		t.Helper()
		if v := at(expr, pt); v >= 0 {
			t.Errorf("%s at %v = %.4f, want metal (< 0)", expr, pt, v)
		}
	}
	open := func(expr string, pt [3]float64) {
		t.Helper()
		if v := at(expr, pt); v <= 0 {
			t.Errorf("%s at %v = %.4f, want open (> 0)", expr, pt, v)
		}
	}
	on := func(expr string, pt [3]float64) {
		t.Helper()
		if v := at(expr, pt); math.Abs(v) > 1e-3 {
			t.Errorf("%s at %v = %.4f, want on the surface", expr, pt, v)
		}
	}

	// M4 pan head, PH2: m 4.4. The recess seats where the dome crosses its own
	// diameter, 3.1 - 6.5 + sqrt(6.5^2 - 2.2^2) = 2.716. Its 26 degree 30 cone
	// reaches g = 2.29 1.055/tan(26.5) = 2.116 below that, at 0.600, and its 28
	// degree bottom comes to a point t1 = 1.145 tan(28) = 0.609 lower, at -0.0084.
	on("ph", [3]float64{4.0, 0, 0.1})
	open("ph", [3]float64{0, 0, 2.6})
	open("ph", [3]float64{0, 0, 0.4})
	open("ph", [3]float64{0, 0, 0.05})
	metal("ph", [3]float64{0, 0, -0.05})
	on("ph", [3]float64{1.0, 0, -0.0084 + math.Tan(28*math.Pi/180)}) // the bottom, under a wing
	open("ph", [3]float64{1.9, 0, 2.66})                             // along a wing
	metal("ph", [3]float64{1.3435, 1.3435, 2.66})                    // between two, at 45 degrees

	// ISO 4757's b is the CORE, not a wing: the wings are what the ribs leave
	// between them, about 0.7 across at the bottom and under 1.0 at the face.
	// Reading b = 1.47 as a wing width — the old model — would leave this open.
	open("ph", [3]float64{1.9, 0.45, 2.66})
	metal("ph", [3]float64{1.9, 0.6, 2.66})
	metal("ph", [3]float64{1.9, 0.9, 2.66})
	metal("ph", [3]float64{1.9, 1.05, 2.66})

	// M4 pan head, PZ2: m 4.3, seated at 2.734. The 26 degree 30 cone reaches
	// the knee, g = 2.34, 0.98/tan(26.5) = 1.966 below that, at 0.769, and the
	// 28 degree bottom comes to a point 1.17 tan(28) = 0.622 lower, at 0.146.
	open("pz", [3]float64{0, 0, 0.25})
	metal("pz", [3]float64{0, 0, 0.05})
	// The wings are f = 1.03 wide, not b = 1.83 — the old model's reading.
	open("pz", [3]float64{1.9, 0.45, 2.66})
	metal("pz", [3]float64{1.9, 0.6, 2.66})
	// A valley runs out along the diagonal to b/2 = 0.915 at the knee, leaning
	// out 5 degrees 45 going up: 0.958 at z 1.2. Past it is metal.
	open("pz", [3]float64{0.64, 0.64, 1.2})
	metal("pz", [3]float64{0.75, 0.75, 1.2})
	// A mark runs out along the diagonal just under the reference plane, where
	// a Phillips has a rib; beside it, and between a wing and a valley, metal.
	open("pz", [3]float64{1.0607, 1.0607, 2.714})
	metal("pz", [3]float64{1.202, 0.919, 2.714})
	metal("pz", [3]float64{1.452, 1.114, 2.70})

	// Robertson #2: a flat bottom leaves the floor's corner open; the patent's
	// pyramid fills it, and meets itself on the axis at the full depth.
	metal("flat", [3]float64{1.3, 1.3, -2.85}) // inside the void
	open("pyramid", [3]float64{1.3, 1.3, -2.85})
	open("pyramid", [3]float64{1.3, 0, -2.5})
	metal("pyramid", [3]float64{0, 0, -2.85})

	// ISO 4017 M6: flats 10 across, head 4 high.
	on("hx", [3]float64{5.0, 0, 1.0})
	on("hx", [3]float64{0, 0, 4.0})

	// ISO 7380-1 M6: socket 2.08 deep below a 3.3 crown.
	open("bt", [3]float64{0, 0, 1.27})
	metal("bt", [3]float64{0, 0, 1.17})

	// ISO 14579 M6: a T30, 5.6 across its lobes — a lobe at 30 degrees, a
	// flute on the axis — 2.42 deep below the 6.0 head.
	open("st", [3]float64{2.3816, 1.375, 5.5})
	metal("st", [3]float64{2.3, 0, 5.5})
	open("st", [3]float64{0, 0, 3.63})
	metal("st", [3]float64{0, 0, 3.53})
}
