//go:build cgo

package eval

import (
	"context"
	"math"
	"strconv"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// TestDriverOnePunch checks the fact the whole cross tip rests on: a recess
// number is one punch. ISO 7045 gives each size its own recess diameter and
// penetration, and for every size a number goes in they put the 26 degree 30
// wing-end cone at the same radius on the recess floor — so one tip, sized off
// any of them, fits all of them.
func TestDriverOnePunch(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	const doc = `
x = 0.0;
return sphere(1.0);
`
	val := func(expr string) float64 {
		t.Helper()
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{expr}, [][3]float64{{0, 0, 0}})
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		return float64(groups[0].Samples[0][0].Value)
	}
	tan := math.Tan(26.5 * math.Pi / 180)
	for no, sizes := range map[float64][]float64{
		0: {1.6, 2.0}, 1: {2.5, 3.0}, 2: {3.5, 4.0, 5.0}, 3: {6.0}, 4: {8.0, 10.0},
	} {
		gauge := val("driver.phillipsGauge(" + ftoa(no) + ")")
		for _, m := range sizes {
			dia := val("catalog.panPhillipsDia(" + ftoa(m) + ")")
			depth := val("catalog.panPhillipsDepth(" + ftoa(m) + ")")
			if got := dia/2 - depth*tan; math.Abs(got-gauge) > 0.05 {
				t.Errorf("PH%v in M%v: the cone sits at %.3f on the floor, the tip's gauge radius is %.3f",
					no, m, got, gauge)
			}
		}
	}

	// ISO 8764 sizes and the gauge radius of a PH2 and a PZ2.
	for expr, want := range map[string]float64{
		"driver.bladeDia(2.0)":      6.0,
		"driver.bladeLength(2.0)":   100.0,
		"driver.fluteLength(2.0)":   4.37,
		"driver.phillipsGauge(2.0)": 1.0034,
		"driver.pozidrivGauge(2.0)": 0.9833,
	} {
		if got := val(expr); math.Abs(got-want) > 1e-3 {
			t.Errorf("%s = %.4f, want %.4f", expr, got, want)
		}
	}
}

// TestDriverTipsFitTheirScrews seats each tip in the recess it is for and
// measures the overlap: the smallest value of max(bit, screw) over the head. A
// tip that is the recess read from the other side — Torx — seated on the
// floor TOUCHES: the minimum is zero, and it passes through nowhere, so it is
// not negative. A cross tip is ISO 4757's gauge instead, and ISO 7045 says
// where that touches: seated at the minimum penetration it is proven clear,
// and by the maximum it has touched (0.03 past it for Phillips, as
// TestPhillipsGaugeSeatsWithinPenetration allows for M3). A PH2 and a PZ2 are
// tried in all three ISO 7045 sizes they go in, not just the one they are
// sized off.
func TestDriverTipsFitTheirScrews(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	const doc = `
bitPH = driver.bit(driver.phillipsTip(2.0), driver.insertShank(driver.fluteLength(2.0), 0.7));
bitPZ = driver.bit(driver.pozidrivTip(2.0), driver.insertShank(driver.fluteLength(2.0), 0.9));
bitTX = driver.bit(driver.torxTip(25.0, 5.0), driver.insertShank(5.0, 1.6));
fitPH(m) {
  size = m;
  return max(remap(bitPH, [p.x, p.y, p.z - fastener.panPhillipsSeat(size)]),
             fastener.panPhillips(size, 6.0));
}
deepPH(m) {
  size = m;
  return max(remap(bitPH, [p.x, p.y, p.z - (fastener.panPhillipsFloor(size) - 0.03)]),
             fastener.panPhillips(size, 6.0));
}
fitPZ(m) {
  size = m;
  return max(remap(bitPZ, [p.x, p.y, p.z - fastener.panPozidrivSeat(size)]),
             fastener.panPozidriv(size, 6.0));
}
deepPZ(m) {
  size = m;
  return max(remap(bitPZ, [p.x, p.y, p.z - fastener.panPozidrivFloor(size)]),
             fastener.panPozidriv(size, 6.0));
}
ph35 = fitPH(3.5);
ph4 = fitPH(4.0);
ph5 = fitPH(5.0);
deep35 = deepPH(3.5);
deep4 = deepPH(4.0);
deep5 = deepPH(5.0);
pz35 = fitPZ(3.5);
pz4 = fitPZ(4.0);
pz5 = fitPZ(5.0);
deepZ35 = deepPZ(3.5);
deepZ4 = deepPZ(4.0);
deepZ5 = deepPZ(5.0);
tx5 = max(remap(bitTX, [p.x, p.y, p.z - fastener.panTorxFloor(5.0)]), fastener.panTorx(5.0, 6.0));
return ph4;
`
	e.SetCertifyBudget(20000000)
	region := [6]float64{-5.5, -5.5, -1, 5.5, 5.5, 4.5}
	for _, expr := range []string{"ph35", "ph4", "ph5", "pz35", "pz4", "pz5"} {
		res, err := e.Certify(ctx, nil, doc, "", expr, region, 0, 1e-4)
		if err != nil {
			t.Fatalf("%s: %v", expr, err)
		}
		if !res.Proven {
			t.Errorf("%s: the bit at the minimum penetration is not proven clear: disproven %v, %s, counter box %v",
				expr, res.Disproven, res.Reason, res.CounterBox)
		}
	}
	for _, expr := range []string{"deep35", "deep4", "deep5", "deepZ35", "deepZ4", "deepZ5"} {
		res, err := e.Extremum(ctx, nil, doc, "", expr, region, "min", 0.01, 1)
		if err != nil {
			t.Fatalf("%s: %v", expr, err)
		}
		if !(res.Min < 0) {
			t.Errorf("%s: the bit has not touched by the maximum penetration: min %.4f at %v", expr, res.Min, res.MinAt)
		}
	}
	for _, expr := range []string{"tx5"} {
		res, err := e.Extremum(ctx, nil, doc, "", expr, [6]float64{-5.5, -5.5, -1, 5.5, 5.5, 4.5}, "min", 0.01, 1)
		if err != nil {
			t.Fatalf("%s: %v", expr, err)
		}
		if res.Min < -0.002 {
			t.Errorf("%s: the seated bit is %.4f inside the screw at %v", expr, -res.Min, res.MinAt)
		}
		if res.Min > 0.002 {
			t.Errorf("%s: the seated bit never touches the screw (%.4f clear) — it is not seated", expr, res.Min)
		}
	}
}

