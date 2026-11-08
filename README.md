# Omega Core SDK (`omega-core`)

`omega-core` is the standalone Go SDK for the Geometric Specification Language (GSL). It provides
pure, modular compiler libraries and the `gsl` command-line toolchain (`compile`, `serve`, `fix`)
for parsing, analyzing, refactoring, and compiling GSL models into WebGPU compute shaders (WGSL),
3D meshes (STL), bytecode tapes, and audio synthesizers.

## Architecture & Package Topology

- **`parser/`**: Stateless GSL lexer, parser, and AST generator.
- **`assembler/`**: Multi-package linker, type checker, inliner, constant evaluator, and optimizer.
- **`program/`**: Canonical AST definitions, type systems, and visitor models.
- **`backend/`**:
  - `wgsl/`: WGSL compute shader emitter for real-time GPU SDF raymarching.
  - `mesh/`: Unified 3D mesh generator coordinating dual-contouring (Fidget) and CSG (Manifold).
  - `audio/`: Audio synthesizer compiler emitting WebAudio / JavaScript synthesis functions.
  - `tape/`: Bytecode tape compiler for compact embedded geometry evaluation.
- **`stdlib/`**: Compile-time embedded GSL standard library packages (`builtin`, `color`, `vector`, `matrix`, etc.).
- **`cmd/gsl/`**: Unified GSL toolchain CLI driver (`gsl <command> [args]`).
- **`cmd/compile/`**: Standalone multi-backend command-line compiler (`gsl compile`).
- **`cmd/serve/`**: Local HTTP server and interactive WebGPU viewer host (`gsl serve`).
- **`cmd/fix/`**: Source refactoring tool for migrating legacy GSL syntax (`gsl fix`).

### Directory Layout

