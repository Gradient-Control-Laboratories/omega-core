package mesh

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler/meshplan"
	"github.com/Gradient-Control-Laboratories/omega-core/assembler/variation"
	"github.com/Gradient-Control-Laboratories/omega-core/backend/mesh/manifold"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

// AssetFunc returns the triangles of imported mesh slot index, as a mesh(index, ...) placement reads it: nine f32
// per triangle, three corners of xyz, unplaced.
type AssetFunc func(index int) ([]float32, error)

// PlanComposes reports whether prog's returns, as plan slices them, use what MeshBodies cannot mesh: an imported
// mesh placed with mesh, or a mesh boolean (meshUnion, meshIntersect, meshDifference). Such a program is meshed by
// its plan (MeshPlanBodies).
func PlanComposes(plan *meshplan.MeshExportPlan) bool {
	return plan != nil && plan.Root != nil && plan.MeshOps
}

// MeshPlanBodies meshes each top-level return of prog by its mesh plan, as the viewer's browser export does: a
// field part (an sdf leaf, one of subtrees) is meshed as Mesh meshes a program, an imported mesh is placed by its
// matrix, and a boolean composes its two sides with Manifold. A body that is a single field part or a single placed
// mesh is written as it is, without Manifold, which refuses a mesh that is not closed. Bodies come back in return
// order, named as the plan names them; WithBodies narrows them as it narrows MeshBodies.
//
// plan and subtrees are meshplan.SliceDocument's for prog. assets supplies the imported meshes the plan places;
// the oracles a field part reads (bvhSdf, vdbSdf) are registered as for Mesh.
func MeshPlanBodies(prog *program.Program, plan *meshplan.MeshExportPlan, subtrees []meshplan.SdfSubtree, assets AssetFunc, opts ...Option) ([]Body, error) {
	var o options
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	if plan == nil || plan.Root == nil {
		return nil, fmt.Errorf("the document returns nothing to mesh")
	}
	// The region a field part is meshed over: the one asked for, else Mesh's default.
	region := [6]float32{-10, -10, -10, 10, 10, 10}
	if o.hasBbox {
		region = o.bbox
	}
	c := &composer{prog: prog, assets: assets, opts: opts, progress: o.progressFn, region: region,
		subtrees: make(map[string]*program.Operand, len(subtrees)),
		soups:    make(map[*meshplan.MeshPlanNode][]float32)}
	for _, s := range subtrees {
		c.subtrees[s.ID] = s.Return
	}
	var bodies []Body
	for _, b := range plan.Bodies() {
		if o.bodies != nil && !o.bodies[b.ShapeIndex] {
			continue
		}
		stl, err := c.body(b.Root)
		if err != nil {
			return nil, fmt.Errorf("body %s: %w", b.Name, err)
		}
		bodies = append(bodies, Body{Name: b.Name, Index: b.ShapeIndex, STL: stl})
	}
	if len(bodies) == 0 {
		return nil, errNothingToExport
	}
	return bodies, nil
}

type composer struct {
	prog     *program.Program
	subtrees map[string]*program.Operand
	assets   AssetFunc
	opts     []Option
	progress func(stage string, completed, total uint32)
	// region is the box a field part is meshed over, and clip the box a part to be combined is closed by: region
	// grown over the body's placed meshes, so a part that cuts one cuts it through.
	region, clip [6]float32
	// soups holds each placed mesh once it is placed.
	soups map[*meshplan.MeshPlanNode][]float32
}

// body is one body's binary STL.
func (c *composer) body(root *meshplan.MeshPlanNode) ([]byte, error) {
	switch root.Kind {
	case meshplan.MeshPlanKindSdf:
		return c.leaf(root)
	case meshplan.MeshPlanKindAsset:
		soup, err := c.placed(root)
		if err != nil {
			return nil, err
		}
		return soupSTL(soup), nil
	}
	clip, err := c.clipBox(root)
	if err != nil {
		return nil, err
	}
	c.clip = clip
	s, err := c.compose(root)
	if err != nil {
		return nil, err
	}
	if s.empty() {
		return soupSTL(nil), nil
	}
	defer s.h.Release()
	return manifold.ToSTL(s.h)
}

// leaf meshes one field part, with extra options after the request's.
func (c *composer) leaf(node *meshplan.MeshPlanNode, extra ...Option) ([]byte, error) {
	ret, ok := c.subtrees[node.SubtreeID]
	if !ok || ret == nil {
		return nil, fmt.Errorf("the mesh plan names part %s, which the slicer did not carve out", node.SubtreeID)
	}
	return Mesh(variation.Program(c.prog, ret), append(append([]Option(nil), c.opts...), extra...)...)
}

