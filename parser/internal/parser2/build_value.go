package parser2

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/alecthomas/participle/v2"
	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/program"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

func buildValue(buildScope scope.BuilderScope, value Value) (*program.Operand, error) {
	if value == nil {
		return nil, errors.New("missing value")
	}

	switch v := value.(type) {
	case *Number:
		start := posToLocation(v.Pos)
		end := posToLocation(v.EndPos)
		value := *v.Value
		return program.NewValue(value, start, end), nil

	case *ImportRef:
		return buildImportRef(v)

	case *Reference:
		name := v.Name
		start := posToLocation(v.Pos)
		end := posToLocation(v.EndPos)
		return program.NewReference("", name, start, end), nil

	case *QualifiedRef:
		return buildQualifiedRef(buildScope, v)

	case *FunctionCall:
		return buildFunctionCall(buildScope, v)

	case *SubExpression:
		return buildExpression(buildScope, v.Expression)

	case *AnonCall:
		if v.Arguments == nil {
			decl, err := buildAnonFunction(buildScope, v)
			if err != nil {
				return nil, err
			}
			return program.NewReference(
				"", string(decl.Name),
				posToLocation(v.Pos), posToLocation(v.EndPos),
			), nil
		}
		return buildAnonFunctionCall(buildScope, v)

	case *Collection:
		return buildCollection(buildScope, v.Expressions, v.Pos, v.EndPos)

	case *String:
		unquoted, err := strconv.Unquote(v.Value)
		if err != nil {
			return nil, &participle.ParseError{
				Pos: v.Pos,
				Msg: err.Error(),
			}
		}
		if len(unquoted) == 0 {
			return nil, &participle.ParseError{
				Pos: v.Pos,
				Msg: "empty strings are not allowed",
			}
		}
		runes := []rune(unquoted)
		if len(runes) == 1 {
			return program.NewValue(
				float64(runes[0]),
				posToLocation(v.Pos), posToLocation(v.EndPos),
			), nil
		}
		operands := make([]*program.Operand, len(runes))
		for i, r := range runes {
			operands[i] = program.NewValue(
				float64(r),
				posToLocation(v.Pos), posToLocation(v.EndPos),
			)
		}
		return program.NewExpression(
			program.Operation_COLLECT, operands,
			posToLocation(v.Pos), posToLocation(v.EndPos),
		), nil

	case *AnnotatedExpr:
		return buildAnnotatedExpr(buildScope, v)

	default:
		return nil, fmt.Errorf("unhandled value type: %T", v)
	}
}

// buildImportRef lowers the expression form `importStep "path"` / `importMesh
// "path"` to a FunctionCall named `importStep`/`importMesh` whose single
// argument carries the path as a rune list (the same representation strings use
// elsewhere). The whole-module slot-assignment pass later reconstructs the path
// and folds the call to a compiler-assigned slot id.
func buildImportRef(v *ImportRef) (*program.Operand, error) {
	unquoted, err := strconv.Unquote(v.Path)
	if err != nil {
		return nil, &participle.ParseError{Pos: v.Pos, Msg: err.Error()}
	}
	if len(unquoted) == 0 {
		return nil, &participle.ParseError{Pos: v.Pos, Msg: "import path must not be empty"}
	}
	start := posToLocation(v.Pos)
	end := posToLocation(v.EndPos)
	runes := []rune(unquoted)
	pathOps := make([]*program.Operand, len(runes))
	for i, r := range runes {
		pathOps[i] = program.NewValue(float64(r), start, end)
	}
	var pathOp *program.Operand
	if len(pathOps) == 1 {
		pathOp = pathOps[0]
	} else {
		pathOp = program.NewExpression(program.Operation_COLLECT, pathOps, start, end)
	}
	return program.NewFnCall("", v.Kind, []*program.Operand{pathOp}, start, end), nil
}

