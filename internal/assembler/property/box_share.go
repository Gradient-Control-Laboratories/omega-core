package property

import (
	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/program"
)

// SHARING WHAT A PART'S BOX REPEATS. The box domain builds a box's corners as
// trees, and a tree holds every copy of what it reads: a profile's bounds are
// a min and a max over all its points, read again in an arc profile's bulge,
// and a rule such as swept reads one bound in several corners. MarkParts
// writes the six corners into the part's shrinkToVanish, and the codegen writes
// each out whole, evaluated per point for the viewer's culling. So before a
// part is wrapped, each subtree its corners repeat is declared once, as a
// value of its own just before the part, and read by name:
//
//	part = shrinkToVanish(shape, [lo(b(c)), …], [hi(b(c)), …], 1)
//	becomes
//	gslBox = b(c);   part = shrinkToVanish(shape, [lo(gslBox), …], [hi(gslBox), …], 1)
//
// It is a rewrite of the corners' trees and nothing else, so the box is the
// same box; inlining the declared values back gives the corners as they were.
// A subtree is shared only when it prints to at least boxShareMinLength
// characters: a short one, a reference or a sum of two, costs more as a
// declaration than its copies do. Among those, the one that saves the most,
// its copies past the first times its length, goes first; a later one can
// read an earlier one's name, and the values are declared in the order they
// read each other.
//
// Measured 2026-11-18 (backlog M17): profile_cylinder.gsl's WGSL from 190 KB
// to 83 KB, about half its time per point on the GPU; of the example goldens,
// CoreXY Printer's alone moves, 0.5% smaller.

// boxShareMinLength is the least printed length of a subtree that is shared.
const boxShareMinLength = 48

// shareBoxes turns the sharing off, for tests that compare against it.
var shareBoxes = true

// shareBoxRepeats is the declarations of what bx's corners repeat, in the
// order they read each other, and bx with its corners reading them; names are
// drawn fresh from taken. A box with nothing to share comes back as it was,
// with no declarations.
func shareBoxRepeats(bx *bbox, taken map[string]bool) ([]*program.Assignment, *bbox) {
	if !shareBoxes || bx == nil || bx.empty {
		return nil, bx
	}
	corners := make([]*program.Operand, 0, 6)
	for _, c := range append(bx.lo[:], bx.hi[:]...) {
		if c == nil {
			return nil, bx
		}
		corners = append(corners, proto.Clone(c).(*program.Operand))
	}
	var shared []*program.Assignment
	for {
		trees := append([]*program.Operand{}, corners...)
		for _, s := range shared {
			trees = append(trees, s.Expression)
		}
		key, found := mostSavedRepeat(trees)
		if !found {
			break
		}
		name := freshName("gslBox", taken)
		var expr *program.Operand
		for i := range corners {
			corners[i] = replaceSubtree(corners[i], key, name, &expr)
		}
		for _, s := range shared {
			s.Expression = replaceSubtree(s.Expression, key, name, &expr)
		}
		shared = append(shared, &program.Assignment{Variables: []*program.Variable{{Name: name}}, Expression: expr})
	}
	if len(shared) == 0 {
		return nil, bx
	}
	out := &bbox{}
	copy(out.lo[:], corners[:3])
	copy(out.hi[:], corners[3:])
	return inReadOrder(shared), out
}

// subtreeKey identifies a subtree by its encoding, so equal trees share one.
func subtreeKey(op *program.Operand) string {
	b, _ := proto.MarshalOptions{Deterministic: true}.Marshal(op)
	return string(b)
}

// shareable reports whether op is a tree that can be declared on its own: an
// expression or a call. A reference or a number is already as short as a
// name.
func shareable(op *program.Operand) bool {
	switch op.T.(type) {
	case *program.Operand_Expression, *program.Operand_FnCall:
		return true
	}
	return false
}

// eachChild calls f on each operand of op that is a value: an expression's
// operands, but of a swizzle only what it reads, and a call's arguments.
func eachChild(op *program.Operand, f func(i int, child *program.Operand)) {
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		for i, c := range t.Expression.Operands {
			if t.Expression.Operation == program.Operation_ACCESS && i > 0 {
				break
			}
			f(i, c)
		}
	case *program.Operand_FnCall:
		for i, c := range t.FnCall.Arguments {
			f(i, c)
		}
	}
}

// setChild replaces op's i-th child.
func setChild(op *program.Operand, i int, child *program.Operand) {
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		t.Expression.Operands[i] = child
	case *program.Operand_FnCall:
		t.FnCall.Arguments[i] = child
	}
}

// mostSavedRepeat is the key of the subtree of trees that sharing would save
// the most characters of, among those that occur more than once and print to
// at least boxShareMinLength, and whether there is one.
func mostSavedRepeat(trees []*program.Operand) (string, bool) {
	counts := map[string]int{}
	lengths := map[string]int{}
	var visit func(op *program.Operand)
	visit = func(op *program.Operand) {
		if op == nil || !shareable(op) {
			return
		}
		eachChild(op, func(_ int, c *program.Operand) { visit(c) })
		k := subtreeKey(op)
		if counts[k]++; counts[k] == 1 {
			lengths[k] = len(op.PrettyPrint(0))
		}
	}
	for _, t := range trees {
		visit(t)
	}
	best, bestSaving := "", 0
	for k, n := range counts {
		if n < 2 || lengths[k] < boxShareMinLength {
			continue
		}
		if saving := (n - 1) * lengths[k]; saving > bestSaving || (saving == bestSaving && k < best) {
			best, bestSaving = k, saving
		}
	}
	return best, best != ""
}

// replaceSubtree is op with every subtree whose key is key read as name; the
// first such subtree it meets is kept in *found, as the declared value.
func replaceSubtree(op *program.Operand, key, name string, found **program.Operand) *program.Operand {
	if op == nil || !shareable(op) {
		return op
	}
	if subtreeKey(op) == key {
		if *found == nil {
			*found = op
		}
		return ref(name)
	}
	eachChild(op, func(i int, c *program.Operand) {
		setChild(op, i, replaceSubtree(c, key, name, found))
	})
	return op
}

// inReadOrder orders declarations so each follows every other it reads.
func inReadOrder(decls []*program.Assignment) []*program.Assignment {
	byName := make(map[string]*program.Assignment, len(decls))
	for _, d := range decls {
		byName[d.Variables[0].Name] = d
	}
	var out []*program.Assignment
	placed := map[string]bool{}
	var place func(d *program.Assignment)
	place = func(d *program.Assignment) {
		name := d.Variables[0].Name
		if placed[name] {
			return
		}
		placed[name] = true
		reads := map[string]int{}
		countRefs(d.Expression, reads)
		for _, other := range decls {
			if n := other.Variables[0].Name; reads[n] > 0 {
				place(byName[n])
			}
		}
		out = append(out, d)
	}
	for _, d := range decls {
		place(d)
	}
	return out
}