// clipBox is the region grown to hold every mesh root's tree places, and a hundredth of its size more on every
// side, so a part that cuts a placed mesh does not stop on its face.
func (c *composer) clipBox(root *meshplan.MeshPlanNode) ([6]float32, error) {
	box := c.region
	var walk func(n *meshplan.MeshPlanNode) error
	walk = func(n *meshplan.MeshPlanNode) error {
		switch n.Kind {
		case meshplan.MeshPlanKindAsset:
			soup, err := c.placed(n)
			if err != nil {
				return err
			}
			for i := 0; i+2 < len(soup); i += 3 {
				for k := 0; k < 3; k++ {
					box[k] = min(box[k], soup[i+k])
					box[k+3] = max(box[k+3], soup[i+k])
				}
			}
		case meshplan.MeshPlanKindBool:
			if err := walk(n.Left); err != nil {
				return err
			}
			return walk(n.Right)
		}
		return nil
	}
	if err := walk(root); err != nil {
		return box, err
	}
	if box == c.region {
		return box, nil
	}
	margin := float32(0)
	for k := 0; k < 3; k++ {
		margin = max(margin, (box[k+3]-box[k])/100)
	}
	for k := 0; k < 3; k++ {
		box[k] -= margin
		box[k+3] += margin
	}
	return box, nil
}

// placed is one imported mesh placed by its node's matrix, as a triangle soup. A matrix that mirrors turns every
// triangle round, so the mesh still faces out.
func (c *composer) placed(node *meshplan.MeshPlanNode) ([]float32, error) {
	if soup, ok := c.soups[node]; ok {
		return soup, nil
	}
	if c.assets == nil {
		return nil, fmt.Errorf("mesh(%d): no imported meshes to place", node.AssetIndex)
	}
	src, err := c.assets(node.AssetIndex)
	if err != nil {
		return nil, err
	}
	if len(src)%9 != 0 {
		return nil, fmt.Errorf("mesh(%d): %d floats is not whole triangles", node.AssetIndex, len(src))
	}
	var m []float64
	if node.Transform != nil && len(node.Transform.Matrix) == 16 {
		m = node.Transform.Matrix
	}
	if m == nil {
		c.soups[node] = src
		return src, nil
	}
	det := m[0]*(m[5]*m[10]-m[9]*m[6]) - m[4]*(m[1]*m[10]-m[9]*m[2]) + m[8]*(m[1]*m[6]-m[5]*m[2])
	out := make([]float32, len(src))
	for t := 0; t < len(src); t += 9 {
		for k := 0; k < 3; k++ {
			corner := k
			if det < 0 && k > 0 {
				corner = 3 - k // 0, 2, 1
			}
			p := meshplan.ApplyMat4Col(m, [3]float64{float64(src[t+3*k]), float64(src[t+3*k+1]), float64(src[t+3*k+2])})
			for d := 0; d < 3; d++ {
				out[t+3*corner+d] = float32(p[d])
			}
		}
	}
	c.soups[node] = out
	return out, nil
}

// solid is a Manifold solid, or the empty solid (a part with no triangles), which Manifold will not load.
type solid struct{ h manifold.Handle }

func (s solid) empty() bool { return !s.h.Valid() }

// compose is a plan node as a solid the caller releases.
func (c *composer) compose(node *meshplan.MeshPlanNode) (solid, error) {
	switch node.Kind {
	case meshplan.MeshPlanKindSdf:
		// Closed by the clip box, over that box padded by one of its cells so the faces it closes the part with
		// fall inside the mesher's cells.
		stl, err := c.leaf(node, Bbox(c.clip), WithBboxPadCells(1), withClip(c.clip))
		if err != nil {
			return solid{}, err
		}
		verts, _, err := unpackSTLTriangles(stl)
		if err != nil {
			return solid{}, err
		}
		return load(verts, "part "+node.SubtreeID)
	case meshplan.MeshPlanKindAsset:
		soup, err := c.placed(node)
		if err != nil {
			return solid{}, err
		}
		return load(soup, fmt.Sprintf("mesh(%d)", node.AssetIndex))
	case meshplan.MeshPlanKindBool:
		left, err := c.compose(node.Left)
		if err != nil {
			return solid{}, err
		}
		right, err := c.compose(node.Right)
		if err != nil {
			left.h.Release()
			return solid{}, err
		}
		if c.progress != nil {
			c.progress("Combining...", 0, 0)
		}
		return combine(left, right, node.Op)
	}
	return solid{}, fmt.Errorf("the mesh plan has a node of unknown kind %q", node.Kind)
}

