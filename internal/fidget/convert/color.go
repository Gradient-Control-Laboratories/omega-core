package convert

import (
	"errors"
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/internal/fidget"
	"gradientcontrol.com/omega/core/program"
	"gradientcontrol.com/omega/core/program/noise/spinodoid"
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
type ColorShape struct {
	Tree        fidget.TreeIndex
	Name        string
	Color       [3]fidget.TreeIndex
	HasColor    bool
	Material    [materialRows][materialRowSize]fidget.TreeIndex
	HasMaterial bool
	Displaced   bool
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
	return toColorShapes(prog, timeAt, false, func(shapes []ColorShape, _ *Delegate) error {
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
	return toColorShapes(prog, timeAt, false, func(shapes []ColorShape, d *Delegate) error {
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
	return toColorShapes(prog, timeAt, true, func(shapes []ColorShape, d *Delegate) error {
		return f(shapes, d.pixelSizeTree, d.hasPixelSizeTree, d.partProxyTree, d.hasPartProxyTree)
	})
}

func toColorShapes(
	prog *program.Program,
	timeAt float64,
	pixelSizeInput bool,
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
	d.pixelSizeInput = pixelSizeInput
	if pixelSizeInput {
		d.dispDrawn = 1
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
		shape := ColorShape{Tree: trees[0]}
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
		}
		shapes = append(shapes, shape)
	}
	return f(shapes, d)
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
