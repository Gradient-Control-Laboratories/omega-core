# Golden verification: material/colour pair functions

The metal, glass and physical libraries changed from 134 annotated **constants**
paired with their colour by name (`metal.brass` + `metal.brassColor`) to 134
zero-argument **functions** returning the pair (`metal.brass()` → the record and
the colour). 35 WGSL example goldens moved. This is the proof that nothing the
renderer sees changed.

## Executive summary

**Verdict: equivalent.** No material value and no colour value moved anywhere.
Every golden difference is constant folding that became possible because a
package constant is no longer a runtime parameter. Total golden size fell 9.28%
(1,421,383 → 1,289,481 bytes).

| Check | Result |
|---|---|
| Library members compared (material rows + colour) | 134 / 134 **bit-identical** |
| Changed goldens | 35 |
| Goldens with an identical set of numeric literals | 26 / 35 |
| Goldens whose literal set changed | 9 — each explained and verified below |
| Constant-folded sRGB `select` sites picking the spec-correct arm | 162 / 162 |
| Material or colour values that moved | **0** |

## 1. The root claim: library values are unchanged

Both forms were parsed and the two helper functions (`metal.finish(rough)`,
`glass.make(rough, index, absorptionDistance)`) expanded to the rows they
produce, then compared element by element against the new pair functions'
returns:

```
library members compared: 134
MATERIAL+COLOUR value mismatches: 0
```

Everything downstream follows from this plus an unchanged compiler: the same
numbers enter the same code generator.

## 2. Why the goldens move at all

Under the constant form, a package constant reached a library function as a
**captured parameter**:

```wgsl
fn natural_terrain(p: vec3<f32>, physical_grassColor: vec3<f32>, physical_sandColor: vec3<f32>,
                   physical_snowColor: vec3<f32>, physical_sand: mat4x4<f32>, …) -> gsl_struct_2
```

A parameter is opaque to the folder. Under the pair form the value is a literal
inside the body, so the parameter disappears and the arithmetic over it folds:

```wgsl
fn natural_terrain(p: vec3<f32>, scale: f32, snowBias: f32, …) -> gsl_struct_2
```

This produces exactly four kinds of textual change:

1. **A named constant and its uses collapse into the literal.**
   `let metal_brassColor = vec3<f32>(0.95, 0.9, 0.69); return vec4<f32>(metal_brassColor, 1.0);`
   becomes `return vec4<f32>(vec3<f32>(0.95, 0.9, 0.69), 1.0);`
2. **Parameter lists shrink**, as above.
3. **`gsl_param_<line>_<col>` slider ids renumber**, because the migration
   inserted destructuring lines at the top of each example and those ids are
   keyed to source position. A pure renaming; the ordered first-use mapping is a
   bijection in every file.
4. **Declaration order shifts** (e.g. `color_cbrt` moves later in
   `Terrain Planet`), which the golden comparator already treats as canonical.

## 3. The 26 goldens with an identical literal set

For 26 of the 35, the set of distinct numeric literals in the file is unchanged.
Inlining `metal_steel[0].y` as `1.0` adds an occurrence of a number the removed
`let metal_steel = mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), …)` already
carried, so counts change while the *set* does not. No constant appeared and
none vanished.

## 4. The 9 goldens whose literal set changed

### 4a. Five lose exactly one literal, and it was dead

In each case the lost number sat in a row that an override replaced wholesale,
so it never reached the result.

| Golden | Lost | Why it was dead |
|---|---|---|
| Material - Basics | `0.45` | `withRoughness(steelMat, 0.15)` overwrites row 0 x |
| Printer Gantry | `0.45` | same |
| Material - Machined | `0.35` | stainless's base roughness, replaced by the brushed finish's |
| Water Waves | `200.0` | `withTransmission` replaces row 2 with `(0.6, 24.0, …)` |
| Material - Canyon Strata | `1000.0` | `withTransmission` replaces row 2 with `(1.0, 4.0, …)` |

