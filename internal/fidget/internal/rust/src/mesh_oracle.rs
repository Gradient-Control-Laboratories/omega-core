//! Triangle-BSP oracle implementation for `fidget::Op::MeshOracle` leaves.
//!
//! Wired up in two steps by the host:
//!
//!  1. Call [`register_asset`] with the triangle soup for each imported mesh.
//!     The BSP, bbox, and winding-number data are built once and cached.
//!  2. Call [`tree_mesh_oracle`] to get a `Tree` leaf that, when evaluated,
//!     dispatches through [`fidget::mesh_oracle`] back into the registered
//!     [`TriBspOracle`].
//!
//! Evaluation model:
//!  * **Point:** BSP traversal with AABB-to-point lower-bound pruning finds
//!    the nearest triangle; sign is determined by the generalized winding
//!    number (full triangle sum — the BSP doesn't accelerate the sign test
//!    yet, but GWN evaluates to the same thing fidget-mesh would get from a
//!    ray-parity test on a watertight input, and is much more robust to
//!    non-manifold / flipped / duplicate geometry).
//!  * **Interval:** we sample the distance at the 8 corners of the query box
//!    and expand by the box diagonal. The signed distance field is
//!    Lipschitz-1, so `|d(p) - d(q)| <= |p - q| <= box_diameter`, making this
//!    a correct-if-loose bound. Looser bounds mean fidget-mesh's octree
//!    prunes fewer cells, which is fine — correctness is preserved.
//!  * **Gradient:** analytic. The BSP walk already yields the closest point
//!    `q` on the mesh, so the gradient is the unit vector from `q` toward the
//!    query point, oriented by the winding-number sign — exact and sharp at
//!    edges/corners, and ~7× cheaper than the finite difference it replaces
//!    (one BSP walk instead of seven signed-distance evals). See
//!    [`TriBspOracle::grad`]. Consumed by the dual-contouring grad-slice
//!    evaluator for Hermite vertex placement.

use core::ffi::c_ulong;
use std::collections::HashMap;
use std::os::raw::c_int;
use std::sync::{Arc, LazyLock, Mutex};

use fidget::{
    context::Tree,
    mesh_oracle::{self, MeshOracle as MeshOracleTrait},
    types::{Grad, Interval},
};

/// Maximum triangles held in a BSP leaf. Smaller values yield deeper trees
/// with more pruning; larger values amortize the per-leaf overhead.
const LEAF_BUCKET_SIZE: usize = 12;

/// Signed-distance oracle backed by a median-split BSP over a triangle soup.
pub struct TriBspOracle {
    vertices: Vec<[f32; 3]>,
    triangles: Vec<[u32; 3]>,
    root: BspNode,
}

struct BspNode {
    bbox: Aabb,
    /// Aggregate statistics over all triangles in this subtree, used by the
    /// Fast Winding Number traversal to approximate a whole cluster's
    /// contribution from far-field points (Barill et al. 2018, "Fast Winding
    /// Numbers for Soups and Clouds", §3.2).
    cluster: ClusterStats,
    payload: NodePayload,
}

/// Dipole-moment aggregate of a cluster of triangles, precomputed once at
/// BSP construction so the Fast GWN walk can approximate a cluster's
/// contribution to the winding sum in O(1) per node visit when the query
/// point is sufficiently far.
#[derive(Copy, Clone, Default)]
struct ClusterStats {
    /// Σ (2 * area_i * normal_i) — each triangle contributes
    /// `(v1 - v0) × (v2 - v0)`. Kept as 2×(area · normal) because that's the
    /// natural cross-product form; the 0.5 scaling is folded into the final
    /// winding formula.
    normal_moment: [f64; 3],
    /// Σ (2 * area_i * centroid_i) — sum of per-triangle
    /// `(v0 + v1 + v2)/3 * 2*area`, kept weighted so we can divide out by
    /// `area_double` once to get the area-weighted mean position.
    centroid_moment: [f64; 3],
    /// Σ (2 * area_i). The raw cross-product magnitude sum.
    area_double: f64,
    /// Center and radius of a bounding sphere for this subtree. Used as the
    /// "is the query far enough?" criterion in the Fast GWN walk.
    center: [f64; 3],
    radius: f64,
}

enum NodePayload {
    Leaf(Vec<u32>),
    Inner(Box<BspNode>, Box<BspNode>),
}

#[derive(Copy, Clone)]
struct Aabb {
    lo: [f32; 3],
    hi: [f32; 3],
}

impl Aabb {
    fn empty() -> Self {
        Self {
            lo: [f32::INFINITY; 3],
            hi: [f32::NEG_INFINITY; 3],
        }
    }

    fn union_point(&mut self, p: [f32; 3]) {
        for i in 0..3 {
            if p[i] < self.lo[i] {
                self.lo[i] = p[i];
            }
            if p[i] > self.hi[i] {
                self.hi[i] = p[i];
            }
        }
    }

    fn longest_axis(&self) -> usize {
        let ext = [
            self.hi[0] - self.lo[0],
            self.hi[1] - self.lo[1],
            self.hi[2] - self.lo[2],
        ];
        if ext[0] >= ext[1] && ext[0] >= ext[2] {
            0
        } else if ext[1] >= ext[2] {
            1
        } else {
            2
        }
    }

    /// Squared distance from `p` to the nearest point on or inside `self`.
    /// Zero when `p` lies inside the box.
    fn sqdist_to_point(&self, p: [f32; 3]) -> f32 {
        let mut s = 0.0;
        for i in 0..3 {
            let d = if p[i] < self.lo[i] {
                self.lo[i] - p[i]
            } else if p[i] > self.hi[i] {
                p[i] - self.hi[i]
            } else {
                0.0
            };
            s += d * d;
        }
        s
    }
}

/// Per-triangle data precomputed once so the BSP builder and Fast GWN walk
/// don't have to recompute cross products / centroids / areas each time.
struct TriData {
    bbox: Aabb,
    centroid: [f32; 3],
    /// `(v1 - v0) × (v2 - v0)`, i.e. `2 * area * outward_normal`. The Fast
    /// GWN aggregate is a sum of these, which is why we keep them in this
    /// unnormalized form — the 0.5 is folded into the winding formula.
    cross: [f64; 3],
}

/// Result of a nearest-point query: the squared distance to the closest
/// point on the mesh and that point itself. The closest point is what makes
/// the SDF gradient analytic (see [`TriBspOracle::grad`]).
#[derive(Copy, Clone)]
struct Nearest {
    sq: f32,
    point: [f32; 3],
}

