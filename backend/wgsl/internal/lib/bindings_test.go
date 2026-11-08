package wgsl

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/float16"
	"gradientcontrol.com/omega/core/internal/meshasset"
	"gradientcontrol.com/omega/core/internal/noise/spinodoid"
)

func TestPackSpinodoidBinding_VariableBandsAndMetadata(t *testing.T) {
	slots := []SpinodoidSlot{{Slot: 0, Key: "graded_part"}, {Slot: 1, Key: "constant"}}
	tables := map[string]spinodoid.BakedTable{
		"graded_part": {ContentID: "graded", Rows: 2, Origin: [3]int{-17, -5, -6}, Size: [3]int{32, 90, 2}, Bounds: [6]float64{-55, -14, -8, 62, 277, 42}},
		"constant":    {ContentID: "periodic", Rows: 64, Origin: [3]int{-32, -32, -32}, Size: [3]int{64, 64, 64}, Periodic: true},
	}
	raw := map[string][]byte{
		"graded":   bytes.Repeat([]byte{0x3f}, spinodoid.GPUWidth*2*16),
		"periodic": bytes.Repeat([]byte{0x4f}, spinodoid.GPUWidth*64*16),
	}
	binding, err := PackSpinodoidBinding(slots, tables, raw, "")
	if err != nil {
		t.Fatal(err)
	}
	if binding.Dims != [2]int{4096, 67} {
		t.Fatalf("dims = %v", binding.Dims)
	}
	packed, err := base64.StdEncoding.DecodeString(binding.TexelsB64)
	if err != nil {
		t.Fatal(err)
	}
	for slot, want := range [][16]float32{
		{-17, -5, -6, 1, 32, 90, 2, 0, -55, -14, -8, 0, 62, 277, 42, 0},
		{-32, -32, -32, 3, 64, 64, 64, 1},
	} {
		for i, expected := range want {
			offset := slot*64 + i*4
			got := math.Float32frombits(binary.LittleEndian.Uint32(packed[offset : offset+4]))
			if got != expected {
				t.Fatalf("slot%d metadata%d = %v, want %v", slot, i, got, expected)
			}
		}
	}
	if !bytes.Equal(packed[4096*16:3*4096*16], raw["graded"]) || !bytes.Equal(packed[3*4096*16:], raw["periodic"]) {
		t.Fatal("variable table bands were not copied at their header offsets")
	}
	// An unchanged layout can omit all bytes. Moving a domain changes the
	// texture hash even if a server supplied the same content id by mistake.
	held, err := PackSpinodoidBinding(slots, tables, nil, binding.Hash)
	if err != nil || held.TexelsB64 != "" || held.Hash != binding.Hash {
		t.Fatalf("held binding = %v, err %v", held, err)
	}
	moved := tables["graded_part"]
	moved.Origin[1]++
	tables["graded_part"] = moved
	changed, err := PackSpinodoidBinding(slots, tables, raw, binding.Hash)
	if err != nil || changed.Hash == binding.Hash || changed.TexelsB64 == "" {
		t.Fatalf("changed domain retained old binding: %v, err %v", changed, err)
	}
}

func TestPackSpinodoidBinding_RejectsInvalidBands(t *testing.T) {
	slots := []SpinodoidSlot{{Slot: 0, Key: "a"}}
	tables := map[string]spinodoid.BakedTable{"a": {ContentID: "id", Rows: 1}}
	if _, err := PackSpinodoidBinding(slots, tables, map[string][]byte{"id": {0}}, ""); err == nil {
		t.Fatal("accepted truncated table")
	}
	if _, err := PackSpinodoidBinding([]SpinodoidSlot{{Slot: 0, Key: "missing"}}, tables, nil, ""); err == nil {
		t.Fatal("accepted missing table key")
	}
	tooTall := tables["a"]
	tooTall.Rows = spinodoid.GPUMaxRows
	tables["a"] = tooTall
	if _, err := PackSpinodoidBinding(slots, tables, nil, ""); err == nil {
		t.Fatal("accepted a table leaving no room for its header")
	}
	if _, err := PackSpinodoidBinding([]SpinodoidSlot{{Slot: 1024, Key: "a"}}, tables, nil, ""); err == nil {
		t.Fatal("accepted metadata outside the header row")
	}
}

func TestBuildMeshSdfBindings(t *testing.T) {
	slots := []MeshSdfSlot{
		{Slot: 0, AssetIndex: 2},
		{Slot: 1, AssetIndex: 5},
	}
	bakes := map[int]*meshasset.SdfGrid{
		2: {
			Dims:   [3]int{2, 1, 1},
			Bbox:   meshasset.Bbox{Min: [3]float64{-1, -2, -3}, Max: [3]float64{3, 4, 5}},
			Values: []float32{-0.5, 1.25},
		},
		5: {
			Dims:   [3]int{1, 1, 1},
			Bbox:   meshasset.Bbox{Min: [3]float64{0, 0, 0}, Max: [3]float64{10, 20, 30}},
			Values: []float32{2.0},
		},
	}

	got := BuildMeshSdfBindings(slots, bakes)
	if len(got) != 2 {
		t.Fatalf("len(BuildMeshSdfBindings) = %d, want 2", len(got))
	}
	if got[0].Slot != 0 || got[0].AssetIndex != 2 || got[0].Dims != [3]int{2, 1, 1} {
		t.Fatalf("got[0] = %+v", got[0])
	}
	if got[0].BboxMin[0] != -1 || got[0].BboxMin[1] != -2 || got[0].BboxMin[2] != -3 ||
		got[0].BboxSize[0] != 4 || got[0].BboxSize[1] != 6 || got[0].BboxSize[2] != 8 {
		t.Fatalf("got[0] bbox = min %v size %v", got[0].BboxMin, got[0].BboxSize)
	}
	raw0, err := base64.StdEncoding.DecodeString(got[0].BytesB64)
	if err != nil {
		t.Fatalf("decode got[0].BytesB64: %v", err)
	}
	v0 := float16.ToFloat32(binary.LittleEndian.Uint16(raw0[0:2]))
	v1 := float16.ToFloat32(binary.LittleEndian.Uint16(raw0[2:4]))
	if v0 != -0.5 || v1 != 1.25 {
		t.Fatalf("decoded got[0] values = (%v, %v), want (-0.5, 1.25)", v0, v1)
	}

	if got[1].Slot != 1 || got[1].AssetIndex != 5 || got[1].Dims != [3]int{1, 1, 1} {
		t.Fatalf("got[1] = %+v", got[1])
	}
}
