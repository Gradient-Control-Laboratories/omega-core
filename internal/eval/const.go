// Package eval provides a constant-folding evaluator for Module operands.
//
// It is used where a caller needs to collapse an Operand subtree to concrete
// numeric values before dispatch — for example, extracting mesh() placement
// transforms in the canvas agent. Only a side-effect-free, runtime-independent
// subset of GSL is supported: literals, references to top-level constant
// assignments (including constants from imported packages when a GetImportFunc
// is supplied via ConstPkg), arithmetic and trig ops, vector and list
// construction, swizzles, indexing, a whitelist of math builtins, calls to
// user-defined single-return functions whose bodies are simple assignments
// (including body-nested lambdas, which close over their enclosing call's
// names), and map/reduce over such functions. Values carry their shape — a
// list of vectors such as a matrix stays a list of its column vectors, so
// column indexing and matrix products fold — and the exported entry points
// flatten the result to the flat []float64 callers consume. References to p
// or time produce an error.
package eval

import (
	"fmt"
	"math"

	"gradientcontrol.com/omega/core/internal/meshslots"
	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/program"
)

// Const evaluates an Operand to a scalar or vector of float64 values using the
// given top-level declarations as the name scope. Scalars return a length-1
// slice; vectors return a length-N slice. Returns an error if the operand
// cannot be folded to a constant.
func Const(op *program.Operand, assignments []*program.Assignment, fnTypes []*program.FnType) ([]float64, error) {
	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: assignments,
				FnTypes:     fnTypes,
			},
		},
	}
	return ConstPkg(op, pkg, nil)
}

// ConstPkg is Const with import support: references that do not resolve to one
// of the package's own top-level constants are looked up in dot-imported
// packages, and qualified references (prefix.name) resolve through the
// matching prefixed import. Imported packages are loaded lazily via getImport,
// so programs whose constant expressions never touch an import do not require
// one to be loadable. A nil getImport behaves like Const.
func ConstPkg(op *program.Operand, pkg *program.Package, getImport program.GetImportFunc) ([]float64, error) {
	return ConstPkgSlots(op, pkg, getImport, nil)
}

// ConstPkgSlots is ConstPkg with mesh-asset slot resolution: an
// `importStep "f"` / `importMesh "f"` expression folds to the slot id the given
// resolver assigns to (declaring package, kind, path). Pass the resolver from
// meshslots.Assign over the same module graph. A nil resolver makes those
// expressions unfoldable (an error), matching ConstPkg.
func ConstPkgSlots(op *program.Operand, pkg *program.Package, getImport program.GetImportFunc, slots meshslots.Resolver) ([]float64, error) {
	return ConstPkgSlotsAt(op, pkg, pkg.Path, getImport, slots)
}

// ConstPkgSlotsAt evaluates an imported package with its module path.
func ConstPkgSlotsAt(op *program.Operand, pkg *program.Package, pkgPath string, getImport program.GetImportFunc, slots meshslots.Resolver) ([]float64, error) {
	prog := &program.Program{
		Primary: pkg,
	}
	progScope, err := scope.Init(prog)
	if err != nil {
		return nil, err
	}
	ctx := &constCtx{
		progScope:  progScope,
		cur:        progScope.Primary,
		pkgPaths:   map[*scope.PkgScope]string{progScope.Primary: pkgPath},
		getImport:  getImport,
		evaluating: make(map[evalKey]bool),
		calls:      make(map[callKey]bool),
		slots:      slots,
	}
	v, err := ctx.operand(op)
	if err != nil {
		return nil, err
	}
	return v.floats(), nil
}

// ConstElementsPkg evaluates an operand as a collection and returns one
// value per element. A vector yields one scalar per component; a list (e.g.
// a map whose function returns vectors) yields one value per element, each
// possibly a vector. A scalar yields a single element.
func ConstElementsPkg(op *program.Operand, pkg *program.Package, getImport program.GetImportFunc) ([][]float64, error) {
	prog := &program.Program{
		Primary: pkg,
	}
	progScope, err := scope.Init(prog)
	if err != nil {
		return nil, err
	}
	ctx := &constCtx{
		progScope:  progScope,
		cur:        progScope.Primary,
		pkgPaths:   map[*scope.PkgScope]string{progScope.Primary: pkg.Path},
		getImport:  getImport,
		evaluating: make(map[evalKey]bool),
		calls:      make(map[callKey]bool),
	}

	v, err := ctx.operand(op)
	if err != nil {
		return nil, err
	}
	elements := v.elements()
	out := make([][]float64, len(elements))
	for i, e := range elements {
		out[i] = e.floats()
	}
	return out, nil
}

// evalKey identifies an in-flight constant declaration evaluation for cycle
// detection. The scope pointer disambiguates same-named constants in
// different packages.
type evalKey struct {
	scope *scope.PkgScope
	name  string
}

// callKey identifies an in-flight user function call for recursion detection,
// by function identity (names cannot: hoisted lambdas are all gsl_anon* and
// repeat across enclosing functions) plus the return being evaluated.
type callKey struct {
	fn  *program.Function
	ret int // return index; 0 for single-return calls
}

type constCtx struct {
	progScope  *scope.ProgScope
	cur        *scope.PkgScope // scope of the package whose operand is being evaluated
	pkgPaths   map[*scope.PkgScope]string
	getImport  program.GetImportFunc
	evaluating map[evalKey]bool   // declaration cycle detection
	calls      map[callKey]bool   // in-flight user calls (recursion detection)
	slots      meshslots.Resolver // resolves importStep/importMesh calls to slot ids (may be nil)
	checker    ParameterChecker   // literals lifted to runtime parameters are not constants (may be nil)

	// frames holds the local bindings of in-flight user function calls. The
	// innermost frame and its chain of defining frames (see frame.parent) are
	// visible: a named function's body must not see its caller's locals, while
	// a lambda closes over the call it was written in.
	frames []*frame
	cache  map[*program.Operand]value
}

// frame is one in-flight user function call's local scope. vals caches
// parameter values and already-folded locals; locals holds body assignments
// not yet folded, so evaluating one return of a multi-return function folds
// only the locals that return depends on (leaving unrelated, possibly
// non-constant locals untouched). fns holds the body's nested function
// declarations (the parser hoists lambdas here as gsl_anon* names). owner is
// the package scope the body's free references resolve against. parent is
// the defining frame a lambda closes over (nil for named functions), and
// folding guards against cyclic locals.
type frame struct {
	vals    map[string]value
	locals  map[string]*program.Operand
	multi   map[string]multiLocalInfo
	fns     map[string][]*program.Function
	owner   *scope.PkgScope
	parent  *frame
	folding map[string]bool
}

type multiLocalInfo struct {
	expr  *program.Operand
	index int
}

func selectUnifiedFn(elem scope.NamedElem, paramCount int) *program.Function {
	if fnTypeScope, ok := elem.(*scope.FnTypeScope); ok && fnTypeScope != nil {
		if gScope, ok := fnTypeScope.Groups[int32(paramCount)]; ok && gScope != nil {
			if len(gScope.Functions) > 0 {
				return gScope.Functions[0].Function
			}
		}
	}
	return nil
}

// topFrame is the innermost in-flight call frame, or nil outside any call.
func (c *constCtx) topFrame() *frame {
	if len(c.frames) == 0 {
		return nil
	}
	return c.frames[len(c.frames)-1]
}