impl TriBspOracle {
    pub fn new(vertices: Vec<[f32; 3]>, triangles: Vec<[u32; 3]>) -> Self {
        let mut tri_data: Vec<TriData> = Vec::with_capacity(triangles.len());
        for t in triangles.iter() {
            let a = vertices[t[0] as usize];
            let b = vertices[t[1] as usize];
            let c = vertices[t[2] as usize];
            let mut bb = Aabb::empty();
            bb.union_point(a);
            bb.union_point(b);
            bb.union_point(c);
            let centroid = [
                (a[0] + b[0] + c[0]) / 3.0,
                (a[1] + b[1] + c[1]) / 3.0,
                (a[2] + b[2] + c[2]) / 3.0,
            ];
            let e1 = [
                (b[0] - a[0]) as f64,
                (b[1] - a[1]) as f64,
                (b[2] - a[2]) as f64,
            ];
            let e2 = [
                (c[0] - a[0]) as f64,
                (c[1] - a[1]) as f64,
                (c[2] - a[2]) as f64,
            ];
            let cross = [
                e1[1] * e2[2] - e1[2] * e2[1],
                e1[2] * e2[0] - e1[0] * e2[2],
                e1[0] * e2[1] - e1[1] * e2[0],
            ];
            tri_data.push(TriData {
                bbox: bb,
                centroid,
                cross,
            });
        }

        let indices: Vec<u32> = (0..triangles.len() as u32).collect();
        let root = build_bsp(&vertices, &triangles, &tri_data, indices);

        Self {
            vertices,
            triangles,
            root,
        }
    }

    /// Unsigned distance to the closest point on any triangle.
    fn unsigned_distance(&self, p: [f32; 3]) -> f32 {
        self.nearest(p).sq.sqrt()
    }

    /// Nearest point on the mesh to `p`: its squared distance and the closest
    /// surface point itself. `unsigned_distance` needs only the distance; the
    /// closest point is what `grad` turns into an analytic gradient.
    fn nearest(&self, p: [f32; 3]) -> Nearest {
        // Running best. The BSP walk prunes any subtree whose AABB is farther
        // than the current best squared distance.
        let mut best = Nearest {
            sq: f32::INFINITY,
            point: p,
        };
        self.walk_nearest(&self.root, p, &mut best);
        best
    }

    fn walk_nearest(&self, node: &BspNode, p: [f32; 3], best: &mut Nearest) {
        if node.bbox.sqdist_to_point(p) >= best.sq {
            return;
        }
        match &node.payload {
            NodePayload::Leaf(tris) => {
                for &tri_idx in tris {
                    let t = self.triangles[tri_idx as usize];
                    let a = self.vertices[t[0] as usize];
                    let b = self.vertices[t[1] as usize];
                    let c = self.vertices[t[2] as usize];
                    let (d, q) = point_triangle_sq_distance(p, a, b, c);
                    if d < best.sq {
                        best.sq = d;
                        best.point = q;
                    }
                }
            }
            NodePayload::Inner(l, r) => {
                // Descend the closer child first so `best` tightens fast.
                let dl = l.bbox.sqdist_to_point(p);
                let dr = r.bbox.sqdist_to_point(p);
                if dl <= dr {
                    self.walk_nearest(l, p, best);
                    self.walk_nearest(r, p, best);
                } else {
                    self.walk_nearest(r, p, best);
                    self.walk_nearest(l, p, best);
                }
            }
        }
    }

    /// Generalized winding number via hierarchical BVH traversal. The winding
    /// number is from Jacobson-Kavan-Sorkine-Hornung 2013; the hierarchical
    /// far-field dipole approximation is from Barill et al. 2018 §3.2. Far
    /// clusters are approximated by a single dipole term; near clusters
    /// recurse. O(log N)
    /// per query instead of O(N) — the difference between minutes and hours
    /// when meshing a ~10⁵-triangle input at depth 8.
    ///
    /// Returns ~0 for points outside a closed mesh and ~±1 for points inside
    /// (sign depends on vertex winding). The caller uses `|w| > 0.5`.
    fn generalized_winding_number(&self, p: [f32; 3]) -> f32 {
        const INV_FOUR_PI: f64 = 1.0 / (4.0 * std::f64::consts::PI);
        let pd = [p[0] as f64, p[1] as f64, p[2] as f64];
        let w = self.fast_gwn_walk(&self.root, pd);
        (w * INV_FOUR_PI) as f32
    }

    /// Recursive walk summing (unnormalized) solid-angle contributions.
    /// At each node, a Barnes-Hut-style admission criterion `dist > β·radius`
    /// decides whether the cluster is far enough to use the single-dipole
    /// approximation. This is the paper's order-0 term only. The paper's
    /// `β = 2` (about 1e-3 RMS error for triangles) is quoted for its
    /// second-order expansion, so accuracy of order 0 at `β = 2` is a choice
    /// made here, not a figure taken from the paper.
    fn fast_gwn_walk(&self, node: &BspNode, p: [f64; 3]) -> f64 {
        const BETA: f64 = 2.0;
        let stats = &node.cluster;
        let dx = p[0] - stats.center[0];
        let dy = p[1] - stats.center[1];
        let dz = p[2] - stats.center[2];
        let dist2 = dx * dx + dy * dy + dz * dz;

        // Far-field dipole approximation. `normal_moment` is Σ 2·area·normal
        // and `centroid_moment / area_double` is the area-weighted centroid.
        // The 0.5 accounts for the 2x scaling we kept in normal_moment.
        if stats.area_double > 0.0
            && dist2 > (BETA * stats.radius) * (BETA * stats.radius)
        {
            let cx = stats.centroid_moment[0] / stats.area_double;
            let cy = stats.centroid_moment[1] / stats.area_double;
            let cz = stats.centroid_moment[2] / stats.area_double;
            let rx = cx - p[0];
            let ry = cy - p[1];
            let rz = cz - p[2];
            let r2 = rx * rx + ry * ry + rz * rz;
            if r2 > 0.0 {
                let r = r2.sqrt();
                let inv_r3 = 1.0 / (r2 * r);
                let dot = stats.normal_moment[0] * rx
                    + stats.normal_moment[1] * ry
                    + stats.normal_moment[2] * rz;
                return 0.5 * dot * inv_r3;
            }
        }

        match &node.payload {
            NodePayload::Leaf(tris) => {
                let mut w = 0.0;
                for &tri_idx in tris {
                    w += self.exact_triangle_solid_angle(tri_idx, p);
                }
                w
            }
            NodePayload::Inner(l, r) => {
                self.fast_gwn_walk(l, p) + self.fast_gwn_walk(r, p)
            }
        }
    }

