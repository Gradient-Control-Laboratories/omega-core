package lower

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/program"
)

func isParameter(fnStack []*program.Function, op *program.Operand) bool {
	if op == nil || len(fnStack) == 0 {
		return false
	}
	ref, ok := op.T.(*program.Operand_Reference)
	if !ok || ref.Reference == nil || ref.Reference.Prefix != "" {
		return false
	}
	for i := len(fnStack) - 1; i >= 0; i-- {
		fn := fnStack[i]
		if fn != nil {
			for _, p := range fn.Parameters {
				if p.Name == ref.Reference.Name {
					return true
				}
			}
		}
	}
	return false
}

func lowerMap(prog *program.Program, pkg *program.Package, file *program.PackageFile, fnStack []*program.Function, op *program.Operand, expr *program.Expression) (*program.Operand, error) {
	if len(expr.Operands) < 2 {
		return nil, fmt.Errorf("map expects at least 2 operands (a function and a collection), got %d", len(expr.Operands))
	}

	collOp := expr.Operands[1]
	if len(expr.Operands) == 2 && (isParameter(fnStack, collOp) || (len(fnStack) > 0 && collOp.GetReference() != nil)) {
		return op, nil
	}

	fnOp := expr.Operands[0]
	prefix, name, arity := resolveFnArity(prog, pkg, file, fnStack, fnOp)
	if name == "" {
		return nil, fmt.Errorf("map first argument must be a function")
	}

	// Multiple scalar arguments: map(fn, a, b, c)
	if len(expr.Operands) > 2 {
		results := make([]*program.Operand, len(expr.Operands)-1)
		for i, arg := range expr.Operands[1:] {
			var args []*program.Operand
			if arity == 2 {
				args = []*program.Operand{arg, program.NewValue(float64(i), op.Start, op.End)}
			} else {
				args = []*program.Operand{arg}
			}
			results[i] = program.NewFnCall(prefix, name, args, op.Start, op.End)
		}
		return program.NewExpression(program.Operation_COLLECT, results, op.Start, op.End), nil
	}

	elements := extractElements(prog, file, fnStack, collOp)
	if len(elements) == 0 {
		return op, nil
	}

	results := make([]*program.Operand, len(elements))
	for i, elem := range elements {
		var args []*program.Operand
		if arity == 2 {
			args = []*program.Operand{elem, program.NewValue(float64(i), op.Start, op.End)}
		} else {
			args = []*program.Operand{elem}
		}
		results[i] = program.NewFnCall(prefix, name, args, op.Start, op.End)
	}

	if len(elements) == 1 && collOp.Type != nil && !isHardwareMatrix(collOp.Type) && !collOp.Type.IsList() {
		return results[0], nil
	}

	return program.NewExpression(program.Operation_COLLECT, results, op.Start, op.End), nil
}

