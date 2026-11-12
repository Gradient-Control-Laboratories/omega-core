package inline

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/program"
)

// Transform runs a function inlining pass over the entire program AST.
// It inlines single-return functions called with constant arguments and
// multi-return functions at multi-variable assignments within the same package.
// It enforces GSL's no-recursion rule by detecting cycles during inlining traversal.
// It returns true if any inlining modifications were made.
func Transform(prog *program.Program) (bool, error) {
	return transform(prog, nil)
}

// TransformWithParameters is Transform for a program whose runtime
// parameters checker knows, as the constant fold that follows it does. A
// constant number passed as an argument is then bound as its literal (see
// bindArg) rather than copied as the expression the fold would reduce to it.
func TransformWithParameters(prog *program.Program, checker eval.ParameterChecker) (bool, error) {
	return transform(prog, checker)
}

func transform(prog *program.Program, checker eval.ParameterChecker) (bool, error) {
	progScope, err := scope.Init(prog)
	if err != nil {
		return false, err
	}

	inliner := &inliner{
		prog:         prog,
		progScope:    progScope,
		callStack:    make(map[string]bool),
		constRefMemo: make(map[string]bool),
		fnCache:      make(map[string]fnLookup),
		pOrTimeCache: make(map[pOrTimeKey]bool),
	}
	if checker != nil {
		inliner.evaluator = eval.NewEvaluator(prog, progScope, checker)
	}
	inliner.initAnonCounter()

	if err := inliner.transformProgram(); err != nil {
		return false, err
	}
	return inliner.modified, nil
}

func (in *inliner) initAnonCounter() {
	maxVal := 0
	checkName := func(name string) {
		var n int
		// TODO: extract literal "gsl_tmp_" prefix to const
		if strings.HasPrefix(name, "gsl_tmp_") {
			if _, err := fmt.Sscanf(name, "gsl_tmp_%d", &n); err == nil && n > maxVal {
				maxVal = n
			}
		}
	}
	var scanFile func(f *program.PackageFile)
	scanFile = func(f *program.PackageFile) {
		if f == nil {
			return
		}
		for _, a := range f.Assignments {
			for _, v := range a.Variables {
				checkName(v.Name)
			}
		}
		for _, ft := range f.FnTypes {
			checkName(ft.Name)
			for _, grp := range ft.Groups {
				for _, fn := range grp.Functions {
					for _, a := range fn.Assignments {
						for _, v := range a.Variables {
							checkName(v.Name)
						}
					}
					for _, subFt := range fn.FnTypes {
						checkName(subFt.Name)
					}
				}
			}
		}
	}
	if in.prog.Primary != nil {
		for _, f := range in.prog.Primary.Files {
			scanFile(f)
		}
	}
	for _, dep := range in.prog.Dependencies {
		for _, f := range dep.Files {
			scanFile(f)
		}
	}
	in.anonCounter = maxVal
}

type fnLookup struct {
	fnType  *program.FnType
	fn      *program.Function
	pkgPath string
	found   bool
}

type inliner struct {
	prog               *program.Program
	progScope          *scope.ProgScope
	callStack          map[string]bool
	modified           bool
	anonCounter        int
	currentFileScope   *scope.PkgFileScope
	constRefMemo       map[string]bool
	fnCache            map[string]fnLookup
	pOrTimeCache       map[pOrTimeKey]bool
	hoistedAssignments []*program.Assignment
	callCounts         map[string]int
	enclosingFns       []*program.Function
	// remapTargets holds the operands that are a remap's target (see
	// inlinesRemapTarget).
	remapTargets map[*program.Operand]bool
	// evaluator gives the fold's verdict on a constant argument (see
	// bindArg); nil when the parameters are not known.
	evaluator       *eval.Evaluator
	currentPkgScope *scope.PkgScope
	// inUIMarker counts the UI markers whose arguments are being transformed:
	// their literals are bound by operand, and the fold leaves them as written.
	inUIMarker int
}

// deferredHoist is a property's temporaries that read a top-level variable
// declared below the part the property belongs to, with the names they wait
// for.
type deferredHoist struct {
	assignments []*program.Assignment
	pending     map[string]bool
}