```text
omega-core/
├── go.mod                     # module gradientcontrol.com/omega/core
├── assembler/
│   ├── assemble.go            # Assemble(primarySources, opts...) API
│   ├── options.go             # Options (SourcePath, SourceManager, WithProgress...)
│   └── source/                # Source abstraction
│       └── manager.go         # source.Manager interface definition
├── backend/                   # Target code generators
│   ├── audio/                 # Web Audio synthesis generator
│   │   ├── emit.go            # AudioToJs / ExtractAudio API
│   │   └── internal/          # Implementation copied from agent/js
│   │       ├── codegen/       # Copied from agent/js/internal/codegen
│   │       ├── extract/       # Copied from agent/js/internal/extract
│   │       └── lib/           # Copied from agent/js/lib
│   ├── internal/              # Shared backend internal implementations
│   │   ├── common/            # Shared backend generator base
│   │   │   └── build.go       # Shared builder and visitor interfaces
│   │   ├── environment/       # Lighting environment presets for 3D rendering
│   │   ├── fidget/            # Copied from agent/fidget (dual-contouring SDF mesher)
│   │   │   ├── tape.go        # Tape generation helper
│   │   │   ├── internal/
│   │   │   └── lib/
│   │   ├── manifold/          # Copied from agent/manifold/internal/manifold (CSG)
│   │   ├── meshmeasure/       # Mesh measurement utilities
│   │   └── tape/              # Copied from tape generator
│   ├── mesh/                  # Unified 3D mesh generator
│   │   └── emit.go            # Mesh(prog, opts...) & MeshBodies API
│   ├── tape/                  # Bytecode tape compiler
│   │   └── emit.go            # Tape compiler API
│   └── wgsl/                  # WebGPU shader emitter
│       ├── emit.go            # WGSL emission API
│       └── internal/          # Implementation copied from agent/wgsl
│           ├── accel/         # Hardware accelerated WGSL kernel specs
│           └── lib/           # Copied from agent/wgsl/internal/codegen etc.
├── bin/                       # Local build outputs (gsl, compile, serve, fix binaries; gitignored)
├── cmd/                       # Standalone CLI tools
│   ├── compile/               # Unified compiler CLI (emits WGSL, STL, tape, audio)
│   │   └── main.go            # Entry point and CLI command execution
│   ├── fix/                   # GSL source refactoring CLI (rewrites legacy comments to directives)
│   │   └── main.go            # Entry point and source refactoring CLI execution
│   ├── gsl/                   # Unified GSL toolchain CLI driver (delegates to bin/ tools)
│   │   └── main.go            # Entry point and subcommand delegation
│   └── serve/                 # Local HTTP server and WebGPU viewer host
│       └── main.go            # Entry point and HTTP server startup
├── examples/                  # Curated showcase GSL models and assets
│   ├── gears.gsl              # Parametric helical gear pair showcase (from common/web/example/Gears.gsl)
│   ├── glass.gsl              # Transmissive glass and refraction showcase (from common/web/example/Material - Glass.gsl)
│   ├── assets/                # Traditional CAD mesh assets (bracket.stl)
│   └── viewer/                # Interactive WebGPU viewer web application
│       ├── README.md          # Viewer usage and architecture overview
│       ├── index.html         # Single-page HTML shell and import map
│       ├── main.js            # Top-level UI controller and /api/* client
│       ├── style.css          # Layout styles for top bar, editor, and panels
│       ├── docs/              # Built-in GSL language reference and interactive help panel
│       ├── editor/            # CodeMirror 6 GSL/WGSL editor, autocomplete, and search
│       ├── fonts/             # Self-hosted UI (PublicSans) and code (IBMPlexMono) fonts
│       ├── icons/             # SVG toolbar and inspector icons
│       ├── materials/         # PBR material catalog, BRDF lookup tables, and swatches
│       ├── panels/            # Inspector panel controllers, mesh/CSG pipelines, and styles
│       ├── render/            # 3D viewport, WebGPU raymarching, shading, and gizmo modules
│       │   ├── viewer.js      # Three.js/WebGPU raymarching viewport and camera rig
│       │   └── environment/   # Lighting environment presets JSON for 3D rendering
│       ├── ui/                # Generic browser UI widgets (events, toasts, progress, splitters)
│       │   └── error.js       # Diagnostics banner and compiler/shader error controller
│       └── vendor/            # Vendored browser libraries (Three.js, CodeMirror, Manifold)
├── internal/                  # Shared compiler passes and internal AST transformations
│   ├── assembler/             # Copied from agent/assembler/lib (orchestrates passes)
│   │   ├── annotation/        # AST annotation metadata extraction
│   │   ├── constspec/         # Constant specialization
│   │   ├── deadcode/          # Dead code elimination
│   │   ├── inline/            # Function inlining
│   │   ├── meshplan/          # Mesh partition planning (from common/program/meshplan)
│   │   ├── property/          # Property assignments and evaluation
│   │   ├── propval/           # Property value evaluation
│   │   ├── textures/          # Texture descriptor binding
│   │   ├── topofold/          # Topological folding
│   │   ├── unroll/            # Loop unrolling transformations
│   │   ├── variation/         # Sub-program and variation slicing (from common/program/variation)
│   │   └── warpfold/          # Warp fold transformations
│   ├── build/                 # Shared AST visitor and builder copied from common/build
│   ├── closure/               # Lambda closure analysis and capture binding
│   ├── collect/               # Multi-package dependency gathering
│   ├── edit/                  # Source span edit primitives
│   ├── eval/                  # Constant evaluation & expression folding
│   ├── float16/               # IEEE 754 binary16 half-precision float encoding
│   ├── link/                  # Lexical scoping and reference linking
│   ├── lower/                 # Structural AST transformations
│   ├── meshasset/             # Imported triangle mesh and SDF grid assets
│   ├── meshfold/              # Asset slot binding
│   ├── meshslots/             # Mesh slot extraction
│   ├── noise/                 # Procedural noise generators
│   ├── pkg/                   # Package naming and path resolution
│   ├── preamble/              # Builtin preamble generation
│   ├── refs/                  # AST reference resolution (from common/program/refs)
│   ├── scope/                 # Lexical scope trees and symbol tables
│   ├── slider/                # Slider parameter metadata (from common/source/edit/slider)
│   ├── source/                # Concrete source managers (StdlibManager, LocalManager)
│   │   └── manager.go         # LocalManager implementation
│   ├── srctest/               # Shared test runner for GSL source tests
│   ├── test/
│   │   ├── assert/            # Test assertions (from stirigma test/assert)
│   │   └── require/           # Test require assertions (from stirigma test/require)
│   ├── testdata/              # Assembler test fixtures
│   │   └── example/           # Example GSL files for tests
│   ├── types/                 # Semantic type analysis
│   └── ui/                    # UI control metadata extraction (from common/source/edit/ui)
├── parser/                    # Lexer and Parser
│   ├── parser.go              # ParsePackage entry point
│   └── internal/              # Implementation copied from Omega/omega
│       ├── annotation/        # Parser AST annotation metadata extraction
│       ├── api/               # Parser API types
│       ├── identifier2/       # Copied from agent/parser/internal/identifier2
│       ├── lib/               # Copied from agent/parser/lib
│       └── parser2/           # Copied from agent/parser/internal/parser2 (name preserved)
├── program/                   # Core Program/AST models and definitions (from common/program)
│   ├── program.proto          # Protobuf AST schema definitions
│   └── scene/                 # Scene descriptor construction and extent utilities (from common/program/scene)
├── stdlib/                    # GSL Standard Library (mirrors common/source/package)
│   ├── pkg_info.go            # Package registry and compile-time embedded source data
│   ├── builtin/
│   ├── color/
│   ├── ease/
│   ├── glass/
│   ├── lattice/
│   ├── map/
│   ├── material/
│   ├── matrix/
│   ├── mechanical/
│   ├── metal/
│   ├── noise/
│   ├── physical/
│   ├── text/
│   ├── texture/
│   ├── transform/
│   └── vector/
└── tool/                      # Developer build and automation scripts
    ├── build-cli.sh          # Builds compiler CLI executables (gsl, compile, serve, fix) into bin/
    ├── build-fidget.sh        # Builds Fidget and Manifold Rust/C++ native and WASM artifacts
    ├── build-proto.sh         # Compiles program.proto and tape.proto into Go bindings
    ├── build-wgsl-accel-ops.sh # Regenerates WGSL hardware acceleration ops table and inliner registry
    ├── check.sh               # Pre-commit quality gate (gofmt, go vet, tests)
    ├── clean.sh               # Cleans bin/ outputs, test cache, and scratch files
    ├── test.sh                # Runs unit, integration, and golden test suites
    ├── update-goldens.sh      # Regenerates reference test goldens
    └── run-go-tests/          # Go test suite runner, timing history, and golden size reporter (from tool/run-go-tests)
```

