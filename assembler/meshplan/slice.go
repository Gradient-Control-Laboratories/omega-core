package meshplan

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

// SliceProgram translates a GSL whole-program AST into a MeshExportPlan
// (a tree of boolean, step/mesh asset, and analytical SDF nodes) and a slice of
// synthetic analytical SDF subtrees requiring backend meshing. Without the
// document's source, a return that names nothing is body-N; see SliceDocument.
func SliceProgram(
	prog *program.Program,
	idPrefix string,
) (*MeshExportPlan, []SdfSubtree, error) {
	return SliceDocument(prog, idPrefix, "")
}

// SliceDocument is SliceProgram given the document's source, so the plan names
// every return as the Bodies panel does (program.OperandName), its text when
// it names nothing.
func SliceDocument(
	prog *program.Program,
	idPrefix string,
	source string,
) (*MeshExportPlan, []SdfSubtree, error) {
	if prog == nil || prog.Primary == nil {
		return &MeshExportPlan{Root: nil}, nil, nil
	}

	module := prog.Primary
	var returns []*program.Operand
	var assignments []*program.Assignment
	for _, f := range module.Files {
		returns = append(returns, f.Returns...)
		assignments = append(assignments, f.Assignments...)
	}

	ctx := &sliceCtx{
		module:      module,
		idPrefix:    idPrefix,
		nextLeaf:    0,
		varBindings: buildVarBindings(assignments),
		getImport:   prog.GetImport,
	}

	roots := make([]*MeshPlanNode, 0, len(returns))
	bodyNames := make(map[int]string)
	used := make(map[string]bool)
	for retIdx, ret := range returns {
		node, err := ctx.sliceOperand(ret)
		if err != nil {
			return nil, nil, err
		}
		assignShapeIndex(node, retIdx+1)
		roots = append(roots, node)
		if base := program.OperandName(ret, source); base != "" {
			name := base
			for n := 2; used[name]; n++ {
				name = fmt.Sprintf("%s-%d", base, n)
			}
			used[name] = true
			bodyNames[retIdx+1] = name
		}
	}

	var root *MeshPlanNode
	if len(roots) > 0 {
		root = roots[0]
		for _, r := range roots[1:] {
			root = UnionPlanNodes(root, r)
		}
	}

	if err := ctx.materializeSubtrees(); err != nil {
		return nil, nil, err
	}

	if len(bodyNames) == 0 {
		bodyNames = nil
	}
	return &MeshExportPlan{Root: root, BodyNames: bodyNames}, ctx.subtrees, nil
}

type sliceCtx struct {
	module        *program.Package
	idPrefix      string
	nextLeaf      int
	varBindings   map[string]*program.Operand
	pendingLeaves []*pendingSdfLeaf
	subtrees      []SdfSubtree
	getImport     program.GetImportFunc
}

type pendingSdfLeaf struct {
	node *MeshPlanNode
	expr *program.Operand
}

func (c *sliceCtx) allocSubtreeID() string {
	id := fmt.Sprintf("%s%d", c.idPrefix, c.nextLeaf)
	c.nextLeaf++
	return id
}

func (c *sliceCtx) sliceOperand(op *program.Operand) (*MeshPlanNode, error) {
	if op == nil {
		return nil, fmt.Errorf("slicer: nil operand")
	}

	target := op
	if ref := op.GetReference(); ref != nil && ref.Name != "" {
		if resolved := c.resolveRef(op); isMeshCall(resolved) {
			target = resolved
		}
	}

	if fc := target.GetFnCall(); fc != nil {
		if inlined := inlineFnCall(fc, c.module); inlined != nil {
			return c.sliceOperand(inlined)
		}
		if fc.Reference != nil && fc.Reference.Prefix == "" {
			switch fc.Reference.Name {
			case "meshUnion", "meshIntersect", "meshDifference":
				return c.sliceMeshCSG(target, fc)
			case "mesh":
				return c.sliceMeshLiteral(target, fc)
			}
		}
	}

	return c.allocSdfLeaf(op), nil
}

func (c *sliceCtx) resolveRef(op *program.Operand) *program.Operand {
	seen := make(map[string]bool)
	for {
		ref := op.GetReference()
		if ref == nil || ref.Name == "" {
			return op
		}
		name := ref.Name
		if seen[name] {
			return op
		}
		seen[name] = true
		bound := c.varBindings[name]
		if bound == nil {
			return op
		}
		op = bound
	}
}

func isMeshCall(op *program.Operand) bool {
	if op == nil {
		return false
	}
	fc := op.GetFnCall()
	if fc == nil || fc.Reference == nil || fc.Reference.Prefix != "" {
		return false
	}
	switch fc.Reference.Name {
	case "meshUnion", "meshIntersect", "meshDifference", "mesh":
		return true
	}
	return false
}

