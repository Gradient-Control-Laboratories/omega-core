package closure

import (
	"strings"

	"gradientcontrol.com/omega/core/program"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

// qualifyForeignBody readies a clone of a function from definerFile to live in
// hostFile, another package's file. specializeHigherOrderCall has to put such
// a clone where the function it was passed is defined, and there the clone's
// unqualified references to its own package would no longer resolve: a
// library function calling one of its own helpers would lose it. So each name
// the body leaves free that its own package declares is given the prefix that
// reaches that package from hostFile, and each prefix the body already carries
// is given the one that reaches the same package from there. A name the body
// binds itself, as a parameter, a local or a nested function, stays as it is.
func (t *transformer) qualifyForeignBody(fn *program.Function, definerFile, hostFile *program.PackageFile) {
	definer := t.filePkgs[definerFile]
	if definer == nil || definer == t.filePkgs[hostFile] {
		return
	}
	prefix, ok := t.prefixFrom(hostFile, definer.Path)
	if !ok {
		return
	}
	declared := make(map[string]bool)
	for name := range t.pkgFunctions[definer.Path] {
		declared[name] = true
	}
	for _, f := range definer.Files {
		for _, a := range f.Assignments {
			for _, v := range a.Variables {
				declared[v.Name] = true
			}
		}
	}
	q := &qualifier{t: t, definerFile: definerFile, hostFile: hostFile, prefix: prefix, declared: declared}
	q.function(fn, nil)
}

// prefixFrom is the prefix that reaches the package at path from file: the
// file's own import of it, or, for a system library, the prefix the linker
// imports it by when a file names it without importing it.
func (t *transformer) prefixFrom(file *program.PackageFile, path string) (string, bool) {
	if file != nil {
		for _, imp := range file.Imports {
			if imp.Path == path {
				// A dot import reaches the package's names without a prefix, so
				// they resolve as they are.
				if imp.Prefix == "." {
					return "", false
				}
				return imp.PrefixOrSimpleName(), true
			}
		}
	}
	simple := path[strings.LastIndex(path, "/")+1:]
	if systemPath, ok := gsllibs.PathForPrefix(simple); ok && systemPath == path {
		return simple, true
	}
	return "", false
}

type qualifier struct {
	t           *transformer
	definerFile *program.PackageFile
	hostFile    *program.PackageFile
	prefix      string
	declared    map[string]bool
}

// function qualifies fn's body, where bound holds the names its enclosing
// functions bind.
func (q *qualifier) function(fn *program.Function, bound map[string]bool) {
	inner := make(map[string]bool, len(bound))
	for name := range bound {
		inner[name] = true
	}
	for _, prm := range fn.Parameters {
		inner[prm.Name] = true
	}
	for _, a := range fn.Assignments {
		for _, v := range a.Variables {
			inner[v.Name] = true
		}
	}
	for _, ft := range fn.FnTypes {
		inner[ft.Name] = true
	}
	for _, a := range fn.Assignments {
		q.operand(a.Expression, inner)
	}
	for _, ret := range fn.Returns {
		q.operand(ret, inner)
	}
	for _, ft := range fn.FnTypes {
		for _, fg := range ft.Groups {
			for _, nested := range fg.Functions {
				q.function(nested, inner)
			}
		}
	}
}

func (q *qualifier) operand(op *program.Operand, bound map[string]bool) {
	if op == nil {
		return
	}
	switch v := op.T.(type) {
	case *program.Operand_Reference:
		q.reference(v.Reference, bound)
	case *program.Operand_FnCall:
		if v.FnCall != nil {
			q.reference(v.FnCall.Reference, bound)
			for _, arg := range v.FnCall.Arguments {
				q.operand(arg, bound)
			}
		}
	case *program.Operand_Expression:
		if v.Expression != nil {
			for _, sub := range v.Expression.Operands {
				q.operand(sub, bound)
			}
		}
	}
}

func (q *qualifier) reference(ref *program.Reference, bound map[string]bool) {
	if ref == nil {
		return
	}
	if ref.Prefix != "" {
		if bound[ref.Prefix] {
			return
		}
		path := q.t.resolvePrefix(q.definerFile, ref.Prefix)
		if prefix, ok := q.t.prefixFrom(q.hostFile, path); ok && prefix != ref.Prefix {
			ref.Prefix = prefix
			ref.RefPath = ""
			ref.PkgPath = ""
		}
		return
	}
	if bound[ref.Name] || !q.declared[ref.Name] || program.IsBuiltin(ref.Name) ||
		ref.Name == program.PVarName || ref.Name == program.TimeVarName {
		return
	}
	ref.Prefix = q.prefix
	ref.RefPath = ""
	ref.PkgPath = ""
}
