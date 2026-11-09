package program

import (
	"testing"
	"time"

	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

// sharedChain builds an operand graph `depth` levels deep in which every level
// is one expression whose two operands are the SAME pointer. That is what the
// assembler's optimizer produces after inlining and common-subexpression
// elimination: a DAG, not a tree. A walk that follows every path instead of
// every node visits 2^depth operands here, which is why a large optimized
// document could hold the auto-import pass for hours.
func sharedChain(depth int) *Operand {
	loc := &Location{Line: 1, Column: 1}
	op := NewFnCall("color", "palette", nil, loc, loc)
	for i := 0; i < depth; i++ {
		op = NewExpression(Operation_ADD, []*Operand{op, op}, loc, loc)
	}
	return op
}

func TestUpdateImportsWalksASharedOperandOnce(t *testing.T) {
	if _, ok := gsllibs.PathForPrefix("color"); !ok {
		t.Skip("the color package is not registered for auto-import")
	}
	pkg := &Package{
		Path: "dag_test",
		Files: []*PackageFile{{
			Returns: []*Operand{sharedChain(40)},
			Assignments: []*Assignment{{
				Expression: sharedChain(40),
				Variables:  []*Variable{{Name: "a"}},
			}},
		}},
	}

	done := make(chan struct{})
	go func() {
		UpdateImports(pkg, "dag_test")
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("UpdateImports did not finish on a 40-deep shared operand graph: " +
			"the collector is following paths rather than visiting nodes")
	}

	var paths []string
	for _, imp := range pkg.Files[0].Imports {
		paths = append(paths, imp.Path)
	}
	found := false
	for _, p := range paths {
		if p == "color" {
			found = true
		}
	}
	if !found {
		t.Errorf("the shared node's package should still be imported, got %v", paths)
	}
}

// Two files sharing one operand must each get the import: the memo resets per
// file, so skipping a repeat never costs a file its own import line.
func TestUpdateImportsImportsPerFileWhenFilesShareAnOperand(t *testing.T) {
	if _, ok := gsllibs.PathForPrefix("color"); !ok {
		t.Skip("the color package is not registered for auto-import")
	}
	shared := sharedChain(2)
	pkg := &Package{
		Path: "dag_test",
		Files: []*PackageFile{
			{Returns: []*Operand{shared}},
			{Returns: []*Operand{shared}},
		},
	}
	UpdateImports(pkg, "dag_test")

	for i, file := range pkg.Files {
		found := false
		for _, imp := range file.Imports {
			if imp.Path == "color" {
				found = true
			}
		}
		if !found {
			t.Errorf("file %d lost its color import", i)
		}
	}
}