func lowerReduce(prog *program.Program, pkg *program.Package, file *program.PackageFile, fnStack []*program.Function, op *program.Operand, expr *program.Expression) (*program.Operand, error) {
	if len(expr.Operands) < 2 {
		return nil, fmt.Errorf("reduce expects at least 2 arguments (a function and a collection), got %d", len(expr.Operands))
	}

	var fnOp *program.Operand
	var initialOp *program.Operand
	var collOps []*program.Operand

	if len(expr.Operands) == 2 {
		// Form: reduce(fn, collection)
		fnOp = expr.Operands[0]
		collOps = expr.Operands[1:]
	} else if len(expr.Operands) == 3 {
		if isFnOperand(prog, pkg, file, fnStack, expr.Operands[0]) {
			// GSL v2 Form: reduce(fn, collection, initial)
			fnOp = expr.Operands[0]
			collOps = []*program.Operand{expr.Operands[1]}
			initialOp = expr.Operands[2]
		} else {
			// GSL v1 Legacy Form: reduce(initial, fn, collection)
			initialOp = expr.Operands[0]
			fnOp = expr.Operands[1]
			collOps = []*program.Operand{expr.Operands[2]}
		}
	} else {
		// Multi-argument collection: reduce(initial, fn, a, b, c) or reduce(fn, a, b, c)
		if isFnOperand(prog, pkg, file, fnStack, expr.Operands[0]) {
			fnOp = expr.Operands[0]
			collOps = expr.Operands[1:]
		} else {
			initialOp = expr.Operands[0]
			fnOp = expr.Operands[1]
			collOps = expr.Operands[2:]
		}
	}

	var literal []*program.Operand
	if len(collOps) == 1 {
		literal = literalListOf(file, fnStack, collOps[0])
	}
	if len(collOps) == 1 && (isParameter(fnStack, collOps[0]) || (len(fnStack) > 0 && collOps[0].GetReference() != nil && literal == nil)) {
		return op, nil
	}

	prefix, name, arity := resolveFnArity(prog, pkg, file, fnStack, fnOp)
	if name == "" {
		return nil, fmt.Errorf("reduce callback argument must be a function")
	}

	var elements []*program.Operand
	if literal != nil {
		elements = literal
	} else if len(collOps) == 1 {
		elements = extractElements(prog, file, fnStack, collOps[0])
	} else {
		for _, c := range collOps {
			els := extractElements(prog, file, fnStack, c)
			if len(els) > 0 {
				elements = append(elements, els...)
			} else {
				elements = append(elements, c)
			}
		}
	}

	if len(elements) == 0 {
		if len(fnStack) > 0 {
			return op, nil
		}
		return nil, fmt.Errorf("reduce expects at least one element in collection")
	}

	// The backend loops over constants with a number or a short vector as
	// the accumulator, and unrolls anything else it is handed less well than
	// this does, so only such a reduce is kept.
	if len(fnStack) > 0 && initialOp != nil && constantWidth(initialOp) > 0 && len(collOps) == 1 &&
		len(elements) >= LoopFoldThreshold && (arity == 2 || arity == 3) && constantElements(elements) &&
		loopable(prog, pkg, file, fnStack, fnOp) {
		// The list written in place, as the backend reads constants only
		// from a list it can see.
		for i, o := range expr.Operands {
			if o == collOps[0] && o.GetExpression().GetOperation() != program.Operation_COLLECT {
				written := make([]*program.Operand, len(elements))
				for j, el := range elements {
					written[j] = proto.Clone(el).(*program.Operand)
				}
				expr.Operands[i] = program.NewExpression(program.Operation_COLLECT, written, o.Start, o.End)
			}
		}
		return op, nil
	}

	return unrollElements(elements, initialOp, prefix, name, arity, op.Start, op.End)
}

func unrollElements(elements []*program.Operand, initialOp *program.Operand, prefix, name string, arity int, start, end *program.Location) (*program.Operand, error) {
	var acc *program.Operand
	startIdx := 0

	if initialOp != nil {
		acc = initialOp
		startIdx = 0
	} else {
		acc = elements[0]
		startIdx = 1
	}

	for i := startIdx; i < len(elements); i++ {
		var args []*program.Operand
		if arity == 3 {
			args = []*program.Operand{acc, elements[i], program.NewValue(float64(i), start, end)}
		} else {
			args = []*program.Operand{acc, elements[i]}
		}
		acc = program.NewFnCall(prefix, name, args, start, end)
	}

	return acc, nil
}

// constantElements reports whether every element is a number or a list of
// numbers written out, each the same width, so a loop can read them from one
// constant array.
func constantElements(elements []*program.Operand) bool {
	width := -1
	for _, el := range elements {
		w := constantWidth(el)
		if w < 1 || (width >= 0 && w != width) {
			return false
		}
		width = w
	}
	return true
}

