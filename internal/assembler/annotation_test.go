package lib

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/annotation"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/program"
)

// findAnnotation returns the recorded annotation for symbol anywhere in the
// assembled program, so the test asserts what survived the whole pipeline
// rather than what the parser wrote.
func findAnnotation(prog *program.Program, symbol string) *program.Annotation {
	if prog == nil {
		return nil
	}
	pkgs := append([]*program.Package{prog.Primary}, prog.Dependencies...)
	for _, pkg := range pkgs {
		if pkg == nil {
			continue
		}
		for _, file := range pkg.Files {
			if file == nil {
				continue
			}
			for _, ann := range file.Annotations {
				if ann != nil && ann.Symbol == symbol {
					return ann
				}
			}
		}
	}
	return nil
}

func TestAnnotationSurvivesAssembly(t *testing.T) {
	// Several passes rebuild PackageFile field by field — CloneFile and the
	// deadcode rebuild among them — so a new field is easy to drop silently.
	// The browser reads annotations off an assembled program, so assert they are
	// still there after the full assemble, not just after the parse.
	pkgMgr := source.NewLocalManager("../../../common/source/package")
	src := `
@material brass() {
  return material.make(0.3, 1, 1.5), [0.95, 0.90, 0.69];
}

boltMat, boltColor = brass();
bolt = sphere(3);
bolt.material = boltMat;
bolt.color = boltColor;
return bolt;
`
	res, err := Assemble([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ann := findAnnotation(res.Program, "brass")
	if ann == nil {
		t.Fatalf("the @material annotation on brass did not survive assembly")
	}
	if ann.Kind != annotation.Material {
		t.Errorf("got kind %q, want %q", ann.Kind, annotation.Material)
	}
	if ann.Start == nil || ann.Start.Line != 2 {
		t.Errorf("got start %+v, want line 2", ann.Start)
	}
}

func TestAnnotationSurvivesWhenItsSymbolIsUnused(t *testing.T) {
	// An unused material is deadcoded out of the program, but its annotation
	// still describes the source: a browser lists every material a library
	// declares, not only the ones this document draws with.
	pkgMgr := source.NewLocalManager("../../../common/source/package")
	src := `
@material brass() {
  return material.make(0.3, 1, 1.5), [0.95, 0.90, 0.69];
}
return sphere(3);
`
	res, err := Assemble([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if findAnnotation(res.Program, "brass") == nil {
		t.Fatalf("the @material annotation on an unused brass was dropped")
	}
}

func TestAssembleRejectsMaterialConstant(t *testing.T) {
	// The assembler is where the pair contract is enforced, so it holds for
	// documents and libraries alike: a constant can carry the record or the
	// colour, never both.
	pkgMgr := source.NewLocalManager("../../../common/source/package")
	src := `
@material brass = material.make(0.3, 1, 1.5);
brassColor = [0.95, 0.90, 0.69];
return sphere(3);
`
	_, err := Assemble([]string{src}, pkgMgr)
	if err == nil {
		t.Fatalf("expected a constant-is-not-a-material error, got nil")
	}
	if !strings.Contains(err.Error(), "is a constant") {
		t.Fatalf("expected the error to say the symbol is a constant, got: %v", err)
	}
}

func TestAssembleRejectsMaterialReturningOneValue(t *testing.T) {
	// The shape half of the contract, through the whole pipeline: the check
	// runs after type analysis, so this is where it is reachable.
	pkgMgr := source.NewLocalManager("../../../common/source/package")
	src := `
@material brass() {
  return material.make(0.3, 1, 1.5);
}
bolt = sphere(3);
bolt.material = brass();
return bolt;
`
	_, err := Assemble([]string{src}, pkgMgr)
	if err == nil {
		t.Fatalf("expected a wrong-arity error, got nil")
	}
	if !strings.Contains(err.Error(), "want 2") {
		t.Fatalf("expected the error to ask for two returns, got: %v", err)
	}
}
