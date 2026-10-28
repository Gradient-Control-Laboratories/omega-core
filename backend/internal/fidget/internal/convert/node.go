package convert

import (
	"fmt"

	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// Each TreeNode represents a GSL number or vector
// and implements the [build.Expression] interface.
// A GSL number has a single tree index,
// while a GSL vector has one tree index per element in the vector
// A GSL list of vectors is stored flattened in element-major order
// where element i occupies trees[i*elemLen : (i+1)*elemLen].
type TreeNode struct {
	op          *program.Operand
	start       *program.Location
	trees       []fidget.TreeIndex
	constValues []float64
	elements    []build.Expression

	// This is set only if the type of the node is a list of vectors
	elemLen int

	// boolean marks a node whose every tree is 0 or 1: a comparison, a !, or
	// && / || of two such nodes. A select uses it as its 0/1 weight directly;
	// any other condition is first stepped at 0 (see BuildSelect2).
	boolean bool
}

func NewTreeNode(
	op *program.Operand, start *program.Location, trees ...fidget.TreeIndex,
) (*TreeNode, error) {

	// Sanity check
	if len(trees) == 0 {
		err := fmt.Errorf("number of trees %d must be >= 1", len(trees))
		panic(program.InternalErr(err).Error())
	}

	return &TreeNode{
		op:    op,
		start: start,
		trees: trees,
	}, nil
}

func NewListVecTreeNode(
	op *program.Operand, start *program.Location, elemLen int, trees ...fidget.TreeIndex,
) (*TreeNode, error) {

	// Sanity check
	if len(trees) == 0 {
		err := fmt.Errorf("number of trees %d must be >= 1", len(trees))
		panic(program.InternalErr(err).Error())
	}
	if elemLen <= 1 {
		err := fmt.Errorf("element length %d must be > 1", elemLen)
		panic(program.InternalErr(err).Error())
	}
	listLen := len(trees) / elemLen
	if listLen*elemLen != len(trees) {
		err := fmt.Errorf("number of trees %d must be a multiple of elemLen %d", len(trees), elemLen)
		panic(program.InternalErr(err).Error())
	}

	return &TreeNode{
		op:      op,
		start:   start,
		trees:   trees,
		elemLen: elemLen,
	}, nil
}

func (n *TreeNode) Operand() *program.Operand {
	return n.op
}

func (n *TreeNode) Start() *program.Location {
	return n.start
}

func (n *TreeNode) Precedence() build.Precedence {
	return build.PrecedenceHighest
}

func (n *TreeNode) Type() *program.Type {
	if n.elemLen <= 1 {
		if len(n.trees) <= 1 {
			return program.NewScalarType()
		}
		return program.NewListType(int32(len(n.trees)), program.NewScalarType())
	}
	listLen := len(n.trees) / n.elemLen
	return program.NewMatrixType(int32(listLen), program.NewListType(int32(n.elemLen), program.NewScalarType()))
}

func (n *TreeNode) FidgetType() FidgetNumType {
	// The parser guarantees that collections have 2 or more elements.
	if n.elemLen <= 1 {
		if len(n.trees) <= 1 {
			return NumType
		}
		return VecNumType(len(n.trees))
	}
	listLen := len(n.trees) / n.elemLen
	return ListVecNumType(listLen, n.elemLen)
}

func (n *TreeNode) Trees() []fidget.TreeIndex {
	return n.trees
}

func (n *TreeNode) ConstValues() []float64 {
	return n.constValues
}

func (n *TreeNode) Elements() []build.Expression {
	return n.elements
}

// ----------------------------------------------------------------------------

// Each MultiTreeNode represents a multi-value return from a GSL function
// and implements the [build.Expression] interface.
type MultiTreeNode struct {
	nodes []*TreeNode
}

func NewMultiTreeNode(nodes ...*TreeNode) (*MultiTreeNode, error) {
	if len(nodes) < 1 {
		return nil, fmt.Errorf("invalid number of nodes: %d", len(nodes))
	}
	return &MultiTreeNode{
		nodes: nodes,
	}, nil
}

func (n *MultiTreeNode) Operand() *program.Operand {
	return n.nodes[0].Operand()
}

func (n *MultiTreeNode) Start() *program.Location {
	return n.nodes[0].Start()
}

func (n *MultiTreeNode) Precedence() build.Precedence {
	return build.PrecedenceHighest
}

func (n *MultiTreeNode) Type() *program.Type {
	return program.NewListType(int32(len(n.nodes)), program.NewScalarType())
}

func (n *MultiTreeNode) Nodes() []*TreeNode {
	return n.nodes
}
