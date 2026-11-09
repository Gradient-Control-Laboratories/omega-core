package assembler_test

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
)

// TestExtractMeshDeclsImportedConst guards the import-aware constant folding
// of mesh() placement arguments: pi comes from the omega library via a dot
// import, so the extracted placement must fold to concrete radians instead of
// the whole declaration being dropped with an eval error.
func TestExtractMeshDeclsImportedConst(t *testing.T) {
	src := `import . "omega";

return mesh(0, [pi, pi / 2, 1]);
`
	res, err := assembler.Assemble([]string{src}, assembler.ExtractMeshPlan("leaf-"))
	if err != nil {
		t.Fatalf("assemble error: %v", err)
	}

	extract := res.MeshPlan.MeshDecls()
	if len(extract) != 1 {
		t.Fatalf("top-level mesh() count: got %d, want 1", len(extract))
	}
	m := extract[0].Matrix
	if len(m) != 16 {
		t.Fatalf("Matrix length: got %d, want 16", len(m))
	}
	// translation column folds the imported pi.
	want := []float64{math.Pi, math.Pi / 2, 1}
	for i := range want {
		if math.Abs(m[12+i]-want[i]) > 1e-9 {
			t.Fatalf("translation: got [%v,%v,%v], want %v", m[12], m[13], m[14], want)
		}
	}
}

// TestExtractMeshDeclsComposeFolds verifies mesh(index, matrix.compose(...))
// folds to a 16-value Matrix in the canvas: the euler constructors are written
// in closed form so the lightweight const-evaluator can fold them (no
// multiply/map, no intermediate-matrix indexing).
func TestExtractMeshDeclsComposeFolds(t *testing.T) {
	src := "import \"matrix\";\n" +
		"return mesh(0, matrix.compose([3, 0, 0], [0, 0, 0], [1, 1, 1]));\n"
	res, err := assembler.Assemble([]string{src}, assembler.ExtractMeshPlan("leaf-"))
	if err != nil {
		t.Fatalf("assemble failed: %v", err)
	}

	extract := res.MeshPlan.MeshDecls()
	if len(extract) != 1 {
		t.Fatalf("top-level mesh() count: got %d, want 1", len(extract))
	}
	d := extract[0]
	if len(d.Matrix) != 16 {
		t.Fatalf("Matrix length: got %d, want 16; decl=%+v", len(d.Matrix), d)
	}
	// identity rotation/scale, translation (3,0,0) → 4th column.
	if d.Matrix[0] != 1 || d.Matrix[5] != 1 || d.Matrix[10] != 1 {
		t.Errorf("diagonal: got %v %v %v, want 1 1 1", d.Matrix[0], d.Matrix[5], d.Matrix[10])
	}
	if d.Matrix[12] != 3 || d.Matrix[13] != 0 || d.Matrix[14] != 0 {
		t.Errorf("translation: got [%v,%v,%v], want [3,0,0]", d.Matrix[12], d.Matrix[13], d.Matrix[14])
	}
}

// TestExtractMeshDeclsRotationPlaceFolds verifies the mesh(0, position,
// matrix.rotation(spin, orient)) form folds to the 16-value placement
// T(position)·Rz(spin)·R(orient): matrix.rotation builds the real rotation
// product (not an euler sum), and the canvas drops `position` into the 4th
// column. rotation builds from scalars only, so it folds here AND compiles in
// the full build (guarded separately in the wgsl codegen tests).
func TestExtractMeshDeclsRotationPlaceFolds(t *testing.T) {
	// A non-pure-Z base orient, so spin+orient is a genuine composition.
	orient := [3]float64{0.3, -0.7, 1.1}
	spin := 0.5
	pos := [3]float64{2, -5, 7}

	src := "import \"matrix\";\n" +
		"return mesh(0, [2, -5, 7], matrix.rotation(0.5, [0.3, -0.7, 1.1]));\n"
	res, err := assembler.Assemble([]string{src}, assembler.ExtractMeshPlan("leaf-"))
	if err != nil {
		t.Fatalf("assemble failed: %v", err)
	}

	extract := res.MeshPlan.MeshDecls()
	if len(extract) != 1 {
		t.Fatalf("top-level mesh() count: got %d, want 1", len(extract))
	}
	got := extract[0].Matrix
	if len(got) != 16 {
		t.Fatalf("Matrix length: got %d, want 16", len(got))
	}

	// Expected: column-major T(pos)·(Rz(spin)·R(orient)).
	rp := matMul3(rotZ3(spin), rotEuler3(orient))
	var want [16]float64
	for c := 0; c < 3; c++ {
		for r := 0; r < 3; r++ {
			want[c*4+r] = rp[r][c]
		}
	}
	want[12], want[13], want[14] = pos[0], pos[1], pos[2]
	want[15] = 1
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-9 {
			t.Fatalf("Matrix[%d]: got %v, want %v\n got=%v\nwant=%v", i, got[i], want[i], got, want)
		}
	}
}

