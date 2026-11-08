package warpfold

import (
	"fmt"
	"math"

	"gradientcontrol.com/omega/core/assembler/warpfold"
	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/internal/meshfold"
	"gradientcontrol.com/omega/core/assembler/meshslots"
	"gradientcontrol.com/omega/core/program"
)

// GraphPackages returns every package in the import graph rooted at root (with optional byte
// overrides) that contains any warp-related import or builtin call.
func GraphPackages(getImport program.GetImportFunc, root *program.Package, overrides map[string][]byte) []WarpGraphPkg {
	var out []WarpGraphPkg
	if HasWarpCall(root) {
		out = append(out, WarpGraphPkg{Path: "", Pkg: root})
	}
	if getImport == nil || root == nil {
		return out
	}
	resolve := func(path string) *program.Package {
		if bytes, ok := overrides[path]; ok {
			if p, err := program.DecodePackage(bytes); err == nil {
				return p
			}
		}
		child, _ := getImport(path)
		return child
	}
	visited := map[string]bool{}
	var walk func(p *program.Package)
	walk = func(p *program.Package) {
		for _, file := range p.Files {
			for _, imp := range file.Imports {
				if imp == nil || visited[imp.GetPath()] {
					continue
				}
				visited[imp.GetPath()] = true
				child := resolve(imp.GetPath())
				if child == nil {
					continue
				}
				if HasWarpCall(child) {
					out = append(out, WarpGraphPkg{Path: imp.GetPath(), Pkg: child})
				}
				walk(child)
			}
		}
	}
	walk(root)
	return out
}

// HasWarpCall returns true if the package uses any warp-related builtins or imports.
func HasWarpCall(pkg *program.Package) bool {
	if pkg == nil {
		return false
	}
	for _, file := range pkg.Files {
		if len(file.WarpImports) > 0 {
			return true
		}
	}
	found := false
	var walkOp func(op *program.Operand)
	walkOp = func(op *program.Operand) {
		if op == nil || found {
			return
		}
		switch t := op.T.(type) {
		case *program.Operand_FnCall:
			fc := t.FnCall
			if fc != nil && fc.Reference != nil && fc.Reference.Prefix == "" {
				switch fc.Reference.Name {
				case "importWarp", "meshWarped", "vdbSdfWarped", "warpField":
					found = true
					return
				}
			}
			for _, arg := range fc.Arguments {
				walkOp(arg)
			}
		case *program.Operand_Expression:
			for _, sub := range t.Expression.Operands {
				walkOp(sub)
			}
		}
	}

	for _, file := range pkg.Files {
		for _, a := range file.Assignments {
			walkOp(a.Expression)
		}
		for _, ft := range file.FnTypes {
			for _, g := range ft.Groups {
				for _, fn := range g.Functions {
					for _, a := range fn.Assignments {
						walkOp(a.Expression)
					}
					for _, r := range fn.Returns {
						walkOp(r)
					}
				}
			}
		}
		for _, r := range file.Returns {
			walkOp(r)
		}
		for _, r := range file.ColorReturns {
			walkOp(r)
		}
		for _, r := range file.MaterialReturns {
			walkOp(r)
		}
	}
	return found
}

// WarpRef is a simulation file and the directory it resolves against.
type WarpRef struct {
	Dir  string
	Path string
}

// WarpSlots maps simulation files to allocated slots across the module graph.
type WarpSlots struct {
	BySlot    map[int]WarpRef
	ExprCalls map[*program.Operand]int
	ExprByPkg map[string][]*program.Operand
}

func (w *WarpSlots) Ref(slot int) (WarpRef, bool) {
	r, ok := w.BySlot[slot]
	return r, ok
}

func (w *WarpSlots) Empty() bool {
	return len(w.BySlot) == 0 && len(w.ExprCalls) == 0
}

type WarpGraphPkg struct {
	Path string
	Pkg  *program.Package
}

