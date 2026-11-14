# `gsl eval`: a Mesh on Request and a Shader for Any Host

## 1. Overview & Problem Statement

A host that edits a GSL document in a loop (an agent, a CAD viewer) wants three things from the
field after each edit: facts, a picture, and, only when a file is asked for, a mesh. `gsl eval`
gives the facts. The other two have gaps:

- **Slider values do not reach `volume` or the mesh.** Every other op takes `defs`
  (`"skin = 1.6;"`); `volume` does not, and a mesh comes only from `gsl compile`, which meshes the
  document as written. A host that moved a slider cannot ask for that part's volume or its mesh.
- **A mesh comes back without what a host needs to trust it.** `compile -t glb` writes positions
  and indices. The host then re-derives normals from facets and re-measures the mesh to learn
  whether it is closed and how its volume compares with the field's.
- **The picture needs omega's viewer.** `compile -t wgsl` emits the field's functions; the march
  that draws them lives in `examples/viewer` as JavaScript template strings. A WebGPU host has to
  copy that march, and a WebGL2 host (three.js `WebGLRenderer`) cannot use either half.

This track adds `defs` to `volume` and two ops to `gsl eval`: `mesh` and `shader`.

---

## 2. Public API Surface Policy

- **`eval.Evaluator.VolumeOf`** (new): `Volume` with `manifest` and `defs`, in the position the
  other queries have them. `Volume` keeps its signature and calls it with neither.
- **`backend/mesh`:** `WriteGLB` gains an option to write per-vertex normals. Its default output
  is unchanged, so `compile -t glb` writes the bytes it writes today.
- **New Go:** `Evaluator.Mesh` and `Evaluator.Shader` (Section 4), each a thin composition of
  existing entry points (`mesh.MeshBodies`, `wgsl.Render`) with the `defs` path of `prepareQuery`.
- **New, outside Go:** the two ops' JSON shapes, and `Session.mesh` / `Session.shader` in the
  wheel's Python package.
- **Unchanged:** `gsl compile` and every target of it; the viewer under `examples/viewer`.

---

## 3. Functional Requirements

### 3.1 `volume` takes `defs`

`{"op": "volume", "source": "part.gsl", "defs": "skin = 1.6;", ...}` brackets the volume of the
document with those statements applied, as `bounds` and `eval` already do. The usage text drops
"(no defs)". Until this track the op refused `defs` as malformed; the test row that pinned the
refusal is replaced by one that pins the answer.

### 3.2 `mesh`

```json
{"op": "mesh", "source": "part.gsl", "defs": "skin = 1.6;", "out": "part.glb",
 "subdivision": 7, "region": [x0, y0, z0, x1, y1, z1], "algorithm": "mdc", "time": 0}
```

```json
{"ok": true, "out": "part.glb", "region": [...],
 "bodies": [{"name": "bracket", "index": 0, "vertices": 5120, "triangles": 10236,
             "openEdges": 0, "nonManifoldEdges": 0, "volume": 1834.2}]}
```

- One GLB, one node per top-level return, indexed, with a normal per vertex.
- **Normals are the field's gradient at the vertex**, not an average of facets. A vertex where the
  gradient is not finite falls back to the facet average.
- `region` defaults to what `compile --bbox auto` uses; the response states the box that was used.
- `subdivision`, `algorithm` and `time` mean what the `compile` flags mean and default as they do.
- `openEdges`, `nonManifoldEdges` and the mesh's own `volume` are counted from the welded index
  buffer the op already builds, so a host need not load the file to gate an export.
- `out` is required. The op writes a file; it never returns mesh bytes in JSON.

### 3.3 `shader`

```json
{"op": "shader", "source": "part.gsl", "defs": "...", "target": "glsl-es300", "outDir": "cache/"}
{"op": "shader", "source": "part.gsl", "defs": "...", "target": "wgsl", "outDir": "cache/"}
```

```json
{"ok": true, "target": "glsl-es300", "code": "...",
 "entry": {"field": "gsl_field", "march": "gsl_march"},
 "uniformBlock": "gsl_Host",
 "sliders": [{"name": "skin", "uniform": "gsl_param_4_9", "type": "f32",
              "value": 1.6, "min": 0.8, "max": 3.0}],
 "textures": [{"uniform": "noiseTex", "kind": "noise3d", "format": "rgba8",
               "size": [64, 64, 64], "path": "cache/noise3d.bin"}],
 "bodies": [{"name": "bracket", "index": 0}],
 "region": [...]}
```

