package program

import "fmt"

// OperandVisitor defines the visitor interface for the Operand oneof field.
type OperandVisitor[T any] interface {
	VisitValue(*Operand, float64) (T, error)
	VisitReference(*Operand, *Reference) (T, error)
	VisitExpression(*Operand, *Expression) (T, error)
	VisitFnCall(*Operand, *FnCall) (T, error)
	VisitPropRef(*Operand, *PropertyRef) (T, error)
}

// VisitOperand dispatches the call to the appropriate visitor method.
func VisitOperand[T any](op *Operand, v OperandVisitor[T]) (T, error) {
	if op == nil || op.T == nil {
		var zero T
		return zero, nil
	}
	switch t := op.T.(type) {
	case *Operand_Value:
		return v.VisitValue(op, t.Value)
	case *Operand_Reference:
		return v.VisitReference(op, t.Reference)
	case *Operand_Expression:
		return v.VisitExpression(op, t.Expression)
	case *Operand_FnCall:
		return v.VisitFnCall(op, t.FnCall)
	case *Operand_PropRef:
		return v.VisitPropRef(op, t.PropRef)
	default:
		panic(fmt.Sprintf("unknown operand type: %T", t))
	}
}
