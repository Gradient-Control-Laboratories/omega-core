package convert

import (
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/internal/fidget"
)

// Ensure the color wrapper also satisfies the material hook. It rides on the
// same wrapper for the same reason: only the paths that render appearance pay
// for lowering it, so meshing and plain evaluation stay byte-identical.
var _ build.MaterialReturnsBuilder = (*colorDelegate)(nil)

// materialRows is how many vec4 rows a material record carries, and
// materialRowSize how wide each row is. The layout is the one declared in
// common/web/gsl-shading.js and property.materialDefault: surface, emissive,
// volume. Rows are appended, never reordered, so a future row widens this
// without moving what is already here.
const (
	materialRows    = 4
	materialRowSize = 4
)

// The displacement channel's position: the volume row's third component, a
// signed height in millimetres.
const (
	displacementRow  = 2
	displacementComp = 2
)

// BuildPackageMaterialReturns implements [build.MaterialReturnsBuilder]: it
// records each return's lowered material field as three rows of four trees,
// parallel to the returns themselves, so a renderer can evaluate a shape's
// roughness, metallic, emission and so on at a hit point the same way it
// evaluates its distance and its color. Returns that author no material record
// nothing and fall back to the default record, matching the viewer, where an
// absent material field arrives with present = 0.
func (cd *colorDelegate) BuildPackageMaterialReturns(materialExpressions []build.Expression) error {
	d := cd.Delegate
	mats := make([][materialRows][materialRowSize]fidget.TreeIndex, len(materialExpressions))
	has := make([]bool, len(materialExpressions))
	displaced := make([]bool, len(materialExpressions))
	for i, me := range materialExpressions {
		if me == nil {
			continue
		}
		trees := materialTrees(me)
		if len(trees) != materialRows*materialRowSize {
			// A material field that did not lower to the expected rows is not
			// renderable; skip it rather than failing the build, so an odd
			// material expression costs that shape its material and nothing
			// more. This is how an odd color field is treated too.
			continue
		}
		for r := 0; r < materialRows; r++ {
			for c := 0; c < materialRowSize; c++ {
				mats[i][r][c] = trees[r*materialRowSize+c]
			}
		}
		has[i] = true
		displaced[i] = !constantZeroDisplacement(me)
	}
	d.materialReturns = mats
	d.hasMaterialReturn = has
	d.displacedReturn = displaced
	return nil
}

// constantZeroDisplacement reports whether a lowered material is known to have
// no displacement: a constant record whose height is 0, which is every library
// preset. Anything the lowering did not fold to a constant counts as possibly
// displaced; a height that turns out to be 0 everywhere then costs evaluations
// but never changes a value, since d - 0 is d.
func constantZeroDisplacement(expr build.Expression) bool {
	ce, ok := expr.(build.ConstExpression)
	if !ok {
		return false
	}
	vals := ce.ConstValues()
	if len(vals) != materialRows*materialRowSize {
		return false
	}
	return vals[displacementRow*materialRowSize+displacementComp] == 0
}

// materialTrees flattens a lowered material field to its component trees. The
// flattening is the same one colorTrees performs: a list of vec4 arrives as a
// MultiTreeNode whose nodes each carry a row's trees, so reading them in order
// gives the rows end to end.
func materialTrees(expr build.Expression) []fidget.TreeIndex {
	return colorTrees(expr)
}