// AssignWarpSlots numbers simulations across the packages.
func AssignWarpSlots(pkgs []WarpGraphPkg, primaryPkgName string) *WarpSlots {
	w := &WarpSlots{
		BySlot:    make(map[int]WarpRef),
		ExprCalls: make(map[*program.Operand]int),
		ExprByPkg: make(map[string][]*program.Operand),
	}
	next := 0
	bySlotKey := make(map[WarpRef]int)

	dirOf := func(gp WarpGraphPkg) string {
		if gp.Path != "" {
			return gp.Path
		}
		return primaryPkgName
	}

	for _, gp := range pkgs {
		dir := dirOf(gp)
		for _, file := range gp.Pkg.Files {
			for _, imp := range file.WarpImports {
				ref := WarpRef{Dir: dir, Path: imp.GetPath()}
				if _, seen := bySlotKey[ref]; seen {
					continue
				}
				slot := int(imp.GetIndex())
				if gp.Path != "" || w.BySlot[slot] != (WarpRef{}) {
					slot = next
				}
				w.BySlot[slot] = ref
				bySlotKey[ref] = slot
				if slot >= next {
					next = slot + 1
				}
			}
		}
	}

	var walkOperand func(op *program.Operand, gp WarpGraphPkg)
	walkOperand = func(op *program.Operand, gp WarpGraphPkg) {
		if op == nil {
			return
		}
		switch t := op.T.(type) {
		case *program.Operand_FnCall:
			fc := t.FnCall
			if fc != nil && fc.Reference != nil && fc.Reference.Prefix == "" && fc.Reference.Name == "importWarp" {
				if len(fc.Arguments) == 1 {
					path := stringFromOperand(fc.Arguments[0])
					ref := WarpRef{Dir: dirOf(gp), Path: path}
					slot, ok := bySlotKey[ref]
					if !ok {
						slot = next
						next++
						w.BySlot[slot] = ref
						bySlotKey[ref] = slot
					}
					w.ExprCalls[op] = slot
					w.ExprByPkg[gp.Path] = append(w.ExprByPkg[gp.Path], op)
				}
			}
			for _, arg := range fc.Arguments {
				walkOperand(arg, gp)
			}
		case *program.Operand_Expression:
			for _, sub := range t.Expression.Operands {
				walkOperand(sub, gp)
			}
		}
	}

	for _, gp := range pkgs {
		for _, file := range gp.Pkg.Files {
			for _, a := range file.Assignments {
				walkOperand(a.Expression, gp)
			}
			for _, ft := range file.FnTypes {
				for _, g := range ft.Groups {
					for _, fn := range g.Functions {
						for _, a := range fn.Assignments {
							walkOperand(a.Expression, gp)
						}
						for _, r := range fn.Returns {
							walkOperand(r, gp)
						}
					}
				}
			}
			for _, r := range file.Returns {
				walkOperand(r, gp)
			}
			for _, r := range file.ColorReturns {
				walkOperand(r, gp)
			}
			for _, r := range file.MaterialReturns {
				walkOperand(r, gp)
			}
		}
	}

	return w
}

func stringFromOperand(op *program.Operand) string {
	if op == nil {
		return ""
	}
	if expr := op.GetExpression(); expr != nil {
		var runes []rune
		for _, sub := range expr.Operands {
			if v, ok := sub.T.(*program.Operand_Value); ok {
				runes = append(runes, rune(v.Value))
			}
		}
		return string(runes)
	}
	return ""
}

// FoldProgram folds all warp simulations and rewrites meshWarped/vdbSdfWarped calls across
// the program graph using the provided WarpResolver, returning the folded program and the
// deduplicated list of resolved WarpedAsset descriptors.
func FoldProgram(prog *program.Program, resolver warpfold.WarpResolver) (*program.Program, []warpfold.WarpedAsset, error) {
	if prog == nil || resolver == nil {
		return prog, nil, nil
	}

	hasWarp := false
	if prog.Primary != nil && HasWarpCall(prog.Primary) {
		hasWarp = true
	}
	if !hasWarp {
		for _, dep := range prog.Dependencies {
			if dep != nil && HasWarpCall(dep) {
				hasWarp = true
				break
			}
		}
	}
	if !hasWarp {
		return prog, nil, nil
	}

	_, slotResolver := meshslots.Assign(prog.Primary, prog.GetImport)

	// Clone & slot-fold the program graph so that mesh slot parameters are constants.
	foldedProg, err := meshfold.FoldProgramMeshIndices(prog, slotResolver)
	if err != nil {
		return nil, nil, fmt.Errorf("warpfold: mesh slots folding failed: %w", err)
	}

	var foldedPkgs []WarpGraphPkg
	if foldedProg.Primary != nil && HasWarpCall(foldedProg.Primary) {
		foldedPkgs = append(foldedPkgs, WarpGraphPkg{Path: "", Pkg: foldedProg.Primary})
	}
	for _, dep := range foldedProg.Dependencies {
		if dep != nil && HasWarpCall(dep) {
			foldedPkgs = append(foldedPkgs, WarpGraphPkg{Path: dep.Path, Pkg: dep})
		}
	}

	warps := AssignWarpSlots(foldedPkgs, foldedProg.Primary.GetPath())

	// Rewrite all expression-form importWarp calls to their assigned integer slot
	// across every package first so cross-package references (e.g. a tree referencing
	// a branch's warpId) evaluate to the assigned slot during foldPackage.
	for _, gp := range foldedPkgs {
		for _, op := range warps.ExprByPkg[gp.Path] {
			slot := warps.ExprCalls[op]
			*op = program.Operand{
				Start: op.Start,
				End:   op.End,
				T:     &program.Operand_Value{Value: float64(slot)},
			}
		}
	}

	var warpedAssets []warpfold.WarpedAsset
	seenDerived := make(map[int]bool)

	for _, gp := range foldedPkgs {
		folder := &warpFolder{
			prog:         foldedProg,
			pkg:          gp.Pkg,
			pkgPath:      gp.Path,
			resolver:     resolver,
			warps:        warps,
			slotResolver: slotResolver,
			seenDerived:  seenDerived,
			warpedAssets: &warpedAssets,
		}
		if err := folder.foldPackage(); err != nil {
			return nil, nil, err
		}
	}

	return foldedProg, warpedAssets, nil
}

