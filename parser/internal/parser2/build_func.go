package parser2

import (
	"fmt"
	"math"
	"slices"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/parser/internal/annotation"
	"gradientcontrol.com/omega/core/parser/internal/identifier2"
	"gradientcontrol.com/omega/core/program"
)

func buildFunction(parentScope scope.BuilderScope, decl *NamedFunction) error {
	paramNames := make([]string, len(decl.Parameters.Params))
	for i, paramVar := range decl.Parameters.Params {
		paramNames[i] = paramVar.Name
	}
	fnScope, err := parentScope.NewFnScope(string(decl.Name), paramNames)
	if err != nil {
		return &participle.ParseError{
			Pos: decl.Pos,
			Msg: err.Error(),
		}
	}

	for i, paramDecl := range decl.Parameters.Params {
		var rng *program.Range
		kind, anchor := "", ""
		if a := paramDecl.Annotation; a != nil {
			switch a.Name {
			case annotation.Range:
				if len(a.Args) != 2 {
					return &participle.ParseError{
						Pos: a.Pos,
						Msg: "@range annotation on a parameter requires exactly 2 arguments: @range(min, max)",
					}
				}
				rng, err = buildRangeAnnotation(parentScope, a.Pos, a.EndPos, a.Args)
				if err != nil {
					return err
				}
			default:
				// A control kind, @color, @point, @direction or @transform:
				// what the argument is, recorded on the parameter and never
				// lowered, so the function compiles exactly as unannotated.
				if err := annotation.ValidateArgCount(a.Name, len(a.Args)); err != nil {
					return &participle.ParseError{
						Pos: a.Pos,
						Msg: err.Error(),
					}
				}
				kind = a.Name
				// @direction(anchor) on a parameter names the sibling
				// parameter the arrow is drawn from.
				if len(a.Args) == 1 {
					name, ok := plainReference(a.Args[0].Expr)
					if !ok || name == paramDecl.Name || !slices.Contains(paramNames, name) {
						return &participle.ParseError{
							Pos: a.Args[0].Pos,
							Msg: fmt.Sprintf("@%s's anchor on a parameter must name another parameter of the function", a.Name),
						}
					}
					anchor = name
				}
			}
		}
		fnScope.SetParam(i, posToLocation(paramDecl.Pos), rng)
		if kind != "" {
			fnScope.SetParamKind(i, kind, anchor)
		}
	}

	err = buildDeclarations(fnScope, decl.Body)
	if err != nil {
		return err
	}

	return buildReturnExpressions(fnScope, decl.Return)
}

func buildReturnExpressions(buildScope scope.BuilderScope, expression []*Expression) error {
	operands, err := buildExpressions(buildScope, expression)
	if err != nil {
		return err
	}
	buildScope.SetReturnOperands(operands)
	return nil
}

func buildAnonFunctionCall(buildScope scope.BuilderScope, v *AnonCall) (*program.Operand, error) {
	decl, err := buildAnonFunction(buildScope, v)
	if err != nil {
		return nil, err
	}

	// Build the anon function call

	var arguments []*program.Operand
	if v.Arguments != nil {
		arguments, err = buildExpressions(buildScope, v.Arguments.Expressions)
		if err != nil {
			return nil, err
		}
	} else {
		arguments = []*program.Operand{}
	}

	return program.NewFnCall(
		"", string(decl.Name), arguments,
		posToLocation(v.Pos), posToLocation(v.EndPos),
	), nil
}

func buildAnonFunction(buildScope scope.BuilderScope, v *AnonCall) (*NamedFunction, error) {
	anonFn := v.AnonFunction
	if anonFn == nil {
		return nil, fmt.Errorf("missing anonymous function")
	}

	// Build the anonymous function

	decl := anonToNamedFunction(buildScope, anonFn)
	err := buildFunction(buildScope, decl)
	if err != nil {
		return nil, err
	}
	return decl, nil
}

