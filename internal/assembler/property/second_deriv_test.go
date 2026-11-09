package property

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// TestSecondDerivative checks that .dtime.dtime (a DERIV nested in a DERIV)
// yields the genuine second derivative, distinct from the first.
func TestSecondDerivative(t *testing.T) {
	d := newDerivative()

	// f = sin(time): f' = cos(time), f'' = -sin(time).
	first, err := d.temporal(un(program.Operation_SIN, ref(program.TimeVarName)))
	if err != nil {
		t.Fatalf("first temporal: %v", err)
	}
	inner := nary(program.Operation_DERIV, un(program.Operation_SIN, ref(program.TimeVarName)), ref(program.TimeVarName))
	second, err := d.temporal(inner)
	if err != nil {
		t.Fatalf("second temporal: %v", err)
	}

	for _, pt := range []point{{0, 0, 0, 0.4}, {0, 0, 0, 1.3}, {0, 0, 0, -0.7}} {
		got1 := evalScalar(t, first, pt)
		got2 := evalScalar(t, second, pt)
		if !closeEnough(got1, math.Cos(pt.time)) {
			t.Errorf("first at t=%v: got %v want cos=%v", pt.time, got1, math.Cos(pt.time))
		}
		if !closeEnough(got2, -math.Sin(pt.time)) {
			t.Errorf("second at t=%v: got %v want -sin=%v", pt.time, got2, -math.Sin(pt.time))
		}
	}
}

// TestSecondDerivativeVector checks .dtime.dtime on a vector field, mirroring
// circular motion: center = [sin(time), cos(time), 0]; acceleration = -center.
func TestSecondDerivativeVector(t *testing.T) {
	d := newDerivative()
	center := nary(program.Operation_COLLECT, un(program.Operation_SIN, ref(program.TimeVarName)), un(program.Operation_COS, ref(program.TimeVarName)), num(0))
	inner := nary(program.Operation_DERIV, center, ref(program.TimeVarName))
	accel, err := d.temporal(inner)
	if err != nil {
		t.Fatalf("temporal: %v", err)
	}
	for _, pt := range []point{{0, 0, 0, 0.4}, {0, 0, 0, 2.1}} {
		got := evalVec(t, accel, pt)
		want := []float64{-math.Sin(pt.time), -math.Cos(pt.time), 0}
		for k := 0; k < 3; k++ {
			if !closeEnough(got[k], want[k]) {
				t.Errorf("accel at t=%v comp %d: got %v want %v", pt.time, k, got[k], want[k])
			}
		}
	}
}