// constantWidth is how many numbers el is, or 0 if it is not numbers written
// out: a number, a negated one, or a list of up to four of those.
func constantWidth(el *program.Operand) int {
	if el == nil {
		return 0
	}
	if _, ok := el.T.(*program.Operand_Value); ok {
		return 1
	}
	expr := el.GetExpression()
	if expr == nil {
		return 0
	}
	switch expr.Operation {
	case program.Operation_NEG:
		if len(expr.Operands) == 1 && constantWidth(expr.Operands[0]) == 1 {
			return 1
		}
	case program.Operation_COLLECT:
		if len(expr.Operands) < 2 || len(expr.Operands) > 4 {
			return 0
		}
		for _, c := range expr.Operands {
			if constantWidth(c) != 1 {
				return 0
			}
		}
		return len(expr.Operands)
	}
	return 0
}

// literalListOf returns copies of the elements of the list literal that op,
// a plain name inside a function, refers to: the nearest enclosing function's
// assignment to that name, or failing that a file-level one. It returns nil
// outside a function, for anything but a plain name, and when the name is a
// parameter or is bound where it is visible to anything but a list literal,
// so that a reduce over it is left as it was.
//
// Unrolling such a reduce here, as one over a literal written in place is,
// lets the inliner bind each step's accumulator once. Left for the backend,
// the accumulator's expression is copied into every use of it at each step,
// so a body that uses it twice doubles at every element. A map has no
// accumulator, so it is left alone.
func literalListOf(file *program.PackageFile, fnStack []*program.Function, op *program.Operand) []*program.Operand {
	ref := op.GetReference()
	if len(fnStack) == 0 || ref == nil || ref.Prefix != "" {
		return nil
	}
	for i := len(fnStack) - 1; i >= 0; i-- {
		fn := fnStack[i]
		if fn == nil {
			continue
		}
		for _, p := range fn.Parameters {
			if p.Name == ref.Name {
				return nil
			}
		}
		if assign := assignmentTo(fn.Assignments, ref.Name); assign != nil {
			return literalElements(assign)
		}
	}
	if file == nil {
		return nil
	}
	if assign := assignmentTo(file.Assignments, ref.Name); assign != nil {
		return literalElements(assign)
	}
	return nil
}

// assignmentTo returns the assignment among assigns that binds name, if any.
func assignmentTo(assigns []*program.Assignment, name string) *program.Assignment {
	for _, assign := range assigns {
		for _, v := range assign.Variables {
			if v.Name == name {
				return assign
			}
		}
	}
	return nil
}

// literalElements returns copies of the elements of the list literal assign
// binds its one name to, or nil if it binds anything else.
func literalElements(assign *program.Assignment) []*program.Operand {
	if len(assign.Variables) != 1 || assign.Expression == nil {
		return nil
	}
	collect, ok := assign.Expression.T.(*program.Operand_Expression)
	if !ok || collect.Expression.Operation != program.Operation_COLLECT || len(collect.Expression.Operands) == 0 {
		return nil
	}
	elements := make([]*program.Operand, len(collect.Expression.Operands))
	for i, el := range collect.Expression.Operands {
		elements[i] = proto.Clone(el).(*program.Operand)
	}
	return elements
}

// extractElements extracts the constituent element operands from a collection operand.
//
// When lowering map/reduce operations into unrolled expressions, the compiler needs concrete
// operands for each element in the sequence. This function extracts elements from:
//  1. Explicit inline collect(...) expressions
//  2. Variable references resolving to a collect(...) assignment across dependency packages,
//     enclosing functions on the stack (innermost first), or file-level assignments
//  3. Sized non-parameter types (unrolled via index access expressions)
//  4. Scalar operands (treated as a single-element list)
//
// Returns nil if the collection cannot be statically resolved into individual elements.
func extractElements(
	prog *program.Program,
	file *program.PackageFile,
	fnStack []*program.Function,
	op *program.Operand,
) []*program.Operand {
	return extractElementsWithVisited(prog, file, fnStack, op, make(map[string]bool))
}

