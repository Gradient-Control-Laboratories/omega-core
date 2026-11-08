package codegen

import (
	"fmt"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// MeshSdfSlots tracks which imported mesh indices the current Module
// references via bvhSdf(i) or vdbSdf(i), and what densified slot number
// each was assigned. Slots are allocated in the order unique indices are
// first seen during a module walk — e.g. a program that uses bvhSdf(3)
// and bvhSdf(7) gets slots 0 and 1 respectively. Keeping slots contiguous
// keeps the browser-side binding layout tight regardless of how the user
// chose to index imports.
type MeshSdfSlots struct {
	byAssetIndex map[int]int // asset index → slot
	ordered      []*program.MeshSdfSlot
}

// NewMeshSdfSlotsFromProto converts pre-allocated protobuf MeshSdfSlot records
// into MeshSdfSlots without walking the AST or re-collecting mesh references.
func NewMeshSdfSlotsFromProto(protoSlots []*program.MeshSdfSlot) *MeshSdfSlots {
	s := &MeshSdfSlots{
		byAssetIndex: make(map[int]int),
	}
	for _, ps := range protoSlots {
		if ps == nil {
			continue
		}
		slot := int(ps.Slot)
		assetIdx := int(ps.AssetIndex)
		s.byAssetIndex[assetIdx] = slot
		s.ordered = append(s.ordered, ps)
	}
	return s
}

// Slot returns the shader binding slot number assigned to the given asset
// index, or (-1, false) if that index is never referenced by bvhSdf() or
// vdbSdf() in the module. Callers should report an error at the call site
// if this is false.
func (s *MeshSdfSlots) Slot(assetIndex int) (int, bool) {
	slot, ok := s.byAssetIndex[assetIndex]
	return slot, ok
}

// Ordered returns the slot-to-asset mapping in slot order (slot 0, slot 1, …).
// Intended for populating RenderResponse.MeshSdfSlots.
func (s *MeshSdfSlots) Ordered() []*program.MeshSdfSlot {
	return s.ordered
}

// Len returns the number of distinct mesh SDF slots referenced.
func (s *MeshSdfSlots) Len() int { return len(s.ordered) }

// SdShapeParamsSuffix returns the comma-prefixed parameter list that must be
// appended to the generated sdShape signature, one group per slot:
//
//	, sdfTex0: texture_3d<f32>, sdfSmp0: sampler, sdfBMin0: vec3<f32>, sdfBSize0: vec3<f32>
//	, sdfTex1: texture_3d<f32>, sdfSmp1: sampler, sdfBMin1: vec3<f32>, sdfBSize1: vec3<f32>
//
// Empty when no slots are present (no extra params added).
func (s *MeshSdfSlots) SdShapeParamsSuffix() string {
	if len(s.ordered) == 0 {
		return ""
	}
	var sb []byte
	for slot := 0; slot < len(s.ordered); slot++ {
		sb = append(sb, fmt.Appendf(nil,
			", sdfTex%d: texture_3d<f32>, sdfSmp%d: sampler, sdfBMin%d: vec3<f32>, sdfBSize%d: vec3<f32>",
			slot, slot, slot, slot)...)
	}
	return string(sb)
}

// BuildMeshSdfCall is the codegen for a bvhSdf(index) or
// vdbSdf(index, ...) builtin call. It returns an Expression whose Codegen
// emits two let-bindings at function scope (bbox-relative uvw + sampled
// distance with out-of-bbox sentinel) and evaluates to the sampled f32
// value.
//
// Precondition: the index argument is a literal non-negative integer, which
// the parser (buildBvhSdfCall / buildVdbSdfCall in
// agent/parser/internal/parser/build.go) already guarantees. The
// asset-index → slot mapping is looked up on the Delegate, which was
// populated at Module build time.
type MeshSdfExpr struct {
	wgsl.Expression
	slot int
}

func NewMeshSdfExpr(op *program.Operand, slot int) *MeshSdfExpr {
	m := &MeshSdfExpr{slot: slot}
	m.Expression = wgsl.NewExpr(op, wgsl.TypeF32, build.PrecedenceHighest)
	m.Expression.SetOuter(m)
	return m
}

func (m *MeshSdfExpr) Codegen(ctx wgsl.CodegenContext) string {
	slot := m.slot
	pName := ctx.P().WGSLName()
	bMin := fmt.Sprintf("sdfBMin%d", slot)
	bSize := fmt.Sprintf("sdfBSize%d", slot)
	tex := fmt.Sprintf("sdfTex%d", slot)
	smp := fmt.Sprintf("sdfSmp%d", slot)

	uvwVar := ctx.AssignUnique("gsl_msUvw",
		fmt.Sprintf("(%s - %s) / %s", pName, bMin, bSize), pName, bMin, bSize)

	// The texture is clamped-sampled so gsl_msSample is well-defined
	// everywhere: inside the bbox it is the real signed distance (can
	// go negative); outside the bbox it is the boundary-cell value
	// (always positive because the bake pads the bbox by 20% so the
	// mesh never touches the grid boundary).
	sampleVar := ctx.AssignUnique("gsl_msSample", fmt.Sprintf(
		"textureSampleLevel(%s, %s, clamp(%s, vec3<f32>(0.0), vec3<f32>(1.0)), 0.0).r",
		tex, smp, uvwVar), tex, smp, uvwVar)

	// gsl_msOutside = Euclidean distance from p to the bbox (0 when
	// inside). Used to extend the distance field analytically past the
	// bake region. An earlier version returned max(gsl_msOutside,
	// gsl_msSample) here — a loose lower bound that caused the
	// offset-iso-surface to bulge axis-aligned along the bbox faces
	// and converge to a rounded cube for large offsets.
	outsideVar := ctx.AssignUnique("gsl_msOutside", fmt.Sprintf(
		"length(max(vec3<f32>(0.0), max(%s - %s, %s - (%s + %s))))",
		bMin, pName, pName, bMin, bSize), bMin, pName, bSize)

	// Outside the bbox we extend the SDF by the triangle inequality:
	// d(p, mesh) ≤ d(p, q) + d(q, mesh), where q is the closest bbox
	// point to p. For convex meshes this bound is *exact* (a unit
	// sphere mesh at origin with p on +x gives d = |p| - 1 = d(p, q) +
	// d(q, mesh)), so offset surfaces like `bvhSdf(0) - 0.5` render
	// as the true inflated shape instead of the rounded-cube artifact
	// the earlier `max()` form produced. For non-convex meshes the
	// bound is an upper bound, so a ray tracing past the bbox could
	// overshoot the true surface — but the Quilez gradient correction
	// handles the Lipschitz-1 gradient (gradient is a unit vector
	// pointing outward from q plus a constant boundary-sample term)
	// without issue, and the accuracy gain on typical mesh shapes is
	// much more important than that edge case. Inside the bbox we
	// still use the raw sample so negative (in-mesh) values come
	// through.
	msVar := ctx.AssignUnique("gsl_msVal", fmt.Sprintf(
		"select(%s + %s, %s, %s == 0.0)",
		outsideVar, sampleVar, sampleVar, outsideVar), outsideVar, sampleVar)
	return msVar
}

// BuildMeshSdfCall is the codegen for a bvhSdf(index) or
// vdbSdf(index, ...) builtin call. It returns an Expression whose Codegen
// emits two let-bindings at function scope (bbox-relative uvw + sampled
// distance with out-of-bbox sentinel) and evaluates to the sampled f32
// value.
//
// Precondition: the index argument is a literal non-negative integer, which
// the parser (buildBvhSdfCall / buildVdbSdfCall in
// agent/parser/internal/parser/build.go) already guarantees. The
// asset-index → slot mapping is looked up on the Delegate, which was
// populated at Module build time.
func (d *Delegate) BuildMeshSdfCall(
	op *program.Operand, fnCall *program.FnCall,
	expressions []build.Expression,
) (build.Expression, error) {
	// vdbSdf shares this codegen path and may carry up to 3 arguments
	// (idx, voxelDivisor, halfBandVoxels); the trailing two are bake
	// hints consumed by the canvas agent and are irrelevant to the
	// WGSL preview sampler, which only needs the asset index.
	if len(fnCall.Arguments) < 1 || len(fnCall.Arguments) > 3 {
		return nil, fmt.Errorf("internal error: %s expected 1-3 arguments, got %d", fnCall.Reference.Name, len(fnCall.Arguments))
	}
	valOp, ok := fnCall.Arguments[0].T.(*program.Operand_Value)
	if !ok {
		return nil, fmt.Errorf("%s() first argument must be a literal integer index", fnCall.Reference.Name)
	}
	assetIndex := int(valOp.Value)
	if d.meshSdfSlots == nil {
		return nil, fmt.Errorf("internal error: mesh SDF slots not initialized")
	}
	slot, ok := d.meshSdfSlots.Slot(assetIndex)
	if !ok {
		return nil, fmt.Errorf("internal error: %s(%d) has no assigned slot", fnCall.Reference.Name, assetIndex)
	}

	return NewMeshSdfExpr(op, slot), nil
}