The last two were checked against the emitted branch, since losing an
absorption distance would be visible. Old and new agree exactly:

```wgsl
// OLD
let glass_water = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.333, 0.0), …, vec4<f32>(1.0, 200.0, 0.0, 0.0), …);
let gsl_m = mat4x4<f32>(glass_water[0], glass_water[1],
                        vec4<f32>(0.6, 24.0, glass_water[2].z, glass_water[2].w), glass_water[3]);
// NEW
let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.333, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0),
                        vec4<f32>(0.6, 24.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
```

`200.0` is overwritten by `24.0` in both; `glass_water[2].z`/`.w` are `0.0`.

### 4b. Four fold a colour-space conversion

`Material - Concrete`, `Material - Library`, `Terrain Bracket` and
`Terrain Planet` convert a library colour through `color.rgb2oklab`. With the
colour now literal, the whole chain folds. Verified numerically.

**Concrete** — `physical.concrete`'s linear grey is `0.5099919`, and all three
Oklab LMS rows sum to 1, so `l' = m' = s' = cbrt(0.5099919)`:

| Quantity | Computed | Golden |
|---|---|---|
| `L = cbrt(0.5099919)` | 0.798953 | `0.79895276` |
| `a = L·(1.9779985 − 2.4285922 + 0.4505937)` | 6.47e-11 | `0.00000000006467769` |
| `b = L·(0.025904037 + 0.78277177 − 0.80867577)` | 2.98e-8 | `0.000000029780091` |

The near-zero `a` and `b` are correct for a grey, and their exact magnitudes are
the floating-point residue of those two matrix rows — they could only arise from
folding with the same coefficients.

**Terrain** — the sRGB transfer function folds, emitting both arms:

| Quantity | Computed | Golden |
|---|---|---|
| `((0.4002 + 0.055)/1.055)^2.4` | 0.133009 | `0.13300854` (live arm) |
| `0.4002 / 12.92` | 0.0309752 | `0.030975232` (dead arm) |
| `((0.2238 + 0.055)/1.055)^2.4` | 0.041014 | `0.04101066` |

The "gained" small numbers are the *unused* linear arm of the branch.

### 4c. The folded branch picks the right arm

Both arms being present raised a question worth settling: the folded condition
prints as `select(A, B, 0.0 > 0.0)`, which is constantly false. If the condition
had lost its operands, a channel below the 0.04045 threshold would take the
wrong arm.

It has not. GSL has no booleans — a comparison yields 0 or 1 — so the folder
emits the *computed* result as `<0 or 1> > 0.0`. Both forms occur across the
goldens (`0.0 > 0.0` for false, `1.0 > 0.0` for true), so the fold is
direction-correct. Checking every site:

```
select() sites in the new goldens: 1076
  with a CONSTANT-folded condition: 162
  picking the wrong sRGB arm: 0
  channel values folded this way: min 0.223800  max 0.930900
  sRGB linear-arm threshold: 0.04045  -> margin +0.183350
```

All 162 agree with the sRGB specification, and the darkest channel folded this
way sits 0.18 above the branch point, so none is near it.

## 5. Suite sign-off

| Suite | Result |
|---|---|
| `./common/...` | ok (includes the closure pass, 2.0s) |
| `./agent/assembler/...` | ok |
| `./agent/wgsl/...` | ok (codegen 92s) |
| `./agent/js/...` | ok — unchanged, these goldens did not move |
| `./agent/fidget/gsleval/libflattest/` | ok — independently checks every material's colour stand-in against its sampled mean |
| `./agent/fidget/internal/mesh/` | ok |

`libflattest` is worth singling out: it samples each library material and
compares its declared stand-in against the mean of the real thing, number by
number, over roughly sixty library calls including every machined, coating and
printed family. It passing is a behavioural check on the library values that
does not go through the goldens at all.

## 6. Scope note

`Material - Coated.wgsl` was already regenerated earlier on this branch (for the
four levelling plates), so its baseline here is the index version rather than
`HEAD`, and the comparison isolates the pair-function change alone.
