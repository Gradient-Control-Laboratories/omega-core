//! Omega tape (de)serialization and in-process evaluation.
//!
//! The Go side lowers GSL to a fidget `Tree` DAG (the global `TREES` vec built
//! by the `tree_*` FFI in `lib.rs`). `serialize_tape` walks that DAG and emits
//! a `tape::Tape` protobuf — the post-lowering "omega tape" — which any consumer
//! (notably the Python-wrapped render lib used by hodge) replays via
//! `Tape::from_bytes` to get a handle it can `eval` / `grad` / `interval` /
//! `mesh_stl` in-process. Oracle payloads (bvhSdf/vdbSdf) travel out-of-band and
//! are bound by asset id via `register_mesh_oracle` / `register_vdb_oracle`
//! before replay; a spinodoid is baked on load from the arguments its payload
//! carries, as the backend `OracleMode::spinodoid` chooses (a tape without the
//! payload binds a table by slot via `register_hodge_spinodoid`).
//! `tape_oracle_uses` lists what a given tape needs, and loading fails rather
//! than reading an unbound oracle as zero.
//!
//! Two layers live here:
//!   * a safe Rust API (`serialize_outputs`, `Tape`, the `register_*` helpers)
//!     used directly by the pyo3 wrapper crate, and
//!   * a thin C-ABI mirror (`serialize_tape`, `tape_build`, `tape_eval`, …) for
//!     the Go cgo path, which just marshals pointers onto the safe API.
//!
//! The tape mirrors fidget's `TreeOp` one-to-one. `Tree::pow` expands to
//! square/mul at construction, so there is no pow node. `RemapAffine` is not
//! produced by the omega lowering (which only calls `remap_xyz`), so it is
//! rejected explicitly rather than guessed.

use crate::tape_proto;
use crate::{build_render_error, build_render_result, RenderResult};
use core::ffi::c_ulong;
use fidget::context::{BinaryOpcode, Tree, TreeOp, UnaryOpcode};
use fidget::shape::{ShapeBulkEval, ShapeTracingEval, ShapeVars};
use fidget::types::{Grad, Interval};
use fidget::var::Var;
use fidget::vm::{VmFloatSliceEval, VmGradSliceEval, VmIntervalEval, VmShape};
use fidget_mesh::{Octree, Settings as MeshSettings};
use prost::Message;
use std::collections::HashMap;
use std::os::raw::{c_float, c_int};
use std::sync::Arc;

const N: usize = u8::MAX as usize;

// Defaults for building a VDB level set from an embedded mesh on load. These
// mirror GSL's vdbSdf (resolveVdbParams): voxel = longest_bbox_axis / divisor,
// default divisor 256, default half-band 3 — so a `resolution` here means the
// same thing as the divisor in `vdbSdf(index, divisor, halfBand)`.
const DEFAULT_VDB_RESOLUTION: u32 = 256; // voxels along the longest bbox axis
const DEFAULT_HALF_BAND_VOXELS: f32 = 3.0;
const MIN_VOXEL: f32 = 1e-6;

/// Which acceleration structure to build for embedded mesh-oracle payloads on
/// load. The tape always carries the source mesh, so the consumer is free to
/// choose: exact `Bvh`, near-surface `Vdb`, or `Auto` (the kind the GSL author
/// wrote — bvhSdf -> Bvh, vdbSdf -> Vdb).
#[derive(Clone, Copy, PartialEq)]
pub enum OracleForce {
    Auto,
    Bvh,
    Vdb,
}

/// Which spinodoid a payload is rebuilt as on load. The tape carries the
/// field's arguments, and the same arguments describe two fields: `Fast` is
/// omega's lattice (what the viewer draws), `Manufacturing` hodge's
/// jittered-kernel field.
#[derive(Clone, Copy, PartialEq)]
pub enum SpinodoidBackend {
    Fast,
    Manufacturing,
}

/// Load-time policy for turning embedded oracle payloads into oracles.
#[derive(Clone, Copy)]
pub struct OracleMode {
    pub force: OracleForce,
    /// Target VDB voxels along the longest bbox axis (higher = finer, smaller
    /// near-surface band). None uses the tape's authored voxel size, or a
    /// default. Ignored when the chosen structure is BVH.
    pub resolution: Option<u32>,
    /// VDB narrow-band half-width in voxels — how far from the surface distance
    /// stays accurate (band reach in world units = voxel_size * half_band).
    /// None uses the tape's value, or a default of 3. Ignored for BVH.
    pub half_band: Option<f32>,
    /// Which field a spinodoid payload is baked as, whatever its `fast` flag.
    pub spinodoid: SpinodoidBackend,
}

impl Default for OracleMode {
    fn default() -> Self {
        OracleMode {
            force: OracleForce::Auto,
            resolution: None,
            half_band: None,
            spinodoid: SpinodoidBackend::Fast,
        }
    }
}

/// One payload's departures from the tape-wide [`OracleMode`]: each set field
/// replaces the mode's for the payload whose asset id it is keyed by. A tape
/// may carry several meshes and several spinodoids, and each can be read its
/// own way.
#[derive(Clone, Copy, Default)]
pub struct OracleOverride {
    pub force: Option<OracleForce>,
    pub resolution: Option<u32>,
    pub half_band: Option<f32>,
    pub spinodoid: Option<SpinodoidBackend>,
}

impl OracleMode {
    fn with(self, o: Option<&OracleOverride>) -> OracleMode {
        let Some(o) = o else { return self };
        OracleMode {
            force: o.force.unwrap_or(self.force),
            resolution: o.resolution.or(self.resolution),
            half_band: o.half_band.or(self.half_band),
            spinodoid: o.spinodoid.unwrap_or(self.spinodoid),
        }
    }
}

// Every payload a Tape loads is bound under ids of its own, taken from here,
// so two tapes in one process never read each other's oracles even where
// their tapes share asset ids. The block sits below the vdbSdf bit and clear
// of the library-baked and spinodoid blocks (0xE0...), and above the small
// asset ids omega writes.
const LOADED_ORACLE_ID_FIRST: u32 = 0x4000_0000;
const LOADED_ORACLE_ID_END: u32 = 0x7FFF_FFFF;
static NEXT_LOADED_ORACLE_ID: std::sync::atomic::AtomicU32 =
    std::sync::atomic::AtomicU32::new(LOADED_ORACLE_ID_FIRST);

fn fresh_oracle_ids(n: u32) -> Result<u32, String> {
    let first = NEXT_LOADED_ORACLE_ID.fetch_add(n, std::sync::atomic::Ordering::Relaxed);
    if first > LOADED_ORACLE_ID_END - n {
        return Err("this process has loaded too many oracles to give a new tape ids of its own".into());
    }
    Ok(first)
}

/// What a released id reads as: nothing should evaluate it once its Tape is
/// gone, and a NaN says so if something does.
struct ReleasedOracle;

impl fidget::mesh_oracle::MeshOracle for ReleasedOracle {
    fn point(&self, _x: f32, _y: f32, _z: f32) -> f32 {
        f32::NAN
    }

    fn interval(&self, _x: Interval, _y: Interval, _z: Interval) -> Interval {
        Interval::new(f32::NEG_INFINITY, f32::INFINITY)
    }
}

/// The registry ids a Tape bound, released when it is dropped (or when a load
/// fails part way). fidget's registry has no removal, so releasing replaces
/// each oracle with a ReleasedOracle: the mesh or table it held is freed, and
/// a few bytes per id remain.
struct OwnedOracles(Vec<u32>);

impl Drop for OwnedOracles {
    fn drop(&mut self) {
        for &id in &self.0 {
            fidget::mesh_oracle::register(id, Arc::new(ReleasedOracle));
        }
    }
}

// Longest bbox axis — the basis for voxel sizing, matching GSL's
// resolveVdbParams (voxel = longest_axis / divisor).
fn oracle_longest_axis(o: &tape_proto::OraclePayload) -> f32 {
    let (mn, mx) = if o.bbox.len() == 6 {
        ([o.bbox[0], o.bbox[1], o.bbox[2]], [o.bbox[3], o.bbox[4], o.bbox[5]])
    } else {
        let mut mn = [f32::INFINITY; 3];
        let mut mx = [f32::NEG_INFINITY; 3];
        for c in o.vertices.chunks_exact(3) {
            for a in 0..3 {
                mn[a] = mn[a].min(c[a]);
                mx[a] = mx[a].max(c[a]);
            }
        }
        (mn, mx)
    };
    (mx[0] - mn[0]).max(mx[1] - mn[1]).max(mx[2] - mn[2])
}

fn resolve_voxel(o: &tape_proto::OraclePayload, resolution: Option<u32>) -> f32 {
    if let Some(res) = resolution {
        return (oracle_longest_axis(o) / res.max(1) as f32).max(MIN_VOXEL);
    }
    if o.voxel_size > 0.0 {
        return o.voxel_size;
    }
    (oracle_longest_axis(o) / DEFAULT_VDB_RESOLUTION as f32).max(MIN_VOXEL)
}

/// Register one embedded mesh payload as the structure `mode` selects, under
/// `asset_id` as its tree leaf uses it (vdbSdf leaves carry the high bit).
/// Returns the registry key.
fn register_oracle(o: &tape_proto::OraclePayload, asset_id: u32, mode: OracleMode) -> Result<u32, String> {
    let authored_vdb = o.kind == tape_proto::OracleKind::Vdb as i32;
    let key = if authored_vdb {
        asset_id | crate::vdb_oracle::VDB_ASSET_BIT
    } else {
        asset_id
    };
    let want_vdb = match mode.force {
        OracleForce::Auto => authored_vdb,
        OracleForce::Bvh => false,
        OracleForce::Vdb => true,
    };
    if want_vdb {
        let voxel = resolve_voxel(o, mode.resolution);
        let band = mode
            .half_band
            .filter(|b| *b > 0.0)
            .or_else(|| (o.half_band_voxels > 0.0).then_some(o.half_band_voxels))
            .unwrap_or(DEFAULT_HALF_BAND_VOXELS);
        let oracle = crate::vdb_oracle::VdbOracle::from_triangles(&o.vertices, &o.triangles, voxel, band);
        fidget::mesh_oracle::register(key, Arc::new(oracle));
    } else {
        let verts: Vec<[f32; 3]> = o
            .vertices
            .chunks_exact(3)
            .map(|c| [c[0], c[1], c[2]])
            .collect();
        let tris: Vec<[u32; 3]> = o
            .triangles
            .chunks_exact(3)
            .map(|c| [c[0], c[1], c[2]])
            .collect();
        let oracle = crate::mesh_oracle::TriBspOracle::new(verts, tris);
        fidget::mesh_oracle::register(key, Arc::new(oracle));
    }
    Ok(key)
}

