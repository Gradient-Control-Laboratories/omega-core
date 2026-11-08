package meshplan

import "fmt"

// Plan node kind values.
const (
	MeshPlanKindAsset = "asset"
	MeshPlanKindSdf   = "sdf"
	MeshPlanKindBool  = "bool"
)

// Boolean CSG operations for boolean plan nodes.
const (
	MeshPlanOpUnion     = "union"
	MeshPlanOpIntersect = "intersect"
	MeshPlanOpSubtract  = "subtract"
)

// MeshExportPlan describes the boolean tree compiled from GSL mesh operations.
type MeshExportPlan struct {
	Root *MeshPlanNode `json:"root"`
	// BodyNames names each top-level return, keyed by its ShapeIndex, as
	// program.OperandName does, a repeat suffixed -2, -3, ...; Bodies calls a
	// return missing here "body-N".
	BodyNames map[int]string `json:"bodyNames,omitempty"`
}

// MeshPlanBody is one top-level return of a plan: its own subtree, which an
// export meshes as a separate solid instead of welding it into the union.
type MeshPlanBody struct {
	ShapeIndex int
	Name       string
	Root       *MeshPlanNode
}

// Bodies splits Root back into its top-level returns. SliceProgram joins the
// returns with unions that belong to no return (ShapeIndex 0); every node of
// a return carries its index, so the first indexed node down each synthetic
// union is a body's root. Bodies come back in return order.
func (p *MeshExportPlan) Bodies() []MeshPlanBody {
	var bodies []MeshPlanBody
	var walk func(n *MeshPlanNode)
	walk = func(n *MeshPlanNode) {
		if n == nil {
			return
		}
		if n.ShapeIndex == 0 && n.Kind == MeshPlanKindBool && n.Op == MeshPlanOpUnion {
			walk(n.Left)
			walk(n.Right)
			return
		}
		name := p.BodyNames[n.ShapeIndex]
		if name == "" {
			name = fmt.Sprintf("body-%d", n.ShapeIndex)
		}
		bodies = append(bodies, MeshPlanBody{ShapeIndex: n.ShapeIndex, Name: name, Root: n})
	}
	walk(p.Root)
	return bodies
}

// KeepBodies narrows the plan to the returns whose ShapeIndex is in keep,
// joined as SliceProgram joins them. It reports false, leaving the plan
// alone, when that would keep nothing.
func (p *MeshExportPlan) KeepBodies(keep map[int]bool) bool {
	var root *MeshPlanNode
	for _, b := range p.Bodies() {
		if keep[b.ShapeIndex] {
			root = UnionPlanNodes(root, b.Root)
		}
	}
	if root == nil {
		return false
	}
	p.Root = root
	return true
}

// UnionPlanNodes folds a new node into a meshUnion with an existing plan root.
func UnionPlanNodes(root, addition *MeshPlanNode) *MeshPlanNode {
	if root == nil {
		return addition
	}
	if addition == nil {
		return root
	}
	return &MeshPlanNode{
		Kind:  MeshPlanKindBool,
		Op:    MeshPlanOpUnion,
		Left:  root,
		Right: addition,
	}
}

// MeshPlanNode is a single tagged node in a MeshExportPlan tree.
type MeshPlanNode struct {
	Kind string `json:"kind"`

	// ShapeIndex is the 1-based shape this node belongs to (matching top-level return slot).
	ShapeIndex int `json:"shapeIndex,omitempty"`

	// Kind == "asset": index of imported mesh asset and optional placement transform.
	// Deliberately omits omitempty because 0 is a valid asset index.
	AssetIndex int            `json:"assetIndex"`
	Transform  *MeshTransform `json:"transform,omitempty"`

	// Kind == "sdf": identifier for the carved-out analytical SDF subtree.
	// StlUrl is populated downstream when the backend meshing response is received.
	SubtreeID string `json:"subtreeID,omitempty"`
	StlUrl    string `json:"stlUrl,omitempty"`

	// Kind == "bool": CSG boolean operation joining Left and Right subtrees.
	Op    string        `json:"op,omitempty"`
	Left  *MeshPlanNode `json:"left,omitempty"`
	Right *MeshPlanNode `json:"right,omitempty"`
}

// MeshTransform carries an AssetLeaf's 4x4 column-major placement matrix.
type MeshTransform struct {
	Matrix []float64 `json:"matrix,omitempty"`
}

// MeshDecl describes an evaluated mesh(...) asset declaration.
type MeshDecl struct {
	// Index identifies the mesh asset by its zero-based session import order.
	Index int `json:"index"`

	// Matrix carries the 4x4 homogeneous placement transform (16 float64 values,
	// column-major order). An empty slice represents identity placement.
	Matrix []float64 `json:"matrix,omitempty"`

	// ShapeIndex is the 1-based shape this mesh belongs to, matching the top-level
	// return slot. Zero if not associated with a specific return shape.
	ShapeIndex int `json:"shapeIndex,omitempty"`
}

// MeshDecls extracts all placed mesh asset declarations from the plan tree.
func (p *MeshExportPlan) MeshDecls() []MeshDecl {
	// Invariant: callers invoke MeshDecls on valid MeshExportPlan with non-nil Root.
	var decls []MeshDecl
	var walk func(n *MeshPlanNode)
	walk = func(n *MeshPlanNode) {
		switch n.Kind {
		case MeshPlanKindAsset:
			var m []float64
			if n.Transform != nil && len(n.Transform.Matrix) > 0 {
				m = append([]float64(nil), n.Transform.Matrix...)
			}
			decls = append(decls, MeshDecl{
				Index:      n.AssetIndex,
				Matrix:     m,
				ShapeIndex: n.ShapeIndex,
			})
		case MeshPlanKindBool:
			// Invariant: boolean CSG nodes always possess Left and Right subtrees.
			walk(n.Left)
			walk(n.Right)
		}
	}
	walk(p.Root)
	return decls
}

// HasImplicit reports whether the plan tree contains any analytical SDF leaf nodes.
func (p *MeshExportPlan) HasImplicit() bool {
	// Invariant: callers invoke HasImplicit on valid MeshExportPlan with non-nil Root.
	var check func(n *MeshPlanNode) bool
	check = func(n *MeshPlanNode) bool {
		switch n.Kind {
		case MeshPlanKindSdf:
			return true
		case MeshPlanKindBool:
			// Invariant: boolean CSG nodes always possess Left and Right subtrees.
			return check(n.Left) || check(n.Right)
		default:
			return false
		}
	}
	return check(p.Root)
}