func anonToNamedFunction(buildScope scope.BuilderScope, anonFn *AnonFunction) *NamedFunction {
	fnName := program.AnonFnPrefix
	fnAnonCount := 2
	for {
		elem := buildScope.NamedElem(fnName)
		if elem == nil || elem.ContainingScope() != buildScope {
			break
		}
		fnName = fmt.Sprintf("%s_%d", program.AnonFnPrefix, fnAnonCount)
		fnAnonCount++
	}

	return &NamedFunction{
		Pos:        anonFn.Pos,
		Name:       FunctionName(fnName),
		Parameters: anonFn.Parameters,
		Body:       anonFn.Body,
		Return:     anonFn.Return,
		EndPos:     anonFn.EndPos,
	}
}

func buildFunctionCall(buildScope scope.BuilderScope, fnCall *FunctionCall) (*program.Operand, error) {
	pkgName := fnCall.PkgName
	fnName := fnCall.Name
	pos := fnCall.Pos
	endPos := fnCall.EndPos

	// mesh(), bvhSdf(), and vdbSdf() require special handling: their first argument is
	// a non-negative integer literal selecting an imported mesh by session
	// import order, not a scope-checked reference or a runtime expression.
	if fnName == "mesh" {
		if pkgName != "" {
			return nil, &participle.ParseError{
				Pos: pos,
				Msg: "cannot import mesh",
			}
		}
		return buildMeshCall(buildScope, fnCall, pos, endPos)
	}
	if fnName == "bvhSdf" {
		if pkgName != "" {
			return nil, &participle.ParseError{
				Pos: pos,
				Msg: "cannot import bvhSdf",
			}
		}
		return buildBvhSdfCall(buildScope, fnCall, pos, endPos)
	}
	if fnName == "vdbSdf" {
		if pkgName != "" {
			return nil, &participle.ParseError{
				Pos: pos,
				Msg: "cannot import vdbSdf",
			}
		}
		return buildVdbSdfCall(buildScope, fnCall, pos, endPos)
	}

	// name.d(var), the derivative of the variable `name` with respect to
	// `var`, reaches here as a qualified call: `name.d(` reads as a package's
	// function d until the prefix turns out not to be an import.
	if fnName == program.DAccessor && pkgName != "" {
		if _, isImport := buildScope.NamedElem(pkgName).(*scope.PkgRef); !isImport {
			start := posToLocation(pos)
			refEnd := program.NewLoc(
				start.Line,
				start.Column+int32(len(pkgName)),
				start.Offset+int32(len(pkgName)),
			)
			target := program.NewReference("", pkgName, start, refEnd)
			return newDerivWrt(buildScope, target, fnCall.Arguments, pos, endPos)
		}
	}

	arguments, err := buildExpressions(buildScope, fnCall.Arguments)
	if err != nil {
		return nil, err
	}

	primFunc := program.PrimitiveFuncNamed(fnName)
	if primFunc != nil {
		if pkgName != "" {
			return nil, &participle.ParseError{
				Pos: pos,
				Msg: fmt.Sprintf("cannot import primitive function: %s", fnName),
			}
		}
		if err != nil {
			return nil, err
		}
		return buildPrimitiveCall(primFunc, arguments, pos, endPos)
	}

	builtin := program.BuiltinNamed(fnName)
	if builtin != nil {
		if pkgName != "" {
			return nil, &participle.ParseError{
				Pos: pos,
				Msg: fmt.Sprintf("cannot import builtin function: %s", fnName),
			}
		}
		if err != nil {
			return nil, err
		}
		return buildBuiltinCall(builtin, arguments, pos, endPos)
	}

	if pkgName == "" && fnName == "audio" {
		return program.NewFnCall(
			"", fnName, arguments,
			posToLocation(pos), posToLocation(endPos),
		), nil
	}

	if pkgName == "" {
		err := identifier2.Validate(fnName)
		if err != nil {
			return nil, &participle.ParseError{
				Pos: pos,
				Msg: err.Error(),
			}
		}
	} else {
		err := identifier2.Validate(fnName)
		if err != nil {
			return nil, &participle.ParseError{
				Pos: pos,
				Msg: err.Error(),
			}
		}
	}

	return program.NewFnCall(
		pkgName, fnName, arguments,
		posToLocation(pos), posToLocation(endPos),
	), nil
}

