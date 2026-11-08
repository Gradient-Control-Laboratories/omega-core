//! C ABI for the Gabor spinodoid oracles (the gaborSpinodoid /
//! fastGaborSpinodoid builtins and their gradients).
//!
//! The tables and the oracles live in tape-core (spinodoid.rs for the lattice
//! field, spinodoid/hodge.rs for hodge's, each with its own slots); this file is
//! the entry point the Go side calls and the tree builder that pushes onto
//! TREES. Mirrors cellular_oracle.rs.

use std::os::raw::{c_int, c_ulong};

use fidget::context::Tree;
use tape_core::spinodoid;

/// Install a baked table in `slot`. `sites` holds `len` f64s, SITE_STRIDE per
/// lattice site in common/noise/spinodoid's order. Returns 0 on success, -1 on
/// a null pointer, -2 on a table the oracle rejects.
///
/// # Safety
/// `sites` must point to `len` valid f64s.
#[no_mangle]
#[allow(clippy::too_many_arguments)]
pub unsafe extern "C" fn register_spinodoid(
    slot: u32,
    sites: *const f64,
    len: usize,
    origin: *const i64,
    size: *const u32,
    bounds: *const f64,
    periodic: c_int,
    spacing: f64,
    reach: f64,
    a: f64,
    sigma: f64,
    base_beta: f64,
    level: f64,
    local: c_int,
) -> c_int {
    if sites.is_null() || origin.is_null() || size.is_null() || bounds.is_null() {
        return -1;
    }
    let data = unsafe { std::slice::from_raw_parts(sites, len) }.to_vec();
    let origin: [i64; 3] = unsafe { std::slice::from_raw_parts(origin, 3) }
        .try_into()
        .unwrap();
    let size: [u32; 3] = unsafe { std::slice::from_raw_parts(size, 3) }
        .try_into()
        .unwrap();
    let bounds: [f64; 6] = unsafe { std::slice::from_raw_parts(bounds, 6) }
        .try_into()
        .unwrap();
    let table = match spinodoid::Table::new(
        data,
        origin,
        size.map(|s| s as usize),
        bounds,
        periodic != 0,
        spacing,
        reach,
        a,
        sigma,
        base_beta,
        level,
        local != 0,
    ) {
        Ok(t) => t,
        Err(_) => return -2,
    };
    if spinodoid::register(slot, table).is_err() {
        return -2;
    }
    0
}

/// Build a `Tree` leaf evaluating one component (0 the distance, 1..3 its
/// gradient) of the table in `slot` at the implicit (x,y,z); the convert layer
/// remaps those to the query point. Returns the TREES index, or u64::MAX when
/// the slot holds no table.
#[no_mangle]
pub extern "C" fn tree_spinodoid(slot: u32, comp: u32) -> c_ulong {
    if !spinodoid::has(slot) || comp >= spinodoid::component::COUNT {
        return u64::MAX;
    }
    // Keep Var::X/Y/Z in the tape so the octree hands interval() real boxes;
    // the same trick as tree_cellular.
    let (x, y, z) = Tree::axes();
    let zero_via_axes = (x.clone() - x) + (y.clone() - y) + (z.clone() - z);
    let tree = Tree::mesh_oracle(spinodoid::oracle_id(slot, comp)) + zero_via_axes;
    let mut trees = crate::TREES.lock().unwrap();
    trees.push(tree);
    (trees.len() - 1) as c_ulong
}

/// Install a baked table of hodge's field in `slot`. `kernels` holds `len`
/// f64s, KERNEL_STRIDE per kernel in common/noise/spinodoid's order, with
/// `impulses` of them per cell, over the cells `origin`/`size` hold. A
/// stationary field is `periodic` over hodge's own 32-cubed window; a driven
/// one is bounded by `bounds` and hashed on true cell indices, so it is
/// hodge's field everywhere inside the domain. Its slots are its own,
/// separate from the lattice's. Same returns as register_spinodoid.
///
/// # Safety
/// `kernels` must point to `len` valid f64s.
#[no_mangle]
#[allow(clippy::too_many_arguments)]
pub unsafe extern "C" fn register_hodge_spinodoid(
    slot: u32,
    kernels: *const f64,
    len: usize,
    origin: *const i64,
    size: *const u32,
    bounds: *const f64,
    periodic: c_int,
    impulses: usize,
    cell_size: f64,
    a: f64,
    sigma: f64,
    base_beta: f64,
    level: f64,
    local: c_int,
) -> c_int {
    if kernels.is_null() || origin.is_null() || size.is_null() || bounds.is_null() {
        return -1;
    }
    let data = unsafe { std::slice::from_raw_parts(kernels, len) }.to_vec();
    let origin: [i64; 3] = unsafe { std::slice::from_raw_parts(origin, 3) }
        .try_into()
        .unwrap();
    let size: [u32; 3] = unsafe { std::slice::from_raw_parts(size, 3) }
        .try_into()
        .unwrap();
    let bounds: [f64; 6] = unsafe { std::slice::from_raw_parts(bounds, 6) }
        .try_into()
        .unwrap();
    let table = match spinodoid::hodge::Table::new(
        data,
        origin,
        size.map(|s| s as usize),
        bounds,
        periodic != 0,
        impulses,
        cell_size,
        a,
        sigma,
        base_beta,
        level,
        local != 0,
    ) {
        Ok(t) => t,
        Err(_) => return -2,
    };
    if spinodoid::hodge::register(slot, table).is_err() {
        return -2;
    }
    0
}

/// tree_spinodoid for a hodge table: the same leaf shape over hodge's own
/// slots and oracle ids.
#[no_mangle]
pub extern "C" fn tree_hodge_spinodoid(slot: u32, comp: u32) -> c_ulong {
    if !spinodoid::hodge::has(slot) || comp >= spinodoid::component::COUNT {
        return u64::MAX;
    }
    let (x, y, z) = Tree::axes();
    let zero_via_axes = (x.clone() - x) + (y.clone() - y) + (z.clone() - z);
    let tree = Tree::mesh_oracle(spinodoid::hodge::oracle_id(slot, comp)) + zero_via_axes;
    let mut trees = crate::TREES.lock().unwrap();
    trees.push(tree);
    (trees.len() - 1) as c_ulong
}

/// The mesh-oracle leaf id one component of a lattice-spinodoid slot carries.
///
/// A tape export needs this: the leaf travels as a bare id, and the payload
/// describing the field has to be keyed to the same id. Exposed rather than
/// mirrored in Go so there is one definition of the id layout.
/// Returns u32::MAX for an out-of-range component.
#[no_mangle]
pub extern "C" fn spinodoid_oracle_id(slot: u32, comp: u32) -> u32 {
    if comp >= spinodoid::component::COUNT {
        return u32::MAX;
    }
    spinodoid::oracle_id(slot, comp)
}

/// spinodoid_oracle_id for hodge's slots, which are their own id space.
#[no_mangle]
pub extern "C" fn hodge_spinodoid_oracle_id(slot: u32, comp: u32) -> u32 {
    if comp >= spinodoid::component::COUNT {
        return u32::MAX;
    }
    spinodoid::hodge::oracle_id(slot, comp)
}
