package warpfold

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

type mockWarpResolver struct {
	resolvedSlot int
}

func (m *mockWarpResolver) ResolveWarpedSlot(meshSlot int, pkgDir, warpPath string, scale float64) (int, error) {
	return m.resolvedSlot, nil
}

func TestWarpFoldMeshWarped(t *testing.T) {
	resolver := &mockWarpResolver{resolvedSlot: 42}

	prog := &program.Program{
		Primary: &program.Package{
			Files: []*program.PackageFile{
				{
					WarpImports: []*program.WarpImport{
						{Index: 0, Path: "run.castsim"},
					},
					Assignments: []*program.Assignment{
						{
							Variables: []*program.Variable{{Name: "msh"}},
							Expression: &program.Operand{
								T: &program.Operand_FnCall{
									FnCall: &program.FnCall{
										Reference: &program.Reference{Name: "meshWarped"},
										Arguments: []*program.Operand{
											{T: &program.Operand_Value{Value: 0}},    // mesh slot 0
											{T: &program.Operand_Value{Value: 0}},    // warp slot 0
											{T: &program.Operand_Value{Value: -1.0}}, // scale -1.0
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Dependencies: []*program.Package{
			{Path: "builtin"},
		},
	}

	if !HasWarpCall(prog.Primary) {
		t.Fatalf("expected HasWarpCall to be true")
	}

	folded, err := FoldProgram(prog, resolver)
	if err != nil {
		t.Fatalf("FoldProgram failed: %v", err)
	}

	assign := folded.Primary.Files[0].Assignments[0]
	fc := assign.Expression.GetFnCall()
	if fc == nil {
		t.Fatalf("expected function call, got %v", assign.Expression)
	}
	if fc.Reference.Name != "mesh" {
		t.Fatalf("expected meshWarped to be rewritten to mesh, got %s", fc.Reference.Name)
	}
	if len(fc.Arguments) != 1 || fc.Arguments[0].GetValue() != 42 {
		t.Fatalf("expected derived slot 42, got %v", fc.Arguments)
	}
}

func TestFoldWarpFieldToZero(t *testing.T) {
	prog := &program.Program{
		Primary: &program.Package{
			Files: []*program.PackageFile{
				{
					Assignments: []*program.Assignment{
						{
							Variables: []*program.Variable{{Name: "wf"}},
							Expression: &program.Operand{
								T: &program.Operand_FnCall{
									FnCall: &program.FnCall{
										Reference: &program.Reference{Name: "warpField"},
										Arguments: []*program.Operand{
											{T: &program.Operand_Value{Value: 1.0}},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	folded, err := FoldWarpFieldToZero(prog)
	if err != nil {
		t.Fatalf("FoldWarpFieldToZero failed: %v", err)
	}

	assign := folded.Primary.Files[0].Assignments[0]
	expr := assign.Expression.GetExpression()
	if expr == nil || expr.Operation != program.Operation_COLLECT || len(expr.Operands) != 3 {
		t.Fatalf("expected [0,0,0] vector, got %v", assign.Expression)
	}
	for i, op := range expr.Operands {
		if op.GetValue() != 0 {
			t.Fatalf("expected component %d to be 0, got %v", i, op.GetValue())
		}
	}
}
