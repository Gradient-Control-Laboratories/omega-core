package property

import (
	"fmt"
	"math"
	"testing"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/program"
)

// nestedRotations is f0 a box, and each f_i a rotation remap of f_{i-1}, with
// ret(f_k) returned: the shape that took a derivative from milliseconds to
// never finishing, one nesting level at a time.
func nestedRotations(k int, ret func(field *program.Operand) *program.Operand) *program.Package {
	px := nary(program.Operation_ACCESS, ref(program.PVarName), ref(program.XVarName))
	py := nary(program.Operation_ACCESS, ref(program.PVarName), ref(program.YVarName))
	pz := nary(program.Operation_ACCESS, ref(program.PVarName), ref(program.ZVarName))
	c, s := num(math.Cos(0.3)), num(math.Sin(0.3))
	box := nary(program.Operation_MAX,
		add(un(program.Operation_ABS, px), num(-1)),
		add(un(program.Operation_ABS, py), num(-2)),
		add(un(program.Operation_ABS, pz), num(-3)))
	fn := func(name string, body *program.Operand) *program.FnType {
		return &program.FnType{Name: name, Groups: []*program.FnGroup{{
			Functions: []*program.Function{{Returns: []*program.Operand{body}}},
		}}}
	}
	call := func(name string) *program.Operand { return program.NewFnCall("", name, nil, nil, nil) }
	fns := []*program.FnType{fn("f0", box)}
	for i := 1; i <= k; i++ {
		q := nary(program.Operation_COLLECT,
			add(mul(px, c), mul(py, s)),
			add(mul(py, c), mul(un(program.Operation_NEG, px), s)),
			pz)
		fns = append(fns, fn(fmt.Sprintf("f%d", i), nary(program.Operation_REMAP, call(fmt.Sprintf("f%d", i-1)), q)))
	}
	return &program.Package{Files: []*program.PackageFile{{
		FnTypes: fns,
		Returns: []*program.Operand{ret(call(fmt.Sprintf("f%d", k)))},
	}}}
}

func dpOf(field *program.Operand) *program.Operand {
	return nary(program.Operation_DERIV, field, ref(program.PVarName))
}

// unshare writes ExpandShared's output back out inline: each call to a
// synthesized function becomes its return, and each shared local its body,
// recursively. What comes back must be Expand's own tree, node for node.
func unshare(t *testing.T, pkg *program.Package, op *program.Operand) *program.Operand {
	t.Helper()
	fns := map[string]*program.Function{}
	for _, f := range pkg.Files {
		for _, ft := range f.FnTypes {
			if ft != nil && len(ft.Groups) == 1 && len(ft.Groups[0].Functions) == 1 {
				fns[ft.Name] = ft.Groups[0].Functions[0]
			}
		}
	}
	var walk func(op *program.Operand, locals map[string]*program.Operand) *program.Operand
	walk = func(op *program.Operand, locals map[string]*program.Operand) *program.Operand {
		switch v := op.T.(type) {
		case *program.Operand_Reference:
			if body, ok := locals[v.Reference.Name]; ok && v.Reference.Prefix == "" {
				return walk(body, locals)
			}
			return op
		case *program.Operand_FnCall:
			fc := v.FnCall
			if fn, ok := fns[fc.Reference.Name]; ok && len(fc.Arguments) == 0 && isSharedFn(fc.Reference.Name) {
				inner := map[string]*program.Operand{}
				for _, a := range fn.Assignments {
					inner[a.Variables[0].Name] = a.Expression
				}
				return walk(fn.Returns[0], inner)
			}
			args := make([]*program.Operand, len(fc.Arguments))
			for i, a := range fc.Arguments {
				args[i] = walk(a, locals)
			}
			return &program.Operand{Start: op.Start, End: op.End, T: &program.Operand_FnCall{
				FnCall: &program.FnCall{Reference: fc.Reference, Arguments: args},
			}}
		case *program.Operand_Expression:
			e := v.Expression
			ops := make([]*program.Operand, len(e.Operands))
			for i, o := range e.Operands {
				ops[i] = walk(o, locals)
			}
			return &program.Operand{Start: op.Start, End: op.End, T: &program.Operand_Expression{
				Expression: &program.Expression{Operation: e.Operation, Operands: ops},
			}}
		}
		return op
	}
	return walk(op, nil)
}

