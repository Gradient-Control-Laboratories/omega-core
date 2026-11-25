# GSL v1 to GSL v2 Migration Guide

This document describes the differences between the **GSL v1 Language Specification**
([gsl_lang_spec_1.md](gsl_lang_spec_1.md)) and the **GSL v2 Language Specification**
([gsl_lang_spec_2.md](gsl_lang_spec_2.md)).
It details mechanical syntactic conversion approaches for migrating existing GSL source code over time,
and tracks what migrations have been completed versus those currently underway.

---

## 1. Migration Strategy: Mechanical Syntactic Conversion

To migrate a large corpus of GSL programs from v1 to v2 safely and incrementally,
a phased, automated syntactic migration strategy is preferred:

1. **Dual-Syntax Acceptance in Parser**:
   During the transition window, the compiler accepts both legacy v1 syntax and target v2 syntax,
   normalizing both to a unified AST representation.
2. **Parser-Returned `migratedSource` ("Lazy Migration")**:
   When the parser encounters legacy v1 constructs in a source file, it can return an updated
   `migratedSource` string containing the canonical v2 representation alongside the parsed AST.
   If the file is already compliant with GSL v2 and requires no migration, `migratedSource` is unset
   (empty). Host applications (such as editors, document save pipelines, or batch tools) can
   inspect this field to automatically rewrite files on disk or update editor buffers with zero
   runtime friction.
3. **Automated Source Formatter / Migration CLI (`gsl migrate` / `gsl fmt`)**:
   Mechanical regex- or AST-based rewriting tools update source files in-place without human error.
4. **Deprecation Warnings with Actionable Suggestions**:
   When legacy v1 syntax is encountered, the parser issues warnings or suggestions
   (e.g., `warning: 'uiPoint(v)' is deprecated, prefer '@point' annotation`).
5. **Final Sunset**:
   Once all repository programs, examples, and user documents have been mechanically converted,
   the legacy syntax rules and fallback parser logic are retired in GSL v2.

---

## 2. Migration Status & Progress

| Feature / Syntax Element | GSL v1 (Legacy) | GSL v2 (Target) | Status | Details |
| :--- | :--- | :--- | :--- | :--- |
| **Logical Negation** | `not(a)` | `!a` | **Completed** | Unary `!` operator is fully supported in current parser grammar. |
| **Logical Conjunction / Disjunction** | `and(a, b)`<br/>`or(a, b)` | `a && b`<br/>`a \|\| b` | **Completed** | Binary `&&` and `\|\|` operators are fully supported in current parser grammar. |
| **Trailing Commas** | Rejected in some contexts | Permitted across all lists | **Completed** | Trailing commas supported across arguments, parameters, returns, and collections. |
| **Nested Ternaries** | Required parentheses | Right-associative with AST rotation | **Completed** | `fixTernaryAssociativity` automatically balances unparenthesized nested ternaries. |
| **Scene Framing** | `// Scene Size: S`<br/>`// Scene Grid: G` | `@scene(S, G)` | **Underway** | `@scene` directive is implemented in parser; legacy header comments are retained as fallback for existing models. |
| **Point Viewport Manipulator** | `uiPoint(pos)` | `@point` preceding assignment or expr | **Underway** | `@point` annotation on variable assignments is supported; rewrites internally to `uiPoint`. Inline expr annotation planned. |
| **Transform Manipulator** | `uiTransform(m)` | `@transform` preceding assignment or expr | **Underway** | `@transform` annotation on variable assignments is supported; rewrites internally to `uiTransform`. Inline expr annotation planned. |
| **Expression Depth Limit** | 127 expressions | Arbitrary depth | **Underway** | Participle parser limit lifted in newer pipeline passes. |
| **Slider Metadata** | `x.min = 0;`<br/>`x.max = 10;`<br/>`x.period = 4;` | `@range(0, 10, 4)` | **Planned** | Dot-property assignments are accepted in v1; `@range` parser & lowering planned for v2. |
| **Discrete Value Choices** | N/A (or multiple conditionals) | `@values(v1, v2, ...)` | **Planned** | `@values` annotation defined in spec; parser implementation planned for v2. |
| **`reduce` Argument Order** | `reduce(init, fn, coll)` | `reduce(fn, coll [, init])` | **Planned** | Current parser requires `reduce(fn, numeric [, initial])` in newer pipeline; migration tooling needed for legacy files. |
| **Asset Imports** | `importMesh 0 "f.stl";`<br/>`importMesh("f.stl")` | `import name "f.stl";` | **Planned** | Pre-allocated statement and function import forms used in v1; unified asset imports specified for v2. |

---

## 3. Differences Between GSL v1 and GSL v2

| Status | Name | Spec |
| :--- | :--- | :--- |
| `started` | Scene Framing (`// Scene` Comments to `@scene` Directive) | internal |
| `started` | Viewport Manipulators (`uiPoint` / `uiTransform` to `@point` / `@transform`) | internal |
| `specified` | Slider Metadata (Dot-Syntax Properties to `@range` Annotations) | internal |
| `specified` | Discrete Choices (Ad-hoc Logic to `@values` Annotations) | internal |
| `completed` | Logical Operations (Functional Primitives to Operators) | internal |
| `specified` | `reduce` Argument Order Standardization | internal |
| `specified` | Unified Asset Imports Syntax | internal |
| `started` | Expression Depth Limit Removal & Iterative Evaluation | internal |

---

## 4. Summary of Migration Actions for Developers

When updating existing GSL code to be forward-compatible with GSL v2:
1. Prefer `@scene(...)` over `// Scene Size:` and `// Scene Grid:` comments.
2. Prefer `@point` and `@transform` annotations on variable assignments over `uiPoint(...)` and `uiTransform(...)`.
3. Use `!`, `&&`, and `||` instead of `not(...)`, `and(...)`, and `or(...)`.
4. Use `atan(y, x)` instead of legacy `atan2(y, x)`.
5. Prepare to transition slider property assignments (`.min`, `.max`, `.period`) to `@range(...)`.
