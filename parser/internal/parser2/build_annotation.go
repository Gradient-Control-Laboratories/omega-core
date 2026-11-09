package parser2

import (
	"fmt"

	participle "github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/parser/internal/annotation"
	"gradientcontrol.com/omega/core/program"
)

func buildAnnotatedDeclaration(buildScope scope.BuilderScope, decl *AnnotatedDeclaration) error {
	// Invariant: decl.Annotation is guaranteed non-nil by parser grammar ("@@").
	annotName := decl.Annotation.Name

	// The recorded annotations come first, because they are the only ones that
	// can mark a function declaration and so cannot pass the assignment
	// assertion below. Nothing is lowered for them: they are noted on the file
	// and the declaration is built exactly as it would be unannotated.
	if annotName == annotation.Material || annotName == annotation.Finish || annotName == annotation.Surface {
		return buildRecordedAnnotation(buildScope, annotName, decl)
	}

	assign, ok := decl.Declaration.(*Assignment)
	if !ok {
		return &participle.ParseError{
			Pos: decl.Pos,
			Msg: fmt.Sprintf("@%s annotation can only be applied to variable assignments", annotName),
		}
	}
	if err := annotation.ValidateTarget(annotName, annotation.TargetAssignment); err != nil {
		return &participle.ParseError{
			Pos: decl.Pos,
			Msg: err.Error(),
		}
	}
	if err := annotation.ValidateArgCount(annotName, len(decl.Annotation.Args)); err != nil {
		return &participle.ParseError{
			Pos: decl.Annotation.Pos,
			Msg: err.Error(),
		}
	}

	if annotName == annotation.Range {
		rng, err := buildRangeAnnotation(buildScope, decl.Annotation.Pos, decl.Annotation.EndPos, decl.Annotation.Args)
		if err != nil {
			return err
		}
		return buildAssignment(buildScope, assign, rng)
	}

	targetFn := annotation.TargetBuiltin(annotName)
	if targetFn == "" {
		return &participle.ParseError{
			Pos: decl.Pos,
			Msg: fmt.Sprintf("unsupported annotation @%s", annotName),
		}
	}

	targetExpr := assign.Expression
	if annotation.RequiresElementPromotion(annotName) {
		if col := getCollectionFromExpression(targetExpr); col != nil {
			promoteSplineElements(col)
		}
	}

	wrappedExpr := wrapInFunctionCall(targetFn, targetExpr, annotationArgs(decl.Annotation)...)
	wrappedAssign := &Assignment{
		Pos:        assign.Pos,
		Names:      assign.Names,
		Expression: wrappedExpr,
		EndPos:     assign.EndPos,
	}

	return buildAssignment(buildScope, wrappedAssign, nil)
}

func buildAnnotatedExpr(buildScope scope.BuilderScope, v *AnnotatedExpr) (*program.Operand, error) {
	// Invariant: v.Annotation is guaranteed non-nil by parser grammar ("@@").
	annotName := v.Annotation.Name
	if err := annotation.ValidateTarget(annotName, annotation.TargetExpression); err != nil {
		return nil, &participle.ParseError{
			Pos: v.Pos,
			Msg: fmt.Sprintf("unsupported expression annotation @%s", annotName),
		}
	}
	if err := annotation.ValidateArgCount(annotName, len(v.Annotation.Args)); err != nil {
		return nil, &participle.ParseError{
			Pos: v.Annotation.Pos,
			Msg: err.Error(),
		}
	}

	if annotName == annotation.Range {
		return buildExpression(buildScope, v.Expression)
	}

	targetFn := annotation.TargetBuiltin(annotName)
	if targetFn == "" {
		return nil, &participle.ParseError{
			Pos: v.Pos,
			Msg: fmt.Sprintf("unsupported expression annotation @%s", annotName),
		}
	}

	targetExpr := v.Expression
	if annotation.RequiresElementPromotion(annotName) {
		if col := getCollectionFromExpression(targetExpr); col != nil {
			promoteSplineElements(col)
		}
	}

	wrapped := wrapInFunctionCall(targetFn, targetExpr, annotationArgs(v.Annotation)...)
	return buildExpression(buildScope, wrapped)
}

// annotationArgs are an annotation's own arguments, which follow the
// annotated expression in the builtin's call: @direction(anchor) x lowers to
// uiDirection(x, anchor).
func annotationArgs(a *AnnotationDirective) []*Expression {
	var out []*Expression
	for _, arg := range a.Args {
		out = append(out, arg.Expr)
	}
	return out
}

func wrapInFunctionCall(targetFn string, expr *Expression, extra ...*Expression) *Expression {
	return &Expression{
		Pos: expr.Pos,
		Value: &LogicOr{
			Left: &LogicAnd{
				Left: &Equality{
					Left: &Comparison{
						Left: &Terms{
							Left: &Term{
								Left: &Factor{
									Pos: expr.Pos,
									Value: &FunctionCall{
										Pos:       expr.Pos,
										Name:      targetFn,
										Arguments: append([]*Expression{expr}, extra...),
										EndPos:    expr.Pos,
									},
									EndPos: expr.Pos,
								},
							},
						},
					},
				},
			},
		},
	}
}

func getCollectionFromExpression(expr *Expression) *Collection {
	if expr == nil || expr.Value == nil || expr.Value.Left == nil {
		return nil
	}
	logicAnd := expr.Value.Left
	if logicAnd.Left == nil || logicAnd.Left.Left == nil || logicAnd.Left.Left.Left == nil ||
		logicAnd.Left.Left.Left.Left == nil || logicAnd.Left.Left.Left.Left.Left == nil {
		return nil
	}
	factor := logicAnd.Left.Left.Left.Left.Left
	if factor == nil {
		return nil
	}
	col, _ := factor.Value.(*Collection)
	return col
}

