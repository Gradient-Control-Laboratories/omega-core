package codegen_test

import (
	"strings"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// TestImportPathSanitizedInWGSL — a package imported from a multi-segment path
// such as "lib/shapes" captures top-level constants as WGSL function
// parameters. The '/' in the path must not leak into the generated identifier
// (which would be invalid WGSL: "expected ':' for parameter").
func TestImportPathSanitizedInWGSL(t *testing.T) {
	srcs := map[string][]string{
		// radius is a top-level constant; blob references it, so when blob is
		// emitted as a WGSL function the constant rides along as a parameter
		// named from the package path.
		"lib/shapes": {`radius = 2;
blob(x) { return length(p) - radius - x; }
`},
	}
	primary := `import . "lib/shapes";
// remap prevents inlining so blob is emitted as a real function
v = remap(blob(1.0), [p.x, p.y, p.z]);
return v;
`
	mgr := &source.MockManager{PkgSources: srcs}
	res, err := assembler.Assemble([]string{primary}, mgr)
	require.NoError(t, err)

	render, err := renderHelper(primary, res.Program)
	require.NoError(t, err)

	// The captured constant must use a sanitized prefix, never the raw path.
	require.False(t, strings.Contains(render.WGSL, "lib/shapes_radius"))
	require.True(t, strings.Contains(render.WGSL, "lib_shapes_radius"))
}

func TestMultipleDotImports(t *testing.T) {
	srcs := map[string][]string{
		"mybuiltin":                   {`omegaRadius = 1.0; sphere(r) { return length(p) - r; }`},
		"digitalmetal.io/common/base": {`baseRadius = 2.0;`},
	}
	primary := `import . "mybuiltin";
import . "digitalmetal.io/common/base";
return sphere(omegaRadius + baseRadius);
`
	mgr := &source.MockManager{PkgSources: srcs}
	res, err := assembler.Assemble([]string{primary}, mgr)
	require.NoError(t, err)

	_, err = renderHelper(primary, res.Program)
	require.NoError(t, err)
}
