# Embed Standard Library via `embed.FS` and Pre-Parsed `Package` Protobufs

## 1. Overview & Problem Statement

Currently, `omega-core` loads and initializes the 50 GSL standard library packages (~1.44 MB of
`.gsl` source across `stdlib/`) with two drawbacks:

1. **Manual Per-File `//go:embed` Boilerplate (`stdlib/pkg_info.go`):**
   `stdlib/pkg_info.go` declares more than 50 individual `//go:embed <path>.gsl` string variables
   and wires them by hand into `All()`.
2. **Expensive Eager Source Parsing in `parser.init()` (`parser/parser.go`):**
   Every binary or `go test` invocation that imports `gradientcontrol.com/omega/core/parser`
   executes `ParsePackage` across all 50 standard library packages during `init()`, adding ~1.2
   seconds of startup latency, duplicating standard library package storage across `program` and
   `parser`, and coupling `program.GetStdPackage` / `program.GetImport` to `parser.init()` side
   effects via `program.RegisterStdPackage`.

This track replaces the per-file `//go:embed` string declarations in `stdlib` with `embed.FS`,
enforces strict package file conventions (`package.yaml` or `<dirname>.gsl`), embeds pre-parsed
source-hash-prefixed `program.Package` protobuf binaries (`<pkg>/package.pb`), shares a single
lazily unmarshaled AST cache between `program` and `CachingParser`, removes
`program.RegisterStdPackage`, and introduces **zero new public API** across `omega-core` by placing
all binary `.pb`, source-hashing, and hash-lookup machinery inside `internal/stdpkg`.

---

## 2. Public API Surface Policy (Zero New Public API)

None of the new `.pb` embedding, source hashing, or hash-indexed lookup functionality needs to be
exported in `omega-core`'s public API:

- **`stdlib` (`gradientcontrol.com/omega/core/stdlib`):** Public API remains **100% unchanged**
  (`PkgInfo{ImportPath, Sources, Domain}`, `All()`, `PathForPrefix()`, `Prefixes()`). No `Proto`
  field, `SourcesHash`, `Lookup`, or `LookupByHash` functions are added to `stdlib`.
- **`program` (`gradientcontrol.com/omega/core/program`):** Public API **shrinks** by deleting
  `RegisterStdPackage`. Existing `GetStdPackage` and `(*Program).GetImport` keep their current
  signatures.
- **`parser` (`gradientcontrol.com/omega/core/parser`):** Public API remains **100% unchanged**
  (`Result`, `ParsePackage`, `CachingParser`, `NewCachingParser`).
- **`internal/stdpkg` (`gradientcontrol.com/omega/core/internal/stdpkg`):** All shared `.gsl` /
  `package.yaml` `fs.FS` loading, 32-byte SHA-256 source hashing, `<pkg>/package.pb` validation,
  and `ImportPath` / hash-to-`ImportPath` indexing live in `internal/stdpkg`, accessible to
  `stdlib`, `program`, and `parser` within `omega-core` while remaining completely unexported to
  external module consumers.

---

## 3. Functional Requirements

### 3.1 `embed.FS` Loading in `stdlib` & Shared Internal Registry in `internal/stdpkg`

1. **Single `embed.FS` Directive in `stdlib/pkg_info.go`:**
   - Replace the 50+ `//go:embed ...` string variables in `stdlib/pkg_info.go` with a single
     `embed.FS` embedding the standard library directories (`builtin`, `color`, `ease`, `lattice`,
     `map`, `material`, `matrix`, `mechanical`, `noise`, `text`, `texture`, `transform`, `vector`).
   - Go's `embed.FS` directory embedding recursively includes nested subpackages while excluding
     `_`-prefixed helper directories such as `text/_tools/`.
   - In `stdlib`'s `init()`, pass `stdlibFS` and the curated `(ImportPath, Domain)` descriptor
     slice to `stdpkg.RegisterFS(stdlibFS, entries)` (a constant-time pointer assignment performing
     zero file reads or allocations at startup).
2. **Curated Package Order & Strict Source File Convention:**
   - Preserve the curated package list (`ImportPath` and `Domain`) in `stdlib/pkg_info.go` so
     `All()` maintains deterministic documentation order.
   - Enforce the same strict package layout rule as `internal/source/local.go` (`loadLocalSources`)
     inside `stdpkg.LoadSources(fsys, importPath)`: every package directory `<ImportPath>` MUST
     contain **either**:
     - `<ImportPath>/package.yaml` specifying the exact ordered `.gsl` file list under `sources:`
       (such as `builtin/package.yaml`), **or**
     - a single `.gsl` file named `<path.Base(ImportPath)>.gsl` (for example, `color/color.gsl` or
       `lattice/beam/beam.gsl`), with no other `*.gsl` files in that directory.
   - Update `stdlib/pkg_info_test.go` to assert this requirement across all standard library
     package directories on disk (verifying that when `package.yaml` is present, every listed
     source exists and no unlisted `*.gsl` files exist in the directory; otherwise, the directory
     contains exactly one `.gsl` file named `<dirname>.gsl`).