/// The values of the tape's driver nodes at `centers`, one Vec per node: the
/// tape's own trees, evaluated in f32 as omega's bake reads them.
fn eval_driver_nodes(
    tape: &tape_proto::Tape,
    nodes: &[u32],
    centers: &[[f64; 3]],
) -> Result<Vec<Vec<f64>>, String> {
    // Replay just the driver trees, as the outputs of a copy of this tape that
    // carries no oracles: every oracle it could reach is already bound.
    let sub = tape_proto::Tape {
        outputs: nodes.to_vec(),
        oracles: Vec::new(),
        output_names: Vec::new(),
        appearance: Vec::new(),
        ..tape.clone()
    };
    let d = tape_core::decode(&sub, &mut |_| Ok(()))?;
    let n = centers.len();
    let xs: Vec<f32> = centers.iter().map(|c| c[0] as f32).collect();
    let ys: Vec<f32> = centers.iter().map(|c| c[1] as f32).collect();
    let zs: Vec<f32> = centers.iter().map(|c| c[2] as f32).collect();
    let mut out = Vec::with_capacity(nodes.len());
    for shape in &d.outputs {
        let t = shape.float_slice_tape(Default::default());
        let mut eval = ShapeBulkEval::<VmFloatSliceEval<N>>::default();
        let values = eval
            .eval_with_vars(&t, &xs, &ys, &zs, &d.shape_vars)
            .map_err(|e| format!("spinodoid driver: {:?}", e))?;
        out.push(values[..n].iter().map(|&v| v as f64).collect());
    }
    Ok(out)
}

/// Whether `id` is a spinodoid slot's distance leaf, lattice or hodge block:
/// the only id a SPINODOID payload may be keyed by.
fn is_spinodoid_distance_id(id: u32) -> bool {
    use tape_core::spinodoid::{component, hodge};
    let comp = tape_core::spinodoid::leaf_of_id(id)
        .or_else(|| hodge::leaf_of_id(id))
        .map(|(_, comp)| comp);
    comp == Some(component::DISTANCE)
}

/// Bake one SPINODOID payload as `backend` and bind it under the four leaf ids
/// from `first_id`. The payload is everything the field is a function of
/// (tape.proto's SpinodoidParams), so this is the same table omega's Go baker
/// builds from the same arguments; tape-core's bake tests pin the two to each
/// other. `tape` is the tape the payload's driver nodes index into, with every
/// oracle they could read already bound.
fn register_spinodoid(
    tape: &tape_proto::Tape,
    o: &tape_proto::OraclePayload,
    first_id: u32,
    backend: SpinodoidBackend,
) -> Result<(), String> {
    use tape_core::spinodoid::{bake, hodge};

    let s = o
        .spinodoid
        .as_ref()
        .ok_or_else(|| format!("spinodoid oracle {:#x} has no parameters", o.asset_id))?;
    #[allow(deprecated)]
    if s.level_bias != 0.0 {
        return Err(format!(
            "spinodoid oracle {:#x} carries the retired quantile levelBias ({}) rather than a bias; regenerate the tape",
            o.asset_id, s.level_bias
        ));
    }
    if s.poles.len() % 3 != 0 {
        return Err(format!("spinodoid oracle {:#x}: poles are not xyz triples", o.asset_id));
    }

    let constant_scale = match (s.scale_nodes.len(), s.scale.first()) {
        (0, Some(&v)) => Some(v),
        (0, None) => {
            return Err(format!(
                "spinodoid oracle {:#x} has neither scale nor scaleNodes",
                o.asset_id
            ))
        }
        _ => None,
    };
    let constant_rotation = match (s.orientation_nodes.len(), s.orientation.len()) {
        (0, 9) => Some([
            [s.orientation[0], s.orientation[1], s.orientation[2]],
            [s.orientation[3], s.orientation[4], s.orientation[5]],
            [s.orientation[6], s.orientation[7], s.orientation[8]],
        ]),
        (9, _) => None,
        _ => {
            return Err(format!(
                "spinodoid oracle {:#x} needs 9 orientation entries or 9 orientationNodes",
                o.asset_id
            ))
        }
    };

    // The drivers as the bake takes them, mirroring omega's converter: a
    // constant 1 or identity is no driver at all.
    let scale_field = |c: &[[f64; 3]]| -> Result<Vec<f64>, String> {
        Ok(eval_driver_nodes(tape, &s.scale_nodes, c)?.remove(0))
    };
    let scale_constant = |c: &[[f64; 3]]| -> Result<Vec<f64>, String> {
        Ok(vec![constant_scale.unwrap_or(1.0); c.len()])
    };
    let rotation_field = |c: &[[f64; 3]]| -> Result<Vec<[[f64; 3]; 3]>, String> {
        let v = eval_driver_nodes(tape, &s.orientation_nodes, c)?;
        Ok((0..c.len())
            .map(|k| [[v[0][k], v[1][k], v[2][k]], [v[3][k], v[4][k], v[5][k]], [v[6][k], v[7][k], v[8][k]]])
            .collect())
    };
    let rotation_constant = |c: &[[f64; 3]]| -> Result<Vec<[[f64; 3]; 3]>, String> {
        Ok(vec![constant_rotation.unwrap_or_default(); c.len()])
    };
    let identity = [[1.0, 0.0, 0.0], [0.0, 1.0, 0.0], [0.0, 0.0, 1.0]];
    let scale: Option<bake::ScaleDriver> = match constant_scale {
        None => Some(&scale_field),
        Some(v) if v != 1.0 => Some(&scale_constant),
        Some(_) => None,
    };
    let orientation: Option<bake::OrientationDriver> = match constant_rotation {
        None => Some(&rotation_field),
        Some(r) if r != identity => Some(&rotation_constant),
        Some(_) => None,
    };

    let p = bake::Params {
        theta: s.theta_radians.clone(),
        poles: s.poles.chunks_exact(3).map(|c| [c[0], c[1], c[2]]).collect(),
        period: s.period,
        packet_width: s.packet_width,
        impulses: s.impulses as usize,
        pool: s.pool as usize,
        seed: s.seed,
        scale,
        orientation,
        domain: (s.domain.len() == 6).then(|| [s.domain[0], s.domain[1], s.domain[2], s.domain[3], s.domain[4], s.domain[5]]),
        periodic: s.periodic,
    };
    let what = |e: String| format!("spinodoid oracle {:#x}: {}", o.asset_id, e);
    match backend {
        SpinodoidBackend::Fast => {
            let table = bake::lattice(&p, s.bias, s.local_normalize).map_err(what)?;
            tape_core::spinodoid::register_leaves(first_id, table);
        }
        SpinodoidBackend::Manufacturing => {
            let table = bake::manufacturing(&p, s.bias, s.local_normalize).map_err(what)?;
            hodge::register_leaves(first_id, table);
        }
    }
    Ok(())
}

// ============================ serialize ============================

struct TapeBuilder {
    nodes: Vec<tape_proto::Node>,
    // dedup by Arc identity so shared subgraphs are emitted once
    seen: HashMap<*const TreeOp, u32>,
    named: HashMap<Var, (String, f64)>,
    x_var: Option<Var>,
    y_var: Option<Var>,
    z_var: Option<Var>,
}

impl TapeBuilder {
    fn push(&mut self, ptr: *const TreeOp, n: tape_proto::node::N) -> u32 {
        let i = self.nodes.len() as u32;
        self.nodes.push(tape_proto::Node { n: Some(n) });
        self.seen.insert(ptr, i);
        i
    }

    fn visit(&mut self, ptr: *const TreeOp, op: &TreeOp) -> Result<u32, String> {
        if let Some(&i) = self.seen.get(&ptr) {
            return Ok(i);
        }
        use tape_proto::node::N;
        let n = match op {
            TreeOp::Const(c) => N::Constant(f64::from(*c)),
            TreeOp::Input(var) => {
                if Some(*var) == self.x_var {
                    N::Axis(tape_proto::Axis::X as i32)
                } else if Some(*var) == self.y_var {
                    N::Axis(tape_proto::Axis::Y as i32)
                } else if Some(*var) == self.z_var {
                    N::Axis(tape_proto::Axis::Z as i32)
                } else {
                    let (name, default) = self
                        .named
                        .get(var)
                        .cloned()
                        .unwrap_or_else(|| (String::new(), 0.0));
                    N::Input(tape_proto::Input { name, default })
                }
            }
            TreeOp::Unary(opcode, arg) => {
                let a = self.visit(Arc::as_ptr(arg), arg)?;
                N::Unary(tape_proto::Unary {
                    op: map_unary(*opcode)? as i32,
                    arg: a,
                })
            }
            TreeOp::Binary(opcode, lhs, rhs) => {
                let l = self.visit(Arc::as_ptr(lhs), lhs)?;
                let r = self.visit(Arc::as_ptr(rhs), rhs)?;
                N::Binary(tape_proto::Binary {
                    op: map_binary(*opcode)? as i32,
                    lhs: l,
                    rhs: r,
                })
            }
            TreeOp::MeshOracle(id) => N::MeshOracle(*id),
            TreeOp::RemapAxes { target, x, y, z } => {
                let t = self.visit(Arc::as_ptr(target), target)?;
                let xi = self.visit(Arc::as_ptr(x), x)?;
                let yi = self.visit(Arc::as_ptr(y), y)?;
                let zi = self.visit(Arc::as_ptr(z), z)?;
                N::Remap(tape_proto::Remap {
                    target: t,
                    x: xi,
                    y: yi,
                    z: zi,
                })
            }
            TreeOp::RemapAffine { .. } => {
                return Err("RemapAffine is not representable in an omega tape".into());
            }
        };
        Ok(self.push(ptr, n))
    }
}

fn map_unary(op: UnaryOpcode) -> Result<tape_proto::UnaryOp, String> {
    use tape_proto::UnaryOp as U;
    Ok(match op {
        UnaryOpcode::Neg => U::Neg,
        UnaryOpcode::Abs => U::Abs,
        UnaryOpcode::Recip => U::Recip,
        UnaryOpcode::Sqrt => U::Sqrt,
        UnaryOpcode::Square => U::Square,
        UnaryOpcode::Floor => U::Floor,
        UnaryOpcode::Ceil => U::Ceil,
        UnaryOpcode::Round => U::Round,
        UnaryOpcode::Sin => U::Sin,
        UnaryOpcode::Cos => U::Cos,
        UnaryOpcode::Tan => U::Tan,
        UnaryOpcode::Asin => U::Asin,
        UnaryOpcode::Acos => U::Acos,
        UnaryOpcode::Atan => U::Atan,
        UnaryOpcode::Exp => U::Exp,
        UnaryOpcode::Ln => U::Ln,
        UnaryOpcode::Not => U::Not,
        UnaryOpcode::Rand => return Err("tape has no rand op".to_string()),
    })
}

