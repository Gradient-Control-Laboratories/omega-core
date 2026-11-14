# Implementation Plan: `gsl eval` Mesh on Request and a Shader for Any Host

## Overview

This plan implements [spec.md](./spec.md) in five test-driven phases. Phases 1 and 2 need no new
dependency and are complete on their own; Phase 3 is a gate that stops for the owner's approval
before Phase 4 builds on naga.

1. **Phase 1 (`volume` takes `defs`):** a new `Evaluator.VolumeOf` takes `manifest` and `defs`
   and `Volume` calls it; the command and `Session.volume` pass `defs`.
2. **Phase 2 (`mesh` op):** gradient normals as a `WriteGLB` option, `Evaluator.Mesh` over
   `mesh.MeshBodies` with the `defs` path, per-body edge counts and volume, the `mesh` op and
   `Session.mesh`.
3. **Phase 3 (naga gate):** add `naga` (`wgsl-in`, `glsl-out`) to the Rust library, measure what
   it adds, report the crates it pulls in, record it in `tech-stack.md`, and **stop for approval**.
   Then one C entry point and its Go wrapper.
4. **Phase 4 (`shader` op):** the host preamble and the front march in WGSL, `Evaluator.Shader`
   with its two targets (`wgsl` as built, `glsl-es300` transpiled), texture files, the name map,
   the `shader` op and `Session.shader`.
5. **Phase 5 (verification and records):** full Go and Rust suites, `tool/check.sh`, usage text,
   README, CHANGELOG, third-party notices, metrics.

## Execution Metrics & Effort Tracking

| Metric             | Elapsed Time |
| :----------------- | :----------- |
| **Spec Creation**  | `pending`    |
| **Plan Creation**  | `pending`    |
| **Implementation** | `pending`    |
| **Code Review**    | `pending`    |
| **Effort**         | `pending`    |

---

## Phase 1: `volume` Takes `defs`

- [ ] Task: Calculate and record active **Spec Creation** and **Plan Creation** times
  - [ ] Compute active collaboration time from the transcript's turn deltas (10-minute idle cutoff
    per turn) and record it, with the implementation start time, in this file.
- [x] Task: Write failing tests for `volume` with `defs` (Red Phase) dc9a5a0
  - [x] `eval/volume_defs_test.go`: a ball whose radius is a slider; with `defs` that double the
    radius the bracket contains eight times the closed-form volume, and without them it does not;
    `Volume` and `VolumeOf` without defs agree.
  - [x] `cmd/compile/internal/evalcmd`: a `volume` request with `defs` answers with that part's
    bracket. The row of `TestErrors` that pinned the refusal ("no defs", exit 2) is removed, with
    the owner's authorization (2026-11-13).
  - [x] Confirm both fail before implementation.
- [x] Task: Implement (Green & Refactor Phase) dc9a5a0
  - [x] `eval/volume.go`: `VolumeOf(ctx, manifest, code, defs, expr, region, relativeTolerance)`,
    passing `manifest` and `defs` to `prepareQuery`; `Volume` calls it and its callers are
    untouched.
  - [x] `evalcmd.go`: pass `req.Defs`; drop the refusal and "(no defs)" from the usage text.
  - [x] `python/omega-core/src/omega_core/_eval.py`: `Session.volume(..., defs="")`, and the
    wheel README's two rows.
  - [ ] A Python test of `Session.volume(defs=...)`: not written. `AGENTS.md` rules out new Python
    tests, and the package's existing ones need a built wheel. Owner's call.
  - [x] `gofmt`; `go vet`; `go test ./eval -run 'Volume|Certify.*Cache'` and
    `go test ./cmd/compile/...` pass.
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

---

## Phase 2: The `mesh` Op

- [x] Task: Write failing tests for normals in `WriteGLB` (Red Phase) d1ec2bd
  - [x] `backend/mesh`: without the option, `WriteGLB` writes the bytes it writes today for one
    document (pinned before any change to the function).
  - [x] With the option, the GLB has a NORMAL accessor per node, one normal per vertex, unit
    length.
  - [x] Confirm the second fails before implementation.
- [x] Task: Implement the normals option (Green & Refactor Phase) d1ec2bd
  - [x] `GLBNormals`, an option on `WriteGLB` carrying a function from a body's welded mesh to
    its normals, so the package stays free of `eval`; `glb.Body.Normals`; `MeasureIndexed`.
  - [x] `gofmt`; `go test ./backend/mesh/...`.
