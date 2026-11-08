package ir

import "io"

// Block represents a sequence of statements within a function or control-flow body.
type Block struct {
	stmts []Stmt
}

// NewBlock constructs an empty statement block.
func NewBlock() *Block {
	return &Block{stmts: make([]Stmt, 0)}
}

// Statements returns a copy of the statements in the block.
func (b *Block) Statements() []Stmt {
	res := make([]Stmt, len(b.stmts))
	copy(res, b.stmts)
	return res
}

// Add appends a statement to the end of the block.
func (b *Block) Add(s Stmt) {
	b.stmts = append(b.stmts, s)
}

// Insert inserts a statement at the specified index.
func (b *Block) Insert(idx int, s Stmt) {
	b.stmts = append(b.stmts[:idx], append([]Stmt{s}, b.stmts[idx:]...)...)
}

// Remove removes the statement at the specified index.
func (b *Block) Remove(idx int) {
	b.stmts = append(b.stmts[:idx], b.stmts[idx+1:]...)
}

// Clear removes all statements from the block.
func (b *Block) Clear() {
	b.stmts = b.stmts[:0]
}

// Len returns the number of statements in the block.
func (b *Block) Len() int {
	return len(b.stmts)
}

// Reads returns the external identifiers read by statements in the block before being defined locally.
func (b *Block) Reads() []string {
	defined := make(map[string]bool)
	seenReads := make(map[string]bool)
	var externalReads []string

	for _, s := range b.stmts {
		for _, r := range s.Reads() {
			if !defined[r] && !seenReads[r] {
				seenReads[r] = true
				externalReads = append(externalReads, r)
			}
		}
		for _, w := range s.Writes() {
			defined[w] = true
		}
	}
	return externalReads
}

// Writes returns the unique identifiers mutated or declared within the block.
func (b *Block) Writes() []string {
	seenWrites := make(map[string]bool)
	var writes []string
	for _, s := range b.stmts {
		for _, w := range s.Writes() {
			if !seenWrites[w] {
				seenWrites[w] = true
				writes = append(writes, w)
			}
		}
	}
	return writes
}

// Format renders each statement in the block indented by indent.
func (b *Block) Format(w io.Writer, indent string) error {
	for _, s := range b.stmts {
		if err := s.Format(w, indent); err != nil {
			return err
		}
	}
	return nil
}
