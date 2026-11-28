package convert

import (
	"errors"
	"fmt"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/build"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/fidget"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
	"github.com/Gradient-Control-Laboratories/omega-core/program/noise/spinodoid"
)

// colorDelegate is a Delegate that also collects per-return color fields. The
// hook lives on this wrapper rather than on Delegate itself so that only the
// paths that render color pay for it: build.Build skips the color phase
// entirely for a delegate that does not implement the interface, which keeps
// the meshing and evaluation paths byte-identical to what they were before
// color existed here.
type colorDelegate struct {
	*Delegate
}

// Ensure the wrapper satisfies the optional color hook.
var _ build.ColorReturnsBuilder = (*colorDelegate)(nil)

// BuildPackageColorReturns implements [build.ColorReturnsBuilder]: it records
// each return's lowered color field as three trees (r, g, b), parallel to the
// returns themselves, so a renderer can evaluate a shape's color at a hit
// point the same way it evaluates its distance. Returns that author no color
// (name.color = …) record nil and fall back to the caller's palette, matching
// the viewer's alpha-0 "no field" convention.
//
// Unlike the WGSL delegate — which re-derives the color expressions as source
// text in a fresh scope — the trees the build package hands over here are
// already the finished article, because a tree IS the lowered form. Storing
// them is the whole implementation.
func (cd *colorDelegate) BuildPackageColorReturns(colorExpressions []build.Expression) error {
	d := cd.Delegate
	colors := make([][3]fidget.TreeIndex, len(colorExpressions))
	has := make([]bool, len(colorExpressions))
	for i, ce := range colorExpressions {
		if ce == nil {
			continue
		}
		trees := colorTrees(ce)
		if len(trees) != 3 {
			// A color field that did not lower to a vec3 is not renderable;
			// skip it rather than failing the build, so an odd color
			// expression costs that shape its color and nothing more.
			continue
		}
		colors[i] = [3]fidget.TreeIndex{trees[0], trees[1], trees[2]}
		has[i] = true
	}
	d.colorReturns = colors
	d.hasColorReturn = has
	return nil
}

// ColorShape is one renderable output: its distance tree, its display name,
// the (r, g, b) trees of its color field when the document authored one, and
// the rows of its material field when it authored that. A shape with HasColor
// false takes the caller's palette color, matching the viewer, where an absent
// color field reads as alpha 0; one with HasMaterial false takes the default
// record, the viewer's present = 0. The two are independent. Displaced marks a
// material whose height (Material[2][2]) is not known to be the constant zero,
// so a renderer can leave the distance tree alone for every other shape.
//
// InTree, when HasInTree, is the shape's distance displaced inside its
// booleans: each shape in it that sets a material with a height moved by its
// own layer before the booleans that use it (program.DisplacedReturns). A
// renderer marches it in place of Tree minus the material's height. Built only
// by ToColorShapesWithPixelSize and ToDisplacedTreesAtTime.
type ColorShape struct {
	Tree        fidget.TreeIndex
	Name        string
	Color       [3]fidget.TreeIndex
	HasColor    bool
	Material    [materialRows][materialRowSize]fidget.TreeIndex
	HasMaterial bool
	Displaced   bool
	InTree      fidget.TreeIndex
	HasInTree   bool
}

// ToColorShapes is ToTrees for renderers that also want color: it builds the
// program and calls f with one ColorShape per scalar output. Like ToTrees, the
// indices are only valid inside f — the fidget lock and tree arena are
// released on return.
func ToColorShapes(
	prog *program.Program,
	timeAt float64,
	f func(shapes []ColorShape) error,
) error {
	return toColorShapes(prog, timeAt, shapeConfig{}, func(shapes []ColorShape, _ *Delegate) error {
		return f(shapes)
	})
}

// ToColorShapesWithOracles is ToColorShapes for a tape export, which also
// needs the spinodoid tables the build bound so the tape can describe them —
// see ToTapeTrees.
func ToColorShapesWithOracles(
	prog *program.Program,
	timeAt float64,
	f func(shapes []ColorShape, spinodoids []spinodoid.Oracle) error,
) error {
	return toColorShapes(prog, timeAt, shapeConfig{}, func(shapes []ColorShape, d *Delegate) error {
		return f(shapes, d.SpinodoidOracles())
	})
}

// ToColorShapesWithPixelSize is ToColorShapes for a renderer that knows each
// point's pixel footprint: pixelSize() is built as an input var (default 0)
// rather than the constant 0, and f receives its tree to bind per point
// (fidget.EvalGradVars). hasPixelSize is false when the program never calls
// pixelSize(). Such a renderer also marches each shape's displaced field at
// full scale (mesh.litDisplaced), so displacementDrawn() is 1. Its parts read
// as their boxes beyond the radius it binds to partProxy (see
// Delegate.partProxy); hasPartProxy is false when there is no such part.
func ToColorShapesWithPixelSize(
	prog *program.Program,
	timeAt float64,
	f func(shapes []ColorShape, pixelSize fidget.TreeIndex, hasPixelSize bool, partProxy fidget.TreeIndex, hasPartProxy bool) error,
) error {
	lit := shapeConfig{pixelSizeInput: true, displaced: true, build: 1, relief: 1}
	return toColorShapes(prog, timeAt, lit, func(shapes []ColorShape, d *Delegate) error {
		return f(shapes, d.pixelSizeTree, d.hasPixelSizeTree, d.partProxyTree, d.hasPartProxyTree)
	})
}

