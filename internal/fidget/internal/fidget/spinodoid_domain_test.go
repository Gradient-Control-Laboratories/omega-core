package fidget

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/program/noise/spinodoid"
)

func TestSpinodoidFiniteDomainNativePointsAndIntervals(t *testing.T) {
	bounds := [6]float64{980, -30, -10, 1260, 40, 30}
	table, err := spinodoid.Bake(spinodoid.Params{
		Period: 2, PacketWidth: 3, Impulses: 8, Theta: []float64{math.Pi / 2, math.Pi / 2, math.Pi / 2}, Domain: &bounds,
		Scale: func(centers [][3]float64) ([]float64, error) {
			out := make([]float64, len(centers))
			for i, q := range centers {
				out[i] = 1 + .001*q[0]
			}
			return out, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	Lock()
	defer Unlock()
	Clear()
	defer Clear()
	const level = .3
	if err := RegisterSpinodoid(0, table, level, true); err != nil {
		t.Fatal(err)
	}
	canvas := BuildCanvas(SpinodoidTree(0, SpinodoidDistance))
	points := [][3]float32{{981.5, 0, 0}, {1100.25, 20.1, 10.2}, {1259.5, 39, 29}, {800, -270, -10}, {1261, 41, 31}}
	for _, q := range points {
		got, err := EvalGrad(canvas, []float32{q[0]}, []float32{q[1]}, []float32{q[2]})
		if err != nil {
			t.Fatal(err)
		}
		want, grad := table.Distance([3]float64{float64(q[0]), float64(q[1]), float64(q[2])}, level, true)
		actual := []float32{got[0].Value, got[0].Dx, got[0].Dy, got[0].Dz}
		expected := []float64{want, grad[0], grad[1], grad[2]}
		for i, v := range actual {
			if math.Abs(float64(v)-expected[i]) > 1e-5 {
				t.Fatalf("at %v component %d native=%g Go=%g", q, i, v, expected[i])
			}
		}
	}
	boxes := [][6]float32{
		{1100, 20, 10, 1100.5, 20.5, 10.5}, // interior
		{790, -280, -20, 810, -260, 0},     // fully outside, corner distance
		{1259, 39, 29, 1261, 41, 31},       // crosses the domain boundary
	}
	intervals, err := EvalInterval(canvas, boxes)
	if err != nil {
		t.Fatal(err)
	}
	for i, box := range boxes {
		for x := range 4 {
			for y := range 4 {
				for z := range 4 {
					var q [3]float64
					for a, fraction := range [3]int{x, y, z} {
						q[a] = float64(box[a]) + float64(fraction)*(float64(box[a+3])-float64(box[a]))/3
					}
					d, _ := table.Distance(q, level, true)
					if d < float64(intervals[i].Lower) || d > float64(intervals[i].Upper) {
						t.Fatalf("box %v interval %v excludes d=%g at %v", box, intervals[i], d, q)
					}
				}
			}
		}
	}
	// The evaluator can widen coordinates before asking an oracle. The
	// extension must still give finite positive bounds of the right scale.
	if intervals[1].Lower <= 0 || intervals[1].Upper > 500 {
		t.Fatalf("outside interval must bound distance to domain, got %v", intervals[1])
	}
	if !math.IsInf(float64(intervals[2].Lower), -1) || !math.IsInf(float64(intervals[2].Upper), 1) {
		t.Fatalf("boundary interval must cover both field definitions, got %v", intervals[2])
	}
	for axis := range 3 {
		gradientCanvas := BuildCanvas(SpinodoidTree(0, SpinodoidGradX+uint32(axis)))
		bounds, err := EvalInterval(gradientCanvas, boxes[1:2])
		if err != nil {
			t.Fatal(err)
		}
		for _, q := range points[3:] {
			_, g := table.Distance([3]float64{float64(q[0]), float64(q[1]), float64(q[2])}, level, true)
			if g[axis] < float64(bounds[0].Lower) || g[axis] > float64(bounds[0].Upper) {
				t.Fatalf("outside gradient %g excluded by interval %v", g[axis], bounds[0])
			}
		}
	}
}
