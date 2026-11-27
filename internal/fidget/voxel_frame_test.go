package fidget

import (
	"math"
	"testing"
)

// A ball of radius 0.5 rendered through the identity transform, 64 voxels on
// a side, fills fidget's [-1, 1] screen square at its centre: the middle pixel
// hits near the ball's top with a normal toward the viewer, the corners miss,
// and the SSAO is NaN on a miss and an occlusion factor on a hit.
func TestRenderVoxelsAndSSAOOnABall(t *testing.T) {
	Lock()
	defer Unlock()
	Clear()
	defer Clear()
	x, y, z := NewX(), NewY(), NewZ()
	ball := Sub(Sqrt(Add(Add(Square(x), Square(y)), Square(z))), NewConst(0.5))
	canvas := BuildCanvas(ball)

	const n = 64
	identity := [16]float32{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
	f, err := RenderVoxels(canvas, n, n, n, identity, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Z) != n*n || len(f.Normal) != 3*n*n {
		t.Fatalf("%d depths and %d normals for a %dx%d frame", len(f.Z), len(f.Normal), n, n)
	}

	// The pixel at (32, 31) sits at screen (0, 0): fidget's y runs down with
	// an offset of one row. The ball's top, z = 0.5, is voxel 0.5 / k + n/2
	// with k = 2 / n: 48.
	mid := 31*n + 32
	if z := f.Z[mid]; z < 46 || z > 49 {
		t.Fatalf("middle pixel depth %d, want about 48", z)
	}
	nx, ny, nz := f.Normal[3*mid], f.Normal[3*mid+1], f.Normal[3*mid+2]
	if !(nz > 0) || math.Abs(float64(nx)) > 0.2*float64(nz) || math.Abs(float64(ny)) > 0.2*float64(nz) {
		t.Fatalf("middle pixel normal (%g, %g, %g), want toward the viewer", nx, ny, nz)
	}
	for _, i := range []int{0, n - 1, (n - 1) * n, n*n - 1} {
		if f.Z[i] != 0 {
			t.Fatalf("corner pixel %d hit at depth %d", i, f.Z[i])
		}
	}

	occ, err := VoxelSSAO(f)
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsNaN(float64(occ[0])) {
		t.Fatalf("SSAO on a miss is %g, want NaN", occ[0])
	}
	if o := occ[mid]; math.IsNaN(float64(o)) || o < 0 || o > 1 {
		t.Fatalf("SSAO on the ball's top is %g, want an occlusion factor in [0, 1]", o)
	}
}