fn map_binary(op: BinaryOpcode) -> Result<tape_proto::BinaryOp, String> {
    use tape_proto::BinaryOp as B;
    Ok(match op {
        BinaryOpcode::Add => B::Add,
        BinaryOpcode::Sub => B::Sub,
        BinaryOpcode::Mul => B::Mul,
        BinaryOpcode::Div => B::Div,
        BinaryOpcode::Atan => B::Atan2,
        BinaryOpcode::Min => B::Min,
        BinaryOpcode::Max => B::Max,
        BinaryOpcode::Compare => B::Compare,
        BinaryOpcode::Mod => B::Mod,
        BinaryOpcode::And => B::And,
        BinaryOpcode::Or => B::Or,
        BinaryOpcode::Mix => return Err("tape has no mix op".to_string()),
    })
}

/// One output's colour and material trees, for `serialize_outputs_with_appearance`.
pub struct AppearanceTrees {
    /// sRGB [r, g, b] tree indices, when the output has a colour.
    pub color: Option<[usize; 3]>,
    /// The 12 material tree indices, rows flattened, when it has a material.
    pub material: Option<[usize; 12]>,
    /// The material's displacement is not a constant zero.
    pub displaced: bool,
}

/// Walk the global `TREES` DAG for the given output tree indices and return the
/// serialized `tape::Tape` protobuf bytes.
pub fn serialize_outputs(output_indices: &[usize]) -> Result<Vec<u8>, String> {
    serialize_outputs_with_appearance(output_indices, &[])
}

/// As `serialize_outputs`, with each output's colour and material trees in the
/// tape's `appearance` (parallel to the outputs; may be empty), sharing the
/// node DAG with the fields.
pub fn serialize_outputs_with_appearance(
    output_indices: &[usize],
    appearance: &[AppearanceTrees],
) -> Result<Vec<u8>, String> {
    let trees = crate::TREES.lock().unwrap_or_else(std::sync::PoisonError::into_inner);
    let input_names = crate::INPUT_NAMES.lock().unwrap_or_else(std::sync::PoisonError::into_inner);
    let input_values = crate::INPUT_VALUES.lock().unwrap_or_else(std::sync::PoisonError::into_inner);

    // Reverse-map each named input's Var to its (name, value), mirroring the
    // GLSL codegen path. X/Y/Z are special vars handled separately.
    let mut named: HashMap<Var, (String, f64)> = HashMap::new();
    for (idx, name) in input_names.iter() {
        if let Some(var) = trees[*idx].var() {
            let value = input_values.get(idx).copied().unwrap_or(0.0);
            named.insert(var, (name.clone(), value));
        }
    }

    let mut builder = TapeBuilder {
        nodes: Vec::new(),
        seen: HashMap::new(),
        named,
        x_var: Tree::x().var(),
        y_var: Tree::y().var(),
        z_var: Tree::z().var(),
    };

    let mut outputs = Vec::with_capacity(output_indices.len());
    for &oi in output_indices {
        if oi >= trees.len() {
            return Err(format!("output tree index {} out of range", oi));
        }
        let tree = &trees[oi];
        let idx = builder.visit(tree.as_ptr(), tree)?;
        outputs.push(idx);
    }

    let mut visit_all = |ids: &[usize]| -> Result<Vec<u32>, String> {
        ids.iter()
            .map(|&i| {
                let tree = trees.get(i).ok_or_else(|| format!("appearance tree index {} out of range", i))?;
                builder.visit(tree.as_ptr(), tree)
            })
            .collect()
    };
    let mut appearances = Vec::with_capacity(appearance.len());
    for a in appearance {
        appearances.push(tape_proto::Appearance {
            color: match &a.color {
                Some(c) => visit_all(c)?,
                None => Vec::new(),
            },
            material: match &a.material {
                Some(m) => visit_all(m)?,
                None => Vec::new(),
            },
            displaced: a.displaced,
        });
    }

    let tape = tape_proto::Tape {
        version: 1,
        nodes: builder.nodes,
        outputs,
        // bbox, oracle payloads, and output names are metadata the Go agent
        // stamps from the export request; the tree walk produces none of them.
        // noiseHash3 leaves need nothing extra — the tile is library-baked.
        bbox: Vec::new(),
        oracles: Vec::new(),
        output_names: Vec::new(),
        appearance: appearances,
    };
    Ok(tape.encode_to_vec())
}

// ============================ replay / safe API ============================

/// A built, evaluable implicit field loaded from an omega tape.
///
/// A tape can have multiple outputs (one per GSL `return`). `combined` is their
/// `min` (the whole-scene field, matching the mesher); `outputs` holds each one
/// individually, with `names` giving the UI label per output.
pub struct Tape {
    combined: VmShape,
    outputs: Vec<VmShape>,
    names: Vec<String>,
    shape_vars: ShapeVars<f32>,
    // Optional world-space scene bbox [xmin,ymin,zmin,xmax,ymax,zmax] carried as
    // tape metadata (None when the tape didn't record one).
    bbox: Option<[f32; 6]>,
    // The registry ids this tape's payloads are bound under, released with it.
    // Last, so the shapes that read them are dropped first.
    _oracles: OwnedOracles,
}

/// Check that every override names a payload in `tape`, and only sets what
/// that payload's kind reads.
fn check_overrides(tape: &tape_proto::Tape, overrides: &HashMap<u32, OracleOverride>) -> Result<(), String> {
    for (&id, ov) in overrides {
        let Some(o) = tape.oracles.iter().find(|o| o.asset_id == id) else {
            return Err(format!("no oracle payload in this tape has asset id {:#x}", id));
        };
        let spinodoid = o.kind == tape_proto::OracleKind::Spinodoid as i32;
        if spinodoid && (ov.force.is_some() || ov.resolution.is_some() || ov.half_band.is_some()) {
            return Err(format!(
                "oracle {:#x} is a spinodoid: oracles / resolution / half_band are mesh settings",
                id
            ));
        }
        if !spinodoid && ov.spinodoid.is_some() {
            return Err(format!("oracle {:#x} is a mesh: backend is a spinodoid setting", id));
        }
    }
    Ok(())
}

fn build_tape(
    tape: &tape_proto::Tape,
    mode: OracleMode,
    overrides: &HashMap<u32, OracleOverride>,
) -> Result<Tape, String> {
    // The replay is tape-core's, shared with the browser build; only the
    // oracle kinds are render's: each payload's mode (the tape-wide one with
    // its override, if any) chooses whether an embedded source mesh becomes an
    // exact BVH or a VDB level set (at an optional resolution), and which field
    // a spinodoid payload is baked as.
    check_overrides(tape, overrides)?;

    // Bind every payload under fresh ids, and replay a copy of the tape whose
    // leaves name them: the registry is process-wide, and a payload read one
    // way here must not be read that way by another tape, or replace what
    // another tape bound. A leaf with no payload keeps its id; its oracle was
    // registered by hand, for every tape that names it.
    let mut rekeyed = tape.clone();
    let mut owned = OwnedOracles(Vec::new());
    let mut leaf_ids: HashMap<u32, u32> = HashMap::new();
    let mut first_of: Vec<Option<u32>> = vec![None; tape.oracles.len()];
    for (i, o) in tape.oracles.iter().enumerate() {
        if tape_proto::OracleKind::try_from(o.kind).is_err() {
            return Err(format!("oracle {}: unknown kind {}", i, o.kind));
        }
        let spinodoid = o.kind == tape_proto::OracleKind::Spinodoid as i32;
        if spinodoid && !is_spinodoid_distance_id(o.asset_id) {
            return Err(format!(
                "oracle {}: spinodoid oracle {:#x} is not a slot's distance leaf",
                i, o.asset_id
            ));
        }
        // A tape may carry one asset's payload twice; it binds once.
        if leaf_ids.contains_key(&o.asset_id) {
            continue;
        }
        if spinodoid {
            let n = tape_core::spinodoid::component::COUNT;
            let first = fresh_oracle_ids(n)?;
            for c in 0..n {
                leaf_ids.insert(o.asset_id + c, first + c);
            }
            first_of[i] = Some(first);
        } else {
            let first = fresh_oracle_ids(1)?;
            leaf_ids.insert(o.asset_id, first);
            leaf_ids.insert(
                o.asset_id | crate::vdb_oracle::VDB_ASSET_BIT,
                first | crate::vdb_oracle::VDB_ASSET_BIT,
            );
            first_of[i] = Some(first);
        }
    }
    for node in &mut rekeyed.nodes {
        if let Some(tape_proto::node::N::MeshOracle(id)) = &mut node.n {
            if let Some(&new) = leaf_ids.get(id) {
                *id = new;
            }
        }
    }

    // Every oracle is bound before the replay. Meshes go first: a spinodoid's
    // driver trees may read them while it bakes.
    for (i, o) in tape.oracles.iter().enumerate() {
        let Some(first) = first_of[i] else { continue };
        if o.kind == tape_proto::OracleKind::Spinodoid as i32 {
            continue;
        }
        let key = register_oracle(o, first, mode.with(overrides.get(&o.asset_id)))
            .map_err(|e| format!("oracle {}: {}", i, e))?;
        owned.0.push(key);
    }
    for (i, o) in tape.oracles.iter().enumerate() {
        let Some(first) = first_of[i] else { continue };
        if o.kind != tape_proto::OracleKind::Spinodoid as i32 {
            continue;
        }
        let backend = mode.with(overrides.get(&o.asset_id)).spinodoid;
        // Bound before baking, so a failed bake releases what it registered.
        owned.0.extend((0..tape_core::spinodoid::component::COUNT).map(|c| first + c));
        register_spinodoid(&rekeyed, o, first, backend)?;
    }
    let tape = &rekeyed;
    let d = tape_core::decode(tape, &mut |_| Ok(()))?;

    // Every oracle leaf must resolve before this handle is handed out.
    //
    // fidget treats an unregistered id as zero rather than an error, on
    // purpose: a host may render preview frames before a bake has finished
    // (fidget::mesh_oracle::lookup). That is right for the live paths, which
    // decode through tape_core directly, and wrong here — a tape is replayed
    // in a process that did not bake it, so an id nothing registered means a
    // field silently reading zero, which meshes a solid block instead of
    // failing. The check is deliberately at load, not in the evaluator.
    let missing: Vec<String> = tape_core::oracle_uses(tape)
        .into_iter()
        .filter(|u| fidget::mesh_oracle::lookup(u.id).is_none())
        .map(|u| format!("{:#x} ({})", u.id, u.kind))
        .collect();
    if !missing.is_empty() {
        return Err(format!(
            "tape references {} mesh oracle(s) nothing has registered: {}. \
             A spinodoid or unknown leaf carries only its id, so the oracle has to be \
             installed before loading — see tape_oracle_uses to list what a tape needs, \
             and register_hodge_spinodoid / register_mesh_oracle / register_vdb_oracle",
            missing.len(),
            missing.join(", ")
        ));
    }

    Ok(Tape {
        combined: d.combined,
        outputs: d.outputs,
        names: d.names,
        shape_vars: d.shape_vars,
        bbox: d.bbox,
        _oracles: owned,
    })
}

