package convert_test

import (
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/fidget/internal/convert"
	"gradientcontrol.com/omega/core/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/internal/fidget/internal/mesh"
	noise "gradientcontrol.com/omega/core/internal/noise"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// Worley noise, backed by the real baked tile oracle, must actually mesh in
// Fidget — the capability the placeholder float hash could not provide (it
// exploded the interval bounds). A cluster of Worley blobs clipped to a ball
// gives a bounded solid that should produce a non-empty mesh on both algorithms.
func TestWorleyMeshesViaNoiseOracle(t *testing.T) {
	require.NoError(t, fidget.RegisterNoiseTile(noise.Bake()))

	const src = `import "noise";
return max(noise.cellular(3) - 1.5, length(p) - 4);`

	res, err := assembler.Assemble([]string{src}, mockMgr)
	require.NoError(t, err)

	bbox := [6]float32{-5, -5, -5, 5, 5, 5}

	err = convert.ToTrees(res.Program, func(trees []fidget.TreeIndex, _ []string) error {
		require.NotEmpty(t, trees)
		for _, algo := range []string{"marching-cubes", "mdc"} {
			stl, e := mesh.GenerateSTL(trees, 6, bbox, algo, func(uint32, uint32, string) {})
			require.NoError(t, e, "GenerateSTL(%s)", algo)
			require.NotEmpty(t, stl, "empty mesh from %s", algo)
		}
		return nil
	})
	require.NoError(t, err)
}

// The Cellular Noise example remaps noise fields to different positions and
// unions them. remap rebinds the noise oracle's input coordinates, so verify
// that a remapped cellular field still meshes (the path the example exercises).
func TestRemappedCellularMeshes(t *testing.T) {
	require.NoError(t, fidget.RegisterNoiseTile(noise.Bake()))

	const src = `import "noise";
edges = max(noise.cellularEdges(4) - 0.6, length(p) - 4);
return min(remap(edges, [p.x + 6, p.y, p.z]), remap(edges, [p.x - 6, p.y, p.z]));`

	res, err := assembler.Assemble([]string{src}, mockMgr)
	require.NoError(t, err)

	bbox := [6]float32{-12, -8, -8, 12, 8, 8}

	err = convert.ToTrees(res.Program, func(trees []fidget.TreeIndex, _ []string) error {
		require.NotEmpty(t, trees)
		stl, e := mesh.GenerateSTL(trees, 6, bbox, "mdc", func(uint32, uint32, string) {})
		require.NoError(t, e)
		require.NotEmpty(t, stl, "empty mesh")
		return nil
	})
	require.NoError(t, err)
}
