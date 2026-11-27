# Specification: Unified Surface Model (material, coating, tolerance)

> Spec agreed 2026-11-26: every decision in section 6 was made that day. Section 4 records the
> experiments behind it, and [plan.md](./plan.md) the work in order.

## 1. Why

Three models describe the same surface, and none of them reads the others.

- **Material.** A material record carries a displacement and a build (row 2, `[transmission,
  absorptionDistance, displacement, build]`); `relief = displacement - build`
  ([material.gsl](../../../stdlib/material/material.gsl)). Since #162 a shape that sets a material
  with a height moves its own surface by the height it adds over its inherited material, before the
  booleans that use it ([layers.go](../../../internal/assembler/property/layers.go)).
- **Coating.** `coating.levelled` adds a film's build to the base's and levels the relief beneath by
  `material.filmKeeps(wavelength, build)` ([material_composition.md](../../../doc/material_composition.md)).
- **Tolerance.** `mechanical/tolerance` returns ISO 286 limit deviations as scalars, and the geometry
  is drawn at one material condition: `examples/shaft_bearings.gsl` draws its 25 k5 seats at
  `tolerance.maxMaterial` ([tolerance.gsl](../../../stdlib/mechanical/tolerance/tolerance.gsl)).

What that costs today:

- **A coat's build changes the part's size, but only the viewer sees it.** `displacementScale()` is
  the viewer's knob in the march, 1 in the CPU renderer, and 0 for meshing, export and every other
  evaluation ([builtin.go](../../../program/preamble/builtin.go)). A part with 0.1 mm of paint exports
  at the size of the bare substrate, and `eval` and `certify` cannot see the coat.
- **The tolerance model knows nothing about coats.** A 25 k5 journal's zone is 9 µm on the diameter
  (+2 to +11 µm). A 0.1 mm paint coat moves the diameter by 0.2 mm, 22 times the zone.
- **The tolerance model knows nothing about relief.** A turned finish at 0.2 mm feed with a 0.8 mm
  tool nose leaves grooves 6.25 µm deep. That is more than k5's 4.5 µm radial zone at 25 mm, so the
  finish cannot hold the fit. That pair is what `examples/material_machined.gsl` turns its bar with.
- **The size depends on a condition chosen when the part is drawn.** Nominal, maximum material and
  minimum material are three questions about one part, but the drawing commits to one of them.

## 2. The Model

### 2.1 A layer

Every process and every coat is a **layer**, and a layer has four quantities:

| quantity | meaning | today |
| :--- | :--- | :--- |
| build `b` | the layer's mean offset, in mm, positive where it adds material | `material.build`, for films only; machining removes stock in the geometry, not the record |
| relief `r(p)` | the marks on the surface, measured from a datum (D1) | `material.relief`; the whole surface function |
| envelope `[r_lo, r_hi]` | the relief's extremes over the surface | not represented; interval arithmetic computes it once `fract` is tight (section 4) |
| spread `[s_lo, s_hi]` | how far the build, or the machined size, may vary from part to part | not represented |

Appearance reads the build and the relief. A stand-in already carries the relief's mean:
`machined.grooveFinish`'s `.simple` moves the surface by the mean depth, `f^2 / (12 r)`. Tolerance reads
the build, the envelope and the spread.

### 2.2 The surface as made

#162's construction is kept as it is. For a shape whose layers stand on a body `d_nom`, the surface
as made is:

```text
d_made(p) = d_nom(p) - H(p) - sum of each layer's build variation
H(p)      = displacement(top material) = sum of builds + the relief that shows
```

`H` already telescopes. #162 moves each shape by `inc = height(own) - height(inherited)`, and
`coating.levelled` adds each film's build to the one beneath and passes on only the relief its
`filmKeeps` share leaves. The new piece is the spread. Each layer's build varies inside its own range,
and the variations add.

### 2.3 Condition: a question asked of the part, not a choice made when drawing it

The part is drawn at its **nominal** size, and the material condition becomes a parameter of the
evaluation, as `displacementScale()` is today:

