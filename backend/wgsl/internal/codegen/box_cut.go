package codegen

import "fmt"

// A declared name.box applies builtin's cropToBox, whose GSL body caps the shape
// with its box: max(shape, the box's distance), a closed solid, which is what
// fidget and the mesher build. The renderer cuts the shape open instead, the way
// the scene's own bounds cut unbounded geometry, so the inside of a part shows
// through where its box crosses it. Both come from one WGSL function, switched
// by a module variable the mesher's shader starts true.
//
// The cut follows from the cap. Where the cap exceeds both the shape and zero,
// the point is outside the box, and the cut keeps the shape's sign at the cap's
// size, so it never reaches zero there; everywhere else it is the shape. Its
// size is still no more than the distance to the cut surface, so a march can
// step by it, and its sign is the shape's, so normals and the renderer's
// inside test are the uncut shape's.

// cropToBoxWGSLName is the name codegen gives builtin's cropToBox.
const cropToBoxWGSLName = "builtin_cropToBox"

// boxCapOverride is the module variable that selects the cap. It is false
// unless a shader starts it true, so everything but the mesher draws the cut.
//
// It is a private variable rather than a pipeline override: any override at
// all makes naga re-evaluate every function's constants when it builds a
// pipeline, and naga 25 to 29 then rejects a function that indexes a constant
// array, as gsl_materialField does, with ExpressionAlreadyInScope (its Emit
// ranges are remapped by their ends only). Metal through wgpu and Firefox
// compile with naga.
const boxCapOverride = "gsl_boxCap"

// cropsToBox reports whether the function being built is builtin's cropToBox,
// cropToBox(shape, lo, hi).
func cropsToBox(wgslName string, paramNames []string) bool {
	return wgslName == cropToBoxWGSLName && len(paramNames) == 3
}

// boxCapDecl declares the variable. It follows cropToBox rather than leading
// it: the viewers hand the module to three.js's wgslFn, which parses it as a
// function and rejects a module that opens with anything else. WGSL resolves a
// module-scope name wherever it is declared.
func boxCapDecl() string {
	return fmt.Sprintf("var<private> %s: bool = false;", boxCapOverride)
}

// boxCutReturn is cropToBox's return statement: capped is its GSL result, local
// a free name to hold it, and shape its field parameter. The sign is taken with
// select, since sign(0) is 0 and would leave the uncut surface outside the box.
func boxCutReturn(local, capped, shape string) string {
	return fmt.Sprintf(""+
		"  let %[1]s = %[2]s;\n"+
		"  if (%[4]s) { return %[1]s; }\n"+
		"  if (%[1]s <= max(%[3]s, 0.0)) { return %[3]s; }\n"+
		"  return select(-1.0, 1.0, %[3]s >= 0.0) * max(abs(%[3]s), %[1]s);\n",
		local, capped, shape, boxCapOverride)
}
