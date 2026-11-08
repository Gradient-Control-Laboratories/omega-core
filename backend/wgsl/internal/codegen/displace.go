package codegen

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
)

// DISPLACEMENT PER RETURN. A material's volume row carries a height the
// viewer moves the surface by (common/web/gsl-displace.js). Subtracting only
// the height of whichever return wins the undisplaced composite is wrong where
// two returns overlap — a screw in the hole tapped for it — because a point
// inside both takes one part's thread and not the other's, and the union
// opens where one part's groove crosses the other part's metal. So each
// displaced return moves its OWN distance before the composite, as the
// fidget renderer's litDisplaced does:
//
//	gsl_sdShapeDisplaced = gsl_min over i of gsl_shape(d_i - h_i * k, i)
//
// The distances are built once, in gsl_shapeParts, which sdShape reads too, so
// the geometry is not emitted twice; the heights are built together in
// gsl_heights, so a value two heights share is computed once. A program with
// no displaced return keeps its sdShape exactly as before.

const (
	shapePartsFn     = "gsl_shapeParts"
	heightsFn        = "gsl_heights"
	displacedShapeFn = "gsl_sdShapeDisplaced"
)

// The displacement channel's position in a material record: the volume row's
// third component.
const (
	displacementRow  = 2
	displacementComp = 2
	materialRowSize  = 4
)

// buildHeights lowers each return's displacement height, skipping any known to
// be the constant zero (every library preset), and records gsl_heights when at
// least one remains.
func (d *Delegate) buildHeights(materialExpressions []build.Expression) error {
	heights := make([]build.Expression, len(materialExpressions))
	mask := make([]bool, len(materialExpressions))
	anyHeight := false
	for i, me := range materialExpressions {
		if me == nil {
			continue
		}
		h, err := heightExpression(me)
		if err != nil {
			return err
		}
		if h == nil {
			continue
		}
		if z, ok := constHeight(me); ok && z != 0 {
			d.heightEvident = true
		}
		heights[i] = h
		mask[i] = true
		anyHeight = true
	}
	if !anyHeight {
		return nil
	}

	childSrc, bodyLets, caseSrc := d.codegenFieldCasesShared(heightsFn, heights)
	vals := make([]string, len(caseSrc))
	for i, src := range caseSrc {
		if src == "" {
			vals[i] = "0.0"
			continue
		}
		vals[i] = src
	}

	var sb strings.Builder
	if childSrc != "" {
		sb.WriteString(childSrc)
		sb.WriteString("\n\n")
	}
	fmt.Fprintf(&sb, "fn %s(p: vec3<f32>, time: f32%s) -> array<f32, %d> {\n",
		heightsFn, d.fieldFnParamSuffix(), len(caseSrc))
	sb.WriteString(bodyLets)
	fmt.Fprintf(&sb, "  return array<f32, %d>(%s);\n", len(caseSrc), strings.Join(vals, ", "))
	sb.WriteString("}")

	d.heightsSrc = sb.String()
	d.displacedMask = mask
	d.heightSrcs = caseSrc
	return nil
}

// heightsDiffer reports whether two displaced returns carry different heights.
// Only then can a point inside both take the wrong one: where every displaced
// return moves by the same height h, min(d_i - h) is min(d_i) - h, and the
// viewer's plain path, which subtracts the winner's height, is exact. One
// displaced return is the same case. A screw in its tapped hole differs (the
// male and female reliefs are different heights), so it keeps the composite.
func (d *Delegate) heightsDiffer() bool {
	first, seen := "", false
	for i, m := range d.displacedMask {
		if !m || i >= len(d.heightSrcs) {
			continue
		}
		if !seen {
			first, seen = d.heightSrcs[i], true
			continue
		}
		if d.heightSrcs[i] != first {
			return true
		}
	}
	return false
}

// constHeight is the height of a material record that lowered to a constant.
func constHeight(me build.Expression) (float64, bool) {
	if ce, ok := me.(build.ConstExpression); ok {
		if vals := ce.ConstValues(); len(vals) > displacementRow*materialRowSize+displacementComp {
			return vals[displacementRow*materialRowSize+displacementComp], true
		}
	}
	return 0, false
}

// displacementEvident reports whether the program can displace anything: a
// record with a constant non-zero height, or a call to the one library
// function that sets a height, material.withDisplacement, anywhere in the
// module. A height the lowering could not fold — a preset carried through
// withRoughness, a blend of presets — is zero unless one of those put a height
// in; without the evidence the composite would only pay for heights of zero,
// and the viewer's plain path is exact for them.
func (d *Delegate) displacementEvident(module string) bool {
	return d.heightEvident ||
		strings.Contains(module, "withDisplacement(") ||
		strings.Contains(d.materialFieldSrc, "withDisplacement(")
}

