//! C ABI for the cellular oracles (the GSL cellularF1 / cellularF2 /
//! cellularFeature / cellularNearCell / cellularFarCell builtins).
//!
//! The oracles themselves live in tape-core (cellular.rs), which the browser
//! build shares; this file is the entry point the Go side calls and the tree
//! builder that pushes onto TREES. Mirrors noise_tile_oracle.rs.

use std::os::raw::c_ulong;

use fidget::context::Tree;
use tape_core::cellular;

/// Build a `Tree` leaf evaluating one component of a cellular scan at the
/// implicit (x,y,z) — the convert layer remaps those to the lattice coordinate —
/// over `octaves` tile lookups per feature-point hash. Returns the TREES index,
/// or u64::MAX on a bad argument.
#[no_mangle]
pub extern "C" fn tree_cellular(comp: u32, octaves: u32) -> c_ulong {
    if cellular::register(comp, octaves).is_err() {
        return u64::MAX;
    }
    // The (v - v) triplet is zero but keeps Var::X/Y/Z in the tape so the
    // octree feeds real per-axis intervals to interval() — without it the
    // oracle would be handed degenerate inputs and its Lipschitz bound, the
    // whole reason these oracles exist, would never be exercised. Same trick as
    // tree_noise_channel and tree_mesh_oracle.
    let (x, y, z) = Tree::axes();
    let zero_via_axes = (x.clone() - x) + (y.clone() - y) + (z.clone() - z);
    let tree = Tree::mesh_oracle(cellular::oracle_id(comp, octaves)) + zero_via_axes;
    let mut trees = crate::TREES.lock().unwrap();
    trees.push(tree);
    (trees.len() - 1) as c_ulong
}
