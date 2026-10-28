package closure

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/program"
)

// specializeHigherOrderCall transforms call sites that pass functions or closures
// as arguments to higher-order functions (e.g. combiners, mappers, or builders).
//
// Specifically, this method:
//  1. Resolves the callee function in scope and matches function-valued arguments
//     against its formal parameters. Arguments may be closure values from env,
//     named library functions, nested anonymous lambdas, or closure factory calls.
//  2. Synthesizes intermediate static helper functions for any lambda expressions
//     or factory calls passed as arguments.
//  3. Clones the callee function and monomorphizes it into a concrete first-order
//     function (named gcl_closure_<callee>...):
//     - Strips function-typed parameters from the parameter list.
//     - Replaces internal indirect calls to those parameters with direct calls
//     to the concrete combiner/passed functions.
//     - Preserves package prefixes and captured lexical variables.
//  4. Adds the new specialized function to the target file's AST.
//  5. Rewrites the call site (fc) to invoke the specialized function directly,
//     passing only the remaining non-function arguments.
//  6. Recursively transforms the newly synthesized function to resolve any nested
//     higher-order calls or closures inside its body.
func (t *transformer) specializeHigherOrderCall(file *program.PackageFile, enclosingFn *program.Function, fc *program.FnCall, env map[string]*closureVal) error {
	if fc.Reference == nil {
		return nil
	}

	calleeFile, calleeFt := t.lookupFnTypeInScope(enclosingFn, file, fc.Reference.Prefix, fc.Reference.Name)
	if calleeFt == nil {
		return nil
	}

	var calleeFn *program.Function
	for _, fg := range calleeFt.Groups {
		if fg.ParamCount == int32(len(fc.Arguments)) {
			if len(fg.Functions) > 0 {
				calleeFn = fg.Functions[0]
				break
			}
		}
	}
	if calleeFn == nil {
		return nil
	}

	hasFunctionArg := false
	argFnMap := make(map[string]*closureVal)
	listArgs := make(map[string][]*program.Operand)

	for i, arg := range fc.Arguments {
		if i >= len(calleeFn.Parameters) {
			break
		}
		paramName := calleeFn.Parameters[i].Name

		// A list of functions binds to the parameter the same way a single
		// function does — the parameter is dropped and the clone is given the
		// elements, which a fold inside it unrolls against.
		if els := t.functionListElements(file, enclosingFn, arg); els != nil {
			hasFunctionArg = true
			listArgs[paramName] = els
			continue
		} else if calleeFn != nil {
			scopeFile := calleeFile
			if scopeFile == nil {
				scopeFile = file
			}
			if t.paramIsFunctionList(scopeFile, calleeFn, paramName) {
				// A scalar behaves as a list of length one, so a lone function
				// folds as a list of one. It has to be spelled that way rather
				// than as [f]: a collection literal needs two or more elements.
				if t.isFunctionValue(arg, file, enclosingFn) {
					hasFunctionArg = true
					listArgs[paramName] = cloneOperands([]*program.Operand{arg})
					continue
				}
				// The callee folds this parameter and calls what it gets, so
				// the argument has to be functions this pass can see. Nothing
				// resolved, and nothing later can: there is no dispatch on a
				// function at run time to fall back on.
				return fmt.Errorf("a list of functions has to be visible where it is folded: the %s argument of %s does not resolve to one. Write the list at the call — %s([first(...), second(...)]), or %s(only(...)) for a single one — or bind it to a name in the same scope. The backends have no function pointers, so every element has to be resolved here", paramName, fc.Reference.Name, fc.Reference.Name, fc.Reference.Name)
			}
		}

		if argRef, ok := arg.T.(*program.Operand_Reference); ok && argRef.Reference != nil {
			refName := argRef.Reference.Name
			refPrefix := argRef.Reference.Prefix
			if cv, ok := env[refName]; ok {
				hasFunctionArg = true
				argFnMap[paramName] = cv
			} else {
				isLocalVar := false
				if refPrefix == "" {
					// The whole lexical chain, not just the immediately
					// enclosing function: a nested function reads what its
					// parents bind, and a name bound anywhere up the chain is
					// a value however many packages declare a function of that
					// name.
					isLocalVar = t.bindsValueName(enclosingFn, refName)
					if !isLocalVar && file != nil {
						for _, a := range file.Assignments {
							for _, v := range a.Variables {
								if v.Name == refName {
									isLocalVar = true
									break
								}
							}
						}
					}
				}
				if isLocalVar {
					continue
				}

				_, argFt := t.lookupFnTypeInScope(enclosingFn, file, refPrefix, refName)
				if argFt != nil {
					hasFunctionArg = true
					isNested := false
					if enclosingFn != nil {
						for _, nft := range enclosingFn.FnTypes {
							if nft.Name == refName {
								isNested = true
								break
							}
						}
					}
					if isNested || strings.HasPrefix(refName, program.AnonFnPrefix) {
						var targetFn *program.Function
						expectedParamCount := t.findParamCallArgCount(calleeFn, paramName)
						for _, fg := range argFt.Groups {
							if fg.ParamCount == int32(expectedParamCount) && len(fg.Functions) > 0 {
								targetFn = fg.Functions[0]
								break
							}
						}
						if targetFn == nil && len(argFt.Groups) > 0 && len(argFt.Groups[0].Functions) > 0 {
							targetFn = argFt.Groups[0].Functions[0]
						}

						if targetFn != nil {
							var captures []captureBinding
							readsParam := false
							if enclosingFn != nil {
								captures, readsParam = capturesReadBy(enclosingFn, targetFn, fc)
							}
							enclosingName := refName
							if enclosingFn != nil {
								enclosingName = fc.Reference.Name
							}
							cv := &closureVal{
								targetFnName:  refName,
								targetFile:    file,
								targetFn:      targetFn,
								enclosingName: enclosingName,
								captures:      captures,
							}
							// A lambda reading a parameter of the function it is
							// written in is lifted, and the specialization passes
							// the parameter in. Baked, the parameter would be
							// undefined in the synthesized function.
							if readsParam {
								lifted, err := t.liftIfEnclosing(file, enclosingFn, cv, env)
								if err != nil {
									return err
								}
								if lifted != nil {
									argFnMap[paramName] = lifted
									continue
								}
							}
							synthName, err := t.synthesizeFunction(file, cv, env)
							if err != nil {
								return err
							}
							argFnMap[paramName] = &closureVal{
								targetFnName:  synthName,
								enclosingName: synthName,
							}
							continue
						}
					}
					argFnMap[paramName] = &closureVal{
						targetFnName:   refName,
						targetFnPrefix: refPrefix,
						enclosingName:  refName,
					}
				}
			}
		} else if argFc, ok := arg.T.(*program.Operand_FnCall); ok && argFc.FnCall != nil {
			fnName := argFc.FnCall.Reference.Name
			fnPrefix := argFc.FnCall.Reference.Prefix
			factoryFile, factoryFt := t.lookupFnTypeInScope(enclosingFn, file, fnPrefix, fnName)
			if factoryFt != nil {
				var factoryFn *program.Function
				for _, fg := range factoryFt.Groups {
					if fg.ParamCount == int32(len(argFc.FnCall.Arguments)) {
						if len(fg.Functions) > 0 {
							factoryFn = fg.Functions[0]
							break
						}
					}
				}
				if factoryFn != nil && t.functionReturnsFunction(factoryFn) {
					retRef := factoryFn.Returns[0].T.(*program.Operand_Reference).Reference
					retTargetFile, retFt := t.lookupFnTypeInScope(factoryFn, factoryFile, retRef.Prefix, retRef.Name)
					if retFt != nil && len(retFt.Groups) > 0 && len(retFt.Groups[0].Functions) > 0 {
						retFn := retFt.Groups[0].Functions[0]
						var captures []captureBinding
						for pi, p := range factoryFn.Parameters {
							if pi < len(argFc.FnCall.Arguments) {
								captures = append(captures, captureBinding{
									name:  p.Name,
									value: proto.Clone(argFc.FnCall.Arguments[pi]).(*program.Operand),
								})
							}
						}
						cv := &closureVal{
							targetFnName:  retRef.Name,
							targetFile:    retTargetFile,
							targetFn:      retFn,
							enclosingName: fnName,
							captures:      captures,
							callerPrefix:  fnPrefix,
						}
						// synthesizeFunction clones the body into the factory's
						// own file and bakes the captures in as assignments.
						// Those captures are THIS file's expressions, so that
						// only holds when they are closed literals; a document
						// variable would be left undefined in the library. Lift
						// instead, which sends the captures over as arguments.
						// Literal captures keep baking, so every program that
						// compiles today still produces what it did.
						var lifted *closureVal
						if !capturesAreLiteral(cv.captures) {
							var err error
							lifted, err = t.liftIfForeign(file, cv, env)
							if err != nil {
								return err
							}
							if lifted == nil {
								lifted, err = t.liftIfEnclosing(file, enclosingFn, cv, env)
								if err != nil {
									return err
								}
							}
						}
						hasFunctionArg = true
						if lifted != nil {
							argFnMap[paramName] = lifted
							continue
						}
						synthFile := cv.targetFile
						if synthFile == nil {
							synthFile = file
						}
						synthName, err := t.synthesizeFunction(synthFile, cv, env)
						if err != nil {
							return err
						}
						// synthFile is remembered so that, once the file for
						// the specialized body is known, a call to this
						// function from another file can be given the prefix
						// that reaches it.
						argFnMap[paramName] = &closureVal{
							targetFnName:  synthName,
							targetFile:    synthFile,
							enclosingName: synthName,
							callerPrefix:  fnPrefix,
						}
					}
				}
			}
		}
	}

	if !hasFunctionArg {
		return nil
	}

	specializedName := t.nextClosureName(fc.Reference.Name)
	clonedFn := proto.Clone(calleeFn).(*program.Function)

	isNested := false
	if enclosingFn != nil {
		for _, nft := range enclosingFn.FnTypes {
			if nft.Name == fc.Reference.Name {
				isNested = true
				break
			}
		}
	}

	paramEnv := make(map[string]*closureVal)
	for k, v := range env {
		paramEnv[k] = v
	}
	for paramName, cv := range argFnMap {
		paramEnv[paramName] = cv
	}

	var remainingParams []*program.Parameter
	var remainingArgs []*program.Operand
	for i, p := range clonedFn.Parameters {
		_, isFnParam := argFnMap[p.Name]
		_, isListParam := listArgs[p.Name]
		if !isFnParam && !isListParam {
			remainingParams = append(remainingParams, p)
			if i < len(fc.Arguments) {
				remainingArgs = append(remainingArgs, fc.Arguments[i])
			}
		}
	}

	targetFile := calleeFile
	if targetFile == nil {
		targetFile = file
	}
	for _, cv := range argFnMap {
		// A lifted closure's captures are expressions from this file, spliced
		// into the specialized body at the call through it, so that body has
		// to live where they resolve.
		if cv.lifted {
			targetFile = file
			break
		}
		if file != calleeFile {
			if f, _ := t.lookupFnTypeInScope(enclosingFn, file, cv.targetFnPrefix, cv.targetFnName); f == file {
				targetFile = file
				break
			}
		}
	}
	if len(listArgs) > 0 {
		// The elements are this file's expressions, spliced into the clone's
		// body where the fold is unrolled, so that body has to live here too.
		targetFile = file
	}
	if isNested && enclosingFn != nil {
		targetFile = file
	}

	// A function synthesized into one file and called from a body that ended up
	// in another has to be reached by prefix. That only happens once the body's
	// file is settled, which is here.
	for _, cv := range argFnMap {
		if cv.targetFnPrefix == "" && cv.targetFile != nil && cv.targetFile != targetFile {
			cv.targetFnPrefix = cv.callerPrefix
		}
	}

	// A clone moved out of its own package keeps reaching that package's names.
	if targetFile != calleeFile && calleeFile != nil && !(isNested && enclosingFn != nil) {
		t.qualifyForeignBody(clonedFn, calleeFile, targetFile)
	}

	if enclosingFn != nil {
		// The elements are spliced into the clone, which is a function of its
		// own: a parameter or local of the function the list was written in is
		// not in its scope. Every such name an element reads becomes a leading
		// parameter of the clone, passed from the call — the same thing a
		// lifted closure does with its captures.
		//
		// A lifted closure's captures are spliced into the clone the same way,
		// at the call through it, so the names they read are passed in too —
		// unless the clone is nested in the enclosing function, which already
		// sees them.
		spliced := make(map[string][]*program.Operand)
		if len(listArgs) > 0 {
			listArgs = cloneListArgs(listArgs)
			for k, els := range listArgs {
				spliced[k] = els
			}
		}
		for paramName, cv := range argFnMap {
			if isNested || !cv.lifted || len(cv.captures) == 0 {
				continue
			}
			values := make([]*program.Operand, len(cv.captures))
			for i := range cv.captures {
				values[i] = cv.captures[i].value
			}
			spliced[paramName] = values
		}
		if len(spliced) > 0 {
			encParams, encArgs := t.enclosingNamesForList(file, enclosingFn, clonedFn, spliced, fc)
			remainingParams = append(encParams, remainingParams...)
			remainingArgs = append(encArgs, remainingArgs...)
		}
	}

	clonedFn.Parameters = remainingParams
	if len(listArgs) > 0 {
		t.listBindings[clonedFn] = listArgs
	}
	fc.Arguments = remainingArgs
	fc.Reference.Name = specializedName
	if targetFile == file {
		fc.Reference.Prefix = ""
	}
	fc.Reference.RefPath = ""
	fc.Reference.PkgPath = ""

	for _, assign := range clonedFn.Assignments {
		if err := t.rewriteOperand(targetFile, clonedFn, assign.Expression, paramEnv); err != nil {
			return err
		}
	}
	for _, ret := range clonedFn.Returns {
		if err := t.rewriteOperand(targetFile, clonedFn, ret, paramEnv); err != nil {
			return err
		}
	}

	paramNames := make([]string, len(remainingParams))
	for i, p := range remainingParams {
		paramNames[i] = p.Name
	}

	if isNested && enclosingFn != nil {
		enclosingFn.FnTypes, _ = program.AddFunction(enclosingFn.FnTypes, specializedName, paramNames)
		for _, sFt := range enclosingFn.FnTypes {
			if sFt.Name == specializedName {
				for _, fg := range sFt.Groups {
					if fg.ParamCount == int32(len(paramNames)) {
						fg.Functions = []*program.Function{clonedFn}
						break
					}
				}
				break
			}
		}
	} else {
		targetFile.FnTypes, _ = program.AddFunction(targetFile.FnTypes, specializedName, paramNames)
		for _, sFt := range targetFile.FnTypes {
			if sFt.Name == specializedName {
				for _, fg := range sFt.Groups {
					if fg.ParamCount == int32(len(paramNames)) {
						fg.Functions = []*program.Function{clonedFn}
						break
					}
				}
				if fileMap, ok := t.fileFunctions[targetFile]; ok {
					fileMap[specializedName] = sFt
				}
				break
			}
		}
	}

	// Recursively transform the newly specialized function
	return t.transformFunction(targetFile, clonedFn, paramEnv)
}
