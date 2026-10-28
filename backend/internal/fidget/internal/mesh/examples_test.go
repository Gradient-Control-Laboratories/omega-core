package mesh

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
	"gradientcontrol.com/omega/core/program/example"
)

func TestExamples(t *testing.T) {
	runTest(t, "Sphere", example.Sphere())
	runTest(t, "Cube", example.Cube())
	runTest(t, "Container", example.Container())
}

func runTest(t *testing.T, name string, progPkg *program.Package) {
	prog := &program.Program{
		Primary: progPkg,
	}
	_, _, err := GenerateRender(prog)
	if err != nil {
		t.Errorf("[%s] Fidget compilation error: %v", name, err)
	} else {
		t.Logf("[%s] SUCCESS!", name)
	}
}
