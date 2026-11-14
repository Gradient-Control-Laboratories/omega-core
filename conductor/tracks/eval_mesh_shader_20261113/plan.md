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

- [x] Task: Measure and report, then stop
  - [x] `naga = { version = "30", default-features = false, features = ["wgsl-in", "glsl-out"] }`
    in `internal/fidget/internal/rust/Cargo.toml`. `Cargo.lock` gains no package (naga 30.0.1 is
    already resolved for the vendored `fidget-wgpu`), only `render`'s own dependency line.
  - [x] Newly linked into `render`: `naga` and `naga-types` 30.0.1 (MIT OR Apache-2.0, gfx-rs),
    `codespan-reporting` 0.13.1 (Apache-2.0), `bit-set` 0.10.0 and `bit-vec` 0.9.1 (Apache-2.0 OR
    MIT), `rustc-hash` 1.1.0 (Apache-2.0/MIT), `libm` 0.2.16 (MIT), and `cfg_aliases` 0.2.2 (MIT)
    at build time. `spirv` is not: only the SPIR-V features enable it.
  - [x] Size, measured on a scratch crate that links only naga behind one C entry point, at the
    locked versions (the worktree has no submodules, so `render` itself was not built): 1.87 MB
    added to a dead-stripped static link, 1.88 MB to a `cdylib`; an upper bound, since crates
    `render` already links are counted again. Not measured: `render` as a Rust `dylib`, which may
    keep more. Cold build of naga and its dependencies: 10 s wall, 26 s CPU (M4 Max).
  - [x] At naga 30.0.1 the scratch entry point transpiles `corexy_printer` (2,404 lines of WGSL
    to 5,768 of GLSL ES 3.00, about 20 ms), `cellular_noise` and `bvhsdf_basic`, with no optional
    capability enabled; invalid WGSL returns naga's message. Reflection names the uniform block
    and the samplers.
  - [x] Sources scanned for network access, process execution and filesystem writes, and the two
    build scripts (naga's, libm's) read: none found in what these features compile. A pattern
    scan, not a line-by-line audit.
  - [x] **Approved by the owner, 2026-11-13.**
- [x] Task: Record the dependency
  - [x] `conductor/tech-stack.md`: naga's role, with a dated note.
  - [x] `THIRD_PARTY_NOTICES.md` needs no change: its crate count is of `Cargo.lock`, which gains
    no package, and `codespan-reporting` is already among its Apache-2.0-only crates. The wheel's
    licence texts are collected by `tool/third_party_licenses.py` at build time.
- [x] Task: Write failing tests for the transpile entry point (Red Phase) 18da6f1
  - [x] Rust (`internal/fidget/internal/rust/src/glsl.rs`): a module with one uniform block, one
    3D texture and one function transpiles to text containing `sampler3D` and `texelFetch`; the
    names of the block and the sampler are returned and appear in the text; invalid WGSL, a module
    that does not validate and a missing entry point are errors, not panics; the C entry point
    answers in JSON, carries the message of a refusal and refuses null or non-UTF-8 arguments.
  - [x] Go (`internal/fidget/glsl_test.go`): the wrapper returns the GLSL and the name maps; an
    error carries naga's message.
  - [x] Confirmed failing before implementation (eight Rust tests against a stub; the Go tests
    did not compile).
- [x] Task: Implement the entry point (Green & Refactor Phase) 18da6f1
  - [x] Rust: `wgsl_to_glsl_es300_json(wgsl, wgsl_len, entry, entry_len) -> RenderResult`; code 0
    carries JSON (`glsl`, `uniformBlocks`, `textures`), any other code a message; panics caught.
    The maps are keyed by WGSL global: block members keep their names inside the block.
  - [x] Go: `fidget.GLSLES300(wgsl, entryPoint) (*GLSL, error)`. The dynamic link is tested; the
    `omega_static` link was not built.
  - [x] `go test ./internal/fidget/... ./eval/... ./backend/... ./cmd/...` and
    `cargo test --release -p render glsl` pass against a `librender` built in the worktree. The
    two failures recorded under Phases 1 and 2 were the stale linked library and are gone.
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

---

## Phase 4: The `shader` Op

- [x] Task: Write tests for the march 37828e2
  - [x] `backend/wgsl/host_test.go`: the emitted march takes every number of the step rule from
    one Go table (`FrontMarch`); a rule with other numbers writes them and not the viewer's.
  - [x] For a module with parts the march assigns `gsl_vanishEye` and `gsl_vanishView` and clears
    the margin before the gradient at the hit; for one without, it does not mention them.
  - [x] `internal/fidget/internal/gpuparity/march_test.go`: the emitted `gsl_march`, run on the
    GPU through `gpu-eval`, meets a ball where the ray meets it in closed form (and returns the
    unit radial as its gradient), and meets a gyroid of gradient about 3 where a 4 micron sampling
    of the field along the same ray, on the CPU, first finds it. This replaces the Go
    transcription the plan first named: it runs the emitted text itself. With the gradient
    correction removed the gyroid test fails on 12 of 144 rays; restored, it passes.
  - [x] The host declarations (`gsl_shape`, `gsl_material`, `gsl_min`) are held to the viewer's
    own text, read from `examples/viewer/render`.
  - [x] Not red first: `host.go` was written before `host_test.go`. The GPU tests and the tests of
    `Evaluator.Shader` and the op below were written and seen to fail first.
- [x] Task: Emit the preamble and the march 37828e2
  - [x] `backend/wgsl/host.go`: `Host(result, extent)` wraps a rendered module with the host
    declarations, the `gsl_Host` block (time, then sdShape's value arguments, with byte offsets),
    the textures and samplers as bindings, `gsl_field`, `gsl_gradient` and `gsl_march`.
  - [x] The front march as the viewer has it. No displacement, backface or march fade.
- [x] Task: Write failing tests for `Evaluator.Shader` and the op (Red Phase) 37828e2
  - [x] Both targets: the two functions, no entry point, no `#version`; sliders with offsets;
    bodies; the region from `@scene`, or fitted without one.
  - [x] The `wgsl` response validates through naga with an entry point added.
  - [x] `corexy_printer` answers for both targets.
  - [x] The noise tile and a mesh's distance grid are written under `outDir` at the stated size.
  - [x] An unknown or missing `target` exits 2; a texture with no `outDir` exits 1.
- [x] Task: Implement `Evaluator.Shader` and the op (Green & Refactor Phase) 37828e2
  - [x] `eval/shader.go`, with its result types beside it. `eval/eval.go`: `prepareQuery` now
    calls `assembleQuery`, which also returns the assembled source and the meshes read; its
    behaviour is unchanged.
  - [x] `evalcmd.go`: the `shader` op, `target` and `outDir`, usage text.
  - [x] `_eval.py`: `Document.shader`. No Python test was added (`AGENTS.md`); `shader`, `mesh` and
    `volume(defs=...)` were each called once through a scratch-built `gsl`, outside the repository.
  - [x] `go test ./backend/wgsl/... ./eval/... ./cmd/... ./internal/fidget/...` passes.
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
