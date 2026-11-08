package ir

import (
	"fmt"
	"io"
	"strings"
)

// Stmt represents a structured statement in the WGSL statement intermediate representation.
type Stmt interface {
	Reads() []string
	Writes() []string
	Format(w io.Writer, indent string) error
}

// LetStmt represents an immutable value declaration (`let <Name>[: <Type>] = <Expr>;`).
type LetStmt struct {
	// Read-only immutable fields set at construction
	Name  string
	Type  string
	Expr  string
	reads []string
}

// NewLetStmt creates an immutable Let statement.
func NewLetStmt(name, typeName, expr string, reads []string) *LetStmt {
	r := make([]string, len(reads))
	copy(r, reads)
	return &LetStmt{
		Name:  name,
		Type:  typeName,
		Expr:  expr,
		reads: r,
	}
}

func (s *LetStmt) Reads() []string {
	r := make([]string, len(s.reads))
	copy(r, s.reads)
	return r
}

func (s *LetStmt) Writes() []string {
	return []string{s.Name}
}

func (s *LetStmt) Format(w io.Writer, indent string) error {
	if s.Type != "" {
		_, err := fmt.Fprintf(w, "%slet %s: %s = %s;\n", indent, s.Name, s.Type, s.Expr)
		return err
	}
	_, err := fmt.Fprintf(w, "%slet %s = %s;\n", indent, s.Name, s.Expr)
	return err
}

// VarStmt represents a mutable local variable declaration (`var <Name>[: <Type>] = <InitExpr>;`).
type VarStmt struct {
	// Read-only immutable fields set at construction
	Name     string
	Type     string
	InitExpr string
	reads    []string
	writes   []string
}

// NewVarStmt creates a mutable variable declaration statement.
func NewVarStmt(name, typeName, initExpr string, reads []string) *VarStmt {
	r := make([]string, len(reads))
	copy(r, reads)
	return &VarStmt{
		Name:     name,
		Type:     typeName,
		InitExpr: initExpr,
		reads:    r,
		writes:   []string{name},
	}
}

func (s *VarStmt) Reads() []string {
	r := make([]string, len(s.reads))
	copy(r, s.reads)
	return r
}

func (s *VarStmt) Writes() []string {
	w := make([]string, len(s.writes))
	copy(w, s.writes)
	return w
}

func (s *VarStmt) Format(w io.Writer, indent string) error {
	if s.InitExpr != "" {
		if s.Type != "" {
			_, err := fmt.Fprintf(w, "%svar %s: %s = %s;\n", indent, s.Name, s.Type, s.InitExpr)
			return err
		}
		_, err := fmt.Fprintf(w, "%svar %s = %s;\n", indent, s.Name, s.InitExpr)
		return err
	}
	if s.Type != "" {
		_, err := fmt.Fprintf(w, "%svar %s : %s;\n", indent, s.Name, s.Type)
		return err
	}
	_, err := fmt.Fprintf(w, "%svar %s;\n", indent, s.Name)
	return err
}

// GuardStmt represents a conditional branch (`if (<Condition>) { <Body> }`).
type GuardStmt struct {
	// Read-only immutable fields set at construction
	Condition   string
	HeaderExtra string
	Body        *Block
	Inline      bool
	reads       []string
}

// NewGuardStmt creates a conditional branch statement.
func NewGuardStmt(condition string, body *Block, reads []string) *GuardStmt {
	return NewGuardStmtWithHeader(condition, "", body, reads)
}

// NewInlineGuardStmt creates a single-line conditional branch statement (`if (cond) { stmt; }`).
func NewInlineGuardStmt(condition string, body *Block, reads []string) *GuardStmt {
	g := NewGuardStmtWithHeader(condition, "", body, reads)
	g.Inline = true
	return g
}

// NewGuardStmtWithHeader creates a conditional branch statement with extra code on the opening line.
func NewGuardStmtWithHeader(condition, headerExtra string, body *Block, reads []string) *GuardStmt {
	r := make([]string, len(reads))
	copy(r, reads)
	if body == nil {
		body = NewBlock()
	}
	return &GuardStmt{
		Condition:   condition,
		HeaderExtra: headerExtra,
		Body:        body,
		reads:       r,
	}
}

// HeaderReads returns the identifiers read by the guard's condition and header.
func (s *GuardStmt) HeaderReads() []string {
	r := make([]string, len(s.reads))
	copy(r, s.reads)
	return r
}

