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
├── backend/                   # Target code generators
│   ├── audio/                 # Web Audio synthesis generator
│   ├── mesh/                  # Unified 3D mesh generator
│   ├── tape/                  # Bytecode tape compiler
│   └── wgsl/                  # WebGPU shader emitter
├── bin/                       # Local build outputs (gsl, compile, serve, fix binaries; gitignored)
├── cmd/                       # Standalone CLI tools
│   ├── compile/               # Unified compiler CLI (emits WGSL, STL, 3MF, tape, audio)
│   ├── fix/                   # GSL source refactoring CLI (rewrites legacy comments to directives)
│   ├── gsl/                   # Unified GSL toolchain CLI driver (delegates to bin/ tools)
│   └── serve/                 # Local HTTP server and WebGPU viewer host
├── examples/                  # Curated showcase GSL models and assets
│   ├── assets/                # Traditional CAD mesh assets (bracket.stl)
│   └── viewer/                # Interactive WebGPU example viewer web application
│       ├── docs/              # Built-in GSL language reference and interactive help panel
│       ├── editor/            # CodeMirror 6 GSL/WGSL editor, autocomplete, and search
│       ├── fonts/             # Self-hosted UI (PublicSans) and code (IBMPlexMono) fonts
│       ├── icons/             # SVG toolbar and inspector icons
│       ├── materials/         # PBR material catalog, BRDF lookup tables, and swatches
│       └── vendor/            # Vendored browser libraries (Three.js, CodeMirror, Manifold)
├── parser/                    # Lexer and Parser
├── program/                   # Core Program/AST models and definitions
│   ├── program.proto          # Protobuf AST schema definitions
│   └── scene/                 # Scene descriptor construction and extent utilities
├── stdlib/                    # GSL Standard Library
│   ├── pkg_info.go            # Package registry and compile-time embedded source data
│   ├── builtin/
│   ├── color/
│   ├── ...
│   ├── transform/
│   └── vector/
└── tool/                      # Developer build and automation scripts
    ├── build-cli.sh           # Builds compiler CLI executables (gsl, compile, serve, fix) into bin/
    ├── build-fidget.sh        # Builds Fidget and Manifold Rust/C++ native and WASM artifacts
    ├── build-proto.sh         # Compiles program.proto and tape.proto into Go bindings
    ├── build-wgsl-accel-ops.sh # Regenerates WGSL hardware acceleration ops table and inliner registry
    ├── check.sh               # Pre-commit quality gate (gofmt, go vet, tests)
    ├── clean.sh               # Cleans bin/ outputs, test cache, and scratch files
    ├── test-go.sh             # Runs Go unit, integration, and golden test suites
    ├── test-rust.sh           # Runs Rust unit and doc test suites
    ├── test.sh                # Runs both Go and Rust test suites
    ├── update-goldens.sh      # Regenerates reference test goldens
    └── run-go-tests/          # Go test suite runner, timing history, and golden size reporter
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

Each top-level return is meshed as its own body, so bodies that touch are not welded together. A model
with several returns gives a zip of one STL per body (`-o model.zip`, named after the source by default),
or, with `-t 3mf`, one 3MF with each body an object in its own color:

