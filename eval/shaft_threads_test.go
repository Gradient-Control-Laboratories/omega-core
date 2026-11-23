//go:build cgo

package eval

import (
	"context"
	"math"
	"os"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
)

// Shaft Modeling, Phase 5: threads. shaft.thread is thread.male, the exact
// ISO 68-1 field, as a prism segment between its planes. Its basic profile's
// crest is the major diameter and its root the basic minor, d - 1.0825P; at
// height z the middle of the groove faces the angle 2 pi z / P + pi, and the
// middle of the crest 2 pi z / P (thread.helixPhase). A thread runout,
// shaft.threadRunout(length), is a lathe tool or a die retracting: past the
// end plane the groove's floor rises in a straight line from the root to the
// major diameter over `length`.

const (
	threadMajor = 6.0
	threadPitch = 1.0
)

func threadRootR() float64 { return (threadMajor - 1.0825317547305482*threadPitch) / 2 }

func grooveAngle(z float64) float64 { return 2*math.Pi*z/threadPitch + math.Pi }
func crestAngle(z float64) float64  { return 2 * math.Pi * z / threadPitch }

// radialSurfaceAt is radialSurface with an angle of its own at each height.
func radialSurfaceAt(t *testing.T, doc string, angle func(z float64) float64, zs []float64, rMax float64) []float64 {
	t.Helper()
	lo, hi := make([]float64, len(zs)), make([]float64, len(zs))
	for i := range zs {
		hi[i] = rMax
	}
	for iter := 0; iter < 30; iter++ {
		pts := make([][3]float64, len(zs))
		for i, z := range zs {
			r := (lo[i] + hi[i]) / 2
			a := angle(z)
			pts[i] = [3]float64{r * math.Cos(a), r * math.Sin(a), z}
		}
		f := evalShaftNamed(t, doc, "solid", pts)
		for i := range zs {
			if f[i] < 0 {
				lo[i] = (lo[i] + hi[i]) / 2
			} else {
				hi[i] = (lo[i] + hi[i]) / 2
			}
		}
	}
	out := make([]float64, len(zs))
	for i := range zs {
		out[i] = (lo[i] + hi[i]) / 2
	}
	return out
}

// Between its planes, clear of them, the thread segment reads thread.male.
func TestShaftThreads_ThreadIsThreadMale(t *testing.T) {
	doc := `@scene(140)
solid = shaft.body(shaft.thread(6.0, 1.0, 0.0, 20.0));
male = thread.male(6.0, 1.0);
return solid, male;
`
	var pts [][3]float64
	for z := 1.0; z <= 19.0; z += 0.37 {
		for x := -4.0; x <= 4.0; x += 0.21 {
			for y := -4.0; y <= 4.0; y += 0.29 {
				pts = append(pts, [3]float64{x, y, z})
			}
		}
	}
	solid := evalShaftNamed(t, doc, "solid", pts)
	male := evalShaftNamed(t, doc, "male", pts)
	checked, bad, worst := 0, 0, 0.0
	for i, q := range pts {
		if math.Abs(male[i]) > 0.9 {
			continue
		}
		checked++
		miss := math.Abs(solid[i] - male[i])
		worst = math.Max(worst, miss)
		if miss > 1e-5 {
			if bad++; bad <= 3 {
				t.Errorf("at %v: segment %v, thread.male %v", q, solid[i], male[i])
			}
		}
	}
	t.Logf("%d points within 0.9 mm of the thread, worst miss %.2g mm", checked, worst)
}

