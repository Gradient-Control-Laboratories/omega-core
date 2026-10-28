package property

import (
	"fmt"
	"sort"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/program"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

// Expand returns a program in which every property operator reachable from a
// top-level declaration or return has been replaced by its computed value. If
// the program uses no property operators it is returned unchanged; otherwise a
// deep copy is rewritten so the caller's (possibly cached) program is untouched.
//
// Today the only property operators are the derivatives .dtime and .dp. A DERIV
// that survives inside a function body — which needs call-site context — is left
// in place; the build pass reports a clear, located error if such a function is
// actually used.
func Expand(prog *program.Program) (*program.Program, error) {
	return expand(prog, false)
}

// ExpandShared is Expand for a compiler pipeline: each expansion comes back
// with the nodes it reaches more than once written down as named locals (see
// share.go), because everything downstream of Expand copies a DAG out as a
// tree. The value computed is the same; Expand keeps the inline spelling that
// reads as one expression.
func ExpandShared(prog *program.Program) (*program.Program, error) {
	return expand(prog, true)
}

func expand(prog *program.Program, share bool) (*program.Program, error) {
	if prog == nil {
		return nil, nil
	}

	hasOperator := false
	if prog.Primary != nil && packageHasOperator(prog.Primary) {
		hasOperator = true
	} else {
		for _, dep := range prog.Dependencies {
			if dep != nil && packageHasOperator(dep) {
				hasOperator = true
				break
			}
		}
	}

	if !hasOperator {
		return prog, nil
	}

	cloned := proto.Clone(prog).(*program.Program)

	// An import is read as it was written, before any expansion here, except
	// that its declared boxes are applied: a library part's .box is its
	// declaration.
	imported := make([]*program.Package, len(prog.Dependencies))
	for i, dep := range prog.Dependencies {
		if dep != nil && packageDeclaresBox(dep) {
			dep = proto.Clone(dep).(*program.Package)
			if err := applyDeclaredBoxes(dep); err != nil {
				return nil, err
			}
		}
		imported[i] = dep
	}

	getImport := func(path string) (*program.Package, error) {
		if prog != nil {
			for _, dep := range imported {
				if dep != nil {
					if dep.Path == path || (path == "builtin" && (dep.Path == "builtin" || dep.Path == "")) {
						return dep, nil
					}
				}
			}
		}
		return prog.GetImport(path)
	}

	// Declared boxes are applied in every package before any is expanded, so a
	// .box read anywhere sees them, a library's as well as the document's own.
	// Stand-ins for values go with them, all but the document's own top-level
	// ones, which may be its parts' and are MarkParts'.
	for _, pkg := range append([]*program.Package{cloned.Primary}, cloned.Dependencies...) {
		if pkg == nil {
			continue
		}
		if err := applyDeclaredBoxes(pkg); err != nil {
			return nil, err
		}
		// A value's fine detail first, so a value with both is
		// simpleValue(detailValue(coarse, fine), stand).
		if err := eachAssignments(pkg, true, lowerDetailValues); err != nil {
			return nil, err
		}
		if err := applySimpleValues(pkg, pkg != cloned.Primary); err != nil {
			return nil, err
		}
	}

	exp := &expansion{
		getImport: getImport,
		share:     share,
		deps:      cloned.Dependencies,
		scopes:    map[*program.Package]*pkgScope{},
	}
	if err := exp.rewritePackage(cloned.Primary); err != nil {
		return nil, err
	}
	for _, dep := range cloned.Dependencies {
		if err := exp.rewritePackage(dep); err != nil {
			return nil, err
		}
	}
	// Companions are woven in once every package is rewritten: a call in one
	// package can ask for a companion in another, whichever was rewritten
	// first.
	exp.weave(cloned.Primary)
	for _, dep := range cloned.Dependencies {
		exp.weave(dep)
	}

	return cloned, nil
}

// expansion is one run of Expand over a program: the scope of every package it
// touches, built once and shared. A qualified call from one package to a
// differentiating function in another finds the callee's companion registry
// here, and the companion is generated in the callee's own package, where its
// body resolves (see rewrite's FnCall case).
type expansion struct {
	getImport program.GetImportFunc
	share     bool
	deps      []*program.Package
	scopes    map[*program.Package]*pkgScope
}

// pkgScope is a package's namespace as the pass reads it: its own declarations
// and functions with every dependency's, the builtin's and its dot-imports'
// merged in by bare name, its qualified imports, its companion registry and
// the names its shared locals are drawn from.
type pkgScope struct {
	decls      map[string]*program.Operand
	multiDecls map[string]multiBinding
	funcsList  map[string][]*program.Function
	imports    map[string]*namespace
	comp       *companions
	names      *shareNames
}

// scopeFor builds a package's scope the first time it is asked for. The scope
// is registered before its companion registry is built, because computing
// which functions are tainted asks the registries of the packages this one
// imports, which may be built on the way.
func (exp *expansion) scopeFor(pkg *program.Package) *pkgScope {
	if s, ok := exp.scopes[pkg]; ok {
		return s
	}
	decls := collectPackageAssignments(pkg)
	multiDecls := collectPackageMultiAssignments(pkg)
	funcsList := collectPackageFunctionsList(pkg)
	funcsMap := collectPackageFunctions(pkg)

	for _, dep := range exp.deps {
		if dep != nil {
			mergePackageNamespace(dep, decls, funcsList, funcsMap)
		}
	}

	if ip, err := exp.getImport("builtin"); err == nil && ip != nil {
		mergePackageNamespace(ip, decls, funcsList, funcsMap)
	}

	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, imp := range f.Imports {
			if imp.Prefix == "." {
				if ip, err := exp.getImport(imp.Path); err == nil && ip != nil {
					mergePackageNamespace(ip, decls, funcsList, funcsMap)
				}
			}
		}
	}
	s := &pkgScope{
		decls:      decls,
		multiDecls: multiDecls,
		funcsList:  funcsList,
		imports:    collectImportNamespaces(pkg, exp.getImport),
		names: &shareNames{taken: func(name string) bool {
			_, isDecl := decls[name]
			_, isFunc := funcsList[name]
			return isDecl || isFunc
		}},
	}
	exp.scopes[pkg] = s
	s.comp = newCompanions(exp, pkg, decls, funcsList, s.imports)
	if exp.share {
		s.comp.names = s.names
	}
	return s
}