```bash
bin/gsl compile -t 3mf -o model.3mf model.gsl
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

## Roadmap

Here is what has been changed in the [current language specification](doc/gsl_lang_spec_current.md)
and what is planned in the [future language specification](doc/gsl_lang_spec_future.md):

### Completed:
- **Functions, closures, and multiple returns**: Added first-class named and anonymous
  functions (`(params) { ... }`), lexical closures, arity-based overloading across package files,
  multi-value `return` statements, and multi-variable call destructuring (`a, b = fn();`), while
  removing legacy `...` varargs.
- **Trailing commas**: Allowed optional trailing commas in parameter lists, call arguments,
  `return` statements, multi-variable assignments, annotation arguments, and `[...]` literals.
- **Lists, matrices, indexing, and strings**: Expanded `[...]` literals and `collect(...)` to
  arbitrary compile-time sizes (including 1-element lists and lists/matrices wider than 4), with
  bracket indexing (`v[i]`, `m[c][r]`), scalar broadcasting, algebraic matrix multiplication, and
  rune-list string literals.
- **Higher-order `map` and `reduce`**: Added `map(fn, collection)` and standardized
  function-first `reduce(fn, collection [, initial])` over scalar lists and matrices.
- **Package imports and `builtin` prelude**: Added qualified (`import "pkg";`), aliased
  (`import alias "pkg";`), and dot (`import . "pkg";`) imports, optional top-level `return` for
  library files, implicit `builtin` dot-import with shadowing warnings, auto-imported standard
  package qualifiers, and dot-import ambiguity diagnostics.
- **Asset imports (`importMesh`, `importStep`, `importWarp`)**: Added slot-numbered top-level
  declarations and compiler-assigned inline import expressions for `.stl`, `.step`/`.stp`/`.hodge`,
  and `.warp`/`.castsim` assets.
- **Top-level `audio(...)` return**: Replaced `returnAudio` statements with `audio(signal)`
  (mono) and `audio(left, right)` (stereo) as the trailing top-level `return` element.
- **Automatic differentiation and shape accessors**: Added postfix `.dp` (spatial gradient),
  `.dtime` (time derivative), and `.d(var)` (partial derivative with respect to a top-level
  variable) operators, along with `.second`, `.box`, `.u`, and `.v` shape accessors.
- **Declaration properties, `auto` bounds, and cycle detection**: Added `.box` (including
  `auto` analytical bounds inference), `.color`, `.material`, `.u`, `.v`, `.simple`, `.detail`, and
  `.detailSize` properties alongside `.min`, `.max`, and `.period`, with function property
  carry-out and transitive dependency cycle detection.
- **Scene directive (`@scene`)**: Added the top-level `@scene(size [, grid])` package directive
  along with `gsl fix` migration from legacy `// Scene Size:` and `// Scene Grid:` header comments.
- **Slider `@range` annotations**: Added `@range(min, max [, period])` on variable declarations
  and inline expressions and `@range(min, max)` on function parameters, normalizing `.min`, `.max`,
  and `.period` properties into AST range metadata.
- **Viewport, parameter, and catalog annotations**: Added `@point`, `@transform`, `@spline`,
  `@color`, and `@direction([anchor])` on assignments and inline expressions, parameter control
  annotations (`@point`, `@transform`, `@color`, `@direction`, `@frame`), and catalog preset
  annotations (`@material`, `@finish`, `@surface`).

### Future:

Language changes will happen in phases:
- add new syntax or capability
- enhance `gsl fix` to upgrade existing GSL source to the newer syntax
- add warnings to the compiler output pointing out source that needs to be updated
- after some time, retire the legacy syntax

Planned
- **First-class AST annotations**: Preserve all annotations directly as structured AST metadata
  across declarations, functions, parameters, and inline expressions (including dynamic slider
  registration for inline `@range` expressions) instead of desugaring viewport annotations into
  synthetic `ui*` calls, and enforce literal-element constraints on `@point` and `@transform`.
- **Discrete option annotations (`@values`)**: Support `@values(v1, v2, ..., vn)` with two or
  more compile-time constant scalar arguments on scalar variable declarations, function parameters,
  and inline scalar literals for discrete UI selection and parameter presets.
- **Retire slider dot-properties (`.min`, `.max`, `.period`)**: Remove legacy `.min`, `.max`,
  and `.period` property assignments in favor of `@range(min, max [, period])`.
- **Retire viewport manipulator built-ins (`uiPoint`, `uiTransform`, `uiSpline`)**: Remove
  direct calls to `uiPoint`, `uiTransform`, and `uiSpline` from the language surface in favor of
  `@point`, `@transform`, and `@spline` annotations, with compiler migration diagnostics.
- **Retire legacy scene header comments**: Drop fallback support for `// Scene Size:` and
  `// Scene Grid:` comments once migration to `@scene(size [, grid])` is complete.
- **Unified asset import syntax**: Replace `importMesh`, `importStep`, and `importWarp`
  numbered-slot statements and inline import expressions with unified `import <name> "<asset>";`
  declarations.
- **Compilation cancellation API**: Accept cancellation contexts across compilation entry
  points so host applications can cleanly abort long-running builds.

## License

The code and data authored for this repository are released under the [MIT License](LICENSE).
Third-party components keep their own licenses: they are listed, with their license texts, in
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md), and the vendored directories carry their own
license files (for example `internal/fidget/vendor/`, which is MPL-2.0).