// heightExpression is the height of one lowered material record, or nil when
// the record is a constant with no height. A record lowered as its three rows
// gives the volume row directly; anything else is indexed.
func heightExpression(me build.Expression) (build.Expression, error) {
	if ce, ok := me.(build.ConstExpression); ok {
		if vals := ce.ConstValues(); vals != nil {
			if len(vals) > displacementRow*materialRowSize+displacementComp &&
				vals[displacementRow*materialRowSize+displacementComp] == 0 {
				return nil, nil
			}
		}
	}
	var row wgsl.Expression
	if c, ok := me.(*wgsl.Collect); ok && (len(c.Args) == 3 || len(c.Args) == materialRowSize) {
		row = c.Args[displacementRow]
		if z, ok := constRowComponent(c.Element(displacementRow), displacementComp); ok && z == 0 {
			return nil, nil
		}
	} else {
		r, err := wgsl.NewListIndexedAccess(nil, me.(wgsl.Expression), displacementRow, wgsl.TypeVec4)
		if err != nil {
			return nil, err
		}
		row = r
	}
	return wgsl.NewVecMemberAccess("z", nil, row)
}

// constRowComponent reads component comp of a material row when the row is a
// constant: a literal vec4, or a row of a constant record such as a library
// preset (metal.steel[2]), lowered with three rows or as a mat4x4's four. The
// second result is false when it is not known.
func constRowComponent(row build.Expression, comp int) (float64, bool) {
	if ce, ok := row.(build.ConstExpression); ok {
		if vals := ce.ConstValues(); len(vals) == materialRowSize {
			return vals[comp], true
		}
	}
	if a, ok := row.(*wgsl.ArrayAccess); ok && len(a.Indices) == 1 {
		if ce, ok := a.Target.(build.ConstExpression); ok {
			vals := ce.ConstValues()
			i := a.Indices[0]*materialRowSize + comp
			if (len(vals) == 3*materialRowSize || len(vals) == materialRowSize*materialRowSize) && i < len(vals) {
				return vals[i], true
			}
		}
	}
	return 0, false
}

// displacedComposite returns sdShape, reading its returns' distances from
// gsl_shapeParts, and gsl_sdShapeDisplaced, which subtracts each displaced
// return's own height, scaled by gsl_dispK, before the same composite.
// paramSuffix is sdShape's parameter list after p and time.
func (d *Delegate) displacedComposite(paramSuffix string, n int) (sdShapeSrc, displacedSrc string) {
	args := argsFromParams(paramSuffix)
	fieldArgs := argsFromParams(d.fieldFnParamSuffix())

	composite := func(height func(i int) string) string {
		var sb strings.Builder
		chain := ""
		for i := 0; i < n; i++ {
			fmt.Fprintf(&sb, "  let gsl_s%d = gsl_shape(gsl_d[%d]%s, %du);\n", i+1, i, height(i), i+1)
			if i == 0 {
				chain = "gsl_s1"
			} else {
				chain = fmt.Sprintf("gsl_min(%s, gsl_s%d)", chain, i+1)
			}
		}
		fmt.Fprintf(&sb, "  return %s;\n}", chain)
		return sb.String()
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "fn sdShape(p: vec3<f32>, time: f32%s) -> gsl_shape {\n", paramSuffix)
	fmt.Fprintf(&sb, "  let gsl_d = %s(p, time%s);\n", shapePartsFn, args)
	sb.WriteString(composite(func(int) string { return "" }))
	sdShapeSrc = sb.String()

	sb.Reset()
	fmt.Fprintf(&sb, "fn %s(p: vec3<f32>, time: f32, gsl_dispK: f32%s) -> gsl_shape {\n", displacedShapeFn, paramSuffix)
	fmt.Fprintf(&sb, "  let gsl_d = %s(p, time%s);\n", shapePartsFn, args)
	fmt.Fprintf(&sb, "  let gsl_h = %s(p, time%s);\n", heightsFn, fieldArgs)
	sb.WriteString(composite(func(i int) string {
		if i < len(d.displacedMask) && d.displacedMask[i] {
			return fmt.Sprintf(" - gsl_h[%d] * gsl_dispK", i)
		}
		return ""
	}))
	displacedSrc = sb.String()
	return sdShapeSrc, displacedSrc
}

// argsFromParams turns a comma-prefixed parameter list (", a: f32, b: vec3<f32>")
// into the matching argument list (", a, b"). No parameter type here contains a
// comma.
func argsFromParams(params string) string {
	var sb strings.Builder
	for _, param := range strings.Split(params, ",") {
		param = strings.TrimSpace(param)
		if param == "" {
			continue
		}
		name, _, _ := strings.Cut(param, ":")
		sb.WriteString(", ")
		sb.WriteString(strings.TrimSpace(name))
	}
	return sb.String()
}

// dispDrawnDeclWGSL declares what displacementDrawn() reads: 0, nothing
// displaced, until a fragment that marches the displaced field sets it at a
// hit to the weight the displacement was drawn with there (gsl-render-graph.js).
// Every other host (the mesher, bounds, probes, the simple tier) leaves it 0.
const dispDrawnDeclWGSL = "var<private> gsl_dispDrawn: f32 = 0.0;"
