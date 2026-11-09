package assembler_test

import (
	"testing"
	"time"

	"gradientcontrol.com/omega/core/assembler"
)

// A mesh belongs to the return that reaches it, and that return's index is what
// the SHAPES checkbox toggles. A body collected with no index is the failure
// this guards: drawn, listed nowhere, and impossible to hide.

func meshShapeIndices(t *testing.T, src string) []int {
	t.Helper()
	res, err := assembler.Assemble([]string{src}, assembler.ExtractMeshPlan("leaf-"))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	decls := res.MeshPlan.MeshDecls()
	var idx []int
	for _, d := range decls {
		idx = append(idx, d.ShapeIndex)
	}
	return idx
}

func TestMeshInReturnGetsItsShapeIndex(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want []int
	}{
		{
			"mesh() directly in the return",
			"importMesh 0 \"p.stl\";\nreturn mesh(0, [1,2,3]);\n",
			[]int{1},
		},
		{
			// The ordinary way to write a body, and the one that regressed.
			"bound, then returned by name",
			"importMesh 0 \"p.stl\";\nmsh = mesh(0, [1,2,3]);\nreturn msh;\n",
			[]int{1},
		},
		{
			"second of two returns",
			"importMesh 0 \"p.stl\";\nmsh = mesh(0, [1,2,3]);\nreturn 1, msh;\n",
			[]int{2},
		},
		{
			"nested in a boolean in the third return",
			"importMesh 0 \"p.stl\";\nreturn 1, 2, meshUnion(mesh(0, [1,2,3]), mesh(0, [4,5,6]));\n",
			[]int{3, 3},
		},
	}
	for _, c := range cases {
		got := meshShapeIndices(t, c.src)
		if len(got) != len(c.want) {
			t.Errorf("%s: %d meshes, want %d (%v)", c.name, len(got), len(c.want), got)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%s: shape indices %v, want %v", c.name, got, c.want)
				break
			}
		}
		for _, s := range got {
			if s == 0 {
				t.Errorf("%s: a mesh came back with no shape index; no checkbox can reach it", c.name)
			}
		}
	}
}

// A mesh nothing returns is not part of the scene. Drawing it anyway is how a
// left-over experiment becomes an object you cannot select, hide, or explain.
func TestMeshNoReturnReachesIsNotDrawn(t *testing.T) {
	got := meshShapeIndices(t, `importMesh 0 "p.stl";
msh = mesh(0, [1,2,3]);
return 1;
`)
	if len(got) != 0 {
		t.Errorf("a mesh no return reaches should not be collected; got shape indices %v", got)
	}
}

func TestMeshReachedThroughAFunctionCall(t *testing.T) {
	got := meshShapeIndices(t, `importMesh 0 "p.stl";
body() {
  return mesh(0, [1,2,3]);
}
return body();
`)
	if len(got) != 1 || got[0] != 1 {
		t.Errorf("a mesh returned through a call should belong to that return; got %v", got)
	}
}

// Following names must not spin on a program that refers to itself.
func TestMeshReferenceTerminates(t *testing.T) {
	done := make(chan []int, 1)
	go func() {
		done <- meshShapeIndices(t, `importMesh 0 "p.stl";
a = mesh(0, [1,2,3]);
b = meshUnion(a, mesh(0, [4,5,6]));
return b, a;
`)
	}()
	select {
	case got := <-done:
		for _, s := range got {
			if s == 0 {
				t.Errorf("shape index 0 in %v", got)
			}
		}
	case <-time.After(5 * time.Second):
		t.Fatal("reference following did not terminate")
	}
}

