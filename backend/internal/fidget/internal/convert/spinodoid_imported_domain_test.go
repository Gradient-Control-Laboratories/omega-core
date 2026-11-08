package convert_test

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/convert"
	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/fidget"
	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/meshfold"
	"gradientcontrol.com/omega/core/internal/noise"
	"gradientcontrol.com/omega/core/program"
)

// This follows the imported, graded, powder-bed-material project pipeline.
// Remaps inside the material's noise libraries must not switch the unrelated
// geometry's driven spinodoid back to the old repeating table.
func TestImportedGradedSpinodoidWithPrintedMaterialKeepsFiniteDomain(t *testing.T) {
	const assetID = 0
	const src = `// Scene Size: 600 mm
import "lattice";
import "lattice/spinodoid";
import "material/printed";

meshId = importMesh "part.stl";
body = vdbSdf(meshId);
loEnd = [0, 0, 10];
hiEnd = [0, 264, 10];
span = hiEnd - loEnd;
t = clamp(dot(p - loEnd, span) / dot(span, span), 0, 1);
refineLine = mix(0.1, 4, t);
refinePoints = 0;
upright = [[1, 0, 0], [0, 1, 0], [0, 0, 1]];
isotropic = spinodoid.fastGaborGraded(2, [pi / 2, pi / 2, pi / 2], refinePoints + refineLine, upright);
specimen = intersectRound(body, lattice.solidify(isotropic, 0, 0.6), 0.1);
slsMat, slsColor, slsHeight, slsSlope = printed.powderBed([0.86, 0.85, 0.81], 0.06);
specimen.color = slsColor;
specimen.material = slsMat;
return specimen;
`
	res, err := assembler.Assemble([]string{src}, mockMgr)
	if err != nil {
		t.Fatal(err)
	}
	prog, err := meshfold.FoldProgramMeshIndices(res.Program, func(_, kind, path string) (int, bool) {
		return assetID, kind == "mesh" && path == "part.stl"
	})
	if err != nil {
		t.Fatal(err)
	}
	// Canvas supplies these imported bounds before preview/export diverge;
	// the larger authored scene cube must not replace them during conversion.
	domain := []float64{-66, -19, -18, 76, 294, 53}
	if prog.Primary.Scene == nil {
		prog.Primary.Scene = &program.Scene{Size: 600}
	}
	prog.Primary.Scene.SpinodoidBounds = append([]float64(nil), domain...)

	// A small, deterministic mesh substitutes for a user's private STL. Its
	// coarse VDB is enough: the sample points are well inside the solid.
	vertices := []float32{
		-60, -5, -15, 70, -5, -15, 70, 280, -15, -60, 280, -15,
		-60, -5, 50, 70, -5, 50, 70, 280, 50, -60, 280, 50,
	}
	triangles := []uint32{0, 2, 1, 0, 3, 2, 4, 5, 6, 4, 6, 7, 0, 1, 5, 0, 5, 4, 3, 7, 6, 3, 6, 2, 0, 4, 7, 0, 7, 3, 1, 2, 6, 1, 6, 5}
	fidget.Lock()
	err = fidget.RegisterVdbOracle(assetID, vertices, triangles, 5, 3)
	if err == nil {
		err = fidget.RegisterNoiseTile(noise.Bake())
	}
	fidget.Unlock()
	if err != nil {
		t.Fatal(err)
	}

	bakes, err := convert.BakeSpinodoids(prog)
	if err != nil {
		t.Fatal(err)
	}
	if len(bakes) != 1 {
		t.Fatalf("got %d spinodoid bakes, want one", len(bakes))
	}
	var baked convert.BakedSpinodoid
	for _, b := range bakes {
		baked = b
	}
	table := baked.Table
	if table.Periodic {
		t.Fatalf("material changed the imported grading domain: %+v", table.Layout)
	}
	for i, got := range table.Bounds {
		if math.Abs(got-domain[i]) > 1e-3 {
			t.Fatalf("imported domain was replaced: got %v, want %v", table.Bounds, domain)
		}
	}
	xs := make([]float32, 8)
	ys := []float32{10, 80, 112, 116, 120, 160, 220, 260}
	zs := make([]float32, len(ys))
	lastScale := 0.0
	for i, y := range ys {
		zs[i] = 10
		_, _, beta, _ := table.Sample([3]float64{0, float64(y), 10})
		scale, want := table.BaseBeta/beta, 0.1+3.9*float64(y)/264
		if math.Abs(scale-want) > 0.03 || scale <= lastScale {
			t.Fatalf("grading wrapped at y=%g: scale=%g, want %g > %g", y, scale, want, lastScale)
		}
		lastScale = scale
	}

	// Build the material/color path too, which actually visits its noise
	// remaps. The native geometry must still evaluate the finite graded table.
	err = convert.ToColorShapes(prog, 0, func(shapes []convert.ColorShape) error {
		if len(shapes) != 1 || !shapes[0].HasColor || !shapes[0].HasMaterial {
			t.Fatalf("printed specimen lost its material/color: %+v", shapes)
		}
		values, err := fidget.EvalGrad(fidget.BuildCanvas(shapes[0].Tree), xs, ys, zs)
		if err != nil {
			return err
		}
		for i, got := range values {
			distance, _ := table.Distance([3]float64{0, float64(ys[i]), 10}, 0, true)
			want := math.Abs(distance) - 0.3 // solidify width 0.6, inside the VDB
			if math.Abs(float64(got.Value)-want) > 2e-4 {
				t.Fatalf("native material build diverged at y=%g: %g != %g", ys[i], got.Value, want)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// The driver source is unchanged, but a different imported domain needs
	// different bytes and a new cache/content identity.
	prog.Primary.Scene.SpinodoidBounds[4] += 20
	changed, err := convert.BakeSpinodoids(prog)
	if err != nil {
		t.Fatal(err)
	}
	for _, next := range changed {
		if next.ContentID == baked.ContentID || next.Table == table {
			t.Fatal("changed imported bounds reused the old table")
		}
	}
}

func TestDirectSpinodoidQueryUsesTransformedFiniteDomain(t *testing.T) {
	const src = `// Scene Size: 20
upright = [[1, 0, 0], [0, 1, 0], [0, 0, 1]];
return fastGaborSpinodoid(p + [0, 200, 0], 2, [pi / 2, pi / 2, pi / 2], upright, 3, 8, 4096, 0, 0, 1, 1 + 0.001 * p.y, upright);
`
	prog, baked := bakeSingleSpinodoidForDomainTest(t, src)
	if baked.Table.Periodic || !baked.Table.Contains([3]float64{0, 200, 0}) || baked.Table.Contains([3]float64{}) {
		t.Fatalf("translated query did not get a translated finite domain: %+v", baked.Table.Layout)
	}
	_, _, beta, _ := baked.Table.Sample([3]float64{0, 200, 0})
	if scale := baked.Table.BaseBeta / beta; math.Abs(scale-1.2) > 0.002 {
		t.Fatalf("translated kernel centres got scale %g, want 1.2", scale)
	}
	want, _ := baked.Table.Distance([3]float64{0, 200, 0}, 0, true)
	assertSpinodoidProgramValue(t, prog, want)
}

func TestOuterRemappedSpinodoidPreservesLegacyField(t *testing.T) {
	const src = `// Scene Size: 20
import "lattice/spinodoid";
upright = [[1, 0, 0], [0, 1, 0], [0, 0, 1]];
field = spinodoid.fastGaborGraded(2, [pi / 2, pi / 2, pi / 2], 1 + 0.001 * p.y, upright);
return remap(field, p + [0, 200, 0]);
`
	prog, baked := bakeSingleSpinodoidForDomainTest(t, src)
	if !baked.Table.Periodic {
		t.Fatal("outer remap retained a finite table that clips the transformed query")
	}
	want, _ := baked.Table.Distance([3]float64{0, 200, 0}, 0, true)
	assertSpinodoidProgramValue(t, prog, want)
}

func bakeSingleSpinodoidForDomainTest(t *testing.T, src string) (*program.Program, convert.BakedSpinodoid) {
	t.Helper()
	res, err := assembler.Assemble([]string{src}, mockMgr)
	if err != nil {
		t.Fatal(err)
	}
	bakes, err := convert.BakeSpinodoids(res.Program)
	if err != nil {
		t.Fatal(err)
	}
	if len(bakes) != 1 {
		t.Fatalf("got %d tables, want one", len(bakes))
	}
	for _, baked := range bakes {
		return res.Program, baked
	}
	panic("unreachable")
}

func assertSpinodoidProgramValue(t *testing.T, prog *program.Program, want float64) {
	t.Helper()
	err := convert.ToTrees(prog, func(trees []fidget.TreeIndex, _ []string) error {
		if len(trees) != 1 {
			t.Fatalf("got %d outputs, want one", len(trees))
		}
		values, err := fidget.EvalGrad(fidget.BuildCanvas(trees[0]), []float32{0}, []float32{0}, []float32{0})
		if err != nil {
			return err
		}
		if got := float64(values[0].Value); math.Abs(got-want) > 2e-5 {
			t.Fatalf("native transformed field = %g, want baked preview field %g", got, want)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