    /// Exact signed solid angle subtended by a single triangle at `p`, used
    /// in the near-field (leaf) path of the fast GWN walk.
    fn exact_triangle_solid_angle(&self, tri_idx: u32, p: [f64; 3]) -> f64 {
        let t = self.triangles[tri_idx as usize];
        let a = self.vertices[t[0] as usize];
        let b = self.vertices[t[1] as usize];
        let c = self.vertices[t[2] as usize];
        let va = [a[0] as f64 - p[0], a[1] as f64 - p[1], a[2] as f64 - p[2]];
        let vb = [b[0] as f64 - p[0], b[1] as f64 - p[1], b[2] as f64 - p[2]];
        let vc = [c[0] as f64 - p[0], c[1] as f64 - p[1], c[2] as f64 - p[2]];
        let la = (va[0] * va[0] + va[1] * va[1] + va[2] * va[2]).sqrt();
        let lb = (vb[0] * vb[0] + vb[1] * vb[1] + vb[2] * vb[2]).sqrt();
        let lc = (vc[0] * vc[0] + vc[1] * vc[1] + vc[2] * vc[2]).sqrt();
        if la == 0.0 || lb == 0.0 || lc == 0.0 {
            return 0.0;
        }
        let det = va[0] * (vb[1] * vc[2] - vb[2] * vc[1])
            - va[1] * (vb[0] * vc[2] - vb[2] * vc[0])
            + va[2] * (vb[0] * vc[1] - vb[1] * vc[0]);
        let dot_ab = va[0] * vb[0] + va[1] * vb[1] + va[2] * vb[2];
        let dot_bc = vb[0] * vc[0] + vb[1] * vc[1] + vb[2] * vc[2];
        let dot_ca = vc[0] * va[0] + vc[1] * va[1] + vc[2] * va[2];
        let denom = la * lb * lc + dot_ab * lc + dot_bc * la + dot_ca * lb;
        2.0 * det.atan2(denom)
    }

    fn signed_distance(&self, p: [f32; 3]) -> f32 {
        let d = self.unsigned_distance(p);
        if self.generalized_winding_number(p).abs() > 0.5 {
            -d
        } else {
            d
        }
    }
}

impl MeshOracleTrait for TriBspOracle {
    fn point(&self, x: f32, y: f32, z: f32) -> f32 {
        self.signed_distance([x, y, z])
    }

    fn interval(&self, x: Interval, y: Interval, z: Interval) -> Interval {
        // Sample at the 8 corners of the query box; the SDF is 1-Lipschitz so
        // every point in the box is within the box diagonal of one corner.
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
        // For any point in an axis-aligned box, the nearest of the 8 corners
        // is at most diag/2 away. The SDF is 1-Lipschitz, so the value at any
        // interior point can differ from its nearest corner sample by at most
        // diag/2 (NOT diag — using the full diagonal is a wider-than-needed
        // bound that causes the octree to descend into cells that don't
        // actually contain the surface, doubling visited-cell count).
        let half_diag = 0.5 * (dx * dx + dy * dy + dz * dz).sqrt();
        Interval::new(lo - half_diag, hi + half_diag)
    }

    fn grad(&self, x: f32, y: f32, z: f32) -> Grad {
        // Analytic gradient of a nearest-point signed distance field: the
        // unit vector from the closest surface point `q` toward the query
        // point `p`, oriented by the interior/exterior sign. |p - q| is the
        // unsigned distance, so this vector is unit length (the SDF is
        // 1-Lipschitz) and points along increasing signed distance — exact,
        // and sharp across edges/corners where a central difference rounds
        // them off. It reuses the single BSP walk `nearest` already does, so
        // it costs ~1 walk instead of the 7 signed_distance evals a central
        // difference needs.
        let p = [x, y, z];
        let near = self.nearest(p);
        let ud = near.sq.sqrt();
        let sign = if self.generalized_winding_number(p).abs() > 0.5 {
            -1.0
        } else {
            1.0
        };
        let signed = sign * ud;

        if ud > 1e-6 {
            let inv = sign / ud;
            return Grad::new(
                signed,
                (p[0] - near.point[0]) * inv,
                (p[1] - near.point[1]) * inv,
                (p[2] - near.point[2]) * inv,
            );
        }

        // Degenerate: `p` lies (essentially) on the surface, so the direction
        // to the closest point is undefined. This is effectively never hit
        // during dual-contouring edge evaluation; fall back to a central
        // difference so the gradient stays well-defined and never NaN.
        let eps: f32 = 0.001;
        let dx_p = self.signed_distance([x + eps, y, z]);
        let dx_m = self.signed_distance([x - eps, y, z]);
        let dy_p = self.signed_distance([x, y + eps, z]);
        let dy_m = self.signed_distance([x, y - eps, z]);
        let dz_p = self.signed_distance([x, y, z + eps]);
        let dz_m = self.signed_distance([x, y, z - eps]);
        Grad::new(
            signed,
            (dx_p - dx_m) / (2.0 * eps),
            (dy_p - dy_m) / (2.0 * eps),
            (dz_p - dz_m) / (2.0 * eps),
        )
    }
}

fn build_bsp(
    vertices: &[[f32; 3]],
    triangles: &[[u32; 3]],
    tri_data: &[TriData],
    mut indices: Vec<u32>,
) -> BspNode {
    let mut bbox = Aabb::empty();
    for &i in indices.iter() {
        let bb = tri_data[i as usize].bbox;
        bbox.union_point(bb.lo);
        bbox.union_point(bb.hi);
    }

    if indices.len() <= LEAF_BUCKET_SIZE {
        let cluster = compute_cluster_stats(&bbox, &indices, vertices, triangles, tri_data);
        return BspNode {
            bbox,
            cluster,
            payload: NodePayload::Leaf(indices),
        };
    }

    // Median split on the longest axis by centroid. Uses `select_nth_unstable`
    // so the cost is O(n) rather than O(n log n) per split.
    let axis = bbox.longest_axis();
    let mid = indices.len() / 2;
    indices.select_nth_unstable_by(mid, |&a, &b| {
        tri_data[a as usize].centroid[axis]
            .partial_cmp(&tri_data[b as usize].centroid[axis])
            .unwrap_or(std::cmp::Ordering::Equal)
    });
    let right_indices = indices.split_off(mid);
    let left = build_bsp(vertices, triangles, tri_data, indices);
    let right = build_bsp(vertices, triangles, tri_data, right_indices);

    // Aggregate child stats — sums are linear in the per-triangle quantities,
    // so the parent's normal/centroid/area moments are just child sums. The
    // bounding sphere is tightened to fit the parent bbox.
    let mut cluster = ClusterStats::default();
    for i in 0..3 {
        cluster.normal_moment[i] =
            left.cluster.normal_moment[i] + right.cluster.normal_moment[i];
        cluster.centroid_moment[i] =
            left.cluster.centroid_moment[i] + right.cluster.centroid_moment[i];
    }
    cluster.area_double = left.cluster.area_double + right.cluster.area_double;
    cluster.center = [
        ((bbox.lo[0] + bbox.hi[0]) * 0.5) as f64,
        ((bbox.lo[1] + bbox.hi[1]) * 0.5) as f64,
        ((bbox.lo[2] + bbox.hi[2]) * 0.5) as f64,
    ];
    let dx = (bbox.hi[0] - bbox.lo[0]) as f64;
    let dy = (bbox.hi[1] - bbox.lo[1]) as f64;
    let dz = (bbox.hi[2] - bbox.lo[2]) as f64;
    cluster.radius = 0.5 * (dx * dx + dy * dy + dz * dz).sqrt();

    BspNode {
        bbox,
        cluster,
        payload: NodePayload::Inner(Box::new(left), Box::new(right)),
    }
}

