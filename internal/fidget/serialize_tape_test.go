package fidget_test

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/backend/tape"
	"gradientcontrol.com/omega/core/internal/fidget"
)

// Build a unit-sphere SDF (sqrt(x^2+y^2+z^2) - 1) via the tree FFI, serialize it
// to an omega tape, and verify the bytes decode as a tape.Tape whose single
// output is the SUB node — i.e. the Go cgo SerializeTape wrapper and the
// generated Go tape bindings agree on the wire format.
func TestSerializeTapeSphere(t *testing.T) {
	fidget.Lock()
	defer fidget.Unlock()
	fidget.Clear() // seeds X=0, Y=1, Z=2

	x, y, z := fidget.NewX(), fidget.NewY(), fidget.NewZ()
	s := fidget.Add(fidget.Add(fidget.Square(x), fidget.Square(y)), fidget.Square(z))
	r := fidget.Sqrt(s)
	one := fidget.NewConst(1.0)
	sphere := fidget.Sub(r, one)

	bytes, err := fidget.SerializeTape([]fidget.TreeIndex{sphere})
	if err != nil {
		t.Fatalf("SerializeTape: %v", err)
	}

	var tp tape.Tape
	if err := proto.Unmarshal(bytes, &tp); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if len(tp.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(tp.Outputs))
	}
	if len(tp.Nodes) == 0 {
		t.Fatalf("expected nodes, got 0")
	}

	out := tp.Nodes[tp.Outputs[0]]
	bin := out.GetBinary()
	if bin == nil {
		t.Fatalf("output node is not a binary op: %v", out)
	}
	if bin.GetOp() != tape.BinaryOp_SUB {
		t.Fatalf("output op = %v, want SUB", bin.GetOp())
	}
}