- **`code` is a library, not a program**, in the language `target` names. It declares one uniform
  block and the textures, and defines two functions. It has no entry point, no `#version` line and
  no outputs, so a host adds its own and calls in.
- **One source, two targets.** The op always builds the same WGSL module: the field as
  `compile -t wgsl` emits it, the host preamble, and the march. `wgsl` returns that module;
  `glsl-es300` returns its transpile (Section 4). A WebGPU host needs no transpiler, and both hosts
  run the same march.
- **The preamble carries every declaration the emitted field expects of its host:** `gsl_shape`,
  `gsl_material`, `gsl_min`, `gsl_isShapeVisible`, `gsl_fieldFilterIdx`, and the pixel footprint
  `gsl_pw` with its scale `gsl_pwScale`.
- **`gsl_field(p)`** returns distance and body index, with sliders and textures already bound.
- **`gsl_march(origin, direction, tEntry, tExit, pixelScale, ortho, stepScale, detail)`** returns
  whether it hit, the distance along the ray, the body index and the gradient at the hit. The
  distance is the host's to turn into depth, so a field composites with a host's own meshes. It is
  the front face of the viewer's primary march:
  - each step is divided by the tetrahedral gradient magnitude, clamped at 1, times 0.8;
  - a sign change of the same body between two samples is bisected, unless the samples are too
    far from zero for the field to have crossed (the jump-slack test);
  - the hit tolerance grows with the pixel footprint;
  - the step budget is 256 to 1024 by path length through the box, times `stepScale`;
  - for a document with parts, it sets the vanish globals (`gsl_vanishEye`, `gsl_vanishView`) from
    the ray and `detail` before the first sample and clears them after the last, so a part far from
    the ray reads as its box while marching, as in the viewer.
- **Sliders are uniforms**, so a host moves one without asking again. `sliders` maps each slider's
  GSL name to the uniform that carries it. `defs` that replace a slider change its `value`.
- **Textures are files** under `outDir`, raw texels in the stated format. The response names each
  one's sampler; nothing is inlined in JSON.
- `target` is required: `wgsl` or `glsl-es300`. In the `wgsl` response the `uniform` names are the
  WGSL names; in the `glsl-es300` response they are the names after transpiling.

### 3.4 Out of Scope (v1)

- The backface march, displacement, glass, shadows and ambient occlusion. They stay in
  `examples/viewer`. The viewer keeps its own copy of the front march too; moving the
  viewer onto the emitted one is a later track (Section 7).
- Shading. `shader` returns geometry of the hit; the look is the host's.
- Mesh bytes or texture bytes inside a JSON response.

---

## 4. How the `glsl-es300` Target Is Made

Two routes were weighed. This section records why the second was chosen.

### 4.1 A GLSL dialect of the Go backend (rejected)

`backend/wgsl` is about 10,000 lines outside tests, and WGSL syntax is not confined to a printer:

- `internal/ir` statements carry their expressions and types as WGSL strings (`LetStmt.Expr`,
  `LetStmt.Type`) and format themselves as WGSL.
- `internal/codegen` holds hand-written WGSL as Go string templates: the cellular scan, the
  spinodoid sampler, warp and noise textures, vanish, displacement, materials.
- `internal/wgsl` builds expressions as WGSL text (`vec3<f32>(...)`, `select`, reserved words).

A dialect therefore means a second copy of each template and a typed expression layer under the
IR, which is a second backend in all but name. Every later change to a template would be made
twice.

### 4.2 Transpile the emitted WGSL with naga (chosen)

The op emits WGSL as `compile -t wgsl` does, appends the host preamble (`gsl_shape`, `gsl_min`,
the uniform block) and the march, written once in WGSL, and hands the module to naga's GLSL
writer.

- **Measured on six documents** (`blend`, `gyroid`, `hybrid_bracket`, `aerospace_bracket`,
  `cellular_noise`, `bvhsdf_basic`) with naga 29's command-line tool, profile `es300`: all six
  transpile, compile under ANGLE (Chrome, Metal) in 40 to 150 ms, and draw. Texture and sampler
  pairs become one `sampler3D`; `textureLoad` becomes `texelFetch`.