func (fr *frame) initLocals(assignments []*program.Assignment) {
	fr.locals = make(map[string]*program.Operand)
	fr.multi = make(map[string]multiLocalInfo)
	for _, a := range assignments {
		if a == nil {
			continue
		}
		if len(a.Variables) == 1 {
			fr.locals[a.Variables[0].Name] = a.Expression
		} else {
			for i, v := range a.Variables {
				if v != nil {
					fr.multi[v.Name] = multiLocalInfo{
						expr:  a.Expression,
						index: i,
					}
				}
			}
		}
	}
}

// buildFnScope collects a function body's nested function declarations (the
// parser hoists lambdas here as named gsl_anon* functions) by name.
func buildFnScope(fnTypes []*program.FnType) map[string][]*program.Function {
	if len(fnTypes) == 0 {
		return nil
	}
	fns := make(map[string][]*program.Function, len(fnTypes))
	for _, ft := range fnTypes {
		if ft == nil {
			continue
		}
		for _, group := range ft.Groups {
			if group == nil {
				continue
			}
			for _, fn := range group.Functions {
				if fn != nil {
					fns[ft.Name] = append(fns[ft.Name], fn)
				}
			}
		}
	}
	return fns
}

func (c *constCtx) operand(op *program.Operand) (value, error) {
	if op == nil {
		return value{}, fmt.Errorf("nil operand")
	}
	if len(c.frames) == 0 && c.cache != nil {
		if cached, ok := c.cache[op]; ok {
			return cached, nil
		}
	}
	val, err := c.evalOperand(op)
	if err == nil && len(c.frames) == 0 && c.cache != nil {
		c.cache[op] = val
	}
	return val, err
}

func (c *constCtx) evalOperand(op *program.Operand) (value, error) {
	switch v := op.T.(type) {
	case *program.Operand_Value:
		// A literal lifted to a runtime parameter (e.g. a slider uniform) has no
		// constant value. Failing here makes every reference to a variable
		// assigned from it — and every expression over those — non-constant
		// too, since references re-evaluate their assignment through this path.
		if c.checker != nil && c.checker.IsParameter(op) {
			return value{}, ErrParameter
		}
		return scalarVal(v.Value), nil

	case *program.Operand_Reference:
		return c.reference(v.Reference)

	case *program.Operand_Expression:
		return c.expression(v.Expression)

	case *program.Operand_FnCall:
		return c.functionCall(v.FnCall)

	case *program.Operand_PropRef:
		return value{}, fmt.Errorf("property references are not supported in constant context")

	default:
		return value{}, fmt.Errorf("unsupported operand type %T", op.T)
	}
}

func (c *constCtx) reference(ref *program.Reference) (value, error) {
	if ref == nil {
		return value{}, fmt.Errorf("nil reference")
	}

	// 1. Qualified prefix reference
	if ref.Prefix != "" {
		var pkgRef *scope.PkgRef
		for _, fileScope := range c.cur.Files {
			if r, ok := fileScope.Imports[ref.Prefix]; ok {
				pkgRef = r
				break
			}
		}
		if pkgRef == nil {
			var imp *program.Import
			for _, fileScope := range c.cur.Files {
				for _, fileImp := range fileScope.File.Imports {
					if fileImp.PrefixOrSimpleName() == ref.Prefix {
						imp = fileImp
						break
					}
				}
				if imp != nil {
					break
				}
			}
			if imp == nil {
				return value{}, fmt.Errorf("unknown import prefix %q in constant context", ref.Prefix)
			}
			_, err := c.importScope(imp)
			if err != nil {
				return value{}, err
			}
			for _, fileScope := range c.cur.Files {
				if r, ok := fileScope.Imports[ref.Prefix]; ok {
					pkgRef = r
					break
				}
			}
		}
		if pkgRef == nil {
			return value{}, fmt.Errorf("prefix %q is not a package reference", ref.Prefix)
		}
		owner := pkgRef.PkgScope()
		elem := pkgRef.NamedElem(ref.Name)
		if elem == nil {
			return value{}, fmt.Errorf("reference %q not found in import %q", ref.Name, pkgRef.Name())
		}
		if varElem, ok := elem.(*scope.Var); ok {
			assign := varElem.Parent.Assign
			if len(assign.Variables) == 1 {
				return c.evalIn(owner, ref.Name, assign.Expression)
			} else {
				var index int
				for i, vr := range assign.Variables {
					if vr == varElem.Var {
						index = i
						break
					}
				}
				return c.evalReturnK(owner, assign.Expression, index)
			}
		}
		return value{}, fmt.Errorf("qualified reference %q.%q is not a constant variable", ref.Prefix, ref.Name)
	}

	// 2. Unqualified reference
	name := ref.Name

	// Check local/frame scope
	for fr := c.topFrame(); fr != nil; fr = fr.parent {
		if v, ok := fr.vals[name]; ok {
			return v, nil
		}
		if body, ok := fr.locals[name]; ok {
			return c.foldLocal(fr, name, body)
		}
		if info, ok := fr.multi[name]; ok {
			return c.foldMultiLocal(fr, name, info)
		}
	}

	// Check package-level variables
	elem := c.cur.NamedElem(name)
	if elem != nil {
		if varElem, ok := elem.(*scope.Var); ok {
			assign := varElem.Parent.Assign
			if len(assign.Variables) == 1 {
				return c.evalIn(c.cur, name, assign.Expression)
			} else {
				var index int
				for i, vr := range assign.Variables {
					if vr == varElem.Var {
						index = i
						break
					}
				}
				return c.evalReturnK(c.cur, assign.Expression, index)
			}
		}
	}

	// Check dot-imports (including lazy loading)
	var dotImports []*program.Import
	for _, fileScope := range c.cur.Files {
		var builtinImp *program.Import
		for _, imp := range fileScope.File.Imports {
			if imp.PrefixOrSimpleName() == "." {
				if imp.Path == "builtin" {
					builtinImp = imp
				} else {
					dotImports = append(dotImports, imp)
				}
			}
		}
		if builtinImp != nil {
			dotImports = append(dotImports, builtinImp)
		}
	}

	var foundElem scope.NamedElem
	var foundOwner *scope.PkgScope
	for _, imp := range dotImports {
		impScope, err := c.importScope(imp)
		if err != nil {
			return value{}, err
		}
		if elem := impScope.NamedElem(name); elem != nil {
			if impScope.Package.Path == "builtin" {
				foundElem = elem
				foundOwner = impScope
				continue
			}
			foundElem = elem
			foundOwner = impScope
			break
		}
	}
	if foundElem != nil {
		if varElem, ok := foundElem.(*scope.Var); ok {
			assign := varElem.Parent.Assign
			if len(assign.Variables) == 1 {
				return c.evalIn(foundOwner, name, assign.Expression)
			} else {
				var index int
				for i, vr := range assign.Variables {
					if vr == varElem.Var {
						index = i
						break
					}
				}
				return c.evalReturnK(foundOwner, assign.Expression, index)
			}
		}
	}

	return value{}, fmt.Errorf("reference %q is not a top-level constant", name)
}

// foldLocal lazily folds a function body local the first time it is referenced
// and caches the result in the frame, so a local is folded at most once and
// only when actually needed. The body folds with its own frame innermost and
// in its owner's package — a closure may reference an enclosing call's local,
// whose body must not resolve names against the closure's frame.
func (c *constCtx) foldLocal(fr *frame, name string, body *program.Operand) (value, error) {
	if fr.folding[name] {
		return value{}, fmt.Errorf("cyclic reference through %q", name)
	}
	if fr.folding == nil {
		fr.folding = make(map[string]bool)
	}
	fr.folding[name] = true
	prev := c.cur
	c.cur = fr.owner
	c.frames = append(c.frames, fr)
	v, err := c.operand(body)
	c.frames = c.frames[:len(c.frames)-1]
	c.cur = prev
	delete(fr.folding, name)
	if err != nil {
		return value{}, err
	}
	fr.vals[name] = v
	return v, nil
}

