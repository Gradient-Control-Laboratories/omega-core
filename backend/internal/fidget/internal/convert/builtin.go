package convert

import (
	"fmt"
	"math"

	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) BuildBuiltinCall(
	op *program.Operand, call *program.FnCall,
	builtin *program.BuiltinType, args []build.Expression,
) (build.Expression, error) {
	switch call.Reference.Name {
	case "acosh":
		return d.builtinAcosh(op, args[0].(*TreeNode))
	case "asinh":
		return d.builtinAsinh(op, args[0].(*TreeNode))
	case "atanh":
		return d.builtinAtanh(op, args[0].(*TreeNode))
	case "clamp":
		return d.builtinClamp(op, args[0].(*TreeNode), args[1].(*TreeNode), args[2].(*TreeNode))
	case "cosh":
		return d.builtinCosh(op, args[0].(*TreeNode))
	case "cross":
		return d.builtinCross(op, args[0].(*TreeNode), args[1].(*TreeNode))
	case "degrees":
		return d.builtinUnaryScale(op, "degrees", 180.0/math.Pi, args[0].(*TreeNode))
	case "distance":
		return d.builtinDistance(op, args[0].(*TreeNode), args[1].(*TreeNode))
	case "dot":
		return d.builtinDot(op, args[0].(*TreeNode), args[1].(*TreeNode))
	case "exp2":
		return d.builtinExp2(op, args[0].(*TreeNode))
	case "fract":
		return d.builtinFract(op, args[0].(*TreeNode))
	case program.NoiseHash3:
		return d.builtinNoiseHash3(op, call, args[0].(*TreeNode))
	case program.WorleyF1, program.WorleyF2, program.WorleyFeature,
		program.WorleyFarFeature, program.WorleyNearCell, program.WorleyFarCell:
		return d.builtinCellular(op, call, args[0].(*TreeNode))
	case program.GaborSpinodoid, program.GaborSpinodoidGradient,
		program.FastGaborSpinodoid, program.FastGaborSpinodoidGradient:
		return d.builtinGaborSpinodoid(op, call, args)
	case program.PixelSize:
		// Meshing and export draw nothing to a screen, so there is no pixel:
		// full detail. A renderer that knows each point's footprint asks for
		// it as an input instead (pixelSizeInput), shared by every call.
		if d.pixelSizeInput {
			if !d.hasPixelSizeTree {
				d.pixelSizeTree = fidget.NewInput(program.PixelSize, 0)
				d.hasPixelSizeTree = true
			}
			return NewTreeNode(op, op.Start, d.pixelSizeTree)
		}
		return d.BuildValue(op, 0)
	case program.DisplacementDrawn:
		// Meshing and export move no surface by its material's height; the
		// lit renderer moves every one by all of it (dispDrawn).
		return d.BuildValue(op, d.dispDrawn)
	case "length":
		return d.builtinLength(op, args[0].(*TreeNode))
	case "bvhSdf":
		return d.builtinBvhSdf(op, call)
	case "vdbSdf":
		return d.builtinVdbSdf(op, call)
	case "meshUnion", "meshIntersect", "meshDifference":
		// Defense-in-depth: the canvas slicer carves these mesh-CSG
		// operators out into AssetLeaf / SdfLeaf / BoolNode well before
		// any synthetic SDF subtree reaches the fidget convert pass, so
		// this case should be unreachable in normal operation.
		return nil, fmt.Errorf(
			"%s reached fidget convert; the canvas slicer should have "+
				"stripped this mesh-CSG operator", call.Reference.Name)
	case "normalize":
		return d.builtinNormalize(op, args[0].(*TreeNode))
	case "radians":
		return d.builtinUnaryScale(op, "radians", math.Pi/180.0, args[0].(*TreeNode))
	case "sign":
		return d.builtinSign(op, args[0].(*TreeNode))
	case "sinh":
		return d.builtinSinh(op, args[0].(*TreeNode))
	case "tanh":
		return d.builtinTanh(op, args[0].(*TreeNode))
	case "trunc":
		return d.builtinTrunc(op, args[0].(*TreeNode))

	default:
		return nil, fmt.Errorf("unsupported builtin: %s", call.Reference.Name)
	}
}

