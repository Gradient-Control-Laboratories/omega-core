//! Host-side registry for mesh-distance oracles referenced by
//! [`Op::MeshOracle`](crate::context::Op::MeshOracle) /
//! [`RegOp::MeshOracle`](crate::compiler::RegOp::MeshOracle).
//!
//! The fidget tree is closed over a fixed set of math operations, so a
//! mesh-distance leaf cannot embed its triangle data inline. Instead, the host
//! (the rust wrapper around fidget) registers an oracle implementation keyed by
//! a `u32` asset id, and the VM evaluators call into it through this registry
//! at evaluation time. JIT/bytecode backends do not support the oracle and
//! must refuse to compile programs that reference one.

use crate::types::{Grad, Interval};
use std::collections::HashMap;
use std::sync::{Arc, OnceLock, RwLock};

/// Host implementation backing a mesh-distance oracle leaf. Implementations
/// must be cheap to clone-share (the registry hands out `Arc`s) and safe to
/// call from any thread (fidget evaluators run in parallel).
pub trait MeshOracle: Send + Sync {
    /// Signed distance from the point `(x, y, z)` to the mesh surface.
    /// Negative inside, positive outside.
    fn point(&self, x: f32, y: f32, z: f32) -> f32;

    /// Conservative interval bound on the signed distance over the
    /// axis-aligned box described by component-wise intervals. The returned
    /// interval must be inclusive — fidget-mesh's octree pruning relies on
    /// `[lo, hi].contains(true_distance)` for every point in the box. A
    /// looser bound is correct but degrades pruning.
    fn interval(
        &self,
        x: Interval,
        y: Interval,
        z: Interval,
    ) -> Interval;

    /// Optional gradient hook for the gradient evaluator. The default
    /// implementation returns the point distance with zero partials, which is
    /// safe for marching cubes (it uses central-difference normals from point
    /// samples, not the carried gradient) but not useful for any caller that
    /// truly relies on the gradient. Override to provide an analytical
    /// gradient when the BSP can supply one cheaply.
    fn grad(&self, x: f32, y: f32, z: f32) -> Grad {
        Grad::new(self.point(x, y, z), 0.0, 0.0, 0.0)
    }
}

static REGISTRY: OnceLock<RwLock<HashMap<u32, Arc<dyn MeshOracle>>>> =
    OnceLock::new();

fn registry() -> &'static RwLock<HashMap<u32, Arc<dyn MeshOracle>>> {
    REGISTRY.get_or_init(|| RwLock::new(HashMap::new()))
}

/// Register an oracle for the given asset id. Subsequent evaluations of
/// [`MeshOracle(asset_id)`](crate::context::Op::MeshOracle) will dispatch to
/// this implementation. Re-registering with the same id replaces the prior
/// oracle.
pub fn register(asset_id: u32, oracle: Arc<dyn MeshOracle>) {
    registry().write().unwrap().insert(asset_id, oracle);
}

/// Look up the oracle for the given asset id, or `None` if no oracle has been
/// registered. Evaluators that find `None` here will treat the leaf as zero
/// rather than panic, since the host may legitimately render preview frames
/// before the oracle has finished baking.
pub fn lookup(asset_id: u32) -> Option<Arc<dyn MeshOracle>> {
    registry().read().unwrap().get(&asset_id).cloned()
}
