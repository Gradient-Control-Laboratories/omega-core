package refs

import (
	"reflect"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

func val(v float64) *program.Operand {
	return &program.Operand{T: &program.Operand_Value{Value: v}}
}

func refCall(name string, idx float64, extra ...*program.Operand) *program.Operand {
	args := append([]*program.Operand{val(idx)}, extra...)
	return program.NewFnCall("", name, args, nil, nil)
}

func meshCall(idx float64, extra ...*program.Operand) *program.Operand {
	return refCall("mesh", idx, extra...)
}

func bvhSdfCall(idx float64) *program.Operand {
	return refCall("bvhSdf", idx)
}

func vdbSdfCall(idx float64) *program.Operand {
	return refCall("vdbSdf", idx)
}

func expr(op program.Operation, args ...*program.Operand) *program.Operand {
	return &program.Operand{T: &program.Operand_Expression{Expression: &program.Expression{Operation: op, Operands: args}}}
}

func TestCollectBasicReturns(t *testing.T) {
	m := &program.Package{
		Files: []*program.PackageFile{
			{
				Returns: []*program.Operand{
					meshCall(0),
					bvhSdfCall(1),
				},
			},
		},
	}
	got := CollectMeshReferences(m)
	want := []MeshReference{
		{Index: 0, IsSdf: false},
		{Index: 1, IsSdf: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestCollectDeduplicates(t *testing.T) {
	m := &program.Package{
		Files: []*program.PackageFile{
			{
				Returns: []*program.Operand{
					meshCall(0), meshCall(0),
					bvhSdfCall(0), bvhSdfCall(0),
				},
			},
		},
	}
	got := CollectMeshReferences(m)
	want := []MeshReference{
		{Index: 0, IsSdf: false},
		{Index: 0, IsSdf: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestCollectFromNestedFunctionBody(t *testing.T) {
	m := &program.Package{
		Files: []*program.PackageFile{
			{
				FnTypes: []*program.FnType{
					{
						Name: "inner",
						Groups: []*program.FnGroup{
							{
								ParamCount: 0,
								Functions: []*program.Function{
									{
										Returns: []*program.Operand{bvhSdfCall(7)},
									},
								},
							},
						},
					},
				},
				Returns: []*program.Operand{meshCall(2)},
			},
		},
	}
	got := CollectMeshReferences(m)
	want := []MeshReference{
		{Index: 7, IsSdf: true},
		{Index: 2, IsSdf: false},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestCollectFromInsideExpression(t *testing.T) {
	// min(bvhSdf(3), 1.0) in a return position.
	m := &program.Package{
		Files: []*program.PackageFile{
			{
				Returns: []*program.Operand{
					expr(program.Operation_MIN, bvhSdfCall(3), val(1.0)),
				},
			},
		},
	}
	got := CollectMeshReferences(m)
	want := []MeshReference{{Index: 3, IsSdf: true}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestCollectFromInsideRemapArgument(t *testing.T) {
	// remap(bvhSdfFn, p*2) — the bvhSdf call nested as the first argument
	// of a remap expression must still be surfaced by the walker.
	m := &program.Package{
		Files: []*program.PackageFile{
			{
				Returns: []*program.Operand{
					expr(program.Operation_REMAP, bvhSdfCall(5), val(2.0)),
				},
			},
		},
	}
	got := CollectMeshReferences(m)
	want := []MeshReference{{Index: 5, IsSdf: true}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestCollectVdbSdf(t *testing.T) {
	// A program using both bvhSdf(0) and vdbSdf(0) for the same asset
	// must surface as two distinct references so the canvas agent
	// registers both oracles.
	m := &program.Package{
		Files: []*program.PackageFile{
			{
				Returns: []*program.Operand{
					bvhSdfCall(0),
					vdbSdfCall(0),
				},
			},
		},
	}
	got := CollectMeshReferences(m)
	want := []MeshReference{
		{Index: 0, IsSdf: true, IsVdb: false},
		{Index: 0, IsSdf: true, IsVdb: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestCollectIgnoresNonLiteralIndex(t *testing.T) {
	// mesh(somevar) — first arg is a reference, not a Value. The walker
	// should skip it silently; downstream code validates literal form.
	m := &program.Package{
		Files: []*program.PackageFile{
			{
				Returns: []*program.Operand{
					program.NewFnCall("", "mesh", []*program.Operand{
						program.NewReference("", "somevar", nil, nil),
					}, nil, nil),
				},
			},
		},
	}
	got := CollectMeshReferences(m)
	if len(got) != 0 {
		t.Fatalf("got %v, want empty", got)
	}
}

func TestCollectNilModule(t *testing.T) {
	if got := CollectMeshReferences(nil); got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

// fnFile wraps a package file whose functions return the given operands.
func fnFile(returns ...*program.Operand) *program.PackageFile {
	return &program.PackageFile{
		FnTypes: []*program.FnType{
			{
				Groups: []*program.FnGroup{
					{
						Functions: []*program.Function{
							{
								Returns: returns,
							},
						},
					},
				},
			},
		},
	}
}

// TestCollectWithImports — a mesh builtin called inside an imported library
// is collected only when imports are walked (regression for vdbSdf(i) in a
// library function getting "no assigned slot").
func TestCollectWithImports(t *testing.T) {
	lib := &program.Package{
		Files: []*program.PackageFile{
			fnFile(vdbSdfCall(2)),
		},
	}
	primary := &program.Package{
		Files: []*program.PackageFile{
			{
				Imports: []*program.Import{{Path: "lib"}},
			},
		},
	}
	getImport := func(path string) (*program.Package, error) {
		if path == "lib" {
			return lib, nil
		}
		return nil, nil
	}

	if got := CollectMeshReferences(primary); len(got) != 0 {
		t.Fatalf("without imports: got %v, want none", got)
	}
	got := CollectMeshReferencesWithImports(primary, getImport)
	want := []MeshReference{{Index: 2, IsSdf: true, IsVdb: true}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("with imports: got %v, want %v", got, want)
	}
}

// TestCollectWithImportsCycle — import cycles terminate.
func TestCollectWithImportsCycle(t *testing.T) {
	var a, b *program.Package
	a = &program.Package{
		Files: []*program.PackageFile{
			{
				Imports: []*program.Import{{Path: "b"}},
				FnTypes: []*program.FnType{
					{
						Groups: []*program.FnGroup{
							{
								Functions: []*program.Function{
									{
										Returns: []*program.Operand{vdbSdfCall(1)},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	b = &program.Package{
		Files: []*program.PackageFile{
			{
				Imports: []*program.Import{{Path: "a"}},
			},
		},
	}
	getImport := func(path string) (*program.Package, error) {
		switch path {
		case "a":
			return a, nil
		case "b":
			return b, nil
		}
		return nil, nil
	}
	got := CollectMeshReferencesWithImports(a, getImport)
	want := []MeshReference{{Index: 1, IsSdf: true, IsVdb: true}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestCollectAllOperandTypes(t *testing.T) {
	m := &program.Package{
		Files: []*program.PackageFile{
			{
				Returns: []*program.Operand{
					val(1.0),
					program.NewReference("", "somevar", nil, nil),
					program.NewPropertyRef("var", "prop", nil, nil),
					meshCall(4),
					expr(program.Operation_MIN, bvhSdfCall(5), val(1.0)),
				},
			},
		},
	}
	got := CollectMeshReferences(m)
	want := []MeshReference{
		{Index: 4, IsSdf: false},
		{Index: 5, IsSdf: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
