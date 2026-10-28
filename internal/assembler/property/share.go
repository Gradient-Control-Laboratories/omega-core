package property

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

// Writing down the expansion's own sharing.
//
// The engine builds a result as a DAG: a primal named again in its tangent, a
// remap's coordinates named once per mention of p, a vector's components
// named in each axis. In memory that sharing costs nothing. Nothing after
// Expand keeps it: proto.Clone copies a DAG out as a tree, and every rewriting
// pass rebuilds a node once per path that reaches it. Ten nested rotations,
// differentiated, are 1,152 distinct nodes and a tree of 7.9 billion — the
// tree is what the rest of the compiler walks, and a gear profile, five
// remaps deep, never finished.
//
// A named value, by contrast, stays shared all the way through: a chain of
// forty locals, each naming the one before it twice, compiles in milliseconds.
// So a node the result reaches more than once is written once, as a local,
// and every other mention refers to it by name. The expression computed is
// the same one; only its spelling changes.

// shareMinSize is the smallest node worth a name. A shared node smaller than
// this — p.x, c·s — is written out at each mention instead: its size counts a
// named child as one, so every node left inline is under shareMinSize and the
// written-out result stays within shareMinSize times the DAG's edges. Naming
// everything shared costs more than it saves: a function of thousands of
// locals is a shape the passes downstream were never timed on, and two of
// them (constant folding, the inliner's reuse count) turned out quadratic in
// it.
var shareMinSize = 8

// shareNames hands out the names of shared locals and of the functions that
// hold them, unique across a package: generated names use the double
// underscore the companions already do, and anything the package itself
// declares is skipped.
type shareNames struct {
	next  int
	taken func(name string) bool
}

func (n *shareNames) fresh(prefix string) string {
	for {
		name := fmt.Sprintf("%s__%d", prefix, n.next)
		n.next++
		if n.taken == nil || !n.taken(name) {
			return name
		}
	}
}

// shareChildren is the operands of op that are evaluated where op is. A
// remap's target is not: it runs with p rebound, so a node under it cannot be
// hoisted out to where p is the caller's, and it is neither counted nor
// rewritten. A swizzle's selectors are syntax, not values.
func shareChildren(op *program.Operand) []*program.Operand {
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		e := t.Expression
		switch e.Operation {
		case program.Operation_REMAP:
			if len(e.Operands) == 2 {
				return e.Operands[1:]
			}
			return nil
		case program.Operation_ACCESS:
			if len(e.Operands) > 0 {
				return e.Operands[:1]
			}
			return nil
		}
		return e.Operands
	case *program.Operand_FnCall:
		return t.FnCall.Arguments
	}
	return nil
}

// allChildren is every operand of op, whether or not it is evaluated where op
// is — what a copy of op carries along.
func allChildren(op *program.Operand) []*program.Operand {
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		return t.Expression.Operands
	case *program.Operand_FnCall:
		return t.FnCall.Arguments
	}
	return nil
}

// sharedLocals rewrites root so each node it reaches more than once, and is
// big enough to be worth it, is computed once: a local named by names, listed
// before any local that uses it. When nothing qualifies, root comes back
// untouched and there are no locals.
func sharedLocals(root *program.Operand, names *shareNames) (*program.Operand, []*program.Assignment) {
	counts := map[*program.Operand]int{}
	var count func(op *program.Operand)
	count = func(op *program.Operand) {
		if op == nil {
			return
		}
		counts[op]++
		if counts[op] > 1 {
			return
		}
		for _, c := range shareChildren(op) {
			count(c)
		}
	}
	count(root)

	// The size a node is written out at, capped at shareMinSize: a child
	// that will be named costs one, and a child evaluated elsewhere (a
	// remap's target) costs what it would take to copy it.
	evaluated := func(op, c *program.Operand) bool {
		for _, e := range shareChildren(op) {
			if e == c {
				return true
			}
		}
		return false
	}
	named := map[*program.Operand]bool{}
	sizes := map[*program.Operand]int{}
	var size func(op *program.Operand) int
	size = func(op *program.Operand) int {
		if op == nil {
			return 0
		}
		if s, ok := sizes[op]; ok {
			return s
		}
		s := 1
		for _, c := range allChildren(op) {
			if s >= shareMinSize {
				break
			}
			if evaluated(op, c) && named[c] {
				s++
			} else {
				s += size(c)
			}
		}
		if s > shareMinSize {
			s = shareMinSize
		}
		sizes[op] = s
		if counts[op] >= 2 && s >= shareMinSize {
			switch op.T.(type) {
			case *program.Operand_Expression, *program.Operand_FnCall:
				named[op] = true
			}
		}
		return s
	}
	var decide func(op *program.Operand)
	seen := map[*program.Operand]bool{}
	decide = func(op *program.Operand) {
		if op == nil || seen[op] {
			return
		}
		seen[op] = true
		for _, c := range shareChildren(op) {
			decide(c)
		}
		size(op)
	}
	decide(root)
	if len(named) == 0 {
		return root, nil
	}

	nameOf := map[*program.Operand]string{}
	var locals []*program.Assignment
	var build func(op *program.Operand) *program.Operand
	build = func(op *program.Operand) *program.Operand {
		if op == nil {
			return nil
		}
		if name, ok := nameOf[op]; ok {
			return program.NewReference("", name, nil, nil)
		}
		out := op
		switch t := op.T.(type) {
		case *program.Operand_Expression:
			e := t.Expression
			ops := make([]*program.Operand, len(e.Operands))
			copy(ops, e.Operands)
			switch e.Operation {
			case program.Operation_REMAP:
				if len(ops) == 2 {
					ops[1] = build(ops[1])
				}
			case program.Operation_ACCESS:
				if len(ops) > 0 {
					ops[0] = build(ops[0])
				}
			default:
				for i, c := range ops {
					ops[i] = build(c)
				}
			}
			out = &program.Operand{Start: op.Start, End: op.End, T: &program.Operand_Expression{
				Expression: &program.Expression{Operation: e.Operation, Operands: ops},
			}}
		case *program.Operand_FnCall:
			fc := t.FnCall
			args := make([]*program.Operand, len(fc.Arguments))
			for i, a := range fc.Arguments {
				args[i] = build(a)
			}
			out = &program.Operand{Start: op.Start, End: op.End, T: &program.Operand_FnCall{
				FnCall: &program.FnCall{Reference: fc.Reference, Arguments: args},
			}}
		}
		if named[op] {
			name := names.fresh("share")
			nameOf[op] = name
			locals = append(locals, program.NewAssignment([]string{name}, out, nil, nil))
			return program.NewReference("", name, nil, nil)
		}
		return out
	}
	return build(root), locals
}