// readsUndeclared is the set of names in topNames not yet declared that the
// assignments read, nil when there is none.
func readsUndeclared(assignments []*program.Assignment, topNames []string, declared map[string]bool) map[string]bool {
	var pending map[string]bool
	for _, name := range topNames {
		if declared[name] || pending[name] {
			continue
		}
		for _, a := range assignments {
			if readsName(a.Expression, name) {
				if pending == nil {
					pending = map[string]bool{}
				}
				pending[name] = true
				break
			}
		}
	}
	return pending
}

// readsName reports whether op refers to the unqualified name. The member of
// an access, the x of p.x, is a name of its own and not a read.
func readsName(op *program.Operand, name string) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		return t.Reference.Prefix == "" && t.Reference.Name == name
	case *program.Operand_PropRef:
		return t.PropRef != nil && t.PropRef.Reference != nil && t.PropRef.Reference.Prefix == "" && t.PropRef.Reference.Name == name
	case *program.Operand_Expression:
		for i, child := range t.Expression.Operands {
			if i == 1 && t.Expression.Operation == program.Operation_ACCESS {
				continue
			}
			if readsName(child, name) {
				return true
			}
		}
	case *program.Operand_FnCall:
		for _, arg := range t.FnCall.Arguments {
			if readsName(arg, name) {
				return true
			}
		}
	}
	return false
}

// placeDeferred puts each deferred batch after the last assignment declaring
// a name it waited for, the batches in their order.
func placeDeferred(assignments []*program.Assignment, deferred []deferredHoist) []*program.Assignment {
	for _, d := range deferred {
		at := len(assignments) - 1
		for i, a := range assignments {
			for _, v := range a.Variables {
				if d.pending[v.Name] {
					at = i
				}
			}
		}
		placed := make([]*program.Assignment, 0, len(assignments)+len(d.assignments))
		placed = append(placed, assignments[:at+1]...)
		placed = append(placed, d.assignments...)
		placed = append(placed, assignments[at+1:]...)
		assignments = placed
	}
	return assignments
}

func (in *inliner) transformProgram() error {
	in.buildCallCounts()
	if in.prog.Primary != nil {
		if err := in.transformPackage(in.prog.Primary, in.progScope.Primary); err != nil {
			return err
		}
	}
	for _, dep := range in.prog.Dependencies {
		pkgScope := in.progScope.Dependencies[dep.Path]
		if err := in.transformPackage(dep, pkgScope); err != nil {
			return err
		}
	}
	return nil
}

