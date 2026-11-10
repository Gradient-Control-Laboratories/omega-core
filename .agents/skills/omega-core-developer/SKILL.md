---
name: omega-core-developer
description: Guidelines and strict coding constraints for making changes anywhere in the omega-core repository, including diagram choices, GSL documentation, golden verification, and compiler pass separation of concerns.
---

# Omega Core Developer Skill

Use this skill whenever generating code or making changes in the `omega-core` repository.

## 1. Diagrams

There are two primary options for diagrams:
- Compact ASCII art flowcharts (preferred for simple/linear decision flows, state transitions,
  and directory/file trees). Compact ASCII diagrams do NOT use boxes (`+---+`, `|   |`);
  use clean text, branch tags (`[Yes]`, `[No]`), and directional arrows (`-->`, `|`, `v`).
- Mermaid diagrams rendered to PNG via `mmdc` (stored in relative `images/` directories)
  reserved for complicated, involved diagrams (e.g., multi-stage compiler pipelines, complex ASTs).
  - Embed the rendered image: `![Diagram Title](images/diagram.png)`.
  - Place the original Mermaid source in a collapsible `<details>` and `<summary>` block
    below the image.
  - Do NOT wrap Mermaid source in HTML comments (`<!-- ... -->`) to prevent syntax conflicts
    with `-->`.

## 2. Read the GSL Documentation

When writing or changing GSL (`.gsl`) code in `stdlib/` or `examples/`, consult the GSL
documentation in `examples/viewer/docs/` (generated from `program/preamble/`):
- `gsl_description.md` — the GSL language reference: syntax, SDF conventions, and how
  shapes/fields are built. Vectors are component-wise, so prefer vector math
  (`repeatCentered(p.xy, cell)`, `dot(v, v)`, `length(v)`) over per-component scalar code.
- `gsl_libraries.md` — the standard library package reference (`builtin`, `map`, `lattice/*`,
  `mechanical/*`, `material/*`, etc.). Prefer calling existing library functions over re-deriving
  them; if you add a helper, name it in the library's style.

## 3. Standalone SDK Boundary Invariants

`omega-core` (`module gradientcontrol.com/omega/core`) is a standalone, pure compiler and geometry
SDK:
- Never introduce dependencies on Stirigma, gRPC microservice dispatchers, cloud SDKs, or parent
  `omega` packages.
- Prefer functional options (`WithSomeOption`) on `parser`, `assembler`, `eval`, and `backend/*`
  rather than leaking internal AST passes (`internal/*`) into the public API.
- Propagate `context.Context` cleanly in CLI commands (`cmd/*`) and long-running servers
  (`cmd/serve`).

## 4. Golden File Verification & Equivalence Protocol

When modifying compiler passes, AST transformations, standard library math primitives (`stdlib/`),
or code generators (`backend/*`, `internal/fidget`), downstream test golden files (`.wgsl`, `.glsl`,
`.js`, `.assembled`) may change.
Follow the [golden-verification skill](../golden-verification/SKILL.md) to categorize all golden
diffs and prove mathematical and functional equivalence before committing golden updates.

## 5. Strict Compiler Pass Separation of Concerns

Compiler passes and AST transformations across `assembler` and `internal/*` MUST maintain strict
single-responsibility boundaries:
- **Constant Evaluation (`internal/eval`)**: Evaluates constant subexpressions, parameter defaults,
  and compile-time arithmetic.
- **Structural Lowering (`internal/lower`)**: Lowers high-level constructs (`map`, `reduce`) into
  canonical AST representations. Lowering passes MUST NOT import or invoke `eval`.
- **Assembler Pipeline Orchestration (`assembler`)**: Solely responsible for pass sequencing (such
  as running constant evaluation ahead of structural lowering).
- **Pass Independence**: No compiler pass should cross-invoke or conflate the responsibilities of
  another pass.

## 6. Package & Toolchain Map

- `parser/` — Stateless GSL lexer/parser (`ParsePackage`) and LRU `CachingParser`.
- `assembler/` — Multi-package linker, type checker, constant folder, inliner, and optimizer.
- `eval/` — Spatial SDF/interval evaluator, bounds, extrema, and formal certification.
- `program/` — Canonical protobuf AST (`program.proto`), diagnostics, errors, and preambles.
- `stdlib/` — Compile-time embedded GSL standard library packages.
- `backend/{wgsl,mesh,audio,tape}/` — Target emitters for WebGPU shaders, STL meshes, WebAudio JS,
  and bytecode tapes.
- `cmd/{gsl,compile,serve,fix}/` — Unified CLI toolchain built into `bin/` via `./tool/build-cli.sh`.

