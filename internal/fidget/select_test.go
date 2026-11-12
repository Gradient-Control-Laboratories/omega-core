package fidget

import (
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/source"
)

func TestSelectBranchOrder(t *testing.T) {
	res, err := assembler.Assemble([]string{`return p.x ? p.y + 10 : p.y + 20 : p.y + 30;`}, assembler.SourceManager(&source.MockManager{}))
	if err != nil {
		t.Fatal(err)
	}
	Lock()
	defer Unlock()
	Clear()
	scope := NewScope(nil)
	if err := scope.EvaluatePackage(res.Program.Primary); err != nil {
		t.Fatal(err)
	}
	files := res.Program.Primary.Files
	trees, _, err := scope.EvaluateReturns(files[len(files)-1].Returns)
	if err != nil {
		t.Fatal(err)
	}
	values, err := EvalGrad(BuildCanvas(trees[0]), []float32{1, 0, -1}, []float32{2, 2, 2}, []float32{0, 0, 0})
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range []float32{12, 22, 32} {
		if values[i].Value != want {
			t.Fatalf("branch %d: %g, want %g", i, values[i].Value, want)
		}
	}
}
