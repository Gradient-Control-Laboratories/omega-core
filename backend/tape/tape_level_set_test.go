package tape_test

import (
	"math"
	"testing"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/backend/tape"
)

// A tape asked for level sets carries each vdbSdf's baked grid in its oracle
// payload; one that was not asked carries only the triangles.
func TestEmbedLevelSetsCarriesTheBakedGrid(t *testing.T) {
	corners := [8][3]float32{}
	for k := range corners {
		for a := 0; a < 3; a++ {
			corners[k][a] = float32((k>>a)&1)*2 - 1
		}
	}
	faces := [][3]int{{0, 2, 1}, {1, 2, 3}, {4, 5, 6}, {5, 7, 6}, {0, 1, 4}, {1, 5, 4}, {2, 6, 3}, {3, 6, 7}, {0, 4, 2}, {2, 4, 6}, {1, 3, 5}, {3, 7, 5}}
	verts := make([]float32, 0, len(faces)*9)
	tris := make([]uint32, 0, len(faces)*3)
	for _, f := range faces {
		for _, c := range f {
			tris = append(tris, uint32(len(tris)))
			verts = append(verts, corners[c][0], corners[c][1], corners[c][2])
		}
	}
	payloads := []tape.VdbOracle{{AssetID: 7, Vertices: verts, Triangles: tris, VoxelSize: 0.05, HalfBandVoxels: 3}}
	res, err := assembler.Assemble([]string{"return vdbSdf(7, 20, 3);"})
	if err != nil {
		t.Fatal(err)
	}
	stamped, err := tape.Compile(res.Program, tape.Bbox([6]float32{-2, -2, -2, 2, 2, 2}), tape.VdbOracles(payloads))
	if err != nil {
		t.Fatal(err)
	}
	vdbPayload := func(b []byte) *tape.OraclePayload {
		var tp tape.Tape
		if err := proto.Unmarshal(b, &tp); err != nil {
			t.Fatal(err)
		}
		if len(tp.Oracles) != 1 || tp.Oracles[0].Kind != tape.OracleKind_VDB {
			t.Fatalf("tape oracles %v, want one VDB payload", tp.Oracles)
		}
		return tp.Oracles[0]
	}
	if vdbPayload(stamped).LevelSet != nil {
		t.Fatal("a tape not asked for level sets carries one")
	}

	withLevelSets, err := tape.EmbedLevelSets(stamped, payloads)
	if err != nil {
		t.Fatal(err)
	}
	ls := vdbPayload(withLevelSets).LevelSet
	if ls == nil {
		t.Fatal("no level set in the tape")
	}
	blocks := int(ls.BlockDims[0] * ls.BlockDims[1] * ls.BlockDims[2])
	leaves := len(ls.LeafValues) / 512
	if leaves == 0 || len(ls.LeafValues)%512 != 0 || len(ls.BlockLeaf) != blocks || len(ls.BlockValue) != blocks {
		t.Fatalf("level set holds %d leaf values and %d/%d block entries for %d blocks", len(ls.LeafValues), len(ls.BlockLeaf), len(ls.BlockValue), blocks)
	}
	if math.Abs(ls.VoxelSize-0.05) > 1e-7 || math.Abs(float64(ls.Background)-0.15) > 1e-6 {
		t.Fatalf("voxel size %v background %v, want 0.05 and 3 voxels", ls.VoxelSize, ls.Background)
	}
}
