
---

## **Library Packages**

**Call these rather than writing the same code yourself.** The library's versions are SDF-correct, consistently named, and tested. If you need a helper the library doesn't provide, name it in the library's style (e.g. `cylinderHelical`, not `helicalCyl`) so it composes cleanly with the rest.

### Core & Standard Packages

The core package `builtin` is implicitly dot-imported for every file, exposing basic shapes and operations (like `sphere(1)`, `mix(a, b, t)`) without any import statement. All other standard GSL packages (`animated`, `beam`, `catalog`, `coating`, `color`, `drill`, `driver`, `ease`, `fastener`, `foam`, `gears`, `glass`, `honeycomb`, `joint`, `lattice`, `machined`, `map`, `material`, `matrix`, `mechanical`, `metal`, `natural`, `noise`, `physical`, `playful`, `printed`, `spinodoid`, `stone`, `text`, `texture`, `thread`, `tpms`, `transform`, `vector`, `weathering`, `wood`) are automatically imported, allowing you to reference their members directly using the package prefix (e.g. `color.red`, `ease.quadOut(t)`). A nested package is referenced by the last segment of its path alone — write `beam.bcc(10)`, not `lattice.beam.bcc(10)`.

To import a custom or product-specific library, put an `import` line at the top of the file. Two forms:

```gsl
import "path/to/pkg";      // qualified: call as `pkg.func(1)` (prefix = last path segment)
import alias "path/to/pkg";  // aliased:   call as `alias.func(1)`
```

---

### `builtin`

#### Constants

- **`pi = atan(1) * 4`**
- **`tau = pi * 2`**
- **`sqrt2 = sqrt(2)`**
- **`sqrt3 = sqrt(3)`**
- **`e = exp(1)`**
- **`inch = 25.4`** — Millimeters are the unit; these convert into them. Write 2.0 * inch rather than 50.8, so the drawing's number stays visible in the source. `in` is a reserved word, hence `inch`, and there is no `mm` because 1.0 * mm says nothing that 1.0 does not.
- **`foot = 304.8`** — Feet to millimeters.
- **`meter = 1000.0`** — Meters to millimeters.
- **`nothing = 1e5`** — The empty solid: far from everything, so min(d, nothing) and max(d, -nothing) are just d. It is finite on purpose: the mesher lowers c ? a : b as c * a + (1 - c) * b, which drops the branch not taken by multiplying it by 0, and 0 times infinity is not 0.
- **`boxPad = 0.01`** — How far past a declared name.box its crop reaches: ten times the raycaster's hit floor, so a box computed from a part's nominal dimensions never shaves float-level overshoot off a face that touches it.

#### CONSTANTS

- **`firstOf(list)`** — First element of a list. A scalar behaves as a list of length one, so firstOf(x) == x — handy for arguments that broadcast a single value or accept a per-slot list. Pairs with lastOf().
- **`lastOf(list)`** — Last element of a list. A scalar behaves as a list of length one, so lastOf(x) == x. Pairs with firstOf() for broadcast-or-per-slot arguments.

#### BOOLEAN OPERATIONS