fn deinterleave(pts: &[f32], n: usize) -> (Vec<f32>, Vec<f32>, Vec<f32>) {
    let mut xs = Vec::with_capacity(n);
    let mut ys = Vec::with_capacity(n);
    let mut zs = Vec::with_capacity(n);
    for i in 0..n {
        xs.push(pts[i * 3]);
        ys.push(pts[i * 3 + 1]);
        zs.push(pts[i * 3 + 2]);
    }
    (xs, ys, zs)
}

/// Isosurface algorithm for [`Tape::mesh`].
#[derive(Clone, Copy, PartialEq)]
pub enum MeshAlgorithm {
    /// Streaming marching cubes (isotropic grid, triangle soup).
    MarchingCubes,
    /// Manifold dual contouring (indexed, sharp-feature preserving).
    DualContour,
}

/// An in-memory triangle mesh — flat xyz vertices (`3*V`) and flat triangle
/// vertex indices (`3*T`). This is the output of meshing, separate from any
/// serialization format.
pub struct MeshData {
    pub vertices: Vec<f32>,
    pub triangles: Vec<u32>,
}

impl MeshData {
    /// Serialize to binary STL, computing a face normal per triangle. Works for
    /// both indexed (dual-contour) and soup (marching-cubes) meshes.
    pub fn to_stl(&self) -> Vec<u8> {
        let ntri = self.triangles.len() / 3;
        let mut out = vec![0u8; 80 + 4 + ntri * 50];
        out[80..84].copy_from_slice(&(ntri as u32).to_le_bytes());
        let vert = |i: u32| -> [f32; 3] {
            let i = i as usize * 3;
            [self.vertices[i], self.vertices[i + 1], self.vertices[i + 2]]
        };
        let mut dst = 84;
        for i in 0..ntri {
            let a = vert(self.triangles[i * 3]);
            let b = vert(self.triangles[i * 3 + 1]);
            let c = vert(self.triangles[i * 3 + 2]);
            let u = [b[0] - a[0], b[1] - a[1], b[2] - a[2]];
            let w = [c[0] - a[0], c[1] - a[1], c[2] - a[2]];
            let mut nrm = [
                u[1] * w[2] - u[2] * w[1],
                u[2] * w[0] - u[0] * w[2],
                u[0] * w[1] - u[1] * w[0],
            ];
            let len = (nrm[0] * nrm[0] + nrm[1] * nrm[1] + nrm[2] * nrm[2]).sqrt();
            if len > 1e-10 {
                nrm = [nrm[0] / len, nrm[1] / len, nrm[2] / len];
            }
            for v in [nrm, a, b, c] {
                for f in v {
                    out[dst..dst + 4].copy_from_slice(&f.to_le_bytes());
                    dst += 4;
                }
            }
            dst += 2; // attribute byte count (already zero)
        }
        out
    }
}

fn mesh_dual_contour(
    shape: &VmShape,
    shape_vars: &ShapeVars<f32>,
    depth: i32,
    bbox_min: [f32; 3],
    bbox_max: [f32; 3],
) -> Result<MeshData, String> {
    let world_to_model = crate::world_to_model_from_bbox(bbox_min, bbox_max);
    let settings = MeshSettings {
        depth: depth as u8,
        world_to_model,
        ..Default::default()
    };
    let bound = shape.bind(shape_vars).map_err(|e| e.to_string())?;
    let octree = Octree::build(&bound, &settings);
    let octree = octree.ok_or_else(|| "octree build failed".to_string())?;
    let mesh = octree.walk_dual();

    let mut vertices = Vec::with_capacity(mesh.vertices.len() * 3);
    for v in &mesh.vertices {
        vertices.extend_from_slice(&[v[0], v[1], v[2]]);
    }
    let mut triangles = Vec::with_capacity(mesh.triangles.len() * 3);
    for t in &mesh.triangles {
        triangles.extend_from_slice(&[t[0] as u32, t[1] as u32, t[2] as u32]);
    }
    Ok(MeshData { vertices, triangles })
}

fn mesh_marching_cubes(
    shape: &VmShape,
    shape_vars: &ShapeVars<f32>,
    depth: i32,
    bbox_min: [f32; 3],
    bbox_max: [f32; 3],
) -> Result<MeshData, String> {
    use fidget::shape::ShapeBulkEval;

    // Marching cubes uses an isotropic grid; pad the bbox to a cube about its
    // center (matching canvas_mesh_stl_marching_cubes).
    let cx = 0.5 * (bbox_min[0] + bbox_max[0]);
    let cy = 0.5 * (bbox_min[1] + bbox_max[1]);
    let cz = 0.5 * (bbox_min[2] + bbox_max[2]);
    let mut longest = (bbox_max[0] - bbox_min[0])
        .max(bbox_max[1] - bbox_min[1])
        .max(bbox_max[2] - bbox_min[2]);
    if longest <= 0.0 {
        longest = 1.0;
    }
    let scale = 0.5 * longest;
    let origin_x = cx - scale;
    let origin_y = cy - scale;
    let origin_z = cz - scale;

    let resolution = 1usize << (depth.max(0) as usize);
    let n = resolution + 1;
    let cell_size = (2.0 * scale) / resolution as f32;

    let points_per_slice = n * n;
    let bytes_per_slice = points_per_slice * 4;
    let total_ram = crate::get_total_ram_bytes();
    let max_buffer_slices = (total_ram / 2 / bytes_per_slice).max(2).min(n);
    let target_eval_points: usize = 4 * 1024 * 1024;
    let eval_batch_slices = (target_eval_points / points_per_slice).max(1).min(max_buffer_slices);

    let tape = shape.float_slice_tape(Default::default());
    let mut eval = ShapeBulkEval::<VmFloatSliceEval<N>>::default();
    let mut eval_error = false;
    let eval_chunk = |iz_start: usize, count: usize| -> Vec<f32> {
        let total_points = count * points_per_slice;
        let mut xs = Vec::with_capacity(total_points);
        let mut ys = Vec::with_capacity(total_points);
        let mut zs = Vec::with_capacity(total_points);
        for iz in iz_start..iz_start + count {
            let z = origin_z + iz as f32 * cell_size;
            for iy in 0..n {
                for ix in 0..n {
                    xs.push(origin_x + ix as f32 * cell_size);
                    ys.push(origin_y + iy as f32 * cell_size);
                    zs.push(z);
                }
            }
        }
        match eval.eval_with_vars(&tape, &xs, &ys, &zs, shape_vars) {
            Ok(r) => r[..total_points].to_vec(),
            Err(_) => {
                eval_error = true;
                vec![1.0; total_points]
            }
        }
    };

    let triangle_data = crate::marching_cubes::chunked_marching_cubes(
        n,
        origin_x,
        origin_y,
        origin_z,
        cell_size,
        max_buffer_slices,
        eval_batch_slices,
        eval_chunk,
        |_, _| {},
    );
    if eval_error {
        return Err("marching cubes evaluation failed".into());
    }

    // triangle_data is a soup of [nx,ny,nz, v0(3), v1(3), v2(3)] per triangle;
    // drop the normals and emit each triangle's three corners with sequential
    // indices (unwelded, like the STL).
    let ntri = triangle_data.len() / 12;
    let mut vertices = Vec::with_capacity(ntri * 9);
    let mut triangles = Vec::with_capacity(ntri * 3);
    for i in 0..ntri {
        let base = i * 12;
        vertices.extend_from_slice(&triangle_data[base + 3..base + 12]);
        let v = (i * 3) as u32;
        triangles.extend_from_slice(&[v, v + 1, v + 2]);
    }
    Ok(MeshData { vertices, triangles })
}

impl Tape {
    /// Decode a `tape::Tape` and build an evaluable handle, choosing each
    /// oracle's structure as the GSL author wrote it (bvhSdf -> BVH, vdbSdf ->
    /// VDB).
    pub fn from_bytes(bytes: &[u8]) -> Result<Tape, String> {
        Tape::from_bytes_with(bytes, OracleMode::default())
    }

    /// Like `from_bytes`, but `mode` chooses the oracle structure on load — e.g.
    /// force exact BVH everywhere, or VDB at a target resolution.
    pub fn from_bytes_with(bytes: &[u8], mode: OracleMode) -> Result<Tape, String> {
        Tape::from_bytes_with_overrides(bytes, mode, &HashMap::new())
    }

    /// Like `from_bytes_with`, with per-payload departures from `mode`, keyed
    /// by each payload's asset id ([`tape_oracle_payloads`] lists them). An
    /// override must name a payload in the tape and set only what its kind
    /// reads.
    pub fn from_bytes_with_overrides(
        bytes: &[u8],
        mode: OracleMode,
        overrides: &HashMap<u32, OracleOverride>,
    ) -> Result<Tape, String> {
        let tape = tape_proto::Tape::decode(bytes).map_err(|e| format!("decode: {:?}", e))?;
        build_tape(&tape, mode, overrides)
    }

    /// The world-space scene bbox recorded in the tape, if any:
    /// `[xmin, ymin, zmin, xmax, ymax, zmax]`.
    pub fn bbox(&self) -> Option<[f32; 6]> {
        self.bbox
    }

    /// Number of individual outputs (GSL returns).
    pub fn num_outputs(&self) -> usize {
        self.outputs.len()
    }

    /// Per-output display names (parallel to outputs; UI labels). May contain
    /// empty strings for unnamed expression returns.
    pub fn output_names(&self) -> &[String] {
        &self.names
    }

    /// The shape for the requested output, or the combined (min) field when
    /// `output` is None.
    fn shape_for(&self, output: Option<usize>) -> Result<&VmShape, String> {
        match output {
            None => Ok(&self.combined),
            Some(i) => self
                .outputs
                .get(i)
                .ok_or_else(|| format!("output {} out of range (have {})", i, self.outputs.len())),
        }
    }