fn compute_cluster_stats(
    bbox: &Aabb,
    indices: &[u32],
    _vertices: &[[f32; 3]],
    _triangles: &[[u32; 3]],
    tri_data: &[TriData],
) -> ClusterStats {
    let mut stats = ClusterStats::default();
    for &i in indices {
        let td = &tri_data[i as usize];
        // |cross| = 2 * area; we keep everything scaled by 2 and fold the
        // 0.5 into the Fast GWN formula (see fast_gwn_walk). This lets the
        // aggregates be simple linear sums over children at inner nodes.
        let area_double = (td.cross[0] * td.cross[0]
            + td.cross[1] * td.cross[1]
            + td.cross[2] * td.cross[2])
            .sqrt();
        stats.normal_moment[0] += td.cross[0];
        stats.normal_moment[1] += td.cross[1];
        stats.normal_moment[2] += td.cross[2];
        stats.centroid_moment[0] += area_double * td.centroid[0] as f64;
        stats.centroid_moment[1] += area_double * td.centroid[1] as f64;
        stats.centroid_moment[2] += area_double * td.centroid[2] as f64;
        stats.area_double += area_double;
    }
    stats.center = [
        ((bbox.lo[0] + bbox.hi[0]) * 0.5) as f64,
        ((bbox.lo[1] + bbox.hi[1]) * 0.5) as f64,
        ((bbox.lo[2] + bbox.hi[2]) * 0.5) as f64,
    ];
    let dx = (bbox.hi[0] - bbox.lo[0]) as f64;
    let dy = (bbox.hi[1] - bbox.lo[1]) as f64;
    let dz = (bbox.hi[2] - bbox.lo[2]) as f64;
    stats.radius = 0.5 * (dx * dx + dy * dy + dz * dz).sqrt();
    stats
}

/// Squared distance from `p` to triangle `(a, b, c)`, together with the
/// closest point on the triangle. Ericson, Real-Time Collision Detection
/// §5.1.5 — closed-form via barycentric region test. The body below follows
/// the book's `ClosestPtPointTriangle`: from Real-Time Collision Detection
/// by Christer Ericson, published by Morgan Kaufmann Publishers,
/// (c) 2005 Elsevier Inc. The closest point is
/// what `grad` turns into the analytic gradient (the unit vector from it to
/// `p`); `unsigned_distance` uses only the squared distance.
fn point_triangle_sq_distance(
    p: [f32; 3],
    a: [f32; 3],
    b: [f32; 3],
    c: [f32; 3],
) -> (f32, [f32; 3]) {
    let ab = sub(b, a);
    let ac = sub(c, a);
    let ap = sub(p, a);
    let d1 = dot(ab, ap);
    let d2 = dot(ac, ap);
    if d1 <= 0.0 && d2 <= 0.0 {
        return (sqlen(ap), a);
    }
    let bp = sub(p, b);
    let d3 = dot(ab, bp);
    let d4 = dot(ac, bp);
    if d3 >= 0.0 && d4 <= d3 {
        return (sqlen(bp), b);
    }
    let vc = d1 * d4 - d3 * d2;
    if vc <= 0.0 && d1 >= 0.0 && d3 <= 0.0 {
        let v = d1 / (d1 - d3);
        let proj = add(a, scale(ab, v));
        return (sqlen(sub(p, proj)), proj);
    }
    let cp = sub(p, c);
    let d5 = dot(ab, cp);
    let d6 = dot(ac, cp);
    if d6 >= 0.0 && d5 <= d6 {
        return (sqlen(cp), c);
    }
    let vb = d5 * d2 - d1 * d6;
    if vb <= 0.0 && d2 >= 0.0 && d6 <= 0.0 {
        let w = d2 / (d2 - d6);
        let proj = add(a, scale(ac, w));
        return (sqlen(sub(p, proj)), proj);
    }
    let va = d3 * d6 - d5 * d4;
    if va <= 0.0 && (d4 - d3) >= 0.0 && (d5 - d6) >= 0.0 {
        let w = (d4 - d3) / ((d4 - d3) + (d5 - d6));
        let proj = add(b, scale(sub(c, b), w));
        return (sqlen(sub(p, proj)), proj);
    }
    let denom = 1.0 / (va + vb + vc);
    let v = vb * denom;
    let w = vc * denom;
    let proj = add(a, add(scale(ab, v), scale(ac, w)));
    (sqlen(sub(p, proj)), proj)
}

#[inline]
fn sub(a: [f32; 3], b: [f32; 3]) -> [f32; 3] {
    [a[0] - b[0], a[1] - b[1], a[2] - b[2]]
}
#[inline]
fn add(a: [f32; 3], b: [f32; 3]) -> [f32; 3] {
    [a[0] + b[0], a[1] + b[1], a[2] + b[2]]
}
#[inline]
fn scale(a: [f32; 3], s: f32) -> [f32; 3] {
    [a[0] * s, a[1] * s, a[2] * s]
}
#[inline]
fn dot(a: [f32; 3], b: [f32; 3]) -> f32 {
    a[0] * b[0] + a[1] * b[1] + a[2] * b[2]
}
#[inline]
fn sqlen(a: [f32; 3]) -> f32 {
    a[0] * a[0] + a[1] * a[1] + a[2] * a[2]
}

// ============ FFI registry =============================================

/// Per-asset oracle cache. Keyed by the u32 asset id the host assigned during
/// import. Re-registration replaces the prior oracle in place.
static ORACLES: LazyLock<Mutex<HashMap<u32, Arc<TriBspOracle>>>> =
    LazyLock::new(|| Mutex::new(HashMap::new()));

