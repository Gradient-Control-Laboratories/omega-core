# omega-core

The GSL toolchain from omega-core, packaged as a Python wheel. GSL is omega's language for implicit
(signed-distance) models. The wheel contains the `gsl` command and a Python client, standard library only,
for querying a model's field.

```sh
pip install omega-core
```

## What it does

- **Compile** a GSL model to a WebGPU shader (WGSL), a mesh (STL), a picture (PNG), a bytecode tape, Web
  Audio JavaScript, or its AST as JSON. Each comes with a JSON record of the model's sliders, controls and
  scene.
- **Mesh** by dual contouring or marching cubes, at any octree depth, over the `@scene` box, the box the
  field fills, or a box you name.
- **Render** the field itself on the CPU, without meshing it: one or several cameras, on a dark or light
  ground.
- **Query the field**, from the command line as JSON or from Python:
  - values and gradients at points;
  - derivatives in space, in time, and with respect to any slider;
  - a model's sliders and declared names;
  - global extrema;
  - sound proofs that an expression stays above a threshold over a region, for every slider value;
  - the occupied box;
  - points projected onto the surface;
  - a bracketed volume;
  - pictures with sliders set or the part cut open.

All lengths are millimetres.

## Command line

| Command | Does |
|---|---|
| `gsl compile [flags] part.gsl` | compile to a target (below) |
| `gsl render [flags] part.gsl` | shorthand for `gsl compile -t png` |
| `gsl eval REQUEST.json`, `gsl eval -`, `gsl eval --stdio` | answer field queries as JSON |
| `gsl skill`, `gsl skill PACKAGE...`, `gsl skill --all` | the GSL reference for a model to read (below) |
| `gsl --version`, `gsl help [command]` | the release; a command's help |

