# omega-core

The GSL toolchain from omega-core, packaged as a Python wheel. It contains the `gsl` command, which
compiles, meshes, renders and queries implicit (signed-distance) models written in GSL. It also contains
a Python client, standard library only, for querying a model's field.

```sh
pip install omega-core
gsl compile -t stl -o part.stl part.gsl     # mesh: depth 9, dual contouring, over the @scene box
gsl compile -t 3mf -o part.3mf part.gsl     # one file, each returned body an object in its own color
gsl render --views iso,top --light -o part.png part.gsl
gsl eval --help                              # querying the field, as JSON
```

## Querying a model from Python

`omega_core.Session` keeps one `gsl eval --stdio` process running, and `Document` methods send it
questions:

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

| `Document` method | Returns |
|---|---|
| `describe()` | `Description`: `sliders` (`{name: (min, max)}`), `names`, `returns` |
| `eval(exprs, points, defs=)` | `Fields`, one `Field` per expression, found by place or by expression (`f["body"]`). A `Field` holds `values[i]` and `gradients[i]` at point `i`. |
| `extremum(expr, region=, mode=, lipschitz=)` | `Extremum`: `min` / `max`, `min_at` / `max_at`, `min_conditional_bound` / `max_conditional_bound`, `assumed_lipschitz` |
| `certify(expr, region=, threshold=, params= or over_sliders=)` | `Certificate`: `verdict` (`"proven"`, `"disproven"` or `"indeterminate"`), `counter_box`, `counter_params`, `reason` |
| `bounds(expr, region=)` | `Bounds`: `found`, `bbox_min`, `bbox_max`, `cell_size` |
| `surface(points, expr)` | a tuple of `SurfaceSample`: `point`, `normal`, `value`, `converged` |
| `volume(expr, region=)` | `Volume`: `lower_mm3`, `upper_mm3` |
| `render(out, expr, views=, light=, extent=)` | `Picture`, the PNG written: `path`, `width`, `height` |

How the arguments work:

- **Units and expressions:** lengths are millimetres, and a region is `(x0, y0, z0, x1, y1, z1)`.
  An `expr` is any GSL expression in the document's scope. When omitted, it is the union of the
  document's returns; `render` instead draws each return in its own colour.
- **`defs`:** GSL statements applied for one question only. A statement naming a declared slider sets it,
  for example `defs="skin = 1.6;"`.
- **Derivatives:** `.dp`, `.dtime` and `.d(var)` are ordinary expressions.
- **Failures:** a question that fails raises `omega_core.EvalError` in the compiler's words.
- **More detail:** the docstrings (`help(omega_core.Document)`) give every argument.

`certify` is a sound proof by interval arithmetic. `extremum` finds a true value of the field, but its
bound holds only under the Lipschitz constant it assumed (1 unless given). That is correct for a distance
field, and wrong for a sum of two fields, a lattice, or a blend that overstates distance.

`omega_core.executable()` is the path to `gsl`, and `omega_core.version()` is the release, for
example `"0.1.1"`.

## Platforms

Each wheel holds `gsl` and `compile` for one platform: macOS 11+ (arm64, x86_64) and Linux with glibc
2.28+ (x86_64, aarch64). `compile` links omega's render library statically, so neither program loads
anything but the system's libraries. `gsl serve` and `gsl fix` are not included.

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