// combine is left op right; it releases both.
func combine(left, right solid, op string) (solid, error) {
	var mop manifold.Op
	switch op {
	case meshplan.MeshPlanOpUnion:
		mop = manifold.OpUnion
	case meshplan.MeshPlanOpIntersect:
		mop = manifold.OpIntersection
	case meshplan.MeshPlanOpSubtract:
		mop = manifold.OpDifference
	default:
		left.h.Release()
		right.h.Release()
		return solid{}, fmt.Errorf("the mesh plan has a boolean of unknown op %q", op)
	}
	// The empty solid: union keeps the other side, intersection is empty, and difference keeps the left.
	switch {
	case left.empty() && right.empty():
		return solid{}, nil
	case left.empty():
		if mop == manifold.OpUnion {
			return right, nil
		}
		right.h.Release()
		return solid{}, nil
	case right.empty():
		if mop == manifold.OpIntersection {
			left.h.Release()
			return solid{}, nil
		}
		return left, nil
	}
	defer left.h.Release()
	defer right.h.Release()
	h, err := manifold.Boolean(left.h, right.h, mop)
	if err != nil {
		return solid{}, fmt.Errorf("%s: %w", op, err)
	}
	return solid{h: h}, nil
}

// load is a triangle soup as a Manifold solid, its corners welded where their coordinates are equal: the meshers
// write a shared corner the same in every triangle. A soup with no triangles is the empty solid; one Manifold
// refuses (not closed, or not a manifold) is an error naming what.
func load(soup []float32, what string) (solid, error) {
	if len(soup) == 0 {
		return solid{}, nil
	}
	positions, indices := weld(soup)
	if len(indices) == 0 {
		return solid{}, nil
	}
	h, err := manifold.Load(positions, indices)
	if err != nil {
		return solid{}, fmt.Errorf("%s: Manifold cannot take it as a solid (it must be closed, each edge shared by two triangles): %w", what, err)
	}
	return solid{h: h}, nil
}

// weld is a triangle soup as positions and indices, corners joined where their f32 coordinates are bit for bit
// equal (-0 as 0), and a triangle welding collapses dropped.
func weld(soup []float32) ([]float32, []uint32) {
	var positions []float32
	indices := make([]uint32, 0, len(soup)/3)
	byKey := make(map[[3]uint32]uint32, len(soup)/6)
	for t := 0; t+9 <= len(soup); t += 9 {
		var corner [3]uint32
		for k := 0; k < 3; k++ {
			var key [3]uint32
			for d := 0; d < 3; d++ {
				v := soup[t+3*k+d]
				if v == 0 {
					v = 0
				}
				key[d] = math.Float32bits(v)
			}
			i, ok := byKey[key]
			if !ok {
				i = uint32(len(positions) / 3)
				byKey[key] = i
				positions = append(positions, math.Float32frombits(key[0]), math.Float32frombits(key[1]), math.Float32frombits(key[2]))
			}
			corner[k] = i
		}
		if corner[0] == corner[1] || corner[1] == corner[2] || corner[0] == corner[2] {
			continue
		}
		indices = append(indices, corner[0], corner[1], corner[2])
	}
	return positions, indices
}

// soupSTL is a triangle soup as a binary STL, each facet normal the cross product of its edges.
func soupSTL(soup []float32) []byte {
	n := len(soup) / 9
	buf := make([]byte, 84+50*n)
	copy(buf, "omega-core mesh plan")
	binary.LittleEndian.PutUint32(buf[80:], uint32(n))
	for t := 0; t < n; t++ {
		p := soup[9*t : 9*t+9]
		u := [3]float32{p[3] - p[0], p[4] - p[1], p[5] - p[2]}
		v := [3]float32{p[6] - p[0], p[7] - p[1], p[8] - p[2]}
		nrm := [3]float32{u[1]*v[2] - u[2]*v[1], u[2]*v[0] - u[0]*v[2], u[0]*v[1] - u[1]*v[0]}
		off := 84 + 50*t
		for k, x := range nrm {
			binary.LittleEndian.PutUint32(buf[off+4*k:], math.Float32bits(x))
		}
		for k, x := range p {
			binary.LittleEndian.PutUint32(buf[off+12+4*k:], math.Float32bits(x))
		}
	}
	return buf
}