func (c *constCtx) foldMultiLocal(fr *frame, name string, info multiLocalInfo) (value, error) {
	if fr.folding[name] {
		return value{}, fmt.Errorf("cyclic reference through %q", name)
	}
	if fr.folding == nil {
		fr.folding = make(map[string]bool)
	}
	fr.folding[name] = true
	defer delete(fr.folding, name)

	vVal, err := c.operand(info.expr)
	if err != nil {
		return value{}, err
	}
	elements := vVal.elements()
	if info.index < 0 || info.index >= len(elements) {
		return value{}, fmt.Errorf("multi-value index %d out of range (length %d)", info.index, len(elements))
	}
	v := elements[info.index]
	fr.vals[name] = v
	return v, nil
}

// evalReturnK evaluates the index-th return value of a multi-return function
// call — the right-hand side of a destructuring assignment such as
// `a, b = f()`. Only that one return is evaluated, and the callee's body
// locals fold lazily (see frame), so unrelated returns (e.g. p-dependent SDFs)
// and unrelated locals never need to be constant. The assignment is a
// package-level statement, so the in-flight frame stack is hidden while its
// arguments and the callee body evaluate (matching evalIn).
func (c *constCtx) evalReturnK(scope *scope.PkgScope, callOp *program.Operand, index int) (value, error) {
	if callOp == nil {
		return value{}, fmt.Errorf("destructuring assignment has no value")
	}
	fc := callOp.GetFnCall()
	if fc == nil {
		return value{}, fmt.Errorf("destructuring assignment value is not a function call")
	}
	if fc.Reference == nil {
		return value{}, fmt.Errorf("nil function call reference")
	}

	prev := c.cur
	prevFrames := c.frames
	c.cur = scope
	c.frames = nil
	defer func() {
		c.cur = prev
		c.frames = prevFrames
	}()

	args := make([]value, len(fc.Arguments))
	for i, a := range fc.Arguments {
		v, err := c.operand(a)
		if err != nil {
			return value{}, fmt.Errorf("%s arg %d: %w", fc.Reference.Name, i, err)
		}
		args[i] = v
	}

	fn, owner, _, err := c.lookupFunction(fc.Reference.Prefix, fc.Reference.Name, len(args))
	if err != nil {
		return value{}, err
	}
	if index < 0 || index >= len(fn.Returns) {
		return value{}, fmt.Errorf("return index %d out of range for %q (%d returns)", index, fc.Reference.Name, len(fn.Returns))
	}

	key := callKey{fn, index}
	if c.calls[key] {
		return value{}, fmt.Errorf("recursive call to %q in constant context", fc.Reference.Name)
	}
	c.calls[key] = true
	defer delete(c.calls, key)

	fr := &frame{
		vals:  make(map[string]value, len(fn.Parameters)),
		fns:   buildFnScope(fn.FnTypes),
		owner: owner,
	}
	fr.initLocals(fn.Assignments)
	for i, param := range fn.Parameters {
		if i < len(args) && param != nil {
			fr.vals[param.Name] = args[i]
		}
	}

	c.cur = owner
	c.frames = append(c.frames, fr)
	defer func() { c.frames = c.frames[:len(c.frames)-1] }()
	return c.operand(fn.Returns[index])
}

// evalIn evaluates a declaration body in the scope of the package that owns
// it, so references inside an imported constant resolve against that
// package's own declarations and imports. A package-level body is not inside
// any call, so the in-flight frame stack is hidden while it evaluates: its
// free names must not resolve against an unrelated caller's locals or
// lambdas.
func (c *constCtx) evalIn(owner *scope.PkgScope, name string, body *program.Operand) (value, error) {
	key := evalKey{owner, name}
	if c.evaluating[key] {
		return value{}, fmt.Errorf("cyclic reference through %q", name)
	}
	c.evaluating[key] = true
	prev := c.cur
	prevFrames := c.frames
	c.cur = owner
	c.frames = nil
	defer func() {
		c.cur = prev
		c.frames = prevFrames
		delete(c.evaluating, key)
	}()
	return c.operand(body)
}

// importScope loads (once) and returns the scope for an imported package.
func (c *constCtx) importScope(imp *program.Import) (*scope.PkgScope, error) {
	var depScope *scope.PkgScope
	if s, ok := c.progScope.Dependencies[imp.Path]; ok {
		depScope = s
	} else {
		if c.getImport == nil {
			return nil, fmt.Errorf("import %q not available in constant context", imp.Path)
		}
		pkg, err := c.getImport(imp.Path)
		if err != nil {
			return nil, fmt.Errorf("import %q: %w", imp.Path, err)
		}
		if pkg == nil {
			return nil, fmt.Errorf("import %q: package not found", imp.Path)
		}
		if pkg.Path == "" {
			pkg.Path = imp.Path
		}
		depScope, err = c.progScope.AddDependency(pkg)
		if err != nil {
			return nil, err
		}
	}

	impName := imp.PrefixOrSimpleName()
	for _, fileScope := range c.cur.Files {
		for _, fileImp := range fileScope.File.Imports {
			if fileImp.Path == imp.Path {
				fileScope.AddImportRef(impName, depScope)
			}
		}
	}
	if c.pkgPaths != nil {
		c.pkgPaths[depScope] = imp.Path
	}
	return depScope, nil
}