- **Frame cost at 800 by 600**, the surface covering 4 to 9 percent of the picture: 0.5 to 1.2 ms
  for five of them, 27 to 32 ms for `cellular_noise`. A part that fills a high-density viewport
  marches ten to twenty times as many pixels, so `stepScale` and a host-side draft resolution are
  needed, not optional.
- **A large document, `corexy_printer`** (598 lines of GSL, parts and materials): 2,292 lines of
  WGSL become 5,766 lines of GLSL, which ANGLE compiles in 49 ms. It draws at 36 ms a frame at 800
  by 600 with the surface covering 23 percent of the picture, vanish off and a march box of twice
  the scene's bounds. That is the worst case, and it is why vanish is in the march (Section 3.3).
  The same document stopped naga twice until the harness declared `gsl_pw` and `gsl_material`,
  which is why the preamble lists them.
- **naga renames identifiers** (`distance` to `distance_`, a trailing `_` on names ending in a
  digit, its own block and instance names). The op reads the names back from naga's reflection
  output and reports them in `sliders` and `textures`; a host never guesses.

**Cost of this route.** naga is in the workspace only through the vendored `fidget-wgpu`, which
the `compile` program's Rust library does not link. This track would add `naga` (features
`wgsl-in`, `glsl-out`) to `internal/fidget/internal/rust`, one C entry point, and its Go wrapper.
It is a new native dependency of every wheel. The owner chose this route on 2026-11-13; the plan
still stops for approval once the added size and the crates it pulls in are known, and records the
change in `tech-stack.md` before any code that uses it.

---

## 5. Python (`omega_core`)

- `Session.volume(..., defs="")`, as `bounds` and `render` spell it.
- `Session.mesh(out, *, defs="", subdivision=None, region=None, algorithm=None, time=None)`
  returns the response's `bodies` and `region`.
- `Session.shader(target, *, out_dir, defs="")` returns the response as a mapping; `target` has no
  default, as in the op.

Standard library only, as the rest of the package is.

---

## 6. Tests

All in Go, written before the code they pin.

- **`volume` with `defs`:** a ball whose radius is a slider; the bracket for `defs` that double
  the radius contains eight times the closed-form volume, and the bracket without them does not.
- **`mesh`:**
  - a ball: the file parses as GLB, has one node, a NORMAL accessor, zero open and non-manifold
    edges, and every normal within a stated angle of the radial direction;
  - the mesh `volume` lies within a stated tolerance of the closed form;
  - `defs` that change the radius change the vertex extents accordingly;
  - two returns give two nodes with the returns' names;
  - a missing `out` is a malformed request (exit code 2).
- **`WriteGLB` without the option** writes the same bytes as before, pinned against the current
  output for one document.
- **`shader`:**
  - for each target, the response's `code` contains `gsl_field` and `gsl_march` and no entry point;
  - for each target, every slider `describe` reports appears in `sliders` with a `uniform` that
    occurs in `code`;
  - the `wgsl` response parses and validates (through the transpile entry point's parser);
  - a document with parts and materials (`corexy_printer`) answers for both targets, and its march
    assigns the vanish globals;
  - a document with a noise texture lists it in `textures` and writes a file of the stated size;
  - an unknown `target` is a malformed request.
- **The march's arithmetic** is pinned against the CPU renderer, which steps by the same rule
  (`eval/litrender.go`): a Go transcription of the emitted loop and the CPU march hit the same
  points on a non-unit-Lipschitz field (a gyroid) within the hit tolerance.

Not covered by an automated test: that a browser compiles the GLSL. No WebGL context exists in
the test environment; the seven-document measurement above is the evidence, and it is by hand.

---

## 7. Open Questions

1. **One march or two:** v1 leaves the viewer's copy in place. The follow-up has the viewer
   consume the emitted front march and keep only its extras.
2. **`region` for `shader` and `mesh`:** whether the certified automatic box, once it lands,
   replaces the sampled one as the default here.
3. **Analytic gradient in the march:** forward-mode `.dp` in place of four taps would cut the
   field evaluations per step from five to one dual evaluation. Not measured.

Decided (2026-11-13): the transpile lives in the `compile` program, not in the host as
WebAssembly; `shader` has two targets from the start, `wgsl` and `glsl-es300`.
