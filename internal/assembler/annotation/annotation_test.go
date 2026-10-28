package annotation_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/annotation"
	"gradientcontrol.com/omega/core/internal/collect"
	"gradientcontrol.com/omega/core/internal/link"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/types"
	parser "gradientcontrol.com/omega/core/parser"
)

func parseAndValidate(t *testing.T, src string) error {
	t.Helper()
	srcMgr := source.NewLocalManager("../../../common/source/package")
	prog, err := collect.Program([]string{src}, srcMgr, parser.ParsePackage)
	if err != nil {
		return err
	}
	if err := link.References(prog, srcMgr, parser.ParsePackage); err != nil {
		return err
	}
	return annotation.Validate(prog)
}

// parseAndValidateReturns runs the pair-shape check, which needs resolved
// types and so runs after types.Analyze rather than beside Validate.
func parseAndValidateReturns(t *testing.T, src string) error {
	t.Helper()
	srcMgr := source.NewLocalManager("../../../common/source/package")
	prog, err := collect.Program([]string{src}, srcMgr, parser.ParsePackage)
	if err != nil {
		return err
	}
	if err := link.References(prog, srcMgr, parser.ParsePackage); err != nil {
		return err
	}
	if err := annotation.Validate(prog); err != nil {
		return err
	}
	if err := types.Analyze(prog); err != nil {
		return err
	}
	return annotation.ValidateReturns(prog)
}

func TestMaterialFunctionReturningThePair(t *testing.T) {
	// The form every metal, glass and physical member takes: the record and
	// the colour leave together, so a consumer cannot take one without the
	// other.
	src := `
@material brass() {
  return material.make(0.3, 1, 1.5), [0.95, 0.90, 0.69];
}

boltMat, boltColor = brass();
bolt = sphere(1);
bolt.material = boltMat;
bolt.color = boltColor;
return bolt;
`
	if err := parseAndValidateReturns(t, src); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestMaterialConstantIsRejected(t *testing.T) {
	// A material record carries no albedo, so a constant can hold only half of
	// a material. The colour would have to live under some other name, which is
	// the naming convention this contract replaced.
	src := `
@material brass = material.make(0.3, 1, 1.5);
brassColor = [0.95, 0.90, 0.69];
return sphere(1);
`
	err := parseAndValidate(t, src)
	if err == nil {
		t.Fatalf("expected a constant-is-not-a-material error, got nil")
	}
	if !strings.Contains(err.Error(), "is a constant") {
		t.Fatalf("expected the error to say the symbol is a constant, got: %v", err)
	}
	if !strings.Contains(err.Error(), "@finish") {
		t.Fatalf("expected the error to point at @finish for a colourless record, got: %v", err)
	}
}

func TestMaterialFunctionWithParametersIsAccepted(t *testing.T) {
	// A material/* family preset takes the point's frame and so cannot be
	// typed without a call site. Validate accepts it; ValidateReturns
	// deliberately leaves it alone (see its doc comment).
	src := `
@material walnut(axisOrigin, axisDir) {
  return material.make(0.4, 0, 1.5), [0.4, 0.26, 0.13];
}

mat, col = walnut([0, 0, 0], [0, 0, 1]);
board = sphere(5);
board.material = mat;
board.color = col;
return board;
`
	if err := parseAndValidateReturns(t, src); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestMaterialReturningOnlyTheRecordIsRejected(t *testing.T) {
	// The arity half of the pair contract.
	src := `
@material brass() {
  return material.make(0.3, 1, 1.5);
}

bolt = sphere(1);
bolt.material = brass();
return bolt;
`
	err := parseAndValidateReturns(t, src)
	if err == nil {
		t.Fatalf("expected a wrong-arity error, got nil")
	}
	if !strings.Contains(err.Error(), "want 2") {
		t.Fatalf("expected the error to ask for two returns, got: %v", err)
	}
}

func TestMaterialWithSwappedReturnsIsRejected(t *testing.T) {
	// The mistake the naming convention could never catch: both halves are
	// present and correctly spelled, in the wrong order.
	src := `
@material brass() {
  return [0.95, 0.90, 0.69], material.make(0.3, 1, 1.5);
}

boltMat, boltColor = brass();
bolt = sphere(1);
bolt.material = boltMat;
bolt.color = boltColor;
return bolt;
`
	err := parseAndValidateReturns(t, src)
	if err == nil {
		t.Fatalf("expected a wrong-shape error, got nil")
	}
	if !strings.Contains(err.Error(), "first return") {
		t.Fatalf("expected the error to name the first return, got: %v", err)
	}
}

func TestFinishIsAConstantRecord(t *testing.T) {
	// The material library's own presets carry no albedo on purpose — its
	// header says so — which is why they are a separate kind rather than a
	// @material the contract would reject.
	src := `
@finish matte = material.make(1.0, 0, 1.45);

cup = sphere(3);
cup.material = matte;
return cup;
`
	if err := parseAndValidate(t, src); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestSurfacePresetIsAnOverloadOfItsOperator(t *testing.T) {
	// An operator preset takes the slots a document supplies and bakes the
	// rest, as an overload of the operator's own name. Validate accepts it,
	// and ValidateReturns leaves it alone: it has parameters, and the pair
	// contract is a @material's, not a @surface's.
	src := `
@surface roughened(m, @frame frame) {
  r = roughened(m, frame, 0.5);
  return r;
}
roughened(m, frame, rough) {
  return material.withRoughness(m, rough + 0 * frame.x);
}

cup = sphere(3);
cup.material = roughened(material.matte, [p.x, p.y]);
return cup;
`
	if err := parseAndValidateReturns(t, src); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestUnannotatedProgramIsUnaffected(t *testing.T) {
	// Validate must be a no-op on every program that carries no annotation,
	// which is every program written before this one.
	src := `
brass = material.make(0.3, 1, 1.5);
return sphere(1);
`
	if err := parseAndValidate(t, src); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}
