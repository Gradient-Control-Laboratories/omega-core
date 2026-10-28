package textures

import (
	"sort"

	"gradientcontrol.com/omega/core/program"
)

// Extract scans all reachable statements and expressions across prog (Primary and
// Dependencies) to determine dense mesh SDF texture slot allocations and global
// shader texture requirements.
//
// Dense SDF texture slots are assigned to distinct asset indices referenced by
// bvhSdf(i) and vdbSdf(i) calls, sorted deterministically in ascending order (0, 1, ...).
func Extract(prog *program.Program) ([]*program.MeshSdfSlot, *program.GlobalTextures) {
	// Invariant: prog is guaranteed non-nil throughout compiler pipeline.
	global := &program.GlobalTextures{}

	w := &walker{
		global:       global,
		assetIndices: make(map[int32]bool),
	}

	w.walkPackage(prog.Primary)
	for _, dep := range prog.Dependencies {
		w.walkPackage(dep)
	}

	if len(w.assetIndices) == 0 {
		return nil, global
	}

	sortedIndices := make([]int, 0, len(w.assetIndices))
	for idx := range w.assetIndices {
		sortedIndices = append(sortedIndices, int(idx))
	}
	sort.Ints(sortedIndices)

	slots := make([]*program.MeshSdfSlot, len(sortedIndices))
	for slot, assetIdx := range sortedIndices {
		slots[slot] = &program.MeshSdfSlot{
			Slot:       int32(slot),
			AssetIndex: int32(assetIdx),
		}
	}

	return slots, global
}

type walker struct {
	global       *program.GlobalTextures
	assetIndices map[int32]bool
}

func (w *walker) walkPackage(pkg *program.Package) {
	if pkg == nil {
		return
	}
	for _, file := range pkg.Files {
		if file == nil {
			continue
		}
		for _, a := range file.Assignments {
			w.walkAssignment(a)
		}
		for _, ft := range file.FnTypes {
			w.walkFnType(ft)
		}
		for _, r := range file.Returns {
			w.walkOperand(r)
		}
		for _, r := range file.ColorReturns {
			w.walkOperand(r)
		}
		for _, r := range file.MaterialReturns {
			w.walkOperand(r)
		}
	}
}

func (w *walker) walkAssignment(a *program.Assignment) {
	if a != nil {
		w.walkOperand(a.Expression)
	}
}

func (w *walker) walkFnType(ft *program.FnType) {
	if ft == nil {
		return
	}
	for _, g := range ft.Groups {
		if g == nil {
			continue
		}
		for _, fn := range g.Functions {
			w.walkFunction(fn)
		}
	}
}

func (w *walker) walkFunction(fn *program.Function) {
	if fn == nil {
		return
	}
	for _, a := range fn.Assignments {
		w.walkAssignment(a)
	}
	for _, ft := range fn.FnTypes {
		w.walkFnType(ft)
	}
	for _, r := range fn.Returns {
		w.walkOperand(r)
	}
}

func (w *walker) walkOperand(op *program.Operand) {
	if op == nil {
		return
	}
	switch v := op.T.(type) {
	case *program.Operand_FnCall:
		fc := v.FnCall
		if fc == nil {
			return
		}
		if fc.Reference != nil && fc.Reference.Prefix == "" {
			name := fc.Reference.Name
			switch name {
			case "bvhSdf", "vdbSdf":
				if len(fc.Arguments) >= 1 {
					if valOp, ok := fc.Arguments[0].T.(*program.Operand_Value); ok {
						w.assetIndices[int32(valOp.Value)] = true
					}
				}
			case program.NoiseHash3, program.WorleyF1, program.WorleyFeature:
				w.global.UsesNoiseTile = true
			case "warpField":
				w.global.UsesWarpField = true
			case program.GaborSpinodoid, program.GaborSpinodoidGradient,
				program.FastGaborSpinodoid, program.FastGaborSpinodoidGradient:
				// The preview renders the lattice field for all four: only the
				// bake tells hodge's field from the lattice's.
				w.global.UsesSpinodoid = true
			}
		}
		for _, arg := range fc.Arguments {
			w.walkOperand(arg)
		}
	case *program.Operand_Expression:
		if v.Expression != nil {
			for _, operand := range v.Expression.Operands {
				w.walkOperand(operand)
			}
		}
	}
}
