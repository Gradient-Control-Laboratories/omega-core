//! `omega_tape` — Python extension module wrapping omega's render lib so hodge
//! can read omega implicit fields in-process, including `bvhSdf` / `vdbSdf`
//! oracles.
//!
//! Flow: omega serializes a GSL-lowered field to a `tape::Tape` protobuf (or
//! Python authors one directly via `tape_pb2`). The reader loads
//! `Tape(tape_bytes)` — mesh payloads are built and spinodoids baked from the
//! arguments the tape carries — then calls `eval` / `grad` / `interval` /
//! `mesh_stl`. `tape_oracle_uses` says what, if anything, a tape needs
//! registered first. All evaluation is in-process against fidget's VM
//! evaluator.
//!
//! `eval` / `grad` are numpy-native: they take `(N,3)` float32 arrays and
//! return `(N,)` / `(N,4)` — the shape hodge's Walk-on-Spheres / mesh-free
//! Monte-Carlo solvers already use, so an omega tape drops in as a hodge
//! `FieldFn` (distance/boundary/source oracle) with no per-call conversion.

use numpy::{
    IntoPyArray, PyArray1, PyArray2, PyArrayMethods, PyReadonlyArray2, PyUntypedArrayMethods,
};
use pyo3::exceptions::PyValueError;
use pyo3::prelude::*;
use pyo3::types::PyBytes;
use pyo3::types::PyDict;
use render::tape::{
    self, MeshAlgorithm, OracleForce, OracleMode, OracleOverride, SpinodoidBackend, Tape as RawTape,
};
use std::collections::HashMap;

