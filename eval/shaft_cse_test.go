//go:build cgo

package eval

import (
	"os"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/backend/wgsl"
)

// Shaft Modeling, Phase 18: the WGSL a shaft compiles to computes each
// section once. body's fold asks every segment four times, capped and
// extended, and the section does not depend on the extension, so after the
// codegen's common-subexpression pass each one is written once.

func cseExampleShader(t *testing.T, name string) string {
	t.Helper()
	src, err := os.ReadFile("../examples/" + name + ".gsl")
	if err != nil {
		t.Fatal(err)
	}
	return fixesShader(t, string(src))
}

// The pinion's helical gear: its profile is called once, not once per copy.
func TestShaftCSE_PinionProfileOnce(t *testing.T) {
	code := cseExampleShader(t, "shaft_pinion")
	calls := 0
	for _, line := range strings.Split(code, "\n") {
		if strings.HasPrefix(line, "fn gears_gearProfile(") {
			continue
		}
		calls += strings.Count(line, "gears_gearProfile(")
	}
	if calls != 1 {
		t.Errorf("gears_gearProfile is called %d times, want 1", calls)
	}
}

// The turned shaft: no function holds two lets of the same expression. (A
// short repeat such as length(p.xy) is named only where naming it shortens
// the function, so a few of those may stay.)
func TestShaftCSE_TurnedNoRepeatedLet(t *testing.T) {
	code := cseExampleShader(t, "shaft_turned")
	// The document's functions only: the host's field wrapper and its march,
	// hand-written, follow them.
	if cut := strings.Index(code, "\n\nfn "+wgsl.FieldFn+"("); cut >= 0 {
		code = code[:cut]
	}
	for _, fn := range strings.Split(code, "\nfn ")[1:] {
		name := fn[:strings.IndexByte(fn, '(')]
		seen := make(map[string]bool)
		for _, line := range strings.Split(fn, "\n") {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "let ") {
				continue
			}
			eq := strings.Index(trimmed, " = ")
			if eq < 0 {
				continue
			}
			expr := trimmed[eq+3:]
			if seen[expr] {
				t.Errorf("%s holds two lets of %s", name, expr)
			}
			seen[expr] = true
		}
	}
}