    /// Field value at each point. `pts` is `n*3` xyz-interleaved; returns `n`.
    /// `output` selects an individual output, or None for the combined field.
    pub fn eval(&self, pts: &[f32], output: Option<usize>) -> Result<Vec<f32>, String> {
        let shape = self.shape_for(output)?;
        let n = pts.len() / 3;
        let (xs, ys, zs) = deinterleave(pts, n);
        let tape = shape.float_slice_tape(Default::default());
        let mut eval = ShapeBulkEval::<VmFloatSliceEval<N>>::default();
        let values = eval
            .eval_with_vars(&tape, &xs, &ys, &zs, &self.shape_vars)
            .map_err(|e| format!("{:?}", e))?;
        Ok(values[..n].to_vec())
    }

    /// Value + gradient at each point. Returns `n*4` as `[v, dx, dy, dz]`.
    pub fn grad(&self, pts: &[f32], output: Option<usize>) -> Result<Vec<f32>, String> {
        let shape = self.shape_for(output)?;
        let n = pts.len() / 3;
        // Forward-mode AD: seed each axis input with a unit partial so the
        // evaluator carries (value, d/dx, d/dy, d/dz) through the tape.
        let mut xs = Vec::with_capacity(n);
        let mut ys = Vec::with_capacity(n);
        let mut zs = Vec::with_capacity(n);
        for i in 0..n {
            xs.push(Grad::new(pts[i * 3], 1.0, 0.0, 0.0));
            ys.push(Grad::new(pts[i * 3 + 1], 0.0, 1.0, 0.0));
            zs.push(Grad::new(pts[i * 3 + 2], 0.0, 0.0, 1.0));
        }
        let tape = shape.grad_slice_tape(Default::default());
        let mut eval = ShapeBulkEval::<VmGradSliceEval<N>>::default();
        let grads = eval
            .eval_with_vars(&tape, &xs, &ys, &zs, &self.shape_vars)
            .map_err(|e| format!("{:?}", e))?;
        let mut out = Vec::with_capacity(n * 4);
        for g in &grads[..n] {
            out.push(g.v);
            out.push(g.dx);
            out.push(g.dy);
            out.push(g.dz);
        }
        Ok(out)
    }

    /// Interval bound of the field over an axis-aligned box. `bbox` is 6 f32
    /// (xmin,ymin,zmin,xmax,ymax,zmax); returns (lower, upper).
    pub fn interval(&self, bbox: &[f32], output: Option<usize>) -> Result<(f32, f32), String> {
        if bbox.len() < 6 {
            return Err("bbox needs 6 floats".into());
        }
        let shape = self.shape_for(output)?;
        let tape = shape.interval_tape(Default::default());
        let mut eval = ShapeTracingEval::<VmIntervalEval<N>>::default();
        let x = Interval::new(bbox[0], bbox[3]);
        let y = Interval::new(bbox[1], bbox[4]);
        let z = Interval::new(bbox[2], bbox[5]);
        let (value, _trace) = eval
            .eval_with_vars(&tape, x, y, z, &self.shape_vars)
            .map_err(|e| format!("{:?}", e))?;
        Ok((value.lower(), value.upper()))
    }

    /// Mesh the field over a world-space bbox and return binary STL bytes.
    pub fn mesh_stl(
        &self,
        depth: i32,
        bbox_min: [f32; 3],
        bbox_max: [f32; 3],
        output: Option<usize>,
    ) -> Result<Vec<u8>, String> {
        let shape = self.shape_for(output)?;
        let world_to_model = crate::world_to_model_from_bbox(bbox_min, bbox_max);
        let settings = MeshSettings {
            depth: depth as u8,
            world_to_model,
            ..Default::default()
        };
        let bound = shape.bind(&self.shape_vars).map_err(|e| e.to_string())?;
        let octree = Octree::build(&bound, &settings);
        let octree = octree.ok_or_else(|| "octree build failed".to_string())?;
        let mesh = octree.walk_dual();
        let mut stl: Vec<u8> = Vec::new();
        mesh.write_stl(&mut stl).map_err(|e| format!("{:?}", e))?;
        if stl.is_empty() {
            return Err("empty mesh".to_string());
        }
        Ok(stl)
    }

    /// Mesh the field over a world-space bbox with the chosen `algorithm`,
    /// returning an in-memory [`MeshData`] (vertices + triangle indices).
    /// `output` selects one output, or None for the combined field. Output
    /// format is the caller's choice on the result (arrays or `to_stl`).
    pub fn mesh(
        &self,
        algorithm: MeshAlgorithm,
        depth: i32,
        bbox_min: [f32; 3],
        bbox_max: [f32; 3],
        output: Option<usize>,
    ) -> Result<MeshData, String> {
        let shape = self.shape_for(output)?;
        match algorithm {
            MeshAlgorithm::DualContour => {
                mesh_dual_contour(shape, &self.shape_vars, depth, bbox_min, bbox_max)
            }
            MeshAlgorithm::MarchingCubes => {
                mesh_marching_cubes(shape, &self.shape_vars, depth, bbox_min, bbox_max)
            }
        }
    }
}

/// One row of [`tape_oracle_uses`]: `(id, kind, self_contained, slot)`.
pub type OracleUseRow = (u32, String, bool, Option<u32>);

/// What a tape needs bound before it can be loaded: one row per mesh-oracle id
/// it references, as `(id, kind, self_contained, slot)`. `self_contained` is
/// true when loading resolves the id by itself — a library-baked noise or
/// cellular tile, or a source mesh the tape carries — and false when the caller
/// has to register the oracle first, because the tape holds the id and nothing
/// else. `slot` is the table slot to register against where the kind has one.
///
/// This is read-only and needs no registration, so it is the first call a host
/// replaying someone else's tape makes.
///
/// A spinodoid leaf is self-contained here when the tape carries its slot's
/// payload, which `Tape` bakes on load; tape-core alone cannot.
pub fn tape_oracle_uses(bytes: &[u8]) -> Result<Vec<OracleUseRow>, String> {
    let tape = tape_proto::Tape::decode(bytes).map_err(|e| format!("decode: {:?}", e))?;
    let baked = |slot: Option<u32>, kind: &str| {
        let first = match (kind, slot) {
            ("spinodoid", Some(s)) => tape_core::spinodoid::oracle_id(s, 0),
            ("hodgeSpinodoid", Some(s)) => tape_core::spinodoid::hodge::oracle_id(s, 0),
            _ => return false,
        };
        tape.oracles
            .iter()
            .any(|o| o.kind == tape_proto::OracleKind::Spinodoid as i32 && o.asset_id == first)
    };
    Ok(tape_core::oracle_uses(&tape)
        .into_iter()
        .map(|u| {
            let self_contained = u.self_contained || baked(u.slot, u.kind);
            (u.id, u.kind.to_string(), self_contained, u.slot)
        })
        .collect())
}

/// The oracle payloads a tape carries, as `(asset_id, kind)`: the ids
/// per-payload load settings are keyed by. `kind` is "bvh" or "vdb" (a mesh,
/// as the GSL author wrote it) or "spinodoid". A payload the tape repeats is
/// listed once.
pub fn tape_oracle_payloads(bytes: &[u8]) -> Result<Vec<(u32, String)>, String> {
    let tape = tape_proto::Tape::decode(bytes).map_err(|e| format!("decode: {:?}", e))?;
    let mut out: Vec<(u32, String)> = Vec::new();
    for o in &tape.oracles {
        if out.iter().any(|(id, _)| *id == o.asset_id) {
            continue;
        }
        let kind = match tape_proto::OracleKind::try_from(o.kind) {
            Ok(tape_proto::OracleKind::MeshBvh) => "bvh",
            Ok(tape_proto::OracleKind::Vdb) => "vdb",
            Ok(tape_proto::OracleKind::Spinodoid) => "spinodoid",
            Err(_) => return Err(format!("oracle {:#x}: unknown kind {}", o.asset_id, o.kind)),
        };
        out.push((o.asset_id, kind.to_string()));
    }
    Ok(out)
}

/// Install a baked hodge-spinodoid kernel table in `slot`, registering the
/// oracles for every leaf id `slot` covers, so a tape referencing them loads.
///
/// For a tape whose leaf has no payload. A tape with one is baked on load
/// instead (register_spinodoid), and its payload's table replaces this one.
/// hodge's `_gabor_kernels` produces exactly these kernels, and omega's
/// `HodgeBake` is pinned to it (common/noise/spinodoid/hodge_parity_test.go).
///
/// `kernels` is `cells * impulses * 8` f64, in row-major cell order over
/// `size` from `origin`, each kernel `[jitter xyz, dir xyz, beta, gamma]`.
/// A `periodic` table must be hodge's 32-cubed window; a bounded one is
/// located by `bounds` (`[xmin, ymin, zmin, xmax, ymax, zmax]`) and reads as a
/// conservative distance-to-domain outside it.
#[allow(clippy::too_many_arguments)]
pub fn register_hodge_spinodoid(
    slot: u32,
    kernels: &[f64],
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
) -> Result<(), String> {
    let table = tape_core::spinodoid::hodge::Table::new(
        kernels.to_vec(),
        origin,
        size.map(|s| s as usize),
        bounds,
        periodic,
        impulses,
        cell_size,
        a,
        sigma,
        base_beta,
        level,
        local,
    )?;
    tape_core::spinodoid::hodge::register(slot, table)
}

/// Register a bvhSdf mesh oracle payload by asset id. `vertices` is xyz-packed
/// f32 (`3*vertex_count`); `triangles` is index-packed u32 (`3*tri_count`).
/// Returns 0 on success. Safe wrapper over the `mesh_oracle` C-ABI.
pub fn register_mesh_oracle(asset_id: u32, vertices: &[f32], triangles: &[u32]) -> i32 {
    unsafe {
        crate::mesh_oracle::register_mesh_oracle(
            asset_id,
            vertices.as_ptr(),
            vertices.len() / 3,
            triangles.as_ptr(),
            triangles.len() / 3,
        )
    }
}

/// Register a vdbSdf mesh oracle payload by asset id (mesh is rasterized to a
/// narrow-band VDB level set with `voxel_size` / `band_width_voxels`).
/// Returns 0 on success. Safe wrapper over the `vdb_oracle` C-ABI.
pub fn register_vdb_oracle(
    asset_id: u32,
    vertices: &[f32],
    triangles: &[u32],
    voxel_size: f32,
    band_width_voxels: f32,
) -> i32 {
    unsafe {
        crate::vdb_oracle::register_vdb_oracle(
            asset_id,
            vertices.as_ptr(),
            vertices.len() / 3,
            triangles.as_ptr(),
            triangles.len() / 3,
            voxel_size,
            band_width_voxels,
        )
    }
}

// ============================ C-ABI (Go cgo path) ============================