// builtinNoiseHash3 lowers noiseHash3(cell) to three noise-tile oracle leaves
// (one per RGB channel), each remapped so its implicit (x,y,z) reads the cell
// coordinate. The host must have registered the baked tile (RegisterNoiseTile)
// before mesh-out; tree_noise_channel returns the sentinel u64::MAX otherwise.
//
// The optional octaves argument is a literal, so it selects which oracle leaf to
// ask for rather than becoming anything the tape evaluates.
func (d *Delegate) builtinNoiseHash3(
	op *program.Operand, call *program.FnCall, cell *TreeNode,
) (build.Expression, error) {
	c := cell.trees
	if len(c) != 3 {
		return nil, fmt.Errorf("noiseHash3: expected a 3-component cell, got %d components", len(c))
	}
	octaves, err := program.NoiseHash3Octaves(call)
	if err != nil {
		return nil, err
	}
	result := make([]fidget.TreeIndex, 3)
	for k := 0; k < 3; k++ {
		leaf := fidget.NoiseChannelTree(uint32(k), uint32(octaves))
		if leaf == ^fidget.TreeIndex(0) {
			return nil, fmt.Errorf(
				"noiseHash3: noise tile not registered — the fidget agent must call " +
					"RegisterNoiseTile before mesh-out")
		}
		result[k] = fidget.Remap(leaf, c[0], c[1], c[2])
	}
	return NewTreeNode(op, op.Start, result...)
}

// cellularComponents maps each cellular builtin to the oracle component(s) it
// reads. The point-valued ones take three consecutive components, so one leaf
// per axis; the distances take one. Numbering matches
// tape_core::cellular::component and the WGSL emitter.
var cellularComponents = map[string][]uint32{
	program.WorleyF1:         {fidget.WorleyF1},
	program.WorleyF2:         {fidget.WorleyF2},
	program.WorleyFeature:    axisComponents(fidget.WorleyFeatureX),
	program.WorleyFarFeature: axisComponents(fidget.WorleyFarFeatureX),
	program.WorleyNearCell:   axisComponents(fidget.WorleyNearCellX),
	program.WorleyFarCell:    axisComponents(fidget.WorleyFarCellX),
}

// axisComponents is the run of components a point-valued builtin reads, one per
// axis starting at first.
func axisComponents(first uint32) []uint32 {
	out := make([]uint32, fidget.WorleyAxes)
	for i := range out {
		out[i] = first + uint32(i)
	}
	return out
}

// builtinCellular lowers a cellular builtin to oracle leaves remapped so their
// implicit (x,y,z) read the lattice coordinate.
//
// One leaf per returned scalar is the whole point. The same field written as a
// 27-neighbour fold in GSL lowered to 81 hash leaves per evaluation, each of
// which could only report [0,1) once a query box straddled a lattice cell — so
// MDC could not prune until the octree had refined below the noise cell size,
// whatever the surface was doing. The oracles bound a box from its corners using
// the exact 1-Lipschitz constant of F1 and F2 instead, and prune at every depth.
// Measured on a fine-noise mesh, that is the difference between seconds and
// tens of milliseconds.
//
// Several leaves of the same call land on the same scan at the same point; the
// Rust side memoizes one scan, so asking for five components costs one.
func (d *Delegate) builtinCellular(
	op *program.Operand, call *program.FnCall, q *TreeNode,
) (build.Expression, error) {
	c := q.trees
	if len(c) != 3 {
		return nil, fmt.Errorf("%s: expected a 3-component coordinate, got %d components",
			call.Reference.Name, len(c))
	}
	octaves, err := program.NoiseHash3Octaves(call)
	if err != nil {
		return nil, err
	}
	comps, ok := cellularComponents[call.Reference.Name]
	if !ok {
		return nil, fmt.Errorf("unsupported cellular builtin: %s", call.Reference.Name)
	}
	leaves := make([]fidget.TreeIndex, len(comps))
	for i, comp := range comps {
		leaf := fidget.WorleyTree(comp, uint32(octaves))
		if leaf == ^fidget.TreeIndex(0) {
			return nil, fmt.Errorf(
				"%s: noise tile not registered — the fidget agent must call "+
					"RegisterNoiseTile before mesh-out", call.Reference.Name)
		}
		leaves[i] = fidget.Remap(leaf, c[0], c[1], c[2])
	}
	return NewTreeNode(op, op.Start, leaves...)
}