func buildPrimitiveCall(
	primFunc *program.PrimitiveFunc, arguments []*program.Operand, pos, endPos lexer.Position,
) (*program.Operand, error) {

	if len(arguments) < primFunc.MinArgs || len(arguments) > primFunc.MaxArgs {

		var expectedMsg string
		if primFunc.MinArgs == primFunc.MaxArgs {
			expectedMsg = fmt.Sprintf("expected %d arguments", primFunc.MinArgs)
		} else if primFunc.MaxArgs < 10 {
			expectedMsg = fmt.Sprintf("expected between %d and %d arguments", primFunc.MinArgs, primFunc.MaxArgs)
		} else {
			expectedMsg = fmt.Sprintf("expected at least %d arguments", primFunc.MinArgs)
		}

		return nil, &participle.ParseError{
			Pos: pos,
			Msg: fmt.Sprintf("invalid number of arguments for primitive %s: %s, got %d", primFunc.Name, expectedMsg, len(arguments)),
		}
	}

	// GSL atan with 2 arguments maps to the Operation_ATAN2 operation
	primOperation := primFunc.Op
	if primOperation == program.Operation_ATAN && len(arguments) == 2 {
		primOperation = program.Operation_ATAN2
	}

	return program.NewExpression(
		primOperation, arguments,
		posToLocation(pos), posToLocation(endPos),
	), nil
}

func buildBuiltinCall(builtin *program.BuiltinType, arguments []*program.Operand, pos, endPos lexer.Position) (*program.Operand, error) {
	if len(arguments) < builtin.MinArgs || len(arguments) > builtin.MaxArgs {
		if builtin.MinArgs == builtin.MaxArgs {
			return nil, fmt.Errorf(
				"builtin %s requires %d arguments, got %d",
				builtin.Name, builtin.MinArgs, len(arguments))
		}
		return nil, fmt.Errorf(
			"builtin %s requires between %d and %d arguments, got %d",
			builtin.Name, builtin.MinArgs, builtin.MaxArgs, len(arguments))
	}

	return program.NewFnCall(
		"", builtin.Name, arguments,
		posToLocation(pos), posToLocation(endPos),
	), nil
}

// extractPositiveFloatLiteral extracts a strictly-positive float literal
// from an Expression AST node, returning (value, true) on success or
// (0, false) otherwise. Used by builtins like vdbSdf that accept a
// numeric tuning parameter (voxel size, narrow-band width) that must be
// known at compile time so the canvas agent can use it as a bake key.
func extractPositiveFloatLiteral(expr *Expression) (float64, bool) {
	if expr == nil || expr.Value == nil {
		return 0, false
	}
	lor := expr.Value
	if lor.Left == nil || len(lor.Right) > 0 {
		return 0, false
	}
	if lor.Left.Left == nil || len(lor.Left.Right) > 0 {
		return 0, false
	}
	eq := lor.Left.Left
	if eq.Left == nil || eq.Right != nil {
		return 0, false
	}
	if eq.Left.Left == nil || eq.Left.Right != nil {
		return 0, false
	}
	terms := eq.Left.Left
	if terms.Left == nil || len(terms.Right) > 0 {
		return 0, false
	}
	if terms.Left.Left == nil || len(terms.Left.Right) > 0 {
		return 0, false
	}
	factor := terms.Left.Left
	if factor.Negative || factor.Not {
		return 0, false
	}
	if factor.Value == nil {
		return 0, false
	}
	num, ok := factor.Value.(*Number)
	if !ok || num.Value == nil {
		return 0, false
	}
	f := *num.Value
	if f <= 0 || math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, false
	}
	return f, true
}