- [x] Task: Write failing tests for `Evaluator.Mesh` and the op (Red Phase) 7fb7901
  - [x] A ball: one node, zero open and non-manifold edges, every normal within a stated angle of
    the radial direction, mesh volume within a stated tolerance of the closed form.
  - [x] `defs` that change the radius change the vertex extents.
  - [x] Two returns give two nodes named as the returns are.
  - [x] A vertex with a non-finite gradient gets the facet average, not a NaN.
  - [x] `evalcmd`: the response's `bodies`, `region` and `out`; a missing `out`, a negative
    `subdivision`, an unknown `algorithm`, a bad `region` or an `expr` exits 2 and writes no file.
  - [x] Confirm they fail before implementation.
- [x] Task: Implement `Evaluator.Mesh` and the op (Green & Refactor Phase) 7fb7901
  - [x] `eval/mesh.go`: assemble with `defs` as `prepareQuery` does, call `mesh.MeshBodies` with
    the request's subdivision, box, algorithm and time, evaluate the gradient at each welded
    vertex, count open and non-manifold edges and the signed volume from the index buffer.
  - [x] `eval/types.go`: `MeshResult`, `MeshBody`.
  - [x] `evalcmd.go`: the `mesh` op, its request fields (`subdivision`, `algorithm`) and usage row.
  - [x] `_eval.py`: `Document.mesh`, `Mesh` and `MeshBody`; the wheel README's two rows.
  - [x] `gofmt`; `go test ./eval/... ./backend/mesh/... ./cmd/compile/...`.
- [ ] Task: A Python test of `Document.mesh`: not written, for the reason given in Phase 1.
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)
  - [x] `go test ./eval/... ./backend/... ./cmd/...`: everything passes except
    `eval/internal/boxsweep/shard2` (`TestBoundingBoxSoundness`) and `eval/internal/libflattest`
    (`TestTextureStandInMatchesMean`), which fail with "worleyWall: noise tile not registered".
    Both fail the same way on an untouched checkout of `main` (d1b27d7) linked against the same
    native library, which was built from `release-0.1.2` and predates `main`'s `cellular.rs`.
  - [ ] Coverage report and the owner's manual verification: not done.

---

## Phase 3: The naga Gate

- [ ] Task: Measure and report, then stop
  - [ ] On a scratch commit, add `naga = { version = "30", default-features = false, features =
    ["wgsl-in", "glsl-out"] }` to `internal/fidget/internal/rust/Cargo.toml`, matching the version
    `Cargo.lock` already resolves for `fidget-wgpu`.
  - [ ] Report: crates added to `Cargo.lock` beyond those already present, each with version,
    source, licence and maintainers; the size of `compile` before and after
    (`tool/build-fidget.sh`, `tool/build-cli.sh`); the build-time change.
  - [ ] Review the added crates' sources for network access, process execution and build scripts
    before running any code that links them, as `AGENTS.md` asks of a new Go package.
  - [ ] **Stop. Wait for the owner's approval of the dependency.**
- [ ] Task: Record the dependency
  - [ ] `conductor/tech-stack.md`: naga's role (WGSL in, GLSL ES 3.00 out, for `gsl eval shader`),
    with a dated note.
  - [ ] `THIRD_PARTY_NOTICES.md` via `tool/third_party_licenses.py`, if it covers Rust crates;
    otherwise by hand in the file's existing form.
- [ ] Task: Write failing tests for the transpile entry point (Red Phase)
  - [ ] Rust (`tool/test-rust.sh`): a small WGSL module with one uniform block, one 3D texture and
    one function transpiles to text containing `sampler3D` and `texelFetch`; invalid WGSL returns
    an error, not a panic.
  - [ ] Go (`internal/fidget`): the wrapper returns the GLSL and a name map (WGSL identifier to
    GLSL identifier) for block members and samplers; an error carries naga's message.
  - [ ] Confirm they fail before implementation.
- [ ] Task: Implement the entry point (Green & Refactor Phase)
  - [ ] Rust: parse, validate, write GLSL with the `es300` profile for a named fragment entry
    point, return the text and the reflection map as JSON; catch panics at the boundary.
  - [ ] Go: `internal/fidget` wrapper in both `link_static.go` and `link_dynamic.go` builds.
  - [ ] `tool/test.sh` passes for the touched packages.
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

