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

1. **C1. Several loops in one profile.** Today an outline and its holes are combined by hand with `max`.
   Decided 2026-11-17 (Phase 1's proposal): a generic `nested(fields)` in `stdlib/builtin/booleans.gsl`
   takes a list of exact fields whose surfaces do not cross, each wound so the solid is on its negative side
   (counterclockwise outlines and islands, clockwise holes), and gives the nearest surface's distance with
   that surface's sign. That is exact for any such arrangement, in any order, with no nesting tree: the
   segment from a point to its nearest surface point crosses no other surface, so the point lies on the side
   that surface's field reports. It is written as `profileXY`'s pools are (a sign gate combined with `max`,
   the magnitude from an ungated pool), so it stays interval-provable. Loops are passed as fields, not as
   data, because GSL refuses ragged lists; and nothing checks the winding, because GSL cannot raise a
   diagnostic. A mis-wound loop reports its own distance wherever it is the nearest surface: right in sign
   inside it, but void just outside it in the material, with the field jumping sign where another surface
   becomes the nearer. The doc comment states the rule. Being generic, it serves 3D fields too (a solid
   with a void).
2. **C2. Corner blends, per vertex.** "Every corner eventually a blend." Each vertex can take either kind:
   - an **arc fillet** (or a **line chamfer**) generated as edges of the loop, so the distance stays exact,
     as the hand-written L in [profile_cylinder.gsl](../../../examples/profile_cylinder.gsl) is; or
   - a **field blend** at the corner (a round or chamfer operator), where exactness is not needed; the
     distance there is a bound.

   Decided 2026-11-17 (Phase 2's proposal): `profileXY(points, curvatures, corners)` and
   `profileXY(points, curvatures, corners, blend)`. `corners` holds one helper per vertex: `cornerSharp()`,
   `cornerFillet(radius)`, `cornerChamfer(setback)` or `cornerBlend()`; `blend` is any of the stdlib's
   partially applied edge blends (`intersectRound(r)`, `intersectChamfer(d)`, `intersectLp(…)`, …), applied
   at every `cornerBlend()` vertex, its dual (`neg(blend(neg …))`) at a reflex one. GSL passes an operator as
   an argument but cannot hold operators in a list, hence one operator per profile. `everyCorner(points,
   corner)` builds a `corners` list of one kind for every vertex. A fillet or chamfer too large for its edges
   is clamped to fit (each trim at most half of each adjacent edge), documented, since GSL cannot refuse it
   with a message. Corners between two straight edges only: corners next to an arc stay sharp, the second
   part of the phase deferred by the owner on 2026-11-17 in favour of implicit blends and more pressing
   work. `intersectRound` makes the same surface as an exact fillet only at a right angle, so field blends
   are genuinely other shapes.
3. **C6. Quadratic Bezier edges.** The stdlib already has an exact distance to a 3D quadratic Bezier
   (`bezierQuadPointRound` in `stdlib/builtin/quad_bezier_curves.gsl`, Quilez's closed-form cubic solve);
   the 2D case is simpler. A profile edge can be a line, an arc or a quadratic Bezier, and the profile stays
   exact. Higher-order splines are not in this track.
4. **C7. Tighter arc (and Bezier) boxes.** The arc profile's box is the points' box padded by
   `min(span / 2, span^2 * maxCurvature / 4)`: sound, but loose, because a `.box` could not hold the reduce
   an exact box needs (see section 5).
5. **C3. Partial revolve.** A revolve through an angle range, capped at both ends, with `.u` running over the
   range. Decided 2026-11-18: `revolve(profile, axisPoint, axisDirection, angle)`, the angle from the
   sketch plane, counterclockwise seen from the tip of the direction as `revolveAngle` runs, negative the
   other way; flat caps. Exact: the magnitude is the least of the revolved distance (gated off outside the
   sweep by a constant) and each cap's distance, sqrt(h^2 + max(d(r, a), 0)^2) in that cap's plane; the sign
   is negative inside the sweep and the profile. The profile is read three times. Its box and frame come
   from Go rules beside `revolved` and `revolveFrame`, since a body cannot read `profile.box` or `.u`
   (constraint 2): the box is the full revolve's, the frame the full revolve's (`.u` the angle, `.v` the
   profile's `.u`). Taken before C6 and C7 (2026-11-18): revolve and sweep first.
6. **C4. `sweep(section, path, …)`.** Exact first: a closed planar path, from the remap in
   [profile_sweep.gsl](../../../examples/profile_sweep.gsl). Open and non-planar paths come after, with a
   stated Lipschitz bound: there the section's frame is transported and the swept distance stops being exact
   past the path's radius of curvature.

   Decided 2026-11-18: `sweep(section, path)`, the path any exact 2D field in XY whose zero line it is
   (`profileXY`, `nested(...)`, a circle), the section a 2D profile with x across the path (positive outside)
   and y up as z. It is the remap `profile_sweep.gsl` writes by hand, `remap(section, [path, p.z, 0])`:
   1-Lipschitz and zero on the swept surface everywhere, so never more than the distance, and exact where
   p's nearest path point is unique, which holds while the section stays inside the path's tube (on the
   inside of a bend, within its radius; nowhere past the path's medial axis). At a sharp corner of the path
   it sweeps round the outside and mitres the inside, as documented; fillet the path for a smooth sweep.
   Frame and box by Go rules beside revolve's: `.u` the path's `.u`, `.v` the section's `.u`, the box the
   path's grown by the section's reach across and its height.
7. **C5. Caps for sweeps.** Cylinder-style caps where an open sweep ends.

## 5. Constraints Met While Writing #45

Noted on 2026-11-11 and 2026-11-12. Constraints 1–4 were rechecked on `main` (85a0152) on 2026-11-17 with
probe documents compiled to WGSL and queried with `gsl eval`; the result follows each.

1. A `.box` expression could not contain a `reduce` whose function captures anything or takes an index.
   **Lifted:** in a document's function, a box holding a reduce that captures a parameter, and one that reads
   its index and a neighbour (`points[mod(index + 1, count)]`), both compile and clip `bounds` where they
   say. Not yet tried inside `stdlib/builtin`, where #45 met it; C7 confirms it there.
2. `.u`, `.v` and `.box` could not be read inside a function body. **Holds:** each is refused with a message
   to read it at the call site.
3. A function's internal `remap` carries frame, colour and material only when its whole body is one
   `return remap(param, coord)` with `coord` in scalars of `p`, parameters and primitives
   (`internal/assembler/property/remap_wrappers.go`): no locals, no stdlib calls, no component of a call's
   result. `revolve` is written that way, which is why it cannot declare `.v` or a box itself. Backlog `M8`
   (the coordinate as an explicit parameter upstream) would lift this. **Holds:** the rule is unchanged.
4. `m[i][j]` and `m[i].x` did not parse inside a list literal. **Lifted:** both parse and evaluate in a list
   literal at the top level, in a function, and in a reduce's function.
5. A `reduce` whose function reads its accumulator several times per step unrolls to code that grows
   exponentially with the list's length: 3.1 MB of WGSL for eight edges. Keep accumulator reads to one or
   two per step.
6. A new stdlib function can be right on the CPU (eval, certify, fidget) and wrong in WGSL (the original
   `revolve`): its WGSL output has to be checked too.

## 6. Verification

- **Exactness:** a brute-force Go reference (distance to a densely sampled boundary, sign by containment)
  agrees with the field to within 2e-5 mm (float32 evaluation; today's profiles sit under 9e-6) on an L, a
  star, a square with a hole, an arc profile and a slot with smooth joints, and on every new edge or
  operation this track adds.
- **Provability:** interval bounds over cells straddling the surface contain the field and are no wider
  than four cell diagonals, the cells 0.5 mm or, on a profile with fillets, no wider than a fifth of the
  smallest fillet's radius (an interval over a curve is as loose as the cell is wide against its radius:
  at a radius-1 fillet, 4.35 diagonals at 0.5 mm cells, 2.43 at 0.04 mm), and `certify` proves the space just off the part empty, including above a
  rounded rim, and `bounds` (`Evaluator.Enclose`) proves the part's box. Whether `Enclose` proves can hang on
  where its cuts land: Phase 0's star proves over one region and runs out of budget over another. A sweep
  reads its section at the path's distance, so the path's interval feeds the section's and their looseness
  compounds (owner's decision, 2026-11-18): its cells may be six diagonals wide (5.54 measured), `certify`
  proves it with a larger budget (about 300,000 evaluations above it, against a default of 200,000), and
  `bounds` is not asserted, its budget fixed short of what it needs (backlog `M16`).
- **Backends:** each example compiles to WGSL; its size is checked against a budget, and the viewer shows
  what the CPU evaluates.
- **Tests:** in Go. A Python test of the wheel's API only where `AGENTS.md`'s owner allows one.

## 7. Scoping Decisions (2026-11-17)

| Question                        | Decision                                                                                                         |
| :------------------------------ | :--------------------------------------------------------------------------------------------------------------- |
| Which candidates, in what order | All seven, profiles first: C1, C2, C6, C7, then C3, C4, C5 (section 4).                                          |
| Corner blends (C2)              | Both, chosen per vertex: exact arc fillets and line chamfers as edges, or a field blend where exactness is not needed. Field blends are the stdlib's edge-blend operators, one per profile (section 4). |
| Splines (C6)                    | Quadratic Bezier edges, exact, from the existing 3D distance. Higher orders are out of this track.               |
| Sweep guarantee (C4)            | Exact first, on closed planar paths; open and non-planar paths later, with a stated Lipschitz bound.             |
| The `sweepLoft*` family         | Kept as it is; revisited once `sweep()` exists.                                                                  |
| Sketch planes                   | Profiles stay in XY and are placed with `transform`; no sketch-plane argument.                                   |
| Backlog `M8`                    | Not waited for: `revolve`, `sweep` and caps are written within constraint 3, as `revolve` is now.                |