- **`intersectSharp(d1, d2)`** — Sharp CSG intersection of two SDFs (logical AND of solid regions).
- **`intersectSharp(d1, d2, d3)`** — Sharp CSG intersection of three SDFs.
- **`intersectSharp(fields)`** — Sharp CSG intersection of N SDFs. Pass [d1, d2, ...] as one collection.
- **`unionSharp(d1, d2)`** — Sharp CSG union of two SDFs (logical OR of solid regions).
- **`unionSharp(d1, d2, d3)`** — Sharp CSG union of three SDFs.
- **`unionSharp(fields)`** — Sharp CSG union of N SDFs.
- **`differenceSharp(target, tool)`** — Sharp CSG difference: target minus tool.
- **`intersectRound(d1, d2, radius)`** — Round CSG intersection with fillet `radius`. The intersection edge becomes a smooth arc of about that radius. At radius 0 the surface is intersectSharp's, but the field is not: outside a convex corner it measures the Euclidean distance to the corner instead of the max.
- **`intersectRound(d1, d2, d3, radius)`** — Round CSG intersection of three SDFs with fillet `radius`. Fillets all three pairwise edges.
- **`intersectRound(fields, radius)`** — Round CSG intersection of N SDFs.
- **`intersectRound(radius)`** — Partial application: binds the fillet radius and returns the intersection operator itself, ready to hand to a primitive that takes its own blend — `boxCentered(center, size, intersectRound(0.2))` rounds every edge of the box by 0.2. The returned operator takes the list of fields to intersect.
- **`unionRound(d1, d2, radius)`** — Round CSG union with fillet `radius`. The union seam becomes a smooth arc of about that radius.
- **`unionRound(d1, d2, d3, radius)`** — Round CSG union of three SDFs with fillet `radius`.
- **`unionRound(fields, radius)`** — Round CSG union of N SDFs.
- **`unionRound(radius)`** — Partial application of unionRound: binds the fillet radius and returns the operator, ready to hand to a primitive that takes its own blend.
- **`differenceRound(target, tool, radius)`** — Round CSG difference: target minus tool with a fillet of `radius` where the tool carves into the target.
- **`intersectLp(fields, width, exponent)`** — Lp-norm blend intersection of a list of SDFs. Interpolates between a sharp intersection (exponent -> infinity) and a round one (exponent == 2) by combining the clamped per-field insets with an Lp norm instead of intersectRound's L2 norm. A larger exponent is sharper; `width` sets the blend width.
- **`intersectLp(d1, d2, width, exponent)`** — Lp-norm blend intersection of two SDFs (see the list form).
- **`intersectLp(d1, d2, d3, width, exponent)`** — Lp-norm blend intersection of three SDFs (see the list form).
- **`intersectLp(width, exponent)`** — Partial application of intersectLp: binds the blend width and exponent and returns the operator, ready to hand to a primitive that takes its own blend.
- **`unionLp(fields, width, exponent)`** — Lp-norm blend union of a list of SDFs (De Morgan dual of intersectLp): interpolates between a sharp union (exponent -> infinity) and a round one (exponent == 2). A larger exponent is sharper; `width` sets the blend width.
- **`unionLp(d1, d2, width, exponent)`** — Lp-norm blend union of two SDFs (see the list form).
- **`unionLp(d1, d2, d3, width, exponent)`** — Lp-norm blend union of three SDFs (see the list form).
- **`unionLp(width, exponent)`** — Partial application of unionLp: binds the blend width and exponent and returns the operator, ready to hand to a primitive that takes its own blend.
- **`differenceLp(target, tool, width, exponent)`** — Lp-norm blend difference: target minus tool, with an Lp-blended seam (see intersectLp).
- **`intersectExponential(fields, width)`** — Exponential (log-sum-exp) blend intersection of a list of SDFs — a smooth maximum whose softness is set by `width`. A larger width blends more; a small one approaches the sharp intersection.
- **`intersectExponential(d1, d2, width)`** — Exponential blend intersection of two SDFs (see the list form).
- **`intersectExponential(d1, d2, d3, width)`** — Exponential blend intersection of three SDFs (see the list form).
- **`intersectExponential(width)`** — Partial application of intersectExponential: binds the blend width and returns the operator, ready to hand to a primitive that takes its own blend.
- **`unionExponential(fields, width)`** — Exponential (log-sum-exp) blend union of a list of SDFs (De Morgan dual of intersectExponential) — a smooth minimum. A larger width blends more.
- **`unionExponential(d1, d2, width)`** — Exponential blend union of two SDFs (see the list form).
- **`unionExponential(d1, d2, d3, width)`** — Exponential blend union of three SDFs (see the list form).
- **`unionExponential(width)`** — Partial application of unionExponential: binds the blend width and returns the operator, ready to hand to a primitive that takes its own blend.
- **`differenceExponential(target, tool, width)`** — Exponential blend difference: target minus tool, with a log-sum-exp seam (see intersectExponential).
- **`intersectChamferEdges(fields, setback)`** — Strongest pairwise (edge) chamfer plane of a list of SDFs at `setback`.
- **`intersectChamferVertices(fields, setback)`** — Strongest three-way (vertex) chamfer plane of a list of SDFs at `setback`.
- **`unionChamferEdges(fields, setback)`** — Union counterpart of intersectChamferEdges.
- **`unionChamferVertices(fields, setback)`** — Union counterpart of intersectChamferVertices.
- **`intersectChamfer(d1, d2, setback)`** — Chamfered CSG intersection: a bevel set back by `setback` along each face replaces the sharp intersection edge. The bevel is at 45° only where the faces meet at a right angle; its plane's gradient has length sqrt(1 + normal1·normal2), so at obtuse edges the field overstates distance.
- **`intersectChamfer(d1, d2, d3, setback)`** — Chamfered CSG intersection of three SDFs. Chamfers all three pairwise edges and the three-way vertex where they meet.
- **`intersectChamfer(fields, setback)`** — Chamfered CSG intersection of N SDFs. `setback` is a scalar that applies to edges and vertices alike, or pass [edgeSetback, vertexSetback] to set them separately.
- **`intersectChamfer(setback)`** — Partial application of intersectChamfer: binds the chamfer setback and returns the operator, ready to hand to a primitive that takes its own blend.
- **`unionChamfer(d1, d2, setback)`** — Chamfered CSG union with bevel set back by `setback`.
- **`unionChamfer(d1, d2, d3, setback)`** — Chamfered CSG union of three SDFs.
- **`unionChamfer(fields, setback)`** — Chamfered CSG union of N SDFs. `setback` is a scalar, or [edgeSetback, vertexSetback] to set edges and vertices separately.
- **`unionChamfer(setback)`** — Partial application of unionChamfer: binds the chamfer setback and returns the operator, ready to hand to a primitive that takes its own blend.
- **`differenceChamfer(target, tool, setback)`** — Chamfered CSG difference: target minus tool with a chamfered cut.
- **`intersectChamferRound(d1, d2, setback, radius)`** — Chamfered intersection whose chamfer planes are joined by a round fillet of `radius`: intersectChamferBlend with intersectRound(radius). A bevel that fades into a fillet rather than ending in a hard line. Radius 0 gives intersectChamfer's surface.
- **`intersectChamferRound(d1, d2, d3, setback, radius)`** — Chamfered + rounded intersection of three SDFs.
- **`intersectChamferRound(fields, setback, radius)`** — Chamfered + rounded CSG intersection of N SDFs. `setback` is a scalar, or [edgeSetback, vertexSetback] to set edges and vertices separately; `radius` is the fillet radius. One fillet joins the face, edge and vertex planes, so edge and vertex rounding are not independent.
- **`intersectChamferRound(setback, radius)`** — Partial application of intersectChamferRound: binds the chamfer setback and rounding radius and returns the operator, ready to hand to a primitive that takes its own blend.
- **`unionChamferRound(d1, d2, setback, radius)`** — Chamfered union whose chamfer planes are joined by a round fillet of `radius`: unionChamferBlend with unionRound(radius).
- **`unionChamferRound(d1, d2, d3, setback, radius)`** — Chamfered + rounded union of three SDFs.
- **`unionChamferRound(fields, setback, radius)`** — Chamfered + rounded CSG union of N SDFs (see intersectChamferRound for the setback and radius parameters).
- **`unionChamferRound(setback, radius)`** — Partial application of unionChamferRound: binds the chamfer setback and rounding radius and returns the operator, ready to hand to a primitive that takes its own blend.
- **`differenceChamferRound(target, tool, setback, radius)`** — Chamfered + rounded CSG difference: target minus tool with a chamfered cut whose planes are joined by a round fillet of `radius`.
- **`intersectChamferBlend(d1, d2, setback, blend)`** — Chamfered intersection whose chamfer planes are combined by `blend` instead of a fixed fillet. blend is an intersection operator over a list, such as intersectRound(radius), intersectLp(width, exponent) or intersectExponential(width); its width is used as given. intersectSharp gives back intersectChamfer.
- **`intersectChamferBlend(d1, d2, d3, setback, blend)`** — Chamfered intersection of three SDFs with its chamfer planes combined by `blend`.
- **`intersectChamferBlend(fields, setback, blend)`** — Chamfered CSG intersection of N SDFs with its chamfer planes combined by `blend`. `setback` is a scalar, or [edgeSetback, vertexSetback]. One blend joins the face, edge and vertex planes, so edge and vertex blending are not independent.
- **`intersectChamferBlend(setback, blend)`** — Partial application of intersectChamferBlend: binds the chamfer setback and blend and returns the operator, ready to hand to a primitive that takes its own blend.
- **`unionChamferBlend(d1, d2, setback, blend)`** — Chamfered union whose chamfer planes are combined by `blend`, a union operator over a list such as unionRound(radius), unionLp(width, exponent) or unionExponential(width). unionSharp gives back unionChamfer.
- **`unionChamferBlend(d1, d2, d3, setback, blend)`** — Chamfered union of three SDFs with its chamfer planes combined by `blend`.
- **`unionChamferBlend(fields, setback, blend)`** — Chamfered CSG union of N SDFs with its chamfer planes combined by the union operator `blend` (see intersectChamferBlend for setback).
- **`unionChamferBlend(setback, blend)`** — Partial application of unionChamferBlend: binds the chamfer setback and blend and returns the operator, ready to hand to a primitive that takes its own blend.
- **`differenceChamferBlend(target, tool, setback, blend)`** — Chamfered CSG difference: target minus tool, with the chamfer planes combined by the intersection operator `blend` (see intersectChamferBlend).
- **`remapEdges(fields, profile)`** — Edge-coordinate fillets. For each pair of fields in the list, `profile` (a nullary function returning a 2D field of p) is evaluated in the local coordinate where x and y are the two signed distances. The result is the nearest such fillet across all pairs. profile() == length(p.xy) - r gives a round fillet of radius r along every shared edge; other profiles give other cross-sections. Slots that aren't a pair fall back to nothing so they lose the min.
- **`unionRoundExact(field1, field2, normal1, normal2)`** — Exact rounded union of two fields (nullary field1/field2) that uses their surface normals normal1, normal2 to scale the blend, staying accurate where the two gradients differ. The zero set is the sharp union; the field is its exact SDF, so offsetting the result by r rounds the seam with radius r.
- **`unionRoundExact(field1, field2)`** — Exact rounded union of two fields (nullary field1/field2), taking each field's surface normal from its own gradient instead of being handed them. Use the four-argument form when a normal is known analytically, or when the same gradient is already computed elsewhere.
- **`intersectRoundExact(field1, field2, normal1, normal2)`** — Exact rounded intersection of two fields (nullary field1/field2) that uses their surface normals normal1, normal2 to scale the blend. The De Morgan dual of unionRoundExact: negating a field flips its gradient, so the normals are negated alongside it.
- **`intersectRoundExact(field1, field2)`** — Exact rounded intersection of two fields (nullary field1/field2), taking each field's surface normal from its own gradient instead of being handed them. The gradients are of field1 and field2 themselves; the four-argument form negates them along with the fields.
- **`intersectRvachev(fields, band)`** — Rvachev R0 intersection of a list of SDFs, gated to a transition band of width `band`. Beyond -band (well inside the sharp solid) it is the sharp intersection; within the band it eases (cubic smoothstep) toward pure R0, sum(d) + sqrt(sum(d*d)). For two inputs R0's zero set is the sharp intersection's, so band shapes the field, not the surface; with three or more the n-ary form is not associative and the boundary shifts near multi-input corners.
- **`intersectRvachev(d1, d2, band)`** — Rvachev R0 intersection of two SDFs (see the list form).
- **`intersectRvachev(d1, d2, d3, band)`** — Rvachev R0 intersection of three SDFs (see the list form).
- **`intersectRvachev(band)`** — Partial application of intersectRvachev: binds the blend width and returns the operator, ready to hand to a primitive that takes its own blend.
- **`unionRvachev(fields, band)`** — Rvachev R0 union of a list of SDFs (via De Morgan's law).
- **`unionRvachev(d1, d2, band)`** — Rvachev R0 union of two SDFs (see the list form).
- **`unionRvachev(d1, d2, d3, band)`** — Rvachev R0 union of three SDFs (see the list form).
- **`unionRvachev(band)`** — Partial application of unionRvachev: binds the blend width and returns the operator, ready to hand to a primitive that takes its own blend.
- **`differenceRvachev(target, tool, band)`** — Rvachev R0 difference: target minus tool.

#### UTILITIES

- **`mix(a, b, t)`** — Linear interpolation between a and b by parameter t. t = 0 returns a, t = 1 returns b; values outside [0, 1] extrapolate.
- **`linearMap(value, inMin, inMax, outMin, outMax)`** — Linearly remap value from [inMin, inMax] to [outMin, outMax]. Does not clamp — values outside the input range extrapolate. Use ramp() when you want clamping.
- **`ramp(value, inMin, inMax, outMin, outMax)`** — Like linearMap() but clamps the result to the output range, so values outside [inMin, inMax] saturate instead of extrapolating.
- **`rampOffset(field, modulator, modulatorMin, modulatorMax, offsetMin, offsetMax)`** — Like ramp(), but normalizes the ramped offset with the assumption that the modulator is orthogonal to the field to be offset.
- **`rampOffset(field, modulator, modulatorMin, modulatorMax, offsetMin, offsetMax, blendMin, blendMax)`** — Like rampOffset(), but with blend radii.

#### OFFSETS & FIELD BLENDS

- **`shell(d, thickness)`** — Shell / onion: hollow a solid SDF `d` into a wall of `thickness` centered on the surface.
- **`shell(d, thickness, bias)`** — Shell with a `bias` in [-1, 1] that shifts the wall inward (-1) or outward (+1) relative to the surface; 0 is centered.
- **`twoBodyField(d1, d2)`** — Two-body Cayley field: (d1 - d2) / (d1 + d2). A normalized interpolant of two fields whose zero set is where d1 = d2.
- **`twoBodyPolar(d1, d2, angle)`** — Two-body polar blend: d1*cos(angle) + d2*sin(angle). Sweeps between fields d1 and d2 as `angle` goes from 0 to pi/2.

#### 2D PRIMITIVES (XY plane)

- **`circleXY(center, radius)`** — Circle in the XY plane, centered at `center` (2D point) with given radius. SDF is signed: negative inside, positive outside.
- **`annulusXY(center, radius, thickness)`** — Ring (annulus) in the XY plane with given radius and stroke thickness, centered at `center`.
- **`rectangleXYCentered(center, size, blend)`** — Axis-aligned rectangle in the XY plane, centered at `center` (2D point) with given `size` (2D extent), using `blend` to combine the two slab fields — pass `intersectRound(radius)` to round every corner, or any other operator taking a list of fields.
- **`rectangleXYCentered(center, size)`** — Axis-aligned rectangle in the XY plane centered at `center` (2D point) with given `size` (2D extent).
- **`rectangleXY(pointA, pointB, blend)`** — Axis-aligned rectangle in the XY plane defined by two opposite corners, using `blend` to combine the two slab fields — pass `intersectRound(radius)` to round every corner, or any other operator taking a list of fields.
- **`rectangleXY(pointA, pointB)`** — Axis-aligned rectangle in the XY plane defined by two opposite corners.
- **`quadXY(corner0, corner1, corner2, corner3)`** — Quadrilateral in the XY plane with corners corner0..corner3 given in consistent winding order.
- **`arcXY(halfAngleSinCos, radius)`** — Circular arc in the XY plane. `halfAngleSinCos` is [sin, cos] of the arc's half-angle and `radius` the arc radius. Returns the distance to the arc curve; subtract a half-width to give it a stroke.
- **`arcExtendedXY(halfAngleSinCos, radius)`** — Arc (as in arcXY) extended along the tangents at its endpoints. `halfAngleSinCos` is [sin, cos] of the half-angle and `radius` the radius.
- **`arcTwoPointTangentExtendedXY(pointA, pointB, tangentA)`** — Arc specified by two endpoints (2D, XY) and the unit tangent at pointA. Returns a 2D SDF for the arc curve; subtract a half-width for a stroke.

#### 3D PRIMITIVES

- **`sphere(radius)`** — Sphere centered at the origin with the given radius. Use remap() to place it elsewhere or to project it into a different dimension.
- **`sphere(center, radius)`** — Sphere of the given radius centered at `center`.
- **`sphere(center, radius, pole, reference)`** — Sphere of the given radius centered at `center`, its frame (.u, .v) measured about `pole` from `reference`: .u the azimuth round the pole from the reference and .v the angle down from the pole, both in radians. Without the two directions the pole is +Z and the reference +X. `reference` need not be perpendicular to `pole`, only not parallel to it.
- **`axis(center, direction)`** — Distance to an infinite line through `center` along `direction`. `direction` does not need to be normalized. Its frame: .u the angle round the line in radians, .v the signed distance along it from `center` in mm.
- **`axisX(center)`** — Distance to an infinite line through `center` parallel to the X axis.
- **`axisY(center)`** — Distance to an infinite line through `center` parallel to the Y axis.
- **`axisZ(center)`** — Distance to an infinite line through `center` parallel to the Z axis.
- **`cylinder(center, direction, radius)`** — Infinite cylinder along `direction` through `center`, with given radius. `direction` does not need to be normalized. Its frame: .u the angle round the axis in radians, .v the signed distance along it from `center` in mm.
- **`cylinderX(center, radius)`** — Infinite cylinder along the X axis through `center`, with given radius.
- **`cylinderY(center, radius)`** — Infinite cylinder along the Y axis through `center`, with given radius.
- **`cylinderZ(center, radius)`** — Infinite cylinder along the Z axis through `center`, with given radius.
- **`cylinderCapped(pointA, pointB, radius, blend)`** — Finite cylinder between pointA and pointB, with flat caps, using `blend` to combine the radial and axial fields — pass `intersectRound(radius)` to fillet the rim, or any other operator taking a list of fields. Its frame: .u the angle round the axis in radians, on the side and the caps alike; .v the signed distance along the axis from the middle on the side, and the distance from the axis on a cap, in mm — so across the rim .u runs on and .v turns from along the axis to across it.
- **`cylinderCapped(pointA, pointB, radius)`** — Finite cylinder between pointA and pointB, with flat caps.
- **`lineSegment(pointA, pointB)`** — Infinitely thin line segment: unsigned distance from p to the segment pointA -> pointB. Subtract a radius to get a capsule (same as cylinderRounded). The building block for beam and foam lattices.
- **`cylinderRounded(pointA, pointB, radius)`** — Capsule between pointA and pointB: cylinder with hemispherical caps of the given radius.
- **`torus(majorRadius, minorRadius)`** — Torus centered at the origin, lying in the XY plane. `majorRadius` is the ring radius (center to tube center); `minorRadius` is the tube radius. Its frame is a sphere's bent round the ring: .u the azimuth round +Z from +X, and .v the angle round the tube from its top, as a sphere's polar angle is from its pole — 0 on top, pi/2 on the outer equator, pi underneath, -pi/2 on the inner equator — both in radians. Unlike a sphere's, .v wraps, along the bottom of the tube.
- **`cone(origin, direction, angle)`** — Infinite cone with apex at `origin`, opening along `direction`, with given half-angle (in radians). Its frame: .u the angle round the axis in radians, .v the signed distance along it from the apex in mm.
- **`coneCapped(pointA, pointB, radiusA, radiusB)`** — Capped cone between pointA (radius radiusA) and pointB (radius radiusB).
- **`coneRounded(pointA, pointB, radiusA, radiusB)`** — Rounded cone: sphere at pointA (radius radiusA), sphere at pointB (radius radiusB), blended with a tangent cone — like a tapered capsule.
- **`plane(point, normal)`** — Plane through `point` with `normal`. Returns signed distance: negative on the side the normal points away from. The normal does not need to be normalized. Its frame: .u and .v the signed distances from `point` along two perpendicular directions in the plane, in mm, right-handed with the normal.
- **`boxCentered(center, size, blend)`** — Axis-aligned box centered at `center` (3D point) with given `size` (3D extent), using `blend` to combine the three slab fields — pass `intersectRound(radius)` to round every edge, or any other operator taking a list of fields. Its frame, on each face, is the signed distance from the center along the face's two in-plane axes, in mm, taken in turn — (y, z) on the x faces, (z, x) on the y faces, (x, y) on the z faces — so across every edge .u and .v each turn a right angle rather than running on.
- **`boxCentered(center, size)`** — Axis-aligned box centered at `center` (3D point) with given `size` (3D extent).
- **`box(pointA, pointB, blend)`** — Axis-aligned box defined by two opposite corners, using `blend` to combine the three slab fields — pass `intersectRound(radius)` to round every edge, or any other operator taking a list of fields. Its frame is boxCentered's.
- **`box(pointA, pointB)`** — Axis-aligned box defined by two opposite corners.
- **`cropToBox(shape, low, high)`** — `shape` capped by the box between corners `low` and `high`, grown by boxPad: the closed solid that export meshes. A declaration's name.box = [low, high] applies this. The viewer draws it cut open instead, so where the box crosses the part its inside shows through, the way the scene's bounds cut unbounded geometry.
- **`octahedron(size)`** — Regular octahedron at the origin, with vertices at (±size, 0, 0), (0, ±size, 0), (0, 0, ±size).

#### QUADRATIC BEZIER CURVES

- **`bezierQuadPointRound(start, control, end)`** — Quadratic Bézier curve from `start` through `control` to `end`. Returns (distance, t, closestPoint) where t is the curve parameter at the closest point on the curve to p, and closestPoint is that point.
- **`bezierQuadTanRound(start, control, end)`** — Same as bezierQuadPointRound but also returns the unit tangent at the closest point. Returns (distance, t, closestPoint, tangent).
- **`bezierQuadRound(start, control, end)`** — Convenience wrapper around bezierQuadPointRound that returns only the distance to the curve.

#### SWEEP LOFT

- **`bezierTubeDir(start, control, end, radius, orientation)`** — Distance to a quadratic Bezier tube from `start` through `control` to `end` with `radius`; `orientation` orients the inside/outside sign. Returns (signedDistance, t, arcLength, closestPoint, tangent).
- **`sweepLoft(startProfile, endProfile, start, control, end, orientation, blend)`** — Sweep a 2D profile along a quadratic Bezier start -> control -> end, lofting from startProfile at the start to endProfile at the end. The profile arguments are nullary functions returning a 2D field of p.
- **`sweepLoft(startProfile, endProfile, start, control, end, orientation)`** — As sweepLoft, combining the profile with its two end caps sharply.
- **`sweepLoftScale(startProfile, endProfile, start, control, end, orientation, scale, blend)`** — Like sweepLoft, but scales the profile linearly from 1 to `scale` along the sweep.
- **`sweepLoftScale(startProfile, endProfile, start, control, end, orientation, scale)`** — As sweepLoftScale, combining the profile with its two end caps sharply.
- **`sweepLoftScaleSmooth(startProfile, endProfile, start, control, end, orientation, scale, blend)`** — Like sweepLoftScale, but eases the scale with a smoothstep along the sweep instead of linearly.
- **`sweepLoftScaleSmooth(startProfile, endProfile, start, control, end, orientation, scale)`** — As sweepLoftScaleSmooth, combining the profile with its two end caps sharply.
- **`sweepLoftScaleSmoothSeparateCaps(startProfile, endProfile, start, control, end, orientation, scale)`** — Variant of sweepLoftScaleSmooth that returns the swept profile and the two end caps separately as (profileDistance, startCap, endCap).

### `map`

- **`patternPolarXY(count)`** — Polar repeat in the XY plane: returns a remapped p so any shape evaluated against it is replicated `count` times around the Z axis. Use as the second argument to remap().
- **`triangleWave(x, period)`** — triangleWave(x, period) folds x into [-period/2, period/2] as a triangle wave: it equals x near 0, rises to +period/2, then mirrors back — continuous everywhere, unlike a modulo sawtooth which jumps. The full (mirror) period is 2 * period; zeros recur every `period`.
- **`repeatCentered(x, period)`** — repeatCentered(x, period) folds x into [-period/2, period/2] with a modulo sawtooth: every cell is an *identical* translated copy (unlike triangleWave, which mirrors alternate cells). The field jumps at each cell boundary, so use it to tile a self-contained cell, not to fold a single asymmetric shape. repeatCentered also works component-wise on vectors: pass a vec2/vec3 `x` and matching `period` (no zero component) to fold every axis at once.
- **`patternRectangularXY(cell)`** — Rectangular (grid) repeat in the XY plane: returns a remapped p so any shape at the origin tiles across a `cell` = [width, height] grid (pass [s, s] for a square). Built from triangleWave, so adjacent tiles are mirror images and the remapped field stays continuous (no modulo seam). z passes through. Use as the second argument to remap().
- **`patternRepeat(size)`** — 3D translational repeat: returns a remapped p so any shape at the cell center tiles across a vec3 `size` grid by identical (non-mirrored) copies. Built from repeatCentered per axis, so each cell is a translated duplicate with a seam at the boundary — fold a self-contained cell, not an asymmetric motif. Use as the second argument to remap().
- **`patternRepeatMirror(size)`** — 3D mirror repeat: like patternRepeat but reflects each cell into its positive octant (abs of the centered fold), so a motif defined in one octant is mirrored across the three cell-center planes. This is the symmetry behind the cubic beam lattices: one representative strut, mirrored, fills the cell. Use with remap().
- **`patternRepeatXY(size)`** — 2D translational repeat in the XY plane: tiles by identical copies across a vec2-ish `size` grid (size.x, size.y), with z passing through unchanged. The modulo sibling of patternRectangularXY (which mirror-tiles). Use to grid a column or other z-extruded motif. Use with remap().
- **`patternTriangularXY(spacing)`** — Triangular point lattice in the XY plane: returns a remapped p giving the offset to the nearest point of a 6-neighbor close-packed lattice with nearest-neighbor distance `spacing`. Built from two interleaved rectangular sublattices (cell [spacing, spacing*sqrt3]); z passes through. The resulting distance field is continuous across cell boundaries — use it with remap() for a radially-symmetric motif (e.g. a column) to tile it triangularly.
- **`patternHexagonalXY(spacing)`** — Hexagonal (honeycomb) point lattice in the XY plane: returns a remapped p giving the offset to the nearest vertex of a 3-neighbor honeycomb with bond length `spacing`. Built from the four-atom rectangular cell [spacing*sqrt3, spacing*3]; z passes through. Continuous across boundaries — use with remap() for a radially-symmetric motif to tile it hexagonally.
- **`triangleDistance(x, period)`** — triangleDistance(x, period) is the unit-gradient distance to the nearest of a set of parallel planes spaced `period` apart (at x = 0, ±period, ...). abs of triangleWave; range [0, period/2]. The building block for grid and honeycomb lattice fields. Shift x by period/2 to phase the planes onto cell centers instead of boundaries.
- **`triangleWaveEvenPositive(x, period)`** — triangleWaveEvenPositive(x, period): the unit-gradient triangle-wave distance whose zeros land on cell *boundaries* (x = 0, +/-period, ...) and whose peak (period/2) is at the cell center. Identical to triangleDistance — the named form documents the phase. "Even" follows the GCL stdlib convention; note this is the OPPOSITE phase from the lattice package's legacy `planesEven` (which is center-phased, i.e. the Odd family here). Range [0, period/2].
- **`triangleWaveOddPositive(x, period)`** — triangleWaveOddPositive(x, period): triangle-wave distance phased a half cell over from Even, so its zeros land on cell *centers* (x = +/-period/2, ...). Equivalent to the lattice package's legacy `planesEven`. Range [0, period/2].
- **`checker(cellSize)`** — Checkerboard parity on a 3D grid of cell size `cellSize`: +1 or -1 by whether the cell's (x+y+z) integer-coordinate sum is even or odd. Signed, so it works directly as a GSL boolean (`map.checker(2) ? colorA : colorB`) per the "<=0 is false" convention — no threshold/mask needed for a sharp 2-way board.
- **`checkerXY(cellSize)`** — Checkerboard parity in the XY plane only (z passes through unaffected) — the planar sibling of checker, for a board pattern that doesn't repeat along Z.
- **`rotateXY(v, ang)`** — Rotate 3D vector v by `ang` radians around the Z axis (XY plane rotation). v.z passes through unchanged.
- **`rotateAxis(v, axisOrigin, axisDir, angle)`** — Rotate vector `v` by `angle` radians about the line through `axisOrigin` with direction `axisDir` (Rodrigues' formula). `axisDir` need not be normalized. The arbitrary-axis counterpart to rotateXY / rotateYZ / rotateZX.
- **`remapCylindrical(circumference)`** — Wrap a flat XY pattern around a Z cylinder: returns a remapped point where x' = radial distance from the Z axis, y' = arc length around Z scaled so a full turn spans `circumference`, z' = z. Use with remap() to bend a planar field onto a cylinder.
- **`axisAt(point, axis)`** — The remap coordinate that puts a +Z-datum shape at `point` with its datum axis aligned to `axis` — for a hole, the direction the drill enters, which is the outward normal of the face it goes into. Pass it to remap:  remap(drill.hole([drill.thru(6.6)], 0), map.axisAt(faceCenter, faceNormal))  It takes a DIRECTION, where rotateAxis and matrix.rotation take an axis AND an angle. Nothing else here aligns a datum to a direction, which is the usual thing to want when placing a feature on a face whose normal you already have.  Well conditioned at every axis, -Z included. The rotation ABOUT the axis is not continuous across the equator, which is invisible for a round hole and visible for a keyway, a D-flat or a slot: orient those by their own angle, or place them on faces whose normals stay on one side.

### `transform`

#### SHAPE TRANSFORMS

- **`translate(shape, offset)`** — The shape moved by `offset`.
- **`rotateX(shape, angle)`** — The shape turned by `angle` radians about the x axis.
- **`rotateY(shape, angle)`** — The shape turned by `angle` radians about the y axis.
- **`rotateZ(shape, angle)`** — The shape turned by `angle` radians about the z axis.
- **`rotate(shape, origin, direction, angle)`** — The shape turned by `angle` radians about the line through `origin` along `direction`, which need not be normalized.
- **`scale(shape, factor)`** — The shape scaled by `factor` about the origin. The factor must be positive; the distance scales with the shape, so the field stays a distance.
- **`affine(shape, m)`** — The shape carried by the 4×4 affine matrix `m` — build it with matrix.translation, matrix.rotation, matrix.scaling and matrix.compose. The distance is scaled by m's smallest scale, which keeps it a bound when m stretches unevenly.

### `lattice`

- **`solidify(field, bias)`** — solidify(field, bias): grow a unit-normalized lattice field into a solid by offsetting its zero level set outward by `bias` (a length, or a field for graded thickness). The beam / honeycomb / foam lattices start with NO thickness — their skeleton sits at field = 0 — so a positive `bias` is what gives the struts/walls their half-width. TPMS surfaces instead start at ~50% volume (the signed field is already negative inside), so bias = 0 is already solid and `bias` merely shifts the volume fraction.
- **`solidify(field, bias, thickness)`** — solidify(field, bias, thickness): a shell/wall of the given `thickness` straddling the biased level set (|field - bias| <= thickness/2). For a thin lattice this turns the struts into hollow tubes; for a TPMS it carves a sheet of that thickness following the surface.
- **`bcc(cellSize, beamRadius, combine)`** — Body-centered-cubic beam lattice: an infinite 3D lattice of cubes of size `cellSize` with diagonal beams of the given radius connecting each cube center to its eight corners. Useful for printable space-filling support structures. (Legacy; the new GCL-derived lattice/beam.bcc supersedes this for later substitution.)
- **`bcc(cellSize, beamRadius)`** — Body-centered-cubic beam lattice with the four diagonal beams unioned sharply. Pass a blend to `bcc(cellSize, beamRadius, combine)` to fillet the junction where they meet at the cell center.

### `lattice/beam`

- **`octet(sizeIn)`** — Octet truss: the stretch-dominated FCC strut lattice (cell corners to face centers). scalar or vec3 `size` cell.
- **`truncatedOctahedral(sizeIn)`** — Truncated-octahedral (Kelvin-adjacent) strut lattice. scalar or vec3 `size` cell.
- **`bcc(sizeIn)`** — Body-centered-cubic strut lattice: the cell-center-to-corner diagonal, mirrored to all eight corners. scalar or vec3 `size` cell. (The new GCL-derived BCC; the legacy lattice.bcc remains for substitution.)
- **`fcc(sizeIn)`** — Face-centered-cubic strut lattice: cell corner to the three near face centers, mirrored. scalar or vec3 `size` cell.
- **`truncatedCubic(sizeIn, setback)`** — Truncated-cubic strut lattice. `setback` in [0, 1] pulls the strut endpoints in from the cell edges (truncating the corners). scalar or vec3 `size` cell.
- **`kelvinCell(sizeIn, setback)`** — Kelvin cell (tetrakaidecahedron) strut lattice. `setback` in [0, 1] sets the hexagon/square truncation. scalar or vec3 `size` cell.
- **`fluorite(sizeIn)`** — Fluorite (CaF2-type) strut lattice. scalar or vec3 `size` cell.
- **`isotruss(sizeIn, combine)`** — Isotruss: orthogonal axis struts (translational) plus corner-to-face diagonals (mirrored). scalar or vec3 `size` cell.
- **`isotruss(sizeIn)`** — As isotruss, with its strut families unioned sharply. Pass a blend to `isotruss(sizeIn, combine)` to fillet where they meet.
- **`reentrant(sizeIn, setback, combine)`** — Re-entrant (auxetic) strut lattice — inward-folded cell walls plus a central Z column. `setback` in [0, 1] sets the fold depth. scalar or vec3 `size` cell.
- **`reentrant(sizeIn, setback)`** — As reentrant, with its strut families unioned sharply. Pass a blend to `reentrant(sizeIn, setback, combine)` to fillet where they meet.
- **`cubic(sizeIn, combine)`** — Simple cubic lattice: three orthogonal axis struts per cell, joined by `combine`, tiled across the scalar or vec3 `size` grid.
- **`cubic(sizeIn)`** — Simple cubic lattice with the three struts joined sharply.
- **`columnRectangular(cell)`** — Grid of infinite vertical columns (z cylinders) on a rectangular XY lattice `cell` ([cellX, cellY]). The thin field is the distance to the nearest column axis; lattice.solidify turns it into cylinders. z passes through, so the columns run the full height.
- **`columnTriangular(spacing)`** — Infinite vertical columns on a triangular (6-neighbor close-packed) lattice with nearest-neighbor `spacing`.
- **`columnHexagonal(spacing)`** — Infinite vertical columns on a hexagonal (honeycomb, 3-neighbor) lattice with bond length `spacing`.
- **`cubicDouble(sizeIn)`** — Double cubic: orthogonal axis struts (translational) plus corner-to-face diagonals (mirrored) — a stiffer cubic variant. scalar or vec3 `size` cell.

### `lattice/honeycomb`

- **`squareHoneycomb(size, combine)`** — Square honeycomb in the XY plane: two orthogonal plane families spaced by the vec2 `size` (z ignored), joined by `combine`.
- **`squareHoneycomb(size)`** — Square honeycomb with the plane families joined sharply.
- **`squareDiagonalHoneycomb(size, combine)`** — Diagonal (diamond) honeycomb: two 45-degree plane families, independently spaced by size.x and size.y along each diagonal, so the rhombi tile cleanly for any aspect ratio. The diagonal coordinates are unit-speed, so the field stays a true distance.
- **`squareDiagonalHoneycomb(size)`** — Diagonal honeycomb with the two plane families joined sharply.
- **`squareHoneycombReinforced(size, rotation, combine)`** — Square honeycomb with one diagonal reinforcement beam per cell. `rotation` is the diagonal angle as a fraction of pi (0.25 = 45 degrees); `combine` joins both the grid and the diagonal.
- **`squareHoneycombReinforced(size, rotation)`** — Reinforced square honeycomb with everything joined sharply. There is no matching convenience for the two-diagonal form below: with the operator dropped it would take three arguments too, and GSL overloads on arity alone — pass `unionSharp` explicitly there.
- **`squareHoneycombReinforced(size, rotation, rotation2, combine)`** — Square honeycomb with two crossed diagonal reinforcement beams (an X in each cell). `rotation` and `rotation2` are the diagonal angles as fractions of pi; `combine` joins the grid and both diagonals.
- **`octetHoneycomb(size, combine)`** — Octet honeycomb (after Martha Baldwin, CMU): a stretch-dominated lattice combining the square grid, a half-cell-phased plane grid, and a 45-degree rotated grid. vec2 `size`, z ignored.
- **`octetHoneycomb(size)`** — Octet honeycomb with every family joined sharply.
- **`hexagonalHoneycomb(cellSize, setback, combine)`** — Hexagonal honeycomb: a three-armed Y-star at each cell, reflected to tile. `cellSize` sets the hex pitch (the cell is [cellSize, cellSize*sqrt3]); `setback` in [0, 1] positions the star branch point (default 1/3).
- **`hexagonalHoneycomb(cellSize, combine)`** — Hexagonal honeycomb with the default 1/3 setback.
- **`hexagonalHoneycomb(cellSize)`** — Hexagonal honeycomb with the default 1/3 setback and sharp junctions.
- **`triangularHoneycomb(cellSize, combine)`** — Triangular honeycomb: three 60-degree plane families, reflected to tile. `cellSize` sets the pitch (the cell is [cellSize, cellSize*sqrt3]).
- **`triangularHoneycomb(cellSize)`** — Triangular honeycomb with the three plane families joined sharply.
- **`cubicBaldwinReinforced(size, radius, biasX, biasY, biasU, biasV)`** — Cubic-Baldwin-reinforced honeycomb (after Martha Baldwin, CMU): a square cell grid plus axis-aligned and diagonal plane reinforcements, each combined with a Euclidean rounded union of fillet `radius`. The per-family biases (biasX, biasY along the axes; biasU, biasV along the diagonals) offset each plane family independently for anisotropic reinforcement — pass 0 for a uniform thin grid. vec2 `size`.

### `lattice/foam`

- **`cubic(size, combine)`** — Cubic foam: the three orthogonal wall families of a vec3 `size` cell, with junctions joined by `combine`. Boundary-phased (walls on cell boundaries).
- **`cubic(size)`** — Cubic foam with the three wall families joined sharply.
- **`cubicGrid(size, combine)`** — Cubic foam, center-phased: the same three orthogonal wall families spaced by the vec3 `size`, but with the walls on cell centers (the odd-phase sibling of cubic, via triangleWaveOddPositive). Migrated from the flat lattice package's cubicGrid; `combine` joins the plane families.
- **`cubicGrid(size)`** — Center-phased cubic foam with the three wall families joined sharply.
- **`bcc(size, combine)`** — Body-centered-cubic foam: the three sets of diagonal wall families (one per coordinate plane), each tiled by translation, then joined. vec3 `size` cell; `combine` joins the junctions.
- **`bcc(size)`** — Body-centered-cubic foam with the wall families joined sharply.
- **`fcc(size, combine)`** — Face-centered-cubic foam: the four {111}-type wall families (three corner planes plus the cell cap), mirror-folded into the cell. vec3 `size` cell; `combine` joins the junctions.
- **`fcc(size)`** — Face-centered-cubic foam with the wall families joined sharply.

### `lattice/tpms`

- **`thick(field, t)`** — thick(field, t): the solid network on the negative side of the surface (field <= t). `t` shifts the volume fraction; t = 0 is the balanced split, t > 0 grows the solid.
- **`thin(field, t)`** — thin(field, t): a sheet straddling the surface itself — a wall of total thickness ~2t following the minimal surface.
- **`inverse(field, t)`** — inverse(field, t): the complementary solid network (field >= -t) — the other labyrinth, the inverse of thick.
- **`thinInverse(field, t)`** — thinInverse(field, t): the complement of the thin sheet — the two voids either side of the wall, joined where the wall is absent.
- **`gyroid(period, drop)`** — Gyroid: sin(x)cos(y) + sin(y)cos(z) + sin(z)cos(x). The most common TPMS — fully connected, no straight channels. `period` is the vec3 cell size, `drop` per-axis anisotropy.
- **`gyroid(period)`** — Isotropic gyroid (drop = [1, 1, 1]).
- **`schwarz(period, drop, gyroidBlend)`** — Schwarz P morphed toward Gyroid. `gyroidBlend` 0 = Schwarz P (cos x + cos y + cos z), 1 = Gyroid.
- **`schwarz(period)`** — Isotropic Schwarz P (drop = [1, 1, 1], no gyroid blend).
- **`diamond(period, drop, gyroidBlend)`** — Diamond (Schwarz D) morphed toward Gyroid. `gyroidBlend` 0 = Diamond, 1 = Gyroid.
- **`diamond(period)`** — Isotropic Diamond (drop = [1, 1, 1], no gyroid blend).
- **`lidinoid(period, drop, gyroidBlend)`** — Lidinoid (Schoen) morphed toward Gyroid. `gyroidBlend` 0 = Lidinoid, 1 = Gyroid.
- **`lidinoid(period)`** — Isotropic Lidinoid (drop = [1, 1, 1], no gyroid blend).
- **`neovius(period, drop, schwarzBlend)`** — Neovius morphed toward Schwarz P. `schwarzBlend` 0 = Neovius, 1 = Schwarz P.
- **`neovius(period)`** — Isotropic Neovius (drop = [1, 1, 1], no schwarz blend).
- **`iwp(period, drop)`** — I-WP (Schoen). `drop` per-axis anisotropy.
- **`iwp(period)`** — Isotropic I-WP (drop = [1, 1, 1]).
- **`splitP(period, lidinoidBlend, gyroidOctave, schwarzOctave)`** — Split-P: a multi-octave surface mixing Lidinoid, Gyroid, and Schwarz terms. `lidinoidBlend`, `gyroidOctave`, `schwarzOctave` weight the three octave families independently.

### `lattice/spinodoid`

- **`gabor(period, theta)`** — gabor(period, theta): a stationary spinodoid of wavelength `period` with the default packet width (3) and kernel density (8). `theta` holds the cone half-angles in radians about the x, y and z axes, each 0 or between pi/12 and pi/2: [pi/2, pi/2, pi/2] is isotropic, [0, 0, pi/2] lamellar along z, [pi/6, pi/6, 0] columnar. The level set sits at the balanced interface.
- **`gabor(period, theta, levelBias)`** — gabor(period, theta, levelBias): gabor(period, theta) with the level set at quantile `levelBias` of the field, in (0, 1). As a solid network (field <= 0) that is the relative density, and the paper wants 0.3 or more for bicontinuity. A wall deposited on the level set takes its density from the cell size instead, so leave this at 0.5 there.
- **`gabor(period, theta, poles, packetWidth, impulses, seed, levelBias)`** — gabor(period, theta, poles, packetWidth, impulses, seed, levelBias): every parameter of a stationary spinodoid. `poles` lists one cone axis per angle in `theta` (e.g. the six <110> directions reach the gyroid's star). `packetWidth` is the envelope width in wavelengths (default 3); `impulses` the kernel density, in kernels per cube of the kernel's reach (default 8, hodge's; evaluation cost grows with it, and 4 shows banding); `seed` picks the realization.
- **`gaborGraded(period, theta, scale, orientation)`** — gaborGraded(period, theta, scale, orientation): a spinodoid whose wavelength and orientation distribution vary through space. `scale` is a field multiplying `period` (> 0; keep it below ~packetWidth/2 = 1.5, past which a packet holds under two waves); `orientation` is a 3x3 rotation field applied to the wave directions — a frame that follows the part, e.g. a build axis. Both are read once at every kernel centre.
- **`gaborGraded(period, theta, poles, packetWidth, impulses, pool, seed, levelBias, localNormalize, scale, orientation)`** — gaborGraded(period, theta, poles, packetWidth, impulses, pool, seed, levelBias, localNormalize, scale, orientation): every parameter. `pool` is the size of the sampled direction pool (default 4096). `localNormalize` 1 divides by the wavenumber the field carries at each point, keeping lengths real under a graded scale; 0 divides by the wavenumber of `period`.

### `mechanical/catalog`

#### Constants

- **`metric = [[ 1.0, 0.25,  0.0,  0.0,  1.2,  0.0,   0.0,  0.0,  0.0,   0.0,  0.0, 0.0,  0.2],`**
- **`pan7045 = [[ 1.6,  3.2, 1.3,  2.5, 0.0,  1.7, 0.95, 1.6, 0.90, 0.70, 0.65],`**
- **`button7380 = [[ 3.0,  5.7, 1.65,  3.5,  2.0, 1.04],`**
- **`clearance273 = [[ 1.0,  1.1,  1.3],`**
- **`shcs4762 = [[ 1.6,  0.7],`**
- **`shcsLengths = [2.5, 3.0, 4.0, 5.0, 6.0, 8.0, 10.0, 12.0, 16.0, 20.0, 25.0, 30.0,`**
- **`csk10642 = [[ 3.0,  2.0, 1.1],`**
- **`hex4017 = [[ 1.6,  3.2,  1.1],`**
- **`torx14579 = [[ 2.0,  6.0, 0.84],`**
- **`torx14583 = [[ 2.0,  6.0, 0.77],`**
- **`torxTable = [[  1.0,  0.9,  0.6],`**
- **`phillipsTable = [[0.0, 0.61, 0.31,  0.81,   0.0, 7.0,  0.25],`**
- **`phillipsGaugeTable = [[0.0, 0.64,  0.315, 0.813, 0.25,   0.0, 7.0,  0.225],`**
- **`pozidrivTable = [[0.0, 0.76, 0.48, 0.86, 7.0,  46.0,  0.13],`**
- **`pozidrivGaugeTable = [[0.0, 0.711, 0.445, 0.890, 0.0705],`**
- **`robertsonTable = [[0.0, 1.786, 1.68],`**
- **`panB1863 = [[0.1120,  5.563, 2.184, 0.254, 0.0, 1.676],`**
- **`flatB1863 = [[0.1120,  5.385, 0.0, 1.676],`**

#### Lookup

- **`onRow(key, rowKey, tol)`** — 1 if `key` is the row keyed `rowKey`, 0 if it is not.  A CATALOG SIZE IS DISCRETE: there is an M5 and an M6 and nothing between them, so a lookup hits a row or misses. `tol` is slack for floating point only — a caller who writes 25.4 / 4.0 for a quarter inch should still land on the 0.25 row — and it is NOT a blend: every key inside it gives the row's exact value, and everything outside gives nothing. Branchless, so interval evaluation is unaffected.  A lookup is a fold of this over a table, which is how every accessor below is written and how a caller with a table of its own does it:  uncPitch(nominalInch) { k = nominalInch; return reduce((acc, v) { return acc + v; }, map((r) { return r[1] * catalog.onRow(k, r[0], 0.0001); }, inchTable)); }  A size with no row sums to 0, and a zero dimension builds nothing.

#### Columns

- **`pitch(nominal)`** — The ISO 261 coarse pitch of the given nominal diameter in millimetres: catalog.pitch(6.0) is 1.0, the pitch of an M6. A diameter that is not a catalog size returns 0, and a zero pitch will not build a thread.
- **`shcsHeadDia(nominal)`** — ISO 4762 socket head cap screw head diameter dk. Zero where the standard has no screw of that size — every second-choice diameter, among others.
- **`shcsSocket(nominal)`** — ISO 4762 hex socket width across flats.
- **`shcsSocketDepth(nominal)`** — ISO 4762 socket depth t, the minimum key engagement the standard allows — the depth button and countersunk sockets are cut to as well. Zero for a size ISO 4762 does not make.
- **`clearance(nominal)`** — ISO 273 clearance hole, medium series — the hole a screw of this size passes through.
- **`clearance(nominal, fit)`** — ISO 273 clearance hole in the series `fit` names: 1 fine, 2 medium (the same number clearance(nominal) gives), 3 coarse — fastener.fitFine, fitMedium and fitCoarse. catalog.clearance(6.0, 1.0) is 6.4. Zero for any other fit, or a size ISO 273 does not list.
- **`counterbore(nominal)`** — DIN 974-1 counterbore diameter, the pocket an ISO 4762 head sits in.
- **`cskHeadDia(nominal)`** — ISO 10642 countersunk head diameter as a THEORETICAL SHARP cone: 2.24 times nominal up to M12, and less above it — M14 30.8, M16 33.6, M20 40.32, whose 90-degree cones come to Table 1's k max, 8.4, 8.8 and 10.16. Zero outside M3 to M20, which is the whole of that standard.  THE SHARP CONE IS THE HEAD AT MAXIMUM MATERIAL, and it is drawn untrimmed on purpose. ISO 10642 gives dk theoretical max and dk actual MIN — no actual max — and gauges the head against the theoretical cone; its edge is "truncated or rounded", unsized. So nothing bigger than this cone passes, and every real head is this cone with its edge taken down somewhere above dk actual min: M3 5.54, M4 7.53, M5 9.43, M6 11.34, M8 15.24, M10 19.22, M12 23.12, M16 29.01, M20 36.05.  REFERENCES: ISO 10642:2004 Table 1 and Figures 1-2, https://cdn.standards.iteh.ai/samples/34454/32dec7e84d974361b85f744935ab2175/ISO-10642-2004.pdf; ISO 10642:1997, the same values, https://cdn.standards.iteh.ai/samples/18737/b9949b6e0b564fe4b10fb4f8f7334fb9/ISO-10642-1997.pdf.
- **`nutAcrossFlats(nominal)`** — ISO 4032 hex nut width across flats.
- **`nutHeight(nominal)`** — ISO 4032 hex nut height m.
- **`washerOd(nominal)`** — ISO 7089 plain washer outside diameter.
- **`washerId(nominal)`** — ISO 7089 plain washer inside diameter.
- **`washerThickness(nominal)`** — ISO 7089 plain washer thickness.
- **`finePitch(nominal)`** — The first ISO 261 FINE pitch of the given nominal diameter in millimetres: catalog.finePitch(10.0) is 1.25, so an M10 x 1.25. Zero where this library carries no fine pitch for the size, which is not the same as the standard having none — see the column note above.  A size has one coarse pitch and several fine ones, and only the first is here. Nothing forces you through the table: every thread function takes the pitch outright, so thread.male(10.0, 1.0) is an M10 x 1 whether or not the catalog knows about it.

#### Product Standards

- **`shcsLength(minLength)`** — The shortest ISO 4762 nominal length that is at least `minLength`: catalog.shcsLength(17.0) is 20. Zero past 300, the end of the series, and a zero length builds nothing.
- **`panHeadDia(nominal)`** — ISO 7045 pan head diameter dk, the head ISO 14583 shares. Zero outside M1.6 to M10.
- **`panHeadHeight(nominal)`** — ISO 7045 pan head height k.
- **`panCrown(nominal)`** — ISO 7045 pan head crown radius rf.
- **`panRecess(nominal)`** — The ISO 7045 cross recess NUMBER for the size, types H and Z alike: 0 for M1.6 and M2, 1 for M2.5 and M3, 2 for M3.5 to M5, 3 for M6, 4 for M8 and M10. Zero is a real size here, so ask panPhillipsDia whether the recess exists.
- **`panPhillipsDia(nominal)`** — ISO 7045 type H (Phillips) recess diameter m at the face.
- **`panPhillipsDepth(nominal)`** — ISO 7045 type H gauge penetration, the maximum: how far the driver reaches below the face.
- **`panPhillipsDepthMin(nominal)`** — ISO 7045 type H gauge penetration, the minimum: the gauge has to reach at least this far below the face, and no further than panPhillipsDepth.
- **`panPozidrivDia(nominal)`** — ISO 7045 type Z (Pozidriv) recess diameter m at the face.
- **`panPozidrivDepthMin(nominal)`** — ISO 7045 type Z gauge penetration, the minimum.
- **`panPozidrivDepth(nominal)`** — ISO 7045 type Z gauge penetration, the maximum.
- **`buttonHeadDia(nominal)`** — ISO 7380-1 hexagon socket button head diameter dk. Zero outside M3 to M16.
- **`buttonHeadHeight(nominal)`** — ISO 7380-1 button head height k.
- **`buttonCrown(nominal)`** — ISO 7380-1 button crown radius rf, the middle of the range the standard gives. Carried for reference: fastener.buttonHead does not build from it. The head there is the cap through dk max and k max, the maximum-material head, and a cap is fixed by its span and height alone, so a third number could only disagree with the other two.
- **`buttonSocket(nominal)`** — ISO 7380-1 socket width across flats s.
- **`buttonSocketDepth(nominal)`** — ISO 7380-1 socket depth t, the minimum the standard allows.
- **`cskSocket(nominal)`** — ISO 10642 countersunk socket width across flats s. Zero outside M3 to M20.
- **`cskSocketDepth(nominal)`** — ISO 10642 socket depth t, the minimum key engagement the standard allows.
- **`hexHeadAcrossFlats(nominal)`** — ISO 4017 hexagon head width across flats s — the head ISO 4014 bolts carry too. Preferred threads only: M1.6 to M12, M16, M20 and M24.
- **`hexHeadHeight(nominal)`** — ISO 4017 hexagon head height k.
- **`shcsTorx(nominal)`** — The ISO 14579 hexalobular socket for a socket head cap screw of this size, as its T number: catalog.shcsTorx(6.0) is 30. Zero outside M2 to M12.
- **`shcsTorxDepth(nominal)`** — ISO 14579 hexalobular recess depth t, the maximum.
- **`panTorx(nominal)`** — The ISO 14583 hexalobular socket for a pan head screw of this size, as its T number. Zero outside M2 to M10.
- **`panTorxDepth(nominal)`** — ISO 14583 hexalobular recess depth t, the maximum.

#### Torx Drive Sizes (ISO 10664)

- **`torxMajor(tNumber)`** — The A dimension of a Torx size — the diameter ACROSS THE LOBES, which is what a T number is quoted by. catalog.torxMajor(25) is a T25.  Zero for a T number this table has no row for, the same as every other lookup here.
- **`torxMinor(tNumber)`** — The B dimension of a Torx size — the diameter across the flutes between the lobes.

#### Cross Recess Sizes (ISO 4757)

- **`phillipsCore(recessNo)`** — ISO 4757 type H (Phillips) core width b for a recess number: across the ridges of opposite ribs, where they meet the bottom — catalog.phillipsCore(2.0) is 1.47, a PH2. NOT a wing width; a PH2's wing is about 0.7 across.
- **`phillipsFacet(recessNo)`** — ISO 4757 type H facet width e: how wide the blunt alpha facet on each rib's ridge is, the middle of Table 1's range.
- **`phillipsPoint(recessNo)`** — ISO 4757 type H point diameter g: where the 26 degree 30 wing-end cone meets the recess bottom.
- **`phillipsRidge(recessNo)`** — ISO 4757 type H angle alpha across each rib's ridge, in RADIANS; zero for No. 0, whose ridge is a radius — see phillipsRoot.
- **`phillipsFlank(recessNo)`** — ISO 4757 type H ridge lean beta, in RADIANS: 7 degrees for PH0, PH1 and PH4, 5 degrees 45 minutes for PH2 and PH3. Each rib's ridge leans out by it going up, so the ribs close on a driver as it goes in — which is what makes a Phillips driver ride up out of the recess under torque.
- **`phillipsRoot(recessNo)`** — ISO 4757 type H ridge radius: No. 0's in place of the alpha facets, zero for every other number.
- **`phillipsGaugeCore(recessNo)`** — The type H GAUGE's core width b — ISO 4757 Table 2.
- **`phillipsGaugeFacet(recessNo)`** — The type H gauge's facet width e.
- **`phillipsGaugePoint(recessNo)`** — The type H gauge's point diameter g, where its wing ends meet its end.
- **`phillipsGaugeFlat(recessNo)`** — The type H gauge's flat end diameter d.
- **`phillipsGaugeRidge(recessNo)`** — The type H gauge's ridge angle alpha, in RADIANS; zero for No. 0.
- **`phillipsGaugeFlank(recessNo)`** — The type H gauge's ridge lean beta, in RADIANS.
- **`phillipsGaugeRoot(recessNo)`** — The type H gauge's ridge radius: No. 0's, zero for every other number.
- **`pozidrivWing(recessNo)`** — ISO 4757 type Z (Pozidriv) WING WIDTH f for a recess number: across a wing's parallel walls, at the knee — catalog.pozidrivWing(2.0) is 1.03.
- **`pozidrivCore(recessNo)`** — ISO 4757 type Z core width b: across opposite valleys, along the diagonal, at the knee.
- **`pozidrivKnee(recessNo)`** — ISO 4757 type Z knee diameter g: where the 26 degree 30 wing floors meet the 28 degree bottom.
- **`pozidrivLean(recessNo)`** — ISO 4757 type Z valley lean alpha, in RADIANS: how far each valley's floor leans out from the axis going up.
- **`pozidrivValley(recessNo)`** — ISO 4757 type Z delta, in RADIANS: with the fixed 46 degrees, the metal angle where a valley wall meets a wing wall.
- **`pozidrivMarkDepth(recessNo)`** — ISO 4757 type Z mark depth j: how deep the four marks along the diagonals start, at their inner ends.
- **`pozidrivGaugeCore(recessNo)`** — The type Z GAUGE's core width b — ISO 4757 Table 4.
- **`pozidrivGaugeWing(recessNo)`** — The type Z gauge's wing width f.
- **`pozidrivGaugeKnee(recessNo)`** — The type Z gauge's knee diameter g.
- **`pozidrivGaugePoint(recessNo)`** — The type Z gauge's p: from its g plane down to its point.

#### Square Recess Sizes (Robertson)

- **`robertsonAcross(sizeNo)`** — A Robertson (square) recess's width across flats at the face: catalog.robertsonAcross(2.0) is a #2, the red one.
- **`robertsonDepth(sizeNo)`** — A Robertson recess's depth, to its deepest point.

#### Inch Product Standards (ASME B18.6.3)

- **`inchPanHeadDia(nominalInch)`** — ASME B18.6.3 square recessed pan head diameter A, in millimetres, for a nominal diameter in INCHES: catalog.inchPanHeadDia(0.164) is a #8's. Zero outside #4 to 1/4".
- **`inchPanHeadHeight(nominalInch)`** — ASME B18.6.3 pan head height H, in millimetres.
- **`inchPanEdge(nominalInch)`** — The radius of an ASME B18.6.3 pan head's top edge, in millimetres.
- **`inchPanRecess(nominalInch)`** — The square recess size an ASME B18.6.3 pan head takes: 0 for a #4, 1 for a #6, 2 for a #8 and a #10, 3 for a 1/4". Zero is a size, so ask inchPanHeadDia whether the screw exists.
- **`inchPanRecessDepth(nominalInch)`** — The depth of an ASME B18.6.3 pan head's square recess, in millimetres — the regular depth, where its size has two.
- **`inchFlatHeadDia(nominalInch)`** — ASME B18.6.3 82-degree flat head diameter A, the theoretical sharp edge, in millimetres, for a nominal diameter in INCHES. Zero outside #4 to 1/4".
- **`inchFlatRecess(nominalInch)`** — The square recess size an ASME B18.6.3 flat head takes.
- **`inchFlatRecessDepth(nominalInch)`** — The depth of an ASME B18.6.3 flat head's square recess, in millimetres — the short depth for a #6 and a #8, which is what their thinner heads leave room for.

### `mechanical/gears`

#### Involutes

- **`involute(radius)`** — Involute curve of a circle of given radius, in the XY plane. Inside the base circle the field is the perimeter distance (so the curve can be revolved cleanly); outside, it's the true distance to the involute.
- **`involuteRadial(radius)`** — Variant of involute() used for gear-tooth flanks: inside the base circle the field is -p.y (radial), so unioning two reflected involutes through rotation closes off a clean tooth shape.

#### Spur Gears

- **`gearProfile(numTeeth, moduleVal, pressureAngle, rotOffset, dedendumClearance, kSmall)`** — 2D gear profile centered at the origin in the XY plane. `rotOffset` spins the profile about Z by the given angle.
- **`gear(numTeeth, moduleVal, pressureAngle, rotOffset, thickness, helixAngle, standoffDiameter, standoffHeight, standoffChamfer, boreDiameter, dedendumClearance, kSmall)`** — 3D gear: extrudes gearProfile along Z with an optional helix, adds a cylindrical standoff on the +Z face, and drills a concentric through-bore.

#### Racks

- **`rackProfile(moduleVal, pressureAngle, shift, backlash, dedendumClearance, kSmall)`** — 2D basic-rack profile in the XY plane — the straight-flank rack that meshes with gearProfile of the same module and pressure angle. A rack tooth is four planes: two flanks at the pressure angle, the tip, and the root plane its bar hangs from. Teeth point +Y with the pitch line at y = 0 (tips at y = moduleVal, roots at y = -(1 + dedendumClearance) * moduleVal), and the body fills everything below the root line, so the caller bounds the bar — see rack. `shift` slides the rack along X with a tooth centered at x = shift; pass rackShift() to stay in step with a meshing gear. `backlash` thins every tooth by that much in total, giving the mesh a running clearance (0 is the theoretical rack, whose flanks touch the gear's). `dedendumClearance` and `kSmall` mean what they do for gearProfile: extra root depth with a fillet of that radius, and a small chamfer breaking the tip corners.
- **`rackShift(numTeeth, moduleVal, travel)`** — The rack `shift` that holds a rack in mesh with a gear of the given tooth count rolling along it, while the rack travels by `travel` and the gear turns by travel / pitchRadius. The gear's pitch circle unrolls onto the rack's pitch line, so the phase constant is a quarter turn of that arc. Verified by measurement rather than by eye — see TestRackAndPinionMesh, which reads the mesh clearance across a full tooth pitch of travel.
- **`rack(moduleVal, pressureAngle, shift, backlash, length, depth, thickness, dedendumClearance, kSmall)`** — 3D rack: extrudes rackProfile along Z, bounded to a length along X, a depth below the pitch line, and a thickness in Z (centered on the origin except depth, which hangs below the pitch line y = 0).
- **`rackAndPinion(numTeeth, moduleVal, pressureAngle, travel, backlash, dedendumClearance, kSmall)`** — A pinion and the rack it drives, in mesh, returned as a multi-value (pinion, rack) of 2D profiles. The pinion is centered at the origin and the rack's pitch line lies at y = -pitchRadius with its teeth pointing up, so the pair rolls without slipping: `travel` slides the rack along X and turns the pinion by travel / pitchRadius, keeping the mesh in step at every travel. Give `backlash` a few hundredths of the module for a running fit. The involute undercut limit still applies: below about 2/sin(pressureAngle)^2 teeth (17 at 20 degrees) a pure-involute pinion and a full-depth rack foul near the root, exactly as real gears do — use more teeth, or relieve the rack tip.
- **`gearPair(numTeeth1, numTeeth2, moduleVal, pressureAngle, thickness, helixAngle, standoffDiam1, standoffDiam2, standoffHeight, standoffChamfer, boreDiam1, boreDiam2, dedendumClearance, kSmall, phase)`** — Two meshing gears placed along X so their pitch circles are tangent, returned as a multi-value (g1, g2). `phase` advances both gears in lock-step (gear 1 by pitchAng1 * phase * 2; gear 2 by pitchAng2 * phase * 2, in the opposite direction) — pass e.g. mod(time, 1) to animate.

#### Bevel Gears

- **`bevelPitchAngle(numTeeth, mateTeeth, shaftAngle)`** — The pitch cone half-angle of a bevel gear of `numTeeth` running against a mate of `mateTeeth`, on shafts crossing at `shaftAngle`. The two angles of a pair sum to the shaft angle, and the pair shares one apex and one outer cone distance, so this is all bevelPair needs to place both gears. For the usual right-angle drive it reduces to atan(numTeeth / mateTeeth).
- **`bevelTeeth(numTeeth, moduleVal, pressureAngle, pitchAngle, spiralAngle, rotOffset, backlash, dedendumClearance, kSmall)`** — The teeth of a bevel gear as an unbounded conical field: apex at the origin, axis +Z, teeth converging on the apex, and no ends — bevelGear faces it off, or bound it yourself. `pitchAngle` is the pitch cone half-angle (bevelPitchAngle gives the one that meshes a chosen mate) and `moduleVal` is the module at the OUTER end; together they fix the outer cone distance at numTeeth * moduleVal / (2 * sin(pitchAngle)), so a bevel gear's size is not a free parameter the way a spur gear's thickness is. `spiralAngle` is the angle the tooth trace makes with the cone element — 0 is a straight bevel, and it is the bevel analog of gear()'s helixAngle, over-estimating the field on the flanks by the same 1 / cos of itself. `rotOffset` spins the gear about its axis and accepts any value. `backlash` thins every tooth by that much at the outer end, and in proportion at every other cone distance. `dedendumClearance` and `kSmall` mean what they do for gearProfile. `numTeeth` must be a whole number — the tooth pattern closes around the axis only then.
- **`bevelGear(numTeeth, moduleVal, pressureAngle, pitchAngle, spiralAngle, rotOffset, faceWidth, crownToBack, recess, backlash, standoffDiameter, standoffHeight, standoffChamfer, boreDiameter, dedendumClearance, kSmall)`** — 3D bevel gear: bevelTeeth ended the way a real bevel blank is, and drilled with a concentric bore. The teeth run out on the two cones normal to the pitch cone — the back cone at the heel and the front cone at the toe — and each of those is then cut by a plane perpendicular to the axis, giving the flat back face and the flat toe face the tooth ends step down from. `faceWidth` is measured along the cone element inward from the outer end; gear practice keeps it under about a third of the outer cone distance.  `crownToBack` and `recess` place those planes: each is the axial distance from its end's outer tooth-tip corner — where the tip cone meets that end cone, on the outer cylinder — to the flat face, both taken in +Z. At the heel that stands the back face proud of the teeth; at the toe it sets the toe face back so the teeth project past it toward the mate. Either way the tooth end is left running out on its cone in between, and the tip corner itself does not move: the recess only says where the flat face meets that cone. Zero is the degenerate case and is meant to work — the plane passes through the tip corner, its cone binds nowhere there is material, and the end comes out flat.  The standoff is the hub, and it goes on the heel where a bevel gear's hub belongs — a cylinder of `standoffDiameter` rising from the toe face, so it adds nothing inside the blank and projects `standoffHeight` past the back face, blended on with `standoffChamfer`. Measuring from the back face is the difference from gear(), which has a mid-plane to measure from and a bevel gear has not. To ask for no hub at all set the height AND the chamfer to zero: a chamfer union of the hub's top with the back face it is then flush against blends the two, standing the face proud by the usual chamfer / sqrt(2) over the hub's footprint. Any real hub clears that, and the back face outside the footprint is untouched either way.  The apex sits at the origin and the teeth face it, so a mate goes on the far side of the origin, not the far side of the gear.
- **`bevelTooth(moduleVal, pressureAngle, shaftAngle, spiralAngle, faceWidth, crownToBack, recess, backlash, standoffHeight, standoffChamfer, dedendumClearance, kSmall)`** — A bevel tooth form shared by every gear that meshes with it, as a factory: call it once with what mating gears must agree on, and the function it returns makes a gear from what is its own — `numTeeth`, its mate's `mateTeeth`, a `spiralHand` of +1 or -1, its `rotOffset`, and its hub `standoffDiameter` and `boreDiameter`. The pitch angle is derived from the two tooth counts and `shaftAngle` rather than passed, so a gear whose cone disagrees with its mate's stops being expressible — an error no single-pose mesh check would catch. A pair needs one hand of each. Every argument means what it does for bevelGear, and a gear made this way is identical to what bevelPair builds; use it for what bevelPair is not — a train, or two gears that need separate placement.  tooth = gears.bevelTooth(0.5, 0.35, shaftAngle, 0.5, faceWidth, 0.4, 0.4, 0.025, 1.2, 0.1, 0.25, 0.025); pinion = tooth(16, 24,  1, rot1, 3.0, 1.5); wheel  = tooth(24, 16, -1, rot2, 3.6, 2.0);
- **`bevelPair(numTeeth1, numTeeth2, moduleVal, pressureAngle, shaftAngle, spiralAngle, faceWidth, crownToBack, recess, backlash, standoffDiam1, standoffDiam2, standoffHeight, standoffChamfer, boreDiam1, boreDiam2, dedendumClearance, kSmall, phase)`** — Two bevel gears in mesh on shafts crossing at `shaftAngle`, returned as a multi-value (g1, g2). Gear 1 turns about +Z, gear 2's axis lies in the XZ plane at `shaftAngle` from it, both apexes are the origin, and the two share one module and one outer cone distance. `phase` advances the pair by one tooth per unit — gear 1 forward, gear 2 back by the tooth ratio — so mod(time, 1) animates the drive. `spiralAngle` is gear 1's hand; gear 2 is given the opposite hand automatically, which is not a nicety but the meshing condition: a same-hand pair collides rather than meshes. Give `backlash` a few hundredths of the module for a running fit. `crownToBack` and `recess` end both blanks and the standoff arguments hub them, meaning what they do for bevelGear — each gear takes its own hub diameter, and both share a height and a chamfer, as gearPair does. Equal tooth counts at a right angle make a miter pair. For gears that are not simply a pair — a train, or two that need separate placement — bevelTooth declares what they share once and makes each gear from its own tooth count and its mate's; the gears it makes are identical to these.

### `mechanical/fastener`

#### Constants

- **`threadIso = 1.0`** — The ISO 68-1 thread form, for a finish.
- **`threadSin = 2.0`** — A cosmetic sinusoidal thread form, for a finish.
- **`fitFine = 1.0`** — ISO 273 fine series: the tightest clearance, H12.
- **`fitMedium = 2.0`** — ISO 273 medium series, the general-purpose hole, H13.
- **`fitCoarse = 3.0`** — ISO 273 coarse series, H14.

#### Heads

- **`cylindricalHead(dk, k)`** — A CYLINDRICAL head `dk` across and `k` tall with a chamfered top edge — the socket head cap screw's head, ISO 4762's shape.
- **`cylindricalHead(dk, k, edge)`** — A cylindrical head with the top edge chamfer leg given.
- **`countersunkHead(dk, dSmall, included)`** — A COUNTERSUNK head: a cone `dk` across at the seating plane narrowing to `dSmall` where it meets the shank, at the given `included` angle in radians. It sits BELOW z = 0, since a flat head seats flush.  The depth follows from the angle rather than being given: a 90-degree head loses a millimetre of radius per millimetre of depth, which is why ISO 10642's head depth is exactly (dk - d)/2.
- **`countersunkHead(dk, dSmall)`** — A 90-degree countersunk head, the angle every ISO flat head uses.
- **`buttonHead(dk, k)`** — A BUTTON head: the spherical cap `dk` across and `k` tall, ISO 7380's shape. The crown radius follows from the two, since a cap is determined by its span and its height.
- **`panHead(dk, k, crownR)`** — A PAN head: a cylinder `dk` across and `k` tall whose top is domed to `crownR`. ISO 7045's shape, and the everyday cross-recessed machine screw.
- **`panHead(dk, k)`** — A pan head with the crown radius at 0.9 dk, near the usual proportion.
- **`hexHead(acrossFlats, k)`** — A HEX head `acrossFlats` wide and `k` tall, corners chamfered on the top face the way a bolt head or a nut is. ISO 4014 and 4017's shape.
- **`squareHead(acrossFlats, k)`** — A SQUARE head `acrossFlats` wide and `k` tall, corners chamfered on the top face. The oldest machine-made bolt head there is, and still what a set screw on a shaft collar, a plough bolt or a track bolt carries: four flats take a bigger wrench flat than six do at the same size, and an open-end wrench only has to be re-seated every quarter turn.
- **`pentHead(acrossFlats, k)`** — A PENTAGON head `acrossFlats` wide (centre to flat, doubled) and `k` tall.  Five sides is a SECURITY head: no standard wrench has five flats, so the fastener needs its own tool, which is why it holds fire hydrants, water meters, utility cabinets and electrical enclosures shut. Odd-sided also means no pair of opposite faces for an adjustable wrench or a pair of pliers to grip, which is the real reason it works and the reason it is measured centre to flat.
- **`flangeHead(dk, k, dc, c)`** — A FLANGED head: the head `dk` across and `k` tall standing on an integral flange `dc` across and `c` thick, its underside chamfered. The flange spreads the load the way a washer does, without the washer.
- **`noHead()`** — NO head at all — the set screw and the headless shoulder screw. It is the empty solid, so screw() unions nothing and the shank stands alone.

#### Head Marks

- **`cylindricalHead(dk, k, markSection, markDepth)`** — A cylindrical head carrying an identification mark — see markHead.
- **`buttonHead(dk, k, markSection, markDepth)`** — A button head carrying an identification mark — see markHead.
- **`panHead(dk, k, markSection, markDepth)`** — A pan head carrying an identification mark — the Pozidriv ticks, say. `markSection` is a 2D section from drill, engraved `markDepth` into the crown and following it, so it reads at the rim of the dome as well as at the middle.  fastener.panHead(12.0, 4.4, drill.dimpleSection(0.9, 4.4), 0.25)
- **`panHead(dk, k, crownR, markSection, markDepth)`** — A pan head with its crown radius given and an identification mark — the ISO 7045 head, catalog.panCrown, carrying Pozidriv ticks or any other mark.

#### Shanks

- **`plainShank(dia, shankLength, zTop)`** — A PLAIN shank `dia` across hanging `shankLength` below the seating plane, its tip chamfered, running up to `zTop` — inside the head it goes under. The form to prove a fit against, whatever thread the screw is drawn with.
- **`plainShank(dia, shankLength)`** — A plain shank whose top is the seating plane: a headless body's.
- **`threadedShank(dia, pitchVal, shankLength, zTop)`** — A shank threaded its whole length: the ISO 68-1 external form at `pitchVal`, crest at `dia`, running up to `zTop`. DISPLAY ONLY — prove against plainShank, which is its envelope, or an interval proof subdivides to thread scale to tell you what the plain cylinder already did.
- **`threadedShank(dia, pitchVal, shankLength)`** — A threaded shank whose top is the seating plane.
- **`sinShank(dia, pitchVal, shankLength, zTop)`** — A shank threaded with a SINUSOID at the standard depth rather than the standard V: the right envelope, the wrong flanks, and no corners in the field. The cheap display form, running up to `zTop`.
- **`sinShank(dia, pitchVal, shankLength)`** — A sinusoid shank whose top is the seating plane.
- **`partialShank(dia, pitchVal, shankLength, threadLength, zTop)`** — A PARTIALLY threaded shank: plain under the head, threaded for the last `threadLength` of its `shankLength`. What a bolt is, as opposed to a screw — the plain run is what takes shear in a fitted joint.  The thread runs up into the plain run, by a quarter of the diameter or half the plain run, whichever is less, so the two do not meet on one plane.
- **`partialShank(dia, pitchVal, shankLength, threadLength)`** — A partially threaded shank whose top is the seating plane.
- **`noShank()`** — NO shank at all — the mirror of noHead(), and the head on its own.  It is what you want when the head and the shank have to be two OBJECTS rather than one: a colour belongs to a named object, so a screw whose shank is painted and whose head is not has to be built as two and min-ed, and this is the half of that pair that carries the head.
- **`shoulderShank(shoulderDia, shoulderLength, threadDia, pitchVal, threadLength, zTop)`** — A SHOULDER shank: a ground shoulder `shoulderDia` across and `shoulderLength` long, then a smaller thread. The shoulder is the bearing surface — a shoulder screw is a dowel and a screw in one part, so the shoulder is what is toleranced and the thread merely holds it there.  The stud runs up into the shoulder by a quarter of its diameter or half the shoulder, whichever is less, so the two do not meet on one plane.
- **`shoulderShank(shoulderDia, shoulderLength, threadDia, pitchVal, threadLength)`** — A shoulder shank whose top is the seating plane.

#### Ends

- **`ended(shank, cut, add)`** — `shank` finished with an end: `cut` taken out of it, `add` unioned onto it. Pass nothing for whichever half the end does not use.  fastener.screw(fastener.noHead(), fastener.ended(fastener.threadedShank(6.0, 1.0, 10.0), fastener.cupEnd(6.0, -10.0, 1.2, radians(118.0)), nothing), drill.hexSocketAt(0.0, 3.0, 5.0))
- **`flatEnd(dia, zEnd, leg)`** — A FLAT end at `zEnd` with its edge broken by `leg` at 45 degrees. The plain end, and the one every other end here is a modification of. A void.
- **`flatTop(dia, zTop, leg)`** — The same break on the plane `zTop` with the material BELOW it — the chamfer on the socket end of a set screw, which is not decoration: the screw is driven down inside a threaded hole, and a sharp corner there catches on the first thread and raises a burr that keeps it from seating.
- **`cupEnd(dia, zEnd, depth, included)`** — A CUP end: a cone ground into the end face to `depth`, leaving a thin annular rim. The commonest set screw point and the best holder on a hard shaft, because the rim bites rather than bearing flat. Pair it with a flatEnd to break the rim's outside edge. A void.
- **`coneEnd(dia, zEnd, included)`** — A CONE end coming to a point at `zEnd`, `included` radians across. It digs into the shaft and will not come out sideways, so it is the one for a permanent setting — and for a spotted shaft, where the cone matches a drilled detent. A void: the shank intersected with the cone.
- **`ovalEnd(dia, zEnd, radius)`** — An OVAL end: the tip radiused to `radius`, which is larger than the shank radius, so it bears on a curved face without the bite of a cone. The point for a shaft that has to be adjusted and re-set without chewing it up. A void.  IT HAS TO BE BOUNDED ABOVE, and a cone does not. A cone opening upward from the tip is wider than the shank everywhere above the point, so intersecting with it takes nothing there — but a SPHERE closes again, and an unbounded version of this end quietly deletes the whole screw above the ball's equator. So the void stops at zJoin, where the sphere crosses the shank's own diameter, and above that the shank is left alone.
- **`ballEnd(dia, zEnd)`** — A BALL end: the oval end at exactly the shank radius, so the tip is a true hemisphere. Bears on a single point, which is what a ball-end set screw in a V-groove or against a flat is for.
- **`dogPoint(pointDia, zEnd, pointLength)`** — A DOG POINT — an EXTENDED point: an unthreaded cylinder `pointDia` across standing `pointLength` below `zEnd`. It drops into a hole or a slot in the shaft, so the screw locates as well as clamps, and it does not mark the bearing surface the way a cone does. A SOLID, so it goes in ended()'s `add`.  A half dog point is the same thing shorter. The extended point in a softer material — brass or nylon, so it clamps without marking — is the same solid too; make it a separate named object if it wants its own colour.
- **`softTip(dia, zEnd, tipLength)`** — A SOFT TIP: a nylon or brass pad `tipLength` long on the end of the screw, clamping without marking the shaft and locking against vibration.  `dia` IS THE THREAD'S MINOR DIAMETER, not its nominal. The pad sits in a pocket bored into the end of the screw, and that pocket cannot be wider than the root of the thread without cutting through it — a pad at nominal would stand out past the crests and be sheared off going into the hole. Pass thread.minorDia(major, pitch).  A SOLID, and one you should keep as its own named object rather than folding into the screw, because the whole point of it is that it is a different material — and a colour belongs to a named object.
- **`selfDrillPoint(dia, zEnd, pointLength, included, fluteWidth)`** — A SELF-DRILLING point: an UNTHREADED drill standing `pointLength` below `zEnd`, at `dia`, ending in a conical point of `included` angle, with two flutes running from the tip up to where the thread starts.  A SOLID — it stands proud of the screw — so it goes in ended()'s `add`.  `dia` IS THE THREAD'S MINOR DIAMETER. That is the whole mechanism: the point drills a hole the size of the thread's ROOT, and the thread then cuts its own crests into the wall of that hole. A drill at nominal would bore away the material the thread needs to bite and the screw would spin in its own hole. Pass thread.minorDia(major, pitch).  THE FLUTES STOP AT THE THREAD. They have to: chips have to clear the cutting lips and get out of the hole, and past the first thread there is nothing left to cut, so a flute run up into the thread would only be a notch out of it.  The flutes are straight here where a real point's are helical. That is enough to draw the screw and to see which way it cuts; it is not a tool geometry, and nothing here should be used to grind one.

#### Composition

- **`screw(head, shank, drive)`** — A screw: the `head` and `shank` solids unioned, with the `drive` recess taken out. Each is an ordinary field, so any head goes with any shank and any drive — drill.hexSocketAt, drill.crossRecessAt, drill.squareSocketAt and drill.slotDriveAt all cut one.  fastener.screw(fastener.panHead(9.5, 3.3), fastener.threadedShank(5.0, 0.8, 16.0), drill.crossRecessAt(3.3, 4.4, 1.2, 2.2, radians(7.0)))
- **`screw(head, shank)`** — A screw with no drive recess — an external drive turns on the head itself, so a hex head needs nothing cut into it.

#### Multi-Material Set Screws

- **`setScrewSoftTip(nominalDia, screwLength, socketAf, socketDepth, tipLength)`** — A SET SCREW WITH A SOFT TIP, as two objects: the steel body, then the nylon or brass pad on the end of it. Multireturn — see the note above.  The pad is sized on the thread's MINOR diameter, because it sits in a pocket bored into the end and that pocket cannot be wider than the thread root without cutting through the crests.
- **`setScrewSoftPoint(nominalDia, screwLength, socketAf, socketDepth,pointDia, pointLength)`** — A SET SCREW WITH AN EXTENDED POINT in a second material — the brass or nylon dog point that locates and clamps without marking the shaft. Two objects: the steel body, then the point.
- **`setScrewDrillPoint(nominalDia, screwLength, socketAf, socketDepth,pointLength, included, fluteWidth)`** — A SELF-DRILLING SCREW as two objects: the threaded body, then the drill point. Two because the point is usually a harder or case-hardened material than the body — it has to cut the hole the body then threads.

#### Standard Screws

- **`shcs(nominalDia, shankLength)`** — ISO 4762 socket head cap screw at catalog nominal: cylindrical head (diameter and hex socket from the size table, height = nominal) above z = 0, plain shank of the given length below, 45-degree tip chamfer. fastener.shcs(3.0, 8.0) is an M3 x 8. Seating plane at z = 0; place with remap(). Prove clearances against this form; display shcsThreaded.
- **`shcsThreaded(nominalDia, shankLength)`** — ISO 4762 socket head cap screw with a display-only helical thread on the shank: the real ISO 68-1 external form at the coarse pitch, crest at nominal. Same dimensions and datum as shcs — use it in display returns; prove against shcs.
- **`shcsThreadedSin(nominalDia, shankLength)`** — ISO 4762 socket head cap screw whose shank thread is a SINUSOID rather than the standard V — the cheap display form, with no corners in its field. Same dimensions and datum as shcs; prove against shcs.
- **`countersunk(nominalDia, overallLength)`** — ISO 10642 countersunk (flat head) screw: 90-degree conical head flush with the surface at z = 0, ISO 10642's own hex socket (a size smaller than shcs's, cut to its minimum key engagement), shank to the given overall length below. Mates with fastener.countersunkHole. Prove against this form; display countersunkThreaded.
- **`countersunkThreaded(nominalDia, overallLength)`** — ISO 10642 countersunk screw with a display-only helical thread on the shank. Same dimensions and datum as countersunk.

#### More Standard Screws

- **`panPhillipsReference(nominalDia)`** — The z of the REFERENCE PLANE of an ISO 7045 type H recess in a pan head of this size: where the recess meets the top of the head, below the crown's apex, and what ISO 4757 measures penetration from.
- **`panPhillipsFloor(nominalDia)`** — The z of the deepest point ISO 7045 lets a type H gauge reach in a pan head of this size: the maximum penetration below the reference plane. The recess itself runs on below it to the point of its 28 degree bottom.
- **`panPhillipsSeat(nominalDia)`** — The z to SEAT a Phillips driver's point at in a pan head of this size: ISO 7045's minimum gauge penetration below the reference plane. The gauge is proven clear of the recess there and touches before the maximum, so a driver.phillipsTip put here is in the recess, a few hundredths off its ribs.
- **`panTorxFloor(nominalDia)`** — The z of the floor of the ISO 14583 hexalobular recess in a pan head of this size — where a Torx driver's end seats.
- **`panPozidrivFloor(nominalDia)`** — The z of the floor of an ISO 7045 type Z recess in a pan head of this size.
- **`panPhillipsDrive(nominalDia)`** — The ISO 7045 type H (PHILLIPS) recess of a pan head of this size, as the void panPhillips cuts: seated on the plane ISO 4757 measures it from, m across there, in the ISO 4757 form of its recess number — see drill.phillipsAt. For a head that carries it with something else — a mark, a different shank.
- **`panPhillipsGauge(nominalDia, pen)`** — The ISO 4757 type H GAUGE for a pan head of this size, seated `pen` below its reference plane and standing 10 above it: the point of the screwdriver the recess is checked with, as a solid.
- **`panPozidrivReference(nominalDia)`** — The z of the REFERENCE PLANE of an ISO 7045 type Z recess in a pan head of this size, where penetration is measured from.
- **`panPozidrivSeat(nominalDia)`** — The z to SEAT a Pozidriv driver's point at in a pan head of this size: ISO 7045's minimum type Z penetration below the reference plane, where the gauge is proven clear and short of touching.
- **`panPozidrivDrive(nominalDia)`** — The ISO 7045 type Z (POZIDRIV) recess of a pan head of this size, as the void panPozidriv cuts: seated on the reference plane, m across there, in the ISO 4757 form of its recess number — see drill.pozidrivAt.
- **`panPozidrivGauge(nominalDia, pen)`** — The ISO 4757 type Z GAUGE for a pan head of this size, its point `pen` below the reference plane and standing 10 above it.
- **`panPhillips(nominalDia, shankLength)`** — ISO 7045 cross recessed PAN HEAD screw, type H — PHILLIPS — at catalog nominal, M1.6 to M10: fastener.panPhillips(4.0, 10.0) is an M4 x 10 with a PH2. Head, recess number, recess diameter and depth all from the standard; the recess is panPhillipsDrive.
- **`panPozidriv(nominalDia, shankLength)`** — ISO 7045 cross recessed pan head screw, type Z — POZIDRIV. The same head as panPhillips; the recess is panPozidrivDrive.
- **`panTorx(nominalDia, shankLength)`** — ISO 14583 hexalobular socket PAN HEAD screw, M2 to M10: ISO 7045's head with the TORX socket ISO 14583 pairs with each size.
- **`button(nominalDia, shankLength)`** — ISO 7380-1 hexagon socket BUTTON HEAD screw, M3 to M16. The crown follows from the head's span and height, and lands inside the radius range the standard gives for it.
- **`shcsTorx(nominalDia, shankLength)`** — ISO 14579 hexalobular socket head cap screw: ISO 4762's head with the TORX socket the standard pairs with each size, M2 to M12.
- **`hexScrew(nominalDia, shankLength)`** — ISO 4017 HEXAGON HEAD SCREW, threaded to the head: M1.6 to M12, M16, M20 and M24. An external drive, so nothing is cut into the head.
- **`hexBolt(nominalDia, shankLength)`** — ISO 4014 HEXAGON HEAD BOLT: the hex screw's head on a shank threaded only its last hexBoltThread(nominalDia, shankLength).  THE SMOOTH FORM IS THE SCREW'S. A bolt and a screw are the same solid at the major diameter; what differs is how much of the shank is threaded, and that is the part displayed rather than proved — pass the thread length to thread.maleRelief as its bound, or to partialShank.
- **`hexBoltThread(nominalDia, shankLength)`** — The threaded length b of an ISO 4014 bolt: 2d + 6 up to 125 long, 2d + 12 over 125 to 200, and 2d + 25 over 200, or the whole shank if that is shorter. The rule reproduces every b ISO 4014 tabulates but one: an M1.6 is 9, not 9.2. The standard leaves b blank where no bolt of that length is made — b1 from M36, b2 at M56 and M64 — and this still answers there.  REFERENCES: ISO 4014:2022 Tables 1-4 (M1.6 to M27) and :2011 Table 1 (M12 to M64), the boundaries from their footnotes, https://cdn.standards.iteh.ai/samples/72579/81f2b1e966b440d7a25f6734e508ea48/ISO-4014-2022.pdf https://cdn.standards.iteh.ai/samples/56447/63b5e00f45d043e6a27a45ae955b88f9/ISO-4014-2011.pdf

#### Inch Square Drive Screws

- **`roundedHead(dk, k, edge)`** — A PAN head the way ASME B18.6.3 dimensions it: a cylinder `dk` across and `k` tall whose top edge is rounded to `edge`, flat on top — the standard gives the edge and not the crown. Exact: the offset, by `edge`, of the corner the flat top and the side make.
- **`panRobertson(nominalInch, shankLength)`** — ASME B18.6.3 type III square recessed PAN HEAD machine screw, #4 to 1/4", nominal diameter in INCHES and shank length in millimetres: fastener.panRobertson(0.164, 12.7) is a #8 x 1/2" with a #2 recess. Seating plane at z = 0, shank down, like every screw here.
- **`flatRobertson(nominalInch, overallLength)`** — ASME B18.6.3 type III square recessed 82-degree FLAT HEAD machine screw, #4 to 1/4", nominal diameter in INCHES and overall length in millimetres — a flat head's length is overall, from the top face. The head sits BELOW z = 0 and seats flush, as countersunk() does.

#### Placement

- **`seatedAt(seatPoint, shankDir)`** — The remap coordinate that seats a +Z-datum fastener at `seatPoint` with its shank running along `shankDir`, any direction: vector.dirZ drives the screw upward, its negation keeps the library's own downward datum, and a face normal's negation drives it square into that face. Pass it to remap: `remap(fastener.shcs(3.0, 18.0), fastener.seatedAt([8, 26.75, -4.5], vector.dirZ))`.  It is map.axisAt with the axis reversed — the datum's +Z is the way the head faces, opposite the shank — so a screw seated here lands in a hole placed by drill.axisAt at the same point with the face's normal. Straight up it is a half turn about x, never a mirror. As with axisAt, the turn ABOUT the shank is not continuous across the equator; a round screw does not show it, a hex head does.
- **`shcsSeated(nominalDia, shankLength, seatPoint, shankDir)`** — An ISO 4762 socket head cap screw already placed: its head seated at `seatPoint` with the shank running along `shankDir`. Identical geometry to shcs(), so prove against it the same way; shcsSeatedThreaded is the display twin. Sizing rule for a through-bolted joint: the shank must cross every part it passes through AND engage at least three pitches of thread in the part it holds — a catalog screw that stops short of that part fastens nothing.
- **`shcsSeatedThreaded(nominalDia, shankLength, seatPoint, shankDir)`** — shcsSeated with the display-only helical thread — same dimensions and placement as shcsSeated; use it in display returns, prove against shcsSeated.
- **`countersunkSeated(nominalDia, overallLength, seatPoint, shankDir)`** — An ISO 10642 countersunk screw already placed: its head flush with the surface at `seatPoint`, shank running along `shankDir`. Mates with fastener.countersunkHole placed with the same seatedAt.

#### Nuts, Washers and Pins

- **`nut(nominalDia)`** — ISO 4032 hex nut at catalog nominal: hex prism (width across flats and height from the size table) resting on z = 0, corner chamfers on both faces, and the bore at the coarse thread's basic MINOR diameter — the nut's maximum material condition, which the screw through it interferes with by design. Draw the thread with thread.femaleRelief(d, P, m, 0) and prove the fit with thread.engaged. Place with remap().
- **`nutThreaded(nominalDia)`** — ISO 4032 hex nut with a display-only helical thread: nut() with the ISO 68-1 internal form cut into its bore, the counterpart of shcsThreaded. Prove against nut().
- **`nutThreadedSin(nominalDia)`** — The hex nut with a SINUSOIDAL thread cut into its bore, the counterpart of shcsThreadedSin.
- **`washer(nominalDia)`** — ISO 7089 plain washer at catalog nominal, resting on z = 0. Place with remap().
- **`dowel(diameter, pinLength)`** — ISO 2338 dowel pin: a plain cylinder of the given diameter and length, centered at the origin along Z, with chamfered ends. Dowels are sized by diameter directly (any value is legal, not table-bound).

#### Finishes

- **`maleFinish(finish, major, pitchVal, zTop, zBottom, form)`** — The material that draws an EXTERNAL thread of `form` on `finish`: the thread between the planes z = `zTop` and z = `zBottom` as a displacement height on a shank at the `major` diameter.
- **`maleColor(major, pitchVal, zTop, zBottom, form)`** — The colour that paints maleFinish's thread on steel where it is not drawn.
- **`femaleFinish(finish, major, pitchVal, zTop, zBottom, form)`** — The material that draws an INTERNAL thread of `form` on `finish`, on a bore at the basic minor diameter between the planes z = `zTop` and z = `zBottom`.
- **`femaleColor(major, pitchVal, zTop, zBottom, form)`** — The colour that paints femaleFinish's thread on steel where it is not drawn.
- **`shcs(nominalDia, shankLength, finish, form)`** — ISO 4762 socket head cap screw, finished: shcs with `finish` and its thread of `form` on the whole shank.
- **`shcs(nominalDia, shankLength, finish, form, headFinish)`** — The finished cap screw with its head in `headFinish`.
- **`countersunk(nominalDia, overallLength, finish, form)`** — ISO 10642 countersunk screw, finished: the thread starts under the cone, (dk - d) / 2 below the face, and runs to the tip.
- **`countersunk(nominalDia, overallLength, finish, form, headFinish)`** — The finished countersunk screw with its head in `headFinish`.
- **`hexScrew(nominalDia, shankLength, finish, form)`** — ISO 4017 hexagon head screw, finished, threaded its whole length.
- **`hexScrew(nominalDia, shankLength, finish, form, headFinish)`** — The finished hexagon head screw with its head in `headFinish`.
- **`nut(nominalDia, finish, form)`** — ISO 4032 hex nut, finished: tapped face to face at the basic minor, the thread of `form` on `finish`.

#### Holes

- **`clearanceHole(nominalDia, fit)`** — The hole a screw of this size passes through, at ISO 273 `fit`: a thru at the clearance diameter, for a part the shank only crosses or a head sits on. A void in the screw's frame — remap it with the screw's own seatedAt and subtract it.
- **`clearanceHole(nominalDia, fit, depth)`** — The clearance stopped `depth` below the seat, flat-bottomed and open above: for a part the screw clears for that far and then threads into, the tapped part and the clamped one being the same solid. Union tappedHole under it for the thread's own hole.
- **`shcsHole(nominalDia, fit)`** — The hole an ISO 4762 cap screw's head sits down in: the DIN 974-1 counterbore, its floor on the seating plane and open above, over the ISO 273 clearance at `fit`. For shcs, shcsTorx and their threaded forms.
- **`countersunkHole(nominalDia, fit)`** — The hole an ISO 10642 countersunk screw seats flush in: a 90-degree countersink whose rim is the head's theoretical sharp diameter on the seating plane — the maximum-material head, so the two cones coincide — over the ISO 273 clearance at `fit`.
- **`tappedHole(nominalDia, shankLength)`** — The hole the thread of a screw `shankLength` long holds in: drilled at the basic MINOR diameter — the tapped hole at maximum material, the same bore thread.female cuts its flanks from — to three pitches past the screw's tip, with a 118-degree drill point. Three pitches is the shop allowance for the drill's point and the tap's lead, not a standard's number. Display the thread on the part with thread.femaleRelief.

### `mechanical/thread`

#### Constants

- **`inchTable = [[0.1120, 0.63500, 0.52917],`**
- **`deviation965 = [[0.2,  0.017,  0.0,    0.0],`**

#### Standard Sizes

- **`metricPitch(nominal)`** — The coarse pitch of an ISO metric thread of the given nominal diameter in millimetres: thread.metricPitch(6.0) is 1.0, the pitch of an M6. A diameter that is not a catalog size returns 0 — threads exist only in standard sizes, and a zero pitch will not build a thread.  ISO 261 first and second choice, M1 through M64, from mechanical/catalog.
- **`metricFinePitch(nominal)`** — The first ISO 261 FINE pitch of an ISO metric thread of the given nominal diameter in millimetres: thread.metricFinePitch(10.0) is 1.25, so an M10 x 1.25. Zero where the catalog carries no fine pitch for the size.  A size has one coarse pitch and several fine ones, and only the first is tabulated. Nothing forces you through the table: every thread function below takes the pitch outright, so thread.male(10.0, 1.0) builds an M10 x 1 whether or not it is a catalog default.
- **`uncPitch(nominalInch)`** — The pitch in millimetres of a Unified Coarse thread of the given nominal diameter in INCHES: thread.uncPitch(0.25) is the 20-per-inch pitch of a 1/4-20, as 1.27 mm.
- **`unfPitch(nominalInch)`** — The pitch in millimetres of a Unified Fine thread, nominal diameter in INCHES: thread.unfPitch(0.25) is a 1/4-28.
- **`minorDia(major, pitchVal)`** — The basic minor diameter of an internal thread — the crest of the nut's own material, and the smallest thing a screw of this size passes.
- **`pitchDia(major, pitchVal)`** — The basic pitch diameter, where the flanks are half a pitch apart. The diameter a thread is actually toleranced on.
- **`tapDrill(major, pitchVal)`** — The drill to run before tapping: nominal minus pitch, the shop rule that lands near 75 percent thread engagement. thread.tapDrill(6.0, 1.0) is 5.0, the drill for an M6.

#### Thread Profiles

- **`male(major, pitchVal)`** — The surface of an EXTERNAL thread of the given major diameter and pitch — the screw itself, negative inside the metal. The ISO 68-1 basic profile, unbounded along z: bound it where you use it.  A true distance field, so it traces and meshes like any other solid. Prove against the smooth cylinder at the major diameter all the same: a proof against the helix subdivides to thread scale for nothing the cylinder does not already give you.
- **`female(major, pitchVal)`** — The void an INTERNAL thread presents — a tapped hole, negative inside the hole. Unbounded along z.  The SAME field as male() at the same size, and that is not an oversight: one basic profile separates the two threads, so a screw and the hole tapped for it meet flank to flank with nothing between them. What opens the fit in a real joint is the tolerance class and the tap drill — and the drill you list under a tap() already cuts the nut's crest back to itself.
- **`maleSin(major, pitchVal)`** — An external thread as a SINUSOID rather than a truncated V. Not a standard form and no longer cheaper to bound, now that the real one is a distance; what it still buys is a field with no corners in it. Reach for it when the thread is scenery.
- **`femaleSin(major, pitchVal)`** — The sinusoidal internal form, the counterpart of maleSin — the same field, for the same reason female() is male().
- **`stripe(major, pitchVal)`** — A thread that costs no geometry: 0 on the major diameter and 1 on the minor, running helically at the given pitch. Paint it onto a SMOOTH bore or shank and the thread reads at any zoom, while the mesher still sees a cylinder and a proof still runs against the true envelope.  shank = fastener.shcs(6.0, 20.0); shank.color = color.mixOklab(steel, dark, thread.stripe(6.0, 1.0));

#### Relief

- **`maleRelief(major, pitchVal, zTop, zBottom)`** — The EXTERNAL thread as a displacement height, in millimetres, for material.withDisplacement: at most zero, the depth threading takes off a smooth shank at the `major` diameter, between the planes z = `zTop` and z = `zBottom`. Hand it to the shank the smooth fastener forms build, and the screw meshes and proves as its major-diameter envelope while it renders with the ISO 68-1 thread.  screw = fastener.shcs(6.0, 20.0); steelMat, steelColor = metal.steel(); screw.material = material.withDisplacement(steelMat, thread.maleRelief(6.0, 1.0, 0.0, -20.0));  `zTop` is where the thread stops under the head — the seating plane z = 0 for a cap screw, the bottom of the cone for a countersunk one — and `zBottom` is the tip. The thread runs out over about a pitch inside each, and nothing beyond a form depth outside the major diameter is touched, so the head keeps its own surface whichever material it wears.
- **`femaleRelief(major, pitchVal, boreDia, zTop, zBottom)`** — The INTERNAL thread as a displacement height: at most zero, the depth a tap takes out of the wall of a bore `boreDia` across, between the planes z = `zTop` (the mouth) and z = `zBottom`. Hand it to the PART, which carries the plain bore.  Wherever the bore is wider than the thread's own crest the tap cuts nothing, so a hole drilled at thread.tapDrill comes out with its crests truncated to the drill, which is how a tapped hole is actually made. `boreDia` above the major diameter leaves no thread at all.
- **`femaleRelief(major, pitchVal, zTop, zBottom)`** — The internal thread relief on a bore at the BASIC MINOR diameter — the maximum material condition of a tapped hole, and the smallest bore the thread's own profile allows. Drill the bore with thread.minorDia(major, pitch) to match.
- **`maleSinRelief(major, pitchVal, zTop, zBottom)`** — The external thread as a COSMETIC SINUSOID, for material.withDisplacement: maleRelief with maleSin's profile in place of the ISO one, and the same arguments and bounds. Not a standard form. What it buys is a height with no corners, which stays smooth where the V's flanks and flats would alias — threads a few pixels a pitch, a board of small screws.  steelMat, steelColor = metal.steel(); screw.material = material.withDisplacement(steelMat, thread.maleSinRelief(6.0, 1.0, 0.0, -20.0));
- **`femaleSinRelief(major, pitchVal, boreDia, zTop, zBottom)`** — The internal thread as a cosmetic sinusoid: femaleRelief with femaleSin's profile, the same arguments and bounds, on a bore `boreDia` across. A sinusoidal screw in a sinusoidal hole meshes, as maleSin and femaleSin do.
- **`femaleSinRelief(major, pitchVal, zTop, zBottom)`** — The sinusoidal internal relief on a bore at the BASIC MINOR diameter.
- **`reliefColor(base, height, pitchVal)`** — The colour a relief stands in as wherever it is not drawn as geometry: `base` on the crest, darkening toward the root by the depth `height` (any relief above, at the same `pitchVal`) takes off there. It fades to plain `base` by displacementDrawn(), so where the renderer moves the surface the thread is shape alone, and where the Detail slider has taken the displacement from a part small on screen — or nothing displaces at all, as in a mesh or an export — the same thread is paint. Too small to resolve, it is `base`.  h = thread.maleRelief(6.0, 1.0, 0.0, -20.0); steelMat, steelColor = metal.steel(); screw.material = material.withDisplacement(steelMat, h); screw.color = thread.reliefColor(steelColor, h, 1.0);

#### Engagement

- **`maleZone(major, pitchVal, zTop, zBottom)`** — The ZONE an EXTERNAL thread's relief works in, as a distance field negative inside: the band a form depth deep under the `major` diameter, between the planes z = `zTop` and z = `zBottom`. Same arguments as maleRelief.
- **`femaleZone(major, pitchVal, boreDia, zTop, zBottom)`** — The ZONE an INTERNAL thread's relief works in: the band from the wall of a bore `boreDia` across out to the `major` diameter, between the planes z = `zTop` and z = `zBottom`. Same arguments as femaleRelief; the pitch changes nothing here and is taken so the two calls read alike.
- **`femaleZone(major, pitchVal, zTop, zBottom)`** — The internal thread's zone on a bore at the BASIC MINOR diameter — the one drill.tap and fastener.nut cut.
- **`engaged(zoneA, zoneB, margin)`** — Where a threaded joint MAY overlap, as a field POSITIVE there: inside both threads' zones, grown by `margin`. Take the max of the two parts with it and prove the result is not negative — it is below zero only where the parts overlap outside the engagement:  screw = fastener.shcs(6.0, 20.0); plate = max(box([-10.0, -10.0, -12.0], [10.0, 10.0, 0.0]), -drill.hole(drill.tap(6.0, 14.0), 0.0)); ok = max(screw, plate, thread.engaged(thread.maleZone(6.0, 1.0, 0.0, -20.0), thread.femaleZone(6.0, 1.0, 0.0, -12.0), 0.1));  and certify `ok` over a region that stops just below the seating plane, since the head resting on the plate is contact of its own. The tap runs past the plate's far face: stopping on it leaves the bore's floor in that face, a membrane no proof gets past. See the Engagement note above for why the margin, and what it lets through.

#### Tolerance Position (ISO 965-1)

- **`deviationE(pitchVal)`** — The fundamental deviation es of an EXTERNAL thread in position e, in millimetres and at most zero — the loosest of the four, for threads that will be plated thick. ISO 965-1 Table 1; zero below P 0.5.
- **`deviationF(pitchVal)`** — The fundamental deviation es of an external thread in position f. ISO 965-1 Table 1; zero below P 0.35.
- **`deviationG(pitchVal)`** — The fundamental deviation es of an external thread in position g — the one a 6g screw carries: thread.deviationG(1.0) is -0.026 for an M6.
- **`deviationH(pitchVal)`** — Position h: no deviation, the external thread at basic.
- **`internalDeviationG(pitchVal)`** — The fundamental deviation EI of an INTERNAL thread in position G, in millimetres and at least zero: the hole opened up by it.
- **`internalDeviationH(pitchVal)`** — Position H: no deviation, the internal thread at basic — the everyday tapped hole.

### `mechanical/drill`

#### Constants

- **`pointAngle118 = radians(118.0)`**
- **`pointAngle135 = radians(135.0)`** — The split point: 135 degrees included. Stainless, hardened, and anything that wants a self-centering start.
- **`centerAngle60 = radians(60.0)`** — The 60-degree included angle of an ANSI B94.11 center, and of the pilot countersink centerHole() cuts.

#### Entry Points

- **`hole(features, zTop)`** — The void solid of a hole whose mouth is the plane z = `zTop`, cut downward. `features` is a list of feature closures from this package — drill.thru, drill.bore, drill.countersink and the rest. Every depth in them is measured from the mouth, so the list reads like the lines of a drawing callout and its order does not matter.  A void solid: subtract it with max(block, -drill.hole([...], 0)). Place it on a face with remap() and drill.axisAt().
- **`hole(features)`** — The same stack with its mouth at z = 0.

#### Internal Geometry

- **`hexXY(acrossFlats)`** — A HEXAGONAL section in XY, `acrossFlats` wide, flats facing +-x: every hex head, nut, socket and bar in the mechanical packages is this one field. Unit-gradient (a max of six half-planes via abs).
- **`squareXY(acrossFlats)`** — A SQUARE section in XY, `acrossFlats` wide: the square head's and the Robertson's.

#### Stacked Features

- **`thru(dia)`** — A drilled hole clean through the part: diameter `dia`, no floor. The THRU of a drawing callout — the part's own faces bound it, and there is no drill point in it because the drill went out the far side.
- **`blind(dia, depth)`** — A BLIND drilled hole of diameter `dia`, `depth` deep to the full-diameter shoulder — the depth a drawing dimensions. Below it sits the 118-degree point the drill leaves, which is never dimensioned: for a 6 mm drill it adds 1.80.
- **`blind(dia, depth, included)`** — A drilled hole with the point angle given, `included` in radians — drill.pointAngle118 or drill.pointAngle135. `depth` still reaches the full-diameter shoulder, not the tip.
- **`bore(dia, depth)`** — A BORED hole of diameter `dia`, `depth` deep, flat bottomed: boring is the operation that opens an existing hole to size and concentricity, so this is also the counterbore, the end-milled pocket and the reamed flat. The counterbore line of a callout.
- **`chamfer(dia)`** — A 45-degree chamfer whose large diameter is `dia` at the mouth. It narrows to nothing, so whatever follows it cuts it off: list it with the bore it leads into.
- **`chamfer(dia, leg)`** — A 45-degree chamfer of leg `leg` on a bore of diameter `dia` — the drawing note "0.5 x 45 CHAMFER" on a 6 mm hole is drill.chamfer(6.0, 0.5).
- **`chamferAt(dia, leg, atDepth)`** — A 45-degree chamfer of leg `leg` on a bore of diameter `dia`, its large diameter `atDepth` below the mouth rather than at it — the lead-in on a bore that starts at the floor of a counterbore, which a drawing dimensions on that bore and not on the face.  The general form of this is a combinator, at(depth, feature), which would shift ANY feature down. It cannot be written yet: it would have to capture a function, and a captured function is the one thing the closure pass still bakes rather than lifts.
- **`countersink(dia, included)`** — A countersink of large diameter `dia` at `included` angle in radians — the drawing spells it "CSK DIA 13 X 90", diameter and angle, never a depth. Use radians(90) for an ISO 10642 flat head, radians(82) for the inch series. To break its own outer edge, list a wider chamfer with it; that needs no new function, which is the point of a feature list.
- **`countersinkAt(dia, included, atDepth)`** — A countersink whose rim sits `atDepth` below the mouth rather than at it.
- **`taper(top, included)`** — A taper from diameter `top` at the mouth, narrowing at `included` radians and running to nothing. The general form chamfer and countersink are special cases of.
- **`taper(top, included, depth)`** — A taper from diameter `top` at the mouth, narrowing at `included` radians, with a flat floor `depth` below the mouth — a frustum rather than a cone run to nothing.
- **`centerHole(pilotDia, csDia, depth)`** — An ANSI B94.11 center: a pilot drill of diameter `pilotDia` to `depth`, under a 60-degree countersink out to `csDia`. The feature a shaft end carries so it can be turned between centers — the likeliest occupant of the two holes at the ends of a shaft.

#### Non-Round Sections

- **`hexSocket(acrossFlats)`** — A hexagonal socket `acrossFlats` wide, through. Flats face +-x.
- **`hexSocket(acrossFlats, depth)`** — A hexagonal socket `acrossFlats` wide, `depth` deep with a flat bottom. Broached hex, a hex drive, a nut pocket.
- **`squareSocket(acrossFlats)`** — A square socket `acrossFlats` wide, through.
- **`squareSocket(acrossFlats, depth)`** — A square socket `acrossFlats` wide, `depth` deep with a flat bottom.
- **`squareSocket(acrossFlats, depth, taper)`** — A TAPERED square socket, `acrossFlats` at the face, `depth` deep, each wall leaning in at `taper` radians.  A wall taper wedges a tapered bit in and still lets it go under a straight pull, which is how a Robertson is usually described. It is not how Robertson drew it — his walls are parallel and the taper is the pyramid at the bottom — and a modern recess holds its bit without one. For the recess itself use drill.robertsonAt.
- **`torx(major, minor)`** — A TORX (hexalobular) socket cut clean through: `major` across the lobes, `minor` across the flutes, at the drawn fillet proportion.
- **`torx(major, minor, depth)`** — A TORX (hexalobular, ISO 10664) socket, `depth` deep with a flat bottom. `major` is the diameter across the lobes — the A dimension a Torx size is quoted by — and `minor` the diameter across the flutes between them.  Six lobes, always: hexalobular is what the standard calls it, and the lobe count is not a parameter of the standard. The flanks are straight and radial rather than tapered, which is the entire reason a Torx bit does not cam out and does not need to be pressed into the screw to stay there.  Sizes come from the catalog: catalog.torxMajor(25) and catalog.torxMinor(25) are a T25.
- **`torx(major, minor, depth, filletR)`** — A Torx socket with its fillet radius given rather than proportioned.

#### Drives as Values

- **`hexSocketAt(zTop, acrossFlats, depth)`** — A hexagonal socket `acrossFlats` wide and `depth` deep, its face at `zTop`.
- **`squareSocketAt(zTop, acrossFlats, depth)`** — A square socket `acrossFlats` wide and `depth` deep, its face at `zTop`, with straight walls.
- **`squareSocketAt(zTop, acrossFlats, depth, taper)`** — A TAPERED square socket, `acrossFlats` at the face and leaning in at `taper` radians per wall — see drill.squareSocket.
- **`robertsonAt(zTop, acrossFlats, depth, leg)`** — A ROBERTSON (square) recess `acrossFlats` wide with its face at `zTop`, `depth` to its deepest point: parallel walls, and a 45-degree chamfer of `leg` running out from the floor to them. A leg of 0 is a flat bottom; a leg of half the width across flats is the four-sided pyramid of the 1909 patent. Sizes come from the catalog:  drill.robertsonAt(zTop, catalog.robertsonAcross(2.0), catalog.robertsonDepth(2.0), 0.5)
- **`robertsonAt(zTop, acrossFlats, depth)`** — A Robertson recess with the patent's pyramid bottom — the chamfer taken all the way to the axis.
- **`pozidrivAt(zRef, dia, core, wing, knee, lean, delta, markDepth)`** — An ISO 4757 TYPE Z (POZIDRIV) recess seated on the reference plane `zRef`, `dia` across there — the screw's m — in the form of one recess number: core b, wing f, knee g, valley lean alpha and delta, and mark depth j, from catalog.pozidrivCore and its neighbours. The marks run out to the recess diameter at 60 degrees across, the near end of ISO's "1 to 1,25 m" and the narrowest of its "60 min, 90 max" — a reading, not a dimension.  fastener.panPozidrivDrive sizes one from the catalog.
- **`pozidrivGauge(zKnee, zTop, core, wing, knee, point, lean, delta)`** — The ISO 4757 type Z GAUGE — a Pozidriv screwdriver's point — as a SOLID, its knee at `zKnee` and cut off at `zTop`; its point is `point` below the knee. Pass Table 4's numbers (catalog.pozidrivGaugeCore and neighbours).
- **`phillipsAt(zRef, dia, core, facet, g, ridge, lean, root)`** — An ISO 4757 TYPE H (PHILLIPS) recess seated on the reference plane `zRef`, `dia` across there — the m of the screw's product standard — in the form of one recess number: core width b, ridge facet e and angle alpha, point diameter g and ridge lean beta, from catalog.phillipsCore and its neighbours. `root` is No. 0's ridge radius and zero for every other number.  The wings are what the four ribs leave between them, so their thickness f is not passed; ISO's Table 1 f comes back out of b, e and the angles. The depth is not passed either: the wing-end cone crosses zRef at `dia`, which sinks the one form of this number as deep as the screw's m puts it.  drill.phillipsAt(2.716, 4.4, 1.47, 0.815, 2.29, radians(140.0), radians(5.75), 0.0)
- **`phillipsGauge(zTip, zTop, core, facet, g, flat, ridge, lean, root)`** — The ISO 4757 type H GAUGE — the point of a Phillips screwdriver — as a SOLID, its flat end at `zTip` and cut off at `zTop`: the one solid in this file, here because it is the recess's own construction read from the other side. Its blades are what the same four ribs leave standing inside the 26 degree 30 cone, which is g across where it meets an 18 degree end chamfer running in to a flat `flat` across. Pass Table 2's numbers (catalog.phillipsGaugeCore and its neighbours), not the recess's.
- **`crossRecessAt(zTop, dia, wingWidth, depth, flank)`** — A cross recess with its face at `zTop` — see drill.crossRecess.
- **`crossRecessAt(zTop, dia, wingWidth, depth, flank, rootR)`** — A cross recess with a plan corner fillet of radius `rootR`, its face at `zTop` — see drill.crossRecess.
- **`crossRecessAt(zTop, dia, wingWidth, depth, flank, rootR, outerTaper)`** — A cross recess with a plan corner fillet AND a conical outer trim leaning out at `outerTaper`, its face at `zTop` — see drill.crossRecess. `dia` is then the recess diameter AT THE FACE, which is where it is measured.
- **`crossRecessRibbedAt(zTop, dia, wingWidth, depth, flank,ribWidth, ribDepth, ribFlank)`** — A ribbed cross recess with its face at `zTop`, ribs running out to the full recess diameter — see drill.crossRecessRibbed.
- **`crossRecessRibbedAt(zTop, dia, wingWidth, depth, flank,ribDia, ribWidth, ribDepth, ribFlank)`** — A ribbed cross recess with its face at `zTop` — see drill.crossRecessRibbed.
- **`crossRecessRibbedAt(zTop, dia, wingWidth, depth, flank,ribDia, ribWidth, ribDepth, ribFlank, rootR)`** — A ribbed cross recess with a plan corner fillet of radius `rootR`, its face at `zTop` — see drill.crossRecessRibbed.
- **`crossRecessRibbedAt(zTop, dia, wingWidth, depth, flank,ribDia, ribWidth, ribDepth, ribFlank, rootR, outerTaper)`** — A ribbed cross recess with a plan corner fillet AND a conical outer trim leaning out at `outerTaper`, its face at `zTop`.
- **`torxAt(zTop, major, minor, depth)`** — A TORX (hexalobular, ISO 10664) socket `depth` deep, its face at `zTop`: `major` across the lobes, `minor` across the flutes. See drill.torx.
- **`torxAt(zTop, major, minor, depth, filletR)`** — A Torx socket with the fillet radius given — see drill.torx.
- **`slotDriveAt(zTop, width, depth)`** — A straight SLOT drive `width` across and `depth` deep, its face at `zTop`, running clean across the head. The screwdriver that predates all the rest.

#### Identification Marks

- **`tickSection(width, innerR, outerR)`** — The section of four TICK MARKS at 45 degrees to the axes, `width` across and running from radius `innerR` out to `outerR`. Unbounded in z, like every other section here — bound it with tickMarksAt for a flat face or engrave for a curved one.
- **`dimpleSection(dia, atR)`** — The section of a round DIMPLE of diameter `dia`, its centre `atR` out from the axis along +X.
- **`tickMarksAt(zTop, width, innerR, outerR, depth)`** — Four shallow tick marks stamped `depth` into the FLAT face at `zTop`.
- **`dimpleAt(zTop, dia, atR, depth)`** — A shallow dimple `depth` into the FLAT face at `zTop`.
- **`engrave(surface, section, depth, zFrom)`** — ENGRAVE `section` into `surface`, `depth` deep, FOLLOWING THE SURFACE rather than cutting down from a plane.  A mark bounded by a plane only works on a flat face. Stamp one on a pan head — which is domed — and at the rim the head has already fallen below that plane, so the mark cuts nothing and silently is not there. Offsetting the surface itself instead works on any shape: the void is the part of the section lying within `depth` of the surface, which is what a stamp does.  The void is OPEN ABOVE and bounded below by the plane z = `zFrom`.  Both of those are load-bearing, and each fixes a different way of getting this wrong.  BOUNDED BELOW, because a surface offset follows EVERY surface the shape has. Written as max(section, -(surface + depth)) alone, the void is every point of the section not deeper than `depth` inside — which is the whole of space outside the part as well. A tick mark written that way does not mark a head; it saws the screw in half down four planes. `zFrom` picks the face the mark belongs on, which for a screw head is the top.  OPEN ABOVE, and NOT closed off with the surface itself. Adding `surface` to the max does bound the void — and it bounds it with the crown, putting the void's top face exactly on the face it opens onto. Then at the crown the field reads 0 from both sides: a membrane, not an opening. It samples correctly at every depth inside the mark and RENDERS AS NO MARK AT ALL, which is the defect this file's header warns about, arrived at from the other direction. The void runs up into the air instead and the head's own crown is what bounds it.  head = fastener.panHead(12.0, 4.4); mark = drill.engrave(head, drill.tickSection(1.0, 3.0, 5.4), 0.45, 2.2);
- **`crossRecess(dia, wingWidth, depth, flank)`** — A CROSS RECESS: four wings of width `wingWidth` reaching out to diameter `dia`, cut `depth` deep, their flanks tapering at `flank` radians from the axis. A GENERIC tapered cross — four slabs — and NOT ISO 4757's type H, whose wings are what four ribs leave between them; that one is drill.phillipsAt.  THE FLANK ANGLE decides whether a driver cams out. Taper the flanks and a driver rides up out of the recess under torque — which is deliberate on a Phillips: cam-out was a torque limiter before torque drivers were cheap. Bring them near parallel and it stops, which is half of what a Pozidriv does differently.  drill.crossRecess(3.5, 1.8, 2.0, radians(7.0))  This is the shape of a recess, not a bit seat: the wing ends are cut by the outer diameter rather than tapering in plan, and the bottom is flat rather than pointed. Enough to draw a cross in a head and to prove a driver clears it; not enough to model a driver seating.
- **`crossRecess(dia, wingWidth, depth, flank, rootR)`** — A cross recess with a PLAN CORNER FILLET of radius `rootR` where adjacent wings meet.  Rounded, the arms of a Phillips appear to flow into one another; left sharp, a cross reads as four straight slots crossing.  It is a plan-view radius, in the section you would see looking down the axis, and NOT a break around the mouth.
- **`crossRecessRibbed(dia, wingWidth, depth, flank,ribDia, ribWidth, ribDepth, ribFlank)`** — A cross recess WITH RIBS: the four wings above plus four secondary ribs at 45 degrees between them, `ribWidth` across and `ribDepth` deep. The Pozidriv and Supadriv family, and the four tick marks stamped between the arms on such a screw head are exactly these.  The ribs are why it does not cam out: they give the bit a second set of faces to bear on that are not part of the tapered wing. Supadriv differs from Pozidriv in the rib shape rather than in having them, so both are this feature at different numbers.  A POZIDRIV'S WINGS ARE VERY NEARLY PARALLEL. Pass a Phillips flank angle here and the result is not a Pozidriv with ribs added; it is a Phillips with ribs added, which cams out exactly as a Phillips does and so has no reason to carry ribs at all. Not camming out IS the design: the ribs give the bit somewhere to bear, and the near-vertical flanks are what stop it climbing.  The ribs are shorter than the wings — `ribDia` rather than `dia` — as well as narrower and shallower.
- **`crossRecessRibbed(dia, wingWidth, depth, flank, ribWidth, ribDepth, ribFlank)`** — A ribbed cross recess whose ribs run out to the full recess diameter.  The older form, kept because it is a shape somebody may want and because giving the ribs their own length was an addition rather than a correction of this. For a Pozidriv pass ribDia explicitly and make it smaller than dia.
- **`crossRecessRibbed(dia, wingWidth, depth, flank,ribDia, ribWidth, ribDepth, ribFlank, rootR)`** — A ribbed cross recess with a plan corner fillet of radius `rootR`.
- **`dFlat(dia, acrossFlat)`** — A D-hole through: a bore of diameter `dia` with one flat, `acrossFlat` measured from the flat to the far side of the bore. The internal counterpart to a D-shaft, and the usual anti-rotation feature for a knob or a light-duty drive. The flat faces +x.
- **`dFlat(dia, acrossFlat, depth)`** — A D-hole `depth` deep with a flat bottom.
- **`doubleD(dia, acrossFlats)`** — A double-D hole through: a bore of diameter `dia` with two opposed flats `acrossFlats` apart, facing +-x.
- **`doubleD(dia, acrossFlats, depth)`** — A double-D hole `depth` deep with a flat bottom.
- **`tap(major, depth)`** — The smooth void a tapped hole presents at its MAXIMUM MATERIAL CONDITION: a cylinder at the thread's basic MINOR diameter, `depth` deep, for the coarse pitch of this `major`. Threading is a machining operation that only ever takes material from there, so the screw that goes in INTERFERES with this bore by design, across the annulus out to the major diameter. Prove the fit with the engagement zone that interference lives in — see thread.engaged — and draw the thread with thread.femaleRelief on the part's material.  A drawing writes the drill and the thread as two lines ("DRILL 5.0 DEEP 20 / M6x1 DEEP 15"), and so do you: list a blind() under a tap(). A tap drill wider than the basic minor is the bore then, which is how a tapped hole is made; pass that bore to thread.femaleRelief as well.
- **`tap(major, pitchVal, depth)`** — tap() for any pitch — a FINE thread's minor diameter is larger than the coarse one of the same major, so the pitch is part of the bore.
- **`tapThreaded(major, pitch, depth)`** — A tapped hole with a display-only thread of the given major diameter and pitch, `depth` deep. The real ISO 68-1 internal form — wide crest flat, narrow root — from the thread package, so this hole and the screw that goes into it are drawn from one definition.  It is tap() with the thread cut into it, so prove against tap(): a 60-degree flank makes an interval proof subdivide to thread scale for no engineering information the smooth form does not already give.
- **`tapSin(major, pitch, depth)`** — A tapped hole whose thread is a SINUSOID rather than the standard V. Not a real form, but it meshes and bounds more cheaply — reach for it when the thread is scenery rather than the subject.

#### Located Features

- **`undercut(dia, fromDepth, toDepth)`** — A relief groove of diameter `dia` between `fromDepth` and `toDepth` below the mouth, toDepth the deeper. The one feature WIDER than the bore above it, so the one that needs both bounds — an undercut with no top would eat the wall it exists to relieve. A thread relief, a retaining-ring groove, an O-ring gland of square section.
- **`spline(count, major, width)`** — A female straight-sided spline cut clean through: `count` grooves outward to diameter `major`, each `width` across, bounded by neither face. The grooves only — list a blind() at the minor diameter with it. count >= 3.
- **`spline(count, major, width, toDepth)`** — A female straight-sided spline: `count` grooves cut outward to diameter `major`, each `width` across, running from the mouth down to `toDepth`. The DIN 5463 parallel-side form. The grooves only — list a blind() or bore() at the minor diameter with it, so the bore reads as its own line, the way a drawing writes it. count >= 3.
- **`splineBetween(count, major, width, fromDepth, toDepth)`** — A spline bounded at both ends, between `fromDepth` and `toDepth` — a set of grooves that neither starts at the mouth nor runs out of it. Give `fromDepth` 0 only if something wider above already opens the face: a groove that stops exactly on the face has a surface lying in it.
- **`keyway(width, depth, boreDia)`** — A keyway cut clean through the wall of a bore of diameter `boreDia`: one slot `width` across, `depth` radially outward, bounded by neither face. On +x. What a keyed bore in a plate or a hub wants.
- **`keyway(width, depth, boreDia, toDepth)`** — A keyway in the wall of a bore of diameter `boreDia`: one slot `width` across, cut `depth` radially outward from the bore, running from the mouth down to `toDepth`. On +x. Takes the bore diameter explicitly, because a keyway is dimensioned on its own detail rather than stacked.
- **`keywayBetween(width, depth, boreDia, fromDepth, toDepth)`** — A keyway bounded at both ends: an internal keyseat that neither starts at the mouth nor runs out of it, between `fromDepth` and `toDepth`. Give `fromDepth` 0 only if something wider above already opens the face.
- **`slot(dia, centers, ang)`** — An obround slot through: diameter `dia`, end centers `centers` apart — which is how a drawing dimensions one, not overall length — with the long axis at `ang` radians from +x.
- **`slot(dia, centers, ang, depth)`** — An obround slot `depth` deep with a flat bottom.
- **`slotChamfer(dia, centers, ang, leg)`** — The chamfer for a slot: a 45-degree lead-in of leg `leg` that FOLLOWS the obround it breaks, given that slot's own `dia`, `centers` and `ang`. List it with the slot, the way chamfer() is listed with its bore.  A round chamfer on a slot breaks the edge at the two ends and cuts a crescent out of the face along the straight, which is neither what the tool does nor what the drawing says. The edge to break is the slot's, so the chamfer has to be the slot's shape.

#### Placement

- **`axisAt(point, axis)`** — The remap coordinate that puts a +Z-datum hole at `point` with its axis along `axis` — the direction the drill enters, which is the outward normal of the face it goes into. Pass it to remap:  remap(drill.hole([drill.thru(6.6)], 0), drill.axisAt(faceCenter, faceNormal))  THE IMPLEMENTATION IS map.axisAt, which is where it belongs: aligning a datum to a direction is not a fact about holes, and map is where rotateAxis and the pattern remaps already live. This forwards so a hole can be placed without reaching outside the package it came from.

### `mechanical/driver`

#### Constants

- **`crossTool = [[0.0,  3.0, 2.78,  60.0, 2.0],`**
- **`handleTable = [[0.0, 23.0, 104.0],`**

#### Tool Sizes

- **`bladeDia(recessNo)`** — ISO 8764-1 blade diameter of a cross recess screwdriver of this recess number: 6 for a PH2 or a PZ2.
- **`fluteLength(recessNo)`** — The fluted length of a cross tip of this recess number — how far up the tip the recess form runs before the tool is solid.
- **`bladeLength(recessNo)`** — ISO 8764-2 series B blade length for this recess number: 100 for a PH2.
- **`handleDia(recessNo)`** — An illustrative screwdriver handle diameter for this recess number.
- **`handleLength(recessNo)`** — An illustrative screwdriver handle length for this recess number.

#### Tips

- **`phillipsGauge(recessNo)`** — The FLOOR RADIUS of a Phillips recess number in ISO 7045: where its 26 degree 30 cone, m wide at the head, comes down to at the maximum penetration — m/2 less the penetration times tan 26.5, 1.00 for a PH2. Every size a number goes in gives the same one, which is what makes the number one recess.  A CHECK ON THE CATALOG, NOT A DIMENSION OF THE TIP. phillipsTip is ISO 4757's gauge and does not read this; it is kept because it is the one number that ties ISO 7045's per-size m and penetration back to the recess number.
- **`pozidrivGauge(recessNo)`** — The floor radius of a Pozidriv recess number in ISO 7045, the same way — a catalog check, not what pozidrivTip is built from.
- **`phillipsTip(recessNo)`** — A PHILLIPS (ISO 4757 type H) tip of this recess number, up the fluted length: ISO 4757's own gauge, whose point the standard says "is identical with the point of the respective screwdriver" — drill.phillipsGauge at Table 2's numbers, its flat end at z = 0. ISO 4757:1983 section 2.2, https://cdn.standards.iteh.ai/samples/10742/d878fa0371e041b39de75de4d4426928/ISO-4757-1983.pdf
- **`pozidrivTip(recessNo)`** — A POZIDRIV (ISO 4757 type Z) tip of this recess number, up the fluted length: ISO 4757's type Z gauge (Table 4, read from IS 7478, https://archive.org/details/gov.in.is.7478.2011) — drill.pozidrivGauge, its point at z = 0.
- **`torxTip(tNumber, length)`** — A TORX (ISO 10664) tip of this T number, `length` long: the hexalobular section, straight-sided, at the recess's own dimensions.
- **`hexTip(acrossFlats, length)`** — A HEX tip `acrossFlats` wide and `length` long — the key for a hex socket.
- **`robertsonTip(sizeNo, length)`** — A ROBERTSON (square) tip of this size number, `length` long, ending in the four-sided point of the 1909 patent: drill.robertsonAt read as a solid.
- **`slotTip(thickness, width, length)`** — A SLOTTED tip, ISO 2380 form A: a blade `thickness` by `width` at its end, `length` long. Its faces open at the standard's 10-degree wedge for the first 1.5, which is the part that sits in the slot, and then at 30 degrees, which is how the blade gets back to the bar it was ground from in the length of a bit rather than of a crowbar.

#### Shanks

- **`insertShank(zLow, neckR)`** — A 1/4-INCH INSERT BIT's shank, ISO 1173 form C 6.3: hex 6.32 across flats, 25 long overall, with the V-groove on its corners 8.2 from the drive end that a magnetic or spring holder catches.
- **`powerShank(zLow, neckR)`** — A 1/4-INCH POWER BIT's shank, ISO 1173 form E 6.3 — the SELF-LOCKING one: a round groove, root diameter 4.7 at 9.5 from the drive end, that the ball of a quick-release chuck drops into. 50 long, hex over the last 25, with a slimmer round torsion zone between it and the tip (4.4 across, ILLUSTRATIVE).
- **`microShank(zLow, neckR)`** — A 4 mm PRECISION BIT's shank, ISO 1173 form C 4: hex 3.94 across flats, 28 long overall.
- **`bit(tip, shank)`** — A bit: `tip` and `shank` together.

#### Screwdrivers

- **`screwdriverBlade(zLow, neckR, bladeDia, bladeLen)`** — A screwdriver BLADE: round, `bladeDia` across, from its tip's working length `zLow` (necked from `neckR`) up `bladeLen` of exposed blade and on another 25 into the handle. Min it with a tip.
- **`screwdriverHandle(zBase, dia, len)`** — A screwdriver HANDLE `dia` across and `len` long, starting at `zBase`: a ferrule taper at the front, a rounded butt, and six shallow grip flutes. ILLUSTRATIVE — the shape of a handle is a maker's, not a standard's.

### `mechanical/joint`

#### Bolted

- **`boltedLength(nominalDia, grip)`** — The ISO 4762 length a bolted joint of this size and grip uses: the shortest catalog length that runs 1.5 diameters past the grip. Zero if no catalog screw is that long.
- **`bolted(clamped, threaded, seatPoint, shankDir, nominalDia, grip, fit)`** — A BOLTED JOINT: an ISO 4762 cap screw seated at `seatPoint` with its shank running along `shankDir`, clamping `clamped` (`grip` thick under the head) to `threaded`, at ISO 273 clearance `fit` (fastener.fitFine, fitMedium or fitCoarse). Returns the screw, then `clamped` with its counterbore and clearance cut, then `threaded` with its tapped hole cut, each a part of its own:  screw, lid, base = joint.bolted(lidBody, baseBody, seat, shankDir, 6.0, 7.5, fastener.fitMedium); lidMat, lidColor = metal.aluminum(); lid.material = lidMat; return screw, lid, base;  The holes are open above and the clearance is a thru, so each part's own faces bound them: the counterbore opens through whatever of `clamped` is above the seat. Chain joints by passing one joint's returned parts to the next.

#### Claims

- **`boltedSeat(seatPoint, shankDir, nominalDia, fit)`** — The SEAT ALLOWANCE: POSITIVE in a band 0.1 either side of the seating plane, out to the counterbore's radius. The head bears on the counterbore's floor, which is contact, not interference; take the max of the screw and a part with this and the floor is let through and nothing else is.
- **`boltedEngaged(seatPoint, shankDir, nominalDia, grip)`** — The ENGAGEMENT ALLOWANCE: positive where the screw's thread and the tapped hole's are both cut, grown by 0.2 — thread.engaged for this joint. The smooth screw at its major diameter overlaps the tapped part at its minor there by design. The margin is twice thread.engaged's usual 0.1 so that a joint on a tilted axis, whose intervals are looser, still proves in a default budget; the threads are the same size and pitch by construction, so the wider band lets nothing through that a mismatch would have made.
- **`boltedGripZone(seatPoint, shankDir, nominalDia, grip, fit)`** — The GRIP ZONE, negative inside: the ring under the head, from the clearance out to the head's own diameter, and along the axis the grip inset from both its faces by the slack (see jointSlack). The clamped part has to fill it — prove max(zone, neg(clampedBody)) is never negative — or the head bears on nothing, or the grip the joint was told is not the parts' grip.
- **`boltedThreadZone(seatPoint, shankDir, nominalDia, grip)`** — The THREAD ZONE, negative inside: a cylinder three quarters of a diameter in radius — a quarter-diameter wall outside the thread's crest — from the slack past the grip to 1.5 diameters past it. The threaded part has to fill it, or the thread breaks out of a wall or engages less than 1.5 d.

### `text`

- **`fontStroke()`** — Default stroke radius, as a fraction of cap height.
- **`fontSpacing()`** — Default inter-glyph spacing (tracking), as a fraction of cap height.
- **`left(codes, position, size, thickness, spacing)`** — Render character codes (e.g. a string literal "Hi") as proportionally-spaced left-aligned text in the XY plane (the position is the lower-left corner of the first glyph); see center() and right() for the other justifications. codes:     character codes / string literal position:  [x, y] of the text origin (lower-left of the first glyph) size:      cap height in output units thickness: stroke radius, as a fraction of cap height spacing:   inter-glyph gap (tracking), as a fraction of cap height thickness and spacing are relative metrics, so they scale with size. Returns a 2D SDF; extrude with e.g. max(..., abs(p.z) - t) for a solid.
- **`left(codes, position, size)`** — Convenience overload with default thickness and spacing.
- **`width(codes, size, spacing)`** — Width of a span of text in output units, matching the horizontal extent left() lays out: from the left edge of the first glyph to the right edge of the last (ink only — the stroke radius is symmetric and not included). codes:    character codes / string literal size:     cap height in output units spacing:  inter-glyph gap (tracking), as a fraction of cap height
- **`width(codes, size)`** — Convenience overload with default spacing.
- **`center(codes, position, size, thickness, spacing)`** — Render text horizontally centered on position.x (position.y is the baseline). Same parameters as left().
- **`center(codes, position, size)`** — Convenience overload with default thickness and spacing.
- **`right(codes, position, size, thickness, spacing)`** — Render text right-justified so position.x is the right edge (position.y is the baseline). Same parameters as left().
- **`right(codes, position, size)`** — Convenience overload with default thickness and spacing.

### `ease`

- **`interp(value, inMin, inMax, outMin, outMax, easeFn)`** — Interpolate value from [inMin, inMax] to [outMin, outMax] through an easing curve. Like ramp() but the interior of the range is shaped by easeFn rather than being linear. Pass any of the curve functions below (or a user-defined one with the same (t)-in-[0,1] signature) as the easeFn argument, e.g. ease.interp(v, 0, 1, a, b, ease.quadratic).
- **`mix(a, b, t, easeFn)`** — Mix between a and b by an eased parameter t: a at t = 0, b at t = 1, with the blend shaped by easeFn instead of being linear — the eased counterpart of the library mix(). Like it, t outside [0, 1] extrapolates per the curve. e.g. ease.mix(a, b, t, ease.cubic).
- **`linear(t)`** — Linear (identity) easing: f(t) = t. The no-op curve — pass to interp() for a straight linear remap, or as a neutral default.
- **`quadratic(t)`** — Smooth quadratic S-curve from 0 to 1 (smoothstep). Symmetric.
- **`inQuadratic(t)`** — Quadratic ease-in. Accelerates from 0.
- **`outQuadratic(t)`** — Quadratic ease-out. Decelerates into 1.
- **`cubic(t)`** — Smoother cubic S-curve from 0 to 1 (smootherstep). Symmetric, with flatter transitions at 0 and 1 than quadratic.
- **`inCubic(t)`** — Cubic ease-in. Steeper acceleration than inQuadratic.
- **`outCubic(t)`** — Cubic ease-out. Steeper deceleration than outQuadratic.
- **`exponential(t)`** — Exponential S-curve. Long flat regions near 0 and 1 with a rapid transition through the middle.
- **`inExponential(t)`** — Exponential ease-in. Very slow start, very fast finish.
- **`outExponential(t)`** — Exponential ease-out. Very fast start, very slow finish.
- **`circular(t)`** — Circular S-curve. Sharper transition than quadratic with a similar overall shape.
- **`inCircular(t)`** — Circular ease-in. Tangent to a vertical at t = 0, approaches 1 sharply at t = 1.
- **`outCircular(t)`** — Circular ease-out. Mirror of inCircular.

### `noise`

- **`cellular(size)`** — cellular(size): cellular noise (a.k.a. Worley) — the distance to the nearest feature point (F1) on a cubic lattice of cell size `size`, in world units. ~0 at the feature points, rising toward cell boundaries. Threshold it (e.g. `cellular(5) - 1`) to grow blobs around the points, or subtract from a body.
- **`cellularF2(size)`** — cellularF2(size): cellular noise F2 — distance to the SECOND-nearest feature point.
- **`cellularF1F2(size)`** — cellularF1F2(size): cellular noise F1 and F2 — the nearest and second-nearest distances, in world units, as two return values.
- **`cellularEdges(size)`** — cellularEdges(size): cellular noise F2 - F1, the Voronoi-edge field — ~0 along the borders equidistant from two feature points, larger inside cells. Negate for a network of cracks/webs along the cell walls.
- **`cellularPair(size)`** — cellularPair(size): cellularCell for the wall between two cells — the nearest and second-nearest feature distances (f1, f2, world units) and the lattice cells of both (nearCell, farCell), as four return values. Along a Voronoi wall the pair names the wall, so a hash of it decides the wall as a whole: noiseHash3(nearCell + farCell + offset) is the same on both sides, letting a texture crack some walls and leave the rest, as veins, joints and crazing do, rather than drawing every cell's outline.
- **`cellularCell(size)`** — cellularCell(size): cellular noise with the identity of each cell — the nearest and second-nearest feature distances (f1, f2, in world units, as cellularF1F2) and cell, the integer lattice cell whose feature point is nearest, as three return values. Every point of one Voronoi cell gets the same cell, so noiseHash3(cell + offset) is a random vec3 in [0,1)^3 that is constant over the whole cell: one colour per crystal, pebble, tile or grain, e.g. `f1, f2, cell = noise.cellularCell(4); tint = noiseHash3(cell + [101, 0, 0]);`. Offset the cell before hashing: noiseHash3(cell) itself places the feature point, so using it unshifted would tie a cell's value to where its point sits. f2 - f1 is ~0 along the cell walls, for dark grain boundaries.
- **`gradient(size)`** — gradient(size): classic Perlin-style gradient noise on a cubic lattice of cell size `size`, in world units. Zero at every lattice vertex; unitless, typically within roughly [-1.5, 1.5]. Unlike cellular(size), the output is NOT scaled by `size` — this is the standard, frequency-only Perlin convention, so multiply by your own amplitude. For a version pre-scaled to be safe to add directly to a distance field, see gradientBump(size).
- **`gradientBump(size)`** — gradientBump(size): gradient(size), pre-scaled so it's safe to ADD directly to a distance field for raymarching, e.g. `sdf(p) + gradientBump(size)`. The scale (size / gradientLipschitzBound()) keeps this term's OWN Lipschitz constant <= 1 in world units (certified — see gradientLipschitzBound). A sum of two 1-Lipschitz fields is itself 2-Lipschitz (not 1), so sphere-trace the combined field with steps of `combined / 2` (or `gsl certify -lipschitz 2`) — never assume the sum is still Lipschitz-1.
- **`gradient4(size, timeScale)`** — gradient4(size, timeScale): true 4D gradient (Perlin) noise — gradient(size) with time as a genuine 4th lattice axis, so the pattern morphs as time advances instead of sliding. `timeScale` is the cell size along the time axis, in seconds: smaller values morph faster. Zero at every 4D lattice vertex; unitless, typically within roughly [-1.5, 1.5]; NOT scaled by `size` (same frequency-only convention as gradient). For a version safe to add directly to a distance field, see gradientBump4(size, timeScale).
- **`gradientBump4(size, timeScale)`** — gradientBump4(size, timeScale): gradient4(size, timeScale), pre-scaled so it's safe to ADD directly to a distance field for raymarching, e.g. `sdf(p) + gradientBump4(size, timeScale)`. The scale keeps this term's OWN SPATIAL Lipschitz constant <= 1 in world units at every instant (certified — see gradient4LipschitzBound). As with gradientBump, a sum of two 1-Lipschitz fields is 2-Lipschitz — sphere-trace with steps of `combined / 2`, never assume the sum is still Lipschitz-1.
- **`fbm2(size, gain, lacunarity)`** — fbm2(size, gain, lacunarity): two-octave fractal noise — gradient(size) plus a `gain`-weighted, `lacunarity`-times-finer second octave. Unscaled, like gradient(size) — no Lipschitz guarantee, for color/height-driving fields only. For a version safe to add directly to a distance field, see fbmBump2.
- **`fbm3(size, gain, lacunarity)`** — fbm3(size, gain, lacunarity): fbm2, plus a third octave (gain^2-weighted, lacunarity^2-times-finer).
- **`fbm4(size, gain, lacunarity)`** — fbm4(size, gain, lacunarity): fbm3, plus a fourth octave (gain^3-weighted, lacunarity^3-times-finer).
- **`fbmBumpLipschitzBound2(gain)`** — fbmBumpLipschitzBound2/3/4(gain): the certified Lipschitz bound on the matching fbmBumpN — each octave is gradientBump(some size), already <=1-Lipschitz for ANY size (gradientLipschitzBound's size/L scaling cancels the size-dependence of the derivative), so the sum's Lipschitz constant is bounded by the sum of the octave weights: a closed-form geometric series, provable by arithmetic — no fresh `gsl certify` pass needed. Also the correct bound for turbulenceBump2/3 (see below): abs() never increases a function's Lipschitz constant (||f(x)|-|f(y)|| <= |f(x)-f(y)|), so these formulas apply unchanged.
- **`fbmBump2(size, gain, lacunarity)`** — fbmBump2(size, gain, lacunarity): fbm2, pre-scaled so it's safe to ADD directly to a distance field for raymarching (certified bound — see fbmBumpLipschitzBound2). Sums of Lipschitz-bounded fields add their bounds, not just 2x — see fbmBumpLipschitzBound2's doc for why summing this with another bounded field needs its OWN combined bound, same caution as gradientBump.
- **`fbmBump3(size, gain, lacunarity)`** — fbmBump3(size, gain, lacunarity): fbm3, pre-scaled (see fbmBumpLipschitzBound3).
- **`fbmBump4(size, gain, lacunarity)`** — fbmBump4(size, gain, lacunarity): fbm4, pre-scaled (see fbmBumpLipschitzBound4).
- **`turbulence2(size, gain, lacunarity)`** — turbulence2(size, gain, lacunarity): two-octave turbulence — like fbm2, but each octave is abs(gradient(...)) before weighting, giving a creased/veiny look (no sign cancellation between octaves) instead of fbm's smooth billowing. Unscaled, no Lipschitz guarantee. `1 - turbulenceBumpN(...)` gives a cheap "ridged" look for free (negation and a constant offset don't change the Lipschitz constant) — not worth a separate named function.
- **`turbulence3(size, gain, lacunarity)`** — turbulence3(size, gain, lacunarity): turbulence2, plus a third octave.
- **`turbulenceBump2(size, gain, lacunarity)`** — turbulenceBump2(size, gain, lacunarity): turbulence2, pre-scaled so it's safe to ADD directly to a distance field (bound: fbmBumpLipschitzBound2 — abs() doesn't change the Lipschitz constant).
- **`turbulenceBump3(size, gain, lacunarity)`** — turbulenceBump3(size, gain, lacunarity): turbulence3, pre-scaled (bound: fbmBumpLipschitzBound3).
- **`wavefield(freq)`** — wavefield(freq): a sum of 8 fixed-direction, fixed-phase cosine waves at frequency `freq`, normalized to roughly [-1,1]. Cheap (no hashing), and a genuinely different look from gradient/cellular noise — threshold it (texture.mask) for a spinodal-decomposition look, or use it directly as a broad, cheap color-blotch field. Unscaled, no Lipschitz guarantee. For a version safe to add directly to a distance field, see wavefieldBump.
- **`wavefieldLipschitzBound(freq)`** — wavefieldLipschitzBound(freq): the Lipschitz bound on the UNSCALED sum of 8 cosine waves (before wavefield's own /8 normalization) — each term dot(dir,p)*freq+phase has gradient magnitude <= freq (|dir| = 1), so the sum of 8 terms is bounded by 8*freq.
- **`wavefieldBump(freq)`** — wavefieldBump(freq): wavefield(freq), pre-scaled so it's safe to ADD directly to a distance field for raymarching (certified-by-arithmetic bound — see wavefieldLipschitzBound; cheap enough to drive displacement, unlike the fBm/turbulence families).
- **`cellularEdgesZBlend(size, zBandSize)`** — cellularEdgesZBlend(size, zBandSize): like cellularEdges, but the 2D feature-point hash is taken at two Z-bands (size zBandSize) and blended through Z, so the Voronoi cell-wall topology visibly reorganizes as Z advances — cells "change neighbors" partway through, evoking (not reproducing) the topology-changing mechanism behind real scutoids. ~0 along cell walls, larger inside cells; negate for a crack/web network that itself twists through depth.
- **`gradientLatticeVector(q)`** — gradientLatticeVector(q): three gradient noises at lattice point q for the cost of one gradientLattice: each corner's gradient is hashed once and dotted with the offset three ways, as it is and turned twice about its axes. The components are uncorrelated enough to warp a domain by, which otherwise takes three noises. Each is gradientLattice-like: 0 at every lattice vertex, typically within roughly [-1.5, 1.5].
- **`rehash(h)`** — rehash(h): three more uniform numbers in [0, 1) from a hash h's three (noiseHash3), for a cell that needs six: each is a mix of all three at scales far apart, taken mod 1, which is cheaper than hashing the cell again.
- **`colonyWallsLattice(q, size, spread, doubledShare, offset)`** — colonyWallsLattice(q, size, spread, doubledShare, offset): the walls between colonies that each start at a point of their own cell of the lattice, at a time of their own up to `spread` of a cell's growth late, and grow at one speed until they meet: a Johnson-Mehl tessellation (Johnson and Mehl 1939), whose walls never cross and meet three at a time, and whose cells vary in size with the spread (0 is plain Voronoi). q is in cells, and `size` is a cell's size in world units, per axis, so a stretched lattice grows faster along its long axis. A share `doubledShare` of the walls are doubled, each colony drawing its own `offset` world units inside its cell, the same for both of a pair. Returns [distance, start.x, start.y, start.z]: the distance to the nearest wall in world units, the gap between the two arrivals over how fast it grows, so walls are as wide whichever way they run; and where the point's own colony started, in cells. The nearest wall of all the owner's neighbours' (Quilez's Voronoi edges) keeps a wall whole where three meet. Warp q before calling to bend the walls. Draw them with texture.coverageLine.
- **`driftingCellsLattice(q, size, life, velocityFn)`** — driftingCellsLattice(q, size, life, velocityFn): cells that are born, grow, drift and die, for plates of crust, floes and bubbles. Each cell of the lattice holds one at a time, born at a point of its own and carried by the flow there, velocityFn(x) in world units per second at a world point x. It grows from nothing, fills its share of the lattice, and shrinks to nothing as its neighbours close over it, `life` seconds after its birth, when the next is born where it was. A cell's lifecycle is the Johnson-Mehl start time of colonyWallsLattice, running late, early and late again, so cells appear and vanish without a jump. Keep the flow's travel over one life under half a cell, velocity * life < size / 2, or cells will outrun the search. q is in cells and `size` a cell's size in world units, per axis. Returns two values: [distance, wall age, age, hash], the distance to the nearest wall in world units; how long the younger of the two cells beside that wall has lived, in seconds, so a new wall is a young one; how long this point's own cell has lived; and a hash in [0, 1) of that cell. And how far that cell has drifted, in world units: subtract it from p to carry a pattern along with the cell.
- **`driftingCellsLatticeXY(q, size, life, velocityFn)`** — driftingCellsLatticeXY(q, size, life, velocityFn): driftingCellsLattice in the plane of x and y, for plates, floes and bubbles on a level surface: the cells are columns along z, so every depth of a slab shows the same ones. It searches 9 cells where the 3D lattice searches 27, a third of the work and of the shader. The same arguments and return values, with q.z and the flow's z ignored.

### `matrix`

#### STRUCTURE

- **`transpose2(m)`** — transpose2(m) returns the transpose of a 2×2 matrix.
- **`transpose3(m)`** — transpose3(m) returns the transpose of a 3×3 matrix.
- **`transpose4(m)`** — transpose4(m) returns the transpose of a 4×4 matrix.

#### PRODUCTS

- **`multiply(a, b)`** — multiply(a, b) is the matrix product a·b. a is R×K, b is K×C, and the result is R×C. Each result column is a·(column of b), formed as the linear combination of a's columns weighted by that column's entries — so no transpose is needed and any rectangular size works. Conformability (a's column count == b's row count) is the caller's responsibility.
- **`outer(u, v)`** — outer(u, v) is the exterior (outer) product u·vᵀ. For a vecR u and a vecC v the result is the R×C matrix whose column j is v[j]·u.

#### ELEMENTWISE

- **`scale(s, m)`** — scale(s, m) multiplies every entry of m by the scalar s. (This is scalar scaling of a matrix; for a scaling-transform matrix see the constructors.)
- **`negate(m)`** — negate(m) negates every entry: −m. The unary `-` operator on a matrix routes here.
- **`plus(a, b)`** — plus(a, b) is the elementwise sum of two equally-shaped matrices. The `+` operator on matrices routes here.
- **`minus(a, b)`** — minus(a, b) is the elementwise difference a − b. The `-` operator on matrices routes here.

#### CONSTRUCTORS

- **`identity3()`** — identity3() is the 3×3 identity matrix.
- **`identity4()`** — identity4() is the 4×4 identity matrix.
- **`rotation3(direction, angle)`** — rotation3(direction, angle) is the 3×3 rotation by `angle` radians about `direction` (normalized internally), via Rodrigues' formula. Use this to rotate directions/normals; to rotate points about an arbitrary line use the 4×4 rotation below.
- **`rotation(axis_origin, axis_direction, angle)`** — rotation(axis_origin, axis_direction, angle) is the 4×4 affine transform that rotates by `angle` radians about the line through `axis_origin` with direction `axis_direction`. Points on that line are fixed. Apply it to points with vector.transformPoint.
- **`translation(t)`** — translation(t) is the 4×4 affine transform that translates by the vec3 t.
- **`scaling(factors)`** — scaling(factors) is the 4×4 transform that scales by the per-axis vec3 `factors` about the origin. Uniform scale: pass [k, k, k].
- **`scaling3(factors)`** — scaling3(factors) is the 3×3 linear scale by the per-axis vec3 `factors`.
- **`rotationEuler3(angles)`** — rotationEuler3(angles) is the 3×3 rotation for the vec3 of euler angles `angles`, matching map.eulerAngles: map.eulerAngles(v, a) = map.rotateXY(map.rotateZX(map.rotateYZ(v, a.x), a.y), a.z). So vector.multiplyVector(rotationEuler3(a), v) == map.eulerAngles(v, a).
- **`rotationEuler(angles)`** — rotationEuler(angles) is the 4×4 homogeneous rotation for the euler vec3 `angles` (about the origin), matching map.eulerAngles.
- **`rotation(spin, orient)`** — rotation(spin, orient) builds the 4×4 rotation matrix Rz(spin)·R(orient): the base euler orientation `orient` (vec3, map.eulerAngles convention) first, then a spin by `spin` radians about +Z. A real matrix product (not an euler-vector sum), closed form from scalars, so it is valid in BOTH the full build and the canvas evaluator. Pass it to mesh(index, position, rotation) to place a part: that overload drops the translation into the 4th column.
- **`compose(position, euler)`** — compose(position, euler) builds the 4×4 homogeneous placement T·R: rotate by the euler vec3 `euler` (map.eulerAngles convention) about the origin, then translate to the vec3 `position`. Closed form from scalars, so it folds in the canvas evaluator and compiles on every backend. Pass it to mesh(index, M) to place a part rigidly (no scale).
- **`compose(position, euler, scale)`** — compose(position, euler, scale) builds the 4×4 homogeneous placement T·R·S: scale by the per-axis vec3 `scale`, rotate by the euler vec3 `euler` (map.eulerAngles convention), then translate to the vec3 `position`. It takes the rigid T·R from the two-argument compose and scales each rotation column by the matching component of `scale` — right-multiplication by the diagonal scale matrix — so the two forms stay in lockstep.

#### DETERMINANT

- **`determinant2(m)`** — determinant2(m) is the determinant of a 2×2 matrix.
- **`determinant3(m)`** — determinant3(m) is the determinant of a 3×3 matrix (the scalar triple product of its columns).
- **`determinant4(m)`** — determinant4(m) is the determinant of a 4×4 matrix, by cofactor expansion along the first column over the four 3×3 minors.

#### INVERSE

- **`inverse2(m)`** — inverse2(m) is the inverse of a 2×2 (linear) matrix. Undefined (divides by zero) when determinant2(m) is 0.
- **`inverse3(m)`** — inverse3(m) is the inverse of a 3×3 (linear) matrix, built from the cross products of its columns (adjugate over determinant). Undefined when determinant3(m) is 0.
- **`inverse4(m)`** — inverse4(m) is the general inverse of an arbitrary 4×4 matrix, via the compact 2×2-minor (GLM) method: 18 paired sub-determinants assembled into cofactor columns, divided by the determinant. Works for affine AND projective matrices. Undefined when the determinant is 0. For a 4×4 you know to be a homogeneous affine transform, prefer `inverse` below — it is much cheaper.
- **`inverse(m)`** — inverse(m) inverts a 4×4 HOMOGENEOUS AFFINE transform (bottom row [0, 0, 0, 1]) far more cheaply than the general inverse4: it inverts the 3×3 linear block A and remaps the translation t to −A⁻¹·t, leaving the bottom row intact. Exact for any invertible affine transform — rotation, scale, shear, translation, and their compositions. NOT valid for projective matrices; use inverse4 for those.

#### SDF SCALE

- **`minScale(m)`** — minScale(m) returns the smallest per-axis scale of m's linear block — the length of its shortest column among the first three. After transforming an SDF's sample point by inverse(m), multiply the resulting distance by this factor to keep the field a conservative (non-overstepping) distance bound: `remap(shape(), vector.multiplyPoint(inverse(m), p)) * matrix.minScale(m)`. Exact for rotation and axis-aligned scale (the linear block's smallest singular value); for shear it is an approximation that may slightly overestimate distance. Works on a 3×3 or the linear block of a 4×4.

### `vector`

#### Constants

- **`zero = [0, 0, 0]`** — The origin, [0, 0, 0]. Also the zero vector for direction arguments.
- **`dirX = [1, 0, 0]`** — Unit vector along +X.
- **`dirY = [0, 1, 0]`** — Unit vector along +Y.
- **`dirZ = [0, 0, 1]`** — Unit vector along +Z.

#### PRODUCTS

- **`multiply(m, v)`** — multiply(m, v) is the matrix–vector product m·v. For an R×C matrix m and a vecC v it returns the vecR formed as the linear combination of m's columns weighted by the components of v. (Same lowering as matrix.multiply, one column of b.)
- **`multiplyPoint(m, point)`** — multiplyPoint(m, point) applies the 4×4 homogeneous transform m to a vec3 POINT — i.e. with w = 1, so translation is included — and returns the vec3 result. To map a world point into a shape's local frame, pass matrix.inverse(transform): `remap(shape(), vector.multiplyPoint(inv, p))`.
- **`multiplyVector(m, dir)`** — multiplyVector(m, dir) applies the 4×4 homogeneous transform m to a vec3 DIRECTION — i.e. with w = 0, so translation is ignored — and returns the vec3 result. This is correct for tangents/ray directions; transforming a surface NORMAL under a non-rigid transform instead needs the inverse-transpose of the linear block.

#### ARITHMETIC OPERATIONS

- **`addVector(a, b)`** — addVector(a, b) computes the componentwise sum of two vectors.
- **`addScalar(v, s)`** — addScalar(v, s) adds a scalar to each component of vector v.
- **`subtractVector(a, b)`** — subtractVector(a, b) computes the componentwise difference a - b.
- **`subtractScalarFrom(v, s)`** — subtractScalarFrom(v, s) subtracts scalar s from each component of vector v.
- **`subtractFromScalar(s, v)`** — subtractFromScalar(s, v) computes the componentwise difference s - v.
- **`multiplyScalar(v, s)`** — multiplyScalar(v, s) multiplies each component of vector v by scalar s.
- **`multiplyComponents(a, b)`** — multiplyComponents(a, b) computes the elementwise product of two vectors.
- **`divideByScalar(v, s)`** — divideByScalar(v, s) divides each component of vector v by scalar s.
- **`divideScalarBy(s, v)`** — divideScalarBy(s, v) divides scalar s by each component of vector v.
- **`divideComponents(a, b)`** — divideComponents(a, b) divides vector a by vector b elementwise.
- **`negate(v)`** — negate(v) negates each component of vector v.

### `color`

#### Constants

- **`pi = atan(1) * 4`**
- **`tau = pi * 2`**
- **`black = [0, 0, 0]`** — Pure black.
- **`darkGray = [0.25, 0.25, 0.25]`** — 25% gray.
- **`gray = [0.5, 0.5, 0.5]`** — 50% gray.
- **`lightGray = [0.75, 0.75, 0.75]`** — 75% gray.
- **`white = [1, 1, 1]`** — Pure white.
- **`red = [1, 0, 0]`** — Primary red.
- **`green = [0, 1, 0]`** — Primary green.
- **`blue = [0, 0, 1]`** — Primary blue.
- **`yellow = [1, 1, 0]`** — Yellow (red + green).
- **`cyan = [0, 1, 1]`** — Cyan (green + blue).
- **`magenta = [1, 0, 1]`** — Magenta (red + blue).
- **`cherry = [0.871, 0.192, 0.388]`** — Cherry red, #de3163.
- **`orange = hsv2rgb([pi / 6, 1, 1])`** — Orange, at hue π/6.
- **`olive = hsv2rgb([pi / 3, 1, 0.5])`** — Olive, #808000: hue π/3 at half value.
- **`teal = hsv2rgb([pi, 1, 0.5])`** — Teal, #008080: hue π at half value.
- **`violet = hsv2rgb([3 * pi / 2, 1, 1])`** — Violet, at hue 3π/2.
- **`rose = hsv2rgb([11 * pi / 6, 1, 1])`** — Rose, at hue 11π/6.
- **`viridisData = [`**
- **`magmaData = [`**
- **`infernoData = [`**
- **`plasmaData = [`**
- **`turboData = [`**
- **`cividisData = [`**
- **`coolwarmData = [[0.230, 0.299, 0.754], [0.865, 0.865, 0.865], [0.706, 0.016, 0.150]]`**
- **`jetData = [`**
- **`hotData = [`**
- **`coolData = [[0.0, 1.0, 1.0], [1.0, 0.0, 1.0]]`**
- **`boneData = [`**
- **`copperData = [[0.0, 0.0, 0.0], [1.0, 0.7812, 0.4975]]`**
- **`spectralData = [`**
- **`rdBuData = [`**
- **`brBgData = [`**
- **`twilightData = [`**
- **`rainbowData = [[1, 0, 0], [1, 1, 0], [0, 1, 0], [0, 1, 1], [0, 0, 1], [0.5, 0, 1]]`**
- **`cyclicRainbowData = [`**

#### Helpers

- **`cbrt(x)`** — Real cube root, valid for negative x.
- **`luminance(c)`** — Relative luminance (Rec. 709 weights) of an sRGB color, in [0, 1]. Uses the gamma-encoded channels — a quick perceptual brightness, not the exact linear-light luminance (for that, take luminance of linearize(c)).
- **`clampColor(c)`** — Clamp a color to the displayable sRGB gamut [0, 1] per channel.
- **`srgbToLinearChannel(s)`** — Convert one sRGB channel in [0, 1] to linear light.
- **`linearToSrgbChannel(l)`** — Convert one linear-light channel in [0, 1] to sRGB.
- **`linearize(c)`** — Convert an sRGB color to linear-light RGB.
- **`gammaEncode(c)`** — Convert a linear-light RGB color back to sRGB.

#### HSV <-> RGB  (hue in radians)

- **`hsv2rgb(c)`** — HSV [hue, saturation, value] (sat/value in [0, 1], hue in radians) to sRGB.
- **`rgb2hsv(c)`** — sRGB to HSV [hue, saturation, value], hue in radians.

#### HSL <-> RGB  (hue in radians)

- **`hsl2rgb(c)`** — HSL [hue, saturation, lightness] (sat/lightness in [0, 1], hue in radians) to sRGB.
- **`rgb2hsl(c)`** — sRGB to HSL [hue, saturation, lightness], hue in radians.
- **`rgb2oklab(c)`** — sRGB to Oklab [L, a, b]. L is perceptual lightness in [0, 1]; a, b are the green–red and blue–yellow axes (~[-0.4, 0.4]).
- **`oklab2rgb(c)`** — Oklab [L, a, b] to sRGB. Out-of-gamut results are not clamped — wrap in clampColor if you need a displayable color.
- **`oklab2oklch(c)`** — Oklab [L, a, b] to OKLCH [L, chroma, hue] (hue in radians, [0, 2π)).
- **`oklch2oklab(c)`** — OKLCH [L, chroma, hue] (hue in radians) to Oklab [L, a, b].
- **`rgb2oklch(c)`** — sRGB to OKLCH [L, chroma, hue], hue in radians.
- **`oklch2rgb(c)`** — OKLCH [L, chroma, hue] (hue in radians) to sRGB.

#### Adjustments & mixing (perceptual, via Oklab / OKLCH)

- **`mixOklab(a, b, t)`** — Perceptually-uniform mix of two sRGB colors: a at t = 0, b at t = 1, interpolated in Oklab. Smoother and more even than a raw RGB lerp.
- **`lighten(c, amt)`** — Lighten an sRGB color by adding amt to its OKLCH lightness (amt in [-1, 1]). Negative amt darkens.
- **`darken(c, amt)`** — Darken an sRGB color by amt (see lighten).
- **`saturate(c, factor)`** — Scale an sRGB color's OKLCH chroma by factor (1 = unchanged, 0 = gray, >1 = more saturated).
- **`desaturate(c)`** — Fully desaturate an sRGB color to its perceptual gray (OKLCH chroma 0), preserving perceived lightness.
- **`hueShift(c, radians)`** — Rotate an sRGB color's hue by radians (in OKLCH, so lightness/chroma hold).
- **`complement(c)`** — Complementary color: hue rotated a half turn (π radians).
- **`grayscale(c)`** — Perceptual grayscale: the color's OKLCH lightness as a neutral sRGB gray.
- **`invert(c)`** — Invert (photographic negative) in sRGB.

#### Palette

- **`sampleMap(palette, t)`** — Sample a colormap (a list of sRGB control colors, evenly spaced over [0, 1]) at parameter t, with piecewise-linear interpolation between the two nearest control points. t is clamped to [0, 1]. Works for any length list.
- **`interp(value, inMin, inMax, cmapFn)`** — Sample a colormap over a value range: remap value from [inMin, inMax] to [0, 1] (clamped) and read cmapFn there. The vec3 analog of ease.interp — pass any colormap FUNCTION, e.g. color.interp(temp, 0, 100, color.viridis).
- **`easedMap(t, cmapFn, easeFn)`** — Read a colormap through an easing curve: cmapFn(easeFn(t)) — reshape how the colormap sweeps, e.g. color.easedMap(t, color.viridis, ease.inCubic). Both arguments are functions, so the color and ease libraries compose directly.
- **`viridis(t)`** — Viridis: perceptually-uniform sequential (dark blue → green → yellow).
- **`magma(t)`** — Magma: sequential (black → purple → orange → cream).
- **`inferno(t)`** — Inferno: sequential (black → red → yellow).
- **`plasma(t)`** — Plasma: sequential (blue → magenta → yellow).
- **`turbo(t)`** — Turbo: an improved rainbow (blue → green → yellow → red).
- **`cividis(t)`** — Cividis: colorblind-friendly sequential (blue → gray → yellow).
- **`coolwarm(t)`** — Cool–warm diverging map (blue → gray → red), centered at t = 0.5.
- **`grayRamp(t)`** — Grayscale ramp (black → white).
- **`jet(t)`** — Jet: the legacy rainbow (blue → cyan → green → yellow → red). High contrast but NOT perceptually uniform — prefer turbo or viridis for real data.
- **`hot(t)`** — Hot: thermal (black → red → yellow → white).
- **`cool(t)`** — Cool: cyan → magenta.
- **`bone(t)`** — Bone: blue-tinted grayscale (black → white).
- **`copper(t)`** — Copper: black → metallic copper.
- **`spectral(t)`** — Spectral: diverging rainbow (red → pale yellow → blue).
- **`rdBu(t)`** — Red–Blue diverging (red → white → blue).
- **`brBg(t)`** — Brown–teal diverging (brown → white → teal).
- **`rainbow(t)`** — Rainbow: full-saturation hue sweep, red → violet.
- **`twilight(t)`** — Twilight: perceptual CYCLIC map (endpoints match) — good for phase/angle.
- **`cyclicRainbow(t)`** — Cyclic rainbow: the full hue wheel; t = 0 and t = 1 both land on red — for angle/phase data.
- **`cubehelix(t)`** — Cubehelix (Green 2011): a perceptually monotonic sequential map that also prints legibly in grayscale, traced by an analytic helix through RGB — no control points. Fixed default parameters (start 0.5, −1.5 rotations).

### `material`

#### Constants

- **`standard = make(0.5, 0, 1.5)`** — The material a shape has when it sets none: roughness 0.5, not metallic, ior 1.5.
- **`matte = make(1.0, 0, 1.45)`** — Fully diffuse: chalk, unglazed clay, flat paint.
- **`rubber = make(0.8, 0, 1.52)`** — Soft, low-gloss: elastomers, grips, tyres.
- **`plastic = make(0.35, 0, 1.46)`** — Moulded plastic with a satin sheen: ABS, nylon, PLA.
- **`shinyPlastic = make(0.1, 0, 1.46)`** — Glossy plastic: polished acrylic, phone cases, car paint before clearcoat.
- **`brushedMetal = make(0.45, 1, 1.5)`** — Brushed or bead-blasted metal, directionless satin. Pair with a colour from the metal library.
- **`polishedMetal = make(0.15, 1, 1.5)`** — Polished metal with soft reflections.
- **`chrome = make(0.04, 1, 1.5)`** — Mirror finish.

#### Constructors

- **`make(rough, metal, index)`** — A material from its three surface numbers: roughness and metallic in [0, 1], ior >= 1. No emission, opaque, undisplaced.
- **`glowing(rgb, strength)`** — A self-lit material: rgb is the emitted colour (sRGB), strength its brightness (1 = white paper, larger values glow). Otherwise `standard`.
- **`translucent(m, transmission, absorptionDistance)`** — A see-through version of m: transmission in [0, 1] is the fraction of light passing through, absorptionDistance (mm) is the depth at which the shape's colour is what comes through (0 = no tint). The glass library has presets.

#### Accessors

- **`roughness(m)`** — Roughness of m, in [0, 1].
- **`metallic(m)`** — Metallic of m, in [0, 1].
- **`ior(m)`** — Index of refraction of m.
- **`emissive(m)`** — Emitted colour of m (sRGB vec3).
- **`emissiveStrength(m)`** — Emissive strength of m.
- **`transmission(m)`** — Transmission of m, in [0, 1].
- **`absorptionDistance(m)`** — Absorption distance of m, in mm.
- **`displacement(m)`** — Displacement of m, in mm.
- **`lay(m)`** — The direction of m's lay (vec3), the way its grooves run; see withLay.
- **`anisotropy(m)`** — Anisotropy of m, in [0, 1]; see withLay.

#### Modifiers

- **`withRoughness(m, rough)`** — m with its roughness replaced.
- **`withMetallic(m, metal)`** — m with its metallic replaced.
- **`withIor(m, index)`** — m with its index of refraction replaced.
- **`withEmissive(m, rgb, strength)`** — m with an emitted colour (sRGB) and strength.
- **`withTransmission(m, transmission, absorptionDistance)`** — m with its transmission and absorption distance replaced; same as `translucent`.
- **`withDisplacement(m, height)`** — m with a displacement height (mm).
- **`withLay(m, direction, anisotropy)`** — m with a lay: grooves too fine to see, running along `direction` (a vector in world space; only its direction across the surface matters), that stretch the highlight across them, as on brushed, turned or ground metal. anisotropy, 0 to 1, sets how much: the lobe's width, roughness^2, is multiplied by (1 - 0.9 anisotropy)^0.5 along the lay and divided by it across (Burley 2012), so 0 is m's own round highlight. Where the lay meets the surface head on, the highlight stays round. The lay may vary over the surface: a turned part's runs round its axis, cross(axis, p - centre).
- **`blend(a, b, t)`** — Linear blend of two materials, row by row: a at t = 0, b at t = 1. One field can drive every property at once, e.g. `part.material = material.blend(material.polishedMetal, material.matte, wear);`

#### Surface composition

- **`filmKeeps(wavelength, build)`** — filmKeeps(wavelength, build): the share of a relief `wavelength` mm across that a film `build` mm thick leaves. A film levels what is finer than it is thick, so a 0.08 mm film of polyurethane fills sanding scratches, half fills pores a millimetre across, and leaves saw marks two millimetres apart as they were. One share governs all three ways a film reaches the lobe: it multiplies a relief's amplitude, squares into its slope variance, and multiplies a lay's anisotropy. Judged.
- **`hiddenShare(fade, standIn)`** — hiddenShare(fade, standIn): how much of a relief is hidden, to be folded into the lobe as roughness rather than drawn as displacement: all of it where the renderer draws no displacement, and all of it for a stand-in (standIn 1), which has no surface to draw on. `fade` is how far the relief is too small to see, as texture.fadeDetail(size) gives it; it comes in as an argument because texture imports this library, so this one cannot call it back.
- **`roughnessFromGloss(gloss)`** — roughnessFromGloss(gloss): the roughness whose 60 degree gloss is `gloss` gloss units, as a finish's data sheet gives its sheen: matte under 5, satin 20 to 35, semi-gloss 35 to 70, gloss 70 to 85, high gloss above (MPI). The share of a GGX surface's reflection (alpha = roughness squared, index 1.5) that falls in a gloss meter's 4.4 by 11.7 degree receptor, against black glass of index 1.567 at 100 (ISO 2813), integrated numerically and fitted to within 0.02 from 1 to 88 units. A smooth coat reads about 89.

### `metal`

#### Ferrous

- **`steel()`** — Mild steel, machined satin. Returns material and colour.
- **`stainless()`** — Stainless steel, brushed. Its colour is a touch brighter and cooler than mild steel's. Returns material and colour.
- **`castIron()`** — Cast iron, rough as-cast surface. Returns material and colour.

#### Light metals

- **`aluminum()`** — Aluminium, machined satin. Returns material and colour.
- **`titanium()`** — Titanium, bead-blasted, a warm gray. Returns material and colour.
- **`zinc()`** — Zinc, as a galvanised or plated dull finish, a cool bluish gray. Returns material and colour.

#### Copper alloys and precious

- **`brass()`** — Brass, polished. Returns material and colour.
- **`copper()`** — Copper, polished. Returns material and colour.
- **`gold()`** — Gold, polished. Returns material and colour.
- **`chromium()`** — Chromium plate, mirror finish. Returns material and colour.

### `glass`

- **`make(rough, index, absorptionDistance)`** — A fully transmissive material: roughness in [0, 1], ior >= 1 (1.5 is glass), absorptionDistance in mm (the depth at which the shape's colour is what comes through; 0 = no tint). Same row layout material.make produces, spelled out so the library needs no imports.

#### Glass

- **`clear()`** — Clear polished glass. Returns material and colour.
- **`frosted()`** — Frosted glass: sand-blasted or acid-etched, blurs what is behind it. Returns material and colour.
- **`smoked()`** — Smoked glass, neutral grey tint over 6 mm. Returns material and colour.
- **`amber()`** — Amber glass, the bottle tint, over 5 mm. Returns material and colour.

#### Plastics and liquids

- **`acrylic()`** — Cast acrylic (PMMA), clearer than glass with a lower index. Returns material and colour.
- **`polycarbonate()`** — Polycarbonate, the high-index engineering plastic, with a faint haze. Returns material and colour.
- **`resin()`** — Clear SLA resin, a satin finish with a pale amber tint over 8 mm. Returns material and colour.
- **`water()`** — Water, with a faint cyan tint over 200 mm. Returns material and colour.

### `physical`

#### Constants

- **`diamondDensity = 3500`** — Density of diamond, kg/m^3.
- **`glassBorosilicateDensity = 2230`** — Density of glass (Borosilicate), kg/m^3.
- **`glassSodaLimeDensity = 2520`** — Density of glass (Soda-lime), kg/m^3.
- **`iceDensity = 917`** — Density of ice, kg/m^3.
- **`marbleDensity = 2700`** — Density of marble, kg/m^3.
- **`quartzDensity = 2600`** — Density of quartz, kg/m^3.
- **`saltDensity = 2170`** — Density of salt, kg/m^3.
- **`sandDensity = 1500`** — Density of sand, kg/m^3.
- **`sapphireDensity = 3980`** — Density of sapphire, kg/m^3.
- **`snowDensity = 300`** — Density of snow, kg/m^3.
- **`bloodDeoxygenatedDensity = 1060`** — Density of blood (Deoxygenated), kg/m^3.
- **`bloodOxygenatedDensity = 1060`** — Density of blood (Oxygenated), kg/m^3.
- **`skinIDensity = 1020`** — Density of skin I, kg/m^3.
- **`skinIIDensity = 1020`** — Density of skin II, kg/m^3.
- **`skinIIIDensity = 1020`** — Density of skin III, kg/m^3.
- **`skinIVDensity = 1020`** — Density of skin IV, kg/m^3.
- **`skinVDensity = 1020`** — Density of skin V, kg/m^3.
- **`skinVIDensity = 1020`** — Density of skin VI, kg/m^3.
- **`beerPaleLagerDensity = 1000`** — Density of beer (Pale Lager), kg/m^3.
- **`coffeeDensity = 1020`** — Density of coffee, kg/m^3.
- **`colaDensity = 1000`** — Density of cola, kg/m^3.
- **`cookingOilDensity = 920`** — Density of cooking Oil, kg/m^3.
- **`creamDensity = 1012`** — Density of cream, kg/m^3.
- **`ethanolDensity = 789`** — Density of ethanol, kg/m^3.
- **`gasolineDensity = 770`** — Density of gasoline, kg/m^3.
- **`honeyLiquidDensity = 1400`** — Density of honey (Liquid), kg/m^3.
- **`juiceAppleDensity = 1000`** — Density of juice (Apple), kg/m^3.
- **`juiceCranberryDensity = 1000`** — Density of juice (Cranberry), kg/m^3.
- **`juiceGrapeDensity = 1000`** — Density of juice (Grape), kg/m^3.
- **`juiceRubyGrapefruitDensity = 1000`** — Density of juice (Ruby Grapefruit), kg/m^3.
- **`juiceWhiteGrapefruitDensity = 1000`** — Density of juice (White Grapefruit), kg/m^3.
- **`milkDensity = 1030`** — Density of milk, kg/m^3.
- **`petroleumDensity = 1000`** — Density of petroleum, kg/m^3.
- **`vodka40Density = 935`** — Density of vodka (40%), kg/m^3.
- **`waterDensity = 1000`** — Density of water, kg/m^3.
- **`wineRedDensity = 1000`** — Density of wine (Red), kg/m^3.
- **`wineWhiteDensity = 1000`** — Density of wine (White), kg/m^3.
- **`brickDensity = 2000`** — Density of brick, kg/m^3.
- **`cardboardDensity = 700`** — Density of cardboard, kg/m^3.
- **`castIronDensity = 7300`** — Density of cast Iron, kg/m^3.
- **`clayDensity = 812`** — Density of clay, kg/m^3.
- **`concreteDensity = 2400`** — Density of concrete, kg/m^3.
- **`officePaperDensity = 800`** — Density of office Paper, kg/m^3.
- **`polystyreneFoamDensity = 75`** — Density of polystyrene (Foam), kg/m^3.
- **`porcelainDensity = 2400`** — Density of porcelain, kg/m^3.
- **`terracottaDensity = 2000`** — Density of terracotta, kg/m^3.
- **`toiletPaperDensity = 50`** — Density of toilet Paper, kg/m^3.
- **`toothpasteDensity = 1400`** — Density of toothpaste, kg/m^3.
- **`aluminumDensity = 2700`** — Density of aluminum, kg/m^3.
- **`aluminumAnodizedRedDensity = 2700`** — Density of aluminum (Anodized Red), kg/m^3.
- **`berylliumDensity = 1850`** — Density of beryllium, kg/m^3.
- **`brassDensity = 8600`** — Density of brass, kg/m^3.
- **`cesiumDensity = 1886`** — Density of cesium, kg/m^3.
- **`chromiumDensity = 7200`** — Density of chromium, kg/m^3.
- **`cobaltDensity = 8900`** — Density of cobalt, kg/m^3.
- **`copperDensity = 8940`** — Density of copper, kg/m^3.
- **`germaniumDensity = 5327`** — Density of germanium, kg/m^3.
- **`goldDensity = 19320`** — Density of gold, kg/m^3.
- **`iridiumDensity = 22562`** — Density of iridium, kg/m^3.
- **`ironDensity = 7870`** — Density of iron, kg/m^3.
- **`leadDensity = 11340`** — Density of lead, kg/m^3.
- **`lithiumDensity = 535`** — Density of lithium, kg/m^3.
- **`magnesiumDensity = 1737`** — Density of magnesium, kg/m^3.
- **`manganeseDensity = 7476`** — Density of manganese, kg/m^3.
- **`mercuryDensity = 13546`** — Density of mercury, kg/m^3.
- **`molybdenumDensity = 10223`** — Density of molybdenum, kg/m^3.
- **`nickelDensity = 8900`** — Density of nickel, kg/m^3.
- **`palladiumDensity = 12007`** — Density of palladium, kg/m^3.
- **`platinumDensity = 21450`** — Density of platinum, kg/m^3.
- **`potassiumDensity = 859`** — Density of potassium, kg/m^3.
- **`rubidiumDensity = 1534`** — Density of rubidium, kg/m^3.
- **`siliconDensity = 2330`** — Density of silicon, kg/m^3.
- **`silverDensity = 10500`** — Density of silver, kg/m^3.
- **`sodiumDensity = 969`** — Density of sodium, kg/m^3.
- **`stainlessSteelDensity = 8000`** — Density of stainless Steel, kg/m^3.
- **`titaniumDensity = 4540`** — Density of titanium, kg/m^3.
- **`titaniumAnodizedDensity = 4540`** — Density of titanium (Anodized), kg/m^3.
- **`tungstenDensity = 19300`** — Density of tungsten, kg/m^3.
- **`vanadiumDensity = 6100`** — Density of vanadium, kg/m^3.
- **`zincDensity = 7000`** — Density of zinc, kg/m^3.
- **`amberDensity = 1080`** — Density of amber, kg/m^3.
- **`boneDensity = 1500`** — Density of bone, kg/m^3.
- **`charcoalDensity = 200`** — Density of charcoal, kg/m^3.
- **`chocolateDensity = 1300`** — Density of chocolate, kg/m^3.
- **`honeyCrystallizedDensity = 1400`** — Density of honey (Crystallized), kg/m^3.
- **`ketchupDensity = 1100`** — Density of ketchup, kg/m^3.
- **`pearlDensity = 2700`** — Density of pearl, kg/m^3.
- **`plasticAcrylicDensity = 1180`** — Density of plastic (Acrylic), kg/m^3.
- **`plasticPCDensity = 1200`** — Density of plastic (PC), kg/m^3.
- **`plasticPETDensity = 1380`** — Density of plastic (PET), kg/m^3.
- **`plasticPLADensity = 1320`** — Density of plastic (PLA), kg/m^3.
- **`plasticPPDensity = 900`** — Density of plastic (PP), kg/m^3.
- **`plasticPURDensity = 1150`** — Density of plastic (PUR), kg/m^3.
- **`plasticPVCDensity = 1300`** — Density of plastic (PVC), kg/m^3.

#### Crystal

- **`diamond()`** — Diamond. Source: Refractive index database. Returns material and colour.
- **`glassBorosilicate()`** — Glass (Borosilicate). Borosilicate glass has numerous advantages over regular soda-lime glass, including having a higher clarity (transparency). Borosilicate is generally slightly more expensive than soda-lime glass. Source: Refractive index database. Returns material and colour.
- **`glassSodaLime()`** — Glass (Soda-lime). Soda-lime glass is the most prevelant type of glass made, and accounts for about 90% of manufactured glass. It's used for windowpanes and glass containers (bottles and jars) for beverages, food, and some commodity items. Source: Refractive index database. Returns material and colour.
- **`ice()`** — Ice. Source: Refractive index database. Returns material and colour.
- **`marble()`** — Marble. Source: A Practical Model for Subsurface Light Transport. Returns material and colour.
- **`quartz()`** — Quartz. Source: Quartz. Returns material and colour.
- **`salt()`** — Salt. Source: Refractive index database. Returns material and colour.
- **`sand()`** — Sand. Source: Feeding a Physically Based Shading Model. Returns material and colour.
- **`sapphire()`** — Sapphire. Source: Refractive index database. Returns material and colour.
- **`snow()`** — Snow. Source: Feeding a Physically Based Shading Model. Returns material and colour.

#### Human

- **`bloodDeoxygenated()`** — Blood (Deoxygenated). The color of blood depends on its oxygen level. When hemoglobin has a full oxygen load, it has a bright red color, and when it's deoxygenated it's a darker red. For example, arterial and capillary blood is bright, and venous blood is dark. Upon exposure to air, deoxygenated blood quickly oxygenates and turns a brighter shade of red. Source: Blood. Returns material and colour.
- **`bloodOxygenated()`** — Blood (Oxygenated). The color of blood depends on its oxygen level. When hemoglobin has a full oxygen load, it has a bright red color, and when it's deoxygenated it's a darker red. For example, arterial and capillary blood is bright, and venous blood is dark. Upon exposure to air, deoxygenated blood quickly oxygenates and turns a brighter shade of red. Source: Blood. Returns material and colour.
- **`eyeCornea()`** — Eye (Cornea). Source: Cornea. Returns material and colour.
- **`eyeLens()`** — Eye (Lens). The refractive index of human lens varies from approximately 1.406 in the central layers down to 1.386 in less dense layers of the lens. This index gradient enhances the optical power of the lens. Source: Lens (vertebrate anatomy). Returns material and colour.
- **`eyeSclera()`** — Eye (Sclera). Source: Optical properties of human sclera in spectral range 370-2500 nm. Returns material and colour.
- **`skinI()`** — Skin I. Source: A Practical Model for Subsurface Light Transport. Returns material and colour.
- **`skinII()`** — Skin II. Source: A Practical Model for Subsurface Light Transport. Returns material and colour.
- **`skinIII()`** — Skin III. Source: A Practical Model for Subsurface Light Transport. Returns material and colour.
- **`skinIV()`** — Skin IV. Source: A Practical Model for Subsurface Light Transport. Returns material and colour.
- **`skinV()`** — Skin V. Source: A Practical Model for Subsurface Light Transport. Returns material and colour.
- **`skinVI()`** — Skin VI. Source: A Practical Model for Subsurface Light Transport. Returns material and colour.

#### Liquid

- **`beerPaleLager()`** — Beer (Pale Lager). While the index of refraction of beer depends on its alcohol and sugar contents, beer contains 90%-95% water, so the index of refraction can in most cases be assumed to be close to that of water. Source: Acquiring Scattering Properties Of Participating Media By Dilution. Returns material and colour.
- **`coffee()`** — Coffee. Source: Photo Reference. Returns material and colour.
- **`cola()`** — Cola. Source: Acquiring Scattering Properties Of Participating Media By Dilution. Returns material and colour.
- **`cookingOil()`** — Cooking Oil. Source: Index of Refraction of Vegetable Oil. Returns material and colour.
- **`cream()`** — Cream. Source: A Practical Model for Subsurface Light Transport. Returns material and colour.
- **`ethanol()`** — Ethanol. Ethanol is also called ethyl alcohol, grain alcohol, drinking alcohol, or simply alcohol. The refractive index of an alcoholic beverage such as vodka depends on the ratio of alcohol and water. For example, vodka with 40% alcohol by volume contains 40% ethanol and 60% water. Source: Ethanol. Returns material and colour.
- **`gasoline()`** — Gasoline. Source: Gasoline. Returns material and colour.
- **`honeyLiquid()`** — Honey (Liquid). Variations in its water content alter its refractive index. Typically, the refractive index for honey ranges from 1.504 at 13% water content to 1.474 at 25%. Fresh honey is liquid and clear but with time it crystallizes and becomes thicker and less clear. The crystallization depends on temperature and the ratio of glucose and fructose. Source: Honey. Returns material and colour.
- **`juiceApple()`** — Juice (Apple). Source: Acquiring Scattering Properties Of Participating Media By Dilution. Returns material and colour.
- **`juiceCranberry()`** — Juice (Cranberry). Source: Acquiring Scattering Properties Of Participating Media By Dilution. Returns material and colour.
- **`juiceGrape()`** — Juice (Grape). Source: Acquiring Scattering Properties Of Participating Media By Dilution. Returns material and colour.
- **`juiceRubyGrapefruit()`** — Juice (Ruby Grapefruit). Source: Acquiring Scattering Properties Of Participating Media By Dilution. Returns material and colour.
- **`juiceWhiteGrapefruit()`** — Juice (White Grapefruit). Source: Acquiring Scattering Properties Of Participating Media By Dilution. Returns material and colour.
- **`milk()`** — Milk. Cow's milk Source: A Practical Model for Subsurface Light Transport. Returns material and colour.
- **`petroleum()`** — Petroleum. The viscosity of crude oil is sensitive to temperature. Source: Petroleum. Returns material and colour.
- **`vodka40()`** — Vodka (40%). The refractive index of vodka depends on the ratio of alcohol and water. Vodka with 40% alcohol by volume contains 40% ethanol and 60% water. Source: Vodka. Returns material and colour.
- **`water()`** — Water. While relatively small quantities of water appear to be colorless, pure water has a slight blue color that becomes deeper as the thickness of the observed sample increases beyond 50 cm. The hue of water is an intrinsic property and is caused by selective absorption and scattering of white light. Dissolved elements or suspended impurities may give water a different color. Source: Refractive index database. Returns material and colour.
- **`wineRed()`** — Wine (Red). Source: Acquiring Scattering Properties Of Participating Media By Dilution. Returns material and colour.
- **`wineWhite()`** — Wine (White). Source: Acquiring Scattering Properties Of Participating Media By Dilution. Returns material and colour.

#### Manmade

- **`asphaltFresh()`** — Asphalt (Fresh). Asphalt consists of crushed stone, gravel, and sand, bound together with bitumen. Source: Spectral Materials Database. Returns material and colour.
- **`blackboard()`** — Blackboard. Source: Spectral Materials Database. Returns material and colour.
- **`brick()`** — Brick. Source: Spectrophotometer Measurement. Returns material and colour.
- **`carPaint()`** — Car Paint. Modern car paints are nearly always an acrylic polyurethane "enamel" with a pigmented basecoat and a clear topcoat. It may be described as "acrylic", "acrylic enamel", "urethane", etc. and the clearcoat in particular may be described as a lacquer. The size of the metal flakes in metallic paints is typically between 100 to 400 microns in diameter. Modern automobile paint is applied in several layers, with a total thickness of around 100 μm (0.1mm). Source: Automotive Paint. Returns material and colour.
- **`cardboard()`** — Cardboard. The material commonly referred to as "cardboard" is more specifically named "corrugated fiberboard". Source: Spectrophotometer Measurement. Returns material and colour.
- **`castIron()`** — Cast Iron. Source: Spectrophotometer Measurement. Returns material and colour.
- **`clay()`** — Clay. Source: Spectrophotometer Measurement. Returns material and colour.
- **`concrete()`** — Concrete. Source: Feeding a Physically Based Shading Model. Returns material and colour.
- **`grayCard()`** — Gray Card. In photography and computer graphics, middle gray is a tone that is perceptually about halfway between black and white on a lightness scale. Light meters and cameras are often calibrated using an 18% gray card or a color reference card such as a ColorChecker. It is typically defined as 18% reflectance in visible light. Source: Gray Card. Returns material and colour.
- **`mitBlack()`** — MIT Black. Source: MIT engineers develop “blackest black” material to date. Returns material and colour.
- **`musouBlack()`** — Musou Black. Musou Black is one of the darkest paints available. Darker materials can be achieved by using carbon nanotube technology, such as the “MIT Black” material. Source: Spectrophotometer Measurement. Returns material and colour.
- **`officePaper()`** — Office Paper. Office paper contains optical brightening additives that absorb light in the ultraviolet and violet region (usually 340-370 nm), and re-emit light in the blue region (typically 420-470 nm) by fluorescence. Source: Spectrophotometer Measurement. Returns material and colour.
- **`polystyreneFoam()`** — Polystyrene (Foam). Source: Spectral Materials Database. Returns material and colour.
- **`porcelain()`** — Porcelain. Source: Spectrophotometer Measurement. Returns material and colour.
- **`rust()`** — Rust. When iron reacts with oxygen it transforms into a non-metal substance called iron oxide. Source: Spectrophotometer Measurement. Returns material and colour.
- **`spectralon()`** — Spectralon. Spectralon is a fluoropolymer that has the highest diffuse reflectance of any known material or coating over the ultraviolet, visible, and near-infrared regions of the spectrum. It is the whitest substance available and reflects 99% of the light. Source: Spectralon. Returns material and colour.
- **`terracotta()`** — Terracotta. Source: Spectrophotometer Measurement. Returns material and colour.
- **`tire()`** — Tire. Car tire. Source: Spectrophotometer Measurement. Returns material and colour.
- **`toiletPaper()`** — Toilet Paper. Source: Spectrophotometer Measurement. Returns material and colour.
- **`tonerBlack()`** — Toner (Black). Black toner printed on office paper with linear sRGB values of 0, 0, 0. Source: Spectrophotometer Measurement. Returns material and colour.
- **`toothpaste()`** — Toothpaste. Source: Spectrophotometer Measurement. Returns material and colour.
- **`whiteboard()`** — Whiteboard. Source: Spectral Materials Database. Returns material and colour.

#### Metal

- **`aluminum()`** — Aluminum. Source: Refractive index database. Returns material and colour.
- **`aluminumAnodizedRed()`** — Aluminum (Anodized Red). Aluminum can be colorized by a process called anodization. It creates a layer of aluminum oxide by supplying different voltages which results in different thickness. Unlike titanium, aluminum doesn't get vivid colors from anodization so dyes are used to color it further. Source: Refractive index database. Returns material and colour.
- **`beryllium()`** — Beryllium. Source: Refractive index database. Returns material and colour.
- **`brass()`** — Brass. Brass is an alloy of copper and zinc, with a bright and gold-like appearance. It is similar to bronze, another alloy containing copper that uses tin instead of zinc. Historically, the distinction between the two alloys has been less consistent and clear, and modern practice in museums and archaeology increasingly avoids both terms for historical objects in favor of the more general "copper alloy". Source: Refractive index database. Returns material and colour.
- **`cesium()`** — Cesium. Source: Refractive index database. Returns material and colour.
- **`chromium()`** — Chromium. Source: Refractive index database. Returns material and colour.
- **`cobalt()`** — Cobalt. Source: Refractive index database. Returns material and colour.
- **`copper()`** — Copper. Source: Refractive index database. Returns material and colour.
- **`germanium()`** — Germanium. Source: Refractive index database. Returns material and colour.
- **`gold()`** — Gold. Source: Refractive index database. Returns material and colour.
- **`iridium()`** — Iridium. Source: Refractive index database. Returns material and colour.
- **`iron()`** — Iron. Source: Refractive index database. Returns material and colour.
- **`lead()`** — Lead. Source: Refractive index database. Returns material and colour.
- **`lithium()`** — Lithium. Source: Refractive index database. Returns material and colour.
- **`magnesium()`** — Magnesium. Source: Refractive index database. Returns material and colour.
- **`manganese()`** — Manganese. Source: Refractive index database. Returns material and colour.
- **`mercury()`** — Mercury. Mercury is the only metallic element that is known to be liquid at standard conditions. Source: Refractive index database. Returns material and colour.
- **`molybdenum()`** — Molybdenum. Source: Refractive index database. Returns material and colour.
- **`nickel()`** — Nickel. Source: Refractive index database. Returns material and colour.
- **`palladium()`** — Palladium. Source: Refractive index database. Returns material and colour.
- **`platinum()`** — Platinum. Source: Refractive index database. Returns material and colour.
- **`potassium()`** — Potassium. Source: Refractive index database. Returns material and colour.
- **`rubidium()`** — Rubidium. Source: Refractive index database. Returns material and colour.
- **`silicon()`** — Silicon. Source: Refractive index database. Returns material and colour.
- **`silver()`** — Silver. Source: Refractive index database. Returns material and colour.
- **`sodium()`** — Sodium. Source: Refractive index database. Returns material and colour.
- **`stainlessSteel()`** — Stainless Steel. Source: Refractive index database. Returns material and colour.
- **`titanium()`** — Titanium. Source: Refractive index database. Returns material and colour.
- **`titaniumAnodized()`** — Titanium (Anodized). Titanium can be colorized by a process called anodization. It creates a layer of titanium dioxide by supplying different voltages which results in different thickness and color. Additional colors can be achieved by adding dyes. Source: Refractive index database. Returns material and colour.
- **`tungsten()`** — Tungsten. Source: Refractive index database. Returns material and colour.
- **`vanadium()`** — Vanadium. Source: Refractive index database. Returns material and colour.
- **`zinc()`** — Zinc. Source: Refractive index database. Returns material and colour.

#### Organic

- **`amber()`** — Amber. Source: Amber. Returns material and colour.
- **`banana()`** — Banana. Source: Spectral Materials Database. Returns material and colour.
- **`bone()`** — Bone. Source: Density of Bone. Returns material and colour.
- **`carrot()`** — Carrot. Source: Spectral Materials Database. Returns material and colour.
- **`charcoal()`** — Charcoal. Source: Feeding a Physically Based Shading Model. Returns material and colour.
- **`chocolate()`** — Chocolate. Source: Spectrophotometer Measurement. Returns material and colour.
- **`eggShellBrown()`** — Egg Shell (Brown). Brown hen egg from grocery store. Source: Spectrophotometer Measurement. Returns material and colour.
- **`eggShellWhite()`** — Egg Shell (White). White hen egg from grocery store. Source: Spectrophotometer Measurement. Returns material and colour.
- **`grass()`** — Grass. Source: Spectral Materials Database. Returns material and colour.
- **`honeyCrystallized()`** — Honey (Crystallized). Variations in its water content alter its refractive index. Typically, the refractive index for honey ranges from 1.504 at 13% water content to 1.474 at 25%. Fresh honey is liquid and clear but with time it crystallizes and becomes thicker and less clear. The crystallization depends on temperature and the ratio of glucose and fructose. Source: Honey. Returns material and colour.
- **`ketchup()`** — Ketchup. Source: A Practical Model for Subsurface Light Transport. Returns material and colour.
- **`lemon()`** — Lemon. Source: Spectral Materials Database. Returns material and colour.
- **`orange()`** — Orange. Source: Spectrophotometer Measurement. Returns material and colour.
- **`pearl()`** — Pearl. The akoya pearl is a saltwater cultured pearl from the akoya oyster. Akoya are considered to be the classic pearl used for necklaces and other pearl jewelry. Nacre is the material of which pearls are composed and consists of aragonite crystal layers separated by conchiolin, a protein secreted by mollusk shells. The thickness of these layers determines the color of reflection from the nacre. The transmission interference color is complementary to the reflection interference color. Source: Structural colors of pearls. Returns material and colour.
- **`soapBubble()`** — Soap Bubble. The colors of a soap bubble are the result of an interaction between the light reflected from the front and back surfaces of the soap film, a phenomenon known as thin-film interference. The thickness of a soap bubble can range between 10 nm and 1000 nm. It's thinner at the top due to gravity. Source: Soap Bubble Wiki - Color and Film Thickness. Returns material and colour.

#### Plastic

- **`plasticAcrylic()`** — Plastic (Acrylic). Source: Poly(methyl Methacrylate). Returns material and colour.
- **`plasticPC()`** — Plastic (PC). Source: Polycarbonate. Returns material and colour.
- **`plasticPET()`** — Plastic (PET). Source: Polyethylene Terephthalate. Returns material and colour.
- **`plasticPLA()`** — Plastic (PLA). Source: Polylactic Acid. Returns material and colour.
- **`plasticPP()`** — Plastic (PP). Source: Polypropylene. Returns material and colour.
- **`plasticPUR()`** — Plastic (PUR). Modern car paints are nearly always an acrylic polyurethane "enamel" with a pigmented basecoat and a clear topcoat. It may be described as "acrylic", "acrylic enamel", "urethane", etc. and the clearcoat in particular may be described as a lacquer. Source: Polyurethane. Returns material and colour.
- **`plasticPVC()`** — Plastic (PVC). Source: Polyvinyl Chloride. Returns material and colour.

### `texture`

- **`mask(field, edge, width)`** — mask(field, edge, width): a soft 0..1 threshold of `field` centered at `edge` with transition half-width `width` (ease.quadratic — smoothstep — over a clamp), i.e. 0 well below edge, 1 well above, smooth between. The shared threshold every paint/paintMaterial call below is built on.
- **`fadeDetail(size)`** — fadeDetail(size): how much of a feature `size` mm across the screen can show where it is being drawn: 1 while it spans four pixels or more, falling to 0 by the time it spans two (pixelSize() against size). Multiply a fine feature's contrast by it, or blend the feature toward its average color, so detail too small to resolve fades instead of shimmering. It is 1 wherever nothing is drawn to a screen (meshing, export), so those keep full detail.
- **`coverageLine(distance, width)`** — coverageLine(distance, width): a line `width` wide, 0 to 1, from the distance to its centre, both in world units. A line thinner than a pixel is drawn a pixel wide (pixelSize()) and as much fainter, so it keeps its share of the pixel rather than fading away or breaking into dashes. Wherever nothing is drawn to a screen it is the line itself.
- **`paint(field, edge, width, colorA, colorB)`** — paint(field, edge, width, colorA, colorB): colorA below `edge`, colorB above, blended across `width` via mask() and mixed perceptually (color.mixOklab) — the generalized form of the existing `color.mixOklab(steel, dark, thread.stripe(...))` idiom, so every texture preset paints a scalar field the same way.
- **`paintMaterial(field, edge, width, matA, matB)`** — paintMaterial(field, edge, width, matA, matB): matA below `edge`, matB above, blended across `width` via mask() and material.blend — the material-record sibling of paint().
- **`domainWarp(size, strength)`** — domainWarp(size, strength): a warped point, for use as remap()'s second argument, built from three noise.gradient taps at fixed decorrelating offsets (the same shear-offset rationale as noise.gsl's cellHash4: avoid aliasing with the hash tile's period). FOR COLOR AND RENDER-ONLY DISPLACEMENT INPUTS ONLY: this warps the sample point, not a certified distance value, so feeding the warped point into real subtractive SDF geometry needs its own fresh Lipschitz analysis — it does not inherit any bound from noise.gradient/gradientBump.
- **`flow(velocity, period, sampleFn)`** — flow(velocity, period, sampleFn): a pattern carried along by a flow that may vary from place to place, without the pattern smearing. sampleFn(q) gives the pattern at a point q in world units; velocity is the flow at p in world units per second, a constant or a field. Two copies of the pattern ride the flow, each for one `period` of seconds before it starts again where it began, half a period apart, and each fades out as it restarts (Vlachos 2010), so neither is ever stretched by more than a period's travel. Each restart draws the copy from a new part of the pattern, so it never repeats. The mix is divided by the root of its weights' sum of squares, so a zero-mean pattern keeps its contrast through the crossfade instead of dimming at its middle: give sampleFn a pattern around zero and add its mean afterwards. A shorter period stretches less and pulses more. The mix is up to 1.42 times as steep as the pattern, and more where the flow shears. Inside a library function, use flowPoints and flowMix instead: a lambda passed from one there captures every local before it.
- **`flowPoints(velocity, period)`** — flowPoints(velocity, period): the two points flow samples its pattern at, and the first one's weight: sample the pattern at both and pass the two values and the weight to flowMix, as flow does. Returns a, b and weight.
- **`flowMix(a, b, weight)`** — flowMix(a, b, weight): flow's mix of the pattern sampled at flowPoints' two points, a with `weight` and b with the rest, divided by the root of the weights' sum of squares.
- **`fadeTo(value, mean, size)`** — fadeTo(value, mean, size): value where a feature `size` mm across shows on screen, fading to `mean`, its average over the pattern, as it shrinks past a few pixels (fadeDetail), so detail too small to resolve settles to what it averages to rather than shimmering.
- **`coverageDisc(distance, radius)`** — coverageDisc(distance, radius): a dot `radius` mm across, 0 to 1, from the distance to its centre, with an edge half a pixel soft. A dot smaller than a pixel is drawn a pixel wide and as much fainter, so it keeps its share of the pixel's area, as coverageLine does for a line.
- **`squareWave(x, period)`** — squareWave(x, period): a square wave of `period`, 0 over the first half of each period and 1 over the second, averaged exactly over a pixel's width about x, so its steps blur by what a pixel sees and settle to 0.5 once the period is under a pixel.
- **`reliefRoughness(roughness, slope, size)`** — reliefRoughness(roughness, slope, size): the roughness of a surface whose relief, `size` mm across with slopes of `slope` (rise over run, root mean square), is too small to resolve or not drawn (displacementDrawn()): its slopes, lost with it, spread the highlight as roughness does (Toksvig 2005). The roughness itself while the relief shows. Fade the relief with fadeDetail(size) alongside.

### `material/stone`

- **`stoneFinish(rough)`** — stoneFinish(rough): shared non-metallic stone material — roughness `rough`, not metallic, ior 1.55 (typical for polished stone/mineral).
- **`marble(scale, veinScale, veinThickness, warpStrength, baseColor, veinColor)`** — marble(scale, veinScale, veinThickness, warpStrength, baseColor, veinColor): veined marble with no preferred direction. Folds `scale` across and `warpStrength` mm deep bend veins that sit `veinScale` apart and are about `veinThickness` mm wide, tapering and fading along their length, with a finer network between them, gathered in zones over a faintly clouded, grainy baseColor ground. See calacatta, statuario and carrara for real stones, with a vein direction. Color only.
- **`calacatta(direction)`** — calacatta(direction): Calacatta marble: warm white with sparse, long grey veins 2 to 14 mm wide running along `direction`, gold where they thin, with soft smoky bands beside them, gathered in zones about 300 mm across with clean stone between (about two thirds of the stone carries veins). Sized for slabs a metre or more across. Color only; pair it with physical.marble.
- **`statuario(direction)`** — statuario(direction): Statuario-type marble (after Statuarietto tiles): bright cool white crazed by a dense network of thin, sharp, jagged dark grey veins, 0.3 to 1.5 mm wide, some thickening into seams, loosely following `direction`, with a finer network between them. Sized for tiles and slabs from 300 mm up. Color only; pair it with physical.marble.
- **`carrara(direction)`** — carrara(direction): Carrara marble: grey-white with soft, feathery grey veins and smudges drawn out along `direction`, low in contrast, over the stone's fine sugary grain. Color only; pair it with physical.marble.
- **`granite(cellSize, baseColor, fleckColorA, fleckColorB)`** — granite(cellSize, baseColor, fleckColorA, fleckColorB): a mosaic of interlocking crystals about cellSize mm across (noise.cellularCell), each one mineral with one tone: about 55% feldspar (baseColor), 30% quartz (fleckColorB, greyed toward smoky glass) and 15% dark mica and hornblende (fleckColorA), every crystal shifted lighter or darker by its own hash, with thin dark seams where crystals meet. The quartz reads as translucent: each grain is darker and cooler toward its core than at its rim, as a glassy crystal looks when you see into it, with a few tiny dark inclusions. Color only; pair it with graniteFinish for glassier quartz and shinier mica.
- **`graniteFinish(cellSize, rough)`** — graniteFinish(cellSize, rough): the material to go with granite(cellSize, ...): the polished stone finish at roughness `rough` on the feldspar, with each quartz crystal glassier (a third of the roughness) and each mica flake shinier (half of it), crystal by crystal from the same mosaic.
- **`graniteFacetFinish(cellSize, rough, tilt, gleam)`** — graniteFacetFinish(cellSize, rough, tilt, gleam): graniteFinish for a granite whose quartz throws light back: each quartz crystal is a small mirror tilted its own way (render-only displacement, rising `tilt` mm per mm across the crystal), as a cleavage plane just under the polish catches the light, so the chunks glint one by one as the view moves; the quartz is glassy (a fifth of `rough`) with a reflection strengthened by `gleam` (0 to 1, the metallic weight standing in for light reflected back from inside the crystal). Feldspar and mica as in graniteFinish.
- **`flagstone(cellSize, baseColor, jointColor)`** — flagstone(cellSize, baseColor, jointColor): cellularEdges-driven joint darkening between flagstone/cobblestone slabs — color only, the joints are a paint job, not relief. For MANUFACTURED joint relief (visible to the mesher, not just the renderer), subtract the same field directly from your slab's geometry instead, e.g. `slab - jointDepth * noise.cellularEdges(cellSize)` — that is the caller's own Lipschitz responsibility, the same as any other direct SDF noise subtraction (Displacement.gsl's pebble example does this uncertified today; a fresh certify pass is only worth it for aggressive depths).

### `material/wood`

#### Constants

- **`defaultLook = [0, 0, 0]`** — The look a preset takes when none is given: [oil, age, weather].
- **`noStain = [[1, 1, 1, 0], [0, 0, 0, 0]]`** — noStain: no stain.
- **`unfinished = topcoat([1, 1, 1], 3, 0, 0, 0, 0)`** — unfinished: bare wood, about 3 gloss units, as sanded alder measures.
- **`oilAmber = [1, 0.91, 0.553]`** — oilAmber: the tint of the oil The Wood Database seals its samples with, as finish() fits it at oil 1: sRGB, what it turns white paper. Much ambler than any finish measured on its own; the presets take a share of it.
- **`tungOil = topcoat(amberTint(0.19), 4, 1, 0.005, 0.89, 0.3)`** — tungOil, linseedOil, danishOil: penetrating oils, about 4 gloss units. Tung ambers the wood more at first; linseed yellows more as it ages, even in the dark.
- **`linseedOil = topcoat(amberTint(0.02), 4, 1, 0.005, 0.91, 0.5)`**
- **`danishOil = topcoat(amberTint(0.15), 6, 1, 0.01, 0.9, 0.4)`**
- **`wax = topcoat(amberTint(0.1), 10, 1, 0.002, 0.5, 0.1)`** — wax: paste wax on bare wood, which wets it only part way.
- **`shellac = topcoat(amberTint(0.57), 60, 2, 0.03, 1, 0.2)`** — shellac, blondeShellac: amber shellac, orange-brown, and dewaxed blonde; thin films, 0.03 mm, at a semi-gloss.
- **`blondeShellac = topcoat(amberTint(0.25), 60, 2, 0.03, 1, 0.1)`**
- **`whiteOakNumber = 0`** — Species numbers, for species, speciesColor and speciesMaterial.
- **`redOakNumber = 1`**
- **`liveOakNumber = 2`**
- **`japaneseElmNumber = 3`**
- **`walnutNumber = 4`**
- **`cherryNumber = 5`**
- **`sugarMapleNumber = 6`**
- **`norwayMapleNumber = 7`**
- **`beechNumber = 8`**
- **`rosewoodNumber = 9`**
- **`macassarEbonyNumber = 10`**
- **`ebonyNumber = 11`**
- **`yewNumber = 12`**
- **`douglasFirNumber = 13`**
- **`pineNumber = 14`**
- **`chestnutNumber = 15`**
- **`japaneseWalnutNumber = 16`**
- **`keyakiNumber = 17`**
- **`cherryBirchNumber = 18`**
- **`sycamoreNumber = 19`**
- **`sakuraNumber = 20`**
- **`horseChestnutNumber = 21`**
- **`hinokiNumber = 22`**
- **`sugiNumber = 23`**
- **`ginkgoNumber = 24`**
- **`poplarNumber = 25`**
- **`darkPoplarNumber = 26`**
- **`butternutNumber = 27`**
- **`englishWalnutNumber = 28`**
- **`honeyLocustNumber = 29`**
- **`hickoryNumber = 30`**
- **`teakNumber = 31`**
- **`padaukNumber = 32`**
- **`mahoganyNumber = 33`**
- **`americanElmNumber = 34`**
- **`dutchElmNumber = 35`**
- **`chineseElmNumber = 36`**
- **`whiteAshNumber = 37`**
- **`europeanAshNumber = 38`**
- **`oliveAshNumber = 39`**
- **`osageOrangeNumber = 40`**
- **`europeanSycamoreNumber = 41`**
- **`purpleheartNumber = 42`**
- **`zebrawoodNumber = 43`**
- **`lacewoodNumber = 44`**
- **`spruceNumber = 45`**
- **`westernRedCedarNumber = 46`**
- **`aromaticCedarNumber = 47`**
- **`hollyNumber = 48`**
- **`whiteOakData = [`**
- **`redOakData = [`**
- **`liveOakData = [`**
- **`japaneseElmData = [`**
- **`walnutData = [`**
- **`cherryData = [`**
- **`sugarMapleData = [`**
- **`norwayMapleData = [`**
- **`beechData = [`**
- **`rosewoodData = [`**
- **`macassarEbonyData = [`**
- **`ebonyData = [`**
- **`yewData = [`**
- **`douglasFirData = [`**
- **`pineData = [`**
- **`chestnutData = [`**
- **`japaneseWalnutData = [`**
- **`keyakiData = [`**
- **`cherryBirchData = [`**
- **`sycamoreData = [`**
- **`sakuraData = [`**
- **`horseChestnutData = [`**
- **`hinokiData = [`**
- **`sugiData = [`**
- **`ginkgoData = [`**
- **`poplarData = [`**
- **`darkPoplarData = [`**
- **`mineralPoplarData = [`**
- **`butternutData = [`**
- **`englishWalnutData = [`**
- **`honeyLocustData = [`**
- **`hickoryData = [`**
- **`teakData = [`**
- **`padaukData = [`**
- **`mahoganyData = [`**
- **`americanElmData = [`**
- **`dutchElmData = [`**
- **`chineseElmData = [`**
- **`whiteAshData = [`**
- **`europeanAshData = [`**
- **`oliveAshData = [`**
- **`osageOrangeData = [`**
- **`europeanSycamoreData = [`**
- **`purpleheartData = [`**
- **`zebrawoodData = [`**
- **`lacewoodData = [`**
- **`spruceData = [`**
- **`westernRedCedarData = [`**
- **`aromaticCedarData = [`**
- **`hollyData = [`**
- **`knotStyleData = [`**
- **`stainUptake = [`**
- **`noFungi = [[0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0]]`** — noFungi: no fungus: the spalting amount changes nothing.
- **`xylaria = [[0.106, 0.09, 0.078, 1], [0.91, 0.886, 0.831, 1], [0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0]]`** — xylaria: Xylaria polymorpha, Trametes versicolor and their kin, which spalt maple, beech, birch and ash: black zone lines, and the white rot's bleaching.
- **`whiteRot = [[0.227, 0.165, 0.118, 0.7], [0.91, 0.886, 0.831, 1], [0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0]]`** — whiteRot: a white rot drawing dark brown zone lines and bleaching strongly, as beech's do.
- **`chlorociboria = [[0, 0, 0, 0], [0, 0, 0, 0], [0.247, 0.549, 0.502, 1], [0, 0, 0, 0], [0, 0, 0, 0]]`** — chlorociboria: Chlorociboria's blue-green, xylindein, the green of green oak; it favours aspen and barely takes to birch.
- **`scytalidiumPink = [[0, 0, 0, 0], [0, 0, 0, 0], [0.89, 0.69, 0.643, 1], [0, 0, 0, 0], [0, 0, 0, 0]]`** — scytalidiumPink: Scytalidium cuboideum's pink, pale on maple and magenta on some woods.
- **`scytalidiumYellow = [[0, 0, 0, 0], [0, 0, 0, 0], [0.839, 0.745, 0.447, 1], [0, 0, 0, 0], [0, 0, 0, 0]]`** — scytalidiumYellow: Scytalidium ganodermophthorum's yellow.
- **`blueStain = [[0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0], [0.49, 0.533, 0.576, 1]]`** — blueStain: the sapwood's blue stain, Ophiostoma, Grosmannia and their kin: blue-grey mold in the sapwood alone, as in blue pine.
- **`redRingRot = [[0.106, 0.09, 0.078, 0.6], [0.929, 0.902, 0.847, 0.8], [0.604, 0.353, 0.384, 1], [0, 0, 0, 0], [0, 0, 0, 0]]`** — redRingRot: Porodaedalea pini in pine and Douglas fir: a pink to purplish-red stain, with black zone lines round white pockets.
- **`beefsteak = [[0, 0, 0, 0], [0, 0, 0, 0], [0.353, 0.227, 0.149, 1], [0, 0, 0, 0], [0, 0, 0, 0]]`** — beefsteak: the beefsteak fungus, Fistulina hepatica, which turns oak chocolate brown: brown oak.
- **`dutchElmDisease = [[0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0], [0.416, 0.322, 0.22, 1]]`** — dutchElmDisease: Dutch elm disease's brown streaks in the outer sapwood rings.
- **`noInsects = [[0, 0, 0, 0], [0, 0, 0, 0], [0, 0, 0, 0]]`** — noInsects: no insects: the worm-hole amount changes nothing.
- **`ambrosiaBeetle = [[0.6, 0.67, 60, 0], [0.522, 0.478, 0.408, 1], [0, 0, 0, 0]]`** — ambrosiaBeetle: the Columbian timber beetle and its kin, in fresh logs: pinholes about 1.2 mm across, 25 to 60 mm in from the bark, each in a grey-brown streak: ambrosia maple, wormy maple.
- **`powderpostBeetle = [[0.6, 0.5, 1000, 1], [0, 0, 0, 0], [0, 0, 0, 0]]`** — powderpostBeetle: lyctid powderpost beetles, in the sapwood of oak, ash, hickory and walnut: holes 0.8 to 1.6 mm, unstained.
- **`timberworm = [[0, 0, 0, 0], [0, 0, 0, 0], [2.4, 0.6, 1, 1]]`** — timberworm: the chestnut and oak timberworms: holes 3.2 to 6.4 mm, unstained, boring on into the heartwood: wormy chestnut.
- **`oldHouseBorer = [[0, 0, 0, 0], [0, 0, 0, 0], [3.5, 0.15, 0, 1.5]]`** — oldHouseBorer: the old house borer, Hylotrupes, in softwood sapwood: oval holes 6 to 10 mm, long along the grain.
- **`pinholeBorer = [[0.5, 0.5, 40, 1], [0.125, 0.106, 0.094, 1], [0, 0, 0, 0]]`** — pinholeBorer: the tropical pinhole borers (Euplatypus in rosewood and padauk): pinholes in the sapwood, blackened around.
- **`furnitureBeetle = [[0, 0, 0, 0], [0, 0, 0, 0], [1.2, 0.3, 0, 1]]`** — furnitureBeetle: anobiid furniture beetles: round holes 1.6 to 3.2 mm, in the sapwood.

- **`finish(raw, oil)`** — finish(raw, oil): the colour a finish gives raw sanded wood of sRGB colour `raw`. Per channel in linear light it is k * raw^gamma, fitted to The Wood Database's sanded and sealed photographs of the same boards, 17 species: the exponent is the median of each board's own fit, which one photograph's exposure cannot move. A clear coat (oil 0) darkens every channel as it does red, which is also what wetting a rough surface predicts (Lekner and Dorf); oil 1 is the fit as photographed, which also cuts green and blue, the amber of an oil finish.
- **`sanded(grit)`** — sanded(grit): sanded with the grain on abrasive of FEPA grade P`grit`: 80 coarse, 180 to 220 the usual last grit before a finish, 400 fine.
- **`sawn(feed, noisiness)`** — sawn(feed, noisiness): sawn, the saw's marks across `feed`: noisiness 0 is a bandsaw's even tooth marks, 1 a chainsaw's coarse, torn ones.
- **`tableSawn(feed, jitter)`** — tableSawn(feed, jitter): ripped on a table saw, fed along `feed`, with a 254 mm blade whose runout is `jitter` mm (0.025 a good blade, 0.1 a worn one): each turn of the blade leaves an arc where its proudest tooth passed. The blade rises through the cut along world z from a saw table at z = 0, its centre 100 mm below it, so it stands 27 mm above the table.
- **`tableSawn(feed, jitter, up, offset)`** — tableSawn(feed, jitter, up, offset): the same with the blade rising through the cut along `up` (only its part across the feed counts), its centre `offset` mm below the plane through the origin across `up`. The blade cuts with its rim, never near its centre: offset should keep the cut between about half the blade's radius, 64 mm, and its rim, 127 mm, from the centre, as a table saw's blade rises only a little above the table.
- **`planed(feed, scallop)`** — planed(feed, scallop): jointed or planed, fed along `feed`, the knife marks `scallop` mm apart, the feed per turn of the head (1.5 to 2 on a shop planer, 1.3 to 1.7 a fine finish): each is a shallow arc of the cutterhead's circle, scallop^2 / (8 * radius) deep, which shows only in the sheen, under a raking light.
- **`torched(depth)`** — torched(depth): torched, shou sugi ban: the surface burned to char `depth` mm deep: 0.2 a light torching that browns the wood and brings up its grain, 1 black, 2 to 3 a heavy char cracked into blocks, alligator char. The soft earlywood burns faster than the latewood, darker and deeper, so the latewood stands proud of it. Torching comes last before an oil and takes no stain. Judged.
- **`stain(color, strength)`** — stain(color, strength): a pigment stain, `color` what one coat wiped off turns sugar maple sanded to P180 under a clear coat, as a sample chip shows it, and `strength` the coats, 1 for one. Other woods take it by their own colour and uptake. Pigment lodges where the surface is open, so it darkens the pores, end grain and coarse scratches most.
- **`dyeStain(color, strength)`** — dyeStain(color, strength): a dye stain, which colours the wood itself, as deep in the dense wood as the soft and the pores alike.
- **`clearcoat(color, gloss)`** — clearcoat(color, gloss): a film finish of tint `color` (white for none) and `gloss` in 60 degree gloss units, as a finish's data sheet gives its sheen: matte under 5, satin 20 to 35, semi-gloss 35 to 70, gloss 70 to 85, high gloss above (MPI); a smooth film reads at most about 89. Its film is 0.08 mm, three coats.
- **`oilCoat(color, gloss)`** — oilCoat(color, gloss): a penetrating oil of tint `color` and `gloss` in 60 degree gloss units: it soaks in, darkening the wood, and leaves no film, so the pores stay open.
- **`oilPolyurethane(gloss)`** — oilPolyurethane(gloss), waterPolyurethane(gloss): polyurethane films of `gloss` in 60 degree gloss units. The oil-based one ambers the wood from the first coat and yellows further with age; the water-based one wets the wood less, is water clear and stays clear.
- **`lacquer(gloss)`** — lacquer(gloss): nitrocellulose lacquer, nearly clear, yellowing with age.
- **`species(number, axisOrigin, axisDir, look)`** — species(number, axisOrigin, axisDir, look): the species numbered `number`, one of the species numbers above, whiteOakNumber to lacewoodNumber. Returns material, color. When the number, the stem or the look varies from point to point, use speciesMaterial and speciesColor instead, one for each property: this one then computes both wherever either is asked for.
- **`species(number, axisOrigin, axisDir, look, knotSpacing, barkRadius, normal, prep, stain, topcoat)`** — species(number, axisOrigin, axisDir, look, knotSpacing, barkRadius, normal, prep, stain, topcoat): the species finished, as material, color; see Finishes. `normal` is the surface's, a shape's gradient (`shape.dp`).
- **`species(number, axisOrigin, axisDir, look, knotSpacing, barkRadius, normal, prep, stain, topcoat, spalting, wormholes)`** — species(number, axisOrigin, axisDir, look, knotSpacing, barkRadius, normal, prep, stain, topcoat, spalting, wormholes): the species finished, spalted and worm-eaten, as material, color; see Spalting and worm holes.
- **`species(number, axisOrigin, axisDir, look, knotSpacing, barkRadius, normal, prep, stain, topcoat, spalting, wormholes, fungi, insects)`** — species(number, axisOrigin, axisDir, look, knotSpacing, barkRadius, normal, prep, stain, topcoat, spalting, wormholes, fungi, insects): the same, by the fungi and insects of the records given.
- **`sapwoodDepth(number)`** — sapwoodDepth(number): how deep the sapwood of species `number` runs in from the bark, mm: its depth in rings times its ring width.
- **`speciesColor(number, axisOrigin, axisDir, look)`** — speciesColor(number, axisOrigin, axisDir, look): the colour of the species numbered `number`; see species. The number may vary from point to point, so one shape can hold many species: the wood is computed once at a point, whichever species it is, where a shape per species computes every species everywhere.
- **`speciesColor(number, axisOrigin, axisDir, look, knotSpacing, barkRadius, normal, prep, stain, topcoat)`** — speciesColor(number, axisOrigin, axisDir, look, knotSpacing, barkRadius, normal, prep, stain, topcoat): the colour of the species finished; see Finishes and colorFromRowsFinished. `normal` is the surface's, a shape's gradient (`shape.dp`); the look's oil is not used.
- **`speciesMaterial(number, axisOrigin, axisDir, look)`** — speciesMaterial(number, axisOrigin, axisDir, look): the material of the species numbered `number`; see species and speciesColor. Its sheen is the wood's sanded fine, rising with the look's oil (oiledRoughness). Open-pored wood, pores 0.25 mm across or more, has its pores sunk as grooves with a matte bottom, roughness 0.9, and oil fills them: their depth and their mattness go as 1 - oil, so at oil 1 they are flat and take the finish. The material is built from a roughness and a height chosen here, never by choosing between two materials, which a shader cannot do as it runs. Fixed numbers fold, so close-pored fresh wood keeps the growth model out of the material, which stays one uniform finish.
- **`speciesMaterial(number, axisOrigin, axisDir, look, knotSpacing, normal, prep, topcoat)`** — speciesMaterial(number, axisOrigin, axisDir, look, knotSpacing, normal, prep, topcoat): the material of the species finished: the prepared surface's relief, the open pores' grooves, both levelled by a film, and the topcoat's sheen; see Finishes. `normal` is the surface's, a shape's gradient (`shape.dp`); the look's oil is not used.
- **`fungi(a, b)`** — fungi(a, b): two fungi in one log: the stronger zone lines, bleaching and mold of the two, and a's pigment and b's, each in patches of its own.
- **`insects(a, b)`** — insects(a, b): two insects in one log: the pinholes of the one that makes them, and the grub holes likewise, a's first.
- **`speciesFungi(number)`** — speciesFungi(number): the fungi documented on species `number`, as a fungi record; see Spalting and worm holes.
- **`speciesInsects(number)`** — speciesInsects(number): the insects documented in species `number`, as an insects record.
- **`speciesColor(number, axisOrigin, axisDir, look, knotSpacing, barkRadius, normal, prep, stain, topcoat, spalting, wormholes)`** — speciesColor(number, axisOrigin, axisDir, look, knotSpacing, barkRadius, normal, prep, stain, topcoat, spalting, wormholes): the finished colour with spalting and worm holes, each 0 to 1: up to 0.5 through the sapwood, then on in from the heartwood's edge toward the pith, by the species' own fungi and insects (speciesFungi, speciesInsects); see Spalting and worm holes. The worms bore in from the bark, so barkRadius must be given. Both are weathered with the wood.
- **`speciesColor(number, axisOrigin, axisDir, look, knotSpacing, barkRadius, normal, prep, stain, topcoat, spalting, wormholes, fungi, insects)`** — speciesColor(number, axisOrigin, axisDir, look, knotSpacing, barkRadius, normal, prep, stain, topcoat, spalting, wormholes, fungi, insects): the same, by the fungi and insects of the records given.
- **`speciesMaterial(number, axisOrigin, axisDir, look, knotSpacing, barkRadius, normal, prep, topcoat, spalting, wormholes)`** — speciesMaterial(number, axisOrigin, axisDir, look, knotSpacing, barkRadius, normal, prep, topcoat, spalting, wormholes): the finished material with spalting and worm holes, by the species' own fungi and insects; see speciesColor with them. White-rotted wood is soft: it sands 0.05 mm lower than the sound wood around it, and bare it is matte, up to roughness 0.8, which an oil or a film mostly takes up. A hole is a pit, matte at 0.9. Judged.
- **`speciesMaterial(number, axisOrigin, axisDir, look, knotSpacing, barkRadius, normal, prep, topcoat, spalting, wormholes, fungi, insects)`** — speciesMaterial(number, axisOrigin, axisDir, look, knotSpacing, barkRadius, normal, prep, topcoat, spalting, wormholes, fungi, insects): the same, by the fungi and insects of the records given.
- **`whiteOak(axisOrigin, axisDir, look)`** — whiteOak(axisOrigin, axisDir, look): white oak, ring-porous: a row of big pores after each ring boundary, 1.5 mm rings, and broad rays that show as dashes on flat-sawn faces and flecks on quarter-sawn ones. Measured from Mokume's Japanese white oak (nara, Quercus crispula).
- **`redOak(axisOrigin, axisDir, look)`** — redOak(axisOrigin, axisDir, look): red oak, ring-porous like white oak but pinker, with open pores that read darker and rays a little smaller.
- **`liveOak(axisOrigin, axisDir, look)`** — liveOak(axisOrigin, axisDir, look): live oak, the evergreen oak: dense, diffuse-porous rather than ring-porous, tan, with oak's broad rays. The ring width and pores are judged, not measured.
- **`japaneseElm(axisOrigin, axisDir, look)`** — japaneseElm(axisOrigin, axisDir, look): Japanese elm (harunire, nire), ring-porous: wide 4.5 mm rings, each opening with a narrow row of big pores, then latewood crossed by elm's wavy pale bands. Measured from Mokume's cubes.
- **`walnut(axisOrigin, axisDir, look)`** — walnut(axisOrigin, axisDir, look): black walnut, semi-ring-porous: 3.3 mm rings, medium pores, 0.16 mm across, scattered through each ring, and a grey-brown that a finish turns deep brown.
- **`cherry(axisOrigin, axisDir, look)`** — cherry(axisOrigin, axisDir, look): black cherry, diffuse-porous: 2.8 mm rings that hardly change colour, fine pores throughout, small rays, and dark red-brown gum streaks. The streaks are judged.
- **`sugarMaple(axisOrigin, axisDir, look)`** — sugarMaple(axisOrigin, axisDir, look): sugar maple, the hard maple: a warm cream, diffuse-porous, 2 mm rings, pores too fine to see, and fine rays. The colour is The Wood Database's; the structure is Mokume's maple.
- **`norwayMaple(axisOrigin, axisDir, look)`** — norwayMaple(axisOrigin, axisDir, look): a soft maple with a Norway maple's greyer, pinker pale: diffuse-porous, 2 mm rings, pores too fine to see, and fine rays. Measured from Mokume's maple cubes (Acer spp.).
- **`beech(axisOrigin, axisDir, look)`** — beech(axisOrigin, axisDir, look): beech, diffuse-porous: pale, 2.7 mm rings, pores too fine to see, and rays that fleck every face.
- **`rosewood(axisOrigin, axisDir, look)`** — rosewood(axisOrigin, axisDir, look): East Indian rosewood: purple-brown with darker streaks along the grain and open pores.
- **`macassarEbony(axisOrigin, axisDir, look)`** — macassarEbony(axisOrigin, axisDir, look): Macassar ebony: bold near-black streaks over a tan-brown ground, following the rings loosely.
- **`ebony(axisOrigin, axisDir, look)`** — ebony(axisOrigin, axisDir, look): Gaboon ebony: near black, with rings you cannot see, fine pores and the odd faint brown streak. It polishes glossy.
- **`yew(axisOrigin, axisDir, look)`** — yew(axisOrigin, axisDir, look): English yew, a softwood without resin canals: tight 1.5 mm rings with a crisp dark latewood line, orange-brown heartwood, cream sapwood 12 rings deep in a stem 78 mm to the bark, and knots, a branch every 120 mm or so along the stem. The ring width and the sapwood are judged, and so are the knots, from yew's irregular clustered branching.
- **`douglasFir(axisOrigin, axisDir, look)`** — douglasFir(axisOrigin, axisDir, look): Douglas fir: pale earlywood ending abruptly in a broad orange latewood band, no pores, and resin canals, fine dark dots and lines in the latewood. The ring width is judged.
- **`pine(axisOrigin, axisDir, knotScale, look)`** — pine(axisOrigin, axisDir, knotScale, look): eastern white pine, a soft pine: pale, 4 mm rings that darken gradually into a faint latewood, no pores, and knots: whorls of branches from the pith every knotScale mm along the stem, with the rings bending around each branch. The colour is measured; the ring width is judged.
- **`chestnut(axisOrigin, axisDir, look)`** — chestnut(axisOrigin, axisDir, look): Japanese chestnut (kuri): ring-porous, a row of big pores opening each 3.5 mm ring, pale grey-tan, few rays. Measured from Mokume's cubes; its aging is judged from its nearest measured relative.
- **`japaneseWalnut(axisOrigin, axisDir, look)`** — japaneseWalnut(axisOrigin, axisDir, look): Japanese walnut (kurumi): semi-ring-porous like black walnut but paler and greyer, with wide 8.8 mm rings. Measured from Mokume's cubes; its aging is judged from its nearest measured relative.
- **`keyaki(axisOrigin, axisDir, look)`** — keyaki(axisOrigin, axisDir, look): Keyaki (zelkova): ring-porous and golden, with strong figure on the face. Measured from Mokume's cubes; its aging is judged from its nearest measured relative.
- **`cherryBirch(axisOrigin, axisDir, look)`** — cherryBirch(axisOrigin, axisDir, look): Japanese cherry birch (mizume): diffuse-porous, fine and even, pale pinkish brown. Measured from Mokume's cubes; its aging is judged from its nearest measured relative.
- **`sycamore(axisOrigin, axisDir, look)`** — sycamore(axisOrigin, axisDir, look): American sycamore (platanus): diffuse-porous, pale, with rays that fleck every face, the lacewood look. Measured from Mokume's cubes; its aging is judged from its nearest measured relative.
- **`sakura(axisOrigin, axisDir, look)`** — sakura(axisOrigin, axisDir, look): Sakura, Japanese cherry: diffuse-porous, warm brown, 4.9 mm rings. Measured from Mokume's cubes; its aging is judged from its nearest measured relative.
- **`horseChestnut(axisOrigin, axisDir, look)`** — horseChestnut(axisOrigin, axisDir, look): Japanese horse chestnut (tochinoki): very pale and fine, pores too small to see. Measured from Mokume's cubes; its aging is judged from its nearest measured relative.
- **`hinoki(axisOrigin, axisDir, look)`** — hinoki(axisOrigin, axisDir, look): Hinoki, Japanese cypress: a pale softwood with thin crisp latewood lines and scattered knots. Measured from Mokume's cubes; its aging is judged from its nearest measured relative.
- **`sugi(axisOrigin, axisDir, look)`** — sugi(axisOrigin, axisDir, look): Sugi, Japanese cedar: a softwood with pinkish-red heartwood, pale sapwood 15 rings deep in a stem 168 mm to the bark with a pink ring at its boundary, wide rings and knots in loose whorls. Measured from Mokume's cubes; its aging is judged from its nearest measured relative.
- **`ginkgo(axisOrigin, axisDir, look)`** — ginkgo(axisOrigin, axisDir, look): Ginkgo (icho): plain pale yellow, faint rings, no pores. Measured from Mokume's cubes; its aging is judged from its nearest measured relative.
- **`poplar(axisOrigin, axisDir, look)`** — poplar(axisOrigin, axisDir, look): Yellow poplar: pale creamy yellow, diffuse-porous, rings barely there, and faint greenish heartwood streaks. The colour is The Wood Database's; the structure and aging are judged.
- **`poplar(axisOrigin, axisDir, look, mineral)`** — poplar(axisOrigin, axisDir, look, mineral): poplar with `mineral` of the mineral stain it takes up from the soil, 0 to 1: 0 is plain poplar, 0.5 darkPoplar, the olive-cream heartwood streaked purple-grey, and 1 the heaviest, greenish grey with purple-black streaks. Returns material, color.
- **`poplarColor(axisOrigin, axisDir, look, mineral)`** — poplarColor(axisOrigin, axisDir, look, mineral): the colour alone of poplar with `mineral` of stain; see poplar. Use it, with speciesMaterial(poplarNumber, ...), where the mineral or the look varies from point to point.
- **`poplarColor(axisOrigin, axisDir, look, mineral, normal, prep, stain, topcoat)`** — poplarColor(axisOrigin, axisDir, look, mineral, normal, prep, stain, topcoat): the same finished; see Finishes.
- **`darkPoplar(axisOrigin, axisDir, look)`** — darkPoplar(axisOrigin, axisDir, look): mineral-stained poplar, the same as poplar with mineral 0.5: poplar's olive-cream heartwood heavily streaked with purple-grey mineral stain. The colours are judged from The Wood Database's rainbow poplar; the structure and aging are poplar's.
- **`butternut(axisOrigin, axisDir, look)`** — butternut(axisOrigin, axisDir, look): Butternut: the pale walnut, semi-ring-porous with open pores, light tan. The colour is The Wood Database's; the structure and aging are judged.
- **`englishWalnut(axisOrigin, axisDir, look)`** — englishWalnut(axisOrigin, axisDir, look): English walnut: greyer and lighter than black walnut, semi-ring-porous, with darker streaks. The colour is The Wood Database's; the structure and aging are judged.
- **`honeyLocust(axisOrigin, axisDir, look)`** — honeyLocust(axisOrigin, axisDir, look): Honey locust: ring-porous, light reddish tan with big earlywood pores. The colour is The Wood Database's; the structure and aging are judged.
- **`hickory(axisOrigin, axisDir, look)`** — hickory(axisOrigin, axisDir, look): Hickory: ring-porous, brown heartwood inside cream sapwood 17 rings deep in a stem 140 mm to the bark, the calico look. The colour is The Wood Database's; the structure and aging are judged.
- **`teak(axisOrigin, axisDir, look)`** — teak(axisOrigin, axisDir, look): Teak: golden olive-brown and oily, semi-ring-porous, with dark streaks. The colour is The Wood Database's; the structure and aging are judged.
- **`padauk(axisOrigin, axisDir, look)`** — padauk(axisOrigin, axisDir, look): Padauk: vivid orange-red with large open pores; light turns it brown. The colour is The Wood Database's; the structure and aging are judged.
- **`mahogany(axisOrigin, axisDir, look)`** — mahogany(axisOrigin, axisDir, look): Mahogany: warm reddish brown, diffuse-porous with medium pores and a ribbon of streaks. The colour is The Wood Database's; the structure and aging are judged.
- **`americanElm(axisOrigin, axisDir, look)`** — americanElm(axisOrigin, axisDir, look): American elm: ring-porous, light tan, a narrow row of big pores opening each ring and latewood crossed by wavy pale bands, the feathery figure on flat-sawn faces. The heartwood and sapwood colours are measured side by side on The Wood Database's sample; the structure is judged from its end grain.
- **`dutchElm(axisOrigin, axisDir, look)`** — dutchElm(axisOrigin, axisDir, look): Dutch elm: the European elm hybrid, reddish brown, with elm's pore row and wavy latewood bands. The colour is The Wood Database's sample of wych elm, one of its parents; the sapwood is American elm's measured difference; the structure is judged.
- **`chineseElm(axisOrigin, axisDir, look)`** — chineseElm(axisOrigin, axisDir, look): Chinese elm (lacebark elm): the hardest elm, reddish tan, ring- to semi-ring-porous with fine wavy latewood bands. The colour is from a dealer's photograph, estimated back to unfinished; the sapwood, off-white, and the structure are judged.
- **`whiteAsh(axisOrigin, axisDir, look)`** — whiteAsh(axisOrigin, axisDir, look): White ash: ring-porous like oak, light brown, big earlywood pores two to four rows deep, fine rays and faint pale bands around the latewood pores. The colour is The Wood Database's; the structure and aging are judged.
- **`europeanAsh(axisOrigin, axisDir, look)`** — europeanAsh(axisOrigin, axisDir, look): European ash: white ash's structure in a paler, pinker cream. The colour is The Wood Database's.
- **`oliveAsh(axisOrigin, axisDir, look)`** — oliveAsh(axisOrigin, axisDir, look): Olive ash: European ash with the dark brown streaks of its false heartwood. The ground and streak colours are The Wood Database's.
- **`osageOrange(axisOrigin, axisDir, look)`** — osageOrange(axisOrigin, axisDir, look): Osage orange: bright golden yellow, ring-porous with pores plugged by tyloses, so they read pale, and strong pale bands across the dark latewood. It browns quickly in light. The colour is The Wood Database's; the aging is judged.
- **`europeanSycamore(axisOrigin, axisDir, look)`** — europeanSycamore(axisOrigin, axisDir, look): European sycamore (sycamore maple): a pale diffuse-porous maple, fine and even. The colour is The Wood Database's sample, which is warmer than the near-white of fresh stock.
- **`purpleheart(axisOrigin, axisDir, look)`** — purpleheart(axisOrigin, axisDir, look): Purpleheart: vivid purple, diffuse-porous, turning dark brown with light. The colour and the aging are The Wood Database's fresh and aged samples; the structure is judged.
- **`zebrawood(axisOrigin, axisDir, look)`** — zebrawood(axisOrigin, axisDir, look): Zebrawood: light gold with narrow dark brown stripes, the rings' latewood, sharp and uneven. The two colours are The Wood Database's light and dark stripes; the structure is judged.
- **`lacewood(axisOrigin, axisDir, look)`** — lacewood(axisOrigin, axisDir, look): Brazilian lacewood: reddish brown with very wide pale rays that make the lace on quarter-sawn faces. The colour is The Wood Database's; the rays are judged.
- **`spruce(axisOrigin, axisDir, look)`** — spruce(axisOrigin, axisDir, look): Sitka spruce: a pale cream softwood, tight 1.2 mm rings darkening gradually into a narrow latewood, sparse resin canals, sapwood that shades into the heartwood, and small knots; light yellows it.
- **`westernRedCedar(axisOrigin, axisDir, look)`** — westernRedCedar(axisOrigin, axisDir, look): Western red cedar: a soft, light brown to reddish-brown softwood with abrupt dark latewood lines, darker streaks, nearly white narrow sapwood, no resin canals, and knots; it darkens in light.
- **`aromaticCedar(axisOrigin, axisDir, look)`** — aromaticCedar(axisOrigin, axisDir, look): Aromatic cedar: purple-red heartwood with pale yellow sapwood in streaks, fine even texture, gradual rings, no resin canals and many small knots; light fades it to a reddish brown within weeks.
- **`holly(axisOrigin, axisDir, look)`** — holly(axisOrigin, axisDir, look): American holly: the whitest wood, ivory white and diffuse-porous with rings barely there, fine pores and rays that range from fine to wide; light yellows it slightly.
- **`weatheredOak(axisOrigin, axisDir)`** — weatheredOak(axisOrigin, axisDir): oak left out in the weather for years: silver grey over the grain, the softer earlywood worn back so the latewood stands proud by about 0.4 mm, and drying checks opening along the grain. The relief is render-only displacement. Returns material, color.
- **`weatheredWalnut(axisOrigin, axisDir)`** — weatheredWalnut(axisOrigin, axisDir): weathered walnut, the soft latewood standing only about 0.25 mm proud. Returns material, color.
- **`weatheredPine(axisOrigin, axisDir, knotScale)`** — weatheredPine(axisOrigin, axisDir, knotScale): weathered pine, the hard latewood bands standing about 0.8 mm proud as they do on an old deck or fence board, with the knots of pine(). Returns material, color.

### `material/weathering`

- **`weathered(baseMaterial, baseColor, corrosionColor, amount)`** — weathered(baseMaterial, baseColor, corrosionColor, amount): a broad noise.fbm3 blotch mask plus fine noise.cellularEdges pitting, thresholded via texture.paintMaterial (toward a dead-matte, non-metallic corroded material — corrosion both roughens and de-metallizes) and texture.paint together. `amount` in [0,1] raises how much of the surface reads as corroded. Returns material, color.  Compiler note for any caller wrapping this: forward its result through an explicit destructure (`m, c = weathered(...); return m, c;`), never a bare `return weathered(...);` — tail-forwarding a multi-return call's result undestructured breaks the WGSL property-gradient compiler ("dot expects vectors of the same size") once that result drives a shape's .material or .color. Root-caused by direct bisection against the real WGSL codegen test (agent/wgsl/internal/codegen), not the CLI eval tool, which does not reproduce it. A real, narrow compiler limitation (likely in common/program/property or agent/wgsl's differentiation of a forwarded multi-return), out of scope to fix here — rustedSteel/patinaBronze below show the working pattern.
- **`rustedSteel(amount)`** — rustedSteel(amount): metal.steel weathered toward a rust-orange corrosion color. Returns material, color.
- **`patinaColor()`** — patinaBronze(amount): metal.copper weathered toward a verdigris-green patina, derived from the existing color.teal (mixed toward white in Oklab space, which lightens and desaturates in one step) rather than a new hardcoded color. Deliberately avoids color.desaturate/rgb2hsv here: composed with this function's noise-heavy field, that path triggers a severe compile-time blowup (multi-GB, multi-minute) — likely the rgb2hsv ternary chain interacting badly with the reduce-based noise lattice folds. Root-caused as a compiler performance issue, not a texture-library design problem; worth its own investigation separately, but out of scope here. mixOklab-toward-white sidesteps it entirely.
- **`weatheredRunoff(baseMaterial, baseColor, corrosionColor, amount, flowOrigin, flowDir, streakLength)`** — weatheredRunoff(baseMaterial, baseColor, corrosionColor, amount, flowOrigin, flowDir, streakLength): anisotropic rust-runoff streaking — corrosion visibly runs and pools downward along flowDir instead of forming isotropic blotches. Decomposes p via map.axisAt(flowOrigin, flowDir), compresses the along-flow noise coordinate before sampling (anisotropic noise via non-uniform domain scale), and adds a linear along-flow bias before thresholding. Returns material, color.

### `material/playful`

- **`checkerboard(cellSize, colorA, colorB)`** — checkerboard(cellSize, colorA, colorB): a sharp 3D checkerboard, colorA on the cells of map.checker's even parity. Each axis's squares are averaged over the pixel (texture.squareWave) and the three combined as the chance their parity is odd, so the edges blur by what a pixel sees and a board too fine to resolve settles to the two colours' even mix.
- **`polkaDot(spacing, dotRadius, baseColor, dotColor)`** — polkaDot(spacing, dotRadius, baseColor, dotColor): a hexagonally-packed dot pattern — map.patternHexagonalXY gives the offset to the nearest dot center, and texture.coverageDisc draws the dot there, a pixel soft at its edge. Dots too small or too close to resolve settle to their share of the honeycomb, pi r^2 over 3 sqrt(3) / 4 spacing^2.
- **`confetti(cellSize, paletteFn)`** — confetti(cellSize, paletteFn): a scattered per-region color sweep — combines noise.cellular and noise.gradient at the same p into a pseudo-id fed through paletteFn (e.g. color.cyclicRainbow). An honest approximation, not a true per-cell constant (the per-cell hash itself isn't public) — the id drifts smoothly within a region rather than jumping cleanly at cell walls, which reads as confetti-scatter rather than flat-per-cell color, and is arguably the more festive result anyway.
- **`glitter(baseMaterial, baseColor, flakeCellSize, flakeColor, density)`** — glitter(baseMaterial, baseColor, flakeCellSize, flakeColor, density): sparse bright metallic flecks, one candidate per Voronoi cell — uses cellularF1F2's F1 channel (distance to the nearest feature point, NOT F2-F1 edges) tightly thresholded near 0, since F1 alone is the right primitive for "a bright dot per cell" (edges is for cracks/joints between cells, not points within them). `density` is the flake radius as a fraction of flakeCellSize, so the look holds at any cell size. Returns material, color.
- **`rainbowSheen(scale, cmapFn)`** — rainbowSheen(scale, cmapFn): a rainbow swept along Z, one period every `scale` — explicitly "rainbow-in-space," not true angle-dependent iridescence: material/color fields see only p and time, never a normal or view direction, so real iridescence isn't achievable here.
- **`cellSkin(size, zBandSize, colorA, colorB)`** — cellSkin(size, zBandSize, colorA, colorB): colors noise.cellularEdgesZBlend — an organic "dragon scale / epithelial tissue" surface pattern whose cell-wall topology visibly reorganizes through depth, evoking (not reproducing — see noise.gsl's doc comment on cellularEdgesZBlend) the topology-changing mechanism behind real scutoids.
- **`glitterTwinkle(baseMaterial, baseColor, flakeCellSize, flakeColor, density, sparkle)`** — glitterTwinkle(baseMaterial, baseColor, flakeCellSize, flakeColor, density, sparkle): glitter whose flakes glint as they would under a moving light. Each flake of glitter() is a tiny mirror at an angle of its own, and it glints only while the half-way vector between the light and the eye lies within a few degrees of its normal, the discrete microfacet picture of glints (Jakob et al. 2014). A colour field sees neither the light nor the eye, so a half-way vector of its own turns slowly through every direction in time instead: flakes facing alike flash together and the share of them alight at once is the real one, about 0.5% for a 8 degree glint. Each glint is emitted at `sparkle` times white paper. Returns material, color.
- **`rainbowCycle(scale, period, cmapFn)`** — rainbowCycle(scale, period, cmapFn): rainbowSheen that runs along Z, one band every `scale`, moving up one band every `period` seconds.
- **`polkaDotDrift(spacing, dotRadius, baseColor, dotColor, velocity)`** — polkaDotDrift(spacing, dotRadius, baseColor, dotColor, velocity): polkaDot carried along at `velocity`, in world units per second.

### `material/natural`

- **`water(rippleScale, rippleTimeScale, rippleAmount)`** — water(rippleScale, rippleTimeScale, rippleAmount): glass.water with an animated ripple, via a SINGLE noise.gradientBump4 tap. Water is transmissive, so it is already marched twice (the primary gradient-corrected march plus the secondary refraction march) — keep the ripple to one tap, not a multi-octave stack, or the double-march cost compounds. For water that moves by its own physics, with waves that run at the speed their length gives them, see animated.ocean. Returns material, color.
- **`lava(crackScale, crustColor, glowColor, glowStrength, flowTimeScale)`** — lava(crackScale, crustColor, glowColor, glowStrength, flowTimeScale): a dark rocky crust (noise.cellularEdges crack network) with glowing molten cracks (material.glowing) — primarily a material/color effect, no geometric relief by design. glowStrength is modulated by a slow noise.gradient4 time-morph so the glow seems to pulse and flow along the crack network. For a lava lake whose plates drift and tear open, with seams that glow by a Planck fit, see animated.lavaLake. Returns material, color.
- **`lavaHero(crackScale, crustColor, glowColor, glowStrength, flowTimeScale, crustDepth)`** — lavaHero(crackScale, crustColor, glowColor, glowStrength, flowTimeScale, crustDepth): lava(), plus one cheap noise.gradientBump crust-roughness displacement — an explicit, separately-named EXPENSIVE tier (real geometric relief on top of the already-animated color/material effect). See animated.lavaLake for the lake's physics. Returns material, color.
- **`fogMaterial(density, tintColor)`** — fogMaterial(density, tintColor): a pure material-row (no noise, free) surface approximation of fog — transmission and absorptionDistance eased by `density` in [0,1]. Drape it over a caller-supplied soft blob-union or slab shape; generating that geometry is out of scope here. Returns material, color.
- **`cloudMaterial(density, tintColor, puffScale, puffTimeScale, puffAmount)`** — cloudMaterial(density, tintColor, puffScale, puffTimeScale, puffAmount): fogMaterial, plus one cheap noise.gradientBump4 puff displacement so the surface visibly billows over time. Drape over a caller-supplied soft blob-union or slab shape. Returns material, color.
- **`terrain(scale, snowBias)`** — terrain(scale, snowBias): a world map painted on any shape — continents about `scale` mm across, with mountain chains raised along ridges, coloured by height through sea, beach sand, grass and forest, granite rock and snow (the measured sand, grass and snow presets). Returns material, color and height: the land's height in mm above sea level (negative out at sea, where the colour deepens), proportional to scale, for relief. Subtract max(height, 0) from a shape for real mountains, as Terrain Planet does, or pass it to material.withDisplacement for render-only relief. The terrain bands sit at fixed heights in units of scale / 12: grass from 0.2, rock from 1.2 and snow from 2. snowBias is added to the height, in those units, before the snow line only: 0 for none, or a field that grows toward the poles (up to about 3) for polar caps.

### `material/animated`

- **`ocean(scale, windSpeed, windDirection, choppiness)`** — ocean(scale, windSpeed, windDirection, choppiness): the open sea a wind has raised, as a model `scale` mm to the real metre. The waves are the Pierson-Moskowitz spectrum of a fully developed sea under a wind of `windSpeed` m/s (Pierson and Moskowitz 1964), as eight wave trains spread about `windDirection` (a vector across the surface; waves run along it and turn about z), each running at the speed deep water gives its length, omega^2 = g k, in real seconds: long waves outrun short ones, so the crests keep forming and breaking up. At 10 m/s the peak waves are 83 m long and 7.3 s apart, and the significant height is 2.1 m, all times scale. choppiness, 0 to 1, pulls the crests sharp and the troughs wide as Gerstner waves do (Tessendorf 2001), up to where they would fold. Whitecaps cover the share of the sea Monahan and O'Muircheartaigh (1980) measured for that wind, 1% at 10 m/s and 10% at 20 m/s, on the highest crests. The ripples too small to model roughen the water by the slopes Cox and Munk (1954) measured in sun glitter. Returns material, color, height and slope.
- **`incandescentColor(celsius)`** — incandescentColor(celsius): the hue of a hot body at `celsius`, as sRGB with red at 1: a Planck spectrum through the CIE 1931 observer, fitted from 600 to 1200 C, deep red at 600 and orange by 1200. Every such hue is outside sRGB; its brightness, incandescentStrength, is what turns seams yellow and white once the tone mapping clips it.
- **`incandescentStrength(celsius)`** — incandescentStrength(celsius): how bright a hot body at `celsius` glows, relative to one at 1175 C, basalt lava fresh from a crack: Wien's law at 0.6 um, fitted to a Planck spectrum's luminance, which grows about 4000 times from 700 to 1200 C. Under 550 C (Draper's point) it is too faint to see.
- **`lavaLake(plateSize, life, velocityFn, glowStrength)`** — lavaLake(plateSize, life, velocityFn, glowStrength): the crust of a lava lake, black glassy plates about `plateSize` mm across that drift on the flow velocityFn(x), in mm/s at a point x, tear open along glowing seams and are swallowed, each in `life` seconds. At Erta Ale the plates cross the lake at 1 to 15 cm/s (Spampinato et al. 2008): scale it as ocean is. A seam is lava the plates pulling apart keep baring, 1175 C at its middle (Spampinato et al. 2008), cooling across it to the crust. A newborn plate opens a wide seam that narrows and cools to 1050 C as the plate ages, and is itself fresh lava that skins over and darkens within seconds, its crust thickening with the root of its age (Hon et al. 1994). The glow follows incandescentStrength, times glowStrength. The ropy wrinkles of each plate ride with it. Keep the drift over one life under half a plate, velocity * life < plateSize / 2. The lake lies across z, and its plates are columns along it. The seams and wrinkles are the material's displacement, and the height is 0: the lake is flat at the scale of its plates. Returns material, color, height and slope.
- **`rain(intensity, dropSize)`** — rain(intensity, dropSize): rain falling on still water, `intensity` mm/h of drops about `dropSize` mm across (1 for drizzle, 4 for a downpour), at true scale: about 1000 impacts a square metre a second at 5 mm/h (Marshall and Palmer 1948). Each drop sends out a ring of capillary waves, short ones leading, around a calm disc that grows at 177 mm/s, dying as viscosity takes the short waves (Le Mehaute 1988); a ring crosses 30 mm in about a tenth of a second, as on a real puddle. The water lies across z, as a pool does. The rings are the material's displacement, and the height is 0. Returns material, color, height and slope.
- **`river(scale, velocityFn, turbidity)`** — river(scale, velocityFn, turbidity): a river's surface, as a model `scale` mm to the real metre, flowing at velocityFn(x) m/s at a model point x (a vector; 0.5 to 3 m/s, rapids to 5). Ripples 15 cm across ride the flow (texture.flowPoints, restarting every 1.5 s), and boils, smooth glassy upwellings about the river's depth across, drift slower and quiet the ripples, so the difference in gloss shows them more than their height. Where the flow is fast, standing waves stand still in the current, as long as the flow's speed gives them, 2 pi U^2 / g: 5.8 m at 3 m/s. The colour runs from clear, 0, to silty brown, 1, with `turbidity`. The ripples, boils and standing waves are the material's displacement, and the height is 0. Returns material, color, height and slope.
- **`mudPot(bubbleSize, period, thinness)`** — mudPot(bubbleSize, period, thinness): a boiling mud pot at true scale, bubbles about `bubbleSize` mm across rising from vents every `period` seconds or so (0.5 to 5 s), each swelling into a dome and bursting as its film tears at about 0.5 m/s, in a hundredth of a second (Bubble ascent and rupture in mud volcanoes, 2024). The mud's yield stress decides what the burst leaves: stiff mud, thinness 0, keeps its crater and rim until the next bubble; watery mud, 1, lets them go at once and rings like water. Grey kaolinite mud, dark and glossy where it is wet near the vents and paler and matte where it dries between them. The mud lies across z, as a pool does. The bubbles are the material's displacement, and the height is 0. Returns material, color, height and slope.
- **`iceFloes(floeSize, life, velocityFn, stage, openWater)`** — iceFloes(floeSize, life, velocityFn, stage, openWater): pack ice, floes about `floeSize` mm across drifting on dark water at velocityFn(x) mm/s, about 2% of the wind (Nansen's rule), that form and melt away over `life` seconds, with open leads between them, from closed pack, openWater 0, to scattered floes, 1. `stage` runs through the World Meteorological Organization's stages of sea ice with Perovich's albedos: grease ice (0.1, matte, since it damps the ripples), dark and light nilas (0.15 and 0.25, glossy), pancakes with rims their collisions raise (0.5), grey and first-year ice (0.35 and 0.6), and snow (0.85). Keep the drift over one life under half a floe. The ice lies across z. The floes' freeboard and rims are the material's displacement, and the height is 0. Returns material, color, height and slope.
- **`sandRipples(scale, windSpeed, windDirection)`** — sandRipples(scale, windSpeed, windDirection): wind ripples on dry sand, as a model `scale` mm to the real metre, under a wind of `windSpeed` m/s along windDirection. Their wavelength grows with the wind, about 12 cm at 8 m/s, and they stand a seventeenth as high, a gentle stoss slope to the wind and a steep lee (Nield et al. 2025). They creep downwind once the wind lifts sand, from about 5 m/s, at 0.35 mm/s at 8 m/s and with the cube of the wind, in real seconds, and saltating grains stream over them at about 1 m/s. Dark heavy minerals gather in the troughs. The crests wander and fork. The ripples are the material's displacement, and the height is 0. Returns material, color, height and slope.

### `material/machined`

- **`turned(baseMaterial, baseColor, axisOrigin, axisDir, feed, noseRadius)`** — turned(baseMaterial, baseColor, axisOrigin, axisDir, feed, noseRadius): a surface turned on a lathe about the axis through axisOrigin along axisDir, at `feed` mm a turn with a tool nose `noseRadius` mm round (0.05 to 0.3 feed and 0.2 to 1.6 nose are usual). The tool leaves a helix of grooves round the axis, f^2 / (8 r) deep: the highlight streaks along the axis and the grooves show where they are big enough to see. Returns material, color, height and slope.
- **`faced(baseMaterial, baseColor, axisOrigin, axisDir, feed, noseRadius)`** — faced(baseMaterial, baseColor, axisOrigin, axisDir, feed, noseRadius): the face of a part turned on a lathe, its tool fed in toward the axis: the same grooves as turned, as rings round the axis every `feed` mm, and the highlight streaked across them, a fan of light from the centre. Returns material, color, height and slope.
- **`faceMilled(baseMaterial, baseColor, origin, feedDir, normal, cutterRadius, feedPerRev, cornerRadius)`** — faceMilled(baseMaterial, baseColor, origin, feedDir, normal, cutterRadius, feedPerRev, cornerRadius): a face milled flat, facing along `normal`, by a cutter `cutterRadius` mm across its path fed along feedDir through origin, `feedPerRev` mm a turn, its inserts' corners `cornerRadius` mm round. The inserts leave arcs of the cutter's radius, one every feed, whose cusps are f^2 / (8 r) high: the highlight streaks across the arcs. Returns material, color, height and slope.
- **`ground(baseMaterial, baseColor, direction, roughnessRa)`** — ground(baseMaterial, baseColor, direction, roughnessRa): a surface ground along `direction`, to `roughnessRa` micrometres Ra (0.1 to 0.8; a fine grind 0.1 to 0.4). Its scratches, a few micrometres across, are never big enough to see: the highlight streaks across them, its width across from slopes of about 0.05 to 0.15 over that range (judged). Returns material, color, height and slope.
- **`brushed(baseMaterial, baseColor, direction, grit)`** — brushed(baseMaterial, baseColor, direction, grit): a satin finish brushed along `direction` with an abrasive of `grit` (120 to 240; No. 4 stainless is 150 to 180). Ra falls with the grit, 1.06 um at 150 and 0.76 at 180 (Euro Inox), and the lobe across the lay with it, 0.3 at 150 (judged). Returns material, color, height and slope.
- **`blasted(baseMaterial, baseColor, mediaSize, angular)`** — blasted(baseMaterial, baseColor, mediaSize, angular): a surface blasted with media `mediaSize` micrometres across (50 to 400), round glass beads at angular 0 and sharp sand or grit at 1. Its dimples, a fraction of the media's size, are never big enough to see: they scatter the highlight evenly into a satin to matte sheen, a lobe 0.3 to 0.5 wide for beads and 0.5 to 0.7 for sand, growing with the media (judged from Ra 0.8 to 3.4 um). Returns material, color, height and slope.
- **`knurled(baseMaterial, baseColor, axisOrigin, axisDir, diameter, pitch, diamond, raised)`** — knurled(baseMaterial, baseColor, axisOrigin, axisDir, diameter, pitch, diamond, raised): a knurl rolled on a round `diameter` mm across, about the axis through axisOrigin along axisDir, to ISO 13444: 90 degree teeth `pitch` mm apart round it (0.4, 0.5, 0.6, 0.8, 1.0, 1.2 or 1.6), straight along the axis (diamond 0) or crossed at 30 degrees either way to it (diamond 1), their points raised (raised 1) or sunk (raised 0). The count of teeth round it is whole, so the knurl closes on itself. The teeth are real shape, pitch / 2 deep: subtract the height from the part and divide by 1 + slope. Returns material, color, height and slope.

### `material/coating`

- **`anodized(baseMaterial, baseColor, dyeColor)`** — anodized(baseMaterial, baseColor, dyeColor): aluminium anodized (Type II, 5 to 25 um) and dyed `dyeColor`, white for clear. The oxide is clear and the dye in its pores tints the light through it and back, so the part stays a metal, its colour the metal's times the dye's, a little dimmed and hazed by the oxide (judged). The oxide follows the surface, so the machining beneath shows through: its lay and grooves are kept. Returns material, color, height and slope.
- **`hardAnodized(baseMaterial, baseColor, thickness)`** — hardAnodized(baseMaterial, baseColor, thickness): hard anodize (Type III), `thickness` um of oxide (25 to 100). Undyed, it greys and darkens toward a bronze-black as it thickens, the metal's glint fading into the film's (judged from 25 um grey to 75 um near black). Returns material, color, height and slope.
- **`blackOxide(baseMaterial, oiled)`** — blackOxide(baseMaterial, oiled): steel blackened by black oxide, a film of magnetite about a micrometre thick. It is too thin to hide the machining, so the lay and grooves are kept, and black, its glint the magnetite's (n about 2.4); the oil it is sealed with (oiled 1) glosses it. Returns material, color, height and slope.
- **`phosphate(baseMaterial, manganese)`** — phosphate(baseMaterial, manganese): steel phosphated, a crust of crystals a few micrometres across that hides the machining under a matte grey, dark for manganese phosphate (manganese 1) and lighter for zinc (0). The crystals are far too small to see: they scatter the highlight evenly and take the lay with them. Returns material, color, height and slope.
- **`glossRoughness(glossUnits)`** — glossRoughness(glossUnits): the roughness of a coat whose 60 degree gloss is `glossUnits` (90 high gloss, 50 satin, 20 low, 5 matte).  This was its own closed form, the lobe width 0.044 sqrt(100 / GU - 1), judged from a GGX lobe's tail. It read far too rough. Integrating the GGX/Smith/Schlick BRDF over the receptor ISO 2813 specifies -- 4.4 by 11.7 degrees at 60 degree incidence, against the black glass standard of index 1.567 -- and asking what each roughness would MEASURE on that meter:  stated GU   old form reads   material.roughnessFromGloss reads 20             16.2                              19.9 45             31.4                              49.5 70             48.9                              76.4 85             64.0                              85.8  The old form is off by up to 21 gloss units through the range coatings actually live in; the material library's fit, which is numerically integrated rather than judged, is within about 7 at worst and within 1 at both ends. So this now defers to it, and the name is kept because it reads better at the call sites and because a coat's gloss is a coating idea.
- **`powderCoat(coatColor, glossUnits, peel)`** — powderCoat(coatColor, glossUnits, peel): a powder coat of `coatColor` whose gloss is `glossUnits`, with orange peel: the coat flows out into gentle waves, `peel` from 0 (flat) to 1 (heavy), which a wave-scan measures in two bands, long waves of a few millimetres and short ones of a few tenths. A few micrometres high, they wobble reflections up close; seen smaller than a pixel their slopes widen the lobe instead. Returns material, color, height and slope.
- **`wrinkleCoat(coatColor, size)`** — wrinkleCoat(coatColor, size): a wrinkle powder coat, its skin cured faster than beneath it and puckered into ridges `size` mm apart (0.3 to 1), low gloss. Returns material, color, height and slope.
- **`sandCoat(coatColor, grain)`** — sandCoat(coatColor, grain): a sand-textured powder coat, grains `grain` mm across (0.1 to 0.3) standing out of a matte coat. Returns material, color, height and slope.
- **`hammertone(coatColor, cellSize)`** — hammertone(coatColor, cellSize): a hammertone powder coat, aluminium flake in a coat that sets in dimples `cellSize` mm across (2 to 6), as if beaten: the flake glints and each dimple catches the light its own way. Returns material, color, height and slope.
- **`edgeWear(faceA, faceB, width)`** — edgeWear(faceA, faceB, width): how worn a part is near the edge where two of its faces meet, from their distances: 1 on the edge, falling to 0 `width` mm onto either face. Take the max over the edges that wear, e.g. `wear = max(coating.edgeWear(slabX, slabY, 3), coating.edgeWear(slabY, slabZ, 3));`, and hand it to chippedPaint.
- **`chippedPaint(baseMaterial, baseColor, paintColor, primerColor, glossUnits, wear, chipSize)`** — chippedPaint(baseMaterial, baseColor, paintColor, primerColor, glossUnits, wear, chipSize): paint of `paintColor` and `glossUnits` gloss over a primer of `primerColor` over the part, chipped where `wear` (0 to 1, see edgeWear) is high: chips about `chipSize` mm across show the primer at their rims and the bare part within, sunk by the coats' 0.08 mm. Chips too small to see settle to their share. Returns material, color, height and slope.
- **`galvanized(spangleSize, age)`** — galvanized(spangleSize, age): hot-dip galvanized steel, its zinc frozen into spangle, crystals `spangleSize` mm across (1 to 10 regular, under 1 minimized). Each crystal's grains lie its own way, so it takes a lay of its own and catches the light alone, bright or dark as the part turns, between thin grain boundaries. `age` from 0 (fresh) to 1 dulls it to the even grey of weathered zinc. Crystals too small to see settle to their average, the lay's anisotropy kept with no direction to streak. Returns material, color, height and slope.
- **`levelled(coat, baseMaterial, build, underSize)`** — levelled(coat, baseMaterial, build, underSize): `coat` laid as a flowed film `build` mm thick over `baseMaterial`, whose relief is `underSize` mm across. What is left of the surface beneath survives in three ways at once, all governed by the one share: its displacement is cut by it, its lay's anisotropy by it, and its lobe's variance by its square, variance going as amplitude squared. The coat's own metallic and index are kept, so a clear film passes them through and an opaque one replaces them.
- **`clearCoat(baseMaterial, baseColor, tint, glossUnits, build, underSize)`** — clearCoat(baseMaterial, baseColor, tint, glossUnits, build, underSize): a clear film over the part — a urethane, lacquer or clear powder — of `tint` (white for none) and 60 degree gloss `glossUnits`, `build` mm thick (0.03 a shellac, 0.08 three coats of polyurethane, 0.12 a clear powder), flowed over a surface whose relief is `underSize` mm across. The film fills what is finer than it is thick: a 0.08 mm urethane leaves 1.2% of a P120 brushed lay (0.133 mm) and 1.6% of a lathe's 0.15 mm feed, so the turning all but disappears, but 16% of a 0.5 mm as-cast skin; a 0.002 mm wax over that same lathe finish leaves it untouched.  The light passes through the film to the metal, so the part keeps its own colour, metallic and index, dimmed a little by the film. What one lobe cannot carry is the film's own reflection, which is a dielectric's and so white where the metal's is coloured: a coated metal's highlight is tinted here where it should have a white film glint over it. That wants a second lobe, which the material record has no row for yet. Returns material, color, height and slope.
- **`powderCoat(baseMaterial, coatColor, glossUnits, peel, build, underSize)`** — powderCoat(baseMaterial, coatColor, glossUnits, peel, build, underSize): a powder coat sprayed over a prepared surface, which shows through where the coat is too thin to fill it. Otherwise powderCoat(coatColor, glossUnits, peel). Returns material, color, height and slope.
- **`wrinkleCoat(baseMaterial, coatColor, size, build, underSize)`** — wrinkleCoat(baseMaterial, coatColor, size, build, underSize): a wrinkle powder coat over a prepared surface. Its ridges are coarse, so a surface under it has to be coarser still to read through them. Otherwise wrinkleCoat(coatColor, size). Returns material, color, height and slope.
- **`sandCoat(baseMaterial, coatColor, grain, build, underSize)`** — sandCoat(baseMaterial, coatColor, grain, build, underSize): a sand-textured powder coat over a prepared surface. Otherwise sandCoat(coatColor, grain). Returns material, color, height and slope.
- **`hammertone(baseMaterial, coatColor, cellSize, build, underSize)`** — hammertone(baseMaterial, coatColor, cellSize, build, underSize): a hammertone powder coat over a prepared surface. Otherwise hammertone(coatColor, cellSize). Returns material, color, height and slope.

### `material/printed`

- **`fdm(plasticColor, buildDir, center, layerHeight, glossUnits)`** — fdm(plasticColor, buildDir, center, layerHeight, glossUnits): a part printed by fused filament, `layerHeight` mm a layer (0.1 to 0.3) along buildDir. Each layer is a bead whose side bulges round, so a wall is a chain of semicircles a layer across, Ra 0.091 L (the chain's mean deviation from its own mean, pi L / 8), and the plastic's gloss is `glossUnits` (PLA about 40, PETG 70, matte filaments 10). `center` is a point near the part's middle, for the lay. While the beads show they are relief; seen smaller than a pixel their slopes, about a radian either way, streak the highlight across the layers instead. Returns material, color, height and slope.
- **`resin(resinColor, buildDir, center, layerHeight, clarity)`** — resin(resinColor, buildDir, center, layerHeight, clarity): a part printed in resin by stereolithography or masked SLA, `layerHeight` mm a layer (0.025 to 0.1): near glossy, its layers faint steps a few micrometres deep, sharp where a shallow slope steps through them and a faint streak across them where they are too fine to see. `clarity` from 0 (opaque, the usual grey or colour) to 1 (a clear resin) lets light through. `center` is a point near the part's middle, for the lay. Returns material, color, height and slope.
- **`powderBed(powderColor, grain)`** — powderBed(powderColor, grain): a part sintered or fused from nylon powder, by selective laser sintering (white, grain about 0.06 mm) or multi jet fusion (grey, often dyed black): matte and grainy, the half-fused grains of its skin `grain` mm across, its layers lost among them. Returns material, color, height and slope.
- **`laserMelted(baseMaterial, baseColor, roughnessRa)`** — laserMelted(baseMaterial, baseColor, roughnessRa): metal printed by laser powder-bed fusion (DMLS, SLM), as built, `roughnessRa` micrometres Ra (5 to 15). Its skin is crusted with powder half-melted onto it, about 0.04 mm across, far too fine to shade as relief: its slopes, 0.6 at Ra 10, spread the highlight evenly, and up close each speck of powder is a shade lighter or darker, settling to the metal's own colour as it shrinks below a pixel. The layers, a few hundredths of a millimetre, are lost beneath it. Returns material, color, height and slope.