func extractElementsWithVisited(
	prog *program.Program,
	file *program.PackageFile,
	fnStack []*program.Function,
	op *program.Operand,
	visited map[string]bool,
) []*program.Operand {
	if op == nil {
		return nil
	}

	if expr, ok := op.T.(*program.Operand_Expression); ok && expr.Expression.Operation == program.Operation_COLLECT {
		return expr.Expression.Operands
	}

	if ref, ok := op.T.(*program.Operand_Reference); ok && ref.Reference != nil {
		refKey := ref.Reference.Prefix + "." + ref.Reference.Name
		if visited[refKey] {
			return nil
		}
		visited[refKey] = true

		// Check cross-package imported variable references (e.g. prefix.varName)
		// by searching matching dependency packages for an assigned collect(...) expression.
		if ref.Reference.Prefix != "" && prog != nil {
			for _, dep := range prog.Dependencies {
				if dep == nil {
					continue
				}
				simpleName := dep.Path[strings.LastIndex(dep.Path, "/")+1:]
				if dep.Path == ref.Reference.Prefix || simpleName == ref.Reference.Prefix || strings.HasSuffix(dep.Path, "/"+ref.Reference.Prefix) {
					for _, f := range dep.Files {
						for _, assign := range f.Assignments {
							for _, v := range assign.Variables {
								if v.Name == ref.Reference.Name && assign.Expression != nil {
									if elems := extractElementsWithVisited(prog, f, nil, assign.Expression, visited); len(elems) > 0 {
										return elems
									}
								}
							}
						}
					}
				}
			}
		}

		// 1. Check local assignments in enclosing functions
		for i := len(fnStack) - 1; i >= 0; i-- {
			if fn := fnStack[i]; fn != nil {
				for _, assign := range fn.Assignments {
					for _, v := range assign.Variables {
						if v.Name == ref.Reference.Name && assign.Expression != nil {
							if elems := extractElementsWithVisited(prog, file, fnStack[:i+1], assign.Expression, visited); len(elems) > 0 {
								return elems
							}
						}
					}
				}
			}
		}
		// 2. Check file-level assignments
		if file != nil {
			for _, assign := range file.Assignments {
				for _, v := range assign.Variables {
					if v.Name == ref.Reference.Name && assign.Expression != nil {
						if elems := extractElementsWithVisited(prog, file, nil, assign.Expression, visited); len(elems) > 0 {
							return elems
						}
					}
				}
			}
		}
	}

	if len(fnStack) == 0 && op.Type != nil && op.Type.Size > 1 {
		elements := make([]*program.Operand, op.Type.Size)
		for i := 0; i < int(op.Type.Size); i++ {
			elements[i] = makeIndexAccess(op, i, op.Start, op.End)
		}
		return elements
	}

	if op.Type != nil && !isHardwareMatrix(op.Type) && !op.Type.IsList() && !isParameter(fnStack, op) {
		return []*program.Operand{op}
	}

	return nil
}

func searchFnTypes(fnTypes []*program.FnType, targetName, name string) (int, bool) {
	for _, ft := range fnTypes {
		if ft != nil && (ft.Name == targetName || ft.Name == name) {
			for _, g := range ft.Groups {
				if len(g.Functions) > 0 {
					return len(g.Functions[0].Parameters), true
				}
			}
		}
		if ft != nil {
			for _, g := range ft.Groups {
				for _, fn := range g.Functions {
					if len(fn.FnTypes) > 0 {
						if arity, ok := searchFnTypes(fn.FnTypes, targetName, name); ok {
							return arity, true
						}
					}
				}
			}
		}
	}
	return 0, false
}

