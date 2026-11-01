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
	// Invariant: prog, prog.Primary, and prog.Dependencies elements are guaranteed non-nil throughout compiler pipeline.
	global := &program.GlobalTextures{}
	if !hasPackageReturns(prog.Primary) {
		return nil, global
	}

	w := &walker{
		global:     global,
		assetSlots: make(map[int32]sdfSlotInfo),
	}

	w.walkPackage(prog.Primary)
	for _, dep := range prog.Dependencies {
		w.walkPackage(dep)
	}

	if len(w.assetSlots) == 0 {
		return nil, global
	}

	sortedIndices := make([]int, 0, len(w.assetSlots))
	for idx := range w.assetSlots {
		sortedIndices = append(sortedIndices, int(idx))
	}
	sort.Ints(sortedIndices)

	slots := make([]*program.MeshSdfSlot, len(sortedIndices))
	for slot, assetIdx := range sortedIndices {
		info := w.assetSlots[int32(assetIdx)]
		slots[slot] = &program.MeshSdfSlot{
			Slot:           int32(slot),
			AssetIndex:     int32(assetIdx),
			Kind:           info.kind,
			VoxelDivisor:   info.voxelDivisor,
			HalfBandVoxels: info.halfBandVoxels,
		}
	}

	return slots, global
}

func hasPackageReturns(pkg *program.Package) bool {
	// Invariant: pkg and pkg.Files elements are guaranteed non-nil.
	for _, f := range pkg.Files {
		if len(f.Returns) > 0 {
			return true
		}
	}
	return false
}

type sdfSlotInfo struct {
	kind           program.MeshSdfKind
	voxelDivisor   int32
	halfBandVoxels float64
}

type walker struct {
	global     *program.GlobalTextures
	assetSlots map[int32]sdfSlotInfo
}

func (w *walker) walkPackage(pkg *program.Package) {
	// Invariant: pkg and pkg.Files elements are guaranteed non-nil.
	for _, file := range pkg.Files {
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
	// Invariant: a and a.Expression are guaranteed non-nil.
	w.walkOperand(a.Expression)
	for _, v := range a.Variables {
		for _, prop := range v.Properties {
			w.walkOperand(prop.Expression)
		}
	}
}

func (w *walker) walkFnType(ft *program.FnType) {
	// Invariant: ft and ft.Groups elements are guaranteed non-nil.
	for _, g := range ft.Groups {
		for _, fn := range g.Functions {
			w.walkFunction(fn)
		}
	}
}

func (w *walker) walkFunction(fn *program.Function) {
	// Invariant: fn is guaranteed non-nil.
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
	// Invariant: op and its oneof payload fields (FnCall, Reference, Expression) are guaranteed non-nil.
	switch v := op.T.(type) {
	case *program.Operand_FnCall:
		fc := v.FnCall
		if fc.Reference.Prefix == "" {
			name := fc.Reference.Name
			switch name {
			case "bvhSdf", "vdbSdf":
				if len(fc.Arguments) >= 1 {
					if valOp, ok := fc.Arguments[0].T.(*program.Operand_Value); ok {
						idx := int32(valOp.Value)
						info := w.assetSlots[idx]
						if name == "vdbSdf" {
							info.kind = program.MeshSdfKind_MESH_SDF_KIND_VDB
							if len(fc.Arguments) >= 2 {
								if vv2, ok := fc.Arguments[1].T.(*program.Operand_Value); ok {
									info.voxelDivisor = int32(vv2.Value)
								}
							}
							if len(fc.Arguments) >= 3 {
								if vv3, ok := fc.Arguments[2].T.(*program.Operand_Value); ok {
									info.halfBandVoxels = vv3.Value
								}
							}
						} else if info.kind == program.MeshSdfKind_MESH_SDF_KIND_UNSPECIFIED {
							info.kind = program.MeshSdfKind_MESH_SDF_KIND_BVH
						}
						w.assetSlots[idx] = info
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
		for _, operand := range v.Expression.Operands {
			w.walkOperand(operand)
		}
	}
}
