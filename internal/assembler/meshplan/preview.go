package meshplan

import "gradientcontrol.com/omega/core/program"

// PreviewBody is one top-level return as the Shapes panel lists it, sent with
// the preview when the program returns meshes. The shader numbers its slots by
// SDF leaf while mesh() placements carry their return index, so the panel uses
// this map to show one row per return. Slots are the 1-based shader slots the
// return draws; Mesh indicates whether it places an imported mesh.
type PreviewBody struct {
	Name  string `json:"name"`
	Slots []int  `json:"slots,omitempty"`
	Mesh  bool   `json:"mesh,omitempty"`
}

// PreviewBodies lists the plan's top-level returns for the Shapes panel of a
// preview that draws meshes. The preview shader returns one slot per SDF
// subtree, in subtree order, so slot i+1 belongs to the return whose plan
// holds subtrees[i]; a return that places an imported mesh is marked so the
// panel can show and hide it too.
func PreviewBodies(plan *MeshExportPlan, subtrees []SdfSubtree) []PreviewBody {
	if plan == nil || plan.Root == nil {
		return nil
	}
	leaves := make(map[string]*MeshPlanNode)
	IndexSdfLeaves(plan.Root, leaves)
	slots := make(map[int][]int)
	for i, s := range subtrees {
		if leaf := leaves[s.ID]; leaf != nil {
			slots[leaf.ShapeIndex] = append(slots[leaf.ShapeIndex], i+1)
		}
	}
	var bodies []PreviewBody
	for _, b := range plan.Bodies() {
		bodies = append(bodies, PreviewBody{
			Name:  b.Name,
			Slots: slots[b.ShapeIndex],
			Mesh:  placesMesh(b.Root),
		})
	}
	return bodies
}

func placesMesh(n *MeshPlanNode) bool {
	switch n.Kind {
	case MeshPlanKindAsset:
		return true
	case MeshPlanKindBool:
		return placesMesh(n.Left) || placesMesh(n.Right)
	default:
		return false
	}
}

// AdditiveSdfSubtrees returns the subtrees whose plan leaf adds material:
// those under no subtract right-hand side. The preview draws each leaf whole,
// so it draws the union of these, which holds the plan's result. A leaf under
// two is still left out: a − (b − c) is (a − b) ∪ (a ∩ c), so c comes back only
// inside a, and drawn whole it would show outside a too. A subtree the plan
// does not name is kept.
func AdditiveSdfSubtrees(plan *MeshExportPlan, subtrees []SdfSubtree) []SdfSubtree {
	if plan == nil || plan.Root == nil {
		return subtrees
	}
	subtracted := make(map[string]bool)
	var walk func(node *MeshPlanNode, negated bool)
	walk = func(node *MeshPlanNode, negated bool) {
		switch node.Kind {
		case MeshPlanKindSdf:
			subtracted[node.SubtreeID] = negated
		case MeshPlanKindBool:
			walk(node.Left, negated)
			walk(node.Right, negated || node.Op == MeshPlanOpSubtract)
		}
	}
	walk(plan.Root, false)
	out := make([]SdfSubtree, 0, len(subtrees))
	for _, sub := range subtrees {
		if !subtracted[sub.ID] {
			out = append(out, sub)
		}
	}
	return out
}

// PreviewProgram returns the SDF program that the WGSL preview shader should
// evaluate. When the plan places meshes, Three.js draws the placed meshes from
// plan.MeshDecls() and WGSL draws the additive SDF subtrees over prog; when no
// meshes are placed, prog itself is returned unchanged so multi-return
// documents keep their per-return colour and material fields.
func PreviewProgram(prog *program.Program, plan *MeshExportPlan, subtrees []SdfSubtree) *program.Program {
	if plan == nil || plan.Root == nil || len(plan.MeshDecls()) == 0 {
		return prog
	}
	additive := AdditiveSdfSubtrees(plan, subtrees)
	returns := make([]*program.Operand, 0, len(additive))
	for _, sub := range additive {
		returns = append(returns, sub.Return)
	}
	return subtreeProgram(prog, returns, previewControlRoots(prog.Primary))
}

// previewControlRoots returns references to top-level range sliders and UI
// control assignments so dead-code pruning in PreviewProgram keeps their
// declarations aligned with the viewer's uniform parameter list.
func previewControlRoots(pkg *program.Package) []*program.Operand {
	// Invariant: pkg, its Files, Assignments, and Variables are guaranteed non-nil by parser/assembler invariants.
	var roots []*program.Operand
	for _, f := range pkg.Files {
		for _, a := range f.Assignments {
			keep := a.Range != nil || containsUIControlCall(a.Expression)
			if !keep {
				continue
			}
			for _, v := range a.Variables {
				if v.Name != "" {
					roots = append(roots, &program.Operand{
						T: &program.Operand_Reference{
							Reference: &program.Reference{Name: v.Name},
						},
					})
				}
			}
			if a.Range != nil {
				if a.Range.Min != nil {
					roots = append(roots, a.Range.Min)
				}
				if a.Range.Max != nil {
					roots = append(roots, a.Range.Max)
				}
				if a.Range.Period != nil {
					roots = append(roots, a.Range.Period)
				}
			}
		}
	}
	return roots
}

func containsUIControlCall(op *program.Operand) bool {
	if op == nil {
		return false
	}
	switch v := op.T.(type) {
	case *program.Operand_FnCall:
		if v.FnCall == nil {
			return false
		}
		if ref := v.FnCall.Reference; ref != nil && ref.Prefix == "" {
			switch ref.Name {
			case "uiPoint", "uiTransform", "uiColor", "uiDirection", "uiSpline":
				return true
			}
		}
		for _, arg := range v.FnCall.Arguments {
			if containsUIControlCall(arg) {
				return true
			}
		}
	case *program.Operand_Expression:
		if v.Expression == nil {
			return false
		}
		for _, child := range v.Expression.Operands {
			if containsUIControlCall(child) {
				return true
			}
		}
	}
	return false
}

// SubtreesInPlan keeps the subtrees whose SdfLeaf is still in the plan.
func SubtreesInPlan(plan *MeshExportPlan, subtrees []SdfSubtree) []SdfSubtree {
	if plan == nil || plan.Root == nil {
		return nil
	}
	inPlan := make(map[string]*MeshPlanNode)
	IndexSdfLeaves(plan.Root, inPlan)
	kept := make([]SdfSubtree, 0, len(subtrees))
	for _, s := range subtrees {
		if inPlan[s.ID] != nil {
			kept = append(kept, s)
		}
	}
	return kept
}