/// A load setting given either as one value (the default for every payload)
/// or as a dict keyed by payload asset id (for those payloads alone). None
/// leaves the default to the caller.
fn per_payload<T>(
    arg: Option<&Bound<'_, PyAny>>,
    parse: impl Fn(&Bound<'_, PyAny>) -> PyResult<T>,
) -> PyResult<(Option<T>, Vec<(u32, T)>)> {
    let Some(arg) = arg.filter(|a| !a.is_none()) else {
        return Ok((None, Vec::new()));
    };
    if let Ok(dict) = arg.downcast::<PyDict>() {
        let mut out = Vec::with_capacity(dict.len());
        for (k, v) in dict.iter() {
            out.push((k.extract::<u32>()?, parse(&v)?));
        }
        return Ok((None, out));
    }
    Ok((Some(parse(arg)?), Vec::new()))
}

/// A built, evaluable implicit field loaded from an omega tape.
#[pyclass(unsendable)]
struct Tape {
    inner: RawTape,
}

#[pymethods]
impl Tape {
    /// Load an evaluable field from serialized `tape::Tape` protobuf bytes.
    ///
    /// `oracles` chooses how embedded mesh oracles are built: "auto" (as the
    /// GSL author wrote them), "bvh" (exact, everywhere), or "vdb" (fast,
    /// near-surface). `resolution` is the target VDB voxels along the longest
    /// bbox axis (higher = finer); `half_band` is the VDB narrow-band half-width
    /// in voxels (how far from the surface distance stays accurate). Both
    /// default to the tape's values or sensible defaults, and are ignored for
    /// BVH.
    ///
    /// `backend` chooses which field a spinodoid is rebuilt as from its
    /// payload's arguments: "fast" (omega's lattice, what the viewer draws) or
    /// "manufacturing" (hodge's jittered-kernel field), whatever the payload's
    /// `fast` flag says.
    ///
    /// Each of the four is either one value, for every payload it applies to,
    /// or a dict keyed by payload asset id (see `tape_oracle_payloads`), for
    /// those payloads alone; a payload a dict leaves out takes the default.
    /// A key must be a payload in the tape, of a kind the setting applies to.
    #[new]
    #[pyo3(signature = (tape, oracles=None, resolution=None, half_band=None, backend=None))]
    fn new(
        tape: &[u8],
        oracles: Option<&Bound<'_, PyAny>>,
        resolution: Option<&Bound<'_, PyAny>>,
        half_band: Option<&Bound<'_, PyAny>>,
        backend: Option<&Bound<'_, PyAny>>,
    ) -> PyResult<Self> {
        let parse_force = |v: &Bound<'_, PyAny>| -> PyResult<OracleForce> {
            let s: String = v.extract()?;
            match s.to_ascii_lowercase().as_str() {
                "auto" => Ok(OracleForce::Auto),
                "bvh" => Ok(OracleForce::Bvh),
                "vdb" => Ok(OracleForce::Vdb),
                other => Err(PyValueError::new_err(format!(
                    "oracles must be 'auto', 'bvh', or 'vdb', got '{other}'"
                ))),
            }
        };
        let parse_backend = |v: &Bound<'_, PyAny>| -> PyResult<SpinodoidBackend> {
            let s: String = v.extract()?;
            match s.to_ascii_lowercase().as_str() {
                "fast" => Ok(SpinodoidBackend::Fast),
                "manufacturing" => Ok(SpinodoidBackend::Manufacturing),
                other => Err(PyValueError::new_err(format!(
                    "backend must be 'fast' or 'manufacturing', got '{other}'"
                ))),
            }
        };
        let (force, forces) = per_payload(oracles, parse_force)?;
        let (resolution, resolutions) = per_payload(resolution, |v| v.extract::<u32>())?;
        let (half_band, half_bands) = per_payload(half_band, |v| v.extract::<f32>())?;
        let (spinodoid, backends) = per_payload(backend, parse_backend)?;

        let mode = OracleMode {
            force: force.unwrap_or(OracleForce::Auto),
            resolution,
            half_band,
            spinodoid: spinodoid.unwrap_or(SpinodoidBackend::Fast),
        };
        let mut overrides: HashMap<u32, OracleOverride> = HashMap::new();
        for (id, v) in forces {
            overrides.entry(id).or_default().force = Some(v);
        }
        for (id, v) in resolutions {
            overrides.entry(id).or_default().resolution = Some(v);
        }
        for (id, v) in half_bands {
            overrides.entry(id).or_default().half_band = Some(v);
        }
        for (id, v) in backends {
            overrides.entry(id).or_default().spinodoid = Some(v);
        }
        RawTape::from_bytes_with_overrides(tape, mode, &overrides)
            .map(|inner| Tape { inner })
            .map_err(PyValueError::new_err)
    }

    /// Number of individual outputs (one per GSL `return`).
    #[getter]
    fn num_outputs(&self) -> usize {
        self.inner.num_outputs()
    }

    /// Per-output names matching the omega UI's shape labels (parallel to the
    /// outputs; entries may be empty for unnamed expression returns).
    #[getter]
    fn output_names(&self) -> Vec<String> {
        self.inner.output_names().to_vec()
    }

    /// Field value at each point. `points` is an `(N,3)` float32 array; returns
    /// an `(N,)` float32 array. `output` selects one output (0-based); omit it
    /// for the combined (min over all outputs) field. Directly usable as a hodge
    /// `FieldFn` — for a true SDF the value is the WoS distance-to-boundary.
    #[pyo3(signature = (points, output=None))]
    fn eval<'py>(
        &self,
        py: Python<'py>,
        points: PyReadonlyArray2<'py, f32>,
        output: Option<usize>,
    ) -> PyResult<Bound<'py, PyArray1<f32>>> {
        let pts = points
            .as_slice()
            .map_err(|e| PyValueError::new_err(format!("points must be C-contiguous (N,3) float32: {e}")))?;
        let vals = self.inner.eval(pts, output).map_err(PyValueError::new_err)?;
        Ok(vals.into_pyarray(py))
    }

    /// Value + gradient at each point. `points` is `(N,3)` float32; returns an
    /// `(N,4)` float32 array `[v, dx, dy, dz]` — the gradient is the field's
    /// (unnormalized) normal / 1-form, what WoSt wants for Neumann/Robin.
    /// `output` selects one output; omit for the combined field.
    #[pyo3(signature = (points, output=None))]
    fn grad<'py>(
        &self,
        py: Python<'py>,
        points: PyReadonlyArray2<'py, f32>,
        output: Option<usize>,
    ) -> PyResult<Bound<'py, PyArray2<f32>>> {
        let n = points.shape()[0];
        let pts = points
            .as_slice()
            .map_err(|e| PyValueError::new_err(format!("points must be C-contiguous (N,3) float32: {e}")))?;
        let flat = self.inner.grad(pts, output).map_err(PyValueError::new_err)?;
        flat.into_pyarray(py)
            .reshape([n, 4])
            .map_err(Into::into)
    }

    /// The world-space scene bounding box recorded in the tape as
    /// `[xmin, ymin, zmin, xmax, ymax, zmax]`, or `None` if the tape carries no
    /// bounds. Use it to seed a sampling / Walk-on-Spheres domain.
    #[getter]
    fn bbox(&self) -> Option<Vec<f32>> {
        self.inner.bbox().map(|b| b.to_vec())
    }

    /// Interval bound over an axis-aligned box `[xmin,ymin,zmin,xmax,ymax,zmax]`;
    /// returns `(lower, upper)`. `output` selects one output; omit for combined.
    #[pyo3(signature = (bbox, output=None))]
    fn interval(&self, bbox: Vec<f32>, output: Option<usize>) -> PyResult<(f32, f32)> {
        self.inner
            .interval(&bbox, output)
            .map_err(PyValueError::new_err)
    }

    /// Mesh the field into an in-memory `Mesh` (vertices + triangles).
    ///
    /// Two explicit axes:
    ///   - **how:** `algorithm` is `"marching_cubes"` (default) or
    ///     `"dual_contour"`.
    ///   - **what:** the returned `Mesh` exposes `.vertices` / `.triangles`
    ///     (numpy) and `.to_stl()`; the caller picks the representation.
    ///
    /// `bbox` is `[xmin,ymin,zmin,xmax,ymax,zmax]`; if omitted it uses the
    /// tape's scene bbox. `depth` is the grid/octree depth. `output` selects one
    /// output, or omit for the combined field.
    #[pyo3(signature = (bbox=None, depth=6, algorithm="marching_cubes", output=None))]
    fn mesh(
        &self,
        bbox: Option<Vec<f32>>,
        depth: i32,
        algorithm: &str,
        output: Option<usize>,
    ) -> PyResult<Mesh> {
        let alg = match algorithm.to_ascii_lowercase().as_str() {
            "marching_cubes" | "mc" => MeshAlgorithm::MarchingCubes,
            "dual_contour" | "dc" => MeshAlgorithm::DualContour,
            other => {
                return Err(PyValueError::new_err(format!(
                    "algorithm must be 'marching_cubes' or 'dual_contour', got '{other}'"
                )))
            }
        };
        let b = match bbox.or_else(|| self.inner.bbox().map(|x| x.to_vec())) {
            Some(b) if b.len() >= 6 => b,
            Some(_) => {
                return Err(PyValueError::new_err(
                    "bbox must have 6 elements [xmin,ymin,zmin,xmax,ymax,zmax]",
                ))
            }
            None => {
                return Err(PyValueError::new_err(
                    "no bbox: pass bbox=[...] or load a tape that carries a scene bbox",
                ))
            }
        };
        let inner = &self.inner;
        let result = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
            inner.mesh(
                alg,
                depth,
                [b[0], b[1], b[2]],
                [b[3], b[4], b[5]],
                output,
            )
        }));
        match result {
            Ok(Ok(md)) => Ok(Mesh { inner: md }),
            Ok(Err(e)) => Err(PyValueError::new_err(e)),
            Err(_) => Err(PyValueError::new_err("mesher panicked")),
        }
    }
}