---

## Phase 4: The `shader` Op

- [ ] Task: Write failing tests for the march (Red Phase)
  - [ ] `backend/wgsl`: the emitted march source contains the step rule's constants (fudge 0.8,
    jump slack 4, budget 256 to 1024) taken from one Go table, so the text and the table cannot
    drift.
  - [ ] `backend/wgsl`: for a document with parts the march assigns `gsl_vanishEye` and
    `gsl_vanishView`; for one without, it does not mention them.
  - [ ] `eval`: a Go transcription of the emitted loop, driven by that table, and the CPU march
    (`eval/litrender.go`) hit the same points on a gyroid within the hit tolerance.
  - [ ] Confirm they fail before implementation.
- [ ] Task: Emit the preamble and the march (Green & Refactor Phase)
  - [ ] `backend/wgsl`: an option that appends the host preamble (`gsl_shape`, `gsl_material`,
    `gsl_min`, `gsl_isShapeVisible`, `gsl_fieldFilterIdx`, `gsl_pw`, `gsl_pwScale`, the `gsl_Host`
    block built from the sliders), `gsl_field` and `gsl_march`.
  - [ ] The front march as the viewer has it (`examples/viewer/render/gsl-render-graph.js`):
    gradient-corrected step, same-body sign-change bisection with the jump-slack test,
    pixel-footprint hit tolerance, step budget times `stepScale`. For a document that declares the
    vanish globals, set them before the first sample and clear them after the last, as the viewer
    does. No displacement or backface.
- [ ] Task: Write failing tests for `Evaluator.Shader` and the op (Red Phase)
  - [ ] For each target, the response's `code` defines `gsl_field` and `gsl_march` and has no
    entry point; the `glsl-es300` one has no `#version` line.
  - [ ] For each target, every slider `describe` reports appears in `sliders` with a `uniform`
    that occurs in `code`; `defs` replacing a slider change its `value`.
  - [ ] The `wgsl` response parses and validates through the transpile entry point's parser.
  - [ ] `corexy_printer` (parts and materials) answers for both targets.
  - [ ] A document with a noise texture lists it in `textures` and writes a file of the stated
    size under `outDir`; a mesh-SDF document does the same for its distance grid.
  - [ ] An unknown `target`, or a missing one, exits 2.
  - [ ] Confirm they fail before implementation.
- [ ] Task: Implement `Evaluator.Shader` and the op (Green & Refactor Phase)
  - [ ] `eval/shader.go`: assemble with `defs`, render WGSL with the preamble option, write
    textures. `wgsl`: return the module and the WGSL names. `glsl-es300`: add a stub fragment
    entry that keeps both functions alive through naga's dead-code pass, transpile, strip the stub
    and the version line, map names through the reflection output.
  - [ ] `eval/types.go`: `ShaderResult` and its parts.
  - [ ] `evalcmd.go`: the `shader` op, `target` and `outDir`, and its usage row.
  - [ ] `_eval.py`: `Session.shader`, with its test.
  - [ ] `gofmt`; `go test ./eval/... ./backend/wgsl/... ./cmd/compile/...`.
- [ ] Task: Check the seven documents by hand
  - [ ] Run `shader` with `glsl-es300` on `blend`, `gyroid`, `hybrid_bracket`,
    `aerospace_bracket`, `cellular_noise`, `bvhsdf_basic` and `corexy_printer`, load each result
    in a WebGL2 page outside the repository, and record compile time and frame time in this file.
    For `corexy_printer`, record the frame time with vanish on against the 36 ms measured with it
    off. This is the only browser evidence the track has.
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

---

## Phase 5: Verification and Records

- [ ] Task: Full verification
  - [ ] `./tool/test.sh` and `./tool/check.sh`; no golden under `backend/wgsl` or `backend/mesh`
    moves. A golden that moves stops the track for the owner's decision; none is regenerated.
  - [ ] Build the wheel (`tool/build-wheel.sh`) and run its tests.
- [ ] Task: Records
  - [ ] `README.md` and the wheel's README: the two ops and `volume`'s `defs`, in the form the
    existing ops are documented.
  - [ ] `CHANGELOG.md`.
  - [ ] `conductor/tracks.md`: register this track; add `index.md` and `metadata.json` beside
    this file.
- [ ] Task: Finalize Execution Metrics & Effort Tracking
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)