/// Build a `TriBspOracle` from raw pointers and install it under `asset_id`.
///
/// # Safety
///
/// * `vertices_ptr` must point to `vertex_count * 3` valid f32 values.
/// * `triangles_ptr` must point to `tri_count * 3` valid u32 values.
/// * All triangle indices must be `< vertex_count`.
///
/// The function validates index bounds and returns `-1` on violation; it
/// never panics on bad FFI input.
#[no_mangle]
pub unsafe extern "C" fn register_mesh_oracle(
    asset_id: u32,
    vertices_ptr: *const f32,
    vertex_count: usize,
    triangles_ptr: *const u32,
    tri_count: usize,
) -> c_int {
    if vertices_ptr.is_null() || triangles_ptr.is_null() {
        return -1;
    }
    // Safety contract from the docstring: caller guarantees the slices cover
    // valid memory with the lengths reported.
    let vert_slice: &[f32] = unsafe {
        std::slice::from_raw_parts(vertices_ptr, vertex_count * 3)
    };
    let tri_slice: &[u32] = unsafe {
        std::slice::from_raw_parts(triangles_ptr, tri_count * 3)
    };

    let mut vertices = Vec::with_capacity(vertex_count);
    for i in 0..vertex_count {
        vertices.push([
            vert_slice[i * 3],
            vert_slice[i * 3 + 1],
            vert_slice[i * 3 + 2],
        ]);
    }

    let mut triangles = Vec::with_capacity(tri_count);
    for i in 0..tri_count {
        let a = tri_slice[i * 3];
        let b = tri_slice[i * 3 + 1];
        let c = tri_slice[i * 3 + 2];
        if a as usize >= vertex_count
            || b as usize >= vertex_count
            || c as usize >= vertex_count
        {
            return -1;
        }
        triangles.push([a, b, c]);
    }

    let oracle = Arc::new(TriBspOracle::new(vertices, triangles));
    eprintln!(
        "register_mesh_oracle: asset={} verts={} tris={} d(origin)={:.4}",
        asset_id,
        vertex_count,
        tri_count,
        oracle.signed_distance([0.0, 0.0, 0.0])
    );
    {
        let mut map = ORACLES.lock().unwrap_or_else(std::sync::PoisonError::into_inner);
        map.insert(asset_id, oracle.clone());
    }
    // Also install into the fidget-core registry so VM evaluators can
    // dispatch to it.
    mesh_oracle::register(asset_id, oracle);
    0
}

/// Returns `true` if an oracle has been registered for this asset id.
#[no_mangle]
pub extern "C" fn has_mesh_oracle(asset_id: u32) -> c_int {
    let map = ORACLES.lock().unwrap_or_else(std::sync::PoisonError::into_inner);
    if map.contains_key(&asset_id) { 1 } else { 0 }
}

/// Build a `Tree` leaf that evaluates to the oracle's signed distance at the
/// implicit `(x, y, z)`. Pushes the tree onto the global TREES vec and
/// returns its index (matching the convention of the other tree_* FFIs).
///
/// The returned tree is `MeshOracle(asset_id) + (x-x) + (y-y) + (z-z)`. The
/// `(v - v)` triplet is mathematically zero but is preserved by fidget's
/// constant-folder (since neither operand is itself a constant) — its
/// purpose is to keep `Var::X`, `Var::Y`, and `Var::Z` in the resulting
/// tape's variable map. Without that, fidget-mesh's `Octree::build` writes
/// each cell's per-axis intervals into `axes[]` slots that don't exist, and
/// the oracle's `interval(x, y, z)` method gets the degenerate `[0, 0]`
/// intervals on every cell — silently producing an empty mesh.
#[no_mangle]
pub extern "C" fn tree_mesh_oracle(asset_id: u32) -> c_ulong {
    let (x, y, z) = Tree::axes();
    let zero_via_axes =
        (x.clone() - x) + (y.clone() - y) + (z.clone() - z);
    let tree = Tree::mesh_oracle(asset_id) + zero_via_axes;
    let mut trees = crate::TREES.lock().unwrap_or_else(std::sync::PoisonError::into_inner);
    trees.push(tree);
    (trees.len() - 1) as c_ulong
}

#[cfg(test)]
mod tests {
    use super::*;
    use fidget::{shape::EzShape, vm::VmShape};

    /// Cube spanning [-1, 1]^3 as a closed triangle soup. 12 triangles, 24
    /// duplicated vertex slots (4 per face, 6 faces of 2 tris each).
    fn unit_cube_mesh() -> (Vec<[f32; 3]>, Vec<[u32; 3]>) {
        let v: Vec<[f32; 3]> = vec![
            // 8 corners
            [-1.0, -1.0, -1.0], // 0
            [1.0, -1.0, -1.0],  // 1
            [1.0, 1.0, -1.0],   // 2
            [-1.0, 1.0, -1.0],  // 3
            [-1.0, -1.0, 1.0],  // 4
            [1.0, -1.0, 1.0],   // 5
            [1.0, 1.0, 1.0],    // 6
            [-1.0, 1.0, 1.0],   // 7
        ];
        let t: Vec<[u32; 3]> = vec![
            // -z face
            [0, 2, 1],
            [0, 3, 2],
            // +z face
            [4, 5, 6],
            [4, 6, 7],
            // -y face
            [0, 1, 5],
            [0, 5, 4],
            // +y face
            [3, 7, 6],
            [3, 6, 2],
            // -x face
            [0, 4, 7],
            [0, 7, 3],
            // +x face
            [1, 2, 6],
            [1, 6, 5],
        ];
        (v, t)
    }

    #[test]
    fn point_eval_unit_cube_signs() {
        let (verts, tris) = unit_cube_mesh();
        let oracle = Arc::new(TriBspOracle::new(verts, tris));
        // Origin is deeply inside the cube; signed distance should be negative
        // and approximately -1 (closest face is 1 unit away).
        let d_inside = oracle.signed_distance([0.0, 0.0, 0.0]);
        assert!(
            d_inside < -0.9 && d_inside > -1.1,
            "inside distance was {}",
            d_inside
        );
        // Far outside on +x; signed distance should be ~1.0 (face is at x=1).
        let d_outside = oracle.signed_distance([2.0, 0.0, 0.0]);
        assert!(
            d_outside > 0.9 && d_outside < 1.1,
            "outside distance was {}",
            d_outside
        );
        // Right at a face; near zero.
        let d_face = oracle.signed_distance([1.0, 0.0, 0.0]);
        assert!(d_face.abs() < 0.01, "on-face distance was {}", d_face);
    }