func (c *constCtx) expression(expr *program.Expression) (value, error) {
	if expr == nil {
		return value{}, fmt.Errorf("nil expression")
	}

	// COLLECT of scalars builds a vector (the common [x, y, z] case); when any
	// operand is itself a vector or list, the collected values keep their shape
	// as a list of elements, so a list-of-vectors literal such as a 4x4 matrix
	// [[..],[..],[..],[..]] stays a list of its column vectors (its flattened
	// form is the same column-major order as before).
	if expr.Operation == program.Operation_COLLECT {
		vals := make([]value, len(expr.Operands))
		allScalar := true
		for i, arg := range expr.Operands {
			v, err := c.operand(arg)
			if err != nil {
				return value{}, fmt.Errorf("collect operand %d: %w", i, err)
			}
			vals[i] = v
			if !v.isScalar() {
				allScalar = false
			}
		}
		if allScalar {
			out := make([]float64, len(vals))
			for i, v := range vals {
				out[i] = v.scalars[0]
			}
			return vectorVal(out), nil
		}
		return listVal(vals), nil
	}

	// ACCESS: first operand is a vector, second encodes the swizzle as a
	// 1-based component index stored as a literal value (parser convention).
	if expr.Operation == program.Operation_ACCESS {
		return c.access(expr.Operands)
	}

	// INDEX, MAP, and REDUCE resolve their function/target operands
	// themselves, so they dispatch before the eager operand evaluation below.
	if expr.Operation == program.Operation_INDEX {
		return c.index(expr.Operands)
	}
	if expr.Operation == program.Operation_MAP {
		return c.mapValues(expr.Operands)
	}
	if expr.Operation == program.Operation_REDUCE {
		return c.reduceValues(expr.Operands)
	}
	if expr.Operation == program.Operation_SELECT {
		return c.selectValues(expr.Operands)
	}

	// Evaluate all operands first for the element-wise ops below.
	vals := make([]value, len(expr.Operands))
	for i, arg := range expr.Operands {
		v, err := c.operand(arg)
		if err != nil {
			return value{}, fmt.Errorf("%s operand %d: %w", expr.Operation.Name(), i, err)
		}
		vals[i] = v
	}

	switch expr.Operation {
	// Variadic reductions over scalars/vectors.
	case program.Operation_ADD:
		return reduceBinary(vals, func(a, b float64) float64 { return a + b })
	case program.Operation_SUB:
		return reduceBinary(vals, func(a, b float64) float64 { return a - b })
	case program.Operation_MUL:
		return mulValues(vals)
	case program.Operation_DIV:
		return reduceBinary(vals, func(a, b float64) float64 { return a / b })
	case program.Operation_MODULO:
		return reduceBinary(vals, func(a, b float64) float64 {
			if b == 0 {
				return math.NaN()
			}
			return a - b*math.Floor(a/b)
		})
	case program.Operation_MIN:
		return reduceBinary(vals, math.Min)
	case program.Operation_MAX:
		return reduceBinary(vals, math.Max)

	// Unary scalar-or-vector ops (applied elementwise).
	case program.Operation_NEG:
		return unary(vals, func(x float64) float64 { return -x })
	case program.Operation_ABS:
		return unary(vals, math.Abs)
	case program.Operation_SQUARE:
		return unary(vals, func(x float64) float64 { return x * x })
	case program.Operation_SQRT:
		return unary(vals, math.Sqrt)
	case program.Operation_FLOOR:
		return unary(vals, math.Floor)
	case program.Operation_CEIL:
		return unary(vals, math.Ceil)
	case program.Operation_ROUND:
		return unary(vals, func(x float64) float64 { return math.Round(x) })
	case program.Operation_EXP:
		return unary(vals, math.Exp)
	case program.Operation_LN:
		return unary(vals, math.Log)
	case program.Operation_SIN:
		return unary(vals, math.Sin)
	case program.Operation_COS:
		return unary(vals, math.Cos)
	case program.Operation_TAN:
		return unary(vals, math.Tan)
	case program.Operation_ASIN:
		return unary(vals, math.Asin)
	case program.Operation_ACOS:
		return unary(vals, math.Acos)
	case program.Operation_ATAN:
		return unary(vals, math.Atan)
	case program.Operation_ATAN2:
		return reduceBinary(vals, math.Atan2)
	case program.Operation_POW:
		return reduceBinary(vals, math.Pow)
	case program.Operation_EQ:
		return reduceBinary(vals, func(a, b float64) float64 {
			if a == b {
				return 1
			}
			return 0
		})
	case program.Operation_NE:
		return reduceBinary(vals, func(a, b float64) float64 {
			if a != b {
				return 1
			}
			return 0
		})
	case program.Operation_LT:
		return reduceBinary(vals, func(a, b float64) float64 {
			if a < b {
				return 1
			}
			return 0
		})
	case program.Operation_LE:
		return reduceBinary(vals, func(a, b float64) float64 {
			if a <= b {
				return 1
			}
			return 0
		})
	case program.Operation_GT:
		return reduceBinary(vals, func(a, b float64) float64 {
			if a > b {
				return 1
			}
			return 0
		})
	case program.Operation_GE:
		return reduceBinary(vals, func(a, b float64) float64 {
			if a >= b {
				return 1
			}
			return 0
		})
	case program.Operation_AND:
		return reduceBinary(vals, func(a, b float64) float64 {
			if a != 0 && b != 0 {
				return 1
			}
			return 0
		})
	case program.Operation_OR:
		return reduceBinary(vals, func(a, b float64) float64 {
			if a != 0 || b != 0 {
				return 1
			}
			return 0
		})
	case program.Operation_NOT:
		return unary(vals, func(x float64) float64 {
			if x == 0 {
				return 1
			}
			return 0
		})
	}

	return value{}, fmt.Errorf("operation %s not supported in constant context", expr.Operation.Name())
}

// access handles an ACCESS expression: a vector operand plus the swizzle.
// The parser stores the swizzle as a reference operand whose name is the
// member string (e.g. "x", "xy", "zyx"); a literal value operand encoding a
// 1-based component index is also accepted.
func (c *constCtx) access(operands []*program.Operand) (value, error) {
	if len(operands) < 2 {
		return value{}, fmt.Errorf("access requires at least 2 operands")
	}
	base, err := c.operand(operands[0])
	if err != nil {
		return value{}, fmt.Errorf("access base: %w", err)
	}
	var vec []float64
	if base.isList() {
		allScalars := true
		for _, e := range base.elems {
			if !e.isScalar() {
				allScalars = false
				break
			}
		}
		if !allScalars {
			return value{}, fmt.Errorf("member access requires a vector, not a list")
		}
		vec = base.floats()
	} else {
		vec = base.scalars
	}
	out := make([]float64, 0, len(operands)-1)
	for i, idxOp := range operands[1:] {
		if ref := idxOp.GetReference(); ref != nil && ref.Prefix == "" {
			member := ref.Name
			for _, ch := range member {
				var idx int
				switch ch {
				case 'x':
					idx = 0
				case 'y':
					idx = 1
				case 'z':
					idx = 2
				case 'w':
					idx = 3
				default:
					return value{}, fmt.Errorf("access member %q has invalid component %q", member, string(ch))
				}
				if idx >= len(vec) {
					return value{}, fmt.Errorf("access component %q out of range for length %d", string(ch), len(vec))
				}
				out = append(out, vec[idx])
			}
			continue
		}
		idxVals, err := c.operand(idxOp)
		if err != nil {
			return value{}, fmt.Errorf("access index %d: %w", i, err)
		}
		if !idxVals.isScalar() {
			return value{}, fmt.Errorf("access index %d must be scalar", i)
		}
		idx := int(idxVals.scalars[0]) - 1 // 1-based in the literal encoding
		if idx < 0 || idx >= len(vec) {
			return value{}, fmt.Errorf("access index %d out of range for length %d", idx+1, len(vec))
		}
		out = append(out, vec[idx])
	}
	return vectorVal(out), nil
}

// importSlot folds an importStep/importMesh call to its compiler-assigned slot
// id in the declaring package.
func (c *constCtx) importSlot(kind string, arg *program.Operand) (value, error) {
	path, ok := meshslots.PathFromArg(arg)
	if !ok {
		return value{}, fmt.Errorf("import%s path must be a string literal", kind)
	}
	if c.slots == nil {
		return value{}, fmt.Errorf("import%s %q: slot assignment unavailable", kind, path)
	}
	pkgPath := ""
	if path, ok := c.pkgPaths[c.cur]; ok {
		pkgPath = path
	} else if c.cur != nil && c.cur.Package != nil {
		pkgPath = c.cur.Package.Path
	}
	slot, ok := c.slots(pkgPath, kind, path)
	if !ok {
		return value{}, fmt.Errorf("import%s %q: no slot assigned", kind, path)
	}
	return scalarVal(float64(slot)), nil
}

