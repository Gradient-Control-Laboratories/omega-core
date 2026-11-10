# omega-gsl

omega-core's `gsl` program, packaged as a Python wheel with a small launcher. It is what
text-to-cad's `cad[omega]` installs to mesh GSL implicit models.

```sh
pip install omega-gsl
gsl compile -t stl --algo mdc --subdivision 8 --bbox auto -o part.stl part.gsl
```

```python
import omega_gsl

omega_gsl.executable()            # Path to gsl
omega_gsl.executable("compile")   # Path to compile, which `gsl compile` runs
omega_gsl.version()               # "0.1.0"
```

Each wheel holds `gsl` and `compile` for one platform: macOS 11+ (arm64, x86_64) and Linux with glibc
2.28+ (x86_64, aarch64). `compile` links omega's render library statically, so neither program loads
anything but the system's libraries. `gsl serve` and `gsl fix` are not included.

## Building

From an omega-core checkout with its submodules (`git submodule update --init --recursive`), Go,
Rust and CMake:

```sh
tool/build-wheel.sh          # builds the programs into target/static/bin, then the wheel into dist/
```

The wheel's version is omega-core's own (`internal/version/version.go`). CI
(`.github/workflows/wheels.yml`) builds and tests every platform's wheel and publishes them to PyPI.

## Licence

omega-core and this launcher are MIT. The programs link code under other licences: the Go and Rust
standard libraries, three Go modules, and omega's render library, with Fidget and dynasm (MPL-2.0),
OpenVDB and oneTBB (Apache-2.0), Imath and c-blosc (BSD-3-Clause), zlib, about 200 Rust crates, and
egui's default fonts. `omega_gsl/licenses/THIRD_PARTY_LICENSES.txt` in the wheel lists every one with
its licence texts; the build writes it from the sources it built (`tool/third_party_licenses.py`), and
no wheel is built without it. The MPL-2.0 source is omega-core's repository at the commit the wheel's
`omega_gsl/_bin/manifest.json` names, and crates.io.