3. **Internal Lookup & Protobuf Access in `internal/stdpkg`:**
   - Keep `internal/stdpkg` free of imports of `program` or `parser` so `stdlib`, `program`, and
     `parser` can all import `internal/stdpkg` with an acyclic dependency graph:
     ```text
     parser  --> program --> stdlib --> internal/stdpkg
       |            |                        ^
       +------------+------------------------+
     ```
   - `internal/stdpkg` lazily indexes packages once via `sync.Once` and provides internal helpers:
     - `SourcesHash(sources []string) [32]byte`: computes the canonical length-prefixed SHA-256 of
       package source strings.
     - `Proto(importPath string) ([]byte, bool)`: returns `<ImportPath>/package.pb`'s marshaled
       `Package` payload (`raw[32:]`) if and only if `len(raw) >= 32` and
       `[32]byte(raw[:32]) == SourcesHash(sources)`.
     - `RawProto(importPath string) ([]byte, bool)`: returns the raw `<ImportPath>/package.pb`
       bytes (including the 32-byte header) for freshness verification in `parser` tests.
     - `ImportPathByHash(hash [32]byte) (string, bool)`: O(1) lookup from `SourcesHash(sources)` to
       `ImportPath` for packages with a valid, hash-matching `package.pb`.

### 3.2 Binary Format of `<pkg>/package.pb` (32-Byte Source-Hash Prefix)

Each `<ImportPath>/package.pb` file uses a minimal, zero-framing header:

```text
Bytes 0..31  : SHA-256 digest of Sources (8-byte little-endian length prefix per source file)
Bytes 32..N  : Deterministic proto.MarshalOptions{Deterministic: true} bytes of *program.Package
```

- **Stale-Safe Verification & Zero Clone Overhead:**
  - A `package.pb` file is valid and current if and only if `len(raw) >= 32` and
    `[32]byte(raw[:32]) == SourcesHash(sources)`.
  - `raw[:32]` is sliced off before `proto.Unmarshal(raw[32:], pkg)`, so the unmarshaled
    `*program.Package` and subsequent `proto.Clone` calls incur zero extra bytes or overhead.

### 3.3 Shared Lazy Stdlib Cache Across `program` and `CachingParser` & Removal of `RegisterStdPackage`

1. **Remove `program.RegisterStdPackage` Completely:**
   - Delete `RegisterStdPackage` and `var stdPkgs = make(map[string]*Package)` from
     `program/get_import.go`. Neither `omega` nor any production caller outside `parser.init()`
     uses `RegisterStdPackage`:
     - Remove the obsolete dummy `builtin` `init()` registration from
       `internal/types/types_test.go`.
     - Update `program/get_import_test.go` to test real embedded standard library resolution (e.g.,
       `"vector"` and `"builtin"`) directly.
2. **Single Lazy Unmarshaled Cache in `program` (`program/get_import.go`):**
   - Unmarshal each standard library package at most **once per process** on first demand (using an
     unexported per-package `sync.OnceValue` cache in `program/get_import.go` backed by
     `stdpkg.Proto(path)`), storing the immutable canonical `*program.Package`:
     - `(*Program).GetImport(path)` returns that shared read-only `*program.Package` directly.
     - `program.GetStdPackage(path)` returns `proto.Clone(pkg).(*program.Package)` so callers may
       mutate the returned copy safely.
3. **Share `program.GetStdPackage` with `parser.CachingParser` & Delete `parser.init()`:**
   - Delete `init()` from `parser/parser.go` and remove `stdCachedPkg`, `stdCachedPkgs`,
     `sourcesKey`, and the 50-package pre-warm loop from `parser/caching.go`.
   - In `(*CachingParser).ParsePackage(sources []string)`:
     1. Compute `key := stdpkg.SourcesHash(sources)`.
     2. Check `if importPath, ok := stdpkg.ImportPathByHash(key); ok` — if matched, return
        `&Result{Package: program.GetStdPackage(importPath)}, nil` directly without acquiring
        `c.mu`, without consuming `CachingParser`'s LRU byte limit, and without duplicating
        unmarshaled `*program.Package` structs between `program` and `parser`.
     3. Otherwise, proceed through `c.get(key)` -> `ParsePackage(sources)` -> `c.put(...)` as usual.
   - **Live Edit Fallback:** If a developer edits a `.gsl` file in `stdlib/` and runs `go test`
     before `package.pb` is regenerated, `stdpkg.ImportPathByHash(key)` misses because `SourcesHash`
     no longer matches `raw[:32]`, so `CachingParser.ParsePackage` automatically falls through to
     parsing the edited `.gsl` source live and caching it in the LRU cache.

