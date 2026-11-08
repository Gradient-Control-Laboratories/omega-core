package meshslots

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

func runeList(s string) *program.Operand {
	rs := []rune(s)
	ops := make([]*program.Operand, len(rs))
	for i, r := range rs {
		ops[i] = &program.Operand{T: &program.Operand_Value{Value: float64(r)}}
	}
	return &program.Operand{T: &program.Operand_Expression{Expression: &program.Expression{
		Operation: program.Operation_COLLECT, Operands: ops,
	}}}
}

func importCall(name, path string) *program.Operand {
	return &program.Operand{T: &program.Operand_FnCall{FnCall: &program.FnCall{
		Reference: &program.Reference{Name: name},
		Arguments: []*program.Operand{runeList(path)},
	}}}
}

func assignDecl(name string, body *program.Operand) *program.Assignment {
	return &program.Assignment{
		Variables:  []*program.Variable{{Name: name}},
		Expression: body,
	}
}

func TestAssignTableAndKeys(t *testing.T) {
	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					assignDecl("a", importCall("importStep", "a.stp")),
					assignDecl("b", importCall("importMesh", "b.stl")),
					assignDecl("a2", importCall("importStep", "a.stp")), // dup -> shares a's slot
				},
			},
		},
	}
	table, resolver := Assign(pkg, nil)
	if len(table) != 2 {
		t.Fatalf("expected 2 entries, got %d: %+v", len(table), table)
	}
	if table[0].Kind != "step" || table[0].Path != "a.stp" || table[0].PkgPath != "" {
		t.Fatalf("entry 0 = %+v", table[0])
	}
	if table[1].Kind != "mesh" || table[1].Path != "b.stl" {
		t.Fatalf("entry 1 = %+v", table[1])
	}
	if s, ok := resolver("", "step", "a.stp"); !ok || s != table[0].Slot {
		t.Fatalf("resolver step a.stp = %d ok=%v, want %d", s, ok, table[0].Slot)
	}
	if _, ok := resolver("", "step", "missing.stp"); ok {
		t.Fatalf("resolver should miss for unknown path")
	}
}

func TestAssignBaseOffsetAboveNumbered(t *testing.T) {
	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				StepImports: []*program.StepImport{{Index: 7, Path: "legacy.stp"}},
				Assignments: []*program.Assignment{assignDecl("a", importCall("importStep", "new.stp"))},
			},
		},
	}
	table, _ := Assign(pkg, nil)
	if len(table) != 1 || table[0].Slot != 8 {
		t.Fatalf("expected slot 8 (above numbered 7), got %+v", table)
	}
}
