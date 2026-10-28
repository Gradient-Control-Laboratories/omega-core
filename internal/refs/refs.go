// Package refs walks a parsed Package and reports the mesh / bvhSdf / vdbSdf
// indices it references. The canvas agent uses this to ensure every asset a
// program touches is baked before the package is forwarded to the codegen
// agents.
package refs

import (
	"gradientcontrol.com/omega/core/program"
)

// MeshReference is one call site's request for an imported mesh asset.
// IsSdf is true for any builtin that participates in the SDF distance
// pipeline (both bvhSdf() and vdbSdf()) and false for the scene-only
// mesh() builtin. IsVdb further selects the OpenVDB-backed oracle when
// IsSdf is true; otherwise the BVH/triangle-soup oracle is used.
//
// VoxelDivisor and HalfBandVoxels carry the optional 2nd and 3rd
// arguments of vdbSdf(idx, voxelDivisor, halfBandVoxels). VoxelDivisor
// is the integer count of voxels along the mesh's longest bbox axis;
// the world-space voxel size is computed as bbox-longest / VoxelDivisor.
// A zero value means "use the canvas-side default" (256 for the
// divisor, 3.0 for HalfBandVoxels). Both fields are ignored unless
// IsVdb is true.
type MeshReference struct {
	Index          int
	IsSdf          bool
	IsVdb          bool
	VoxelDivisor   int
	HalfBandVoxels float64
}

// CollectMeshReferences returns the set of unique (index, isSdf, isVdb)
// triples the package references via mesh() / bvhSdf() / vdbSdf().
// Nested function bodies and all subexpression trees are walked. Only the
// given package is walked; use CollectMeshReferencesWithImports when a mesh
// builtin may be called from an imported library.
func CollectMeshReferences(m *program.Package) []MeshReference {
	return CollectMeshReferencesWithImports(m, nil)
}

// CollectMeshReferencesWithImports is like CollectMeshReferences but also
// walks the packages reachable through `import`, so a mesh builtin called
// inside an imported library (e.g. vdbSdf(0) in a library function the
// program uses) is still collected. getImport resolves an import path to its
// package; if it is nil — or returns nil/error for a path — imports are
// skipped, making this identical to CollectMeshReferences.
func CollectMeshReferencesWithImports(m *program.Package, getImport program.GetImportFunc) []MeshReference {
	if m == nil {
		return nil
	}
	w := &walker{seen: make(map[MeshReference]bool)}
	w.walkPackage(m, getImport, make(map[string]bool))
	return w.collect()
}

func (w *walker) walkPackage(m *program.Package, getImport program.GetImportFunc, visited map[string]bool) {
	if m == nil {
		return
	}
	for _, f := range m.Files {
		if f == nil {
			continue
		}
		for _, a := range f.Assignments {
			w.assignment(a)
		}
		for _, ft := range f.FnTypes {
			w.fnType(ft)
		}
		for _, r := range f.Returns {
			w.operand(r)
		}
		for _, r := range f.ColorReturns {
			w.operand(r)
		}
		for _, r := range f.MaterialReturns {
			w.operand(r)
		}
		if getImport == nil {
			continue
		}
		for _, imp := range f.Imports {
			if imp == nil || visited[imp.Path] {
				continue
			}
			visited[imp.Path] = true
			pkg, err := getImport(imp.Path)
			if err != nil || pkg == nil {
				continue
			}
			w.walkPackage(pkg, getImport, visited)
		}
	}
}

type walker struct {
	seen  map[MeshReference]bool
	order []MeshReference
}

func (w *walker) collect() []MeshReference {
	return w.order
}

func (w *walker) record(ref MeshReference) {
	if w.seen[ref] {
		return
	}
	w.seen[ref] = true
	w.order = append(w.order, ref)
}

func (w *walker) assignment(a *program.Assignment) {
	if a == nil {
		return
	}
	w.operand(a.Expression)
	for _, v := range a.Variables {
		for _, prop := range v.Properties {
			w.operand(prop.Expression)
		}
	}
}

func (w *walker) fnType(ft *program.FnType) {
	if ft == nil {
		return
	}
	for _, group := range ft.Groups {
		if group == nil {
			continue
		}
		for _, fn := range group.Functions {
			w.function(fn)
		}
	}
}

func (w *walker) function(fn *program.Function) {
	if fn == nil {
		return
	}
	for _, a := range fn.Assignments {
		w.assignment(a)
	}
	for _, ft := range fn.FnTypes {
		w.fnType(ft)
	}
	for _, r := range fn.Returns {
		w.operand(r)
	}
}

func (w *walker) operand(op *program.Operand) {
	if op == nil {
		return
	}
	switch v := op.T.(type) {
	case *program.Operand_Value:
	case *program.Operand_Reference:
	case *program.Operand_PropRef:
	case *program.Operand_Expression:
		if v.Expression == nil {
			return
		}
		for _, o := range v.Expression.Operands {
			w.operand(o)
		}
	case *program.Operand_FnCall:
		if v.FnCall == nil {
			return
		}
		w.fnCall(v.FnCall)
	}
}

func (w *walker) fnCall(fc *program.FnCall) {
	if fc == nil {
		return
	}
	ref := fc.Reference
	if ref != nil {
		isMesh := ref.Name == "mesh"
		isBvhSdf := ref.Name == "bvhSdf"
		isVdbSdf := ref.Name == "vdbSdf"
		if isMesh || isBvhSdf || isVdbSdf {
			if len(fc.Arguments) > 0 {
				if vv, ok := fc.Arguments[0].T.(*program.Operand_Value); ok {
					meshRef := MeshReference{
						Index: int(vv.Value),
						IsSdf: isBvhSdf || isVdbSdf,
						IsVdb: isVdbSdf,
					}
					if isVdbSdf {
						if len(fc.Arguments) >= 2 {
							if vv2, ok := fc.Arguments[1].T.(*program.Operand_Value); ok {
								meshRef.VoxelDivisor = int(vv2.Value)
							}
						}
						if len(fc.Arguments) >= 3 {
							if vv3, ok := fc.Arguments[2].T.(*program.Operand_Value); ok {
								meshRef.HalfBandVoxels = vv3.Value
							}
						}
					}
					w.record(meshRef)
				}
			}
		}
	}
	// Recurse into arguments either way — nested calls may reference more meshes.
	for _, arg := range fc.Arguments {
		w.operand(arg)
	}
}