// builtinBvhSdf turns a bvhSdf(i) call into an oracle tree leaf. The asset
// triangles must already have been registered with fidget (the canvas agent
// ships them as MeshOracle payloads before dispatching the mesh-out request);
// if no oracle is registered for the index we report an error here rather than
// silently meshing empty geometry.
func (d *Delegate) builtinBvhSdf(
	op *program.Operand, call *program.FnCall,
) (build.Expression, error) {
	if len(call.Arguments) != 1 {
		return nil, fmt.Errorf("bvhSdf: expected 1 argument, got %d", len(call.Arguments))
	}
	valOp, ok := call.Arguments[0].T.(*program.Operand_Value)
	if !ok {
		return nil, fmt.Errorf("bvhSdf: argument must be a literal integer index")
	}
	if valOp.Value < 0 {
		return nil, fmt.Errorf("bvhSdf: asset index must be non-negative, got %g", valOp.Value)
	}
	assetID := uint32(valOp.Value)
	if !fidget.HasMeshOracle(assetID) {
		return nil, fmt.Errorf(
			"bvhSdf(%d): no mesh asset registered for index %d — the canvas "+
				"agent must call RegisterMeshOracle before dispatching mesh-out",
			assetID, assetID)
	}
	tree := fidget.MeshOracleTree(assetID)
	return NewTreeNode(op, op.Start, tree)
}

// builtinVdbSdf turns a vdbSdf(i) call into an OpenVDB-backed oracle tree leaf.
// The bake must already have happened on the Rust side via RegisterVdbOracle
// (the canvas agent ensures this before dispatching mesh-out). VDB and BVH
// oracles for the same asset id coexist via high-bit asset_id namespacing on
// the Rust side. Trailing vdbSdf args (voxelDivisor, halfBandVoxels) are
// bake-time hints already consumed by the canvas agent, so we only validate
// arity and read the asset index here.
func (d *Delegate) builtinVdbSdf(
	op *program.Operand, call *program.FnCall,
) (build.Expression, error) {
	if len(call.Arguments) < 1 || len(call.Arguments) > 3 {
		return nil, fmt.Errorf("vdbSdf: expected 1-3 arguments, got %d", len(call.Arguments))
	}
	valOp, ok := call.Arguments[0].T.(*program.Operand_Value)
	if !ok {
		return nil, fmt.Errorf("vdbSdf: first argument must be a literal integer index")
	}
	if valOp.Value < 0 {
		return nil, fmt.Errorf("vdbSdf: asset index must be non-negative, got %g", valOp.Value)
	}
	assetID := uint32(valOp.Value)
	if !fidget.HasVdbOracle(assetID) {
		return nil, fmt.Errorf(
			"vdbSdf(%d): no VDB asset registered for index %d — the canvas "+
				"agent must call RegisterVdbOracle before dispatching mesh-out",
			assetID, assetID)
	}
	tree := fidget.VdbOracleTree(assetID)
	return NewTreeNode(op, op.Start, tree)
}

// acosh(x) = ln(x + sqrt(x² - 1)) — element-wise
func (d *Delegate) builtinAcosh(op *program.Operand, arg *TreeNode) (*TreeNode, error) {
	v := arg.trees
	result := make([]fidget.TreeIndex, len(v))
	for i := range v {
		clampedSquare := fidget.Max(fidget.Sub(fidget.Square(v[i]), d.one), d.zero)
		result[i] = fidget.Ln(fidget.Max(fidget.Add(v[i], fidget.Sqrt(clampedSquare)), d.zero))
	}
	return NewTreeNode(op, op.Start, result...)
}

// asinh(x) = ln(x + sqrt(x² + 1)) — element-wise
func (d *Delegate) builtinAsinh(op *program.Operand, arg *TreeNode) (*TreeNode, error) {
	v := arg.trees
	result := make([]fidget.TreeIndex, len(v))
	for i := range v {
		clampedSquare := fidget.Max(fidget.Add(fidget.Square(v[i]), d.one), d.zero)
		result[i] = fidget.Ln(fidget.Max(fidget.Add(v[i], fidget.Sqrt(clampedSquare)), d.zero))
	}
	return NewTreeNode(op, op.Start, result...)
}

// atanh(x) = 0.5 * ln((1 + x) / (1 - x)) — element-wise
func (d *Delegate) builtinAtanh(op *program.Operand, arg *TreeNode) (*TreeNode, error) {
	v := arg.trees
	result := make([]fidget.TreeIndex, len(v))
	for i := range v {
		clampedInput := fidget.Max(fidget.Div(fidget.Add(d.one, v[i]), fidget.Sub(d.one, v[i])), d.zero)
		result[i] = fidget.Mul(d.half, fidget.Ln(clampedInput))
	}
	return NewTreeNode(op, op.Start, result...)
}

