package compiler_test

import (
	"encoding/base64"
	"encoding/binary"
	"math"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler/meshplan"
	"gradientcontrol.com/omega/core/cmd/serve/internal/compiler"
	"gradientcontrol.com/omega/core/internal/meshimport"
	"gradientcontrol.com/omega/core/internal/test/stlbox"
	"gradientcontrol.com/omega/core/program/meshasset"
)

var (
	boxLo = [3]float64{-10, -5, -3}
	boxHi = [3]float64{10, 5, 3}
)

func boxFolder(t *testing.T) meshimport.Finder {
	t.Helper()
	dir := t.TempDir()
	stlbox.Write(t, dir, "box.stl", boxLo, boxHi)
	return meshimport.Dirs(dir)
}

// A compile of a program that reads an imported mesh carries the preview grid
// the shader samples for its slot, as half floats.
func TestCompile_ImportedMeshCarriesItsPreviewGrid(t *testing.T) {
	resp := compiler.Compile(compiler.CompileRequest{
		ID:     "m",
		Source: "importMesh 0 \"box.stl\";\nreturn bvhSdf(0);\n",
		Meshes: boxFolder(t),
	}, nil, nil)
	if len(resp.Errors) > 0 {
		t.Fatalf("compile errors: %+v", resp.Errors)
	}
	if len(resp.MeshSdfBindings) != 1 {
		t.Fatalf("got %d mesh bindings, want 1", len(resp.MeshSdfBindings))
	}
	b := resp.MeshSdfBindings[0]
	dims := meshimport.GridDims(meshasset.Bbox{Min: boxLo, Max: boxHi})
	if b.Dims != [3]int{int(dims[0]), int(dims[1]), int(dims[2])} {
		t.Errorf("dims %v, want %v", b.Dims, dims)
	}
	raw, err := base64.StdEncoding.DecodeString(b.BytesB64)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 2*b.Dims[0]*b.Dims[1]*b.Dims[2] {
		t.Errorf("%d bytes of half floats for %v cells", len(raw), b.Dims)
	}
	for k := 0; k < 3; k++ {
		if b.BboxMin[k] > boxLo[k] || b.BboxMin[k]+b.BboxSize[k] < boxHi[k] {
			t.Errorf("the grid's box (min %v, size %v) does not cover the mesh", b.BboxMin, b.BboxSize)
		}
	}
	if !strings.Contains(resp.WGSL, "sdfTex0") {
		t.Error("the shader samples no sdfTex0")
	}
}

func TestCompile_ImportedMeshFailures(t *testing.T) {
	resp := compiler.Compile(compiler.CompileRequest{
		Source: "importMesh 0 \"box.stl\";\nreturn bvhSdf(0);\n",
	}, nil, nil)
	if len(resp.Errors) == 0 || !strings.Contains(resp.Errors[0].Message, "no folder") {
		t.Errorf("without a mesh folder: %+v", resp.Errors)
	}
	resp = compiler.Compile(compiler.CompileRequest{
		Source: "importMesh 0 \"gone.stl\";\nreturn bvhSdf(0);\n",
		Meshes: boxFolder(t),
	}, nil, nil)
	if len(resp.Errors) == 0 || !strings.Contains(resp.Errors[0].Message, `importMesh "gone.stl"`) {
		t.Errorf("a missing file: %+v", resp.Errors)
	}
	// A program that imports a mesh only to place it reads no file.
	resp = compiler.Compile(compiler.CompileRequest{
		Source: "importMesh 0 \"gone.stl\";\nreturn length(p) - 2;\n",
	}, nil, nil)
	if len(resp.Errors) > 0 {
		t.Errorf("no distance builtin, yet: %+v", resp.Errors)
	}
}

// A mesh the program places with mesh(i, placement) comes back as a
// declaration the viewer draws itself, and the shader draws only the SDF parts.
func TestCompile_PlacedMeshIsADeclaration(t *testing.T) {
	resp := compiler.Compile(compiler.CompileRequest{
		Source: "importMesh 0 \"box.stl\";\nxf = uiTransform([[1, 0, 0, 0], [0, 1, 0, 0], [0, 0, 1, 0], [0, 0, 0, 1]]);\n" +
			"return bvhSdf(0) - 1, mesh(0, xf);\n",
		Meshes: boxFolder(t),
	}, nil, nil)
	if len(resp.Errors) > 0 {
		t.Fatalf("compile errors: %+v", resp.Errors)
	}
	decls, ok := resp.MeshDecls.([]meshplan.MeshDecl)
	if !ok || len(decls) != 1 || decls[0].Index != 0 || decls[0].ShapeIndex != 2 {
		t.Fatalf("mesh declarations %#v, want mesh 0 as the second return", resp.MeshDecls)
	}
	if len(resp.MeshSdfBindings) != 1 {
		t.Errorf("got %d mesh bindings, want 1 for bvhSdf", len(resp.MeshSdfBindings))
	}
	if strings.Contains(resp.WGSL, "1e30") {
		t.Error("the shader still carries the placed mesh's sentinel")
	}
}

// Meshing on the server loads the mesh too: the box's volume, 1200 mm³.
func TestMesh_ImportedMesh(t *testing.T) {
	stl, err := compiler.Mesh(compiler.MeshRequest{
		Source:      "importMesh 0 \"box.stl\";\nreturn vdbSdf(0);\n",
		Subdivision: 6,
		Algo:        "mdc",
		Bbox:        [6]float32{-12, -7, -5, 12, 7, 5},
		HasBbox:     true,
		Meshes:      boxFolder(t),
	}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	n := int(binary.LittleEndian.Uint32(stl[80:84]))
	volume := 0.0
	for i := 0; i < n; i++ {
		var v [3][3]float64
		for c := 0; c < 3; c++ {
			for k := 0; k < 3; k++ {
				off := 84 + 50*i + 12 + 12*c + 4*k
				v[c][k] = float64(math.Float32frombits(binary.LittleEndian.Uint32(stl[off:])))
			}
		}
		volume += (v[0][0]*(v[1][1]*v[2][2]-v[1][2]*v[2][1]) - v[0][1]*(v[1][0]*v[2][2]-v[1][2]*v[2][0]) +
			v[0][2]*(v[1][0]*v[2][1]-v[1][1]*v[2][0])) / 6
	}
	if math.Abs(volume-1200) > 0.02*1200 {
		t.Errorf("volume %g, want 1200", volume)
	}
}
