package codegen

import (
	"gradientcontrol.com/omega/core/assembler/ui"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/ir"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
	"gradientcontrol.com/omega/core/program/edit"
)

type Delegate struct {
	wgslSource string
	module     *ir.Module

	imports []*ImportInfo

	rootBScope *build.Scope
	currBScope *build.Scope
	prevBScope *build.Scope
	// primaryBScope is the primary (top-level, unnamed) build scope. Captured
	// in BuildPackage. Remap builds its wrapper helper with this as the current
	// build scope so a captured p-dependent field re-derives under the remapped
	// p — the same self-contained codegen a top-level remap already produces.
	primaryBScope *build.Scope

	rootPVar    *wgsl.VarInfo
	rootTimeVar *wgsl.VarInfo
	rootCtx     BuildContext

	prog             *program.Program
	sliderVars       map[string]bool
	sliderParamNames []string
	sliderParamVars  map[string]*wgsl.VarInfo
	uiParams         []ui.Param
	uiParamVars      map[edit.Span]*wgsl.VarInfo
	uiParamOperands  map[*program.Operand]edit.Span
	structTypeCache  map[string]wgsl.TypeInfo

	meshSdfSlots *MeshSdfSlots
	// cellularUsed is true once a cellularF1 / cellularFeature call has been
	// emitted, so build.go appends the scan helper the call refers to.
	cellularUsed bool

	// noiseUsed is true when the program (or an import) calls noiseHash3, so the
	// noise tile texture is threaded into the shader functions.
	noiseUsed bool
	// spinodoidUsed is true when the program (or an import) calls a
	// gaborSpinodoid builtin, so the table texture is threaded into the shader
	// functions; spinodoidEmitted once a call has been emitted, so build.go
	// declares gsl_spinodoid. spinodoidSlots are the table bands handed out, by
	// spinodoid.Call key (spinodoid.go).
	spinodoidUsed        bool
	spinodoidEmitted     bool
	spinodoidGradEmitted bool
	spinodoidSlots       []string
	spinodoidSlotByKey   map[string]int
	// warpUsed is true when the program (or an import) calls warpField, so the
	// displacement texture is threaded into the shader functions.
	warpUsed bool
	// vanishUsed is true when a part is guarded by shrinkToVanish, so the camera
	// it reads and gsl_vanish are declared after sdShape (vanish.go).
	vanishUsed bool
	// simpleUsed is true when a part switches to a stand-in, so gsl_simple and
	// its rung are declared after sdShape (simple.go). deferSimple is set while
	// a part's guard generates its shape, and pendingSimple then holds the
	// shape's switch for the guard to write.
	simpleUsed    bool
	deferSimple   bool
	pendingSimple *simpleSwitch
	// simpleValueUsed is true when a colour or material has a stand-in, so its
	// weight is declared after sdShape (simple.go). simpleValues is each
	// function's switches written so far, so a copy of one reads its value.
	simpleValueUsed bool
	simpleValues    map[*Scope]map[string]string
	// dispDrawnUsed is true when the program reads displacementDrawn(), so its
	// var is declared after sdShape (function_call.go).
	dispDrawnUsed bool
	// detailUsed is true when the program draws fine detail (name.detail), so
	// the pixel its guards read is declared (detailPwDeclWGSL).
	detailUsed bool
	// foldConsts names each constant array a loop's elements are read from,
	// by its value, and foldConstDecls declares them, after sdShape
	// (fold_loop.go).
	foldConsts     map[string]string
	foldConstDecls []string
	// foldScopes is the scope of each fold whose step function is being
	// built, innermost last, and foldLoops writes each fold's loop where its
	// value is first read (fold_loop.go).
	foldScopes []*build.Scope
	foldLoops  map[*wgsl.VarInfo]func(ctx wgsl.CodegenContext)
	// parts is the box of each guarded part whose box is constant, in the
	// order they are guarded (vanish.go).
	parts [][2][3]float64
	// assignedNames records every name handed out by AssignUnique so successive
	// calls with the same prefix get distinct suffixes (otherwise two calls to
	// AssignUnique("x", …) would both emit `let x = …` and clobber each other).
	assignedNames map[string]bool

	// remapHelpers maps canonical helper signatures and bodies to the first synthesized function name.
	remapHelpers map[string]string

	// colorFieldSrc is the generated gsl_colorField function (empty when the
	// program authors no color). Built in BuildPackageColorReturns and woven into
	// the module by BuildPackageReturns, before the struct declarations.
	colorFieldSrc string
	// materialFieldSrc is the generated gsl_materialField function (empty when
	// the program authors no material), woven in after colorFieldSrc.
	materialFieldSrc string
	// heightsSrc is the generated gsl_heights function, every return's
	// displacement height at once (empty when no return is displaced), and
	// displacedMask marks the returns it carries a height for.
	heightsSrc    string
	displacedMask []bool
	// heightSrcs is each return's height as gsl_heights writes it, empty for
	// a return with none, so heightsDiffer can compare them.
	heightSrcs []string
	// heightEvident is set when a return's material lowered to a constant
	// record with a non-zero height.
	heightEvident bool
	// displacedFieldSrc is the generated gsl_sdShapeDisplaced, the composite of
	// each return's own displaced distance; woven in after heightsSrc.
	displacedFieldSrc string
}

// Ensure that Delegate implements the build.Delegate interface.
var _ build.Delegate = (*Delegate)(nil)