// packageByPath is the package of the program being expanded that an import
// path names, or nil: the primary has no path, and a qualified call never
// names it.
func (exp *expansion) packageByPath(path string) *program.Package {
	for _, dep := range exp.deps {
		if dep != nil && dep.Path == path {
			return dep
		}
	}
	return nil
}

// qualified is the companion registry a prefix names from within pkg: the
// registry of the package the prefix imports, so a call to a differentiating
// library function is routed to a companion generated beside it, in its own
// package. Nil when the prefix does not resolve to a package of the program.
func (exp *expansion) qualified(pkg *program.Package, prefix string) *companions {
	ns, ok := exp.scopeFor(pkg).imports[prefix]
	if !ok || ns == nil {
		return nil
	}
	dep := exp.packageByPath(ns.path)
	if dep == nil {
		return nil
	}
	return exp.scopeFor(dep).comp
}

// rewritePackage replaces every property operator in pkg's declarations,
// properties and returns with its expansion, routing calls to differentiating
// functions through their companions.
func (exp *expansion) rewritePackage(pkg *program.Package) error {
	if pkg == nil {
		return nil
	}
	s := exp.scopeFor(pkg)
	// A package with no operator of its own still needs rewriting when it
	// calls a function whose body has one: the call becomes a companion call
	// carrying the arguments' tangents, which only the caller has. Left alone,
	// the callee's v.dp on a parameter has nothing to differentiate against.
	if !packageHasOperator(pkg) && !packageCallsTainted(pkg, s.comp) {
		return nil
	}
	decls, multiDecls, funcsList, imports := s.decls, s.multiDecls, s.funcsList, s.imports
	getImport, share, names := exp.getImport, exp.share, s.names

	sec := &second{decls: decls, multiDecls: multiDecls, funcs: funcsList}
	fr := newFrames(pkg, getImport, sec)
	r := &rewriter{
		deriv:  &derivative{decls: decls, multiDecls: multiDecls, funcs: funcsList, imports: imports, own: ownFunctionNames(pkg), frames: fr},
		sec:    sec,
		box:    &boxProp{decls: decls, multiDecls: multiDecls, funcs: funcsList, imports: imports, own: ownFunctionNames(pkg), sliders: sliderNames(pkg)},
		frames: fr,
		comp:   s.comp,
	}

	// Rewrite all assignments and returns in all files
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		// At package scope an expansion's shared nodes become the locals of a
		// function synthesized beside it, and the expansion a call to it: a
		// package function sees p and every global, which is all the inline
		// expansion saw, and under a remap the call is evaluated at the rebound
		// p exactly as the inline tree would have been.
		file := f
		r.share = nil
		if share {
			r.share = func(expanded *program.Operand) *program.Operand {
				root, locals := sharedLocals(expanded, names)
				if len(locals) == 0 {
					return expanded
				}
				fnName := names.fresh("shared")
				file.FnTypes = append(file.FnTypes, &program.FnType{
					Name: fnName,
					Groups: []*program.FnGroup{{
						ParamCount: 0,
						Functions:  []*program.Function{{Assignments: locals, Returns: []*program.Operand{root}}},
					}},
				})
				return program.NewFnCall("", fnName, nil, nil, nil)
			}
		}
		for _, assign := range f.Assignments {
			if assign == nil {
				continue
			}
			rewritten, err := r.rewrite(assign.Expression)
			if err != nil {
				return err
			}
			assign.Expression = rewritten
			// A property's expression is rewritten too, so a read where the
			// shape is named — bar.material = turned(..., bar.u, ...), or a
			// colour from s.dp — expands as it would in a declaration. Only
			// one holding an operator: every other property stays exactly as
			// written.
			for _, v := range assign.Variables {
				if v == nil {
					continue
				}
				for _, prop := range v.Properties {
					if prop == nil || !operandHasOperator(prop.Expression) {
						continue
					}
					rewritten, err := r.rewrite(prop.Expression)
					if err != nil {
						return err
					}
					prop.Expression = rewritten
				}
			}
		}

		for i, ret := range f.Returns {
			rewritten, err := r.rewrite(ret)
			if err != nil {
				return err
			}
			f.Returns[i] = rewritten
		}
	}

	return nil
}