    #[test]
    fn grad_is_outward_unit_normal() {
        // The analytic gradient is the outward unit normal: it points from the
        // closest surface point toward the query point, has unit length, and
        // flips correctly with the interior/exterior sign.
        let (verts, tris) = unit_cube_mesh();
        let oracle = TriBspOracle::new(verts, tris);

        // Outside, off the +x face: gradient ≈ +x, value ≈ +0.5.
        let g = oracle.grad(1.5, 0.0, 0.0);
        assert!(g.v > 0.4 && g.v < 0.6, "value {}", g.v);
        assert!((g.dx - 1.0).abs() < 1e-3, "dx {}", g.dx);
        assert!(g.dy.abs() < 1e-3 && g.dz.abs() < 1e-3, "dy {} dz {}", g.dy, g.dz);

        // Outside, off the +y face: gradient ≈ +y.
        let g = oracle.grad(0.0, 1.6, 0.0);
        assert!((g.dy - 1.0).abs() < 1e-3, "dy {}", g.dy);
        assert!(g.dx.abs() < 1e-3 && g.dz.abs() < 1e-3, "dx {} dz {}", g.dx, g.dz);

        // Inside, nearest the +x face (0.5 away, vs 1.0 to the ±y/±z faces):
        // gradient still ≈ +x (points along increasing signed distance, i.e.
        // outward), value ≈ −0.5.
        let g = oracle.grad(0.5, 0.0, 0.0);
        assert!(g.v > -0.6 && g.v < -0.4, "interior value {}", g.v);
        assert!((g.dx - 1.0).abs() < 1e-3, "interior dx {}", g.dx);
        assert!(
            g.dy.abs() < 1e-3 && g.dz.abs() < 1e-3,
            "interior dy {} dz {}",
            g.dy,
            g.dz
        );
    }

    #[test]
    fn grad_matches_central_difference() {
        // Cross-check: the analytic gradient agrees with a manual central
        // difference of the signed distance at generic off-surface points, and
        // is unit length (the field is 1-Lipschitz). Points are chosen well
        // inside a single feature's Voronoi region (face, edge, corner) so the
        // finite-difference reference doesn't straddle a medial axis.
        let (verts, tris) = thin_slab_mesh();
        let oracle = TriBspOracle::new(verts, tris);

        let eps = 0.01_f32;
        for &pt in &[
            [0.0_f32, 0.0, 1.0], // above the +z face (exterior)
            [4.0, 0.0, 1.0],     // beyond the +x/+z edge (exterior)
            [4.2, 4.2, 1.0],     // beyond the +x/+y/+z corner (exterior)
            [1.2, -0.8, 0.9],    // generic point above the +z face (exterior)
            [0.0, 0.0, 0.1],     // interior, nearest the +z face
        ] {
            let g = oracle.grad(pt[0], pt[1], pt[2]);
            let fd = |i: usize| {
                let mut a = pt;
                let mut b = pt;
                a[i] += eps;
                b[i] -= eps;
                (oracle.signed_distance(a) - oracle.signed_distance(b)) / (2.0 * eps)
            };
            let (rx, ry, rz) = (fd(0), fd(1), fd(2));
            assert!((g.dx - rx).abs() < 2e-2, "pt {:?}: dx analytic {} fd {}", pt, g.dx, rx);
            assert!((g.dy - ry).abs() < 2e-2, "pt {:?}: dy analytic {} fd {}", pt, g.dy, ry);
            assert!((g.dz - rz).abs() < 2e-2, "pt {:?}: dz analytic {} fd {}", pt, g.dz, rz);
            let mag = (g.dx * g.dx + g.dy * g.dy + g.dz * g.dz).sqrt();
            assert!((mag - 1.0).abs() < 1e-3, "pt {:?}: |grad| {}", pt, mag);
        }
    }

    #[test]
    fn vm_tree_dispatches_through_oracle() {
        // Round-trip: register oracle → build tree from `bvhSdf(99) - 0.2` →
        // evaluate via VmShape → confirm the value matches a direct oracle
        // call shifted by 0.2. Uses asset id 99 to avoid colliding with
        // anything else in the global registry across tests.
        let (verts, tris) = unit_cube_mesh();
        let oracle = Arc::new(TriBspOracle::new(verts.clone(), tris.clone()));
        mesh_oracle::register(99, oracle.clone());

        let tree = Tree::mesh_oracle(99) - Tree::constant(0.2);
        let mut ctx = fidget::Context::new();
        let node = ctx.import(&tree);
        let shape = VmShape::new(&ctx, node).expect("VmShape::new");

        let mut eval = VmShape::new_point_eval();
        let tape = shape.ez_point_tape();
        let (v_origin, _) = eval
            .eval(&tape, 0.0, 0.0, 0.0)
            .expect("eval at origin");
        let direct = oracle.signed_distance([0.0, 0.0, 0.0]) - 0.2;
        assert!(
            (v_origin - direct).abs() < 1e-5,
            "VM eval={}, direct={}",
            v_origin,
            direct
        );
    }

    /// Builds a thin slab mesh similar to the user's mok.stl: extent ~7×7
    /// in x/y, ~0.7 in z. Constructed as a closed box (top + bottom +
    /// 4 sides) so GWN is unambiguous.
    fn thin_slab_mesh() -> (Vec<[f32; 3]>, Vec<[u32; 3]>) {
        let h = 0.334;
        let w = 3.5;
        let v: Vec<[f32; 3]> = vec![
            [-w, -w, -h], // 0
            [w, -w, -h],  // 1
            [w, w, -h],   // 2
            [-w, w, -h],  // 3
            [-w, -w, h],  // 4
            [w, -w, h],   // 5
            [w, w, h],    // 6
            [-w, w, h],   // 7
        ];
        let t: Vec<[u32; 3]> = vec![
            [0, 2, 1], [0, 3, 2],   // -z
            [4, 5, 6], [4, 6, 7],   // +z
            [0, 1, 5], [0, 5, 4],   // -y
            [3, 7, 6], [3, 6, 2],   // +y
            [0, 4, 7], [0, 7, 3],   // -x
            [1, 2, 6], [1, 6, 5],   // +x
        ];
        (v, t)
    }

