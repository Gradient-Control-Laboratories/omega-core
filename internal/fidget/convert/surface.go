package convert

import (
	"errors"
	"fmt"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/build"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/fidget"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

// Surface is where a query's field puts each layered surface: moved by Build
// times its materials' builds and Relief times their relief, each from 0 to 1.
// The zero Surface is the field as built, the one ToTrees builds.
type Surface struct {
	Build, Relief float64
}

// Moved reports whether s moves any surface.
func (s Surface) Moved() bool { return s.Build != 0 || s.Relief != 0 }

// ToSurfaceTrees builds prog for a query that reads its field at s: each
// scalar return is its distance as a displaced mesh meshes it
// (ToDisplacedTreesAtTime: its layered distance, built with buildScale() and
// displacementScale() at s, or moved as a whole by DisplacedTree), and a vector
// or matrix return gives its component trees as ToTreeGroups gives them. With
// params it builds the named declarations as input vars, as ToTreesWithParams
// does, and f receives each one's var; without them paramTrees is nil.
// displacementDrawn() is s.Relief, as for a displaced mesh. The indices are
// only valid inside f.
func ToSurfaceTrees(
	prog *program.Program,
	s Surface,
	params []string,
	f func(treeIndices []fidget.TreeIndex, groups []ReturnGroup, paramTrees map[string]fidget.TreeIndex) error,
) error {
	if prog == nil || prog.Primary == nil {
		return fmt.Errorf("program or primary package is nil")
	}
	var gslSource string
	if len(prog.Primary.Files) > 0 {
		gslSource = prog.Primary.Files[0].Source
	}
	// Mark and build the same copy: markParamLiterals keys operands by pointer.
	prog = BuildCopy(prog)
	d := NewDelegate(gslSource)
	d.recordInTree = true
	d.buildScale, d.reliefScale, d.dispDrawn = s.Build, s.Relief, s.Relief
	d.SetParamVars(params)
	d.markParamLiterals(prog)

	fidget.Lock()
	defer fidget.Unlock()
	defer fidget.Clear()

	if err := build.Build(&colorDelegate{d}, prog, true); err != nil {
		return build.ErrWithPkgSrc(prog, err)
	}
	if len(d.outputs) == 0 {
		return errors.New("no outputs")
	}
	if err := d.bindParams(params); err != nil {
		return err
	}

	var treeIndices []fidget.TreeIndex
	var groups []ReturnGroup
	for i, node := range d.outputs {
		typeName := "num"
		if ft := node.FidgetType(); ft != nil {
			typeName = ft.String()
		}
		trees := node.Trees()
		if len(trees) == 0 {
			return build.ErrLoc(d.GetScope(), node.Start(),
				fmt.Errorf("expected numeric return values, got %s", typeName))
		}
		if len(trees) == 1 {
			trees = []fidget.TreeIndex{DisplacedTree(d.colorShape(i, trees[0]), s.Build, s.Relief)}
		}
		treeIndices = append(treeIndices, trees...)
		groups = append(groups, ReturnGroup{TypeName: typeName, Count: len(trees)})
	}
	if params == nil {
		return f(treeIndices, groups, nil)
	}
	return f(treeIndices, groups, d.paramTrees)
}