// weave declares pkg's generated companions in its files, each immediately
// after the function it was made from, in a fixed order.
func (exp *expansion) weave(pkg *program.Package) {
	if pkg == nil {
		return
	}
	s, ok := exp.scopes[pkg]
	if !ok || len(s.comp.generated) == 0 {
		return
	}

	// Map function name to file index
	fnFileIdx := make(map[string]int)
	for idx, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, ft := range f.FnTypes {
			if ft != nil {
				fnFileIdx[ft.Name] = idx
			}
		}
	}

	// Group the companions by the file they belong to, in name order.
	keys := make([]string, 0, len(s.comp.generated))
	for key := range s.comp.generated {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	companionsByFile := make(map[int][]*companionFn)
	for _, key := range keys {
		comp := s.comp.generated[key]
		fileIdx, ok := fnFileIdx[comp.src]
		if !ok {
			// fallback to the last file if we can't find it
			fileIdx = len(pkg.Files) - 1
		}
		companionsByFile[fileIdx] = append(companionsByFile[fileIdx], comp)
	}

	for idx, f := range pkg.Files {
		if f == nil {
			continue
		}
		fileComps := companionsByFile[idx]
		if len(fileComps) == 0 {
			continue
		}

		placed := map[*companionFn]bool{}
		newFnTypes := make([]*program.FnType, 0, len(f.FnTypes)+len(fileComps))
		for _, ft := range f.FnTypes {
			newFnTypes = append(newFnTypes, ft)
			if ft == nil {
				continue
			}
			for _, comp := range fileComps {
				if comp.src == ft.Name {
					newFnTypes = append(newFnTypes, companionFnType(comp))
					placed[comp] = true
				}
			}
		}

		// Any leftover companions for this file
		for _, comp := range fileComps {
			if !placed[comp] {
				newFnTypes = append(newFnTypes, companionFnType(comp))
			}
		}
		f.FnTypes = newFnTypes
	}
}

func companionFnType(comp *companionFn) *program.FnType {
	return &program.FnType{
		Name: comp.name,
		Groups: []*program.FnGroup{
			{
				ParamCount: int32(len(comp.fn.Parameters)),
				Functions:  []*program.Function{comp.fn},
			},
		},
	}
}

func collectPackageAssignments(pkg *program.Package) map[string]*program.Operand {
	m := make(map[string]*program.Operand)
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, a := range f.Assignments {
			if a == nil || len(a.Variables) != 1 {
				continue
			}
			m[a.Variables[0].Name] = a.Expression
		}
	}
	return m
}

// collectImportNamespaces resolves the package's qualified imports (map, noise,
// …) to their scopes, so a property that differentiates can inline a library
// call rather than treat it as a given. Each namespace absorbs its own package's
// dot-imports, since a library body resolves those names unqualified.
func collectImportNamespaces(pkg *program.Package, getImport program.GetImportFunc) map[string]*namespace {
	out := map[string]*namespace{}
	collectQualifiedNamespaces(pkg, getImport, out, map[string]bool{})
	// The builder auto-imports matrix and vector without a GSL import
	// statement (see build's auto-import list), and dot-imported library
	// bodies reference them qualified (builtin's cylinderZ uses vector.dirZ),
	// so the engine binds them the same way when they resolve.
	for _, auto := range []string{"matrix", "vector"} {
		if _, done := out[auto]; done {
			continue
		}
		if ip, err := getImport(auto); err == nil && ip != nil {
			ns := &namespace{
				path:       auto,
				decls:      collectPackageAssignments(ip),
				multiDecls: collectPackageMultiAssignments(ip),
				funcs:      collectPackageFunctionsList(ip),
			}
			absorbDotImports(ip, getImport, ns, map[string]bool{auto: true})
			out[auto] = ns
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// collectQualifiedNamespaces adds pkg's qualified imports to out, then follows
// pkg's dot-imports transitively: a dot-imported library's functions merge into
// the primary scope, so the qualified prefixes their bodies reference (e.g.
// builtin's cylinderZ using vector.dirZ) must resolve at the top level too.
// The first package to claim a prefix keeps it.
func collectQualifiedNamespaces(pkg *program.Package, getImport program.GetImportFunc,
	out map[string]*namespace, seen map[string]bool) {
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, imp := range f.Imports {
			if imp.Prefix == "." {
				continue // dot-imports merge into the primary scope instead
			}
			prefix := imp.PrefixOrSimpleName()
			if _, done := out[prefix]; done {
				continue
			}
			ip, err := getImport(imp.Path)
			if err != nil || ip == nil {
				continue
			}
			ns := &namespace{
				path:       imp.Path,
				decls:      collectPackageAssignments(ip),
				multiDecls: collectPackageMultiAssignments(ip),
				funcs:      collectPackageFunctionsList(ip),
			}
			absorbDotImports(ip, getImport, ns, map[string]bool{imp.Path: true})
			out[prefix] = ns
			// A qualified import's body may itself call a second package
			// qualified, and that prefix has to resolve here too. Same
			// argument as the dot-import walk below — and the reason
			// builtin's vector.dirZ needed a hardcoded special case above.
			if !seen[imp.Path] {
				seen[imp.Path] = true
				collectQualifiedNamespaces(ip, getImport, out, seen)
			}
		}
	}
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, imp := range f.Imports {
			if imp.Prefix != "." || seen[imp.Path] {
				continue
			}
			seen[imp.Path] = true
			if ip, err := getImport(imp.Path); err == nil && ip != nil {
				collectQualifiedNamespaces(ip, getImport, out, seen)
			}
		}
	}
}

// absorbDotImports merges a library's own dot-imported names into its namespace,
// following them transitively; seen guards against an import cycle.
func absorbDotImports(pkg *program.Package, getImport program.GetImportFunc, ns *namespace, seen map[string]bool) {
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		var builtinImp *program.Import
		for _, imp := range f.Imports {
			if imp.Prefix != "." || seen[imp.Path] {
				continue
			}
			if imp.Path == "builtin" {
				builtinImp = imp
				continue
			}
			seen[imp.Path] = true
			ip, err := getImport(imp.Path)
			if err != nil || ip == nil {
				continue
			}
			mergePackageNamespace(ip, ns.decls, ns.funcs, nil)
			absorbDotImports(ip, getImport, ns, seen)
		}
		if builtinImp != nil {
			seen[builtinImp.Path] = true
			if ip, err := getImport(builtinImp.Path); err == nil && ip != nil {
				mergePackageNamespace(ip, ns.decls, ns.funcs, nil)
				absorbDotImports(ip, getImport, ns, seen)
			}
		}
	}
}