/// Serialize the given output tree indices into a `tape::Tape` protobuf,
/// returned as C-heap bytes (free with `render_result_free`).
///
/// # Safety
/// `outputs_ptr` must point to `count` valid `c_ulong` values.
#[no_mangle]
pub unsafe extern "C" fn serialize_tape(outputs_ptr: *const c_ulong, count: usize) -> RenderResult {
    if outputs_ptr.is_null() {
        return build_render_error(-40);
    }
    let raw = unsafe { std::slice::from_raw_parts(outputs_ptr, count) };
    let indices: Vec<usize> = raw.iter().map(|&v| v as usize).collect();
    match std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| serialize_outputs(&indices))) {
        Ok(Ok(bytes)) => build_render_result(bytes),
        Ok(Err(_)) => build_render_error(-41),
        Err(_) => build_render_error(-42),
    }
}

/// Serialize the outputs with their appearance (see
/// `serialize_outputs_with_appearance`): `appearance_ptr` holds 15 tree indices
/// per output, colour r, g, b then the 12 material entries, `u64::MAX` for a
/// colour or material the output does not have (all three, or all twelve), and
/// `displaced_ptr` one flag per output.
///
/// # Safety
/// `outputs_ptr` and `displaced_ptr` must point to `count` values and
/// `appearance_ptr` to `15 * count`.
#[no_mangle]
pub unsafe extern "C" fn serialize_tape_appearance(
    outputs_ptr: *const c_ulong,
    count: usize,
    appearance_ptr: *const c_ulong,
    displaced_ptr: *const u8,
) -> RenderResult {
    if outputs_ptr.is_null() || appearance_ptr.is_null() || displaced_ptr.is_null() {
        return build_render_error(-40);
    }
    let raw = unsafe { std::slice::from_raw_parts(outputs_ptr, count) };
    let indices: Vec<usize> = raw.iter().map(|&v| v as usize).collect();
    let app = unsafe { std::slice::from_raw_parts(appearance_ptr, 15 * count) };
    let disp = unsafe { std::slice::from_raw_parts(displaced_ptr, count) };
    let absent = c_ulong::MAX;
    let appearance: Vec<AppearanceTrees> = (0..count)
        .map(|k| {
            let a = &app[15 * k..15 * k + 15];
            AppearanceTrees {
                color: (a[0] != absent).then(|| [a[0] as usize, a[1] as usize, a[2] as usize]),
                material: (a[3] != absent).then(|| std::array::from_fn(|i| a[3 + i] as usize)),
                displaced: disp[k] != 0,
            }
        })
        .collect();
    match std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        serialize_outputs_with_appearance(&indices, &appearance)
    })) {
        Ok(Ok(bytes)) => build_render_result(bytes),
        Ok(Err(_)) => build_render_error(-41),
        Err(_) => build_render_error(-42),
    }
}

/// Decode a tape into a `Tape` handle. Returns null on failure.
///
/// # Safety
/// `tape_ptr` must point to `tape_len` valid bytes.
#[no_mangle]
pub unsafe extern "C" fn tape_build(tape_ptr: *const u8, tape_len: usize) -> *mut Tape {
    if tape_ptr.is_null() {
        return std::ptr::null_mut();
    }
    let bytes = unsafe { std::slice::from_raw_parts(tape_ptr, tape_len) };
    match std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| Tape::from_bytes(bytes))) {
        Ok(Ok(h)) => Box::into_raw(Box::new(h)),
        _ => std::ptr::null_mut(),
    }
}

/// Free a `Tape` handle returned by `tape_build`. (Named `tape_handle_free`
/// rather than `tape_free` to avoid colliding with the legacy GLSL-codegen
/// `tape_free`, which frees a CString.)
///
/// # Safety
/// `handle` must be a pointer from `tape_build`, freed at most once.
#[no_mangle]
pub unsafe extern "C" fn tape_handle_free(handle: *mut Tape) {
    if !handle.is_null() {
        unsafe {
            drop(Box::from_raw(handle));
        }
    }
}

/// Field value at `n` points. `pts`: `n*3` floats; `out`: `n` floats.
///
/// # Safety
/// `pts` must point to `n*3` floats and `out` to `n` floats.
#[no_mangle]
pub unsafe extern "C" fn tape_eval(
    handle: *mut Tape,
    pts: *const c_float,
    n: usize,
    out: *mut c_float,
) -> c_int {
    if handle.is_null() || pts.is_null() || out.is_null() {
        return -1;
    }
    let h = unsafe { &*handle };
    let pts = unsafe { std::slice::from_raw_parts(pts, n * 3) };
    let out = unsafe { std::slice::from_raw_parts_mut(out, n) };
    match std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| h.eval(pts, None))) {
        Ok(Ok(v)) => {
            out.copy_from_slice(&v);
            0
        }
        Ok(Err(_)) => -2,
        Err(_) => -3,
    }
}

/// Value + gradient at `n` points. `out`: `n*4` floats as `[v, dx, dy, dz]`.
///
/// # Safety
/// `pts` must point to `n*3` floats and `out` to `n*4` floats.
#[no_mangle]
pub unsafe extern "C" fn tape_grad(
    handle: *mut Tape,
    pts: *const c_float,
    n: usize,
    out: *mut c_float,
) -> c_int {
    if handle.is_null() || pts.is_null() || out.is_null() {
        return -1;
    }
    let h = unsafe { &*handle };
    let pts = unsafe { std::slice::from_raw_parts(pts, n * 3) };
    let out = unsafe { std::slice::from_raw_parts_mut(out, n * 4) };
    match std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| h.grad(pts, None))) {
        Ok(Ok(v)) => {
            out.copy_from_slice(&v);
            0
        }
        Ok(Err(_)) => -2,
        Err(_) => -3,
    }
}

/// Interval bound over a box. `bbox`: 6 floats; `out`: 2 floats (lower, upper).
///
/// # Safety
/// `bbox` must point to 6 floats and `out` to 2 floats.
#[no_mangle]
pub unsafe extern "C" fn tape_interval(
    handle: *mut Tape,
    bbox: *const c_float,
    out: *mut c_float,
) -> c_int {
    if handle.is_null() || bbox.is_null() || out.is_null() {
        return -1;
    }
    let h = unsafe { &*handle };
    let b = unsafe { std::slice::from_raw_parts(bbox, 6) };
    let out = unsafe { std::slice::from_raw_parts_mut(out, 2) };
    match std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| h.interval(b, None))) {
        Ok(Ok((lo, hi))) => {
            out[0] = lo;
            out[1] = hi;
            0
        }
        Ok(Err(_)) => -2,
        Err(_) => -3,
    }
}

/// Mesh the field over a world-space bbox; returns binary STL (C-heap; free
/// with `render_result_free`).
#[no_mangle]
pub extern "C" fn tape_mesh_stl(
    handle: *mut Tape,
    depth: c_int,
    bbox_min_x: c_float,
    bbox_min_y: c_float,
    bbox_min_z: c_float,
    bbox_max_x: c_float,
    bbox_max_y: c_float,
    bbox_max_z: c_float,
) -> RenderResult {
    if handle.is_null() {
        return build_render_error(-50);
    }
    let h = unsafe { &*handle };
    let result = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        h.mesh_stl(
            depth,
            [bbox_min_x, bbox_min_y, bbox_min_z],
            [bbox_max_x, bbox_max_y, bbox_max_z],
            None,
        )
    }));
    match result {
        Ok(Ok(stl)) => build_render_result(stl),
        Ok(Err(_)) => build_render_error(-51),
        Err(_) => build_render_error(-52),
    }
}

// ============================ tests ============================

#[cfg(test)]
mod tests {
    use super::*;
    use crate::{new_const, tree_add, tree_sqrt, tree_square, tree_sub};

    // Build a unit sphere SDF: sqrt(x^2 + y^2 + z^2) - 1, serialize it from the
    // global TREES DAG, replay it into a Tape, and check that point / gradient /
    // interval / mesh all behave through the round trip.
    #[test]
    fn sphere_round_trip() {
        crate::clear(); // seeds X=0, Y=1, Z=2
        let (x, y, z) = (0u64, 1u64, 2u64);
        let s = tree_add(tree_add(tree_square(x), tree_square(y)), tree_square(z));
        let r = tree_sqrt(s);
        let one = new_const(1.0);
        let sphere = tree_sub(r, one);

        let bytes = serialize_outputs(&[sphere as usize]).expect("serialize");
        let h = Tape::from_bytes(&bytes).expect("build");

        assert_eq!(h.num_outputs(), 1);

        let vals = h
            .eval(&[0.0, 0.0, 0.0, 1.0, 0.0, 0.0, 2.0, 0.0, 0.0], None)
            .expect("eval");
        assert!((vals[0] + 1.0).abs() < 1e-4, "center: {}", vals[0]);
        assert!(vals[1].abs() < 1e-4, "surface: {}", vals[1]);
        assert!((vals[2] - 1.0).abs() < 1e-4, "outside: {}", vals[2]);

        // output 0 matches the combined field for a single-output tape
        let v0 = h.eval(&[2.0, 0.0, 0.0], Some(0)).expect("eval out0");
        assert!((v0[0] - 1.0).abs() < 1e-4, "output 0: {}", v0[0]);
        assert!(h.eval(&[0.0, 0.0, 0.0], Some(1)).is_err(), "out-of-range output");

        let g = h.grad(&[1.0, 0.0, 0.0], None).expect("grad");
        assert!(g[0].abs() < 1e-4, "grad value {}", g[0]);
        assert!((g[1] - 1.0).abs() < 1e-3, "grad dx {}", g[1]);
        assert!(g[2].abs() < 1e-3 && g[3].abs() < 1e-3, "grad dy/dz {} {}", g[2], g[3]);

        let (lo, hi) = h
            .interval(&[-2.0, -2.0, -2.0, 2.0, 2.0, 2.0], None)
            .expect("interval");
        assert!(lo <= 0.0 && hi >= 0.0, "interval [{}, {}]", lo, hi);

        let stl = h
            .mesh_stl(5, [-2.0, -2.0, -2.0], [2.0, 2.0, 2.0], None)
            .expect("mesh");
        assert!(stl.len() > 84, "STL too small: {}", stl.len());
    }

