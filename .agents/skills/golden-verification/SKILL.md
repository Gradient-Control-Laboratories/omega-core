---
name: golden-verification
description: Verifies that any changes or regenerations to compiler test golden files (.wgsl, .glsl, .js, .assembled, etc.) are mathematically equivalent to prior goldens or represent documented, intentional improvements. Use whenever updating, modifying, or reviewing test goldens across compiler passes, standard libraries, and backend code generators.
---

# Golden Verification & Mathematical Equivalence Skill

When compiler passes, standard library functions, or backend code generators are refactored or
enhanced, downstream test golden files (such as `.wgsl`, `.glsl`, `.js`, and `.assembled`) may
change.

This skill defines the standardized protocol for evaluating, categorizing, and proving the
**mathematical and functional equivalence** of updated golden files so no regressions or unintended
numerical deviations enter the codebase.

---

## 1. When to Apply Golden Verification

Apply this verification protocol whenever:
- A new compiler pass, AST transformation, or optimization is added (e.g., closure
  monomorphization, constant folding, deadcode removal).
- Standard library geometry or math primitives are modified (e.g., SDF equations in `stdlib/`).
- Code generation backends (`backend/wgsl`, `backend/audio`, `backend/mesh`, `backend/tape`,
  `internal/fidget`) are updated.
- A Conductor track includes a phase for golden verification, or a track review touches modified
  test goldens.

---

## 2. Verification Protocol

### Step 1: Detect & Inventory Golden Changes

Run `git diff` against the target base revision to list all modified, added, or deleted golden
files:

```bash
git diff --name-status <base_revision> HEAD | grep -E '\.(wgsl|glsl|js|assembled)$'
```

### Step 2: Categorize Golden Modifications

Group every changed golden file into one of the following standard categories:

1. **Monomorphization / AST Inlining / Optimization**:
   - Higher-order functions, lambdas, or closures specialized into static first-order routines
     (`gcl_closure_*`).
   - Constant propagation or deadcode removal of unused variables/imports.
   - *Requirement*: Prove that the generated instruction sequence evaluates to the exact same
     symbolic mathematical expression as the un-inlined or dynamic version.

2. **Geometric Primitive Enhancements**:
   - Updates to standard library signed-distance fields (SDFs) or blending operators.
   - *Requirement*: Formulate the mathematical equations before vs after (e.g., Euclidean $L_2$
     norm vs $L_\infty$ bounding approximation) and demonstrate that the update is exact,
     continuous, and strictly superior.

3. **Lattice, Gallery, and Kernel Dispatchers**:
   - Parameterized procedural lattices (beam, foam, honeycomb), TPMS surfaces, or font rasterizer
     kernels.
   - *Requirement*: Confirm spatial periodicity, strut blending, and glyph rasterization formulas
     remain identical without unintended coordinate offsets.

4. **Backend Audio / Signal Generators**:
   - JavaScript audio code generation from GSL audio graphs.
   - *Requirement*: Confirm that synthesized audio sample calculations evaluate to identical
     floating-point values.

5. **AST Snapshots & Serializations**:
   - Deadcoded AST rebuilds (`.assembled`, `.gsl`).
   - *Requirement*: Confirm that structural deadcode elimination retains all reachable
     declarations.

---

### Step 3: Author the Verification Report

Author `golden_verification_report.md` (placed in the active track's directory, e.g.,
`conductor/tracks/<track_id>/golden_verification_report.md`):

The report MUST include:
1. **Executive Summary**: Overview of changes and overall verification verdict.
2. **Category-by-Category Mathematical Analysis**:
   - Affected golden files listed per category.
   - Detailed symbolic and numeric proofs of equivalence or documented improvements.
3. **Summary Table**:
   - Column 1: Category Name
   - Column 2: Golden Count
   - Column 3: Equivalence Status (`Identical`, `Numerically Superior`, etc.)
   - Column 4: Notes
4. **Determinism & Suite Sign-Off**:
   - Verification that running tests repeatedly produces 100% deterministic bitwise output.

---

## 3. Reviewer Checklist

When acting as a code reviewer or running `conductor-review`:
- [ ] Confirm `golden_verification_report.md` exists if any goldens were modified.
- [ ] Cross-check the git diff against the files listed in the report.
- [ ] Verify that every mathematical claim in the report is sound.
- [ ] Ensure all tests pass across all compiler packages (`./tool/test.sh`).

