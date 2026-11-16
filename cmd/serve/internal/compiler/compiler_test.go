package compiler_test

import (
	"encoding/binary"
	"testing"

	"gradientcontrol.com/omega/core/cmd/serve/internal/compiler"
)

func TestMesh_BinarySTLAndProgress(t *testing.T) {
	src := `return sphere(5);`
	var meshStages []string
	stlBytes, err := compiler.Mesh(compiler.MeshRequest{
		ID:          "mesh-1",
		Source:      src,
		Subdivision: 2,
		Bbox:        [6]float32{-8, -8, -8, 8, 8, 8},
		HasBbox:     true,
		Algo:        "marching-cubes",
		Time:        0,
	}, []string{"."}, func(stage, label string, completed, total uint32) {
		meshStages = append(meshStages, stage)
	})
	if err != nil {
		t.Fatalf("unexpected Mesh error: %v", err)
	}
	if len(stlBytes) < 84 {
		t.Fatalf("expected binary STL >= 84 bytes, got %d", len(stlBytes))
	}
	triCount := binary.LittleEndian.Uint32(stlBytes[80:84])
	if triCount == 0 {
		t.Errorf("expected non-zero triangle count in binary STL")
	}
	expectedLen := 84 + int(triCount)*50
	if len(stlBytes) != expectedLen {
		t.Errorf("expected STL byte length %d for %d triangles, got %d", expectedLen, triCount, len(stlBytes))
	}
	if len(meshStages) == 0 {
		t.Errorf("expected progress callbacks during Mesh")
	}
}

func TestMesh_ErrorOnInvalidSource(t *testing.T) {
	_, err := compiler.Mesh(compiler.MeshRequest{
		ID:     "mesh-err",
		Source: "return nonexistent(1);",
	}, nil, nil)
	if err == nil {
		t.Fatalf("expected error from Mesh on invalid GSL source")
	}
}

func TestMesh_WithAudioBlock(t *testing.T) {
	src := `return sphere(5), audio(sin(440 * 2 * pi * time));`
	stlBytes, err := compiler.Mesh(compiler.MeshRequest{
		ID:          "mesh-audio",
		Source:      src,
		Subdivision: 2,
		Bbox:        [6]float32{-8, -8, -8, 8, 8, 8},
		HasBbox:     true,
		Algo:        "marching-cubes",
	}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected Mesh error: %v", err)
	}
	if len(stlBytes) < 84 {
		t.Fatalf("expected binary STL >= 84 bytes, got %d", len(stlBytes))
	}
}
