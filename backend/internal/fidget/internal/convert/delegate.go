package convert

import (
	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// Delegate implements the [build.Delegate] interface.
type Delegate struct {
	gslSource  string
	rootBScope *build.Scope
	currBScope *build.Scope

	// Cached tree indices for common values
	// These must be initialized in BuildModule() after fidget.Clear()
	two     fidget.TreeIndex
	one     fidget.TreeIndex
	half    fidget.TreeIndex
	epsilon fidget.TreeIndex // 1e-12
	zero    fidget.TreeIndex
	negOne  fidget.TreeIndex

	// list of tree indices that are outputs
	outputs []*TreeNode

	// spinodoid oracle slots registered this build, by table and leveling
	// (see builtinGaborSpinodoid).
	spinodoidSlots    map[string]uint32
	spinodoidNextSlot uint32

	// spinodoid tables baked this build, by spinodoid.Call key.
	spinodoidBakes  map[string]BakedSpinodoid
	spinodoidDomain *[6]float64

	// spinodoidOracles describes every spinodoid table bound this build, one
	// per slot, for a tape export to stamp into tape.oracles. A slot's leaf id
	// means nothing outside this process, so the arguments have to travel with
	// it — see SpinodoidOracle.
	spinodoidOracles []SpinodoidOracle

	// per-output display names (parallel to outputs), matching the UI's
	// per-shape labels — see outputName.
	outputNames []string

	// per-return color fields as (r, g, b) trees, with hasColorReturn
	// marking which returns authored one. Both are indexed by RETURN, which
	// is not the same as outputs when a return carries several trees — see
	// BuildPackageColorReturns and ToColorTrees.
	colorReturns   [][3]fidget.TreeIndex
	hasColorReturn []bool

	// per-return material fields as three rows of four trees, with
	// hasMaterialReturn marking which returns authored one. Indexed by RETURN
	// like the color fields, and independent of them: a return may have either
	// property, both or neither.
	materialReturns   [][materialRows][materialRowSize]fidget.TreeIndex
	hasMaterialReturn []bool
	// displacedReturn marks the returns whose displacement (volume.z) is not
	// known to be the constant zero, so a renderer knows which surfaces move.
	displacedReturn []bool

	// outputReturnIdx maps each output back to the return it came from, so
	// per-return data (color, material) can be attached to the flattened
	// outputs.
	outputReturnIdx []int

	// The implied p and time vars, retained so BuildPackage can re-point them
	// at live trees after it clears the arena — see NewImpliedVars.
	pNode    *TreeNode
	timeNode *TreeNode

	// time is the animation clock every `time` reference evaluates to.
	time float64

	// pixelSizeInput builds pixelSize() as one input var, pixelSizeTree
	// (default 0), for a renderer that supplies each point's pixel footprint;
	// otherwise pixelSize() is the constant 0. See ToColorShapesWithPixelSize.
	pixelSizeInput   bool
	pixelSizeTree    fidget.TreeIndex
	hasPixelSizeTree bool

	// partProxyTree is the radius beyond which a part whose solid is in its
	// box reads as the box, one input var (default 0, never), built with
	// pixelSizeInput for a renderer that marches as the viewer does. See
	// partProxy.
	partProxyTree    fidget.TreeIndex
	hasPartProxyTree bool

	// dispDrawn is what displacementDrawn() builds as: 0, nothing displaced,
	// unless a renderer that marches the displaced field says otherwise. See
	// ToColorShapesWithPixelSize.
	dispDrawn float64

	// paramRequest names top-level declarations to build as input VARS
	// rather than baked constants; paramTrees records the var tree built
	// for each. See SetParamVars.
	paramRequest map[string]bool
	// paramRejected records a requested parameter whose declaration was seen
	// but is not a slider-style literal — distinct from one never seen at all,
	// which simply means the expression does not reference it.
	paramRejected map[string]bool
	paramTrees    map[string]fidget.TreeIndex

	// paramLiterals holds the literal each requested parameter is declared
	// with, which IsParameter reports so the fold keeps everything derived
	// from it dynamic. See markParamLiterals.
	paramLiterals map[*program.Operand]bool
}

// SetTime sets the animation clock this build evaluates `time` at. It must be
// called before the build; the default is 0, the moment a document opens at.
func (d *Delegate) SetTime(t float64) {
	d.time = t
}

// Ensure that Delegate implements the build.Delegate interface.
var _ build.Delegate = (*Delegate)(nil)

func NewDelegate(gslSource string) *Delegate {
	return &Delegate{
		gslSource: gslSource,
	}
}

func (d *Delegate) BuildVarRef(op *program.Operand, varInfo build.VarInfo) (build.Expression, error) {
	return varInfo.Expr(), nil
}

func (d *Delegate) BuildValue(op *program.Operand, value float64) (build.ConstExpression, error) {
	node, err := NewTreeNode(op, op.Start, fidget.NewConst(value))
	if err != nil {
		return nil, err
	}
	node.constValues = []float64{value}
	return node, nil
}

func (d *Delegate) GetScope() *build.Scope {
	return d.currBScope
}

func (d *Delegate) SetScope(bScope *build.Scope) *build.Scope {
	prev := d.currBScope
	d.currBScope = bScope

	s := d.ensureAgentScope(bScope)
	if prev != nil {
		if prevScope, ok := prev.GetAgentScope().(*Scope); ok {
			s.importingScope = prevScope
		}
	}

	return prev
}

func (d *Delegate) ensureAgentScope(bScope *build.Scope) *Scope {
	if existing, ok := bScope.GetAgentScope().(*Scope); ok && existing != nil {
		return existing
	}

	var parentScope *Scope
	if bScope.Parent() != nil {
		parentScope = d.ensureAgentScope(bScope.Parent())
	}

	s := NewScope(parentScope, bScope)
	bScope.SetAgentScope(s)
	return s
}