// BodyReads returns the identifiers read inside the guard's body block.
func (s *GuardStmt) BodyReads() []string {
	return s.Body.Reads()
}

func (s *GuardStmt) Reads() []string {
	seen := make(map[string]bool)
	var all []string
	for _, id := range s.reads {
		if !seen[id] {
			seen[id] = true
			all = append(all, id)
		}
	}
	for _, id := range s.Body.Reads() {
		if !seen[id] {
			seen[id] = true
			all = append(all, id)
		}
	}
	return all
}

func (s *GuardStmt) Writes() []string {
	return s.Body.Writes()
}

func (s *GuardStmt) Format(w io.Writer, indent string) error {
	if s.Inline && len(s.Body.Statements()) == 1 {
		var singleBuf strings.Builder
		if err := s.Body.Statements()[0].Format(&singleBuf, ""); err != nil {
			return err
		}
		stmtStr := strings.TrimSpace(singleBuf.String())
		_, err := fmt.Fprintf(w, "%sif (%s) { %s }\n", indent, s.Condition, stmtStr)
		return err
	}

	if s.HeaderExtra != "" {
		if _, err := fmt.Fprintf(w, "%sif (%s) { %s\n", indent, s.Condition, s.HeaderExtra); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, "%sif (%s) {\n", indent, s.Condition); err != nil {
			return err
		}
	}
	if err := s.Body.Format(w, indent+"  "); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "%s}\n", indent)
	return err
}

// AssignStmt represents an assignment to an existing variable (`<Target> = <Expr>;`).
type AssignStmt struct {
	// Read-only immutable fields set at construction
	Target string
	Expr   string
	reads  []string
	writes []string
}

// NewAssignStmt creates an assignment statement.
func NewAssignStmt(target, expr string, reads []string) *AssignStmt {
	r := make([]string, len(reads))
	copy(r, reads)
	return &AssignStmt{
		Target: target,
		Expr:   expr,
		reads:  r,
		writes: []string{target},
	}
}

func (s *AssignStmt) Reads() []string {
	r := make([]string, len(s.reads))
	copy(r, s.reads)
	return r
}

func (s *AssignStmt) Writes() []string {
	w := make([]string, len(s.writes))
	copy(w, s.writes)
	return w
}

func (s *AssignStmt) Format(w io.Writer, indent string) error {
	_, err := fmt.Fprintf(w, "%s%s = %s;\n", indent, s.Target, s.Expr)
	return err
}

// ReturnStmt represents a function return (`return <Expr>;`).
type ReturnStmt struct {
	// Read-only immutable fields set at construction
	Expr  string
	reads []string
}

// NewReturnStmt creates a return statement.
func NewReturnStmt(expr string, reads []string) *ReturnStmt {
	r := make([]string, len(reads))
	copy(r, reads)
	return &ReturnStmt{
		Expr:  expr,
		reads: r,
	}
}

func (s *ReturnStmt) Reads() []string {
	r := make([]string, len(s.reads))
	copy(r, s.reads)
	return r
}

func (s *ReturnStmt) Writes() []string {
	return nil
}

func (s *ReturnStmt) Format(w io.Writer, indent string) error {
	_, err := fmt.Fprintf(w, "%sreturn %s;\n", indent, s.Expr)
	return err
}

// OpaqueStmt represents raw statements preserved as text lines with explicit read/write tracking.
type OpaqueStmt struct {
	// Read-only immutable fields set at construction
	Lines  []string
	reads  []string
	writes []string
}

// NewOpaqueStmt creates an opaque statement block.
func NewOpaqueStmt(lines []string, reads, writes []string) *OpaqueStmt {
	l := make([]string, len(lines))
	copy(l, lines)
	r := make([]string, len(reads))
	copy(r, reads)
	w := make([]string, len(writes))
	copy(w, writes)
	return &OpaqueStmt{
		Lines:  l,
		reads:  r,
		writes: w,
	}
}

func (s *OpaqueStmt) Reads() []string {
	r := make([]string, len(s.reads))
	copy(r, s.reads)
	return r
}

func (s *OpaqueStmt) Writes() []string {
	w := make([]string, len(s.writes))
	copy(w, s.writes)
	return w
}

func (s *OpaqueStmt) Format(w io.Writer, indent string) error {
	for _, l := range s.Lines {
		trimmed := strings.TrimSpace(l)
		if trimmed == "" {
			continue
		}
		if _, err := fmt.Fprintf(w, "%s%s\n", indent, trimmed); err != nil {
			return err
		}
	}
	return nil
}
