# omega_tape — Python wrapper for GSL implicit fields

A native Python extension (pyo3) that lets Python callers read **omega tapes** in-process,
including `bvhSdf` / `vdbSdf` oracles. It wraps `omega-core`'s `render` lib (fidget +
OpenVDB), which is statically embedded, so the built module is self-contained.

```
                     backend/tape/tape.proto  (single source)
                       /                     \
          protoc --go_out              protoc --python_out  +  prost-build
              |                              |                      |
        tape.pb.go                     tape_pb2.py             (render lib)
       (Go: serialize_tape)        (Python: author tape)   (replay + eval/mesh)
                       \                     |                      /
                        \--------- tape::Tape  (bytes) ------------/
                                            |
                                   omega_tape.Tape(bytes)
```

The **omega tape** is the post-GSL-lowering fidget `TreeOp` DAG as protobuf —
content is fidget-shaped (plus `omega-core`'s `MeshOracle` leaves), in the
`backend/tape` format. All GSL semantics (inlining, map/reduce unroll, const-eval) are already
resolved by the Go `build.Build` pass, so replay needs no GSL machinery.

## Build

The module is a `cdylib` built against Python's stable ABI (abi3, 3.9+), so one
build works across interpreters (verified: built on 3.14, imported on 3.10).

With maturin (preferred for packaging):

```sh
cd internal/fidget/internal/tape-py
PYO3_USE_ABI3_FORWARD_COMPATIBILITY=1 maturin build --release
```

Or with plain cargo (no maturin), then rename the artifact to an importable
module name:

```sh
PYO3_USE_ABI3_FORWARD_COMPATIBILITY=1 cargo build -p omega-tape --release
cp target/release/libomega_tape.dylib /somewhere/on/sys.path/omega_tape.so   # .so on Linux too
```

`PYO3_USE_ABI3_FORWARD_COMPATIBILITY=1` is only needed when building against a
Python newer than the pyo3 release knows about; it is not needed at import time.

## Usage

```python
import numpy as np
import omega_tape

# Mesh oracles travel with the tape: the source mesh for each one and the scene
# bbox are embedded, so Tape(bytes) builds and registers them on load. A
# spinodoid is baked on load from the arguments it carries — see "Spinodoid
# oracles" below.
field = omega_tape.Tape(tape_bytes)

P = np.ascontiguousarray(pts, dtype=np.float32)   # (N,3)
values = field.eval(P)                            # (N,)   float32
grads  = field.grad(P)                            # (N,4)  [v, dx, dy, dz]
lo, hi = field.interval([xmin,ymin,zmin, xmax,ymax,zmax])
domain = field.bbox                               # [xmin..zmax] or None
```

### Meshing

`mesh` separates the two choices that used to be hidden in one call: **how** to
mesh (the algorithm) and **what** representation you get (the returned `Mesh`).

```python
m = field.mesh(depth=6)                            # how: marching cubes (default)
m = field.mesh(depth=6, algorithm="dual_contour")  # or dual contouring (sharp features)
V = m.vertices       # (V,3) float32   — what: in-memory arrays
T = m.triangles      # (T,3) uint32
stl = m.to_stl()     # what: binary STL bytes (a serialization of the same mesh)
```

`bbox` defaults to the tape's scene bbox if omitted (`field.mesh(bbox=[xmin..zmax], depth=…)`
to override); `output=` selects one output of a multi-output tape. Marching
cubes returns a triangle soup (`V == 3*T`); dual contouring returns an indexed
mesh.

### Choosing the oracle on load (BVH vs VDB)

A tape always carries the source **mesh** for each `bvhSdf`/`vdbSdf` asset, so
the *reader* picks how to evaluate it — there is no oracle choice baked into the
file:

```python
omega_tape.Tape(b)                               # "auto": as the GSL author wrote
omega_tape.Tape(b, oracles="bvh")                # exact distance everywhere (slower)
omega_tape.Tape(b, oracles="vdb")                # fast, near-surface
omega_tape.Tape(b, oracles="vdb", resolution=256)# finer VDB (default ~128)
omega_tape.Tape(b, oracles="vdb", half_band=16)  # wider accurate band (default 3)
```

- **`bvh`** — exact closest-point distance with generalized-winding-number sign;
  valid at any range (what Walk-on-Spheres wants), but slower.
- **`vdb`** — an OpenVDB narrow-band level set rebuilt from the mesh at load.
  Fast, but distance is only accurate within a few voxels of the surface (see
  half-band below); far from the surface it clamps to a constant.
