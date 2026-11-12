//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// catalogEval evaluates catalog expressions at a single point. Every value in
// the catalog is a constant, so the point is irrelevant and only there because
// the evaluator wants one.
func catalogEval(t *testing.T, exprs []string) []float64 {
	t.Helper()
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	groups, err := e.EvalGsl(context.Background(), nil, "return p.x;", "", exprs, [][3]float64{{0, 0, 0}})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	out := make([]float64, len(exprs))
	for i := range exprs {
		out[i] = float64(groups[i].Samples[0][0].Value)
	}
	return out
}

// TestCatalogSizesAreDiscrete is the property that matters about a table of
// standard parts: a lookup hits a row or it misses, and there is nothing in
// between.
//
// The lookup used to be a hat — 1 on the row, sloping to 0 over a half-width
// either side — so that it was total and branchless the way a field has to be.
// A catalog is not a field. catalog.pitch(5.03) answered 0.32 mm, three
// eighths of the M5 pitch, and a caller built a screw out of it and was told
// nothing. Blending two rows does not give a conservative answer or a smooth
// one, it gives a fastener that does not exist.
//
// So a near miss must read EXACTLY zero, not a small number: 0.32 and 0.016
// are both "wrong screw", and only 0 means "not a size I have".
func TestCatalogSizesAreDiscrete(t *testing.T) {
	exprs := []string{
		"catalog.pitch(5.0)",   // 0: on the row
		"catalog.pitch(5.001)", // 1: a hair off, still nothing
		"catalog.pitch(5.03)",  // 2: the old hat answered 0.32 here
		"catalog.pitch(5.049)", // 3: the old hat answered 0.016 here
		"catalog.pitch(5.5)",   // 4: plainly between sizes
		"catalog.pitch(6.0)",   // 5: the next row
	}
	got := catalogEval(t, exprs)

	if math.Abs(got[0]-0.8) > 1e-6 {
		t.Errorf("catalog.pitch(5.0) = %v, want the M5 coarse pitch 0.8", got[0])
	}
	if math.Abs(got[5]-1.0) > 1e-6 {
		t.Errorf("catalog.pitch(6.0) = %v, want the M6 coarse pitch 1.0", got[5])
	}
	for _, miss := range []struct {
		i    int
		expr string
	}{{1, exprs[1]}, {2, exprs[2]}, {3, exprs[3]}, {4, exprs[4]}} {
		if got[miss.i] != 0 {
			t.Errorf("%s = %v, want exactly 0 — a size the catalog does not have "+
				"must build nothing, not a fraction of the nearest size", miss.expr, got[miss.i])
		}
	}
}

// TestCatalogToleratesFloatDrift is the other half: the tolerance exists so a
// caller's own arithmetic still lands on a row. A quarter inch computed rather
// than typed has to find the 0.25 row.
func TestCatalogToleratesFloatDrift(t *testing.T) {
	got := catalogEval(t, []string{
		"catalog.pitch(5.0 + 0.00001)",
		"thread.uncPitch(0.25)",
		"thread.uncPitch(25.4 / 4.0 / 25.4)",
	})
	if math.Abs(got[0]-0.8) > 1e-6 {
		t.Errorf("catalog.pitch(5.0 + 1e-5) = %v, want 0.8 — the tolerance is float "+
			"slack and a key a hair off a row still names that row", got[0])
	}
	if math.Abs(got[1]-1.27) > 1e-4 {
		t.Errorf("thread.uncPitch(0.25) = %v, want the 1/4-20 pitch 1.27 mm", got[1])
	}
	if math.Abs(got[2]-got[1]) > 1e-6 {
		t.Errorf("a computed quarter inch read %v where a written one read %v; "+
			"the tolerance exists exactly so these agree", got[2], got[1])
	}
}

// TestCatalogFinePitchIsAColumn pins that a fine series is a default and not a
// second key. A screw's head, socket, nut and washer are tabulated against the
// thread DIAMETER and do not change with pitch — an M10 socket head cap screw
// has the same head at 1.5 and at 1.25 — and every thread function takes the
// pitch outright rather than looking it up. So the row stays keyed by diameter
// and the fine series is one more column beside the coarse one.
func TestCatalogFinePitchIsAColumn(t *testing.T) {
	exprs := []string{
		"catalog.pitch(10.0)",         // 0
		"catalog.finePitch(10.0)",     // 1
		"catalog.finePitch(8.0)",      // 2
		"catalog.finePitch(12.0)",     // 3
		"catalog.finePitch(6.0)",      // 4
		"thread.metricFinePitch(8.0)", // 5
		"catalog.shcsHeadDia(10.0)",   // 6
		// The pitch reaches the geometry: a fine thread has a shallower form,
		// so its minor diameter is larger than the coarse one at the same size.
		"thread.minorDia(10.0, catalog.pitch(10.0))",     // 7
		"thread.minorDia(10.0, catalog.finePitch(10.0))", // 8
	}
	got := catalogEval(t, exprs)

	for _, want := range []struct {
		i    int
		v    float64
		what string
	}{
		{0, 1.5, "M10 coarse pitch"},
		{1, 1.25, "M10 first fine pitch"},
		{2, 1.0, "M8 first fine pitch"},
		{3, 1.5, "M12 first fine pitch"},
		{4, 0.75, "M6 first fine pitch"},
		{5, 1.0, "thread.metricFinePitch forwards to the catalog"},
		{6, 16.0, "ISO 4762 M10 head diameter, the same at any pitch"},
	} {
		if math.Abs(got[want.i]-want.v) > 1e-6 {
			t.Errorf("%s = %v, want %v (%s)", exprs[want.i], got[want.i], want.v, want.what)
		}
	}
	if !(got[8] > got[7]) {
		t.Errorf("minor diameter at the fine pitch (%v) must exceed the coarse one (%v): "+
			"a finer thread cuts a shallower form, and this is what proves the pitch "+
			"reaches the geometry rather than stopping at the table", got[8], got[7])
	}
}

// TestCatalogRowIsOneSize pins the arrangement the catalog exists for: one row
// per size, so a pitch and a screw head cannot disagree about which sizes are
// real. thread carried pitches for sizes fastener had no rows for, and an M16
// screw came back as a zero head on a zero shank with no complaint.
func TestCatalogRowIsOneSize(t *testing.T) {
	exprs := []string{
		"catalog.pitch(16.0)",          // 0
		"catalog.shcsHeadDia(16.0)",    // 1
		"catalog.nutAcrossFlats(16.0)", // 2
		"catalog.pitch(18.0)",          // 3: second choice, thread but no ISO 4762 screw
		"catalog.shcsHeadDia(18.0)",    // 4
		"catalog.nutAcrossFlats(18.0)", // 5
	}
	got := catalogEval(t, exprs)

	for _, want := range []struct {
		i    int
		v    float64
		what string
	}{
		{0, 2.0, "M16 coarse pitch"},
		{1, 24.0, "ISO 4762 M16 head diameter"},
		{2, 24.0, "ISO 4032 M16 nut across flats"},
		{3, 2.5, "M18 coarse pitch"},
		{5, 27.0, "ISO 4032 M18 nut across flats"},
	} {
		if math.Abs(got[want.i]-want.v) > 1e-6 {
			t.Errorf("%s = %v, want %v", exprs[want.i], got[want.i], want.v)
		}
	}
	// A zero cell is information: ISO 4762 has no M18, so asking for one builds
	// nothing rather than inventing a head.
	if got[4] != 0 {
		t.Errorf("catalog.shcsHeadDia(18.0) = %v, want 0 — ISO 4762 has no M18", got[4])
	}
}
