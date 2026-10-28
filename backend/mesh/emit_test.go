package mesh_test

import (
	"bytes"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/backend/mesh"
)

func TestMesh_Sphere(t *testing.T) {
	src := `return sphere(5);`
	res, err := assembler.Assemble([]string{src})
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	progressCalled := false
	progressFn := func(stage string, completed, total uint32) {
		progressCalled = true
	}

	stl, err := mesh.Mesh(res.Program,
		mesh.WithSubdivision(3),
		mesh.WithAlgorithm("marching-cubes"),
		mesh.WithBbox([6]float32{-6, -6, -6, 6, 6, 6}),
		mesh.WithTime(0.0),
		mesh.WithProgress(progressFn),
	)
	if err != nil {
		t.Fatalf("Mesh failed: %v", err)
	}
	if len(stl) < 84 {
		t.Fatalf("expected binary STL to have at least 84 bytes, got %d", len(stl))
	}
	if !progressCalled {
		t.Error("expected progress callback to be invoked")
	}

	// Verify MeshBodies
	bodies, err := mesh.MeshBodies(res.Program, mesh.WithSubdivision(3))
	if err != nil {
		t.Fatalf("MeshBodies failed: %v", err)
	}
	if len(bodies) == 0 {
		t.Fatal("expected at least one body returned")
	}
	if bodies[0].Name == "" {
		t.Error("expected non-empty body name")
	}
	if !bytes.Equal(bodies[0].STL[:80], stl[:80]) {
		t.Logf("body STL header matches generated STL")
	}
}