- **`resolution`** is the VDB voxel count along the longest bbox axis. Higher =
  finer surface but a *narrower* accurate band; `None` uses the tape's value or
  a default.
- **`half_band`** is how many voxels wide the accurate shell is (default 3). The
  band reaches `voxel_size * half_band` in world units, so widen it to get valid
  distance further from the surface — at the cost of voxelization time/memory.

### Per-payload settings

A tape can carry several meshes and several spinodoids. `oracles`,
`resolution`, `half_band` and `backend` each take either one value, for every
payload it applies to, or a dict keyed by payload asset id, for those payloads
alone; a payload the dict leaves out takes the default. `tape_oracle_payloads`
lists the ids:

```python
omega_tape.tape_oracle_payloads(b)   # [(1, "bvh"), (2, "bvh"), (0xe05b1000, "spinodoid")]
omega_tape.Tape(b, oracles={2: "vdb"}, resolution={2: 256},
                backend={0xE05B1000: "manufacturing"})
```

A key must be a payload in the tape, and of a kind the setting applies to: a
mesh takes `oracles` / `resolution` / `half_band`, a spinodoid `backend`.

Each `Tape` binds its payloads under oracle ids of its own, so two tapes in one
process never read each other's oracles, even when the tapes share asset ids
and are loaded with different settings. The oracles are released when the
`Tape` is dropped.

`register_mesh_oracle` / `register_vdb_oracle` remain for advanced manual
authoring (a tape that references oracles by id without embedding the mesh);
not needed for tapes `omega-core` produces.

The `noiseHash3` oracle (procedural cellular/Worley noise) needs no reader
choice and carries no payload: its baked 64³ tile is a build-time constant baked
into the render lib (the same bytes the GPU path uses), so a noise field
serializes as plain oracle leaf ids and `Tape(...)` registers the channel
oracles from the library-baked tile automatically.

### Spinodoid oracles: baked on load, as the backend you choose

A spinodoid leaf carries only an id, but the tape carries the field's
*arguments* beside it (`tape.proto`'s `SpinodoidParams`), and `Tape(...)` bakes
the kernel table from them on load, with the same bake `omega-core`'s Go side uses
(tape-core `spinodoid/bake.rs`, pinned to it). The same arguments describe two
fields, and `backend` picks which one every spinodoid in the tape is rebuilt
as, whatever the payload's `fast` flag says:

```python
field = omega_tape.Tape(blob)                           # "fast": omega-core's lattice,
                                                        #   what the viewer draws
field = omega_tape.Tape(blob, backend="manufacturing")  # hodge's spinodoid_gabor
```

Field drivers (a graded scale or orientation) travel as nodes of the tape and
are read at every kernel centre, as `omega-core`'s bake reads them.

A tape whose spinodoid leaf has no payload still needs its table handed over
before it loads. `tape_oracle_uses` says what a tape needs:

```python
import numpy as np, omega_tape

blob = open("field.tape", "rb").read()
for oid, kind, self_contained, slot in omega_tape.tape_oracle_uses(blob):
    print(f"{oid:#x} {kind} self_contained={self_contained} slot={slot}")
# 0xe05b2000 hodgeSpinodoid self_contained=False slot=0   (no payload)

omega_tape.register_hodge_spinodoid(
    slot=0,
    kernels=K,                      # (cells*impulses, 8) float64, see below
    origin=[-16, -16, -16], size=[32, 32, 32],
    bounds=[0, 0, 0, 0, 0, 0], periodic=True,
    impulses=8, cell_size=7.2644,
    a=a0, sigma=sigma, base_beta=2 * np.pi / period,
    level=bias, local=True,         # the payload's bias is the threshold; 0 is the median
)
field = omega_tape.Tape(blob)       # now it loads
```

`kernels` is row-major over `size` from `origin`, `impulses` rows per cell, each
row `[jitter x, jitter y, jitter z, dir x, dir y, dir z, beta, gamma]`, with a
kernel centre at `(cell + jitter) * cell_size`. hodge's own `_gabor_kernels`
produces exactly these, and `omega-core`'s `HodgeBake` is pinned to it (bit-exact on
the direction pool, 2e-14 on the field —
`program/noise/spinodoid/hodge_parity_test.go`). A `periodic=True` table must be
hodge's 32-cubed window; a bounded one is located by `bounds` and reads as a
conservative distance-to-domain outside it. A tape with a payload ignores this:
its baked table replaces a registered one.