// rotZ3 is the reference 3×3 Z rotation (omega.eulerAngles convention for
// [0,0,spin]), row-indexed m[row][col].
func rotZ3(spin float64) [3][3]float64 {
	c, s := math.Cos(spin), math.Sin(spin)
	return [3][3]float64{
		{c, s, 0},
		{-s, c, 0},
		{0, 0, 1},
	}
}

// rotEuler3 is the reference 3×3 euler rotation (omega.eulerAngles
// convention), as a row-indexed m[row][col].
func rotEuler3(a [3]float64) [3][3]float64 {
	cx, sx := math.Cos(a[0]), math.Sin(a[0])
	cy, sy := math.Cos(a[1]), math.Sin(a[1])
	cz, sz := math.Cos(a[2]), math.Sin(a[2])
	return [3][3]float64{
		{cz * cy, cz*sy*sx + sz*cx, -cz*sy*cx + sz*sx},
		{-sz * cy, -sz*sy*sx + cz*cx, sz*sy*cx + cz*sx},
		{sy, -cy * sx, cy * cx},
	}
}

func matMul3(a, b [3][3]float64) [3][3]float64 {
	var p [3][3]float64
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			for k := 0; k < 3; k++ {
				p[r][c] += a[r][k] * b[k][c]
			}
		}
	}
	return p
}

// TestExtractMeshDeclsMapIndexPlacement guards mesh() placements that index
// into a map whose function returns a vector through an imported function:
//
//	radials = map((ang) { return rotateXY(dirX * shift, ang); }, angles);
//	mesh(0, radials[0], ...);
func TestExtractMeshDeclsMapIndexPlacement(t *testing.T) {
	src := `import . "omega";
import map "map";

partPlacementShift = 4;
angles = [-pi / 3, -tau / 3, pi / 3, tau / 3];

radials = map(
  (ang) { return map.rotateXY(vector.dirX * partPlacementShift, ang); },
  angles
);

return mesh(0, radials[0]);
`
	res, err := assembler.Assemble([]string{src}, assembler.ExtractMeshPlan("leaf-"))
	if err != nil {
		t.Fatalf("assemble error: %v", err)
	}

	extract := res.MeshPlan.MeshDecls()
	if len(extract) != 1 {
		t.Fatalf("top-level mesh() count: got %d, want 1", len(extract))
	}

	// radials[0] = rotateXY([4, 0, 0], -pi/3) = [2, 2*sqrt(3), 0] -> translation column
	m := extract[0].Matrix
	if len(m) != 16 {
		t.Fatalf("Matrix length: got %d, want 16", len(m))
	}
	want := []float64{2, 2 * math.Sqrt(3), 0}
	for i := range want {
		if math.Abs(m[12+i]-want[i]) > 1e-9 {
			t.Fatalf("translation: got [%v,%v,%v], want %v", m[12], m[13], m[14], want)
		}
	}
}

// TestMapMeshUnionPlacements guards mesh() calls written inside a map over a
// list of placement vectors:
//
//	meshUnion(map((value, index) { return mesh(0, value, matrix.rotation(0, orient)); }, radials))
//
// The map unrolls to one placement per radials element for both the preview
// decls and the export plan.
func TestMapMeshUnionPlacements(t *testing.T) {
	src := `import . "omega";
import map "map";
import matrix "matrix";

partPlacementShift = 4;
partOrient = [0, pi / 2, 0];
angles = [-pi / 3, -tau / 3, pi / 3, tau / 3];

radials = map(
  (ang) { return map.rotateXY(vector.dirX * partPlacementShift, -ang); },
  angles
);

solid = length(p) - 1;

return meshDifference([
    meshUnion(map(
      (value, index) { return mesh(0, value, matrix.rotation(-angles[index], partOrient)); },
      radials
    )),
    solid
  ],
  solid
);
`
	res, err := assembler.Assemble([]string{src}, assembler.ExtractMeshPlan("leaf-"))
	if err != nil {
		t.Fatalf("assemble error: %v", err)
	}

	// rotateXY([4, 0, 0], ang) = [4 cos(ang), -4 sin(ang), 0]
	s3 := 2 * math.Sqrt(3)
	wantPositions := [][3]float64{
		{2, -s3, 0},  // -pi/3
		{-2, -s3, 0}, // -2pi/3
		{2, s3, 0},   // pi/3
		{-2, s3, 0},  // 2pi/3
	}

	extract := res.MeshPlan.MeshDecls()
	if len(extract) != len(wantPositions) {
		t.Fatalf("nested mesh() count: got %d, want %d", len(extract), len(wantPositions))
	}
	for k, want := range wantPositions {
		m := extract[k].Matrix
		if len(m) != 16 {
			t.Fatalf("nested decl %d Matrix length: got %d, want 16", k, len(m))
		}
		for i := range want {
			if math.Abs(m[12+i]-want[i]) > 1e-6 {
				t.Fatalf("nested decl %d translation: got [%v,%v,%v], want %v", k, m[12], m[13], m[14], want)
			}
		}
	}
}

