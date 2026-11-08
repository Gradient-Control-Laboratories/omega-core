package fidget

import (
	"testing"

	"google.golang.org/protobuf/proto"
	tape "gradientcontrol.com/omega/core/internal/fidget/internal/tape"
)

// Build a unit-sphere SDF (sqrt(x^2+y^2+z^2) - 1) via the tree FFI, serialize it
// to an omega tape, and verify the bytes decode as a tape.Tape whose single
// output is the SUB node — i.e. the Go cgo SerializeTape wrapper and the
// generated Go tape bindings agree on the wire format.
func TestSerializeTapeSphere(t *testing.T) {
	Lock()
	defer Unlock()
	Clear() // seeds X=0, Y=1, Z=2

	x, y, z := NewX(), NewY(), NewZ()
	s := Add(Add(Square(x), Square(y)), Square(z))
	r := Sqrt(s)
	one := NewConst(1.0)
	sphere := Sub(r, one)

	bytes, err := SerializeTape([]TreeIndex{sphere})
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