type warpFolder struct {
	prog         *program.Program
	pkg          *program.Package
	pkgPath      string
	resolver     warpfold.WarpResolver
	warps        *WarpSlots
	slotResolver meshslots.Resolver
	seenDerived  map[int]bool
	warpedAssets *[]warpfold.WarpedAsset
}

func (f *warpFolder) foldPackage() error {
	for _, file := range f.pkg.Files {
		for _, a := range file.Assignments {
			out, err := f.foldOperand(a.Expression)
			if err != nil {
				return err
			}
			a.Expression = out
		}
		for _, ft := range file.FnTypes {
			if err := f.foldFnType(ft); err != nil {
				return err
			}
		}
		for i, ret := range file.Returns {
			out, err := f.foldOperand(ret)
			if err != nil {
				return err
			}
			file.Returns[i] = out
		}
		for i, ret := range file.ColorReturns {
			out, err := f.foldOperand(ret)
			if err != nil {
				return err
			}
			file.ColorReturns[i] = out
		}
		for i, ret := range file.MaterialReturns {
			out, err := f.foldOperand(ret)
			if err != nil {
				return err
			}
			file.MaterialReturns[i] = out
		}
	}
	return nil
}

func (f *warpFolder) foldFnType(ft *program.FnType) error {
	if ft == nil {
		return nil
	}
	for _, g := range ft.Groups {
		for _, fn := range g.Functions {
			if fn == nil {
				continue
			}
			for _, a := range fn.Assignments {
				out, err := f.foldOperand(a.Expression)
				if err != nil {
					return err
				}
				a.Expression = out
			}
			for _, sub := range fn.FnTypes {
				if err := f.foldFnType(sub); err != nil {
					return err
				}
			}
			for i, ret := range fn.Returns {
				out, err := f.foldOperand(ret)
				if err != nil {
					return err
				}
				fn.Returns[i] = out
			}
		}
	}
	return nil
}

func (f *warpFolder) foldOperand(op *program.Operand) (*program.Operand, error) {
	if op == nil {
		return nil, nil
	}
	switch v := op.T.(type) {
	case *program.Operand_FnCall:
		fc := v.FnCall
		if fc == nil || fc.Reference == nil {
			return op, nil
		}
		for i, arg := range fc.Arguments {
			out, err := f.foldOperand(arg)
			if err != nil {
				return nil, err
			}
			fc.Arguments[i] = out
		}
		if fc.Reference.Prefix != "" {
			return op, nil
		}
		switch fc.Reference.Name {
		case "meshWarped":
			return op, f.rewrite(fc, op, "mesh")
		case "vdbSdfWarped":
			return op, f.rewrite(fc, op, "vdbSdf")
		}
	case *program.Operand_Expression:
		if v.Expression != nil {
			for i, o := range v.Expression.Operands {
				out, err := f.foldOperand(o)
				if err != nil {
					return nil, err
				}
				v.Expression.Operands[i] = out
			}
		}
	}
	return op, nil
}

func (f *warpFolder) rewrite(fc *program.FnCall, op *program.Operand, plain string) error {
	name := fc.Reference.Name
	fail := func(format string, args ...any) error {
		return fmt.Errorf("%s: %s", name, fmt.Sprintf(format, args...))
	}
	if len(fc.Arguments) < 3 {
		return fail("needs a mesh slot, a warp slot and a scale")
	}
	meshSlot, ok := f.constIndex(fc.Arguments[0])
	if !ok {
		return fail("the mesh id must be a constant, non-negative integer")
	}
	warpSlot, ok := f.constIndex(fc.Arguments[1])
	if !ok {
		return fail("the warp id must be a constant, non-negative integer")
	}
	scale, ok := f.constScalar(fc.Arguments[2])
	if !ok {
		return fail("the scale must be a constant number")
	}

	ref, ok := f.warps.Ref(warpSlot)
	if !ok {
		return fail("warp id %d names no imported simulation", warpSlot)
	}

	derivedSlot, err := f.resolver.ResolveWarpedSlot(meshSlot, ref.Dir, ref.Path, scale)
	if err != nil {
		return fail("%v", err)
	}
	// Invariant: f.warpedAssets and f.seenDerived are guaranteed non-nil by FoldProgram.
	if !f.seenDerived[derivedSlot] {
		f.seenDerived[derivedSlot] = true
		*f.warpedAssets = append(*f.warpedAssets, warpfold.WarpedAsset{
			DerivedSlot: derivedSlot,
			MeshSlot:    meshSlot,
			WarpDir:     ref.Dir,
			WarpPath:    ref.Path,
			Scale:       scale,
		})
	}

	fc.Reference.Name = plain
	rest := fc.Arguments[3:]
	args := make([]*program.Operand, 0, 1+len(rest))
	args = append(args, &program.Operand{
		Start: fc.Arguments[0].Start,
		End:   fc.Arguments[0].End,
		T:     &program.Operand_Value{Value: float64(derivedSlot)},
	})
	args = append(args, rest...)
	fc.Arguments = args
	return nil
}

