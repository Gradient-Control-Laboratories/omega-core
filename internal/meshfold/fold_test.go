package meshfold

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

// TestFoldingGetImportStripsImportStep proves a branch fetched through the
// folding getImport has its importStep declaration and vdbSdf arg rewritten to
// the literal slot — so an agent (e.g. fidget) that re-resolves the branch and
// builds it never sees the importStep builtin.
func TestFoldingGetImportStripsImportStep(t *testing.T) {
	branch := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					assignDecl("importId", importCall("importStep", "f.stp")),
					// castVDB = vdbSdf(importId, 256, 3)
					assignDecl("castVDB", &program.Operand{T: &program.Operand_FnCall{FnCall: &program.FnCall{
						Reference: &program.Reference{Name: "vdbSdf"},
						Arguments: []*program.Operand{
							refOp("importId"),
							{T: &program.Operand_Value{Value: 256}},
							{T: &program.Operand_Value{Value: 3}},
						},
					}}}),
				},
			},
		},
	}
	getImport := func(path string) (*program.Package, error) {
		if path == "branches/x" {
			return branch, nil
		}
		return nil, nil
	}
	_, resolver := meshslots.Assign(&program.Package{
		Files: []*program.PackageFile{
			{
				Imports: []*program.Import{{Path: "branches/x"}},
			},
		},
	}, getImport)
	slot, _ := resolver("branches/x", "step", "f.stp")

	folded, err := FoldingGetImport(getImport, resolver)("branches/x")
	if err != nil {
		t.Fatalf("folding getImport: %v", err)
	}
	// importId decl must be the literal slot, not an importStep call.
	rhs := folded.Files[0].Assignments[0].Expression
	if v, ok := rhs.T.(*program.Operand_Value); !ok || int(v.Value) != slot {
		t.Fatalf("importId should fold to literal slot %d, got %+v", slot, rhs)
	}
	// vdbSdf index arg must be the literal slot too.
	vdbArg := folded.Files[0].Assignments[1].Expression.GetFnCall().Arguments[0]
	if v, ok := vdbArg.T.(*program.Operand_Value); !ok || int(v.Value) != slot {
		t.Fatalf("vdbSdf arg should fold to literal slot %d, got %+v", slot, vdbArg)
	}
	// The original branch must be untouched (still an importStep call).
	if _, isVal := branch.Files[0].Assignments[0].Expression.T.(*program.Operand_Value); isVal {
		t.Fatalf("original branch was mutated")
	}
}

// TestFoldMeshIndices proves the fold-to-literal pass rewrites mesh(partId)
// into mesh(<slot>) so downstream literal readers keep working, while leaving a
// legacy numbered mesh(7) untouched.
func TestFoldMeshIndices(t *testing.T) {
	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					assignDecl("partId", importCall("importStep", "part.stp")),
				},
				Returns: []*program.Operand{
					// mesh(partId) -> should fold to mesh(<slot>)
					{T: &program.Operand_FnCall{FnCall: &program.FnCall{
						Reference: &program.Reference{Name: "mesh"}, Arguments: []*program.Operand{refOp("partId")},
					}}},
					// mesh(7) -> legacy literal, untouched
					{T: &program.Operand_FnCall{FnCall: &program.FnCall{
						Reference: &program.Reference{Name: "mesh"}, Arguments: []*program.Operand{{T: &program.Operand_Value{Value: 7}}},
					}}},
				},
			},
		},
	}
	_, resolver := meshslots.Assign(pkg, nil)
	slot, _ := resolver("", "step", "part.stp")

	folded := FoldMeshIndices(pkg, nil, resolver)

	arg0 := folded.Files[0].Returns[0].GetFnCall().Arguments[0]
	v, ok := arg0.T.(*program.Operand_Value)
	if !ok || int(v.Value) != slot {
		t.Fatalf("mesh(partId) arg should fold to literal slot %d, got %+v", slot, arg0)
	}
	legacy := folded.Files[0].Returns[1].GetFnCall().Arguments[0]
	lv, ok := legacy.T.(*program.Operand_Value)
	if !ok || int(lv.Value) != 7 {
		t.Fatalf("legacy mesh(7) must stay literal 7, got %+v", legacy)
	}
	// The importStep declaration itself must fold to a literal so the build
	// never sees the importStep builtin: partId = importStep "f" -> partId = slot.
	declRHS := folded.Files[0].Assignments[0].Expression
	dv, ok := declRHS.T.(*program.Operand_Value)
	if !ok || int(dv.Value) != slot {
		t.Fatalf("importStep decl should fold to literal slot %d, got %+v", slot, declRHS)
	}
	// The clone must not mutate the original (still a reference).
	if _, isVal := pkg.Files[0].Returns[0].GetFnCall().Arguments[0].T.(*program.Operand_Value); isVal {
		t.Fatalf("original package was mutated")
	}
}
