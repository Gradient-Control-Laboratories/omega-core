package property

import "gradientcontrol.com/omega/core/program"

// materialLayRow is the material record's row for a lay: the direction grooves
// too fine to see run in, and the anisotropy they give the highlight.
const materialLayRow = 3

// layFunction is the material library's one way to give a material a lay.
const layFunction = "withLay"

// usesCall reports whether a call to a function named name is reachable from
// pkg's top level: through pkg's own functions and constants, and through those
// of the packages it imports, directly or through another. A material's lay row
// is blended across its units like any other row, which costs a derivation of
// every unit's weight and nearest-unit fill; a program that never gives a
// material a lay has a lay of 0 everywhere, so it skips all of that for a
// constant.
//
// Reachability is by name, every overload of a called function followed, and a
// bare reference to a function or to a constant reaches it too, since a
// function handed on as a value may be called anywhere. What no path reaches
// does not count: a library can carry a preset that calls name, as material's
// own `withLay(m)` does, without every document that imports it paying for a
// lay. Without getImport only pkg itself is read.
func usesCall(pkg *program.Package, getImport program.GetImportFunc, name string) bool {
	w := &reach{getImport: getImport, name: name, seen: map[string]bool{}}
	for _, f := range pkg.Files {
		if f != nil && w.body(pkg, f, f.Assignments, f.Returns, nil) {
			return true
		}
	}
	return false
}

// reach walks the declarations a program can reach, looking for a call.
type reach struct {
	getImport program.GetImportFunc
	name      string
	// seen is the declarations already read, by package path and name, so a
	// recursive or shared function is read once.
	seen map[string]bool
}

// body reads one scope: its assignments with their properties, its returns,
// and the functions declared inside it, which are reached with it (a thunk
// among them is what a caller is handed).
func (w *reach) body(pkg *program.Package, file *program.PackageFile, assigns []*program.Assignment, returns []*program.Operand, fnTypes []*program.FnType) bool {
	for _, a := range assigns {
		if a == nil {
			continue
		}
		if w.operand(pkg, file, a.Expression) {
			return true
		}
		for _, v := range a.Variables {
			if v == nil {
				continue
			}
			for _, prop := range v.Properties {
				if prop != nil && w.operand(pkg, file, prop.Expression) {
					return true
				}
			}
		}
	}
	for _, r := range returns {
		if w.operand(pkg, file, r) {
			return true
		}
	}
	for _, ft := range fnTypes {
		if ft == nil {
			continue
		}
		for _, fg := range ft.Groups {
			if fg == nil {
				continue
			}
			for _, fn := range fg.Functions {
				if fn != nil && w.body(pkg, file, fn.Assignments, fn.Returns, fn.FnTypes) {
					return true
				}
			}
		}
	}
	return false
}

// operand reads one expression: a call of name is the answer, and every call,
// reference and property read is followed to what it names.
func (w *reach) operand(pkg *program.Package, file *program.PackageFile, op *program.Operand) bool {
	if op == nil {
		return false
	}
	switch v := op.T.(type) {
	case *program.Operand_FnCall:
		if v.FnCall == nil {
			return false
		}
		if ref := v.FnCall.Reference; ref != nil {
			if ref.Name == w.name || w.follow(pkg, file, ref) {
				return true
			}
		}
		for _, arg := range v.FnCall.Arguments {
			if w.operand(pkg, file, arg) {
				return true
			}
		}
	case *program.Operand_Reference:
		if v.Reference != nil && w.follow(pkg, file, v.Reference) {
			return true
		}
	case *program.Operand_PropRef:
		if v.PropRef != nil && v.PropRef.Reference != nil && w.follow(pkg, file, v.PropRef.Reference) {
			return true
		}
	case *program.Operand_Expression:
		if v.Expression == nil {
			return false
		}
		for _, sub := range v.Expression.Operands {
			if w.operand(pkg, file, sub) {
				return true
			}
		}
	}
	return false
}

// follow reads what a reference names: the functions of that name at every
// arity, and a top-level constant of it, in the package the reference is into.
// An unqualified name is the package's own or a dot import's; a qualified one
// is the import the prefix names. A name no package declares, a builtin or a
// local, is left alone.
func (w *reach) follow(pkg *program.Package, file *program.PackageFile, ref *program.Reference) bool {
	if ref.Prefix == "" {
		if w.declaration(pkg, ref.Name) {
			return true
		}
		for _, dep := range w.imports(file, ".") {
			if w.declaration(dep, ref.Name) {
				return true
			}
		}
		return false
	}
	for _, dep := range w.imports(file, ref.Prefix) {
		if w.declaration(dep, ref.Name) {
			return true
		}
	}
	return false
}

// imports resolves the imports of file known by prefix: "." for the dot
// imports, else the import written with that prefix. Nothing without
// getImport.
func (w *reach) imports(file *program.PackageFile, prefix string) []*program.Package {
	if w.getImport == nil || file == nil {
		return nil
	}
	var out []*program.Package
	for _, imp := range file.Imports {
		if imp == nil || imp.PrefixOrSimpleName() != prefix {
			continue
		}
		if dep, err := w.getImport(imp.Path); err == nil && dep != nil {
			out = append(out, dep)
		}
	}
	return out
}

// declaration reads the declarations of name in pkg, once: each function of
// the name, and a constant of it with its properties.
func (w *reach) declaration(pkg *program.Package, name string) bool {
	if pkg == nil {
		return false
	}
	key := pkg.Path + "\x00" + name
	if w.seen[key] {
		return false
	}
	w.seen[key] = true
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, ft := range f.FnTypes {
			if ft == nil || ft.Name != name {
				continue
			}
			for _, fg := range ft.Groups {
				if fg == nil {
					continue
				}
				for _, fn := range fg.Functions {
					if fn != nil && w.body(pkg, f, fn.Assignments, fn.Returns, fn.FnTypes) {
						return true
					}
				}
			}
		}
		for _, a := range f.Assignments {
			if a == nil {
				continue
			}
			for _, v := range a.Variables {
				if v == nil || v.Name != name {
					continue
				}
				if w.operand(pkg, f, a.Expression) {
					return true
				}
				for _, prop := range v.Properties {
					if prop != nil && w.operand(pkg, f, prop.Expression) {
						return true
					}
				}
			}
		}
	}
	return false
}