### 3.4 Single-Source-of-Truth Generator (`stdlib/internal/gen`) & `tool/build-stdlib.sh`

1. **Single Source of Truth in `stdlib/internal/gen` (`stdlib/generate.go` & `stdlib/internal/gen/main.go`):**
   - Add `//go:generate go run ./internal/gen` in `stdlib/generate.go`.
   - `stdlib/internal/gen` is the **sole generator** of `<ImportPath>/package.pb` files and accepts
     an optional `--force` flag:
     - **Default (Incremental SHA-256 Skip):** Before parsing `lib.Sources`, read existing
       `<ImportPath>/package.pb` from disk. If `len(existing) >= 32` and
       `[32]byte(existing[:32]) == stdpkg.SourcesHash(lib.Sources)`, skip both
       `parser.ParsePackage` and writing. Only packages with missing or hash-mismatched
       `package.pb` files are parsed and written atomically.
     - **`--force` (Full Rebuild):** Bypass the SHA-256 skip, parse all 50 packages in
       `gsllibs.All()`, marshal with `proto.MarshalOptions{Deterministic: true}`, prepend
       `stdpkg.SourcesHash(lib.Sources)`, and write every `<ImportPath>/package.pb` (ensuring
       parser or `program.proto` changes regenerate all `.pb` files even when `.gsl` sources did
       not change).
2. **Fast `mtime`-Guarded `./tool/build-stdlib.sh` Delegating to `go run ./stdlib/internal/gen`:**
   - Create `tool/build-stdlib.sh` supporting `--force` and delegating all `.pb` generation to
     `go run ./stdlib/internal/gen`:
     - If `--force` is passed, or if any file in `parser/` (`*.go`), `program/program.pb.go`, or
       `stdlib/internal/gen/` (`*.go`) is newer than the oldest `<ImportPath>/package.pb`, invoke
       `go run ./stdlib/internal/gen --force`.
     - Else if any `<ImportPath>/package.pb` is missing or any file in `stdlib/` (`*.gsl`,
       `package.yaml`, `pkg_info.go`) is newer than its `<ImportPath>/package.pb`, invoke
       `go run ./stdlib/internal/gen` (incremental mode, parsing only the modified stdlib
       packages).
     - Otherwise, exit immediately (~10 ms shell check without invoking `go`).
   - Hook `./tool/build-stdlib.sh` into `./tool/build.sh`, `./tool/build-cli.sh`,
     `./tool/build-wheel.sh`, `./tool/test-go.sh`, `./tool/check.sh`, `./tool/update-goldens.sh`,
     and `./tool/build-proto.sh` (with `--force` after `protoc` regenerates `program.pb.go`).
3. **Freshness Unit Test (`parser`):**
   - Add `TestStdlibProtosUpToDate` in `parser` that parses every package in `gsllibs.All()` from
     `.gsl` sources and verifies via `stdpkg.RawProto(lib.ImportPath)` that:
     1. `len(raw) >= 32` and `[32]byte(raw[:32]) == stdpkg.SourcesHash(lib.Sources)`.
     2. Unmarshaling `raw[32:]` yields a `*program.Package` that is `proto.Equal` to the freshly
        parsed `res.Package`.
   - On mismatch, fail with an explicit message instructing the caller to run
     `./tool/build-stdlib.sh --force`.

### 3.5 Code Documentation Requirements (Explaining *What* and *Why*)

Clear, concise comments explaining **what** each mechanism does and **why** it is structured that
way MUST be included at every non-obvious boundary:

1. **`stdlib/pkg_info.go`:**
   - Explain why `//go:embed` lists the top-level library directories (recursively embedding nested
     packages while automatically excluding `_`-prefixed tool directories like `text/_tools/`).
   - Explain why `All()` keeps a hand-curated package order (stable generated documentation) and why
     `init()` only registers `stdlibFS` with `internal/stdpkg` without reading files eagerly.