## Building the Toolchain

Compile the `gsl`, `compile`, `serve`, and `fix` binaries into `bin/`:

```bash
./tool/build-cli.sh
```

Run pre-commit quality checks (gofmt, go vet, and tests):

```bash
./tool/check.sh
```

Run unit and integration test suites:

```bash
./tool/test.sh
```

Clean build outputs and test cache:

```bash
./tool/clean.sh
```

## Toolchain Driver (`bin/gsl`) Usage

```text
gsl <command> [arguments]
```

Subcommands delegate to sibling tool binaries in `bin/`:
- `gsl compile` (aliases `build`, `c`): Compile GSL sources to target outputs (`bin/compile`).
- `gsl serve`: Launch the local HTTP server and interactive WebGPU viewer (`bin/serve`).
- `gsl fix`: Refactor GSL source files to current syntax (`bin/fix`).
- `gsl help [command]`: Display help for `gsl` or a specific subcommand.

### Supported Compiler Targets (`-t, --target`)

- `wgsl`: WebGPU signed distance field shader code (default).
- `stl`: 3D printable watertight binary polygon mesh via dual-contouring.
- `tape`: Compact vectorized spatial bytecode evaluation tape.
- `audio`: Web Audio JavaScript oscillator and synthesis code.
- `ast`: Formatted JSON AST dump for structural inspection.

### Companion Metadata (`<output>.json`)

When compiling to output artifacts (such as `.wgsl` or `.stl`), `gsl compile` automatically emits a
companion `<output>.json` metadata file alongside the output. This document includes:
- Extracted `@slider` definitions (names, min, max, initial values, and precision).
- Extracted UI controls and layout descriptors.
- Scene bounding box dimensions and camera presets.
- Non-fatal compiler diagnostics and warning spans.

To specify a custom metadata destination, pass `-m <file>`. To disable companion metadata emission,
pass `--no-meta`.

### Common Workflow Examples

Compile a model to a WebGPU shader with auto-emitted `model.json` metadata:

```bash
bin/gsl compile -t wgsl -o model.wgsl model.gsl
```

Compile a standalone WebGPU shader bundle with full-screen raymarcher harness (`vs_main`, `fs_main`):

```bash
bin/gsl compile --standalone -o preview.wgsl model.gsl
```

Compile a model into a watertight binary STL mesh:

```bash
bin/gsl compile -t stl --subdivision 5 --algo marching-cubes -o model.stl model.gsl
```

Compile from standard input (`stdin`) and stream to standard output (`stdout`):

```bash
cat model.gsl | bin/gsl compile -t wgsl -o -
```

Refactor `.gsl` files or directories in-place (or preview with `-diff`):

```bash
bin/gsl fix -diff examples
bin/gsl fix examples
```