func (f *warpFolder) constScalar(op *program.Operand) (float64, bool) {
	if len(op.ConstValues) == 1 {
		return op.ConstValues[0], true
	}
	if v, ok := op.T.(*program.Operand_Value); ok {
		return v.Value, true
	}
	// Invariant: f.prog is guaranteed non-nil by FoldProgram.
	vals, err := eval.ConstPkgSlotsAt(op, f.pkg, f.pkgPath, f.prog.GetImport, f.slotResolver)
	if err == nil && len(vals) == 1 {
		return vals[0], true
	}
	return 0, false
}

func (f *warpFolder) constIndex(op *program.Operand) (int, bool) {
	v, ok := f.constScalar(op)
	if !ok {
		return 0, false
	}
	i := int(math.Round(v))
	if i < 0 || math.Abs(v-float64(i)) > 1e-9 {
		return 0, false
	}
	return i, true
}

// FoldWarpFieldToZero replaces every warpField(...) call across the program with a constant [0,0,0].
func FoldWarpFieldToZero(prog *program.Program) (*program.Program, error) {
	if prog == nil {
		return nil, nil
	}
	if prog.Primary != nil {
		foldWarpFieldPackage(prog.Primary)
	}
	for _, dep := range prog.Dependencies {
		if dep != nil {
			foldWarpFieldPackage(dep)
		}
	}
	return prog, nil
}

func foldWarpFieldPackage(pkg *program.Package) {
	for _, file := range pkg.Files {
		for _, a := range file.Assignments {
			a.Expression = foldWarpFieldOperand(a.Expression)
		}
		for _, ft := range file.FnTypes {
			foldWarpFieldFnType(ft)
		}
		for i, ret := range file.Returns {
			file.Returns[i] = foldWarpFieldOperand(ret)
		}
		for i, ret := range file.ColorReturns {
			file.ColorReturns[i] = foldWarpFieldOperand(ret)
		}
		for i, ret := range file.MaterialReturns {
			file.MaterialReturns[i] = foldWarpFieldOperand(ret)
		}
	}
}

func foldWarpFieldFnType(ft *program.FnType) {
	if ft == nil {
		return
	}
	for _, g := range ft.Groups {
		for _, fn := range g.Functions {
			if fn == nil {
				continue
			}
			for _, a := range fn.Assignments {
				a.Expression = foldWarpFieldOperand(a.Expression)
			}
			for _, sub := range fn.FnTypes {
				foldWarpFieldFnType(sub)
			}
			for i, ret := range fn.Returns {
				fn.Returns[i] = foldWarpFieldOperand(ret)
			}
		}
	}
}

func foldWarpFieldOperand(op *program.Operand) *program.Operand {
	if op == nil {
		return nil
	}
	switch v := op.T.(type) {
	case *program.Operand_FnCall:
		fc := v.FnCall
		if fc != nil && fc.Reference != nil && fc.Reference.Prefix == "" && fc.Reference.Name == "warpField" {
			return constZeroVec(op)
		}
		for i, a := range fc.Arguments {
			fc.Arguments[i] = foldWarpFieldOperand(a)
		}
	case *program.Operand_Expression:
		if v.Expression != nil {
			for i, o := range v.Expression.Operands {
				v.Expression.Operands[i] = foldWarpFieldOperand(o)
			}
		}
	}
	return op
}

func constZeroVec(from *program.Operand) *program.Operand {
	ops := make([]*program.Operand, 3)
	for i := 0; i < 3; i++ {
		ops[i] = &program.Operand{
			Start: from.GetStart(),
			End:   from.GetEnd(),
			T:     &program.Operand_Value{Value: 0},
		}
	}
	return &program.Operand{
		Start: from.GetStart(),
		End:   from.GetEnd(),
		T: &program.Operand_Expression{Expression: &program.Expression{
			Operation: program.Operation_COLLECT,
			Operands:  ops,
		}},
	}
}
