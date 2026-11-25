//go:build cgo

package eval

import (
	"strings"
	"testing"
)

// Shaft Modeling, Phase 17 (a): a shaft with detail stands in as its envelope.
// shaft.body sets .simple to shaft.envelope when any segment answers "has
// detail" (want 12: a gear, spline, thread, knurl or bevel), and to nothing
// otherwise; the codegen writes a part whose stand-in is the constant nothing
// as the part alone, so a plain turned shaft compiles as it did.

func standInShader(t *testing.T, segments string) string {
	t.Helper()
	return fixesShader(t, "laid = transform.translate(shaft.body("+segments+"), [0.0, 0.0, 0.0]);\nreturn laid;\n")
}

// A shaft with a gear, a thread, a spline, a knurl or a bevel gets the
// viewer's stand-in switch.
func TestShaftStandIn_DetailedShaftsHaveOne(t *testing.T) {
	for name, segs := range map[string]string{
		"gear":   "[shaft.diameter(20.0, 0.0, 10.0), shaft.gear(18.0, 2.0, 0.35, 0.0, 10.0, 30.0)]",
		"thread": "[shaft.thread(10.0, 1.5, 0.0, 20.0), shaft.diameter(16.0, 20.0, 30.0)]",
		"spline": "[shaft.spline(6.0, 16.0, 20.0, 4.0, 0.0, 20.0), shaft.diameter(26.0, 20.0, 30.0)]",
		"knurl":  "[shaft.diameter(12.0, 0.0, 6.0), shaft.knurl(20.0, 1.0, shaft.knurlDiamond, 6.0, 30.0)]",
		"bevel":  "[shaft.bevel(20.0, 20.0, 2.0, 0.35, 1.5707963267948966, 0.0, 12.0, 0.0, 0.0, 12.7, 20.6), shaft.diameter(10.0, 20.6, 30.0)]",
	} {
		if !strings.Contains(standInShader(t, segs), "fn gsl_simple(") {
			t.Errorf("%s: the shaft has detail but no stand-in", name)
		}
	}
}

// A plain turned shaft, fillets and all, gets none.
func TestShaftStandIn_PlainShaftHasNone(t *testing.T) {
	code := standInShader(t, "[shaft.diameter(20.0, 0.0, 10.0), shaft.fillet(1.0), shaft.diameter(30.0, 10.0, 20.0), shaft.taper(30.0, 20.0, 20.0, 25.0)]")
	if strings.Contains(code, "fn gsl_simple(") {
		t.Errorf("a plain shaft carries a stand-in switch")
	}
}

// The codegen rule on its own: a part whose stand-in is nothing is the part.
func TestSimple_NothingIsNoStandIn(t *testing.T) {
	code := fixesShader(t, "ball = transform.translate(sphere(2.0), [1.0, 0.0, 0.0]);\nball.simple = nothing;\nreturn ball;\n")
	if strings.Contains(code, "fn gsl_simple(") {
		t.Errorf("a stand-in of nothing still writes the switch")
	}
}
