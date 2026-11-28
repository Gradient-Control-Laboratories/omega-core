package codegen_test

import (
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
)

// toleranceDoc is a block with a bored hole and a shaft, the bore's and the
// shaft's radii working tolerances.
const toleranceDoc = `
boreR = 12.51175;
boreR.tolerance = tolerance.symmetric(0.003);
shaftR = 12.49325;
shaftR.tolerance = tolerance.symmetric(0.0016);
block = boxCentered([0, 0, 0], [40.0, 40.0, 20.0]);
holed = max(block, 0.0 - (length(p.xy) - boreR));
shaft = max(length(p.xy) - shaftR, abs(p.z) - 15.0);
return holed, shaft;
`

// A document with working tolerances carries each return's tolerance to the
// viewer as gsl_toleranceField, one case a return; one without carries none,
// and compiles as before.
func TestToleranceField_OnlyForTolerancedDocuments(t *testing.T) {
	for _, c := range []struct {
		name, doc string
		want      bool
	}{
		{"toleranced", toleranceDoc, true},
		{"plain", strings.NewReplacer("boreR.tolerance = tolerance.symmetric(0.003);\n", "", "shaftR.tolerance = tolerance.symmetric(0.0016);\n", "").Replace(toleranceDoc), false},
	} {
		t.Run(c.name, func(t *testing.T) {
			res, err := assembler.Assemble([]string{c.doc}, assembler.SourceManager(mockMgr))
			if err != nil {
				t.Fatalf("assemble: %v", err)
			}
			tols := res.Program.ToleranceReturns()
			if c.want && len(tols) != 2 {
				t.Fatalf("%d tolerance returns, want one a return (2)", len(tols))
			}
			if !c.want && len(tols) != 0 {
				t.Fatalf("%d tolerance returns for a document with no tolerance", len(tols))
			}
			out, err := renderHelper(c.doc, res.Program)
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			has := strings.Contains(out.WGSL, "fn gsl_toleranceField(")
			if has != c.want {
				t.Fatalf("gsl_toleranceField emitted %v, want %v", has, c.want)
			}
			if c.want && (!strings.Contains(out.WGSL, "gsl_tolIdx == 1u") || !strings.Contains(out.WGSL, "gsl_tolIdx == 2u")) {
				t.Errorf("gsl_toleranceField lacks a case for each return")
			}
		})
	}
}