// resolveFnArity determines the reference prefix, function name, and parameter count (arity)
// for the callable operand passed to higher-order operations like map or reduce.
//
// When lowering map/reduce expressions into concrete unrolled function calls, the lowerer needs
// to know how many parameters the target callback expects (e.g., 1-arg value vs 2-arg index+value
// in map, or 2-arg accumulator+value vs 3-arg accumulator+value+index in reduce).
//
// It performs a hierarchical lexical search across:
// 1. Enclosing functions on the stack (innermost lexical scope first)
// 2. Current file function types
// 3. Package-level function types
// 4. Dependency packages matching the target prefix or package path
// 5. Built-in functions
func resolveFnArity(
	prog *program.Program,
	pkg *program.Package,
	file *program.PackageFile,
	fnStack []*program.Function,
	fnOp *program.Operand,
) (string, string, int) {
	if fnOp == nil {
		return "", "", 1
	}

	var prefix, name string
	if ref, ok := fnOp.T.(*program.Operand_Reference); ok && ref.Reference != nil {
		prefix = ref.Reference.Prefix
		name = ref.Reference.Name
	} else if fc, ok := fnOp.T.(*program.Operand_FnCall); ok && fc.FnCall != nil && fc.FnCall.Reference != nil {
		prefix = fc.FnCall.Reference.Prefix
		name = fc.FnCall.Reference.Name
	}

	targetName := name
	targetPrefix := prefix
	// Inlined anonymous functions are hoisted directly into the target package file
	// with a "gsl_inlined_" prefix. Clear targetPrefix so they resolve locally in the
	// target package rather than attempting cross-package path resolution.
	if strings.HasPrefix(name, program.InlinedFnPrefix) {
		targetName = name
		targetPrefix = ""
	} else if fnOp.Type != nil && fnOp.Type.FnPath != "" {
		targetName = fnOp.Type.FnPath
		if idx := strings.LastIndex(targetName, "/"); idx >= 0 {
			targetPrefix = targetName[:idx]
			targetName = targetName[idx+1:]
		}
		if fnOp.Type.PkgPath != "" {
			targetPrefix = fnOp.Type.PkgPath
		}
	}

	// 1. Check enclosing functions on stack (innermost first)
	for i := len(fnStack) - 1; i >= 0; i-- {
		if fn := fnStack[i]; fn != nil {
			if arity, ok := searchFnTypes(fn.FnTypes, targetName, name); ok {
				return prefix, name, arity
			}
		}
	}

	// 2. Check current file (top-level and nested functions)
	if file != nil {
		if arity, ok := searchFnTypes(file.FnTypes, targetName, name); ok {
			return prefix, name, arity
		}
	}

	// 3. Check package files
	if pkg != nil {
		for _, f := range pkg.Files {
			if arity, ok := searchFnTypes(f.FnTypes, targetName, name); ok {
				return prefix, name, arity
			}
		}
	}

	// 4. Check program dependencies
	if prog != nil {
		for _, dep := range prog.Dependencies {
			if dep == nil {
				continue
			}
			simpleName := dep.Path[strings.LastIndex(dep.Path, "/")+1:]
			if targetPrefix == "" || dep.Path == targetPrefix || simpleName == targetPrefix || strings.HasSuffix(dep.Path, "/"+targetPrefix) {
				for _, f := range dep.Files {
					if arity, ok := searchFnTypes(f.FnTypes, targetName, name); ok {
						return prefix, name, arity
					}
				}
			}
		}
	}

	// 5. Check builtins
	if builtin := program.BuiltinNamed(targetName); builtin != nil && builtin.MinArgs > 0 {
		return prefix, name, builtin.MinArgs
	}
	if builtin := program.BuiltinNamed(name); builtin != nil && builtin.MinArgs > 0 {
		return prefix, name, builtin.MinArgs
	}

	return prefix, name, 1
}

