package preamble

import (
	"fmt"
	"strings"
)

// builtinGroup is a set of specialist built-ins the skill index leaves out:
// a reader asks for one by its path (`gsl skill builtin/mesh`). Every built-in
// in no group is listed in the index.
type builtinGroup struct {
	Path    string
	Summary string
	Names   []string
}

// builtinGroups are the specialist built-ins, in the order the index names
// them. A built-in belongs to at most one.
var builtinGroups = []builtinGroup{
	{
		Path:    "builtin/mesh",
		Summary: "Imported meshes and STEP bodies: importing and placing them, distances to them, mesh booleans at export, and simulated warps.",
		Names: []string{
			"importMesh", "importStep", "importWarp", "mesh", "meshWarped",
			"bvhSdf", "vdbSdf", "vdbSdfWarped", "warpField",
			"meshUnion", "meshIntersect", "meshDifference",
		},
	},
	{
		Path:    "builtin/topology",
		Summary: "Points, normals and measures read off imported STEP bodies, folded to constants at compile time.",
		Names: []string{
			"facePoint", "facePointNormal", "edgePoint", "edgePointNormal",
			"vertexPoint", "vertexPointNormal", "circleCenter", "circleCenterNormal",
			"meshVolume", "meshArea", "meshMaxThickness", "meshThickestPoint", "meshThickestPoints",
		},
	},
	{
		Path:    "builtin/noise",
		Summary: "The low-level cellular, hash and spinodoid primitives the noise and lattice libraries are built on.",
		Names: []string{
			"noiseHash3",
			"worleyF1", "worleyF2", "worleyFeature", "worleyFarFeature",
			"worleyNearCell", "worleyFarCell", "worleyWall", "worleyWallFeature",
			"gaborSpinodoid", "gaborSpinodoidGradient", "fastGaborSpinodoid", "fastGaborSpinodoidGradient",
		},
	},
	{
		Path:    "builtin/render",
		Summary: "What the renderer is doing where a color or material is shaded: the size of a pixel and how much displacement is drawn.",
		Names:   []string{"pixelSize", "displacementDrawn", "displacementScale"},
	},
}

// builtinGroupOf maps each grouped built-in to its group's path.
var builtinGroupOf = func() map[string]string {
	m := make(map[string]string)
	for _, g := range builtinGroups {
		for _, n := range g.Names {
			m[n] = g.Path
		}
	}
	return m
}()

// BuiltinGroupNames returns the paths of the specialist built-in groups.
func BuiltinGroupNames() []string {
	out := make([]string, len(builtinGroups))
	for i, g := range builtinGroups {
		out[i] = g.Path
	}
	return out
}

// builtinGroupText returns one group's documentation: its summary and each of
// its built-ins in full.
func builtinGroupText(path string) (string, bool) {
	for _, g := range builtinGroups {
		if g.Path != path {
			continue
		}
		var sb strings.Builder
		fmt.Fprintf(&sb, "### `%s`\n\n%s\n\n", g.Path, g.Summary)
		for _, n := range g.Names {
			fmt.Fprintf(&sb, "* **`%s`**: %s\n", n, builtinDescriptions[n])
		}
		return sb.String(), true
	}
	return "", false
}

// writeBuiltinGroupIndex names each group with its summary and how to read it.
func writeBuiltinGroupIndex(sb *strings.Builder) {
	sb.WriteString("\nSpecialist built-ins are left out above; read a group in full with `gsl skill GROUP`:\n\n")
	for _, g := range builtinGroups {
		fmt.Fprintf(sb, "- **`%s`** (%s): %s\n", g.Path, strings.Join(g.Names, ", "), g.Summary)
	}
}
