package convert

import (
	"errors"
	"fmt"

	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/internal/noise/spinodoid"
	"gradientcontrol.com/omega/core/program"
)

// Build trees from the given program and call the given function with the trees.
// Callers should not hold onto the tree indices after the function returns.
func ToTrees(
	prog *program.Program,
	f func(treeIndices []fidget.TreeIndex, outputNames []string) error,
) error {
	return ToTreesAtTime(prog, 0, f)
}

// ToTreesAtTime is ToTrees with the animation clock set: every `time`
// reference in the program evaluates to timeAt. Meshing passes the time the
// request carries, so an export matches the frame the viewer is showing;
// ToTrees keeps the 0 a document opens at.
func ToTreesAtTime(
	prog *program.Program,
	timeAt float64,
	f func(treeIndices []fidget.TreeIndex, outputNames []string) error,
) error {
	return toTreesAtTime(prog, timeAt, func(t []fidget.TreeIndex, names []string, _ *Delegate) error {
		return f(t, names)
	})
}

// ToTapeTrees is ToTrees for a tape export, which also needs what the build
// bound to the oracle: a spinodoid leaf travels as a bare slot id that means
// nothing in another process, so the tape has to carry the field's arguments
// beside it (see SpinodoidOracle).
func ToTapeTrees(
	prog *program.Program,
	f func(treeIndices []fidget.TreeIndex, outputNames []string, spinodoids []SpinodoidOracle) error,
) error {
	return toTreesAtTime(prog, 0, func(t []fidget.TreeIndex, names []string, d *Delegate) error {
		return f(t, names, d.SpinodoidOracles())
	})
}

func toTreesAtTime(
	prog *program.Program,
	timeAt float64,
	f func(treeIndices []fidget.TreeIndex, outputNames []string, d *Delegate) error,
) error {
	if prog == nil || prog.Primary == nil {
		return fmt.Errorf("program or primary package is nil")
	}

	var gslSource string
	if len(prog.Primary.Files) > 0 {
		gslSource = prog.Primary.Files[0].Source
	}

	d := NewDelegate(gslSource)
	d.SetTime(timeAt)

	// Only one build can happen at a time
	// because current fidget usage has global state
	// Don't release the lock until all trees are built
	// and all operations on the trees (e.g. stl generation) are completed.
	fidget.Lock()
	defer fidget.Unlock()
	defer fidget.Clear()

	err := build.Build(d, prog, true)
	if err != nil {
		return build.ErrWithPkgSrc(prog, err)
	}

	nodes := d.outputs
	if len(nodes) == 0 {
		return errors.New("no outputs")
	}

	treeIndices := make([]fidget.TreeIndex, 0, len(nodes))
	for _, node := range nodes {
		typeName := "num"
		if ft := node.FidgetType(); ft != nil {
			typeName = ft.String()
		}
		if len(node.Trees()) != 1 {
			return build.ErrLoc(d.GetScope(), node.Start(),
				fmt.Errorf("expected numeric return values, got %s", typeName))
		}
		treeIndices = append(treeIndices, node.Trees()[0])
	}

	return f(treeIndices, d.outputNames, d)
}

// ReturnGroup describes one return value's shape: its GSL type name and how
// many scalar component trees it contributed to the flat tree list.
type ReturnGroup struct {
	TypeName string
	Count    int
}

// ToTreeGroups is ToTrees for callers that accept non-scalar returns: a
// vector or matrix return contributes all its component trees (in order)
// plus a ReturnGroup describing the grouping, instead of erroring. The
// meshing pipeline keeps using ToTrees; this serves ad-hoc evaluation
// (gsleval), where models legitimately ask for vector- and matrix-valued
// expressions.
func ToTreeGroups(
	prog *program.Program,
	f func([]fidget.TreeIndex, []ReturnGroup) error,
) error {
	if prog == nil || prog.Primary == nil {
		return fmt.Errorf("program or primary package is nil")
	}

	var gslSource string
	if len(prog.Primary.Files) > 0 {
		gslSource = prog.Primary.Files[0].Source
	}

	d := NewDelegate(gslSource)

	fidget.Lock()
	defer fidget.Unlock()
	defer fidget.Clear()

	err := build.Build(d, prog, true)
	if err != nil {
		return build.ErrWithPkgSrc(prog, err)
	}

	nodes := d.outputs
	if len(nodes) == 0 {
		return errors.New("no outputs")
	}

	var treeIndices []fidget.TreeIndex
	groups := make([]ReturnGroup, 0, len(nodes))
	for _, node := range nodes {
		typeName := "num"
		if ft := node.FidgetType(); ft != nil {
			typeName = ft.String()
		}
		trees := node.Trees()
		if len(trees) == 0 {
			return build.ErrLoc(d.GetScope(), node.Start(),
				fmt.Errorf("expected numeric return values, got %s", typeName))
		}
		treeIndices = append(treeIndices, trees...)
		groups = append(groups, ReturnGroup{TypeName: typeName, Count: len(trees)})
	}

	return f(treeIndices, groups)
}