// collectPackageMultiAssignments gathers the destructuring assignments
// (a, b = f()) of every file in the package; collectPackageAssignments takes
// the single-name ones.
func collectPackageMultiAssignments(pkg *program.Package) map[string]multiBinding {
	var m map[string]multiBinding
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for name, mb := range collectMultiAssignments(f.Assignments) {
			if m == nil {
				m = make(map[string]multiBinding)
			}
			m[name] = mb
		}
	}
	return m
}

func collectPackageFunctions(pkg *program.Package) map[string]*program.Function {
	m := make(map[string]*program.Function)
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, ft := range f.FnTypes {
			if ft == nil {
				continue
			}
			for _, group := range ft.Groups {
				if group == nil {
					continue
				}
				for _, fn := range group.Functions {
					if fn != nil {
						m[ft.Name] = fn
					}
				}
			}
		}
	}
	return m
}

func collectPackageFunctionsList(pkg *program.Package) map[string][]*program.Function {
	m := make(map[string][]*program.Function)
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, ft := range f.FnTypes {
			if ft == nil {
				continue
			}
			for _, group := range ft.Groups {
				if group == nil {
					continue
				}
				for _, fn := range group.Functions {
					if fn != nil {
						m[ft.Name] = append(m[ft.Name], fn)
					}
				}
			}
		}
	}
	return m
}

// rewriter walks operands replacing property-operator subtrees with their
// expansion. It is the routing layer: each operator op is dispatched to the
// property that owns it. deriv is the current scope's derivative property
// (top-level, or function-scoped while generating a companion body); comp is
// the shared companion registry.
type rewriter struct {
	deriv  *derivative
	sec    *second
	box    *boxProp
	frames *frames
	comp   *companions
	// share writes an expansion's shared nodes down where this rewriter's
	// scope can hold them (see share.go), returning what stands in for the
	// expansion. Nil leaves an expansion as the engine built it.
	share func(expanded *program.Operand) *program.Operand
	// pending holds shared locals a companion body's share produced for the
	// statement being rewritten; rewriteAssignments places them before it.
	pending []*program.Assignment
}

// expandOperator returns (expansion, true, err) if expr is a property operator.
func (r *rewriter) expandOperator(expr *program.Expression) (*program.Operand, bool, error) {
	switch expr.Operation {
	case program.Operation_DERIV:
		op, err := r.deriv.expandDeriv(expr.Operands)
		return op, true, err
	case program.Operation_SECONDCLOSEST:
		if r.sec == nil {
			// The companion-body rewriter carries no second property: a
			// .second of a function parameter would need call-site context it
			// does not have. Apply .second at the call site instead.
			return nil, true, fmt.Errorf("the .second operator is not supported inside a differentiated function body; apply .second at the call site")
		}
		op, err := r.sec.expandSecond(expr.Operands)
		return op, true, err
	case program.Operation_BOUNDINGBOX:
		if r.box == nil {
			// The companion-body rewriter carries no box property, for the
			// same reason it carries no second.
			return nil, true, fmt.Errorf("the .box operator is not supported inside a differentiated function body; apply .box at the call site")
		}
		op, err := r.box.expandBox(expr.Operands)
		return op, true, err
	case program.Operation_SURFACEU, program.Operation_SURFACEV:
		if r.frames == nil {
			// The companion-body rewriter carries no frames, for the same
			// reason it carries no box: a material reads the frame of the shape
			// it is on, which only the call site names.
			return nil, true, fmt.Errorf("the .u and .v operators are not supported inside a function body; read the frame at the call site and pass it in")
		}
		op, err := r.frames.expand(expr.Operation, expr.Operands)
		if err != nil {
			return nil, true, err
		}
		// A frame is an ordinary expression of p, so one written with .dp
		// still holds an operator for this rewriter to expand.
		op, err = r.rewrite(op)
		return op, true, err
	}
	return nil, false, nil
}