func (in *inliner) transformPackage(pkg *program.Package, pkgScope *scope.PkgScope) error {
	currentPkgPath := pkg.Path
	in.currentPkgScope = pkgScope
	for _, file := range pkg.Files {
		var fileScope *scope.PkgFileScope
		if pkgScope != nil {
			for _, fs := range pkgScope.Files {
				if fs.File == file {
					fileScope = fs
					break
				}
			}
		}
		in.currentFileScope = fileScope
		in.constRefMemo = make(map[string]bool)

		// Transform functions first so their internal bodies are simplified
		for _, fnType := range file.FnTypes {
			for _, group := range fnType.Groups {
				for _, fn := range group.Functions {
					if err := in.transformFunction(fn, fileScope, file, currentPkgPath); err != nil {
						return err
					}
				}
			}
		}

		// Transform top-level assignments. The file's names in order, and the
		// ones declared so far, place a property's temporaries (below).
		var topNames []string
		for _, a := range file.Assignments {
			for _, v := range a.Variables {
				topNames = append(topNames, v.Name)
			}
		}
		declared := map[string]bool{}
		var deferred []deferredHoist
		var newAssignments []*program.Assignment
		for _, assign := range file.Assignments {
			// Check if this assignment is a multi-return call that can be expanded
			if len(assign.Variables) > 1 && assign.Expression != nil {
				if fc := assign.Expression.GetFnCall(); fc != nil && fc.Reference != nil {
					fnDecl, fn, calleePkgPath := in.findFunction(fc, fileScope, len(fc.Arguments))
					if fn != nil && len(fn.Returns) == len(assign.Variables) && !in.isBuiltinOrSpecial(fc) && in.shouldInlineMultiReturnCall(fc, fn, calleePkgPath, currentPkgPath, fileScope) {
						expanded, err := in.inlineMultiReturnAssignment(assign, fnDecl, fn, calleePkgPath, fileScope, file, currentPkgPath)
						if err != nil {
							return err
						}
						if expanded != nil {
							in.modified = true
							if len(in.hoistedAssignments) > 0 {
								newAssignments = append(newAssignments, in.hoistedAssignments...)
								in.hoistedAssignments = nil
							}
							for _, expAssign := range expanded {
								in.hoistedAssignments = nil
								if err := in.transformOperand(expAssign.Expression, fileScope, file, currentPkgPath); err != nil {
									return err
								}
								if len(in.hoistedAssignments) > 0 {
									newAssignments = append(newAssignments, in.hoistedAssignments...)
									in.hoistedAssignments = nil
								}
								newAssignments = append(newAssignments, expAssign)
								for _, v := range expAssign.Variables {
									declared[v.Name] = true
								}
							}
							continue
						}
					}
				}
			}

			in.hoistedAssignments = nil
			if err := in.transformOperand(assign.Expression, fileScope, file, currentPkgPath); err != nil {
				return err
			}
			if len(in.hoistedAssignments) > 0 {
				newAssignments = append(newAssignments, in.hoistedAssignments...)
				in.hoistedAssignments = nil
			}
			for _, v := range assign.Variables {
				for _, prop := range v.Properties {
					if err := in.transformOperand(prop.Expression, fileScope, file, currentPkgPath); err != nil {
						return err
					}
				}
			}
			// A property is written below the part it belongs to and may read a
			// top-level variable declared between the two, in order as written:
			// a = sphere(5); x = 8; a.color = pattern(x). Its temporaries go in
			// front of the part, as the expression's do, unless one reads a
			// variable not yet declared; then they wait for its declaration.
			if len(in.hoistedAssignments) > 0 {
				if pending := readsUndeclared(in.hoistedAssignments, topNames, declared); len(pending) > 0 {
					deferred = append(deferred, deferredHoist{assignments: in.hoistedAssignments, pending: pending})
				} else {
					newAssignments = append(newAssignments, in.hoistedAssignments...)
				}
				in.hoistedAssignments = nil
			}
			newAssignments = append(newAssignments, assign)
			for _, v := range assign.Variables {
				declared[v.Name] = true
			}
		}
		file.Assignments = placeDeferred(newAssignments, deferred)

		// Transform top-level returns
		for _, ret := range file.Returns {
			in.hoistedAssignments = nil
			if err := in.transformOperand(ret, fileScope, file, currentPkgPath); err != nil {
				return err
			}
			if len(in.hoistedAssignments) > 0 {
				file.Assignments = append(file.Assignments, in.hoistedAssignments...)
				in.hoistedAssignments = nil
			}
		}
		for _, ret := range file.ColorReturns {
			in.hoistedAssignments = nil
			if err := in.transformOperand(ret, fileScope, file, currentPkgPath); err != nil {
				return err
			}
			if len(in.hoistedAssignments) > 0 {
				file.Assignments = append(file.Assignments, in.hoistedAssignments...)
				in.hoistedAssignments = nil
			}
		}
		for _, ret := range file.MaterialReturns {
			in.hoistedAssignments = nil
			if err := in.transformOperand(ret, fileScope, file, currentPkgPath); err != nil {
				return err
			}
			if len(in.hoistedAssignments) > 0 {
				file.Assignments = append(file.Assignments, in.hoistedAssignments...)
				in.hoistedAssignments = nil
			}
		}
	}
	return nil
}

