package property

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// The primitives declare their frames, and only a program that reads one pays
// for them: in any other the carry pass erases the declarations, putting a
// function that named its part only to frame it back to returning the part
// itself, so it compiles as it did before frames existed. These check that the
// erasure takes exactly the frame and nothing else.

// frameFn is a function with no parameters whose body declares one local and
// returns ret: form = 1, with the given properties.
func frameFn(name string, props []*program.Property, extra []*program.Assignment, ret *program.Operand) *program.FnType {
	body := append([]*program.Assignment{{
		Variables:  []*program.Variable{{Name: "form", Properties: props}},
		Expression: num(1),
	}}, extra...)
	return &program.FnType{Name: name, Groups: []*program.FnGroup{{
		Functions: []*program.Function{{Assignments: body, Returns: []*program.Operand{ret}}},
	}}}
}

func frameProp(name string) *program.Property {
	return &program.Property{Name: name, Expression: num(2)}
}

// frameProgram is a primary package holding fns, returning ret.
func frameProgram(ret *program.Operand, fns ...*program.FnType) *program.Program {
	return &program.Program{Primary: &program.Package{Files: []*program.PackageFile{{
		FnTypes: fns,
		Returns: []*program.Operand{ret},
	}}}}
}

func fnNamed(prog *program.Program, name string) *program.Function {
	for _, ft := range prog.Primary.Files[0].FnTypes {
		if ft.Name == name {
			return ft.Groups[0].Functions[0]
		}
	}
	return nil
}

func TestFramesAreErasedWhenNoneIsRead(t *testing.T) {
	framed := frameFn("framed", []*program.Property{frameProp(uProperty), frameProp(vProperty)}, nil, ref("form"))
	painted := frameFn("painted", []*program.Property{{Name: colorProperty, Expression: rgbLiteral(1)}, frameProp(uProperty)}, nil, ref("form"))
	named := frameFn("named", nil, nil, ref("form"))
	readAgain := frameFn("readAgain", []*program.Property{frameProp(uProperty)},
		[]*program.Assignment{{Variables: []*program.Variable{{Name: "other"}}, Expression: bin(program.Operation_ADD, ref("form"), num(1))}},
		ref("form"))
	prog := frameProgram(num(0), framed, painted, named, readAgain)

	if err := CarryReturnProperties(prog); err != nil {
		t.Fatalf("CarryReturnProperties: %v", err)
	}

	// Framed and nothing else: the local goes, and the function returns the
	// part itself, exactly as if it had been written return 1.
	if fn := fnNamed(prog, "framed"); len(fn.Assignments) != 0 || !isValue(fn.Returns[0]) {
		t.Errorf("framed: %d assignments, return %v; want none and the part itself", len(fn.Assignments), fn.Returns[0])
	}
	// Still carrying a colour: the local stays, without its frame.
	if fn := fnNamed(prog, "painted"); len(fn.Assignments) != 1 || len(fn.Assignments[0].Variables[0].Properties) != 1 ||
		fn.Assignments[0].Variables[0].Properties[0].Name != colorProperty {
		t.Errorf("painted: want the local kept with its colour and no frame")
	}
	// Named without a frame: as written.
	if fn := fnNamed(prog, "named"); len(fn.Assignments) != 1 || fn.Returns[0].GetReference() == nil {
		t.Errorf("named: a local named without a frame must be left as written")
	}
	// Read elsewhere in the body: the frame goes, the local stays.
	if fn := fnNamed(prog, "readAgain"); len(fn.Assignments) != 2 || len(fn.Assignments[0].Variables[0].Properties) != 0 {
		t.Errorf("readAgain: want the local kept, without its frame, since the body reads it")
	}
}

func TestFramesAreKeptWhereOneIsRead(t *testing.T) {
	framed := frameFn("framed", []*program.Property{frameProp(uProperty)}, nil, ref("form"))
	read := nary(program.Operation_SURFACEU, ref("bar"))
	prog := frameProgram(read, framed)

	if err := CarryReturnProperties(prog); err != nil {
		t.Fatalf("CarryReturnProperties: %v", err)
	}

	fn := fnNamed(prog, "framed")
	if len(fn.Assignments) != 1 || len(fn.Assignments[0].Variables[0].Properties) != 1 {
		t.Fatalf("framed: want the local and its frame kept in a program that reads one")
	}
	if fnNamed(prog, "framed__u") == nil {
		t.Errorf("framed carries no u companion; a call to it would hand its caller no frame")
	}
}

func isValue(op *program.Operand) bool {
	_, ok := op.T.(*program.Operand_Value)
	return ok
}
