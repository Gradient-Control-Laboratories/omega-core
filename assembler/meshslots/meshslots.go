// Package meshslots assigns compiler-determined slot ids to the expression form
// of mesh/STEP asset imports: `importStep "file.stp"` and `importMesh
// "file.stl"`. The parser lowers each to a FunctionCall named
// `importStep`/`importMesh` whose single argument carries the path as a rune
// list. Assign walks the whole module graph (primary package plus its imports),
// keys each import by (declaring package, kind, asset path), and hands out a
// stable slot id per distinct key.
//
// Keying is per-asset (declaring package, kind, path), not per textual
// occurrence. The resolution is deterministic, so independent agents that run
// Assign on the same graph agree on the numbering.
package meshslots

import (
	"gradientcontrol.com/omega/core/program"
)

// Import describes a compiler-assigned slot for an imported STEP or mesh asset.
type Import struct {
	Slot    int    `json:"slot"`
	Kind    string `json:"kind"`
	Path    string `json:"path"`
	PkgPath string `json:"pkgPath,omitempty"`
}

type Resolver func(pkgPath, kind, path string) (int, bool)

func KindOf(fnName string) (string, bool) {
	switch fnName {
	case "importStep":
		return "step", true
	case "importMesh":
		return "mesh", true
	default:
		return "", false
	}
}

// PathFromArg reconstructs the asset path from the single argument of a lowered
// import call. The path is a rune list: a COLLECT expression of char codes, or
// a bare value for a single-character path. Returns false for a non-literal arg
// (e.g. a path passed as a function parameter), which is not supported.
func PathFromArg(op *program.Operand) (string, bool) {
	if op == nil {
		return "", false
	}
	switch v := op.T.(type) {
	case *program.Operand_Value:
		return string(rune(int(v.Value))), true
	case *program.Operand_Expression:
		expr := v.Expression
		if expr == nil || expr.Operation != program.Operation_COLLECT {
			return "", false
		}
		runes := make([]rune, 0, len(expr.Operands))
		for _, el := range expr.Operands {
			val, ok := el.T.(*program.Operand_Value)
			if !ok {
				return "", false
			}
			runes = append(runes, rune(int(val.Value)))
		}
		return string(runes), true
	default:
		return "", false
	}
}

type slotKey struct {
	pkgPath string
	kind    string
	path    string
}

// Assign walks pkg and its imports, assigns a stable slot to each distinct
// asset import expression, and returns the table (in slot order) plus a
// resolver. Slots start above any author-chosen numbered index found in the
// graph so the expression form coexists with the legacy `importStep N` form.
func Assign(pkg *program.Package, getImport program.GetImportFunc) ([]Import, Resolver) {
	w := &walker{
		getImport: getImport,
		index:     map[slotKey]int{},
		visited:   map[string]bool{},
		nextSlot:  maxAuthorIndex(pkg, getImport) + 1,
	}
	w.walkPackage(pkg, pkg.Path)

	byKey := w.index
	return w.imports, func(pkgPath, kind, path string) (int, bool) {
		slot, ok := byKey[slotKey{pkgPath, kind, path}]
		if !ok && pkgPath == "" && pkg.Path != "" {
			slot, ok = byKey[slotKey{pkg.Path, kind, path}]
		}
		return slot, ok
	}
}

type walker struct {
	getImport program.GetImportFunc
	index     map[slotKey]int
	imports   []Import
	visited   map[string]bool
	nextSlot  int
}

func (w *walker) record(pkgPath, kind, path string) {
	key := slotKey{pkgPath, kind, path}
	if _, seen := w.index[key]; seen {
		return
	}
	slot := w.nextSlot
	w.nextSlot++
	w.index[key] = slot
	w.imports = append(w.imports, Import{Slot: slot, Kind: kind, Path: path, PkgPath: pkgPath})
}

func (w *walker) walkPackage(pkg *program.Package, pkgPath string) {
	if pkg == nil {
		return
	}
	for _, file := range pkg.Files {
		if file == nil {
			continue
		}
		for _, a := range file.Assignments {
			w.walkAssignment(a, pkgPath)
		}
		for _, ft := range file.FnTypes {
			w.walkFnType(ft, pkgPath)
		}
		for _, r := range file.Returns {
			w.walkOperand(r, pkgPath)
		}
		for _, r := range file.ColorReturns {
			w.walkOperand(r, pkgPath)
		}
		for _, r := range file.MaterialReturns {
			w.walkOperand(r, pkgPath)
		}
	}
	if w.getImport == nil {
		return
	}
	for _, file := range pkg.Files {
		if file == nil {
			continue
		}
		for _, imp := range file.Imports {
			if imp == nil || w.visited[imp.Path] {
				continue
			}
			w.visited[imp.Path] = true
			child, err := w.getImport(imp.Path)
			if err != nil || child == nil {
				continue
			}
			w.walkPackage(child, imp.Path)
		}
	}
}

func (w *walker) walkAssignment(a *program.Assignment, pkgPath string) {
	if a != nil {
		w.walkOperand(a.Expression, pkgPath)
	}
}

func (w *walker) walkFnType(ft *program.FnType, pkgPath string) {
	if ft == nil {
		return
	}
	for _, g := range ft.Groups {
		if g == nil {
			continue
		}
		for _, fn := range g.Functions {
			if fn != nil {
				w.walkFunction(fn, pkgPath)
			}
		}
	}
}

func (w *walker) walkFunction(fn *program.Function, pkgPath string) {
	for _, a := range fn.Assignments {
		w.walkAssignment(a, pkgPath)
	}
	for _, ft := range fn.FnTypes {
		w.walkFnType(ft, pkgPath)
	}
	for _, r := range fn.Returns {
		w.walkOperand(r, pkgPath)
	}
}

func (w *walker) walkOperand(op *program.Operand, pkgPath string) {
	if op == nil {
		return
	}
	switch v := op.T.(type) {
	case *program.Operand_FnCall:
		fc := v.FnCall
		if fc != nil {
			if fc.Reference != nil && fc.Reference.Prefix == "" {
				if kind, ok := KindOf(fc.Reference.Name); ok && len(fc.Arguments) == 1 {
					if path, ok := PathFromArg(fc.Arguments[0]); ok {
						w.record(pkgPath, kind, path)
					}
				}
			}
			for _, a := range fc.Arguments {
				w.walkOperand(a, pkgPath)
			}
		}
	case *program.Operand_Expression:
		if v.Expression != nil {
			for _, el := range v.Expression.Operands {
				w.walkOperand(el, pkgPath)
			}
		}
	}
}

// maxAuthorIndex returns the largest author-chosen index across the graph's
// legacy numbered imports, or -1 if there are none.
func maxAuthorIndex(pkg *program.Package, getImport program.GetImportFunc) int {
	max := -1
	visited := map[string]bool{}
	var walk func(p *program.Package)
	walk = func(p *program.Package) {
		if p == nil {
			return
		}
		for _, file := range p.Files {
			if file == nil {
				continue
			}
			for _, mi := range file.MeshImports {
				if int(mi.Index) > max {
					max = int(mi.Index)
				}
			}
			for _, si := range file.StepImports {
				if int(si.Index) > max {
					max = int(si.Index)
				}
			}
		}
		if getImport == nil {
			return
		}
		for _, file := range p.Files {
			if file == nil {
				continue
			}
			for _, imp := range file.Imports {
				if imp == nil || visited[imp.Path] {
					continue
				}
				visited[imp.Path] = true
				child, err := getImport(imp.Path)
				if err == nil {
					walk(child)
				}
			}
		}
	}
	walk(pkg)
	return max
}
