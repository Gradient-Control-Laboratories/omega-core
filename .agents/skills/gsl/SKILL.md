---
name: gsl-authoring
description: Author and verify GSL (.gsl) geometry files from the command line — evaluate expressions, box shapes, measure certified clearances, prove non-interference, and project points onto surfaces. Use whenever writing or editing .gsl documents and before declaring any spatial relationship (touching, clearing, containing) correct.
---

# Skill: GSL Authoring with Verified Geometry

GSL documents are implicit geometry: every shape is a signed-distance field (SDF) over the
implicit point `p` — negative inside, positive outside, magnitude = distance to the surface.
Read `examples/viewer/docs/gsl_description.md` and `examples/viewer/docs/gsl_libraries.md` (or
`program/preamble/`) for the language and standard libraries; this skill is about *verifying* what
you write.

## Toolchain & Evaluation

Build the `omega-core` CLI tools once per session:

```bash
./tool/build-cli.sh
./bin/gsl help
```

For spatial evaluation, bounds, extrema, interval certification, and surface projection, use the
`eval` package (`gradientcontrol.com/omega/core/eval`) or the `gsl` evaluation CLI if available.

## The Contract for Changing Geometry

- Open with a brief before writing any code: the target dimensions, datums, and frames, and the
  acceptance checks you will run (`Eval`, `Bounds`, `Extremum`, `Certify`, or `bin/gsl compile`).
- Run every check in the brief before calling the work done; when one fails, apply the smallest
  fix that addresses the failure and re-run that check. Never weaken a check to make it pass.

## Conventions That Matter

- GSL matrices are lists of COLUMNS: `[[col0...], [col1...], ...]`.
  `vector.multiplyPoint(m, p)` transforms a point (w = 1); `vector.multiplyVector(m, d)` transforms
  a direction (w = 0, translation ignored). Place a shape under transform `T` with
  `remap(shape, vector.multiplyPoint(matrix.inverse(T), p))`.
- Rotation constructors: `matrix.rotationEuler` (4x4), `matrix.rotationEuler3` (3x3),
  `matrix.rotation(...)`. Do not invent constructors that are not in `gsl_libraries.md`.
- Prefer vector math (`repeatCentered(p.xy, cell)`, `dot(v, v)`, `length(v)`) over per-component
  scalar code.

## Known Failure Classes

- Double-call: `shape(args)` already yields the field value — `shape(args)()` is a parse error.
- `remap`'s first argument must be a field VALUE, so an inline thunk must be called:
  `remap(() { return cyl(r, h); }(), q)` works; `remap(() { ... }, q)` fails with
  `expected numeric type, got FnType`.
- `remap` takes a POINT second argument: `p.xy` is a vec2 and will not build — use a vec3 (such as
  `p.zyx` or `[p.x, p.y, 0.0]`).
- Builds but shows nothing: check the bounding box of each returned shape; if empty, the field is
  positive everywhere — check sign convention (negative = inside) and dimension scale.