// extractNonNegativeIntLiteral extracts a non-negative integer literal from an
// Expression AST node, returning (value, true) on success or (0, false) if the
// expression is not a plain non-negative integer.
func extractNonNegativeIntLiteral(expr *Expression) (int, bool) {
	if expr == nil || expr.Value == nil {
		return 0, false
	}
	lor := expr.Value
	if lor.Left == nil || len(lor.Right) > 0 {
		return 0, false
	}
	if lor.Left.Left == nil || len(lor.Left.Right) > 0 {
		return 0, false
	}
	eq := lor.Left.Left
	if eq.Left == nil || eq.Right != nil {
		return 0, false
	}
	if eq.Left.Left == nil || eq.Left.Right != nil {
		return 0, false
	}
	terms := eq.Left.Left
	if terms.Left == nil || len(terms.Right) > 0 {
		return 0, false
	}
	if terms.Left.Left == nil || len(terms.Left.Right) > 0 {
		return 0, false
	}
	factor := terms.Left.Left
	if factor.Negative || factor.Not {
		return 0, false
	}
	if factor.Value == nil {
		return 0, false
	}
	num, ok := factor.Value.(*Number)
	if !ok || num.Value == nil {
		return 0, false
	}
	f := *num.Value
	if f < 0 || f != math.Trunc(f) {
		return 0, false
	}
	return int(f), true
}

func extractStringLiteral(expr *Expression) (*String, bool) {
	if expr == nil || expr.Value == nil {
		return nil, false
	}
	lor := expr.Value
	if lor.Left == nil || len(lor.Right) > 0 {
		return nil, false
	}
	land := lor.Left
	if land.Left == nil || len(land.Right) > 0 {
		return nil, false
	}
	eq := land.Left
	if eq.Left == nil || eq.Right != nil {
		return nil, false
	}
	comp := eq.Left
	if comp.Left == nil || comp.Right != nil {
		return nil, false
	}
	terms := comp.Left
	if terms.Left == nil || len(terms.Right) > 0 {
		return nil, false
	}
	term := terms.Left
	if term.Left == nil || len(term.Right) > 0 {
		return nil, false
	}
	factor := term.Left
	if factor.Negative || factor.Not {
		return nil, false
	}
	if factor.Value == nil {
		return nil, false
	}
	str, ok := factor.Value.(*String)
	return str, ok
}

// buildMeshCall handles mesh(index, position, scale, rotation) where the first
// argument is a non-negative integer literal identifying an imported mesh by
// session import order.
func buildMeshCall(buildScope scope.BuilderScope, funcCall *FunctionCall, pos, endPos lexer.Position) (*program.Operand, error) {
	builtin := program.BuiltinNamed("mesh")
	rawArgs := funcCall.Arguments
	if len(rawArgs) < builtin.MinArgs || len(rawArgs) > builtin.MaxArgs {
		return nil, &participle.ParseError{
			Pos: pos,
			Msg: fmt.Sprintf("mesh requires between %d and %d arguments, got %d",
				builtin.MinArgs, builtin.MaxArgs, len(rawArgs)),
		}
	}

	// First argument: a non-negative integer literal, or any expression that
	// folds to one (e.g. an id returned by an importStep/importMesh import).
	indexOp, err := buildMeshIndexArg(buildScope, rawArgs[0], pos, endPos)
	if err != nil {
		return nil, err
	}

	// Remaining arguments (position, scale, rotation) are built normally
	var restArgs []*program.Operand
	if len(rawArgs) > 1 {
		var err error
		restArgs, err = buildExpressions(buildScope, rawArgs[1:])
		if err != nil {
			return nil, err
		}
	}

	arguments := append([]*program.Operand{indexOp}, restArgs...)
	return program.NewFnCall(
		"", "mesh", arguments,
		posToLocation(pos), posToLocation(endPos),
	), nil
}

// buildBvhSdfCall handles bvhSdf(index), which returns the signed distance
// from the implicit position p to an imported mesh sampled by a
// BVH-accelerated triangle-set distance oracle. The single argument must be
// a non-negative integer literal; it selects the mesh by session-scoped
// import order.
func buildBvhSdfCall(buildScope scope.BuilderScope, funcCall *FunctionCall, pos, endPos lexer.Position) (*program.Operand, error) {
	builtin := program.BuiltinNamed("bvhSdf")
	rawArgs := funcCall.Arguments
	if len(rawArgs) != builtin.MinArgs {
		return nil, &participle.ParseError{
			Pos: pos,
			Msg: fmt.Sprintf("bvhSdf requires exactly %d argument, got %d", builtin.MinArgs, len(rawArgs)),
		}
	}
	indexOp, err := buildMeshIndexArg(buildScope, rawArgs[0], pos, endPos)
	if err != nil {
		return nil, err
	}
	return program.NewFnCall(
		"", "bvhSdf", []*program.Operand{indexOp},
		posToLocation(pos), posToLocation(endPos),
	), nil
}

