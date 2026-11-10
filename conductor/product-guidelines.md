# Product Guidelines: Omega Core (`omega-core`)

## Primary Principles

- **Clean Compiler-Style Public API:** Expose cohesive, minimal top-level packages (`parser`,
  `assembler`, `eval`, `program`, `stdlib`, `backend/*`) configured via `With...` functional
  options rather than leaking internal AST passes (`internal/*`).
- **Zero External Service Coupling:** `omega-core` is a pure library and CLI SDK with zero
  dependencies on Stirigma, gRPC buses, or cloud infrastructure.
- **Fast Startup & Compilation:** Prioritize low initialization overhead, fast compilation, and
  minimal heap allocations along error-free compilation paths.
- **Precision & Actionable Diagnostics:** Report lexical, semantic, type, and backend errors via
  structured `program.Error` and `program.Diagnostic` values with accurate file, line, and column
  spans.

## Interactive Viewer UX (`examples/viewer`)

- **Theme:** Clean dark-mode CAD viewport paired with a monospaced CodeMirror 6 GSL editor.
- **Responsiveness:** Real-time WebGPU raymarching, parameter sliders, and unobtrusive error
  banners.

## Communication & Voice

- **Tone:** Technical, precise, and concise without marketing fluff.
- **Invariants:** Document package and function invariants explicitly in GoDoc comments at public
  API boundaries.

