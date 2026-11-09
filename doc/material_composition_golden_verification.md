# Golden Equivalence & Mathematical Verification Report

**Work**: Material composition, step 1 — relocate the pure surface-composition
operators from `material/bulk/wood` to `material`
**Date**: 2026-10-11
**Status**: Verified & Equivalent

No conductor track exists for this work, so this report sits beside the design
it belongs to (`doc/material_composition.md`) rather than under
`conductor/tracks/<track_id>/`.

---

## 1. Executive Summary

Three scalar operators moved from `stdlib/material/bulk/wood/wood.gsl`
into `stdlib/material/material.gsl`, where they are no longer
specific to wood:

- `filmKeeps(wavelength, build)` — how much of a relief a film leaves
- `hiddenShare(fade, standIn)` — how much relief is folded into the lobe
  rather than drawn (reshaped to take the fade as an argument, because
  `texture` imports `material` and so `material` cannot call `texture` back)
- `roughnessFromGloss(gloss)` — the ISO 2813 gloss-to-roughness fit

`wood.gsl` now delegates to all three. Nothing else moved: `hiddenSlopeFinish`
stays in `wood.gsl` — see section 5.

Five WGSL example goldens changed. **All five are mathematically equivalent**:
two function renames, one beta-substitution, and one blank line. No function
was lost from any golden, and no arithmetic changed anywhere.

---

## 2. Inventory of Golden Changes

```
M  backend/wgsl/internal/lib/internal/codegen/testdata/example/Material - Knotty Planks.wgsl
M  backend/wgsl/internal/lib/internal/codegen/testdata/example/Material - Spalted and Wormy.wgsl
M  backend/wgsl/internal/lib/internal/codegen/testdata/example/Material - Wood Finishes.wgsl
M  backend/wgsl/internal/lib/internal/codegen/testdata/example/Material - Wood Planks.wgsl
M  backend/wgsl/internal/lib/internal/codegen/testdata/example/Material - Wood Species.wgsl
```

`UPDATE_GOLDEN=1 go test ./backend/wgsl/internal/lib/internal/codegen/` touched exactly these
five and no others. In particular every non-wood golden is byte-unchanged —
`material_only`, `material_remap`, `material_blend`, `color_and_material`,
`Drill` and `Thread_Relief` among them — which is the evidence that the three
operators cost nothing to documents that do not use them.

| golden | before | after | delta |
| --- | --- | --- | --- |
| Material - Knotty Planks | 96,985 B | 97,185 B | +200 |
| Material - Spalted and Wormy | 228,502 B | 228,702 B | +200 |
| Material - Wood Finishes | 108,236 B | 108,436 B | +200 |
| Material - Wood Planks | 105,284 B | 105,484 B | +200 |
| Material - Wood Species | 106,349 B | 106,357 B | +8 |

The largest relative growth is 0.21%.

---

## 3. Category-by-Category Mathematical Verification

The change is a **symbol relocation** — category 1 of the protocol (AST
inlining / optimization), in its simplest form: no pass was added, no primitive
changed, and no backend was touched.

### A. Renamed functions — byte-identical bodies

`wood_filmKeeps` → `material_filmKeeps` and `wood_roughnessFromGloss` →
`material_roughnessFromGloss`. After applying that rename to the old goldens,
the function bodies compare byte-identical, e.g.

```wgsl
fn material_roughnessFromGloss(gloss: f32, noiseTex: texture_3d<f32>) -> f32 {
  let g = clamp(gloss, 0.5, 90.0);
  return clamp(0.647 - 0.1243 * log(g) + 0.008 * log(1.0 - g / 91.5), 0.045, 1.0);
}
```

The emitted prefix is the only difference, which is where the +200 bytes come
from (`material_` is longer than `wood_`, over every call site).

**Verification**: comparing the set of `name -> (signature, body)` pairs across
all five goldens after the rename gives, in every file, *no function present in
the old golden and absent from the new*. Arithmetic is untouched.

### B. `hiddenShare` — exact by beta-substitution

The one genuine restructuring. Before, one function did the whole job:

```wgsl
// old wood_hiddenShare(size, standIn)
return select(1.0 - texture_fadeDetail(size, noiseTex) * gsl_dispDrawn, 1.0, standIn > 0.5);
```