// ToDisplacedTreesAtTime is ToTreesAtTime for a displaced mesh: each return's
// distance moved by build times its materials' builds and relief times their
// relief, each 0 or 1. A return whose layers are set inside it
// (program.DisplacedReturns) is its layered distance, built with buildScale()
// and displacementScale() at those values; any other return whose material
// has a height is moved as a whole (DisplacedTree). displacementDrawn() is
// relief: where relief is geometry, the colours and materials that stand in
// for it when it is not switch off, as in the lit renderer.
func ToDisplacedTreesAtTime(
	prog *program.Program,
	timeAt float64,
	build, relief float64,
	f func(treeIndices []fidget.TreeIndex, outputNames []string) error,
) error {
	cfg := shapeConfig{displaced: true, build: build, relief: relief}
	return toColorShapes(prog, timeAt, cfg, func(shapes []ColorShape, _ *Delegate) error {
		trees := make([]fidget.TreeIndex, len(shapes))
		names := make([]string, len(shapes))
		for i, s := range shapes {
			trees[i] = DisplacedTree(s, build, relief)
			names[i] = s.Name
		}
		return f(trees, names)
	})
}

// DisplacedTree is the distance a shape is drawn or meshed at when its
// materials move it by build times their build and relief times their relief:
// its layered distance when it has one (its scales were built into it), else
// Tree - build * b - relief * (h - b), h and b its material's height and
// build, for a material whose height is not known to be zero; else Tree. With
// the two scales equal it is Tree - s * h, exactly as before builds were
// scaled apart.
func DisplacedTree(s ColorShape, build, relief float64) fidget.TreeIndex {
	if s.HasInTree {
		return s.InTree
	}
	if !s.HasMaterial || !s.Displaced || (build == 0 && relief == 0) {
		return s.Tree
	}
	h := s.Material[displacementRow][displacementComp]
	if build == relief {
		if build == 1 {
			return fidget.Sub(s.Tree, h)
		}
		return fidget.Sub(s.Tree, fidget.Mul(fidget.NewConst(build), h))
	}
	b := s.Material[displacementRow][buildComp]
	moved := fidget.Add(fidget.Mul(fidget.NewConst(build), b),
		fidget.Mul(fidget.NewConst(relief), fidget.Sub(h, b)))
	return fidget.Sub(s.Tree, moved)
}

// shapeConfig is how toColorShapes builds a program: pixelSize() as an input
// var for a renderer that knows each point's footprint; displaced to keep the
// layered distances, with buildScale() and displacementScale() at build and
// relief, and displacementDrawn() at relief.
type shapeConfig struct {
	pixelSizeInput bool
	displaced      bool
	build, relief  float64
}

func toColorShapes(
	prog *program.Program,
	timeAt float64,
	cfg shapeConfig,
	f func(shapes []ColorShape, d *Delegate) error,
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
	d.pixelSizeInput = cfg.pixelSizeInput
	if cfg.displaced {
		d.recordInTree = true
		d.buildScale, d.reliefScale, d.dispDrawn = cfg.build, cfg.relief, cfg.relief
	}

	fidget.Lock()
	defer fidget.Unlock()
	defer fidget.Clear()

	if err := build.Build(&colorDelegate{d}, BuildCopy(prog), true); err != nil {
		return build.ErrWithPkgSrc(prog, err)
	}
	if len(d.outputs) == 0 {
		return errors.New("no outputs")
	}

	shapes := make([]ColorShape, 0, len(d.outputs))
	for i, node := range d.outputs {
		trees := node.Trees()
		if len(trees) != 1 {
			return build.ErrLoc(d.GetScope(), node.Start(),
				fmt.Errorf("expected numeric return values, got %s", node.Type().String()))
		}
		shapes = append(shapes, d.colorShape(i, trees[0]))
	}
	return f(shapes, d)
}

// colorShape is output i, whose distance is tree, with what the build recorded
// for its return: name, color, material and layered distance.
func (d *Delegate) colorShape(i int, tree fidget.TreeIndex) ColorShape {
	shape := ColorShape{Tree: tree}
	if i < len(d.outputNames) {
		shape.Name = d.outputNames[i]
	}
	if i < len(d.outputReturnIdx) {
		r := d.outputReturnIdx[i]
		if r < len(d.hasColorReturn) && d.hasColorReturn[r] {
			shape.Color = d.colorReturns[r]
			shape.HasColor = true
		}
		if r < len(d.hasMaterialReturn) && d.hasMaterialReturn[r] {
			shape.Material = d.materialReturns[r]
			shape.HasMaterial = true
			shape.Displaced = d.displacedReturn[r]
		}
		if r < len(d.hasInTreeReturn) && d.hasInTreeReturn[r] {
			shape.InTree = d.inTreeReturns[r]
			shape.HasInTree = true
		}
	}
	return shape
}

// colorTrees flattens a lowered color expression to its component trees. A
// vec3 arrives either as one TreeNode carrying three trees or as a
// MultiTreeNode of three single-tree nodes, depending on how it was built.
func colorTrees(expr build.Expression) []fidget.TreeIndex {
	switch node := expr.(type) {
	case *TreeNode:
		return node.Trees()
	case *MultiTreeNode:
		var trees []fidget.TreeIndex
		for _, n := range node.Nodes() {
			trees = append(trees, n.Trees()...)
		}
		return trees
	}
	return nil
}
