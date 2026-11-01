package program

import "strings"

// OperandName is a human-readable name for a return operand: the variable it
// names, else the function or property it calls, else its text in source (the
// file that holds it). It is what the Bodies panel and the exports call a
// return, and "" when there is nothing to go on.
func OperandName(op *Operand, source string) string {
	if op == nil {
		return ""
	}
	if ref := op.GetReference(); ref != nil && ref.Name != "" {
		return ref.Name
	}
	if fc := op.GetFnCall(); fc != nil && fc.Reference != nil && fc.Reference.Name != "" {
		return fc.Reference.Name
	}
	if prop := op.GetPropRef(); prop != nil && prop.Reference != nil && prop.Reference.Name != "" {
		return prop.Reference.Name
	}
	// Fall back to the source text between the operand's start and end offsets
	if op.Start != nil && op.End != nil {
		start := int(op.Start.Offset)
		end := int(op.End.Offset)
		if start >= 0 && end > start && end <= len(source) {
			return strings.TrimSpace(source[start:end])
		}
	}
	return ""
}
