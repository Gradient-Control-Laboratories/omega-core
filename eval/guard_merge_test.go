//go:build cgo

package eval

import (
	"strings"
	"testing"
)

// Shaft Modeling, Phase 21: a split form's two guards are one branch. A
// thread's crests and roots, and a gear's tips and gaps, share one size and so
// one weight; merged, the form is evaluated once inside a single guard.

// documentPart is the shader up to the end of sdShape: the document's own
// functions, without the host's march and shading.
func documentPart(t *testing.T, code string) string {
	t.Helper()
	start := strings.Index(code, "fn sdShape(")
	if start < 0 {
		t.Fatal("no sdShape")
	}
	end := strings.Index(code[start:], "\n}\n")
	if end < 0 {
		t.Fatal("sdShape has no end")
	}
	return code[:start+end]
}

func TestGuardMerge_ThreadFormOnce(t *testing.T) {
	code, err := pruneShader(t, "s = max(thread.male(6.0, 1.0), abs(p.z) - 10.0);\nreturn s;\n")
	if err != nil {
		t.Fatal(err)
	}
	doc := documentPart(t, code)
	if n := strings.Count(doc, "if (gsl_featureWeight"); n != 1 {
		t.Errorf("%d feature guards, want the crests' and the roots' merged into 1", n)
	}
	if n := strings.Count(doc, "acos("); n != 1 {
		t.Errorf("the helix's acos is written %d times, want once", n)
	}
}

func TestGuardMerge_GearProfileOnce(t *testing.T) {
	code, err := pruneShader(t, "g = max(gears.gearProfile(60.0, 2.0, 0.3490658503988659, 0.0, 0.25, 0.1), abs(p.z - 10.0) - 10.0);\nreturn g;\n")
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(documentPart(t, code), "if (gsl_featureWeight"); n != 1 {
		t.Errorf("%d feature guards, want the tips' and the gaps' merged into 1", n)
	}
}
