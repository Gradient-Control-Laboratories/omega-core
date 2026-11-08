# Golden verification: `metal` takes the measured colors

Each `metal.*` preset is now a measured substance with the finish the library
names: `physical`'s color (the Physically Based database) or `alloy`'s, and the
roughness `metal` always baked in. Cast iron takes iron's color, since the
database's Cast Iron is a spectrophotometer reading filed as Manmade, not a
metal's reflectance. Galvanizing gains an overload over a base, which moved its
stand-in into a helper, `galvanizedFlat`. Material - Coated calls galvanizing,
and its WGSL differs from before only in the colors below.

Nineteen goldens moved, all under `agent/wgsl/internal/codegen/testdata/`:
eighteen examples and the `TestBuild` fixture `color_and_material`. Regenerated
with Blake's go.

## What changed

**Color constants (70 lines).** The line is the same once numbers are stripped:
a `metal.*` color became the measured one.

| Preset | Before | After |
|---|---|---|
| `metal.steel` | `[0.77, 0.78, 0.78]` | unchanged (`alloy.steel` was always this) |
| `metal.stainless` | `[0.80, 0.81, 0.82]` | `physical.stainlessSteel`, `[0.8373, 0.8204, 0.7966]` |
| `metal.castIron` | `[0.45, 0.45, 0.46]` | `physical.iron`, `[0.7548, 0.7439, 0.7314]` |
| `metal.aluminum` | `[0.96, 0.96, 0.97]` | `physical.aluminum`, `[0.9621, 0.9654, 0.9658]` |
| `metal.titanium` | `[0.76, 0.74, 0.70]` | `physical.titanium`, `[0.6951, 0.6652, 0.635]` |
| `metal.zinc` | `[0.80, 0.82, 0.84]` | `physical.zinc`, `[0.9103, 0.928, 0.9381]` |
| `metal.brass` | `[0.95, 0.90, 0.69]` | `physical.brass`, `[0.9593, 0.8952, 0.6822]` |
| `metal.copper` | `[0.98, 0.82, 0.76]` | `physical.copper`, `[0.9695, 0.8112, 0.7497]` |
| `metal.gold` | `[1.0, 0.89, 0.62]` | `physical.gold`, `[1, 0.8927, 0.59]` |
| `metal.chromium` | `[0.77, 0.78, 0.77]` | `physical.chromium`, `[0.8289, 0.8461, 0.8548]` |

**Steel's color bound through a temporary (6 places).** Where an example names
steel's pair, the color is now `let c = vec3(…); let steelColor = c;`, the
value `alloy.steel` returns, rather than one line. The material is unchanged:
`metal` keeps its own `finish(rough)` record, so each preset still folds to
one constant and still has nothing for Adjust to lift.

## Determinism

`go test -count=1 ./internal/codegen/` passes against the regenerated goldens.
No other codegen golden moved.