func (r *rewriter) rewrite(op *program.Operand) (*program.Operand, error) {
	if op == nil {
		return nil, nil
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		expr := t.Expression
		if expanded, ok, err := r.expandOperator(expr); err != nil {
			return nil, err
		} else if ok {
			if r.share != nil {
				return r.share(expanded), nil
			}
			return expanded, nil
		}
		operands := make([]*program.Operand, len(expr.Operands))
		for i, sub := range expr.Operands {
			rr, err := r.rewrite(sub)
			if err != nil {
				return nil, err
			}
			operands[i] = rr
		}
		return &program.Operand{Start: op.Start, End: op.End, T: &program.Operand_Expression{
			Expression: &program.Expression{Operation: expr.Operation, Operands: operands},
		}}, nil

	case *program.Operand_FnCall:
		fc := t.FnCall
		args := make([]*program.Operand, len(fc.Arguments))
		for i, a := range fc.Arguments {
			rr, err := r.rewrite(a)
			if err != nil {
				return nil, err
			}
			args[i] = rr
		}
		// A call to a function whose body contains a derivative operator is
		// replaced by a call to its companion, passing the argument tangents.
		// A qualified call's companion lives in the package the prefix names,
		// and is called by the same prefix.
		if fc.Reference != nil && r.comp != nil {
			ref := fc.Reference
			if ref.Prefix == "" {
				if r.comp.isTainted(ref.Name, len(args)) {
					return r.comp.call("", ref.Name, args, r.deriv)
				}
			} else if dep := r.comp.qualified(ref.Prefix); dep != nil && dep.isTainted(ref.Name, len(args)) {
				return dep.call(ref.Prefix, ref.Name, args, r.deriv)
			}
		}
		return &program.Operand{Start: op.Start, End: op.End, T: &program.Operand_FnCall{
			FnCall: &program.FnCall{
				Reference: fc.Reference,
				Arguments: args,
			},
		}}, nil

	case *program.Operand_Value, *program.Operand_Reference, *program.Operand_PropRef:
		return op, nil
	default:
		return nil, fmt.Errorf("internal error: unhandled operand type %T in rewrite; please add a new case to this switch", t)
	}
}

// --- declared boxes ---------------------------------------------------------

// boxPropName is the declared bounding box, name.box = [lo, hi].
const boxPropName = "box"

// cropToBoxName is builtin's crop, which a declared box applies.
const cropToBoxName = "cropToBox"

// declaredBoxName is builtin's marker for a standard library part's box, which
// a declared box applies there instead of the crop.
const declaredBoxName = "declaredBox"

// libraryPaths are the standard library's packages (gsllibs).
var libraryPaths = func() map[string]bool {
	paths := map[string]bool{}
	for _, lib := range gsllibs.All() {
		paths[lib.ImportPath] = true
	}
	return paths
}()

// applyDeclaredBoxes rewrites every declaration that sets name.box = [lo, hi],
// at the top level or inside a function, and drops the property. In a
// document, or any package outside the standard library, it becomes
// cropToBox(expression, lo, hi): the crop is what the box promises, and plain
// GSL carries it to every backend. In the standard library, whose boxes are
// proven against its parts, it becomes declaredBox(expression, lo, hi), which
// is the expression wherever it is built and only the box Domain reads.
func applyDeclaredBoxes(pkg *program.Package) error {
	apply := cropToBoxName
	if libraryPaths[pkg.Path] {
		apply = declaredBoxName
	}
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		if err := cropAssignments(f.Assignments, apply, pkg); err != nil {
			return err
		}
		for _, ft := range f.FnTypes {
			if err := cropFnType(ft, apply, pkg); err != nil {
				return err
			}
		}
	}
	return nil
}

