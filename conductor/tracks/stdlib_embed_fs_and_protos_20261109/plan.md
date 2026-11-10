# Implementation Plan: Embed Standard Library via `embed.FS` and Pre-Parsed `Package` Protobufs

## Overview

This plan implements [spec.md](./spec.md) across four modular, test-driven phases with **zero new
public API** in `omega-core` (and removes `program.RegisterStdPackage`):

1. **Phase 1 (`internal/stdpkg` & `stdlib` `embed.FS`):** Build the internal `stdpkg` registry
   (`SourcesHash`, strict `package.yaml` / `<dirname>.gsl` `fs.FS` loader, and 32-byte SHA-256
   header validation) and replace the 50+ per-file `//go:embed` string variables in
   `stdlib/pkg_info.go` with a single `embed.FS`.
2. **Phase 2 (`stdlib/internal/gen` & `tool/build-stdlib.sh`):** Create the single-source-of-truth
   `package.pb` generator (incremental SHA-256 skip by default, full rebuild under `--force`),
   generate all 50 `stdlib/**/package.pb` files, wire fast `mtime`-guarded `tool/build-stdlib.sh`
   into all build/test scripts, and enforce freshness via `TestStdlibProtosUpToDate` in `parser`.
3. **Phase 3 (Shared Lazy Cache in `program` & `CachingParser`):** Replace `program.RegisterStdPackage`
   and `parser.init()` with per-package `sync.Once` lazy unmarshaling in `program/get_import.go`,
   shared lock-free by `(*parser.CachingParser).ParsePackage` via `stdpkg.ImportPathByHash` ->
   `program.GetStdPackage`.
4. **Phase 4 (Full Verification, Guard Audit & Metrics):** Run full Go and Rust suites, verify zero
   golden drift, audit all modified Go files with `clean-go-test-guards`, and finalize effort
   metrics.

## Execution Metrics & Effort Tracking

| Metric             | Elapsed Time |
| :----------------- | :----------- |
| **Spec Creation**  | `~60 mins`   |
| **Plan Creation**  | `~24 mins`   |
| **Implementation** | `TBD`        |
| **Code Review**    | `TBD`        |
| **Effort**         | `TBD tokens` |

---

## Phase 1: Internal `stdpkg` Foundation & `embed.FS` Migration in `stdlib` [checkpoint: b04bdb0]

- [x] Task: Calculate and record active **Spec Creation** and **Plan Creation** times
  - [x] Compute active collaboration time from conversation transcript turn deltas (with a 10-minute
    idle cutoff per turn), display the turn breakdown table in chat, and record active durations and
    implementation start time in `plan.md`.
- [x] Task: Write failing unit tests for `internal/stdpkg` and `stdlib` package layout invariants (Red Phase) b04bdb0
  - [x] Create `internal/stdpkg/stdpkg_test.go` testing `SourcesHash` determinism and boundary
    collision resistance (`["ab", "c"]` vs. `["a", "bc"]`), `LoadSources` over in-memory `fstest.MapFS`
    (`package.yaml` source ordering, single `<dirname>.gsl` loading, rejection of unlisted or extra
    `*.gsl` files, and missing source errors), and `Proto` / `RawProto` / `ImportPathByHash`
    validation (valid 32-byte SHA-256 header, truncated `< 32` byte rejection, and stale hash
    mismatch rejection).
  - [x] Update `stdlib/pkg_info_test.go` to assert that every standard library directory on disk
    contains either a `package.yaml` file (whose `sources:` exist on disk and cover all `*.gsl`
    files in that directory) or a single `<dirname>.gsl` file matching the directory's base name.
  - [x] Confirm new tests fail before implementation.
