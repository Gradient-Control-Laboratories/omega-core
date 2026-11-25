package build_test

import (
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/build"
)

func TestBuild_NestedConstantList(t *testing.T) {
	res, err := assembler.Assemble([]string{`m = [[1,0,0,0],[0,1,0,0],[0,0,1,0],[3,4,5,1]];
matrices = [m, m, m];
last = matrices[2];
return last;`}, assembler.SourceManager(mockMgr))
	if err != nil {
		t.Fatal(err)
	}
	if err := build.Build(newMockDelegate(), res.Program, false); err != nil {
		t.Fatal(err)
	}
}
