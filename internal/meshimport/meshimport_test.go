package meshimport_test

import (
	"encoding/binary"
	"errors"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/meshimport"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
	"gradientcontrol.com/omega/core/internal/test/stlbox"
	"gradientcontrol.com/omega/core/program"
	"gradientcontrol.com/omega/core/program/meshasset"
)

// The box every test imports: 20 x 10 x 6 mm, centred on the origin.
var (
	lo = [3]float64{-10, -5, -3}
	hi = [3]float64{10, 5, 3}
)

func boxDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	stlbox.Write(t, dir, "box.stl", lo, hi)
	return dir
}

func resolve(t *testing.T, source string, find meshimport.Finder) ([]meshimport.Use, error) {
	t.Helper()
	res, err := assembler.Assemble([]string{source})
	require.NoError(t, err)
	return meshimport.Resolve(res.Program, res.Slots, find)
}

func TestResolve_NumberedDeclaration(t *testing.T) {
	uses, err := resolve(t, "importMesh 3 \"box.stl\";\nreturn bvhSdf(3);\n", meshimport.Dirs(boxDir(t)))
	require.NoError(t, err)
	require.Len(t, uses, 1)
	u := uses[0]
	assert.Equal(t, 3, u.AssetIndex)
	assert.Equal(t, program.MeshSdfKind_MESH_SDF_KIND_BVH, u.Kind)
	assert.Equal(t, "box.stl", u.Asset.Filename)
	assert.Len(t, u.Asset.Mesh.Triangles, 12)
	assert.Len(t, u.Asset.Vertices, 12*9)
	assert.Equal(t, []uint32{0, 1, 2}, u.Asset.Triangles[:3])
}

// The expression form gets its slot from the assembler, and vdbSdf's arguments
// set the level set: 64 voxels along the longest axis, a 2-voxel half band.
func TestResolve_ExpressionFormAndVdbArguments(t *testing.T) {
	uses, err := resolve(t, "return vdbSdf(importMesh \"box.stl\", 64, 2.0);\n", meshimport.Dirs(boxDir(t)))
	require.NoError(t, err)
	require.Len(t, uses, 1)
	assert.Equal(t, program.MeshSdfKind_MESH_SDF_KIND_VDB, uses[0].Kind)
	assert.Equal(t, float32(20.0/64), uses[0].VoxelSize)
	assert.Equal(t, float32(2), uses[0].HalfBandVoxels)
}

func TestResolve_NothingToLoad(t *testing.T) {
	never := func(string, string) ([]byte, error) {
		t.Fatal("a program that reads no mesh looked for a file")
		return nil, nil
	}
	uses, err := resolve(t, "importMesh 0 \"box.stl\";\nreturn length(p) - 1;\n", never)
	require.NoError(t, err)
	assert.Len(t, uses, 0)
}

func TestResolve_Failures(t *testing.T) {
	dir := boxDir(t)
	_, err := resolve(t, "importMesh 0 \"gone.stl\";\nreturn bvhSdf(0);\n", meshimport.Dirs(dir))
	require.Error(t, err)
	assert.True(t, errors.Is(err, fs.ErrNotExist), err)
	assert.Contains(t, err.Error(), `importMesh "gone.stl"`)

	_, err = resolve(t, "return bvhSdf(4);\n", meshimport.Dirs(dir))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no importMesh declares slot 4")

	_, err = resolve(t, "importStep 0 \"part.step\";\nreturn bvhSdf(0);\n", meshimport.Dirs(dir))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only importMesh .stl files load here")

	require.NoError(t, os.WriteFile(filepath.Join(dir, "junk.stl"), []byte("solid nothing\nendsolid\n"), 0o644))
	_, err = resolve(t, "importMesh 0 \"junk.stl\";\nreturn bvhSdf(0);\n", meshimport.Dirs(dir))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "junk.stl")
}

// A slot read by both builtins gets both oracles; reading it twice does not
// duplicate one.
func TestOracles_OnePerSlotAndKind(t *testing.T) {
	uses, err := resolve(t, "importMesh 0 \"box.stl\";\nreturn min(bvhSdf(0), vdbSdf(0)), bvhSdf(0) + 1;\n",
		meshimport.Dirs(boxDir(t)))
	require.NoError(t, err)
	bvh, vdb := meshimport.Oracles(uses)
	require.Len(t, bvh, 1)
	require.Len(t, vdb, 1)
	assert.Equal(t, uint32(0), bvh[0].AssetID)
	assert.Equal(t, uint32(0), vdb[0].AssetID)
}