- [x] Task: Implement `internal/stdpkg` and migrate `stdlib/pkg_info.go` to `embed.FS` (Green & Refactor Phase) b04bdb0
  - [x] Implement `internal/stdpkg` (`RegisterFS`, `LoadSources`, `SourcesHash`, `Proto`,
    `RawProto`, `ImportPathByHash`) with zero imports of `program` or `parser`, documenting *what*
    and *why* for the 32-byte SHA-256 header layout, length-prefixed source hashing, stale-hash
    rejection, and acyclic package boundary.
  - [x] Replace the 50+ `//go:embed` string variables in `stdlib/pkg_info.go` with a single
    `embed.FS` directive over the top-level library directories, register `stdlibFS` and curated
    package entries with `internal/stdpkg` in `init()` (zero file reads at `init()`), and load
    `All()` via `internal/stdpkg` while keeping `stdlib`'s public API (`PkgInfo`, `All`,
    `PathForPrefix`, `Prefixes`) 100% unchanged.
  - [x] Add *what*-and-*why* comments in `stdlib/pkg_info.go` explaining recursive directory
    embedding (`text/_tools/` exclusion), curated `All()` documentation order, and zero-I/O `init()`.
  - [x] Run `gofmt` and verify targeted tests pass:
    `go test -v ./internal/stdpkg/... ./stdlib/... ./internal/source/...`.
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`) b04bdb0

---

## Phase 2: Single-Source-of-Truth `package.pb` Generator & `tool/build-stdlib.sh` Automation

- [~] Task: Write failing tests for `stdlib/internal/gen` and `parser` `.pb` freshness check (Red Phase)
  - [ ] Create `stdlib/internal/gen/gen_test.go` testing importable generator functions on a
    temporary copy/fixture directory: initial `<ImportPath>/package.pb` creation, default
    incremental SHA-256 skip (verifying unchanged packages are neither re-parsed nor rewritten),
    `--force` full rebuild (verifying packages are rewritten even when `.gsl` hashes match), and
    byte-for-byte determinism across repeated runs.
  - [ ] Create `parser/stdlib_proto_test.go` (`TestStdlibProtosUpToDate`) asserting that every
    package in `gsllibs.All()` has a non-empty `stdpkg.RawProto(lib.ImportPath)` whose 32-byte
    header equals `stdpkg.SourcesHash(lib.Sources)` and whose unmarshaled `*program.Package` is
    `proto.Equal` to `ParsePackage(lib.Sources)` (failing initially because `<pkg>/package.pb` files
    do not exist yet).
- [ ] Task: Implement `stdlib/internal/gen`, `tool/build-stdlib.sh`, and build/test hooks (Green & Refactor Phase)
  - [ ] Create `stdlib/generate.go` (`//go:generate go run ./internal/gen`) and
    `stdlib/internal/gen/main.go` implementing single-source-of-truth `package.pb` generation with
    incremental SHA-256 skip by default, `--force` full rebuild, deterministic `proto.MarshalOptions`,
    atomic writes, and clear *what*-and-*why* comments.
  - [ ] Create executable `tool/build-stdlib.sh` with `--force` support and fast shell `mtime`
    guards (`< 50 ms` no-op when `stdlib/`, `parser/`, and `program/program.pb.go` are older than
    `package.pb`; incremental `go run ./stdlib/internal/gen` when only `stdlib/` changed;
    `go run ./stdlib/internal/gen --force` when `--force` is passed or `parser/`,
    `program/program.pb.go`, or `stdlib/internal/gen/` changed), including *what*-and-*why*
    comments.
  - [ ] Run `./tool/build-stdlib.sh --force` to generate all 50 `stdlib/**/package.pb` files.
  - [ ] Hook `./tool/build-stdlib.sh` into `tool/build.sh`, `tool/build-cli.sh`,
    `tool/build-wheel.sh`, `tool/test-go.sh`, `tool/check.sh`, `tool/update-goldens.sh`, and
    `tool/build-proto.sh` (`--force`).
  - [ ] Run `gofmt` and verify targeted tests and generator idempotence:
    `./tool/build-stdlib.sh --force && ./tool/build-stdlib.sh && go test -v ./stdlib/... ./parser/...`.
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

---

## Phase 3: Shared Lazy Stdlib Cache in `program` & `CachingParser`, Removing `RegisterStdPackage` and `parser.init()`

- [ ] Task: Write failing tests for standalone `program` stdlib resolution and shared `CachingParser` fast path (Red Phase)
  - [ ] Update `program/get_import_test.go` to remove `RegisterStdPackage` calls and test standalone
    resolution of real embedded standard library packages (`"builtin"`, `"vector"`) via both
    `(*Program).GetImport` and `GetStdPackage` (without importing `parser`), including mutation
    isolation on `GetStdPackage` clones.
  - [ ] Update `parser/parse_cache_test.go` to assert that `(*CachingParser).ParsePackage` serves
    standard library packages via the shared `program.GetStdPackage` fast path without increasing
    `CachingParser`'s internal LRU `size` counter, preserves clone mutation safety, and falls back
    to live `ParsePackage` + LRU caching when `.gsl` source text does not match a pre-baked stdlib
    hash.
  - [ ] Remove the obsolete `init()` stub calling `program.RegisterStdPackage("builtin", ...)` in
    `internal/types/types_test.go`.
- [ ] Task: Implement lazy unmarshaling in `program` and lock-free stdlib lookup in `CachingParser` (Green & Refactor Phase)
  - [ ] Update `program/get_import.go` to delete `RegisterStdPackage` and `stdPkgs`, lazily
    unmarshal each standard library package at most once per process (`sync.OnceValue` over
    `stdpkg.Proto(path)`), return the shared read-only `*Package` from `(*Program).GetImport`, and
    return `proto.Clone(pkg).(*Package)` from `GetStdPackage`, with clear *what*-and-*why* comments.
  - [ ] Delete `init()` (and unused imports) from `parser/parser.go`.
  - [ ] Update `parser/caching.go` to delete `stdCachedPkg`, `stdCachedPkgs`, `sourcesKey`, and the
    pre-warm loop in `NewCachingParser()`, routing `(*CachingParser).ParsePackage` through
    `stdpkg.SourcesHash` -> `stdpkg.ImportPathByHash` -> `program.GetStdPackage` before checking
    the LRU cache, with clear *what*-and-*why* comments.
  - [ ] Run `gofmt` and verify targeted package tests pass with near-zero startup overhead:
    `go test -v ./program/... ./parser/... ./internal/types/... ./internal/link/... ./assembler/...`.
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

---

## Phase 4: Repository-Wide Verification, Guard Cleanup & Metrics Finalization

- [ ] Task: Run full quality checks and repository-wide test suites
  - [ ] Execute `./tool/check.sh` and `./tool/test.sh` across `omega-core`, verifying all Go and
    Rust tests pass with zero regressions and zero compiler golden diffs.
  - [ ] Verify `./tool/build-stdlib.sh` is a fast no-op on clean state and `git status` shows
    deterministic `stdlib/**/package.pb` files.
- [ ] Task: Audit created and modified Go code with `clean-go-test-guards`
  - [ ] Follow `.agents/skills/clean-go-test-guards/SKILL.md` across `internal/stdpkg`, `stdlib`,
    `program`, `parser`, and `stdlib/internal/gen` to remove any test-induced defensive checks,
    unused parameters, or dead unexported helpers while preserving boundary invariants and
    *what*-and-*why* comments.
  - [ ] Re-run `./tool/check.sh`.
- [ ] Task: Finalize Execution Metrics & Effort Tracking
  - [ ] Calculate total active **Implementation** time, accumulated **Code Review** time, and
    estimated **Effort** tokens across track transcripts, and record them in `plan.md`.
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)
