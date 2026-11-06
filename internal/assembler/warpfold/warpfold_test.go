package warpfold

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

type mockWarpResolver struct {
	resolvedSlot int
	lastDir      string
	lastPath     string
}

func (m *mockWarpResolver) ResolveWarpedSlot(meshSlot int, pkgDir, warpPath string, scale float64) (int, error) {
	m.lastDir = pkgDir
	m.lastPath = warpPath
	return m.resolvedSlot, nil
}

func TestWarpFoldMeshWarped(t *testing.T) {
	resolver := &mockWarpResolver{resolvedSlot: 42}

	prog := &program.Program{
		Primary: &program.Package{
			Path: "pkg_primary",
			Files: []*program.PackageFile{
				{
					WarpImports: []*program.WarpImport{
						{Index: 0, Path: "run.warp"},
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

	folded, warpedAssets, err := FoldProgram(prog, resolver)
	if err != nil {
		t.Fatalf("FoldProgram failed: %v", err)
	}
	if resolver.lastDir != "pkg_primary" {
		t.Errorf("expected ResolveWarpedSlot pkgDir = \"pkg_primary\", got %q", resolver.lastDir)
	}
	if len(warpedAssets) != 1 {
		t.Fatalf("expected 1 WarpedAsset, got %d", len(warpedAssets))
	}
	wantAsset := WarpedAsset{
		DerivedSlot: 42,
		MeshSlot:    0,
		WarpDir:     "pkg_primary",
		WarpPath:    "run.warp",
		Scale:       -1.0,
	}
	if warpedAssets[0] != wantAsset {
		t.Errorf("warpedAssets[0] = %+v, want %+v", warpedAssets[0], wantAsset)
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

func TestGraphPackages(t *testing.T) {
	dep := &program.Package{
		Path: "branches/b1",
		Files: []*program.PackageFile{{
			WarpImports: []*program.WarpImport{{Index: 0, Path: "branch.warp"}},
		}},
	}
	root := &program.Package{
		Files: []*program.PackageFile{{
			Imports: []*program.Import{{Path: "branches/b1"}},
		}},
	}
	pkgs := GraphPackages(func(path string) (*program.Package, error) {
		if path == "branches/b1" {
			return dep, nil
		}
		return nil, nil
	}, root, nil)
	if len(pkgs) != 1 || pkgs[0].Path != "branches/b1" {
		t.Fatalf("GraphPackages = %+v, want [branches/b1]", pkgs)
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

func stringExprOperand(s string) *program.Operand {
	ops := make([]*program.Operand, len(s))
	for i, r := range s {
		ops[i] = &program.Operand{T: &program.Operand_Value{Value: float64(r)}}
	}
	return &program.Operand{
		T: &program.Operand_Expression{
			Expression: &program.Expression{
				Operation: program.Operation_COLLECT,
				Operands:  ops,
			},
		},
	}
}

func importWarpExprPkg(paths ...string) *program.Package {
	var assigns []*program.Assignment
	for i, p := range paths {
		assigns = append(assigns, &program.Assignment{
			Variables: []*program.Variable{{Name: string(rune('a' + i))}},
			Expression: &program.Operand{
				T: &program.Operand_FnCall{
					FnCall: &program.FnCall{
						Reference: &program.Reference{Name: "importWarp"},
						Arguments: []*program.Operand{stringExprOperand(p)},
					},
				},
			},
		})
	}
	return &program.Package{
		Files: []*program.PackageFile{{Assignments: assigns}},
	}
}

func TestAssignWarpSlotsBothImportForms(t *testing.T) {
	numbered := &program.Package{
		Files: []*program.PackageFile{{
			WarpImports: []*program.WarpImport{{Index: 0, Path: "run.warp"}},
		}},
	}
	w := AssignWarpSlots([]WarpGraphPkg{{Path: "", Pkg: numbered}}, "trees/T1")
	if got, ok := w.Ref(0); !ok || got.Path != "run.warp" || got.Dir != "trees/T1" {
		t.Fatalf("numbered form: slot 0 = %+v (ok=%v), want run.warp in trees/T1", got, ok)
	}
	if len(w.ExprCalls) != 0 {
		t.Errorf("numbered form should produce no expression calls, got %d", len(w.ExprCalls))
	}

	expr := importWarpExprPkg("run.warp")
	w2 := AssignWarpSlots([]WarpGraphPkg{{Path: "", Pkg: expr}}, "trees/T1")
	if got, ok := w2.Ref(0); !ok || got.Path != "run.warp" {
		t.Fatalf("expression form: slot 0 = %+v (ok=%v)", got, ok)
	}
	if len(w2.ExprCalls) != 1 {
		t.Fatalf("expression form should leave one call to rewrite, got %d", len(w2.ExprCalls))
	}
	if len(w2.ExprByPkg[""]) != 1 {
		t.Errorf("the call should be attributed to the document, got %v", w2.ExprByPkg)
	}
}

func TestAssignWarpSlotsSeparatesIdenticalFilenames(t *testing.T) {
	pkgs := []WarpGraphPkg{
		{Path: "", Pkg: importWarpExprPkg("part.warp")},
		{Path: "branches/A", Pkg: importWarpExprPkg("part.warp")},
		{Path: "branches/B", Pkg: importWarpExprPkg("part.warp")},
	}
	w := AssignWarpSlots(pkgs, "trees/T1")

	if len(w.BySlot) != 3 {
		t.Fatalf("three packages naming part.warp should get three slots, got %d: %+v",
			len(w.BySlot), w.BySlot)
	}
	dirs := map[string]bool{}
	for _, ref := range w.BySlot {
		if ref.Path != "part.warp" {
			t.Errorf("unexpected path %q", ref.Path)
		}
		if dirs[ref.Dir] {
			t.Errorf("directory %q used for two slots", ref.Dir)
		}
		dirs[ref.Dir] = true
	}
	for _, want := range []string{"trees/T1", "branches/A", "branches/B"} {
		if !dirs[want] {
			t.Errorf("no slot resolves against %q", want)
		}
	}
	for _, p := range []string{"", "branches/A", "branches/B"} {
		if len(w.ExprByPkg[p]) != 1 {
			t.Errorf("package %q has %d calls to rewrite, want 1", p, len(w.ExprByPkg[p]))
		}
	}
}

func TestAssignWarpSlotsReusesOneFile(t *testing.T) {
	pkg := importWarpExprPkg("run.warp", "run.warp")
	w := AssignWarpSlots([]WarpGraphPkg{{Path: "", Pkg: pkg}}, "trees/T1")
	if len(w.BySlot) != 1 {
		t.Fatalf("one file should occupy one slot, got %d: %+v", len(w.BySlot), w.BySlot)
	}
	if len(w.ExprCalls) != 2 {
		t.Fatalf("both calls should still be rewritten, got %d", len(w.ExprCalls))
	}
	for _, slot := range w.ExprCalls {
		if slot != 0 {
			t.Errorf("call resolved to slot %d, want 0", slot)
		}
	}
}

func TestAssignWarpSlotsRenumbersImportedNumberedForms(t *testing.T) {
	doc := &program.Package{
		Files: []*program.PackageFile{{
			WarpImports: []*program.WarpImport{{Index: 0, Path: "tree.warp"}},
		}},
	}
	branch := &program.Package{
		Files: []*program.PackageFile{{
			WarpImports: []*program.WarpImport{{Index: 0, Path: "branch.warp"}},
		}},
	}
	w := AssignWarpSlots([]WarpGraphPkg{
		{Path: "", Pkg: doc},
		{Path: "branches/A", Pkg: branch},
	}, "trees/T1")

	if len(w.BySlot) != 2 {
		t.Fatalf("want two slots, got %+v", w.BySlot)
	}
	if got, ok := w.Ref(0); !ok || got.Path != "tree.warp" {
		t.Errorf("the document keeps slot 0, got %+v", got)
	}
	found := false
	for slot, ref := range w.BySlot {
		if ref.Path == "branch.warp" {
			found = true
			if slot == 0 {
				t.Errorf("the branch's import collided with the document's slot 0")
			}
			if ref.Dir != "branches/A" {
				t.Errorf("branch warp resolves against %q, want branches/A", ref.Dir)
			}
		}
	}
	if !found {
		t.Error("the branch's simulation never got a slot")
	}
}

type errWarpResolver struct {
	err error
}

func (e *errWarpResolver) ResolveWarpedSlot(meshSlot int, pkgDir, warpPath string, scale float64) (int, error) {
	return 0, e.err
}

func TestFoldProgram_NestedFunctionsAndErrors(t *testing.T) {
	resolver := &mockWarpResolver{resolvedSlot: 99}

	wEmpty := &WarpSlots{}
	if !wEmpty.Empty() {
		t.Errorf("expected empty WarpSlots to report Empty() == true")
	}

	nestedCall := func(name string, args ...*program.Operand) *program.Operand {
		return &program.Operand{
			T: &program.Operand_FnCall{
				FnCall: &program.FnCall{
					Reference: &program.Reference{Name: name},
					Arguments: args,
				},
			},
		}
	}
	valOp := func(v float64) *program.Operand {
		return &program.Operand{T: &program.Operand_Value{Value: v}}
	}

	dep := &program.Package{
		Path: "branches/dep",
		Files: []*program.PackageFile{{
			WarpImports: []*program.WarpImport{{Index: 0, Path: "dep.warp"}},
			FnTypes: []*program.FnType{{
				Name: "helper",
				Groups: []*program.FnGroup{{
					ParamCount: 0,
					Functions: []*program.Function{{
						Assignments: []*program.Assignment{{
							Variables:  []*program.Variable{{Name: "inner"}},
							Expression: nestedCall("vdbSdfWarped", valOp(0), valOp(0), valOp(1.0), valOp(64), valOp(5)),
						}},
						FnTypes: []*program.FnType{{
							Name: "subHelper",
							Groups: []*program.FnGroup{{
								ParamCount: 0,
								Functions: []*program.Function{{
									Returns: []*program.Operand{nestedCall("warpField", valOp(0))},
								}},
							}},
						}},
						Returns: []*program.Operand{nestedCall("meshWarped", valOp(0), valOp(0), valOp(1.0))},
					}},
				}},
			}},
			Returns: []*program.Operand{
				{
					T: &program.Operand_Expression{
						Expression: &program.Expression{
							Operation: program.Operation_ADD,
							Operands: []*program.Operand{
								nestedCall("vdbSdfWarped", valOp(0), valOp(0), valOp(1.0)),
								nestedCall("warpField", valOp(0)),
							},
						},
					},
				},
			},
			ColorReturns:    []*program.Operand{nestedCall("warpField", valOp(0))},
			MaterialReturns: []*program.Operand{nestedCall("warpField", valOp(0))},
		}},
	}

	prog := &program.Program{
		Primary: &program.Package{
			Path:  "trees/main",
			Files: []*program.PackageFile{{}},
		},
		Dependencies: []*program.Package{{Path: "builtin"}, dep},
	}

	folded, assets, err := FoldProgram(prog, resolver)
	if err != nil {
		t.Fatalf("FoldProgram failed: %v", err)
	}
	if len(assets) != 1 || assets[0].DerivedSlot != 99 {
		t.Fatalf("assets = %+v, want [DerivedSlot: 99]", assets)
	}

	foldedZero, err := FoldWarpFieldToZero(folded)
	if err != nil {
		t.Fatalf("FoldWarpFieldToZero failed: %v", err)
	}
	if HasWarpCall(foldedZero.Primary) {
		t.Errorf("expected no warp calls in Primary")
	}

	// Error paths in rewrite
	errCases := []struct {
		name string
		args []*program.Operand
	}{
		{"too few args", []*program.Operand{valOp(0), valOp(0)}},
		{"negative mesh id", []*program.Operand{valOp(-1), valOp(0), valOp(1)}},
		{"negative warp id", []*program.Operand{valOp(0), valOp(-1), valOp(1)}},
		{"non-const scale", []*program.Operand{valOp(0), valOp(0), &program.Operand{T: &program.Operand_Reference{Reference: &program.Reference{Name: "p"}}}}},
		{"unmapped warp slot", []*program.Operand{valOp(0), valOp(7), valOp(1)}},
	}
	for _, tc := range errCases {
		badProg := &program.Program{
			Primary: &program.Package{
				Files: []*program.PackageFile{{
					WarpImports: []*program.WarpImport{{Index: 0, Path: "sim.warp"}},
					Returns:     []*program.Operand{nestedCall("meshWarped", tc.args...)},
				}},
			},
			Dependencies: []*program.Package{{Path: "builtin"}},
		}
		if _, _, err := FoldProgram(badProg, resolver); err == nil {
			t.Errorf("case %q: expected error, got nil", tc.name)
		}
	}
}