func (in *inliner) transformFunction(fn *program.Function, fileScope *scope.PkgFileScope, targetFile *program.PackageFile, currentPkgPath string) error {
	in.enclosingFns = append(in.enclosingFns, fn)
	defer func() {
		in.enclosingFns = in.enclosingFns[:len(in.enclosingFns)-1]
	}()

	for _, nestedFnType := range fn.FnTypes {
		for _, group := range nestedFnType.Groups {
			for _, nestedFn := range group.Functions {
				if err := in.transformFunction(nestedFn, fileScope, targetFile, currentPkgPath); err != nil {
					return err
				}
			}
		}
	}

	var newAssignments []*program.Assignment
	for _, assign := range fn.Assignments {
		if len(assign.Variables) > 1 && assign.Expression != nil {
			if fc := assign.Expression.GetFnCall(); fc != nil && fc.Reference != nil && fc.Reference.Prefix == "" {
				fnDecl, calleeFn, calleePkgPath := in.findFunction(fc, fileScope, len(fc.Arguments))
				if calleeFn != nil && len(calleeFn.Returns) == len(assign.Variables) && in.isAllConstArgs(fc) && !in.hasPOrTime(calleeFn, fileScope) && !in.isBuiltinOrSpecial(fc) && calleePkgPath == currentPkgPath {
					expanded, err := in.inlineMultiReturnAssignment(assign, fnDecl, calleeFn, calleePkgPath, fileScope, targetFile, currentPkgPath)
					if err != nil {
						return err
					}
					if expanded != nil {
						in.modified = true
						if len(in.hoistedAssignments) > 0 {
							newAssignments = append(newAssignments, in.hoistedAssignments...)
							in.hoistedAssignments = nil
						}
						for _, expAssign := range expanded {
							in.hoistedAssignments = nil
							if err := in.transformOperand(expAssign.Expression, fileScope, targetFile, currentPkgPath); err != nil {
								return err
							}
							if len(in.hoistedAssignments) > 0 {
								newAssignments = append(newAssignments, in.hoistedAssignments...)
								in.hoistedAssignments = nil
							}
							newAssignments = append(newAssignments, expAssign)
						}
						continue
					}
				}
			}
		}

		in.hoistedAssignments = nil
		if err := in.transformOperand(assign.Expression, fileScope, targetFile, currentPkgPath); err != nil {
			return err
		}
		for _, v := range assign.Variables {
			for _, prop := range v.Properties {
				if err := in.transformOperand(prop.Expression, fileScope, targetFile, currentPkgPath); err != nil {
					return err
				}
			}
		}
		if len(in.hoistedAssignments) > 0 {
			newAssignments = append(newAssignments, in.hoistedAssignments...)
			in.hoistedAssignments = nil
		}
		newAssignments = append(newAssignments, assign)
	}
	fn.Assignments = newAssignments

	for _, ret := range fn.Returns {
		in.hoistedAssignments = nil
		if err := in.transformOperand(ret, fileScope, targetFile, currentPkgPath); err != nil {
			return err
		}
		if len(in.hoistedAssignments) > 0 {
			fn.Assignments = append(fn.Assignments, in.hoistedAssignments...)
			in.hoistedAssignments = nil
		}
	}
	return nil
}

