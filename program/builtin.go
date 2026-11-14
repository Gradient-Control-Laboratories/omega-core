package program

import "fmt"

type BuiltinType struct {
	Name    string
	MinArgs int
	MaxArgs int

	// Reduce, when not NoReduce, marks a binary builtin that also accepts
	// lists of meshes and folds them left into a chain of binary calls at
	// build time, so the backends only ever see the binary form. See
	// ReduceMode.
	Reduce ReduceMode
}

// ReduceMode describes how a reducing mesh-CSG builtin collapses a list of
// meshes into a left fold of binary calls.
type ReduceMode int

const (
	// NoReduce is an ordinary builtin with no list-folding behavior.
	NoReduce ReduceMode = iota

	// ReduceAll folds every operand together. Both the variadic form
	// meshUnion(a, b, c) and the single-list form meshUnion([a, b, c]) are
	// accepted and lower to meshUnion(meshUnion(a, b), c).
	ReduceAll

	// ReduceTail takes exactly two operands: a base mesh and a second operand
	// that may be a single mesh or a list of meshes. The base is folded
	// against each element of the second operand, so meshDifference(a, [b, c])
	// lowers to meshDifference(meshDifference(a, b), c).
	ReduceTail
)

func Builtins() []string {
	builtins := make([]string, 0, len(builtinsMap))
	for name := range builtinsMap {
		builtins = append(builtins, name)
	}
	return builtins
}

func IsBuiltin(name string) bool {
	return BuiltinNamed(name) != nil
}

// Return the BuiltinNamed type if it exists, otherwise return nil
func BuiltinNamed(name string) *BuiltinType {
	return builtinsMap[name]
}

