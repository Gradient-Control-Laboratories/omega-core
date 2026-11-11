# `gsl eval`: Field Queries from the Command Line and Python

## 1. Overview & Problem Statement

`gradientcontrol.com/omega/core/eval` answers the questions a designer asks of a field: its value
and gradient at points, `.dp` / `.dtime` / `.d(var)` derivatives, a global extremum, a sound
proof that `expr >= threshold` over a region (also over slider ranges), the occupied box, points
projected onto the surface, and a bracketed volume. Only Go code can call it. The `gsl` command
uses two of these calls (`Bounds` for `compile --bbox auto`, the lit renderer for `compile -t png`),
and the `omega-core` wheel exposes nothing but `executable()` and `version()`.

So a Python user cannot ask "what is `wall` at these points, how does it move with `skin`, and is
it at least 1.2 mm everywhere for every slider setting?" without writing Go. This track adds one
command, `gsl eval`, that speaks JSON, and a standard-library Python client for it in `omega_core`.

---

## 2. Public API Surface Policy (One New Go Method)

- **`eval` (`gradientcontrol.com/omega/core/eval`):** unchanged. `gsl eval` is a thin JSON layer
  over `eval.New(...)` and the methods it already has (`EvalGsl`, `Extremum`, `Certify`,
  `CertifyParams`, `CertifyOverSliders`, `Bounds`, `SurfacePoint`, `Volume`, `SliderRanges`,
  `TopLevelNames`, `TopLevelReturnExprs`).
- **`Evaluator.RenderViews`** (new): `Render` from any cameras (`lib.LitView`: presets or `x,y,z`
  directions, each with its ground) over a scene of half-size `extent` at animation time `timeAt`,
  so `render` matches `compile -t png`. `Render` now calls it, unchanged in behaviour (a test pins
  that both draw the same bytes for the same cameras).
- **New, outside Go:** the `gsl eval` command and its JSON request/response schema (Section 4),
  and `omega_core.Session` in the wheel's Python package (Section 6).

---

## 3. Where the Command Lives

`gsl` runs a sibling program per command (`cmd/gsl/internal/cli`: `compile`, `serve`, `fix`). The
wheel ships `gsl` and `compile`; `compile` (~25 MB) carries Fidget and the render crate statically.

- **Option A (built): a mode of the `compile` program.** `gsl eval ARGS` runs
  `compile eval ARGS`. The wheel keeps one Fidget-linked program; `hatch_build.py`'s `PROGRAMS`
  and the wheel layout are unchanged. Costs a small change to `MapSubcommand` (a command that maps
  to a binary *and* a leading argument).
- **Option B: a new `eval` program.** Cleaner separation; adds a second ~25 MB static binary to
  every wheel and a third entry to `PROGRAMS`, `link_static.go` checks and `test_wheel.py`.

---

## 4. Functional Requirements: the Command

### 4.1 Invocation

```
gsl eval [-I DIR]... REQUEST.json        # one request, one response on stdout
gsl eval [-I DIR]... -                   # the same, the request on stdin
gsl eval [-I DIR]... --stdio             # a session: one JSON request per line in, one per line out
```

- `-I` is `compile`'s `--source-path`, passed to the evaluator as `assembler.SourcePaths`.
- A session keeps one `Evaluator` with `CacheQueries(n)` on (n = 64), so a query asked again of an
  unchanged document is not assembled again. One-shot mode leaves the cache off.
- One-shot exit codes: 0 answered (a disproven claim is an answer), 1 the query failed (GSL error,
  bad region), 2 a malformed request. A session never exits on a failed request; it answers it.

### 4.2 Common Request Fields