// buildMeshIndexArg builds the leading mesh-index argument of mesh / bvhSdf /
// vdbSdf. A non-negative integer literal is emitted directly; otherwise the
// argument is built as a normal expression — it must fold to a slot id later
// (e.g. a variable bound to an importStep/importMesh import). This is what lets
// `mesh(part, ...)` work where `part = importStep "f.stp"`.
func buildMeshIndexArg(buildScope scope.BuilderScope, expr *Expression, pos, endPos lexer.Position) (*program.Operand, error) {
	if meshIndex, ok := extractNonNegativeIntLiteral(expr); ok {
		return &program.Operand{
			Start: posToLocation(pos),
			End:   posToLocation(endPos),
			T:     &program.Operand_Value{Value: float64(meshIndex)},
		}, nil
	}
	ops, err := buildExpressions(buildScope, []*Expression{expr})
	if err != nil {
		return nil, err
	}
	return ops[0], nil
}

// buildVdbSdfCall handles vdbSdf(index [, voxelDivisor [, halfBandVoxels]])
// which returns the signed distance from the implicit position p to an
// imported mesh sampled from a pre-baked OpenVDB narrow-band level set.
//
//   - vdbSdf(index)                                 → defaults: voxelDivisor = 256,
//     halfBandVoxels = 3
//   - vdbSdf(index, voxelDivisor)                   → custom resolution, default band
//   - vdbSdf(index, voxelDivisor, halfBandVoxels)   → both custom
//
// `voxelDivisor` is the number of voxels along the mesh's longest bbox
// axis — the actual world-space cell size is mesh-bbox-longest /
// voxelDivisor. Higher values give sharper bakes at quadratic memory
// cost. All arguments must be compile-time literals (the canvas agent
// uses them as bake-cache keys before the program runs).
func buildVdbSdfCall(buildScope scope.BuilderScope, funcCall *FunctionCall, pos, endPos lexer.Position) (*program.Operand, error) {
	builtin := program.BuiltinNamed("vdbSdf")
	rawArgs := funcCall.Arguments
	if len(rawArgs) < builtin.MinArgs || len(rawArgs) > builtin.MaxArgs {
		return nil, &participle.ParseError{
			Pos: pos,
			Msg: fmt.Sprintf("vdbSdf requires between %d and %d arguments, got %d",
				builtin.MinArgs, builtin.MaxArgs, len(rawArgs)),
		}
	}
	indexOp, err := buildMeshIndexArg(buildScope, rawArgs[0], pos, endPos)
	if err != nil {
		return nil, err
	}
	args := []*program.Operand{indexOp}
	if len(rawArgs) >= 2 {
		divisor, ok := extractNonNegativeIntLiteral(rawArgs[1])
		if !ok || divisor <= 0 {
			return nil, &participle.ParseError{
				Pos: pos,
				Msg: "vdbSdf() second argument (voxelDivisor) must be a positive integer literal — voxels along the longest bbox axis",
			}
		}
		args = append(args, &program.Operand{
			Start: posToLocation(pos),
			End:   posToLocation(endPos),
			T:     &program.Operand_Value{Value: float64(divisor)},
		})
	}
	if len(rawArgs) >= 3 {
		halfBand, ok := extractPositiveFloatLiteral(rawArgs[2])
		if !ok {
			return nil, &participle.ParseError{
				Pos: pos,
				Msg: "vdbSdf() third argument (halfBandVoxels) must be a positive numeric literal",
			}
		}
		args = append(args, &program.Operand{
			Start: posToLocation(pos),
			End:   posToLocation(endPos),
			T:     &program.Operand_Value{Value: halfBand},
		})
	}
	return program.NewFnCall(
		"", "vdbSdf", args,
		posToLocation(pos), posToLocation(endPos),
	), nil
}