/// An in-memory triangle mesh produced by `Tape.mesh`. The output format is the
/// caller's choice on this object: numpy `vertices` / `triangles`, or `to_stl`.
#[pyclass]
struct Mesh {
    inner: tape::MeshData,
}

#[pymethods]
impl Mesh {
    /// `(V, 3)` float32 vertex positions.
    #[getter]
    fn vertices<'py>(&self, py: Python<'py>) -> PyResult<Bound<'py, PyArray2<f32>>> {
        let n = self.inner.vertices.len() / 3;
        self.inner
            .vertices
            .clone()
            .into_pyarray(py)
            .reshape([n, 3])
            .map_err(Into::into)
    }

    /// `(T, 3)` uint32 triangle vertex indices.
    #[getter]
    fn triangles<'py>(&self, py: Python<'py>) -> PyResult<Bound<'py, PyArray2<u32>>> {
        let n = self.inner.triangles.len() / 3;
        self.inner
            .triangles
            .clone()
            .into_pyarray(py)
            .reshape([n, 3])
            .map_err(Into::into)
    }

    #[getter]
    fn num_vertices(&self) -> usize {
        self.inner.vertices.len() / 3
    }

    #[getter]
    fn num_triangles(&self) -> usize {
        self.inner.triangles.len() / 3
    }

    /// Binary STL bytes.
    fn to_stl<'py>(&self, py: Python<'py>) -> Bound<'py, PyBytes> {
        PyBytes::new(py, &self.inner.to_stl())
    }

    fn __repr__(&self) -> String {
        format!(
            "<Mesh {} verts, {} tris>",
            self.inner.vertices.len() / 3,
            self.inner.triangles.len() / 3
        )
    }
}

