//go:build cgo

package eval

import (
	"math"
	"testing"
)

// Shaft Modeling, Phase 16: cross holes and set screws. shaft.crossVoid
// (features, dia, z, angle) is a mechanical/drill stack drilled radially into
// the shaft, its mouth the plane tangent to the cylinder `dia` across at
// height z, `angle` from +x about the axis, so every depth is measured from
// the surface as a drawing gives it; shaft.crossHole(void) is the entry in a
// shaft's list that cuts it. The fold passes the entry; body cuts its void
// after the fold, as it cuts a runout's.
// shaft.crossPlace(part, dia, z, angle) puts a part built on a +Z datum (a
// set screw's seating plane at z = 0, its shank down) on the same axis, with
// transform, so it keeps its box.

const crossHoleDoc = `
plainShaft = shaft.body([shaft.diameter(20.0, 0.0, 40.0)]);
blindShaft = shaft.body([shaft.diameter(20.0, 0.0, 40.0), shaft.crossHole(shaft.crossVoid(drill.blind(5.0, 8.0), 20.0, 15.0, 0.0))]);
thruShaft = shaft.body([shaft.diameter(20.0, 0.0, 40.0), shaft.crossHole(shaft.crossVoid(drill.thru(6.0), 20.0, 25.0, pi / 2.0))]);
tappedShaft = shaft.body([shaft.diameter(20.0, 0.0, 40.0), shaft.crossHole(shaft.crossVoid([drill.tap(6.0, 10.0), drill.blind(5.0, 12.0)], 20.0, 15.0, pi))]);
filleted = shaft.body([shaft.diameter(20.0, 0.0, 20.0), shaft.fillet(2.0), shaft.diameter(30.0, 20.0, 40.0)]);
filletedHoled = shaft.body([shaft.diameter(20.0, 0.0, 20.0), shaft.fillet(2.0), shaft.diameter(30.0, 20.0, 40.0), shaft.crossHole(shaft.crossVoid(drill.thru(4.0), 30.0, 32.0, 0.5))]);
pin = max(length(p.xy) - 1.0, p.z, -10.0 - p.z);
placedPin = shaft.crossPlace(pin, 20.0, 15.0, 0.0);
return plainShaft;
`

// A blind cross hole: open from the surface down its axis to its depth, its
// drill point below that, and solid beside it.
func TestShaftCrossHole_Blind(t *testing.T) {
	pts := [][3]float64{
		{9, 0, 15},   // 1 mm in, on the axis
		{5, 0, 15},   // 5 mm in, where the wall is nearer than the drill point
		{-2, 0, 15},  // 12 mm in, past the drill point
		{5, 4, 15},   // beside the hole
		{9.5, 0, 22}, // well away from it, in the shaft
	}
	v := undercutEval(t, crossHoleDoc, []string{"blindShaft"}, pts)[0]
	if v[0] < 2.4 || v[1] < 2.4 {
		t.Errorf("down the hole: %.4f and %.4f, want open by the drill's radius 2.5", v[0], v[1])
	}
	for _, i := range []int{2, 3, 4} {
		if v[i] > -0.4 {
			t.Errorf("at %v: %.4f, want solid", pts[i], v[i])
		}
	}
}

// A through hole goes out the far side; placed at angle pi/2 it runs along y.
func TestShaftCrossHole_Through(t *testing.T) {
	pts := [][3]float64{{0, 0, 25}, {0, -9, 25}, {0, 9, 25}, {4, 0, 25}}
	v := undercutEval(t, crossHoleDoc, []string{"thruShaft"}, pts)[0]
	for i := 0; i < 3; i++ {
		if math.Abs(v[i]-3) > 1e-3 {
			t.Errorf("along the through hole at %v: %.4f, want its radius 3", pts[i], v[i])
		}
	}
	if v[3] > -0.5 {
		t.Errorf("beside the through hole: %.4f, want solid", v[3])
	}
}

// A tapped hole, entering from -x at angle pi: open on its axis to its depth.
func TestShaftCrossHole_Tapped(t *testing.T) {
	pts := [][3]float64{{-8, 0, 15}, {-4, 0, 15}, {-12, 0, 15}}
	v := undercutEval(t, crossHoleDoc, []string{"tappedShaft"}, pts)[0]
	if v[0] < 2.0 || v[1] < 2.0 {
		t.Errorf("down the tapped hole: %.4f and %.4f, want open", v[0], v[1])
	}
	if v[2] < 1.9 {
		t.Errorf("just outside the shaft, before the mouth: %.4f, want open", v[2])
	}
}

// Away from its hole the shaft is unchanged, and so is its box.
func TestShaftCrossHole_ElsewhereUnchanged(t *testing.T) {
	pts := [][3]float64{{10.5, 0, 35}, {0, 9, 5}, {-6, -6, 30}, {0, 0, 41}, {12, 0, 25}}
	v := undercutEval(t, crossHoleDoc, []string{"blindShaft", "plainShaft"}, pts)
	for i := range pts {
		if math.Abs(v[0][i]-v[1][i]) > 1e-5 {
			t.Errorf("at %v the holed shaft reads %.5f, the plain one %.5f", pts[i], v[0][i], v[1][i])
		}
	}
	lo, hi := readBox(t, crossHoleDoc, "blindShaft")
	plo, phi := readBox(t, crossHoleDoc, "plainShaft")
	if lo != plo || hi != phi {
		t.Errorf("box %v .. %v, the plain shaft's %v .. %v", lo, hi, plo, phi)
	}
}

// The fold passes the entry: a fillet listed before it still blends its joint.
func TestShaftCrossHole_FoldPassesIt(t *testing.T) {
	pts := [][3]float64{{11, 0, 21}, {10.6, 0, 20.6}, {11.5, 0, 21.5}}
	v := undercutEval(t, crossHoleDoc, []string{"filletedHoled", "filleted"}, pts)
	for i := range pts {
		if math.Abs(v[0][i]-v[1][i]) > 1e-5 {
			t.Errorf("in the fillet at %v: %.5f with the hole listed after it, %.5f without", pts[i], v[0][i], v[1][i])
		}
	}
}

// A part placed on the cross axis: a pin 2 mm across, 10 mm below its datum,
// runs in from the surface along +x at z = 15, and keeps its box.
func TestShaftCrossPlace(t *testing.T) {
	pts := [][3]float64{{5, 0, 15}, {9.5, 0, 15}, {11, 0, 15}, {5, 2, 15}, {-1, 0, 15}}
	v := undercutEval(t, crossHoleDoc, []string{"placedPin"}, pts)[0]
	if v[0] > -0.9 || v[1] > -0.4 {
		t.Errorf("inside the pin: %.4f and %.4f, want solid", v[0], v[1])
	}
	if math.Abs(v[2]-1) > 1e-3 || math.Abs(v[3]-1) > 1e-3 || math.Abs(v[4]-1) > 1e-3 {
		t.Errorf("a millimetre off the pin: %.4f, %.4f, %.4f, want 1", v[2], v[3], v[4])
	}
	lo, hi := readBox(t, crossHoleDoc, "placedPin")
	want := [2][3]float64{{0, -1, 14}, {10, 1, 16}}
	for k := 0; k < 3; k++ {
		if math.Abs(lo[k]-want[0][k]) > 0.05 || math.Abs(hi[k]-want[1][k]) > 0.05 {
			t.Errorf("placed pin's box %v .. %v, want %v .. %v", lo, hi, want[0], want[1])
			break
		}
	}
}
