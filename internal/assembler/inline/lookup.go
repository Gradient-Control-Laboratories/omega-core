package inline

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/program"
)

// fileScopeOf finds the file of package pkgPath that declares fn, or nil.
func (in *inliner) fileScopeOf(fn *program.Function, pkgPath string) *scope.PkgFileScope {
	if in.progScope == nil {
		return nil
	}
	pkgScope := in.progScope.Dependencies[pkgPath]
	if in.prog != nil && in.prog.Primary != nil && in.prog.Primary.Path == pkgPath {
		pkgScope = in.progScope.Primary
	}
	if pkgScope == nil {
		return nil
	}
	for _, fs := range pkgScope.Files {
		if fs == nil || fs.File == nil {
			continue
		}
		for _, fnType := range fs.File.FnTypes {
			for _, group := range fnType.Groups {
				for _, f := range group.Functions {
					if f == fn {
						return fs
					}
				}
			}
		}
	}
	return nil
}

func (in *inliner) isNameShadowed(name string, targetFile *program.PackageFile, fileScope *scope.PkgFileScope) bool {
	if fileScope != nil {
		if _, ok := fileScope.Variables[name]; ok {
			return true
		}
		if _, ok := fileScope.Functions[name]; ok {
			return true
		}
	}
	if targetFile != nil {
		for _, a := range targetFile.Assignments {
			for _, v := range a.Variables {
				if v.Name == name {
					return true
				}
			}
		}
		for _, ft := range targetFile.FnTypes {
			if ft.Name == name {
				return true
			}
		}
	}
	return false
}

func (in *inliner) isBuiltinOrSpecialName(name string) bool {
	if strings.HasPrefix(name, program.AnonFnPrefix) || strings.HasPrefix(name, program.InlinedFnPrefix) {
		return false
	}
	if program.BuiltinNamed(name) != nil || name == "remap" ||
		name == program.PVarName || name == program.TimeVarName ||
		strings.HasPrefix(name, "gsl_") {
		return true
	}
	if in.progScope != nil {
		if builtinScope, ok := in.progScope.Dependencies["builtin"]; ok && builtinScope != nil {
			if elem := builtinScope.NamedElem(name); elem != nil {
				return true
			}
		}
	}
	return false
}

// isNativeOrSpecial reports whether fc calls something with no GSL body to
// look into: a native builtin (length, a noise oracle), remap, or p, time or a
// gsl_ name. Unlike isBuiltinOrSpecial it leaves out the builtin package's
// own functions, sphere and the rest, which are GSL with bodies.
func (in *inliner) isNativeOrSpecial(fc *program.FnCall) bool {
	if fc.Reference == nil {
		return true
	}
	if fc.Reference.Prefix != "" {
		return false
	}
	name := fc.Reference.Name
	if strings.HasPrefix(name, program.AnonFnPrefix) || strings.HasPrefix(name, program.InlinedFnPrefix) {
		return false
	}
	return program.BuiltinNamed(name) != nil || name == "remap" ||
		name == program.PVarName || name == program.TimeVarName ||
		strings.HasPrefix(name, "gsl_")
}

func (in *inliner) isBuiltinOrSpecial(fc *program.FnCall) bool {
	if fc.Reference == nil {
		return true
	}
	if fc.Reference.Prefix == "" {
		if in.isBuiltinOrSpecialName(fc.Reference.Name) {
			return true
		}
	}
	return false
}

func (in *inliner) findFunction(fc *program.FnCall, fileScope *scope.PkgFileScope, paramCount int) (*program.FnType, *program.Function, string) {
	if fc.Reference == nil {
		return nil, nil, ""
	}

	cacheKey := ""
	if in.fnCache != nil {
		var innermostEncFn *program.Function
		if len(in.enclosingFns) > 0 {
			innermostEncFn = in.enclosingFns[len(in.enclosingFns)-1]
		}
		if fileScope != nil {
			cacheKey = fmt.Sprintf("%p|%p|%s|%s|%d", fileScope, innermostEncFn, fc.Reference.Prefix, fc.Reference.Name, paramCount)
		} else {
			cacheKey = fmt.Sprintf("nil|%p|%s|%s|%d", innermostEncFn, fc.Reference.Prefix, fc.Reference.Name, paramCount)
		}
		if res, ok := in.fnCache[cacheKey]; ok {
			if res.found {
				return res.fnType, res.fn, res.pkgPath
			}
			return nil, nil, ""
		}
	}

	fnType, fn, pkgPath := in.findFunctionUncached(fc, fileScope, paramCount)
	if in.fnCache != nil && cacheKey != "" {
		in.fnCache[cacheKey] = fnLookup{
			fnType:  fnType,
			fn:      fn,
			pkgPath: pkgPath,
			found:   fn != nil,
		}
	}
	return fnType, fn, pkgPath
}

