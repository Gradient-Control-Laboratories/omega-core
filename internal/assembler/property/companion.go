package property

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

// companions generates and caches companion functions for user functions whose
// bodies contain a derivative operator (.dp / .dtime). A companion takes the
// original parameters plus the spatial/temporal tangents of the parameters its
// operators depend on, and resolves those operators against the passed-in
// tangents — so a function-internal derivative is computed with call-site
// context without inlining the body at every call.
//
// One registry belongs to each package of the program being expanded (see
// expansion in expand.go). A companion is generated in the package that
// declares its function, where its body resolves, and a call from another
// package reaches it qualified: `machined.turned(…)` becomes
// `machined.turned__d(…)` with the tangents the caller computes. Functions are
// told apart by name AND arity throughout, since a name's overloads are
// distinct functions with distinct bodies and needs.
type companions struct {
	exp     *expansion                     // the run this registry belongs to; nil only in isolation
	pkg     *program.Package               // the package whose functions these are
	decls   map[string]*program.Operand    // package-level assignment bodies
	funcs   map[string][]*program.Function // package-level functions (with overloads)
	own     map[string]int                 // overloads per name declared by pkg itself, for naming
	imports map[string]*namespace          // qualified imports, so a tangent can read vector.dirY
	tainted map[string]bool                // fnKey of functions that transitively contain an operator

	generated  map[string]*companionFn // fnKey -> generated companion
	needed     map[string]paramNeeds   // fnKey -> which params need which tangents
	needing    map[string]bool         // needs computation in progress (cycle guard)
	generating map[string]bool         // cycle detection

	// names, when set, writes each companion body's shared nodes down as its
	// own locals (see share.go); nil leaves the body as the engine built it.
	names *shareNames
}

// companionFn is one generated companion: the function it was made from, by
// name and arity, the name it is declared under, and its body.
type companionFn struct {
	src   string
	arity int
	name  string
	fn    *program.Function
}

// paramNeeds records, per parameter name, whether the function's operators
// require that parameter's spatial (dp) and/or temporal (dt) tangent.
type paramNeeds map[string]struct{ dp, dt bool }

// callRef is a call as the taint analysis sees it: the function it names,
// qualified or not, and how many arguments it passes, which picks the overload.
type callRef struct {
	prefix, name string
	argc         int
}

func (r callRef) key() string { return fnKey(r.name, r.argc) }

// fnKey tells a name's overloads apart: companions, needs and taint are all
// kept by it.
func fnKey(name string, arity int) string { return fmt.Sprintf("%s/%d", name, arity) }

func newCompanions(exp *expansion, pkg *program.Package, decls map[string]*program.Operand,
	funcs map[string][]*program.Function, imports map[string]*namespace) *companions {
	c := &companions{
		exp:        exp,
		pkg:        pkg,
		decls:      decls,
		funcs:      funcs,
		own:        map[string]int{},
		imports:    imports,
		generated:  map[string]*companionFn{},
		needed:     map[string]paramNeeds{},
		needing:    map[string]bool{},
		generating: map[string]bool{},
	}
	if pkg != nil {
		for name, fns := range collectPackageFunctionsList(pkg) {
			c.own[name] = arities(fns)
		}
	}
	c.tainted = taintedFunctions(funcs, func(ref callRef) bool {
		dep := c.qualified(ref.prefix)
		return dep != nil && dep.tainted[ref.key()]
	})
	return c
}

func (c *companions) findFunc(name string, arity int) *program.Function {
	if fns, ok := c.funcs[name]; ok && len(fns) > 0 {
		for _, f := range fns {
			if arity < 0 || len(f.Parameters) == arity {
				return f
			}
		}
		return fns[0]
	}
	return nil
}

// isTainted reports whether the overload of name taking argc arguments
// transitively contains an operator.
func (c *companions) isTainted(name string, argc int) bool {
	return c.tainted[fnKey(name, argc)]
}

// callTainted is isTainted for a call as written: a qualified call asks the
// registry of the package its prefix names.
func (c *companions) callTainted(ref callRef) bool {
	if ref.prefix == "" {
		return c.tainted[ref.key()]
	}
	dep := c.qualified(ref.prefix)
	return dep != nil && dep.tainted[ref.key()]
}

