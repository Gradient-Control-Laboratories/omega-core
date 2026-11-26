package compiler_test

import (
	"math"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/backend/mesh"
	"github.com/Gradient-Control-Laboratories/omega-core/cmd/serve/internal/compiler"
)

// planBody meshes source on the server and checks it comes back as one closed body.
func planBody(t *testing.T, req compiler.MeshRequest) mesh.STLStats {
	t.Helper()
	bodies, err := compiler.MeshBodies(req, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(bodies) != 1 {
		t.Fatalf("%d bodies, want 1", len(bodies))
	}
	stats, err := mesh.MeasureClosedSTL(bodies[0].STL)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Empty() {
		t.Fatal("body has no triangles")
	}
	return stats
}

// A mesh boolean of two fields meshes on the server: a sphere of radius 5 less one of radius 4 four along x, whose
// volume is the first sphere's less the lens they share, π(R+r−d)²(d²+2dr−3r²+2dR+6rR−3R²)/12d.
func TestMeshBodies_MeshDifferenceOfFields(t *testing.T) {
	const R, r, d = 5.0, 4.0, 4.0
	lens := math.Pi * (R + r - d) * (R + r - d) * (d*d + 2*d*r - 3*r*r + 2*d*R + 6*r*R - 3*R*R) / (12 * d)
	want := 4.0/3*math.Pi*R*R*R - lens
	for _, algo := range []string{"mdc", "marching-cubes"} {
		t.Run(algo, func(t *testing.T) {
			stats := planBody(t, compiler.MeshRequest{
				Source:      "a = length(p) - 5;\nb = length(p - [4, 0, 0]) - 4;\nreturn meshDifference(a, b);\n",
				Subdivision: 7,
				Algo:        algo,
				Bbox:        [6]float32{-10, -10, -10, 10, 10, 10},
				HasBbox:     true,
			})
			if math.Abs(stats.Volume-want) > 0.02*want {
				t.Errorf("volume %g, want %g", stats.Volume, want)
			}
		})
	}
}

// A field drills an imported mesh on the server: the 20 × 10 × 6 box less a cylinder of radius 2 along z. The
// cylinder never ends, so the box meshed (z within ±5) cuts it off past the box's faces, and the volume is the
// box's less the hole, 1200 − 24π.
func TestMeshBodies_FieldDrillsPlacedMesh(t *testing.T) {
	want := 1200 - math.Pi*4*6
	stats := planBody(t, compiler.MeshRequest{
		Source:      "importMesh 0 \"box.stl\";\nreturn meshDifference(mesh(0), length([p.x, p.y]) - 2);\n",
		Subdivision: 7,
		Algo:        "mdc",
		Bbox:        [6]float32{-5, -5, -5, 5, 5, 5},
		HasBbox:     true,
		Meshes:      boxFolder(t),
	})
	if math.Abs(stats.Volume-want) > 0.01*want {
		t.Errorf("volume %g, want %g", stats.Volume, want)
	}
}

// An imported mesh placed by mesh() unions with a field on the server: the 20 × 10 × 6 box moved 30 along x, apart
// from a sphere of radius 5, so the volume is both of theirs, 1200 + 500π/3.
func TestMeshBodies_PlacedMeshUnionField(t *testing.T) {
	want := 1200 + 4.0/3*math.Pi*125
	stats := planBody(t, compiler.MeshRequest{
		Source:      "importMesh 0 \"box.stl\";\nreturn meshUnion(mesh(0, [30, 0, 0]), length(p) - 5);\n",
		Subdivision: 7,
		Algo:        "mdc",
		Bbox:        [6]float32{-10, -10, -10, 10, 10, 10},
		HasBbox:     true,
		Meshes:      boxFolder(t),
	})
	if math.Abs(stats.Volume-want) > 0.02*want {
		t.Errorf("volume %g, want %g", stats.Volume, want)
	}
	if stats.Bbox[3] < 39.99 || stats.Bbox[3] > 40.01 {
		t.Errorf("x reaches %g; the placed box should reach 40", stats.Bbox[3])
	}
}
