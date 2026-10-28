package eval

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/meshslots"
	"gradientcontrol.com/omega/core/program"
)

// runeList builds the rune-list operand the parser lowers a string literal to.
func runeList(s string) *program.Operand {
	rs := []rune(s)
	if len(rs) == 1 {
		return &program.Operand{T: &program.Operand_Value{Value: float64(rs[0])}}
	}
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

func refOp(name string) *program.Operand {
	return &program.Operand{T: &program.Operand_Reference{Reference: &program.Reference{Name: name}}}
}

func qrefOp(prefix, name string) *program.Operand {
	return &program.Operand{T: &program.Operand_Reference{Reference: &program.Reference{Prefix: prefix, Name: name}}}
}

func foldScalar(t *testing.T, op *program.Package, getImport program.GetImportFunc, slots meshslots.Resolver, e *program.Operand) int {
	t.Helper()
	vals, err := ConstPkgSlots(e, op, getImport, slots)
	if err != nil {
		t.Fatalf("fold error: %v", err)
	}
	if len(vals) != 1 {
		t.Fatalf("expected scalar, got %v", vals)
	}
	return int(vals[0])
}

// TestImportSlotFoldsAcrossImports proves that `importStep "f"` folds to a
// compiler-assigned slot, that a tree gets each imported branch's slot through
// the branch's returned id, that the same relative filename in two different
// packages gets distinct slots (package-relative keying), and that the same
// file referenced twice within one package shares a slot.
func TestImportSlotFoldsAcrossImports(t *testing.T) {
	branchA := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{assignDecl("aId", importCall("importStep", "part.stp"))},
				Returns:     []*program.Operand{refOp("aId")},
			},
		},
	}
	branchB := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{assignDecl("bId", importCall("importStep", "part.stp"))},
				Returns:     []*program.Operand{refOp("bId")},
			},
		},
	}
	tree := &program.Package{
		Files: []*program.PackageFile{
			{
				Imports: []*program.Import{
					{Prefix: "a", Path: "branchA"},
					{Prefix: "b", Path: "branchB"},
				},
				Assignments: []*program.Assignment{
					// Same filename as branchA's import, but in the root package -> distinct slot.
					assignDecl("x", importCall("importStep", "part.stp")),
					// Same file again within the root package -> must share x's slot.
					assignDecl("dup", importCall("importStep", "part.stp")),
				},
			},
		},
	}

	getImport := func(path string) (*program.Package, error) {
		switch path {
		case "branchA":
			return branchA, nil
		case "branchB":
			return branchB, nil
		}
		return nil, nil
	}

	table, resolver := meshslots.Assign(tree, getImport)
	if len(table) != 3 {
		t.Fatalf("expected 3 package-local asset imports, got %d: %+v", len(table), table)
	}

	slotA := foldScalar(t, tree, getImport, resolver, qrefOp("a", "aId"))
	slotB := foldScalar(t, tree, getImport, resolver, qrefOp("b", "bId"))
	slotX := foldScalar(t, tree, getImport, resolver, refOp("x"))
	slotDup := foldScalar(t, tree, getImport, resolver, refOp("dup"))

	if slotX != slotDup {
		t.Fatalf("same file must share a slot: x=%d dup=%d", slotX, slotDup)
	}
	if slotA == slotX || slotA == slotB || slotB == slotX {
		t.Fatalf("same filename in distinct packages must use distinct slots: a=%d b=%d x=%d", slotA, slotB, slotX)
	}

	if got, _ := resolver("branchA", "step", "part.stp"); got != slotA {
		t.Fatalf("resolver(branchA,step,part.stp)=%d, folded aId=%d", got, slotA)
	}
}

// TestImportSlotInList folds a list literal of ids — the shape a tree uses to
// map mesh() placements over a returned id list.
func TestImportSlotInList(t *testing.T) {
	tree := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					assignDecl("p0", importCall("importStep", "p0.stp")),
					assignDecl("p1", importCall("importMesh", "p1.stl")),
				},
				Returns: []*program.Operand{
					{T: &program.Operand_Expression{Expression: &program.Expression{
						Operation: program.Operation_COLLECT,
						Operands:  []*program.Operand{refOp("p0"), refOp("p1")},
					}}},
				},
			},
		},
	}
	_, resolver := meshslots.Assign(tree, nil)
	vals, err := ConstPkgSlots(tree.Files[0].Returns[0], tree, nil, resolver)
	if err != nil {
		t.Fatalf("fold list: %v", err)
	}
	if len(vals) != 2 || vals[0] == vals[1] {
		t.Fatalf("expected two distinct slot ids in list, got %v", vals)
	}
}