// qualified is the registry a prefix names from this package, or nil when the
// prefix does not resolve to a package of the program being expanded.
func (c *companions) qualified(prefix string) *companions {
	if c.exp == nil || c.pkg == nil {
		return nil
	}
	return c.exp.qualified(c.pkg, prefix)
}

func companionName(fnName string) string { return fnName + "__d" }

// nameFor is the name a function's companion is declared under: the plain
// `__d` form when the name has one overload, so the programs written so far
// read as they did, and `__d<arity>` when it has several, since two
// companions of one name would collide as overloads once their tangent
// parameters are counted.
func (c *companions) nameFor(fnName string, arity int) string {
	count := c.own[fnName]
	if count == 0 {
		count = arities(c.funcs[fnName])
	}
	if count <= 1 {
		return companionName(fnName)
	}
	return fmt.Sprintf("%s__d%d", fnName, arity)
}

// arities counts a name's overloads as its distinct parameter counts, however
// many times a namespace lists each: a dependency is merged into a package's
// namespace as itself and again as the dot import that names it, so the
// builtin's normalizeField is listed more than once and is still one overload.
func arities(fns []*program.Function) int {
	seen := map[int]bool{}
	for _, f := range fns {
		if f != nil {
			seen[len(f.Parameters)] = true
		}
	}
	return len(seen)
}

// dpParamName is the companion parameter carrying a parameter's derivative
// along one spatial axis. Three of them rather than one packed gradient, so
// each has the parameter's own shape; see paramTangent.
func dpParamName(param string, axis int) string {
	return fmt.Sprintf("%s__dp%d", param, axis)
}

// call builds a call to fnName's companion, given the already-rewritten
// arguments, computing the required parameter tangents with deriv (the calling
// scope's derivative property, so tangents compose through nested companions).
// prefix is how the caller names this registry's package, "" from within it.
func (c *companions) call(prefix, fnName string, args []*program.Operand, deriv *derivative) (*program.Operand, error) {
	if err := c.ensure(fnName, len(args)); err != nil {
		return nil, err
	}
	fn := c.findFunc(fnName, len(args))
	if fn == nil {
		return nil, fmt.Errorf("unknown function %q", fnName)
	}
	if len(args) != len(fn.Parameters) {
		return nil, fmt.Errorf("function %q expects %d arguments, got %d", fnName, len(fn.Parameters), len(args))
	}
	needs := c.needed[fnKey(fnName, len(args))]
	callArgs := append([]*program.Operand(nil), args...)
	for _, name := range orderedParams(fn, needs) {
		i := paramIndex(fn, name)
		n := needs[name]
		if n.dp {
			for k := 0; k < 3; k++ {
				t, err := deriv.axisTangent(args[i], k)
				if err != nil {
					return nil, fmt.Errorf("companion %q arg %q gradient: %w", fnName, name, err)
				}
				callArgs = append(callArgs, t)
			}
		}
		if n.dt {
			t, err := deriv.temporal(args[i])
			if err != nil {
				return nil, fmt.Errorf("companion %q arg %q temporal: %w", fnName, name, err)
			}
			callArgs = append(callArgs, t)
		}
	}
	return &program.Operand{
		T: &program.Operand_FnCall{
			FnCall: &program.FnCall{
				Reference: &program.Reference{Prefix: prefix, Name: c.generated[fnKey(fnName, len(args))].name},
				Arguments: callArgs,
			},
		},
	}, nil
}

