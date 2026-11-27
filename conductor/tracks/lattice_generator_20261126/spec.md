# Lattices Generated From Points

## 1. Idea

Today each lattice is hand-built: a few struts or walls placed in one cell and mirror-folded. Generate them from
the point lattice instead. Given a lattice's points (simple cubic, BCC, FCC, HCP, or any periodic set with a
motif), one function gives the **Voronoi walls**, the exact distance to the bisector planes of the nearest
neighbours, and another gives the **bond struts**. Each takes a `combine`. Every primal and dual beam/foam pair
then comes from one definition, and a dual is never missing.

## 2. Why

The [API cleanup](../api_cleanup_20261126/spec.md) (section 6a) found the hand-built set hard to read: a
foam and a beam lattice of the same name are sometimes the same complex (`foam.bcc` holds `beam.bcc`'s struts,
`foam.cubic` holds `beam.cubic`'s) and sometimes neither (`foam.fcc` sits on the other face-centred lattice of the
cell), duals hide under other names (`beam.fluorite` is the rhombic-dodecahedral strut lattice, the dual of
`foam.fcc`), and two Voronoi foams were missing until that track added `foam.kelvin` and
`foam.rhombicDodecahedral`. A generator makes the
relation explicit: struts are Delaunay edges, walls are Voronoi faces.

## 3. Cost

Measured in generated WGSL operations (beyond a bare sphere), 2026-11-26:

| Variant | Ops | Strut or plane evaluations per sample |
| :------ | --: | :------------------------------------ |
| `beam.bcc`, folded | 30 | 1 segment |
| `foam.bcc`, sharp | 38 | 6 planes |
| Kelvin foam from points, folded (nearest point per sublattice, `abs`, `max` over sublattices) | 34 | 2 × 2 planes |
| `lattice.bcc`, 4 explicit diagonals, sharp / round | 79 / 122 | 4 segments |
| BCC bonds from points, explicit, sharp / round | 104 / 204 | 16 segments |

- Using the lattice's symmetry, a generator costs what a hand-built lattice does.
- A generic generator, with no fold, pays per neighbour: 14 planes for a BCC Voronoi cell, 12 to 16 struts per
  sample. That is the cost to weigh.
- Blending (`combine`) is what costs, folded or not: it needs every junction's parts written out.
- The sublattice choice avoids a select by taking `max` of each sublattice's signed wall distance (non-negative
  only inside its own cell). Selects lower to `lerp` and widen fidget's interval bounds; `round`/`floor`
  jumps are as today's mod folds. Fidget interval quality was not measured.

## 4. Open

- Whether the generic form is worth its per-neighbour cost against folded special cases, measured on the GPU
  frame time and on fidget meshing time rather than operation counts.
- Which point sets to support first (HCP needs a two-point motif).
- Whether the hand-built lattices become wrappers over the generator.
