package ir

import (
	"fmt"
	"io"
)

// LoopStmt represents a counted loop (`for (var <Index> = 0; <Index> < <Count>; ...) { <Body> }`).
type LoopStmt struct {
	// Read-only immutable fields set at construction
	Index string
	Count int
	Body  *Block
}

// NewLoopStmt creates a loop running body count times, its index counting up
// from 0.
func NewLoopStmt(index string, count int, body *Block) *LoopStmt {
	if body == nil {
		body = NewBlock()
	}
	return &LoopStmt{Index: index, Count: count, Body: body}
}

// Reads returns what the body reads from outside the loop; the index is the
// loop's own.
func (s *LoopStmt) Reads() []string {
	var r []string
	for _, id := range s.Body.Reads() {
		if id != s.Index {
			r = append(r, id)
		}
	}
	return r
}

func (s *LoopStmt) Writes() []string {
	return s.Body.Writes()
}

func (s *LoopStmt) Format(w io.Writer, indent string) error {
	if _, err := fmt.Fprintf(w, "%sfor (var %[2]s = 0; %[2]s < %[3]d; %[2]s = %[2]s + 1) {\n", indent, s.Index, s.Count); err != nil {
		return err
	}
	if err := s.Body.Format(w, indent+"  "); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "%s}\n", indent)
	return err
}