// ensure generates the companion of fnName's overload taking arity arguments
// once, recursively generating companions for any tainted functions it calls.
func (c *companions) ensure(fnName string, arity int) error {
	key := fnKey(fnName, arity)
	if _, ok := c.generated[key]; ok {
		return nil
	}
	if c.generating[key] {
		return fmt.Errorf("recursive function %q cannot be differentiated", fnName)
	}
	fn := c.findFunc(fnName, arity)
	if fn == nil {
		return fmt.Errorf("unknown function %q", fnName)
	}
	if len(fn.Returns) == 0 {
		return fmt.Errorf("cannot differentiate function %q with no return", fnName)
	}
	c.generating[key] = true
	defer delete(c.generating, key)

	needs, err := c.needsOf(fnName, arity)
	if err != nil {
		return err
	}

	// Build the scoped derivative property: package scope + this function's
	// locals/nested functions, with each needed parameter bound to its tangent
	// (a reference to the corresponding extra companion parameter).
	pts := map[string]*paramTangent{}
	for name, n := range needs {
		pt := &paramTangent{}
		if n.dp {
			for k := 0; k < 3; k++ {
				pt.dpAxis[k] = ref(dpParamName(name, k))
			}
		}
		if n.dt {
			pt.dt = ref(name + "__dt")
		}
		pts[name] = pt
	}
	mergedFuncs := make(map[string][]*program.Function)
	for name, fnVals := range c.funcs {
		mergedFuncs[name] = append(mergedFuncs[name], fnVals...)
	}
	for name, childFns := range collectFunctions(fn.FnTypes) {
		mergedFuncs[name] = append(mergedFuncs[name], childFns...)
	}
	decls := merge(c.decls, collectAssignments(fn.Assignments))
	// A parameter shadows a package declaration of the same name: Slider
	// Mechanism's top-level `a` was differentiated in place of
	// normalizeCurve's parameter a.
	for _, p := range fn.Parameters {
		delete(decls, p.Name)
	}
	scoped := &derivative{
		decls:   decls,
		funcs:   mergedFuncs,
		imports: c.imports,
		params:  pts,
	}
	r := &rewriter{deriv: scoped, comp: c}
	if c.names != nil {
		// A companion body's expansion may name the companion's own
		// parameters, so its shared nodes are locals of the body, each set
		// before the statement that uses it.
		r.share = func(expanded *program.Operand) *program.Operand {
			root, locals := sharedLocals(expanded, c.names)
			r.pending = append(r.pending, locals...)
			return root
		}
	}

	// Rewrite the body: resolve operators, preserve everything else.
	assigns, err := rewriteAssignments(r, fn.Assignments)
	if err != nil {
		return err
	}
	r.pending = nil
	// Every return, so a multi-return function's companion destructures the
	// same way the function does.
	rets := make([]*program.Operand, len(fn.Returns))
	for i, ret := range fn.Returns {
		rets[i], err = r.rewrite(ret)
		if err != nil {
			return err
		}
	}
	assigns = append(assigns, r.pending...)
	r.pending = nil

	params := make([]*program.Parameter, len(fn.Parameters))
	for i, p := range fn.Parameters {
		params[i] = &program.Parameter{Name: p.Name}
	}
	for _, name := range orderedParams(fn, needs) {
		n := needs[name]
		if n.dp {
			for k := 0; k < 3; k++ {
				params = append(params, &program.Parameter{Name: dpParamName(name, k)})
			}
		}
		if n.dt {
			params = append(params, &program.Parameter{Name: name + "__dt"})
		}
	}

	c.generated[key] = &companionFn{
		src:   fnName,
		arity: arity,
		name:  c.nameFor(fnName, arity),
		fn: &program.Function{
			Parameters:  params,
			Assignments: assigns,
			FnTypes:     fn.FnTypes,
			Returns:     rets,
		},
	}
	return nil
}

// rewriteAssignments rewrites the assignment bodies of a function's declarations,
// preserving names, nested functions, and structure.
func rewriteAssignments(r *rewriter, assigns []*program.Assignment) ([]*program.Assignment, error) {
	out := make([]*program.Assignment, 0, len(assigns))
	for _, a := range assigns {
		if a == nil {
			out = append(out, nil)
			continue
		}
		r.pending = nil
		body, err := r.rewrite(a.Expression)
		if err != nil {
			return nil, err
		}
		// Shared locals the rewrite wrote down go just before the statement.
		out = append(out, r.pending...)
		r.pending = nil
		out = append(out, &program.Assignment{
			Start:      a.Start,
			End:        a.End,
			Variables:  a.Variables,
			Expression: body,
		})
	}
	return out, nil
}

// --- taint analysis ---------------------------------------------------------

// taintedFunctions returns the set of functions, by fnKey, that transitively
// contain a derivative operator: directly in their body, or via a function
// they call. qualified answers for a callee named through a prefix, whose
// taint belongs to another package's registry; nil treats such calls as
// clean.
func taintedFunctions(funcs map[string][]*program.Function, qualified func(ref callRef) bool) map[string]bool {
	tainted := map[string]bool{}
	for name, fnList := range funcs {
		for _, fn := range fnList {
			if funcBodyHasOperator(fn) {
				tainted[fnKey(name, len(fn.Parameters))] = true
			}
		}
	}
	for changed := true; changed; {
		changed = false
		for name, fnList := range funcs {
			for _, fn := range fnList {
				key := fnKey(name, len(fn.Parameters))
				if tainted[key] {
					continue
				}
				for callee := range calledFunctions(fn) {
					calleeTainted := false
					if callee.prefix == "" {
						calleeTainted = tainted[callee.key()]
					} else if qualified != nil {
						calleeTainted = qualified(callee)
					}
					if calleeTainted {
						tainted[key] = true
						changed = true
						break
					}
				}
			}
		}
	}
	return tainted
}