    #[test]
    fn thin_slab_oracle_sanity() {
        // Sanity-check the oracle directly at points whose true SDF we know
        // analytically, so we can localize a bug to BSP / GWN / sign before
        // looking at the octree.
        let (verts, tris) = thin_slab_mesh();
        let oracle = TriBspOracle::new(verts, tris);

        // Inside the slab at the origin: should be ~ -0.334 (distance to top
        // or bottom face).
        let d_origin = oracle.signed_distance([0.0, 0.0, 0.0]);
        let gwn_origin = oracle.generalized_winding_number([0.0, 0.0, 0.0]);
        eprintln!(
            "  d(origin) = {:.4}, gwn = {:.4}, expected ~ -0.334",
            d_origin, gwn_origin
        );
        assert!(d_origin < -0.3 && d_origin > -0.4, "d(origin) was {}", d_origin);

        // Above the slab at z=1.0: should be ~ +0.666.
        let d_above = oracle.signed_distance([0.0, 0.0, 1.0]);
        let gwn_above = oracle.generalized_winding_number([0.0, 0.0, 1.0]);
        eprintln!(
            "  d(0,0,1) = {:.4}, gwn = {:.4}, expected ~ +0.666",
            d_above, gwn_above
        );
        assert!(d_above > 0.6 && d_above < 0.7, "d(0,0,1) was {}", d_above);

        // Far outside at (0,0,5): should be ~ +4.666.
        let d_far = oracle.signed_distance([0.0, 0.0, 5.0]);
        let gwn_far = oracle.generalized_winding_number([0.0, 0.0, 5.0]);
        eprintln!(
            "  d(0,0,5) = {:.4}, gwn = {:.4}, expected ~ +4.666",
            d_far, gwn_far
        );
        assert!(d_far > 4.6 && d_far < 4.7, "d(0,0,5) was {}", d_far);

        // Near the surface at (0,0,0.4): should be ~ +0.066 (just outside top).
        let d_near = oracle.signed_distance([0.0, 0.0, 0.4]);
        eprintln!(
            "  d(0,0,0.4) = {:.4}, expected ~ +0.066",
            d_near
        );
        assert!(d_near > 0.05 && d_near < 0.08, "d(0,0,0.4) was {}", d_near);
    }

    #[test]
    fn thin_slab_offset_produces_non_empty_octree() {
        // Mimics the user-reported failure: a thin slab mesh with `bvhSdf - 1.0`
        // (the inflated shell at distance 1). Verify the octree builds and walks
        // a non-empty mesh.
        use fidget_mesh::{Octree, Settings as MeshSettings};
        use nalgebra::Vector3;
        use fidget::gui::View3;

        let (verts, tris) = thin_slab_mesh();
        let oracle = Arc::new(TriBspOracle::new(verts, tris));
        mesh_oracle::register(7, oracle);

        // Build the tree the same way the FFI does: oracle leaf augmented with
        // (x-x)+(y-y)+(z-z) so the tape's var_map includes Var::X/Y/Z. See
        // tree_mesh_oracle for the rationale.
        let (x, y, z) = Tree::axes();
        let zero_via_axes = (x.clone() - x) + (y.clone() - y) + (z.clone() - z);
        let augmented = Tree::mesh_oracle(7) + zero_via_axes;
        let tree = augmented - Tree::constant(1.0);
        let mut ctx = fidget::Context::new();
        let node = ctx.import(&tree);
        assert!(ctx.contains_mesh_oracle(node), "tree must contain MeshOracle");

        let shape = VmShape::new(&ctx, node).expect("VmShape::new");

        let view = View3::from_center_and_scale(
            Vector3::new(0.0_f32, 0.0, 0.0),
            3.0_f32,
        );
        let settings = MeshSettings {
            depth: 6,
            world_to_model: view.world_to_model(),
            ..Default::default()
        };

        let vars = fidget::shape::ShapeVars::<f32>::new();
        let bound = shape.bind(&vars).expect("bind");
        let octree = Octree::build(&bound, &settings).expect("Octree::build");
        let mesh = octree.walk_dual();
        eprintln!(
            "thin_slab: produced mesh with {} verts, {} tris",
            mesh.vertices.len(),
            mesh.triangles.len()
        );
        assert!(
            !mesh.triangles.is_empty(),
            "expected non-empty mesh for inflated thin slab"
        );
    }

    /// Builds a tessellated thin slab so we can measure the Fast GWN
    /// improvement on a triangle-soup with non-trivial size. Each face of
    /// the box is split into a grid of `subdiv × subdiv` tris.
    fn dense_slab_mesh(subdiv: u32) -> (Vec<[f32; 3]>, Vec<[u32; 3]>) {
        let h = 0.334;
        let w = 3.5;
        let mut verts = Vec::new();
        let mut tris = Vec::new();
        let add_face =
            |corners: [[f32; 3]; 4], flip: bool, verts: &mut Vec<[f32; 3]>, tris: &mut Vec<[u32; 3]>| {
                // Subdivide a quad into a grid of `subdiv × subdiv` cells.
                let base = verts.len() as u32;
                let n = subdiv as usize;
                let np1 = (subdiv + 1) as u32;
                for i in 0..=n {
                    for j in 0..=n {
                        let s = i as f32 / subdiv as f32;
                        let t = j as f32 / subdiv as f32;
                        // Bilinear blend of the 4 corners (00 10 11 01).
                        let p = [
                            (1.0 - s) * (1.0 - t) * corners[0][0]
                                + s * (1.0 - t) * corners[1][0]
                                + s * t * corners[2][0]
                                + (1.0 - s) * t * corners[3][0],
                            (1.0 - s) * (1.0 - t) * corners[0][1]
                                + s * (1.0 - t) * corners[1][1]
                                + s * t * corners[2][1]
                                + (1.0 - s) * t * corners[3][1],
                            (1.0 - s) * (1.0 - t) * corners[0][2]
                                + s * (1.0 - t) * corners[1][2]
                                + s * t * corners[2][2]
                                + (1.0 - s) * t * corners[3][2],
                        ];
                        verts.push(p);
                    }
                }
                for i in 0..subdiv {
                    for j in 0..subdiv {
                        let v00 = base + i * np1 + j;
                        let v10 = base + (i + 1) * np1 + j;
                        let v11 = base + (i + 1) * np1 + (j + 1);
                        let v01 = base + i * np1 + (j + 1);
                        if flip {
                            tris.push([v00, v01, v11]);
                            tris.push([v00, v11, v10]);
                        } else {
                            tris.push([v00, v11, v01]);
                            tris.push([v00, v10, v11]);
                        }
                    }
                }
            };
        // -z (looking from -z, CCW is in +x +y order)
        add_face(
            [[-w, -w, -h], [w, -w, -h], [w, w, -h], [-w, w, -h]],
            true,
            &mut verts,
            &mut tris,
        );
        // +z
        add_face(
            [[-w, -w, h], [w, -w, h], [w, w, h], [-w, w, h]],
            false,
            &mut verts,
            &mut tris,
        );
        // -y
        add_face(
            [[-w, -w, -h], [w, -w, -h], [w, -w, h], [-w, -w, h]],
            false,
            &mut verts,
            &mut tris,
        );
        // +y
        add_face(
            [[-w, w, -h], [w, w, -h], [w, w, h], [-w, w, h]],
            true,
            &mut verts,
            &mut tris,
        );
        // -x
        add_face(
            [[-w, -w, -h], [-w, w, -h], [-w, w, h], [-w, -w, h]],
            true,
            &mut verts,
            &mut tris,
        );
        // +x
        add_face(
            [[w, -w, -h], [w, w, -h], [w, w, h], [w, -w, h]],
            false,
            &mut verts,
            &mut tris,
        );
        (verts, tris)
    }