| condition | builds | relief | used by |
| :--- | :--- | :--- | :--- |
| display | nominal | the relief function, scaled by the Detail knob | viewer, CPU renderer (today's behaviour) |
| nominal | nominal | off | mass properties; mesh export and section mode with the build flag on (D2) |
| maximum material | top of each spread | at the envelope's outermost point | envelope claims, assembly clearance |
| minimum material | bottom of each spread | at the envelope's innermost point | least-material claims, wall thickness |

This needs **build and relief scaled apart**. `displacementScale()` scales both today, through
`disp(v) = under(v) - displacementScale() * inc(v)`. The proposal is a separate scale for build,
set by the condition, and `displacementScale()` for relief alone.

### 2.4 Tolerance as a property of the shape

A zone is a band of allowed offsets from the nominal surface, `[o_lo, o_hi]` mm, set the way a material
is set:

```gsl
// Illustrative: the names are not settled.
seat = journalAt(25.0);                                      // drawn at nominal, no condition baked in
seat.tolerance = tolerance.shaftZone(25.0, tolerance.k5);    // radial band [+0.001, +0.0055]
seat.material = turnedMat;                                   // its relief and envelope come from the process
```

**The general zone is ISO 1101 profile of a surface**: the band `|offset| <= t/2` about the nominal
field. In a unit-gradient field this band is just two offsets of the field. ISO 286 fits are the
case for cylinders, written as constructors that return a band. The two-point size ISO 14405-1 makes
the default is not the same check; the field checks the envelope (Ⓔ) and a local radial size, which
is what bearing seats call for.

Unit gradient is what makes an offset band equal the geometric zone. Where a field drops it at edges
(the treatments the UGF goal already makes exceptions for), the zone needs its own rule there. GD&T
has rules at edges too, so the two should be designed together (D4).

### 2.5 What the one model answers

1. **Render the part at a condition, or render its zone.** That needs the condition parameter and a band
   shader; there is no new data.
2. **Prove it conforms.** At maximum material, there is no material outside the band's outer offset,
   `max(o_hi - d_nom, d_made) >= 0`. At minimum material, there is material everywhere inside its
   inner offset, `max(d_nom - o_lo, -d_made) >= 0`. Both read offsets as distances, so both hold only
   where `d_nom` has unit gradient. `eval.Certify` proves both, given a slack for the resolution of
   the measurement, because a claim that touches its bound everywhere cannot be proven (section 4).
3. **Prove the worst case over every spread.** Make each spread a document parameter with a range, and
   `CertifyParams` / `CertifyOverSliders` prove the claim "for every value of every listed parameter
   simultaneously" ([certify.go](../../../eval/certify.go)). That is worst-case stack-up analysis by
   branch and bound, with no new record fields.
4. **Solve the inverse.** The size to machine before coating is the band less the builds. A finish is
   capable of a fit when its spread plus its envelope's width fits inside the band.

A statistical (root-sum-square) stack-up is a separate calculation over each spread's distribution,
not a proof. It is left out (D6).

## 3. What Exists

| piece | where |
| :--- | :--- |
| Material record, `build`, `relief`, `withBuild`, `addDisplacement` | [material.gsl](../../../stdlib/material/material.gsl) |
| `filmKeeps`, `coating.levelled`, flowed against conformal films | [material_composition.md](../../../doc/material_composition.md) |
| Displaced distance where a material is set, layer shading | [layers.go](../../../internal/assembler/property/layers.go), [layers_height.go](../../../internal/assembler/property/layers_height.go) |
| `displacementScale()`, `pixelSize()`, `fadeDetail` | [preamble/builtin.go](../../../program/preamble/builtin.go), [texture.gsl](../../../stdlib/texture/texture.gsl) |
| Machined relief from the process (`nosePitch`, `f^2 / (8 r)`), stand-ins at the mean | [machined.gsl](../../../stdlib/material/process/machined/machined.gsl) |
| ISO 286 deviations and grades | [tolerance.gsl](../../../stdlib/mechanical/tolerance/tolerance.gsl), [tolerance-tables-sources.md](../../../doc/tolerance-tables-sources.md) |
| Bearing seats at a condition | [bearing.gsl](../../../stdlib/mechanical/bearing/bearing.gsl) (`seatDia`, `housingDia`) |
| Sound proofs over space and parameter ranges | [certify.go](../../../eval/certify.go) (`Certify`, `CertifyParams`, `CertifyOverSliders`) |

## 4. Findings (2026-11-26)

A standalone document posed a 25 k5 journal (radius 12.5010 to 12.5055 mm) as a field: the drawn
cylinder, minus the turned material's own height (`material.displacement` of `machined.turned`), plus a
coat's build. It was checked with `gsl eval`'s certify op. The displacement had to be written out by
hand, because certify evaluates an expression as a plain field and `displacementScale()` is 0 there.
The turned relief is measured down from the cusps, so drawing at maximum material puts the cusp tips on
the envelope.

| claim, over a 17 µm × 0.1 mm × 0.5 mm patch | verdict | interval evaluations |
| :--- | :--- | ---: |
| groove bottoms stay outside the least-material cylinder (0.1 mm feed, 0.4 mm nose, 3.14 µm deep) | proven | 169 |
| the same, 0.2 mm feed, 0.8 mm nose (6.25 µm deep) | disproven, with a box | 5,145 |
| cusp tips stay inside the envelope, no slack | undecided | budget |
| the same, 0.1 µm slack | proven | 41,657 |
| 5 µm of plating, drawn at maximum material | disproven | 2,217 |
| 5 µm of plating, machined 5 µm under | proven | 41,625 |

**The bound that mattered was `fract`'s.** The fidget path built `fract(x)` as `x - floor(x)`, and
interval arithmetic takes the two x's as independent: over z in [10, 11] it bounds `fract(z / 0.1)` by
[-10, 10], so the groove's depth bound fell to [-0.397, +0.003] mm. An envelope claim needs the
depth's lower bound, so certify had to subdivide below the feed everywhere. A least-material claim only
needs the upper bound, which is why it was always cheap. Building `fract` as fidget's modulo
(`rem_euclid`, tight to [0, 1]) gives exactly [0, 0.00314] and changes no value. That change is
prepared separately (branch `blake/tight-fract`).

| envelope claim, whole 15 mm seat, profile plane | `x - floor(x)` | modulo |
| :--- | ---: | ---: |
| today's octree search | undecided at 2,000,000 | proven, 43,689 |
| split by interval gradient (width × the magnitude of `expr.dp`'s interval), prototype | undecided at 2,000,000 | proven, 7,690 |

Also found:

- **Tight bounds are the lever, not the search.** With modulo `fract`, the octree's cost on this claim
  stops growing with the seat's length. The interval-gradient split is about six times cheaper at
  15 mm, grows linearly, and is worse on least-material claims. It is a later, optional change.
- **`.dp` fails through material records** ("collection elements must be all scalars or all vectors").
  Anything that needs the gradient of the surface as made, from an interval-gradient search to a check
  of unit gradient, needs it to pass through.
- **The claim in 3D over the full circle stays undecided** in every mode. Posing it in the profile
  plane (`r = abs(p.x)`) is what decides it, and a revolved package (`shaft.body`) can guarantee that
  symmetry itself.
- `pixelSize()` is 0 outside the renderer, so `fadeDetail` is 1 and eval sees the full relief.

## 5. Work, in Order

1. **Tight `fract`** (prepared). Test baseline outstanding: on a stale dylib, `cmd/compile`'s
   `TestRun_STL_MarchingCubesIsWatertight` failed and `eval` timed out. Neither is yet compared with
   unchanged `origin/main`.
2. **Split build from relief** in the displaced distance: a build scale set by the condition, and
   `displacementScale()` for relief alone. Per D2, flags for build and for relief on mesh export and
   on the viewer's section mode, whose shaders today show the field as built.
3. **Let eval see the surface as made**: a condition for an evaluation, so `certify`, `eval`, `bounds`
   and `volume` read #162's displaced distance and do not have to rebuild it by hand.
4. **Envelopes**: each process declares its relief's extremes where it knows them (`machined` does:
   `[0, f^2 / (8 r)]` below the cusps), and interval evaluation checks the declaration.
5. **Spreads** as process and coat parameters with ranges (D3).
6. **`.tolerance`**, zone constructors (profile band; ISO 286 as bands), and the claims of section 2.5
   generated from it, with a slack for measurement resolution. Per D5, the condition arguments of
   `tolerance` and `bearing` are replaced, with a `gsl fix` migration.
7. **A datum audit** of the library's reliefs (D1), and a tag on each process saying whether its
   heights are measured (machined) or judged for appearance (wood, `filmKeeps`'s fitted coefficient).
8. **`.dp` through material records.**
9. Optional: the interval-gradient split as a certify option.

## 6. Decisions Needed

| # | question | options | recommendation |
| :--- | :--- | :--- | :--- |
| D1 | What is relief measured from? | the outermost envelope (the peaks) / the mean line | **Decided 2026-11-26: the peaks.** `machined` already works this way, a fit binds at the peaks, and the mean is what each stand-in carries |
| D2 | Do mesh export and the section view include displacement? | no (today: both show the field as built, `gsl-displace.js`) / a flag | **Decided 2026-11-26: flags in mesh export and in section mode, separate ones for build and for relief** |
| D3 | Where do spreads live? | document parameters with ranges / a new record row | **Decided 2026-11-26: ranged parameters.** `CertifyParams` already proves over them, and the record stays as it is |
| D4 | What is the primitive zone? | ISO 286 size / profile of a surface | **Decided 2026-11-26: profile bands fixed in the part's frame**, a for-all claim certify proves. ISO 286 fits are band constructors on features a package knows (shaft diameters, bores). Floating, form-only zones (a there-exists-for-all claim) are deferred. The edge rule is designed with the UGF edge exceptions: a true distance offsets a convex edge into the round GD&T's zone has, and a `max` of planes overstates the zone there |
| D5 | What happens to `maxMaterial` / `minMaterial` arguments (`shaftDia`, `holeDia`, `bearing.seatDia`, `bearing.housingDia`)? | keep / deprecate later / replace | **Decided 2026-11-26: replaced in this track.** The condition arguments go, callers draw at nominal and set `.tolerance`, and `gsl fix` migrates them. A breaking change |
| D6 | Is a statistical stack-up in scope? | yes / later | **Decided 2026-11-26: later.** Worst case by proof comes first |
| D7 | Which surface is drawn: the substrate (coats grow out from it) or the finished surface (the cut derived as finished less the builds, "dimensions apply after plating")? | substrate / finished / per feature | **Decided 2026-11-27: the substrate, as today.** A coat is a deviation from nominal, not a tolerance: its build is a known offset applied in every condition but as-built, and only spreads are tolerance. Coats are laid on the cut's surface (a chain telescopes onto it), so a cut's size spread, a ranged parameter in that surface, moves every coat on it together, correlated. "After plating" sizes are reached by solving the inverse (section 2.5, item 4): machine to the band less the builds |

## 7. Constraints

- `omega-core` stays decoupled from the `omega` application (AGENTS.md). The condition is an evaluation
  option configured with `With...`, not a viewer concept.
- A document that sets no `.tolerance` and no build compiles exactly as before.
- The WGSL viewer path changes only through the condition knob and, later, a band view.
- Existing tests keep their fixtures and their goldens; any golden this track moves is shown and
  agreed before it is regenerated.

## 8. Verification

- The journal cases of section 4 as certify tests: proven, disproven and undecided at the stated slack,
  each over the patch and the whole seat. New tests, added once agreed.
- Export at each condition: the volume of a coated block grows by its area times its build, to the
  mesher's tolerance.
- No change to any existing golden from steps 1 to 3 with the D2 flags off.