func isUiTransformCall(expr *Expression) bool {
	if expr == nil || expr.Value == nil || expr.Value.Left == nil {
		return false
	}
	logicAnd := expr.Value.Left
	if logicAnd.Left == nil || logicAnd.Left.Left == nil || logicAnd.Left.Left.Left == nil ||
		logicAnd.Left.Left.Left.Left == nil || logicAnd.Left.Left.Left.Left.Left == nil {
		return false
	}
	factor := logicAnd.Left.Left.Left.Left.Left
	if factor == nil {
		return false
	}
	fc, ok := factor.Value.(*FunctionCall)
	return ok && fc != nil && fc.Name == "uiTransform"
}

func isReference(expr *Expression) bool {
	_, ok := plainReference(expr)
	return ok
}

// plainReference is the name an expression consists of, when it is a bare
// identifier and nothing else: no sign, no field, no index, no operator.
func plainReference(expr *Expression) (string, bool) {
	if expr == nil || expr.Value == nil || expr.Value.Left == nil || len(expr.Value.Right) > 0 {
		return "", false
	}
	logicAnd := expr.Value.Left
	if len(logicAnd.Right) > 0 || logicAnd.Left == nil || logicAnd.Left.Right != nil ||
		logicAnd.Left.Left == nil || logicAnd.Left.Left.Right != nil ||
		logicAnd.Left.Left.Left == nil || len(logicAnd.Left.Left.Left.Right) > 0 ||
		logicAnd.Left.Left.Left.Left == nil || len(logicAnd.Left.Left.Left.Left.Right) > 0 {
		return "", false
	}
	factor := logicAnd.Left.Left.Left.Left.Left
	if factor == nil || factor.Not || factor.Negative || factor.Access != nil || factor.Index != nil {
		return "", false
	}
	ref, ok := factor.Value.(*Reference)
	if !ok || ref == nil {
		return "", false
	}
	return ref.Name, true
}

func promoteSplineElements(col *Collection) {
	if col == nil {
		return
	}
	for i, expr := range col.Expressions {
		if isUiTransformCall(expr) || isReference(expr) {
			continue
		}
		col.Expressions[i] = wrapInFunctionCall("uiTransform", expr)
	}
}

func buildRangeAnnotation(
	buildScope scope.BuilderScope,
	startPos, endPos lexer.Position,
	args []*AnnotationArg,
) (*program.Range, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, &participle.ParseError{
			Pos: startPos,
			Msg: "@range annotation requires 2 or 3 arguments: @range(min, max[, period])",
		}
	}

	minOp, err := buildExpression(buildScope, args[0].Expr)
	if err != nil {
		return nil, err
	}

	maxOp, err := buildExpression(buildScope, args[1].Expr)
	if err != nil {
		return nil, err
	}

	var periodOp *program.Operand
	if len(args) == 3 {
		periodOp, err = buildExpression(buildScope, args[2].Expr)
		if err != nil {
			return nil, err
		}
	}

	return &program.Range{
		Start:  posToLocation(startPos),
		End:    posToLocation(endPos),
		Min:    minOp,
		Max:    maxOp,
		Period: periodOp,
	}, nil
}

// buildRecordedAnnotation notes the annotation on the file, then builds the
// declaration exactly as it would an unannotated one: a recorded annotation
// never changes what its symbol compiles to.
func buildRecordedAnnotation(buildScope scope.BuilderScope, kind string, decl *AnnotatedDeclaration) error {
	symbol, symbolPos, target, err := annotatedSymbol(kind, decl)
	if err != nil {
		return err
	}
	if err := annotation.ValidateTarget(kind, target); err != nil {
		return &participle.ParseError{Pos: decl.Pos, Msg: err.Error()}
	}
	if err := annotation.ValidateArgCount(kind, len(decl.Annotation.Args)); err != nil {
		return &participle.ParseError{Pos: decl.Annotation.Pos, Msg: err.Error()}
	}
	end := posToLocation(symbolPos)
	end.Column += int32(len(symbol))
	end.Offset += int32(len(symbol))
	if err := buildScope.AddAnnotation(kind, symbol, posToLocation(decl.Pos), end); err != nil {
		return &participle.ParseError{Pos: decl.Pos, Msg: err.Error()}
	}
	return buildDeclaration(buildScope, decl.Declaration)
}

// annotatedSymbol is the name a recorded annotation marks, where that name
// sits, and which target it is. An annotation marks a name rather than an
// arity, so every overload of an annotated function is covered by the one
// annotation.
func annotatedSymbol(kind string, decl *AnnotatedDeclaration) (string, lexer.Position, annotation.Target, error) {
	switch inner := decl.Declaration.(type) {
	case *Assignment:
		if len(inner.Names.Names) != 1 {
			return "", decl.Pos, annotation.TargetAssignment, &participle.ParseError{
				Pos: decl.Pos,
				Msg: fmt.Sprintf("@%s annotates a single name, got %d",
					kind, len(inner.Names.Names)),
			}
		}
		return inner.Names.Names[0].Name, inner.Names.Names[0].Pos, annotation.TargetAssignment, nil
	case *NamedFunction:
		return string(inner.Name), inner.Pos, annotation.TargetFunction, nil
	default:
		return "", decl.Pos, annotation.TargetExpression, &participle.ParseError{
			Pos: decl.Pos,
			Msg: fmt.Sprintf("@%s can only annotate a variable assignment or a function declaration",
				kind),
		}
	}
}