type Scope struct {
	pkgPath            string
	name               string
	parent             *Scope
	importingScope     *Scope
	delegate           *Delegate
	bScope             *build.Scope
	block              *ir.Block
	funcs              []*ir.FuncDecl
	prefixImpCache     map[string]*Scope
	globalRefs         []*wgsl.VarInfo
	generated          map[string]bool
	localAssignedNames map[string]bool
	isFunctionScope    bool
}

func (s *Scope) IsAgentScope() {}

func (s *Scope) PkgPath() string {
	if s.pkgPath != "" {
		return s.pkgPath
	}
	if s.parent != nil {
		return s.parent.PkgPath()
	}
	return ""
}

func newScope(pkgPath, name string, parent *Scope, delegate *Delegate, bScope *build.Scope) *Scope {
	return &Scope{
		pkgPath:            pkgPath,
		name:               name,
		parent:             parent,
		delegate:           delegate,
		bScope:             bScope,
		block:              ir.NewBlock(),
		generated:          make(map[string]bool),
		localAssignedNames: make(map[string]bool),
	}
}

func (s *Scope) Block() *ir.Block {
	return s.block
}

func (s *Scope) AddStmt(stmt ir.Stmt) {
	s.block.Add(stmt)
}

func (s *Scope) copy(name string) *Scope {
	return &Scope{
		pkgPath:            s.pkgPath,
		name:               name,
		parent:             s.parent,
		delegate:           s.delegate,
		bScope:             s.bScope,
		block:              ir.NewBlock(),
		prefixImpCache:     s.prefixImpCache,
		generated:          make(map[string]bool),
		localAssignedNames: make(map[string]bool),
	}
}

func (s *Scope) functionScope() *Scope {
	curr := s
	for curr.parent != nil && !curr.isFunctionScope {
		curr = curr.parent
	}
	return curr
}

func (d *Delegate) getAssignedNamesMap() map[string]bool {
	if d.assignedNames == nil {
		d.assignedNames = make(map[string]bool)
	}
	return d.assignedNames
}

func (d *Delegate) getLocalAssignedNamesMap() map[string]bool {
	fnScope := d.currentScope().functionScope()
	if fnScope.localAssignedNames == nil {
		fnScope.localAssignedNames = make(map[string]bool)
	}
	return fnScope.localAssignedNames
}

// markGenerated marks a variable as generated
// and returns true if it was not already generated.
func (d *Delegate) markGenerated(name string) bool {
	return d.currentScope().markGenerated(name)
}

// markGenerated marks a variable as generated
// and returns true if it was not already generated.
func (s *Scope) markGenerated(name string) bool {
	if s.generated[name] {
		return false
	}
	s.generated[name] = true
	return true
}

func (d *Delegate) currentScope() *Scope {
	return d.currBScope.GetAgentScope().(*Scope)
}

func (d *Delegate) rootScope() *Scope {
	return d.rootBScope.GetAgentScope().(*Scope)
}

func (d *Delegate) GetScope() *build.Scope {
	return d.currBScope
}

func (d *Delegate) SetScope(bScope *build.Scope) *build.Scope {
	prev := d.currBScope
	d.currBScope = bScope
	if prev != bScope && prev != nil {
		d.prevBScope = prev
	}

	s := d.ensureAgentScope(bScope)
	if prev != nil {
		if prevScope, ok := prev.GetAgentScope().(*Scope); ok {
			if s.importingScope == nil && s != prevScope {
				curr := prevScope
				hasCycle := false
				for curr != nil {
					if curr == s {
						hasCycle = true
						break
					}
					curr = curr.importingScope
				}
				if !hasCycle {
					s.importingScope = prevScope
				}
			}
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

	s := newScope("", bScope.Name(), parentScope, d, bScope)
	bScope.SetAgentScope(s)
	return s
}

func (d *Delegate) isLocalScope(bScope *build.Scope) bool {
	curr := bScope
	for curr.Parent() != nil && curr.Parent() != d.rootBScope {
		curr = curr.Parent()
	}
	return curr == d.primaryBScope
}

func (d *Delegate) emitVar(ref *wgsl.VarInfo, ctx wgsl.CodegenContext) {
	if ref == nil || !ref.CanRederive() {
		return
	}
	if loop, ok := d.foldLoops[ref]; ok {
		loop(ctx)
		return
	}
	expr := ref.Expr()
	if expr.Type().IsFunction() {
		return
	}
	wgslExpr := expr.(wgsl.Expression)
	val := wgslExpr.Codegen(ctx)
	reads := wgslExpr.Reads()
	if len(reads) == 0 && val != "" {
		reads = []string{val}
	}
	d.currentScope().AddStmt(ir.NewLetStmt(ref.WGSLName(), "", val, reads))
}

// withScopedAgent runs fn inside tempScope attached to primaryBScope,
// restoring the previous build scope and agent scope upon return.
func (d *Delegate) withScopedAgent(tempScope *Scope, fn func()) {
	prevBScope := d.currBScope
	prevPrimaryAgent := d.primaryBScope.GetAgentScope()
	d.primaryBScope.SetAgentScope(tempScope)
	d.currBScope = d.primaryBScope
	defer func() {
		d.currBScope = prevBScope
		d.primaryBScope.SetAgentScope(prevPrimaryAgent)
	}()
	fn()
}