/// What a tape needs bound before `Tape(...)` will load it: one tuple per
/// mesh-oracle id it references, as `(id, kind, self_contained, slot)`.
///
/// `kind` is "noise", "cellular", "spinodoid", "hodgeSpinodoid", "mesh" or
/// "unknown". `self_contained` is True when loading resolves the id by itself
/// (a library-baked tile, a source mesh carried in the tape, or a spinodoid
/// whose payload the tape carries) and False when this process has to
/// register the oracle first. `slot` is what to pass as
/// `register_hodge_spinodoid`'s first argument, where the kind has a slot.
///
/// Read-only, and it registers nothing, so this is the first call to make on a
/// tape written somewhere else:
///
/// ```python
/// for oid, kind, ok, slot in omega_tape.tape_oracle_uses(blob):
///     if not ok:
///         print(f"must bind {oid:#x} ({kind}) slot {slot}")
/// ```
#[pyfunction]
fn tape_oracle_uses(tape: &[u8]) -> PyResult<Vec<tape::OracleUseRow>> {
    tape::tape_oracle_uses(tape).map_err(PyValueError::new_err)
}

/// The oracle payloads a tape carries, as `(asset_id, kind)` with `kind`
/// "bvh", "vdb" (a mesh, as the GSL author wrote it) or "spinodoid": the keys
/// `Tape(...)`'s per-payload settings take.
///
/// ```python
/// for asset_id, kind in omega_tape.tape_oracle_payloads(blob):
///     print(f"{asset_id:#x} {kind}")
/// field = omega_tape.Tape(blob, backend={0xE05B1000: "manufacturing"})
/// ```
#[pyfunction]
fn tape_oracle_payloads(tape: &[u8]) -> PyResult<Vec<(u32, String)>> {
    tape::tape_oracle_payloads(tape).map_err(PyValueError::new_err)
}

/// Install a baked hodge-spinodoid kernel table in `slot`, so a tape whose
/// leaves name that slot will load.
///
/// The table is the caller's: hodge's own `_gabor_kernels` produces exactly
/// these kernels, and omega's `HodgeBake` is pinned to it bit-for-bit on the
/// direction pool and to 2e-14 on the field, so handing the table over is
/// handing over the field hodge computes. Nothing is re-derived here.
///
/// `kernels` is `(cells * impulses, 8)` float64 in row-major cell order over
/// `size` from `origin`, each row `[jitter x, jitter y, jitter z, dir x,
/// dir y, dir z, beta, gamma]` — a centre is `(cell + jitter) * cell_size`.
/// `periodic` tables must be hodge's 32-cubed window and ignore `bounds`; a
/// bounded table is located by `bounds` and reads as a conservative
/// distance-to-domain outside it. `level` is the field threshold, the
/// payload's `bias` as it stands, and `local` divides by the
/// wavenumber carried at each point rather than the base one.
#[pyfunction]
#[allow(clippy::too_many_arguments)]
#[pyo3(signature = (slot, kernels, origin, size, bounds, periodic, impulses,
                    cell_size, a, sigma, base_beta, level, local))]
