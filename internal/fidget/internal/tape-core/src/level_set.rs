//! A vdbSdf oracle read from the level set a tape carries (tape.proto's
//! VdbLevelSet), for a reader without OpenVDB. It answers exactly as render's
//! `vdb_oracle::VdbOracle` does over the grid the writer baked: the point
//! value is OpenVDB's `BoxSampler` (Interpolation.h) at the world point, and
//! the distance extension, interval and gradient are VdbOracle's, so a mesh
//! of the same field over the same box comes out the same.

use std::sync::Arc;

use fidget_core::mesh_oracle::{self, MeshOracle};
use fidget_core::types::{Grad, Interval};

use crate::tape_proto;

const LEAF_DIM: i32 = 8;
const LEAF_SIZE: usize = 512;

pub struct LevelSetOracle {
    inv_voxel: f64,
    voxel_size: f32,
    background: f32,
    bbox_min: [f32; 3],
    bbox_max: [f32; 3],
    origin: [i32; 3],
    dims: [i32; 3],
    block_leaf: Vec<i32>,
    block_value: Vec<f32>,
    leaf_values: Vec<f32>,
}

/// The block (8^3 voxels) holding index coordinate v.
fn block_of(v: i32) -> i32 {
    v.div_euclid(LEAF_DIM)
}

impl LevelSetOracle {
    pub fn new(ls: &tape_proto::VdbLevelSet) -> Result<Self, String> {
        if ls.voxel_size <= 0.0 {
            return Err(format!("level set voxel size {}", ls.voxel_size));
        }
        if ls.active_bbox.len() != 6 || ls.block_origin.len() != 3 || ls.block_dims.len() != 3 {
            return Err("level set is missing its box or block table".into());
        }
        let dims = [ls.block_dims[0] as i32, ls.block_dims[1] as i32, ls.block_dims[2] as i32];
        let blocks = dims.iter().map(|&d| d as usize).product::<usize>();
        if ls.block_leaf.len() != blocks || ls.block_value.len() != blocks {
            return Err(format!(
                "level set block table holds {} leaves and {} values for {} blocks",
                ls.block_leaf.len(),
                ls.block_value.len(),
                blocks
            ));
        }
        let leaves = ls.leaf_values.len() / LEAF_SIZE;
        if ls.leaf_values.len() % LEAF_SIZE != 0 || ls.block_leaf.iter().any(|&l| l >= leaves as i32) {
            return Err(format!("level set leaf data does not match its {} leaves", leaves));
        }
        let b = &ls.active_bbox;
        Ok(Self {
            // ScaleMap::applyInverseMap multiplies by 1 / scale.
            inv_voxel: 1.0 / ls.voxel_size,
            voxel_size: ls.voxel_size as f32,
            background: ls.background,
            bbox_min: [b[0], b[1], b[2]],
            bbox_max: [b[3], b[4], b[5]],
            origin: [ls.block_origin[0], ls.block_origin[1], ls.block_origin[2]],
            dims,
            block_leaf: ls.block_leaf.clone(),
            block_value: ls.block_value.clone(),
            leaf_values: ls.leaf_values.clone(),
        })
    }

    /// The tree's value at voxel (i, j, k), as getValue would read it.
    fn value(&self, i: i32, j: i32, k: i32) -> f32 {
        let b = [block_of(i) - self.origin[0], block_of(j) - self.origin[1], block_of(k) - self.origin[2]];
        if (0..3).any(|a| b[a] < 0 || b[a] >= self.dims[a]) {
            return self.background;
        }
        let block = (b[0] + self.dims[0] * (b[1] + self.dims[1] * b[2])) as usize;
        match self.block_leaf[block] {
            -1 => self.block_value[block],
            leaf => {
                let offset = (((i & 7) << 6) | ((j & 7) << 3) | (k & 7)) as usize;
                self.leaf_values[leaf as usize * LEAF_SIZE + offset]
            }
        }
    }

