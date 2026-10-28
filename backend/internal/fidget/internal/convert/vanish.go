package convert

import (
	"fmt"

	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// shrinkToVanishName is builtin's marker for a part (3d.gsl),
// shrinkToVanish(shape, lo, hi, solid), whose body is the shape.
const shrinkToVanishName = "shrinkToVanish"

// partProxyName names the input a renderer binds partProxy's radius to.
const partProxyName = "gsl_partProxy"

// boxFloor is the least squared distance partProxy takes the root of; see
// partProxy.
const boxFloor = 1e-30

// partProxy is a part whose solid is in its box [lo, hi], as the viewer marches
// it (agent/wgsl/internal/codegen/vanish.go): the distance to its box wherever
// that is more than the radius bound to partProxyTree, and the shape itself
// within it, or everywhere while the radius is 0, its default. A renderer
// that marches as the viewer does binds the viewer's radius while it marches
// and leaves it 0 while it shades, so both take the same steps: the box is
// never farther than the part, so either is a step a march can take, but they
// are different steps, and a march stops anywhere within its hit tolerance.
// The parts never shrink here; the viewer's Detail ladder is off for it.
func (d *Delegate) partProxy(op *program.Operand, shapeExpr, loExpr, hiExpr build.Expression) (build.Expression, error) {
	shape, ok1 := shapeExpr.(*TreeNode)
	lo, ok2 := loExpr.(*TreeNode)
	hi, ok3 := hiExpr.(*TreeNode)
	if !ok1 || !ok2 || !ok3 || len(shape.trees) != 1 || len(lo.trees) != 3 || len(hi.trees) != 3 || len(d.pNode.trees) != 3 {
		return nil, build.ErrLoc(d.GetScope(), op.Start, fmt.Errorf("shrinkToVanish: want a number and two vec3 corners"))
	}
	if !d.hasPartProxyTree {
		d.partProxyTree = fidget.NewInput(partProxyName, 0)
		d.hasPartProxyTree = true
	}
	var sum fidget.TreeIndex
	for i := 0; i < 3; i++ {
		p := d.pNode.trees[i]
		outside := fidget.Max(fidget.Max(fidget.Sub(lo.trees[i], p), fidget.Sub(p, hi.trees[i])), d.zero)
		if i == 0 {
			sum = fidget.Square(outside)
		} else {
			sum = fidget.Add(sum, fidget.Square(outside))
		}
	}
	// The box's distance is floored under its root. Inside the box, where all
	// of the part's surface is, the sum is 0 and the root's slope there is
	// infinite, so the gradient is ∞·0, NaN — and far's 0 does not cancel a
	// NaN, so the part's gradient, the normal a renderer shades by, would be
	// NaN wherever it has a surface. Floored, the slope inside is 0 and the
	// part's gradient there is exactly its shape's. The floor raises the box's
	// distance to 1e-15 at most, far under any radius a march binds.
	box := fidget.Sqrt(fidget.Max(sum, fidget.NewConst(boxFloor)))
	radius := d.partProxyTree
	// 1 where the radius is set and the box is beyond it, else 0, as a
	// comparison builds a condition (see BuildSelect2).
	far := fidget.Mul(fidget.Max(d.zero, fidget.Compare(radius, d.zero)), fidget.Max(d.zero, fidget.Compare(box, radius)))
	part := fidget.Add(fidget.Mul(far, box), fidget.Mul(fidget.Sub(d.one, far), shape.trees[0]))
	return NewTreeNode(op, op.Start, part)
}