fn register_hodge_spinodoid(
    slot: u32,
    kernels: PyReadonlyArray2<f64>,
    origin: [i64; 3],
    size: [u32; 3],
    bounds: [f64; 6],
    periodic: bool,
    impulses: usize,
    cell_size: f64,
    a: f64,
    sigma: f64,
    base_beta: f64,
    level: f64,
    local: bool,
) -> PyResult<()> {
    let shape = kernels.shape();
    if shape.len() != 2 || shape[1] != 8 {
        return Err(PyValueError::new_err(format!(
            "kernels must be (N,8) float64 — [jitter xyz, dir xyz, beta, gamma] — got {:?}",
            shape
        )));
    }
    let k = kernels.as_slice().map_err(|e| {
        PyValueError::new_err(format!("kernels must be C-contiguous (N,8) float64: {e}"))
    })?;
    tape::register_hodge_spinodoid(
        slot, k, origin, size, bounds, periodic, impulses, cell_size, a, sigma, base_beta, level,
        local,
    )
    .map_err(PyValueError::new_err)
}

/// Register a bvhSdf mesh oracle payload by asset id. `vertices` is `(M,3)`
/// float32, `triangles` is `(K,3)` uint32. Must be called before loading a
/// `Tape` whose tape references this asset id.
#[pyfunction]
fn register_mesh_oracle(
    asset_id: u32,
    vertices: PyReadonlyArray2<f32>,
    triangles: PyReadonlyArray2<u32>,
) -> PyResult<i32> {
    let v = vertices
        .as_slice()
        .map_err(|e| PyValueError::new_err(format!("vertices must be C-contiguous (M,3) float32: {e}")))?;
    let t = triangles
        .as_slice()
        .map_err(|e| PyValueError::new_err(format!("triangles must be C-contiguous (K,3) uint32: {e}")))?;
    Ok(tape::register_mesh_oracle(asset_id, v, t))
}

/// Register a vdbSdf mesh oracle payload by asset id (rasterized to a
/// narrow-band VDB level set). `vertices` `(M,3)` float32, `triangles` `(K,3)`
/// uint32.
#[pyfunction]
#[pyo3(signature = (asset_id, vertices, triangles, voxel_size = 0.05, band_width_voxels = 3.0))]
fn register_vdb_oracle(
    asset_id: u32,
    vertices: PyReadonlyArray2<f32>,
    triangles: PyReadonlyArray2<u32>,
    voxel_size: f32,
    band_width_voxels: f32,
) -> PyResult<i32> {
    let v = vertices
        .as_slice()
        .map_err(|e| PyValueError::new_err(format!("vertices must be C-contiguous (M,3) float32: {e}")))?;
    let t = triangles
        .as_slice()
        .map_err(|e| PyValueError::new_err(format!("triangles must be C-contiguous (K,3) uint32: {e}")))?;
    Ok(tape::register_vdb_oracle(asset_id, v, t, voxel_size, band_width_voxels))
}

#[pymodule]
fn omega_tape(m: &Bound<'_, PyModule>) -> PyResult<()> {
    m.add_class::<Tape>()?;
    m.add_class::<Mesh>()?;
    m.add_function(wrap_pyfunction!(tape_oracle_uses, m)?)?;
    m.add_function(wrap_pyfunction!(tape_oracle_payloads, m)?)?;
    m.add_function(wrap_pyfunction!(register_mesh_oracle, m)?)?;
    m.add_function(wrap_pyfunction!(register_vdb_oracle, m)?)?;
    m.add_function(wrap_pyfunction!(register_hodge_spinodoid, m)?)?;
    Ok(())
}
