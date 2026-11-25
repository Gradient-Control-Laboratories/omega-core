//go:build cgo

package eval

import (
	"context"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
)

// Threads simplify to their pitch cylinder via .simple + .featureSize (P / 2,
// as a gear's is pi m / 2). Far enough away a screw, a tapped hole or a
// shaft.thread reads its pitch cylinder while evaluating the thread form once.

// Every geometric thread in the library opts in: shaft.thread, the threaded
// fastener shanks and nuts, drill.tapThreaded, and the sinusoid forms. Each
// thread emits a gsl_simpleSized switch; a drilled hole also has a
// gsl_featureSize guard, so a tapped hole has one of each.
func TestThreadPrune_LibraryOptsIn(t *testing.T) {
	for _, c := range []struct {
		name     string
		doc      string
		switches int
		guards   int
	}{
		{"shaft.thread", "s = shaft.body([shaft.thread(6.0, 1.0, 0.0, 20.0)]);\nreturn s;\n", 1, 0},
		{"threadedShank", "s = fastener.threadedShank(6.0, 1.0, 20.0);\nreturn s;\n", 1, 0},
		{"sinShank", "s = fastener.sinShank(6.0, 1.0, 20.0);\nreturn s;\n", 1, 0},
		{"nutThreaded", "n = fastener.nutThreaded(6.0);\nreturn n;\n", 1, 0},
		{"drill.tapThreaded", "plate = box([-20.0, -20.0, -10.0], [20.0, 20.0, 0.0]);\ntapped = max(plate, neg(drill.hole(drill.tapThreaded(6.0, 1.0, 8.0), 0.0)));\nreturn tapped;\n", 1, 1},
		{"thread.maleSin", "s = max(thread.maleSin(6.0, 1.0), abs(p.z) - 10.0);\nreturn s;\n", 1, 0},
		{"thread.female", "s = max(box([-10.0, -10.0, -5.0], [10.0, 10.0, 5.0]), neg(thread.female(6.0, 1.0)));\nreturn s;\n", 1, 0},
	} {
		code, err := pruneShader(t, c.doc)
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if n := strings.Count(code, "gsl_simpleSized("); n < c.switches {
			t.Errorf("%s: %d stand-in switches, want at least %d", c.name, n, c.switches)
		}
		if n := strings.Count(code, "let gsl_featureSize"); n < c.guards {
			t.Errorf("%s: %d feature guards, want at least %d", c.name, n, c.guards)
		}
	}
}

// A thread's stand-in switches at half a pitch: an M6 x 1's is 0.5.
func TestThreadPrune_FeatureSizeIsHalfAPitch(t *testing.T) {
	code, err := pruneShader(t, "s = max(thread.male(6.0, 1.0), abs(p.z) - 10.0);\nreturn s;\n")
	if err != nil {
		t.Fatal(err)
	}
	sizes := regexp.MustCompile(`gsl_simpleSized\((-?[0-9.e+-]+)\)`).FindAllStringSubmatch(code, -1)
	if len(sizes) != 1 {
		t.Fatalf("want one literal stand-in size, found %d:\n%s", len(sizes), code)
	}
	for _, m := range sizes {
		if v, _ := strconv.ParseFloat(m[1], 64); math.Abs(v-0.5) > 1e-5 {
			t.Errorf("a thread stand-in's size is %g, want P / 2 = 0.5", v)
		}
	}
}

// thread.basicForm and thread.sinForm as they were before Phase 20, frozen as
// the reference the re-authored forms must keep: the same solid.
const oldThreadFormDoc = `
oldHelixPhase(pitchVal) {
  radial = max(length(p.xy), 0.2);
  ang = 2.0 * pi * p.z / max(pitchVal, 0.01);
  along = cos(ang) * (p.x / radial) + sin(ang) * (p.y / radial);
  return acos(clamp(along, -1.0, 1.0)) / pi;
}
oldLeadCos(major, pitchVal) {
  reference = max(major / 2.0 - 0.5412659 * pitchVal - 0.65 * pitchVal, 0.01);
  return reference / length([reference, pitchVal / (2.0 * pi)]);
}
oldGroove(major, pitchVal) {
  minorR = major / 2.0 - 0.5412659 * pitchVal;
  rad = length(p.xy);
  flank = (1.0 - oldHelixPhase(pitchVal)) * pitchVal * 0.5 - pitchVal / 8.0 - (rad - minorR) * 0.57735027;
  scale = 1.0 / length([0.8660254 / oldLeadCos(major, pitchVal), 0.5]);
  return max(minorR - rad, flank * 0.8660254 * scale);
}
oldBasicForm(major, pitchVal) {
  return max(length(p.xy) - major / 2.0, neg(oldGroove(major, pitchVal)));
}
oldSinForm(major, pitchVal) {
  radial = max(length(p.xy), 0.2);
  ang = 2.0 * pi * p.z / max(pitchVal, 0.01);
  along = cos(ang) * (p.x / radial) + sin(ang) * (p.y / radial);
  radiusAt = major / 2.0 - 0.5412659 * pitchVal * 0.5 * (1.0 - along);
  return (length(p.xy) - radiusAt) * 0.7;
}
newIso = thread.male(6.0, 1.0);
oldIso = oldBasicForm(6.0, 1.0);
newSin = thread.maleSin(6.0, 1.0);
oldSin = oldSinForm(6.0, 1.0);
return newIso;
`