func (in *inliner) findFunctionUncached(fc *program.FnCall, fileScope *scope.PkgFileScope, paramCount int) (*program.FnType, *program.Function, string) {
	// First check local function scopes in innermost-to-outermost lexical order
	if fc.Reference.Prefix == "" && len(in.enclosingFns) > 0 {
		for i := len(in.enclosingFns) - 1; i >= 0; i-- {
			encFn := in.enclosingFns[i]
			for _, fnType := range encFn.FnTypes {
				if fnType.Name == fc.Reference.Name {
					for _, group := range fnType.Groups {
						if int(group.ParamCount) == paramCount && len(group.Functions) > 0 {
							pkgPath := ""
							if fileScope != nil && fileScope.PkgScope() != nil && fileScope.PkgScope().Package != nil {
								pkgPath = fileScope.PkgScope().Package.Path
							}
							return fnType, group.Functions[0], pkgPath
						}
					}
				}
			}
		}
	}

	if fileScope != nil {
		if fc.Reference.Prefix != "" {
			if imp, ok := fileScope.Imports[fc.Reference.Prefix]; ok && imp != nil {
				targetPkgScope := imp.PkgScope()
				if targetPkgScope != nil && targetPkgScope.Package != nil {
					for _, file := range targetPkgScope.Package.Files {
						for _, fnType := range file.FnTypes {
							if fnType.Name == fc.Reference.Name {
								for _, group := range fnType.Groups {
									if int(group.ParamCount) == paramCount && len(group.Functions) > 0 {
										return fnType, group.Functions[0], targetPkgScope.Package.Path
									}
								}
							}
						}
					}
				}
			}
		} else {
			// Check current package
			targetPkgScope := fileScope.PkgScope()
			if targetPkgScope != nil && targetPkgScope.Package != nil {
				for _, file := range targetPkgScope.Package.Files {
					for _, fnType := range file.FnTypes {
						if fnType.Name == fc.Reference.Name {
							for _, group := range fnType.Groups {
								if int(group.ParamCount) == paramCount && len(group.Functions) > 0 {
									return fnType, group.Functions[0], targetPkgScope.Package.Path
								}
							}
						}
					}
				}
			}
			// Check dot-imports
			for _, dotImp := range fileScope.DotImports {
				if dotImp != nil && dotImp.PkgScope() != nil && dotImp.PkgScope().Package != nil {
					for _, file := range dotImp.PkgScope().Package.Files {
						for _, fnType := range file.FnTypes {
							if fnType.Name == fc.Reference.Name {
								for _, group := range fnType.Groups {
									if int(group.ParamCount) == paramCount && len(group.Functions) > 0 {
										return fnType, group.Functions[0], dotImp.PkgScope().Package.Path
									}
								}
							}
						}
					}
				}
			}
		}
	}

	if fc.Reference.Prefix != "" {
		if in.progScope != nil {
			for depPath, depScope := range in.progScope.Dependencies {
				pkgParts := strings.Split(depPath, "/")
				if depPath == fc.Reference.Prefix || pkgParts[len(pkgParts)-1] == fc.Reference.Prefix {
					if depScope != nil && depScope.Package != nil {
						for _, file := range depScope.Package.Files {
							for _, fnType := range file.FnTypes {
								if fnType.Name == fc.Reference.Name {
									for _, group := range fnType.Groups {
										if int(group.ParamCount) == paramCount && len(group.Functions) > 0 {
											return fnType, group.Functions[0], depScope.Package.Path
										}
									}
								}
							}
						}
					}
				}
			}
		}
		return nil, nil, ""
	}

	// Fallback to Primary package
	if in.progScope != nil {
		targetPkgScope := in.progScope.Primary
		calleePkgPath := ""
		if in.prog.Primary != nil {
			calleePkgPath = in.prog.Primary.Path
		}

		if targetPkgScope != nil && targetPkgScope.Package != nil {
			for _, file := range targetPkgScope.Package.Files {
				for _, fnType := range file.FnTypes {
					if fnType.Name == fc.Reference.Name {
						for _, group := range fnType.Groups {
							if int(group.ParamCount) == paramCount && len(group.Functions) > 0 {
								return fnType, group.Functions[0], calleePkgPath
							}
						}
					}
				}
			}
		}
	}

	return nil, nil, ""
}