    /// Depth-8 perf measurement. Same setup as `fast_gwn_perf_floor` but at
    /// the resolution the user actually wants. Marked `#[ignore]` because it
    /// takes ~30s and we don't want to slow down the default test run.
    #[test]
    #[ignore]
    fn fast_gwn_perf_depth8() {
        use fidget::gui::View3;
        use fidget_mesh::{Octree, Settings as MeshSettings};
        use nalgebra::Vector3;
        use std::time::Instant;

        let (verts, tris) = dense_slab_mesh(64);
        let tri_count = tris.len();
        let oracle = Arc::new(TriBspOracle::new(verts, tris));
        mesh_oracle::register(199, oracle);

        let (x, y, z) = Tree::axes();
        let zero_via_axes =
            (x.clone() - x) + (y.clone() - y) + (z.clone() - z);
        let tree =
            (Tree::mesh_oracle(199) + zero_via_axes) - Tree::constant(1.0);

        let mut ctx = fidget::Context::new();
        let node = ctx.import(&tree);
        let shape = VmShape::new(&ctx, node).unwrap();

        let view = View3::from_center_and_scale(
            Vector3::new(0.0_f32, 0.0, 0.0),
            3.0_f32,
        );
        let settings = MeshSettings {
            depth: 8,
            world_to_model: view.world_to_model(),
            ..Default::default()
        };

        let start = Instant::now();
        let vars = fidget::shape::ShapeVars::<f32>::new();
        let bound = shape.bind(&vars).unwrap();
        let octree = Octree::build(&bound, &settings).unwrap();
        let octree_elapsed = start.elapsed();
        let walk_start = Instant::now();
        let mesh = octree.walk_dual();
        let walk_elapsed = walk_start.elapsed();
        eprintln!(
            "depth8: {} tris -> octree {:.2}s walk {:.2}s, mesh: {} verts {} tris",
            tri_count,
            octree_elapsed.as_secs_f64(),
            walk_elapsed.as_secs_f64(),
            mesh.vertices.len(),
            mesh.triangles.len()
        );
    }

    /// Sanity perf-floor for Fast GWN: a depth-7 mesh-out on a ~49k-tri slab
    /// completes in well under a minute. The assertion only fires in release
    /// builds (where production actually runs) — the debug build is roughly
    /// 13× slower on the hot loop and would spuriously fail a tight bound.
    #[test]
    fn fast_gwn_perf_floor() {
        use fidget::gui::View3;
        use fidget_mesh::{Octree, Settings as MeshSettings};
        use nalgebra::Vector3;
        use std::time::Instant;

        // 64-subdiv face = 64*64*2 tris per face × 6 faces = 49 152 tris.
        let (verts, tris) = dense_slab_mesh(64);
        let tri_count = tris.len();
        let oracle = Arc::new(TriBspOracle::new(verts, tris));
        mesh_oracle::register(99, oracle);

        let (x, y, z) = Tree::axes();
        let zero_via_axes =
            (x.clone() - x) + (y.clone() - y) + (z.clone() - z);
        let tree =
            (Tree::mesh_oracle(99) + zero_via_axes) - Tree::constant(1.0);

        let mut ctx = fidget::Context::new();
        let node = ctx.import(&tree);
        let shape = VmShape::new(&ctx, node).unwrap();

        let view = View3::from_center_and_scale(
            Vector3::new(0.0_f32, 0.0, 0.0),
            3.0_f32,
        );
        let settings = MeshSettings {
            depth: 7,
            world_to_model: view.world_to_model(),
            ..Default::default()
        };

        let start = Instant::now();
        let vars = fidget::shape::ShapeVars::<f32>::new();
        let bound = shape.bind(&vars).expect("bind");
        let octree = Octree::build(&bound, &settings).expect("Octree::build");
        let octree_elapsed = start.elapsed();
        let walk_start = Instant::now();
        let mesh = octree.walk_dual();
        let walk_elapsed = walk_start.elapsed();
        eprintln!(
            "fast_gwn_perf_floor: {} tris -> octree {:.2}s walk {:.2}s, mesh: {} verts {} tris",
            tri_count,
            octree_elapsed.as_secs_f64(),
            walk_elapsed.as_secs_f64(),
            mesh.vertices.len(),
            mesh.triangles.len()
        );
        assert!(
            !mesh.triangles.is_empty(),
            "expected non-empty mesh"
        );
        #[cfg(not(debug_assertions))]
        assert!(
            octree_elapsed.as_secs() < 30,
            "depth-7 octree on a 49k-tri slab should be <30s in release, got {:.2}s",
            octree_elapsed.as_secs_f64()
        );
    }

    #[test]
    fn interval_eval_brackets_corner_min_max() {
        // The interval bound must contain the true distance at every point in
        // the box. Pick a box that straddles a face — the true min/max are
        // determined by the corner samples — and verify the returned interval
        // contains every corner distance.
        let (verts, tris) = unit_cube_mesh();
        let oracle = TriBspOracle::new(verts, tris);
        let x = Interval::new(0.5, 1.5);
        let y = Interval::new(-0.3, 0.3);
        let z = Interval::new(-0.3, 0.3);
        let i = oracle.interval(x, y, z);
        // Sample interior corners; all of them must lie inside the interval.
        for &cx in &[0.5_f32, 1.5] {
            for &cy in &[-0.3_f32, 0.3] {
                for &cz in &[-0.3_f32, 0.3] {
                    let d = oracle.signed_distance([cx, cy, cz]);
                    assert!(
                        d >= i.lower() && d <= i.upper(),
                        "corner ({},{},{}) d={} outside interval [{},{}]",
                        cx,
                        cy,
                        cz,
                        d,
                        i.lower(),
                        i.upper()
                    );
                }
            }
        }
    }

}

