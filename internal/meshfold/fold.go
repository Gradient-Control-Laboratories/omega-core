package meshfold

import (
	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/internal/closure"
	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/internal/meshslots"
	"gradientcontrol.com/omega/core/internal/types"
	"gradientcontrol.com/omega/core/program"
)

// TODO: Inline this func
// FoldProgram assigns mesh slots across the program graph and folds mesh indices to constants.
func FoldProgram(prog *program.Program) (*program.Program, error) {
	p, _, err := FoldProgramWithSlots(prog)
	return p, err
}

// FoldProgramWithSlots assigns mesh slots across the program graph, folds mesh indices to constants,
// and returns the folded program along with the assigned import slot table.
// TODO: is the returned program a clone? If so, then update the comment to indicate as much
func FoldProgramWithSlots(prog *program.Program) (*program.Program, []meshslots.Import, error) {
	return FoldProgramWithSlotsChecked(prog, nil)
}

// FoldProgramWithSlotsChecked is FoldProgramWithSlots for a program whose
// runtime parameters are known: newChecker builds the parameter checker for
// the primary package the fold runs over. The constant fold that resolves the
// slot variables runs over the WHOLE program, and with no checker it folds
// through a slider as if it were its current literal: `r > 0.5 ? a : b` was
// pruned to one branch and r dropped from the shader, so the viewer's live
// slider never reached it. The assembler passes the checker its own folds use.
//
// A constructor rather than a checker because FoldMeshIndicesAt CLONES the
// packages first, and the assembler's checker knows a UI marker's literals by
// operand: built before the clone it would not recognise them, and under a
// checker the fold also collapses the constants it evaluates, so pt.x * 2.0 +
// 1.0 became 3 and the marker's declaration went as dead code. A nil
// newChecker folds everything, as before.
func FoldProgramWithSlotsChecked(prog *program.Program, newChecker func(primary *program.Package) eval.ParameterChecker) (*program.Program, []meshslots.Import, error) {
	if prog == nil || prog.Primary == nil {
		return prog, nil, nil
	}
	table, slots := meshslots.Assign(prog.Primary, prog.GetImport)
	prog.Primary = FoldMeshIndicesAt(prog.Primary, prog.Primary.Path, prog.GetImport, slots)
	for i, dep := range prog.Dependencies {
		prog.Dependencies[i] = FoldMeshIndicesAt(dep, dep.Path, prog.GetImport, slots)
	}
	var checker eval.ParameterChecker
	if newChecker != nil {
		checker = newChecker(prog.Primary)
	}
	if _, err := eval.Fold(prog, checker); err != nil {
		return nil, nil, err
	}
	prog.Primary = FoldMeshIndicesAt(prog.Primary, prog.Primary.Path, prog.GetImport, slots)
	for i, dep := range prog.Dependencies {
		prog.Dependencies[i] = FoldMeshIndicesAt(dep, dep.Path, prog.GetImport, slots)
	}
	return prog, table, nil
}

// FoldProgramMeshIndices returns a clone of prog in which every package (Primary
// and Dependencies) has its mesh() / bvhSdf() / vdbSdf() index arguments folded.
func FoldProgramMeshIndices(prog *program.Program, slots meshslots.Resolver) (*program.Program, error) {
	if prog == nil {
		return nil, nil
	}
	clone := &program.Program{}

	clone.Primary = FoldMeshIndicesAt(prog.Primary, prog.Primary.Path, prog.GetImport, slots)
	if len(prog.Dependencies) > 0 {
		clone.Dependencies = make([]*program.Package, len(prog.Dependencies))
		for i, dep := range prog.Dependencies {
			clone.Dependencies[i] = FoldMeshIndicesAt(dep, dep.Path, prog.GetImport, slots)
		}
	}

	// FoldProgramMeshIndices operates on a fresh, cloned program graph built
	// from source-manager packages prior to the main backend assembler pipeline.
	// Because this function runs its own private types.Analyze and eval.Fold pass
	// to constant-fold slot variables, closure.Transform must be called here to
	// monomorphize closures, specialize higher-order templates (e.g., combiners
	// in boxCentered/cylinderCapped), and prune unspecialized templates so
	// types.Analyze does not encounter unresolved higher-order function parameters.
	if err := closure.Transform(clone); err != nil {
		return nil, err
	}

	// Re-run type analysis and constant evaluation pass on the clone to propagate slots
	if err := types.Analyze(clone); err != nil {
		return nil, err
	}
	if _, err := eval.Fold(clone, nil); err != nil {
		return nil, err
	}

	// Second pass of FoldMeshIndices to fold the resolved index variables
	clone.Primary = FoldMeshIndicesAt(clone.Primary, prog.Primary.Path, prog.GetImport, slots)
	for i, dep := range clone.Dependencies {
		clone.Dependencies[i] = FoldMeshIndicesAt(dep, dep.Path, prog.GetImport, slots)
	}

	return clone, nil
}