    /// BoxSampler::sample at the world point: the float values of the eight
    /// voxels around it, blended in double precision weights.
    fn sample(&self, p: [f32; 3]) -> f32 {
        let c = [p[0] as f64 * self.inv_voxel, p[1] as f64 * self.inv_voxel, p[2] as f64 * self.inv_voxel];
        let i = [c[0].floor() as i32, c[1].floor() as i32, c[2].floor() as i32];
        let u = [c[0] - i[0] as f64, c[1] - i[1] as f64, c[2] - i[2] as f64];
        let lerp = |a: f32, b: f32, w: f64| -> f32 { a + ((b - a) as f64 * w) as f32 };
        let v = |di: i32, dj: i32, dk: i32| self.value(i[0] + di, i[1] + dj, i[2] + dk);
        lerp(
            lerp(lerp(v(0, 0, 0), v(0, 0, 1), u[2]), lerp(v(0, 1, 0), v(0, 1, 1), u[2]), u[1]),
            lerp(lerp(v(1, 0, 0), v(1, 0, 1), u[2]), lerp(v(1, 1, 0), v(1, 1, 1), u[2]), u[1]),
            u[0],
        )
    }

    // The rest is render's VdbOracle, line for line.

    fn signed_distance(&self, p: [f32; 3]) -> f32 {
        let sample = self.sample(p);
        let dx = (self.bbox_min[0] - p[0]).max(0.0).max(p[0] - self.bbox_max[0]);
        let dy = (self.bbox_min[1] - p[1]).max(0.0).max(p[1] - self.bbox_max[1]);
        let dz = (self.bbox_min[2] - p[2]).max(0.0).max(p[2] - self.bbox_max[2]);
        let outside_dist = (dx * dx + dy * dy + dz * dz).sqrt();
        if outside_dist == 0.0 {
            sample
        } else {
            sample + outside_dist
        }
    }
}

impl MeshOracle for LevelSetOracle {
    fn point(&self, x: f32, y: f32, z: f32) -> f32 {
        self.signed_distance([x, y, z])
    }

    fn interval(&self, x: Interval, y: Interval, z: Interval) -> Interval {
        let xs = [x.lower(), x.upper()];
        let ys = [y.lower(), y.upper()];
        let zs = [z.lower(), z.upper()];
        let mut lo = f32::INFINITY;
        let mut hi = f32::NEG_INFINITY;
        for &cx in &xs {
            for &cy in &ys {
                for &cz in &zs {
                    let d = self.signed_distance([cx, cy, cz]);
                    if d < lo {
                        lo = d;
                    }
                    if d > hi {
                        hi = d;
                    }
                }
            }
        }
        let dx = xs[1] - xs[0];
        let dy = ys[1] - ys[0];
        let dz = zs[1] - zs[0];
        let half_diag = 0.5 * (dx * dx + dy * dy + dz * dz).sqrt();
        Interval::new(lo - half_diag, hi + half_diag)
    }

    fn grad(&self, x: f32, y: f32, z: f32) -> Grad {
        let eps = (self.voxel_size * 0.5).max(1e-3);
        let v = self.signed_distance([x, y, z]);
        let dxp = self.signed_distance([x + eps, y, z]);
        let dxm = self.signed_distance([x - eps, y, z]);
        let dyp = self.signed_distance([x, y + eps, z]);
        let dym = self.signed_distance([x, y - eps, z]);
        let dzp = self.signed_distance([x, y, z + eps]);
        let dzm = self.signed_distance([x, y, z - eps]);
        let gx = (dxp - dxm) / (2.0 * eps);
        let gy = (dyp - dym) / (2.0 * eps);
        let gz = (dzp - dzm) / (2.0 * eps);
        Grad::new(v, gx, gy, gz)
    }
}

/// Registers a VDB payload's carried level set under the key its leaves use.
/// Returns false, registering nothing, when the payload carries none.
pub fn register(o: &tape_proto::OraclePayload) -> Result<bool, String> {
    let Some(ls) = &o.level_set else {
        return Ok(false);
    };
    let oracle = LevelSetOracle::new(ls)?;
    mesh_oracle::register(o.asset_id | crate::VDB_ASSET_BIT, Arc::new(oracle));
    Ok(true)
}
