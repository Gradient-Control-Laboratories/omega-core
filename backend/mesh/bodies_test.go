package mesh

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"io"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func triCount(stl []byte) uint32 { return binary.LittleEndian.Uint32(stl[80:84]) }

func bodiesOpts() []Option {
	return []Option{WithSubdivision(6), WithAlgorithm("mdc"), WithBbox([6]float32{-20, -20, -20, 20, 20, 20})}
}

// Each top-level return is meshed on its own and named as the Bodies panel
// names it; disjoint bodies add up to the union, and a repeated return is a
// body of its own rather than vanishing into the union.
func TestMeshBodiesSplitsReturns(t *testing.T) {
	src := "@scene(40)\nball = length(p - [8, 0, 0]) - 5;\nblk = box([-14, -4, -4], [-4, 4, 4]);\nreturn ball, blk, ball;\n"
	res, err := assembler.Assemble([]string{src})
	require.NoError(t, err)

	whole, err := Mesh(res.Program, bodiesOpts()...)
	require.NoError(t, err)
	bodies, err := MeshBodies(res.Program, bodiesOpts()...)
	require.NoError(t, err)

	require.Len(t, bodies, 3)
	for i, want := range []string{"ball", "blk", "ball-2"} {
		assert.Equal(t, want, bodies[i].Name)
		assert.Equal(t, i+1, bodies[i].Index)
		assert.True(t, triCount(bodies[i].STL) > 0, "%s has triangles", want)
	}
	assert.Equal(t, triCount(whole), triCount(bodies[0].STL)+triCount(bodies[1].STL))
	assert.True(t, bytes.Equal(bodies[0].STL, bodies[2].STL), "the repeated return meshes the same")
}

// A program with one return is meshed whole, byte for byte as Mesh meshes it.
func TestMeshBodiesOneReturnIsMesh(t *testing.T) {
	res, err := assembler.Assemble([]string{"@scene(40)\nball = length(p) - 5;\nreturn ball;\n"})
	require.NoError(t, err)
	whole, err := Mesh(res.Program, bodiesOpts()...)
	require.NoError(t, err)
	bodies, err := MeshBodies(res.Program, bodiesOpts()...)
	require.NoError(t, err)
	require.Len(t, bodies, 1)
	assert.Equal(t, "ball", bodies[0].Name)
	assert.Equal(t, 1, bodies[0].Index)
	assert.True(t, bytes.Equal(whole, bodies[0].STL), "one body is Mesh's STL")
}

// A return with no variable or call to name it takes its source text, as
// program.OperandName gives it.
func TestMeshBodiesUnnamedReturn(t *testing.T) {
	_, names := topLevelReturnsOf(t, "@scene(40)\nball = length(p) - 5;\nreturn ball, 1.0;\n")
	assert.Equal(t, "ball", names[0])
	assert.Equal(t, "1.0", names[1])
}

func topLevelReturnsOf(t *testing.T, src string) (int, []string) {
	t.Helper()
	res, err := assembler.Assemble([]string{src})
	require.NoError(t, err)
	returns, names := topLevelReturns(res.Program)
	return len(returns), names
}

func TestZipEntryNames(t *testing.T) {
	got := zipEntryNames([]string{"ball", "a/b:c", "  two\n words ", "", "ball", "ball"})
	assert.Equal(t, []string{"ball.stl", "a_b_c.stl", "two words.stl", "body.stl", "ball-2.stl", "ball-3.stl"}, got)
}

func TestPackSTL(t *testing.T) {
	one := []Body{{Name: "a", Index: 1, STL: []byte("stl-a")}}
	data, zipped, err := PackSTL(one)
	require.NoError(t, err)
	assert.False(t, zipped)
	assert.Equal(t, "stl-a", string(data))

	two := append(one, Body{Name: "a", Index: 2, STL: []byte("stl-b")})
	data, zipped, err = PackSTL(two)
	require.NoError(t, err)
	assert.True(t, zipped)
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	var entries []string
	for _, f := range zr.File {
		rc, err := f.Open()
		require.NoError(t, err)
		b, err := io.ReadAll(rc)
		require.NoError(t, err)
		rc.Close()
		entries = append(entries, f.Name+"="+string(b))
	}
	assert.Equal(t, []string{"a.stl=stl-a", "a-2.stl=stl-b"}, entries)
}

// Each body is a 3MF object in the palette's color for its return index, or
// the color given for it.
func TestWrite3MFColors(t *testing.T) {
	res, err := assembler.Assemble([]string{"@scene(40)\nball = length(p - [8, 0, 0]) - 5;\nblk = box([-14, -4, -4], [-4, 4, 4]);\nreturn ball, blk;\n"})
	require.NoError(t, err)
	bodies, err := MeshBodies(res.Program, bodiesOpts()...)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, Write3MF(&buf, bodies, map[int]string{2: "#123456"}))
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)
	var model string
	for _, f := range zr.File {
		if f.Name == "3D/3dmodel.model" {
			rc, err := f.Open()
			require.NoError(t, err)
			b, err := io.ReadAll(rc)
			require.NoError(t, err)
			rc.Close()
			model = string(b)
		}
	}
	assert.True(t, strings.Contains(model, "   <m:color color=\"#8987FF\"/>\n   <m:color color=\"#123456\"/>\n"), "colors in body order")
	assert.True(t, strings.Contains(model, `name="ball" pid="1" pindex="0"`))
	assert.True(t, strings.Contains(model, `name="blk" pid="1" pindex="1"`))
}

// WithBodies meshes only the returns it lists, keeping their return indices;
// listing none of them is an error.
func TestMeshBodiesWithBodies(t *testing.T) {
	res, err := assembler.Assemble([]string{"@scene(40)\nball = length(p - [8, 0, 0]) - 5;\nblk = box([-14, -4, -4], [-4, 4, 4]);\nreturn ball, blk;\n"})
	require.NoError(t, err)
	bodies, err := MeshBodies(res.Program, append(bodiesOpts(), WithBodies(2))...)
	require.NoError(t, err)
	require.Len(t, bodies, 1)
	assert.Equal(t, "blk", bodies[0].Name)
	assert.Equal(t, 2, bodies[0].Index)

	_, err = MeshBodies(res.Program, append(bodiesOpts(), WithBodies())...)
	require.Error(t, err)
	assert.Equal(t, "nothing to export: every body is hidden", err.Error())
}