    // A tape with a noiseHash3 leaf carries only the oracle leaf id — no tile.
    // Replay must register the channel oracle from the library-baked tile and
    // evaluate to the same values as the reference hash. Builds the tape by hand
    // (no TREES / clear()) so it can't race the sphere round trip above.
    #[test]
    fn noise_leaf_replays_from_lib_baked_tile() {
        use prost::Message;
        use tape_proto::node::N;
        let id = crate::noise_tile_oracle::NOISE_ORACLE_ID_BASE; // channel 0
        let tape = tape_proto::Tape {
            version: 1,
            nodes: vec![tape_proto::Node {
                n: Some(N::MeshOracle(id)),
            }],
            outputs: vec![0],
            bbox: Vec::new(),
            oracles: Vec::new(),
            output_names: Vec::new(),
            appearance: Vec::new(),
        };

        let h = Tape::from_bytes(&tape.encode_to_vec()).expect("build");
        let pts = [0.0, 0.0, 0.0, 5.0, 9.0, 13.0, -1.0, -7.0, -2.0];
        let vals = h.eval(&pts, None).expect("eval");
        for (i, c) in pts.chunks(3).enumerate() {
            let mut out = [0f32; 3];
            unsafe { crate::noise_tile_oracle::noise_hash3(c[0], c[1], c[2], out.as_mut_ptr()) };
            assert!(
                (vals[i] - out[0]).abs() < 1e-6,
                "cell {:?}: tape {} != ref {}",
                c,
                vals[i],
                out[0]
            );
        }
    }

    // A one-leaf tape over a hodge spinodoid slot.
    fn hodge_leaf_tape(slot: u32) -> Vec<u8> {
        use prost::Message;
        use tape_proto::node::N;
        let id = tape_core::spinodoid::hodge::oracle_id(slot, 0);
        tape_proto::Tape {
            version: 1,
            nodes: vec![tape_proto::Node {
                n: Some(N::MeshOracle(id)),
            }],
            outputs: vec![0],
            bbox: Vec::new(),
            oracles: Vec::new(),
            output_names: Vec::new(),
            appearance: Vec::new(),
        }
        .encode_to_vec()
    }

    // A smallest valid bounded table: 2x2x2 cells of one kernel each, every
    // kernel the same plane wave. Enough to bind a slot; the field's agreement
    // with hodge is pinned in Go (hodge_parity_test.go), not here.
    fn flat_table() -> Vec<f64> {
        let mut k = Vec::new();
        for _ in 0..8 {
            k.extend_from_slice(&[0.5, 0.5, 0.5, 1.0, 0.0, 0.0, TWO_PI, 0.0]);
        }
        k
    }
    const TWO_PI: f64 = std::f64::consts::TAU;

    // The trap this guards: fidget reads an unregistered oracle leaf as zero on
    // purpose, so before this check a tape of a spinodoid replayed in a process
    // that had not baked it meshed a 100%-solid block and reported nothing.
    // Replay is exactly that process, so here it has to fail loudly.
    #[test]
    fn spinodoid_leaf_without_a_table_is_an_error() {
        let bytes = hodge_leaf_tape(250);
        let err = match Tape::from_bytes(&bytes) {
            Ok(_) => panic!("a slot nothing registered must not load"),
            Err(e) => e,
        };
        assert!(
            err.contains("hodgeSpinodoid") && err.contains("register_hodge_spinodoid"),
            "error should name the kind and the way out, got: {err}"
        );
    }

    // And with the table bound first, the same tape loads and evaluates. This
    // is the hodge/elasticam path: inspect, register, load.
    #[test]
    fn spinodoid_leaf_replays_from_a_registered_table() {
        let slot = 251;
        register_hodge_spinodoid(
            slot,
            &flat_table(),
            [0, 0, 0],
            [2, 2, 2],
            [0.0, 0.0, 0.0, 2.0, 2.0, 2.0],
            false,
            1,
            1.0,
            1.0,
            1.0,
            TWO_PI,
            0.0,
            false,
        )
        .expect("register");

        let bytes = hodge_leaf_tape(slot);
        let h = Tape::from_bytes(&bytes).expect("build");
        // On a kernel centre and half a wavelength along it: crest against
        // trough. (A quarter wavelength out, every kernel contributes cos of
        // +-pi/2 and the sum is zero whatever the table — not a useful probe.)
        let pts = [0.5, 0.5, 0.5, 1.0, 0.5, 0.5];
        let vals = h.eval(&pts, None).expect("eval");
        assert!(
            vals.iter().all(|v| v.is_finite()),
            "non-finite field: {vals:?}"
        );
        // Negative is solid, so a crest (every kernel in phase at its own
        // centre) reads positive and the trough half a wave along reads
        // negative. Straddling the level is what says the table is bound and
        // the wave is being evaluated, not that a leaf folded to zero.
        assert!(
            vals[0] > 0.0 && vals[1] < 0.0,
            "crest and trough of a cos wave should straddle the level: {vals:?}"
        );
    }

    // A tape as omega writes it: one spinodoid leaf with its payload. The
    // arguments are spinodoid.gabor(6, [pi/2, pi/2, pi/2]) at bias 0.3.
    fn payload_tape(asset_id: u32) -> tape_proto::Tape {
        use tape_proto::node::N;
        tape_proto::Tape {
            version: 1,
            nodes: vec![tape_proto::Node {
                n: Some(N::MeshOracle(asset_id)),
            }],
            outputs: vec![0],
            oracles: vec![tape_proto::OraclePayload {
                kind: tape_proto::OracleKind::Spinodoid as i32,
                asset_id,
                spinodoid: Some(tape_proto::SpinodoidParams {
                    fast: true,
                    period: 6.0,
                    theta_radians: vec![std::f64::consts::FRAC_PI_2; 3],
                    poles: vec![1.0, 0.0, 0.0, 0.0, 1.0, 0.0, 0.0, 0.0, 1.0],
                    packet_width: 3.0,
                    impulses: 8,
                    pool: 4096,
                    seed: 0,
                    local_normalize: true,
                    scale: vec![1.0],
                    orientation: vec![1.0, 0.0, 0.0, 0.0, 1.0, 0.0, 0.0, 0.0, 1.0],
                    periodic: true,
                    bias: 0.3,
                    ..Default::default()
                }),
                ..Default::default()
            }],
            ..Default::default()
        }
    }

    // What omega's Go baker gives for payload_tape's arguments at these
    // points (Bake / HodgeBake, Distance with local normalization).
    const PAYLOAD_POINTS: [f32; 9] = [0.0, 0.0, 0.0, 2.5, -3.75, 6.125, -18.5, 11.25, 4.0];
    const GO_LATTICE: [f64; 3] = [0.52092486296795693, 1.0646893390705905, -0.02383445563949289];
    const GO_MANUFACTURING: [f64; 3] = [-0.41810563283014013, -1.0595184216823339, 0.18173290695875885];

    fn load(t: &tape_proto::Tape, backend: Option<SpinodoidBackend>) -> Tape {
        use prost::Message;
        let mut mode = OracleMode::default();
        if let Some(b) = backend {
            mode.spinodoid = b;
        }
        Tape::from_bytes_with(&t.encode_to_vec(), mode).expect("load")
    }

    fn assert_near(got: &[f32], want: &[f64], tolerance: f64, what: &str) {
        for (g, w) in got.iter().zip(want) {
            assert!(
                (*g as f64 - w).abs() < tolerance,
                "{what}: got {got:?}, want {want:?}"
            );
        }
    }

    // The payload is enough: no registration, and the backend picks which of
    // the two fields the same arguments describe. f32 evaluation, so the
    // tolerance is float32 rounding on values of order one.
    #[test]
    fn spinodoid_payload_loads_as_either_backend() {
        let t = payload_tape(tape_core::spinodoid::oracle_id(240, 0));
        let fast = load(&t, None).eval(&PAYLOAD_POINTS, None).expect("eval");
        assert_near(&fast, &GO_LATTICE, 1e-5, "default backend (fast)");

        let t = payload_tape(tape_core::spinodoid::oracle_id(241, 0));
        let fast = load(&t, Some(SpinodoidBackend::Fast)).eval(&PAYLOAD_POINTS, None).expect("eval");
        assert_near(&fast, &GO_LATTICE, 1e-5, "fast");

        let t = payload_tape(tape_core::spinodoid::oracle_id(242, 0));
        let mfg = load(&t, Some(SpinodoidBackend::Manufacturing))
            .eval(&PAYLOAD_POINTS, None)
            .expect("eval");
        assert_near(&mfg, &GO_MANUFACTURING, 1e-5, "manufacturing");
    }

    // A tape written before omega marked every leaf fast has its leaf in
    // hodge's id block. The backend still decides: with none chosen it reads
    // as the lattice, and the table is bound under the ids the tape carries.
    #[test]
    fn hodge_block_leaf_follows_the_backend() {
        let mut t = payload_tape(tape_core::spinodoid::hodge::oracle_id(243, 0));
        t.oracles[0].spinodoid.as_mut().unwrap().fast = false;
        let fast = load(&t, None).eval(&PAYLOAD_POINTS, None).expect("eval");
        assert_near(&fast, &GO_LATTICE, 1e-5, "hodge-block leaf, default backend");

        let mut t = payload_tape(tape_core::spinodoid::hodge::oracle_id(244, 0));
        t.oracles[0].spinodoid.as_mut().unwrap().fast = false;
        let mfg = load(&t, Some(SpinodoidBackend::Manufacturing))
            .eval(&PAYLOAD_POINTS, None)
            .expect("eval");
        assert_near(&mfg, &GO_MANUFACTURING, 1e-5, "hodge-block leaf, manufacturing");
    }

    // Two spinodoids from the same arguments, one output each: the second
    // payload's override reads it as the manufacturing field, and the first
    // keeps the tape-wide default.
    fn two_spinodoid_tape() -> tape_proto::Tape {
        use tape_proto::node::N;
        let mut t = payload_tape(tape_core::spinodoid::oracle_id(0, 0));
        let second = payload_tape(tape_core::spinodoid::oracle_id(1, 0));
        t.nodes.push(tape_proto::Node {
            n: Some(N::MeshOracle(second.oracles[0].asset_id)),
        });
        t.outputs.push(1);
        t.oracles.push(second.oracles[0].clone());
        t
    }

    fn overrides(entries: &[(u32, OracleOverride)]) -> HashMap<u32, OracleOverride> {
        entries.iter().copied().collect()
    }

    #[test]
    fn each_spinodoid_payload_takes_its_own_backend() {
        use prost::Message;
        let bytes = two_spinodoid_tape().encode_to_vec();
        let mfg = OracleOverride {
            spinodoid: Some(SpinodoidBackend::Manufacturing),
            ..Default::default()
        };
        let h = Tape::from_bytes_with_overrides(
            &bytes,
            OracleMode::default(),
            &overrides(&[(tape_core::spinodoid::oracle_id(1, 0), mfg)]),
        )
        .expect("load");
        let first = h.eval(&PAYLOAD_POINTS, Some(0)).expect("eval");
        let second = h.eval(&PAYLOAD_POINTS, Some(1)).expect("eval");
        assert_near(&first, &GO_LATTICE, 1e-5, "payload without an override");
        assert_near(&second, &GO_MANUFACTURING, 1e-5, "payload overridden to manufacturing");
    }