After, the fade is computed by the caller and the decision by the callee:

```wgsl
// new wood_hiddenShare(size, standIn)
return material_hiddenShare(texture_fadeDetail(size, noiseTex), standIn, noiseTex);

// new material_hiddenShare(fade, standIn)
return select(1.0 - fade * gsl_dispDrawn, 1.0, standIn > 0.5);
```

Substituting `fade := texture_fadeDetail(size, noiseTex)` into the callee
reproduces the old body **exactly, term for term**. This is the only added
function (`material_hiddenShare`) and the only changed body
(`wood_hiddenShare`) in any of the five goldens.

### C. Residual text — one blank line

After excluding the two functions above, the remaining difference in the
non-function text of each golden is a single blank line (578 → 579 bytes of
non-function text in *Material - Wood Finishes*), an artifact of functions being
emitted at a different file position now that they belong to the `material`
package. No declaration, entry point or uniform changed.

`Material - Wood Species` has no `hiddenShare` at all: its function set is
*identical* and its non-function text is *identical*, leaving +8 bytes of pure
reordering. It is exactly equivalent with no substitution argument needed.

---

## 4. Independent Runtime Verification

Equivalence was also checked at runtime, before the goldens were regenerated
and independently of them, with the `gsl` CLI over a board carrying four preps
across the parameter space:

- preps: `sanded(180)`, `sawn([1,0,0], 0.4)`, `tableSawn([1,0,0], 0.05)`,
  `planed([1,0,0], 1.6)`
- coats: `oilPolyurethane(45)` (a film) and `tungOil` (a penetrating oil), so
  `filmKeeps` is exercised at two very different builds
- species: walnut and white oak
- 4 points, 6 accessors each — `roughness`, `anisotropy`, `displacement`,
  `lay`, `metallic`, `ior`

The change was stashed and restored with a path-scoped stash to isolate it.
Output was **byte-identical** across all 96 values and their gradients.

---

## 5. What was deliberately NOT moved

`hiddenSlopeFinish` (the slope-variance tensor that becomes the lobe's lay)
stays in `wood.gsl`. Moving it to `material.gsl` as `slopeLobe` was tried and
reverted: merely *declaring* it there makes material row 3 — the lay row —
live in **every program that imports `material`**, including programs that
never call it. `material_only.gsl` grew 26,808 → 35,209 bytes, +31%, gaining
two functions reading `material_chrome[3]` and `material_plastic[3]` for a
function that document does not call.

Bisected: `filmKeeps` + `hiddenShare` + `roughnessFromGloss` leave every
non-wood golden untouched; adding `slopeLobe` moves `material_only`,
`material_remap`, `color_and_material`, `Drill` and `Thread_Relief`. The
`displacementDrawn()` builtin in `hiddenShare` was tested first and is not the
cause.

That is a size regression across every document that uses a material, not
golden churn, so a function reaching the lay row does not belong in
`material.gsl` at all. It stayed in `wood`, and `material/process/machined.layFinish`
is the equivalent the repo already had for metals — see
`doc/material_composition.md`, which was rewritten after that library was
found. The measurement stands on its own regardless: **declaring** a function
that calls `withLay` in a widely imported library costs every document that
imports it, whether or not it composes a surface.

---

## 6. Summary Table

| Category | Goldens | Equivalence Status | Notes |
| --- | --- | --- | --- |
| A. Symbol relocation (rename) | 5 | Identical | `wood_*` → `material_*`; bodies byte-identical after rename |
| B. `hiddenShare` split | 4 | Identical | Exact by beta-substitution of the fade argument |
| C. Emission order | 5 | Identical | One blank line; no declaration or entry point changed |

No golden lost a function. No arithmetic changed. No numerical improvement is
claimed, because none was intended.

---

## 7. Determinism & Suite Sign-Off

- `go test ./backend/wgsl/internal/lib/...` — green after regeneration
- `go test ./internal/preamble/...` — green
- Re-running the codegen suite reproduces the regenerated goldens bitwise, with
  no further diff.

The full `tool/test-local` gate has not been run as part of this report.