// FoldMeshIndices returns a clone of pkg in which every mesh() / bvhSdf() /
// vdbSdf() index argument that is not already an integer literal is folded to
// its compiler-assigned slot id (using slots). This lets the downstream readers
// — build/codegen, refs, and the mesh-export extractor — keep reading a literal
// index even when the program wrote `mesh(partId, ...)` with an id returned by
// an importStep/importMesh expression.
//
// The original pkg (which may be a srcMgr-cached import) is never mutated.
// Folding that fails or yields a non-scalar is left untouched, so legacy
// numbered imports and genuinely-dynamic indices pass through unchanged.
func FoldMeshIndices(pkg *program.Package, getImport program.GetImportFunc, slots meshslots.Resolver) *program.Package {
	pkgPath := ""
	if pkg != nil {
		pkgPath = pkg.Path
	}
	return FoldMeshIndicesAt(pkg, pkgPath, getImport, slots)
}

// FoldMeshIndicesAt folds an imported package with its module path.
func FoldMeshIndicesAt(pkg *program.Package, pkgPath string, getImport program.GetImportFunc, slots meshslots.Resolver) *program.Package {
	if pkg == nil {
		return nil
	}
	clone := proto.Clone(pkg).(*program.Package)
	f := &meshFolder{scope: clone, pkgPath: pkgPath, getImport: getImport, slots: slots}
	for _, file := range clone.Files {
		if file != nil {
			f.file(file)
		}
	}
	return clone
}

// FoldingGetImport wraps getImport so every fetched package has its mesh index
// arguments folded (via FoldMeshIndices) before the build inlines it. Pair it
// with FoldMeshIndices on the root module so the whole graph the build sees has
// literal indices.
func FoldingGetImport(getImport program.GetImportFunc, slots meshslots.Resolver) program.GetImportFunc {
	return func(path string) (*program.Package, error) {
		pkg, err := getImport(path)
		if err != nil || pkg == nil {
			return pkg, err
		}
		return FoldMeshIndicesAt(pkg, path, getImport, slots), nil
	}
}

type meshFolder struct {
	scope     *program.Package
	pkgPath   string
	getImport program.GetImportFunc
	slots     meshslots.Resolver
}

func (f *meshFolder) file(file *program.PackageFile) {
	for _, a := range file.Assignments {
		f.assignment(a)
	}
	for _, ft := range file.FnTypes {
		f.fnType(ft)
	}
	for _, r := range file.Returns {
		f.operand(r)
	}
	for _, r := range file.ColorReturns {
		f.operand(r)
	}
	for _, r := range file.MaterialReturns {
		f.operand(r)
	}
}

func (f *meshFolder) assignment(a *program.Assignment) {
	if a != nil {
		f.operand(a.Expression)
	}
}

func (f *meshFolder) fnType(ft *program.FnType) {
	if ft == nil {
		return
	}
	for _, g := range ft.Groups {
		if g == nil {
			continue
		}
		for _, fn := range g.Functions {
			if fn != nil {
				f.function(fn)
			}
		}
	}
}

func (f *meshFolder) function(fn *program.Function) {
	for _, a := range fn.Assignments {
		f.assignment(a)
	}
	for _, ft := range fn.FnTypes {
		f.fnType(ft)
	}
	for _, r := range fn.Returns {
		f.operand(r)
	}
}

func (f *meshFolder) operand(op *program.Operand) {
	if op == nil {
		return
	}
	switch v := op.T.(type) {
	case *program.Operand_Expression:
		if v.Expression != nil {
			for _, el := range v.Expression.Operands {
				f.operand(el)
			}
		}
	case *program.Operand_FnCall:
		fc := v.FnCall
		if fc == nil {
			return
		}
		// Replace an importStep/importMesh call with its slot literal in place,
		// so the build never sees the (otherwise-unknown) importStep builtin and
		// the declaration `partId = importStep "f"` becomes `partId = <slot>`.
		if fc.Reference != nil && fc.Reference.Prefix == "" {
			name := fc.Reference.Name
			if kind, ok := meshslots.KindOf(name); ok && len(fc.Arguments) == 1 {
				if path, pok := meshslots.PathFromArg(fc.Arguments[0]); pok {
					if f.slots != nil {
						if slot, sok := f.slots(f.pkgPath, kind, path); sok {
							op.T = &program.Operand_Value{Value: float64(slot)}
							return
						}
					}
				}
			}
			// The warped forms carry the same mesh slot in argument 0; their warp
			// slot is a separate namespace and only ever appears as a literal, so
			// it needs no folding.
			if (name == "mesh" || name == "bvhSdf" || name == "vdbSdf" ||
				name == "meshWarped" || name == "vdbSdfWarped") && len(fc.Arguments) > 0 {
				fc.Arguments[0] = f.foldIndex(fc.Arguments[0])
			}
		}
		for _, a := range fc.Arguments {
			f.operand(a)
		}
	}
}

// foldIndex replaces a non-literal mesh index argument with the constant it
// folds to; literals and unfoldable args are returned unchanged.
func (f *meshFolder) foldIndex(arg *program.Operand) *program.Operand {
	if arg == nil {
		return arg
	}
	if _, ok := arg.T.(*program.Operand_Value); ok {
		return arg // already a literal (legacy numbered form)
	}
	if len(arg.ConstValues) == 1 {
		return &program.Operand{Start: arg.Start, End: arg.End, T: &program.Operand_Value{Value: arg.ConstValues[0]}}
	}
	vals, err := eval.ConstPkgSlotsAt(arg, f.scope, f.pkgPath, f.getImport, f.slots)
	if err != nil || len(vals) != 1 {
		return arg
	}
	return &program.Operand{Start: arg.Start, End: arg.End, T: &program.Operand_Value{Value: vals[0]}}
}