func isSharedFn(name string) bool {
	var n int
	_, err := fmt.Sscanf(name, "shared__%d", &n)
	return err == nil
}

// Written back out inline, ExpandShared's result is exactly Expand's: the same
// expression, only spelled with names. Checked on nested rotations (the
// derivative's worst case), on .second over a union, and on the derivative of
// .second.
func TestExpandSharedIsExpandSpelledWithNames(t *testing.T) {
	cases := map[string]*program.Package{}
	for k := 1; k <= 4; k++ {
		cases[fmt.Sprintf(".dp through %d nested rotations", k)] = nestedRotations(k, dpOf)
	}
	cases[".second of a union"] = threeSpheresPkg(sc(nary(program.Operation_MIN, ref("a"), ref("b"), ref("c"))))
	cases[".dp of .second"] = threeSpheresPkg(dpOf(sc(nary(program.Operation_MIN, ref("a"), ref("b"), ref("c")))))
	cases[".second through 3 nested rotations"] = nestedRotations(3, sc)

	for name, pkg := range cases {
		t.Run(name, func(t *testing.T) {
			plain, err := Expand(&program.Program{Primary: proto.Clone(pkg).(*program.Package)})
			if err != nil {
				t.Fatalf("Expand: %v", err)
			}
			shared, err := ExpandShared(&program.Program{Primary: proto.Clone(pkg).(*program.Package)})
			if err != nil {
				t.Fatalf("ExpandShared: %v", err)
			}
			want := plain.Primary.Files[0].Returns[0]
			got := unshare(t, shared.Primary, shared.Primary.Files[0].Returns[0])
			if !proto.Equal(got, want) {
				t.Errorf("written back out inline, the shared expansion is not Expand's")
			}
		})
	}
}

// What is written down is linear in the nesting, not exponential: every node
// of the shared output — its return and every synthesized function — counted.
func TestExpandSharedStaysSmall(t *testing.T) {
	size := func(pkg *program.Package) int {
		n := 0
		var walk func(op *program.Operand)
		walk = func(op *program.Operand) {
			if op == nil {
				return
			}
			n++
			for _, c := range allChildren(op) {
				walk(c)
			}
		}
		for _, f := range pkg.Files {
			for _, r := range f.Returns {
				walk(r)
			}
			for _, ft := range f.FnTypes {
				for _, g := range ft.Groups {
					for _, fn := range g.Functions {
						for _, a := range fn.Assignments {
							walk(a.Expression)
						}
						for _, r := range fn.Returns {
							walk(r)
						}
					}
				}
			}
		}
		return n
	}
	var prev int
	for _, k := range []int{4, 8, 16} {
		shared, err := ExpandShared(&program.Program{Primary: nestedRotations(k, dpOf)})
		if err != nil {
			t.Fatalf("ExpandShared: %v", err)
		}
		n := size(shared.Primary)
		// Unshared this is a tree of ~5.6^k nodes: 309 thousand at k=4, 7.9
		// billion at k=10. Written down with names it grows by a constant per
		// level.
		if prev > 0 && n > 3*prev {
			t.Errorf("k=%d: %d nodes, more than 3x the %d at half the depth — growing faster than linear", k, n, prev)
		}
		if n > 100000 {
			t.Errorf("k=%d: %d nodes written down", k, n)
		}
		prev = n
	}
}

// With nothing shared worth a name, ExpandShared is Expand exactly.
func TestExpandSharedLeavesAnUnsharedExpansionAlone(t *testing.T) {
	pkg := threeSpheresPkg(dpOf(ref("b")))
	plain, err := Expand(&program.Program{Primary: proto.Clone(pkg).(*program.Package)})
	if err != nil {
		t.Fatal(err)
	}
	shared, err := ExpandShared(&program.Program{Primary: proto.Clone(pkg).(*program.Package)})
	if err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(plain, shared) {
		t.Errorf("a small expansion should come back from ExpandShared exactly as from Expand")
	}
}