// A stud cut by shaft.thread in a plate tapped M6x1 by drill.tap, both at
// maximum material, proves within its engagement (thread.engaged), as a cap
// screw does; with no engagement allowed, the same pair is disproven.
func TestShaftThreads_EngagesDrillTap(t *testing.T) {
	doc := `
stud = shaft.body([shaft.thread(6.0, 1.0, -20.0, 0.0), shaft.diameter(10.0, 0.0, 5.0)]);
plate = max(box([-10.0, -10.0, -12.0], [10.0, 10.0, 0.0]),
            -drill.hole(drill.tap(6.0, 1.0, 14.0), 0.0));
ok = max(stud, plate, thread.engaged(thread.maleZone(6.0, 1.0, 0.0, -20.0),
                                     thread.femaleZone(6.0, 1.0, 0.0, -12.0), 0.1));
bare = max(stud, plate);
return ok;
`
	region := [6]float64{-8, -8, -14, 8, 8, -0.3}
	e := New(source.NewLocalManager("testpkg", "../stdlib"), nil)
	r, err := e.Certify(context.Background(), nil, doc, "", "ok", region, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Proven {
		t.Errorf("the threaded stud should prove within its engagement, got %+v", r)
	}
	r, err = e.Certify(context.Background(), nil, doc, "", "bare", region, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Disproven {
		t.Errorf("with no engagement allowed the stud and plate should be disproven, got %+v", r)
	}
}

// A thread run out over 2.5 mm into a plain diameter at the major, then a
// shoulder: in the middle of the groove the floor rises in a straight line
// from the root to the major diameter, and the crest stays at the major; the
// runout band reads as a void.
func TestShaftThreads_RunoutFollowsItsLine(t *testing.T) {
	doc := `@scene(140)
solid = shaft.body([shaft.thread(6.0, 1.0, 0.0, 20.0, shaft.planeEnd(), shaft.threadRunout(2.5)),
                    shaft.diameter(6.0, 20.0, 25.0),
                    shaft.diameter(10.0, 25.0, 30.0)]);
return solid;
`
	rootR, majorR, length := threadRootR(), threadMajor/2, 2.5
	zs := spanZs(15, 25, 0.0625)
	floor := func(z float64) float64 {
		switch {
		case z < 20:
			return rootR
		case z < 20+length:
			return rootR + (majorR-rootR)*(z-20)/length
		}
		return majorR
	}
	checkRadii(t, "groove", zs, radialSurfaceAt(t, doc, grooveAngle, zs, 4.5), floor, 1e-3)
	checkRadii(t, "crest", zs, radialSurfaceAt(t, doc, crestAngle, zs, 4.5), func(z float64) float64 { return majorR }, 1e-3)

	var band [][3]float64
	for z := 20.05; z < 20+length-0.1; z += 0.05 {
		a := grooveAngle(z)
		for r := floor(z) + 0.02; r < majorR-0.02; r += 0.02 {
			band = append(band, [3]float64{r * math.Cos(a), r * math.Sin(a), z})
		}
	}
	f := evalShaftNamed(t, doc, "solid", band)
	bad := 0
	for i, q := range band {
		if f[i] <= 0.005 {
			if bad++; bad <= 3 {
				t.Errorf("runout band at %v reads %v, not a void", q, f[i])
			}
		}
	}
	t.Logf("runout band: %d points, %d not reading empty", len(band), bad)
	if len(band) == 0 {
		t.Fatal("no points in the runout band")
	}
}

// A thread chamfered 0.6 mm at its start: the crest stands at
// major/2 - (0.6 - z) within 0.6 of the end plane, and at the major further on.
func TestShaftThreads_ChamferedStart(t *testing.T) {
	doc := `@scene(140)
solid = shaft.body([shaft.thread(6.0, 1.0, 0.0, 20.0, shaft.chamferEnd(0.6), shaft.planeEnd()),
                    shaft.diameter(10.0, 20.0, 30.0)]);
return solid;
`
	zs := spanZs(0, 2, 0.0625)
	checkRadii(t, "crest", zs, radialSurfaceAt(t, doc, crestAngle, zs, 4.5), func(z float64) float64 {
		return threadMajor/2 - math.Max(0, 0.6-z)
	}, 1e-3)
}

// No sheet inside the solid on the joint planes of a threaded stud, run out
// into its plain diameter and then a shoulder: every point inside the
// material on both sides reads solid.
func TestShaftThreads_NoSheetOnJoints(t *testing.T) {
	doc := `@scene(140)
solid = shaft.body([shaft.thread(6.0, 1.0, 0.0, 20.0, shaft.planeEnd(), shaft.threadRunout(2.5)),
                    shaft.diameter(6.0, 20.0, 25.0),
                    shaft.diameter(10.0, 25.0, 30.0)]);
return solid;
`
	var pts [][3]float64
	for _, z := range []float64{20, 25} {
		for x := -5.0; x <= 5; x += 0.1 {
			for y := -5.0; y <= 5; y += 0.1 {
				pts = append(pts, [3]float64{x, y, z})
			}
		}
	}
	solid := evalShaftNamed(t, doc, "solid", pts)
	checked, bad := 0, 0
	for i, q := range pts {
		r := math.Hypot(q[0], q[1])
		inside := r < threadRootR()-0.05 // the thread's core; the diameter of 6 is larger
		if q[2] == 25 {
			inside = r < 3-0.05 // the diameter of 6; the shoulder is larger
		}
		if !inside {
			continue
		}
		checked++
		if solid[i] > -1e-3 {
			if bad++; bad <= 3 {
				t.Errorf("joint point %v inside both sides reads %v", q, solid[i])
			}
		}
	}
	t.Logf("%d joint points inside both sides, %d not solid", checked, bad)
}

// The threaded example compiles to WGSL (its size logged as Phase 5's budget)
// and evaluates: solid in the thread's core, the neck and the shoulder, empty
// beyond the end.
func TestShaftThreads_ThreadedExampleCompilesToWGSL(t *testing.T) {
	src, err := os.ReadFile("../examples/shaft_threaded.gsl")
	if err != nil {
		t.Fatal(err)
	}
	e := New(nil, nil)
	res, err := e.Shader(t.Context(), nil, string(src), "", ShaderOptions{Target: "wgsl", OutDir: t.TempDir()})
	if err != nil {
		t.Fatalf("shader: %v", err)
	}
	t.Logf("examples/shaft_threaded.gsl: %d bytes of WGSL", len(res.Code))
	groups, err := e.EvalGsl(t.Context(), nil, string(src), "", []string{"threadedShaft"},
		[][3]float64{{0, 0, 10}, {0, 0, 23}, {8, 0, 33}, {0, 0, 72}})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	got := groups[0].Samples[0]
	if got[0].Value >= 0 || got[1].Value >= 0 || got[2].Value >= 0 || got[3].Value <= 0 {
		t.Errorf("thread %v, neck %v, shoulder %v, beyond the end %v: want solid three times, then empty",
			got[0].Value, got[1].Value, got[2].Value, got[3].Value)
	}
}