func cropFnType(ft *program.FnType, apply string, pkg *program.Package) error {
	if ft == nil {
		return nil
	}
	for _, group := range ft.Groups {
		if group == nil {
			continue
		}
		for _, fn := range group.Functions {
			if fn == nil {
				continue
			}
			if err := cropAssignments(fn.Assignments, apply, pkg); err != nil {
				return err
			}
			for _, nested := range fn.FnTypes {
				if err := cropFnType(nested, apply, pkg); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func cropAssignments(assigns []*program.Assignment, apply string, pkg *program.Package) error {
	for _, a := range assigns {
		if a == nil {
			continue
		}
		for _, v := range a.Variables {
			if v == nil {
				continue
			}
			kept := v.Properties[:0]
			var declared *program.Property
			for _, prop := range v.Properties {
				if prop != nil && prop.Name == boxPropName {
					declared = prop
					continue
				}
				kept = append(kept, prop)
			}
			if declared == nil {
				continue
			}
			if len(a.Variables) != 1 {
				return fmt.Errorf("%s.box: a box can only be declared on a single-name assignment", v.Name)
			}
			v.Properties = kept

			if isAuto(declared.Expression) {
				continue
			}

			lo, hi := boxCorners(declared.Expression)
			if containsAuto(lo) || containsAuto(hi) {
				b := &boxProp{
					decls:      collectPackageAssignments(pkg),
					multiDecls: collectPackageMultiAssignments(pkg),
					funcs:      collectPackageFunctionsList(pkg),
					own:        ownFunctionNames(pkg),
					sliders:    sliderNames(pkg),
				}
				calcBox, _ := b.expandBox([]*program.Operand{a.Expression})
				calcLo, calcHi := boxCorners(calcBox)
				lo = resolveAutoCorner(lo, calcLo)
				hi = resolveAutoCorner(hi, calcHi)
			}
			a.Expression = call(apply, a.Expression, lo, hi)
		}
	}
	return nil
}

func isAuto(op *program.Operand) bool {
	if op == nil {
		return false
	}
	if ref, ok := op.T.(*program.Operand_Reference); ok && ref.Reference != nil {
		return ref.Reference.Prefix == "" && ref.Reference.Name == "auto"
	}
	return false
}

func containsAuto(op *program.Operand) bool {
	if op == nil {
		return false
	}
	if isAuto(op) {
		return true
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		if t.Expression != nil {
			for _, child := range t.Expression.Operands {
				if containsAuto(child) {
					return true
				}
			}
		}
	case *program.Operand_FnCall:
		if t.FnCall != nil {
			for _, arg := range t.FnCall.Arguments {
				if containsAuto(arg) {
					return true
				}
			}
		}
	}
	return false
}

func replaceAuto(op *program.Operand, repl *program.Operand) *program.Operand {
	if op == nil {
		return nil
	}
	if isAuto(op) {
		if repl != nil {
			return program.CloneOperand(repl)
		}
		return op
	}
	res := program.CloneOperand(op)
	switch t := res.T.(type) {
	case *program.Operand_Expression:
		if t.Expression != nil {
			for i, child := range t.Expression.Operands {
				t.Expression.Operands[i] = replaceAuto(child, repl)
			}
		}
	case *program.Operand_FnCall:
		if t.FnCall != nil {
			for i, arg := range t.FnCall.Arguments {
				t.FnCall.Arguments[i] = replaceAuto(arg, repl)
			}
		}
	}
	return res
}

func resolveAutoCorner(corner *program.Operand, calcCorner *program.Operand) *program.Operand {
	if corner == nil {
		return nil
	}
	if isAuto(corner) {
		return program.CloneOperand(calcCorner)
	}
	if e, ok := corner.T.(*program.Operand_Expression); ok && e.Expression != nil && e.Expression.Operation == program.Operation_COLLECT {
		var calcItems []*program.Operand
		if ce, ok := calcCorner.T.(*program.Operand_Expression); ok && ce.Expression != nil && ce.Expression.Operation == program.Operation_COLLECT {
			calcItems = ce.Expression.Operands
		}
		newItems := make([]*program.Operand, len(e.Expression.Operands))
		for i, op := range e.Expression.Operands {
			var repl *program.Operand
			if i < len(calcItems) {
				repl = calcItems[i]
			}
			newItems[i] = replaceAuto(op, repl)
		}
		return program.NewExpression(program.Operation_COLLECT, newItems, corner.Start, corner.End)
	}
	return replaceAuto(corner, calcCorner)
}

// DropDeclaredBoxes replaces every declaredBox(shape, lo, hi) with its shape. A
// library part's box is only there to be read, and MarkParts is the last to
// read one; its corners, catalog lookups as often as not, would otherwise be
// inlined and folded by every later pass for a shader that never uses them.
// A package that declares a declaredBox of its own is left as it is.
//
// It changes prog in place. Only Expand writes a declaredBox, into the copy it
// returns, so the program holding one is never a caller's own.
func DropDeclaredBoxes(prog *program.Program) {
	if prog == nil {
		return
	}
	eachPackage(prog, func(pkg *program.Package) {
		if ownsDeclaredBox(pkg) {
			return
		}
		rewritePackageOperands(pkg, func(op *program.Operand) {
			for shape := declaredBoxShape(op); shape != nil; shape = declaredBoxShape(op) {
				op.T = shape.T
				op.ConstValues = shape.ConstValues
				if op.Type == nil {
					op.Type = shape.Type
				}
			}
		})
	})
}

// declaredBoxShape is the shape of a declaredBox(shape, lo, hi) call, or nil.
func declaredBoxShape(op *program.Operand) *program.Operand {
	fc := op.GetFnCall()
	if fc == nil || fc.Reference == nil || fc.Reference.Prefix != "" ||
		fc.Reference.Name != declaredBoxName || len(fc.Arguments) != 3 {
		return nil
	}
	return fc.Arguments[0]
}

// ownsDeclaredBox reports whether a package declares a function of that name,
// which its own calls then mean.
func ownsDeclaredBox(pkg *program.Package) bool {
	// builtin defines the marker itself, and its own parts declare boxes too
	// (boxCentered).
	if pkg.Path == "builtin" {
		return false
	}
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, ft := range f.FnTypes {
			if ft != nil && ft.Name == declaredBoxName {
				return true
			}
		}
	}
	return false
}

// simpleValueName is builtin's marker for a value with a stand-in:
// simpleValue(value, stand) is the value wherever it is built, and the viewer's
// WGSL draws the stand-in below the Detail slider's Textures rung, a colour's
// and a material's alike. simpleTextureName is the same marker for a stand-in
// that has a stand-in of its own, a simpler texture before the flat look: both
// switch below the Textures rung.
const (
	simpleValueName   = "simpleValue"
	simpleTextureName = "simpleTexture"
)

// applySimpleValues rewrites every declaration inside a function that sets
// name.simple, and at the top level too when topLevel is set, to
// simpleValue(expression, stand), and drops the property. Inside a function
// a stand-in is always a value's, a colour's or a material's; a part's shape
// stand-in is the document's alone (MarkParts).
func applySimpleValues(pkg *program.Package, topLevel bool) error {
	return eachAssignments(pkg, topLevel, func(assigns []*program.Assignment) error {
		return lowerSimpleValues(assigns, nil)
	})
}

// lowerSimpleValues rewrites each of assigns that sets name.simple to
// simpleValue(expression, stand), or to simpleTexture(expression, stand) when
// nests reports the stand-in has a stand-in of its own, and drops the
// property. A nil nests finds none.
func lowerSimpleValues(assigns []*program.Assignment, nests func(*program.Operand) bool) error {
	for _, a := range assigns {
		if a == nil {
			continue
		}
		for _, v := range a.Variables {
			if v == nil {
				continue
			}
			kept := v.Properties[:0]
			var stand *program.Operand
			for _, prop := range v.Properties {
				if prop != nil && prop.Name == simplePropName {
					stand = prop.Expression
					continue
				}
				kept = append(kept, prop)
			}
			if stand == nil {
				continue
			}
			if len(a.Variables) != 1 {
				return fmt.Errorf("%s.simple: a stand-in can only be declared on a single-name assignment", v.Name)
			}
			v.Properties = kept
			marker := simpleValueName
			if nests != nil && nests(stand) {
				marker = simpleTextureName
			}
			a.Expression = call(marker, a.Expression, stand)
		}
	}
	return nil
}

// detailPropName is the property that gives a value its fine detail,
// detailSizePropName the size of the feature it draws, in mm, and
// detailValueName builtin's marker for them: detailValue(coarse, fine, size)
// is fine where a pixel is small enough to show a feature that size, and the
// viewer's WGSL evaluates fine only there. Without a size, the feature is
// defaultDetailSize, wood's fibre.
const (
	detailPropName     = "detail"
	detailSizePropName = "detailSize"
	detailValueName    = "detailValue"
	defaultDetailSize  = 0.15
)

// lowerDetailValues rewrites each of assigns that sets name.detail to
// detailValue(expression, fine, size), and drops the properties. The
// expression is the value's coarse look, drawn where its fine detail cannot
// show; size is name.detailSize, or defaultDetailSize. A size without fine
// detail is an error.
func lowerDetailValues(assigns []*program.Assignment) error {
	for _, a := range assigns {
		if a == nil {
			continue
		}
		for _, v := range a.Variables {
			if v == nil {
				continue
			}
			kept := v.Properties[:0]
			var fine, size *program.Operand
			for _, prop := range v.Properties {
				if prop != nil && prop.Name == detailPropName {
					fine = prop.Expression
					continue
				}
				if prop != nil && prop.Name == detailSizePropName {
					size = prop.Expression
					continue
				}
				kept = append(kept, prop)
			}
			if fine == nil {
				if size != nil {
					return fmt.Errorf("%s.detailSize: the size of fine detail needs %s.detail", v.Name, v.Name)
				}
				continue
			}
			if len(a.Variables) != 1 {
				return fmt.Errorf("%s.detail: fine detail can only be declared on a single-name assignment", v.Name)
			}
			if size == nil {
				size = num(defaultDetailSize)
			}
			v.Properties = kept
			a.Expression = call(detailValueName, a.Expression, fine, size)
		}
	}
	return nil
}

// packageDeclaresDetail reports whether any declaration in the package, at
// any depth, sets name.detail.
func packageDeclaresDetail(pkg *program.Package) bool {
	found := false
	_ = eachAssignments(pkg, true, func(assigns []*program.Assignment) error {
		for _, a := range assigns {
			if a == nil {
				continue
			}
			for _, v := range a.Variables {
				if v != nil && hasProp(v, detailPropName) {
					found = true
				}
			}
		}
		return nil
	})
	return found
}

// packageDeclaresSimple reports whether any declaration in the package, at
// any depth, sets name.simple.
func packageDeclaresSimple(pkg *program.Package) bool {
	found := false
	_ = eachAssignments(pkg, true, func(assigns []*program.Assignment) error {
		for _, a := range assigns {
			if a == nil {
				continue
			}
			for _, v := range a.Variables {
				if v != nil && hasProp(v, simplePropName) {
					found = true
				}
			}
		}
		return nil
	})
	return found
}

// eachAssignments calls f with every function's list of assignments in the
// package, nested functions included, and each file's too when topLevel is
// set.
func eachAssignments(pkg *program.Package, topLevel bool, f func([]*program.Assignment) error) error {
	var fnType func(ft *program.FnType) error
	fnType = func(ft *program.FnType) error {
		if ft == nil {
			return nil
		}
		for _, group := range ft.Groups {
			if group == nil {
				continue
			}
			for _, fn := range group.Functions {
				if fn == nil {
					continue
				}
				if err := f(fn.Assignments); err != nil {
					return err
				}
				for _, nested := range fn.FnTypes {
					if err := fnType(nested); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	for _, file := range pkg.Files {
		if file == nil {
			continue
		}
		if topLevel {
			if err := f(file.Assignments); err != nil {
				return err
			}
		}
		for _, ft := range file.FnTypes {
			if err := fnType(ft); err != nil {
				return err
			}
		}
	}
	return nil
}

func eachPackage(prog *program.Program, visit func(*program.Package)) {
	for _, pkg := range append([]*program.Package{prog.Primary}, prog.Dependencies...) {
		if pkg != nil {
			visit(pkg)
		}
	}
}

// rewritePackageOperands calls visit on every operand in the package, outermost
// first. visit may rewrite the operand it is given; its operands are visited
// after.
func rewritePackageOperands(pkg *program.Package, visit func(*program.Operand)) {
	var walk func(op *program.Operand)
	walk = func(op *program.Operand) {
		if op == nil {
			return
		}
		visit(op)
		switch t := op.T.(type) {
		case *program.Operand_Expression:
			for _, o := range t.Expression.Operands {
				walk(o)
			}
		case *program.Operand_FnCall:
			for _, a := range t.FnCall.Arguments {
				walk(a)
			}
		}
	}
	var walkAssigns func(assigns []*program.Assignment)
	walkAssigns = func(assigns []*program.Assignment) {
		for _, a := range assigns {
			if a == nil {
				continue
			}
			walk(a.Expression)
			for _, v := range a.Variables {
				if v == nil {
					continue
				}
				for _, prop := range v.Properties {
					if prop != nil {
						walk(prop.Expression)
					}
				}
			}
		}
	}
	var walkFnType func(ft *program.FnType)
	walkFnType = func(ft *program.FnType) {
		if ft == nil {
			return
		}
		for _, group := range ft.Groups {
			if group == nil {
				continue
			}
			for _, fn := range group.Functions {
				if fn == nil {
					continue
				}
				walkAssigns(fn.Assignments)
				for _, ret := range fn.Returns {
					walk(ret)
				}
				for _, nested := range fn.FnTypes {
					walkFnType(nested)
				}
			}
		}
	}
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		walkAssigns(f.Assignments)
		for _, ret := range f.Returns {
			walk(ret)
		}
		for _, ft := range f.FnTypes {
			walkFnType(ft)
		}
	}
}

// boxCorners splits a declared [lo, hi] into its corners.
func boxCorners(op *program.Operand) (*program.Operand, *program.Operand) {
	if e, ok := op.T.(*program.Operand_Expression); ok &&
		e.Expression.Operation == program.Operation_COLLECT && len(e.Expression.Operands) == 2 {
		return e.Expression.Operands[0], e.Expression.Operands[1]
	}
	return nary(program.Operation_INDEX, op, num(0)), nary(program.Operation_INDEX, op, num(1))
}

// packageDeclaresBox reports whether any declaration in the package, at any
// depth, sets name.box.
func packageDeclaresBox(pkg *program.Package) bool {
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		if assignsDeclareBox(f.Assignments) {
			return true
		}
		for _, ft := range f.FnTypes {
			if fnTypeDeclaresBox(ft) {
				return true
			}
		}
	}
	return false
}

func fnTypeDeclaresBox(ft *program.FnType) bool {
	if ft == nil {
		return false
	}
	for _, group := range ft.Groups {
		if group == nil {
			continue
		}
		for _, fn := range group.Functions {
			if fn == nil {
				continue
			}
			if assignsDeclareBox(fn.Assignments) {
				return true
			}
			for _, nested := range fn.FnTypes {
				if fnTypeDeclaresBox(nested) {
					return true
				}
			}
		}
	}
	return false
}

func assignsDeclareBox(assigns []*program.Assignment) bool {
	for _, a := range assigns {
		if a == nil {
			continue
		}
		for _, v := range a.Variables {
			if v == nil {
				continue
			}
			for _, prop := range v.Properties {
				if prop != nil && prop.Name == boxPropName {
					return true
				}
			}
		}
	}
	return false
}

// --- operator scanning ------------------------------------------------------

func packageHasOperator(pkg *program.Package) bool {
	if packageDeclaresBox(pkg) || packageDeclaresSimple(pkg) || packageDeclaresDetail(pkg) {
		return true
	}
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, a := range f.Assignments {
			if assignmentHasOperator(a) {
				return true
			}
		}
		for _, ft := range f.FnTypes {
			if fnTypeHasOperator(ft) {
				return true
			}
		}
		for _, r := range f.Returns {
			if operandHasOperator(r) {
				return true
			}
		}
	}
	return false
}

func assignmentHasOperator(a *program.Assignment) bool {
	if a == nil {
		return false
	}
	if operandHasOperator(a.Expression) {
		return true
	}
	for _, v := range a.Variables {
		if v == nil {
			continue
		}
		for _, prop := range v.Properties {
			if prop != nil && operandHasOperator(prop.Expression) {
				return true
			}
		}
	}
	return false
}

func fnTypeHasOperator(ft *program.FnType) bool {
	if ft == nil {
		return false
	}
	for _, group := range ft.Groups {
		if group == nil {
			continue
		}
		for _, fn := range group.Functions {
			if fn == nil {
				continue
			}
			for _, a := range fn.Assignments {
				if assignmentHasOperator(a) {
					return true
				}
			}
			for _, ft := range fn.FnTypes {
				if fnTypeHasOperator(ft) {
					return true
				}
			}
			for _, ret := range fn.Returns {
				if operandHasOperator(ret) {
					return true
				}
			}
		}
	}
	return false
}

func operandHasOperator(op *program.Operand) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		if isOperator(t.Expression.Operation) {
			return true
		}
		for _, sub := range t.Expression.Operands {
			if operandHasOperator(sub) {
				return true
			}
		}
	case *program.Operand_FnCall:
		for _, a := range t.FnCall.Arguments {
			if operandHasOperator(a) {
				return true
			}
		}
	case *program.Operand_Value, *program.Operand_Reference, *program.Operand_PropRef:
		return false
	default:
		panic(fmt.Sprintf("internal error: unhandled operand type %T in operandHasOperator; please add a new case to this switch", t))
	}
	return false
}

// isOperator reports whether an operation is a property operator (expanded by
// this package rather than lowered to a backend).
func isOperator(operation program.Operation) bool {
	return operation == program.Operation_DERIV || operation == program.Operation_SECONDCLOSEST ||
		operation == program.Operation_BOUNDINGBOX ||
		operation == program.Operation_SURFACEU || operation == program.Operation_SURFACEV
}

// sliderNames names a package's slider declarations: those given a .min, .max
// or .period, which the viewer turns into live controls.
func sliderNames(pkg *program.Package) map[string]bool {
	out := map[string]bool{}
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, a := range f.Assignments {
			if a == nil {
				continue
			}
			for _, v := range a.Variables {
				if v == nil {
					continue
				}
				for _, prop := range v.Properties {
					if prop != nil && (prop.Name == "min" || prop.Name == "max" || prop.Name == "period") {
						out[v.Name] = true
					}
				}
			}
		}
	}
	return out
}

// ownFunctionNames names the functions a package declares itself, at any
// arity, so a property reading builtin primitives by name can tell a user's
// own box() from builtin's.
func ownFunctionNames(pkg *program.Package) map[string]bool {
	out := map[string]bool{}
	for name := range collectPackageFunctionsList(pkg) {
		out[name] = true
	}
	return out
}
