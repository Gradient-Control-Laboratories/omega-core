package types

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/internal"
	"gradientcontrol.com/omega/core/program"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

func inferSingleType(op *program.Operand, env *TypeEnv, file *program.PackageFile, pa *PackageAnalyzer) (*program.Type, error) {
	types, err := inferOperandType(op, env, file, pa)
	if err != nil {
		return nil, err
	}
	if len(types) != 1 {
		return nil, fmt.Errorf("expression returned multiple values where a single value was expected: multi-value functions (like *Normal) cannot be used directly in expressions; assign them first")
	}
	op.Type = types[0]
	return types[0], nil
}

func inferOperandType(op *program.Operand, env *TypeEnv, file *program.PackageFile, pa *PackageAnalyzer) ([]*program.Type, error) {
	if op == nil || op.T == nil {
		return nil, nil
	}

	// If already computed and not inside a user function body, return cached type
	if op.Type != nil && pa.inUserFn == 0 {
		return []*program.Type{op.Type}, nil
	}

	var res []*program.Type
	var err error

	switch v := op.T.(type) {
	case *program.Operand_Value:
		res = []*program.Type{program.NewScalarType()}

	case *program.Operand_Reference:
		ref := v.Reference
		if ref.Prefix != "" {
			fi := pa.fileImports[file]
			importedEnv, exists := fi.named[ref.Prefix]
			if !exists {
				importPath := ref.Prefix
				if path, ok := gsllibs.PathForPrefix(ref.Prefix); ok {
					importPath = path
				}
				importedPkg, err := pa.prog.GetImport(importPath)
				if err == nil && importedPkg != nil {
					importedEnv, err = analyzePackage(importedPkg, importPath, pa.prog, pa.pkgTypes, pa.fnFiles, pa.fileImports, pa.filePkgs)
					if err == nil {
						pa.pkgTypes[importPath] = importedEnv
						fi.named[ref.Prefix] = importedEnv
						exists = true
					}
				}
			}
			if !exists {
				keys := make([]string, 0, len(fi.named))
				for k := range fi.named {
					keys = append(keys, k)
				}
				return nil, fmt.Errorf("unknown package prefix %q. Available prefixes: %v in file %s", ref.Prefix, keys, file.Path)
			}
			t := importedEnv.getVar(ref.Name)
			if t == nil {
				if fnType := importedEnv.getFn(ref.Name); fnType != nil {
					t = program.NewFunctionType(ref.Prefix, ref.Name)
				} else {
					return nil, fmt.Errorf("undefined reference %s in package %s", ref.Name, ref.Prefix)
				}
			}
			res = []*program.Type{t}
		} else {
			name := ref.Name
			if name == "p" {
				res = []*program.Type{program.NewListType(3, program.NewScalarType())}
			} else if name == "time" {
				res = []*program.Type{program.NewScalarType()}
			} else if t := env.getVar(name); t != nil {
				res = []*program.Type{t}
			} else if fnType := env.getFn(name); fnType != nil {
				res = []*program.Type{program.NewFunctionType(pa.currentPkg.Path, name)}
			} else {
				// Check dot imports
				var dotFound *program.Type
				fi := pa.fileImports[file]
				for _, importedEnv := range fi.dots {
					if t := importedEnv.getVar(name); t != nil {
						dotFound = t
						break
					}
				}
				if dotFound != nil {
					res = []*program.Type{dotFound}
				} else {
					var foreignFound *program.Type
					declPkg := pa.filePkgs[file]
					if declPkg != nil && declPkg.Path != pa.currentPkg.Path {
						if depEnv := pa.pkgTypes[declPkg.Path]; depEnv != nil {
							if t := depEnv.getVar(name); t != nil {
								foreignFound = t
							} else if f := depEnv.getFn(name); f != nil {
								foreignFound = program.NewFunctionType(declPkg.Path, name)
							}
						}
					}
					if foreignFound != nil {
						res = []*program.Type{foreignFound}
					} else {
						var dotFnFound *program.FnType
						for _, importedEnv := range fi.dots {
							if f := importedEnv.getFn(name); f != nil {
								dotFnFound = f
								break
							}
						}
						if dotFnFound != nil {
							res = []*program.Type{program.NewFunctionType("", name)}
						} else {
							// Check package-level vars of the current package
							t, err := pa.resolvePackageVar(name)
							if err != nil {
								return nil, err
							}
							if t != nil {
								res = []*program.Type{t}
							} else {
								return nil, fmt.Errorf("undefined identifier: %s", name)
							}
						}
					}
				}
			}
		}

	case *program.Operand_FnCall:
		fc := v.FnCall
		name := fc.Reference.Name
		if program.IsBuiltin(name) {
			res, err = pa.inferBuiltinCall(name, fc.Arguments, env, file)
		} else {
			var fnType *program.FnType
			if fc.Reference.Prefix != "" {
				fi := pa.fileImports[file]
				importedEnv, exists := fi.named[fc.Reference.Prefix]
				if !exists {
					importPath := fc.Reference.Prefix
					if path, ok := gsllibs.PathForPrefix(fc.Reference.Prefix); ok {
						importPath = path
					}
					importedPkg, err := pa.prog.GetImport(importPath)
					if err == nil && importedPkg != nil {
						importedEnv, err = analyzePackage(importedPkg, importPath, pa.prog, pa.pkgTypes, pa.fnFiles, pa.fileImports, pa.filePkgs)
						if err == nil {
							pa.pkgTypes[importPath] = importedEnv
							fi.named[fc.Reference.Prefix] = importedEnv
							exists = true
						}
					}
				}
				if !exists {
					return nil, fmt.Errorf("unknown package prefix: %s", fc.Reference.Prefix)
				}
				fnType = importedEnv.getFn(name)
			} else {
				fnType = env.getFn(name)
				if fnType == nil {
					fi := pa.fileImports[file]
					for _, importedEnv := range fi.dots {
						if f := importedEnv.getFn(name); f != nil {
							fnType = f
							break
						}
					}
				}
				if fnType == nil {
					declPkg := pa.filePkgs[file]
					if declPkg != nil && declPkg.Path != pa.currentPkg.Path {
						if depEnv := pa.pkgTypes[declPkg.Path]; depEnv != nil {
							fnType = depEnv.getFn(name)
						}
					}
				}
			}
			if fnType == nil {
				if varType := env.getVar(name); varType != nil && varType.FnPath != "" {
					fnName := varType.FnPath
					// If the function reference belongs to an external package (e.g. standard library),
					// look up its signature in that package's type environment. Otherwise, fallback
					// to the current package environment.
					if varType.PkgPath != "" {
						if targetEnv := pa.pkgTypes[varType.PkgPath]; targetEnv != nil {
							fnType = targetEnv.getFn(fnName)
						}
					}
					if fnType == nil {
						fnType = env.getFn(fnName)
					}
					if fnType == nil {
						fi := pa.fileImports[file]
						for _, importedEnv := range fi.dots {
							if f := importedEnv.getFn(fnName); f != nil {
								fnType = f
								break
							}
						}
					}
				}
			}
			if fnType == nil {
				return nil, fmt.Errorf("undefined identifier: %s", name)
			}

			// Find group with matching paramCount
			var targetFn *program.Function
			for _, group := range fnType.Groups {
				if int(group.ParamCount) == len(fc.Arguments) {
					if len(group.Functions) > 0 {
						targetFn = group.Functions[0]
					}
					break
				}
			}
			if targetFn == nil {
				// Construct a friendly compiler error containing all available overload signatures
				// for this function name (e.g., "; perhaps one of these? sphere(radius), sphere(center, radius)")
				// to help the user identify parameter count mismatches.
				var sigs []string
				fullName := fnType.Name
				if fc.Reference.Prefix != "" {
					fullName = fc.Reference.Prefix + "." + fnType.Name
				}
				for _, g := range fnType.Groups {
					if len(g.Functions) == 0 {
						continue
					}
					var params []string
					for _, p := range g.Functions[0].Parameters {
						params = append(params, p.Name)
					}
					sigs = append(sigs, fmt.Sprintf("%s(%s)", fullName, strings.Join(params, ", ")))
				}
				var suggestion string
				if len(sigs) > 0 {
					suggestion = "; perhaps one of these? " + strings.Join(sigs, ", ")
				}
				return nil, fmt.Errorf("no function named %q with %d parameters%s", fnType.Name, len(fc.Arguments), suggestion)
			}

			// Infer argument types
			argTypes := make([]*program.Type, len(fc.Arguments))
			for i, arg := range fc.Arguments {
				t, err := inferSingleType(arg, env, file, pa)
				if err != nil {
					return nil, err
				}
				argTypes[i] = t
			}

			res, err = pa.inferUserCall(targetFn, argTypes, env, file)
		}

	case *program.Operand_PropRef:
		return nil, fmt.Errorf("property references are not supported yet")

	case *program.Operand_Expression:
		expr := v.Expression
		switch expr.Operation {
		// Modulo, min, and max: infer all operand types and unify them into a common type.
		// These operations evaluate elementwise across matching vector, list, or scalar
		// dimensions, preserving the broadest component shape.
		case program.Operation_MODULO, program.Operation_MIN, program.Operation_MAX:
			var maxType *program.Type
			for _, argOp := range expr.Operands {
				t, err := inferSingleType(argOp, env, file, pa)
				if err != nil {
					return nil, err
				}
				if maxType == nil {
					maxType = t
				} else {
					maxType, err = commonType(maxType, t)
					if err != nil {
						return nil, err
					}
				}
			}
			res = []*program.Type{maxType}

		// Arithmetic operators: check for hardware vector/matrix types. When present,
		// chain DeriveBinaryOpType across operands to handle linear algebra rules (such
		// as matrix-matrix or matrix-vector multiplication). Otherwise, unify scalar or
		// component-wise collection types using commonType.
		case program.Operation_ADD, program.Operation_SUB, program.Operation_DIV, program.Operation_MUL:
			// Invariant: parser grammar and primitive validation guarantee len(expr.Operands) >= 2.
			operandTypes := make([]*program.Type, len(expr.Operands))
			anyMatrixOrVector := false
			for i, argOp := range expr.Operands {
				t, err := inferSingleType(argOp, env, file, pa)
				if err != nil {
					return nil, err
				}
				operandTypes[i] = t
				if t.IsHardwareMatrix() || t.IsHardwareVector() {
					anyMatrixOrVector = true
				}
			}
			if anyMatrixOrVector {
				accType := operandTypes[0]
				for i := 1; i < len(operandTypes); i++ {
					stepType := operandTypes[i]
					derived, err := program.DeriveBinaryOpType(expr.Operation, accType, stepType, i+1)
					if err != nil {
						return nil, err
					}
					accType = derived
				}
				res = []*program.Type{accType}
			} else {
				// Pure scalar / non-matrix/vector
				maxType := operandTypes[0]
				for i := 1; i < len(operandTypes); i++ {
					var err error
					maxType, err = commonType(maxType, operandTypes[i])
					if err != nil {
						return nil, err
					}
				}
				res = []*program.Type{maxType}
			}

		// Unary mathematical functions: infer the operand's type. These functions
		// (trigonometric, square root, rounding, etc.) operate elementwise and preserve
		// the operand's shape without altering its type.
		case program.Operation_NEG, program.Operation_SQUARE, program.Operation_SQRT,
			program.Operation_FLOOR, program.Operation_CEIL, program.Operation_ROUND,
			program.Operation_ABS, program.Operation_EXP, program.Operation_LN,
			program.Operation_SIN, program.Operation_COS, program.Operation_TAN,
			program.Operation_ASIN, program.Operation_ACOS, program.Operation_ATAN:
			// Invariant: parser grammar and primitive validation guarantee exactly 1 operand.
			t, err := inferSingleType(expr.Operands[0], env, file, pa)
			if err != nil {
				return nil, err
			}
			res = []*program.Type{t}

		// Binary mathematical operations (pow, atan2): infer both operand types to
		// compute their common unified type. This ensures both operands are compatible
		// and supports elementwise evaluation or scalar broadcasting.
		case program.Operation_POW, program.Operation_ATAN2:
			// Invariant: parser primitive validation guarantees exactly 2 operands.
			t1, err := inferSingleType(expr.Operands[0], env, file, pa)
			if err != nil {
				return nil, err
			}
			t2, err := inferSingleType(expr.Operands[1], env, file, pa)
			if err != nil {
				return nil, err
			}
			common, err := commonType(t1, t2)
			if err != nil {
				return nil, err
			}
			res = []*program.Type{common}

		// Comparison operators (eq, ne, lt, le, gt, ge, compare): ensure all operands are
		// dimensionally compatible (disallowing direct scalar-to-collection comparisons)
		// and resolve to a common boolean shape matching the operands' dimensions.
		case program.Operation_EQ, program.Operation_NE, program.Operation_LT, program.Operation_LE, program.Operation_GT, program.Operation_GE, program.Operation_COMPARE:
			var maxType *program.Type
			for _, argOp := range expr.Operands {
				t, err := inferSingleType(argOp, env, file, pa)
				if err != nil {
					return nil, err
				}
				if maxType == nil {
					maxType = t
				} else {
					if (maxType.IsList() || maxType.IsMatrix()) && t.IsScalar() || (maxType.IsScalar() && (t.IsList() || t.IsMatrix())) {
						return nil, fmt.Errorf("cannot compare %s with %s", maxType.TypeString(), t.TypeString())
					}
					maxType, err = commonType(maxType, t)
					if err != nil {
						return nil, err
					}
				}
			}
			res = []*program.Type{maxType} // Boolean of the same shape

		// Logical operators (and, or, not): infer operand types and unify them into a
		// common type. Logical operations evaluate elementwise over boolean values and
		// preserve component shapes across scalar or collection operands.
		case program.Operation_AND, program.Operation_OR, program.Operation_NOT:
			var maxType *program.Type
			for _, argOp := range expr.Operands {
				t, err := inferSingleType(argOp, env, file, pa)
				if err != nil {
					return nil, err
				}
				if maxType == nil {
					maxType = t
				} else {
					maxType, err = commonType(maxType, t)
					if err != nil {
						return nil, err
					}
				}
			}
			res = []*program.Type{maxType}

		// Coordinate remapping (remap): infer the types of transformed coordinate arguments
		// (operands 1..N) to validate them, and return the type of the base shape/field
		// (operand 0), since remapping changes spatial evaluation points without altering output type.
		case program.Operation_REMAP:
			// Invariant: parser primitive validation guarantees at least 2 operands (target shape + coordinate mappings).
			for i := 1; i < len(expr.Operands); i++ {
				_, err := inferSingleType(expr.Operands[i], env, file, pa)
				if err != nil {
					return nil, err
				}
			}
			t, err := inferOperandType(expr.Operands[0], env, file, pa)
			if err != nil {
				return nil, err
			}
			res = t

		// Collection literal construction: validate that all elements are uniformly scalars
		// or uniformly vectors of identical length. If vectors, produce a Matrix type; if scalars,
		// produce a List/Vector type whose length equals the operand count.
		case program.Operation_COLLECT:
			var elemVecType *program.Type
			for i, argOp := range expr.Operands {
				t, err := inferSingleType(argOp, env, file, pa)
				if err != nil {
					return nil, err
				}
				if i == 0 {
					if t.Size > 1 && t.ElemType != nil {
						elemVecType = t
					}
					continue
				}
				isVec := t.Size > 1 && t.ElemType != nil
				expectedVec := elemVecType != nil
				if isVec != expectedVec {
					return nil, fmt.Errorf("collection elements must be all scalars or all vectors")
				}
				if isVec && t.Size != elemVecType.Size {
					return nil, fmt.Errorf("collection elements must be vectors of the same length")
				}
			}
			if elemVecType != nil {
				res = []*program.Type{program.NewMatrixType(int32(len(expr.Operands)), elemVecType)}
			} else {
				res = []*program.Type{program.NewListType(int32(len(expr.Operands)), program.NewScalarType())}
			}

		// Collection reduction: collapse a collection down to an accumulated value. If an initial
		// accumulator operand is provided (>= 3 operands), return its type; otherwise, return the
		// collection element's type, defaulting to scalar.
		case program.Operation_REDUCE:
			// Invariant: parser primitive validation guarantees at least 2 operands (collection + reducer).
			var firstType *program.Type
			var secondType *program.Type
			for i, argOp := range expr.Operands {
				t, err := inferSingleType(argOp, env, file, pa)
				if err != nil {
					return nil, err
				}
				switch i {
				case 0:
					firstType = t
				case 1:
					secondType = t
				}
			}
			if len(expr.Operands) >= 3 {
				res = []*program.Type{firstType}
			} else {
				var elemType *program.Type
				if secondType != nil && secondType.ElemType != nil {
					elemType = secondType.ElemType
				} else {
					elemType = program.NewScalarType()
				}
				res = []*program.Type{elemType}
			}

		// Collection mapping: apply a function over collection elements. Infer the collection's
		// element type and simulate invoking the mapping function on it. Return a List (if the mapped
		// result is scalar) or a Matrix (if the mapped result is a vector) matching collection size.
		case program.Operation_MAP:
			// Invariant: parser primitive validation guarantees at least 2 operands (function + collection).
			var collType *program.Type
			for i, argOp := range expr.Operands {
				t, err := inferSingleType(argOp, env, file, pa)
				if err != nil {
					return nil, err
				}
				if i == len(expr.Operands)-1 {
					collType = t
				}
			}
			var elemType *program.Type
			if collType != nil && collType.ElemType != nil {
				elemType = collType.ElemType
			} else {
				elemType = program.NewScalarType()
			}
			fnOp := expr.Operands[0]
			var fnType *program.FnType
			if refOp, ok := fnOp.T.(*program.Operand_Reference); ok {
				name := refOp.Reference.Name
				fnType = env.getFn(name)
				if fnType == nil {
					fi := pa.fileImports[file]
					for _, importedEnv := range fi.dots {
						if f := importedEnv.getFn(name); f != nil {
							fnType = f
							break
						}
					}
				}
			}
			var retType *program.Type = program.NewScalarType()
			if fnType != nil && len(fnType.Groups) > 0 && len(fnType.Groups[0].Functions) > 0 {
				targetFn := fnType.Groups[0].Functions[0]
				var argTypes []*program.Type
				if len(targetFn.Parameters) == 2 {
					argTypes = []*program.Type{elemType, program.NewScalarType()}
				} else {
					argTypes = []*program.Type{elemType}
				}
				retTypes, err := pa.inferUserCall(targetFn, argTypes, env, file)
				if err == nil && len(retTypes) > 0 {
					retType = retTypes[0]
				}
			}
			if retType.IsScalar() {
				if collType != nil && collType.IsScalar() {
					res = []*program.Type{retType}
				} else {
					res = []*program.Type{program.NewListType(collType.Size, retType)}
				}
			} else {
				res = []*program.Type{program.NewMatrixType(collType.Size, retType)}
			}

		// Conditional selection (select): ensure all value branches (operands 1..N) unify to a
		// common type, and verify that the condition operand (operand 0) is compatible with that
		// common shape, returning the unified value type.
		case program.Operation_SELECT:
			var commonValType *program.Type
			for i := 1; i < len(expr.Operands); i++ {
				t, err := inferSingleType(expr.Operands[i], env, file, pa)
				if err != nil {
					return nil, err
				}
				if commonValType == nil {
					commonValType = t
				} else {
					commonValType, err = commonType(commonValType, t)
					if err != nil {
						return nil, err
					}
				}
			}
			condType, err := inferSingleType(expr.Operands[0], env, file, pa)
			if err != nil {
				return nil, err
			}
			_, err = commonType(condType, commonValType)
			if err != nil {
				return nil, err
			}
			res = []*program.Type{commonValType}

		// Collection indexing: verify that the target operand is a list or matrix and the index is
		// a scalar. Returns the target's element type, stepping down one dimension in the type hierarchy.
		case program.Operation_INDEX:
			targetType, err := inferSingleType(expr.Operands[0], env, file, pa)
			if err != nil {
				return nil, err
			}
			indexType, err := inferSingleType(expr.Operands[1], env, file, pa)
			if err != nil {
				return nil, err
			}
			if !targetType.IsList() && !targetType.IsMatrix() {
				return nil, fmt.Errorf("index operator expects collection target, got %s", targetType.TypeString())
			}
			if !indexType.IsScalar() {
				return nil, fmt.Errorf("index operator expects scalar index, got %s", indexType.TypeString())
			}
			if targetType.ElemType != nil {
				res = []*program.Type{targetType.ElemType}
			} else {
				res = []*program.Type{program.NewScalarType()}
			}

		// Member access and swizzling: extract components (e.g. .x, .xyz) from scalars or vectors.
		// Validates that requested swizzle letters stay within target bounds and returns either a
		// scalar (for 1-char access) or a list/vector of scalars (for multi-char swizzles).
		case program.Operation_ACCESS:
			// Invariant: parser member syntax (target.member) guarantees exactly 2 operands (target + reference).
			targetType, err := inferSingleType(expr.Operands[0], env, file, pa)
			if err != nil {
				return nil, err
			}
			refOp, ok := expr.Operands[1].T.(*program.Operand_Reference)
			if !ok {
				return nil, fmt.Errorf("access operator expects second operand to be a reference")
			}
			member := refOp.Reference.Name

			if targetType.IsMatrix() {
				return nil, fmt.Errorf("member access is not supported on %s; use indexing instead", targetType.TypeString())
			}
			if targetType.IsScalar() {
				if len(member) == 1 {
					res = []*program.Type{program.NewScalarType()}
				} else {
					res = []*program.Type{program.NewListType(int32(len(member)), program.NewScalarType())}
				}
				break
			}

			// Bounds check
			switch targetType.Size {
			case 3:
				if strings.Contains(member, "w") {
					return nil, fmt.Errorf("cannot use \"w\" to access a collection of 3 elements")
				}
			case 2:
				if strings.Contains(member, "z") || strings.Contains(member, "w") {
					return nil, fmt.Errorf("cannot use \"z\" or \"w\" to access a collection of 2 elements")
				}
			case 1:
				if strings.Contains(member, "y") || strings.Contains(member, "z") || strings.Contains(member, "w") {
					return nil, fmt.Errorf("cannot use \"y\", \"z\", or \"w\" to access a collection of 1 element")
				}
			}

			if len(member) == 1 {
				res = []*program.Type{program.NewScalarType()}
			} else {
				res = []*program.Type{program.NewListType(int32(len(member)), program.NewScalarType())}
			}

		// High-level property queries: fail fast if encountered during type analysis, as these
		// operations must be desugared and expanded into core AST nodes by property.Expand beforehand.
		case program.Operation_DERIV, program.Operation_SECONDCLOSEST, program.Operation_BOUNDINGBOX,
			program.Operation_SURFACEU, program.Operation_SURFACEV:
			return nil, internal.InternalErr(fmt.Errorf("%s should be expanded by property.Expand before type analysis", expr.Operation))

		// Unspecified operations: fail with an error because uninitialized or blank operation
		// enum values indicate a malformed AST.
		case program.Operation_UNSPECIFIED_OPERAND:
			return nil, internal.InternalErr(fmt.Errorf("unspecified operation"))

		// Catch-all for unhandled operations: fail fast if new operation enum variants have been
		// added to the AST without corresponding type inference rules.
		default:
			return nil, internal.InternalErr(fmt.Errorf("type inference case for operation %s is not implemented", expr.Operation))
		}

	default:
		return nil, internal.InternalErr(fmt.Errorf("type inference case for operand type %T is not implemented", op.T))
	}

	if len(res) == 1 && pa.inUserFn == 0 {
		op.Type = res[0]
	}
	return res, err
}