// funcBodyHasOperator reports whether a function's own body (assignments and
// returns) contains a derivative operator.
func funcBodyHasOperator(fn *program.Function) bool {
	for _, a := range fn.Assignments {
		if a != nil && operandHasOperator(a.Expression) {
			return true
		}
	}
	for _, ret := range fn.Returns {
		if operandHasOperator(ret) {
			return true
		}
	}
	return false
}

// calledFunctions collects the user functions called anywhere in a function's
// body (including nested function bodies), qualified or not.
func calledFunctions(fn *program.Function) map[callRef]bool {
	out := map[callRef]bool{}
	for _, a := range fn.Assignments {
		if a != nil {
			collectCalls(a.Expression, out)
		}
	}
	for _, ft := range fn.FnTypes {
		if ft == nil {
			continue
		}
		for _, group := range ft.Groups {
			if group == nil {
				continue
			}
			for _, subFn := range group.Functions {
				if subFn != nil {
					for ref := range calledFunctions(subFn) {
						out[ref] = true
					}
				}
			}
		}
	}
	for _, ret := range fn.Returns {
		collectCalls(ret, out)
	}
	return out
}

// packageCallsTainted reports whether any of pkg's own declarations, function
// bodies or returns calls a tainted function, unqualified in its own scope or
// qualified in another package's.
func packageCallsTainted(pkg *program.Package, c *companions) bool {
	calls := map[callRef]bool{}
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, a := range f.Assignments {
			if a != nil {
				collectCalls(a.Expression, calls)
			}
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
						for ref := range calledFunctions(fn) {
							calls[ref] = true
						}
					}
				}
			}
		}
		for _, r := range f.Returns {
			collectCalls(r, calls)
		}
	}
	for ref := range calls {
		if c.callTainted(ref) {
			return true
		}
	}
	return false
}

func collectCalls(op *program.Operand, out map[callRef]bool) {
	if op == nil {
		return
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		for _, sub := range t.Expression.Operands {
			collectCalls(sub, out)
		}
	case *program.Operand_FnCall:
		if r := t.FnCall.Reference; r != nil {
			out[callRef{prefix: r.Prefix, name: r.Name, argc: len(t.FnCall.Arguments)}] = true
		}
		for _, arg := range t.FnCall.Arguments {
			collectCalls(arg, out)
		}
	}
}

// --- reachable-parameter analysis -------------------------------------------

// needsOf computes (and memoizes) which of fnName's parameters need which
// tangents: the DERIV operators in its own body, plus — transitively — its
// calls to other tainted functions, whose needy argument positions induce the
// same need on whatever parameters those arguments reach. Without the
// propagation, a function that only forwards its parameters into a
// differentiating callee (e.g. lengthNormalized -> normalizeField(f.dp)) got no
// tangent parameters, so the callee saw zero tangents — length(f.dp) folded
// to 0 and the division produced a build error.
func (c *companions) needsOf(fnName string, arity int) (paramNeeds, error) {
	key := fnKey(fnName, arity)
	if n, ok := c.needed[key]; ok {
		return n, nil
	}
	if c.needing[key] {
		// A call cycle; ensure reports it properly when generating.
		return paramNeeds{}, nil
	}
	fn := c.findFunc(fnName, arity)
	if fn == nil {
		return nil, fmt.Errorf("unknown function %q", fnName)
	}
	c.needing[key] = true
	defer delete(c.needing, key)
	needs, err := c.reachableParams(fn)
	if err != nil {
		return nil, err
	}
	c.needed[key] = needs
	return needs, nil
}

