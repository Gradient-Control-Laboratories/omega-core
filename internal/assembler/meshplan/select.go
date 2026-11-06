package meshplan

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/assembler/variation"
	"gradientcontrol.com/omega/core/program"
)

// SelectPrimarySdfProgram selects the primary analytical SDF subprogram and
// placed mesh declarations for downstream consumer dispatch (tape, simulation,
// measurement, litrender).
func SelectPrimarySdfProgram(prog *program.Program, plan *MeshExportPlan, subtrees []SdfSubtree) (*program.Program, []MeshDecl) {
	var meshDecls []MeshDecl
	if plan != nil && plan.Root != nil {
		meshDecls = plan.MeshDecls()
	}
	if len(subtrees) > 0 && subtrees[0].Return != nil {
		return variation.Program(prog, subtrees[0].Return), meshDecls
	}
	return prog, meshDecls
}

// SelectBodySdfProgram returns the SDF program of one top-level return
// (1-based), for a tape that meshes that body alone. The body must be a single
// SDF leaf: a tape holds no mesh booleans or imported meshes.
func SelectBodySdfProgram(prog *program.Program, plan *MeshExportPlan, subtrees []SdfSubtree, body int) (*program.Program, error) {
	if plan == nil || plan.Root == nil {
		return nil, fmt.Errorf("the program has no body %d", body)
	}
	for _, b := range plan.Bodies() {
		if b.ShapeIndex != body {
			continue
		}
		if b.Root.Kind != MeshPlanKindSdf {
			return nil, fmt.Errorf("%s combines meshes, which dual contouring cannot export; use the local export", b.Name)
		}
		for _, s := range subtrees {
			if s.ID != b.Root.SubtreeID || s.Return == nil {
				continue
			}
			return variation.Program(prog, s.Return), nil
		}
		return nil, fmt.Errorf("%s has no SDF program", b.Name)
	}
	return nil, fmt.Errorf("the program has no body %d", body)
}
