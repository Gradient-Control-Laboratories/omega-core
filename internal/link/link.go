package link

import (
	"errors"
	"fmt"
	"sort"

	"gradientcontrol.com/omega/core/assembler/source"
	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

var errIsPropRef = errors.New("is property reference")

// References resolves all symbol references in-place, discarding diagnostics.
// Used for internal re-linking throughout the assembler pipeline.
func References(
	origProg *program.Program,
	srcMgr source.Manager,
	parsePkg func(sources []string) (*parser.Result, error),
) error {
	_, err := ReferencesAndDiagnostics(origProg, srcMgr, parsePkg)
	return err
}

// ReferencesAndDiagnostics resolves references and returns compiler diagnostics.
// Used at Step 1 of the assembler pipeline.
func ReferencesAndDiagnostics(
	origProg *program.Program,
	srcMgr source.Manager,
	parsePkg func(sources []string) (*parser.Result, error),
) ([]program.Diagnostic, error) {
	progScope, err := scope.Init(origProg)
	if err != nil {
		return nil, err
	}

	sink := &diagnosticSink{}
	l := newLinker(origProg, progScope, srcMgr, parsePkg, sink)
	err = l.link()
	return sink.diags, err
}

// linker resolves references and performs diagnostic collection.
type linker struct {
	prog      *program.Program
	progScope *scope.ProgScope
	srcMgr    source.Manager
	parsePkg  func(sources []string) (*parser.Result, error)
	sink      *diagnosticSink
}

func newLinker(
	origProg *program.Program,
	progScope *scope.ProgScope,
	srcMgr source.Manager,
	parsePkg func(sources []string) (*parser.Result, error),
	sink *diagnosticSink,
) *linker {
	return &linker{
		prog:      origProg,
		progScope: progScope,
		srcMgr:    srcMgr,
		parsePkg:  parsePkg,
		sink:      sink,
	}
}

func (l *linker) link() error {
	// Phase 1: Reference Resolution for primary package
	// Invariant: scope.Init guarantees progScope.Primary is non-nil
	if err := l.linkPkg(l.progScope.Primary); err != nil {
		return err
	}

	// Link dependent packages (index-based to accommodate dynamically added packages)
	for i := 0; i < len(l.prog.Dependencies); i++ {
		depPkg := l.prog.Dependencies[i]
		// Invariant: scope.Init and AddDependency guarantee all program dependencies have non-nil scopes
		depScope := l.progScope.Dependencies[depPkg.Path]
		if err := l.linkPkg(depScope); err != nil {
			return err
		}
	}

	return nil
}

func fnGroupLocation(group *program.FnGroup) *program.Location {
	// Invariant: Parser guarantees every FnGroup has at least one Function
	f := group.Functions[0]
	if len(f.Parameters) > 0 && f.Parameters[0].Start != nil {
		return f.Parameters[0].Start
	}
	if len(f.Returns) > 0 && f.Returns[0].Start != nil {
		return f.Returns[0].Start
	}
	return nil
}

func (l *linker) checkDotImportCollisions(fileScope *scope.PkgFileScope) {
	if len(fileScope.DotImports) <= 1 {
		return
	}

	type dotEntry struct {
		imp      *program.Import
		pkgScope *scope.PkgScope
	}
	var entries []dotEntry
	for _, imp := range fileScope.File.Imports {
		if imp.PrefixOrSimpleName() == "." {
			if depScope := l.progScope.Dependencies[imp.Path]; depScope != nil {
				entries = append(entries, dotEntry{imp: imp, pkgScope: depScope})
			}
		}
	}

	for i := 1; i < len(entries); i++ {
		curr := entries[i]
		if curr.pkgScope.Package.Path == "builtin" {
			continue
		}
		for j := 0; j < i; j++ {
			prev := entries[j]
			if prev.pkgScope.Package.Path == "builtin" || prev.pkgScope == curr.pkgScope {
				continue
			}

			// Variable collisions
			for _, currFile := range curr.pkgScope.Files {
				for varName := range currFile.Variables {
					if elem := prev.pkgScope.NamedElem(varName); elem != nil {
						if _, isVar := elem.(*scope.Var); isVar {
							l.sink.addWarning(
								curr.imp.Start,
								fileScope.Pkg.Name(),
								fileScope.File.Path,
								fmt.Sprintf(
									"variable %q declared in both import %q and %q",
									varName,
									prev.pkgScope.Package.Path,
									curr.pkgScope.Package.Path,
								),
							)
						}
					}
				}
			}

			// Function collisions
			for _, currFile := range curr.pkgScope.Files {
				for fnName, fnTypeScope := range currFile.Functions {
					if prevElem := prev.pkgScope.NamedElem(fnName); prevElem != nil {
						if prevFnScope, ok := prevElem.(*scope.FnTypeScope); ok {
							for paramCount := range fnTypeScope.Groups {
								if _, ok := prevFnScope.Groups[paramCount]; ok {
									l.sink.addWarning(
										curr.imp.Start,
										fileScope.Pkg.Name(),
										fileScope.File.Path,
										fmt.Sprintf(
											"function %q with %d parameters declared in both import %q and %q",
											fnName,
											paramCount,
											prev.pkgScope.Package.Path,
											curr.pkgScope.Package.Path,
										),
									)
								}
							}
						}
					}
				}
			}
		}
	}
}

func (l *linker) checkLocalDeclarationCollisions(fileScope *scope.PkgFileScope) error {
	hasExplicitDotImport := false
	for _, dotImp := range fileScope.DotImports {
		// Invariant: scope.Init guarantees every PkgRef in DotImports has a non-nil PkgScope and Package
		if dotImp.PkgScope().Package.Path != "builtin" {
			hasExplicitDotImport = true
			break
		}
	}
	if !hasExplicitDotImport {
		return nil
	}

	collidesWithDotImport := func(name string) bool {
		for _, dotImp := range fileScope.DotImports {
			pkg := dotImp.PkgScope()
			if pkg.Package.Path != "builtin" {
				if elem := pkg.NamedElem(name); elem != nil {
					switch elem.(type) {
					case *scope.Var, *scope.FnTypeScope:
						return true
					}
				}
			}
		}
		return false
	}

	// Check file-level assignments (variables)
	for _, assign := range fileScope.File.Assignments {
		for _, v := range assign.Variables {
			if collidesWithDotImport(v.Name) {
				return fileScope.Err(assign.Start, fmt.Errorf("variable or function named %q already defined", v.Name))
			}
		}
	}

	// Check file-level functions
	for _, fnType := range fileScope.File.FnTypes {
		if collidesWithDotImport(fnType.Name) {
			var loc *program.Location
			if len(fnType.Groups) > 0 {
				loc = fnGroupLocation(fnType.Groups[0])
			}
			return fileScope.Err(loc, fmt.Errorf("variable or function named %q already defined", fnType.Name))
		}
	}

	return nil
}

func (l *linker) warnShadowedBuiltin(fileScope *scope.PkgFileScope, group *program.FnGroup, fnName string, pc int) {
	loc := fnGroupLocation(group)
	l.sink.addWarning(
		loc,
		l.prog.Primary.Path,
		fileScope.File.Path,
		fmt.Sprintf(
			"function %q with %d parameters shadows implicitly imported builtin function",
			fnName,
			pc,
		),
	)
}

func (l *linker) checkBuiltinShadowing(fileScope *scope.PkgFileScope) {
	if fileScope.PkgScope() != l.progScope.Primary {
		return
	}
	builtinScope := l.progScope.Dependencies["builtin"]

	for _, fnType := range fileScope.File.FnTypes {
		for _, group := range fnType.Groups {
			pc := int(group.ParamCount)
			if prim := program.PrimitiveFuncNamed(fnType.Name); prim != nil {
				if pc >= prim.MinArgs && pc <= prim.MaxArgs {
					l.warnShadowedBuiltin(fileScope, group, fnType.Name, pc)
					continue
				}
			}
			if builtin := program.BuiltinNamed(fnType.Name); builtin != nil {
				if pc >= builtin.MinArgs && (builtin.MaxArgs == 0 || pc <= builtin.MaxArgs) {
					l.warnShadowedBuiltin(fileScope, group, fnType.Name, pc)
					continue
				}
			}
			if builtinScope != nil {
				if decl := builtinScope.NamedElem(fnType.Name); decl != nil {
					if fnScope, ok := decl.(*scope.FnTypeScope); ok {
						if _, ok := fnScope.Groups[int32(pc)]; ok {
							l.warnShadowedBuiltin(fileScope, group, fnType.Name, pc)
						}
					}
				}
			}
		}
	}
}

func (l *linker) linkPkg(pkgScope *scope.PkgScope) error {
	for _, fileScope := range pkgScope.Files {
		if err := l.linkFile(fileScope); err != nil {
			return err
		}
	}
	return nil
}

func (l *linker) linkFile(fileScope *scope.PkgFileScope) error {
	origCount := len(fileScope.File.Imports)

	// Check dot-import collisions, local declaration collisions, and builtin shadowing
	l.checkDotImportCollisions(fileScope)
	if err := l.checkLocalDeclarationCollisions(fileScope); err != nil {
		return err
	}
	l.checkBuiltinShadowing(fileScope)

	visitor := &linkVisitor{
		l:         l,
		currScope: fileScope,
	}

	// Link assignments in the file
	for _, assign := range fileScope.File.Assignments {
		if _, err := program.VisitOperand(assign.Expression, visitor); err != nil {
			return err
		}
		for _, v := range assign.Variables {
			for _, prop := range v.Properties {
				if _, err := program.VisitOperand(prop.Expression, visitor); err != nil {
					return err
				}
			}
		}
	}

	// Link returns in the file
	for _, ret := range fileScope.File.Returns {
		if _, err := program.VisitOperand(ret, visitor); err != nil {
			return err
		}
	}
	for _, ret := range fileScope.File.ColorReturns {
		if _, err := program.VisitOperand(ret, visitor); err != nil {
			return err
		}
	}
	for _, ret := range fileScope.File.MaterialReturns {
		if _, err := program.VisitOperand(ret, visitor); err != nil {
			return err
		}
	}

	// Link package-level functions in the file
	for _, typeScope := range fileScope.Functions {
		for _, groupScope := range typeScope.Groups {
			for _, fnScope := range groupScope.Functions {
				if err := l.linkFn(fnScope); err != nil {
					return err
				}
			}
		}
	}

	// Sort dynamically added imports alphabetically to keep code generation order deterministic
	if len(fileScope.File.Imports) > origCount {
		autoImports := fileScope.File.Imports[origCount:]
		sort.Slice(autoImports, func(i, j int) bool {
			return autoImports[i].Path < autoImports[j].Path
		})
		fileScope.File.Imports = append(fileScope.File.Imports[:origCount], autoImports...)
	}

	return nil
}

func (l *linker) linkFn(fnScope *scope.FnScope) error {
	// Initialize local function scope parameters, variables, nested functions
	fnScope.NamedElem("")

	visitor := &linkVisitor{
		l:         l,
		currScope: fnScope,
	}

	// Link assignments in the function
	for _, assign := range fnScope.Function.Assignments {
		if _, err := program.VisitOperand(assign.Expression, visitor); err != nil {
			return err
		}
		for _, v := range assign.Variables {
			for _, prop := range v.Properties {
				if _, err := program.VisitOperand(prop.Expression, visitor); err != nil {
					return err
				}
			}
		}
	}

	// Link returns in the function
	for _, ret := range fnScope.Function.Returns {
		if _, err := program.VisitOperand(ret, visitor); err != nil {
			return err
		}
	}

	// Link nested functions recursively
	for _, typeScope := range fnScope.Functions {
		for _, groupScope := range typeScope.Groups {
			for _, nestedFnScope := range groupScope.Functions {
				if err := l.linkFn(nestedFnScope); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func isLexicalSymbol(s scope.Scope, name string) bool {
	for curr := s; curr != nil; curr = curr.ContainingScope() {
		switch sc := curr.(type) {
		case *scope.FnScope:
			sc.NamedElem("")
			if sc.Parameters != nil && sc.Parameters[name] != nil {
				return true
			}
			if sc.Variables != nil && sc.Variables[name] != nil {
				return true
			}
			if sc.Functions != nil && sc.Functions[name] != nil {
				return true
			}
		case *scope.PkgFileScope:
			if sc.Variables != nil && sc.Variables[name] != nil {
				return true
			}
			if sc.Functions != nil && sc.Functions[name] != nil {
				return true
			}
		case *scope.PkgScope:
			if elem := sc.NamedElem(name); elem != nil {
				if _, isVar := elem.(*scope.Var); isVar {
					return true
				}
				if _, isFn := elem.(*scope.FnTypeScope); isFn {
					return true
				}
			}
		}
	}
	return false
}

type linkVisitor struct {
	l         *linker
	currScope scope.Scope
}

func (v *linkVisitor) VisitValue(op *program.Operand, value float64) (any, error) {
	return nil, nil
}

func (v *linkVisitor) VisitReference(op *program.Operand, ref *program.Reference) (any, error) {
	if ref.Prefix == "" && !program.IsBuiltin(ref.Name) && ref.Name != program.PVarName && ref.Name != program.TimeVarName && !isTopologyBuiltin(ref.Name) && ref.Name != "auto" && !isLexicalSymbol(v.currScope, ref.Name) {
		fileScope := v.currScope.PkgFileScope()
		if fileScope != nil && len(fileScope.DotImports) > 1 {
			var matching []string
			for _, dotImp := range fileScope.DotImports {
				pkg := dotImp.PkgScope()
				if pkg.Package.Path != "builtin" {
					if elem := pkg.NamedElem(ref.Name); elem != nil {
						if _, isVar := elem.(*scope.Var); isVar {
							matching = append(matching, pkg.Package.Path)
						}
					}
				}
			}
			if len(matching) > 1 {
				return nil, v.currScope.Err(
					op.Start,
					fmt.Errorf("ambiguous reference to variable %q: defined in import %q and %q", ref.Name, matching[0], matching[1]),
				)
			}
		}
	}

	elem, err := v.l.resolveRef(op, ref, v.currScope)
	if err == errIsPropRef {
		prefix := ref.Prefix
		name := ref.Name

		propRef := &program.PropertyRef{
			Reference: &program.Reference{
				Prefix: "",
				Name:   prefix,
			},
			Property: name,
		}

		op.T = &program.Operand_PropRef{PropRef: propRef}
		return program.VisitOperand(op, v)
	}
	if err != nil {
		return nil, err
	}
	if elem == nil {
		return nil, nil
	}

	pkgPath, refPath := scope.ResolvePaths(v.currScope, elem)
	ref.PkgPath = pkgPath
	ref.RefPath = refPath
	return nil, nil
}

func (v *linkVisitor) VisitExpression(op *program.Operand, expr *program.Expression) (any, error) {
	if expr.Operation == program.Operation_ACCESS {
		return program.VisitOperand(expr.Operands[0], v)
	}

	for _, operand := range expr.Operands {
		_, err := program.VisitOperand(operand, v)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (v *linkVisitor) VisitFnCall(op *program.Operand, fc *program.FnCall) (any, error) {
	if fc.Reference.Prefix == "" && !program.IsBuiltin(fc.Reference.Name) && !isTopologyBuiltin(fc.Reference.Name) && !isLexicalSymbol(v.currScope, fc.Reference.Name) {
		fileScope := v.currScope.PkgFileScope()
		if fileScope != nil && len(fileScope.DotImports) > 1 {
			var matching []string
			argCount := int32(len(fc.Arguments))
			for _, dotImp := range fileScope.DotImports {
				pkg := dotImp.PkgScope()
				if pkg.Package.Path != "builtin" {
					if elem := pkg.NamedElem(fc.Reference.Name); elem != nil {
						if fnTypeScope, ok := elem.(*scope.FnTypeScope); ok {
							if _, ok := fnTypeScope.Groups[argCount]; ok {
								matching = append(matching, pkg.Package.Path)
							}
						}
					}
				}
			}
			if len(matching) > 1 {
				return nil, v.currScope.Err(
					op.Start,
					fmt.Errorf("ambiguous call to %q with %d arguments: defined in import %q and %q", fc.Reference.Name, len(fc.Arguments), matching[0], matching[1]),
				)
			}
		}
	}

	elem, err := v.l.resolveRef(op, fc.Reference, v.currScope)
	if err != nil {
		return nil, err
	}

	for _, arg := range fc.Arguments {
		_, err := program.VisitOperand(arg, v)
		if err != nil {
			return nil, err
		}
	}

	if elem == nil {
		return nil, nil
	}

	pkgPath, refPath := scope.ResolvePaths(v.currScope, elem)
	fc.Reference.PkgPath = pkgPath
	fc.Reference.RefPath = refPath
	return nil, nil
}

func (v *linkVisitor) VisitPropRef(op *program.Operand, pr *program.PropertyRef) (any, error) {
	if pr.Reference.Prefix == "" && !isLexicalSymbol(v.currScope, pr.Reference.Name) {
		fileScope := v.currScope.PkgFileScope()
		if fileScope != nil && len(fileScope.DotImports) > 1 {
			var matching []string
			for _, dotImp := range fileScope.DotImports {
				pkg := dotImp.PkgScope()
				if pkg.Package.Path != "builtin" {
					if elem := pkg.NamedElem(pr.Reference.Name); elem != nil {
						if _, isVar := elem.(*scope.Var); isVar {
							matching = append(matching, pkg.Package.Path)
						}
					}
				}
			}
			if len(matching) > 1 {
				return nil, v.currScope.Err(
					op.Start,
					fmt.Errorf("ambiguous reference to variable %q: defined in import %q and %q", pr.Reference.Name, matching[0], matching[1]),
				)
			}
		}
	}

	elem, err := v.l.resolveRef(op, pr.Reference, v.currScope)
	if err != nil {
		return nil, err
	}
	if elem == nil {
		return nil, v.currScope.Err(op.Start, fmt.Errorf("variable %q is not defined", pr.Reference.Name))
	}

	pkgPath, refPath := scope.ResolvePaths(v.currScope, elem)
	pr.Reference.PkgPath = pkgPath
	pr.Reference.RefPath = refPath

	if pr.Property == "box" {
		targetOp := &program.Operand{
			Start: op.Start,
			End:   op.End,
			T: &program.Operand_Reference{
				Reference: &program.Reference{
					Prefix:  pr.Reference.Prefix,
					Name:    pr.Reference.Name,
					PkgPath: pkgPath,
					RefPath: refPath,
				},
			},
		}
		op.T = &program.Operand_Expression{
			Expression: &program.Expression{
				Operation: program.Operation_BOUNDINGBOX,
				Operands:  []*program.Operand{targetOp},
			},
		}
		return nil, nil
	}

	if vr, ok := elem.(*scope.Var); ok && vr.Var != nil {
		hasProp := false
		for _, p := range vr.Var.Properties {
			if p.Name == pr.Property {
				hasProp = true
				break
			}
		}
		if !hasProp {
			return nil, v.currScope.Err(op.Start, fmt.Errorf("variable %q has no property %q", pr.Reference.Name, pr.Property))
		}
	} else {
		return nil, v.currScope.Err(op.Start, fmt.Errorf("variable %q has no property %q", pr.Reference.Name, pr.Property))
	}

	return nil, v.currScope.Err(op.Start, fmt.Errorf("reading property %q in expressions is not supported; properties are declaration metadata", pr.Property))
}

func (l *linker) resolveRef(op *program.Operand, ref *program.Reference, currScope scope.Scope) (scope.NamedElem, error) {
	if ref.Prefix != "" {
		if ref.Prefix == program.PVarName || ref.Prefix == program.TimeVarName {
			return nil, errIsPropRef
		}

		prefElem := currScope.NamedElem(ref.Prefix)
		if prefElem == nil {
			// Report unresolved reference if not a system package. The prefix is
			// the last segment of the path, so the path has to be looked up
			// rather than assumed equal to it: `beam` means `lattice/beam`.
			pkgPath, isSystem := gsllibs.PathForPrefix(ref.Prefix)
			if !isSystem {
				return nil, currScope.Err(op.Start, fmt.Errorf("undefined identifier: %s", ref.Prefix))
			}

			// Check program dependencies first
			var depPkgScope *scope.PkgScope
			for _, dep := range l.prog.Dependencies {
				if dep.Path == pkgPath {
					depPkgScope = l.progScope.Dependencies[pkgPath]
					break
				}
			}

			// Query srcMgr for the system package if not already loaded
			if depPkgScope == nil {
				if l.srcMgr == nil {
					return nil, currScope.Err(op.Start, fmt.Errorf("undefined identifier: %s", ref.Prefix))
				}
				var err error
				depPkgScope, err = l.loadAndScopeDependency(pkgPath, currScope, op)
				if err != nil {
					return nil, currScope.Err(op.Start, fmt.Errorf("undefined identifier: %s (failed to load: %w)", ref.Prefix, err))
				}
			}

			// Dynamically add the import and update scopes in-place. An empty
			// import prefix means "derive it from the last path segment", which
			// is what the parser records for an unaliased import of this path.
			if depPkgScope != nil {
				fileScope := currScope.PkgFileScope()
				autoImp := program.NewImport("", pkgPath, &program.Location{Line: 1, Column: 1}, &program.Location{Line: 1, Column: 1})
				fileScope.File.Imports = append(fileScope.File.Imports, autoImp)
				fileScope.AddImportRef(ref.Prefix, depPkgScope)

				prefElem = currScope.NamedElem(ref.Prefix)
			}
		}

		if prefElem == nil {
			return nil, currScope.Err(op.Start, fmt.Errorf("undefined identifier: %s", ref.Prefix))
		}

		pkgRef, isPkgRef := prefElem.(*scope.PkgRef)
		if !isPkgRef {
			return nil, errIsPropRef
		}

		elem := pkgRef.NamedElem(ref.Name)
		if elem == nil {
			pkgPath := ""
			if pkgRef.PkgScope() != nil && pkgRef.PkgScope().Package != nil {
				pkgPath = pkgRef.PkgScope().Package.Path
			}
			if pkgPath != "" {
				var freshPkg *program.Package
				if std := program.GetStdPackage(pkgPath); std != nil {
					freshPkg = std
				} else if l.srcMgr != nil {
					if sources, _, err := l.srcMgr.Sources(pkgPath); err == nil && len(sources) > 0 {
						if freshRes, err := l.parsePkg(sources); err == nil && freshRes != nil {
							freshPkg = freshRes.Package
						}
					}
				}
				if freshPkg != nil {
					pkgRef.PkgScope().ReloadDeclarations(freshPkg)
					elem = pkgRef.NamedElem(ref.Name)
				}
			}
		}
		if elem == nil {
			return nil, currScope.Err(op.Start, fmt.Errorf("name %q not found in package %q", ref.Name, pkgRef.Name()))
		}
		return elem, nil

	} else {
		if program.IsBuiltin(ref.Name) || ref.Name == program.PVarName || ref.Name == program.TimeVarName || isTopologyBuiltin(ref.Name) || ref.Name == "auto" {
			return nil, nil
		}
	}

	elem := currScope.NamedElem(ref.Name)
	if elem == nil {
		if pkgPath, isSystem := gsllibs.PathForPrefix(ref.Name); isSystem {
			var depPkgScope *scope.PkgScope
			for _, dep := range l.prog.Dependencies {
				if dep.Path == pkgPath {
					depPkgScope = l.progScope.Dependencies[pkgPath]
					break
				}
			}

			if depPkgScope == nil && l.srcMgr != nil {
				var err error
				depPkgScope, err = l.loadAndScopeDependency(pkgPath, currScope, op)
				if err != nil {
					return nil, currScope.Err(op.Start, fmt.Errorf("undefined identifier: %s (failed to load: %w)", ref.Name, err))
				}
			}

			if depPkgScope != nil {
				fileScope := currScope.PkgFileScope()
				autoImp := program.NewImport("", pkgPath, &program.Location{Line: 1, Column: 1}, &program.Location{Line: 1, Column: 1})
				fileScope.File.Imports = append(fileScope.File.Imports, autoImp)
				fileScope.AddImportRef(ref.Name, depPkgScope)

				elem = currScope.NamedElem(ref.Name)
			}
		}
	}

	if elem == nil {
		return nil, currScope.Err(op.Start, fmt.Errorf("undefined identifier: %s", ref.Name))
	}
	return elem, nil
}

func (l *linker) loadAndScopeDependency(
	pkgPath string,
	currScope scope.Scope,
	op *program.Operand,
) (*scope.PkgScope, error) {
	if pkgPath == "builtin" || pkgPath == l.prog.Primary.Path {
		return nil, nil
	}
	if s, ok := l.progScope.Dependencies[pkgPath]; ok {
		return s, nil
	}

	sources, _, err := l.srcMgr.Sources(pkgPath)
	if err != nil {
		return nil, err
	}

	depRes, err := l.parsePkg(sources)
	if err != nil {
		return nil, err
	}
	if depRes == nil || depRes.Package == nil {
		return nil, fmt.Errorf("package %q not found or empty", pkgPath)
	}
	depPkg := depRes.Package

	program.UpdateImports(depPkg, pkgPath)
	if depPkg.Path == "" {
		depPkg.Path = pkgPath
	}

	// Recursively load and scope dependencies bottom-up
	for _, file := range depPkg.Files {
		for _, imp := range file.Imports {
			if _, err := l.loadAndScopeDependency(imp.Path, currScope, op); err != nil {
				return nil, err
			}
		}
	}

	// Add package to program
	l.prog.Dependencies = append(l.prog.Dependencies, depPkg)

	// Register package in ProgScope
	depScope, err := l.progScope.AddDependency(depPkg)
	if err != nil {
		return nil, err
	}

	return depScope, nil
}

func isTopologyBuiltin(name string) bool {
	switch name {
	case "facePoint", "edgePoint", "vertexPoint", "circleCenter",
		"facePointNormal", "edgePointNormal", "vertexPointNormal", "circleCenterNormal",
		"meshVolume", "meshArea", "meshMaxThickness", "meshThickestPoint", "meshThickestPoints":
		return true
	default:
		return false
	}
}