    // A tetrahedron, outward-wound, as a mesh payload.
    fn tetra_payload(asset_id: u32) -> tape_proto::OraclePayload {
        tape_proto::OraclePayload {
            kind: tape_proto::OracleKind::MeshBvh as i32,
            asset_id,
            vertices: vec![1.0, 1.0, 1.0, 1.0, -1.0, -1.0, -1.0, 1.0, -1.0, -1.0, -1.0, 1.0],
            triangles: vec![0, 1, 2, 0, 3, 1, 0, 2, 3, 1, 3, 2],
            ..Default::default()
        }
    }

    // Two copies of one mesh. Loaded as authored, both are exact BVHs and
    // agree; with the second forced to a coarse VDB, only the second moves.
    #[test]
    fn each_mesh_payload_takes_its_own_structure() {
        use prost::Message;
        use tape_proto::node::N;
        let t = tape_proto::Tape {
            version: 1,
            nodes: vec![
                tape_proto::Node { n: Some(N::MeshOracle(1)) },
                tape_proto::Node { n: Some(N::MeshOracle(2)) },
            ],
            outputs: vec![0, 1],
            oracles: vec![tetra_payload(1), tetra_payload(2)],
            ..Default::default()
        };
        let vdb = OracleOverride {
            force: Some(OracleForce::Vdb),
            resolution: Some(16),
            ..Default::default()
        };
        let bytes = t.encode_to_vec();
        // Off a vertex along x the exact distance is 9.
        let far = [10.0, 1.0, 1.0];
        let authored = Tape::from_bytes(&bytes).expect("load as authored");
        let a0 = authored.eval(&far, Some(0)).expect("eval")[0];
        let a1 = authored.eval(&far, Some(1)).expect("eval")[0];
        assert!((a0 - 9.0).abs() < 1e-4, "BVH distance: {a0}");
        assert_eq!(a0, a1, "two BVHs of one mesh");

        let h = Tape::from_bytes_with_overrides(&bytes, OracleMode::default(), &overrides(&[(2, vdb)]))
            .expect("load");
        let bvh = h.eval(&far, Some(0)).expect("eval")[0];
        let vdb = h.eval(&far, Some(1)).expect("eval")[0];
        assert_eq!(bvh, a0, "the payload without an override is still the BVH");
        assert!((vdb - bvh).abs() > 1e-3, "the overridden payload should be the coarse VDB: {vdb} vs {bvh}");
    }

    // Two tapes with the same asset ids, read different ways, in one process:
    // each keeps its own oracle, and dropping one releases only its own.
    #[test]
    fn tapes_sharing_asset_ids_keep_their_own_oracles() {
        use prost::Message;
        let bytes = payload_tape(tape_core::spinodoid::oracle_id(0, 0)).encode_to_vec();
        let fast = Tape::from_bytes(&bytes).expect("load fast");
        let mut mode = OracleMode::default();
        mode.spinodoid = SpinodoidBackend::Manufacturing;
        let mfg = Tape::from_bytes_with(&bytes, mode).expect("load manufacturing");
        assert_near(&fast.eval(&PAYLOAD_POINTS, None).expect("eval"), &GO_LATTICE, 1e-5, "first tape");
        assert_near(&mfg.eval(&PAYLOAD_POINTS, None).expect("eval"), &GO_MANUFACTURING, 1e-5, "second tape");

        let released = mfg._oracles.0.clone();
        drop(mfg);
        for id in released {
            let o = fidget::mesh_oracle::lookup(id).expect("still registered");
            assert!(o.point(0.0, 0.0, 0.0).is_nan(), "{id:#x} should read as released");
        }
        assert_near(&fast.eval(&PAYLOAD_POINTS, None).expect("eval"), &GO_LATTICE, 1e-5, "survivor");
    }

    #[test]
    fn overrides_must_fit_the_tape() {
        use prost::Message;
        let bytes = payload_tape(tape_core::spinodoid::oracle_id(0, 0)).encode_to_vec();
        let mode = OracleMode::default();
        let backend = OracleOverride {
            spinodoid: Some(SpinodoidBackend::Manufacturing),
            ..Default::default()
        };
        let err = Tape::from_bytes_with_overrides(&bytes, mode, &overrides(&[(7, backend)]))
            .err()
            .expect("an id the tape lacks");
        assert!(err.contains("no oracle payload"), "{err}");
        let force = OracleOverride {
            force: Some(OracleForce::Vdb),
            ..Default::default()
        };
        let err = Tape::from_bytes_with_overrides(
            &bytes,
            mode,
            &overrides(&[(tape_core::spinodoid::oracle_id(0, 0), force)]),
        )
        .err()
        .expect("a mesh setting on a spinodoid");
        assert!(err.contains("mesh settings"), "{err}");

        let mesh = tape_proto::Tape {
            version: 1,
            nodes: vec![tape_proto::Node {
                n: Some(tape_proto::node::N::MeshOracle(1)),
            }],
            outputs: vec![0],
            oracles: vec![tetra_payload(1)],
            ..Default::default()
        };
        let err = Tape::from_bytes_with_overrides(&mesh.encode_to_vec(), mode, &overrides(&[(1, backend)]))
            .err()
            .expect("a backend on a mesh");
        assert!(err.contains("spinodoid setting"), "{err}");
        assert_eq!(
            tape_oracle_payloads(&two_spinodoid_tape().encode_to_vec()).expect("payloads"),
            vec![
                (tape_core::spinodoid::oracle_id(0, 0), "spinodoid".to_string()),
                (tape_core::spinodoid::oracle_id(1, 0), "spinodoid".to_string()),
            ]
        );
    }

    // With its payload in the tape, a spinodoid leaf needs nothing registered.
    #[test]
    fn oracle_uses_counts_a_payload_as_self_contained() {
        use prost::Message;
        let uses = tape_oracle_uses(&payload_tape(tape_core::spinodoid::oracle_id(247, 0)).encode_to_vec())
            .expect("uses");
        assert_eq!(uses.len(), 1);
        assert_eq!(uses[0].1, "spinodoid");
        assert!(uses[0].2, "the tape carries this slot's payload");
    }

    // Field drivers travel as nodes of the tape and are read at every kernel
    // centre: scale 1 + 0.004 x, and a rotation about z by 0.05 y, over a
    // bounded domain. The Go reference reads the drivers in f64 and this
    // reader in f32, as omega's own bake does, hence the looser bound.
    #[test]
    fn spinodoid_payload_reads_its_driver_nodes() {
        use tape_proto::node::N;
        use tape_proto::{BinaryOp as B, UnaryOp as U};
        let asset_id = tape_core::spinodoid::oracle_id(245, 0);
        let mut t = payload_tape(asset_id);
        let mut push = |n: N| {
            t.nodes.push(tape_proto::Node { n: Some(n) });
            (t.nodes.len() - 1) as u32
        };
        let bin = |op: B, lhs, rhs| N::Binary(tape_proto::Binary { op: op as i32, lhs, rhs });
        let un = |op: U, arg| N::Unary(tape_proto::Unary { op: op as i32, arg });
        let x = push(N::Axis(tape_proto::Axis::X as i32));
        let y = push(N::Axis(tape_proto::Axis::Y as i32));
        let one = push(N::Constant(1.0));
        let zero = push(N::Constant(0.0));
        let k = push(N::Constant(0.004));
        let kx = push(bin(B::Mul, x, k));
        let scale = push(bin(B::Add, one, kx));
        let w = push(N::Constant(0.05));
        let a = push(bin(B::Mul, y, w));
        let c = push(un(U::Cos, a));
        let s = push(un(U::Sin, a));
        let ns = push(un(U::Neg, s));
        let p = t.oracles[0].spinodoid.as_mut().unwrap();
        p.theta_radians = vec![
            std::f64::consts::FRAC_PI_2,
            std::f64::consts::FRAC_PI_2,
            std::f64::consts::FRAC_PI_6,
        ];
        p.seed = 3;
        p.bias = 0.25;
        p.scale = Vec::new();
        p.orientation = Vec::new();
        p.scale_nodes = vec![scale];
        p.orientation_nodes = vec![c, ns, zero, s, c, zero, zero, zero, one];
        p.periodic = false;
        p.domain = vec![-20.0, -20.0, -20.0, 20.0, 20.0, 20.0];

        let pts = [0.0, 0.0, 0.0, 2.5, -3.75, 6.125, -18.5, 11.25, 4.0, 30.0, 0.0, 0.0];
        let fast = load(&t, None).eval(&pts, None).expect("eval");
        assert_near(
            &fast,
            &[-1.7870893793673694, 0.86395264880181899, -0.22853634702410267, 10.0],
            1e-4,
            "driven lattice",
        );

        let mut t = t;
        let id = tape_core::spinodoid::oracle_id(246, 0);
        t.oracles[0].asset_id = id;
        t.nodes[0].n = Some(N::MeshOracle(id));
        let mfg = load(&t, Some(SpinodoidBackend::Manufacturing)).eval(&pts, None).expect("eval");
        assert_near(
            &mfg,
            &[-1.9002615102383427, 0.44457195544755918, -0.28181459383269619, 10.0],
            1e-4,
            "driven manufacturing",
        );
    }

    // What a host has to ask before it can bind anything: which ids a tape
    // references, and which of them the tape cannot resolve by itself.
    #[test]
    fn oracle_uses_separates_self_contained_leaves_from_bindable_ones() {
        let uses = tape_oracle_uses(&hodge_leaf_tape(7)).expect("uses");
        assert_eq!(uses.len(), 1);
        let (id, kind, self_contained, slot) = &uses[0];
        assert_eq!(*id, tape_core::spinodoid::hodge::oracle_id(7, 0));
        assert_eq!(kind, "hodgeSpinodoid");
        assert!(!*self_contained, "a spinodoid leaf carries no table");
        assert_eq!(*slot, Some(7));

        use prost::Message;
        use tape_proto::node::N;
        let noise = tape_proto::Tape {
            version: 1,
            nodes: vec![tape_proto::Node {
                n: Some(N::MeshOracle(
                    crate::noise_tile_oracle::NOISE_ORACLE_ID_BASE,
                )),
            }],
            outputs: vec![0],
            bbox: Vec::new(),
            oracles: Vec::new(),
            output_names: Vec::new(),
            appearance: Vec::new(),
        }
        .encode_to_vec();
        let uses = tape_oracle_uses(&noise).expect("uses");
        assert_eq!(uses.len(), 1);
        assert_eq!(uses[0].1, "noise");
        assert!(uses[0].2, "a noise tile is library-baked");
    }
}