func (c *constCtx) functionCall(fc *program.FnCall) (value, error) {
	if fc == nil {
		return value{}, fmt.Errorf("nil function call")
	}
	if fc.Reference == nil {
		return value{}, fmt.Errorf("nil function call reference")
	}

	// importStep "f" / importMesh "f" fold to their compiler-assigned slot id.
	// Intercept before arg folding: the single argument is the path rune list,
	// not a value to evaluate.
	if kind, ok := meshslots.KindOf(fc.Reference.Name); ok && fc.Reference.Prefix == "" && len(fc.Arguments) == 1 {
		return c.importSlot(kind, fc.Arguments[0])
	}

	args := make([]value, len(fc.Arguments))
	for i, a := range fc.Arguments {
		v, err := c.operand(a)
		if err != nil {
			return value{}, fmt.Errorf("%s arg %d: %w", fc.Reference.Name, i, err)
		}
		args[i] = v
	}

	// The vector builtins below consume flat component values; a list argument
	// flattens to the same values the flat evaluator produced.
	flat := func(i int) []float64 { return args[i].floats() }

	name := fc.Reference.Name
	switch name {
	case "uiPoint", "uiTransform", "uiSpline", "uiColor":
		if len(args) != 1 {
			return value{}, fmt.Errorf("%s requires 1 argument", name)
		}
		return args[0], nil
	case "uiDirection":
		// The second argument, the arrow's anchor, is for the viewer alone.
		if len(args) < 1 || len(args) > 2 {
			return value{}, fmt.Errorf("%s requires 1 or 2 arguments", name)
		}
		return args[0], nil
	case "radians":
		return unary(args, func(x float64) float64 { return x * math.Pi / 180 })
	case "degrees":
		return unary(args, func(x float64) float64 { return x * 180 / math.Pi })
	case "sinh":
		return unary(args, math.Sinh)
	case "cosh":
		return unary(args, math.Cosh)
	case "tanh":
		return unary(args, math.Tanh)
	case "asinh":
		return unary(args, math.Asinh)
	case "acosh":
		return unary(args, math.Acosh)
	case "atanh":
		return unary(args, math.Atanh)
	case "exp2":
		return unary(args, math.Exp2)
	case "fract":
		return unary(args, func(x float64) float64 { _, f := math.Modf(x); return f })
	case "sign":
		return unary(args, func(x float64) float64 {
			switch {
			case x > 0:
				return 1
			case x < 0:
				return -1
			}
			return 0
		})
	case "trunc":
		return unary(args, math.Trunc)
	case "clamp":
		if len(args) != 3 {
			return value{}, fmt.Errorf("clamp requires exactly 3 arguments")
		}
		return clampVal(args[0], args[1], args[2])
	case "length":
		if len(args) != 1 {
			return value{}, fmt.Errorf("length requires 1 argument")
		}
		return scalarVal(vecLength(flat(0))), nil
	case "distance":
		if len(args) != 2 {
			return value{}, fmt.Errorf("distance requires two vectors of equal length")
		}
		a, b := flat(0), flat(1)
		if len(a) != len(b) {
			return value{}, fmt.Errorf("distance requires two vectors of equal length")
		}
		diff := make([]float64, len(a))
		for i := range diff {
			diff[i] = a[i] - b[i]
		}
		return scalarVal(vecLength(diff)), nil
	case "dot":
		if len(args) != 2 {
			return value{}, fmt.Errorf("dot requires two vectors of equal length")
		}
		a, b := flat(0), flat(1)
		if len(a) != len(b) {
			return value{}, fmt.Errorf("dot requires two vectors of equal length")
		}
		var sum float64
		for i := range a {
			sum += a[i] * b[i]
		}
		return scalarVal(sum), nil
	case "cross":
		if len(args) != 2 {
			return value{}, fmt.Errorf("cross requires two vec3 arguments")
		}
		a, b := flat(0), flat(1)
		if len(a) != 3 || len(b) != 3 {
			return value{}, fmt.Errorf("cross requires two vec3 arguments")
		}
		return vectorVal([]float64{
			a[1]*b[2] - a[2]*b[1],
			a[2]*b[0] - a[0]*b[2],
			a[0]*b[1] - a[1]*b[0],
		}), nil
	case "normalize":
		if len(args) != 1 {
			return value{}, fmt.Errorf("normalize requires 1 argument")
		}
		v := flat(0)
		l := vecLength(v)
		if l == 0 {
			return value{}, fmt.Errorf("cannot normalize zero vector")
		}
		out := make([]float64, len(v))
		for i, x := range v {
			out[i] = x / l
		}
		return vectorVal(out), nil
	}
	// TODO extract literal strings "matrix" (package name) and "multiply" (function name) to constants
	if fc.Reference.Prefix == "matrix" && name == "multiply" && len(args) == 2 {
		if args[0].isList() && args[1].isList() {
			return matMulValues(args[0], args[1])
		}
		if args[0].isList() && !args[1].isList() {
			return matVecValues(args[0], args[1])
		}
		out, err := matrixMultiply(flat(0), flat(1))
		if err != nil {
			return value{}, err
		}
		return vectorVal(out), nil
	}
	// TODO extract literal strings "vector" (package name) and "multiply" (function name) to constants
	if fc.Reference.Prefix == "vector" && name == "multiply" && len(args) == 2 {
		if args[0].isList() && !args[1].isList() {
			return matVecValues(args[0], args[1])
		}
	}
	if fc.Reference.Prefix == "vector" {
		switch name {
		case "addVector", "addScalar":
			if len(args) == 2 {
				return applyBinary(args[0], args[1], func(x, y float64) float64 { return x + y })
			}
		case "subtractVector", "subtractScalarFrom", "subtractFromScalar":
			if len(args) == 2 {
				return applyBinary(args[0], args[1], func(x, y float64) float64 { return x - y })
			}
		case "multiplyScalar", "multiplyComponents":
			if len(args) == 2 {
				return applyBinary(args[0], args[1], func(x, y float64) float64 { return x * y })
			}
		case "divideByScalar", "divideScalarBy", "divideComponents":
			if len(args) == 2 {
				return applyBinary(args[0], args[1], func(x, y float64) float64 { return x / y })
			}
		case "negate":
			if len(args) == 1 {
				return unary(args, func(x float64) float64 { return -x })
			}
		}
	}
	return c.userFunctionCall(fc.Reference.Prefix, name, args)
}

// matrixMultiply is the flat column-major square-matrix product backing the
// import-free matrix.multiply fallback above. Only square sizes are
// recognized: a flat argument's length cannot distinguish rectangular
// shapes.
func matrixMultiply(a, b []float64) ([]float64, error) {
	var k int
	switch {
	case len(a) == 4 && len(b) == 4:
		k = 2
	case len(a) == 9 && len(b) == 9:
		k = 3
	case len(a) == 16 && len(b) == 16:
		k = 4
	default:
		return nil, fmt.Errorf("matrix.multiply folds only square matrices (2x2, 3x3, 4x4), got %d and %d values", len(a), len(b))
	}
	out := make([]float64, k*k)
	for j := 0; j < k; j++ {
		for i := 0; i < k; i++ {
			var sum float64
			for m := 0; m < k; m++ {
				sum += a[m*k+i] * b[j*k+m]
			}
			out[j*k+i] = sum
		}
	}
	return out, nil
}

// userFunctionCall evaluates a call to a user-defined GSL function:
// the arguments bind to a fresh frame, the body's assignments fold lazily
// on first reference, and the return operands are evaluated.
func (c *constCtx) userFunctionCall(prefix, name string, args []value) (value, error) {
	fn, teardown, err := c.enterUserCall(prefix, name, args)
	if err != nil {
		return value{}, err
	}
	defer teardown()

	if len(fn.Returns) == 1 {
		return c.operand(fn.Returns[0])
	}
	elements := make([]value, len(fn.Returns))
	for i, r := range fn.Returns {
		el, err := c.operand(r)
		if err != nil {
			return value{}, err
		}
		elements[i] = el
	}
	return listVal(elements), nil
}

