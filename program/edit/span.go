package edit

import "fmt"

// Span represents a single source code span
type Span struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn"`
	EndLine     int `json:"endLine"`
	EndColumn   int `json:"endColumn"`
}

// Equal returns true if the two spans represent the exact same source range.
func (s Span) Equal(other Span) bool {
	return s.StartLine == other.StartLine &&
		s.StartColumn == other.StartColumn &&
		s.EndLine == other.EndLine &&
		s.EndColumn == other.EndColumn
}

// VarName returns the parameter variable name for the span.
func (s Span) VarName() string {
	return fmt.Sprintf("gsl_param_%d_%d", s.StartLine, s.StartColumn)
}