// findCallbackFunction resolves a higher-order operation's callback operand (e.g. from
// Operation_REDUCE or Operation_MAP) to its underlying Function AST.
func (in *inliner) findCallbackFunction(cbOp *program.Operand, fileScope *scope.PkgFileScope) *program.Function {
	if cbOp == nil {
		return nil
	}
	var ref *program.Reference
	switch t := cbOp.T.(type) {
	case *program.Operand_Reference:
		ref = t.Reference
	case *program.Operand_FnCall:
		ref = t.FnCall.Reference
	}
	if ref == nil {
		return nil
	}

	// 1. Check enclosing functions from innermost to outermost lexical scope.
	for i := len(in.enclosingFns) - 1; i >= 0; i-- {
		enc := in.enclosingFns[i]
		for _, ft := range enc.FnTypes {
			if ft.Name == ref.Name {
				for _, grp := range ft.Groups {
					if len(grp.Functions) > 0 {
						return grp.Functions[0]
					}
				}
			}
		}
	}

	// 2. Check fileScope imports if reference has a package prefix.
	if fileScope != nil && ref.Prefix != "" {
		if imp, ok := fileScope.Imports[ref.Prefix]; ok && imp != nil && imp.PkgScope() != nil && imp.PkgScope().Package != nil {
			for _, file := range imp.PkgScope().Package.Files {
				for _, ft := range file.FnTypes {
					if ft.Name == ref.Name {
						for _, grp := range ft.Groups {
							if len(grp.Functions) > 0 {
								return grp.Functions[0]
							}
						}
					}
				}
			}
		}
	}

	// 3. Check current fileScope package files.
	if fileScope != nil && fileScope.PkgScope() != nil && fileScope.PkgScope().Package != nil {
		for _, file := range fileScope.PkgScope().Package.Files {
			for _, ft := range file.FnTypes {
				if ft.Name == ref.Name {
					for _, grp := range ft.Groups {
						if len(grp.Functions) > 0 {
							return grp.Functions[0]
						}
					}
				}
			}
		}
	}

	// 4. Check primary package files.
	if in.prog != nil && in.prog.Primary != nil {
		for _, file := range in.prog.Primary.Files {
			for _, ft := range file.FnTypes {
				if ft.Name == ref.Name {
					for _, grp := range ft.Groups {
						if len(grp.Functions) > 0 {
							return grp.Functions[0]
						}
					}
				}
			}
		}
	}

	// 5. Check dependency package files.
	if in.prog != nil {
		for _, dep := range in.prog.Dependencies {
			pkgParts := strings.Split(dep.Path, "/")
			if ref.Prefix != "" && dep.Path != ref.Prefix && pkgParts[len(pkgParts)-1] != ref.Prefix {
				continue
			}
			for _, file := range dep.Files {
				for _, ft := range file.FnTypes {
					if ft.Name == ref.Name {
						for _, grp := range ft.Groups {
							if len(grp.Functions) > 0 {
								return grp.Functions[0]
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (in *inliner) hasUnresolvedFreeVars(fnType *program.FnType, env map[string]*program.Operand, fileScope *scope.PkgFileScope, calleePkgPath string) bool {
	var calleeScope *scope.PkgScope
	if in.progScope != nil {
		if in.prog.Primary != nil && in.prog.Primary.Path == calleePkgPath {
			calleeScope = in.progScope.Primary
		} else if depScope, ok := in.progScope.Dependencies[calleePkgPath]; ok {
			calleeScope = depScope
		}
	}

	for _, grp := range fnType.Groups {
		for _, fn := range grp.Functions {
			localNames := make(map[string]bool)
			for _, p := range fn.Parameters {
				localNames[p.Name] = true
			}
			for _, a := range fn.Assignments {
				for _, v := range a.Variables {
					localNames[v.Name] = true
				}
			}
			for _, subFt := range fn.FnTypes {
				localNames[subFt.Name] = true
			}
			var checkOp func(op *program.Operand) bool
			checkOp = func(op *program.Operand) bool {
				if op == nil {
					return false
				}
				switch t := op.T.(type) {
				case *program.Operand_Reference:
					if t.Reference.Prefix == "" {
						name := t.Reference.Name
						if in.isBuiltinOrSpecialName(name) || localNames[name] {
							return false
						}
						if env != nil {
							if repl, ok := env[name]; ok && repl != nil {
								if in.isHoistableConst(repl) {
									return false
								}
								return true
							}
						}
						if calleeScope != nil && calleeScope.NamedElem(name) != nil {
							return false
						}
						if fileScope != nil && fileScope.NamedElem(name) != nil {
							return false
						}
						return true
					}
				case *program.Operand_Expression:
					for i, child := range t.Expression.Operands {
						if t.Expression.Operation == program.Operation_ACCESS && i > 0 {
							continue
						}
						if checkOp(child) {
							return true
						}
					}
				case *program.Operand_FnCall:
					if t.FnCall.Reference != nil && t.FnCall.Reference.Prefix == "" {
						name := t.FnCall.Reference.Name
						if !in.isBuiltinOrSpecialName(name) && !localNames[name] {
							if env != nil {
								if repl, ok := env[name]; ok && repl != nil {
									if !in.isHoistableConst(repl) {
										return true
									}
								}
							}
							if (calleeScope == nil || calleeScope.NamedElem(name) == nil) &&
								(fileScope == nil || fileScope.NamedElem(name) == nil) {
								return true
							}
						}
					}
					for _, arg := range t.FnCall.Arguments {
						if checkOp(arg) {
							return true
						}
					}
				}
				return false
			}

			for _, a := range fn.Assignments {
				if checkOp(a.Expression) {
					return true
				}
			}
			for _, r := range fn.Returns {
				if checkOp(r) {
					return true
				}
			}
			for _, subFt := range fn.FnTypes {
				subEnv := make(map[string]*program.Operand)
				for k, v := range env {
					subEnv[k] = v
				}
				for k := range localNames {
					subEnv[k] = &program.Operand{}
				}
				if in.hasUnresolvedFreeVars(subFt, subEnv, fileScope, calleePkgPath) {
					return true
				}
			}
		}
	}
	return false
}