// TestDriverShanks checks the drive ends ISO 1173 gives the shanks: the 1/4"
// hex 6.32 across flats, the V-groove a form C bit's corners carry, the round
// groove a form E power bit locks by, and the 4 mm hex of a precision bit.
func TestDriverShanks(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	const doc = `
insert = driver.insertShank(5.0, 1.0);
power = driver.powerShank(5.0, 1.0);
micro = driver.microShank(3.0, 0.6);
return insert;
`
	at := func(expr string, pt [3]float64) float64 {
		t.Helper()
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{expr}, [][3]float64{pt})
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		return float64(groups[0].Samples[0][0].Value)
	}
	on := func(expr string, pt [3]float64) {
		t.Helper()
		if v := at(expr, pt); math.Abs(v) > 1e-3 {
			t.Errorf("%s at %v = %.4f, want on the surface", expr, pt, v)
		}
	}
	metal := func(expr string, pt [3]float64) {
		t.Helper()
		if v := at(expr, pt); v >= 0 {
			t.Errorf("%s at %v = %.4f, want inside", expr, pt, v)
		}
	}
	cut := func(expr string, pt [3]float64) {
		t.Helper()
		if v := at(expr, pt); v <= 0 {
			t.Errorf("%s at %v = %.4f, want cut away", expr, pt, v)
		}
	}

	// Form C 6.3: flats 6.32 across. A corner is 3.65 out; 8.2 from the drive
	// end the V-groove takes it down to 3.35, and 3 mm either side it is whole.
	on("insert", [3]float64{3.16, 0, 12.0})
	corner := func(r, z float64) [3]float64 { return [3]float64{r * math.Cos(math.Pi/6), r * math.Sin(math.Pi/6), z} }
	cut("insert", corner(3.5, 16.8))
	metal("insert", corner(3.5, 13.0))
	metal("insert", corner(3.3, 16.8))

	// Form E 6.3: the round groove, root 4.7 across at 9.5 from the drive end
	// of a 50-long bit, cuts through the flats; above and below it the hex is
	// whole.
	cut("power", [3]float64{2.5, 0, 40.5})
	metal("power", [3]float64{2.2, 0, 40.5})
	metal("power", [3]float64{3.0, 0, 45.0})
	metal("power", [3]float64{3.0, 0, 30.0})

	// Form C 4: flats 3.94 across.
	on("micro", [3]float64{1.97, 0, 15.0})
}

func ftoa(v float64) string {
	return strconv.FormatFloat(v, 'f', 1, 64)
}