func (in *inliner) transformOperand(op *program.Operand, fileScope *scope.PkgFileScope, targetFile *program.PackageFile, currentPkgPath string) error {
	if op == nil {
		return nil
	}

	switch t := op.T.(type) {
	case *program.Operand_Expression:
		if t.Expression.Operation == program.Operation_REMAP && len(t.Expression.Operands) > 0 {
			if in.remapTargets == nil {
				in.remapTargets = map[*program.Operand]bool{}
			}
			in.remapTargets[t.Expression.Operands[0]] = true
		}
		if t.Expression.Operation == program.Operation_SELECT && len(t.Expression.Operands) >= 3 {
			if isConst, condVal := in.evalConstCondition(t.Expression.Operands[0]); isConst {
				if condVal {
					// Condition is true: only transform true branch (operand 1)
					return in.transformOperand(t.Expression.Operands[1], fileScope, targetFile, currentPkgPath)
				}
				// Condition is false: only transform false branch (operand 2)
				if len(t.Expression.Operands) == 3 {
					return in.transformOperand(t.Expression.Operands[2], fileScope, targetFile, currentPkgPath)
				}
			}
		}
		for _, child := range t.Expression.Operands {
			if err := in.transformOperand(child, fileScope, targetFile, currentPkgPath); err != nil {
				return err
			}
		}
	case *program.Operand_PropRef:
		refOp := &program.Operand{T: &program.Operand_Reference{Reference: t.PropRef.Reference}}
		if err := in.transformOperand(refOp, fileScope, targetFile, currentPkgPath); err != nil {
			return err
		}
	case *program.Operand_FnCall:
		fc := t.FnCall
		if isUIMarker(fc) {
			in.inUIMarker++
			defer func() { in.inUIMarker-- }()
		}
		for _, arg := range fc.Arguments {
			if err := in.transformOperand(arg, fileScope, targetFile, currentPkgPath); err != nil {
				return err
			}
		}

		// A function of the builtin package is left as a call, with one
		// exception: one that remaps a parameter (revolve) has to be inlined
		// as any other package's is (see remapsFieldArg), or a backend that
		// emits it as a function returns the field where the caller stood.
		// Whether the argument depends on p is not asked here: operandUsesP
		// does not see a p read inside a callee's map or reduce function,
		// which is where a profile reads it, and inlining a call whose
		// argument is the same at every point changes nothing.
		builtinFn := false
		if in.isBuiltinOrSpecial(fc) {
			if in.isNativeOrSpecial(fc) {
				return nil
			}
			builtinFn = true
		}

		fnDecl, fn, calleePkgPath := in.findFunction(fc, fileScope, len(fc.Arguments))
		if builtinFn {
			fnDecl, fn, calleePkgPath = in.findBuiltinFunction(fc, len(fc.Arguments))
		}
		if fn == nil || fnDecl == nil {
			return nil
		}

		if len(fn.Returns) != 1 {
			return nil
		}

		if builtinFn && !in.remapsAnyParam(fn, calleePkgPath) {
			return nil
		}

		callKey := fmt.Sprintf("%s/%s/%d", calleePkgPath, fnDecl.Name, len(fc.Arguments))
		if in.callStack[callKey] {
			return fmt.Errorf("recursion not allowed in GSL: %q is called recursively", fnDecl.Name)
		}

		if !in.inlinesRemapTarget(op, fc, fileScope) && !in.shouldInlineCall(fc, fn, calleePkgPath, currentPkgPath, fileScope) {
			return nil
		}

		// Perform inlining substitution
		in.callStack[callKey] = true
		defer func() {
			delete(in.callStack, callKey)
		}()

		inlinedOp, err := in.inlineSingleReturn(fc, fn, fileScope, targetFile, currentPkgPath, calleePkgPath)
		if err != nil {
			return err
		}
		if inlinedOp != nil {
			in.modified = true
			// Replace op with inlinedOp
			op.T = inlinedOp.T
			op.ConstValues = inlinedOp.ConstValues
			if op.Type == nil {
				op.Type = inlinedOp.Type
			}
			// Recursively transform in case inlined body has more calls
			return in.transformOperand(op, fileScope, targetFile, currentPkgPath)
		}
	}

	return nil
}

func (in *inliner) buildCallCounts() {
	in.callCounts = make(map[string]int)
	var countInOperand func(op *program.Operand)
	countInOperand = func(op *program.Operand) {
		if op == nil {
			return
		}
		switch t := op.T.(type) {
		case *program.Operand_Expression:
			for _, child := range t.Expression.Operands {
				countInOperand(child)
			}
		case *program.Operand_FnCall:
			if t.FnCall.Reference != nil {
				name := t.FnCall.Reference.Name
				in.callCounts[name]++
				if t.FnCall.Reference.Prefix != "" {
					in.callCounts[t.FnCall.Reference.Prefix+"."+name]++
				}
			}
			for _, arg := range t.FnCall.Arguments {
				countInOperand(arg)
			}
		}
	}

	var countInPackage func(pkg *program.Package)
	countInPackage = func(pkg *program.Package) {
		if pkg == nil {
			return
		}
		for _, file := range pkg.Files {
			if file == nil {
				continue
			}
			for _, a := range file.Assignments {
				countInOperand(a.Expression)
			}
			for _, r := range file.Returns {
				countInOperand(r)
			}
			for _, r := range file.ColorReturns {
				countInOperand(r)
			}
			for _, r := range file.MaterialReturns {
				countInOperand(r)
			}
			for _, ft := range file.FnTypes {
				for _, g := range ft.Groups {
					for _, fn := range g.Functions {
						for _, a := range fn.Assignments {
							countInOperand(a.Expression)
						}
						for _, r := range fn.Returns {
							countInOperand(r)
						}
					}
				}
			}
		}
	}

	if in.prog != nil {
		countInPackage(in.prog.Primary)
		for _, dep := range in.prog.Dependencies {
			countInPackage(dep)
		}
	}
}
