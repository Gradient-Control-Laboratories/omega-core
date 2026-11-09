package closure

import (
	"gradientcontrol.com/omega/core/program"
)

func (t *transformer) resolvePrefix(file *program.PackageFile, prefix string) string {
	if prefix == "" {
		return ""
	}
	if file != nil {
		for _, imp := range file.Imports {
			if imp.Prefix == prefix || imp.SimpleName() == prefix {
				return imp.Path
			}
		}
	}
	return prefix
}

func (t *transformer) lookupFunction(file *program.PackageFile, prefix, name string, paramCount int) *program.Function {
	_, ft := t.lookupFnType(file, prefix, name)
	if ft == nil {
		return nil
	}

	for _, fg := range ft.Groups {
		if fg.ParamCount == int32(paramCount) || paramCount < 0 {
			if len(fg.Functions) > 0 {
				return fg.Functions[0]
			}
		}
	}
	return nil
}

func (t *transformer) lookupFnTypeInScope(enclosingFn *program.Function, file *program.PackageFile, prefix, name string) (*program.PackageFile, *program.FnType) {
	if enclosingFn != nil && prefix == "" {
		for _, nestedFt := range enclosingFn.FnTypes {
			if nestedFt.Name == name {
				return file, nestedFt
			}
		}
	}
	return t.lookupFnType(file, prefix, name)
}

func (t *transformer) lookupFnType(file *program.PackageFile, prefix, name string) (*program.PackageFile, *program.FnType) {
	pkgPath := t.resolvePrefix(file, prefix)
	if pkgPath == "" {
		if fileMap, ok := t.fileFunctions[file]; ok {
			if ft, ok := fileMap[name]; ok {
				return file, ft
			}
		}
		if file != nil {
			pkg := t.filePkgs[file]
			if pkg != nil {
				for _, f := range pkg.Files {
					if fileMap, ok := t.fileFunctions[f]; ok {
						if ft, ok := fileMap[name]; ok {
							return f, ft
						}
					}
				}
			}
			// An unqualified name reaches another package only through a dot
			// import, as it does in the scope; a prefixed import's functions
			// need their prefix.
			for _, imp := range file.Imports {
				if imp.PrefixOrSimpleName() != "." {
					continue
				}
				if dep, ok := t.pkgByPath[imp.Path]; ok {
					for _, f := range dep.Files {
						if fileMap, ok := t.fileFunctions[f]; ok {
							if ft, ok := fileMap[name]; ok {
								return f, ft
							}
						}
					}
				}
			}
		}
		if dep, ok := t.pkgByPath["builtin"]; ok {
			for _, f := range dep.Files {
				if fileMap, ok := t.fileFunctions[f]; ok {
					if ft, ok := fileMap[name]; ok {
						return f, ft
					}
				}
			}
		}
	} else {
		if dep, ok := t.pkgByPath[pkgPath]; ok {
			for _, f := range dep.Files {
				if fileMap, ok := t.fileFunctions[f]; ok {
					if ft, ok := fileMap[name]; ok {
						return f, ft
					}
				}
			}
		}
	}
	return nil, nil
}

func (t *transformer) findParamCallArgCount(fn *program.Function, paramName string) int {
	if fn == nil {
		return -1
	}
	for _, assign := range fn.Assignments {
		if c := t.operandCallArgCount(assign.Expression, paramName); c >= 0 {
			return c
		}
	}
	for _, ret := range fn.Returns {
		if c := t.operandCallArgCount(ret, paramName); c >= 0 {
			return c
		}
	}
	return -1
}

func (t *transformer) operandCallArgCount(op *program.Operand, paramName string) int {
	if op == nil {
		return -1
	}
	switch v := op.T.(type) {
	case *program.Operand_FnCall:
		if v.FnCall != nil && v.FnCall.Reference != nil && v.FnCall.Reference.Prefix == "" && v.FnCall.Reference.Name == paramName {
			return len(v.FnCall.Arguments)
		}
		if v.FnCall != nil {
			for _, arg := range v.FnCall.Arguments {
				if c := t.operandCallArgCount(arg, paramName); c >= 0 {
					return c
				}
			}
		}
	case *program.Operand_Expression:
		if v.Expression != nil {
			for _, sub := range v.Expression.Operands {
				if c := t.operandCallArgCount(sub, paramName); c >= 0 {
					return c
				}
			}
		}
	}
	return -1
}
