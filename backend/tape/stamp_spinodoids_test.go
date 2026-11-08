package tape

import (
	"testing"

	"google.golang.org/protobuf/proto"
)

// A field driver travels as nodes appended to the tape. The splice has to
// shift every reference by where the nodes land, leave the existing nodes and
// outputs alone, and hand the payload the drivers' new indices — the
// orientation reordered from the column-major order GSL hands it over in to
// the row-major order the proto promises.
func TestStampSpinodoidsSplicesDriverNodes(t *testing.T) {
	main := &Tape{
		Version: 1,
		Nodes: []*Node{
			{N: &Node_MeshOracle{MeshOracle: 0xE05B2000}},
		},
		Outputs: []uint32{0},
	}

	// The driver tape: x, then scale = x * 0.5, then nine orientation
	// components that are constants 0..8 in column-major order.
	sub := &Tape{Version: 1}
	add := func(n *Node) uint32 {
		sub.Nodes = append(sub.Nodes, n)
		return uint32(len(sub.Nodes) - 1)
	}
	x := add(&Node{N: &Node_Axis{Axis: Axis_X}})
	half := add(&Node{N: &Node_Constant{Constant: 0.5}})
	scale := add(&Node{N: &Node_Binary{Binary: &Binary{Op: BinaryOp_MUL, Lhs: x, Rhs: half}}})
	sub.Outputs = append(sub.Outputs, scale)
	for k := range 9 {
		sub.Outputs = append(sub.Outputs, add(&Node{N: &Node_Constant{Constant: float64(k)}}))
	}
	subBytes, err := proto.Marshal(sub)
	if err != nil {
		t.Fatal(err)
	}

	err = StampSpinodoids(main, []SpinodoidOracle{{
		AssetID: 0xE05B2000, Period: 2, DriverFingerprint: "fp", DriverTape: subBytes,
	}})
	if err != nil {
		t.Fatalf("StampSpinodoids: %v", err)
	}

	if len(main.Outputs) != 1 || main.Outputs[0] != 0 {
		t.Fatalf("outputs moved: %v", main.Outputs)
	}
	if got, want := len(main.Nodes), 1+len(sub.Nodes); got != want {
		t.Fatalf("got %d nodes, want %d", got, want)
	}
	p := main.Oracles[0].GetSpinodoid()
	if len(p.ScaleNodes) != 1 {
		t.Fatalf("scaleNodes %v", p.ScaleNodes)
	}
	mul := main.Nodes[p.ScaleNodes[0]].GetBinary()
	if mul == nil || mul.Op != BinaryOp_MUL {
		t.Fatalf("scale node is %v, want the MUL", main.Nodes[p.ScaleNodes[0]])
	}
	if main.Nodes[mul.Lhs].GetN().(*Node_Axis) == nil || main.Nodes[mul.Rhs].GetConstant() != 0.5 {
		t.Fatalf("MUL operands not shifted: lhs %v rhs %v", main.Nodes[mul.Lhs], main.Nodes[mul.Rhs])
	}
	// Column-major constant k sits at (k%3, k/3); row-major slot 3i+j must
	// hold element (i, j), i.e. constant 3j+i.
	for i := range 3 {
		for j := range 3 {
			got := main.Nodes[p.OrientationNodes[3*i+j]].GetConstant()
			if want := float64(3*j + i); got != want {
				t.Errorf("orientation (%d,%d) = %v, want %v", i, j, got, want)
			}
		}
	}
}
