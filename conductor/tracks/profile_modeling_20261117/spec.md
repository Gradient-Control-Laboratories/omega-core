# Specification: Profile Modeling

> Spec agreed 2026-11-17; no plan yet. Section 2 records the decisions behind the profiles already on
> `main`, section 4 the work in this track in its order, and section 7 the decisions that scoped it.

## 1. Why

A designer who comes from CAD thinks in sketches: a closed outline of lines and arcs, extruded, revolved or
swept. GSL can now say that directly, and say it with an **exact** signed distance, which is what lets
`gsl eval` prove things about the part (`certify`, `bounds`) instead of estimating them. Agents in the
`gsl eval` pilots already reach for it: the bracket in [eval_fixes_20261115](../eval_fixes_20261115/spec.md)
was an extruded `profileXY`.

The backlog ([backlog_20261110](../backlog_20261110/spec.md)) holds no modeling work beyond one lattice
(`M12`). This track is the modeling thread: what a profile is, what it can become, and what each step
guarantees about its distance.

## 2. Decisions Already Made

| Topic                | Decision                                                                                                                                                                                                                                       |
| :------------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Distance             | Exact signed distance first. Nearest-feature formulation: 2N features per loop (each vertex and each edge), pooled by sign. Not a winding-number or crossing-count formulation.                                                                |
| Sign pooling         | A vertex's sign comes from its turn (convex +, reflex −); an edge's from its side. The magnitude comes from an ungated pool and only the sign from comparing the two gated pools, so surface cells stay interval-provable.                      |
| Winding              | Counterclockwise outer loops, clockwise holes.                                                                                                                                                                                                 |
| Edges                | Vertex-first circlines: the points, plus one signed curvature per edge (0 straight, + bulging left along the loop). Minor arcs only: `abs(curvature * chord) <= 2`.                                                                            |
| Smooth joints        | A joint whose turn is within 1e-6 of zero (tangent-continuous, or a collinear point) joins neither sign pool.                                                                                                                                  |
| Revolve              | CAD-like, on an axis's distance-angle-along basis (`axisCylindrical`): the axis is a 2D line in the sketch plane, the profile read on the right of its direction.                                                                               |
| Frames               | A profile's `.u` is the running length round its loop. `extrude`: `.u` the profile's, `.v` the height z. `revolve`: `.u` the angle round the axis (the surface-of-revolution convention turned materials read), `.v` the profile's `.u`.       |
| Boxes                | A profile declares its XY box from its points (arcs padded); `extrude` and `revolve` carry boxes of their own.                                                                                                                                 |
| Revolve in a backend | `revolve(profile, …)` is inlined because it is handed a field it remaps (`remapsFieldArg`, #53), not by a special case.                                                                                                                        |

## 3. What Exists

On `main` (6f5462b):

- `profileXY(points)` and `profileXY(points, curvatures)` in `stdlib/builtin/2d.gsl`, with
  `profileLengthXY` and the box, run and pool helpers.
- `extrude(profile, low, high[, blend])` and `revolve(profile, axisPoint, axisDirection)` with
  `revolveAngle` in `stdlib/builtin/sweep.gsl`.
- `axis`, `axisBasis`, `axisAngle`, `axisDistance`, `axisCylindrical` in `stdlib/builtin/3d.gsl`.
- The older `sweepLoft*` family in `stdlib/builtin/sweep.gsl`: nullary profile functions lofted along a
  quadratic Bezier, with optional scaling and separate caps. It predates `profileXY` and does not take one.
- Examples: [profile_star.gsl](../../../examples/profile_star.gsl),
  [profile_arcs.gsl](../../../examples/profile_arcs.gsl),
  [profile_cylinder.gsl](../../../examples/profile_cylinder.gsl) (an L revolved and wrapped onto a cylinder),
  and [profile_sweep.gsl](../../../examples/profile_sweep.gsl), which sweeps a section along a closed profile
  as a remap: the section read at (the path's signed distance, z), with `.u` the path's running length. A
  sweep along a closed planar path is therefore already expressible; it is not yet a library function.

## 4. Scope, in Order

Profiles first, then the operations on them. The numbering is the order.

1. **C1. Several loops in one profile.** Today an outline and its holes are combined by hand with `max`. For
   properly nested loops (holes inside the outline, islands inside holes) `max`/`min` of exact loop distances
   is exact everywhere, because each region's nearest boundary belongs to one loop. `profileXY` takes a list
   of loops, combines them that way, and checks that they nest.
2. **C2. Corner blends, per vertex.** "Every corner eventually a blend." Each vertex can take either kind:
   - an **arc fillet** (or a **line chamfer**) generated as edges of the loop, so the distance stays exact,
     as the hand-written L in [profile_cylinder.gsl](../../../examples/profile_cylinder.gsl) is; or
   - a **field blend** at the corner (a round or chamfer operator), where exactness is not needed; the
     distance there is a bound.
3. **C6. Quadratic Bezier edges.** The stdlib already has an exact distance to a 3D quadratic Bezier
   (`bezierQuadPointRound` in `stdlib/builtin/quad_bezier_curves.gsl`, Quilez's closed-form cubic solve);
   the 2D case is simpler. A profile edge can be a line, an arc or a quadratic Bezier, and the profile stays
   exact. Higher-order splines are not in this track.
4. **C7. Tighter arc (and Bezier) boxes.** The arc profile's box is the points' box padded by
   `min(span / 2, span^2 * maxCurvature / 4)`: sound, but loose, because a `.box` could not hold the reduce
   an exact box needs (see section 5).
5. **C3. Partial revolve.** A revolve through an angle range, capped at both ends, with `.u` running over the
   range.
6. **C4. `sweep(section, path, …)`.** Exact first: a closed planar path, from the remap in
   [profile_sweep.gsl](../../../examples/profile_sweep.gsl). Open and non-planar paths come after, with a
   stated Lipschitz bound: there the section's frame is transported and the swept distance stops being exact
   past the path's radius of curvature.
7. **C5. Caps for sweeps.** Cylinder-style caps where an open sweep ends.

**Upstream, in its own PR:** `cylinderCapped(pointA, pointB, radius[, blend])` is placed by its two end
points, which makes it a beam rather than a cylinder. It is fixed upstream before the caps of C3 and C5 are
modelled on it.

## 5. Constraints Met While Writing #45

Noted on 2026-11-11 and 2026-11-12; each is to be rechecked on `main` before planning, since S1 (postfix
modifier chaining) and other front-end work have landed since.

1. A `.box` expression could not contain a `reduce` whose function captures anything or takes an index.
2. `.u`, `.v` and `.box` could not be read inside a function body.
3. A function's internal `remap` carries frame, colour and material only when its whole body is one
   `return remap(param, coord)` with `coord` in scalars of `p`, parameters and primitives
   (`internal/assembler/property/remap_wrappers.go`): no locals, no stdlib calls, no component of a call's
   result. `revolve` is written that way, which is why it cannot declare `.v` or a box itself. Backlog `M8`
   (the coordinate as an explicit parameter upstream) would lift this.
4. `m[i][j]` and `m[i].x` did not parse inside a list literal.
5. A `reduce` whose function reads its accumulator several times per step unrolls to code that grows
   exponentially with the list's length: 3.1 MB of WGSL for eight edges. Keep accumulator reads to one or
   two per step.
6. A new stdlib function can be right on the CPU (eval, certify, fidget) and wrong in WGSL (the original
   `revolve`): its WGSL output has to be checked too.

## 6. Verification

- **Exactness:** a brute-force Go reference (distance to a densely sampled boundary, sign by containment)
  agrees with the field to about 1e-6 on an L, a star, a square with a hole and an arc profile, and on every
  new edge or operation this track adds.
- **Provability:** `certify` and `bounds` prove on surface cells, including rounded rims above an extrude's
  caps.
- **Backends:** each example compiles to WGSL; its size is checked against a budget, and the viewer shows
  what the CPU evaluates.
- **Tests:** in Go. A Python test of the wheel's API only where `AGENTS.md`'s owner allows one.

## 7. Scoping Decisions (2026-11-17)

| Question                        | Decision                                                                                                         |
| :------------------------------ | :--------------------------------------------------------------------------------------------------------------- |
| Which candidates, in what order | All seven, profiles first: C1, C2, C6, C7, then C3, C4, C5 (section 4).                                          |
| Corner blends (C2)              | Both, chosen per vertex: exact arc fillets and line chamfers as edges, or a field blend where exactness is not needed. |
| Splines (C6)                    | Quadratic Bezier edges, exact, from the existing 3D distance. Higher orders are out of this track.               |
| Sweep guarantee (C4)            | Exact first, on closed planar paths; open and non-planar paths later, with a stated Lipschitz bound.             |
| The `sweepLoft*` family         | Kept as it is; revisited once `sweep()` exists.                                                                  |
| Sketch planes                   | Profiles stay in XY and are placed with `transform`; no sketch-plane argument.                                   |
| Backlog `M8`                    | Not waited for: `revolve`, `sweep` and caps are written within constraint 3, as `revolve` is now.                |
| `cylinderCapped`                | A beam by its end points; fixed in its own upstream PR before C3 and C5 model caps on it.                        |