// enterUserCall resolves a user-defined function, pushes a frame
// binding its parameters to args, and switches the evaluation scope to the
// function's package. Body assignments are held unfolded in the frame's locals/multi
// and fold lazily on first reference.
func (c *constCtx) enterUserCall(prefix, name string, args []value) (*program.Function, func(), error) {
	fn, owner, def, err := c.lookupFunction(prefix, name, len(args))
	if err != nil {
		return nil, nil, err
	}

	key := callKey{fn, 0}
	if c.calls[key] {
		return nil, nil, fmt.Errorf("recursive call to %q in constant context", name)
	}
	c.calls[key] = true

	fr := &frame{
		vals:   make(map[string]value, len(fn.Parameters)),
		fns:    buildFnScope(fn.FnTypes),
		owner:  owner,
		parent: def,
	}
	fr.initLocals(fn.Assignments)
	for i, param := range fn.Parameters {
		if param != nil {
			fr.vals[param.Name] = args[i]
		}
	}

	prev := c.cur
	c.cur = owner
	c.frames = append(c.frames, fr)
	teardown := func() {
		c.cur = prev
		c.frames = c.frames[:len(c.frames)-1]
		delete(c.calls, key)
	}
	return fn, teardown, nil
}

// lookupFunction resolves a (possibly package-qualified) function name to the
// overload with the given parameter count, along with the scope of the
// package that declares it and — for a body-nested lambda — the frame that
// declares it (nil for package-level functions). Unqualified names check the
// frame chain first (the innermost call's hoisted gsl_anon* lambdas, then its
// defining frames'), then the current package, then its dot imports.
func (c *constCtx) lookupFunction(prefix, name string, paramCount int) (*program.Function, *scope.PkgScope, *frame, error) {
	if prefix != "" {
		var pkgRef *scope.PkgRef
		for _, fileScope := range c.cur.Files {
			if r, ok := fileScope.Imports[prefix]; ok {
				pkgRef = r
				break
			}
		}
		if pkgRef == nil {
			var imp *program.Import
			for _, fileScope := range c.cur.Files {
				for _, fileImp := range fileScope.File.Imports {
					if fileImp.PrefixOrSimpleName() == prefix {
						imp = fileImp
						break
					}
				}
				if imp != nil {
					break
				}
			}
			if imp == nil {
				return nil, nil, nil, fmt.Errorf("unknown import prefix %q in constant context", prefix)
			}
			_, err := c.importScope(imp)
			if err != nil {
				return nil, nil, nil, err
			}
			for _, fileScope := range c.cur.Files {
				if r, ok := fileScope.Imports[prefix]; ok {
					pkgRef = r
					break
				}
			}
		}
		if pkgRef == nil {
			return nil, nil, nil, fmt.Errorf("prefix %q is not a package reference", prefix)
		}
		impScope := pkgRef.PkgScope()
		elem := impScope.NamedElem(name)
		if fn := selectUnifiedFn(elem, paramCount); fn != nil {
			return fn, impScope, nil, nil
		}
		return nil, nil, nil, fmt.Errorf("function %q.%q with %d parameters not supported in constant context", prefix, name, paramCount)
	}

	for fr := c.topFrame(); fr != nil; fr = fr.parent {
		fns, ok := fr.fns[name]
		if !ok {
			continue
		}
		// The innermost declaration of a hoisted lambda name shadows any
		// outer or package-level one (anon names repeat across scopes), so
		// the search stops at the first frame declaring the name.
		if fn := selectOverload(fns, paramCount); fn != nil {
			return fn, fr.owner, fr, nil
		}
		return nil, nil, nil, fmt.Errorf("function %q with %d parameters not supported in constant context", name, paramCount)
	}

	elem := c.cur.NamedElem(name)
	if fn := selectUnifiedFn(elem, paramCount); fn != nil {
		return fn, c.cur, nil, nil
	}

	var dotImports []*program.Import
	for _, fileScope := range c.cur.Files {
		var builtinImp *program.Import
		for _, imp := range fileScope.File.Imports {
			if imp.PrefixOrSimpleName() == "." {
				if imp.Path == "builtin" {
					builtinImp = imp
				} else {
					dotImports = append(dotImports, imp)
				}
			}
		}
		if builtinImp != nil {
			dotImports = append(dotImports, builtinImp)
		}
	}
	for _, imp := range dotImports {
		impScope, err := c.importScope(imp)
		if err != nil {
			return nil, nil, nil, err
		}
		elem := impScope.NamedElem(name)
		if fn := selectUnifiedFn(elem, paramCount); fn != nil {
			return fn, impScope, nil, nil
		}
	}

	return nil, nil, nil, fmt.Errorf("function %q not supported in constant context", name)
}

func selectOverload(fns []*program.Function, paramCount int) *program.Function {
	for _, fn := range fns {
		if len(fn.Parameters) == paramCount {
			return fn
		}
	}
	return nil
}

// index evaluates target[k]. Indexing a plain vector yields the scalar
// element; indexing a list yields the whole element value, so a
// list-of-vectors such as a matrix yields its k-th COLUMN vector — matching
// the full build.
func (c *constCtx) index(operands []*program.Operand) (value, error) {
	if len(operands) != 2 {
		return value{}, fmt.Errorf("index requires 2 operands, got %d", len(operands))
	}
	idxVals, err := c.operand(operands[1])
	if err != nil {
		return value{}, fmt.Errorf("index value: %w", err)
	}
	if !idxVals.isScalar() || idxVals.scalars[0] != math.Trunc(idxVals.scalars[0]) {
		return value{}, fmt.Errorf("index must be a scalar integer")
	}
	k := int(idxVals.scalars[0])

	target, err := c.operand(operands[0])
	if err != nil {
		return value{}, fmt.Errorf("index target: %w", err)
	}
	return target.elementAt(k)
}

// mapValues fully evaluates map(fn, ...collections): each element maps
// through the function, scalar results forming a vector and vector results a
// list of vectors. A single result is returned as-is, since a singleton is
// equivalent to a list of length one — both mirroring the build-time map.
func (c *constCtx) mapValues(operands []*program.Operand) (value, error) {
	elements, err := c.mapElements(operands)
	if err != nil {
		return value{}, err
	}
	results := make([]value, len(elements))
	allScalar := true
	for i, element := range elements {
		v, err := c.callMapFn(operands[0], float64(i), element)
		if err != nil {
			return value{}, err
		}
		results[i] = v
		if !v.isScalar() {
			allScalar = false
		}
	}
	if len(results) == 1 {
		return results[0], nil
	}
	if allScalar {
		out := make([]float64, len(results))
		for i, v := range results {
			out[i] = v.scalars[0]
		}
		return vectorVal(out), nil
	}
	return listVal(results), nil
}

// reduceValues folds reduce in either of its two forms, mirroring the build:
// reduce(fn, collection) seeds the accumulator with the collection's first
// element and folds fn over the rest, while reduce(initial, fn, ...collections)
// folds from the explicit initial value, which may be a vector (it is not
// expanded to elements). fn takes (accumulator, value) and returns the next
// accumulator; the final accumulator is the result.
func (c *constCtx) reduceValues(operands []*program.Operand) (value, error) {
	if len(operands) < 2 {
		return value{}, fmt.Errorf("reduce expects at least 2 arguments (a function and a collection, optionally preceded by an initial value)")
	}
	var acc value
	var fnOp *program.Operand
	var elements []value
	var err error
	if len(operands) == 2 {
		fnOp = operands[0]
		elements, err = c.mapElements(operands)
		if err != nil {
			return value{}, err
		}
		if len(elements) == 0 {
			return value{}, fmt.Errorf("reduce of an empty collection")
		}
		acc = elements[0]
		elements = elements[1:]
	} else {
		acc, err = c.operand(operands[0])
		if err != nil {
			return value{}, fmt.Errorf("reduce initial value: %w", err)
		}
		fnOp = operands[1]
		elements, err = c.mapElements(operands[1:])
		if err != nil {
			return value{}, err
		}
	}
	for _, element := range elements {
		acc, err = c.callFn2(fnOp, acc, element)
		if err != nil {
			return value{}, err
		}
	}
	return acc, nil
}