`python -m omega_core` runs `gsl` too. The wheel does not include `gsl serve` (omega's interactive viewer)
or `gsl fix`.

### `gsl compile`

| `-t` | Writes | Default output |
|---|---|---|
| `wgsl` (default) | a WebGPU raymarching shader; `--standalone` appends a runner harness | `part.wgsl` |
| `stl` | a binary STL mesh | `part.stl` |
| `png` | a picture of the field, raymarched on the CPU | `part.png` |
| `tape` | the field as a bytecode tape (for Fidget) | `part.tape` |
| `audio` | Web Audio JavaScript for the model's sound | `part.js` |
| `ast` | the linked program as JSON | `part.ast.json` |

General flags:

| Flag | Meaning |
|---|---|
| `-o PATH` | where to write (`-` for stdout) |
| `-m PATH`, `--no-meta` | where to write the companion JSON (sliders, controls, scene, diagnostics), or not at all |
| `-I DIR` | a root for package resolution (repeatable; default: the model's folder) |
| `--time S` | the animation time |
| `-O 0` | debug build: keeps dead code |
| `-v` | verbose |

With no source files, the model is read from stdin.

Mesh flags (`-t stl`):

| Flag | Default | Meaning |
|---|---|---|
| `--algo` | `mdc` | `mdc` (manifold dual contouring: flat faces stay flat, sharp edges sharp) or `marching-cubes` |
| `--subdivision` | `9` | octree depth: 2^N cells across the box |
| `--bbox` | the `@scene` box | `x0,y0,z0,x1,y1,z1`, or `auto` for the box the field fills. The `@scene` box is an edge of the scene's size centred on the origin; without `@scene`, `auto` is used. |

Picture flags (`-t png`, `gsl render`):

| Flag | Default | Meaning |
|---|---|---|
| `--views` | `iso` | presets (`iso`, `iso-below`, `front`, `back`, `left`, `right`, `top`, `bottom`) or `x,y,z` directions, comma-separated; several tile two across |
| `--size` | `640x480` | each view's size in pixels |
| `--light` | off | a light studio ground instead of the dark canvas |
| `--extent` | `0` | the scene's half-size in mm; 0 frames the geometry |

### `gsl eval`

`gsl eval` answers one JSON request from a file or stdin (`-`). With `--stdio` it answers one request per
line until stdin closes, keeping its evaluator between requests so a repeated question isn't reassembled.
A request names its `op` and its document, by path (`"source"`) or by text (`"code"`):

```json
{"id": 1, "op": "eval", "source": "part.gsl", "exprs": ["body", "body.dp"], "points": [[0, 0, 5]]}
```

| `op` | Fields | Answers |
|---|---|---|
| `eval` | `exprs`, `points` | each expression's value and gradient at each point |
| `extremum` | `expr`, `region`, `mode` (`min`, `max`, `both`), `tolerance`, `lipschitz` | the global min and/or max, where it is, and a bound that assumes the Lipschitz constant |
| `certify` | `expr`, `region`, `threshold`, `tolerance`, `params` or `overSliders` | a sound proof that `expr >= threshold` over the region and the parameter ranges, or a box where it fails |
| `bounds` | `expr`, `region`, `resolution` | the box where `expr < 0` |
| `surface` | `expr`, `points`, `maxIters`, `tolerance` | each point moved onto `expr = 0`, with the normal there |
| `volume` | `expr`, `region`, `relativeTolerance` | the volume where `expr < 0`, bracketed (no `defs`) |
| `describe` | (none) | the sliders and their ranges, top-level names and returns |
| `render` | `out`, `expr`, `views`, `width`, `height`, `light`, `extent`, `time` | a PNG at `out`, as `gsl compile -t png` draws it |

- **Responses:** each is `{"id", "ok": true, ...}` or `{"id", "ok": false, "error"}`. A value that is not
  finite is `null`.
- **Exit codes** (one request): 0 answered, 1 the query failed, 2 a malformed request.
- **More detail:** `gsl eval --help`.

### `gsl skill`

`gsl skill` prints the GSL reference, written for a model to read before it writes GSL and generated from
the standard library this release compiles. Without arguments it prints the language, the built-ins, how
packages are referenced and one line per package, about 70 KB. `gsl skill PACKAGE...` prints those
packages in full, named by import path or prefix (`gsl skill builtin wood`), and `gsl skill --all` prints
everything, about 470 KB. From Python, `omega_core.skill()` and `omega_core.skill("wood")` return the same text.

## Python

```python
import omega_core

with omega_core.Session() as gsl:
    part = gsl.document("part.gsl")
    part.describe().sliders                        # {"skin": (1.2, 4.0)}
    f = part.eval(["body", "body.d(skin)"], points=[(40, 0, 9)], defs="skin = 1.6;")
    f["body"].values[0], f["body"].gradients[0]    # value and (dx, dy, dz) at the point
    part.certify("body", region=(-60, -60, 0, 60, 60, 50), threshold=0, over_sliders=True).verdict
    part.render("cut.png", "max(body, p.y)", views="iso,top", light=True)
```

### `Session`

`omega_core.Session(source_paths=(), program=None)` runs one `gsl eval --stdio` process until `close()`
or the end of the `with` block.

- `source_paths` are its `-I` roots; by default, each document's folder.
- `program` is the `gsl` to run; by default, this package's.

| Method | Returns |
|---|---|
| `document(path)` | a `Document` read from `path` again for each question, so an edit between questions is seen |
| `code(text)` | a `Document` given as its text |
| `request(mapping)` | one raw `gsl eval` request's response, as a dict |
| `close()` | stops the process |

### `Document`

| Method | Returns |
|---|---|
| `describe()` | `Description` |
| `eval(exprs, points, *, defs="")` | `Fields` |
| `extremum(expr=None, *, region, mode="min", tolerance=0, lipschitz=0, defs="")` | `Extremum` |
| `certify(expr=None, *, region, threshold, tolerance=0, params=None, over_sliders=False, defs="")` | `Certificate` |
| `bounds(expr=None, *, region=None, resolution=0, defs="")` | `Bounds` |
| `surface(points, expr=None, *, max_iters=0, tolerance=0, defs="")` | a tuple of `SurfaceSample` |
| `volume(expr=None, *, region, relative_tolerance=0)` | `Volume` |
| `render(out, expr=None, *, views="iso", width=640, height=480, light=False, extent=0, time=0, defs="")` | `Picture` |

A zero `tolerance`, `lipschitz`, `resolution`, `max_iters` or `relative_tolerance` takes the evaluator's
default. `views` is a string as for `--views`, or a list of them.

### Results

All results are frozen dataclasses of floats and tuples.

| Type | Fields |
|---|---|
| `Description` | `sliders` (`{name: (min, max)}`), `names`, `returns` |
| `Fields` | the `Field`s an eval asked for, by place or by expression (`f["body"]`) |
| `Field` | `expr`; `type` (the GSL type, `"num"` for a scalar); `values[i]` and `gradients[i]` at point `i`: a float and a `(dx, dy, dz)` for a scalar, a tuple of them per component for a vector |
| `Extremum` | `min` and `min_at`, `max` and `max_at`; `min_conditional_bound` and `max_conditional_bound`; `min_converged` and `max_converged`; `assumed_lipschitz`, `tolerance`, `evals` |
| `Certificate` | `verdict` (`"proven"`, `"disproven"` or `"indeterminate"`) and `.proven`; `counter_box`, `counter_upper` and `counter_params` where it fails; `reason`, `worst_box`, `worst_lower` and `worst_params` when undecided; `params`, `region_lower`, `region_upper`, `evals` |
| `Bounds` | `found`, `bbox_min`, `bbox_max`, `cell_size`, `resolution`; `min_value` and `min_at`, the least sample |
| `SurfaceSample` | `point`, `normal`, `value`, `converged` |
| `Volume` | `lower_mm3`, `upper_mm3`, `evals` |
| `Picture` | `path`, and the sheet's `width` and `height` in pixels |

A question that fails raises `omega_core.EvalError` in the compiler's words.

`omega_core.skill(*packages, everything=False)` is the text `gsl skill` prints (above).

`omega_core.executable()` is the path to the bundled `gsl`; it raises `FileNotFoundError` on a platform
the wheel was not built for. `omega_core.version()` is the release, such as `"0.1.1"`.

## Conventions

- **Expressions:** an `expr` is any GSL expression in the document's scope. When omitted, it is the union
  of the document's returns; `render` instead draws each return in its own colour.
- **`defs`:** GSL statements for one question only. One that names a declared name, `@range` sliders
  included, replaces it for that question (`"skin = 1.6;"`), even outside the slider's range.
- **Derivatives:** `f.dp` (space), `f.dtime` (time) and `f.d(var)` (a top-level scalar, such as a slider)
  are expressions like any other. `-f.d(var) / length(f.dp)` is how fast the surface moves along its
  normal as `var` grows.
- **Regions:** `(x0, y0, z0, x1, y1, z1)` in mm. `extremum`, `certify` and `volume` need one, because a
  proof's region is part of its claim.
- **Numbers:** values are float32, as the evaluator computes them. A value that is not finite is `None`.
- **What each answer guarantees:**
  - `certify` is a sound proof by interval arithmetic.
  - `extremum`'s sample is a true value of the field, but its bound holds only under the Lipschitz
    constant it assumed (1 unless given). That is correct for a distance field, and wrong for a sum of two
    fields, a lattice, or a blend that overstates distance. To claim a minimum, certify it.
  - `bounds` can fall a cell short of the surface.

## Platforms

Each wheel holds `gsl` and `compile` for one platform: macOS 11+ (arm64, x86_64) and Linux with glibc
2.28+ (x86_64, aarch64). `compile` links omega's render library statically, so neither program loads
anything but the system's libraries.

## Building

You need an omega-core checkout with its submodules (`git submodule update --init --recursive`), plus
Go, Rust, CMake and Python's `build` package:

```sh
tool/build-wheel.sh          # builds the programs into target/static/bin, then the wheel into dist/
```

The wheel's version is omega-core's own (`internal/version/version.go`). CI
(`.github/workflows/wheels.yml`) builds and tests every platform's wheel and publishes them to PyPI.

## Licence

omega-core and this launcher are MIT. The programs link code under other licences:

- the Go and Rust standard libraries, three Go modules, about 200 Rust crates, and egui's default fonts;
- omega's render library, with Fidget and dynasm (MPL-2.0), OpenVDB and oneTBB (Apache-2.0), Imath and
  c-blosc (BSD-3-Clause), and zlib.

`omega_core/licenses/THIRD_PARTY_LICENSES.txt` in the wheel lists every one with its licence text. The
build writes it from the sources it built (`tool/third_party_licenses.py`), and no wheel is built without
it. The MPL-2.0 source is omega-core's repository at the commit named in the wheel's
`omega_core/_bin/manifest.json`, together with crates.io.
