package property

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// TestColorMapOverQualifiedLibraryList: a shape whose body maps over a list a
// qualified library exports — every fastener size-table lookup does this —
// unioned with a colored sibling that is a remap of it. The engine used to
// resolve the qualified list as ONE opaque value, so the map closure received
// the whole table as its element and its `.y` became member access on the
// entire list; the builder then rejected the color field with "member access is
// not supported on vec10; use indexing instead" (vec10 = the table's rows).
func TestColorMapOverQualifiedLibraryList(t *testing.T) {
	// lib: tbl = [[1, 2], [3, 4], [5, 6]]
	row := func(a, b float64) *program.Operand { return nary(program.Operation_COLLECT, num(a), num(b)) }
	lib := &program.Package{Files: []*program.PackageFile{{
		Assignments: []*program.Assignment{assignDecl("tbl", nary(program.Operation_COLLECT, row(1, 2), row(3, 4), row(5, 6)))},
	}}}
	getImport := func(path string) (*program.Package, error) { return lib, nil }

	// sel(pr) { return pr.y; }   sum(acc, v) { return acc + v; }
	sel := &program.Function{
		Parameters: []*program.Parameter{{Name: "pr"}},
		Returns:    []*program.Operand{nary(program.Operation_ACCESS, ref("pr"), ref(program.YVarName))},
	}
	sum := &program.Function{
		Parameters: []*program.Parameter{{Name: "acc"}, {Name: "v"}},
		Returns:    []*program.Operand{add(ref("acc"), ref("v"))},
	}
	// rad() { return reduce(sum, map(sel, lib.tbl)); }
	rad := &program.Function{
		FnTypes: []*program.FnType{
			{Name: "sel", Groups: []*program.FnGroup{{ParamCount: 1, Functions: []*program.Function{sel}}}},
			{Name: "sum", Groups: []*program.FnGroup{{ParamCount: 2, Functions: []*program.Function{sum}}}},
		},
		Returns: []*program.Operand{nary(program.Operation_REDUCE, ref("sum"),
			nary(program.Operation_MAP, ref("sel"),
				&program.Operand{T: &program.Operand_Reference{Reference: &program.Reference{Prefix: "lib", Name: "tbl"}}}))},
	}
	// jaw() { return length(p - [2, 0, 0]) - rad() * 0.1; }
	jaw := &program.Function{
		Returns: []*program.Operand{sub(sphereSDF(2, 0, 0, 0), mul(call("rad"), num(0.1)))},
	}
	mirror := nary(program.Operation_COLLECT,
		neg(nary(program.Operation_ACCESS, ref(program.PVarName), ref(program.XVarName))),
		neg(nary(program.Operation_ACCESS, ref(program.PVarName), ref(program.YVarName))),
		nary(program.Operation_ACCESS, ref(program.PVarName), ref(program.ZVarName)))
	pkg := &program.Package{Files: []*program.PackageFile{{
		Imports: []*program.Import{{Path: "lib"}},
		Assignments: []*program.Assignment{
			coloredShape("a", call("jaw"), rgb(1, 0, 0)),
			coloredShape("b", nary(program.Operation_REMAP, call("jaw"), mirror), rgb(0, 0, 1)),
		},
		FnTypes: []*program.FnType{
			{Name: "rad", Groups: []*program.FnGroup{{ParamCount: 0, Functions: []*program.Function{rad}}}},
			{Name: "jaw", Groups: []*program.FnGroup{{ParamCount: 0, Functions: []*program.Function{jaw}}}},
		},
		Returns: []*program.Operand{bin(program.Operation_MIN, ref("a"), ref("b"))},
	}}}

	fields, err := ColorFields(pkg, getImport)
	if err != nil {
		t.Fatalf("ColorFields: %v", err)
	}
	if len(fields) != 1 || fields[0] == nil {
		t.Fatalf("expected one color field, got %v", fields)
	}
	if bad := qualifiedAccessIn(fields[0]); bad != "" {
		t.Fatalf("color field reads a member of a whole qualified list: %s\n%s", bad, fields[0].PrettyPrint(0))
	}
}

// qualifiedAccessIn returns "prefix.name" for the first ACCESS whose base is a
// qualified reference — a member read on an opaque library value — or "".
func qualifiedAccessIn(op *program.Operand) string {
	if op == nil {
		return ""
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		e := t.Expression
		if e.Operation == program.Operation_ACCESS && len(e.Operands) > 0 {
			if r, ok := e.Operands[0].T.(*program.Operand_Reference); ok && r.Reference.Prefix != "" {
				return r.Reference.Prefix + "." + r.Reference.Name
			}
		}
		for _, o := range e.Operands {
			if s := qualifiedAccessIn(o); s != "" {
				return s
			}
		}
	case *program.Operand_FnCall:
		for _, a := range t.FnCall.Arguments {
			if s := qualifiedAccessIn(a); s != "" {
				return s
			}
		}
	}
	return ""
}
