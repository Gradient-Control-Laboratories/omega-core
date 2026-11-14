# Technology Stack: Omega Core (`omega-core`)

## Primary Languages & Runtimes

- **Go (`go 1.26`):** Core implementation language for the GSL lexer/parser, assembler passes,
  evaluator, WGSL/Audio/Tape backends, and `gsl` CLI tools (`module gradientcontrol.com/omega/core`).
- **Rust (`Cargo.toml`):** Native Fidget interval-arithmetic / Manifold Dual Contouring (MDC)
  meshing and bytecode VM evaluation static library linked via CGO (`internal/fidget`).
- **C / C++:** Native Manifold 3D CSG boolean library linked via CGO (`backend/internal/manifold`).
- **HTML5 / CSS / Vanilla JavaScript:** Standalone local WebGPU viewer (`examples/viewer/`) using
  CodeMirror 6, Three.js, WebGPU raymarching shaders, and Manifold WASM.

## Core Libraries & Data Serialization

- **Participle v2 (`github.com/alecthomas/participle/v2`):** Struct-tag grammar lexer and parser
  in `parser/internal/parser2`.
- **Protocol Buffers (`google.golang.org/protobuf`):** Canonical AST schema (`program/program.proto`)
  and VM tape bundle schema (`backend/tape/tape.proto`).
- **Go `embed` (`//go:embed`, `embed.FS`):** Compile-time embedding of GSL standard library sources
  (`stdlib/`) and pre-compiled assets.
- **naga 30 (`wgsl-in`, `glsl-out`), in the `render` Rust library:** transpiles the WGSL that
  `backend/wgsl` emits to GLSL ES 3.00, for the `glsl-es300` target of `gsl eval`'s `shader` op, so
  a WebGL2 host can draw a field. Added 2026-11-13 (track `eval_mesh_shader_20261113`), approved by
  the owner after measurement: no package added to `Cargo.lock` (the vendored `fidget-wgpu`
  already resolves naga 30.0.1), eight crates newly linked into `render` (`naga`, `naga-types`,
  `codespan-reporting`, `bit-set`, `bit-vec`, `rustc-hash` 1.1, `libm`, and `cfg_aliases` at build
  time), about 1.9 MB of code as an upper bound.

## Compiler Architecture & Pass Boundaries

- **Parser (`parser`):** Stateless `ParsePackage` and thread-safe LRU `CachingParser`.
- **Assembler (`assembler` & `internal/*`):**
  - Orchestrates dependency collection (`internal/collect`), linking (`internal/link`), closures
    (`internal/closure`), type checking (`internal/types`), constant folding (`internal/eval`),
    structural lowering (`internal/lower`), loop unrolling, inlining, and dead-code elimination.
  - Enforces strict pass separation: `internal/lower` MUST NOT import or invoke `internal/eval`.
- **Evaluator (`eval` & `eval/joints`):** Spatial point/gradient evaluation, bounding boxes,
  global extrema, and sound interval-arithmetic certification.
- **Backends (`backend/wgsl`, `backend/mesh`, `backend/tape`, `backend/audio`):** Shared AST
  builder traversal (`internal/build`) decoupled from semantic analysis.

## Build & Verification Tooling (`tool/`)

- `./tool/build-stdlib.sh`: Generates and validates pre-parsed standard library `package.pb`
  protobufs.
- `./tool/build-cli.sh`: Builds `bin/{gsl,compile,serve,fix}`.
- `./tool/build-proto.sh`: Compiles `program.proto` and `tape.proto`.
- `./tool/build-fidget.sh`: Builds native Rust/C++ Fidget and Manifold libraries.
- `./tool/test.sh` (`./tool/test-go.sh`, `./tool/test-rust.sh`): Runs Go and Rust test suites.
- `./tool/check.sh`: Runs `gofmt`, `go vet`, and test quality gates.

