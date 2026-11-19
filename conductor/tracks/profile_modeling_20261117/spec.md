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

   Decided 2026-11-18 (taken after C4 and C5): an edge is named by a descriptor, `edgeLine()`,
   `edgeArc(curvature)` or `edgeBezier(control)`, and a profile of mixed edges is `profileEdgesXY(points,
   edges)`, one descriptor per edge; a name of its own, since overloads are told apart by arity alone and
   `profileXY` already takes two and three arguments. Open chains too: `openProfileEdgesXY(points, edges)`
   and `sweepOpenEdges(section, points, edges)`, one descriptor per edge (one fewer than the points).
   Corners on these forms, and corners next to a Bezier, come later. A Bezier edge joins the three pools as
   an arc does: its distance is the least over the cubic solve's candidate parameters, each clamped to the
   edge and each a point on it, with its two ends, so no select chooses among them; its side is the sign
   of the implicit parabola it lies on, F(q) = D (g x q) - (q x f)^2 with g = 2 (control - start),
   f = start - 2 control + end, D = g x f and q = p - start, or the chord's side where D vanishes (a choice
   made by the points alone, so exact over any cell); it is held off past its ends by twice the overshoot,
   or gated by its parameter, whichever the reference and surface cells bear out. `.u` along it is its arc
   length, in closed form; its box takes the hull of its points and control.

   Found 2026-11-18 while building it. The overshoot penalty is not valid for a Bezier as it is for an arc:
   on its concave side the normals meet, and a point before an end normal can still be nearest a point
   within the edge (the mixed loop's, 0.94 mm off with the penalty in all three pools). Gating by "a point
   within the edge is nearer than both its ends" is exact, but where an end is nearest it compares two equal
   quantities, so no interval decides it, and certify failed everywhere. Kept: the penalty in the two sided
   pools only, the nearest pool taking the distance itself, so the magnitude is always exact and the sign
   exact on every reference. Over a cell the cubic's roots can span the edge, so its distance is helped
   without changing a value: three fixed points of the edge (t = 1/4, 1/2, 3/4) bound it above, and the
   triangle of its points and control below; the trigonometric root's divisor is held below zero, which
   ended a NaN in intervals. `.u` takes one Newton step on the root, which float32 gives to a part in ten
   thousand. Certify proves the test loops in the default budget; surface cells on a curved Bezier reach
   5.93 diagonals, allowed 6 by the owner (lines and arcs keep 4).

   Found by the owner in the viewer, the same day: the side from the parabola's sign was wrong. Past the edge
   the parabola swings back, and where the control polygon turns sharply its other arm passes between the
   edge and points near it, so the sign flipped inside the example's plate (24 mm off; the test loops had
   not had so tight an edge). The side is now the edge's own: concave where p is across the chord from the
   control, or within the triangle of the points and control and inside the parabola, which within the
   triangle is sound, since the edge leaves it along two of its sides. The plate is now a test loop. Each
   candidate root also takes one Newton step (float32 left 3e-4 mm on the tight edge), the roots as solved
   kept beside the stepped ones so the bounds do not widen. Its cells reach 9.18 diagonals; the owner set
   the limit for curved Beziers to 10. Two WGSL-only failures the CPU tests could not see, both found by
   compiling the example: a list indexed through a parameter (not a constant in WGSL), and an `||` mixed
   with `&&`, whose brackets the codegen drops and WGSL refuses.
4. **C7. Tighter arc (and Bezier) boxes.** The arc profile's box is the points' box padded by
   `min(span / 2, span^2 * maxCurvature / 4)`: sound, but loose, because a `.box` could not hold the reduce
   an exact box needs (see section 5).

   Decided 2026-11-19 (taken after C6): exact, per edge, for every form. An arc edge's box is its ends and
   whichever of its circle's four axis extremes lie on it, those on the side of the chord the arc bulges
   to; a Bezier edge's, its ends and its point at each axis's extremum, t* = (start - control) /
   (start - 2 control + end) where that is in (0, 1). The loop's box is the union over its edges, padded
   0.01 mm as now. The corner forms take their loop's box without the corners: a fillet, chamfer or blend
   lies within its corner's triangle, which that box holds. The open forms' bounds helpers, which the open
   sweeps box with, are exact likewise. All of it reads the points alone, so its selects are the same
   everywhere. Measured: the moulding's concave cove padded its box by 4.62 mm on every side, which every
   sweep box built from it carried.
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

   Decided 2026-11-18 (open paths): an open profile is `openProfileXY(points)` or
   `openProfileXY(points, curvatures)`, a chain of lines and arcs with one curvature per edge (one fewer than
   the points). Its two end edges are extended along their end tangents as rays, so the chain splits the
   plane and the closed profile's pools apply unchanged, the rays standing in for the closing edge and the
   end corners: the exact signed distance to the extended chain, positive to the right of travel (where a
   counterclockwise loop's outside is), so a piece of a closed path sweeps as the closed path does. That
   holds while neither ray crosses the chain or the other ray; a chain that turns past half a turn, a C,
   breaks the sign, which the doc comment states since GSL cannot refuse it. Its `.u` is the arc length to
   the nearest point of the extended chain, negative along the start ray and past the chain's length along
   the end ray. It has no box, the side it calls negative being unbounded, and `abs()` of it is a strip
   running out along both rays, documented; finite 2D strokes come later. Non-planar paths are split off
   into a proposal of their own, with the stated Lipschitz bound.
7. **C5. Caps for sweeps.** Decided 2026-11-18: `sweepOpen(section, points)` and
   `sweepOpen(section, points, curvatures)`, a name of its own because overloads are told apart by arity
   alone and `sweep(section, path)` takes two. The path is data and not a field: the caps need the chain's
   ends, and a body cannot read a field's `.u` (constraint 2). Flat caps: the magnitude is the least of the
   wall's distance, |section(across, z)| gated off where the nearest point of the extended chain is on a ray,
   and each cap's sqrt(h^2 + max(section(lateral, z), 0)^2) in its own plane; the sign is negative where the
   wall's is and the nearest point is on the chain. The section is read three times, as the partial revolve
   reads its profile, so its cost is timed on the GPU before it is called done. Frame and box by Go rules
   beside `sweep`'s.

   Found 2026-11-18, against the swept solid sampled on its own: this is not exact everywhere, as the
   proposal said, and not a bound. The partial revolve's symmetry puts the nearest wall point at a cap
   plane wherever the wall is gated off; a curved chain does not. Beyond a cap, another part of the sweep
   can be nearer than the cap, and the field reads the cap's distance: 1.73 mm over at 16 mm from the S's
   end, where the moulding's side on a leg is nearer. The sign stays right, so certify's proofs of empty
   space hold. Measured alternative: one read of the section cut by a slab coordinate,
   min(end-plane depth, (chain distance - ray distance) / 2), combined as `extrude` combines its slab. It
   was never over in either test (worst 0.0003 mm), but under by up to 3.8 mm beyond the ends, and its bound
   has an unproven gap for asymmetric sections. Decided by the owner: keep the three reads, exact near the
   chain and over the caps, with the condition in the doc comment. On a chain with sharp corners, surface
   cells near a corner's bisector are bounded soundly but wide (9.00 diagonals; the pools' sign test is
   undecided there), so only their soundness is asserted.

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
