package compiler_test

import (
	"encoding/binary"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/cmd/serve/internal/compiler"
	"gradientcontrol.com/omega/core/internal/meshimport"
	"gradientcontrol.com/omega/core/internal/test/stlbox"
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