// callFn2 invokes a two-parameter function operand — a reduce's
// (accumulator, value) fold function — with the given arguments.
func (c *constCtx) callFn2(fnOp *program.Operand, a, b value) (value, error) {
	ref := fnOp.GetReference()
	if ref == nil {
		return value{}, fmt.Errorf("reduce expects a function reference")
	}
	return c.userFunctionCall(ref.Prefix, ref.Name, []value{a, b})
}

// mapElements expands a map call's collection operands to elements, matching
// the two calling conventions: a single collection operand expands to its
// top-level elements (a list of vectors yields its column vectors), while
// multiple scalar operands each contribute one.
func (c *constCtx) mapElements(operands []*program.Operand) ([]value, error) {
	if len(operands) < 2 {
		return nil, fmt.Errorf("map expects a function and a collection")
	}
	first, err := c.operand(operands[1])
	if err != nil {
		return nil, fmt.Errorf("map collection: %w", err)
	}
	if first.count() > 1 || first.isList() {
		if len(operands) > 2 {
			return nil, fmt.Errorf("map expects either a single collection or multiple scalar arguments")
		}
		return first.elements(), nil
	}
	elements := first.elements()
	for i, op := range operands[2:] {
		v, err := c.operand(op)
		if err != nil {
			return nil, fmt.Errorf("map argument %d: %w", i+2, err)
		}
		if !v.isScalar() {
			return nil, fmt.Errorf("map expects either a single collection or multiple scalar arguments")
		}
		elements = append(elements, v)
	}
	return elements, nil
}

// callMapFn invokes the map function operand with the element value, plus
// the element index when the function takes two parameters. A one-parameter
// overload is preferred, mirroring the build-time map dispatch.
func (c *constCtx) callMapFn(fnOp *program.Operand, index float64, element value) (value, error) {
	var prefix, name string
	if ref := fnOp.GetReference(); ref != nil {
		prefix, name = ref.Prefix, ref.Name
	} else {
		return value{}, fmt.Errorf("map expects a function reference as its first argument")
	}
	if _, _, _, err := c.lookupFunction(prefix, name, 1); err == nil {
		return c.userFunctionCall(prefix, name, []value{element})
	}
	return c.userFunctionCall(prefix, name, []value{element, scalarVal(index)})
}

// reduceBinary folds a variadic op left-to-right. Scalars promote to any
// vector operand length; mixed vectors must all match.
func reduceBinary(vals []value, op func(a, b float64) float64) (value, error) {
	if len(vals) == 0 {
		return value{}, fmt.Errorf("binary op requires at least 1 operand")
	}
	acc := vals[0]
	for i := 1; i < len(vals); i++ {
		next := vals[i]
		var err error
		acc, err = applyBinary(acc, next, op)
		if err != nil {
			return value{}, err
		}
	}
	return acc, nil
}

// mulValues dispatches `*` by operand shape, mirroring the build's matrix
// operator routing: when a list (matrix) operand is involved the product is a
// real matrix product — matrix*matrix multiplies (rectangular allowed),
// matrix*vector applies the matrix to a column vector, and a scalar scales.
// Before values carried shape this case silently folded as an elementwise
// (Hadamard) product of the flattened matrices. Leaf operands keep the
// elementwise/broadcast fold.
func mulValues(vals []value) (value, error) {
	anyList := false
	for _, v := range vals {
		if v.isList() {
			anyList = true
			break
		}
	}
	if !anyList {
		return reduceBinary(vals, func(a, b float64) float64 { return a * b })
	}
	if len(vals) != 2 {
		return value{}, fmt.Errorf("matrix '*' takes two operands; parenthesize chained products")
	}
	a, b := vals[0], vals[1]
	mul := func(x, s float64) float64 { return x * s }
	switch {
	case a.isList() && b.isList():
		return matMulValues(a, b)
	case a.isList() && b.isScalar():
		return applyBinary(a, b, mul)
	case a.isScalar() && b.isList():
		return applyBinary(a, b, mul)
	case a.isList():
		return matVecValues(a, b)
	}
	return value{}, fmt.Errorf("vector * matrix needs a transposed row vector (not yet supported); write M * v, or vector.multiply(matrix.transpose(M), v)")
}

// matMulValues is the matrix product a·b for two lists of column vectors:
// a is R×K, b is K×C, and result column j is the linear combination of a's
// columns weighted by b's column j — matching matrix.multiply, with any
// rectangular sizes allowed.
func matMulValues(a, b value) (value, error) {
	k := len(a.elems)
	if k == 0 {
		return value{}, fmt.Errorf("cannot multiply an empty matrix")
	}
	rows := -1
	for _, acol := range a.elems {
		if acol.isList() {
			return value{}, fmt.Errorf("matrix '*' requires matrices of column vectors")
		}
		if rows == -1 {
			rows = len(acol.scalars)
		} else if len(acol.scalars) != rows {
			return value{}, fmt.Errorf("matrix '*' requires uniform column lengths")
		}
	}
	out := make([]value, len(b.elems))
	for j, bcol := range b.elems {
		if bcol.isList() {
			return value{}, fmt.Errorf("matrix '*' requires matrices of column vectors")
		}
		if len(bcol.scalars) != k {
			return value{}, fmt.Errorf("cannot multiply a matrix of %d columns by a matrix with %d-row columns", k, len(bcol.scalars))
		}
		col := make([]float64, rows)
		for m, w := range bcol.scalars {
			for i, x := range a.elems[m].scalars {
				col[i] += x * w
			}
		}
		out[j] = vectorVal(col)
	}
	return listVal(out), nil
}

// matVecValues applies a matrix (list of column vectors) to a column vector:
// the linear combination of the columns weighted by v's components —
// matching vector.multiply.
func matVecValues(m, v value) (value, error) {
	if len(m.elems) != len(v.scalars) {
		return value{}, fmt.Errorf("cannot multiply a matrix of %d columns by a vector of length %d", len(m.elems), len(v.scalars))
	}
	var out []float64
	for k, col := range m.elems {
		if col.isList() {
			return value{}, fmt.Errorf("matrix '*' requires a matrix of column vectors")
		}
		if out == nil {
			out = make([]float64, len(col.scalars))
		} else if len(col.scalars) != len(out) {
			return value{}, fmt.Errorf("matrix '*' requires uniform column lengths")
		}
		for i, x := range col.scalars {
			out[i] += x * v.scalars[k]
		}
	}
	return vectorVal(out), nil
}

// applyBinary combines two shaped values elementwise: matching lists recurse
// per element, a scalar broadcasts over a list, and leaves combine as flat
// vectors (equal lengths elementwise, a scalar broadcasting over a vector).
// Structurally mixed shapes fall back to combining their flattened values,
// preserving the flat evaluator's behavior.
func applyBinary(a, b value, op func(a, b float64) float64) (value, error) {
	switch {
	case a.isList() && b.isList() && len(a.elems) == len(b.elems):
		out := make([]value, len(a.elems))
		for i := range a.elems {
			v, err := applyBinary(a.elems[i], b.elems[i], op)
			if err != nil {
				return value{}, err
			}
			out[i] = v
		}
		return listVal(out), nil
	case a.isList() && b.isScalar():
		out := make([]value, len(a.elems))
		for i := range a.elems {
			v, err := applyBinary(a.elems[i], b, op)
			if err != nil {
				return value{}, err
			}
			out[i] = v
		}
		return listVal(out), nil
	case a.isScalar() && b.isList():
		out := make([]value, len(b.elems))
		for i := range b.elems {
			v, err := applyBinary(a, b.elems[i], op)
			if err != nil {
				return value{}, err
			}
			out[i] = v
		}
		return listVal(out), nil
	case !a.isList() && !b.isList():
		return applyBinaryFlat(a.scalars, b.scalars, op)
	}
	return applyBinaryFlat(a.floats(), b.floats(), op)
}

