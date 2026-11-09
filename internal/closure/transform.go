package closure

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/program"
)

func (t *transformer) transformFile(file *program.PackageFile) error {
	env := make(map[string]*closureVal)
	var newAssignments []*program.Assignment

	for _, assign := range file.Assignments {
		handled, err := t.processAssignment(file, nil, assign, env)
		if err != nil {
			return err
		}
		if !handled {
			if err := t.rewriteOperand(file, nil, assign.Expression, env); err != nil {
				return err
			}
			if err := t.rewriteProperties(file, nil, assign, env); err != nil {
				return err
			}
			newAssignments = append(newAssignments, assign)
		}
	}
	file.Assignments = newAssignments

	for _, ret := range file.Returns {
		if err := t.rewriteOperand(file, nil, ret, env); err != nil {
			return err
		}
	}

	for _, ft := range file.FnTypes {
		for _, fg := range ft.Groups {
			for _, fn := range fg.Functions {
				if !t.isHigherOrderTemplate(file, fn) {
					if err := t.transformFunction(file, fn, env); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (t *transformer) transformFunction(file *program.PackageFile, fn *program.Function, parentEnv map[string]*closureVal) error {
	if t.processedFns[fn] {
		return nil
	}
	t.processedFns[fn] = true

	// The lexical chain, so a nested function can be told that a bare name its
	// parent binds is a value rather than some package's function of that name.
	t.fnStack = append(t.fnStack, fn)
	defer func() { t.fnStack = t.fnStack[:len(t.fnStack)-1] }()

	fnEnv := make(map[string]*closureVal)
	for k, v := range parentEnv {
		fnEnv[k] = v
	}

	var newAssignments []*program.Assignment
	for _, assign := range fn.Assignments {
		handled, err := t.processAssignment(file, fn, assign, fnEnv)
		if err != nil {
			return err
		}
		if !handled {
			if err := t.rewriteOperand(file, fn, assign.Expression, fnEnv); err != nil {
				return err
			}
			if err := t.rewriteProperties(file, fn, assign, fnEnv); err != nil {
				return err
			}
			newAssignments = append(newAssignments, assign)
		}
	}
	fn.Assignments = newAssignments

	for _, ret := range fn.Returns {
		if err := t.rewriteOperand(file, fn, ret, fnEnv); err != nil {
			return err
		}
	}

	for _, nft := range fn.FnTypes {
		for _, nfg := range nft.Groups {
			for _, nfn := range nfg.Functions {
				if err := t.transformFunction(file, nfn, fnEnv); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (t *transformer) processAssignment(file *program.PackageFile, enclosingFn *program.Function, assign *program.Assignment, env map[string]*closureVal) (bool, error) {
	if assign.Expression == nil || len(assign.Variables) == 0 {
		return false, nil
	}

	// A list holding a function is not a value — it names no runtime object,
	// and the factories in it are about to be pruned. It is recorded under the
	// names it was assigned to and the assignment is dropped, the same way an
	// assignment of a closure is, so that a fold reading the name later finds
	// the elements themselves.
	if els := t.functionListElements(file, enclosingFn, assign.Expression); els != nil {
		byName := t.listBindings[enclosingFn]
		if byName == nil {
			byName = make(map[string][]*program.Operand)
			t.listBindings[enclosingFn] = byName
		}
		for _, v := range assign.Variables {
			byName[v.Name] = els
		}
		return true, nil
	}

	if ref, ok := assign.Expression.T.(*program.Operand_Reference); ok && ref.Reference != nil {
		refPrefix := ref.Reference.Prefix
		refName := ref.Reference.Name
		if existing, exists := env[refName]; exists {
			for _, v := range assign.Variables {
				env[v.Name] = existing
			}
			return true, nil
		}
		targetFile, targetFt := t.lookupFnTypeInScope(enclosingFn, file, refPrefix, refName)
		if targetFt != nil && len(targetFt.Groups) > 0 && len(targetFt.Groups[0].Functions) > 0 {
			targetFn := targetFt.Groups[0].Functions[0]
			cv := &closureVal{
				targetFnName:  refName,
				targetFile:    targetFile,
				targetFn:      targetFn,
				enclosingName: refName,
			}
			for _, v := range assign.Variables {
				env[v.Name] = cv
			}
			return true, nil
		}
	}

	fnCall, ok := assign.Expression.T.(*program.Operand_FnCall)
	if !ok || fnCall.FnCall == nil || fnCall.FnCall.Reference == nil {
		return false, nil
	}

	fc := fnCall.FnCall
	calleeName := fc.Reference.Name
	calleePrefix := fc.Reference.Prefix

	if existing, exists := env[calleeName]; exists {
		// A call supplying every parameter of the closure's target is a FULL
		// application: it yields a value, not another closure. It has to stay
		// an assignment and be rewritten into a direct call to the function
		// already synthesized for this closure, which is what rewriteOperand
		// does for any env name. Routing it through instantiateClosure instead
		// re-specializes the original template with only these arguments
		// bound, dropping the captures that synthesizeFunction had baked in —
		// so the body was left referring to a capture that no longer existed
		// ("undefined identifier"), or, with nothing captured, the assignment
		// was dropped and the name bound to a function rather than a value.
		// Partial application still belongs to instantiateClosure, and so does
		// a target that returns a function, since that level is a closure too.
		if existing.targetFn != nil && !t.functionReturnsFunction(existing.targetFn) &&
			len(fc.Arguments) == len(remainingParams(existing)) {
			return false, nil
		}
		res, err := t.instantiateClosure(file, existing, fc.Arguments, env)
		if err != nil {
			return false, err
		}
		if res != nil {
			for _, v := range assign.Variables {
				env[v.Name] = res
			}
			return true, nil
		}
		return false, nil
	}

	targetFile, ft := t.lookupFnTypeInScope(enclosingFn, file, calleePrefix, calleeName)
	if ft == nil || len(ft.Groups) == 0 {
		return false, nil
	}

	var calleeFn *program.Function
	for _, fg := range ft.Groups {
		if fg.ParamCount == int32(len(fc.Arguments)) {
			if len(fg.Functions) > 0 {
				calleeFn = fg.Functions[0]
				break
			}
		}
	}
	if calleeFn == nil && len(ft.Groups[0].Functions) > 0 {
		calleeFn = ft.Groups[0].Functions[0]
	}
	if calleeFn == nil {
		return false, nil
	}

	if !t.functionReturnsFunction(calleeFn) {
		return false, nil
	}

	retRef := calleeFn.Returns[0].T.(*program.Operand_Reference).Reference
	retTargetFile, retFt := t.lookupFnTypeInScope(calleeFn, targetFile, retRef.Prefix, retRef.Name)
	if retFt == nil || len(retFt.Groups) == 0 || len(retFt.Groups[0].Functions) == 0 {
		return false, fmt.Errorf("cannot resolve returned function %q", retRef.Name)
	}

	retFn := retFt.Groups[0].Functions[0]

	var captures []captureBinding
	for i, param := range calleeFn.Parameters {
		if i < len(fc.Arguments) {
			argVal := fc.Arguments[i]
			if argRef, ok := argVal.T.(*program.Operand_Reference); ok && argRef.Reference != nil {
				if cv, ok := env[argRef.Reference.Name]; ok {
					argVal = program.NewReference("", cv.targetFnName, argVal.Start, argVal.End)
				}
			}
			captures = append(captures, captureBinding{
				name:  param.Name,
				value: proto.Clone(argVal).(*program.Operand),
			})
		}
	}

	cv := &closureVal{
		targetFnName:   retRef.Name,
		targetFile:     retTargetFile,
		targetFn:       retFn,
		enclosingName:  calleeName,
		captures:       captures,
		isNestedReturn: t.functionReturnsFunction(retFn),
		callerPrefix:   calleePrefix,
	}

	if cv.isNestedReturn {
		for _, v := range assign.Variables {
			env[v.Name] = cv
		}
		return true, nil
	}

	lifted, err := t.liftIfForeign(file, cv, env)
	if err != nil {
		return false, err
	}
	if lifted != nil {
		for _, v := range assign.Variables {
			env[v.Name] = lifted
		}
		return true, nil
	}

	synthesizedName, err := t.synthesizeFunction(file, cv, env)
	if err != nil {
		return false, err
	}

	// The captures ride along even though synthesizeFunction has already baked
	// them into synthesizedName. A FULL application of this closure never reads
	// them — it becomes a direct call to that function. A PARTIAL one does: it
	// re-specializes targetFn, which is still the original template, so without
	// the captures here its body would be left referring to them.
	finalCv := &closureVal{
		targetFnName:  synthesizedName,
		targetFile:    file,
		targetFn:      retFn,
		enclosingName: calleeName,
		captures:      captures,
	}
	for _, v := range assign.Variables {
		env[v.Name] = finalCv
	}

	return true, nil
}

func (t *transformer) instantiateClosure(file *program.PackageFile, parentCv *closureVal, args []*program.Operand, env map[string]*closureVal) (*closureVal, error) {
	if parentCv.lifted {
		return t.applyLifted(file, parentCv, args, env)
	}
	targetFn := parentCv.targetFn
	var newCaptures []captureBinding
	newCaptures = append(newCaptures, parentCv.captures...)

	for i, param := range targetFn.Parameters {
		if i < len(args) {
			argVal := args[i]
			if argRef, ok := argVal.T.(*program.Operand_Reference); ok && argRef.Reference != nil {
				if cv, ok := env[argRef.Reference.Name]; ok {
					argVal = program.NewReference("", cv.targetFnName, argVal.Start, argVal.End)
				}
			}
			found := false
			for ci, c := range newCaptures {
				if c.name == param.Name {
					newCaptures[ci] = captureBinding{
						name:  param.Name,
						value: proto.Clone(argVal).(*program.Operand),
					}
					found = true
					break
				}
			}
			if !found {
				newCaptures = append(newCaptures, captureBinding{
					name:  param.Name,
					value: proto.Clone(argVal).(*program.Operand),
				})
			}
		}
	}

	if !t.functionReturnsFunction(targetFn) {
		cv := &closureVal{
			targetFnName:   parentCv.targetFnName,
			targetFile:     parentCv.targetFile,
			targetFn:       targetFn,
			enclosingName:  parentCv.enclosingName,
			captures:       newCaptures,
			isNestedReturn: false,
			callerPrefix:   parentCv.callerPrefix,
		}
		synthFile := cv.targetFile
		if synthFile == nil {
			synthFile = file
		}
		synthesizedName, err := t.synthesizeFunction(synthFile, cv, env)
		if err != nil {
			return nil, err
		}
		return &closureVal{
			targetFnName:  synthesizedName,
			targetFile:    synthFile,
			targetFn:      targetFn,
			enclosingName: parentCv.enclosingName,
			captures:      newCaptures,
		}, nil
	}

	retRef := targetFn.Returns[0].T.(*program.Operand_Reference).Reference
	retTargetFile, retFt := t.lookupFnTypeInScope(targetFn, parentCv.targetFile, retRef.Prefix, retRef.Name)
	if retFt == nil || len(retFt.Groups) == 0 || len(retFt.Groups[0].Functions) == 0 {
		return nil, fmt.Errorf("cannot resolve returned function %q", retRef.Name)
	}
	nextFn := retFt.Groups[0].Functions[0]

	cv := &closureVal{
		targetFnName:   retRef.Name,
		targetFile:     retTargetFile,
		targetFn:       nextFn,
		enclosingName:  parentCv.enclosingName,
		captures:       newCaptures,
		isNestedReturn: t.functionReturnsFunction(nextFn),
		callerPrefix:   parentCv.callerPrefix,
	}

	if cv.isNestedReturn {
		return cv, nil
	}

	lifted, err := t.liftIfForeign(file, cv, env)
	if err != nil {
		return nil, err
	}
	if lifted != nil {
		return lifted, nil
	}

	synthFile := cv.targetFile
	if synthFile == nil {
		synthFile = file
	}
	synthesizedName, err := t.synthesizeFunction(synthFile, cv, env)
	if err != nil {
		return nil, err
	}
	return &closureVal{
		targetFnName:  synthesizedName,
		targetFile:    synthFile,
		targetFn:      nextFn,
		enclosingName: parentCv.enclosingName,
		captures:      newCaptures,
	}, nil
}

// remainingParams returns the parameters of cv's target that the closure has
// not already bound as a capture. Those are the parameters the function
// synthesized for cv actually takes, and therefore the arity a call must supply
// to be fully applied. A factory's captures are normally disjoint from its
// target's parameters, so this is usually just the target's own list; it
// differs once a closure has been PARTIALLY applied, because the arguments
// supplied then are recorded as captures of parameters the target still
// declares.
func remainingParams(cv *closureVal) []string {
	if cv == nil || cv.targetFn == nil {
		return nil
	}
	captured := make(map[string]bool, len(cv.captures))
	for _, c := range cv.captures {
		captured[c.name] = true
	}
	out := make([]string, 0, len(cv.targetFn.Parameters))
	for _, prm := range cv.targetFn.Parameters {
		if !captured[prm.Name] {
			out = append(out, prm.Name)
		}
	}
	return out
}

func (t *transformer) synthesizeFunction(file *program.PackageFile, cv *closureVal, env map[string]*closureVal) (string, error) {
	name := t.nextClosureName(cv.enclosingName)
	clonedFn := proto.Clone(cv.targetFn).(*program.Function)
	clonedFn.FnTypes = nil

	var prependedAssignments []*program.Assignment
	seenNames := make(map[string]bool)
	for i := len(cv.captures) - 1; i >= 0; i-- {
		cap := cv.captures[i]
		if seenNames[cap.name] {
			continue
		}
		seenNames[cap.name] = true
		assign := program.NewAssignment([]string{cap.name}, proto.Clone(cap.value).(*program.Operand), nil, nil)
		prependedAssignments = append([]*program.Assignment{assign}, prependedAssignments...)
	}
	clonedFn.Assignments = append(prependedAssignments, clonedFn.Assignments...)

	// Drop the parameters the captures have bound: they are supplied by the
	// prepended assignments above, so leaving them declared would keep the
	// synthesized function at the template's arity and a later call would not
	// match it.
	paramNames := remainingParams(cv)
	keptParams := make([]*program.Parameter, 0, len(clonedFn.Parameters))
	for _, prm := range clonedFn.Parameters {
		if !seenNames[prm.Name] {
			keptParams = append(keptParams, prm)
		}
	}
	clonedFn.Parameters = keptParams

	file.FnTypes, _ = program.AddFunction(file.FnTypes, name, paramNames)
	for _, ft := range file.FnTypes {
		if ft.Name == name {
			for _, fg := range ft.Groups {
				if fg.ParamCount == int32(len(paramNames)) {
					fg.Functions = []*program.Function{clonedFn}
					break
				}
			}
			if fileMap, ok := t.fileFunctions[file]; ok {
				fileMap[name] = ft
			}
			if pkgMap, ok := t.pkgFunctions[""]; ok {
				pkgMap[name] = ft
			}
			break
		}
	}

	if env != nil {
		for _, assign := range clonedFn.Assignments {
			if err := t.rewriteOperand(file, clonedFn, assign.Expression, env); err != nil {
				return "", err
			}
		}
		for _, ret := range clonedFn.Returns {
			if err := t.rewriteOperand(file, clonedFn, ret, env); err != nil {
				return "", err
			}
		}
	}

	return name, nil
}

// applyLifted is instantiateClosure for a lifted closure. A partial
// application binds the leading unbound parameters of the SAME lifted
// function, so it synthesizes nothing: the result is the parent with more
// captures, which reach the function as more leading arguments at the call
// that eventually completes it.
func (t *transformer) applyLifted(file *program.PackageFile, parentCv *closureVal, args []*program.Operand, env map[string]*closureVal) (*closureVal, error) {
	remaining := remainingParams(parentCv)
	newCaptures := append([]captureBinding(nil), parentCv.captures...)
	for i, arg := range args {
		if i >= len(remaining) {
			break
		}
		if !t.isValueOperand(file, env, arg) {
			return nil, fmt.Errorf("cannot pass a function to %s: it is a closure from another package, and only values cross that boundary", parentCv.enclosingName)
		}
		newCaptures = append(newCaptures, captureBinding{
			name:  remaining[i],
			value: proto.Clone(arg).(*program.Operand),
		})
	}
	cv := *parentCv
	cv.captures = newCaptures
	return &cv, nil
}

// liftIfForeign handles a closure whose target lives in a different package
// than the file creating it. synthesizeFunction cannot: it clones the target's
// body into the creating file, where the body's unqualified references to its
// own package no longer resolve, and cloning into the target's file instead
// strands the captures, which are expressions from the creating file's scope.
// The body wants the definer's scope and the captures want the caller's, so
// the two are kept apart. The body stays in its own file with the captures as
// leading PARAMETERS, and every call through the closure passes the capture
// values as leading arguments — in the caller, where they resolve.
//
// One lifted function then serves every instantiation of the factory, so two
// closures from it cannot collide on a capture: they differ in their
// arguments, not in their function.
//
// It returns nil when cv is not a case for lifting, and the caller bakes it
// as before: a target in the creating file's own package, which keeps the
// output of every existing program unchanged; a captured function, which
// still needs specialization rather than an argument; or a capture that
// shadows another capture or a target parameter, where baking's
// last-binding-wins semantics apply and a parameter list could not express
// them.
func (t *transformer) liftIfForeign(file *program.PackageFile, cv *closureVal, env map[string]*closureVal) (*closureVal, error) {
	if cv == nil || cv.targetFn == nil || cv.targetFile == nil || cv.targetFile == file {
		return nil, nil
	}
	if t.filePkgs[cv.targetFile] == nil || t.filePkgs[cv.targetFile] == t.filePkgs[file] {
		return nil, nil
	}
	return t.lift(file, cv, env)
}

// liftIfEnclosing lifts cv when a capture reads a parameter or local of
// enclosingFn, the function whose call hands the closure on. Baked in as an
// assignment, that name would be undefined in the synthesized function, which
// lives at the top of its file, in whatever package. Lifted, the capture is
// spliced into the specialized body at the call through it, and the
// specialization passes the enclosing name in (see enclosingNamesForList).
func (t *transformer) liftIfEnclosing(file *program.PackageFile, enclosingFn *program.Function, cv *closureVal, env map[string]*closureVal) (*closureVal, error) {
	if enclosingFn == nil || cv == nil || cv.targetFn == nil || cv.targetFile == nil {
		return nil, nil
	}
	if !t.capturesReadEnclosing(file, enclosingFn, cv.captures) {
		return nil, nil
	}
	return t.lift(file, cv, env)
}

// lift synthesizes cv's target with its captures as leading parameters, when
// the captures are plain values that can be passed that way.
func (t *transformer) lift(file *program.PackageFile, cv *closureVal, env map[string]*closureVal) (*closureVal, error) {
	if t.functionReturnsFunction(cv.targetFn) {
		return nil, nil
	}
	seen := make(map[string]bool, len(cv.captures))
	for _, c := range cv.captures {
		if seen[c.name] || !t.isValueOperand(file, env, c.value) {
			return nil, nil
		}
		seen[c.name] = true
	}
	for _, prm := range cv.targetFn.Parameters {
		if seen[prm.Name] {
			return nil, nil
		}
	}

	name, err := t.liftFunction(cv.targetFile, cv)
	if err != nil {
		return nil, err
	}
	return &closureVal{
		targetFnName:   name,
		targetFnPrefix: cv.callerPrefix,
		targetFile:     cv.targetFile,
		targetFn:       cv.targetFn,
		enclosingName:  cv.enclosingName,
		captures:       cv.captures,
		callerPrefix:   cv.callerPrefix,
		lifted:         true,
	}, nil
}

// liftFunction synthesizes cv's target into definerFile, the file that
// declares it, as a first-order function whose leading parameters are cv's
// captures in order, followed by the target's own. Nothing is prepended to
// the body and nothing in it is renamed: it is the definer's code, in the
// definer's file, and it resolves there. The captures reach it as arguments,
// which is also why the function is shared: two instantiations of a factory
// differ only in those arguments.
func (t *transformer) liftFunction(definerFile *program.PackageFile, cv *closureVal) (string, error) {
	paramNames := make([]string, 0, len(cv.captures)+len(cv.targetFn.Parameters))
	params := make([]*program.Parameter, 0, len(cv.captures)+len(cv.targetFn.Parameters))
	for _, c := range cv.captures {
		paramNames = append(paramNames, c.name)
		params = append(params, &program.Parameter{Name: c.name})
	}
	key := liftKey{target: cv.targetFn, captures: strings.Join(paramNames, ",")}
	if name, ok := t.liftedFns[key]; ok {
		return name, nil
	}

	name := t.nextClosureName(cv.enclosingName)
	t.liftedFns[key] = name
	clonedFn := proto.Clone(cv.targetFn).(*program.Function)
	clonedFn.FnTypes = nil
	for _, prm := range clonedFn.Parameters {
		paramNames = append(paramNames, prm.Name)
		params = append(params, prm)
	}
	clonedFn.Parameters = params

	definerFile.FnTypes, _ = program.AddFunction(definerFile.FnTypes, name, paramNames)
	for _, ft := range definerFile.FnTypes {
		if ft.Name != name {
			continue
		}
		for _, fg := range ft.Groups {
			if fg.ParamCount == int32(len(paramNames)) {
				fg.Functions = []*program.Function{clonedFn}
				break
			}
		}
		if fileMap, ok := t.fileFunctions[definerFile]; ok {
			fileMap[name] = ft
		}
		if pkg := t.filePkgs[definerFile]; pkg != nil {
			if pkgMap, ok := t.pkgFunctions[pkg.Path]; ok {
				pkgMap[name] = ft
			}
		}
		break
	}

	// The body may create closures or call higher-order functions of its own.
	// Those are handled as for any first-order function, in its own file.
	if err := t.transformFunction(definerFile, clonedFn, nil); err != nil {
		return "", err
	}
	return name, nil
}

// isValueOperand reports whether op denotes a value rather than a function.
// A closure name from env, a reference that resolves to a function, an
// anonymous function, a factory call, and a partial application all denote
// functions. Expressions, literals, file-level variables, and calls that
// supply every parameter of a value-returning function are values. A call
// whose callee the pass cannot see is taken to be a value: it is a builtin,
// or a name the linker rejects regardless.
func (t *transformer) isValueOperand(file *program.PackageFile, env map[string]*closureVal, op *program.Operand) bool {
	if op == nil {
		return true
	}
	switch v := op.T.(type) {
	case *program.Operand_Reference:
		ref := v.Reference
		if ref == nil {
			return true
		}
		if ref.Prefix == "" {
			if _, ok := env[ref.Name]; ok {
				return false
			}
			// A parameter or local of a function the descent is inside is a
			// value, whatever any linked package happens to call a function of
			// the same name. Asked before lookupFnType below, whose
			// last-resort scan reaches every dependency in the program and so
			// would answer for a package this file never imported.
			if t.bindsValueName(nil, ref.Name) {
				return true
			}
			if file != nil {
				for _, a := range file.Assignments {
					for _, variable := range a.Variables {
						if variable.Name == ref.Name {
							return true
						}
					}
				}
			}
		}
		if strings.HasPrefix(ref.Name, program.AnonFnPrefix) {
			return false
		}
		_, ft := t.lookupFnType(file, ref.Prefix, ref.Name)
		return ft == nil
	case *program.Operand_FnCall:
		fc := v.FnCall
		if fc == nil || fc.Reference == nil {
			return true
		}
		if fc.Reference.Prefix == "" {
			if cv, ok := env[fc.Reference.Name]; ok {
				return cv.targetFn != nil && !t.functionReturnsFunction(cv.targetFn) &&
					len(fc.Arguments) == len(remainingParams(cv))
			}
		}
		_, ft := t.lookupFnType(file, fc.Reference.Prefix, fc.Reference.Name)
		if ft == nil {
			return true
		}
		for _, fg := range ft.Groups {
			if fg.ParamCount == int32(len(fc.Arguments)) {
				return len(fg.Functions) == 0 || !t.functionReturnsFunction(fg.Functions[0])
			}
		}
		return false
	case *program.Operand_Expression:
		// A list is a value only if every element is. A list holding a function
		// must not be lifted into a leading parameter: the fold that consumes
		// it would then be reading a parameter, which has no one value at
		// compile time, and could never be unrolled.
		if v.Expression != nil && v.Expression.Operation == program.Operation_COLLECT {
			for _, sub := range v.Expression.Operands {
				if !t.isValueOperand(file, env, sub) {
					return false
				}
			}
		}
		return true
	}
	return true
}

func (t *transformer) rewriteOperand(file *program.PackageFile, enclosingFn *program.Function, op *program.Operand, env map[string]*closureVal) error {
	if op == nil {
		return nil
	}
	switch val := op.T.(type) {
	case *program.Operand_FnCall:
		fc := val.FnCall
		if fc == nil || fc.Reference == nil {
			return nil
		}

		// A lifted closure's captures are the leading parameters of its
		// function, so a call through it supplies their values first. They are
		// cloned per call site, since each site's arguments are rewritten in
		// place below.
		if cv, ok := env[fc.Reference.Name]; ok && cv.lifted {
			leading := make([]*program.Operand, 0, len(cv.captures)+len(fc.Arguments))
			for _, c := range cv.captures {
				leading = append(leading, proto.Clone(c.value).(*program.Operand))
			}
			fc.Arguments = append(leading, fc.Arguments...)
		}

		for _, arg := range fc.Arguments {
			if err := t.rewriteOperand(file, enclosingFn, arg, env); err != nil {
				return err
			}
		}

		calleeName := fc.Reference.Name
		if cv, ok := env[calleeName]; ok {
			fc.Reference.Name = cv.targetFnName
			fc.Reference.Prefix = cv.targetFnPrefix
			fc.Reference.RefPath = ""
			fc.Reference.PkgPath = ""
			return nil
		}

		if err := t.specializeHigherOrderCall(file, enclosingFn, fc, env); err != nil {
			return err
		}

		// Reachability: if this is a first-order call to an unprocessed function, transform it recursively
		calleeFile, calleeFt := t.lookupFnTypeInScope(enclosingFn, file, fc.Reference.Prefix, fc.Reference.Name)
		if calleeFt != nil {
			var calleeFn *program.Function
			for _, fg := range calleeFt.Groups {
				if fg.ParamCount == int32(len(fc.Arguments)) {
					if len(fg.Functions) > 0 {
						calleeFn = fg.Functions[0]
						break
					}
				}
			}
			if calleeFn != nil && !t.processedFns[calleeFn] && !t.isHigherOrderTemplate(calleeFile, calleeFn) {
				if err := t.transformFunction(calleeFile, calleeFn, env); err != nil {
					return err
				}
			}
		}

	case *program.Operand_Expression:
		// A fold whose collection is a list of functions becomes the calls it
		// denotes, here, before the operands below are walked: those calls are
		// what needs specializing, not the map/reduce that named them.
		if val.Expression != nil &&
			(val.Expression.Operation == program.Operation_MAP || val.Expression.Operation == program.Operation_REDUCE) {
			unrolled, err := t.unrollFunctionFold(file, enclosingFn, op, val.Expression, env)
			if err != nil {
				return err
			}
			if unrolled {
				return t.rewriteOperand(file, enclosingFn, op, env)
			}
		}
		for _, subOp := range val.Expression.Operands {
			if err := t.rewriteOperand(file, enclosingFn, subOp, env); err != nil {
				return err
			}
		}
	case *program.Operand_PropRef:
		if val.PropRef != nil && val.PropRef.Reference != nil {
			if cv, ok := env[val.PropRef.Reference.Name]; ok {
				val.PropRef.Reference.Name = cv.targetFnName
				val.PropRef.Reference.Prefix = cv.targetFnPrefix
				val.PropRef.Reference.RefPath = ""
				val.PropRef.Reference.PkgPath = ""
			}
		}
	case *program.Operand_Reference:
		ref := val.Reference
		if ref != nil && ref.Prefix == "" {
			// A bare reference to a lifted closure cannot name a first-order
			// function, since its captures travel as arguments. It is left as
			// it is for specializeHigherOrderCall, which finds the closure in
			// env and specializes the function it is passed to.
			if cv, ok := env[ref.Name]; ok && !cv.lifted {
				ref.Name = cv.targetFnName
				ref.Prefix = cv.targetFnPrefix
				ref.RefPath = ""
				ref.PkgPath = ""
			}
		}
	}
	return nil
}

// rewriteProperties rewrites the expressions of `name.prop = ...` statements,
// which live on the assignment's variables rather than in its expression.
func (t *transformer) rewriteProperties(file *program.PackageFile, fn *program.Function, assign *program.Assignment, env map[string]*closureVal) error {
	for _, v := range assign.Variables {
		for _, prop := range v.Properties {
			if err := t.rewriteOperand(file, fn, prop.Expression, env); err != nil {
				return err
			}
		}
	}
	return nil
}
