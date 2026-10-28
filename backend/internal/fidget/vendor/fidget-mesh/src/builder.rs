//! Mesh builder data structure and implementation
use super::{
    Mesh, Octree,
    cell::{CellIndex, CellVertex},
    dc,
    frame::Frame,
};

/// Container used during construction of a [`Mesh`]
#[derive(Default)]
pub struct MeshBuilder {
    /// Map from indexes in [`Octree::verts`](super::Octree::verts) to
    /// `out.vertices`
    ///
    /// `usize::MAX` is used a marker for an unmapped vertex
    map: Vec<usize>,
    out: Mesh,
}

impl MeshBuilder {
    pub fn take(self) -> Mesh {
        self.out
    }

    pub(crate) fn cell(&mut self, octree: &Octree, cell: CellIndex<3>) {
        dc::dc_cell(octree, cell, self);
    }

    pub(crate) fn face<F: Frame>(
        &mut self,
        octree: &Octree,
        a: CellIndex<3>,
        b: CellIndex<3>,
    ) {
        dc::dc_face::<F>(octree, a, b, self)
    }

    /// Handles four cells that share a common edge aligned on axis `T`
    ///
    /// Cells positions are in the order `[0, U, U | V, U]`, i.e. a right-handed
    /// winding about `+T` (where `T, U, V` is a right-handed coordinate frame)
    pub(crate) fn edge<F: Frame>(
        &mut self,
        octree: &Octree,
        a: CellIndex<3>,
        b: CellIndex<3>,
        c: CellIndex<3>,
        d: CellIndex<3>,
    ) {
        dc::dc_edge::<F>(octree, a, b, c, d, self)
    }

    /// Triangulates a dual polygon without folding around its edge intersection.
    pub(crate) fn quad(&mut self, vertices: [usize; 4]) {
        let p = vertices.map(|i| self.out.vertices[i]);
        let normals = [
            [(p[1] - p[0]).cross(&(p[2] - p[0])), (p[2] - p[0]).cross(&(p[3] - p[0]))],
            [(p[1] - p[0]).cross(&(p[3] - p[0])), (p[2] - p[1]).cross(&(p[3] - p[1]))],
        ];
        let normal = normals[0][0] + normals[0][1];
        let valid = normals.map(|n| n.iter().all(|v| v.dot(&normal) >= 0.0));
        let area = normals.map(|n| n[0].norm() + n[1].norm());
        let tris = if valid[1] && (!valid[0] || area[1] < area[0]) {
            [[0, 1, 3], [1, 2, 3]]
        } else {
            [[0, 1, 2], [0, 2, 3]]
        };
        for tri in tris {
            let [a, b, c] = tri.map(|i| vertices[i]);
            if a != b && b != c && a != c {
                self.out.triangles.push(nalgebra::Vector3::new(a, b, c));
            }
        }
    }

    /// Looks up the given vertex, localizing it within a cell
    ///
    /// `v` is an absolute offset into `verts`, which should be a reference to
    /// [`Octree::verts`](super::Octree::verts).
    pub(crate) fn vertex(
        &mut self,
        v: usize,
        verts: &[CellVertex<3>],
    ) -> usize {
        if v >= self.map.len() {
            self.map.resize(v + 1, usize::MAX);
        }
        match self.map[v] {
            usize::MAX => {
                let next_vert = self.out.vertices.len();
                self.out.vertices.push(verts[v].pos);
                self.map[v] = next_vert;

                next_vert
            }
            u => u,
        }
    }
}

#[cfg(test)]
mod test {
    use super::*;
    use nalgebra::Vector3;

    #[test]
    fn quad_triangulation() {
        for vertices in [[0, 1, 2, 3], [0, 3, 2, 1], [0, 1, 2, 2]] {
            let mut builder = MeshBuilder::default();
            builder.out.vertices = vec![
                Vector3::new(0.0, 0.0, 0.0), Vector3::new(2.0, 0.0, 0.0),
                Vector3::new(0.5, 0.5, 0.0), Vector3::new(0.0, 2.0, 0.0),
            ];
            builder.quad(vertices);
            let area: f32 = builder.out.triangles.iter().map(|t| {
                let [a, b, c] = [t.x, t.y, t.z].map(|i| builder.out.vertices[i]);
                (b - a).cross(&(c - a)).norm() * 0.5
            }).sum();
            assert_eq!(area, if vertices[2] == vertices[3] { 0.5 } else { 1.0 });
        }
    }
}
