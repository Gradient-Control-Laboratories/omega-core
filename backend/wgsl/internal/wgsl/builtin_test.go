package wgsl

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

func TestSanityCheck(t *testing.T) {
	for name := range wgslBuiltins {
		modelBuiltin := program.BuiltinNamed(name)
		if modelBuiltin == nil {
			t.Errorf("builtin %s not found in program.Builtins", name)
		}
	}
	for _, name := range program.Builtins() {
		// mesh() is a scene declaration handled by the canvas agent before
		// the Module reaches the WGSL agent — it has no WGSL representation.
		// bvhSdf() / vdbSdf() sample a 3D distance-field texture bound by
		// the canvas agent; their codegen path lives in mesh_sdf.go, not
		// in wgslBuiltins. meshDifference has a custom emitter in
		// codegen/function_call.go (lowers to max(L, -R)).
		// noiseHash3 samples the baked noise tile; its codegen lives in
		// function_call.go (a vec3 expression), not in wgslBuiltins.
		// UI annotations are stripped by common/build before codegen.
		// and the mesh-property queries (meshVolume) are folded to constants
		// by the canvas agent's early pass and never reach WGSL codegen.
		// meshWarped / vdbSdfWarped are rewritten to mesh() / vdbSdf() over a
		// derived asset by the canvas agent, for the same reason the topology
		// builtins are folded there: the warp is baked into geometry, so by the
		// time a module reaches any backend the warped forms are gone.
		if name == "mesh" || name == "importStep" || name == "importMesh" || name == "importWarp" || name == "bvhSdf" || name == "vdbSdf" || name == "meshDifference" || name == program.NoiseHash3 || name == program.PixelSize || name == program.DisplacementDrawn ||
			// The worley* builtins lower to the gsl_cellularScan helper in
			// codegen/cellular.go, not to a wgslBuiltins alias.
			name == program.WorleyF1 || name == program.WorleyF2 ||
			name == program.WorleyFeature || name == program.WorleyFarFeature ||
			name == program.WorleyNearCell ||
			name == program.WorleyFarCell ||
			// worleyWall / worleyWallFeature lower to gsl_cellularWallScan,
			// beside it in codegen/cellular.go.
			name == program.WorleyWall || name == program.WorleyWallFeature ||
			// gaborSpinodoid lowers to the gsl_spinodoid helper in
			// codegen/spinodoid.go.
			program.IsGaborSpinodoid(name) ||
			// erfinv takes only a constant and is folded by the compiler, so
			// it never reaches codegen (a density's spinodoid level, mostly).
			name == "erfinv" ||
			name == "meshWarped" || name == "vdbSdfWarped" ||
			name == "uiPoint" || name == "uiTransform" || name == "uiSpline" || name == "uiColor" || name == "uiDirection" ||
			name == "facePoint" || name == "facePointNormal" ||
			name == "edgePoint" || name == "edgePointNormal" ||
			name == "vertexPoint" || name == "vertexPointNormal" ||
			name == "circleCenter" || name == "circleCenterNormal" ||
			name == "meshVolume" || name == "meshArea" || name == "meshMaxThickness" ||
			name == "meshThickestPoint" || name == "meshThickestPoints" ||
			name == "warpField" {
			continue
		}
		modelBuiltin := program.BuiltinNamed(name)
		wgslBuiltin, ok := wgslBuiltins[name]
		if !ok {
			t.Errorf("builtin %s not found in wgslBuiltins", name)
			continue
		}
		for _, wgslOverload := range wgslBuiltin.Overloads {
			if len(wgslOverload.ParamTypes) < modelBuiltin.MinArgs || len(wgslOverload.ParamTypes) > modelBuiltin.MaxArgs {
				t.Errorf("builtin %s has overload with %d params, which is not between %d and %d",
					name, len(wgslOverload.ParamTypes), modelBuiltin.MinArgs, modelBuiltin.MaxArgs)
			}
		}
	}
}