func (pa *PackageAnalyzer) inferUserCall(fn *program.Function, argTypes []*program.Type, callerEnv *TypeEnv, file *program.PackageFile) ([]*program.Type, error) {
	pa.inUserFn++
	defer func() { pa.inUserFn-- }()

	if pa.inUserFn > 64 {
		return nil, fmt.Errorf("recursion detected; recursive functions are prohibited in GSL")
	}

	if len(fn.Parameters) != len(argTypes) {
		return nil, fmt.Errorf("function expects %d arguments, got %d", len(fn.Parameters), len(argTypes))
	}

	declFile := pa.fnFiles[fn]
	if declFile == nil {
		declFile = file
	}

	var parentEnv *TypeEnv = pa.env
	if pa.isPkgFn[fn] {
		if declPkg := pa.filePkgs[declFile]; declPkg != nil {
			if depEnv := pa.pkgTypes[declPkg.Path]; depEnv != nil {
				parentEnv = depEnv
			}
		}
	} else {
		parentEnv = callerEnv
	}

	localEnv := newEnv(parentEnv)
	for i, param := range fn.Parameters {
		param.Type = argTypes[i]
		localEnv.setVar(param.Name, argTypes[i])
	}
	for _, localFn := range fn.FnTypes {
		localEnv.fns[localFn.Name] = localFn
	}

	for _, assign := range fn.Assignments {
		rhsTypes, err := inferOperandType(assign.Expression, localEnv, declFile, pa)
		if err != nil {
			return nil, err
		}
		if len(rhsTypes) != len(assign.Variables) {
			return nil, fmt.Errorf("assignment count mismatch in function: expected %d values, got %d; multi-value functions (like *Normal) must assign both values", len(assign.Variables), len(rhsTypes))
		}
		for i, v := range assign.Variables {
			v.Type = rhsTypes[i]
			localEnv.setVar(v.Name, rhsTypes[i])
		}
	}

	var retTypes []*program.Type
	for _, retOp := range fn.Returns {
		ts, err := inferOperandType(retOp, localEnv, declFile, pa)
		if err != nil {
			return nil, err
		}
		retTypes = append(retTypes, ts...)
	}

	return retTypes, nil
}
