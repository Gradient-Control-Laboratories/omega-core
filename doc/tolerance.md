# Tolerances, Coats and Requirements

GSL describes what you make, not what a drawing asks for. A part is drawn at the size each operation
makes it to. Coats and finishes are layers on that part. Each operation's scatter is a **working
tolerance** on the number it uses. A **requirement**, what the part must do, is a claim you ask of
that description and prove with `certify`. This page shows how the pieces fit. The design and its
decisions are in [the track's spec](../conductor/tracks/unified_surface_20261126/spec.md).

## The surface as made

A material carries three things that move a surface:

| | what it is | where it is set | moves the surface by |
| :--- | :--- | :--- | :--- |
| build | the film a coat stands proud by, mm | `material.withBuild(m, t)`, `coating.levelled` | `buildScale()` |
| relief | the marks on a surface, measured down from their peaks: 0 at the peaks, negative into the part | `material.withDisplacement`, `addDisplacement`, the process libraries | `displacementScale()` |
| relief depth | how far the relief reaches below its peaks, its whole envelope | `material.withReliefDepth(m, d)`, read by `material.reliefDepth(m)` | (read by requirements) |

The drawn shape is the substrate, the cut. Coats grow out from it, and a coat laid on a shape is moved
by what its material adds over the one beneath, so coats stack and a cut taken after them stays flat
(see [material_library.md](material_library.md), "Relief"). The process and coat families
(`machined`, `cast`, `printed`, `molded`, `sheet`, `coating`) declare their relief depths. A random
relief such as a sand casting's has no highest point, so its peak is the one a profilometer reports:
the mean peak height over five sampling lengths, Rpm (ISO 21920). Bulk and collection materials
(wood, stone, water, studs) are judged for appearance and are not measured from their peaks; the
[datum audit](../conductor/tracks/unified_surface_20261126/datum_audit.md) lists every one.

### Conditions

A field query reads the part at a **condition** (`eval.Evaluator.SetCondition`, or `"condition"` on a
`gsl eval` request):

| condition | builds | relief |
| :--- | :--- | :--- |
| `as-built` (default) | off | off: the field as the document builds it |
| `nominal` | on | off |
| `as-made` | on | on: the part as the CPU renderer draws it |
| `max-material` | on | at its peaks (the same surface as `nominal`) |

`min-material` is not a field. A feature's relief depth is a constant, so a least-material claim
shifts its band by `material.reliefDepth` instead (see below). Eval, Extremum, Certify, Bounds,
Enclose, Volume, SurfacePoint and Mesh read the condition; render, describe and shader do not. A
`.d(x)` or `.dp` in a query's expressions is taken of the field the condition reads, coats included.

```json
{"op": "certify", "source": "journal.gsl", "expr": "envelope",
 "region": [12.495, -0.05, 0, 12.512, 0.05, 0.5], "threshold": -0.0001, "condition": "as-made"}
```

## Working tolerances

A working tolerance is the range an operation's number may take: a bore's radius as bored, a coat's
thickness, a shaft's radius as turned. It is set as a property of that scalar, as offsets from its value:

```gsl
boreR = 12.51175;
boreR.tolerance = tolerance.symmetric(0.003);        // ±3 µm
zinc = 0.0065;
zinc.tolerance = tolerance.limits(-0.0015, 0.0015);  // 5 to 8 µm
```

`x.tolerance` reads back `[lo, hi]`, so `boreR + boreR.tolerance[0]` is the smallest bore. A limit may
vary over the part (a field of `p`), as plating builds heavier at edges.

ISO 286 sizes give the size a part is made to, the zone's middle, and the zone as its working tolerance:

```gsl
seat = tolerance.shaftDia(25.0, tolerance.k5);         // 25.0065
seat.tolerance = tolerance.shaftZone(25.0, tolerance.k5);   // ±4.5 µm
```

`holeDia` / `holeZone`, `bearing.seatDia` / `seatZone` and `bearing.housingDia` / `housingZone` work
the same way. The limit deviations themselves stay available (`tolerance.shaftUpper`, `shaftLower`,
`holeUpper`, `holeLower`).

### Propagation

A value that declares no tolerance reads its own, to first order from the document's working
tolerances: the sum of `y.d(x)` times each tolerance's midpoint (its bias), plus and minus the sum of
`|y.d(x)|` times each half-width.

```gsl
finished = boreR - zinc;   // finished.tolerance reads [-0.0045, 0.0045]
```

A shape's tolerance is a field. On the plated block above, the bore face reads the boring and the
zinc (±4.5 µm as made), and the outside reads the zinc alone (±1.5 µm), because a boolean takes the
derivative of the face that wins at each point. First order is exact where a value is linear in its
tolerances, as sizes and coats are. Near a crease it is an estimate, and certify is the referee.

## Requirements as claims

A requirement is a claim that `certify` proves `>= 0`, less a slack for what a measurement can
resolve, over the region it names:

```gsl
refField = r - 12.5;                          // the nominal 25 mm cylinder's field
band = tolerance.shaftBand(25.0, tolerance.k5);    // [ei / 2, es / 2] about it
most = tolerance.mostMaterialClaim(plated, refField, band[1]);
least = tolerance.leastMaterialClaim(plated, refField, band[0]);
edge = tolerance.gradientClaim(refField, 0.0001);
```

`mostMaterialClaim` says there is no material beyond the band's outer edge. `leastMaterialClaim` says
there is material everywhere inside its inner edge. `holeBand` gives a hole's band about the field
of the material round it. Declare each claim on its own: a value destructured from a call
(`a, b = f(...)`) is not moved by its coats, so it reads as built.

Certify over the document's tolerances to prove a claim for every value each takes between its
limits (`eval.CertifyOverTolerances`, or `"overTolerances": true`). A disproof names each
tolerance's failing share, from 0 at its low limit to 1 at its high one:

```json
{"op": "certify", "source": "fit.gsl", "expr": "max(plated, shaft)", "threshold": 0.0032,
 "region": [12.48, -0.2, -5, 12.53, 0.2, 5], "condition": "as-made", "overTolerances": true}
```

On the H7/g6 plated bore this disproves a 3.2 µm clearance at the worst corner: the bore bored
smallest, the zinc and the shaft largest.

**The edge rule.** A band only means "within t of the surface" where the reference field is a true
distance. Certify `gradientClaim` over the region first. It does not see every field that is not a
true distance: a `max` of planes keeps unit gradient up to its edges, where its offsets are sharp
corners and overstate the zone. Read a band near a convex edge from an exact round.

## In the viewer

- The **Section** panel's **Coatings** and **Relief** toggles draw the cut with builds and relief on.
  **Band** shades where each surface may lie between its most and least material, at least a pixel
  wide.
- **Render › Detail › Tolerance** draws a part with working tolerances at its **Most material**,
  **Nominal** or **Least material**, and **Exaggerate** (x1 to x1000) scales the offsets and the band,
  so micrometre tolerances read at a glance.
- The viewer's tolerance view takes its derivatives as built, so a coat's thickness tolerance does not
  show in it; `gsl eval` queries count it. The CPU device (Fidget only) does not take the view.

## Not yet

- Statistical stack-up (sigma, Cpk, ASME's statistical tolerance) is planned later; tolerances are
  worst-case limits.
- Datums and setups (where a part is located for each operation) and assembly variation (mates, datum
  reference frames, floating zones) are proposed as their own tracks.
