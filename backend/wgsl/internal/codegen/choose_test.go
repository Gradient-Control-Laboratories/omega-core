package codegen_test

import (
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
)

func chooseWGSL(t *testing.T, src string) string {
	t.Helper()
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mockMgr))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	out, err := renderHelper(src, res.Program)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return out.WGSL
}

// choose(condition, onTrue, onFalse) is written as a branch, each side in a
// guard of its own, so only the side taken is computed; the select that
// follows picks between the two results, not between two computations.
func TestChoose_IsABranch(t *testing.T) {
	wgsl := chooseWGSL(t, `@scene(40)
@range(0, 1) kind = 1.0;
solid = choose(kind > 0.5, length(p - [3, 0, 0]) - 2, max(abs(p.x), abs(p.y), abs(p.z)) - 2);
return solid;
`)
	for _, want := range []string{
		"let gsl_branch = kind > 0.5;",
		"if (gsl_branch) {\n    gsl_onTrue = length(p - vec3<f32>(3.0, 0.0, 0.0)) - 2.0;",
		"if (!gsl_branch) {\n    gsl_onFalse = max(max(abs(p.x), abs(p.y)), abs(p.z)) - 2.0;",
		"select(gsl_onFalse, gsl_onTrue, gsl_branch)",
	} {
		if !strings.Contains(wgsl, want) {
			t.Errorf("want %q in the WGSL:\n%s", want, wgsl)
		}
	}
	if strings.Contains(wgsl, "builtin_choose") {
		t.Error("choose is written as a function; every call should be its branch")
	}
}

// A condition that reads nothing is a constant the GPU compiler folds out of a
// select, so choose leaves it one rather than add a branch.
func TestChoose_ConstantConditionStaysASelect(t *testing.T) {
	wgsl := chooseWGSL(t, `@scene(40)
solid = choose(1 > 0, length(p) - 2, length(p) - 3);
return solid;
`)
	if strings.Contains(wgsl, "gsl_branch") || !strings.Contains(wgsl, "select(") {
		t.Errorf("want a select and no branch:\n%s", wgsl)
	}
}