func buildVarBindings(assignments []*program.Assignment) map[string]*program.Operand {
	bindings := make(map[string]*program.Operand)
	for _, a := range assignments {
		if len(a.Variables) == 1 && a.Variables[0] != nil {
			bindings[a.Variables[0].Name] = a.Expression
		}
	}
	return bindings
}

func (c *sliceCtx) sliceMeshCSG(op *program.Operand, fc *program.FnCall) (*MeshPlanNode, error) {
	builtin := program.BuiltinNamed(fc.Reference.Name)
	if builtin == nil {
		return nil, fmt.Errorf("slicer: unknown mesh-CSG builtin %q", fc.Reference.Name)
	}
	if builtin.Reduce == program.ReduceTail {
		return c.sliceMeshDifference(fc)
	}

	var elements []*program.Operand
	if len(fc.Arguments) == 1 {
		elements = c.expandMeshList(fc.Arguments[0])
	} else {
		elements = fc.Arguments
	}
	return c.foldMeshCSG(fc.Reference.Name, elements)
}

func (c *sliceCtx) sliceMeshDifference(fc *program.FnCall) (*MeshPlanNode, error) {
	if len(fc.Arguments) != 2 {
		return nil, fmt.Errorf("slicer: meshDifference expects 2 arguments, got %d", len(fc.Arguments))
	}
	base, err := c.foldMeshCSG("meshUnion", c.expandMeshList(fc.Arguments[0]))
	if err != nil {
		return nil, err
	}
	acc := base
	for _, sub := range c.expandMeshList(fc.Arguments[1]) {
		right, err := c.sliceOperand(sub)
		if err != nil {
			return nil, err
		}
		acc, err = c.combineMeshCSG("meshDifference", acc, right)
		if err != nil {
			return nil, err
		}
	}
	return acc, nil
}

func (c *sliceCtx) foldMeshCSG(builtinName string, elements []*program.Operand) (*MeshPlanNode, error) {
	if len(elements) == 0 {
		return nil, fmt.Errorf("slicer: %s expects at least one mesh, got an empty list", builtinName)
	}
	acc, err := c.sliceOperand(elements[0])
	if err != nil {
		return nil, err
	}
	for _, el := range elements[1:] {
		right, err := c.sliceOperand(el)
		if err != nil {
			return nil, err
		}
		acc, err = c.combineMeshCSG(builtinName, acc, right)
		if err != nil {
			return nil, err
		}
	}
	return acc, nil
}

func (c *sliceCtx) combineMeshCSG(builtinName string, left, right *MeshPlanNode) (*MeshPlanNode, error) {
	if isSdfLeaf(left) && isSdfLeaf(right) {
		return c.mergeSdfLeaves(builtinName, left, right), nil
	}
	planOp, err := planOpForBuiltin(builtinName)
	if err != nil {
		return nil, err
	}
	return &MeshPlanNode{
		Kind:  MeshPlanKindBool,
		Op:    planOp,
		Left:  left,
		Right: right,
	}, nil
}

func (c *sliceCtx) expandMeshList(op *program.Operand) []*program.Operand {
	if op == nil {
		return nil
	}
	if ex, ok := op.T.(*program.Operand_Expression); ok && ex.Expression != nil {
		if ex.Expression.Operation == program.Operation_COLLECT {
			var out []*program.Operand
			for _, sub := range ex.Expression.Operands {
				out = append(out, c.expandMeshList(sub)...)
			}
			return out
		}
	}
	if ref := op.GetReference(); ref != nil && ref.Name != "" {
		if bound := c.resolveRef(op); bound != op {
			return c.expandMeshList(bound)
		}
	}
	if fc := op.GetFnCall(); fc != nil {
		if inlined := inlineFnCall(fc, c.module); inlined != nil {
			return c.expandMeshList(inlined)
		}
	}
	return []*program.Operand{op}
}

func (c *sliceCtx) sliceMeshLiteral(op *program.Operand, fc *program.FnCall) (*MeshPlanNode, error) {
	decl, err := parseMeshArgs(fc, c.module, c.getImport)
	if err != nil {
		return nil, fmt.Errorf("slicer: mesh() arg eval failed: %w", err)
	}
	return &MeshPlanNode{
		Kind:       MeshPlanKindAsset,
		AssetIndex: decl.Index,
		Transform:  MeshTransformFromDecl(decl.Matrix),
	}, nil
}

