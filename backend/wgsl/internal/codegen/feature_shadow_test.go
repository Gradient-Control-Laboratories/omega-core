package codegen_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
)

// Two features in one function. Each writes its weight as a let beside the
// module function gsl_featureWeight it calls; a let of the function's own name
// shadows it from there on, and the next feature's call then names the let,
// which Tint rejects ("cannot use 'let gsl_featureWeight' as call target").
// No local may take the name of a module-scope function.
func TestFeatureWeightLocalsDoNotShadowTheFunction(t *testing.T) {
	const src = `a = length(p - [10.0, 0.0, 0.0]) - 5.0;
a.featureSize = 1.0;
b = length(p + [10.0, 0.0, 0.0]) - 5.0;
b.featureSize = 2.0;
return min(a, b);
`
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mockMgr))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	out, err := renderHelper(src, res.Program)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if got := strings.Count(out.WGSL, "= gsl_featureWeight("); got != 2 {
		t.Fatalf("want two features calling gsl_featureWeight in one function, got %d calls:\n%s", got, out.WGSL)
	}

	fns := map[string]bool{}
	for _, m := range regexp.MustCompile(`(?m)^fn (\w+)\(`).FindAllStringSubmatch(out.WGSL, -1) {
		fns[m[1]] = true
	}
	for _, m := range regexp.MustCompile(`(?m)^\s+(?:let|var) (\w+)\b`).FindAllStringSubmatch(out.WGSL, -1) {
		if fns[m[1]] {
			t.Errorf("local %q shadows the module function of the same name", m[1])
		}
	}
}