// applyBinaryFlat is the flat-leaf combination: equal lengths apply
// elementwise and a scalar broadcasts over a vector.
func applyBinaryFlat(a, b []float64, op func(a, b float64) float64) (value, error) {
	switch {
	case len(a) == len(b):
		out := make([]float64, len(a))
		for i := range a {
			out[i] = op(a[i], b[i])
		}
		return vectorVal(out), nil
	case len(a) == 1:
		out := make([]float64, len(b))
		for i := range b {
			out[i] = op(a[0], b[i])
		}
		return vectorVal(out), nil
	case len(b) == 1:
		out := make([]float64, len(a))
		for i := range a {
			out[i] = op(a[i], b[0])
		}
		return vectorVal(out), nil
	}
	return value{}, fmt.Errorf("cannot combine vectors of length %d and %d", len(a), len(b))
}

func unary(vals []value, op func(x float64) float64) (value, error) {
	if len(vals) != 1 {
		return value{}, fmt.Errorf("unary op requires exactly 1 operand, got %d", len(vals))
	}
	return unaryVal(vals[0], op), nil
}

// unaryVal applies op to every scalar of a shaped value, preserving shape.
func unaryVal(v value, op func(x float64) float64) value {
	if v.isList() {
		out := make([]value, len(v.elems))
		for i, e := range v.elems {
			out[i] = unaryVal(e, op)
		}
		return listVal(out)
	}
	out := make([]float64, len(v.scalars))
	for i, x := range v.scalars {
		out[i] = op(x)
	}
	return vectorVal(out)
}

func vecLength(v []float64) float64 {
	var sum float64
	for _, x := range v {
		sum += x * x
	}
	return math.Sqrt(sum)
}

func clampVal(v, loVal, hiVal value) (value, error) {
	if v.isList() {
		if loVal.isList() && len(loVal.elems) != len(v.elems) {
			return value{}, fmt.Errorf("clamp bounds size mismatch: lo %d elements != target %d elements", len(loVal.elems), len(v.elems))
		}
		if hiVal.isList() && len(hiVal.elems) != len(v.elems) {
			return value{}, fmt.Errorf("clamp bounds size mismatch: hi %d elements != target %d elements", len(hiVal.elems), len(v.elems))
		}
		out := make([]value, len(v.elems))
		for i, e := range v.elems {
			var lo, hi value
			if loVal.isList() {
				lo = loVal.elems[i]
			} else {
				lo = loVal
			}
			if hiVal.isList() {
				hi = hiVal.elems[i]
			} else {
				hi = hiVal
			}
			res, err := clampVal(e, lo, hi)
			if err != nil {
				return value{}, err
			}
			out[i] = res
		}
		return listVal(out), nil
	}

	if loVal.isList() || hiVal.isList() {
		return value{}, fmt.Errorf("clamp bounds incompatible with target vector")
	}

	if !loVal.isScalar() && len(loVal.scalars) != len(v.scalars) {
		return value{}, fmt.Errorf("clamp lo bound length %d != target vector length %d", len(loVal.scalars), len(v.scalars))
	}
	if !hiVal.isScalar() && len(hiVal.scalars) != len(v.scalars) {
		return value{}, fmt.Errorf("clamp hi bound length %d != target vector length %d", len(hiVal.scalars), len(v.scalars))
	}

	out := make([]float64, len(v.scalars))
	for i, x := range v.scalars {
		var lo, hi float64
		if loVal.isScalar() {
			lo = loVal.scalars[0]
		} else {
			lo = loVal.scalars[i]
		}
		if hiVal.isScalar() {
			hi = hiVal.scalars[0]
		} else {
			hi = hiVal.scalars[i]
		}
		out[i] = math.Max(lo, math.Min(hi, x))
	}
	return vectorVal(out), nil
}

func (c *constCtx) selectValues(operands []*program.Operand) (value, error) {
	if len(operands) != 3 && len(operands) != 4 {
		return value{}, fmt.Errorf("select requires 3 or 4 operands, got %d", len(operands))
	}
	cond, err := c.operand(operands[0])
	if err != nil {
		return value{}, fmt.Errorf("select condition: %w", err)
	}
	condFloats := cond.floats()
	if len(condFloats) == 0 {
		return value{}, fmt.Errorf("select condition has no values")
	}

	if len(operands) == 3 {
		trueVal, err := c.operand(operands[1])
		if err != nil {
			return value{}, fmt.Errorf("select true branch: %w", err)
		}
		falseVal, err := c.operand(operands[2])
		if err != nil {
			return value{}, fmt.Errorf("select false branch: %w", err)
		}

		if cond.isScalar() {
			if condFloats[0] != 0 {
				return trueVal, nil
			}
			return falseVal, nil
		}

		// Vector condition: elementwise select
		trueFloats := trueVal.floats()
		falseFloats := falseVal.floats()
		out := make([]float64, len(condFloats))
		for i, cv := range condFloats {
			if cv != 0 {
				if i < len(trueFloats) {
					out[i] = trueFloats[i]
				} else if len(trueFloats) > 0 {
					out[i] = trueFloats[len(trueFloats)-1]
				}
			} else {
				if i < len(falseFloats) {
					out[i] = falseFloats[i]
				} else if len(falseFloats) > 0 {
					out[i] = falseFloats[len(falseFloats)-1]
				}
			}
		}
		return vectorVal(out), nil
	}

	// 4 operands (3-way select: cond > 0 ? pos : cond == 0 ? zero : neg)
	posVal, err := c.operand(operands[1])
	if err != nil {
		return value{}, fmt.Errorf("select pos branch: %w", err)
	}
	zeroVal, err := c.operand(operands[2])
	if err != nil {
		return value{}, fmt.Errorf("select zero branch: %w", err)
	}
	negVal, err := c.operand(operands[3])
	if err != nil {
		return value{}, fmt.Errorf("select neg branch: %w", err)
	}

	if cond.isScalar() {
		cv := condFloats[0]
		if cv > 0 {
			return posVal, nil
		} else if cv == 0 {
			return zeroVal, nil
		}
		return negVal, nil
	}

	posFloats := posVal.floats()
	zeroFloats := zeroVal.floats()
	negFloats := negVal.floats()
	out := make([]float64, len(condFloats))
	for i, cv := range condFloats {
		if cv > 0 {
			if i < len(posFloats) {
				out[i] = posFloats[i]
			} else if len(posFloats) > 0 {
				out[i] = posFloats[len(posFloats)-1]
			}
		} else if cv == 0 {
			if i < len(zeroFloats) {
				out[i] = zeroFloats[i]
			} else if len(zeroFloats) > 0 {
				out[i] = zeroFloats[len(zeroFloats)-1]
			}
		} else {
			if i < len(negFloats) {
				out[i] = negFloats[i]
			} else if len(negFloats) > 0 {
				out[i] = negFloats[len(negFloats)-1]
			}
		}
	}
	return vectorVal(out), nil
}