// clamp(x, lo, hi) = min(max(x, lo), hi) — element-wise with broadcasting
func (d *Delegate) builtinClamp(op *program.Operand, x, lo, hi *TreeNode) (*TreeNode, error) {
	cType, err := build.CommonType(x.Type(), lo.Type())
	if err != nil {
		return nil, err
	}
	cType, err = build.CommonType(cType, hi.Type())
	if err != nil {
		return nil, err
	}
	numType := toFidgetNumType(cType)
	xTrees, _ := numType.ExpandTrees(x.trees)
	loTrees, _ := numType.ExpandTrees(lo.trees)
	hiTrees, _ := numType.ExpandTrees(hi.trees)

	result := make([]fidget.TreeIndex, len(xTrees))
	for i := range xTrees {
		result[i] = fidget.Min(fidget.Max(xTrees[i], loTrees[i]), hiTrees[i])
	}
	return NewTreeNode(op, op.Start, result...)
}

// cosh(x) = (exp(x) + exp(-x)) / 2 — element-wise
func (d *Delegate) builtinCosh(op *program.Operand, arg *TreeNode) (*TreeNode, error) {
	v := arg.trees
	result := make([]fidget.TreeIndex, len(v))
	for i := range v {
		ep := fidget.Exp(v[i])
		en := fidget.Exp(fidget.Neg(v[i]))
		result[i] = fidget.Div(fidget.Add(ep, en), d.two)
	}
	return NewTreeNode(op, op.Start, result...)
}

// cross(a, b) = (a.y*b.z - a.z*b.y, a.z*b.x - a.x*b.z, a.x*b.y - a.y*b.x)
func (d *Delegate) builtinCross(op *program.Operand, a, b *TreeNode) (*TreeNode, error) {
	cType, err := build.CommonType(a.Type(), b.Type())
	if err != nil {
		return nil, err
	}
	numType := toFidgetNumType(cType)
	aTrees, _ := numType.ExpandTrees(a.trees)
	bTrees, _ := numType.ExpandTrees(b.trees)

	rx := fidget.Sub(fidget.Mul(aTrees[1], bTrees[2]), fidget.Mul(aTrees[2], bTrees[1]))
	ry := fidget.Sub(fidget.Mul(aTrees[2], bTrees[0]), fidget.Mul(aTrees[0], bTrees[2]))
	rz := fidget.Sub(fidget.Mul(aTrees[0], bTrees[1]), fidget.Mul(aTrees[1], bTrees[0]))
	return NewTreeNode(op, op.Start, rx, ry, rz)
}

// distance(a, b) = length(a - b)
func (d *Delegate) builtinDistance(op *program.Operand, a, b *TreeNode) (*TreeNode, error) {
	cType, err := build.CommonType(a.Type(), b.Type())
	if err != nil {
		return nil, err
	}
	numType := toFidgetNumType(cType)
	aTrees, _ := numType.ExpandTrees(a.trees)
	bTrees, _ := numType.ExpandTrees(b.trees)

	diff := make([]fidget.TreeIndex, len(aTrees))
	for i := range aTrees {
		diff[i] = fidget.Sub(aTrees[i], bTrees[i])
	}
	return NewTreeNode(op, op.Start, d.vecLength(diff))
}

// dot(a, b) = sum(a_i * b_i)
func (d *Delegate) builtinDot(op *program.Operand, a, b *TreeNode) (*TreeNode, error) {
	cType, err := build.CommonType(a.Type(), b.Type())
	if err != nil {
		return nil, err
	}
	numType := toFidgetNumType(cType)
	aTrees, _ := numType.ExpandTrees(a.trees)
	bTrees, _ := numType.ExpandTrees(b.trees)

	result := fidget.Mul(aTrees[0], bTrees[0])
	for i := 1; i < len(aTrees); i++ {
		result = fidget.Add(result, fidget.Mul(aTrees[i], bTrees[i]))
	}
	return NewTreeNode(op, op.Start, result)
}