var builtinsMap = map[string]*BuiltinType{

	// GSL operations

	"acosh":     {Name: "acosh", MinArgs: 1, MaxArgs: 1},
	"asinh":     {Name: "asinh", MinArgs: 1, MaxArgs: 1},
	"atanh":     {Name: "atanh", MinArgs: 1, MaxArgs: 1},
	"clamp":     {Name: "clamp", MinArgs: 3, MaxArgs: 3},
	"cosh":      {Name: "cosh", MinArgs: 1, MaxArgs: 1},
	"cross":     {Name: "cross", MinArgs: 2, MaxArgs: 2},
	"degrees":   {Name: "degrees", MinArgs: 1, MaxArgs: 1},
	"distance":  {Name: "distance", MinArgs: 2, MaxArgs: 2},
	"dot":       {Name: "dot", MinArgs: 2, MaxArgs: 2},
	"erfinv":    {Name: "erfinv", MinArgs: 1, MaxArgs: 1},
	"exp2":      {Name: "exp2", MinArgs: 1, MaxArgs: 1},
	"fract":     {Name: "fract", MinArgs: 1, MaxArgs: 1},
	"length":    {Name: "length", MinArgs: 1, MaxArgs: 1},
	"normalize": {Name: "normalize", MinArgs: 1, MaxArgs: 1},
	"radians":   {Name: "radians", MinArgs: 1, MaxArgs: 1},
	"sign":      {Name: "sign", MinArgs: 1, MaxArgs: 1},
	"sinh":      {Name: "sinh", MinArgs: 1, MaxArgs: 1},
	"tanh":      {Name: "tanh", MinArgs: 1, MaxArgs: 1},
	"trunc":     {Name: "trunc", MinArgs: 1, MaxArgs: 1},

	// Mesh operations

	// noiseHash3(cell) folds two tile lookups; noiseHash3(cell, octaves) takes
	// the octave count as a literal 1 or 2 (see the description in preamble).
	NoiseHash3: {Name: NoiseHash3, MinArgs: 1, MaxArgs: 2},

	// cellularF1(q) / cellularF1(q, octaves): F1 on the unit cellular
	// lattice at coordinate q, and the feature point it measures to.
	WorleyF1:         {Name: WorleyF1, MinArgs: 1, MaxArgs: 2},
	WorleyF2:         {Name: WorleyF2, MinArgs: 1, MaxArgs: 2},
	WorleyFeature:    {Name: WorleyFeature, MinArgs: 1, MaxArgs: 2},
	WorleyNearCell:   {Name: WorleyNearCell, MinArgs: 1, MaxArgs: 2},
	WorleyFarCell:    {Name: WorleyFarCell, MinArgs: 1, MaxArgs: 2},
	WorleyFarFeature: {Name: WorleyFarFeature, MinArgs: 1, MaxArgs: 2},
	// worleyWall(q) / worleyWall(q, octaves): the exact distance to the nearest
	// Voronoi wall, and the feature point whose wall it is.
	WorleyWall:        {Name: WorleyWall, MinArgs: 1, MaxArgs: 2},
	WorleyWallFeature: {Name: WorleyWallFeature, MinArgs: 1, MaxArgs: 2},

	// gaborSpinodoid(q, period, theta, poles, packetWidth, impulses, pool,
	// seed, bias, localNormalize, scale, orientation) and its gradient.
	// Everything but q, scale and orientation must fold to a constant: it
	// keys the baked kernel table.
	GaborSpinodoid:         {Name: GaborSpinodoid, MinArgs: GaborSpinodoidArgs, MaxArgs: GaborSpinodoidArgs},
	GaborSpinodoidGradient: {Name: GaborSpinodoidGradient, MinArgs: GaborSpinodoidArgs, MaxArgs: GaborSpinodoidArgs},

	// fastGaborSpinodoid and its gradient: the lattice field, same arguments.
	FastGaborSpinodoid:         {Name: FastGaborSpinodoid, MinArgs: GaborSpinodoidArgs, MaxArgs: GaborSpinodoidArgs},
	FastGaborSpinodoidGradient: {Name: FastGaborSpinodoidGradient, MinArgs: GaborSpinodoidArgs, MaxArgs: GaborSpinodoidArgs},

	// The world-space size of one pixel where a color or material is being
	// evaluated for display; 0 anywhere else.
	PixelSize: {Name: PixelSize, MinArgs: 0, MaxArgs: 0},

	// How much of a material's displacement is drawn where a color or
	// material is being evaluated for display, 0 to 1; 0 anywhere else.
	DisplacementDrawn: {Name: DisplacementDrawn, MinArgs: 0, MaxArgs: 0},

	"uiPoint":     {Name: "uiPoint", MinArgs: 1, MaxArgs: 1},
	"uiTransform": {Name: "uiTransform", MinArgs: 1, MaxArgs: 1},
	"uiSpline":    {Name: "uiSpline", MinArgs: 1, MaxArgs: 1},
	"uiColor":     {Name: "uiColor", MinArgs: 1, MaxArgs: 1},
	"uiDirection": {Name: "uiDirection", MinArgs: 1, MaxArgs: 2}, // (direction[, anchor])

	// meshWarped / vdbSdfWarped are the warped forms of mesh / vdbSdf. Both
	// take (meshSlot, warpSlot, scale) ahead of whatever the unwarped form
	// takes, so their trailing arguments read identically to it. The canvas
	// agent bakes the displacement into a derived mesh asset and rewrites the
	// calls to the plain forms, so nothing downstream ever sees them.
	"mesh":           {Name: "mesh", MinArgs: 1, MaxArgs: 4},
	"importStep":     {Name: "importStep", MinArgs: 1, MaxArgs: 2},
	"importMesh":     {Name: "importMesh", MinArgs: 1, MaxArgs: 2},
	"importWarp":     {Name: "importWarp", MinArgs: 1, MaxArgs: 2},
	"bvhSdf":         {Name: "bvhSdf", MinArgs: 1, MaxArgs: 1},
	"vdbSdf":         {Name: "vdbSdf", MinArgs: 1, MaxArgs: 3},
	"meshWarped":     {Name: "meshWarped", MinArgs: 3, MaxArgs: 6},
	"vdbSdfWarped":   {Name: "vdbSdfWarped", MinArgs: 3, MaxArgs: 5},
	"meshUnion":      {Name: "meshUnion", MinArgs: 1, MaxArgs: 100, Reduce: ReduceAll},
	"meshIntersect":  {Name: "meshIntersect", MinArgs: 1, MaxArgs: 100, Reduce: ReduceAll},
	"meshDifference": {Name: "meshDifference", MinArgs: 2, MaxArgs: 2, Reduce: ReduceTail},

	// Topology-reference builtins: resolve a B-rep entity on an annotated
	// (STEP-imported) body to constant world-frame geometry. The *Normal
	// variants return two values (point, normal). All are folded to constants
	// by the canvas agent's server-side early pass (webai canvas
	// topo_fold.go) before any backend sees the module.

	"facePoint":          {Name: "facePoint", MinArgs: 3, MaxArgs: 4},
	"facePointNormal":    {Name: "facePointNormal", MinArgs: 3, MaxArgs: 4},
	"edgePoint":          {Name: "edgePoint", MinArgs: 3, MaxArgs: 5},
	"edgePointNormal":    {Name: "edgePointNormal", MinArgs: 3, MaxArgs: 5},
	"vertexPoint":        {Name: "vertexPoint", MinArgs: 2, MaxArgs: 3},
	"vertexPointNormal":  {Name: "vertexPointNormal", MinArgs: 2, MaxArgs: 3},
	"circleCenter":       {Name: "circleCenter", MinArgs: 2, MaxArgs: 3},
	"circleCenterNormal": {Name: "circleCenterNormal", MinArgs: 2, MaxArgs: 3},

	// Mesh-property queries, folded to constants by the same canvas early
	// pass against the imported mesh registry (STL or STEP).

	"meshVolume":         {Name: "meshVolume", MinArgs: 1, MaxArgs: 1},
	"meshArea":           {Name: "meshArea", MinArgs: 1, MaxArgs: 1},
	"meshMaxThickness":   {Name: "meshMaxThickness", MinArgs: 1, MaxArgs: 1},
	"meshThickestPoint":  {Name: "meshThickestPoint", MinArgs: 1, MaxArgs: 2},
	"meshThickestPoints": {Name: "meshThickestPoints", MinArgs: 1, MaxArgs: 5},

	// Simulated displacement field, sampled at a position. Unlike the mesh SDF
	// oracles this is not per-asset: a simulation produces one field
	// for the whole scene, so there is no index to take.

	"warpField": {Name: "warpField", MinArgs: 1, MaxArgs: 1},
}

// NoiseHash3Octaves reads the octave count out of a noiseHash3 call: how many
// tile lookups the hash folds together. One argument means the two-lookup
// default (fine plus supercell, repeating every TileSize*TileSize cells); a
// second argument chooses explicitly, and must be a literal so every backend
// can pick its lowering at compile time rather than branching per evaluation.
//
// Both backends call this so the accepted values and the error text cannot
// drift apart.
func NoiseHash3Octaves(call *FnCall) (int, error) {
	if call == nil || len(call.Arguments) < 2 {
		return 2, nil
	}
	valOp, ok := call.Arguments[1].T.(*Operand_Value)
	if !ok {
		return 0, fmt.Errorf("noiseHash3: octaves must be a literal 1 or 2")
	}
	switch valOp.Value {
	case 1:
		return 1, nil
	case 2:
		return 2, nil
	}
	return 0, fmt.Errorf("noiseHash3: octaves must be 1 or 2, got %g", valOp.Value)
}
