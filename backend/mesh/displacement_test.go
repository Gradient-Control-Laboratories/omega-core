package mesh_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
	"github.com/Gradient-Control-Laboratories/omega-core/backend/mesh"
)

// A 20 mm cube with a 0.1 mm build, laid two ways: as a coat over the cube
// (a layered shape, moved by its displaced distance) and as the cube's own
// material (a whole return, moved by its height). Neither material has
// relief.
var coatedBlocks = map[string]string{
	"layered coat": `
blank = boxCentered(vector.zero, [20.0, 20.0, 20.0]);
blank.material = material.make(0.5, 0, 1.5);
coat = blank;
coat.material = material.withBuild(material.make(0.3, 0, 1.5), 0.1);
return coat;`,
	"whole return": `
whole = boxCentered(vector.zero, [20.0, 20.0, 20.0]);
whole.material = material.withBuild(material.make(0.3, 0, 1.5), 0.1);
return whole;`,
}

func meshCoatedBlock(t *testing.T, src string, opts ...mesh.Option) []byte {
	t.Helper()
	res, err := assembler.Assemble([]string{src})
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	base := []mesh.Option{
		mesh.WithSubdivision(8),
		mesh.WithAlgorithm("mdc"),
		mesh.WithBbox([6]float32{-11, -11, -11, 11, 11, 11}),
	}
	stl, err := mesh.Mesh(res.Program, append(base, opts...)...)
	if err != nil {
		t.Fatalf("Mesh: %v", err)
	}
	return stl
}

// TestMesh_BuildGrowsACoatedBlock: meshed with its build, the block's volume
// grows by its area times the build, as the cube offset by 0.1 mm encloses,
// a^3 + 6 a^2 b + 3 pi a b^2 + 4/3 pi b^3; with relief alone, which these
// materials do not have, it stays the bare cube's.
func TestMesh_BuildGrowsACoatedBlock(t *testing.T) {
	const a, b = 20.0, 0.1
	bare := a * a * a
	built := bare + 6*a*a*b + 3*math.Pi*a*b*b + 4.0/3.0*math.Pi*b*b*b
	for name, src := range coatedBlocks {
		t.Run(name, func(t *testing.T) {
			for _, c := range []struct {
				build, relief bool
				want          float64
			}{
				{false, false, bare},
				{true, false, built},
				{false, true, bare},
				{true, true, built},
			} {
				stats, err := mesh.MeasureSTL(meshCoatedBlock(t, src, mesh.WithDisplacement(c.build, c.relief)))
				if err != nil {
					t.Fatalf("MeasureSTL: %v", err)
				}
				if rel := math.Abs(stats.SignedVolume-c.want) / c.want; rel > 1e-4 {
					t.Errorf("build %v relief %v: volume %.3f, want %.3f (relative error %.2g)",
						c.build, c.relief, stats.SignedVolume, c.want, rel)
				}
			}
		})
	}
}

// TestMesh_DisplacementOffIsTheFieldAsBuilt: with both flags off the mesher
// takes the path it always took, so the mesh is byte for byte the one meshed
// with no displacement option at all.
func TestMesh_DisplacementOffIsTheFieldAsBuilt(t *testing.T) {
	for name, src := range coatedBlocks {
		t.Run(name, func(t *testing.T) {
			plain := meshCoatedBlock(t, src)
			off := meshCoatedBlock(t, src, mesh.WithDisplacement(false, false))
			if !bytes.Equal(plain, off) {
				t.Errorf("meshed with both flags off: %d bytes, differing from the %d meshed without the option",
					len(off), len(plain))
			}
		})
	}
}