// exp2(x) = exp(x * ln(2)) — element-wise
func (d *Delegate) builtinExp2(op *program.Operand, arg *TreeNode) (*TreeNode, error) {
	v := arg.trees
	ln2 := fidget.NewConst(math.Ln2)
	result := make([]fidget.TreeIndex, len(v))
	for i := range v {
		result[i] = fidget.Exp(fidget.Mul(v[i], ln2))
	}
	return NewTreeNode(op, op.Start, result...)
}

// fract(x) = x - floor(x) — element-wise
func (d *Delegate) builtinFract(op *program.Operand, arg *TreeNode) (*TreeNode, error) {
	v := arg.trees
	result := make([]fidget.TreeIndex, len(v))
	for i := range v {
		result[i] = fidget.Sub(v[i], fidget.Floor(v[i]))
	}
	return NewTreeNode(op, op.Start, result...)
}

// length(v) = sqrt(sum(v_i^2))
func (d *Delegate) builtinLength(op *program.Operand, arg *TreeNode) (*TreeNode, error) {
	return NewTreeNode(op, op.Start, d.vecLength(arg.trees))
}

// normalize(v) = v / length(v)
func (d *Delegate) builtinNormalize(op *program.Operand, arg *TreeNode) (*TreeNode, error) {
	v := arg.trees
	l := d.vecLength(v)
	result := make([]fidget.TreeIndex, len(v))
	for i := range v {
		result[i] = fidget.Div(v[i], l)
	}
	return NewTreeNode(op, op.Start, result...)
}

// sign(x) = compare(x, 0) — element-wise
func (d *Delegate) builtinSign(op *program.Operand, arg *TreeNode) (*TreeNode, error) {
	v := arg.trees
	result := make([]fidget.TreeIndex, len(v))
	for i := range v {
		result[i] = fidget.Compare(v[i], d.zero)
	}
	return NewTreeNode(op, op.Start, result...)
}

// sinh(x) = (exp(x) - exp(-x)) / 2 — element-wise
func (d *Delegate) builtinSinh(op *program.Operand, arg *TreeNode) (*TreeNode, error) {
	v := arg.trees
	result := make([]fidget.TreeIndex, len(v))
	for i := range v {
		ep := fidget.Exp(v[i])
		en := fidget.Exp(fidget.Neg(v[i]))
		result[i] = fidget.Div(fidget.Sub(ep, en), d.two)
	}
	return NewTreeNode(op, op.Start, result...)
}

// tanh(x) = (exp(x) - exp(-x)) / (exp(x) + exp(-x)) — element-wise
func (d *Delegate) builtinTanh(op *program.Operand, arg *TreeNode) (*TreeNode, error) {
	v := arg.trees
	result := make([]fidget.TreeIndex, len(v))
	for i := range v {
		ep := fidget.Exp(v[i])
		en := fidget.Exp(fidget.Neg(v[i]))
		result[i] = fidget.Div(fidget.Sub(ep, en), fidget.Add(ep, en))
	}
	return NewTreeNode(op, op.Start, result...)
}

// trunc(x) = sign(x) * floor(abs(x)) — element-wise
func (d *Delegate) builtinTrunc(op *program.Operand, arg *TreeNode) (*TreeNode, error) {
	v := arg.trees
	result := make([]fidget.TreeIndex, len(v))
	for i := range v {
		result[i] = fidget.Mul(fidget.Compare(v[i], d.zero), fidget.Floor(fidget.Abs(v[i])))
	}
	return NewTreeNode(op, op.Start, result...)
}

// builtinUnaryScale multiplies each component by a constant — used for degrees and radians
func (d *Delegate) builtinUnaryScale(op *program.Operand, name string, factor float64, arg *TreeNode) (*TreeNode, error) {
	v := arg.trees
	k := fidget.NewConst(factor)
	result := make([]fidget.TreeIndex, len(v))
	for i := range v {
		result[i] = fidget.Mul(v[i], k)
	}
	return NewTreeNode(op, op.Start, result...)
}

// vecLength computes sqrt(sum(v_i^2))
func (d *Delegate) vecLength(v []fidget.TreeIndex) fidget.TreeIndex {
	sum := fidget.Square(v[0])
	for i := 1; i < len(v); i++ {
		sum = fidget.Add(sum, fidget.Square(v[i]))
	}
	// Clamp sum to >= 0 to prevent NaN bounds during unconditional interval evaluations
	return fidget.Sqrt(fidget.Max(sum, d.zero))
}
