//! `omega_tape` — Python extension module wrapping omega's render lib so hodge
//! can read omega implicit fields in-process, including `bvhSdf` / `vdbSdf`
//! oracles.
//!
//! Flow: omega serializes a GSL-lowered field to a `tape::Tape` protobuf (or
//! Python authors one directly via `tape_pb2`). hodge asks `tape_oracle_uses`
//! what the tape cannot resolve by itself, registers that — a bvhSdf/vdbSdf
//! payload by asset id, a spinodoid kernel table by slot — loads
//! `Tape(tape_bytes)`, then calls `eval` / `grad` / `interval` / `mesh_stl`.
//! All evaluation is in-process against fidget's VM evaluator.
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
use render::tape::{self, MeshAlgorithm, OracleForce, OracleMode, Tape as RawTape};

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
    #[new]
    #[pyo3(signature = (tape, oracles="auto", resolution=None, half_band=None))]
    fn new(
        tape: &[u8],
        oracles: &str,
        resolution: Option<u32>,
        half_band: Option<f32>,
    ) -> PyResult<Self> {
        let force = match oracles.to_ascii_lowercase().as_str() {
            "auto" => OracleForce::Auto,
            "bvh" => OracleForce::Bvh,
            "vdb" => OracleForce::Vdb,
            other => {
                return Err(PyValueError::new_err(format!(
                    "oracles must be 'auto', 'bvh', or 'vdb', got '{other}'"
                )))
            }
        };
        let mode = OracleMode {
            force,
            resolution,
            half_band,
        };
        RawTape::from_bytes_with(tape, mode)
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
/// (a library-baked tile, or a source mesh carried in the tape) and False when
/// this process has to register the oracle first — a spinodoid leaf carries
/// only its id, never its kernel table. `slot` is what to pass as
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
    m.add_function(wrap_pyfunction!(register_mesh_oracle, m)?)?;
    m.add_function(wrap_pyfunction!(register_vdb_oracle, m)?)?;
    m.add_function(wrap_pyfunction!(register_hodge_spinodoid, m)?)?;
    Ok(())
}