func TestDirsAndChain(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "parts", "brackets"), 0o755))
	stlbox.Write(t, filepath.Join(root, "parts", "brackets"), "dep.stl", lo, hi)
	data, err := meshimport.Dirs(t.TempDir(), root)("parts/brackets", "dep.stl")
	require.NoError(t, err)
	assert.Equal(t, stlbox.Bytes(lo, hi), data)

	first := func(pkgPath, name string) ([]byte, error) {
		if name == "up.stl" {
			return []byte("uploaded"), nil
		}
		return nil, fs.ErrNotExist
	}
	find := meshimport.Chain(first, meshimport.Dirs(root))
	data, err = find("", "up.stl")
	require.NoError(t, err)
	assert.Equal(t, "uploaded", string(data))
	_, err = find("", "none.stl")
	assert.True(t, errors.Is(err, fs.ErrNotExist), err)
}

func TestGridDims(t *testing.T) {
	assert.Equal(t, [3]uint32{256, 128, 77}, meshimport.GridDims(meshimportBox()))
	thin := meshimportBox()
	thin.Max[2] = thin.Min[2] + 0.01
	assert.Equal(t, uint32(32), meshimport.GridDims(thin)[2])
}

// The preview grid of a box: the level set sampled at GridDims cells, a box
// that covers the mesh, inside negative at the centre and outside positive at
// a corner.
func TestGrids_PreviewOfABox(t *testing.T) {
	uses, err := resolve(t, "importMesh 0 \"box.stl\";\nreturn bvhSdf(0);\n", meshimport.Dirs(boxDir(t)))
	require.NoError(t, err)
	grids, err := meshimport.Grids(uses)
	require.NoError(t, err)
	g := grids[0]
	require.True(t, g != nil, "no grid for slot 0")
	dims := meshimport.GridDims(uses[0].Asset.Mesh.Bbox)
	assert.Equal(t, [3]int{int(dims[0]), int(dims[1]), int(dims[2])}, g.Dims)
	assert.Equal(t, g.Len(), len(g.Values))
	for k := 0; k < 3; k++ {
		assert.True(t, g.Bbox.Min[k] <= lo[k] && hi[k] <= g.Bbox.Max[k], "grid box %v covers the mesh", g.Bbox)
	}
	assert.True(t, g.At(g.Dims[0]/2, g.Dims[1]/2, g.Dims[2]/2) < 0, "the centre is inside")
	assert.True(t, g.At(0, 0, 0) > 0, "a corner of the grid is outside")

	again, err := meshimport.Grids(uses)
	require.NoError(t, err)
	assert.True(t, again[0] == g, "the same mesh and parameters reuse the baked grid")
}

// fidget samples the box faces; the decoded grid's box reaches half a cell
// past them, to the cells' outer edges.
func TestDecodeDenseGrid(t *testing.T) {
	var raw []byte
	put := func(v uint32) { raw = binary.LittleEndian.AppendUint32(raw, v) }
	for _, d := range []uint32{3, 2, 1} {
		put(d)
	}
	for _, f := range []float32{0, 0, 0, 2, 4, 6} {
		put(math.Float32bits(f))
	}
	for i := 0; i < 6; i++ {
		put(math.Float32bits(float32(i)))
	}
	g, err := meshimport.DecodeDenseGrid(raw)
	require.NoError(t, err)
	assert.Equal(t, [3]int{3, 2, 1}, g.Dims)
	assert.Equal(t, [3]float64{-0.5, -2, 0}, g.Bbox.Min)
	assert.Equal(t, [3]float64{2.5, 6, 6}, g.Bbox.Max)
	assert.Equal(t, float32(5), g.At(2, 1, 0))

	_, err = meshimport.DecodeDenseGrid(raw[:len(raw)-4])
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "do not match"), err)
}

func meshimportBox() meshasset.Bbox {
	return meshasset.Bbox{Min: lo, Max: hi}
}