// threadPoints samples two pitches of an M6 x 1 along +x and at a quarter
// turn, from inside the root to outside the crest.
func threadPoints() [][3]float64 {
	var pts [][3]float64
	for _, a := range []float64{0, math.Pi / 2} {
		for i := 0; i <= 60; i++ {
			r := 2.2 + 1.2*float64(i)/60
			for j := 0; j <= 80; j++ {
				z := -1 + 2*float64(j)/80
				pts = append(pts, [3]float64{r * math.Cos(a), r * math.Sin(a), z})
			}
		}
	}
	return pts
}

// The re-authored forms are the same solid: wherever the old one is clearly
// inside or outside, so is the new one.
func TestThreadForm_SameSolid(t *testing.T) {
	e := New(&source.MockManager{}, nil)
	pts := threadPoints()
	g, err := e.EvalGsl(context.Background(), nil, oldThreadFormDoc, "", []string{"newIso", "oldIso", "newSin", "oldSin"}, pts)
	if err != nil {
		t.Fatal(err)
	}
	for _, form := range []struct {
		name     string
		new, old int
	}{{"iso", 0, 1}, {"sin", 2, 3}} {
		bad := 0
		for i := range pts {
			n, o := float64(g[form.new].Samples[0][i].Value), float64(g[form.old].Samples[0][i].Value)
			if math.Abs(o) > 0.01 && (n < 0) != (o < 0) {
				if bad < 5 {
					t.Errorf("%s at %.3f, %.3f, %.3f: new %g, old %g", form.name, pts[i][0], pts[i][1], pts[i][2], n, o)
				}
				bad++
			}
		}
		if bad > 0 {
			t.Errorf("%s: %d of %d points disagree on inside or outside", form.name, bad, len(pts))
		}
	}
}

// On the pitch cylinder, where the cylinder meets the crests and the roots,
// the new form is as deep inside a crest as half the old one's depth there,
// and as far outside in a root as half the old one's: no sheet of zero, nor a
// shallow one, lies inside the metal or across the groove.
func TestThreadForm_NoShallowSheetAtPitch(t *testing.T) {
	pitchR := (6.0 - 0.6495) / 2
	var pts [][3]float64
	for j := 0; j <= 200; j++ {
		pts = append(pts, [3]float64{pitchR, 0, -1 + 2*float64(j)/200})
	}
	e := New(&source.MockManager{}, nil)
	g, err := e.EvalGsl(context.Background(), nil, oldThreadFormDoc, "", []string{"newIso", "oldIso", "newSin", "oldSin"}, pts)
	if err != nil {
		t.Fatal(err)
	}
	for _, form := range []struct {
		name     string
		new, old int
	}{{"iso", 0, 1}, {"sin", 2, 3}} {
		deepest, farthest := 0, 0
		for i := range pts {
			v := g[form.old].Samples[0][i].Value
			if v < g[form.old].Samples[0][deepest].Value {
				deepest = i
			}
			if v > g[form.old].Samples[0][farthest].Value {
				farthest = i
			}
		}
		for _, c := range []struct {
			where string
			i     int
		}{{"inside a crest", deepest}, {"across a root", farthest}} {
			n, o := float64(g[form.new].Samples[0][c.i].Value), float64(g[form.old].Samples[0][c.i].Value)
			if math.Abs(o) < 0.05 {
				t.Fatalf("%s: the reference is only %g %s on the pitch cylinder", form.name, o, c.where)
			}
			if math.Abs(n) < 0.5*math.Abs(o) || (n < 0) != (o < 0) {
				t.Errorf("%s %s on the pitch cylinder: new %g, old %g", form.name, c.where, n, o)
			}
		}
	}
}