func buildQualifiedRef(buildScope scope.BuilderScope, v *QualifiedRef) (*program.Operand, error) {
	start := posToLocation(v.Pos)

	if v.Name == program.DTimeAccessor || v.Name == program.DPAccessor {
		// Forward-mode derivative: .dtime is d/d(time), .dp is the spatial
		// gradient d/d(p). Expanded by the autodiff pass after conversion to
		// the model. The target is resolved downstream, like any reference.
		refEnd := program.NewLoc(
			start.Line,
			start.Column+int32(len(v.Prefix)),
			start.Offset+int32(len(v.Prefix)),
		)
		target := program.NewReference("", v.Prefix, start, refEnd)
		return newDeriv(target, v.Name, v.EndPos), nil
	}

	if v.Name == program.SecondAccessor {
		// Second-closest primitive-surface distance (name.second). Expanded by
		// the self-computing-property pass. The target is a plain reference.
		refEnd := program.NewLoc(
			start.Line,
			start.Column+int32(len(v.Prefix)),
			start.Offset+int32(len(v.Prefix)),
		)
		target := program.NewReference("", v.Prefix, start, refEnd)
		return newSecond(target, v.EndPos), nil
	}

	if v.Name == program.BoxAccessor {
		// Bounding box of a field (name.box), unless the prefix names an
		// import, where pkg.box is that package's value. Expanded by the
		// self-computing-property pass; the target is a plain reference.
		if _, isImport := buildScope.NamedElem(v.Prefix).(*scope.PkgRef); !isImport {
			refEnd := program.NewLoc(
				start.Line,
				start.Column+int32(len(v.Prefix)),
				start.Offset+int32(len(v.Prefix)),
			)
			target := program.NewReference("", v.Prefix, start, refEnd)
			return newBox(target, v.EndPos), nil
		}
	}

	if v.Name == program.UAccessor || v.Name == program.VAccessor {
		// Surface frame coordinate of a field (name.u, name.v), unless the
		// prefix names an import, where pkg.u is that package's value, as with
		// .box. Expanded by the self-computing-property pass; the target is a
		// plain reference.
		if _, isImport := buildScope.NamedElem(v.Prefix).(*scope.PkgRef); !isImport {
			refEnd := program.NewLoc(
				start.Line,
				start.Column+int32(len(v.Prefix)),
				start.Offset+int32(len(v.Prefix)),
			)
			target := program.NewReference("", v.Prefix, start, refEnd)
			return newSurfaceFrame(target, v.Name, v.EndPos), nil
		}
	}

	elem := buildScope.NamedElem(v.Prefix)
	isLocal := false
	if elem != nil {
		if _, ok := elem.(*scope.Var); ok {
			isLocal = true
		} else if _, ok := elem.(*scope.Param); ok {
			isLocal = true
		}
	}

	isImport := false
	if !isLocal {
		if _, ok := elem.(*scope.PkgRef); ok {
			isImport = true
		} else {
			_, isImport = gsllibs.PathForPrefix(v.Prefix)
		}
	}

	if !isImport {
		isAccessor, err := IsAccessor(v.Name)
		if err != nil {
			return nil, &participle.ParseError{
				Pos: v.Pos,
				Msg: err.Error(),
			}
		}
		if isAccessor {
			name := v.Prefix
			accessor := v.Name
			refEnd := program.NewLoc(
				start.Line,
				start.Column+int32(len(name)),
				start.Offset+int32(len(name)),
			)
			target := program.NewReference("", name, start, refEnd)
			return program.NewAccess(target, accessor, posToLocation(v.EndPos)), nil
		}
	}

	if elem != nil {
		if vVar, ok := elem.(*scope.Var); ok && vVar != nil {
			return program.NewPropertyRef(vVar.Var.Name, v.Name, start, posToLocation(v.EndPos)), nil
		}
		if vParam, ok := elem.(*scope.Param); ok && vParam != nil {
			return program.NewPropertyRef(vParam.Param.Name, v.Name, start, posToLocation(v.EndPos)), nil
		}
	}

	return program.NewReference(v.Prefix, v.Name, start, posToLocation(v.EndPos)), nil
}
