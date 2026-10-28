package codegen

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) NewFnVar(name string, decl *program.FnType, enclosingScope *build.Scope) build.VarInfo {
	fnProgType := program.NewFunctionType("", name)
	expr := wgsl.NewOpaqueExpr(&program.Operand{Type: fnProgType}, wgsl.TypeInfo{}, build.PrecedenceHighest, name, name)
	return wgsl.NewVar(enclosingScope, name, expr, wgsl.TypeInfo{})
}

// Build the top-level WGSL function name from the given function name and enclosing scope
func (d *Delegate) buildFnName(fnName string, enclosingScope *build.Scope) string {
	name := d.prefixWithScopeNames(enclosingScope, fnName)
	name = wgsl.SanitizeFnName(name)

	// all WGSL functions are at the top level, so ensure it is unique

	return d.uniqueTopLevelName(name)
}

// Prefix the given name with the names of the scopes.
// This does not check if the name is unique or a valid WGSL identifier.
func (d *Delegate) prefixWithScopeNames(bScope *build.Scope, name string) string {
	for {
		if bScope == d.rootBScope || bScope.Name() == "" {
			return wgsl.SanitizeFnName(name)
		}
		scopeName := bScope.Name()
		if !bScope.IsFileScope() {
			name = wgsl.SanitizeIdent(scopeName[strings.LastIndex(scopeName, "/")+1:]) + "_" + name
		}
		bScope = bScope.Parent()
	}
}

// Return top level name that has not been assigned before.
// This does not check variable and function definitions.
func (d *Delegate) uniqueTopLevelName(base string) string {
	origBase := base
	if idx := strings.LastIndex(base, "_"); idx != -1 && idx < len(base)-1 {
		isDigits := true
		for _, r := range base[idx+1:] {
			if r < '0' || r > '9' {
				isDigits = false
				break
			}
		}
		if isDigits {
			base = base[:idx]
		}
	}
	name := origBase
	count := 1
	assignedNames := d.getAssignedNamesMap()
	for assignedNames[name] {
		count++
		name = fmt.Sprintf("%s_%d", base, count)
	}
	assignedNames[name] = true
	return name
}

func (d *Delegate) uniqueLocalName(base string) string {
	origBase := base
	if idx := strings.LastIndex(base, "_"); idx != -1 && idx < len(base)-1 {
		isDigits := true
		for _, r := range base[idx+1:] {
			if r < '0' || r > '9' {
				isDigits = false
				break
			}
		}
		if isDigits {
			base = base[:idx]
		}
	}
	name := origBase
	count := 1
	assignedNames := d.getLocalAssignedNamesMap()
	for assignedNames[name] {
		count++
		name = fmt.Sprintf("%s_%d", base, count)
	}
	assignedNames[name] = true
	return name
}

// Return the scope associated with the given imported package name
// excluding dot-imported packages. Return nil if not found.
func (d *Delegate) getScope(pkgName string) *Scope {
	return d.getScopeFrom(d.currentScope(), pkgName)
}

// getScopeFrom searches for an imported package scope matching the prefix pkgName.
// It starts from the start scope and traverses up the parent scope hierarchy, checking
// the prefixImpCache of each scope along the way. Returns nil if not found.
func (d *Delegate) getScopeFrom(start *Scope, pkgName string) *Scope {
	s := start
	if s == nil {
		return nil
	}
	if pkgName == "" {
		return s
	}

	for s != nil {
		if s.prefixImpCache != nil {
			if impScope, ok := s.prefixImpCache[pkgName]; ok {
				return impScope
			}
		}
		if s == d.rootScope() {
			break
		}
		s = s.parent
	}
	return nil
}

// findEnclosingScope walks up from the current scope or package scope to find
// the delegate scope that contains the declaration for given function type.
func (d *Delegate) findEnclosingScope(fnDecl *program.FnType, pkgName string) *Scope {
	fnName := fnDecl.Name

	var s *Scope
	if pkgName != "" {
		if cur := d.currentScope(); cur != nil {
			importing := cur.importingScope
			for importing != nil && importing.PkgPath() == cur.PkgPath() {
				importing = importing.importingScope
			}
			if importing != nil {
				s = d.getScopeFrom(importing, pkgName)
			} else {
				s = d.getScope(pkgName)
			}
		} else {
			s = d.getScope(pkgName)
		}

		for s != nil {
			if s.bScope.ExistingFnDecl(fnName) != nil {
				retScope := s
				for retScope.parent != nil && retScope.parent != d.rootScope() {
					retScope = retScope.parent
				}
				return retScope
			}
			s = s.parent
		}
		return d.rootScope()
	}

	// pkgName == ""
	fnScope := d.GetScope().FindFnScope(fnName)
	if fnScope == nil {
		fnScope = d.rootBScope
	}
	callSitePkgPath := ""
	if d.prevBScope != nil {
		if agScope, ok := d.prevBScope.GetAgentScope().(*Scope); ok && agScope != nil {
			callSitePkgPath = agScope.PkgPath()
		}
	}
	fnPkgPath := d.ensureAgentScope(fnScope).PkgPath()

	retScope := d.rootScope()
	if callSitePkgPath == fnPkgPath {
		retScope = d.ensureAgentScope(fnScope)
	}
	for retScope.parent != nil && retScope.parent != d.rootScope() {
		retScope = retScope.parent
	}
	return retScope
}

func (d *Delegate) addPredefinedFunc(fnName string, args []build.Expression, returnType wgsl.TypeInfo, funcSrc string) error {

	// Get or create FnType in the root scope

	fnDecl := d.rootBScope.ExistingFnDecl(fnName)
	var fn *program.Function
	if fnDecl == nil {
		argNames := make([]string, len(args))
		for i, arg := range args {
			// The context used when building the function body
			funcCtx := NewBuildContext(d, d.rootPVar)
			argNames[i] = arg.(wgsl.Expression).Codegen(funcCtx)
		}
		params := make([]*program.Parameter, len(argNames))
		for i, an := range argNames {
			params[i] = &program.Parameter{Name: an}
		}
		fn = &program.Function{
			Parameters: params,
		}
		decl := &program.FnType{
			Name: fnName,
			Groups: []*program.FnGroup{
				{
					ParamCount: int32(len(args)),
					Functions:  []*program.Function{fn},
				},
			},
		}
		var err error
		fnDecl, err = d.rootBScope.RegisterFnDecl(fnName, decl, d.NewFnVar, d.GetScope())
		if err != nil {
			return program.InternalErr(err)
		}
	} else {
		for _, g := range fnDecl.Groups {
			if int(g.ParamCount) == len(args) && len(g.Functions) > 0 {
				fn = g.Functions[0]
				break
			}
		}
	}

	// Check if the function has already been built for these argument types
	fnInfo := d.rootBScope.InfoForArgs(fn, args)
	if fnInfo != nil {
		return nil
	}

	var paramNames []string
	if fn != nil {
		paramNames = make([]string, len(fn.Parameters))
		for i, p := range fn.Parameters {
			paramNames[i] = p.Name
		}
	}

	// Create new FnInfo
	fnWgslName := d.buildFnName(fnName, d.rootBScope)
	wgslFnInfo := wgsl.NewFnInfo(fnName, fnWgslName, paramNames, args)

	// Mark built and make unique
	wgslFnInfo.Built(nil, returnType, nil)

	// Add to the common build scope
	err := d.rootBScope.AddFnInfo(fn, wgslFnInfo)
	if err != nil {
		return program.InternalErr(err)
	}

	// Append content to root and register on module
	d.module.AddPredefined(funcSrc)
	return nil
}