Loading **fails** if any referenced oracle is unregistered, and that is new
behavior worth knowing about. fidget reads an unregistered leaf as zero on
purpose, so a preview can draw before a bake finishes — which on a replay meant
a spinodoid tape quietly evaluating to `min 0.0 max 0.0`, i.e. meshing a
100%-solid block with no error. The check runs at load, not in the evaluator, so
the live preview paths keep their old behavior.

`eval`/`grad` take `(N,3)` **float32** arrays and return numpy arrays — the same
shape hodge's Walk-on-Spheres / Walk-on-Stars solvers use, so a GSL field
drops in as a `FieldFn`. For a true SDF, `eval` is the WoS distance-to-boundary
and its sign (GWN, baked into `bvhSdf`) is the inside/outside test — mesh-free.

## Reading a tape (the hodge entry point)

Reading needs only the built `omega_tape` module — the `.tape` protobuf is
decoded inside the module (prost), so `tape_pb2.py` is NOT required to read one
(it's only for authoring):

```python
import numpy as np, omega_tape

field = omega_tape.Tape(open("field.tape", "rb").read())
P = np.ascontiguousarray(points, dtype=np.float32)   # (N,3)
sdf  = field.eval(P)                                  # (N,)  combined (min of outputs)
g    = field.grad(P)                                  # (N,4): value, ∂x, ∂y, ∂z
```

### Multiple outputs

A program with several `return`s exposes each one individually; the no-arg form
is their `min` (the whole scene). `output_names` matches the UI's shape labels.

```python
field.num_outputs            # e.g. 2
field.output_names           # e.g. ["body", "lid"]  (UI labels; "" if unnamed)
field.eval(P)                # combined field (min over outputs)
field.eval(P, output=0)      # just the "body" field
g = field.grad(P, output=1)  # gradient of "lid"
# interval(...) and mesh(...) take the same optional `output=` argument.
```

### Mapping to hodge's solver primitives

The field is a signed implicit in the GSL source's own world coordinates (no
transform is applied on `eval`). For the Walk-on-Spheres / Walk-on-Stars stack:

- **domain** = `field.bbox` (`[xmin,ymin,zmin,xmax,ymax,zmax]` or `None`) — the
  scene bbox to seed sampling / the WoS domain.
- **distance to boundary** (WoS sphere radius) = `abs(field.eval(P))`
- **inside test** = `field.eval(P) < 0` (sign is GWN-correct for `bvhSdf`)
- **boundary normal / 1-form** = `field.grad(P)[:, 1:4]` (normalize for a unit
  normal); column 0 is the value, so `grad` gives value+normal in one pass.

`field.eval` already matches `FieldFn = Callable[[np.ndarray], Scalar]`, so it
drops in directly as a distance/boundary/source oracle — no mesh required.
Oracle (`bvhSdf`/`vdbSdf`) payloads are embedded in the tape and auto-registered
on load, so even mesh-referencing tapes are read with a bare `Tape(bytes)`, and
so are spinodoid tapes `omega-core` writes, which carry their payload; pass
`backend="manufacturing"` for hodge's field.

### Making it a native hodge feature

`omega_tape` statically embeds `omega-core`'s render lib (fidget + OpenVDB), so the
only artifact hodge needs at runtime is the built `omega_tape.{so,pyd}` on its
path — there is no separate `omega-core` process and no `.tape`-specific C dependency.
To wrap it natively, hodge would:

1. Add a build step that produces the wheel from an `omega-core` checkout
   (`maturin build --release` in `internal/fidget/internal/tape-py`, or
   `cargo build -p omega-tape --release` + place `libomega_tape.dylib` as
   `omega_tape.so`), pinned to an `omega-core` revision.
2. Add a thin `hodge` reader, e.g. `hodge.io.read_omega_tape(path) -> Field`,
   that calls `omega_tape.Tape(...)` and adapts `eval`/`grad`/`bbox` to whatever
   distance/normal interface hodge's solvers expect (per the mapping above).

Mesh, noise and spinodoid oracles need no external asset registry or manual
registration; step 2 chooses the spinodoid's `backend`.

## Notes / follow-ups

- `eval`/`grad` require contiguous **float32** `(N,3)` input (hodge already
  downcasts geometry queries to f32). An f64 path could be added if needed.
- `Tape` is `unsendable` (holds a fidget VM shape) — use it from one thread.
- On the Go side, `backend/tape` (`tape.Compile`) wraps `serialize_tape` to emit
  self-contained `.tape` payloads directly from assembled GSL programs.