func (c *sliceCtx) allocSdfLeaf(expr *program.Operand) *MeshPlanNode {
	id := c.allocSubtreeID()
	node := &MeshPlanNode{
		Kind:      MeshPlanKindSdf,
		SubtreeID: id,
	}
	c.pendingLeaves = append(c.pendingLeaves, &pendingSdfLeaf{
		node: node,
		expr: expr,
	})
	return node
}

func (c *sliceCtx) mergeSdfLeaves(builtinName string, left, right *MeshPlanNode) *MeshPlanNode {
	leftPending := c.findPending(left)
	rightPending := c.findPending(right)
	if leftPending == nil || rightPending == nil {
		op, _ := planOpForBuiltin(builtinName)
		return &MeshPlanNode{
			Kind:  MeshPlanKindBool,
			Op:    op,
			Left:  left,
			Right: right,
		}
	}

	var combined *program.Operand
	switch builtinName {
	case "meshUnion":
		combined = program.NewExpression(program.Operation_MIN, []*program.Operand{leftPending.expr, rightPending.expr}, nil, nil)
	case "meshIntersect":
		combined = program.NewExpression(program.Operation_MAX, []*program.Operand{leftPending.expr, rightPending.expr}, nil, nil)
	case "meshDifference":
		negR := program.NewExpression(program.Operation_NEG, []*program.Operand{rightPending.expr}, nil, nil)
		combined = program.NewExpression(program.Operation_MAX, []*program.Operand{leftPending.expr, negR}, nil, nil)
	default:
		combined = leftPending.expr
	}
	leftPending.expr = combined
	c.removePending(rightPending)
	return left
}

func (c *sliceCtx) findPending(node *MeshPlanNode) *pendingSdfLeaf {
	for _, p := range c.pendingLeaves {
		if p.node == node {
			return p
		}
	}
	return nil
}

func (c *sliceCtx) removePending(target *pendingSdfLeaf) {
	out := c.pendingLeaves[:0]
	for _, p := range c.pendingLeaves {
		if p != target {
			out = append(out, p)
		}
	}
	c.pendingLeaves = out
}

func (c *sliceCtx) materializeSubtrees() error {
	c.subtrees = make([]SdfSubtree, 0, len(c.pendingLeaves))
	for _, p := range c.pendingLeaves {
		c.subtrees = append(c.subtrees, SdfSubtree{
			ID:     p.node.SubtreeID,
			Return: p.expr,
		})
	}
	return nil
}

// IndexSdfLeaves indexes every SdfLeaf in the plan node hierarchy by SubtreeID.
func IndexSdfLeaves(root *MeshPlanNode, leavesByID map[string]*MeshPlanNode) {
	if root == nil || leavesByID == nil {
		return
	}
	switch root.Kind {
	case MeshPlanKindSdf:
		leavesByID[root.SubtreeID] = root
	case MeshPlanKindBool:
		IndexSdfLeaves(root.Left, leavesByID)
		IndexSdfLeaves(root.Right, leavesByID)
	}
}

// MeshTransformFromDecl constructs a MeshTransform from a 16-element column-major matrix.
func MeshTransformFromDecl(matrix []float64) *MeshTransform {
	if len(matrix) == 16 {
		return &MeshTransform{Matrix: append([]float64(nil), matrix...)}
	}
	return &MeshTransform{
		Matrix: []float64{
			1, 0, 0, 0,
			0, 1, 0, 0,
			0, 0, 1, 0,
			0, 0, 0, 1,
		},
	}
}

// ApplyMat4Col applies a 4x4 column-major transform to a 3D point.
func ApplyMat4Col(m []float64, v [3]float64) [3]float64 {
	if len(m) < 16 {
		return v
	}
	return [3]float64{
		m[0]*v[0] + m[4]*v[1] + m[8]*v[2] + m[12],
		m[1]*v[0] + m[5]*v[1] + m[9]*v[2] + m[13],
		m[2]*v[0] + m[6]*v[1] + m[10]*v[2] + m[14],
	}
}

func isSdfLeaf(node *MeshPlanNode) bool {
	return node != nil && node.Kind == MeshPlanKindSdf
}

func planOpForBuiltin(name string) (string, error) {
	switch name {
	case "meshUnion":
		return MeshPlanOpUnion, nil
	case "meshIntersect":
		return MeshPlanOpIntersect, nil
	case "meshDifference":
		return MeshPlanOpSubtract, nil
	}
	return "", fmt.Errorf("slicer: unknown mesh-CSG builtin %q", name)
}

func assignShapeIndex(node *MeshPlanNode, shapeIndex int) {
	if node == nil {
		return
	}
	if node.ShapeIndex == 0 {
		node.ShapeIndex = shapeIndex
	}
	assignShapeIndex(node.Left, shapeIndex)
	assignShapeIndex(node.Right, shapeIndex)
}
