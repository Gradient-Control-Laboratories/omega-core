package property

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

// repeatedBox is a box whose corners repeat a long subtree, a bound over
// three slider points, in all six, and a short one, r + 1, in two.
func repeatedBox() *bbox {
	long := func() *program.Operand {
		return call("max", call("max", nary(program.Operation_ADD, ref("pointA"), ref("spread")),
			nary(program.Operation_ADD, ref("pointB"), ref("spread"))),
			nary(program.Operation_MUL, ref("pointC"), ref("spread")))
	}
	short := func() *program.Operand { return nary(program.Operation_ADD, ref("r"), num(1)) }
	bx := &bbox{}
	for i := 0; i < 3; i++ {
		bx.lo[i] = nary(program.Operation_SUB, num(float64(i)), long())
		bx.hi[i] = nary(program.Operation_ADD, num(float64(i)), long())
	}
	bx.lo[2] = nary(program.Operation_SUB, bx.lo[2], short())
	bx.hi[2] = nary(program.Operation_ADD, bx.hi[2], short())
	return bx
}

// inlined is op with every reference to a declaration replaced by its value,
// recursively.
func inlined(op *program.Operand, decls map[string]*program.Operand) *program.Operand {
	op = proto.Clone(op).(*program.Operand)
	var walk func(o *program.Operand) *program.Operand
	walk = func(o *program.Operand) *program.Operand {
		if r := o.GetReference(); r != nil {
			if v, ok := decls[r.Name]; ok {
				return walk(proto.Clone(v).(*program.Operand))
			}
			return o
		}
		eachChild(o, func(i int, c *program.Operand) { setChild(o, i, walk(c)) })
		return o
	}
	return walk(op)
}

// Sharing rewrites the corners and nothing else: inlining the declarations
// back gives the corners as they were, so the box is the same box. The long
// subtree is declared once; the short one, under boxShareMinLength, is left
// where it is; each declaration reads only those before it.
func TestShareBoxRepeats_IsTheSameBox(t *testing.T) {
	want := repeatedBox()
	decls, got := shareBoxRepeats(repeatedBox(), map[string]bool{"gslBox": true})
	if len(decls) == 0 {
		t.Fatal("nothing shared")
	}
	values := map[string]*program.Operand{}
	for _, d := range decls {
		name := d.Variables[0].Name
		if name == "gslBox" {
			t.Errorf("a declaration took the taken name %q", name)
		}
		reads := map[string]int{}
		countRefs(d.Expression, reads)
		for other := range reads {
			if _, declared := values[other]; !declared && strings.HasPrefix(other, "gslBox") {
				t.Errorf("%s reads %s, which is not declared before it", name, other)
			}
		}
		values[name] = d.Expression
	}
	for i := 0; i < 3; i++ {
		for _, c := range []struct {
			corner    string
			got, want *program.Operand
		}{{"lo", got.lo[i], want.lo[i]}, {"hi", got.hi[i], want.hi[i]}} {
			if back := inlined(c.got, values); !proto.Equal(back, c.want) {
				t.Errorf("%s[%d] inlined back is %s, want %s", c.corner, i, back.PrettyPrint(0), c.want.PrettyPrint(0))
			}
		}
	}
	for name, v := range values {
		if len(v.PrettyPrint(0)) < boxShareMinLength {
			t.Errorf("%s = %s is shorter than boxShareMinLength", name, v.PrettyPrint(0))
		}
	}
	if s := got.hi[2].PrettyPrint(0); len(decls) != 1 {
		t.Errorf("want the long subtree alone declared, got %d declarations (hi[2] = %s)", len(decls), s)
	}
}

// With sharing off, or nothing long enough repeated, the box comes back as it
// was, with no declarations.
func TestShareBoxRepeats_LeavesWhatItCannotShare(t *testing.T) {
	shareBoxes = false
	decls, got := shareBoxRepeats(repeatedBox(), map[string]bool{})
	shareBoxes = true
	if len(decls) != 0 || !proto.Equal(got.lo[0], repeatedBox().lo[0]) {
		t.Errorf("sharing off: got %d declarations", len(decls))
	}
	short := &bbox{}
	for i := 0; i < 3; i++ {
		short.lo[i] = nary(program.Operation_SUB, ref("x"), num(1))
		short.hi[i] = nary(program.Operation_ADD, ref("x"), num(1))
	}
	if decls, _ := shareBoxRepeats(short, map[string]bool{}); len(decls) != 0 {
		t.Errorf("short repeats: got %d declarations, want none", len(decls))
	}
}

// nestedBox repeats a long subtree, outer, three times, and inside it and
// twice on its own a shorter one, inner. outer saves the most and is shared
// first; inner, shared after, is read by outer's declaration, which must
// therefore come after it.
func nestedBox() *bbox {
	inner := func() *program.Operand {
		return call("max", nary(program.Operation_ADD, ref("pointAlpha"), ref("spreadWidth")),
			nary(program.Operation_MUL, ref("pointBravo"), ref("spreadWidth")))
	}
	outer := func() *program.Operand {
		other := call("min", call("min", nary(program.Operation_SUB, ref("pointC"), ref("spread")),
			nary(program.Operation_SUB, ref("pointD"), ref("spread"))),
			call("min", nary(program.Operation_SUB, ref("pointE"), ref("spread")),
				nary(program.Operation_SUB, ref("pointF"), ref("spread"))))
		return call("max", inner(), other)
	}
	bx := &bbox{}
	bx.lo[0], bx.hi[0], bx.lo[1] = outer(), outer(), outer()
	bx.hi[1], bx.lo[2] = inner(), inner()
	bx.hi[2] = num(1)
	return bx
}

func TestShareBoxRepeats_DeclaresInReadOrder(t *testing.T) {
	want := nestedBox()
	decls, got := shareBoxRepeats(nestedBox(), map[string]bool{})
	if len(decls) != 2 {
		t.Fatalf("want 2 declarations, got %d", len(decls))
	}
	values := map[string]*program.Operand{}
	for _, d := range decls {
		reads := map[string]int{}
		countRefs(d.Expression, reads)
		for other := range reads {
			if _, declared := values[other]; !declared && strings.HasPrefix(other, "gslBox") {
				t.Errorf("%s reads %s, which is declared after it", d.Variables[0].Name, other)
			}
		}
		values[d.Variables[0].Name] = d.Expression
	}
	for i := 0; i < 3; i++ {
		if back := inlined(got.lo[i], values); !proto.Equal(back, want.lo[i]) {
			t.Errorf("lo[%d] inlined back differs", i)
		}
		if back := inlined(got.hi[i], values); !proto.Equal(back, want.hi[i]) {
			t.Errorf("hi[%d] inlined back differs", i)
		}
	}
}
