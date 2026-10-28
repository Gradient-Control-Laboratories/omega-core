package property

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// innerLibPkg is a library that ANOTHER library calls qualified: one function
// whose body indexes a collection by a parameter, aliased into a local first
// the way a fold callback's capture has to be.
func innerLibPkg() *program.Package {
	// pick(row, col) { at = col; return row[at]; }
	pick := &program.Function{
		Parameters:  []*program.Parameter{{Name: "row"}, {Name: "col"}},
		Assignments: []*program.Assignment{assignDecl("at", ref("col"))},
		Returns:     []*program.Operand{nary(program.Operation_INDEX, ref("row"), ref("at"))},
	}
	return &program.Package{Files: []*program.PackageFile{{
		FnTypes: []*program.FnType{
			{Name: "pick", Groups: []*program.FnGroup{{ParamCount: 2, Functions: []*program.Function{pick}}}},
		},
	}}}
}

// outerLibPkg is the library the document imports. It calls inner qualified,
// which is the arrangement the engine has to follow: mechanical/fastener reads
// its sizes from mechanical/catalog, and a document imports only fastener.
func outerLibPkg() *program.Package {
	// second(v) { return inner.pick(v, 1); }
	second := &program.Function{
		Parameters: []*program.Parameter{{Name: "v"}},
		Returns: []*program.Operand{{T: &program.Operand_FnCall{FnCall: &program.FnCall{
			Reference: &program.Reference{Prefix: "inner", Name: "pick"},
			Arguments: []*program.Operand{ref("v"), num(1)},
		}}}},
	}
	return &program.Package{Files: []*program.PackageFile{{
		Imports: []*program.Import{{Path: "inner"}},
		FnTypes: []*program.FnType{
			{Name: "second", Groups: []*program.FnGroup{{ParamCount: 1, Functions: []*program.Function{second}}}},
		},
	}}}
}

// expandThroughTwoLibs differentiates lib.second([p.x, p.y, p.z]) — a call into
// a library that calls a second library, whose body indexes by a parameter.
func expandThroughTwoLibs(t *testing.T) *program.Program {
	t.Helper()
	inner := innerLibPkg()
	inner.Path = "inner"
	outer := outerLibPkg()
	outer.Path = "lib"

	vec := nary(program.Operation_COLLECT,
		ref(program.XVarName), ref(program.YVarName), ref(program.ZVarName))
	deriv := nary(program.Operation_DERIV, qualCall("second", vec), ref(program.PVarName))

	prog, err := Expand(&program.Program{
		Primary:      importPkg(deriv),
		Dependencies: []*program.Package{outer, inner},
	})
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	return prog
}

// TestQualifiedImportOfQualifiedImport: a library the document imports may
// itself call a second library qualified, and that prefix has to resolve.
//
// The engine used to collect only the DOCUMENT's qualified imports, following
// dot-imports transitively but not qualified ones — so a two-deep library call
// failed with "package %q is not resolvable here". builtin's vector.dirZ needed
// a hardcoded special case for exactly this reason.
func TestQualifiedImportOfQualifiedImport(t *testing.T) {
	prog := expandThroughTwoLibs(t)
	if prog == nil || prog.Primary == nil {
		t.Fatal("expanded program has no primary package")
	}
}

// TestIndexOperandFromInlinedImport is the companion bug, and the sharper one.
//
// dom.Index takes the index as an OPERAND rather than a Value, because a
// backend indexes by a compile-time constant and the operand is what gets
// emitted. Inside an inlined foreign function body that operand may name one of
// that function's locals — `at` here — and those locals are not emitted
// anywhere. The reference then reaches the output naming nothing, and type
// analysis rejects it with "undefined identifier: at".
//
// So no reference to a foreign local may survive into the expanded program.
func TestIndexOperandFromInlinedImport(t *testing.T) {
	prog := expandThroughTwoLibs(t)
	for _, f := range prog.Primary.Files {
		for _, ret := range f.Returns {
			if leaked := findRef(ret, "at"); leaked {
				t.Error(`the expanded program references "at", a local of an inlined ` +
					`library function, which is not declared anywhere it can be read`)
			}
			if leaked := findRef(ret, "col"); leaked {
				t.Error(`the expanded program references "col", a parameter of an ` +
					`inlined library function, which is not declared in the output`)
			}
		}
	}
}

// findRef reports whether op mentions a bare (unprefixed) reference to name.
func findRef(op *program.Operand, name string) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		return t.Reference != nil && t.Reference.Prefix == "" && t.Reference.Name == name
	case *program.Operand_Expression:
		if t.Expression == nil {
			return false
		}
		for _, sub := range t.Expression.Operands {
			if findRef(sub, name) {
				return true
			}
		}
	case *program.Operand_FnCall:
		if t.FnCall == nil {
			return false
		}
		for _, sub := range t.FnCall.Arguments {
			if findRef(sub, name) {
				return true
			}
		}
	}
	return false
}