func isFnOperand(prog *program.Program, pkg *program.Package, file *program.PackageFile, fnStack []*program.Function, op *program.Operand) bool {
	if op == nil {
		return false
	}
	if op.Type != nil && op.Type.FnPath != "" {
		return true
	}
	if ref, ok := op.T.(*program.Operand_Reference); ok && ref.Reference != nil {
		name := ref.Reference.Name
		if strings.HasPrefix(name, program.AnonFnPrefix) {
			return true
		}
		for i := len(fnStack) - 1; i >= 0; i-- {
			if fn := fnStack[i]; fn != nil {
				if _, ok := searchFnTypes(fn.FnTypes, name, name); ok {
					return true
				}
			}
		}
		if file != nil {
			if _, ok := searchFnTypes(file.FnTypes, name, name); ok {
				return true
			}
		}
		if pkg != nil {
			for _, f := range pkg.Files {
				if _, ok := searchFnTypes(f.FnTypes, name, name); ok {
					return true
				}
			}
		}
		if prog != nil {
			for _, dep := range prog.Dependencies {
				for _, f := range dep.Files {
					if _, ok := searchFnTypes(f.FnTypes, name, name); ok {
						return true
					}
				}
			}
		}
		if program.BuiltinNamed(name) != nil {
			return true
		}
	}
	return false
}

// UnrollReduce traverses prog and unrolls any remaining Operation_REDUCE expressions
// into sequential function calls.
func UnrollReduce(prog *program.Program) error {
	if prog == nil {
		return nil
	}
	unrollPkg := func(pkg *program.Package) error {
		for _, file := range pkg.Files {
			if file != nil {
				if err := walkFile(file, func(op *program.Operand, fnStack []*program.Function) (*program.Operand, error) {
					if op == nil {
						return nil, nil
					}
					expr := op.GetExpression()
					if expr != nil && expr.Operation == program.Operation_REDUCE {
						return unrollReduceOp(prog, pkg, file, fnStack, op, expr)
					}
					return op, nil
				}); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if prog.Primary != nil {
		if err := unrollPkg(prog.Primary); err != nil {
			return err
		}
	}
	for _, dep := range prog.Dependencies {
		if dep != nil {
			if err := unrollPkg(dep); err != nil {
				return err
			}
		}
	}
	return nil
}

func unrollReduceOp(prog *program.Program, pkg *program.Package, file *program.PackageFile, fnStack []*program.Function, op *program.Operand, expr *program.Expression) (*program.Operand, error) {
	if len(expr.Operands) < 2 {
		return nil, fmt.Errorf("reduce expects at least 2 arguments (a function and a collection), got %d", len(expr.Operands))
	}

	var fnOp *program.Operand
	var initialOp *program.Operand
	var collOps []*program.Operand

	if len(expr.Operands) == 2 {
		fnOp = expr.Operands[0]
		collOps = expr.Operands[1:]
	} else if len(expr.Operands) == 3 {
		if isFnOperand(prog, pkg, file, fnStack, expr.Operands[0]) {
			fnOp = expr.Operands[0]
			collOps = []*program.Operand{expr.Operands[1]}
			initialOp = expr.Operands[2]
		} else {
			initialOp = expr.Operands[0]
			fnOp = expr.Operands[1]
			collOps = []*program.Operand{expr.Operands[2]}
		}
	} else {
		if isFnOperand(prog, pkg, file, fnStack, expr.Operands[0]) {
			fnOp = expr.Operands[0]
			collOps = expr.Operands[1:]
		} else {
			initialOp = expr.Operands[0]
			fnOp = expr.Operands[1]
			collOps = expr.Operands[2:]
		}
	}

	var literal []*program.Operand
	if len(collOps) == 1 {
		literal = literalListOf(file, fnStack, collOps[0])
	}

	prefix, name, arity := resolveFnArity(prog, pkg, file, fnStack, fnOp)
	if name == "" {
		return nil, fmt.Errorf("reduce callback argument must be a function")
	}

	var elements []*program.Operand
	if literal != nil {
		elements = literal
	} else if len(collOps) == 1 {
		elements = extractElements(prog, file, fnStack, collOps[0])
	} else {
		for _, c := range collOps {
			els := extractElements(prog, file, fnStack, c)
			if len(els) > 0 {
				elements = append(elements, els...)
			} else {
				elements = append(elements, c)
			}
		}
	}

	if len(elements) == 0 {
		return op, nil
	}

	return unrollElements(elements, initialOp, prefix, name, arity, op.Start, op.End)
}