2. **`internal/stdpkg`:**
   - Document the 32-byte SHA-256 header + deterministic protobuf payload layout of `package.pb`,
     why source strings are length-prefixed before hashing (preventing boundary collisions across
     multi-file packages), and why `Proto` / `ImportPathByHash` validate the 32-byte header against
     current `.gsl` sources (preventing stale ASTs during local `.gsl` edits before regeneration).
   - Document why `internal/stdpkg` avoids importing `program` or `parser` (preserving the acyclic
     `parser -> program -> stdlib -> internal/stdpkg` package graph while keeping all `.pb`
     plumbing out of public APIs).
3. **`program/get_import.go`:**
   - Explain why standard library packages are unmarshaled lazily once per package (`sync.Once`),
     why `(*Program).GetImport` returns the shared read-only `*Package` while `GetStdPackage`
     returns a `proto.Clone` copy (callers mutate returned `Package` ASTs in place), and why this
     single cache is shared with `parser.CachingParser`.
4. **`parser/caching.go`:**
   - Explain why `(*CachingParser).ParsePackage` checks `stdpkg.ImportPathByHash` ->
     `program.GetStdPackage` before acquiring `c.mu` (lock-free stdlib hits, zero consumption of
     the 32 MiB LRU budget, and automatic fallthrough to live parsing if `.gsl` sources were edited
     locally).
5. **`stdlib/internal/gen/main.go` & `tool/build-stdlib.sh`:**
   - Explain why `stdlib/internal/gen` is the single source of truth for `package.pb` generation,
     why default mode skips unchanged packages via the 32-byte SHA-256 header while `--force`
     rebuilds all packages (covering `parser/` or `program.proto` changes where `.gsl` text did not
     change), and why `tool/build-stdlib.sh` uses `mtime` guards to keep no-op builds instantaneous.

---

## 4. Non-Functional Requirements

- **Zero New Public API:** No new exported types, struct fields, functions, or methods in `stdlib`,
  `program`, or `parser` (and `program.RegisterStdPackage` is removed).
- **Startup Latency:** Importing `gradientcontrol.com/omega/core/parser` or
  `gradientcontrol.com/omega/core/program` must perform zero `.gsl` source parsing or file reads
  during `init()`, reducing package initialization overhead from ~1.2 s to ~0 ms.
- **Fast Incremental Regeneration:** Editing a single `.gsl` file in `stdlib/` only re-parses that
  single package in `stdlib/internal/gen`, while no-op runs of `./tool/build-stdlib.sh` finish in
  < 50 ms.
- **Determinism:** Regenerating `stdlib/**/package.pb` twice from identical inputs must produce
  byte-for-byte identical files with zero `git diff`.

---

## 5. Acceptance Criteria

- [ ] Zero new exported symbols are added to `stdlib`, `program`, or `parser`, and
  `program.RegisterStdPackage` is deleted.
- [ ] All 50+ per-file `//go:embed` string variables are removed from `stdlib/pkg_info.go` and
  replaced by `embed.FS` loading enforcing either `<ImportPath>/package.yaml` or a single
  `<dirname>.gsl` file.
- [ ] `stdlib/pkg_info_test.go` asserts that every standard library directory on disk has either a
  valid `package.yaml` or a single `<dirname>.gsl` file matching its directory name.
- [ ] Every package directory in `stdlib/` includes a deterministic, source-hash-prefixed
  `package.pb` managed internally via `internal/stdpkg`.
- [ ] `parser.init()` is removed; `program` lazily unmarshals each standard library package at most
  once and shares it with both `(*Program).GetImport` and `(*parser.CachingParser).ParsePackage`.
- [ ] `stdlib/internal/gen` skips unchanged packages via the 32-byte SHA-256 header by default and
  rebuilds all packages when `--force` is passed; `./tool/build-stdlib.sh` delegates exclusively to
  `go run ./stdlib/internal/gen` with fast `mtime` guards and is wired into
  `./tool/{build,build-cli,build-wheel,test-go,check,update-goldens,build-proto}.sh`.
- [ ] Code comments explaining *what* and *why* are present across `stdlib/pkg_info.go`,
  `internal/stdpkg`, `program/get_import.go`, `parser/caching.go`, `stdlib/internal/gen`, and
  `tool/build-stdlib.sh`.
- [ ] `./tool/check.sh` and `./tool/test.sh` pass cleanly in `omega-core`.

---

## 6. Out of Scope

- Changing the public signatures of `parser.ParsePackage`, `(*parser.CachingParser).ParsePackage`,
  `program.GetStdPackage`, or `gsllibs.All`.
- Modifying GSL standard library `.gsl` definitions or compiler pass semantics.