func (d *Delegate) NewImpliedVars(bScope *build.Scope) ([]build.VarInfo, error) {
	d.rootBScope = bScope.Parent()

	pLoc := &program.Location{Line: -1, Column: -1, Offset: -1}
	pNode, err := NewTreeNode(nil, pLoc, fidget.NewX(), fidget.NewY(), fidget.NewZ())
	if err != nil {
		return nil, fmt.Errorf("failed to create root p node: %w", err)
	}
	pVarInfo := NewTreeVarInfo(bScope, program.PVarName, pNode)

	timeLoc := &program.Location{Line: -1, Column: -1, Offset: -1}
	timeNode, err := NewTreeNode(nil, timeLoc, fidget.NewConst(0))
	if err != nil {
		return nil, fmt.Errorf("failed to create root time node: %w", err)
	}
	timeVarInfo := NewTreeVarInfo(bScope, program.TimeVarName, timeNode)

	// build calls NewImpliedVars BEFORE BuildPackage, and BuildPackage clears
	// the tree arena — so these indices are stale by the time anything is
	// built against them. p survives only by luck (x, y, z are re-pushed at
	// the same indices); time's constant does not, and the index it kept
	// aliased another tree. Keep both nodes so BuildPackage can re-point them
	// at fresh trees once the arena is live.
	d.pNode = pNode
	d.timeNode = timeNode

	return []build.VarInfo{pVarInfo, timeVarInfo}, nil
}

// The first method called by the build package
func (d *Delegate) BuildPackage(module *program.Package) error {
	d.spinodoidDomain = spinodoid.SceneDomain(module.GetScene().GetSpinodoidBounds(), module.GetScene().GetSize())

	// Clear the global state
	fidget.Clear()
	d.hasPixelSizeTree = false
	d.hasPartProxyTree = false

	d.rootBScope = d.GetScope().Parent()

	// Initialize cached tree indices for common values
	d.two = fidget.NewConst(2)
	d.one = fidget.NewConst(1)
	d.half = fidget.NewConst(0.5)
	d.epsilon = fidget.NewConst(1e-12)
	d.zero = fidget.NewConst(0)
	d.negOne = fidget.NewConst(-1)

	// Re-point the implied vars at trees from the live arena (see
	// NewImpliedVars): the indices they captured were invalidated by the
	// Clear above.
	if d.pNode != nil {
		d.pNode.trees = []fidget.TreeIndex{fidget.NewX(), fidget.NewY(), fidget.NewZ()}
	}
	if d.timeNode != nil {
		if d.paramRequest[program.TimeVarName] {
			// time is being proven over a range: build the clock as a tape
			// var (at its point value) so interval evaluation can widen it —
			// a motion envelope. Requested before any literal declaration
			// could shadow it, so paramAssignment leaves a same-named
			// document constant alone.
			varTree := fidget.NewInput(program.TimeVarName, d.time)
			d.timeNode.trees = []fidget.TreeIndex{varTree}
			d.paramTrees[program.TimeVarName] = varTree
		} else {
			d.timeNode.trees = []fidget.TreeIndex{fidget.NewConst(d.time)}
		}
	}

	return nil
}

// The last method called by the build package
func (d *Delegate) BuildPackageReturns(ops []*program.Operand, expressions []build.Expression) error {
	var outputs []*TreeNode
	var names []string
	var returnIdx []int
	for i, expr := range expressions {
		var op *program.Operand
		if i < len(ops) {
			op = ops[i]
		}
		name := outputName(op)
		switch node := expr.(type) {
		case *TreeNode:
			outputs = append(outputs, node)
			names = append(names, name)
			returnIdx = append(returnIdx, i)
		case *MultiTreeNode:
			ns := node.Nodes()
			for j := range ns {
				outputs = append(outputs, ns[j])
				if len(ns) == 1 {
					names = append(names, name)
				} else {
					names = append(names, fmt.Sprintf("%s[%d]", name, j))
				}
				returnIdx = append(returnIdx, i)
			}
		default:
			return fmt.Errorf("invalid return expression type: %T", node)
		}
	}
	d.outputs = outputs
	d.outputNames = names
	d.outputReturnIdx = returnIdx
	return nil
}

// outputName mirrors the UI's per-shape naming: a return's variable or function
// name — the same string webai/webedit show in the shapes list. Returns "" for
// an unnamed expression return (the consumer falls back to a positional name);
// the source-text fallback the UI uses isn't available at this layer.
func outputName(op *program.Operand) string {
	if op == nil {
		return ""
	}
	if ref := op.GetReference(); ref != nil && ref.Name != "" {
		return ref.Name
	}
	if fc := op.GetFnCall(); fc != nil && fc.Reference != nil && fc.Reference.Name != "" {
		return fc.Reference.Name
	}
	return ""
}