| Field | Meaning |
|---|---|
| `id` | Echoed in the response; any JSON value. |
| `op` | `eval`, `extremum`, `certify`, `bounds`, `surface`, `volume`, `describe`. |
| `source` / `code` | A `.gsl` path (read on every request, so edits are seen; relative imports resolve from its folder), or the text itself. Exactly one. |
| `defs` | GSL statements in the document's scope before the query. A statement naming a declared top-level name replaces that declaration in place (existing `prepareQuery` behaviour): `"skin = 1.6;"` sets a slider for this query. |
| `expr` | One scalar expression. When omitted: the document's returns, unioned (`min(r1, r2, ...)`). |
| `region` | `[x0, y0, z0, x1, y1, z1]` mm. Required for `extremum`, `certify` and `volume` (no silent 1024 mm default: a proof's region is part of its claim); optional for `bounds`. |

### 4.3 Operations

| `op` | Request (beyond the common fields) | Response |
|---|---|---|
| `eval` | `exprs: [..]`, `points: [[x,y,z], ...]` | `groups: [{expr, type, values, gradients}]`: `values[component][point]`, `gradients[component][point] = [dx, dy, dz]`, as `GradGroup` orders them. A call returning several values yields several groups for one `expr`. |
| `extremum` | `mode` (`min`/`max`/`both`), `tolerance`, `lipschitz` | `ExtremumResult`'s fields, plus `assumes: {lipschitz: L}` (Section 5). |
| `certify` | `threshold`, `tolerance`; `params: {name: [lo, hi]}` or `overSliders: true` | `verdict` (`proven` / `disproven` / `indeterminate`), `reason`, `counterBox`, `counterUpper`, `counterParams`, `worstBox`, `worstLower`, `worstParams`, `params`, `regionLower`, `regionUpper`, `evals`. |
| `bounds` | `resolution` | `BoundsResult`'s fields (`found`, `bboxMin`, `bboxMax`, `cellSize`, `minValue`, `minAt`). |
| `surface` | `points`, `maxIters`, `tolerance` | `samples: [{point, normal, value, converged}]`. |
| `volume` | `relativeTolerance`; no `defs` (`Evaluator.Volume` takes none) | `{lowerMm3, upperMm3, evals}`. |
| `describe` | (none) | `sliders: {name: [min, max]}`, `names: [...]`, `returns: [...]`. |
| `render` | `out` (required), `views` (compile's `--views` syntax), `width`, `height` (640 by 480), `light`, `extent`, `time`; `expr` defaults to the returns, each in its own colour | `{out, width, height, bytes}`: the PNG written, byte for byte what `compile -t png` writes with the same options. |

Every response is `{id, ok: true, ...}` or `{id, ok: false, error}`, with `error` the evaluator's
message verbatim (GSL parse and build errors already read as ordinary GSL errors). Field names are
camelCase, matching the JSON tags `GradSample` and `VolumeBounds` already carry.

### 4.4 Found While Building

- **`defs` could not override the first name a document declares, nor any `@range` slider
  (fixed in this track).** `prepareQuery` splits the document into statements at `;` and names each
  by its leading identifier (`leadingIdentifier`). The `@scene(...)` line has no `;`, so the first
  declaration was read as part of a statement starting with `@`, with no name; a declaration behind
  its annotation (`@range(5, 15) r = 10.0;`) had none either; and `r.min = 5.0;` was named `r`.
  An override was then appended as a second declaration (`"r" is an existing variable`) or replaced
  a `.min` line. `leadingIdentifier` now skips `@scene(...)` and the declaration's annotations and
  names no property line. An override keeps what precedes the declaration (`@scene`, comments)
  and replaces the declaration with its annotation, so a query may set a slider outside its
  `@range`. `TopLevelNames` now includes annotated declarations.
- **`bounds` can fall short of the surface by up to a cell.** A radius-10 sphere reports a box of
  ±8 with 8 mm cells. `compile --bbox auto` already grows the box by two cells; the Python client's
  `Bounds` says so.
- **`volume` at 1% relative tolerance takes about 3 s** (51M evaluations for a sphere); at 5% about
  0.1 s.

### 4.5 Out of Scope (v1)

- Documents with mesh assets (`importStep`, `importMesh`, `bvhSdf`, `vdbSdf`): they need the canvas
  agent's `AssetFetcher`. Refused with a sentence naming the asset call.
- Certify budget knobs (`SetCertifyBudget`): the default budget, reported through `reason`.

---

## 5. Soundness Is Labelled, Not Implied

`Extremum` and `Certify` answer different questions, and the schema must keep them apart:

- **`certify` is sound** by interval arithmetic. `proven` holds for every point in the region (and
  every parameter value in range); no constant is assumed.
- **`extremum` is conditional.** Its reported point and value are a true field sample, but the
  bracket (`minLower`, `maxUpper`) holds only if `|grad expr| <= lipschitz` everywhere. With no
  `lipschitz` the evaluator assumes 1: right for one distance field and for min/max of distance
  fields, wrong for a sum of two (2), for a TPMS lattice, and for nested exact blends that claim
  more distance than there is. A default-1 bracket on those can prune a real basin.
- Hence every `extremum` response carries `assumes: {lipschitz: L}`, and the Python client names
  the bracket `conditional_bound`. To *claim* a minimum, certify it: `certify` with
  `threshold = min - margin`.

---

## 6. Functional Requirements: Python (`omega_core`)

Standard library only, like the rest of the package. One `gsl eval --stdio` process per session.

```python
import omega_core

with omega_core.Session(source_paths=[]) as gsl:
    part = gsl.document("lamp_base.gsl")

    part.describe().sliders                        # {"skin": (1.2, 4.0), ...}
    r = part.eval(["body", "body.d(skin)"], points=[(40, 0, 9)], defs="skin = 1.6;")
    r["body"].values[0], r["body"].gradients[0]    # value, (dx, dy, dz) at each point
    r["body.d(skin)"].values[0]                    # d body / d skin; with body.dp: -f.d(skin)/|grad f|
                                                   # is the surface's normal speed as skin grows

    m = part.extremum("body", region=box, mode="min", lipschitz=1.0)
    m.min, m.min_at, m.min_conditional_bound       # bound holds only under lipschitz

    c = part.certify("wall - 1.2", region=box, threshold=0, params={"skin": (1.2, 4.0)})
    c.verdict, c.counter_box, c.counter_params     # "disproven", box, {"skin": (1.2, 1.4)}

    part.render("cut.png", "max(body, p.y)", views="iso,top", light=True, defs="skin = 1.6;")
```

- Results are frozen dataclasses of tuples and floats (no NumPy dependency). An eval's `Field` is
  point-major (`values[i]` is the value at point `i`: a float for a scalar, a tuple for a vector),
  though the wire is component-major as `GradGroup` is.
- A failed request raises `omega_core.EvalError` carrying the evaluator's message.
- `document(path)` sends the path on each request; the session process reads the file, so an edit
  between calls is seen.

---

## 7. Costs to Measure Before Optimising

- **A parameter sweep assembles once per value.** The query cache is keyed by the assembled
  source text, and `defs` changes that text. Measure assembly time on a library-heavy document
  (fastener + drill + tpms) before deciding whether a sweep needs parameters as program inputs.
- **Points per request.** `EvalGsl` evaluates all points in one batch; the client should send
  thousands per request rather than one point per round trip.
- **Precision.** Samples are float32: about 7 significant digits, so ~1e-5 mm on a 100 mm part.

---

## 8. Tests

- Go, `cmd/compile/internal/evalcmd`: each op on a sphere with known answers; the union of the
  returns as the default expr; a GSL error relayed verbatim (exit 1) and each malformed request
  (exit 2); `defs` overriding an `@range` slider declared first; `.d(var)`; a session answering a failed request and a bad
  line and then good ones; non-finite numbers as `null`.
- Go, `render`: the picture at `out` (views tiled two across, a direction, the light ground, the
  default size); `defs` and `expr` reach it; refusals write nothing; and `cmd/compile` pins that it
  is byte for byte `compile -t png`'s. `internal/gsleval`: `Render` and `RenderViews` draw the same.
- Go, `cmd/gsl/internal/cli`: `gsl eval ARGS` and `gsl help eval` run `compile eval`.
- Wheel, `test_wheel.py`, on every platform: `omega_core.Session` against the real program (values,
  gradients, `.d(var)`, `defs`, certify over sliders, extremum's assumption, a failure that the
  session survives) and `gsl eval -`'s exit codes. The client is tested against the real program
  rather than a stand-in, so the repository gains no Python test tooling beyond the wheel's test.
- Not tested: that a repeated query is served from the cache (not observable from outside).

---

## 9. Open Questions

1. Should `lipschitz` be required (no default) in the Python client, so a caller must state it?
2. Default `expr` when omitted: the union of returns (as the private `gslinfo` does), or refuse?
3. text-to-cad: a `cad gsl eval` / `certify` surface later, or Python through `cad run` only?