// reachableParams determines, for each derivative operator in a function's body
// (and each needy parameter of a tainted callee it invokes), which parameters
// the target depends on (transitively through locals), and hence which
// parameter tangents the companion must accept.
func (c *companions) reachableParams(fn *program.Function) (paramNeeds, error) {
	params := map[string]bool{}
	for _, p := range fn.Parameters {
		params[p.Name] = true
	}
	locals := collectAssignments(fn.Assignments)
	needs := paramNeeds{}

	mark := func(target *program.Operand, dp bool) {
		reached := map[string]bool{}
		reachParams(target, params, locals, map[string]bool{}, reached)
		for name := range reached {
			n := needs[name]
			if dp {
				n.dp = true
			} else {
				n.dt = true
			}
			needs[name] = n
		}
	}

	var walkErr error
	var walk func(op *program.Operand)
	walk = func(op *program.Operand) {
		if op == nil || walkErr != nil {
			return
		}
		switch t := op.T.(type) {
		case *program.Operand_Expression:
			if t.Expression.Operation == program.Operation_DERIV && len(t.Expression.Operands) == 2 {
				if ref, ok := t.Expression.Operands[1].T.(*program.Operand_Reference); ok {
					mark(t.Expression.Operands[0], ref.Reference.Name == program.PVarName)
				}
			}
			for _, sub := range t.Expression.Operands {
				walk(sub)
			}
		case *program.Operand_FnCall:
			// A call to a tainted function: each argument bound to a needy
			// callee parameter needs the same tangent of whatever caller
			// parameters that argument reaches. A qualified call's callee
			// belongs to the registry its prefix names.
			if r := t.FnCall.Reference; r != nil {
				argc := len(t.FnCall.Arguments)
				owner := c
				if r.Prefix != "" {
					owner = c.qualified(r.Prefix)
				}
				var callee *program.Function
				if owner != nil && owner.isTainted(r.Name, argc) {
					callee = owner.findFunc(r.Name, argc)
				}
				if callee != nil {
					calleeNeeds, err := owner.needsOf(r.Name, argc)
					if err != nil {
						walkErr = err
						return
					}
					for i, p := range callee.Parameters {
						n := calleeNeeds[p.Name]
						if n.dp {
							mark(t.FnCall.Arguments[i], true)
						}
						if n.dt {
							mark(t.FnCall.Arguments[i], false)
						}
					}
				}
			}
			for _, arg := range t.FnCall.Arguments {
				walk(arg)
			}
		}
	}
	for _, a := range fn.Assignments {
		if a != nil {
			walk(a.Expression)
		}
	}
	for _, r := range fn.Returns {
		walk(r)
	}
	return needs, walkErr
}

// reachParams collects the parameter names a target operand depends on,
// following local declarations.
func reachParams(op *program.Operand, params map[string]bool, locals map[string]*program.Operand, visiting, reached map[string]bool) {
	if op == nil {
		return
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		name := t.Reference.Name
		if params[name] {
			reached[name] = true
			return
		}
		if body, ok := locals[name]; ok && !visiting[name] {
			visiting[name] = true
			reachParams(body, params, locals, visiting, reached)
			delete(visiting, name)
		}
	case *program.Operand_Expression:
		switch t.Expression.Operation {
		case program.Operation_ACCESS:
			// operands[1:] are swizzle selectors, not value references.
			reachParams(t.Expression.Operands[0], params, locals, visiting, reached)
		case program.Operation_DERIV:
			reachParams(t.Expression.Operands[0], params, locals, visiting, reached)
		default:
			for _, sub := range t.Expression.Operands {
				reachParams(sub, params, locals, visiting, reached)
			}
		}
	case *program.Operand_FnCall:
		for _, a := range t.FnCall.Arguments {
			reachParams(a, params, locals, visiting, reached)
		}
	}
}

// --- ordering / scope helpers -----------------------------------------------

// orderedParams lists the parameters needing tangents in declaration order, for
// a deterministic companion signature and matching call-site argument order.
func orderedParams(fn *program.Function, needs paramNeeds) []string {
	var out []string
	for _, p := range fn.Parameters {
		if n, ok := needs[p.Name]; ok && (n.dp || n.dt) {
			out = append(out, p.Name)
		}
	}
	return out
}

func paramIndex(fn *program.Function, name string) int {
	for i, p := range fn.Parameters {
		if p.Name == name {
			return i
		}
	}
	return -1
}

func merge(a, b map[string]*program.Operand) map[string]*program.Operand {
	out := make(map[string]*program.Operand, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}
